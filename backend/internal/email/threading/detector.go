// Package threading detects whether an inbound email belongs to an existing
// ticket by inspecting RFC 2822 threading headers (In-Reply-To, References)
// and falling back to normalised subject matching.
package threading

import (
	"context"
	"strings"

	"github.com/ayush/supportiq/internal/email/providers"
	"github.com/ayush/supportiq/internal/repositories"
	"github.com/google/uuid"
)

// Detector resolves inbound emails to existing ticket IDs.
type Detector struct {
	messageRepo *repositories.EmailMessageRepository
}

// NewDetector creates a Detector backed by the email message repository.
func NewDetector(repo *repositories.EmailMessageRepository) *Detector {
	return &Detector{messageRepo: repo}
}

// Detect returns the ticket UUID that owns the thread this email belongs to.
// Returns uuid.Nil when no matching ticket is found (= new ticket).
func (d *Detector) Detect(ctx context.Context, p *providers.ParsedEmail) (uuid.UUID, error) {
	// 1. In-Reply-To header — most reliable signal
	if p.InReplyTo != "" {
		if id, err := d.messageRepo.FindTicketByMessageID(ctx, p.InReplyTo); err == nil {
			return id, nil
		}
	}

	// 2. References header — contains full chain of prior Message-IDs
	for _, ref := range parseReferences(p.References) {
		if id, err := d.messageRepo.FindTicketByMessageID(ctx, ref); err == nil {
			return id, nil
		}
	}

	// 3. Thread-ID header (Gmail X-GM-THRID, etc.)
	if p.ThreadID != "" {
		if id, err := d.messageRepo.FindTicketByThreadID(ctx, p.ThreadID); err == nil {
			return id, nil
		}
	}

	// 4. Subject matching (strip Re:/Fwd: prefixes and compare)
	if sub := normaliseSubject(p.Subject); sub != "" {
		if id, err := d.messageRepo.FindTicketBySubject(ctx, sub, p.FromAddress); err == nil {
			return id, nil
		}
	}

	return uuid.Nil, nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// parseReferences splits the References header into individual Message-IDs.
func parseReferences(refs string) []string {
	var out []string
	for _, part := range strings.Fields(refs) {
		part = strings.Trim(part, "<>")
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

// normaliseSubject strips common reply/forward prefixes.
func normaliseSubject(subject string) string {
	subject = strings.TrimSpace(subject)
	for {
		lower := strings.ToLower(subject)
		switch {
		case strings.HasPrefix(lower, "re:"):
			subject = strings.TrimSpace(subject[3:])
		case strings.HasPrefix(lower, "fwd:"):
			subject = strings.TrimSpace(subject[4:])
		case strings.HasPrefix(lower, "fw:"):
			subject = strings.TrimSpace(subject[3:])
		default:
			return subject
		}
	}
}
