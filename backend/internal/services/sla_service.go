package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ayush/supportiq/internal/dto"
	"github.com/ayush/supportiq/internal/events"
	"github.com/ayush/supportiq/internal/models"
	"github.com/ayush/supportiq/internal/repositories"
	"github.com/ayush/supportiq/internal/utils"
	appws "github.com/ayush/supportiq/internal/websocket"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SLAService manages SLA policies and monitors ticket SLA compliance.
type SLAService struct {
	slaRepo      *repositories.SLARepository
	ticketRepo   *repositories.TicketRepository
	activityRepo *repositories.ActivityRepository
	hub          *appws.Hub
}

func NewSLAService(
	slaRepo *repositories.SLARepository,
	ticketRepo *repositories.TicketRepository,
	activityRepo *repositories.ActivityRepository,
	hub *appws.Hub,
) *SLAService {
	return &SLAService{
		slaRepo:      slaRepo,
		ticketRepo:   ticketRepo,
		activityRepo: activityRepo,
		hub:          hub,
	}
}

// ─── Policy CRUD ──────────────────────────────────────────────────────────────

func (s *SLAService) ListPolicies(tenantID uuid.UUID) ([]dto.SLAPolicyResponse, int, error) {
	policies, err := s.slaRepo.List(tenantID)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	resp := make([]dto.SLAPolicyResponse, len(policies))
	for i, p := range policies {
		resp[i] = toSLAPolicyResponse(&p)
	}
	return resp, http.StatusOK, nil
}

func (s *SLAService) CreatePolicy(tenantID uuid.UUID, req *dto.CreateSLAPolicyRequest) (*dto.SLAPolicyResponse, int, error) {
	if req.IsDefault {
		if err := s.slaRepo.UnsetAllDefaults(tenantID); err != nil {
			return nil, http.StatusInternalServerError, err
		}
	}
	p := &models.SLAPolicy{
		TenantID:             tenantID,
		Name:                 req.Name,
		Priority:             models.TicketPriority(req.Priority),
		FirstResponseMinutes: req.FirstResponseMinutes,
		ResolutionMinutes:    req.ResolutionMinutes,
		IsDefault:            req.IsDefault,
	}
	if err := s.slaRepo.Create(p); err != nil {
		return nil, http.StatusInternalServerError, err
	}
	r := toSLAPolicyResponse(p)
	return &r, http.StatusCreated, nil
}

func (s *SLAService) GetPolicy(tenantID uuid.UUID, id uint) (*dto.SLAPolicyResponse, int, error) {
	p, err := s.slaRepo.FindByID(tenantID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, http.StatusNotFound, fmt.Errorf("SLA policy not found")
		}
		return nil, http.StatusInternalServerError, err
	}
	r := toSLAPolicyResponse(p)
	return &r, http.StatusOK, nil
}

func (s *SLAService) UpdatePolicy(tenantID uuid.UUID, id uint, req *dto.UpdateSLAPolicyRequest) (*dto.SLAPolicyResponse, int, error) {
	p, err := s.slaRepo.FindByID(tenantID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, http.StatusNotFound, fmt.Errorf("SLA policy not found")
		}
		return nil, http.StatusInternalServerError, err
	}
	if req.Name != "" {
		p.Name = req.Name
	}
	if req.Priority != "" {
		p.Priority = models.TicketPriority(req.Priority)
	}
	if req.FirstResponseMinutes > 0 {
		p.FirstResponseMinutes = req.FirstResponseMinutes
	}
	if req.ResolutionMinutes > 0 {
		p.ResolutionMinutes = req.ResolutionMinutes
	}
	if req.IsDefault != nil {
		if *req.IsDefault {
			if err := s.slaRepo.UnsetAllDefaults(tenantID); err != nil {
				return nil, http.StatusInternalServerError, err
			}
		}
		p.IsDefault = *req.IsDefault
	}
	if err := s.slaRepo.Update(p); err != nil {
		return nil, http.StatusInternalServerError, err
	}
	r := toSLAPolicyResponse(p)
	return &r, http.StatusOK, nil
}

func (s *SLAService) DeletePolicy(tenantID uuid.UUID, id uint) (int, error) {
	if _, err := s.slaRepo.FindByID(tenantID, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return http.StatusNotFound, fmt.Errorf("SLA policy not found")
		}
		return http.StatusInternalServerError, err
	}
	if err := s.slaRepo.Delete(tenantID, id); err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}

// ─── SLA assignment ───────────────────────────────────────────────────────────

// AssignSLAToTicket looks up the best matching SLA policy for the ticket's
// priority and sets the deadline fields. Silently no-ops if no policy exists.
func (s *SLAService) AssignSLAToTicket(tenantID uuid.UUID, ticket *models.Ticket) {
	policy, err := s.slaRepo.FindByPriority(tenantID, ticket.Priority)
	if err != nil {
		// Fallback: try the tenant's default policy
		policy, err = s.slaRepo.FindDefault(tenantID)
		if err != nil {
			return // No SLA configured — skip silently
		}
	}

	now := ticket.CreatedAt
	if now.IsZero() {
		now = time.Now()
	}
	firstResponseDue := now.Add(time.Duration(policy.FirstResponseMinutes) * time.Minute)
	resolutionDue := now.Add(time.Duration(policy.ResolutionMinutes) * time.Minute)

	ticket.SLAPolicyID = &policy.ID
	ticket.FirstResponseDueAt = &firstResponseDue
	ticket.ResolutionDueAt = &resolutionDue
	ticket.SLAStatus = string(models.SLAStatusOnTrack)

	if err := s.ticketRepo.UpdateSLAFields(ticket); err != nil {
		utils.Logger.WithError(err).Warn("SLA: failed to assign policy to ticket")
		return
	}

	s.logActivity(tenantID, ticket.ID, ticket.CreatedBy,
		models.ActivitySLAAssigned, "", policy.Name,
		fmt.Sprintf("SLA policy '%s' assigned — resolution due %s",
			policy.Name, resolutionDue.Format("02 Jan 2006 15:04 MST")))
}

// MarkFirstResponse records when the first agent response occurred.
// Called when a ticket transitions to IN_PROGRESS status.
func (s *SLAService) MarkFirstResponse(tenantID uuid.UUID, ticket *models.Ticket) {
	if ticket.FirstResponseDueAt == nil || ticket.FirstResponseCompletedAt != nil {
		return
	}
	now := time.Now()
	ticket.FirstResponseCompletedAt = &now

	if err := s.ticketRepo.UpdateSLAFields(ticket); err != nil {
		utils.Logger.WithError(err).Warn("SLA: failed to mark first response")
		return
	}

	s.logActivity(tenantID, ticket.ID, 0, models.ActivitySLACompleted,
		"", "first_response",
		fmt.Sprintf("First response completed in %.0f minutes",
			now.Sub(ticket.CreatedAt).Minutes()))
}

// MarkResolved marks the SLA as COMPLETED or BREACHED when a ticket is resolved.
func (s *SLAService) MarkResolved(tenantID uuid.UUID, ticket *models.Ticket) {
	if ticket.ResolutionDueAt == nil {
		return
	}
	now := time.Now()
	ticket.ResolvedAt = &now

	if now.Before(*ticket.ResolutionDueAt) {
		ticket.SLAStatus = string(models.SLAStatusCompleted)
		s.logActivity(tenantID, ticket.ID, 0, models.ActivitySLACompleted,
			"", string(models.SLAStatusCompleted),
			fmt.Sprintf("Ticket resolved within SLA — %.0f minutes remaining",
				ticket.ResolutionDueAt.Sub(now).Minutes()))
	} else {
		ticket.SLAStatus = string(models.SLAStatusBreached)
		s.logActivity(tenantID, ticket.ID, 0, models.ActivitySLABreached,
			"", string(models.SLAStatusBreached),
			"Ticket resolved after SLA breach")
	}

	if err := s.ticketRepo.UpdateSLAFields(ticket); err != nil {
		utils.Logger.WithError(err).Warn("SLA: failed to mark resolved")
	}
	s.broadcastSLAEvent(ticket)
}

// ─── Dashboard ────────────────────────────────────────────────────────────────

func (s *SLAService) GetDashboard(tenantID uuid.UUID) (*dto.SLADashboardResponse, int, error) {
	nearBreach, err := repositories.FindNearBreach(s.ticketRepo.DB(), tenantID, 120)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	breached, err := repositories.FindBreachedOpen(s.ticketRepo.DB(), tenantID)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	stats, err := repositories.SLAStats(s.ticketRepo.DB(), tenantID)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	compliance := 0.0
	if stats.Total > 0 {
		compliance = float64(stats.Total-stats.Breached) / float64(stats.Total) * 100
	}

	return &dto.SLADashboardResponse{
		NearBreach:          toTicketSLASummaries(nearBreach),
		Breached:            toTicketSLASummaries(breached),
		AvgFirstResponseMin: stats.AvgFirstResponseMin,
		AvgResolutionMin:    stats.AvgResolutionMin,
		CompliancePercent:   compliance,
		TotalWithSLA:        stats.Total,
		BreachedCount:       stats.Breached,
		CompletedOnTime:     stats.CompletedOnTime,
	}, http.StatusOK, nil
}

// ─── SLA Monitor ──────────────────────────────────────────────────────────────

// StartMonitor runs the SLA monitoring loop in the calling goroutine.
// It blocks until ctx is cancelled.
func (s *SLAService) StartMonitor(ctx context.Context, interval time.Duration) {
	utils.Logger.WithField("interval", interval).Info("SLAMonitor: started")
	s.runCheck()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			utils.Logger.Info("SLAMonitor: stopped")
			return
		case <-ticker.C:
			s.runCheck()
		}
	}
}

func (s *SLAService) runCheck() {
	tickets, err := s.ticketRepo.FindAllOpenWithSLA()
	if err != nil {
		utils.Logger.WithError(err).Warn("SLAMonitor: failed to fetch tickets")
		return
	}
	if len(tickets) == 0 {
		return
	}

	now := time.Now()
	for i := range tickets {
		s.evaluateTicket(&tickets[i], now)
	}
}

func (s *SLAService) evaluateTicket(t *models.Ticket, now time.Time) {
	if t.ResolutionDueAt == nil {
		return
	}

	total := t.ResolutionDueAt.Sub(t.CreatedAt).Minutes()
	if total <= 0 {
		return
	}
	pct := now.Sub(t.CreatedAt).Minutes() / total * 100

	old := models.SLAStatus(t.SLAStatus)
	var next models.SLAStatus

	switch {
	case pct >= float64(models.SLAThresholdBreach):
		next = models.SLAStatusBreached
	case pct >= float64(models.SLAThresholdAtRisk):
		next = models.SLAStatusAtRisk
	default:
		next = models.SLAStatusOnTrack
	}

	// Log threshold crossing activities
	if next == models.SLAStatusAtRisk && old == models.SLAStatusOnTrack {
		s.logActivity(t.TenantID, t.ID, 0, models.ActivitySLAAtRisk,
			string(old), string(next),
			fmt.Sprintf("SLA at risk — %.0f%% of resolution time elapsed, notifying assigned agent", pct))
	}

	// 90% escalation — only log once per ticket
	if pct >= float64(models.SLAThresholdEscalate) &&
		!s.hasActivityOfType(t.TenantID, t.ID, models.ActivitySLAEscalated) {
		s.logActivity(t.TenantID, t.ID, 0, models.ActivitySLAEscalated,
			"", "",
			fmt.Sprintf("SLA escalated — %.0f%% of resolution time elapsed, notifying team lead", pct))
	}

	if next == models.SLAStatusBreached && old != models.SLAStatusBreached {
		s.logActivity(t.TenantID, t.ID, 0, models.ActivitySLABreached,
			string(old), string(models.SLAStatusBreached),
			fmt.Sprintf("SLA breached — ticket overdue by %.0f minutes",
				now.Sub(*t.ResolutionDueAt).Minutes()))
	}

	if next == old {
		return
	}

	t.SLAStatus = string(next)
	if err := s.ticketRepo.UpdateSLAFields(t); err != nil {
		utils.Logger.WithError(err).WithField("ticket_id", t.ID).
			Warn("SLAMonitor: failed to update SLA status")
		return
	}
	s.broadcastSLAEvent(t)
}

func (s *SLAService) hasActivityOfType(tenantID, ticketID uuid.UUID, actType string) bool {
	var count int64
	s.activityRepo.DB().Model(&models.TicketActivity{}).
		Where("tenant_id = ? AND ticket_id = ? AND activity_type = ?", tenantID, ticketID, actType).
		Count(&count)
	return count > 0
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func (s *SLAService) logActivity(tenantID, ticketID uuid.UUID, userID uint, actType, oldVal, newVal, desc string) {
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

func (s *SLAService) broadcastSLAEvent(t *models.Ticket) {
	if s.hub == nil {
		return
	}
	s.hub.BroadcastToTenant(t.TenantID, events.New(events.SLAUpdated, t.ID.String(), 0, "", map[string]interface{}{
		"sla_status":        t.SLAStatus,
		"resolution_due_at": t.ResolutionDueAt,
		"ticket_number":     t.TicketNumber,
	}))
}

func toSLAPolicyResponse(p *models.SLAPolicy) dto.SLAPolicyResponse {
	return dto.SLAPolicyResponse{
		ID:                   p.ID,
		TenantID:             p.TenantID,
		Name:                 p.Name,
		Priority:             string(p.Priority),
		FirstResponseMinutes: p.FirstResponseMinutes,
		ResolutionMinutes:    p.ResolutionMinutes,
		IsDefault:            p.IsDefault,
		CreatedAt:            p.CreatedAt,
		UpdatedAt:            p.UpdatedAt,
	}
}

func toTicketSLASummaries(tickets []models.Ticket) []dto.TicketSLASummary {
	now := time.Now()
	summaries := make([]dto.TicketSLASummary, len(tickets))
	for i, t := range tickets {
		var remainingMin, pctElapsed float64
		if t.ResolutionDueAt != nil {
			remainingMin = t.ResolutionDueAt.Sub(now).Minutes()
			total := t.ResolutionDueAt.Sub(t.CreatedAt).Minutes()
			if total > 0 {
				pctElapsed = now.Sub(t.CreatedAt).Minutes() / total * 100
			}
		}
		summaries[i] = dto.TicketSLASummary{
			TicketID:         t.ID,
			TicketNumber:     t.TicketNumber,
			Subject:          t.Subject,
			Priority:         string(t.Priority),
			SLAStatus:        t.SLAStatus,
			ResolutionDueAt:  t.ResolutionDueAt,
			TimeRemainingMin: remainingMin,
			PercentElapsed:   pctElapsed,
			AssignedTo:       t.AssignedTo,
		}
	}
	return summaries
}
