package dto

import "time"

// AIAnalysisResponse is the dedicated payload for GET /api/v1/tickets/:id/ai-analysis.
type AIAnalysisResponse struct {
	ProcessingStatus string     `json:"processing_status"`
	Category         string     `json:"category,omitempty"`
	Priority         string     `json:"priority,omitempty"`
	Sentiment        string     `json:"sentiment,omitempty"`
	RecommendedTeam  string     `json:"recommended_team,omitempty"`
	Confidence       *int       `json:"confidence,omitempty"`
	Summary          string     `json:"summary,omitempty"`
	Tags             []string   `json:"tags,omitempty"`
	ProcessedAt      *time.Time `json:"processed_at,omitempty"`
}
