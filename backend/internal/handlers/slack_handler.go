package handlers

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	emailcrypto "github.com/ayush/supportiq/internal/email/crypto"
	"github.com/ayush/supportiq/internal/models"
	"github.com/ayush/supportiq/internal/repositories"
	"github.com/ayush/supportiq/internal/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SlackHandler processes inbound Slack Event API webhooks.
// Endpoint: POST /api/v1/slack/events/:integrationID  (public — no JWT)
type SlackHandler struct {
	integRepo    *repositories.IntegrationRepository
	ticketRepo   *repositories.TicketRepository
	commentRepo  *repositories.CommentRepository
	activityRepo *repositories.ActivityRepository
	encKey       string
}

func NewSlackHandler(db *gorm.DB, encKey string) *SlackHandler {
	return &SlackHandler{
		integRepo:    repositories.NewIntegrationRepository(db),
		ticketRepo:   repositories.NewTicketRepository(db),
		commentRepo:  repositories.NewCommentRepository(db),
		activityRepo: repositories.NewActivityRepository(db),
		encKey:       encKey,
	}
}

// ── Slack payload types ───────────────────────────────────────────────────────

type slackEnvelope struct {
	Type      string     `json:"type"`
	Challenge string     `json:"challenge"`
	Event     slackEvent `json:"event"`
}

type slackEvent struct {
	Type      string `json:"type"`
	Text      string `json:"text"`
	User      string `json:"user"`
	Channel   string `json:"channel"`
	Timestamp string `json:"ts"`
	ThreadTS  string `json:"thread_ts"`
	BotID     string `json:"bot_id"`
	SubType   string `json:"subtype"`
}

// ── HandleEvents ──────────────────────────────────────────────────────────────

func (h *SlackHandler) HandleEvents(c *gin.Context) {
	integIDStr := c.Param("integrationID")
	integID64, err := strconv.ParseUint(integIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid integration id"})
		return
	}

	// Read raw body (required for HMAC verification before JSON parse)
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, 1<<20))
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	// Load integration (no tenant scoping — public endpoint)
	intg, err := h.integRepo.FindByIDNoTenant(uint(integID64))
	if err != nil || !intg.Enabled || intg.Provider != "slack" {
		c.Status(http.StatusNotFound)
		return
	}

	// Decrypt config to get signing_secret, bot_token, support_channel_id
	plaintext, err := emailcrypto.Decrypt(h.encKey, intg.Configuration)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal([]byte(plaintext), &cfg); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	signingSecret, _ := cfg["signing_secret"].(string)
	botToken, _ := cfg["bot_token"].(string)
	supportChannelID, _ := cfg["support_channel_id"].(string)

	// Verify Slack request signature (skip if signing_secret not configured yet)
	if signingSecret != "" {
		if !verifySlackSignature(
			c.GetHeader("X-Slack-Request-Timestamp"),
			c.GetHeader("X-Slack-Signature"),
			signingSecret, body,
		) {
			c.Status(http.StatusUnauthorized)
			return
		}
	}

	var envelope slackEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	// Slack one-time URL verification challenge
	if envelope.Type == "url_verification" {
		c.JSON(http.StatusOK, gin.H{"challenge": envelope.Challenge})
		return
	}

	// Acknowledge immediately — Slack retries if response > 3 seconds
	c.Status(http.StatusOK)

	// Process in background so we return 200 before doing DB work
	go h.processEvent(envelope.Event, intg, botToken, supportChannelID)
}

// ── Event processing ──────────────────────────────────────────────────────────

func (h *SlackHandler) processEvent(
	evt slackEvent,
	intg *models.Integration,
	botToken, supportChannelID string,
) {
	if evt.Type != "message" {
		return
	}
	// Ignore bot messages, message_changed, message_deleted, etc.
	if evt.BotID != "" || evt.SubType != "" {
		return
	}
	// Only process messages from the configured support channel
	if supportChannelID != "" && evt.Channel != supportChannelID {
		return
	}
	if strings.TrimSpace(evt.Text) == "" {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Thread reply → append as comment to existing ticket
	if evt.ThreadTS != "" && evt.ThreadTS != evt.Timestamp {
		h.handleThreadReply(ctx, intg, evt, botToken)
		return
	}

	// New top-level message → create ticket
	h.handleNewMessage(ctx, intg, evt, botToken)
}

// handleNewMessage creates a ticket from a new Slack message.
func (h *SlackHandler) handleNewMessage(
	ctx context.Context,
	intg *models.Integration,
	evt slackEvent,
	botToken string,
) {
	tenantID := intg.TenantID
	title := slackExtractTitle(evt.Text)

	var ticket models.Ticket
	err := h.ticketRepo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ticketNum, err := h.ticketRepo.NextTicketNumber(tenantID, tx)
		if err != nil {
			return err
		}
		ticket = models.Ticket{
			TenantID:      tenantID,
			Subject:       title,
			Description:   evt.Text,
			CustomerEmail: fmt.Sprintf("%s@slack.user", evt.User),
			CustomerName:  evt.User,
			TicketNumber:  ticketNum,
			Status:        models.TicketStatusOpen,
			Priority:      models.TicketPriorityMedium,
			Category:      models.TicketCategoryGeneral,
			Source:        models.TicketSource("SLACK"),
			CreatedBy:     0,
		}
		return h.ticketRepo.Create(tx, &ticket)
	})
	if err != nil {
		utils.Logger.WithError(err).Error("Slack inbound: ticket creation failed")
		return
	}

	// Activity log
	_ = h.activityRepo.Create(&models.TicketActivity{
		TenantID:     tenantID,
		TicketID:     ticket.ID,
		UserID:       0,
		ActivityType: models.ActivityCreateTicket,
		Description:  "Ticket created from Slack message",
	})

	// Store Slack thread_ts → ticket mapping for future thread replies
	slackURL := fmt.Sprintf("https://slack.com/archives/%s/p%s",
		evt.Channel, strings.ReplaceAll(evt.Timestamp, ".", ""))
	_ = h.integRepo.CreateTicketIntegration(&models.TicketIntegration{
		TenantID:      tenantID,
		TicketID:      ticket.ID,
		IntegrationID: intg.ID,
		ExternalID:    evt.Timestamp, // used as thread_ts key for replies
		ExternalKey:   evt.Channel,
		ExternalURL:   slackURL,
	})

	// Reply in the Slack thread to confirm ticket creation
	if botToken != "" {
		msg := fmt.Sprintf(
			"✅ *Ticket %s created!*\nWe've received your request and will get back to you soon.\n> _%s_",
			ticket.TicketNumber, title,
		)
		if err := slackPostMessage(ctx, botToken, evt.Channel, evt.Timestamp, msg); err != nil {
			utils.Logger.WithError(err).Warn("Slack inbound: failed to post confirmation reply")
		}
	}

	utils.Logger.
		WithField("ticket", ticket.TicketNumber).
		WithField("slack_channel", evt.Channel).
		WithField("slack_ts", evt.Timestamp).
		Info("Slack inbound: ticket created")
}

// handleThreadReply appends a Slack thread reply as a ticket comment.
func (h *SlackHandler) handleThreadReply(
	ctx context.Context,
	intg *models.Integration,
	evt slackEvent,
	botToken string,
) {
	ti, err := h.integRepo.FindBySlackThread(intg.ID, evt.ThreadTS)
	if err != nil {
		// Thread not tracked by SupportIQ — silently ignore
		return
	}

	comment := &models.TicketComment{
		TenantID:    intg.TenantID,
		TicketID:    ti.TicketID,
		UserID:      0,
		Message:     fmt.Sprintf("**[Slack — @%s]**\n\n%s", evt.User, evt.Text),
		CommentType: models.CommentTypePublic,
	}
	if err := h.commentRepo.Create(comment); err != nil {
		utils.Logger.WithError(err).Error("Slack inbound: comment creation failed")
		return
	}

	utils.Logger.
		WithField("ticket_id", ti.TicketID).
		WithField("slack_ts", evt.Timestamp).
		Info("Slack inbound: comment added from thread reply")
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// verifySlackSignature validates the HMAC-SHA256 signature Slack attaches to every request.
func verifySlackSignature(timestamp, signature, signingSecret string, body []byte) bool {
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false
	}
	// Reject requests older than 5 minutes (replay attack prevention)
	if abs(time.Now().Unix()-ts) > 300 {
		return false
	}
	base := "v0:" + timestamp + ":" + string(body)
	mac := hmac.New(sha256.New, []byte(signingSecret))
	mac.Write([]byte(base))
	expected := "v0=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

// slackExtractTitle returns the first line of text, capped at 80 chars.
func slackExtractTitle(text string) string {
	first := strings.SplitN(strings.TrimSpace(text), "\n", 2)[0]
	first = strings.TrimSpace(first)
	if len(first) > 80 {
		first = first[:77] + "..."
	}
	if first == "" {
		return "Slack support request"
	}
	return first
}

// slackPostMessage calls Slack's chat.postMessage API to reply in a thread.
func slackPostMessage(ctx context.Context, botToken, channel, threadTS, text string) error {
	payload := map[string]interface{}{
		"channel":   channel,
		"text":      text,
		"thread_ts": threadTS,
	}
	b, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://slack.com/api/chat.postMessage", bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+botToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}
