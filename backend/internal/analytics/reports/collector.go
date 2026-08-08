package reports

import (
	"fmt"
	"time"

	"github.com/ayush/supportiq/internal/dto"
	"github.com/ayush/supportiq/internal/models"
	"gorm.io/gorm"
)

// DataCollector implements the CollectData interface used by Service.Schedule.
// It queries the database to produce tabular rows for each report type.
type DataCollector struct {
	db *gorm.DB
}

// NewDataCollector creates a DataCollector.
func NewDataCollector(db *gorm.DB) *DataCollector {
	return &DataCollector{db: db}
}

// CollectData dispatches to the correct table builder based on reportType.
func (c *DataCollector) CollectData(reportType string, f dto.DateFilter) (*ReportData, error) {
	switch reportType {
	case "tickets":
		return c.collectTickets(f)
	case "agents":
		return c.collectAgents(f)
	case "ai":
		return c.collectAI(f)
	case "email":
		return c.collectEmail(f)
	default:
		return nil, fmt.Errorf("unknown report type: %s", reportType)
	}
}

// ─── Tickets report ──────────────────────────────────────────────────────────

func (c *DataCollector) collectTickets(f dto.DateFilter) (*ReportData, error) {
	var tickets []models.Ticket
	q := c.db.Preload("Assignee").
		Where("created_at BETWEEN ? AND ?", f.StartDate, f.EndDate)
	if f.Priority != "" {
		q = q.Where("priority = ?", f.Priority)
	}
	if f.Category != "" {
		q = q.Where("category = ?", f.Category)
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.Source != "" {
		q = q.Where("source = ?", f.Source)
	}
	if f.AgentID != nil {
		q = q.Where("assigned_to = ?", *f.AgentID)
	}
	if err := q.Find(&tickets).Error; err != nil {
		return nil, err
	}

	headers := []string{
		"Ticket Number", "Subject", "Status", "Priority", "Category",
		"Source", "Customer", "Assigned To", "AI Confidence", "Created At", "Updated At",
	}
	rows := make([][]string, 0, len(tickets))
	for _, t := range tickets {
		assignee := ""
		if t.Assignee != nil {
			assignee = t.Assignee.Name
		}
		conf := ""
		if t.AIConfidence != nil {
			conf = fmt.Sprintf("%d%%", *t.AIConfidence)
		}
		rows = append(rows, []string{
			t.TicketNumber, t.Subject, string(t.Status), string(t.Priority),
			string(t.Category), string(t.Source), t.CustomerName,
			assignee, conf,
			t.CreatedAt.Format(time.RFC3339),
			t.UpdatedAt.Format(time.RFC3339),
		})
	}

	return &ReportData{
		Title:   "Tickets Report",
		Headers: headers,
		Rows:    rows,
		Meta: map[string]string{
			"From":  f.StartDate.Format("2006-01-02"),
			"To":    f.EndDate.Format("2006-01-02"),
			"Total": fmt.Sprintf("%d", len(tickets)),
		},
	}, nil
}

// ─── Agents report ───────────────────────────────────────────────────────────

func (c *DataCollector) collectAgents(f dto.DateFilter) (*ReportData, error) {
	var metrics []models.AgentMetrics
	if err := c.db.Preload("User").Find(&metrics).Error; err != nil {
		return nil, err
	}

	headers := []string{
		"Agent Name", "Email", "Tickets Assigned", "Tickets Resolved",
		"Avg Resolution (h)", "Avg Reply Time (h)", "Last Calculated",
	}
	rows := make([][]string, 0, len(metrics))
	for _, m := range metrics {
		name, email := "", ""
		if m.User != nil {
			name = m.User.Name
			email = m.User.Email
		}
		rows = append(rows, []string{
			name, email,
			fmt.Sprintf("%d", m.TicketsAssigned),
			fmt.Sprintf("%d", m.TicketsResolved),
			fmt.Sprintf("%.2f", m.AverageResolutionTime),
			fmt.Sprintf("%.2f", m.AverageReplyTime),
			m.LastCalculated.Format("2006-01-02 15:04:05"),
		})
	}

	return &ReportData{
		Title:   "Agent Performance Report",
		Headers: headers,
		Rows:    rows,
		Meta: map[string]string{
			"Total Agents": fmt.Sprintf("%d", len(metrics)),
			"Generated":    time.Now().Format("2006-01-02"),
		},
	}, nil
}

// ─── AI report ───────────────────────────────────────────────────────────────

func (c *DataCollector) collectAI(f dto.DateFilter) (*ReportData, error) {
	var metrics []models.AIMetrics
	if err := c.db.Where("date BETWEEN ? AND ?", f.StartDate, f.EndDate).
		Order("date ASC").Find(&metrics).Error; err != nil {
		return nil, err
	}

	headers := []string{
		"Date", "Analyses", "Replies", "Avg Confidence", "Avg Gen (ms)",
		"Approval %", "Edit %", "Rejection %", "Retry %",
	}
	rows := make([][]string, 0, len(metrics))
	for _, m := range metrics {
		rows = append(rows, []string{
			m.Date.Format("2006-01-02"),
			fmt.Sprintf("%d", m.AnalysisGenerated),
			fmt.Sprintf("%d", m.RepliesGenerated),
			fmt.Sprintf("%.1f", m.AverageConfidence),
			fmt.Sprintf("%.0f", m.AverageGenerationTime),
			fmt.Sprintf("%.1f", m.ApprovalRate),
			fmt.Sprintf("%.1f", m.EditRate),
			fmt.Sprintf("%.1f", m.RejectionRate),
			fmt.Sprintf("%.1f", m.RetryRate),
		})
	}

	return &ReportData{
		Title:   "AI Performance Report",
		Headers: headers,
		Rows:    rows,
		Meta: map[string]string{
			"From": f.StartDate.Format("2006-01-02"),
			"To":   f.EndDate.Format("2006-01-02"),
		},
	}, nil
}

// ─── Email report ────────────────────────────────────────────────────────────

func (c *DataCollector) collectEmail(f dto.DateFilter) (*ReportData, error) {
	var messages []models.EmailMessage
	if err := c.db.Where("created_at BETWEEN ? AND ?", f.StartDate, f.EndDate).
		Order("created_at ASC").Find(&messages).Error; err != nil {
		return nil, err
	}

	headers := []string{
		"Direction", "Status", "Subject", "Sender", "Recipient",
		"Account ID", "Retry Count", "Created At", "Sent At",
	}
	rows := make([][]string, 0, len(messages))
	for _, m := range messages {
		sentAt := ""
		if m.SentAt != nil {
			sentAt = m.SentAt.Format(time.RFC3339)
		}
		rows = append(rows, []string{
			string(m.Direction), string(m.Status), m.Subject,
			m.Sender, m.Recipient,
			fmt.Sprintf("%d", m.AccountID),
			fmt.Sprintf("%d", m.RetryCount),
			m.CreatedAt.Format(time.RFC3339),
			sentAt,
		})
	}

	return &ReportData{
		Title:   "Email Activity Report",
		Headers: headers,
		Rows:    rows,
		Meta: map[string]string{
			"From":  f.StartDate.Format("2006-01-02"),
			"To":    f.EndDate.Format("2006-01-02"),
			"Total": fmt.Sprintf("%d", len(messages)),
		},
	}, nil
}
