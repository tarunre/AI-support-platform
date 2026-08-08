package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/ayush/supportiq/internal/analytics"
	"github.com/ayush/supportiq/internal/analytics/reports"
	"github.com/ayush/supportiq/internal/dto"
	"github.com/ayush/supportiq/internal/middleware"
	"github.com/ayush/supportiq/internal/models"
	"github.com/ayush/supportiq/internal/utils"
	"github.com/gin-gonic/gin"
)

// AnalyticsHandler handles all /api/v1/analytics/** routes.
type AnalyticsHandler struct {
	svc       *analytics.Service
	reportSvc *reports.Service
	collector *reports.DataCollector
}

// NewAnalyticsHandler creates an AnalyticsHandler.
func NewAnalyticsHandler(svc *analytics.Service, reportSvc *reports.Service, collector *reports.DataCollector) *AnalyticsHandler {
	return &AnalyticsHandler{svc: svc, reportSvc: reportSvc, collector: collector}
}

// ─── Overview ────────────────────────────────────────────────────────────────

// Overview handles GET /api/v1/analytics/overview
func (h *AnalyticsHandler) Overview(c *gin.Context) {
	resp, err := h.svc.GetOverview(middleware.GetTenantID(c))
	if err != nil {
		utils.SendInternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// ─── Tickets ─────────────────────────────────────────────────────────────────

// TicketStats handles GET /api/v1/analytics/tickets
func (h *AnalyticsHandler) TicketStats(c *gin.Context) {
	f := h.parseFilter(c)
	resp, err := h.svc.GetTicketStats(middleware.GetTenantID(c), f)
	if err != nil {
		utils.SendInternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// ─── Agents ──────────────────────────────────────────────────────────────────

// AgentStats handles GET /api/v1/analytics/agents
// Admins see all agents; SupportAgents see only their own metrics.
func (h *AnalyticsHandler) AgentStats(c *gin.Context) {
	role, _ := c.Get("userRole")
	userID, _ := c.Get("userID")

	if role == string(models.RoleSupportAgent) {
		uid, ok := userID.(uint)
		if !ok {
			utils.SendError(c, http.StatusUnauthorized, "invalid user context")
			return
		}
		resp, err := h.svc.GetPersonalAgentStats(middleware.GetTenantID(c), uid)
		if err != nil {
			utils.SendInternalError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": resp})
		return
	}

	resp, err := h.svc.GetAgentStats(middleware.GetTenantID(c))
	if err != nil {
		utils.SendInternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// ─── AI ──────────────────────────────────────────────────────────────────────

// AIStats handles GET /api/v1/analytics/ai
func (h *AnalyticsHandler) AIStats(c *gin.Context) {
	f := h.parseFilter(c)
	resp, err := h.svc.GetAIStats(middleware.GetTenantID(c), f)
	if err != nil {
		utils.SendInternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// ─── Queue ────────────────────────────────────────────────────────────────────

// QueueStats handles GET /api/v1/analytics/queues
func (h *AnalyticsHandler) QueueStats(c *gin.Context) {
	resp, err := h.svc.GetQueueStats(middleware.GetTenantID(c))
	if err != nil {
		utils.SendInternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// ─── Email ────────────────────────────────────────────────────────────────────

// EmailStats handles GET /api/v1/analytics/email
func (h *AnalyticsHandler) EmailStats(c *gin.Context) {
	f := h.parseFilter(c)
	resp, err := h.svc.GetEmailStats(middleware.GetTenantID(c), f)
	if err != nil {
		utils.SendInternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// ─── Trends ──────────────────────────────────────────────────────────────────

// Trends handles GET /api/v1/analytics/trends
func (h *AnalyticsHandler) Trends(c *gin.Context) {
	f := h.parseFilter(c)
	resp, err := h.svc.GetTrends(middleware.GetTenantID(c), f)
	if err != nil {
		utils.SendInternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// ─── Aggregation trigger ──────────────────────────────────────────────────────

// TriggerAggregation handles POST /api/v1/analytics/aggregate (admin only)
func (h *AnalyticsHandler) TriggerAggregation(c *gin.Context) {
	go h.svc.TriggerAggregation()
	c.JSON(http.StatusAccepted, gin.H{"message": "aggregation started"})
}

// ─── Reports ──────────────────────────────────────────────────────────────────

// GenerateReport handles POST /api/v1/analytics/reports
func (h *AnalyticsHandler) GenerateReport(c *gin.Context) {
	var req dto.GenerateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	userID, _ := c.Get("userID")
	uid, _ := userID.(uint)
	tenantID := middleware.GetTenantID(c)

	report, err := h.reportSvc.Schedule(&req, tenantID, uid, h.collector)
	if err != nil {
		utils.SendInternalError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": toReportResponse(report)})
}

// ListReports handles GET /api/v1/analytics/reports
func (h *AnalyticsHandler) ListReports(c *gin.Context) {
	role, _ := c.Get("userRole")
	userID, _ := c.Get("userID")

	var generatedBy *uint
	if role == string(models.RoleSupportAgent) {
		uid, _ := userID.(uint)
		generatedBy = &uid
	}

	reports, err := h.reportSvc.ListReports(middleware.GetTenantID(c), generatedBy)
	if err != nil {
		utils.SendInternalError(c, err)
		return
	}

	out := make([]dto.ReportResponse, len(reports))
	for i, r := range reports {
		out[i] = toReportResponse(&r)
	}
	c.JSON(http.StatusOK, gin.H{"data": out, "total": len(out)})
}

// GetReport handles GET /api/v1/analytics/reports/:id
func (h *AnalyticsHandler) GetReport(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "invalid report id")
		return
	}
	report, err := h.reportSvc.GetReport(middleware.GetTenantID(c), uint(id))
	if err != nil {
		utils.SendError(c, http.StatusNotFound, "report not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": toReportResponse(report)})
}

// DownloadReport handles GET /api/v1/analytics/reports/:id/download
func (h *AnalyticsHandler) DownloadReport(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "invalid report id")
		return
	}

	data, mime, filename, err := h.reportSvc.DownloadReport(middleware.GetTenantID(c), uint(id))
	if err != nil {
		utils.SendError(c, http.StatusNotFound, err.Error())
		return
	}

	c.Header("Content-Disposition", `attachment; filename="`+utils.SafeFilename(filename)+`"`)
	c.Header("Content-Type", mime)
	c.Data(http.StatusOK, mime, data)
}

// ─── Filter parsing ───────────────────────────────────────────────────────────

func (h *AnalyticsHandler) parseFilter(c *gin.Context) dto.DateFilter {
	period := c.DefaultQuery("period", "last30")
	now := time.Now()
	f := dto.DateFilter{}

	switch period {
	case "today":
		f.StartDate = truncateDay(now)
		f.EndDate = now
	case "yesterday":
		y := truncateDay(now).AddDate(0, 0, -1)
		f.StartDate = y
		f.EndDate = y.Add(24*time.Hour - time.Second)
	case "last7":
		f.StartDate = truncateDay(now).AddDate(0, 0, -7)
		f.EndDate = now
	case "last30":
		f.StartDate = truncateDay(now).AddDate(0, 0, -30)
		f.EndDate = now
	case "last90":
		f.StartDate = truncateDay(now).AddDate(0, 0, -90)
		f.EndDate = now
	case "custom":
		if s := c.Query("start_date"); s != "" {
			if t, err := time.Parse("2006-01-02", s); err == nil {
				f.StartDate = t
			}
		}
		if e := c.Query("end_date"); e != "" {
			if t, err := time.Parse("2006-01-02", e); err == nil {
				f.EndDate = t.Add(24*time.Hour - time.Second)
			}
		}
	default:
		f.StartDate = truncateDay(now).AddDate(0, 0, -30)
		f.EndDate = now
	}

	if f.StartDate.IsZero() {
		f.StartDate = truncateDay(now).AddDate(0, 0, -30)
	}
	if f.EndDate.IsZero() {
		f.EndDate = now
	}

	if raw := c.Query("agent_id"); raw != "" {
		if n, err := strconv.ParseUint(raw, 10, 64); err == nil {
			uid := uint(n)
			f.AgentID = &uid
		}
	}
	f.Priority = c.Query("priority")
	f.Category = c.Query("category")
	f.Status = c.Query("status")
	f.Source = c.Query("source")

	return f
}

func truncateDay(t time.Time) time.Time {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func toReportResponse(r *models.Report) dto.ReportResponse {
	return dto.ReportResponse{
		ID:           r.ID,
		Name:         r.Name,
		ReportType:   r.Type,
		Format:       string(r.Format),
		Status:       string(r.Status),
		FileSize:     r.FileSize,
		Filters:      r.Parameters,
		GeneratedBy:  r.GeneratedBy,
		ErrorMessage: r.ErrorMsg,
		CreatedAt:    r.CreatedAt,
		CompletedAt:  nil,
	}
}
