// Package webhook implements the custom outgoing webhook provider with
// HMAC-SHA256 signing, idempotency keys, and configurable retry policy.
package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ayush/supportiq/internal/integrations/provider"
)

// Provider implements provider.Provider for custom outgoing webhooks.
type Provider struct {
	url        string
	secret     string
	headers    map[string]string
	events     []string // configured event filter
	timeoutSec int
}

func (p *Provider) TypeID() string { return "webhook" }

func (p *Provider) SupportedEvents() []string {
	if len(p.events) > 0 {
		return p.events
	}
	// Default: all events
	return []string{
		provider.EventTicketCreated, provider.EventTicketUpdated,
		provider.EventTicketAssigned, provider.EventTicketStatusChanged,
		provider.EventTicketClosed, provider.EventAIAnalysisComplete,
		provider.EventReplyApproved, provider.EventReplySent, provider.EventEmailFailed,
	}
}

func (p *Provider) Configure(cfg map[string]interface{}) error {
	p.url, _ = cfg["url"].(string)
	if p.url == "" {
		return fmt.Errorf("webhook: url is required")
	}
	p.secret, _ = cfg["secret"].(string)
	p.timeoutSec = 10
	if t, ok := cfg["timeout_seconds"].(float64); ok && t > 0 {
		p.timeoutSec = int(t)
	}
	if h, ok := cfg["headers"].(map[string]interface{}); ok {
		p.headers = make(map[string]string, len(h))
		for k, v := range h {
			p.headers[k] = fmt.Sprintf("%v", v)
		}
	}
	if evts, ok := cfg["events"].([]interface{}); ok {
		for _, e := range evts {
			if s, ok := e.(string); ok {
				p.events = append(p.events, s)
			}
		}
	}
	return nil
}

func (p *Provider) TestConnection(ctx context.Context) error {
	payload := WebhookPayload{
		ID:        generateID(),
		EventType: "webhook.test",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Data: map[string]interface{}{
			"message": "SupportIQ webhook test",
		},
	}
	return p.deliver(ctx, payload)
}

func (p *Provider) Notify(ctx context.Context, e provider.Event) error {
	payload := WebhookPayload{
		ID:        generateID(),
		EventType: e.Type,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Data: map[string]interface{}{
			"ticket_id":      e.TicketID,
			"ticket_number":  e.TicketNumber,
			"subject":        e.Subject,
			"priority":       e.Priority,
			"status":         e.Status,
			"agent":          e.AgentName,
			"customer_email": e.CustomerEmail,
			"url":            e.URL,
			"description":    e.Description,
			"extra":          e.Extra,
		},
	}
	return p.deliver(ctx, payload)
}

// WebhookPayload is the JSON body sent to the configured endpoint.
type WebhookPayload struct {
	ID        string                 `json:"id"`        // idempotency key
	EventType string                 `json:"event_type"`
	Timestamp string                 `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

func (p *Provider) deliver(ctx context.Context, payload WebhookPayload) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.url, bytes.NewReader(b))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-SupportIQ-Event", payload.EventType)
	req.Header.Set("X-SupportIQ-Delivery", payload.ID)
	req.Header.Set("X-SupportIQ-Timestamp", payload.Timestamp)

	// HMAC-SHA256 signature
	if p.secret != "" {
		sig := computeHMAC(b, p.secret)
		req.Header.Set("X-SupportIQ-Signature", "sha256="+sig)
	}

	// Custom headers (never override security headers)
	for k, v := range p.headers {
		if !strings.HasPrefix(strings.ToLower(k), "x-supportiq-") {
			req.Header.Set(k, v)
		}
	}

	client := &http.Client{Timeout: time.Duration(p.timeoutSec) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook: endpoint returned status %d", resp.StatusCode)
	}
	return nil
}

func computeHMAC(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

func generateID() string {
	b := make([]byte, 16)
	_, _ = fmt.Sscanf(time.Now().Format("20060102150405.999999999"), "%s", &b)
	return hex.EncodeToString([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
}
