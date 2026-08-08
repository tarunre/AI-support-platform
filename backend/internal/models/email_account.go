package models

import (
	"time"

	"github.com/google/uuid"
)

// EmailProvider identifies the email service provider.
type EmailProvider string

const (
	EmailProviderSMTPIMAP EmailProvider = "SMTP_IMAP"
	EmailProviderGmail    EmailProvider = "GMAIL"
	EmailProviderOutlook  EmailProvider = "OUTLOOK"
	EmailProviderSendGrid EmailProvider = "SENDGRID"
	EmailProviderSES      EmailProvider = "SES"
	EmailProviderMailgun  EmailProvider = "MAILGUN"
)

// EmailAccount holds IMAP/SMTP credentials for a configured mailbox.
// Passwords are stored AES-256-GCM encrypted and never returned through APIs.
type EmailAccount struct {
	ID                uint          `gorm:"primarykey;autoIncrement" json:"id"`
	TenantID          uuid.UUID     `gorm:"type:uuid;not null;default:'00000000-0000-0000-0000-000000000000';uniqueIndex:idx_tenant_email_addr;index" json:"tenant_id"`
	Provider          EmailProvider `gorm:"type:varchar(50);not null" json:"provider"`
	EmailAddress      string        `gorm:"type:varchar(255);not null;uniqueIndex:idx_tenant_email_addr" json:"email_address"`
	DisplayName       string        `gorm:"type:varchar(100)" json:"display_name"`
	IMAPHost          string        `gorm:"type:varchar(255)" json:"imap_host"`
	IMAPPort          int           `gorm:"default:993" json:"imap_port"`
	IMAPUseTLS        bool          `gorm:"not null;default:true" json:"imap_use_tls"`
	SMTPHost          string        `gorm:"type:varchar(255)" json:"smtp_host"`
	SMTPPort          int           `gorm:"default:587" json:"smtp_port"`
	SMTPImplicitTLS   bool          `gorm:"not null;default:false" json:"smtp_implicit_tls"`
	Username          string        `gorm:"type:varchar(255);not null" json:"username"`
	EncryptedPassword string        `gorm:"type:text;not null" json:"-"`
	IsActive          bool          `gorm:"not null;default:true;index" json:"is_active"`
	LastSyncAt        *time.Time    `json:"last_sync_at,omitempty"`
	CreatedAt         time.Time     `json:"created_at"`
	UpdatedAt         time.Time     `json:"updated_at"`
}
