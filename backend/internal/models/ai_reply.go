package models

import (
	"time"

	"github.com/google/uuid"
)

// AIReplyStatus represents the lifecycle state of an AI-generated reply draft.
type AIReplyStatus string

const (
	AIReplyStatusGenerated   AIReplyStatus = "GENERATED"
	AIReplyStatusApproved    AIReplyStatus = "APPROVED"
	AIReplyStatusRejected    AIReplyStatus = "REJECTED"
	AIReplyStatusRegenerated AIReplyStatus = "REGENERATED"
	AIReplyStatusSent        AIReplyStatus = "SENT"
)

// AIReply is a versioned AI-generated reply draft for a support ticket.
// Every generation creates a new record; previous versions are never deleted.
type AIReply struct {
	ID             uint          `gorm:"primarykey;autoIncrement" json:"id"`
	TenantID       uuid.UUID     `gorm:"type:uuid;not null;default:'00000000-0000-0000-0000-000000000000';index" json:"tenant_id"`
	TicketID       uuid.UUID     `gorm:"type:uuid;not null;index" json:"ticket_id"`
	GeneratedReply string        `gorm:"type:text;not null" json:"generated_reply"`
	EditedReply    string        `gorm:"type:text" json:"edited_reply,omitempty"`
	Confidence     int           `gorm:"not null" json:"confidence"`
	Status         AIReplyStatus `gorm:"type:varchar(20);not null;default:'GENERATED'" json:"status"`
	Model          string        `gorm:"type:varchar(100)" json:"model"`
	PromptVersion  string        `gorm:"type:varchar(20)" json:"prompt_version"`
	GenerationTime int64         `gorm:"not null" json:"generation_time"` // milliseconds
	ApprovedBy     *uint         `json:"approved_by,omitempty"`
	ApprovedAt     *time.Time    `json:"approved_at,omitempty"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`

	Approver *User `gorm:"foreignKey:ApprovedBy" json:"approver,omitempty"`
}
