package workers

import (
	"context"
	"time"

	"github.com/ayush/supportiq/internal/services"
	"github.com/ayush/supportiq/internal/utils"
)

// StartOutboundWorker processes the outbound email queue on a tick, retrying
// failed messages after a back-off period.
func StartOutboundWorker(ctx context.Context, emailSvc *services.EmailService, pollInterval time.Duration, maxRetries int) {
	utils.Logger.WithField("interval", pollInterval).Info("OutboundWorker: started")

	tick := time.NewTicker(pollInterval)
	retryTick := time.NewTicker(pollInterval * 5) // retry check every 5× interval
	defer tick.Stop()
	defer retryTick.Stop()

	for {
		select {
		case <-ctx.Done():
			utils.Logger.Info("OutboundWorker: stopping")
			return
		case <-tick.C:
			emailSvc.ProcessQueuedOutbound(ctx)
		case <-retryTick.C:
			emailSvc.RetryFailedOutbound(ctx, maxRetries)
		}
	}
}
