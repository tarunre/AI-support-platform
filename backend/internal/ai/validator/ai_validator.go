package validator

import (
	"fmt"

	"github.com/ayush/supportiq/internal/ai/parser"
)

var (
	validCategories = map[string]bool{
		"Payment": true, "Authentication": true, "Technical Issue": true,
		"Refund": true, "Account": true, "Subscription": true, "General": true,
	}
	validPriorities = map[string]bool{
		"Low": true, "Medium": true, "High": true, "Urgent": true,
	}
	validSentiments = map[string]bool{
		"Positive": true, "Neutral": true, "Frustrated": true,
		"Angry": true, "Confused": true,
	}
	validTeams = map[string]bool{
		"Finance": true, "Support": true, "Engineering": true,
		"Sales": true, "Security": true,
	}
)

// Validate checks every field of the parsed AI response.
// Returns the first validation error encountered, or nil if valid.
func Validate(resp *parser.RawAIResponse) error {
	if resp.Category == "" {
		return fmt.Errorf("missing field: category")
	}
	if !validCategories[resp.Category] {
		return fmt.Errorf("invalid category %q: must be one of Payment, Authentication, Technical Issue, Refund, Account, Subscription, General", resp.Category)
	}
	if resp.Priority == "" {
		return fmt.Errorf("missing field: priority")
	}
	if !validPriorities[resp.Priority] {
		return fmt.Errorf("invalid priority %q: must be one of Low, Medium, High, Urgent", resp.Priority)
	}
	if resp.Sentiment == "" {
		return fmt.Errorf("missing field: sentiment")
	}
	if !validSentiments[resp.Sentiment] {
		return fmt.Errorf("invalid sentiment %q: must be one of Positive, Neutral, Frustrated, Angry, Confused", resp.Sentiment)
	}
	if resp.RecommendedTeam == "" {
		return fmt.Errorf("missing field: recommended_team")
	}
	if !validTeams[resp.RecommendedTeam] {
		return fmt.Errorf("invalid recommended_team %q: must be one of Finance, Support, Engineering, Sales, Security", resp.RecommendedTeam)
	}
	if resp.Confidence < 0 || resp.Confidence > 100 {
		return fmt.Errorf("confidence %d is outside the valid range 0–100", resp.Confidence)
	}
	if resp.Summary == "" {
		return fmt.Errorf("missing field: summary")
	}
	if len(resp.Tags) == 0 {
		return fmt.Errorf("missing field: tags (must have at least one tag)")
	}
	return nil
}
