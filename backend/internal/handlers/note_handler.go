package handlers

import (
	"net/http"

	"github.com/ayush/supportiq/internal/dto"
	"github.com/ayush/supportiq/internal/middleware"
	"github.com/ayush/supportiq/internal/services"
	"github.com/ayush/supportiq/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// NoteHandler is a thin HTTP layer for ticket notes.
type NoteHandler struct {
	service *services.NoteService
}

func NewNoteHandler(service *services.NoteService) *NoteHandler {
	return &NoteHandler{service: service}
}

// Create handles POST /api/v1/tickets/:id/notes
func (h *NoteHandler) Create(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid ticket ID")
		return
	}

	var req dto.CreateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

		resp, statusCode, err := h.service.Create(middleware.GetTenantID(c), ticketID, &req, userID)
	if err != nil {
		utils.SendError(c, statusCode, err.Error())
		return
	}
	utils.SendSuccess(c, http.StatusCreated, "Note added successfully", resp)
}

// List handles GET /api/v1/tickets/:id/notes
func (h *NoteHandler) List(c *gin.Context) {
	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid ticket ID")
		return
	}

	resp, statusCode, err := h.service.List(middleware.GetTenantID(c), ticketID)
	if err != nil {
		utils.SendError(c, statusCode, err.Error())
		return
	}
	utils.SendSuccess(c, http.StatusOK, "Notes retrieved", resp)
}
