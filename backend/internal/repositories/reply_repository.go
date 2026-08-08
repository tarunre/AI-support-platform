package repositories

import (
	"github.com/ayush/supportiq/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ReplyRepository handles all database access for ai_replies.
type ReplyRepository struct {
	db *gorm.DB
}

func NewReplyRepository(db *gorm.DB) *ReplyRepository {
	return &ReplyRepository{db: db}
}

func (r *ReplyRepository) Create(reply *models.AIReply) error {
	return r.db.Create(reply).Error
}

func (r *ReplyRepository) FindByID(tenantID uuid.UUID, id uint) (*models.AIReply, error) {
	var reply models.AIReply
	if err := r.db.Where("tenant_id = ? AND id = ?", tenantID, id).First(&reply).Error; err != nil {
		return nil, err
	}
	return &reply, nil
}

func (r *ReplyRepository) ListByTicketID(tenantID uuid.UUID, ticketID uuid.UUID) ([]models.AIReply, error) {
	var replies []models.AIReply
	err := r.db.
		Where("tenant_id = ? AND ticket_id = ?", tenantID, ticketID).
		Order("created_at DESC").
		Find(&replies).Error
	return replies, err
}

func (r *ReplyRepository) Update(reply *models.AIReply) error {
	return r.db.Save(reply).Error
}

func (r *ReplyRepository) FindLatestByTicketID(tenantID uuid.UUID, ticketID uuid.UUID) (*models.AIReply, error) {
	var reply models.AIReply
	err := r.db.Where("tenant_id = ? AND ticket_id = ?", tenantID, ticketID).Order("created_at DESC").First(&reply).Error
	if err != nil {
		return nil, err
	}
	return &reply, nil
}

func (r *ReplyRepository) FindAllByTicketID(tenantID uuid.UUID, ticketID uuid.UUID) ([]models.AIReply, error) {
	var replies []models.AIReply
	err := r.db.Where("tenant_id = ? AND ticket_id = ?", tenantID, ticketID).Order("created_at DESC").Find(&replies).Error
	return replies, err
}
