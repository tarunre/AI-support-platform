package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TenantStatus represents the operational state of a tenant account.
type TenantStatus string

// SubscriptionPlan is the billing tier for a tenant.
type SubscriptionPlan string

const (
	TenantStatusActive    TenantStatus = "ACTIVE"
	TenantStatusSuspended TenantStatus = "SUSPENDED"
	TenantStatusDeleted   TenantStatus = "DELETED"
)

const (
	PlanFree       SubscriptionPlan = "FREE"
	PlanPro        SubscriptionPlan = "PRO"
	PlanEnterprise SubscriptionPlan = "ENTERPRISE"
)

// Tenant represents one company/organisation using the platform.
// Every piece of business data (tickets, users, KB, analytics…) is scoped to a
// tenant via a foreign-key tenant_id column.
type Tenant struct {
	ID               uuid.UUID        `gorm:"type:uuid;primarykey"                               json:"id"`
	Name             string           `gorm:"type:varchar(200);not null"                         json:"name"`
	Slug             string           `gorm:"type:varchar(100);uniqueIndex;not null"             json:"slug"`
	LogoURL          string           `gorm:"type:varchar(500)"                                  json:"logo_url,omitempty"`
	SubscriptionPlan SubscriptionPlan `gorm:"type:varchar(50);not null;default:'FREE'"           json:"subscription_plan"`
	Status           TenantStatus     `gorm:"type:varchar(20);not null;default:'ACTIVE';index"   json:"status"`

	// Tenant-level settings stored inline for simplicity.
	Timezone          string `gorm:"type:varchar(100);not null;default:'UTC'"       json:"timezone"`
	BrandColor        string `gorm:"type:varchar(10);not null;default:'#3B82F6'"    json:"brand_color"`
	DefaultAIModel    string `gorm:"type:varchar(100)"                              json:"default_ai_model,omitempty"`
	SupportEmail      string `gorm:"type:varchar(255)"                              json:"support_email,omitempty"`
	WorkingHoursStart int    `gorm:"not null;default:9"                             json:"working_hours_start"`
	WorkingHoursEnd   int    `gorm:"not null;default:17"                            json:"working_hours_end"`
	MaxUsersAllowed   int    `gorm:"not null;default:10"                            json:"max_users_allowed"`
	MaxTicketsPerMonth int   `gorm:"not null;default:500"                           json:"max_tickets_per_month"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BeforeCreate auto-generates a UUID.
func (t *Tenant) BeforeCreate(_ *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}
