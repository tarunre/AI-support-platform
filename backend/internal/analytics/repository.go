package analytics

import (
	"time"

	"github.com/ayush/supportiq/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// AnalyticsRepository encapsulates all raw database queries used by analytics.
type AnalyticsRepository struct {
	db *gorm.DB
}

func NewAnalyticsRepository(db *gorm.DB) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

// ─── Live ticket counts ─────────────────────────────────────────────────────

type StatusCount struct {
	Status string
	Count  int64
}

type LabelCount struct {
	Label string
	Count int64
}

func (r *AnalyticsRepository) CountTicketsByStatus(tenantID uuid.UUID) ([]StatusCount, error) {
	var rows []struct {
		Status string
		Count  int64
	}
	err := r.db.Model(&models.Ticket{}).
		Select("status, COUNT(*) as count").
		Where("tenant_id = ?", tenantID).
		Group("status").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]StatusCount, len(rows))
	for i, row := range rows {
		out[i] = StatusCount{Status: row.Status, Count: row.Count}
	}
	return out, nil
}

func (r *AnalyticsRepository) CountTicketsByPriority(tenantID uuid.UUID, start, end time.Time) ([]LabelCount, error) {
	var rows []LabelCount
	err := r.db.Model(&models.Ticket{}).
		Select("priority as label, COUNT(*) as count").
		Where("tenant_id = ? AND created_at BETWEEN ? AND ?", tenantID, start, end).
		Group("priority").Order("count DESC").Find(&rows).Error
	return rows, err
}

func (r *AnalyticsRepository) CountTicketsByCategory(tenantID uuid.UUID, start, end time.Time) ([]LabelCount, error) {
	var rows []LabelCount
	err := r.db.Model(&models.Ticket{}).
		Select("COALESCE(NULLIF(ai_category,''), category) as label, COUNT(*) as count").
		Where("tenant_id = ? AND created_at BETWEEN ? AND ?", tenantID, start, end).
		Group("label").Order("count DESC").Find(&rows).Error
	return rows, err
}

func (r *AnalyticsRepository) CountTicketsBySource(tenantID uuid.UUID, start, end time.Time) ([]LabelCount, error) {
	var rows []LabelCount
	err := r.db.Model(&models.Ticket{}).
		Select("source as label, COUNT(*) as count").
		Where("tenant_id = ? AND created_at BETWEEN ? AND ?", tenantID, start, end).
		Group("source").Order("count DESC").Find(&rows).Error
	return rows, err
}

func (r *AnalyticsRepository) CountTicketsByHour(tenantID uuid.UUID, date time.Time) ([]LabelCount, error) {
	dayStart := truncateDay(date)
	dayEnd := dayStart.Add(24 * time.Hour)
	var rows []LabelCount
	err := r.db.Model(&models.Ticket{}).
		Select("TO_CHAR(created_at, 'HH24') as label, COUNT(*) as count").
		Where("tenant_id = ? AND created_at BETWEEN ? AND ?", tenantID, dayStart, dayEnd).
		Group("label").Order("label ASC").Find(&rows).Error
	return rows, err
}

func (r *AnalyticsRepository) CountTicketsInRange(tenantID uuid.UUID, start, end time.Time) (int64, error) {
	var count int64
	err := r.db.Model(&models.Ticket{}).
		Where("tenant_id = ? AND created_at BETWEEN ? AND ?", tenantID, start, end).
		Count(&count).Error
	return count, err
}

func (r *AnalyticsRepository) CountTicketsCreatedToday(tenantID uuid.UUID) (int64, error) {
	return r.CountTicketsInRange(tenantID, truncateDay(time.Now()), time.Now())
}

func (r *AnalyticsRepository) CountTicketsResolvedToday(tenantID uuid.UUID) (int64, error) {
	today := truncateDay(time.Now())
	var count int64
	err := r.db.Model(&models.Ticket{}).
		Where("tenant_id = ? AND status IN ('RESOLVED','CLOSED') AND updated_at >= ?", tenantID, today).
		Count(&count).Error
	return count, err
}

func (r *AnalyticsRepository) AvgResolutionHours(tenantID uuid.UUID, start, end time.Time) (float64, error) {
	var result struct{ Avg float64 }
	err := r.db.Model(&models.Ticket{}).
		Select("COALESCE(AVG(EXTRACT(EPOCH FROM (updated_at - created_at)) / 3600), 0) as avg").
		Where("tenant_id = ? AND status IN ('RESOLVED','CLOSED') AND updated_at BETWEEN ? AND ?", tenantID, start, end).
		Find(&result).Error
	return result.Avg, err
}

func (r *AnalyticsRepository) CountTicketsByStatusSnapshot(tenantID uuid.UUID) (map[string]int64, error) {
	rows, err := r.CountTicketsByStatus(tenantID)
	if err != nil {
		return nil, err
	}
	m := make(map[string]int64)
	for _, row := range rows {
		m[row.Status] = row.Count
	}
	return m, nil
}

// ─── Live AI counts ──────────────────────────────────────────────────────────

type AIReplySummary struct {
	Total    int64
	Approved int64
	Rejected int64
	Edited   int64
	Retried  int64
	AvgConf  float64
	AvgGenMs float64
}

func (r *AnalyticsRepository) SummariseAIReplies(tenantID uuid.UUID, start, end time.Time) (AIReplySummary, error) {
	var row struct {
		Total    int64
		Approved int64
		Rejected int64
		Edited   int64
		Retried  int64
		AvgConf  float64
		AvgGenMs float64
	}
	err := r.db.Model(&models.AIReply{}).
		Select(`
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'APPROVED')    as approved,
			COUNT(*) FILTER (WHERE status = 'REJECTED')    as rejected,
			COUNT(*) FILTER (WHERE status = 'APPROVED' AND edited_reply != '') as edited,
			COUNT(*) FILTER (WHERE status = 'REGENERATED') as retried,
			COALESCE(AVG(confidence), 0)                   as avg_conf,
			COALESCE(AVG(generation_time), 0)              as avg_gen_ms
		`).
		Where("tenant_id = ? AND created_at BETWEEN ? AND ?", tenantID, start, end).
		Find(&row).Error
	if err != nil {
		return AIReplySummary{}, err
	}
	return AIReplySummary{
		Total: row.Total, Approved: row.Approved, Rejected: row.Rejected,
		Edited: row.Edited, Retried: row.Retried,
		AvgConf: row.AvgConf, AvgGenMs: row.AvgGenMs,
	}, nil
}

func (r *AnalyticsRepository) AvgAIConfidence(tenantID uuid.UUID, start, end time.Time) (float64, error) {
	var result struct{ Avg float64 }
	err := r.db.Model(&models.Ticket{}).
		Select("COALESCE(AVG(ai_confidence), 0) as avg").
		Where("tenant_id = ? AND ai_confidence IS NOT NULL AND created_at BETWEEN ? AND ?", tenantID, start, end).
		Find(&result).Error
	return result.Avg, err
}

func (r *AnalyticsRepository) CountAIAnalysesCompleted(tenantID uuid.UUID, start, end time.Time) (int64, error) {
	var count int64
	err := r.db.Model(&models.Ticket{}).
		Where("tenant_id = ? AND ai_processing_status = 'COMPLETED' AND processed_at BETWEEN ? AND ?", tenantID, start, end).
		Count(&count).Error
	return count, err
}

func (r *AnalyticsRepository) CountAIFailures(tenantID uuid.UUID, start, end time.Time) (int64, error) {
	var count int64
	err := r.db.Model(&models.BackgroundJob{}).
		Where("tenant_id = ? AND job_type IN ('AI_ANALYSIS','GENERATE_REPLY','RETRY_AI','RETRY_REPLY') AND status = 'FAILED' AND created_at BETWEEN ? AND ?", tenantID, start, end).
		Count(&count).Error
	return count, err
}

func (r *AnalyticsRepository) TopAICategories(tenantID uuid.UUID, start, end time.Time, limit int) ([]LabelCount, error) {
	var rows []LabelCount
	err := r.db.Model(&models.Ticket{}).
		Select("ai_category as label, COUNT(*) as count").
		Where("tenant_id = ? AND ai_category != '' AND created_at BETWEEN ? AND ?", tenantID, start, end).
		Group("ai_category").Order("count DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *AnalyticsRepository) TopAISentiments(tenantID uuid.UUID, start, end time.Time) ([]LabelCount, error) {
	var rows []LabelCount
	err := r.db.Model(&models.Ticket{}).
		Select("ai_sentiment as label, COUNT(*) as count").
		Where("tenant_id = ? AND ai_sentiment != '' AND created_at BETWEEN ? AND ?", tenantID, start, end).
		Group("ai_sentiment").Order("count DESC").Find(&rows).Error
	return rows, err
}

// ─── Queue counts ────────────────────────────────────────────────────────────

type QueueSnapshot struct {
	Queued      int64
	Processing  int64
	Completed   int64
	Failed      int64
	Dead        int64
	Retrying    int64
	AvgQueueSec float64
}

func (r *AnalyticsRepository) SnapshotQueue(tenantID uuid.UUID) (QueueSnapshot, error) {
	var rows []struct {
		Status string
		Count  int64
	}
	err := r.db.Model(&models.BackgroundJob{}).
		Select("status, COUNT(*) as count").
		Where("tenant_id = ?", tenantID).
		Group("status").Find(&rows).Error
	if err != nil {
		return QueueSnapshot{}, err
	}
	snap := QueueSnapshot{}
	for _, row := range rows {
		switch row.Status {
		case "QUEUED":
			snap.Queued = row.Count
		case "PROCESSING":
			snap.Processing = row.Count
		case "COMPLETED":
			snap.Completed = row.Count
		case "FAILED":
			snap.Failed = row.Count
		case "DEAD":
			snap.Dead = row.Count
		case "RETRYING":
			snap.Retrying = row.Count
		}
	}
	var avgRow struct{ Avg float64 }
	_ = r.db.Model(&models.BackgroundJob{}).
		Select("COALESCE(AVG(EXTRACT(EPOCH FROM (COALESCE(started_at, NOW()) - created_at))), 0) as avg").
		Where("tenant_id = ? AND created_at > NOW() - INTERVAL '24 hours' AND status IN ('QUEUED','PROCESSING','COMPLETED')", tenantID).
		Find(&avgRow).Error
	snap.AvgQueueSec = avgRow.Avg
	return snap, nil
}

func (r *AnalyticsRepository) CountJobsByType(tenantID uuid.UUID) ([]LabelCount, error) {
	var rows []LabelCount
	err := r.db.Model(&models.BackgroundJob{}).
		Select("job_type as label, COUNT(*) as count").
		Where("tenant_id = ?", tenantID).
		Group("job_type").Order("count DESC").Find(&rows).Error
	return rows, err
}

// ─── Email counts ─────────────────────────────────────────────────────────────

type EmailSnapshot struct {
	Received    int64
	Sent        int64
	Failed      int64
	Queued      int64
	AvgDelivSec float64
}

func (r *AnalyticsRepository) SnapshotEmail(tenantID uuid.UUID, start, end time.Time) (EmailSnapshot, error) {
	var rows []struct {
		Direction string
		Status    string
		Count     int64
	}
	err := r.db.Model(&models.EmailMessage{}).
		Select("direction, status, COUNT(*) as count").
		Where("tenant_id = ? AND created_at BETWEEN ? AND ?", tenantID, start, end).
		Group("direction, status").Find(&rows).Error
	if err != nil {
		return EmailSnapshot{}, err
	}
	snap := EmailSnapshot{}
	for _, row := range rows {
		switch {
		case row.Direction == "INBOUND" && row.Status == "RECEIVED":
			snap.Received += row.Count
		case row.Direction == "OUTBOUND" && row.Status == "SENT":
			snap.Sent += row.Count
		case row.Direction == "OUTBOUND" && row.Status == "FAILED":
			snap.Failed += row.Count
		case row.Direction == "OUTBOUND" && row.Status == "QUEUED":
			snap.Queued += row.Count
		}
	}
	var avgRow struct{ Avg float64 }
	_ = r.db.Model(&models.EmailMessage{}).
		Select("COALESCE(AVG(EXTRACT(EPOCH FROM (sent_at - created_at))), 0) as avg").
		Where("tenant_id = ? AND direction = 'OUTBOUND' AND status = 'SENT' AND sent_at BETWEEN ? AND ?", tenantID, start, end).
		Find(&avgRow).Error
	snap.AvgDelivSec = avgRow.Avg
	return snap, nil
}

func (r *AnalyticsRepository) CountEmailsProcessedToday(tenantID uuid.UUID) (int64, error) {
	today := truncateDay(time.Now())
	var count int64
	err := r.db.Model(&models.EmailMessage{}).
		Where("tenant_id = ? AND created_at >= ?", tenantID, today).
		Count(&count).Error
	return count, err
}

// ─── Aggregated metrics table CRUD ──────────────────────────────────────────

func (r *AnalyticsRepository) UpsertDailyTicketMetrics(m *models.DailyTicketMetrics) error {
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "tenant_id"}, {Name: "date"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"tickets_created", "tickets_closed", "tickets_resolved",
			"tickets_reopened", "average_resolution_time",
			"average_first_response_time", "average_ai_processing_time",
		}),
	}).Create(m).Error
}

func (r *AnalyticsRepository) GetDailyTicketMetrics(tenantID uuid.UUID, start, end time.Time) ([]models.DailyTicketMetrics, error) {
	var rows []models.DailyTicketMetrics
	err := r.db.Where("tenant_id = ? AND date BETWEEN ? AND ?", tenantID, start, end).
		Order("date ASC").Find(&rows).Error
	return rows, err
}

func (r *AnalyticsRepository) UpsertAIMetrics(m *models.AIMetrics) error {
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "tenant_id"}, {Name: "date"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"analysis_generated", "replies_generated", "average_confidence",
			"average_generation_time", "approval_rate", "edit_rate",
			"rejection_rate", "retry_rate",
		}),
	}).Create(m).Error
}

func (r *AnalyticsRepository) GetAIMetrics(tenantID uuid.UUID, start, end time.Time) ([]models.AIMetrics, error) {
	var rows []models.AIMetrics
	err := r.db.Where("tenant_id = ? AND date BETWEEN ? AND ?", tenantID, start, end).
		Order("date ASC").Find(&rows).Error
	return rows, err
}

func (r *AnalyticsRepository) UpsertAgentMetrics(m *models.AgentMetrics) error {
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "tenant_id"}, {Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"tickets_assigned", "tickets_resolved",
			"average_resolution_time", "average_reply_time",
			"average_customer_rating", "last_calculated",
		}),
	}).Create(m).Error
}

func (r *AnalyticsRepository) GetAllAgentMetrics(tenantID uuid.UUID) ([]models.AgentMetrics, error) {
	var rows []models.AgentMetrics
	err := r.db.Preload("User").
		Where("tenant_id = ?", tenantID).
		Order("tickets_resolved DESC").Find(&rows).Error
	return rows, err
}

func (r *AnalyticsRepository) GetAgentMetricsByUserID(tenantID uuid.UUID, userID uint) (*models.AgentMetrics, error) {
	var m models.AgentMetrics
	err := r.db.Preload("User").
		Where("tenant_id = ? AND user_id = ?", tenantID, userID).
		First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *AnalyticsRepository) CountActiveTicketsForAgent(tenantID uuid.UUID, userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.Ticket{}).
		Where("tenant_id = ? AND assigned_to = ? AND status IN ('OPEN','IN_PROGRESS')", tenantID, userID).
		Count(&count).Error
	return count, err
}

func (r *AnalyticsRepository) AllSupportAgents(tenantID uuid.UUID) ([]models.User, error) {
	var users []models.User
	err := r.db.Where("tenant_id = ? AND is_active = true", tenantID).
		Order("id ASC").Find(&users).Error
	return users, err
}

// ─── Aggregation raw queries ─────────────────────────────────────────────────

type AgentRawMetrics struct {
	UserID         uint
	Assigned       int64
	Resolved       int64
	AvgResolutionH float64
	AvgReplyH      float64
}

func (r *AnalyticsRepository) ComputeAgentMetrics(tenantID uuid.UUID, userID uint) (AgentRawMetrics, error) {
	var row struct {
		Assigned       int64
		Resolved       int64
		AvgResolutionH float64
	}
	err := r.db.Model(&models.Ticket{}).
		Select(`
			COUNT(*) as assigned,
			COUNT(*) FILTER (WHERE status IN ('RESOLVED','CLOSED')) as resolved,
			COALESCE(AVG(EXTRACT(EPOCH FROM (updated_at - created_at)) / 3600)
			  FILTER (WHERE status IN ('RESOLVED','CLOSED')), 0) as avg_resolution_h
		`).
		Where("tenant_id = ? AND assigned_to = ?", tenantID, userID).
		Find(&row).Error
	if err != nil {
		return AgentRawMetrics{}, err
	}

	var replyRow struct{ AvgH float64 }
	_ = r.db.Model(&models.TicketComment{}).
		Select(`COALESCE(AVG(EXTRACT(EPOCH FROM (ticket_comments.created_at - t.created_at)) / 3600), 0) as avg_h`).
		Joins("JOIN tickets t ON t.id = ticket_comments.ticket_id").
		Where("ticket_comments.tenant_id = ? AND ticket_comments.user_id = ? AND t.assigned_to = ? AND ticket_comments.comment_type = 'PUBLIC'",
			tenantID, userID, userID).
		Find(&replyRow).Error

	return AgentRawMetrics{
		UserID: userID, Assigned: row.Assigned, Resolved: row.Resolved,
		AvgResolutionH: row.AvgResolutionH, AvgReplyH: replyRow.AvgH,
	}, nil
}

func (r *AnalyticsRepository) ComputeDailyTickets(tenantID uuid.UUID, date time.Time) (models.DailyTicketMetrics, error) {
	dayStart := truncateDay(date)
	dayEnd := dayStart.Add(24 * time.Hour)

	var row struct {
		Created  int
		Closed   int
		Resolved int
		Reopened int
		AvgRes   float64
		AvgAI    float64
	}

	err := r.db.Model(&models.Ticket{}).
		Select(`
			COUNT(*) FILTER (WHERE created_at BETWEEN ? AND ?)  as created,
			COUNT(*) FILTER (WHERE status = 'CLOSED'   AND updated_at BETWEEN ? AND ?) as closed,
			COUNT(*) FILTER (WHERE status = 'RESOLVED' AND updated_at BETWEEN ? AND ?) as resolved,
			COALESCE(AVG(EXTRACT(EPOCH FROM (updated_at - created_at)) / 3600)
			  FILTER (WHERE status IN ('RESOLVED','CLOSED') AND updated_at BETWEEN ? AND ?), 0) as avg_res,
			COALESCE(AVG(EXTRACT(EPOCH FROM (processed_at - created_at)))
			  FILTER (WHERE ai_processing_status = 'COMPLETED' AND processed_at BETWEEN ? AND ?), 0) as avg_ai
		`,
			dayStart, dayEnd,
			dayStart, dayEnd,
			dayStart, dayEnd,
			dayStart, dayEnd,
			dayStart, dayEnd,
		).Where("tenant_id = ?", tenantID).Find(&row).Error
	if err != nil {
		return models.DailyTicketMetrics{}, err
	}

	var frRow struct{ AvgH float64 }
	_ = r.db.Model(&models.TicketComment{}).
		Select(`COALESCE(AVG(EXTRACT(EPOCH FROM (tc.created_at - t.created_at)) / 3600), 0) as avg_h`).
		Table("ticket_comments tc").
		Joins("JOIN tickets t ON t.id = tc.ticket_id").
		Where(`tc.tenant_id = ? AND tc.created_at BETWEEN ? AND ?
			AND tc.id = (SELECT MIN(tc2.id) FROM ticket_comments tc2 WHERE tc2.ticket_id = tc.ticket_id AND tc2.comment_type = 'PUBLIC')`,
			tenantID, dayStart, dayEnd).
		Find(&frRow).Error

	return models.DailyTicketMetrics{
		TenantID:                 tenantID,
		Date:                     dayStart,
		TicketsCreated:           row.Created,
		TicketsClosed:            row.Closed,
		TicketsResolved:          row.Resolved,
		AverageResolutionTime:    row.AvgRes,
		AverageFirstResponseTime: frRow.AvgH,
		AverageAIProcessingTime:  row.AvgAI,
	}, nil
}

func (r *AnalyticsRepository) ComputeAIMetrics(tenantID uuid.UUID, date time.Time) (models.AIMetrics, error) {
	dayStart := truncateDay(date)
	dayEnd := dayStart.Add(24 * time.Hour)

	var analysisCount int64
	_ = r.db.Model(&models.Ticket{}).
		Where("tenant_id = ? AND ai_processing_status = 'COMPLETED' AND processed_at BETWEEN ? AND ?", tenantID, dayStart, dayEnd).
		Count(&analysisCount).Error

	s, err := r.SummariseAIReplies(tenantID, dayStart, dayEnd)
	if err != nil {
		return models.AIMetrics{}, err
	}
	approvalRate, editRate, rejectionRate, retryRate := calcRates(s)

	return models.AIMetrics{
		TenantID:              tenantID,
		Date:                  dayStart,
		AnalysisGenerated:     int(analysisCount),
		RepliesGenerated:      int(s.Total),
		AverageConfidence:     s.AvgConf,
		AverageGenerationTime: s.AvgGenMs,
		ApprovalRate:          approvalRate,
		EditRate:              editRate,
		RejectionRate:         rejectionRate,
		RetryRate:             retryRate,
	}, nil
}

// ─── Report CRUD ─────────────────────────────────────────────────────────────

func (r *AnalyticsRepository) CreateReport(report *models.Report) error {
	return r.db.Create(report).Error
}

func (r *AnalyticsRepository) FindReport(tenantID uuid.UUID, id uint) (*models.Report, error) {
	var report models.Report
	err := r.db.Preload("Generator").Where("tenant_id = ? AND id = ?", tenantID, id).First(&report).Error
	return &report, err
}

func (r *AnalyticsRepository) UpdateReport(report *models.Report) error {
	return r.db.Save(report).Error
}

func (r *AnalyticsRepository) ListReports(tenantID uuid.UUID, generatedBy *uint) ([]models.Report, error) {
	var reports []models.Report
	q := r.db.Where("tenant_id = ?", tenantID).Order("created_at DESC")
	if generatedBy != nil {
		q = q.Where("generated_by = ?", *generatedBy)
	}
	err := q.Find(&reports).Error
	return reports, err
}

func (r *AnalyticsRepository) ListOldReports(retentionDays int) ([]models.Report, error) {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	var reports []models.Report
	err := r.db.Where("created_at < ?", cutoff).Find(&reports).Error
	return reports, err
}

func (r *AnalyticsRepository) DeleteReport(id uint) error {
	return r.db.Delete(&models.Report{}, id).Error
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func truncateDay(t time.Time) time.Time {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func calcRates(s AIReplySummary) (approval, edit, rejection, retry float64) {
	if s.Total == 0 {
		return 0, 0, 0, 0
	}
	total := float64(s.Total)
	return round2(float64(s.Approved) / total * 100),
		round2(float64(s.Edited) / total * 100),
		round2(float64(s.Rejected) / total * 100),
		round2(float64(s.Retried) / total * 100)
}

func round2(f float64) float64 {
	return float64(int(f*100+0.5)) / 100
}
