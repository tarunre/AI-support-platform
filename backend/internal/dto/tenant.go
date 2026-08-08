package dto

import "time"

// CreateTenantRequest is the body for POST /tenants (SuperAdmin only).
type CreateTenantRequest struct {
	Name             string `json:"name"   binding:"required,min=2,max=200"`
	Slug             string `json:"slug"   binding:"required,min=2,max=100,alphanum"`
	LogoURL          string `json:"logo_url"`
	SubscriptionPlan string `json:"subscription_plan"`
}

// UpdateTenantRequest is the body for PUT /tenants/:id.
type UpdateTenantRequest struct {
	Name              *string `json:"name"`
	LogoURL           *string `json:"logo_url"`
	SubscriptionPlan  *string `json:"subscription_plan"`
	Status            *string `json:"status"`
	Timezone          *string `json:"timezone"`
	BrandColor        *string `json:"brand_color"`
	DefaultAIModel    *string `json:"default_ai_model"`
	SupportEmail      *string `json:"support_email"`
	WorkingHoursStart *int    `json:"working_hours_start"`
	WorkingHoursEnd   *int    `json:"working_hours_end"`
}

// TenantResponse is the public representation of a Tenant record.
type TenantResponse struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Slug             string    `json:"slug"`
	LogoURL          string    `json:"logo_url,omitempty"`
	SubscriptionPlan string    `json:"subscription_plan"`
	Status           string    `json:"status"`
	Timezone         string    `json:"timezone"`
	BrandColor       string    `json:"brand_color"`
	DefaultAIModel   string    `json:"default_ai_model,omitempty"`
	SupportEmail     string    `json:"support_email,omitempty"`
	WorkingHoursStart int      `json:"working_hours_start"`
	WorkingHoursEnd   int      `json:"working_hours_end"`
	MaxUsersAllowed   int      `json:"max_users_allowed"`
	MaxTicketsPerMonth int     `json:"max_tickets_per_month"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// TenantStatsResponse contains aggregate statistics for a tenant (SuperAdmin view).
type TenantStatsResponse struct {
	TenantID     string `json:"tenant_id"`
	TenantName   string `json:"tenant_name"`
	TotalUsers   int64  `json:"total_users"`
	TotalTickets int64  `json:"total_tickets"`
	OpenTickets  int64  `json:"open_tickets"`
	AIRequests   int64  `json:"ai_requests"`
}

// SuperAdminOverview is the SuperAdmin dashboard overview.
type SuperAdminOverview struct {
	TotalTenants   int64  `json:"total_tenants"`
	ActiveTenants  int64  `json:"active_tenants"`
	TotalUsers     int64  `json:"total_users"`
	TotalTickets   int64  `json:"total_tickets"`
	AIUsageToday   int64  `json:"ai_usage_today"`
	TenantStats    []TenantStatsResponse `json:"tenant_stats"`
}

// RegisterWithTenantRequest extends RegisterRequest with tenant creation fields.
type RegisterWithTenantRequest struct {
	Name        string `json:"name"         binding:"required,min=2,max=100"`
	Email       string `json:"email"        binding:"required,email"`
	Password    string `json:"password"     binding:"required,min=8"`
	CompanyName string `json:"company_name" binding:"required,min=2,max=200"`
	CompanySlug string `json:"company_slug"`
}
