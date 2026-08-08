package dto

import (
	"time"

	"github.com/google/uuid"
)

type ActivityResponse struct {
	ID           uint          `json:"id"`
	TicketID     uuid.UUID     `json:"ticket_id"`
	ActivityType string        `json:"activity_type"`
	OldValue     string        `json:"old_value,omitempty"`
	NewValue     string        `json:"new_value,omitempty"`
	Description  string        `json:"description"`
	CreatedAt    time.Time     `json:"created_at"`
	User         *UserResponse `json:"user,omitempty"`
}
