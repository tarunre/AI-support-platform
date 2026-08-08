package models

import (
	"time"

	"github.com/google/uuid"
)

// EmailDirection indicates whether a message was received or sent.
type EmailDirection string

// EmailStatus tracks delivery lifecycle.
type EmailStatus string

const (
	EmailDirectionInbound  EmailDirection = "INBOUND"
	EmailDirectionOutbound EmailDirection = "OUTBOUND"
)

const (
	EmailStatusReceived  EmailStatus = "RECEIVED"
	EmailStatusQueued    EmailStatus = "QUEUED"
	EmailStatusSent      EmailStatus = "SENT"
	EmailStatusFailed    EmailStatus = "FAILED"
	EmailStatusDelivered EmailStatus = "DELIVERED"
	EmailStatusRead      EmailStatus = "READ"
)

// EmailAttachment stores metadata for a single file attachment.
type EmailAttachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
	StoragePath string `json:"storage_path"`
}

// EmailMessage represents one email in either direction linked to a ticket.
type EmailMessage struct {
	ID               uint              `gorm:"primarykey;autoIncrement" json:"id"`
	TenantID         uuid.UUID         `gorm:"type:uuid;not null;default:'00000000-0000-0000-0000-000000000000';index" json:"tenant_id"`
	TicketID         uuid.UUID         `gorm:"type:uuid;not null;index" json:"ticket_id"`
	AccountID        uint              `gorm:"index" json:"account_id"`
	MessageID        string            `gorm:"type:varchar(500);uniqueIndex" json:"message_id"`
	ThreadID         string            `gorm:"type:varchar(500);index" json:"thread_id,omitempty"`
	InReplyTo        string            `gorm:"type:varchar(500)" json:"in_reply_to,omitempty"`
	References       string            `gorm:"type:text" json:"references,omitempty"`
	Direction        EmailDirection    `gorm:"type:varchar(20);not null;index" json:"direction"`
	Sender           string            `gorm:"type:varchar(255);not null" json:"sender"`
	Recipient        string            `gorm:"type:varchar(255);not null" json:"recipient"`
	Subject          string            `gorm:"type:varchar(500)" json:"subject"`
	Body             string            `gorm:"type:text" json:"body"`
	HTMLBody         string            `gorm:"type:text" json:"html_body,omitempty"`
	Status           EmailStatus       `gorm:"type:varchar(20);not null;default:'RECEIVED';index" json:"status"`
	AttachmentsCount int               `gorm:"not null;default:0" json:"attachments_count"`
	Attachments      []EmailAttachment `gorm:"serializer:json" json:"attachments,omitempty"`
	ErrorMessage     string            `gorm:"type:text" json:"error_message,omitempty"`
	RetryCount       int               `gorm:"not null;default:0" json:"retry_count"`
	RawHeaders       string            `gorm:"type:text" json:"-"`
	ReceivedAt       *time.Time        `json:"received_at,omitempty"`
	SentAt           *time.Time        `json:"sent_at,omitempty"`
	CreatedAt        time.Time         `json:"created_at"`

	Ticket  *Ticket       `gorm:"foreignKey:TicketID"  json:"ticket,omitempty"`
	Account *EmailAccount `gorm:"foreignKey:AccountID" json:"account,omitempty"`
}
