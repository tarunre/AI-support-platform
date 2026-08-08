package models

import (
	"time"

	"github.com/google/uuid"
)

// SLAStatus represents the current SLA compliance state of a ticket.
type SLAStatus string

const (
	SLAStatusOnTrack   SLAStatus = "ON_TRACK"
	SLAStatusAtRisk    SLAStatus = "AT_RISK"
	SLAStatusBreached  SLAStatus = "BREACHED"
	SLAStatusCompleted SLAStatus = "COMPLETED"
)

// SLA escalation thresholds — percentage of resolution time elapsed.
const (
	SLAThresholdAtRisk   = 80  // Mark AT_RISK, notify assigned agent
	SLAThresholdEscalate = 90  // Notify team lead
	SLAThresholdBreach   = 100 // Mark BREACHED
)

// SLAPolicy defines first-response and resolution time targets for a given
// ticket priority within a tenant. One policy per priority per tenant is
// the expected configuration; a single default policy acts as the catch-all.
type SLAPolicy struct {
	ID                   uint           `gorm:"primarykey;autoIncrement"                                                           json:"id"`
	TenantID             uuid.UUID      `gorm:"type:uuid;not null;default:'00000000-0000-0000-0000-000000000000';index"            json:"tenant_id"`
	Name                 string         `gorm:"type:varchar(255);not null"                                                         json:"name"`
	Priority             TicketPriority `gorm:"type:varchar(20);not null"                                                          json:"priority"`
	FirstResponseMinutes int            `gorm:"not null;default:60"                                                                json:"first_response_minutes"`
	ResolutionMinutes    int            `gorm:"not null;default:480"                                                               json:"resolution_minutes"`
	IsDefault            bool           `gorm:"not null;default:false"                                                             json:"is_default"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
}
