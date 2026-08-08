package dto

import (
	"time"

	"github.com/ayush/supportiq/internal/models"
)

// ── Email Account DTOs ────────────────────────────────────────────────────────

type CreateEmailAccountRequest struct {
	Provider        string `json:"provider"          binding:"required"`
	EmailAddress    string `json:"email_address"     binding:"required,email"`
	DisplayName     string `json:"display_name"`
	IMAPHost        string `json:"imap_host"`
	IMAPPort        int    `json:"imap_port"`
	IMAPUseTLS      bool   `json:"imap_use_tls"`
	SMTPHost        string `json:"smtp_host"`
	SMTPPort        int    `json:"smtp_port"`
	SMTPImplicitTLS bool   `json:"smtp_implicit_tls"`
	Username        string `json:"username"          binding:"required"`
	Password        string `json:"password"          binding:"required"`
	IsActive        bool   `json:"is_active"`
}

type UpdateEmailAccountRequest struct {
	DisplayName     *string `json:"display_name"`
	IMAPHost        *string `json:"imap_host"`
	IMAPPort        *int    `json:"imap_port"`
	IMAPUseTLS      *bool   `json:"imap_use_tls"`
	SMTPHost        *string `json:"smtp_host"`
	SMTPPort        *int    `json:"smtp_port"`
	SMTPImplicitTLS *bool   `json:"smtp_implicit_tls"`
	Username        *string `json:"username"`
	Password        *string `json:"password"` // if nil, existing password retained
	IsActive        *bool   `json:"is_active"`
}

// EmailAccountResponse is what the API returns — never includes the password.
type EmailAccountResponse struct {
	ID              uint       `json:"id"`
	Provider        string     `json:"provider"`
	EmailAddress    string     `json:"email_address"`
	DisplayName     string     `json:"display_name"`
	IMAPHost        string     `json:"imap_host"`
	IMAPPort        int        `json:"imap_port"`
	IMAPUseTLS      bool       `json:"imap_use_tls"`
	SMTPHost        string     `json:"smtp_host"`
	SMTPPort        int        `json:"smtp_port"`
	SMTPImplicitTLS bool       `json:"smtp_implicit_tls"`
	Username        string     `json:"username"`
	IsActive        bool       `json:"is_active"`
	LastSyncAt      *time.Time `json:"last_sync_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// ToEmailAccountResponse converts a model to a safe response DTO.
func ToEmailAccountResponse(a *models.EmailAccount) EmailAccountResponse {
	return EmailAccountResponse{
		ID:              a.ID,
		Provider:        string(a.Provider),
		EmailAddress:    a.EmailAddress,
		DisplayName:     a.DisplayName,
		IMAPHost:        a.IMAPHost,
		IMAPPort:        a.IMAPPort,
		IMAPUseTLS:      a.IMAPUseTLS,
		SMTPHost:        a.SMTPHost,
		SMTPPort:        a.SMTPPort,
		SMTPImplicitTLS: a.SMTPImplicitTLS,
		Username:        a.Username,
		IsActive:        a.IsActive,
		LastSyncAt:      a.LastSyncAt,
		CreatedAt:       a.CreatedAt,
		UpdatedAt:       a.UpdatedAt,
	}
}

// ── Email Message DTOs ────────────────────────────────────────────────────────

type SendEmailRequest struct {
	To      string `json:"to"      binding:"required,email"`
	Subject string `json:"subject" binding:"required"`
	Body    string `json:"body"    binding:"required"`
}

// EmailMessageResponse is the API representation of a single email message.
type EmailMessageResponse struct {
	ID               uint                     `json:"id"`
	TicketID         string                   `json:"ticket_id"`
	Direction        string                   `json:"direction"`
	Sender           string                   `json:"sender"`
	Recipient        string                   `json:"recipient"`
	Subject          string                   `json:"subject"`
	Body             string                   `json:"body"`
	HTMLBody         string                   `json:"html_body,omitempty"`
	Status           string                   `json:"status"`
	AttachmentsCount int                      `json:"attachments_count"`
	Attachments      []models.EmailAttachment `json:"attachments,omitempty"`
	ErrorMessage     string                   `json:"error_message,omitempty"`
	ReceivedAt       *time.Time               `json:"received_at,omitempty"`
	SentAt           *time.Time               `json:"sent_at,omitempty"`
	CreatedAt        time.Time                `json:"created_at"`
}

// ToEmailMessageResponse converts a model to an API response.
func ToEmailMessageResponse(m *models.EmailMessage) EmailMessageResponse {
	return EmailMessageResponse{
		ID:               m.ID,
		TicketID:         m.TicketID.String(),
		Direction:        string(m.Direction),
		Sender:           m.Sender,
		Recipient:        m.Recipient,
		Subject:          m.Subject,
		Body:             m.Body,
		HTMLBody:         m.HTMLBody,
		Status:           string(m.Status),
		AttachmentsCount: m.AttachmentsCount,
		Attachments:      m.Attachments,
		ErrorMessage:     m.ErrorMessage,
		ReceivedAt:       m.ReceivedAt,
		SentAt:           m.SentAt,
		CreatedAt:        m.CreatedAt,
	}
}

// ── Email Monitor DTO ─────────────────────────────────────────────────────────

type EmailMonitorStats struct {
	TotalAccounts     int64      `json:"total_accounts"`
	ActiveAccounts    int64      `json:"active_accounts"`
	UnreadCount       int64      `json:"unread_count"`
	QueuedCount       int64      `json:"queued_count"`
	FailedCount       int64      `json:"failed_count"`
	SentToday         int64      `json:"sent_today"`
	ReceivedToday     int64      `json:"received_today"`
	LastSyncAt        *time.Time `json:"last_sync_at,omitempty"`
	Accounts          []EmailAccountResponse `json:"accounts"`
}
