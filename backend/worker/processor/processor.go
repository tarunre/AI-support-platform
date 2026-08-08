package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ayush/supportiq/internal/events"
	"github.com/ayush/supportiq/internal/queue"
	"github.com/ayush/supportiq/internal/queue/redisqueue"
	"github.com/ayush/supportiq/internal/repositories"
	"github.com/ayush/supportiq/internal/utils"
	"github.com/google/uuid"
)

// Handler processes a single job. Each job type has its own Handler implementation.
type Handler interface {
	Handle(ctx context.Context, job queue.Job) error
}

// Processor manages the worker pool that drains the Redis queue.
type Processor struct {
	queue       queue.Queue
	redisQ      *redisqueue.Client
	jobRepo     *repositories.JobRepository
	ticketRepo  *repositories.TicketRepository
	handlers    map[string]Handler
	workerCount int
	maxRetries  int
	retryDelay  int // base delay in seconds
}

// New creates a Processor with the given queue client and handler registry.
func New(
	q queue.Queue,
	rq *redisqueue.Client,
	jobRepo *repositories.JobRepository,
	workerCount, maxRetries, retryDelay int,
) *Processor {
	return &Processor{
		queue:       q,
		redisQ:      rq,
		jobRepo:     jobRepo,
		handlers:    make(map[string]Handler),
		workerCount: workerCount,
		maxRetries:  maxRetries,
		retryDelay:  retryDelay,
	}
}

// SetTicketRepo injects the ticket repository so exhausted jobs can mark tickets as FAILED.
func (p *Processor) SetTicketRepo(r *repositories.TicketRepository) { p.ticketRepo = r }

// RegisterHandler binds a job type string to its Handler implementation.
func (p *Processor) RegisterHandler(jobType string, h Handler) {
	p.handlers[jobType] = h
}

// Start launches workerCount goroutines and the retry poller.
// It blocks until ctx is cancelled, then waits for all workers to drain.
func (p *Processor) Start(ctx context.Context) {
	utils.Logger.WithField("workers", p.workerCount).Info("Worker: Starting processor")

	go p.retryPoller(ctx)

	done := make(chan struct{}, p.workerCount)
	for i := 0; i < p.workerCount; i++ {
		go func(wid int) {
			defer func() { done <- struct{}{} }()
			p.runWorker(ctx, wid)
		}(i + 1)
	}

	<-ctx.Done()
	utils.Logger.Info("Worker: Shutdown signal received — draining workers")
	for i := 0; i < p.workerCount; i++ {
		<-done
	}
	utils.Logger.Info("Worker: All workers stopped")
}

// ─── Worker loop ─────────────────────────────────────────────────────────────

func (p *Processor) runWorker(ctx context.Context, workerID int) {
	log := utils.Logger.WithField("worker_id", workerID)
	log.Info("Worker: Started")

	for {
		job, err := p.queue.Dequeue(ctx)
		if err != nil {
			if ctx.Err() != nil {
				log.Info("Worker: Context cancelled, stopping")
				return
			}
			log.WithError(err).Warn("Worker: Dequeue error")
			time.Sleep(time.Second)
			continue
		}
		p.processJob(ctx, workerID, *job)
	}
}

func (p *Processor) processJob(ctx context.Context, workerID int, job queue.Job) {
	start := time.Now()
	log := utils.Logger.
		WithField("worker_id", workerID).
		WithField("job_type", job.Type).
		WithField("ticket_id", job.TicketID).
		WithField("db_job_id", job.DBJobID).
		WithField("retry_count", job.RetryCount)

	log.Info("Worker: Processing job")
	_ = p.jobRepo.MarkProcessing(job.DBJobID)

	handler, ok := p.handlers[job.Type]
	if !ok {
		errMsg := fmt.Sprintf("no handler registered for job type: %s", job.Type)
		log.Error("Worker: " + errMsg)
		_ = p.jobRepo.MarkFailed(job.DBJobID, errMsg)
		p.publishEvent(ctx, events.New(events.JobFailed, job.TicketID, job.DBJobID, job.Type, nil))
		return
	}

	jobCtx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()

	err := handler.Handle(jobCtx, job)
	elapsed := time.Since(start).Milliseconds()

	if err == nil {
		log.WithField("elapsed_ms", elapsed).Info("Worker: Job completed successfully")
		_ = p.jobRepo.MarkCompleted(job.DBJobID)
		p.publishEvent(ctx, events.New(events.JobCompleted, job.TicketID, job.DBJobID, job.Type, nil))
		return
	}

	log.WithError(err).Warn("Worker: Job failed")

	if job.RetryCount < p.maxRetries {
		newCount := job.RetryCount + 1
		// Exponential backoff: 5s, 10s, 20s
		delaySecs := p.retryDelay * (1 << (newCount - 1))

		_ = p.jobRepo.MarkRetrying(job.DBJobID, newCount, err.Error())

		retryJob := job
		retryJob.RetryCount = newCount
		if qErr := p.queue.EnqueueDelayed(ctx, retryJob, delaySecs); qErr != nil {
			log.WithError(qErr).Error("Worker: Failed to schedule retry")
		} else {
			log.WithField("delay_secs", delaySecs).
				WithField("retry_attempt", newCount).
				Info("Worker: Job scheduled for retry")
		}
		return
	}

	// All retries exhausted → dead letter + mark ticket as FAILED so UI shows retry button
	log.WithField("max_retries", p.maxRetries).Error("Worker: Job exhausted all retries — moving to dead letter")
	_ = p.jobRepo.MarkDead(job.DBJobID, err.Error())
	_ = p.queue.MoveToDeadLetter(ctx, job)
	if p.ticketRepo != nil && job.TicketID != "" {
		if ticketID, parseErr := uuid.Parse(job.TicketID); parseErr == nil {
			_ = p.ticketRepo.MarkAIFailed(ticketID)
		}
	}
	p.publishEvent(ctx, events.New(events.JobFailed, job.TicketID, job.DBJobID, job.Type,
		map[string]string{"error": err.Error()}))
}

// ─── Retry poller ─────────────────────────────────────────────────────────────

func (p *Processor) retryPoller(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			moved, err := p.redisQ.MoveDueRetryJobs(ctx)
			if err != nil {
				utils.Logger.WithError(err).Warn("Worker: Retry poller error")
			} else if moved > 0 {
				utils.Logger.WithField("moved", moved).Info("Worker: Moved retry jobs to main queue")
			}
		}
	}
}

// ─── Event publishing ─────────────────────────────────────────────────────────

func (p *Processor) publishEvent(ctx context.Context, evt events.Event) {
	data, err := json.Marshal(evt)
	if err != nil {
		utils.Logger.WithError(err).Warn("Worker: Failed to marshal event")
		return
	}

	pubCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := p.redisQ.PublishEvent(pubCtx, data); err != nil {
		utils.Logger.WithError(err).Warn("Worker: Failed to publish event")
	}
}
