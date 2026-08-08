package dto

import "time"

// ─── Filter ────────────────────────────────────────────────────────────────

// DateFilter holds the parsed time-range and dimension filters for analytics
// queries. It is constructed once per request from query-string parameters.
type DateFilter struct {
	StartDate time.Time
	EndDate   time.Time
	AgentID   *uint
	Priority  string
	Category  string
	Status    string
	Source    string
}

// GenerateReportRequest is the request body for POST /analytics/reports.
type GenerateReportRequest struct {
	Name       string `json:"name"        binding:"required"`
	ReportType string `json:"report_type" binding:"required"` // tickets | agents | ai | email
	Format     string `json:"format"      binding:"required"` // CSV | EXCEL | HTML
	Period     string `json:"period"`                         // today|yesterday|last7|last30|last90|custom
	StartDate  string `json:"start_date"`                     // YYYY-MM-DD  (required when period=custom)
	EndDate    string `json:"end_date"`                       // YYYY-MM-DD
	AgentID    *uint  `json:"agent_id"`
	Priority   string `json:"priority"`
	Category   string `json:"category"`
	Status     string `json:"status"`
	Source     string `json:"source"`
}

// ─── Overview ──────────────────────────────────────────────────────────────

// OverviewResponse is returned by GET /analytics/overview.
type OverviewResponse struct {
	TotalTickets        int64   `json:"total_tickets"`
	OpenTickets         int64   `json:"open_tickets"`
	InProgressTickets   int64   `json:"in_progress_tickets"`
	ResolvedTickets     int64   `json:"resolved_tickets"`
	ClosedTickets       int64   `json:"closed_tickets"`
	HighPriorityTickets int64   `json:"high_priority_tickets"`
	UrgentTickets       int64   `json:"urgent_tickets"`
	CreatedToday        int64   `json:"created_today"`
	ResolvedToday       int64   `json:"resolved_today"`
	AvgResolutionHours  float64 `json:"avg_resolution_hours"`
	AvgAIConfidence     float64 `json:"avg_ai_confidence"`
	AIApprovalRate      float64 `json:"ai_approval_rate"`
	QueuedJobs          int64   `json:"queued_jobs"`
	FailedJobs          int64   `json:"failed_jobs"`
	EmailsProcessedToday int64  `json:"emails_processed_today"`
}

// ─── Tickets ────────────────────────────────────────────────────────────────

// CountPair holds a label and its integer count (used for distributions).
type CountPair struct {
	Label string `json:"label"`
	Count int64  `json:"count"`
}

// TicketStatsResponse is returned by GET /analytics/tickets.
type TicketStatsResponse struct {
	TotalInRange      int64       `json:"total_in_range"`
	ByStatus          []CountPair `json:"by_status"`
	ByPriority        []CountPair `json:"by_priority"`
	ByCategory        []CountPair `json:"by_category"`
	BySource          []CountPair `json:"by_source"`
	ByHour            []CountPair `json:"by_hour"`
	DailyTrend        interface{} `json:"daily_trend"` // []DailyTicketMetrics
	AvgResolutionHours float64    `json:"avg_resolution_hours"`
}

// ─── Agents ─────────────────────────────────────────────────────────────────

// AgentRow is one agent in the leaderboard.
type AgentRow struct {
	UserID                uint    `json:"user_id"`
	Name                  string  `json:"name"`
	Email                 string  `json:"email"`
	TicketsAssigned       int     `json:"tickets_assigned"`
	TicketsResolved       int     `json:"tickets_resolved"`
	ActiveTickets         int64   `json:"active_tickets"`
	AverageResolutionTime float64 `json:"average_resolution_time"`
	AverageReplyTime      float64 `json:"average_reply_time"`
	LastCalculated        time.Time `json:"last_calculated"`
}

// AgentStatsResponse is returned by GET /analytics/agents.
type AgentStatsResponse struct {
	Agents     []AgentRow `json:"agents"`
	Leaderboard []AgentRow `json:"leaderboard"` // top 10 by resolved
}

// ─── AI ──────────────────────────────────────────────────────────────────────

// AIStatsResponse is returned by GET /analytics/ai.
type AIStatsResponse struct {
	TotalAnalyses      int64       `json:"total_analyses"`
	TotalReplies       int64       `json:"total_replies"`
	AvgConfidence      float64     `json:"avg_confidence"`
	AvgGenerationMs    float64     `json:"avg_generation_ms"`
	ApprovalRate       float64     `json:"approval_rate"`
	EditRate           float64     `json:"edit_rate"`
	RejectionRate      float64     `json:"rejection_rate"`
	RetryRate          float64     `json:"retry_rate"`
	FailureCount       int64       `json:"failure_count"`
	RetryCount         int64       `json:"retry_count"`
	TopCategories      []CountPair `json:"top_categories"`
	TopSentiments      []CountPair `json:"top_sentiments"`
	TopTags            []CountPair `json:"top_tags"`
	DailyTrend         interface{} `json:"daily_trend"` // []AIMetrics
}

// ─── Queue ───────────────────────────────────────────────────────────────────

// QueueStatsResponse is returned by GET /analytics/queues.
type QueueStatsResponse struct {
	TotalQueued       int64   `json:"total_queued"`
	TotalProcessing   int64   `json:"total_processing"`
	TotalCompleted    int64   `json:"total_completed"`
	TotalFailed       int64   `json:"total_failed"`
	TotalDead         int64   `json:"total_dead"`
	TotalRetrying     int64   `json:"total_retrying"`
	AvgQueueSeconds   float64 `json:"avg_queue_seconds"`
	ByJobType         []CountPair `json:"by_job_type"`
}

// ─── Email ───────────────────────────────────────────────────────────────────

// EmailStatsResponse is returned by GET /analytics/email.
type EmailStatsResponse struct {
	ReceivedTotal     int64   `json:"received_total"`
	SentTotal         int64   `json:"sent_total"`
	FailedTotal       int64   `json:"failed_total"`
	QueuedTotal       int64   `json:"queued_total"`
	AvgDeliverySeconds float64 `json:"avg_delivery_seconds"`
	ByStatus          []CountPair `json:"by_status"`
}

// ─── Trends ──────────────────────────────────────────────────────────────────

// TrendPoint is a single data point in a time-series.
type TrendPoint struct {
	Date            string  `json:"date"`
	TicketsCreated  int     `json:"tickets_created"`
	TicketsClosed   int     `json:"tickets_closed"`
	TicketsResolved int     `json:"tickets_resolved"`
	AIAnalyses      int     `json:"ai_analyses"`
	AIReplies       int     `json:"ai_replies"`
	AvgConfidence   float64 `json:"avg_confidence"`
	EmailsReceived  int     `json:"emails_received"`
	EmailsSent      int     `json:"emails_sent"`
}

// TrendsResponse is returned by GET /analytics/trends.
type TrendsResponse struct {
	Points []TrendPoint `json:"points"`
}

// ─── Reports ─────────────────────────────────────────────────────────────────

// ReportResponse is the API representation of a Report record.
type ReportResponse struct {
	ID           uint      `json:"id"`
	Name         string    `json:"name"`
	ReportType   string    `json:"report_type"`
	Format       string    `json:"format"`
	Status       string    `json:"status"`
	FileSize     int64     `json:"file_size"`
	Filters      string    `json:"filters"`
	GeneratedBy  uint      `json:"generated_by"`
	ErrorMessage string    `json:"error_message,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
}
