package analytics

import (
	"time"

	"github.com/ayush/supportiq/internal/models"
	"github.com/ayush/supportiq/internal/utils"
	"github.com/google/uuid"
)

// Aggregator computes and persists pre-aggregated metrics rows.
type Aggregator struct {
	repo        *AnalyticsRepository
	tenantRepo  TenantLister
}

// TenantLister is a minimal interface so the aggregator can list all active tenants
// without importing the full tenant repository package (avoids circular deps).
type TenantLister interface {
	AllActiveTenantIDs() ([]uuid.UUID, error)
}

func NewAggregator(repo *AnalyticsRepository, tenantRepo TenantLister) *Aggregator {
	return &Aggregator{repo: repo, tenantRepo: tenantRepo}
}

// RunAll aggregates tickets, AI, and agent metrics for all active tenants.
func (a *Aggregator) RunAll() {
	tenantIDs, err := a.tenantRepo.AllActiveTenantIDs()
	if err != nil {
		utils.Logger.WithError(err).Error("Aggregator: failed to list tenants")
		return
	}

	today := truncateDay(time.Now())
	yesterday := today.AddDate(0, 0, -1)

	for _, tenantID := range tenantIDs {
		a.AggregateDailyTickets(tenantID, today)
		a.AggregateDailyTickets(tenantID, yesterday)
		a.AggregateAIMetrics(tenantID, today)
		a.AggregateAIMetrics(tenantID, yesterday)
		a.AggregateAllAgents(tenantID)
	}
}

func (a *Aggregator) AggregateDailyTickets(tenantID uuid.UUID, date time.Time) {
	m, err := a.repo.ComputeDailyTickets(tenantID, date)
	if err != nil {
		utils.Logger.WithError(err).WithField("date", date).
			Error("Aggregator: failed to compute daily ticket metrics")
		return
	}
	if err := a.repo.UpsertDailyTicketMetrics(&m); err != nil {
		utils.Logger.WithError(err).WithField("date", date).
			Error("Aggregator: failed to upsert daily ticket metrics")
	}
}

func (a *Aggregator) AggregateAIMetrics(tenantID uuid.UUID, date time.Time) {
	m, err := a.repo.ComputeAIMetrics(tenantID, date)
	if err != nil {
		utils.Logger.WithError(err).WithField("date", date).
			Error("Aggregator: failed to compute AI metrics")
		return
	}
	if err := a.repo.UpsertAIMetrics(&m); err != nil {
		utils.Logger.WithError(err).WithField("date", date).
			Error("Aggregator: failed to upsert AI metrics")
	}
}

func (a *Aggregator) AggregateAllAgents(tenantID uuid.UUID) {
	agents, err := a.repo.AllSupportAgents(tenantID)
	if err != nil {
		utils.Logger.WithError(err).Error("Aggregator: failed to fetch agents")
		return
	}
	for _, agent := range agents {
		a.aggregateAgent(tenantID, agent)
	}
}

func (a *Aggregator) aggregateAgent(tenantID uuid.UUID, agent models.User) {
	raw, err := a.repo.ComputeAgentMetrics(tenantID, agent.ID)
	if err != nil {
		utils.Logger.WithError(err).WithField("userID", agent.ID).
			Error("Aggregator: failed to compute agent metrics")
		return
	}
	m := &models.AgentMetrics{
		TenantID:              tenantID,
		UserID:                agent.ID,
		TicketsAssigned:       int(raw.Assigned),
		TicketsResolved:       int(raw.Resolved),
		AverageResolutionTime: raw.AvgResolutionH,
		AverageReplyTime:      raw.AvgReplyH,
		LastCalculated:        time.Now().UTC(),
	}
	if err := a.repo.UpsertAgentMetrics(m); err != nil {
		utils.Logger.WithError(err).WithField("userID", agent.ID).
			Error("Aggregator: failed to upsert agent metrics")
	}
}
