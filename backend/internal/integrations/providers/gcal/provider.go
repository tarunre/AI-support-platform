// Package gcal implements the Google Calendar provider for scheduling support meetings.
package gcal

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

// Provider implements provider.CalendarProvider for Google Calendar.
type Provider struct {
	accessToken string
	calendarID  string // defaults to "primary"
}

func (p *Provider) TypeID() string { return "gcal" }

func (p *Provider) SupportedEvents() []string {
	// gcal does not auto-create meetings from ticket events; meetings are
	// created explicitly via CreateMeeting.  Notify is a no-op.
	return []string{}
}

func (p *Provider) Configure(cfg map[string]interface{}) error {
	p.accessToken, _ = cfg["access_token"].(string)
	p.calendarID, _ = cfg["calendar_id"].(string)
	if p.calendarID == "" {
		p.calendarID = "primary"
	}
	if p.accessToken == "" {
		return fmt.Errorf("gcal: access_token is required")
	}
	return nil
}

func (p *Provider) TestConnection(ctx context.Context) error {
	url := "https://www.googleapis.com/calendar/v3/users/me/calendarList?maxResults=1"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	p.setAuth(req)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("gcal: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("gcal: auth failed (status %d)", resp.StatusCode)
	}
	return nil
}

func (p *Provider) Notify(_ context.Context, _ provider.Event) error { return nil }

func (p *Provider) CreateMeeting(ctx context.Context, mr provider.MeetingRequest) (*provider.MeetingResult, error) {
	attendees := make([]map[string]string, 0, len(mr.Attendees))
	for _, a := range mr.Attendees {
		attendees = append(attendees, map[string]string{"email": a})
	}

	requestID := fmt.Sprintf("supportiq-%d", time.Now().UnixNano())
	body := map[string]interface{}{
		"summary":     mr.Title,
		"description": mr.Description,
		"start": map[string]string{
			"dateTime": mr.Start.UTC().Format(time.RFC3339),
			"timeZone": "UTC",
		},
		"end": map[string]string{
			"dateTime": mr.End.UTC().Format(time.RFC3339),
			"timeZone": "UTC",
		},
		"attendees":  attendees,
		"conferenceData": map[string]interface{}{
			"createRequest": map[string]interface{}{
				"requestId": requestID,
				"conferenceSolutionKey": map[string]string{
					"type": "hangoutsMeet",
				},
			},
		},
	}

	b, _ := json.Marshal(body)
	apiURL := fmt.Sprintf(
		"https://www.googleapis.com/calendar/v3/calendars/%s/events?conferenceDataVersion=1&sendUpdates=all",
		p.calendarID,
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	p.setAuth(req)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gcal: %w", err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gcal: create event failed (%d): %s", resp.StatusCode, truncate(string(data), 200))
	}

	var result struct {
		ID             string `json:"id"`
		HTMLURL        string `json:"htmlLink"`
		ConferenceData struct {
			EntryPoints []struct {
				EntryPointType string `json:"entryPointType"`
				URI            string `json:"uri"`
			} `json:"entryPoints"`
		} `json:"conferenceData"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("gcal: decode: %w", err)
	}

	meetLink := ""
	for _, ep := range result.ConferenceData.EntryPoints {
		if ep.EntryPointType == "video" {
			meetLink = ep.URI
			break
		}
	}

	return &provider.MeetingResult{
		EventID:  result.ID,
		MeetLink: meetLink,
		HTMLURL:  result.HTMLURL,
	}, nil
}

func (p *Provider) setAuth(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+p.accessToken)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
