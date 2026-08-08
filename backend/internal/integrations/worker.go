// Package integrations provides the background worker that polls TicketActivity
// rows and dispatches integration events to external providers.
package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	emailcrypto "github.com/ayush/supportiq/internal/email/crypto"
	"github.com/ayush/supportiq/internal/integrations/provider"
	"github.com/ayush/supportiq/internal/models"
	"github.com/ayush/supportiq/internal/repositories"
	"github.com/ayush/supportiq/internal/utils"
	"github.com/google/uuid"
)

const (
	defaultPollInterval = 30 * time.Second
	maxRetries          = 5
	eventBatchSize      = 50
)

// Worker polls TicketActivity rows, converts them to IntegrationEvents, and
// dispatches them to all enabled provider integrations.
type Worker struct {
	integrationRepo *repositories.IntegrationRepository
	activityRepo    *repositories.ActivityRepository
	ticketRepo      *repositories.TicketRepository
	tenantRepo      TenantIDLister
	registry        *Registry
	encryptionKey   string
	lastActivityID  uint
	pollInterval    time.Duration
}

// TenantIDLister is a minimal interface for listing active tenant IDs.
type TenantIDLister interface {
	AllActiveTenantIDs() ([]uuid.UUID, error)
}

func NewWorker(
	integrationRepo *repositories.IntegrationRepository,
	activityRepo *repositories.ActivityRepository,
	ticketRepo *repositories.TicketRepository,
	tenantRepo TenantIDLister,
	registry *Registry,
	encryptionKey string,
) *Worker {
	return &Worker{
		integrationRepo: integrationRepo,
		activityRepo:    activityRepo,
		ticketRepo:      ticketRepo,
		tenantRepo:      tenantRepo,
		registry:        registry,
		encryptionKey:   encryptionKey,
		pollInterval:    defaultPollInterval,
	}
}

func (w *Worker) Start(ctx context.Context) {
	// Seed the cursor to the current max activity ID so we only process
	// NEW activities after startup — prevents duplicate notifications on restart.
	if w.lastActivityID == 0 {
		w.lastActivityID = w.activityRepo.MaxID()
		utils.Logger.WithField("last_activity_id", w.lastActivityID).
			Info("Integration worker: seeded cursor from DB")
	}

	// Recover any events left in PROCESSING state from a previous crash.
	if err := w.integrationRepo.ResetStuckProcessingEvents(); err != nil {
		utils.Logger.WithError(err).Warn("Integration worker: could not reset stuck PROCESSING events")
	}

	activityTicker := time.NewTicker(w.pollInterval)
	eventTicker := time.NewTicker(w.pollInterval)
	defer activityTicker.Stop()
	defer eventTicker.Stop()

	utils.Logger.Info("Integration worker started")
	for {
		select {
		case <-ctx.Done():
			utils.Logger.Info("Integration worker stopped")
			return
		case <-activityTicker.C:
			w.pollActivities(ctx)
		case <-eventTicker.C:
			w.processEvents(ctx)
		}
	}
}

func (w *Worker) pollActivities(ctx context.Context) {
	tenantIDs, err := w.tenantRepo.AllActiveTenantIDs()
	if err != nil {
		utils.Logger.WithError(err).Error("Integration worker: failed to list tenants")
		return
	}

	for _, tenantID := range tenantIDs {
		activities, err := w.activityRepo.FindActivitiesSince(tenantID, w.lastActivityID, eventBatchSize)
		if err != nil {
			utils.Logger.WithError(err).Error("Integration worker: poll activities failed")
			continue
		}
		for _, act := range activities {
			if act.ID > w.lastActivityID {
				w.lastActivityID = act.ID
			}
			eventType := activityToEventType(act)
			if eventType == "" {
				continue
			}
			if err := w.createEventsForActivity(ctx, tenantID, act, eventType); err != nil {
				utils.Logger.WithError(err).WithField("activity_id", act.ID).
					Error("Integration worker: create events failed")
			}
		}
	}
}

func activityToEventType(act models.TicketActivity) string {
	switch act.ActivityType {
	case models.ActivityCreateTicket:
		return provider.EventTicketCreated
	case models.ActivityStatusChanged:
		if act.NewValue == string(models.TicketStatusClosed) {
			return provider.EventTicketClosed
		}
		return provider.EventTicketStatusChanged
	case models.ActivityAssignTicket:
		return provider.EventTicketAssigned
	case models.ActivityAIAnalysisCompleted:
		return provider.EventAIAnalysisComplete
	case models.ActivityReplyApproved:
		return provider.EventReplyApproved
	case models.ActivityEmailFailed:
		return provider.EventEmailFailed
	default:
		return ""
	}
}

func (w *Worker) createEventsForActivity(ctx context.Context, tenantID uuid.UUID, act models.TicketActivity, eventType string) error {
	integrations, err := w.integrationRepo.FindEnabled(tenantID)
	if err != nil {
		return err
	}

	ticket, err := w.ticketRepo.FindByIDUnscoped(act.TicketID)
	if err != nil {
		return fmt.Errorf("find ticket %s: %w", act.TicketID, err)
	}

	payload := map[string]interface{}{
		"ticket_id":      ticket.ID.String(),
		"ticket_number":  ticket.TicketNumber,
		"subject":        ticket.Subject,
		"priority":       string(ticket.Priority),
		"status":         string(ticket.Status),
		"customer_email": ticket.CustomerEmail,
		"customer_name":  ticket.CustomerName,
		"activity_id":    act.ID,
		"agent_id":       act.UserID,
	}
	payloadJSON, _ := json.Marshal(payload)

	for _, intg := range integrations {
		prov, err := w.buildProvider(intg)
		if err != nil {
			continue
		}
		supported := false
		for _, evtType := range prov.SupportedEvents() {
			if evtType == eventType {
				supported = true
				break
			}
		}
		if !supported {
			continue
		}

		// Skip if we already created an event for this activity+integration
		// (guards against replays if the cursor is ever lost).
		if w.integrationRepo.EventExistsForActivity(act.ID, intg.ID) {
			continue
		}

		event := &models.IntegrationEvent{
			TenantID:      tenantID,
			IntegrationID: intg.ID,
			ActivityID:    act.ID,
			EventType:     eventType,
			Payload:       string(payloadJSON),
			Status:        models.IntEventPending,
		}
		if err := w.integrationRepo.CreateEvent(event); err != nil {
			utils.Logger.WithError(err).WithField("integration_id", intg.ID).
				Error("Integration worker: store event")
		}
	}
	return nil
}

func (w *Worker) processEvents(ctx context.Context) {
	// ClaimPendingEvents atomically marks events as PROCESSING before returning
	// them — prevents two concurrent workers from dispatching the same event.
	events, err := w.integrationRepo.ClaimPendingEvents(eventBatchSize)
	if err != nil {
		utils.Logger.WithError(err).Error("Integration worker: claim pending events")
		return
	}
	for _, evt := range events {
		w.dispatchEvent(ctx, evt)
	}
}

func (w *Worker) dispatchEvent(ctx context.Context, evt models.IntegrationEvent) {
	intg, err := w.integrationRepo.FindByID(evt.TenantID, evt.IntegrationID)
	if err != nil || !intg.Enabled {
		_ = w.integrationRepo.UpdateEventStatus(evt.ID, models.IntEventDead, "integration not found or disabled")
		return
	}

	prov, err := w.buildProvider(*intg)
	if err != nil {
		_ = w.integrationRepo.UpdateEventStatus(evt.ID, models.IntEventFailed, err.Error())
		return
	}

	var payload map[string]interface{}
	_ = json.Unmarshal([]byte(evt.Payload), &payload)

	e := provider.Event{
		Type:          evt.EventType,
		TicketID:      strVal(payload, "ticket_id"),
		TicketNumber:  strVal(payload, "ticket_number"),
		Subject:       strVal(payload, "subject"),
		Priority:      strVal(payload, "priority"),
		Status:        strVal(payload, "status"),
		CustomerEmail: strVal(payload, "customer_email"),
	}

	dispatchCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	if notifyErr := prov.Notify(dispatchCtx, e); notifyErr != nil {
		_ = w.integrationRepo.IncrementEventRetry(evt.ID)
		if evt.RetryCount+1 >= maxRetries {
			_ = w.integrationRepo.MarkEventDead(evt.ID, notifyErr.Error())
		} else {
			_ = w.integrationRepo.UpdateEventStatus(evt.ID, models.IntEventFailed, notifyErr.Error())
		}
		return
	}

	_ = w.integrationRepo.UpdateEventStatus(evt.ID, models.IntEventProcessed, "")
}

func (w *Worker) buildProvider(intg models.Integration) (provider.Provider, error) {
	plaintext, err := emailcrypto.Decrypt(w.encryptionKey, intg.Configuration)
	if err != nil {
		return nil, fmt.Errorf("decrypt config: %w", err)
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal([]byte(plaintext), &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return w.registry.Build(string(intg.Provider), cfg)
}

func strVal(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
