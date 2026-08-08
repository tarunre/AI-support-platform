package analytics

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ayush/supportiq/internal/utils"
	appws "github.com/ayush/supportiq/internal/websocket"
)

// Scheduler runs the aggregator on a fixed interval and broadcasts a WebSocket
// event after every successful cycle so dashboard clients can refresh.
type Scheduler struct {
	aggregator *Aggregator
	service    *Service
	tenantRepo TenantLister
	wsHub      *appws.Hub
	interval   time.Duration
}

// NewScheduler creates a Scheduler.
func NewScheduler(agg *Aggregator, svc *Service, hub *appws.Hub, interval time.Duration) *Scheduler {
	return &Scheduler{
		aggregator: agg,
		service:    svc,
		tenantRepo: agg.tenantRepo,
		wsHub:      hub,
		interval:   interval,
	}
}

// Start runs the aggregation loop in the calling goroutine.
// It blocks until ctx is cancelled.
func (s *Scheduler) Start(ctx context.Context) {
	// Run once at startup so the tables are populated immediately.
	s.runCycle()

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.runCycle()
		}
	}
}

func (s *Scheduler) runCycle() {
	utils.Logger.Info("Analytics: running aggregation cycle")
	s.aggregator.RunAll()
	utils.Logger.Info("Analytics: aggregation cycle complete")
	s.broadcastRefresh()
}

type analyticsRefreshEvent struct {
	Type string `json:"type"`
	At   string `json:"at"`
}

func (s *Scheduler) broadcastRefresh() {
	if s.wsHub == nil {
		return
	}
	payload := analyticsRefreshEvent{
		Type: "ANALYTICS_REFRESH",
		At:   time.Now().UTC().Format(time.RFC3339),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	// Broadcast only to users of each active tenant.
	if s.tenantRepo != nil {
		if tenantIDs, err := s.tenantRepo.AllActiveTenantIDs(); err == nil {
			for _, tid := range tenantIDs {
				s.wsHub.BroadcastToTenantRaw(tid, data)
			}
			return
		}
	}
	// Fallback: global broadcast (single-tenant or tenant list unavailable).
	s.wsHub.BroadcastRaw(data)
}
