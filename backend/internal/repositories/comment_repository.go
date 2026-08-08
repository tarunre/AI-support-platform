package repositories

import (
	"github.com/ayush/supportiq/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CommentRepository handles all database access for ticket comments.
type CommentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) Create(comment *models.TicketComment) error {
	return r.db.Create(comment).Error
}

func (r *CommentRepository) FindByID(tenantID uuid.UUID, id uint) (*models.TicketComment, error) {
	var comment models.TicketComment
	if err := r.db.Preload("User").Where("tenant_id = ? AND id = ?", tenantID, id).First(&comment).Error; err != nil {
		return nil, err
	}
	return &comment, nil
}

func (r *CommentRepository) ListByTicketID(tenantID uuid.UUID, ticketID uuid.UUID) ([]models.TicketComment, error) {
	var comments []models.TicketComment
	err := r.db.
		Preload("User").
		Where("tenant_id = ? AND ticket_id = ?", tenantID, ticketID).
		Order("created_at ASC").
		Find(&comments).Error
	return comments, err
}

// ListPublicByTicketUnscoped returns PUBLIC + CUSTOMER comments for a ticket
// without a tenant filter — used by the customer portal.
func (r *CommentRepository) ListPublicByTicketUnscoped(ticketID uuid.UUID) ([]models.TicketComment, error) {
	var comments []models.TicketComment
	err := r.db.
		Preload("User").
		Where("ticket_id = ? AND comment_type IN ('PUBLIC','CUSTOMER')", ticketID).
		Order("created_at ASC").
		Find(&comments).Error
	return comments, err
}
