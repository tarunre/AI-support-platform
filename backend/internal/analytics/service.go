package analytics

import (
	"math"
	"time"

	"github.com/ayush/supportiq/internal/dto"
	"github.com/ayush/supportiq/internal/models"
	"github.com/google/uuid"
)

// Service orchestrates analytics queries and returns frontend-ready DTOs.
type Service struct {
	repo       *AnalyticsRepository
	aggregator *Aggregator
}

func NewService(repo *AnalyticsRepository, aggregator *Aggregator) *Service {
	return &Service{repo: repo, aggregator: aggregator}
}

func (s *Service) TriggerAggregation() {
	s.aggregator.RunAll()
}

// ─── Overview ────────────────────────────────────────────────────────────────

func (s *Service) GetOverview(tenantID uuid.UUID) (*dto.OverviewResponse, error) {
	statusMap, err := s.repo.CountTicketsByStatusSnapshot(tenantID)
	if err != nil {
		return nil, err
	}

	createdToday, _ := s.repo.CountTicketsCreatedToday(tenantID)
	resolvedToday, _ := s.repo.CountTicketsResolvedToday(tenantID)

	end := time.Now()
	start := end.AddDate(0, 0, -30)

	avgRes, _ := s.repo.AvgResolutionHours(tenantID, start, end)
	avgConf, _ := s.repo.AvgAIConfidence(tenantID, start, end)

	aiSum, _ := s.repo.SummariseAIReplies(tenantID, start, end)
	approvalRate := 0.0
	if aiSum.Total > 0 {
		approvalRate = round2(float64(aiSum.Approved) / float64(aiSum.Total) * 100)
	}

	queueSnap, _ := s.repo.SnapshotQueue(tenantID)
	emailsToday, _ := s.repo.CountEmailsProcessedToday(tenantID)

	return &dto.OverviewResponse{
		TotalTickets:         sumStatusMap(statusMap),
		OpenTickets:          statusMap["OPEN"],
		InProgressTickets:    statusMap["IN_PROGRESS"],
		ResolvedTickets:      statusMap["RESOLVED"],
		ClosedTickets:        statusMap["CLOSED"],
		HighPriorityTickets:  s.countByPriority(tenantID, "HIGH"),
		UrgentTickets:        s.countByPriority(tenantID, "URGENT"),
		CreatedToday:         createdToday,
		ResolvedToday:        resolvedToday,
		AvgResolutionHours:   roundF(avgRes, 2),
		AvgAIConfidence:      roundF(avgConf, 1),
		AIApprovalRate:       approvalRate,
		QueuedJobs:           queueSnap.Queued,
		FailedJobs:           queueSnap.Failed,
		EmailsProcessedToday: emailsToday,
	}, nil
}

func (s *Service) countByPriority(tenantID uuid.UUID, p string) int64 {
	var count int64
	s.repo.db.Model(&models.Ticket{}).
		Where("tenant_id = ? AND priority = ? AND status NOT IN ('RESOLVED','CLOSED')", tenantID, p).
		Count(&count)
	return count
}

// ─── Ticket Analytics ─────────────────────────────────────────────────────────

func (s *Service) GetTicketStats(tenantID uuid.UUID, f dto.DateFilter) (*dto.TicketStatsResponse, error) {
	total, err := s.repo.CountTicketsInRange(tenantID, f.StartDate, f.EndDate)
	if err != nil {
		return nil, err
	}

	byPriority, _ := s.repo.CountTicketsByPriority(tenantID, f.StartDate, f.EndDate)
	byCategory, _ := s.repo.CountTicketsByCategory(tenantID, f.StartDate, f.EndDate)
	bySource, _ := s.repo.CountTicketsBySource(tenantID, f.StartDate, f.EndDate)
	byHour, _ := s.repo.CountTicketsByHour(tenantID, time.Now())
	avgRes, _ := s.repo.AvgResolutionHours(tenantID, f.StartDate, f.EndDate)
	dailyMetrics, _ := s.repo.GetDailyTicketMetrics(tenantID, f.StartDate, f.EndDate)
	statusSnap, _ := s.repo.CountTicketsByStatusSnapshot(tenantID)

	return &dto.TicketStatsResponse{
		TotalInRange:       total,
		ByStatus:           statusMapToPairs(statusSnap),
		ByPriority:         labelCountsToPairs(byPriority),
		ByCategory:         labelCountsToPairs(byCategory),
		BySource:           labelCountsToPairs(bySource),
		ByHour:             labelCountsToPairs(byHour),
		DailyTrend:         dailyMetrics,
		AvgResolutionHours: roundF(avgRes, 2),
	}, nil
}

// ─── Agent Analytics ──────────────────────────────────────────────────────────

func (s *Service) GetAgentStats(tenantID uuid.UUID) (*dto.AgentStatsResponse, error) {
	rows, err := s.repo.GetAllAgentMetrics(tenantID)
	if err != nil {
		return nil, err
	}

	agents := make([]dto.AgentRow, 0, len(rows))
	for _, m := range rows {
		active, _ := s.repo.CountActiveTicketsForAgent(tenantID, m.UserID)
		name, email := userInfo(m.User)
		agents = append(agents, dto.AgentRow{
			UserID:                m.UserID,
			Name:                  name,
			Email:                 email,
			TicketsAssigned:       m.TicketsAssigned,
			TicketsResolved:       m.TicketsResolved,
			ActiveTickets:         active,
			AverageResolutionTime: roundF(m.AverageResolutionTime, 2),
			AverageReplyTime:      roundF(m.AverageReplyTime, 2),
			LastCalculated:        m.LastCalculated,
		})
	}

	top := agents
	if len(top) > 10 {
		top = top[:10]
	}

	return &dto.AgentStatsResponse{
		Agents:      agents,
		Leaderboard: top,
	}, nil
}

func (s *Service) GetPersonalAgentStats(tenantID uuid.UUID, userID uint) (*dto.AgentStatsResponse, error) {
	m, err := s.repo.GetAgentMetricsByUserID(tenantID, userID)
	if err != nil {
		return nil, err
	}
	active, _ := s.repo.CountActiveTicketsForAgent(tenantID, userID)
	name, email := userInfo(m.User)
	row := dto.AgentRow{
		UserID:                m.UserID,
		Name:                  name,
		Email:                 email,
		TicketsAssigned:       m.TicketsAssigned,
		TicketsResolved:       m.TicketsResolved,
		ActiveTickets:         active,
		AverageResolutionTime: roundF(m.AverageResolutionTime, 2),
		AverageReplyTime:      roundF(m.AverageReplyTime, 2),
		LastCalculated:        m.LastCalculated,
	}
	return &dto.AgentStatsResponse{
		Agents:      []dto.AgentRow{row},
		Leaderboard: []dto.AgentRow{},
	}, nil
}

// ─── AI Analytics ─────────────────────────────────────────────────────────────

func (s *Service) GetAIStats(tenantID uuid.UUID, f dto.DateFilter) (*dto.AIStatsResponse, error) {
	aiSum, err := s.repo.SummariseAIReplies(tenantID, f.StartDate, f.EndDate)
	if err != nil {
		return nil, err
	}

	totalAnalyses, _ := s.repo.CountAIAnalysesCompleted(tenantID, f.StartDate, f.EndDate)
	failures, _ := s.repo.CountAIFailures(tenantID, f.StartDate, f.EndDate)
	avgConf, _ := s.repo.AvgAIConfidence(tenantID, f.StartDate, f.EndDate)
	topCats, _ := s.repo.TopAICategories(tenantID, f.StartDate, f.EndDate, 10)
	topSents, _ := s.repo.TopAISentiments(tenantID, f.StartDate, f.EndDate)
	aiMetrics, _ := s.repo.GetAIMetrics(tenantID, f.StartDate, f.EndDate)

	approvalRate, editRate, rejectionRate, retryRate := calcRates(aiSum)

	return &dto.AIStatsResponse{
		TotalAnalyses:   totalAnalyses,
		TotalReplies:    aiSum.Total,
		AvgConfidence:   roundF(avgConf, 1),
		AvgGenerationMs: roundF(aiSum.AvgGenMs, 0),
		ApprovalRate:    approvalRate,
		EditRate:        editRate,
		RejectionRate:   rejectionRate,
		RetryRate:       retryRate,
		FailureCount:    failures,
		RetryCount:      aiSum.Retried,
		TopCategories:   labelCountsToPairs(topCats),
		TopSentiments:   labelCountsToPairs(topSents),
		TopTags:         []dto.CountPair{},
		DailyTrend:      aiMetrics,
	}, nil
}

// ─── Queue Analytics ──────────────────────────────────────────────────────────

func (s *Service) GetQueueStats(tenantID uuid.UUID) (*dto.QueueStatsResponse, error) {
	snap, err := s.repo.SnapshotQueue(tenantID)
	if err != nil {
		return nil, err
	}
	byType, _ := s.repo.CountJobsByType(tenantID)

	return &dto.QueueStatsResponse{
		TotalQueued:     snap.Queued,
		TotalProcessing: snap.Processing,
		TotalCompleted:  snap.Completed,
		TotalFailed:     snap.Failed,
		TotalDead:       snap.Dead,
		TotalRetrying:   snap.Retrying,
		AvgQueueSeconds: roundF(snap.AvgQueueSec, 1),
		ByJobType:       labelCountsToPairs(byType),
	}, nil
}

// ─── Email Analytics ──────────────────────────────────────────────────────────

func (s *Service) GetEmailStats(tenantID uuid.UUID, f dto.DateFilter) (*dto.EmailStatsResponse, error) {
	snap, err := s.repo.SnapshotEmail(tenantID, f.StartDate, f.EndDate)
	if err != nil {
		return nil, err
	}

	var rows []struct {
		Direction string
		Status    string
		Count     int64
	}
	s.repo.db.Model(&models.EmailMessage{}).
		Select("direction, status, COUNT(*) as count").
		Where("tenant_id = ? AND created_at BETWEEN ? AND ?", tenantID, f.StartDate, f.EndDate).
		Group("direction, status").
		Find(&rows)

	byStatus := make([]dto.CountPair, 0, len(rows))
	for _, row := range rows {
		byStatus = append(byStatus, dto.CountPair{
			Label: row.Direction + "_" + row.Status,
			Count: row.Count,
		})
	}

	return &dto.EmailStatsResponse{
		ReceivedTotal:      snap.Received,
		SentTotal:          snap.Sent,
		FailedTotal:        snap.Failed,
		QueuedTotal:        snap.Queued,
		AvgDeliverySeconds: roundF(snap.AvgDelivSec, 1),
		ByStatus:           byStatus,
	}, nil
}

// ─── Trends ───────────────────────────────────────────────────────────────────

func (s *Service) GetTrends(tenantID uuid.UUID, f dto.DateFilter) (*dto.TrendsResponse, error) {
	dailyMetrics, err := s.repo.GetDailyTicketMetrics(tenantID, f.StartDate, f.EndDate)
	if err != nil {
		return nil, err
	}
	aiMetrics, _ := s.repo.GetAIMetrics(tenantID, f.StartDate, f.EndDate)

	aiByDate := make(map[string]models.AIMetrics, len(aiMetrics))
	for _, m := range aiMetrics {
		aiByDate[m.Date.Format("2006-01-02")] = m
	}

	emailByDate := s.buildEmailByDate(tenantID, f.StartDate, f.EndDate)

	points := make([]dto.TrendPoint, 0, len(dailyMetrics))
	for _, dm := range dailyMetrics {
		key := dm.Date.Format("2006-01-02")
		ai := aiByDate[key]
		em := emailByDate[key]
		points = append(points, dto.TrendPoint{
			Date:            key,
			TicketsCreated:  dm.TicketsCreated,
			TicketsClosed:   dm.TicketsClosed,
			TicketsResolved: dm.TicketsResolved,
			AIAnalyses:      ai.AnalysisGenerated,
			AIReplies:       ai.RepliesGenerated,
			AvgConfidence:   roundF(ai.AverageConfidence, 1),
			EmailsReceived:  em.received,
			EmailsSent:      em.sent,
		})
	}

	return &dto.TrendsResponse{Points: points}, nil
}

type emailDayCount struct{ received, sent int }

func (s *Service) buildEmailByDate(tenantID uuid.UUID, start, end time.Time) map[string]emailDayCount {
	var rows []struct {
		Day       string
		Direction string
		Count     int
	}
	s.repo.db.Model(&models.EmailMessage{}).
		Select("TO_CHAR(DATE(created_at), 'YYYY-MM-DD') as day, direction, COUNT(*) as count").
		Where("tenant_id = ? AND created_at BETWEEN ? AND ?", tenantID, start, end).
		Group("day, direction").
		Find(&rows)

	m := make(map[string]emailDayCount)
	for _, row := range rows {
		e := m[row.Day]
		if row.Direction == "INBOUND" {
			e.received += row.Count
		} else {
			e.sent += row.Count
		}
		m[row.Day] = e
	}
	return m
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func sumStatusMap(m map[string]int64) int64 {
	var total int64
	for _, v := range m {
		total += v
	}
	return total
}

func statusMapToPairs(m map[string]int64) []dto.CountPair {
	pairs := make([]dto.CountPair, 0, len(m))
	for k, v := range m {
		pairs = append(pairs, dto.CountPair{Label: k, Count: v})
	}
	return pairs
}

func labelCountsToPairs(rows []LabelCount) []dto.CountPair {
	pairs := make([]dto.CountPair, len(rows))
	for i, r := range rows {
		pairs[i] = dto.CountPair{Label: r.Label, Count: r.Count}
	}
	return pairs
}

func userInfo(u *models.User) (name, email string) {
	if u == nil {
		return "Unknown", ""
	}
	return u.Name, u.Email
}

func roundF(f float64, decimals int) float64 {
	pow := math.Pow(10, float64(decimals))
	return math.Round(f*pow) / pow
}
