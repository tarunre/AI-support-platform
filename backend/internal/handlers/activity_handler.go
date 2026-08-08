package handlers

import (
	"net/http"

	"github.com/ayush/supportiq/internal/dto"
	"github.com/ayush/supportiq/internal/middleware"
	"github.com/ayush/supportiq/internal/models"
	"github.com/ayush/supportiq/internal/repositories"
	"github.com/ayush/supportiq/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ActivityHandler serves read-only activity timeline endpoints.
type ActivityHandler struct {
	activityRepo *repositories.ActivityRepository
}

func NewActivityHandler(activityRepo *repositories.ActivityRepository) *ActivityHandler {
	return &ActivityHandler{activityRepo: activityRepo}
}

// ListByTicket handles GET /api/v1/tickets/:id/activity
func (h *ActivityHandler) ListByTicket(c *gin.Context) {
	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid ticket ID")
		return
	}

	activities, err := h.activityRepo.ListByTicketID(middleware.GetTenantID(c), ticketID)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to retrieve activities")
		return
	}

	responses := make([]dto.ActivityResponse, len(activities))
	for i := range activities {
		responses[i] = toActivityResponse(&activities[i])
	}
	utils.SendSuccess(c, http.StatusOK, "Activities retrieved", responses)
}

// ListRecent handles GET /api/v1/activities (global feed for dashboard)
func (h *ActivityHandler) ListRecent(c *gin.Context) {
	activities, err := h.activityRepo.ListRecent(middleware.GetTenantID(c), 20)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to retrieve activities")
		return
	}

	responses := make([]dto.ActivityResponse, len(activities))
	for i := range activities {
		responses[i] = toActivityResponse(&activities[i])
	}
	utils.SendSuccess(c, http.StatusOK, "Recent activities retrieved", responses)
}

func toActivityResponse(a *models.TicketActivity) dto.ActivityResponse {
	resp := dto.ActivityResponse{
		ID:           a.ID,
		TicketID:     a.TicketID,
		ActivityType: a.ActivityType,
		OldValue:     a.OldValue,
		NewValue:     a.NewValue,
		Description:  a.Description,
		CreatedAt:    a.CreatedAt,
	}
	if a.User != nil {
		ur := dto.UserResponse{ID: a.User.ID, Name: a.User.Name, Email: a.User.Email, Role: string(a.User.Role)}
		resp.User = &ur
	}
	return resp
}
