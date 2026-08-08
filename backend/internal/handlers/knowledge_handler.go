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

// KnowledgeHandler serves the knowledge base CRUD endpoints (admin only).
type KnowledgeHandler struct {
	svc *services.KnowledgeService
}

func NewKnowledgeHandler(svc *services.KnowledgeService) *KnowledgeHandler {
	return &KnowledgeHandler{svc: svc}
}

// List handles GET /api/v1/knowledge-base
func (h *KnowledgeHandler) List(c *gin.Context) {
	var q dto.ListKnowledgeQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	_, _, resp, err := h.svc.List(middleware.GetTenantID(c), q)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to fetch knowledge base")
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Knowledge base retrieved", resp)
}

// Create handles POST /api/v1/knowledge-base
func (h *KnowledgeHandler) Create(c *gin.Context) {
	var req dto.CreateKnowledgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	doc, err := h.svc.Create(middleware.GetTenantID(c), req)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusCreated, "Knowledge base document created", dto.KnowledgeResponse{
		ID:        doc.ID,
		Title:     doc.Title,
		Category:  string(doc.Category),
		Content:   doc.Content,
		IsActive:  doc.IsActive,
		CreatedAt: doc.CreatedAt,
		UpdatedAt: doc.UpdatedAt,
	})
}

// Update handles PUT /api/v1/knowledge-base/:id
func (h *KnowledgeHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid document ID")
		return
	}

	var req dto.UpdateKnowledgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	doc, err := h.svc.Update(middleware.GetTenantID(c), uint(id), req)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Knowledge base document updated", dto.KnowledgeResponse{
		ID:        doc.ID,
		Title:     doc.Title,
		Category:  string(doc.Category),
		Content:   doc.Content,
		IsActive:  doc.IsActive,
		CreatedAt: doc.CreatedAt,
		UpdatedAt: doc.UpdatedAt,
	})
}

// Delete handles DELETE /api/v1/knowledge-base/:id
func (h *KnowledgeHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid document ID")
		return
	}

	if err := h.svc.Delete(middleware.GetTenantID(c), uint(id)); err != nil {
		utils.SendError(c, http.StatusNotFound, "Document not found")
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Knowledge base document deleted", nil)
}
