package services

import (
	"errors"
	"net/http"

	"github.com/ayush/supportiq/internal/dto"
	"github.com/ayush/supportiq/internal/models"
	"github.com/ayush/supportiq/internal/repositories"
	"github.com/google/uuid"
)

// TenantService handles business logic for tenant management.
type TenantService struct {
	repo *repositories.TenantRepository
}

func NewTenantService(repo *repositories.TenantRepository) *TenantService {
	return &TenantService{repo: repo}
}

// List returns all tenants with optional status filter.
func (s *TenantService) List(status string, page, limit int) ([]dto.TenantResponse, int64, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	tenants, total, err := s.repo.FindAll(status, page, limit)
	if err != nil {
		return nil, 0, http.StatusInternalServerError, err
	}
	resp := make([]dto.TenantResponse, 0, len(tenants))
	for _, t := range tenants {
		resp = append(resp, toTenantResponse(t))
	}
	return resp, total, http.StatusOK, nil
}

// GetByID returns a tenant by ID.
func (s *TenantService) GetByID(id uuid.UUID) (*dto.TenantResponse, int, error) {
	t, err := s.repo.FindByID(id)
	if err != nil {
		return nil, http.StatusNotFound, errors.New("tenant not found")
	}
	resp := toTenantResponse(*t)
	return &resp, http.StatusOK, nil
}

// Create creates a new tenant (SuperAdmin operation).
func (s *TenantService) Create(req *dto.CreateTenantRequest) (*dto.TenantResponse, int, error) {
	// Check slug uniqueness
	if _, err := s.repo.FindBySlug(req.Slug); err == nil {
		return nil, http.StatusConflict, errors.New("slug already in use")
	}

	plan := models.PlanFree
	if req.SubscriptionPlan != "" {
		plan = models.SubscriptionPlan(req.SubscriptionPlan)
	}

	t := &models.Tenant{
		Name:             req.Name,
		Slug:             req.Slug,
		LogoURL:          req.LogoURL,
		SubscriptionPlan: plan,
		Status:           models.TenantStatusActive,
		Timezone:         "UTC",
		BrandColor:       "#3B82F6",
	}

	if err := s.repo.Create(t); err != nil {
		return nil, http.StatusInternalServerError, err
	}
	resp := toTenantResponse(*t)
	return &resp, http.StatusCreated, nil
}

// Update modifies a tenant's settings.
func (s *TenantService) Update(id uuid.UUID, req *dto.UpdateTenantRequest) (*dto.TenantResponse, int, error) {
	t, err := s.repo.FindByID(id)
	if err != nil {
		return nil, http.StatusNotFound, errors.New("tenant not found")
	}

	if req.Name != nil {
		t.Name = *req.Name
	}
	if req.LogoURL != nil {
		t.LogoURL = *req.LogoURL
	}
	if req.SubscriptionPlan != nil {
		t.SubscriptionPlan = models.SubscriptionPlan(*req.SubscriptionPlan)
	}
	if req.Status != nil {
		t.Status = models.TenantStatus(*req.Status)
	}
	if req.Timezone != nil {
		t.Timezone = *req.Timezone
	}
	if req.BrandColor != nil {
		t.BrandColor = *req.BrandColor
	}
	if req.DefaultAIModel != nil {
		t.DefaultAIModel = *req.DefaultAIModel
	}
	if req.SupportEmail != nil {
		t.SupportEmail = *req.SupportEmail
	}
	if req.WorkingHoursStart != nil {
		t.WorkingHoursStart = *req.WorkingHoursStart
	}
	if req.WorkingHoursEnd != nil {
		t.WorkingHoursEnd = *req.WorkingHoursEnd
	}

	if err := s.repo.Update(t); err != nil {
		return nil, http.StatusInternalServerError, err
	}
	resp := toTenantResponse(*t)
	return &resp, http.StatusOK, nil
}

// Delete soft-deletes a tenant (marks as DELETED).
func (s *TenantService) Delete(id uuid.UUID) (int, error) {
	if _, err := s.repo.FindByID(id); err != nil {
		return http.StatusNotFound, errors.New("tenant not found")
	}
	if err := s.repo.Delete(id); err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}

// GetPlatformOverview returns the SuperAdmin dashboard overview.
func (s *TenantService) GetPlatformOverview() (*dto.SuperAdminOverview, int, error) {
	stats, err := s.repo.PlatformStats()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	// Per-tenant stats
	tenants, _, err := s.repo.FindAll("ACTIVE", 1, 50)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	tenantStats := make([]dto.TenantStatsResponse, 0, len(tenants))
	for _, t := range tenants {
		users, _ := s.repo.CountUsers(t.ID)
		tickets, _ := s.repo.CountTickets(t.ID)
		open, _ := s.repo.CountOpenTickets(t.ID)
		ai, _ := s.repo.CountAIRequests(t.ID)
		tenantStats = append(tenantStats, dto.TenantStatsResponse{
			TenantID:     t.ID.String(),
			TenantName:   t.Name,
			TotalUsers:   users,
			TotalTickets: tickets,
			OpenTickets:  open,
			AIRequests:   ai,
		})
	}

	return &dto.SuperAdminOverview{
		TotalTenants:  stats.TotalTenants,
		ActiveTenants: stats.ActiveTenants,
		TotalUsers:    stats.TotalUsers,
		TotalTickets:  stats.TotalTickets,
		AIUsageToday:  stats.AIUsageToday,
		TenantStats:   tenantStats,
	}, http.StatusOK, nil
}

func toTenantResponse(t models.Tenant) dto.TenantResponse {
	return dto.TenantResponse{
		ID:                t.ID.String(),
		Name:              t.Name,
		Slug:              t.Slug,
		LogoURL:           t.LogoURL,
		SubscriptionPlan:  string(t.SubscriptionPlan),
		Status:            string(t.Status),
		Timezone:          t.Timezone,
		BrandColor:        t.BrandColor,
		DefaultAIModel:    t.DefaultAIModel,
		SupportEmail:      t.SupportEmail,
		WorkingHoursStart: t.WorkingHoursStart,
		WorkingHoursEnd:   t.WorkingHoursEnd,
		MaxUsersAllowed:   t.MaxUsersAllowed,
		MaxTicketsPerMonth: t.MaxTicketsPerMonth,
		CreatedAt:         t.CreatedAt,
		UpdatedAt:         t.UpdatedAt,
	}
}
