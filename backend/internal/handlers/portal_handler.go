package handlers

import (
	"fmt"
	"net/http"

	"github.com/ayush/supportiq/internal/dto"
	jwtpkg "github.com/ayush/supportiq/internal/jwt"
	"github.com/ayush/supportiq/internal/models"
	"github.com/ayush/supportiq/internal/repositories"
	"github.com/ayush/supportiq/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ReplyTrigger is a minimal interface so the portal handler can trigger AI
// reply generation without importing the full services package.
type ReplyTrigger interface {
	GenerateAndSendPortalReply(tenantID uuid.UUID, ticketID uuid.UUID, userID uint)
}

// PortalHandler serves the customer-facing magic-link portal endpoints.
type PortalHandler struct {
	ticketRepo       *repositories.TicketRepository
	messageRepo      *repositories.EmailMessageRepository
	commentRepo      *repositories.CommentRepository
	activityRepo     *repositories.ActivityRepository
	emailAccountRepo *repositories.EmailAccountRepository
	jwtSecret        string
	replyTrigger     ReplyTrigger
}

func NewPortalHandler(
	ticketRepo *repositories.TicketRepository,
	messageRepo *repositories.EmailMessageRepository,
	activityRepo *repositories.ActivityRepository,
	jwtSecret string,
) *PortalHandler {
	return &PortalHandler{
		ticketRepo:   ticketRepo,
		messageRepo:  messageRepo,
		activityRepo: activityRepo,
		jwtSecret:    jwtSecret,
	}
}

func (h *PortalHandler) SetEmailAccountRepo(r *repositories.EmailAccountRepository) {
	h.emailAccountRepo = r
}
func (h *PortalHandler) SetCommentRepo(r *repositories.CommentRepository) {
	h.commentRepo = r
}

func (h *PortalHandler) SetReplyTrigger(rt ReplyTrigger) { h.replyTrigger = rt }

func (h *PortalHandler) validatePortalToken(c *gin.Context) (*jwtpkg.PortalClaims, bool) {
	tokenStr := c.Query("token")
	if tokenStr == "" {
		utils.SendError(c, http.StatusUnauthorized, "Missing portal token")
		return nil, false
	}
	claims, err := jwtpkg.ValidatePortalToken(tokenStr, h.jwtSecret)
	if err != nil {
		utils.SendError(c, http.StatusUnauthorized, "Invalid or expired portal link")
		return nil, false
	}
	return claims, true
}

// GetConversation handles GET /api/v1/portal/conversation?token=<jwt>
// Returns ticket info + all PUBLIC/CUSTOMER comments (Conversation tab messages).
func (h *PortalHandler) GetConversation(c *gin.Context) {
	claims, ok := h.validatePortalToken(c)
	if !ok {
		return
	}
	ticketID, err := uuid.Parse(claims.TicketID)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid ticket reference")
		return
	}
	ticket, err := h.ticketRepo.FindByIDUnscoped(ticketID)
	if err != nil {
		utils.SendError(c, http.StatusNotFound, "Ticket not found")
		return
	}
	if ticket.CustomerEmail != claims.CustomerEmail {
		utils.SendError(c, http.StatusForbidden, "Access denied")
		return
	}

	comments, err := h.commentRepo.ListPublicByTicketUnscoped(ticketID)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to load conversation")
		return
	}

	portalMsgs := make([]dto.PortalMessage, 0, len(comments))
	for _, cmt := range comments {
		isCustomer := string(cmt.CommentType) == "CUSTOMER"
		sender := "Support Team"
		if cmt.User != nil && !isCustomer {
			sender = cmt.User.Name
		}
		if isCustomer {
			sender = ticket.CustomerName
		}
		portalMsgs = append(portalMsgs, dto.PortalMessage{
			ID:        cmt.ID,
			Direction: map[bool]string{true: "INBOUND", false: "OUTBOUND"}[isCustomer],
			Body:      cmt.Message,
			Sender:    sender,
			CreatedAt: cmt.CreatedAt,
		})
	}

	utils.SendSuccess(c, http.StatusOK, "Conversation loaded", dto.PortalConversationResponse{
		Ticket: dto.PortalTicketInfo{
			TicketNumber: ticket.TicketNumber,
			Subject:      ticket.Subject,
			Status:       string(ticket.Status),
			CustomerName: ticket.CustomerName,
			CreatedAt:    ticket.CreatedAt,
		},
		Messages: portalMsgs,
	})
}

// Reply handles POST /api/v1/portal/reply
// Saves the customer message as a CUSTOMER-type ticket comment (shows in Conversation tab).
func (h *PortalHandler) Reply(c *gin.Context) {
	var req dto.PortalReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	claims, err := jwtpkg.ValidatePortalToken(req.Token, h.jwtSecret)
	if err != nil {
		utils.SendError(c, http.StatusUnauthorized, "Invalid or expired portal link")
		return
	}

	ticketID, err := uuid.Parse(claims.TicketID)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid ticket reference")
		return
	}

	ticket, err := h.ticketRepo.FindByIDUnscoped(ticketID)
	if err != nil {
		utils.SendError(c, http.StatusNotFound, "Ticket not found")
		return
	}
	if ticket.CustomerEmail != claims.CustomerEmail {
		utils.SendError(c, http.StatusForbidden, "Access denied")
		return
	}

	// Save as a CUSTOMER comment — visible in the Conversation tab
	comment := &models.TicketComment{
		TenantID:    ticket.TenantID,
		TicketID:    ticketID,
		UserID:      ticket.CreatedBy, // system user; UI distinguishes via CommentType
		Message:     req.Message,
		CommentType: models.CommentTypeCustomer,
	}
	if err := h.commentRepo.Create(comment); err != nil {
		utils.Logger.WithError(err).Error("Portal: failed to save customer comment")
		utils.SendError(c, http.StatusInternalServerError, "Failed to save message")
		return
	}

	_ = h.activityRepo.Create(&models.TicketActivity{
		TenantID:     ticket.TenantID,
		TicketID:     ticketID,
		UserID:       ticket.CreatedBy,
		ActivityType: models.ActivityEmailReceived,
		Description:  fmt.Sprintf("Customer replied via portal: %s", ticket.CustomerEmail),
	})

	// Trigger AI reply in background
	if h.replyTrigger != nil {
		go h.replyTrigger.GenerateAndSendPortalReply(ticket.TenantID, ticketID, ticket.CreatedBy)
	}

	utils.SendSuccess(c, http.StatusCreated, "Message sent", dto.PortalMessage{
		ID:        comment.ID,
		Direction: "INBOUND",
		Body:      comment.Message,
		Sender:    ticket.CustomerName,
		CreatedAt: comment.CreatedAt,
	})
}
