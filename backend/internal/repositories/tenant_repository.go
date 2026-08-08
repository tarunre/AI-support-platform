package repositories

import (
	"time"

	"github.com/ayush/supportiq/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TenantRepository handles all database access for tenants.
type TenantRepository struct {
	db *gorm.DB
}

func NewTenantRepository(db *gorm.DB) *TenantRepository {
	return &TenantRepository{db: db}
}

func (r *TenantRepository) Create(t *models.Tenant) error {
	return r.db.Create(t).Error
}

func (r *TenantRepository) FindByID(id uuid.UUID) (*models.Tenant, error) {
	var t models.Tenant
	if err := r.db.First(&t, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TenantRepository) FindBySlug(slug string) (*models.Tenant, error) {
	var t models.Tenant
	if err := r.db.Where("slug = ?", slug).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TenantRepository) FindAll(status string, page, limit int) ([]models.Tenant, int64, error) {
	q := r.db.Model(&models.Tenant{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	var tenants []models.Tenant
	err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&tenants).Error
	return tenants, total, err
}

func (r *TenantRepository) Update(t *models.Tenant) error {
	return r.db.Save(t).Error
}

func (r *TenantRepository) Delete(id uuid.UUID) error {
	// Deactivate all users in this tenant so they cannot log in.
	if err := r.db.Model(&models.User{}).
		Where("tenant_id = ?", id).
		Update("is_active", false).Error; err != nil {
		return err
	}
	// Soft-delete: mark status as DELETED
	return r.db.Model(&models.Tenant{}).Where("id = ?", id).
		Update("status", models.TenantStatusDeleted).Error
}

// CountUsers returns the number of active users for a tenant.
func (r *TenantRepository) CountUsers(tenantID uuid.UUID) (int64, error) {
	var n int64
	return n, r.db.Model(&models.User{}).
		Where("tenant_id = ? AND is_active = true", tenantID).Count(&n).Error
}

// CountTickets returns the total ticket count for a tenant.
func (r *TenantRepository) CountTickets(tenantID uuid.UUID) (int64, error) {
	var n int64
	return n, r.db.Model(&models.Ticket{}).
		Where("tenant_id = ?", tenantID).Count(&n).Error
}

// CountOpenTickets returns the open ticket count for a tenant.
func (r *TenantRepository) CountOpenTickets(tenantID uuid.UUID) (int64, error) {
	var n int64
	return n, r.db.Model(&models.Ticket{}).
		Where("tenant_id = ? AND status = 'OPEN'", tenantID).Count(&n).Error
}

// CountAIRequests returns AI reply count for a tenant today.
func (r *TenantRepository) CountAIRequests(tenantID uuid.UUID) (int64, error) {
	var n int64
	today := time.Now().Truncate(24 * time.Hour)
	return n, r.db.Model(&models.AIReply{}).
		Where("tenant_id = ? AND created_at >= ?", tenantID, today).Count(&n).Error
}

// PlatformStats returns global counts across all tenants (SuperAdmin).
type PlatformStats struct {
	TotalTenants  int64
	ActiveTenants int64
	TotalUsers    int64
	TotalTickets  int64
	AIUsageToday  int64
}

func (r *TenantRepository) PlatformStats() (*PlatformStats, error) {
	var stats PlatformStats
	r.db.Model(&models.Tenant{}).Count(&stats.TotalTenants)
	r.db.Model(&models.Tenant{}).Where("status = 'ACTIVE'").Count(&stats.ActiveTenants)
	r.db.Model(&models.User{}).Where("role != 'SuperAdmin'").Count(&stats.TotalUsers)
	r.db.Model(&models.Ticket{}).Count(&stats.TotalTickets)
	today := time.Now().Truncate(24 * time.Hour)
	r.db.Model(&models.AIReply{}).Where("created_at >= ?", today).Count(&stats.AIUsageToday)
	return &stats, nil
}

// AllActiveTenantIDs returns all tenant IDs with ACTIVE status.
// Used by the analytics aggregator and workers to iterate all tenants.
func (r *TenantRepository) AllActiveTenantIDs() ([]uuid.UUID, error) {
	var tenants []models.Tenant
	if err := r.db.Select("id").Where("status = 'ACTIVE'").Find(&tenants).Error; err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, len(tenants))
	for i, t := range tenants {
		ids[i] = t.ID
	}
	return ids, nil
}
