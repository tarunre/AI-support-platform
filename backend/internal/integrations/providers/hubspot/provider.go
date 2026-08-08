// Package hubspot implements the HubSpot CRM provider via REST API v3.
package hubspot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ayush/supportiq/internal/integrations/provider"
)

const hubspotBaseURL = "https://api.hubapi.com"

// Provider implements provider.CRMProvider for HubSpot.
type Provider struct {
	accessToken string
}

func (p *Provider) TypeID() string { return "hubspot" }

func (p *Provider) SupportedEvents() []string {
	return []string{provider.EventTicketCreated, provider.EventTicketClosed}
}

func (p *Provider) Configure(cfg map[string]interface{}) error {
	p.accessToken, _ = cfg["access_token"].(string)
	if p.accessToken == "" {
		// Legacy API key support
		if key, _ := cfg["api_key"].(string); key != "" {
			p.accessToken = key
		}
	}
	if p.accessToken == "" {
		return fmt.Errorf("hubspot: access_token (or api_key) is required")
	}
	return nil
}

func (p *Provider) TestConnection(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		hubspotBaseURL+"/crm/v3/objects/contacts?limit=1", nil)
	if err != nil {
		return err
	}
	p.setAuth(req)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("hubspot: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("hubspot: auth failed (status %d)", resp.StatusCode)
	}
	return nil
}

func (p *Provider) Notify(ctx context.Context, e provider.Event) error {
	if e.Type != provider.EventTicketCreated {
		return nil
	}
	// Create a HubSpot Ticket object
	payload := map[string]interface{}{
		"properties": map[string]interface{}{
			"subject":       fmt.Sprintf("[%s] %s", e.TicketNumber, e.Subject),
			"content":       e.Description,
			"hs_pipeline":   "0",
			"hs_pipeline_stage": "1",
			"hs_ticket_priority": hsTicketPriority(e.Priority),
		},
	}
	_, err := p.post(ctx, "/crm/v3/objects/tickets", payload)
	return err
}

func (p *Provider) LookupCustomer(ctx context.Context, email string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"filterGroups": []interface{}{
			map[string]interface{}{
				"filters": []interface{}{
					map[string]interface{}{
						"propertyName": "email",
						"operator":     "EQ",
						"value":        email,
					},
				},
			},
		},
		"properties": []string{"firstname", "lastname", "email", "phone"},
		"limit":      1,
	}

	data, err := p.post(ctx, "/crm/v3/objects/contacts/search", payload)
	if err != nil {
		return nil, err
	}
	var result struct {
		Total   int                        `json:"total"`
		Results []map[string]interface{} `json:"results"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	if result.Total == 0 {
		return nil, nil
	}
	return result.Results[0], nil
}

func (p *Provider) CreateCustomer(ctx context.Context, name, email string) (string, error) {
	payload := map[string]interface{}{
		"properties": map[string]interface{}{
			"email":     email,
			"firstname": name,
		},
	}
	data, err := p.post(ctx, "/crm/v3/objects/contacts", payload)
	if err != nil {
		return "", err
	}
	var result struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return "", err
	}
	return result.ID, nil
}

func (p *Provider) post(ctx context.Context, path string, payload interface{}) ([]byte, error) {
	b, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, hubspotBaseURL+path, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	p.setAuth(req)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("hubspot: %w", err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("hubspot: %s returned %d: %s", path, resp.StatusCode, truncate(string(data), 200))
	}
	return data, nil
}

func (p *Provider) setAuth(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+p.accessToken)
}

func hsTicketPriority(p string) string {
	switch p {
	case "URGENT", "HIGH":
		return "HIGH"
	case "MEDIUM":
		return "MEDIUM"
	default:
		return "LOW"
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
