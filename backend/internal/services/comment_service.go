package services

import (
	"errors"
	"net/http"

	"github.com/ayush/supportiq/internal/dto"
	"github.com/ayush/supportiq/internal/models"
	"github.com/ayush/supportiq/internal/repositories"
	"github.com/google/uuid"
)

// CommentService handles business logic for ticket comments.
type CommentService struct {
	commentRepo  *repositories.CommentRepository
	activityRepo *repositories.ActivityRepository
	ticketRepo   *repositories.TicketRepository
}

func NewCommentService(commentRepo *repositories.CommentRepository, activityRepo *repositories.ActivityRepository) *CommentService {
	return &CommentService{commentRepo: commentRepo, activityRepo: activityRepo}
}

func (s *CommentService) SetTicketRepo(r *repositories.TicketRepository) {
	s.ticketRepo = r
}

func (s *CommentService) Create(tenantID uuid.UUID, ticketID uuid.UUID, req *dto.CreateCommentRequest, userID uint) (*dto.CommentResponse, int, error) {
	// Verify the ticket exists and belongs to this tenant
	if s.ticketRepo != nil {
		if _, err := s.ticketRepo.FindByID(tenantID, ticketID); err != nil {
			return nil, http.StatusNotFound, errors.New("ticket not found")
		}
	}
	commentType := models.CommentTypePublic
	if req.CommentType == "INTERNAL" {
		commentType = models.CommentTypeInternal
	}

	comment := &models.TicketComment{
		TenantID:    tenantID,
		TicketID:    ticketID,
		UserID:      userID,
		Message:     req.Message,
		CommentType: commentType,
	}
	if err := s.commentRepo.Create(comment); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	loaded, err := s.commentRepo.FindByID(tenantID, comment.ID)
	if err != nil {
		r := toCommentResponse(comment)
		return &r, http.StatusCreated, nil
	}

	_ = s.activityRepo.Create(&models.TicketActivity{
		TenantID:     tenantID,
		TicketID:     ticketID,
		UserID:       userID,
		ActivityType: models.ActivityCommentAdded,
		Description:  "Comment added",
	})

	r := toCommentResponse(loaded)
	return &r, http.StatusCreated, nil
}

func (s *CommentService) List(tenantID uuid.UUID, ticketID uuid.UUID) ([]dto.CommentResponse, int, error) {
	comments, err := s.commentRepo.ListByTicketID(tenantID, ticketID)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	responses := make([]dto.CommentResponse, len(comments))
	for i := range comments {
		responses[i] = toCommentResponse(&comments[i])
	}
	return responses, http.StatusOK, nil
}

func toCommentResponse(c *models.TicketComment) dto.CommentResponse {
	resp := dto.CommentResponse{
		ID:          c.ID,
		TicketID:    c.TicketID,
		Message:     c.Message,
		CommentType: string(c.CommentType),
		CreatedAt:   c.CreatedAt,
	}
	if c.User != nil {
		ur := dto.UserResponse{ID: c.User.ID, Name: c.User.Name, Email: c.User.Email, Role: string(c.User.Role)}
		resp.User = &ur
	}
	return resp
}
