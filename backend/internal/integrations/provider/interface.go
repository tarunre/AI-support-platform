// Package provider defines the interfaces and shared types for all integration
// providers. No provider-specific code lives here — only contracts.
package provider

import (
	"context"
	"time"

	"github.com/ayush/supportiq/internal/models"
)

// ─── Event types ─────────────────────────────────────────────────────────────

const (
	EventTicketCreated       = "ticket.created"
	EventTicketUpdated       = "ticket.updated"
	EventTicketAssigned      = "ticket.assigned"
	EventTicketStatusChanged = "ticket.status_changed"
	EventTicketClosed        = "ticket.closed"
	EventAIAnalysisComplete  = "ai.analysis_complete"
	EventAIFailed            = "ai.failed"
	EventReplyApproved       = "reply.approved"
	EventReplySent           = "reply.sent"
	EventEmailFailed         = "email.failed"
	EventQueueFailed         = "queue.failed"
)

// ─── Shared types ─────────────────────────────────────────────────────────────

// Event is the canonical notification payload sent to all providers.
type Event struct {
	Type          string
	TicketID      string
	TicketNumber  string
	Subject       string
	Priority      string
	Status        string
	AgentName     string
	CustomerEmail string
	URL           string
	Description   string
	Extra         map[string]interface{}
}

// IssueRef is returned by IssueProvider.CreateIssue.
type IssueRef struct {
	ExternalID  string
	ExternalKey string
	ExternalURL string
}

// MeetingRequest is passed to CalendarProvider.CreateMeeting.
type MeetingRequest struct {
	Title       string
	Description string
	Start       time.Time
	End         time.Time
	Attendees   []string
}

// MeetingResult is returned by CalendarProvider.CreateMeeting.
type MeetingResult struct {
	EventID  string
	MeetLink string
	HTMLURL  string
}

// ─── Interfaces ──────────────────────────────────────────────────────────────

// Provider is the base interface every integration provider must implement.
type Provider interface {
	// TypeID returns the canonical provider identifier (e.g. "slack").
	TypeID() string
	// Configure validates and stores the configuration map for this instance.
	Configure(cfg map[string]interface{}) error
	// TestConnection makes a lightweight API call to verify credentials.
	TestConnection(ctx context.Context) error
	// Notify sends a notification to the external service.
	Notify(ctx context.Context, event Event) error
	// SupportedEvents returns the event types this provider handles.
	SupportedEvents() []string
}

// IssueProvider extends Provider with issue-tracking capabilities.
// Implemented by Jira, Linear, and GitHub providers.
type IssueProvider interface {
	Provider
	CreateIssue(ctx context.Context, ticket *models.Ticket) (*IssueRef, error)
}

// CRMProvider extends Provider with customer relationship management.
// Implemented by Salesforce and HubSpot providers.
type CRMProvider interface {
	Provider
	LookupCustomer(ctx context.Context, email string) (map[string]interface{}, error)
	CreateCustomer(ctx context.Context, name, email string) (string, error)
}

// CalendarProvider extends Provider with meeting scheduling.
// Implemented by the Google Calendar provider.
type CalendarProvider interface {
	Provider
	CreateMeeting(ctx context.Context, req MeetingRequest) (*MeetingResult, error)
}
