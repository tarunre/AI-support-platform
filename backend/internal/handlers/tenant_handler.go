package handlers

import (
	"net/http"
	"strconv"

	"github.com/ayush/supportiq/internal/dto"
	"github.com/ayush/supportiq/internal/services"
	"github.com/ayush/supportiq/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TenantHandler handles tenant management endpoints (SuperAdmin only).
type TenantHandler struct {
	svc *services.TenantService
}

func NewTenantHandler(svc *services.TenantService) *TenantHandler {
	return &TenantHandler{svc: svc}
}

// List handles GET /api/v1/tenants
func (h *TenantHandler) List(c *gin.Context) {
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	tenants, total, statusCode, err := h.svc.List(status, page, limit)
	if err != nil {
		utils.SendError(c, statusCode, err.Error())
		return
	}
	utils.SendSuccess(c, http.StatusOK, "Tenants retrieved", gin.H{
		"tenants": tenants,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

// GetByID handles GET /api/v1/tenants/:id
func (h *TenantHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid tenant ID")
		return
	}
	tenant, statusCode, err := h.svc.GetByID(id)
	if err != nil {
		utils.SendError(c, statusCode, err.Error())
		return
	}
	utils.SendSuccess(c, http.StatusOK, "Tenant retrieved", tenant)
}

// Create handles POST /api/v1/tenants
func (h *TenantHandler) Create(c *gin.Context) {
	var req dto.CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}
	tenant, statusCode, err := h.svc.Create(&req)
	if err != nil {
		utils.SendError(c, statusCode, err.Error())
		return
	}
	utils.SendSuccess(c, http.StatusCreated, "Tenant created", tenant)
}

// Update handles PUT /api/v1/tenants/:id
func (h *TenantHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid tenant ID")
		return
	}
	var req dto.UpdateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}
	tenant, statusCode, err := h.svc.Update(id, &req)
	if err != nil {
		utils.SendError(c, statusCode, err.Error())
		return
	}
	utils.SendSuccess(c, http.StatusOK, "Tenant updated", tenant)
}

// Delete handles DELETE /api/v1/tenants/:id
func (h *TenantHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid tenant ID")
		return
	}
	statusCode, err := h.svc.Delete(id)
	if err != nil {
		utils.SendError(c, statusCode, err.Error())
		return
	}
	utils.SendSuccess(c, http.StatusOK, "Tenant deleted", nil)
}

// Overview handles GET /api/v1/superadmin/overview
func (h *TenantHandler) Overview(c *gin.Context) {
	overview, statusCode, err := h.svc.GetPlatformOverview()
	if err != nil {
		utils.SendError(c, statusCode, err.Error())
		return
	}
	utils.SendSuccess(c, http.StatusOK, "Overview retrieved", overview)
}

// GetSettings handles GET /api/v1/settings (current tenant's settings)
func (h *TenantHandler) GetSettings(c *gin.Context) {
	tenantID, ok := c.Get("tenantID")
	if !ok {
		utils.SendError(c, http.StatusUnauthorized, "No tenant context")
		return
	}
	id, _ := tenantID.(uuid.UUID)
	tenant, statusCode, err := h.svc.GetByID(id)
	if err != nil {
		utils.SendError(c, statusCode, err.Error())
		return
	}
	utils.SendSuccess(c, http.StatusOK, "Settings retrieved", tenant)
}

// UpdateSettings handles PUT /api/v1/settings (current tenant Admin)
func (h *TenantHandler) UpdateSettings(c *gin.Context) {
	tenantID, ok := c.Get("tenantID")
	if !ok {
		utils.SendError(c, http.StatusUnauthorized, "No tenant context")
		return
	}
	id, _ := tenantID.(uuid.UUID)
	var req dto.UpdateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}
	tenant, statusCode, err := h.svc.Update(id, &req)
	if err != nil {
		utils.SendError(c, statusCode, err.Error())
		return
	}
	utils.SendSuccess(c, http.StatusOK, "Settings updated", tenant)
}
