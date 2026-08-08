package handlers

import (
	"net/http"

	"github.com/ayush/supportiq/internal/dto"
	"github.com/ayush/supportiq/internal/services"
	"github.com/ayush/supportiq/internal/utils"
	"github.com/ayush/supportiq/internal/validators"
	"github.com/gin-gonic/gin"
)

// AuthHandler exposes HTTP endpoints for authentication.
type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register handles POST /api/v1/auth/register
// Creates a new tenant + admin user.
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterWithTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := validators.ValidatePasswordStrength(req.Password); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	resp, statusCode, err := h.authService.RegisterWithTenant(&req)
	if err != nil {
		utils.SendError(c, statusCode, err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusCreated, "Registration successful", resp)
}

// AgentJoin handles POST /api/v1/auth/agent-join
// Registers a new SupportAgent under an existing tenant by company slug.
func (h *AuthHandler) AgentJoin(c *gin.Context) {
	var req dto.AgentJoinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := validators.ValidatePasswordStrength(req.Password); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}
	resp, statusCode, err := h.authService.AgentJoin(&req)
	if err != nil {
		utils.SendError(c, statusCode, err.Error())
		return
	}
	utils.SendSuccess(c, http.StatusCreated, "Agent registered successfully", resp)
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	resp, statusCode, err := h.authService.Login(&req)
	if err != nil {
		utils.SendError(c, statusCode, err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Login successful", resp)
}

// Logout handles POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	utils.SendSuccess(c, http.StatusOK, "Logged out successfully", nil)
}

// Refresh handles POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refreshToken" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	resp, statusCode, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		utils.SendError(c, statusCode, err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Token refreshed", resp)
}

// Me handles GET /api/v1/auth/me
func (h *AuthHandler) Me(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.SendError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	resp, statusCode, err := h.authService.GetUserByID(userID.(uint))
	if err != nil {
		utils.SendError(c, statusCode, err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusOK, "User retrieved", resp)
}
