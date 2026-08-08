// Package teams implements the Microsoft Teams notification provider via
// incoming webhooks with Adaptive Cards.
package teams

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ayush/supportiq/internal/integrations/provider"
)

// Provider implements provider.Provider for Microsoft Teams incoming webhooks.
type Provider struct {
	webhookURL string
}

func (p *Provider) TypeID() string { return "teams" }

func (p *Provider) SupportedEvents() []string {
	return []string{
		provider.EventTicketCreated, provider.EventTicketAssigned,
		provider.EventTicketStatusChanged, provider.EventTicketClosed,
		provider.EventAIAnalysisComplete, provider.EventReplyApproved,
		provider.EventEmailFailed, provider.EventQueueFailed,
	}
}

func (p *Provider) Configure(cfg map[string]interface{}) error {
	url, _ := cfg["webhook_url"].(string)
	if url == "" {
		return fmt.Errorf("teams: webhook_url is required")
	}
	p.webhookURL = url
	return nil
}

func (p *Provider) TestConnection(ctx context.Context) error {
	card := p.buildCard(provider.Event{
		Type:         "test",
		TicketNumber: "TEST",
		Subject:      "SupportIQ connected successfully",
		Priority:     "LOW",
		Status:       "ACTIVE",
	})
	return p.post(ctx, card)
}

func (p *Provider) Notify(ctx context.Context, e provider.Event) error {
	return p.post(ctx, p.buildCard(e))
}

// buildCard constructs a MessageCard (O365 Connector Card) payload.
func (p *Provider) buildCard(e provider.Event) map[string]interface{} {
	themeColor := teamsColor(e.Priority, e.Status)
	facts := []map[string]string{
		{"name": "Priority", "value": e.Priority},
		{"name": "Status", "value": e.Status},
		{"name": "Customer", "value": ifEmpty(e.CustomerEmail, "—")},
		{"name": "Agent", "value": ifEmpty(e.AgentName, "Unassigned")},
		{"name": "Event", "value": e.Type},
	}
	card := map[string]interface{}{
		"@type":      "MessageCard",
		"@context":   "http://schema.org/extensions",
		"themeColor": themeColor,
		"summary":    fmt.Sprintf("[%s] %s", e.TicketNumber, e.Subject),
		"sections": []interface{}{
			map[string]interface{}{
				"activityTitle":    fmt.Sprintf("**[%s]** %s", e.TicketNumber, e.Subject),
				"activitySubtitle": e.Description,
				"facts":            facts,
				"markdown":         true,
			},
		},
	}
	if e.URL != "" {
		card["potentialAction"] = []interface{}{
			map[string]interface{}{
				"@type": "OpenUri",
				"name":  "View Ticket",
				"targets": []map[string]string{
					{"os": "default", "uri": e.URL},
				},
			},
		}
	}
	return card
}

func (p *Provider) post(ctx context.Context, payload interface{}) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.webhookURL, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("teams: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("teams: unexpected status %d", resp.StatusCode)
	}
	return nil
}

func teamsColor(priority, status string) string {
	if status == "CLOSED" || status == "RESOLVED" {
		return "00B050"
	}
	switch priority {
	case "URGENT":
		return "FF0000"
	case "HIGH":
		return "FF8C00"
	default:
		return "0078D7"
	}
}

func ifEmpty(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}
