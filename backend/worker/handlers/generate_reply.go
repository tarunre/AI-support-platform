package handlers

import (
	"context"
	"fmt"

	"github.com/ayush/supportiq/internal/queue"
	"github.com/ayush/supportiq/internal/services"
	"github.com/ayush/supportiq/internal/utils"
	"github.com/google/uuid"
)

// GenerateReplyHandler processes GENERATE_REPLY and REGENERATE_REPLY jobs.
type GenerateReplyHandler struct {
	replySvc *services.ReplyService
}

func NewGenerateReplyHandler(replySvc *services.ReplyService) *GenerateReplyHandler {
	return &GenerateReplyHandler{replySvc: replySvc}
}

// Handle generates an AI reply for the given ticket job.
func (h *GenerateReplyHandler) Handle(ctx context.Context, job queue.Job) error {
	ticketID, err := uuid.Parse(job.TicketID)
	if err != nil {
		return fmt.Errorf("invalid ticket UUID %q: %w", job.TicketID, err)
	}

	log := utils.Logger.
		WithField("job_type", job.Type).
		WithField("ticket_id", ticketID)

	log.Info("Reply Generation: Starting")

	tenantID, _ := uuid.Parse(job.TenantID)

	if job.Type == "REGENERATE_REPLY" {
		_, err = h.replySvc.Regenerate(ctx, tenantID, ticketID, job.UserID)
	} else {
		_, err = h.replySvc.Generate(ctx, tenantID, ticketID, job.UserID)
	}

	if err != nil {
		return fmt.Errorf("reply generation failed: %w", err)
	}

	log.Info("Reply Generation: Completed successfully")
	return nil
}
