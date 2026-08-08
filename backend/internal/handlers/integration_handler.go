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

// IntegrationHandler handles HTTP requests for the integrations resource.
type IntegrationHandler struct {
	svc *services.IntegrationService
}

// NewIntegrationHandler creates a new IntegrationHandler.
func NewIntegrationHandler(svc *services.IntegrationService) *IntegrationHandler {
	return &IntegrationHandler{svc: svc}
}

// List returns all configured integrations.
// GET /integrations
func (h *IntegrationHandler) List(c *gin.Context) {
	list, err := h.svc.List(middleware.GetTenantID(c))
	if err != nil {
		utils.SendInternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"integrations": list})
}

// Create adds a new integration.
// POST /integrations
func (h *IntegrationHandler) Create(c *gin.Context) {
	var req dto.CreateIntegrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDVal, _ := c.Get("userID")
	userID, _ := userIDVal.(uint)
	resp, err := h.svc.Create(middleware.GetTenantID(c), req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"integration": resp})
}

// Update modifies an existing integration.
// PUT /integrations/:id
func (h *IntegrationHandler) Update(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return
	}
	var req dto.UpdateIntegrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.svc.Update(middleware.GetTenantID(c), id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"integration": resp})
}

// Delete removes an integration.
// DELETE /integrations/:id
func (h *IntegrationHandler) Delete(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return
	}
	if err := h.svc.Delete(middleware.GetTenantID(c), id); err != nil {
		utils.SendInternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Integration deleted"})
}

// TestConnection runs a live connectivity test.
// POST /integrations/:id/test
func (h *IntegrationHandler) TestConnection(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return
	}
	if err := h.svc.TestConnection(c.Request.Context(), middleware.GetTenantID(c), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Connection successful"})
}

// ListEvents returns the most recent event delivery log for an integration.
// GET /integrations/:id/events
func (h *IntegrationHandler) ListEvents(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return
	}
	events, err := h.svc.ListEvents(middleware.GetTenantID(c), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"events": events})
}

// GetTicketIntegrations returns external issue links for a ticket.
// GET /tickets/:id/integrations
func (h *IntegrationHandler) GetTicketIntegrations(c *gin.Context) {
	ticketID, err := parseUUIDParam(c, "id")
	if err != nil {
		return
	}
	items, err := h.svc.GetTicketIntegrations(middleware.GetTenantID(c), ticketID)
	if err != nil {
		utils.SendInternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"integrations": items})
}

// CreateJiraIssue creates a Jira issue from a ticket.
// POST /tickets/:id/create-jira
func (h *IntegrationHandler) CreateJiraIssue(c *gin.Context) {
	ticketID, err := parseUUIDParam(c, "id")
	if err != nil {
		return
	}
	resp, err := h.svc.CreateJiraIssue(c.Request.Context(), middleware.GetTenantID(c), ticketID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"integration": resp})
}

// CreateLinearIssue creates a Linear issue from a ticket.
// POST /tickets/:id/create-linear
func (h *IntegrationHandler) CreateLinearIssue(c *gin.Context) {
	ticketID, err := parseUUIDParam(c, "id")
	if err != nil {
		return
	}
	resp, err := h.svc.CreateLinearIssue(c.Request.Context(), middleware.GetTenantID(c), ticketID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"integration": resp})
}

// CreateGitHubIssue creates a GitHub issue from a ticket.
// POST /tickets/:id/create-github-issue
func (h *IntegrationHandler) CreateGitHubIssue(c *gin.Context) {
	ticketID, err := parseUUIDParam(c, "id")
	if err != nil {
		return
	}
	resp, err := h.svc.CreateGitHubIssue(c.Request.Context(), middleware.GetTenantID(c), ticketID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"integration": resp})
}

// ─── helpers ────────────────────────────────────────────────────────────────

func parseUintParam(c *gin.Context, param string) (uint, error) {
	raw := c.Param(param)
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return 0, err
	}
	return uint(id), nil
}

func parseUUIDParam(c *gin.Context, param string) (uuid.UUID, error) {
	raw := c.Param(param)
	id, err := uuid.Parse(raw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return uuid.Nil, err
	}
	return id, nil
}
