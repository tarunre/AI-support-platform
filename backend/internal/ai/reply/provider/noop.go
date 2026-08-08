package replyprovider

import (
	"context"
	"fmt"
)

// NoopReplyProvider is used when no API key is configured.
// It returns a clear error so the service layer can surface it gracefully.
type NoopReplyProvider struct{}

func (n *NoopReplyProvider) GenerateReply(_ context.Context, _ ReplyRequest) (*ReplyResult, error) {
	return nil, fmt.Errorf("AI reply provider not configured: set GEMINI_API_KEY to enable reply generation")
}
