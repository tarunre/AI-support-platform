package events

import "time"

// Event type constants — these are the WebSocket event names sent to the frontend.
const (
	TicketAICompleted    = "ticket.ai.completed"
	TicketReplyGenerated = "ticket.reply.generated"
	TicketReplyFailed    = "ticket.reply.failed"
	TicketUpdated        = "ticket.updated"
	JobCompleted         = "job.completed"
	JobFailed            = "job.failed"

	// SLA events
	SLAUpdated  = "sla.updated"
	SLAAtRisk   = "sla.at_risk"
	SLABreached = "sla.breached"
)

// Channel is the Redis pub/sub channel used for worker → API server events.
const Channel = "events:notifications"

// Event is the payload published via Redis pub/sub and forwarded over WebSocket.
type Event struct {
	Type      string      `json:"event"`
	TicketID  string      `json:"ticket_id,omitempty"`
	JobID     uint        `json:"job_id,omitempty"`
	JobType   string      `json:"job_type,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// New creates a new Event with the current timestamp.
func New(eventType, ticketID string, jobID uint, jobType string, data interface{}) Event {
	return Event{
		Type:      eventType,
		TicketID:  ticketID,
		JobID:     jobID,
		JobType:   jobType,
		Data:      data,
		Timestamp: time.Now(),
	}
}
