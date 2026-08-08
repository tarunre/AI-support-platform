package services

import (
	"errors"
	"net/http"

	"github.com/ayush/supportiq/internal/dto"
	"github.com/ayush/supportiq/internal/models"
	"github.com/ayush/supportiq/internal/repositories"
	"github.com/google/uuid"
)

// NoteService handles business logic for internal ticket notes.
type NoteService struct {
	noteRepo     *repositories.NoteRepository
	activityRepo *repositories.ActivityRepository
	ticketRepo   *repositories.TicketRepository
}

func NewNoteService(noteRepo *repositories.NoteRepository, activityRepo *repositories.ActivityRepository) *NoteService {
	return &NoteService{noteRepo: noteRepo, activityRepo: activityRepo}
}

func (s *NoteService) SetTicketRepo(r *repositories.TicketRepository) {
	s.noteRepo = s.noteRepo
	s.ticketRepo = r
}

func (s *NoteService) Create(tenantID uuid.UUID, ticketID uuid.UUID, req *dto.CreateNoteRequest, userID uint) (*dto.NoteResponse, int, error) {
	// Verify the ticket exists and belongs to this tenant
	if s.ticketRepo != nil {
		if _, err := s.ticketRepo.FindByID(tenantID, ticketID); err != nil {
			return nil, http.StatusNotFound, errors.New("ticket not found")
		}
	}
	note := &models.TicketNote{
		TenantID:   tenantID,
		TicketID:   ticketID,
		UserID:     userID,
		Note:       req.Note,
		IsInternal: true,
	}
	if err := s.noteRepo.Create(note); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	loaded, err := s.noteRepo.FindByID(tenantID, note.ID)
	if err != nil {
		r := toNoteResponse(note)
		return &r, http.StatusCreated, nil
	}

	_ = s.activityRepo.Create(&models.TicketActivity{
		TenantID:     tenantID,
		TicketID:     ticketID,
		UserID:       userID,
		ActivityType: models.ActivityInternalNoteAdded,
		Description:  "Internal note added",
	})

	r := toNoteResponse(loaded)
	return &r, http.StatusCreated, nil
}

func (s *NoteService) List(tenantID uuid.UUID, ticketID uuid.UUID) ([]dto.NoteResponse, int, error) {
	notes, err := s.noteRepo.ListByTicketID(tenantID, ticketID)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	responses := make([]dto.NoteResponse, len(notes))
	for i := range notes {
		responses[i] = toNoteResponse(&notes[i])
	}
	return responses, http.StatusOK, nil
}

func toNoteResponse(n *models.TicketNote) dto.NoteResponse {
	resp := dto.NoteResponse{
		ID:         n.ID,
		TicketID:   n.TicketID,
		Note:       n.Note,
		IsInternal: n.IsInternal,
		CreatedAt:  n.CreatedAt,
	}
	if n.User != nil {
		ur := dto.UserResponse{ID: n.User.ID, Name: n.User.Name, Email: n.User.Email, Role: string(n.User.Role)}
		resp.User = &ur
	}
	return resp
}
