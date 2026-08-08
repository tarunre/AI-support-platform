package services

import (
	"context"
	"fmt"
	"time"

	"github.com/ayush/supportiq/internal/ai/provider"
	"github.com/ayush/supportiq/internal/models"
	"github.com/ayush/supportiq/internal/repositories"
	"github.com/ayush/supportiq/internal/utils"
	"github.com/google/uuid"
)

// AIService orchestrates AI ticket analysis.
// Analysis always runs in a background goroutine so ticket creation is
// never blocked by the AI provider.
type AIService struct {
	provider     provider.Provider
	ticketRepo   *repositories.TicketRepository
	userRepo     *repositories.UserRepository
	activityRepo *repositories.ActivityRepository
	replySvc     *ReplyService // optional; set via SetReplyService after construction
}

func NewAIService(
	p provider.Provider,
	ticketRepo *repositories.TicketRepository,
	activityRepo *repositories.ActivityRepository,
) *AIService {
	return &AIService{provider: p, ticketRepo: ticketRepo, activityRepo: activityRepo}
}

func (s *AIService) SetUserRepo(ur *repositories.UserRepository) { s.userRepo = ur }

// SetReplyService injects the reply service so AI analysis can trigger automatic
// reply generation on completion. Called after both services are constructed
// to avoid circular dependency in the DI wiring.
func (s *AIService) SetReplyService(rs *ReplyService) {
	s.replySvc = rs
}

// AnalyzeTicket queues an async analysis for a newly created ticket.
func (s *AIService) AnalyzeTicket(ticketID uuid.UUID) {
	go s.run(ticketID)
}

// RetryAnalysis re-queues analysis for a ticket whose previous attempt failed.
func (s *AIService) RetryAnalysis(ticketID uuid.UUID) {
	go s.run(ticketID)
}

func (s *AIService) run(ticketID uuid.UUID) {
	log := utils.Logger.WithField("ticket_id", ticketID)

	ticket, err := s.ticketRepo.FindByIDUnscoped(ticketID)
	if err != nil {
		log.WithError(err).Error("AI: Could not load ticket for analysis")
		return
	}

	log = log.WithField("ticket_number", ticket.TicketNumber)
	log.Info("AI: Starting analysis")

	// Mark the ticket as in-progress so the UI shows the spinner
	ticket.AIProcessingStatus = models.AIStatusProcessing
	if err := s.ticketRepo.UpdateAIFields(ticket); err != nil {
		log.WithError(err).Warn("AI: Failed to set PROCESSING status")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := s.provider.Analyze(ctx, provider.AnalysisRequest{
		Subject:      ticket.Subject,
		Description:  ticket.Description,
		CustomerName: ticket.CustomerName,
		Category:     string(ticket.Category),
		Priority:     string(ticket.Priority),
	})

	if err != nil {
		log.WithError(err).Error("AI: Analysis failed")
		ticket.AIProcessingStatus = models.AIStatusFailed
		_ = s.ticketRepo.UpdateAIFields(ticket)
		return
	}

	// Persist the result
	now := time.Now()
	ticket.AICategory = result.Category
	ticket.AIPriority = result.Priority
	ticket.AISentiment = result.Sentiment
	ticket.AITeam = result.RecommendedTeam
	ticket.AIConfidence = &result.Confidence
	ticket.AISummary = result.Summary
	ticket.AITags = result.Tags
	ticket.AIProcessingStatus = models.AIStatusCompleted
	ticket.ProcessedAt = &now

	if err := s.ticketRepo.UpdateAIFields(ticket); err != nil {
		log.WithError(err).Error("AI: Failed to persist analysis result")
		return
	}

	// Append to the activity timeline
	_ = s.activityRepo.Create(&models.TicketActivity{
		TenantID:     ticket.TenantID,
		TicketID:     ticketID,
		UserID:       ticket.CreatedBy,
		ActivityType: models.ActivityAIAnalysisCompleted,
		NewValue:     result.Category,
		Description:  "AI analysis completed",
	})

	log.WithField("category", result.Category).
		WithField("priority", result.Priority).
		WithField("confidence", result.Confidence).
		Info("AI: Analysis completed successfully")

	// Auto-assign: SupportAgent from recommended team → any SupportAgent → Admin last resort
	if ticket.AssignedTo == nil && s.userRepo != nil {
		var assignee *models.User

		// 1st: SupportAgent in the AI-recommended team (Finance or Engineering)
		if result.RecommendedTeam != "" {
			if teamUsers, err := s.userRepo.ListByTeam(ticket.TenantID, result.RecommendedTeam); err == nil {
				for i := range teamUsers {
					if teamUsers[i].Role == models.RoleSupportAgent {
						u := teamUsers[i]
						assignee = &u
						break
					}
				}
			}
		}
		// 2nd: any SupportAgent in the tenant
		if assignee == nil {
			if agents, err := s.userRepo.ListByRole(ticket.TenantID, models.RoleSupportAgent); err == nil && len(agents) > 0 {
				u := agents[0]
				assignee = &u
			}
		}
		// 3rd: Admin only as last resort
		if assignee == nil {
			if admins, err := s.userRepo.ListByRole(ticket.TenantID, models.RoleAdmin); err == nil && len(admins) > 0 {
				u := admins[0]
				assignee = &u
			}
		}

		if assignee != nil {
			ticket.AssignedTo = &assignee.ID
			_ = s.ticketRepo.Update(ticket)
			_ = s.activityRepo.Create(&models.TicketActivity{
				TenantID:     ticket.TenantID,
				TicketID:     ticketID,
				UserID:       ticket.CreatedBy,
				ActivityType: models.ActivityAssignTicket,
				NewValue:     assignee.Name,
				Description:  fmt.Sprintf("Auto-assigned to %s (%s team)", assignee.Name, result.RecommendedTeam),
			})
			log.WithField("agent", assignee.Name).WithField("team", result.RecommendedTeam).Info("AI: Ticket auto-assigned")
		}
	}

	// Automatically trigger reply generation now that we have AI context
	if s.replySvc != nil {
		s.replySvc.GenerateForTicket(ticket.TenantID, ticketID, ticket.CreatedBy)
	}
}
