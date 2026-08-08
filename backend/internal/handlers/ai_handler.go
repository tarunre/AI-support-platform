package handlers

import (
	"net/http"

	"github.com/ayush/supportiq/internal/dto"
	"github.com/ayush/supportiq/internal/middleware"
	"github.com/ayush/supportiq/internal/repositories"
	"github.com/ayush/supportiq/internal/services"
	"github.com/ayush/supportiq/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AIHandler serves the AI analysis endpoints.
type AIHandler struct {
	ticketRepo *repositories.TicketRepository
	aiService  *services.AIService
}

func NewAIHandler(ticketRepo *repositories.TicketRepository, aiService *services.AIService) *AIHandler {
	return &AIHandler{ticketRepo: ticketRepo, aiService: aiService}
}

// GetAnalysis handles GET /api/v1/tickets/:id/ai-analysis
func (h *AIHandler) GetAnalysis(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid ticket ID")
		return
	}

	ticket, err := h.ticketRepo.FindByID(middleware.GetTenantID(c), id)
	if err != nil {
		utils.SendError(c, http.StatusNotFound, "Ticket not found")
		return
	}

	utils.SendSuccess(c, http.StatusOK, "AI analysis retrieved", dto.AIAnalysisResponse{
		ProcessingStatus: ticket.AIProcessingStatus,
		Category:         ticket.AICategory,
		Priority:         ticket.AIPriority,
		Sentiment:        ticket.AISentiment,
		RecommendedTeam:  ticket.AITeam,
		Confidence:       ticket.AIConfidence,
		Summary:          ticket.AISummary,
		Tags:             ticket.AITags,
		ProcessedAt:      ticket.ProcessedAt,
	})
}

// RetryAnalysis handles POST /api/v1/tickets/:id/retry-ai
func (h *AIHandler) RetryAnalysis(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid ticket ID")
		return
	}

	if _, err := h.ticketRepo.FindByID(middleware.GetTenantID(c), id); err != nil {
		utils.SendError(c, http.StatusNotFound, "Ticket not found")
		return
	}

	h.aiService.RetryAnalysis(id)
	utils.SendSuccess(c, http.StatusAccepted, "AI analysis queued for retry", nil)
}
