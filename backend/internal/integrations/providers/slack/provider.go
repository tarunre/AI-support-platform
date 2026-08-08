// Package slack implements the Slack notification provider via incoming webhooks.
package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ayush/supportiq/internal/integrations/provider"
)

// priorityEmoji maps ticket priority to a Slack emoji.
var priorityEmoji = map[string]string{
	"URGENT": "🚨", "HIGH": "⚠️", "MEDIUM": "🔵", "LOW": "⚪",
}

// Provider implements provider.Provider for Slack incoming webhooks.
type Provider struct {
	webhookURL string
	channel    string // optional override
}

func (p *Provider) TypeID() string { return "slack" }

func (p *Provider) SupportedEvents() []string {
	return []string{
		provider.EventTicketCreated, provider.EventTicketAssigned,
		provider.EventTicketStatusChanged, provider.EventTicketClosed,
		provider.EventAIAnalysisComplete, provider.EventAIFailed,
		provider.EventReplyApproved, provider.EventEmailFailed,
		provider.EventQueueFailed,
	}
}

func (p *Provider) Configure(cfg map[string]interface{}) error {
	url, _ := cfg["webhook_url"].(string)
	if url == "" {
		return fmt.Errorf("slack: webhook_url is required")
	}
	p.webhookURL = url
	p.channel, _ = cfg["channel"].(string)
	return nil
}

func (p *Provider) TestConnection(ctx context.Context) error {
	payload := map[string]interface{}{
		"text": "✅ SupportIQ connected successfully.",
	}
	return p.post(ctx, payload)
}

func (p *Provider) Notify(ctx context.Context, e provider.Event) error {
	emoji := priorityEmoji[e.Priority]
	if emoji == "" {
		emoji = "🎫"
	}
	color := slackColor(e.Priority, e.Status)

	attachment := map[string]interface{}{
		"color": color,
		"blocks": []interface{}{
			map[string]interface{}{
				"type": "section",
				"text": map[string]string{
					"type": "mrkdwn",
					"text": fmt.Sprintf("%s *<%s|%s>* — %s", emoji, e.URL, e.TicketNumber, e.Subject),
				},
			},
			map[string]interface{}{
				"type": "section",
				"fields": []map[string]string{
					{"type": "mrkdwn", "text": fmt.Sprintf("*Priority:*\n%s", e.Priority)},
					{"type": "mrkdwn", "text": fmt.Sprintf("*Status:*\n%s", e.Status)},
					{"type": "mrkdwn", "text": fmt.Sprintf("*Customer:*\n%s", e.CustomerEmail)},
					{"type": "mrkdwn", "text": fmt.Sprintf("*Agent:*\n%s", ifEmpty(e.AgentName, "Unassigned"))},
				},
			},
		},
	}
	if e.Description != "" {
		attachment["blocks"] = append(attachment["blocks"].([]interface{}), map[string]interface{}{
			"type": "context",
			"elements": []map[string]string{
				{"type": "mrkdwn", "text": truncate(e.Description, 150)},
			},
		})
	}

	payload := map[string]interface{}{
		"attachments": []interface{}{attachment},
	}
	if p.channel != "" {
		payload["channel"] = p.channel
	}
	return p.post(ctx, payload)
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
		return fmt.Errorf("slack: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack: unexpected status %d", resp.StatusCode)
	}
	return nil
}

func slackColor(priority, status string) string {
	if status == "CLOSED" || status == "RESOLVED" {
		return "#36a64f"
	}
	switch priority {
	case "URGENT":
		return "#ff0000"
	case "HIGH":
		return "#ff8c00"
	default:
		return "#0066cc"
	}
}

func ifEmpty(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
