package handlers

import (
	"net/http"

	"github.com/ayush/supportiq/internal/utils"
	"github.com/gin-gonic/gin"
)

// HealthHandler handles health-check related endpoints.
type HealthHandler struct{}

// NewHealthHandler constructs a HealthHandler via dependency injection.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Check responds with a 200 OK and a success message to confirm the service is running.
func (h *HealthHandler) Check(c *gin.Context) {
	utils.SendSuccess(c, http.StatusOK, "Backend running successfully", nil)
}
