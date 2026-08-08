// Package linear implements Linear issue creation via the GraphQL API.
package linear

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

const linearAPIURL = "https://api.linear.app/graphql"

// Provider implements provider.IssueProvider for Linear.
type Provider struct {
	apiKey string
	teamID string
}

func (p *Provider) TypeID() string { return "linear" }

func (p *Provider) SupportedEvents() []string {
	return []string{provider.EventTicketCreated, provider.EventTicketStatusChanged}
}

func (p *Provider) Configure(cfg map[string]interface{}) error {
	p.apiKey, _ = cfg["api_key"].(string)
	p.teamID, _ = cfg["team_id"].(string)
	if p.apiKey == "" || p.teamID == "" {
		return fmt.Errorf("linear: api_key and team_id are required")
	}
	return nil
}

func (p *Provider) TestConnection(ctx context.Context) error {
	query := `query { viewer { id name email } }`
	var res struct {
		Data struct {
			Viewer struct {
				ID string `json:"id"`
			} `json:"viewer"`
		} `json:"data"`
		Errors []struct{ Message string } `json:"errors"`
	}
	if err := p.gql(ctx, query, nil, &res); err != nil {
		return err
	}
	if len(res.Errors) > 0 {
		return fmt.Errorf("linear: %s", res.Errors[0].Message)
	}
	if res.Data.Viewer.ID == "" {
		return fmt.Errorf("linear: authentication failed")
	}
	return nil
}

func (p *Provider) Notify(_ context.Context, _ provider.Event) error { return nil }

func (p *Provider) CreateIssue(ctx context.Context, ticket *models.Ticket) (*provider.IssueRef, error) {
	desc := fmt.Sprintf(
		"**Customer:** %s (%s)\n\n**Description:**\n%s\n\n_Created from SupportIQ ticket %s_",
		ticket.CustomerName, ticket.CustomerEmail, ticket.Description, ticket.TicketNumber,
	)

	mutation := `
mutation CreateIssue($input: IssueCreateInput!) {
  issueCreate(input: $input) {
    success
    issue {
      id
      identifier
      url
    }
  }
}`
	vars := map[string]interface{}{
		"input": map[string]interface{}{
			"teamId":      p.teamID,
			"title":       fmt.Sprintf("[%s] %s", ticket.TicketNumber, ticket.Subject),
			"description": desc,
			"priority":    linearPriority(string(ticket.Priority)),
		},
	}

	var res struct {
		Data struct {
			IssueCreate struct {
				Success bool `json:"success"`
				Issue   struct {
					ID         string `json:"id"`
					Identifier string `json:"identifier"`
					URL        string `json:"url"`
				} `json:"issue"`
			} `json:"issueCreate"`
		} `json:"data"`
		Errors []struct{ Message string } `json:"errors"`
	}

	if err := p.gql(ctx, mutation, vars, &res); err != nil {
		return nil, err
	}
	if len(res.Errors) > 0 {
		return nil, fmt.Errorf("linear: %s", res.Errors[0].Message)
	}
	if !res.Data.IssueCreate.Success {
		return nil, fmt.Errorf("linear: issue creation failed")
	}
	issue := res.Data.IssueCreate.Issue
	return &provider.IssueRef{
		ExternalID:  issue.ID,
		ExternalKey: issue.Identifier,
		ExternalURL: issue.URL,
	}, nil
}

func (p *Provider) gql(ctx context.Context, query string, variables map[string]interface{}, out interface{}) error {
	body := map[string]interface{}{"query": strings.TrimSpace(query)}
	if variables != nil {
		body["variables"] = variables
	}
	b, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, linearAPIURL, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("linear: %w", err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("linear: status %d: %s", resp.StatusCode, truncate(string(data), 200))
	}
	return json.Unmarshal(data, out)
}

// linearPriority maps SupportIQ priority to Linear priority int (0–4).
func linearPriority(p string) int {
	switch p {
	case "URGENT":
		return 1 // Urgent
	case "HIGH":
		return 2 // High
	case "MEDIUM":
		return 3 // Medium
	default:
		return 4 // Low
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
