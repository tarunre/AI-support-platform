package repositories

import (
	"time"

	"github.com/ayush/supportiq/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SLARepository manages database access for SLA policies.
type SLARepository struct {
	db *gorm.DB
}

func NewSLARepository(db *gorm.DB) *SLARepository {
	return &SLARepository{db: db}
}

func (r *SLARepository) scoped(tenantID uuid.UUID) *gorm.DB {
	return r.db.Where("tenant_id = ?", tenantID)
}

func (r *SLARepository) Create(p *models.SLAPolicy) error {
	return r.db.Create(p).Error
}

func (r *SLARepository) List(tenantID uuid.UUID) ([]models.SLAPolicy, error) {
	var policies []models.SLAPolicy
	err := r.scoped(tenantID).Order("priority ASC, created_at ASC").Find(&policies).Error
	return policies, err
}

func (r *SLARepository) FindByID(tenantID uuid.UUID, id uint) (*models.SLAPolicy, error) {
	var p models.SLAPolicy
	err := r.scoped(tenantID).First(&p, id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// FindDefault returns the tenant's default SLA policy.
func (r *SLARepository) FindDefault(tenantID uuid.UUID) (*models.SLAPolicy, error) {
	var p models.SLAPolicy
	err := r.scoped(tenantID).Where("is_default = true").First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// FindByPriority returns a policy matching the given ticket priority for the tenant.
func (r *SLARepository) FindByPriority(tenantID uuid.UUID, priority models.TicketPriority) (*models.SLAPolicy, error) {
	var p models.SLAPolicy
	err := r.scoped(tenantID).Where("priority = ?", priority).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// UnsetAllDefaults clears the is_default flag on every policy for a tenant.
// Call this before setting a new default.
func (r *SLARepository) UnsetAllDefaults(tenantID uuid.UUID) error {
	return r.db.Model(&models.SLAPolicy{}).
		Where("tenant_id = ?", tenantID).
		Update("is_default", false).Error
}

func (r *SLARepository) Update(p *models.SLAPolicy) error {
	return r.db.Save(p).Error
}

func (r *SLARepository) Delete(tenantID uuid.UUID, id uint) error {
	return r.scoped(tenantID).Delete(&models.SLAPolicy{}, id).Error
}

// ─── Ticket SLA queries ───────────────────────────────────────────────────────

// SLATicketStats holds aggregated SLA compliance numbers for a tenant.
type SLATicketStats struct {
	Total               int64
	Breached            int64
	CompletedOnTime     int64
	AvgFirstResponseMin float64
	AvgResolutionMin    float64
}

// SLAStats returns aggregated SLA compliance numbers for a tenant.
func SLAStats(db *gorm.DB, tenantID uuid.UUID) (SLATicketStats, error) {
	var s SLATicketStats

	db.Model(&models.Ticket{}).
		Where("tenant_id = ? AND sla_policy_id IS NOT NULL", tenantID).
		Count(&s.Total)

	db.Model(&models.Ticket{}).
		Where("tenant_id = ? AND sla_status = ?", tenantID, string(models.SLAStatusBreached)).
		Count(&s.Breached)

	db.Model(&models.Ticket{}).
		Where("tenant_id = ? AND sla_status = ?", tenantID, string(models.SLAStatusCompleted)).
		Count(&s.CompletedOnTime)

	db.Raw(`
		SELECT COALESCE(AVG(EXTRACT(EPOCH FROM (first_response_completed_at - created_at)) / 60), 0)
		FROM tickets
		WHERE tenant_id = ? AND first_response_completed_at IS NOT NULL AND deleted_at IS NULL
	`, tenantID).Scan(&s.AvgFirstResponseMin)

	db.Raw(`
		SELECT COALESCE(AVG(EXTRACT(EPOCH FROM (resolved_at - created_at)) / 60), 0)
		FROM tickets
		WHERE tenant_id = ? AND resolved_at IS NOT NULL AND deleted_at IS NULL
	`, tenantID).Scan(&s.AvgResolutionMin)

	return s, nil
}

// FindNearBreach returns open tickets whose resolution_due_at is within withinMinutes from now.
func FindNearBreach(db *gorm.DB, tenantID uuid.UUID, withinMinutes int) ([]models.Ticket, error) {
	var tickets []models.Ticket
	cutoff := time.Now().Add(time.Duration(withinMinutes) * time.Minute)
	err := db.
		Where("tenant_id = ?", tenantID).
		Where("status IN ?", []string{"OPEN", "IN_PROGRESS"}).
		Where("resolution_due_at IS NOT NULL AND resolution_due_at <= ?", cutoff).
		Where("sla_status = ?", string(models.SLAStatusAtRisk)).
		Order("resolution_due_at ASC").
		Limit(50).
		Find(&tickets).Error
	return tickets, err
}

// FindBreachedOpen returns open/in-progress tickets that have already breached their SLA.
func FindBreachedOpen(db *gorm.DB, tenantID uuid.UUID) ([]models.Ticket, error) {
	var tickets []models.Ticket
	err := db.
		Where("tenant_id = ?", tenantID).
		Where("status IN ?", []string{"OPEN", "IN_PROGRESS"}).
		Where("sla_status = ?", string(models.SLAStatusBreached)).
		Order("resolution_due_at ASC").
		Limit(50).
		Find(&tickets).Error
	return tickets, err
}
