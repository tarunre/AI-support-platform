// Package github implements GitHub issue creation via the REST API v3.
package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ayush/supportiq/internal/integrations/provider"
	"github.com/ayush/supportiq/internal/models"
)

// Provider implements provider.IssueProvider for GitHub.
type Provider struct {
	token string
	owner string
	repo  string
}

func (p *Provider) TypeID() string { return "github" }

func (p *Provider) SupportedEvents() []string {
	return []string{provider.EventTicketCreated, provider.EventTicketStatusChanged}
}

func (p *Provider) Configure(cfg map[string]interface{}) error {
	p.token, _ = cfg["token"].(string)
	p.owner, _ = cfg["owner"].(string)
	p.repo, _ = cfg["repo"].(string)
	if p.token == "" || p.owner == "" || p.repo == "" {
		return fmt.Errorf("github: token, owner, and repo are required")
	}
	return nil
}

func (p *Provider) TestConnection(ctx context.Context) error {
	url := "https://api.github.com/user"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	p.setAuth(req)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("github: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("github: auth failed (status %d)", resp.StatusCode)
	}
	return nil
}

func (p *Provider) Notify(_ context.Context, _ provider.Event) error { return nil }

func (p *Provider) CreateIssue(ctx context.Context, ticket *models.Ticket) (*provider.IssueRef, error) {
	body := fmt.Sprintf(
		"**Customer:** %s (%s)\n\n**Description:**\n%s\n\n---\n_Created from SupportIQ ticket %s_",
		ticket.CustomerName, ticket.CustomerEmail, ticket.Description, ticket.TicketNumber,
	)
	labels := []string{"supportiq", strings.ToLower(string(ticket.Priority))}

	payload := map[string]interface{}{
		"title":  fmt.Sprintf("[%s] %s", ticket.TicketNumber, ticket.Subject),
		"body":   body,
		"labels": labels,
	}

	b, _ := json.Marshal(payload)
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues", p.owner, p.repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	p.setAuth(req)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github: %w", err)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("github: create issue failed (%d): %s", resp.StatusCode, truncate(string(data), 200))
	}

	var result struct {
		ID      int64  `json:"id"`
		Number  int    `json:"number"`
		HTMLURL string `json:"html_url"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("github: decode response: %w", err)
	}

	return &provider.IssueRef{
		ExternalID:  fmt.Sprintf("%d", result.ID),
		ExternalKey: fmt.Sprintf("#%d", result.Number),
		ExternalURL: result.HTMLURL,
	}, nil
}

func (p *Provider) setAuth(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+p.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
