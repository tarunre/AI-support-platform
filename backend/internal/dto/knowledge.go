package dto

import "time"

// ─── Request DTOs ─────────────────────────────────────────────────────────────

type CreateKnowledgeRequest struct {
	Title    string `json:"title"    binding:"required,min=3,max=255"`
	Category string `json:"category" binding:"required"`
	Content  string `json:"content"  binding:"required,min=10"`
}

type UpdateKnowledgeRequest struct {
	Title    string `json:"title"     binding:"omitempty,min=3,max=255"`
	Category string `json:"category"  binding:"omitempty"`
	Content  string `json:"content"   binding:"omitempty,min=10"`
	IsActive *bool  `json:"is_active"`
}

type ListKnowledgeQuery struct {
	Page       int    `form:"page"`
	Limit      int    `form:"limit"`
	Search     string `form:"search"`
	Category   string `form:"category"`
	ActiveOnly bool   `form:"active_only"`
}

// ─── Response DTOs ────────────────────────────────────────────────────────────

type KnowledgeResponse struct {
	ID        uint      `json:"id"`
	Title     string    `json:"title"`
	Category  string    `json:"category"`
	Content   string    `json:"content"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ListKnowledgeResponse struct {
	Items       []KnowledgeResponse `json:"items"`
	TotalCount  int64               `json:"total_count"`
	CurrentPage int                 `json:"current_page"`
	TotalPages  int                 `json:"total_pages"`
	Limit       int                 `json:"limit"`
}
