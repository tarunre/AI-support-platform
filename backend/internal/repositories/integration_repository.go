package repositories

import (
	"time"

	"github.com/ayush/supportiq/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IntegrationRepository handles all database access for integrations.
type IntegrationRepository struct {
	db *gorm.DB
}

func NewIntegrationRepository(db *gorm.DB) *IntegrationRepository {
	return &IntegrationRepository{db: db}
}

// ─── Integration CRUD ───────────────────────────────────────────────────────

func (r *IntegrationRepository) Create(integration *models.Integration) error {
	return r.db.Create(integration).Error
}

func (r *IntegrationRepository) FindByID(tenantID uuid.UUID, id uint) (*models.Integration, error) {
	var integration models.Integration
	err := r.db.Preload("Creator").Where("tenant_id = ? AND id = ?", tenantID, id).First(&integration).Error
	if err != nil {
		return nil, err
	}
	return &integration, nil
}

// FindByIDNoTenant loads an integration by ID without tenant scoping.
// Used by the public Slack events webhook where tenant is unknown.
func (r *IntegrationRepository) FindByIDNoTenant(id uint) (*models.Integration, error) {
	var integration models.Integration
	err := r.db.Where("id = ?", id).First(&integration).Error
	if err != nil {
		return nil, err
	}
	return &integration, nil
}

// FindBySlackThread looks up a TicketIntegration by the Slack thread_ts.
// Used to map thread replies back to their originating ticket.
func (r *IntegrationRepository) FindBySlackThread(integrationID uint, threadTS string) (*models.TicketIntegration, error) {
	var ti models.TicketIntegration
	err := r.db.Where("integration_id = ? AND external_id = ?", integrationID, threadTS).
		First(&ti).Error
	if err != nil {
		return nil, err
	}
	return &ti, nil
}

func (r *IntegrationRepository) FindAll(tenantID uuid.UUID) ([]models.Integration, error) {
	var integrations []models.Integration
	err := r.db.Where("tenant_id = ?", tenantID).Order("created_at DESC").Find(&integrations).Error
	return integrations, err
}

func (r *IntegrationRepository) FindEnabled(tenantID uuid.UUID) ([]models.Integration, error) {
	var integrations []models.Integration
	err := r.db.Where("tenant_id = ? AND enabled = true AND status != ?", tenantID, models.IntegrationStatusError).
		Order("created_at DESC").
		Find(&integrations).Error
	return integrations, err
}

// FindAllEnabled returns enabled integrations across all tenants (used by background workers).
func (r *IntegrationRepository) FindAllEnabled() ([]models.Integration, error) {
	var integrations []models.Integration
	err := r.db.Where("enabled = true AND status != ?", models.IntegrationStatusError).
		Order("created_at DESC").
		Find(&integrations).Error
	return integrations, err
}

func (r *IntegrationRepository) FindByProvider(tenantID uuid.UUID, provider models.IntegrationProvider) ([]models.Integration, error) {
	var integrations []models.Integration
	err := r.db.Where("tenant_id = ? AND provider = ? AND enabled = true", tenantID, provider).Find(&integrations).Error
	return integrations, err
}

func (r *IntegrationRepository) Update(integration *models.Integration) error {
	return r.db.Save(integration).Error
}

func (r *IntegrationRepository) Delete(tenantID uuid.UUID, id uint) error {
	return r.db.Where("tenant_id = ?", tenantID).Delete(&models.Integration{}, id).Error
}

// ─── IntegrationEvent ───────────────────────────────────────────────────────

func (r *IntegrationRepository) CreateEvent(event *models.IntegrationEvent) error {
	return r.db.Create(event).Error
}

// EventExistsForActivity returns true if an IntegrationEvent was already
// created for this (activityID, integrationID) pair. Skipped for legacy
// rows that have activityID=0 (created before the dedup field existed).
func (r *IntegrationRepository) EventExistsForActivity(activityID, integrationID uint) bool {
	if activityID == 0 {
		return false
	}
	var count int64
	r.db.Model(&models.IntegrationEvent{}).
		Where("activity_id = ? AND integration_id = ?", activityID, integrationID).
		Count(&count)
	return count > 0
}

func (r *IntegrationRepository) FindPendingEvents(tenantID uuid.UUID, limit int) ([]models.IntegrationEvent, error) {
	var events []models.IntegrationEvent
	err := r.db.
		Where("tenant_id = ? AND status IN (?, ?)", tenantID, models.IntEventPending, models.IntEventFailed).
		Where("retry_count < 5").
		Order("created_at ASC").
		Limit(limit).
		Find(&events).Error
	return events, err
}

// FindAllPendingEvents returns pending events across all tenants (for background worker).
func (r *IntegrationRepository) FindAllPendingEvents(limit int) ([]models.IntegrationEvent, error) {
	var events []models.IntegrationEvent
	err := r.db.
		Where("status IN (?, ?)", models.IntEventPending, models.IntEventFailed).
		Where("retry_count < 5").
		Order("created_at ASC").
		Limit(limit).
		Find(&events).Error
	return events, err
}

// ClaimPendingEvents atomically marks up to `limit` PENDING/FAILED events as
// PROCESSING and returns them. Uses a single UPDATE…RETURNING statement so that
// two concurrent workers can never claim the same event, eliminating the
// duplicate-notification race condition.
func (r *IntegrationRepository) ClaimPendingEvents(limit int) ([]models.IntegrationEvent, error) {
	var events []models.IntegrationEvent
	err := r.db.Raw(`
		UPDATE integration_events
		SET status = ?
		WHERE id IN (
			SELECT id FROM integration_events
			WHERE status IN (?, ?) AND retry_count < 5
			ORDER BY created_at ASC
			LIMIT ?
		)
		RETURNING *`,
		models.IntEventProcessing,
		models.IntEventPending, models.IntEventFailed,
		limit,
	).Scan(&events).Error
	return events, err
}

// ResetStuckProcessingEvents resets any events left in PROCESSING state back
// to PENDING. Called on worker startup to recover from a previous crash.
func (r *IntegrationRepository) ResetStuckProcessingEvents() error {
	return r.db.Model(&models.IntegrationEvent{}).
		Where("status = ?", models.IntEventProcessing).
		Update("status", models.IntEventPending).Error
}

func (r *IntegrationRepository) UpdateEventStatus(id uint, status models.IntegrationEventStatus, errMsg string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":        status,
		"error_message": errMsg,
	}
	if status == models.IntEventProcessed {
		updates["processed_at"] = now
	}
	return r.db.Model(&models.IntegrationEvent{}).Where("id = ?", id).Updates(updates).Error
}

func (r *IntegrationRepository) IncrementEventRetry(id uint) error {
	return r.db.Model(&models.IntegrationEvent{}).Where("id = ?", id).
		UpdateColumn("retry_count", gorm.Expr("retry_count + 1")).Error
}

func (r *IntegrationRepository) MarkEventDead(id uint, errMsg string) error {
	return r.UpdateEventStatus(id, models.IntEventDead, errMsg)
}

// ─── TicketIntegration ──────────────────────────────────────────────────────

func (r *IntegrationRepository) CreateTicketIntegration(ti *models.TicketIntegration) error {
	return r.db.Create(ti).Error
}

func (r *IntegrationRepository) FindTicketIntegrations(tenantID uuid.UUID, ticketID uuid.UUID) ([]models.TicketIntegration, error) {
	var items []models.TicketIntegration
	err := r.db.Preload("Integration").
		Where("tenant_id = ? AND ticket_id = ?", tenantID, ticketID).
		Find(&items).Error
	return items, err
}

func (r *IntegrationRepository) FindTicketIntegrationByProvider(tenantID uuid.UUID, ticketID uuid.UUID, integrationID uint) (*models.TicketIntegration, error) {
	var ti models.TicketIntegration
	err := r.db.Where("tenant_id = ? AND ticket_id = ? AND integration_id = ?", tenantID, ticketID, integrationID).
		First(&ti).Error
	if err != nil {
		return nil, err
	}
	return &ti, nil
}
