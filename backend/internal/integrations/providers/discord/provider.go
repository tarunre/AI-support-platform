// Package discord implements the Discord notification provider via webhooks.
package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ayush/supportiq/internal/integrations/provider"
)

// Provider implements provider.Provider for Discord webhooks.
type Provider struct {
	webhookURL string
	username   string // bot display name (optional)
}

func (p *Provider) TypeID() string { return "discord" }

func (p *Provider) SupportedEvents() []string {
	return []string{
		provider.EventTicketCreated, provider.EventTicketAssigned,
		provider.EventTicketStatusChanged, provider.EventTicketClosed,
		provider.EventReplyApproved, provider.EventEmailFailed,
	}
}

func (p *Provider) Configure(cfg map[string]interface{}) error {
	url, _ := cfg["webhook_url"].(string)
	if url == "" {
		return fmt.Errorf("discord: webhook_url is required")
	}
	p.webhookURL = url
	p.username, _ = cfg["username"].(string)
	if p.username == "" {
		p.username = "SupportIQ"
	}
	return nil
}

func (p *Provider) TestConnection(ctx context.Context) error {
	return p.Notify(ctx, provider.Event{
		Type:         "test",
		TicketNumber: "TEST",
		Subject:      "SupportIQ connected successfully",
		Priority:     "LOW",
		Status:       "TEST",
	})
}

func (p *Provider) Notify(ctx context.Context, e provider.Event) error {
	embed := map[string]interface{}{
		"title":       fmt.Sprintf("[%s] %s", e.TicketNumber, e.Subject),
		"color":       discordColor(e.Priority, e.Status),
		"description": truncate(e.Description, 200),
		"fields": []map[string]interface{}{
			{"name": "Priority", "value": ifEmpty(e.Priority, "—"), "inline": true},
			{"name": "Status", "value": ifEmpty(e.Status, "—"), "inline": true},
			{"name": "Customer", "value": ifEmpty(e.CustomerEmail, "—"), "inline": true},
			{"name": "Agent", "value": ifEmpty(e.AgentName, "Unassigned"), "inline": true},
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	if e.URL != "" {
		embed["url"] = e.URL
	}

	payload := map[string]interface{}{
		"username": p.username,
		"embeds":   []interface{}{embed},
	}

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
		return fmt.Errorf("discord: %w", err)
	}
	defer resp.Body.Close()
	// Discord returns 204 No Content on success
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("discord: unexpected status %d", resp.StatusCode)
	}
	return nil
}

// discordColor converts priority/status to a decimal RGB colour int.
func discordColor(priority, status string) int {
	if status == "CLOSED" || status == "RESOLVED" {
		return 0x00B050 // green
	}
	switch priority {
	case "URGENT":
		return 0xFF0000
	case "HIGH":
		return 0xFF8C00
	case "MEDIUM":
		return 0x0078D7
	default:
		return 0x9B9B9B
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
