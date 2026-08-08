package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/ayush/supportiq/internal/ai/provider"
	"github.com/ayush/supportiq/internal/models"
	"github.com/ayush/supportiq/internal/queue"
	"github.com/ayush/supportiq/internal/repositories"
	"github.com/ayush/supportiq/internal/utils"
	"github.com/google/uuid"
)

// ReplyGenerator is a minimal interface so the worker doesn't import the full services package.
type ReplyGenerator interface {
	GenerateForTicket(tenantID uuid.UUID, ticketID uuid.UUID, userID uint)
}

// AIAnalysisHandler processes AI_ANALYSIS and RETRY_AI jobs.
type AIAnalysisHandler struct {
	ticketRepo   *repositories.TicketRepository
	activityRepo *repositories.ActivityRepository
	userRepo     *repositories.UserRepository
	replyGen     ReplyGenerator
	aiProvider   provider.Provider
}

func NewAIAnalysisHandler(
	ticketRepo *repositories.TicketRepository,
	activityRepo *repositories.ActivityRepository,
	aiProvider provider.Provider,
) *AIAnalysisHandler {
	return &AIAnalysisHandler{
		ticketRepo:   ticketRepo,
		activityRepo: activityRepo,
		aiProvider:   aiProvider,
	}
}

func (h *AIAnalysisHandler) SetUserRepo(ur *repositories.UserRepository) { h.userRepo = ur }
func (h *AIAnalysisHandler) SetReplyGenerator(rg ReplyGenerator)         { h.replyGen = rg }

// Handle runs AI analysis for the given ticket job.
func (h *AIAnalysisHandler) Handle(ctx context.Context, job queue.Job) error {
	ticketID, err := uuid.Parse(job.TicketID)
	if err != nil {
		return fmt.Errorf("invalid ticket UUID %q: %w", job.TicketID, err)
	}

	log := utils.Logger.
		WithField("job_type", job.Type).
		WithField("ticket_id", ticketID)

	ticket, err := h.ticketRepo.FindByIDUnscoped(ticketID)
	if err != nil {
		return fmt.Errorf("ticket not found: %w", err)
	}
	tenantID := ticket.TenantID

	// Guard: skip if another goroutine (e.g. API server retry) already completed this.
	if ticket.AIProcessingStatus == models.AIStatusCompleted {
		log.Info("AI Analysis: already completed by another process — skipping")
		return nil
	}

	log.Info("AI Analysis: Starting")

	// Mark ticket as processing
	ticket.AIProcessingStatus = models.AIStatusProcessing
	_ = h.ticketRepo.UpdateAIFields(ticket)

	result, err := h.aiProvider.Analyze(ctx, provider.AnalysisRequest{
		Subject:      ticket.Subject,
		Description:  ticket.Description,
		CustomerName: ticket.CustomerName,
		Category:     string(ticket.Category),
		Priority:     string(ticket.Priority),
	})
	if err != nil {
		ticket.AIProcessingStatus = models.AIStatusFailed
		_ = h.ticketRepo.UpdateAIFields(ticket)
		return fmt.Errorf("AI analysis failed: %w", err)
	}

	// Persist results
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

	if err := h.ticketRepo.UpdateAIFields(ticket); err != nil {
		return fmt.Errorf("failed to persist AI results: %w", err)
	}

	// Activity log
	_ = h.activityRepo.Create(&models.TicketActivity{
		TenantID:     tenantID,
		TicketID:     ticketID,
		UserID:       ticket.CreatedBy,
		ActivityType: models.ActivityAIAnalysisCompleted,
		NewValue:     result.Category,
		Description:  "AI analysis completed",
	})

	log.WithField("category", result.Category).
		WithField("confidence", result.Confidence).
		Info("AI Analysis: Completed successfully")

	// Auto-assign: SupportAgent from recommended team → any SupportAgent → Admin last resort
	if ticket.AssignedTo == nil && h.userRepo != nil {
		var assignee *models.User

		// 1st: SupportAgent in the AI-recommended team
		if result.RecommendedTeam != "" {
			teamUsers, err := h.userRepo.ListByTeam(ticket.TenantID, result.RecommendedTeam)
			log.WithField("team_query", result.RecommendedTeam).
				WithField("team_users_found", len(teamUsers)).
				WithField("team_err", err).
				Info("AI Analysis: team lookup")
			if err == nil {
				for i := range teamUsers {
					if teamUsers[i].Role == models.RoleSupportAgent {
						u := teamUsers[i]
						assignee = &u
						break
					}
				}
			}
		}
		// 2nd: any SupportAgent
		if assignee == nil {
			if agents, err := h.userRepo.ListByRole(ticket.TenantID, models.RoleSupportAgent); err == nil && len(agents) > 0 {
				a := agents[0]
				assignee = &a
			}
		}
		// 3rd: Admin only as last resort
		if assignee == nil {
			if admins, err := h.userRepo.ListByRole(ticket.TenantID, models.RoleAdmin); err == nil && len(admins) > 0 {
				a := admins[0]
				assignee = &a
			}
		}

		if assignee != nil {
			ticket.AssignedTo = &assignee.ID
			_ = h.ticketRepo.Update(ticket)
			_ = h.activityRepo.Create(&models.TicketActivity{
				TenantID:     ticket.TenantID,
				TicketID:     ticketID,
				UserID:       ticket.CreatedBy,
				ActivityType: models.ActivityAssignTicket,
				NewValue:     assignee.Name,
				Description:  fmt.Sprintf("Auto-assigned to %s (%s team)", assignee.Name, result.RecommendedTeam),
			})
			log.WithField("agent", assignee.Name).WithField("agent_team", assignee.Team).
				WithField("ai_team", result.RecommendedTeam).Info("AI Analysis: ticket auto-assigned")
		}
	}

	// Auto-trigger reply generation after successful analysis
	if h.replyGen != nil {
		h.replyGen.GenerateForTicket(ticket.TenantID, ticketID, ticket.CreatedBy)
	}

	return nil
}
