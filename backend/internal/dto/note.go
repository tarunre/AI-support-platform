package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateNoteRequest struct {
	Note string `json:"note" binding:"required"`
}

type NoteResponse struct {
	ID         uint          `json:"id"`
	TicketID   uuid.UUID     `json:"ticket_id"`
	Note       string        `json:"note"`
	IsInternal bool          `json:"is_internal"`
	CreatedAt  time.Time     `json:"created_at"`
	User       *UserResponse `json:"user,omitempty"`
}
