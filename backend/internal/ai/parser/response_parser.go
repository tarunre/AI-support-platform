package parser

import (
	"encoding/json"
	"fmt"
	"strings"
)

// RawAIResponse mirrors the JSON structure the model is asked to produce.
type RawAIResponse struct {
	Category        string   `json:"category"`
	Priority        string   `json:"priority"`
	Sentiment       string   `json:"sentiment"`
	RecommendedTeam string   `json:"recommended_team"`
	Confidence      int      `json:"confidence"`
	Summary         string   `json:"summary"`
	Tags            []string `json:"tags"`
}

// Parse cleans the raw model output and unmarshals it into RawAIResponse.
// It defensively strips markdown code fences in case the model disobeys the
// prompt despite being instructed not to include them.
func Parse(raw string) (*RawAIResponse, error) {
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
		return nil, fmt.Errorf("no JSON object found in AI response")
	}
	cleaned = cleaned[start : end+1]

	var resp RawAIResponse
	if err := json.Unmarshal([]byte(cleaned), &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal AI response: %w", err)
	}
	return &resp, nil
}
