package repositories

import (
	"context"

	"github.com/ayush/supportiq/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// KnowledgeRepository handles all database access for the knowledge_bases table.
type KnowledgeRepository struct {
	db *gorm.DB
}

func NewKnowledgeRepository(db *gorm.DB) *KnowledgeRepository {
	return &KnowledgeRepository{db: db}
}

func (r *KnowledgeRepository) Create(doc *models.KnowledgeBase) error {
	return r.db.Create(doc).Error
}

func (r *KnowledgeRepository) FindByID(tenantID uuid.UUID, id uint) (*models.KnowledgeBase, error) {
	var doc models.KnowledgeBase
	if err := r.db.Where("tenant_id = ? AND id = ?", tenantID, id).First(&doc).Error; err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *KnowledgeRepository) Update(doc *models.KnowledgeBase) error {
	return r.db.Save(doc).Error
}

func (r *KnowledgeRepository) Delete(tenantID uuid.UUID, id uint) error {
	return r.db.Where("tenant_id = ?", tenantID).Delete(&models.KnowledgeBase{}, id).Error
}

func (r *KnowledgeRepository) List(tenantID uuid.UUID, search, category string, activeOnly bool, page, limit int) ([]models.KnowledgeBase, int64, error) {
	q := r.db.Model(&models.KnowledgeBase{}).Where("tenant_id = ?", tenantID)

	if search != "" {
		pattern := "%" + search + "%"
		q = q.Where("title ILIKE ? OR content ILIKE ?", pattern, pattern)
	}
	if category != "" {
		q = q.Where("category = ?", category)
	}
	if activeOnly {
		q = q.Where("is_active = true")
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	var docs []models.KnowledgeBase
	err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&docs).Error
	return docs, total, err
}

// Search performs a keyword search over active documents for a specific tenant.
func (r *KnowledgeRepository) Search(_ context.Context, tenantID uuid.UUID, query string, limit int) ([]models.KnowledgeBase, error) {
	if limit <= 0 {
		limit = 5
	}
	pattern := "%" + query + "%"
	var docs []models.KnowledgeBase
	err := r.db.
		Where("tenant_id = ? AND is_active = true AND (title ILIKE ? OR content ILIKE ?)", tenantID, pattern, pattern).
		Order("updated_at DESC").
		Limit(limit).
		Find(&docs).Error
	return docs, err
}
