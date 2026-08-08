package handlers

import (
	"net/http"
	"strconv"

	"github.com/ayush/supportiq/internal/dto"
	"github.com/ayush/supportiq/internal/middleware"
	"github.com/ayush/supportiq/internal/services"
	"github.com/ayush/supportiq/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// EmailHandler serves all email-related REST endpoints.
type EmailHandler struct {
	accountSvc *services.EmailAccountService
	emailSvc   *services.EmailService
}

func NewEmailHandler(accountSvc *services.EmailAccountService, emailSvc *services.EmailService) *EmailHandler {
	return &EmailHandler{accountSvc: accountSvc, emailSvc: emailSvc}
}

// ── Email Account endpoints (admin only) ──────────────────────────────────────

// ListAccounts handles GET /api/v1/email/accounts
func (h *EmailHandler) ListAccounts(c *gin.Context) {
	accounts, code, err := h.accountSvc.List(middleware.GetTenantID(c))
	if err != nil {
		utils.SendError(c, code, err.Error())
		return
	}
	utils.SendSuccess(c, http.StatusOK, "Email accounts retrieved", accounts)
}

// CreateAccount handles POST /api/v1/email/accounts
func (h *EmailHandler) CreateAccount(c *gin.Context) {
	var req dto.CreateEmailAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}
	account, code, err := h.accountSvc.Create(middleware.GetTenantID(c), &req)
	if err != nil {
		utils.SendError(c, code, err.Error())
		return
	}
	utils.SendSuccess(c, http.StatusCreated, "Email account created", account)
}

// UpdateAccount handles PUT /api/v1/email/accounts/:id
func (h *EmailHandler) UpdateAccount(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid account ID")
		return
	}
	var req dto.UpdateEmailAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}
	account, code, err := h.accountSvc.Update(middleware.GetTenantID(c), uint(id), &req)
	if err != nil {
		utils.SendError(c, code, err.Error())
		return
	}
	utils.SendSuccess(c, http.StatusOK, "Email account updated", account)
}

// DeleteAccount handles DELETE /api/v1/email/accounts/:id
func (h *EmailHandler) DeleteAccount(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid account ID")
		return
	}
	code, err := h.accountSvc.Delete(middleware.GetTenantID(c), uint(id))
	if err != nil {
		utils.SendError(c, code, err.Error())
		return
	}
	utils.SendSuccess(c, http.StatusOK, "Email account deleted", nil)
}

// TestConnection handles POST /api/v1/email/accounts/:id/test
func (h *EmailHandler) TestConnection(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid account ID")
		return
	}

	protocol := c.DefaultQuery("protocol", "smtp")
	var testErr error
	if protocol == "imap" {
		testErr = h.accountSvc.TestIMAP(middleware.GetTenantID(c), uint(id))
	} else {
		testErr = h.accountSvc.TestSMTP(middleware.GetTenantID(c), uint(id))
	}

	if testErr != nil {
		utils.SendError(c, http.StatusBadGateway, testErr.Error())
		return
	}
	utils.SendSuccess(c, http.StatusOK, "Connection successful", nil)
}

// Monitor handles GET /api/v1/email/monitor
func (h *EmailHandler) Monitor(c *gin.Context) {
	stats, err := h.emailSvc.GetMonitorStats()
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to fetch email stats")
		return
	}
	utils.SendSuccess(c, http.StatusOK, "Email monitor stats", stats)
}

// Sync handles POST /api/v1/email/sync — triggers an immediate mailbox poll
func (h *EmailHandler) Sync(c *gin.Context) {
	go h.emailSvc.SyncNow(c.Request.Context())
	utils.SendSuccess(c, http.StatusAccepted, "Mailbox sync triggered", nil)
}

// ── Ticket email endpoints ────────────────────────────────────────────────────

// GetTicketEmails handles GET /api/v1/tickets/:id/emails
func (h *EmailHandler) GetTicketEmails(c *gin.Context) {
	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid ticket ID")
		return
	}
	emails, code, err := h.emailSvc.GetTicketEmails(ticketID)
	if err != nil {
		utils.SendError(c, code, err.Error())
		return
	}
	utils.SendSuccess(c, http.StatusOK, "Ticket emails retrieved", emails)
}

// SendEmail handles POST /api/v1/tickets/:id/send-email
func (h *EmailHandler) SendEmail(c *gin.Context) {
	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid ticket ID")
		return
	}
	var req dto.SendEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}
	userID := c.GetUint("userID")
	if err := h.emailSvc.SendEmail(c.Request.Context(), ticketID, &req, userID); err != nil {
		utils.SendError(c, http.StatusBadGateway, err.Error())
		return
	}
	utils.SendSuccess(c, http.StatusOK, "Email sent", nil)
}
