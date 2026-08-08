package models

import (
	"time"

	"github.com/google/uuid"
)

// CommentType distinguishes public vs. internal comments.
type CommentType string

const (
	CommentTypePublic   CommentType = "PUBLIC"
	CommentTypeInternal CommentType = "INTERNAL"
	CommentTypeCustomer CommentType = "CUSTOMER" // sent by customer via portal
)

// TicketComment stores communication history attached to a ticket.
type TicketComment struct {
	ID          uint        `gorm:"primarykey;autoIncrement" json:"id"`
	TenantID    uuid.UUID   `gorm:"type:uuid;not null;default:'00000000-0000-0000-0000-000000000000';index" json:"tenant_id"`
	TicketID    uuid.UUID   `gorm:"type:uuid;not null;index" json:"ticket_id"`
	UserID      uint        `gorm:"not null" json:"user_id"`
	Message     string      `gorm:"type:text;not null" json:"message"`
	CommentType CommentType `gorm:"type:varchar(20);not null;default:'PUBLIC'" json:"comment_type"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
