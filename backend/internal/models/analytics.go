package models

import (
	"time"

	"github.com/google/uuid"
)

// DailyTicketMetrics stores pre-aggregated per-day ticket statistics per tenant.
// One row per (tenant_id, date); upserted by the nightly aggregation job.
type DailyTicketMetrics struct {
	ID                       uint      `gorm:"primarykey;autoIncrement" json:"id"`
	TenantID                 uuid.UUID `gorm:"type:uuid;not null;default:'00000000-0000-0000-0000-000000000000';uniqueIndex:idx_tenant_daily_metrics;index" json:"tenant_id"`
	Date                     time.Time `gorm:"type:date;not null;uniqueIndex:idx_tenant_daily_metrics" json:"date"`
	TicketsCreated           int       `gorm:"not null;default:0" json:"tickets_created"`
	TicketsClosed            int       `gorm:"not null;default:0" json:"tickets_closed"`
	TicketsResolved          int       `gorm:"not null;default:0" json:"tickets_resolved"`
	TicketsReopened          int       `gorm:"not null;default:0" json:"tickets_reopened"`
	AverageResolutionTime    float64   `gorm:"type:numeric(10,2);not null;default:0" json:"average_resolution_time"`
	AverageFirstResponseTime float64   `gorm:"type:numeric(10,2);not null;default:0" json:"average_first_response_time"`
	AverageAIProcessingTime  float64   `gorm:"type:numeric(10,2);not null;default:0" json:"average_ai_processing_time"`
	CreatedAt                time.Time `json:"created_at"`
}

// AgentMetrics stores per-agent performance snapshot per tenant.
// One row per (tenant_id, user_id); updated in-place.
type AgentMetrics struct {
	ID                    uint      `gorm:"primarykey;autoIncrement" json:"id"`
	TenantID              uuid.UUID `gorm:"type:uuid;not null;default:'00000000-0000-0000-0000-000000000000';uniqueIndex:idx_tenant_agent_metrics;index" json:"tenant_id"`
	UserID                uint      `gorm:"not null;uniqueIndex:idx_tenant_agent_metrics" json:"user_id"`
	TicketsAssigned       int       `gorm:"not null;default:0" json:"tickets_assigned"`
	TicketsResolved       int       `gorm:"not null;default:0" json:"tickets_resolved"`
	AverageResolutionTime float64   `gorm:"type:numeric(10,2);not null;default:0" json:"average_resolution_time"`
	AverageReplyTime      float64   `gorm:"type:numeric(10,2);not null;default:0" json:"average_reply_time"`
	AverageCustomerRating float64   `gorm:"type:numeric(3,2);not null;default:0" json:"average_customer_rating"`
	LastCalculated        time.Time `gorm:"not null" json:"last_calculated"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// AIMetrics stores pre-aggregated per-day AI performance statistics per tenant.
// One row per (tenant_id, date); upserted by the nightly aggregation job.
type AIMetrics struct {
	ID                    uint      `gorm:"primarykey;autoIncrement" json:"id"`
	TenantID              uuid.UUID `gorm:"type:uuid;not null;default:'00000000-0000-0000-0000-000000000000';uniqueIndex:idx_tenant_ai_metrics;index" json:"tenant_id"`
	Date                  time.Time `gorm:"type:date;not null;uniqueIndex:idx_tenant_ai_metrics" json:"date"`
	AnalysisGenerated     int       `gorm:"not null;default:0" json:"analysis_generated"`
	RepliesGenerated      int       `gorm:"not null;default:0" json:"replies_generated"`
	AverageConfidence     float64   `gorm:"type:numeric(5,2);not null;default:0" json:"average_confidence"`
	AverageGenerationTime float64   `gorm:"type:numeric(10,2);not null;default:0" json:"average_generation_time"`
	ApprovalRate          float64   `gorm:"type:numeric(5,2);not null;default:0" json:"approval_rate"`
	EditRate              float64   `gorm:"type:numeric(5,2);not null;default:0" json:"edit_rate"`
	RejectionRate         float64   `gorm:"type:numeric(5,2);not null;default:0" json:"rejection_rate"`
	RetryRate             float64   `gorm:"type:numeric(5,2);not null;default:0" json:"retry_rate"`
	CreatedAt             time.Time `json:"created_at"`
}

// Report stores metadata for a generated analytics report.
type Report struct {
	ID          uint      `gorm:"primarykey;autoIncrement" json:"id"`
	TenantID    uuid.UUID `gorm:"type:uuid;not null;default:'00000000-0000-0000-0000-000000000000';index" json:"tenant_id"`
	Name        string    `gorm:"type:varchar(255);not null" json:"name"`
	Type        string    `gorm:"type:varchar(50);not null" json:"type"`
	Format      string    `gorm:"type:varchar(20);not null;default:'PDF'" json:"format"`
	Status      string    `gorm:"type:varchar(20);not null;default:'PENDING'" json:"status"`
	FilePath    string    `gorm:"type:varchar(500)" json:"file_path,omitempty"`
	FileSize    int64     `gorm:"default:0" json:"file_size"`
	GeneratedBy uint      `gorm:"not null" json:"generated_by"`
	Parameters  string    `gorm:"type:text" json:"parameters,omitempty"`
	ErrorMsg    string    `gorm:"type:text" json:"error_msg,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	GeneratedByUser *User `gorm:"foreignKey:GeneratedBy" json:"generated_by_user,omitempty"`
}
