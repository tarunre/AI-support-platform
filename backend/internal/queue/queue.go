package queue

import (
	"context"
	"encoding/json"
	"time"
)

// Job is the unit of work exchanged between the API server and the worker.
type Job struct {
	ID         string    `json:"id"`          // UUID string
	Type       string    `json:"type"`        // JobType constant
	TicketID   string    `json:"ticket_id"`   // ticket UUID string
	TenantID   string    `json:"tenant_id"`   // tenant UUID string
	UserID     uint      `json:"user_id"`
	DBJobID    uint      `json:"db_job_id"`   // background_jobs.id for status updates
	RetryCount int       `json:"retry_count"`
	MaxRetries int       `json:"max_retries"`
	EnqueuedAt time.Time `json:"enqueued_at"`
}

// RawPayload serialises a Job to JSON bytes, safe for logging (no secrets).
func (j Job) RawPayload() string {
	b, _ := json.Marshal(j)
	return string(b)
}

// Queue is the abstract interface for the job queue broker.
// Replacing Redis with another backend requires only a new implementation.
type Queue interface {
	// Enqueue adds a job to the main work queue.
	Enqueue(ctx context.Context, job Job) error

	// Dequeue blocks until a job is available and returns it.
	// The context cancellation stops the blocking wait.
	Dequeue(ctx context.Context) (*Job, error)

	// EnqueueDelayed adds a job to the retry queue to be processed after delaySeconds.
	EnqueueDelayed(ctx context.Context, job Job, delaySeconds int) error

	// MoveToDeadLetter archives an exhausted job in the dead-letter queue.
	MoveToDeadLetter(ctx context.Context, job Job) error

	// QueueLen returns the number of pending jobs in the main queue.
	QueueLen(ctx context.Context) (int64, error)

	// RetryQueueLen returns the number of delayed jobs awaiting retry.
	RetryQueueLen(ctx context.Context) (int64, error)

	// DeadLetterLen returns the number of permanently failed jobs.
	DeadLetterLen(ctx context.Context) (int64, error)

	// Close releases any resources held by the queue implementation.
	Close() error
}
