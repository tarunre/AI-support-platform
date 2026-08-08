package dto

import "time"

// CreateIntegrationRequest is the request body for POST /integrations.
type CreateIntegrationRequest struct {
	Provider string                 `json:"provider" binding:"required"`
	Name     string                 `json:"name"     binding:"required"`
	Config   map[string]interface{} `json:"config"   binding:"required"`
	Enabled  bool                   `json:"enabled"`
}

// UpdateIntegrationRequest is the request body for PUT /integrations/:id.
type UpdateIntegrationRequest struct {
	Name    *string                `json:"name"`
	Config  map[string]interface{} `json:"config"`
	Enabled *bool                  `json:"enabled"`
}

// IntegrationResponse is the API representation of an Integration record.
// Configuration is never included; only the provider and status are exposed.
type IntegrationResponse struct {
	ID           uint      `json:"id"`
	Provider     string    `json:"provider"`
	Name         string    `json:"name"`
	Status       string    `json:"status"`
	Enabled      bool      `json:"enabled"`
	CreatedBy    uint      `json:"created_by"`
	LastSyncAt   *time.Time `json:"last_sync_at,omitempty"`
	ErrorMessage string    `json:"error_message,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// IntegrationEventResponse is the API representation of an IntegrationEvent.
type IntegrationEventResponse struct {
	ID            uint       `json:"id"`
	IntegrationID uint       `json:"integration_id"`
	EventType     string     `json:"event_type"`
	Status        string     `json:"status"`
	RetryCount    int        `json:"retry_count"`
	ErrorMessage  string     `json:"error_message,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	ProcessedAt   *time.Time `json:"processed_at,omitempty"`
}

// TicketIntegrationResponse is the API representation of a TicketIntegration.
type TicketIntegrationResponse struct {
	ID            uint       `json:"id"`
	IntegrationID uint       `json:"integration_id"`
	Provider      string     `json:"provider"`
	Name          string     `json:"name"`
	ExternalKey   string     `json:"external_key"`
	ExternalURL   string     `json:"external_url"`
	SyncedAt      *time.Time `json:"synced_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}
