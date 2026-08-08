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

// CommentHandler is a thin HTTP layer for ticket comments.
type CommentHandler struct {
	service *services.CommentService
}

func NewCommentHandler(service *services.CommentService) *CommentHandler {
	return &CommentHandler{service: service}
}

// Create handles POST /api/v1/tickets/:id/comments
func (h *CommentHandler) Create(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid ticket ID")
		return
	}

	var req dto.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

		resp, statusCode, err := h.service.Create(middleware.GetTenantID(c), ticketID, &req, userID)
	if err != nil {
		utils.SendError(c, statusCode, err.Error())
		return
	}
	utils.SendSuccess(c, http.StatusCreated, "Comment added successfully", resp)
}

// List handles GET /api/v1/tickets/:id/comments
func (h *CommentHandler) List(c *gin.Context) {
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
	utils.SendSuccess(c, http.StatusOK, "Comments retrieved", resp)
}
