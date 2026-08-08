package replyparser

import (
	"encoding/json"
	"fmt"
	"strings"
)

// RawReplyResponse mirrors the JSON structure the model is asked to produce.
type RawReplyResponse struct {
	Reply      string `json:"reply"`
	Confidence int    `json:"confidence"`
}

// Parse cleans the raw model output and unmarshals it into RawReplyResponse.
// It defensively strips markdown code fences in case the model disobeys the prompt.
func Parse(raw string) (*RawReplyResponse, error) {
	cleaned := strings.TrimSpace(raw)

	// Strip ```json ... ``` or ``` ... ``` wrappers
	if strings.HasPrefix(cleaned, "```") {
		cleaned = strings.TrimPrefix(cleaned, "```json")
		cleaned = strings.TrimPrefix(cleaned, "```")
		if idx := strings.LastIndex(cleaned, "```"); idx != -1 {
			cleaned = cleaned[:idx]
		}
		cleaned = strings.TrimSpace(cleaned)
	}

	// Locate the outermost JSON object in case there is leading/trailing text
	start := strings.Index(cleaned, "{")
	end := strings.LastIndex(cleaned, "}")
	if start == -1 || end == -1 || end < start {
		return nil, fmt.Errorf("no JSON object found in AI reply response")
	}
	cleaned = cleaned[start : end+1]

	var resp RawReplyResponse
	if err := json.Unmarshal([]byte(cleaned), &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal AI reply response: %w", err)
	}
	return &resp, nil
}
