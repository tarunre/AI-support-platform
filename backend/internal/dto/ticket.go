package dto

import (
	"time"

	"github.com/google/uuid"
)

// ─── Request DTOs ─────────────────────────────────────────────────────────────

type CreateTicketRequest struct {
	Subject       string `json:"subject"       binding:"required,min=5,max=150"`
	Description   string `json:"description"   binding:"required"`
	CustomerName  string `json:"customerName"  binding:"required"`
	CustomerEmail string `json:"customerEmail" binding:"required,email"`
	Priority      string `json:"priority"      binding:"omitempty,oneof=LOW MEDIUM HIGH URGENT"`
	Category      string `json:"category"      binding:"omitempty"`
}

type UpdateTicketRequest struct {
	Subject       string `json:"subject"       binding:"omitempty,min=5,max=150"`
	Description   string `json:"description"   binding:"omitempty"`
	Priority      string `json:"priority"      binding:"omitempty,oneof=LOW MEDIUM HIGH URGENT"`
	Category      string `json:"category"      binding:"omitempty"`
	CustomerName  string `json:"customerName"  binding:"omitempty"`
	CustomerEmail string `json:"customerEmail" binding:"omitempty,email"`
}

type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=OPEN IN_PROGRESS RESOLVED CLOSED"`
}

type AssignTicketRequest struct {
	AssignedTo uint `json:"assignedTo" binding:"required"`
}

// ListTicketsQuery holds all supported query parameters for the list endpoint.
type ListTicketsQuery struct {
	Page           int    `form:"page"`
	Limit          int    `form:"limit"`
	Search         string `form:"search"`
	Status         string `form:"status"`
	Priority       string `form:"priority"`
	Category       string `form:"category"`
	SLAStatus      string `form:"sla_status"`
	AssignedTo     *uint  `form:"assigned_to"`
	CreatedBy      *uint  `form:"created_by"`
	UnassignedOnly bool   `form:"unassigned_only"`
}

// ─── Response DTOs ────────────────────────────────────────────────────────────

// TicketResponse is the public, sanitised representation of a single ticket.
type TicketResponse struct {
	ID            uuid.UUID     `json:"id"`
	TicketNumber  string        `json:"ticket_number"`
	Subject       string        `json:"subject"`
	Description   string        `json:"description"`
	Status        string        `json:"status"`
	Priority      string        `json:"priority"`
	Category      string        `json:"category"`
	Source        string        `json:"source"`
	AssignedTo    *uint         `json:"assigned_to"`
	CreatedBy     uint          `json:"created_by"`
	CustomerName  string        `json:"customer_name"`
	CustomerEmail string        `json:"customer_email"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
	Creator       *UserResponse `json:"creator,omitempty"`
	Assignee      *UserResponse `json:"assignee,omitempty"`

	// AI analysis fields — omitted when empty
	AICategory         string     `json:"ai_category,omitempty"`
	AIPriority         string     `json:"ai_priority,omitempty"`
	AISentiment        string     `json:"ai_sentiment,omitempty"`
	AITeam             string     `json:"ai_team,omitempty"`
	AIConfidence       *int       `json:"ai_confidence,omitempty"`
	AISummary          string     `json:"ai_summary,omitempty"`
	AITags             []string   `json:"ai_tags,omitempty"`
	AIProcessingStatus string     `json:"ai_processing_status"`
	ProcessedAt        *time.Time `json:"processed_at,omitempty"`

	// SLA fields — omitted when no SLA is assigned
	SLAPolicyID              *uint      `json:"sla_policy_id,omitempty"`
	FirstResponseDueAt       *time.Time `json:"first_response_due_at,omitempty"`
	ResolutionDueAt          *time.Time `json:"resolution_due_at,omitempty"`
	FirstResponseCompletedAt *time.Time `json:"first_response_completed_at,omitempty"`
	ResolvedAt               *time.Time `json:"resolved_at,omitempty"`
	SLAStatus                string     `json:"sla_status,omitempty"`
}

// ListTicketsResponse wraps a paginated ticket slice with cursor metadata.
type ListTicketsResponse struct {
	Items       []TicketResponse `json:"items"`
	TotalCount  int64            `json:"total_count"`
	CurrentPage int              `json:"current_page"`
	TotalPages  int              `json:"total_pages"`
	Limit       int              `json:"limit"`
}
