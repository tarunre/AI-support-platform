package dto

import (
	"time"

	"github.com/google/uuid"
)

// ─── Request DTOs ─────────────────────────────────────────────────────────────

type EditReplyRequest struct {
	Reply string `json:"reply" binding:"required,min=10"`
}

// ─── Response DTOs ────────────────────────────────────────────────────────────

type ApproverResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type AIReplyResponse struct {
	ID             uint              `json:"id"`
	TicketID       uuid.UUID         `json:"ticket_id"`
	GeneratedReply string            `json:"generated_reply"`
	EditedReply    string            `json:"edited_reply,omitempty"`
	Confidence     int               `json:"confidence"`
	Status         string            `json:"status"`
	Model          string            `json:"model"`
	PromptVersion  string            `json:"prompt_version"`
	GenerationTime int64             `json:"generation_time"`
	ApprovedBy     *uint             `json:"approved_by,omitempty"`
	ApprovedAt     *time.Time        `json:"approved_at,omitempty"`
	Approver       *ApproverResponse `json:"approver,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

type ReplyHistoryResponse struct {
	Items []AIReplyResponse `json:"items"`
	Total int               `json:"total"`
}
