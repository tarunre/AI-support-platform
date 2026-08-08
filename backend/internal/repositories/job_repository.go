package repositories

import (
	"time"

	"github.com/ayush/supportiq/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// JobRepository handles all database access for background_jobs.
type JobRepository struct {
	db *gorm.DB
}

func NewJobRepository(db *gorm.DB) *JobRepository {
	return &JobRepository{db: db}
}

func (r *JobRepository) Create(job *models.BackgroundJob) error {
	return r.db.Create(job).Error
}

// FindByID looks up a job by ID — used by worker/retry logic (no tenant filter needed).
func (r *JobRepository) FindByID(id uint) (*models.BackgroundJob, error) {
	var job models.BackgroundJob
	if err := r.db.First(&job, id).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *JobRepository) Update(job *models.BackgroundJob) error {
	return r.db.Save(job).Error
}

// List returns jobs with optional status/type filter and pagination (no tenant filter — admin use).
func (r *JobRepository) List(status, jobType string, page, limit int) ([]models.BackgroundJob, int64, error) {
	var jobs []models.BackgroundJob
	var total int64
	offset := (page - 1) * limit
	q := r.db.Model(&models.BackgroundJob{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if jobType != "" {
		q = q.Where("job_type = ?", jobType)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&jobs).Error; err != nil {
		return nil, 0, err
	}
	return jobs, total, nil
}

// ListByTenant returns jobs scoped to a specific tenant.
func (r *JobRepository) ListByTenant(tenantID uuid.UUID, page, limit int) ([]models.BackgroundJob, int64, error) {
	var jobs []models.BackgroundJob
	var total int64
	offset := (page - 1) * limit
	q := r.db.Model(&models.BackgroundJob{}).Where("tenant_id = ?", tenantID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&jobs).Error; err != nil {
		return nil, 0, err
	}
	return jobs, total, nil
}

// MarkProcessing sets a job to PROCESSING and records start time.
func (r *JobRepository) MarkProcessing(id uint) error {
	now := time.Now()
	return r.db.Model(&models.BackgroundJob{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     models.JobStatusProcessing,
		"started_at": now,
	}).Error
}

// MarkCompleted sets a job to COMPLETED and records finish time.
func (r *JobRepository) MarkCompleted(id uint) error {
	now := time.Now()
	return r.db.Model(&models.BackgroundJob{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":       models.JobStatusCompleted,
		"completed_at": now,
	}).Error
}

// MarkFailed sets a job to FAILED with an error message.
func (r *JobRepository) MarkFailed(id uint, errMsg string) error {
	return r.db.Model(&models.BackgroundJob{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":        models.JobStatusFailed,
		"error_message": errMsg,
	}).Error
}

// MarkDead sets a job to DEAD status.
func (r *JobRepository) MarkDead(id uint, errMsg string) error {
	return r.db.Model(&models.BackgroundJob{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":        models.JobStatusDead,
		"error_message": errMsg,
	}).Error
}

// MarkRetrying sets a job to RETRYING and resets its retry count.
func (r *JobRepository) MarkRetrying(id uint, retryCount int, note string) error {
	return r.db.Model(&models.BackgroundJob{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":        models.JobStatusRetrying,
		"retry_count":   retryCount,
		"error_message": note,
	}).Error
}

// Statistics returns job counts grouped by status.
func (r *JobRepository) Statistics() (map[string]int64, error) {
	var rows []struct {
		Status string
		Count  int64
	}
	if err := r.db.Model(&models.BackgroundJob{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	stats := make(map[string]int64)
	for _, row := range rows {
		stats[row.Status] = row.Count
	}
	return stats, nil
}
