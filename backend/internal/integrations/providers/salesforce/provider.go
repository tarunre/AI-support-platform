// Package salesforce implements the Salesforce CRM provider via REST API.
package salesforce

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ayush/supportiq/internal/integrations/provider"
)

// Provider implements provider.CRMProvider for Salesforce.
// Authentication uses a stored access_token. Refresh is handled externally
// (the admin pastes a fresh token when it expires, or a future OAuth flow
// can populate it automatically).
type Provider struct {
	instanceURL string
	accessToken string
	apiVersion  string
}

func (p *Provider) TypeID() string { return "salesforce" }

func (p *Provider) SupportedEvents() []string {
	return []string{provider.EventTicketCreated, provider.EventTicketClosed}
}

func (p *Provider) Configure(cfg map[string]interface{}) error {
	p.instanceURL, _ = cfg["instance_url"].(string)
	p.accessToken, _ = cfg["access_token"].(string)
	p.apiVersion, _ = cfg["api_version"].(string)
	if p.apiVersion == "" {
		p.apiVersion = "v58.0"
	}
	if p.instanceURL == "" || p.accessToken == "" {
		return fmt.Errorf("salesforce: instance_url and access_token are required")
	}
	p.instanceURL = strings.TrimRight(p.instanceURL, "/")
	return nil
}

func (p *Provider) TestConnection(ctx context.Context) error {
	url := fmt.Sprintf("%s/services/data/%s/limits", p.instanceURL, p.apiVersion)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	p.setAuth(req)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("salesforce: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("salesforce: auth check failed (status %d)", resp.StatusCode)
	}
	return nil
}

func (p *Provider) Notify(ctx context.Context, e provider.Event) error {
	// Create a Salesforce Case for new / closed tickets
	if e.Type != provider.EventTicketCreated && e.Type != provider.EventTicketClosed {
		return nil
	}
	status := "New"
	if e.Type == provider.EventTicketClosed {
		status = "Closed"
	}
	casePayload := map[string]interface{}{
		"Subject":         fmt.Sprintf("[%s] %s", e.TicketNumber, e.Subject),
		"Description":     e.Description,
		"Status":          status,
		"Priority":        sfPriority(e.Priority),
		"Origin":          "SupportIQ",
		"ExternalId__c":   e.TicketNumber,
	}
	_, err := p.createRecord(ctx, "Case", casePayload)
	return err
}

func (p *Provider) LookupCustomer(ctx context.Context, email string) (map[string]interface{}, error) {
	query := fmt.Sprintf("SELECT Id,Name,Email,Phone FROM Contact WHERE Email = '%s' LIMIT 1",
		url.QueryEscape(strings.ReplaceAll(email, "'", "\\'")))
	apiURL := fmt.Sprintf("%s/services/data/%s/query?q=%s", p.instanceURL, p.apiVersion, url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	p.setAuth(req)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("salesforce: %w", err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)

	var result struct {
		TotalSize int                      `json:"totalSize"`
		Records   []map[string]interface{} `json:"records"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("salesforce: decode: %w", err)
	}
	if result.TotalSize == 0 {
		return nil, nil // not found
	}
	return result.Records[0], nil
}

func (p *Provider) CreateCustomer(ctx context.Context, name, email string) (string, error) {
	parts := strings.SplitN(name, " ", 2)
	firstName, lastName := parts[0], ""
	if len(parts) > 1 {
		lastName = parts[1]
	}
	payload := map[string]interface{}{
		"FirstName": firstName,
		"LastName":  lastName,
		"Email":     email,
	}
	return p.createRecord(ctx, "Contact", payload)
}

func (p *Provider) createRecord(ctx context.Context, objectType string, payload map[string]interface{}) (string, error) {
	b, _ := json.Marshal(payload)
	apiURL := fmt.Sprintf("%s/services/data/%s/sobjects/%s", p.instanceURL, p.apiVersion, objectType)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	p.setAuth(req)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("salesforce: %w", err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("salesforce: create %s failed (%d): %s", objectType, resp.StatusCode, truncate(string(data), 200))
	}
	var result struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(data, &result)
	return result.ID, nil
}

func (p *Provider) setAuth(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+p.accessToken)
	req.Header.Set("Accept", "application/json")
}

func sfPriority(p string) string {
	switch p {
	case "URGENT":
		return "High"
	case "HIGH":
		return "High"
	case "MEDIUM":
		return "Medium"
	default:
		return "Low"
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
