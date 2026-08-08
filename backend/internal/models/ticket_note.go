package models

import (
	"time"

	"github.com/google/uuid"
)

// TicketNote stores internal notes on a ticket, visible to staff only.
type TicketNote struct {
	ID         uint      `gorm:"primarykey;autoIncrement" json:"id"`
	TenantID   uuid.UUID `gorm:"type:uuid;not null;default:'00000000-0000-0000-0000-000000000000';index" json:"tenant_id"`
	TicketID   uuid.UUID `gorm:"type:uuid;not null;index"  json:"ticket_id"`
	UserID     uint      `gorm:"not null"                  json:"user_id"`
	Note       string    `gorm:"type:text;not null"        json:"note"`
	IsInternal bool      `gorm:"not null;default:true"     json:"is_internal"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
