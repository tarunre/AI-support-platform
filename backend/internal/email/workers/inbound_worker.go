// Package workers contains background goroutines for email processing.
package workers

import (
	"context"
	"time"

	"github.com/ayush/supportiq/internal/repositories"
	"github.com/ayush/supportiq/internal/services"
	"github.com/ayush/supportiq/internal/utils"
)

// StartInboundWorker polls all active IMAP mailboxes on every pollInterval tick.
// Runs in the caller's goroutine; returns when ctx is cancelled.
func StartInboundWorker(
	ctx context.Context,
	accountRepo *repositories.EmailAccountRepository,
	emailSvc *services.EmailService,
	accountSvc *services.EmailAccountService,
	pollInterval time.Duration,
) {
	utils.Logger.WithField("interval", pollInterval).Info("InboundWorker: started")

	doPoll := func() {
		accounts, err := accountRepo.FindActive()
		if err != nil {
			utils.Logger.WithError(err).Warn("InboundWorker: load accounts failed")
			return
		}

		for i := range accounts {
			account := &accounts[i]
			if account.IMAPHost == "" {
				continue
			}

			receiver, err := accountSvc.BuildReceiver(account)
			if err != nil {
				utils.Logger.WithError(err).
					WithField("account", account.EmailAddress).
					Warn("InboundWorker: build receiver failed")
				continue
			}

			parsed, err := receiver.FetchUnread(ctx)
			if err != nil {
				utils.Logger.WithError(err).
					WithField("account", account.EmailAddress).
					Warn("InboundWorker: IMAP fetch failed")
				continue
			}

			// Exact-time filter: skip any email whose Date header is before
			// last_sync_at. This prevents processing old inbox emails when an
			// account is first added (IMAP SINCE is day-granularity only).
			var toProcess []struct{ idx int }
			for j := range parsed {
				if account.LastSyncAt != nil && !parsed[j].Date.IsZero() &&
					parsed[j].Date.Before(*account.LastSyncAt) {
					// Email arrived before we started monitoring — skip silently
					if parsed[j].UID > 0 {
						_ = receiver.MarkSeen(ctx, parsed[j].UID)
					}
					continue
				}
				toProcess = append(toProcess, struct{ idx int }{j})
			}

			processed := 0
			for _, item := range toProcess {
				j := item.idx
				if err := emailSvc.ProcessInbound(ctx, account, &parsed[j]); err != nil {
					utils.Logger.WithError(err).
						WithField("message_id", parsed[j].MessageID).
						Warn("InboundWorker: process inbound failed")
					continue
				}
				processed++

				// Mark as seen only after successful processing
				if parsed[j].UID > 0 {
					if err := receiver.MarkSeen(ctx, parsed[j].UID); err != nil {
						utils.Logger.WithError(err).
							WithField("uid", parsed[j].UID).
							Warn("InboundWorker: mark seen failed")
					}
				}
			}

			// Update last_sync_at — only advance forward, never go back.
			// This preserves a manually-set baseline time.
			now := time.Now()
			if account.LastSyncAt == nil || now.After(*account.LastSyncAt) {
				account.LastSyncAt = &now
				if err := accountRepo.Update(account); err != nil {
					utils.Logger.WithError(err).Warn("InboundWorker: update last_sync_at failed")
				}
			}

			utils.Logger.WithField("account", account.EmailAddress).
				WithField("fetched", len(parsed)).
				WithField("processed", processed).
				Info("InboundWorker: poll complete")
		}
	}

	// Run immediately at startup
	doPoll()

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			utils.Logger.Info("InboundWorker: stopping")
			return
		case <-ticker.C:
			doPoll()
		}
	}
}
