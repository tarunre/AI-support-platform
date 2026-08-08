package services

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/ayush/supportiq/internal/models"
	"github.com/ayush/supportiq/internal/queue"
	"github.com/ayush/supportiq/internal/repositories"
	"github.com/google/uuid"
)

// JobService creates background jobs, enqueues them to Redis, and provides
// monitoring data for the admin API.
type JobService struct {
	jobRepo *repositories.JobRepository
	queue   queue.Queue
}

// JobStatistics holds job counts grouped by status.
type JobStatistics struct {
	Queued     int64 `json:"queued"`
	Processing int64 `json:"processing"`
	Completed  int64 `json:"completed"`
	Failed     int64 `json:"failed"`
	Retrying   int64 `json:"retrying"`
	Dead       int64 `json:"dead"`
	Total      int64 `json:"total"`
}

// ListJobsQuery holds filter and pagination params for the job list endpoint.
type ListJobsQuery struct {
	Page    int    `form:"page"`
	Limit   int    `form:"limit"`
	Status  string `form:"status"`
	JobType string `form:"job_type"`
}

func NewJobService(jobRepo *repositories.JobRepository, q queue.Queue) *JobService {
	return &JobService{jobRepo: jobRepo, queue: q}
}

func (s *JobService) IsQueueAvailable() bool {
	return s.queue != nil
}

func (s *JobService) EnqueueAIAnalysis(ticketID uuid.UUID, userID uint) error {
	return s.enqueue(uuid.Nil, models.JobTypeAIAnalysis, ticketID.String(), userID)
}

func (s *JobService) EnqueueAIAnalysisForTenant(tenantID uuid.UUID, ticketID uuid.UUID, userID uint) error {
	return s.enqueue(tenantID, models.JobTypeAIAnalysis, ticketID.String(), userID)
}

func (s *JobService) EnqueueGenerateReply(ticketID uuid.UUID, userID uint) error {
	return s.enqueue(uuid.Nil, models.JobTypeGenerateReply, ticketID.String(), userID)
}

func (s *JobService) EnqueueRegenerateReply(ticketID uuid.UUID, userID uint) error {
	return s.enqueue(uuid.Nil, models.JobTypeRegenerateReply, ticketID.String(), userID)
}

func (s *JobService) RetryJob(id uint) (*models.BackgroundJob, error) {
	job, err := s.jobRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("job not found: %w", err)
	}
	if job.Status != models.JobStatusFailed && job.Status != models.JobStatusDead {
		return nil, fmt.Errorf("only FAILED or DEAD jobs can be retried; current status: %s", job.Status)
	}

	if s.queue == nil {
		return nil, fmt.Errorf("queue not available — Redis is not configured")
	}

	qJob := queue.Job{
		ID:         fmt.Sprintf("%d", job.ID),
		Type:       string(job.JobType),
		TicketID:   job.ReferenceID,
		DBJobID:    job.ID,
		RetryCount: 0,
		MaxRetries: 3,
		EnqueuedAt: time.Now(),
	}
	if err := s.queue.Enqueue(context.Background(), qJob); err != nil {
		return nil, fmt.Errorf("failed to re-enqueue: %w", err)
	}

	_ = s.jobRepo.MarkRetrying(job.ID, 0, "Manual retry initiated")
	return s.jobRepo.FindByID(id)
}

func (s *JobService) GetJob(id uint) (*models.BackgroundJob, error) {
	return s.jobRepo.FindByID(id)
}

func (s *JobService) ListJobs(q ListJobsQuery) ([]models.BackgroundJob, int64, int, error) {
	page := q.Page
	if page < 1 {
		page = 1
	}
	limit := q.Limit
	if limit < 1 || limit > 100 {
		limit = 20
	}

	jobs, total, err := s.jobRepo.List(q.Status, q.JobType, page, limit)
	if err != nil {
		return nil, 0, 0, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	if totalPages == 0 {
		totalPages = 1
	}
	return jobs, total, totalPages, nil
}

func (s *JobService) GetStatistics() (JobStatistics, error) {
	counts, err := s.jobRepo.Statistics()
	if err != nil {
		return JobStatistics{}, err
	}

	stats := JobStatistics{
		Queued:     counts[string(models.JobStatusQueued)],
		Processing: counts[string(models.JobStatusProcessing)],
		Completed:  counts[string(models.JobStatusCompleted)],
		Failed:     counts[string(models.JobStatusFailed)],
		Retrying:   counts[string(models.JobStatusRetrying)],
		Dead:       counts[string(models.JobStatusDead)],
	}
	for _, v := range counts {
		stats.Total += v
	}
	return stats, nil
}

func (s *JobService) QueueStats(ctx context.Context) (map[string]int64, error) {
	if s.queue == nil {
		return map[string]int64{"main": 0, "retry": 0, "dead": 0}, nil
	}
	main, _ := s.queue.QueueLen(ctx)
	retry, _ := s.queue.RetryQueueLen(ctx)
	dead, _ := s.queue.DeadLetterLen(ctx)
	return map[string]int64{"main": main, "retry": retry, "dead": dead}, nil
}

func (s *JobService) enqueue(tenantID uuid.UUID, jobType models.JobType, ticketID string, userID uint) error {
	dbJob := &models.BackgroundJob{
		TenantID:    tenantID,
		JobType:     jobType,
		ReferenceID: ticketID,
		Status:      models.JobStatusQueued,
	}
	if err := s.jobRepo.Create(dbJob); err != nil {
		return fmt.Errorf("failed to create job record: %w", err)
	}

	if s.queue == nil {
		return nil
	}

	qJob := queue.Job{
		ID:         fmt.Sprintf("%d", dbJob.ID),
		Type:       string(jobType),
		TicketID:   ticketID,
		UserID:     userID,
		DBJobID:    dbJob.ID,
		RetryCount: 0,
		MaxRetries: 3,
		EnqueuedAt: time.Now(),
	}
	if err := s.queue.Enqueue(context.Background(), qJob); err != nil {
		_ = s.jobRepo.MarkFailed(dbJob.ID, fmt.Sprintf("Redis enqueue failed: %s", err.Error()))
		return fmt.Errorf("failed to enqueue job: %w", err)
	}

	return nil
}
