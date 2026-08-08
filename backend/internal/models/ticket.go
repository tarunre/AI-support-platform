package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TicketStatus string
type TicketPriority string
type TicketCategory string
type TicketSource string

const (
	TicketStatusOpen       TicketStatus = "OPEN"
	TicketStatusInProgress TicketStatus = "IN_PROGRESS"
	TicketStatusResolved   TicketStatus = "RESOLVED"
	TicketStatusClosed     TicketStatus = "CLOSED"
)

const (
	TicketPriorityLow    TicketPriority = "LOW"
	TicketPriorityMedium TicketPriority = "MEDIUM"
	TicketPriorityHigh   TicketPriority = "HIGH"
	TicketPriorityUrgent TicketPriority = "URGENT"
)

const (
	TicketCategoryGeneral     TicketCategory = "GENERAL"
	TicketCategoryEngineering TicketCategory = "ENGINEERING"
)

const (
	TicketSourceWeb   TicketSource = "WEB"
	TicketSourceEmail TicketSource = "EMAIL"
)

// validStatusTransitions defines the allowed state machine transitions.
// RESOLVED tickets may be reopened (→ OPEN). CLOSED is terminal.
var validStatusTransitions = map[TicketStatus][]TicketStatus{
	TicketStatusOpen:       {TicketStatusInProgress},
	TicketStatusInProgress: {TicketStatusResolved},
	TicketStatusResolved:   {TicketStatusClosed, TicketStatusOpen},
	TicketStatusClosed:     {},
}

// IsValidStatusTransition returns true only if the from → to transition is permitted.
func IsValidStatusTransition(from, to TicketStatus) bool {
	allowed, ok := validStatusTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

// FormatTicketNumber formats an integer into the TKT-XXXXXX display format.
func FormatTicketNumber(n int64) string {
	return fmt.Sprintf("TKT-%06d", n)
}

// AI processing status constants.
const (
	AIStatusPending    = "PENDING"
	AIStatusProcessing = "PROCESSING"
	AIStatusCompleted  = "COMPLETED"
	AIStatusFailed     = "FAILED"
)

// Ticket is the core support ticket entity.
type Ticket struct {
	TenantID      uuid.UUID      `gorm:"type:uuid;not null;default:'00000000-0000-0000-0000-000000000000';index:idx_ticket_tenant_status;index:idx_ticket_tenant_created" json:"tenant_id"`
	ID            uuid.UUID      `gorm:"type:uuid;primarykey"                            json:"id"`
	TicketNumber  string         `gorm:"type:varchar(20);not null"                       json:"ticket_number"`
	Subject       string         `gorm:"type:varchar(150);not null"                      json:"subject"`
	Description   string         `gorm:"type:text;not null"                              json:"description"`
	Status        TicketStatus   `gorm:"type:varchar(20);not null;default:'OPEN';index:idx_ticket_tenant_status"  json:"status"`
	Priority      TicketPriority `gorm:"type:varchar(20);not null;default:'MEDIUM'"      json:"priority"`
	Category      TicketCategory `gorm:"type:varchar(50);not null;default:'GENERAL'"     json:"category"`
	Source        TicketSource   `gorm:"type:varchar(20);not null;default:'WEB'"         json:"source"`
	AssignedTo    *uint          `gorm:"index"                                           json:"assigned_to"`
	CreatedBy     uint           `gorm:"not null;index"                                  json:"created_by"`
	CustomerName  string         `gorm:"type:varchar(100);not null"                      json:"customer_name"`
	CustomerEmail string         `gorm:"type:varchar(255);not null"                      json:"customer_email"`
	CreatedAt     time.Time      `gorm:"index:idx_ticket_tenant_created"                 json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index"                                           json:"-"`

	// AI analysis fields — all nullable; populated asynchronously after creation
	AICategory         string     `gorm:"type:varchar(100)"                   json:"ai_category,omitempty"`
	AIPriority         string     `gorm:"type:varchar(20)"                    json:"ai_priority,omitempty"`
	AISentiment        string     `gorm:"type:varchar(50)"                    json:"ai_sentiment,omitempty"`
	AITeam             string     `gorm:"type:varchar(100)"                   json:"ai_team,omitempty"`
	AIConfidence       *int       `gorm:"type:int"                            json:"ai_confidence,omitempty"`
	AISummary          string     `gorm:"type:text"                           json:"ai_summary,omitempty"`
	AITags             []string   `gorm:"serializer:json"                     json:"ai_tags,omitempty"`
	AIProcessingStatus string     `gorm:"type:varchar(20);default:'PENDING'" json:"ai_processing_status"`
	ProcessedAt        *time.Time `                                           json:"processed_at,omitempty"`

	// SLA fields — populated asynchronously after creation when a matching policy exists
	SLAPolicyID              *uint      `gorm:"index"                        json:"sla_policy_id,omitempty"`
	FirstResponseDueAt       *time.Time `                                     json:"first_response_due_at,omitempty"`
	ResolutionDueAt          *time.Time `                                     json:"resolution_due_at,omitempty"`
	FirstResponseCompletedAt *time.Time `                                     json:"first_response_completed_at,omitempty"`
	ResolvedAt               *time.Time `                                     json:"resolved_at,omitempty"`
	SLAStatus                string     `gorm:"type:varchar(20)"             json:"sla_status,omitempty"`

	// Associations — populated by Preload in the repository layer
	Creator  *User `gorm:"foreignKey:CreatedBy"  json:"creator,omitempty"`
	Assignee *User `gorm:"foreignKey:AssignedTo" json:"assignee,omitempty"`
}

// BeforeCreate auto-generates a UUID for every new ticket row.
func (t *Ticket) BeforeCreate(_ *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// TicketCounter holds the per-tenant monotonically-increasing sequence value
// used to generate TKT-XXXXXX numbers safely under concurrent writes.
// One row per tenant, created on first ticket creation.
type TicketCounter struct {
	TenantID  uuid.UUID `gorm:"type:uuid;primarykey"`
	LastValue int64     `gorm:"not null;default:0"`
}
