package services

import (
	"errors"
	"fmt"
	"math"
	"net/http"

	"github.com/ayush/supportiq/internal/dto"
	"github.com/ayush/supportiq/internal/models"
	"github.com/ayush/supportiq/internal/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrTicketNotFound    = errors.New("ticket not found")
	ErrInvalidTransition = errors.New("invalid status transition")
	ErrForbidden         = errors.New("insufficient permissions")
	ErrAssigneeNotFound  = errors.New("assignee not found")
	ErrAssigneeNotAgent  = errors.New("assignee must have SupportAgent role")
	ErrAlreadyAssigned   = errors.New("ticket is already assigned")
)

// TicketService contains all business logic for ticket management.
type TicketService struct {
	repo         *repositories.TicketRepository
	userRepo     *repositories.UserRepository
	activityRepo *repositories.ActivityRepository
	aiService    *AIService
	jobSvc       *JobService
	slaSvc       *SLAService
}

func NewTicketService(
	repo *repositories.TicketRepository,
	userRepo *repositories.UserRepository,
	activityRepo *repositories.ActivityRepository,
	aiService *AIService,
) *TicketService {
	return &TicketService{repo: repo, userRepo: userRepo, activityRepo: activityRepo, aiService: aiService}
}

func (s *TicketService) SetJobService(js *JobService) {
	s.jobSvc = js
}

func (s *TicketService) SetSLAService(sla *SLAService) {
	s.slaSvc = sla
}

// logActivity is fire-and-forget; errors are silently suppressed.
func (s *TicketService) logActivity(tenantID uuid.UUID, ticketID uuid.UUID, userID uint, actType, oldVal, newVal, desc string) {
	_ = s.activityRepo.Create(&models.TicketActivity{
		TenantID:     tenantID,
		TicketID:     ticketID,
		UserID:       userID,
		ActivityType: actType,
		OldValue:     oldVal,
		NewValue:     newVal,
		Description:  desc,
	})
}

// Create inserts a new ticket, generating its sequential number inside a transaction.
func (s *TicketService) Create(tenantID uuid.UUID, req *dto.CreateTicketRequest, createdBy uint) (*dto.TicketResponse, int, error) {
	var created models.Ticket

	err := s.repo.DB().Transaction(func(tx *gorm.DB) error {
		ticketNum, err := s.repo.NextTicketNumber(tenantID, tx)
		if err != nil {
			return err
		}
		priority := models.TicketPriorityMedium
		if req.Priority != "" {
			priority = models.TicketPriority(req.Priority)
		}
		category := models.TicketCategoryGeneral
		if req.Category != "" {
			category = models.TicketCategory(req.Category)
		}
		created = models.Ticket{
			TenantID:      tenantID,
			Subject:       req.Subject,
			Description:   req.Description,
			CustomerName:  req.CustomerName,
			CustomerEmail: req.CustomerEmail,
			TicketNumber:  ticketNum,
			Status:        models.TicketStatusOpen,
			Priority:      priority,
			Category:      category,
			Source:        models.TicketSourceWeb,
			CreatedBy:     createdBy,
		}
		return s.repo.Create(tx, &created)
	})
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	s.logActivity(tenantID, created.ID, createdBy, models.ActivityCreateTicket, "", "", "Ticket created")

	// Assign SLA deadlines asynchronously (no-op if no policy configured)
	if s.slaSvc != nil {
		s.slaSvc.AssignSLAToTicket(tenantID, &created)
	}

	if s.jobSvc != nil && s.jobSvc.IsQueueAvailable() {
		_ = s.jobSvc.EnqueueAIAnalysis(created.ID, createdBy)
	} else {
		s.aiService.AnalyzeTicket(created.ID)
	}

	full, err := s.repo.FindByID(tenantID, created.ID)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return toTicketResponse(full), http.StatusCreated, nil
}

// List returns a paginated, filtered set of tickets for a tenant.
func (s *TicketService) List(tenantID uuid.UUID, q *dto.ListTicketsQuery) (*dto.ListTicketsResponse, int, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 || q.Limit > 100 {
		q.Limit = 20
	}

	tickets, total, err := s.repo.List(tenantID, q)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	items := make([]dto.TicketResponse, len(tickets))
	for i := range tickets {
		items[i] = *toTicketResponse(&tickets[i])
	}

	totalPages := int(math.Ceil(float64(total) / float64(q.Limit)))
	if totalPages == 0 {
		totalPages = 1
	}

	return &dto.ListTicketsResponse{
		Items:       items,
		TotalCount:  total,
		CurrentPage: q.Page,
		TotalPages:  totalPages,
		Limit:       q.Limit,
	}, http.StatusOK, nil
}

// GetByID returns a single ticket by UUID, scoped to tenant.
func (s *TicketService) GetByID(tenantID uuid.UUID, id uuid.UUID) (*dto.TicketResponse, int, error) {
	t, err := s.repo.FindByID(tenantID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, http.StatusNotFound, ErrTicketNotFound
		}
		return nil, http.StatusInternalServerError, err
	}
	return toTicketResponse(t), http.StatusOK, nil
}

// Update applies editable field changes.
func (s *TicketService) Update(tenantID uuid.UUID, id uuid.UUID, req *dto.UpdateTicketRequest, userID uint) (*dto.TicketResponse, int, error) {
	t, err := s.repo.FindByID(tenantID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, http.StatusNotFound, ErrTicketNotFound
		}
		return nil, http.StatusInternalServerError, err
	}

	oldPriority := string(t.Priority)
	oldCategory := string(t.Category)

	if req.Subject != "" {
		t.Subject = req.Subject
	}
	if req.Description != "" {
		t.Description = req.Description
	}
	if req.Priority != "" {
		t.Priority = models.TicketPriority(req.Priority)
	}
	if req.Category != "" {
		t.Category = models.TicketCategory(req.Category)
	}
	if req.CustomerName != "" {
		t.CustomerName = req.CustomerName
	}
	if req.CustomerEmail != "" {
		t.CustomerEmail = req.CustomerEmail
	}

	if err := s.repo.Update(t); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	if req.Priority != "" && string(t.Priority) != oldPriority {
		s.logActivity(tenantID, id, userID, models.ActivityPriorityChanged, oldPriority, string(t.Priority),
			fmt.Sprintf("Priority changed from %s to %s", oldPriority, string(t.Priority)))
	}
	if req.Category != "" && string(t.Category) != oldCategory {
		s.logActivity(tenantID, id, userID, models.ActivityCategoryChanged, oldCategory, string(t.Category),
			fmt.Sprintf("Category changed from %s to %s", oldCategory, string(t.Category)))
	}
	if req.Subject != "" || req.Description != "" || req.CustomerName != "" || req.CustomerEmail != "" {
		s.logActivity(tenantID, id, userID, models.ActivityUpdateTicket, "", "", "Ticket details updated")
	}

	if req.Subject != "" || req.Description != "" {
		if s.jobSvc != nil && s.jobSvc.IsQueueAvailable() {
			_ = s.jobSvc.EnqueueAIAnalysis(id, userID)
		} else {
			s.aiService.AnalyzeTicket(id)
		}
	}

	return toTicketResponse(t), http.StatusOK, nil
}

// UpdateStatus transitions the ticket to a new status.
func (s *TicketService) UpdateStatus(tenantID uuid.UUID, id uuid.UUID, req *dto.UpdateStatusRequest, userID uint) (*dto.TicketResponse, int, error) {
	t, err := s.repo.FindByID(tenantID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, http.StatusNotFound, ErrTicketNotFound
		}
		return nil, http.StatusInternalServerError, err
	}

	oldStatus := string(t.Status)
	newStatus := models.TicketStatus(req.Status)
	if !models.IsValidStatusTransition(t.Status, newStatus) {
		return nil, http.StatusBadRequest, ErrInvalidTransition
	}

	t.Status = newStatus
	if err := s.repo.Update(t); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	// SLA hooks
	if s.slaSvc != nil {
		switch newStatus {
		case models.TicketStatusInProgress:
			s.slaSvc.MarkFirstResponse(tenantID, t)
		case models.TicketStatusResolved:
			s.slaSvc.MarkResolved(tenantID, t)
		}
	}

	actType := models.ActivityStatusChanged
	if newStatus == models.TicketStatusClosed {
		actType = models.ActivityTicketClosed
	}
	s.logActivity(tenantID, id, userID, actType, oldStatus, string(newStatus),
		fmt.Sprintf("Status changed from %s to %s", oldStatus, string(newStatus)))

	return toTicketResponse(t), http.StatusOK, nil
}

// Assign sets the assigned_to field.
func (s *TicketService) Assign(tenantID uuid.UUID, id uuid.UUID, req *dto.AssignTicketRequest, userID uint, userRole string) (*dto.TicketResponse, int, error) {
	if models.Role(userRole) != models.RoleAdmin {
		return nil, http.StatusForbidden, ErrForbidden
	}

	t, err := s.repo.FindByID(tenantID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, http.StatusNotFound, ErrTicketNotFound
		}
		return nil, http.StatusInternalServerError, err
	}

	assignee, err := s.userRepo.FindByID(req.AssignedTo)
	if err != nil {
		return nil, http.StatusNotFound, ErrAssigneeNotFound
	}
	if assignee.TenantID != tenantID {
		return nil, http.StatusForbidden, errors.New("assignee does not belong to this tenant")
	}
	if assignee.Role != models.RoleSupportAgent {
		return nil, http.StatusBadRequest, ErrAssigneeNotAgent
	}

	t.AssignedTo = &req.AssignedTo
	if err := s.repo.Update(t); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	s.logActivity(tenantID, id, userID, models.ActivityAssignTicket, "", assignee.Name,
		fmt.Sprintf("Assigned to %s", assignee.Name))

	return toTicketResponse(t), http.StatusOK, nil
}

// Delete soft-deletes a ticket.
func (s *TicketService) Delete(tenantID uuid.UUID, id uuid.UUID, userRole string) (int, error) {
	if models.Role(userRole) != models.RoleAdmin {
		return http.StatusForbidden, ErrForbidden
	}

	if _, err := s.repo.FindByID(tenantID, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return http.StatusNotFound, ErrTicketNotFound
		}
		return http.StatusInternalServerError, err
	}

	if err := s.repo.SoftDelete(tenantID, id); err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}

// TakeOwnership lets a SupportAgent claim an unassigned ticket.
func (s *TicketService) TakeOwnership(tenantID uuid.UUID, id uuid.UUID, userID uint, userRole string) (*dto.TicketResponse, int, error) {
	if models.Role(userRole) != models.RoleSupportAgent {
		return nil, http.StatusForbidden, ErrForbidden
	}

	t, err := s.repo.FindByID(tenantID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, http.StatusNotFound, ErrTicketNotFound
		}
		return nil, http.StatusInternalServerError, err
	}

	if t.AssignedTo != nil {
		return nil, http.StatusConflict, ErrAlreadyAssigned
	}

	t.AssignedTo = &userID
	if err := s.repo.Update(t); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	s.logActivity(tenantID, id, userID, models.ActivityTakeOwnership, "", "", "Took ownership of ticket")

	full, _ := s.repo.FindByID(tenantID, id)
	if full == nil {
		return toTicketResponse(t), http.StatusOK, nil
	}
	return toTicketResponse(full), http.StatusOK, nil
}

// MyTickets returns tickets assigned to the given user within a tenant.
func (s *TicketService) MyTickets(tenantID uuid.UUID, userID uint, q *dto.ListTicketsQuery) (*dto.ListTicketsResponse, int, error) {
	q.AssignedTo = &userID
	q.UnassignedOnly = false
	return s.List(tenantID, q)
}

// TeamTickets returns tickets for a shared team account:
// tickets explicitly assigned to this user OR ai_team matches the user's team name.
func (s *TicketService) TeamTickets(tenantID uuid.UUID, userID uint, teamName string, q *dto.ListTicketsQuery) (*dto.ListTicketsResponse, int, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 {
		q.Limit = 50
	}
	tickets, total, err := s.repo.FindTeamTickets(tenantID, userID, teamName, q)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	items := make([]dto.TicketResponse, 0, len(tickets))
	for _, t := range tickets {
		items = append(items, *toTicketResponse(&t))
	}
	totalPages := int(total) / q.Limit
	if int(total)%q.Limit != 0 {
		totalPages++
	}
	return &dto.ListTicketsResponse{
		Items:       items,
		TotalCount:  total,
		CurrentPage: q.Page,
		TotalPages:  totalPages,
		Limit:       q.Limit,
	}, http.StatusOK, nil
}

// ListUnassigned returns unassigned tickets within a tenant.
func (s *TicketService) ListUnassigned(tenantID uuid.UUID, q *dto.ListTicketsQuery) (*dto.ListTicketsResponse, int, error) {
	q.UnassignedOnly = true
	q.AssignedTo = nil
	return s.List(tenantID, q)
}

func toTicketResponse(t *models.Ticket) *dto.TicketResponse {
	resp := &dto.TicketResponse{
		ID:                 t.ID,
		TicketNumber:       t.TicketNumber,
		Subject:            t.Subject,
		Description:        t.Description,
		Status:             string(t.Status),
		Priority:           string(t.Priority),
		Category:           string(t.Category),
		Source:             string(t.Source),
		AssignedTo:         t.AssignedTo,
		CreatedBy:          t.CreatedBy,
		CustomerName:       t.CustomerName,
		CustomerEmail:      t.CustomerEmail,
		CreatedAt:          t.CreatedAt,
		UpdatedAt:          t.UpdatedAt,
		AICategory:         t.AICategory,
		AIPriority:         t.AIPriority,
		AISentiment:        t.AISentiment,
		AITeam:             t.AITeam,
		AIConfidence:       t.AIConfidence,
		AISummary:          t.AISummary,
		AITags:             t.AITags,
		AIProcessingStatus: t.AIProcessingStatus,
		ProcessedAt:        t.ProcessedAt,
		// SLA fields
		SLAPolicyID:              t.SLAPolicyID,
		FirstResponseDueAt:       t.FirstResponseDueAt,
		ResolutionDueAt:          t.ResolutionDueAt,
		FirstResponseCompletedAt: t.FirstResponseCompletedAt,
		ResolvedAt:               t.ResolvedAt,
		SLAStatus:                t.SLAStatus,
	}

	if t.Creator != nil {
		ur := dto.UserResponse{ID: t.Creator.ID, Name: t.Creator.Name, Email: t.Creator.Email, Role: string(t.Creator.Role)}
		resp.Creator = &ur
	}
	if t.Assignee != nil {
		ur := dto.UserResponse{ID: t.Assignee.ID, Name: t.Assignee.Name, Email: t.Assignee.Email, Role: string(t.Assignee.Role)}
		resp.Assignee = &ur
	}
	return resp
}
