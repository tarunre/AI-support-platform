package dto

import (
	"time"

	"github.com/google/uuid"
)

// ─── Request DTOs ──────────────────────────────────────────────────────────────

type CreateSLAPolicyRequest struct {
	Name                 string `json:"name"                   binding:"required,min=2,max=255"`
	Priority             string `json:"priority"               binding:"required,oneof=LOW MEDIUM HIGH URGENT"`
	FirstResponseMinutes int    `json:"first_response_minutes" binding:"required,min=1"`
	ResolutionMinutes    int    `json:"resolution_minutes"     binding:"required,min=1"`
	IsDefault            bool   `json:"is_default"`
}

type UpdateSLAPolicyRequest struct {
	Name                 string `json:"name"                   binding:"omitempty,min=2,max=255"`
	Priority             string `json:"priority"               binding:"omitempty,oneof=LOW MEDIUM HIGH URGENT"`
	FirstResponseMinutes int    `json:"first_response_minutes" binding:"omitempty,min=1"`
	ResolutionMinutes    int    `json:"resolution_minutes"     binding:"omitempty,min=1"`
	IsDefault            *bool  `json:"is_default"`
}

// ─── Response DTOs ─────────────────────────────────────────────────────────────

type SLAPolicyResponse struct {
	ID                   uint      `json:"id"`
	TenantID             uuid.UUID `json:"tenant_id"`
	Name                 string    `json:"name"`
	Priority             string    `json:"priority"`
	FirstResponseMinutes int       `json:"first_response_minutes"`
	ResolutionMinutes    int       `json:"resolution_minutes"`
	IsDefault            bool      `json:"is_default"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// TicketSLASummary is a lightweight ticket row used in the SLA dashboard.
type TicketSLASummary struct {
	TicketID         uuid.UUID  `json:"ticket_id"`
	TicketNumber     string     `json:"ticket_number"`
	Subject          string     `json:"subject"`
	Priority         string     `json:"priority"`
	SLAStatus        string     `json:"sla_status"`
	ResolutionDueAt  *time.Time `json:"resolution_due_at,omitempty"`
	TimeRemainingMin float64    `json:"time_remaining_minutes"`
	PercentElapsed   float64    `json:"percent_elapsed"`
	AssignedTo       *uint      `json:"assigned_to,omitempty"`
}

// SLADashboardResponse is the payload returned by GET /api/v1/tickets/sla.
type SLADashboardResponse struct {
	NearBreach          []TicketSLASummary `json:"near_breach"`
	Breached            []TicketSLASummary `json:"breached"`
	AvgFirstResponseMin float64            `json:"avg_first_response_minutes"`
	AvgResolutionMin    float64            `json:"avg_resolution_minutes"`
	CompliancePercent   float64            `json:"compliance_percent"`
	TotalWithSLA        int64              `json:"total_with_sla"`
	BreachedCount       int64              `json:"breached_count"`
	CompletedOnTime     int64              `json:"completed_on_time"`
}
