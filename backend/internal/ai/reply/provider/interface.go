package replyprovider

import "context"

// RelevantDocument is a single KB document included in the prompt context.
type RelevantDocument struct {
	Title    string
	Category string
	Content  string
}

// ReplyRequest is the input to the AI reply provider.
type ReplyRequest struct {
	Subject      string
	Description  string
	Category     string
	Priority     string
	Sentiment    string
	Documents    []RelevantDocument
	CustomPrompt string // if non-empty, overrides the standard prompt builder
	OrderContext string // optional order status snippet from orders.json
}

// ReplyResult is the parsed, validated AI reply output.
type ReplyResult struct {
	Reply      string
	Confidence int
}

// ReplyProvider is the interface every AI reply backend must satisfy.
// Swapping Gemini for another model requires only a new implementation of
// this interface — no service or handler changes.
type ReplyProvider interface {
	GenerateReply(ctx context.Context, req ReplyRequest) (*ReplyResult, error)
}
