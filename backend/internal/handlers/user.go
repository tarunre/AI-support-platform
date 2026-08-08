package handlers

import (
	"net/http"

	"github.com/ayush/supportiq/internal/dto"
	"github.com/ayush/supportiq/internal/middleware"
	"github.com/ayush/supportiq/internal/models"
	"github.com/ayush/supportiq/internal/repositories"
	"github.com/ayush/supportiq/internal/utils"
	"github.com/gin-gonic/gin"
)

// UserHandler exposes user-related read endpoints for supporting the UI.
type UserHandler struct {
	userRepo *repositories.UserRepository
}

func NewUserHandler(userRepo *repositories.UserRepository) *UserHandler {
	return &UserHandler{userRepo: userRepo}
}

// ListAgents handles GET /api/v1/users/agents
// Returns all active SupportAgent users — used by the ticket assignment UI.
func (h *UserHandler) ListAgents(c *gin.Context) {
	users, err := h.userRepo.ListByRole(middleware.GetTenantID(c), models.RoleSupportAgent)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to retrieve agents")
		return
	}

	responses := make([]dto.UserResponse, len(users))
	for i, u := range users {
		responses[i] = dto.UserResponse{
			ID:       u.ID,
			Name:     u.Name,
			Email:    u.Email,
			Role:     string(u.Role),
			IsActive: u.IsActive,
		}
	}
	utils.SendSuccess(c, http.StatusOK, "Agents retrieved", responses)
}
