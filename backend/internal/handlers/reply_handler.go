package handlers

import (
	"net/http"

	"github.com/ayush/supportiq/internal/middleware"
	"github.com/ayush/supportiq/internal/services"
	"github.com/ayush/supportiq/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ReplyHandler serves the AI reply workflow endpoints.
type ReplyHandler struct {
	replySvc *services.ReplyService
	emailSvc *services.EmailService // optional — queues outbound email on approval
}

func NewReplyHandler(replySvc *services.ReplyService) *ReplyHandler {
	return &ReplyHandler{replySvc: replySvc}
}

// SetEmailService injects the email service after construction.
func (h *ReplyHandler) SetEmailService(svc *services.EmailService) {
	h.emailSvc = svc
}

// GetReply handles GET /api/v1/tickets/:id/reply
func (h *ReplyHandler) GetReply(c *gin.Context) {
	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid ticket ID")
		return
	}

	reply, err := h.replySvc.GetLatest(middleware.GetTenantID(c), ticketID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendSuccess(c, http.StatusOK, "No reply generated yet", nil)
			return
		}
		utils.SendError(c, http.StatusInternalServerError, "Failed to fetch reply")
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Reply retrieved", services.ToReplyResponse(reply))
}

// GenerateReply handles POST /api/v1/tickets/:id/reply/generate
func (h *ReplyHandler) GenerateReply(c *gin.Context) {
	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid ticket ID")
		return
	}

	userID := c.GetUint("userID")

	reply, err := h.replySvc.Generate(c.Request.Context(), middleware.GetTenantID(c), ticketID, userID)
	if err != nil {
		utils.SendError(c, http.StatusUnprocessableEntity, err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusCreated, "Reply generated", services.ToReplyResponse(reply))
}

// RegenerateReply handles POST /api/v1/tickets/:id/reply/regenerate
func (h *ReplyHandler) RegenerateReply(c *gin.Context) {
	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid ticket ID")
		return
	}

	userID := c.GetUint("userID")

	reply, err := h.replySvc.Regenerate(c.Request.Context(), middleware.GetTenantID(c), ticketID, userID)
	if err != nil {
		utils.SendError(c, http.StatusUnprocessableEntity, err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusCreated, "Reply regenerated", services.ToReplyResponse(reply))
}

// EditReply handles PATCH /api/v1/tickets/:id/reply/edit
func (h *ReplyHandler) EditReply(c *gin.Context) {
	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid ticket ID")
		return
	}

	var req struct {
		Reply string `json:"reply" binding:"required,min=10"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	userID := c.GetUint("userID")

	reply, err := h.replySvc.Edit(c.Request.Context(), middleware.GetTenantID(c), ticketID, userID, req.Reply)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Reply updated", services.ToReplyResponse(reply))
}

// ApproveReply handles POST /api/v1/tickets/:id/reply/approve
func (h *ReplyHandler) ApproveReply(c *gin.Context) {
	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid ticket ID")
		return
	}

	userID := c.GetUint("userID")

	reply, err := h.replySvc.Approve(c.Request.Context(), middleware.GetTenantID(c), ticketID, userID)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Queue outbound email automatically when email service is available
	if h.emailSvc != nil {
		// Prefer edited reply if present, otherwise use the generated reply
		replyText := reply.EditedReply
		if replyText == "" {
			replyText = reply.GeneratedReply
		}
		if qErr := h.emailSvc.QueueReplyForTicket(c.Request.Context(), ticketID, replyText, userID); qErr != nil {
			utils.Logger.WithError(qErr).Warn("ApproveReply: failed to queue outbound email")
		}
	}

	utils.SendSuccess(c, http.StatusOK, "Reply approved", services.ToReplyResponse(reply))
}

// RejectReply handles POST /api/v1/tickets/:id/reply/reject
func (h *ReplyHandler) RejectReply(c *gin.Context) {
	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid ticket ID")
		return
	}

	userID := c.GetUint("userID")

	reply, err := h.replySvc.Reject(c.Request.Context(), middleware.GetTenantID(c), ticketID, userID)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Reply rejected", services.ToReplyResponse(reply))
}
