package repositories

import (
	"github.com/ayush/supportiq/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NoteRepository handles all database access for ticket notes.
type NoteRepository struct {
	db *gorm.DB
}

func NewNoteRepository(db *gorm.DB) *NoteRepository {
	return &NoteRepository{db: db}
}

func (r *NoteRepository) Create(note *models.TicketNote) error {
	return r.db.Create(note).Error
}

func (r *NoteRepository) FindByID(tenantID uuid.UUID, id uint) (*models.TicketNote, error) {
	var note models.TicketNote
	if err := r.db.Preload("User").Where("tenant_id = ? AND id = ?", tenantID, id).First(&note).Error; err != nil {
		return nil, err
	}
	return &note, nil
}

func (r *NoteRepository) ListByTicketID(tenantID uuid.UUID, ticketID uuid.UUID) ([]models.TicketNote, error) {
	var notes []models.TicketNote
	err := r.db.
		Preload("User").
		Where("tenant_id = ? AND ticket_id = ?", tenantID, ticketID).
		Order("created_at DESC").
		Find(&notes).Error
	return notes, err
}
