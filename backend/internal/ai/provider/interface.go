package provider

import "context"

// AnalysisRequest contains the data forwarded to the AI provider.
type AnalysisRequest struct {
	Subject      string
	Description  string
	CustomerName string
	Category     string
	Priority     string
}

// AnalysisResult is the validated, normalised AI response.
type AnalysisResult struct {
	Category        string
	Priority        string
	Sentiment       string
	RecommendedTeam string
	Confidence      int
	Summary         string
	Tags            []string
}

// Provider is the interface every AI backend must satisfy.
// Swapping Gemini for OpenAI or Claude requires only a new implementation of
// this interface — no service or handler changes.
type Provider interface {
	Analyze(ctx context.Context, req AnalysisRequest) (*AnalysisResult, error)
}
