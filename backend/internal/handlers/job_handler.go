package handlers

import (
	"net/http"
	"strconv"

	"github.com/ayush/supportiq/internal/services"
	"github.com/ayush/supportiq/internal/utils"
	"github.com/gin-gonic/gin"
)

// JobHandler serves the background job monitoring endpoints (admin only).
type JobHandler struct {
	jobSvc *services.JobService
}

func NewJobHandler(jobSvc *services.JobService) *JobHandler {
	return &JobHandler{jobSvc: jobSvc}
}

// List handles GET /api/v1/jobs
func (h *JobHandler) List(c *gin.Context) {
	var q services.ListJobsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	jobs, total, totalPages, err := h.jobSvc.ListJobs(q)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to fetch jobs")
		return
	}

	page := q.Page
	if page < 1 {
		page = 1
	}
	limit := q.Limit
	if limit < 1 {
		limit = 20
	}

	utils.SendSuccess(c, http.StatusOK, "Jobs retrieved", gin.H{
		"items":        jobs,
		"total_count":  total,
		"current_page": page,
		"total_pages":  totalPages,
		"limit":        limit,
	})
}

// GetByID handles GET /api/v1/jobs/:id
func (h *JobHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid job ID")
		return
	}

	job, err := h.jobSvc.GetJob(uint(id))
	if err != nil {
		utils.SendError(c, http.StatusNotFound, "Job not found")
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Job retrieved", job)
}

// Retry handles POST /api/v1/jobs/:id/retry
func (h *JobHandler) Retry(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid job ID")
		return
	}

	job, err := h.jobSvc.RetryJob(uint(id))
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Job re-queued", job)
}

// Statistics handles GET /api/v1/jobs/statistics
func (h *JobHandler) Statistics(c *gin.Context) {
	stats, err := h.jobSvc.GetStatistics()
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to fetch statistics")
		return
	}

	queueStats, err := h.jobSvc.QueueStats(c.Request.Context())
	if err != nil {
		queueStats = map[string]int64{}
	}

	utils.SendSuccess(c, http.StatusOK, "Statistics retrieved", gin.H{
		"jobs":  stats,
		"queue": queueStats,
	})
}
