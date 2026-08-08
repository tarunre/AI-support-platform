// Package jira implements Jira Cloud issue creation via REST API v3.
package jira

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ayush/supportiq/internal/integrations/provider"
	"github.com/ayush/supportiq/internal/models"
)

// Provider implements provider.IssueProvider for Jira Cloud.
type Provider struct {
	baseURL    string // e.g. https://myorg.atlassian.net
	email      string
	apiToken   string
	projectKey string
	issueType  string // default "Bug"
}

func (p *Provider) TypeID() string { return "jira" }

func (p *Provider) SupportedEvents() []string {
	return []string{provider.EventTicketCreated, provider.EventTicketStatusChanged}
}

func (p *Provider) Configure(cfg map[string]interface{}) error {
	p.baseURL, _ = cfg["base_url"].(string)
	p.email, _ = cfg["email"].(string)
	p.apiToken, _ = cfg["api_token"].(string)
	p.projectKey, _ = cfg["project_key"].(string)
	p.issueType, _ = cfg["issue_type"].(string)
	if p.issueType == "" {
		p.issueType = "Task"
	}
	if p.baseURL == "" || p.email == "" || p.apiToken == "" || p.projectKey == "" {
		return fmt.Errorf("jira: base_url, email, api_token and project_key are required")
	}
	p.baseURL = strings.TrimRight(p.baseURL, "/")
	return nil
}

func (p *Provider) TestConnection(ctx context.Context) error {
	url := fmt.Sprintf("%s/rest/api/3/myself", p.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	p.setAuth(req)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("jira: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("jira: auth failed (status %d)", resp.StatusCode)
	}
	return nil
}

func (p *Provider) Notify(_ context.Context, _ provider.Event) error { return nil }

func (p *Provider) CreateIssue(ctx context.Context, ticket *models.Ticket) (*provider.IssueRef, error) {
	description := fmt.Sprintf(
		"*Customer:* %s (%s)\n\n*Description:*\n%s\n\n_Created from SupportIQ ticket %s_",
		ticket.CustomerName, ticket.CustomerEmail, ticket.Description, ticket.TicketNumber,
	)

	body := map[string]interface{}{
		"fields": map[string]interface{}{
			"project":   map[string]string{"key": p.projectKey},
			"issuetype": map[string]string{"name": p.issueType},
			"summary":   fmt.Sprintf("[%s] %s", ticket.TicketNumber, ticket.Subject),
			"description": map[string]interface{}{
				"type":    "doc",
				"version": 1,
				"content": []interface{}{
					map[string]interface{}{
						"type": "paragraph",
						"content": []interface{}{
							map[string]interface{}{"type": "text", "text": description},
						},
					},
				},
			},
			"priority": map[string]string{"name": jiraPriority(string(ticket.Priority))},
			"labels":   []string{"supportiq", strings.ToLower(string(ticket.Priority))},
		},
	}

	b, _ := json.Marshal(body)
	url := fmt.Sprintf("%s/rest/api/3/issue", p.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	p.setAuth(req)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("jira: %w", err)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("jira: create issue failed (%d): %s", resp.StatusCode, truncate(string(data), 200))
	}

	var result struct {
		ID  string `json:"id"`
		Key string `json:"key"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("jira: decode response: %w", err)
	}

	return &provider.IssueRef{
		ExternalID:  result.ID,
		ExternalKey: result.Key,
		ExternalURL: fmt.Sprintf("%s/browse/%s", p.baseURL, result.Key),
	}, nil
}

func (p *Provider) setAuth(req *http.Request) {
	creds := base64.StdEncoding.EncodeToString([]byte(p.email + ":" + p.apiToken))
	req.Header.Set("Authorization", "Basic "+creds)
	req.Header.Set("Accept", "application/json")
}

func jiraPriority(p string) string {
	switch p {
	case "URGENT":
		return "Highest"
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
