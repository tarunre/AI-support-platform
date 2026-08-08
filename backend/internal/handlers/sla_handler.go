package handlers

import (
	"net/http"
	"strconv"

	"github.com/ayush/supportiq/internal/dto"
	"github.com/ayush/supportiq/internal/middleware"
	"github.com/ayush/supportiq/internal/services"
	"github.com/ayush/supportiq/internal/utils"
	"github.com/gin-gonic/gin"
)

// SLAHandler serves CRUD endpoints for SLA policies and the SLA dashboard.
type SLAHandler struct {
	svc *services.SLAService
}

func NewSLAHandler(svc *services.SLAService) *SLAHandler {
	return &SLAHandler{svc: svc}
}

// ListPolicies handles GET /api/v1/sla-policies
func (h *SLAHandler) ListPolicies(c *gin.Context) {
	policies, code, err := h.svc.ListPolicies(middleware.GetTenantID(c))
	if err != nil {
		utils.SendError(c, code, err.Error())
		return
	}
	utils.SendSuccess(c, http.StatusOK, "SLA policies retrieved", policies)
}

// CreatePolicy handles POST /api/v1/sla-policies
func (h *SLAHandler) CreatePolicy(c *gin.Context) {
	var req dto.CreateSLAPolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}
	policy, code, err := h.svc.CreatePolicy(middleware.GetTenantID(c), &req)
	if err != nil {
		utils.SendError(c, code, err.Error())
		return
	}
	utils.SendSuccess(c, http.StatusCreated, "SLA policy created", policy)
}

// GetPolicy handles GET /api/v1/sla-policies/:id
func (h *SLAHandler) GetPolicy(c *gin.Context) {
	id, err := parseSLAID(c)
	if err != nil {
		return
	}
	policy, code, svcErr := h.svc.GetPolicy(middleware.GetTenantID(c), id)
	if svcErr != nil {
		utils.SendError(c, code, svcErr.Error())
		return
	}
	utils.SendSuccess(c, http.StatusOK, "SLA policy retrieved", policy)
}

// UpdatePolicy handles PUT /api/v1/sla-policies/:id
func (h *SLAHandler) UpdatePolicy(c *gin.Context) {
	id, err := parseSLAID(c)
	if err != nil {
		return
	}
	var req dto.UpdateSLAPolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}
	policy, code, svcErr := h.svc.UpdatePolicy(middleware.GetTenantID(c), id, &req)
	if svcErr != nil {
		utils.SendError(c, code, svcErr.Error())
		return
	}
	utils.SendSuccess(c, http.StatusOK, "SLA policy updated", policy)
}

// DeletePolicy handles DELETE /api/v1/sla-policies/:id
func (h *SLAHandler) DeletePolicy(c *gin.Context) {
	id, err := parseSLAID(c)
	if err != nil {
		return
	}
	code, svcErr := h.svc.DeletePolicy(middleware.GetTenantID(c), id)
	if svcErr != nil {
		utils.SendError(c, code, svcErr.Error())
		return
	}
	utils.SendSuccess(c, http.StatusOK, "SLA policy deleted", nil)
}

// GetDashboard handles GET /api/v1/tickets/sla
func (h *SLAHandler) GetDashboard(c *gin.Context) {
	dashboard, code, err := h.svc.GetDashboard(middleware.GetTenantID(c))
	if err != nil {
		utils.SendError(c, code, err.Error())
		return
	}
	utils.SendSuccess(c, http.StatusOK, "SLA dashboard retrieved", dashboard)
}

func parseSLAID(c *gin.Context) (uint, error) {
	raw := c.Param("id")
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "invalid SLA policy id")
		return 0, err
	}
	return uint(id), nil
}
