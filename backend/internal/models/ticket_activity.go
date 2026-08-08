package models

import (
	"time"

	"github.com/google/uuid"
)

// Activity type constants used for audit trail logging.
const (
	ActivityCreateTicket      = "CREATE_TICKET"
	ActivityUpdateTicket      = "UPDATE_TICKET"
	ActivityAssignTicket      = "ASSIGN_TICKET"
	ActivityTakeOwnership     = "TAKE_OWNERSHIP"
	ActivityStatusChanged     = "STATUS_CHANGED"
	ActivityPriorityChanged   = "PRIORITY_CHANGED"
	ActivityCategoryChanged   = "CATEGORY_CHANGED"
	ActivityCommentAdded      = "COMMENT_ADDED"
	ActivityInternalNoteAdded = "INTERNAL_NOTE_ADDED"
	ActivityTicketClosed        = "TICKET_CLOSED"
	ActivityTicketReopened      = "TICKET_REOPENED"
	ActivityAIAnalysisCompleted = "AI_ANALYSIS_COMPLETED"

	// AI reply workflow activity types
	ActivityReplyGenerated   = "AI_REPLY_GENERATED"
	ActivityReplyApproved    = "AI_REPLY_APPROVED"
	ActivityReplyRejected    = "AI_REPLY_REJECTED"
	ActivityReplyEdited      = "AI_REPLY_EDITED"
	ActivityReplyRegenerated = "AI_REPLY_REGENERATED"

	// SLA activity types
	ActivitySLAAssigned  = "SLA_ASSIGNED"
	ActivitySLAAtRisk    = "SLA_AT_RISK"
	ActivitySLAEscalated = "SLA_ESCALATED"
	ActivitySLABreached  = "SLA_BREACHED"
	ActivitySLACompleted = "SLA_COMPLETED"

	// Email activity types
	ActivityEmailReceived    = "EMAIL_RECEIVED"
	ActivityEmailQueued      = "EMAIL_QUEUED"
	ActivityEmailSent        = "EMAIL_SENT"
	ActivityEmailFailed      = "EMAIL_FAILED"
	ActivityEmailDelivered   = "EMAIL_DELIVERED"
	ActivityAttachmentAdded  = "ATTACHMENT_ADDED"
)

// TicketActivity is an immutable audit-log row. Never edited after creation.
type TicketActivity struct {
	ID           uint      `gorm:"primarykey;autoIncrement"  json:"id"`
	TenantID     uuid.UUID `gorm:"type:uuid;not null;default:'00000000-0000-0000-0000-000000000000';index" json:"tenant_id"`
	TicketID     uuid.UUID `gorm:"type:uuid;not null;index"  json:"ticket_id"`
	UserID       uint      `gorm:"not null"                  json:"user_id"`
	ActivityType string    `gorm:"type:varchar(50);not null" json:"activity_type"`
	OldValue     string    `gorm:"type:varchar(255)"         json:"old_value,omitempty"`
	NewValue     string    `gorm:"type:varchar(255)"         json:"new_value,omitempty"`
	Description  string    `gorm:"type:text"                 json:"description"`
	CreatedAt    time.Time `json:"created_at"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
