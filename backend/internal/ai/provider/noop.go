package provider

import (
	"context"
	"errors"
)

// NoopProvider is used when no AI API key is configured.
// It immediately returns an error so the ticket is marked FAILED instead of
// hanging, allowing the operator to retry once a key is added.
type NoopProvider struct{}

func (n *NoopProvider) Analyze(_ context.Context, _ AnalysisRequest) (*AnalysisResult, error) {
	return nil, errors.New("AI provider not configured: set GEMINI_API_KEY in environment")
}
