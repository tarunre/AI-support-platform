package models

import (
	"time"

	"github.com/google/uuid"
)

// IntegrationProvider is the canonical identifier for a supported provider.
type IntegrationProvider string

// IntegrationStatus represents the operational state of a configured integration.
type IntegrationStatus string

const (
	IntegrationProviderSlack      IntegrationProvider = "slack"
	IntegrationProviderTeams      IntegrationProvider = "teams"
	IntegrationProviderDiscord    IntegrationProvider = "discord"
	IntegrationProviderJira       IntegrationProvider = "jira"
	IntegrationProviderLinear     IntegrationProvider = "linear"
	IntegrationProviderGitHub     IntegrationProvider = "github"
	IntegrationProviderWebhook    IntegrationProvider = "webhook"
	IntegrationProviderSalesforce IntegrationProvider = "salesforce"
	IntegrationProviderHubSpot    IntegrationProvider = "hubspot"
	IntegrationProviderGoogleCal  IntegrationProvider = "gcal"
)

const (
	IntegrationStatusActive   IntegrationStatus = "ACTIVE"
	IntegrationStatusError    IntegrationStatus = "ERROR"
	IntegrationStatusInactive IntegrationStatus = "INACTIVE"
)

// Integration represents a configured external service connection.
type Integration struct {
	ID            uint                `gorm:"primarykey;autoIncrement" json:"id"`
	TenantID      uuid.UUID           `gorm:"type:uuid;not null;default:'00000000-0000-0000-0000-000000000000';index" json:"tenant_id"`
	Provider      IntegrationProvider `gorm:"type:varchar(50);not null;index" json:"provider"`
	Name          string              `gorm:"type:varchar(200);not null" json:"name"`
	Configuration string              `gorm:"type:text" json:"-"`
	Status        IntegrationStatus   `gorm:"type:varchar(20);not null;default:'INACTIVE'" json:"status"`
	Enabled       bool                `gorm:"not null;default:false" json:"enabled"`
	CreatedBy     uint                `gorm:"not null;index" json:"created_by"`
	LastSyncAt    *time.Time          `json:"last_sync_at,omitempty"`
	ErrorMessage  string              `gorm:"type:text" json:"error_message,omitempty"`
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`

	Creator *User `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

// IntegrationEventStatus is the delivery lifecycle state of an event.
type IntegrationEventStatus string

const (
	IntEventPending    IntegrationEventStatus = "PENDING"
	IntEventProcessing IntegrationEventStatus = "PROCESSING" // claimed by a worker, in-flight
	IntEventProcessed  IntegrationEventStatus = "PROCESSED"
	IntEventFailed     IntegrationEventStatus = "FAILED"
	IntEventDead       IntegrationEventStatus = "DEAD"
)

// IntegrationEvent is a durable outbound event record used for reliable delivery.
type IntegrationEvent struct {
	ID            uint                   `gorm:"primarykey;autoIncrement" json:"id"`
	TenantID      uuid.UUID              `gorm:"type:uuid;not null;default:'00000000-0000-0000-0000-000000000000';index" json:"tenant_id"`
	IntegrationID uint                   `gorm:"not null;index" json:"integration_id"`
	ActivityID    uint                   `gorm:"not null;default:0;index" json:"activity_id"`
	EventType     string                 `gorm:"type:varchar(100);not null;index" json:"event_type"`
	Payload       string                 `gorm:"type:text" json:"payload"`
	Status        IntegrationEventStatus `gorm:"type:varchar(20);not null;default:'PENDING';index" json:"status"`
	RetryCount    int                    `gorm:"not null;default:0" json:"retry_count"`
	ErrorMessage  string                 `gorm:"type:text" json:"error_message,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	ProcessedAt   *time.Time             `json:"processed_at,omitempty"`
}

// TicketIntegration stores the link between a ticket and an external issue.
type TicketIntegration struct {
	ID            uint       `gorm:"primarykey;autoIncrement" json:"id"`
	TenantID      uuid.UUID  `gorm:"type:uuid;not null;default:'00000000-0000-0000-0000-000000000000';index" json:"tenant_id"`
	TicketID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"ticket_id"`
	IntegrationID uint       `gorm:"not null;index" json:"integration_id"`
	ExternalID    string     `gorm:"type:varchar(500)" json:"external_id"`
	ExternalKey   string     `gorm:"type:varchar(200)" json:"external_key"`
	ExternalURL   string     `gorm:"type:varchar(1000)" json:"external_url"`
	SyncedAt      *time.Time `json:"synced_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`

	Integration *Integration `gorm:"foreignKey:IntegrationID" json:"integration,omitempty"`
}
