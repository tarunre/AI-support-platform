package repositories

import (
	"time"

	"github.com/ayush/supportiq/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EmailMessageRepository handles all database access for email_messages.
type EmailMessageRepository struct {
	db *gorm.DB
}

func NewEmailMessageRepository(db *gorm.DB) *EmailMessageRepository {
	return &EmailMessageRepository{db: db}
}

func (r *EmailMessageRepository) Create(msg *models.EmailMessage) error {
	return r.db.Create(msg).Error
}

func (r *EmailMessageRepository) FindByMessageID(tenantID uuid.UUID, messageID string) (*models.EmailMessage, error) {
	var msg models.EmailMessage
	if err := r.db.Where("tenant_id = ? AND message_id = ?", tenantID, messageID).First(&msg).Error; err != nil {
		return nil, err
	}
	return &msg, nil
}

func (r *EmailMessageRepository) ListByTicketID(tenantID uuid.UUID, ticketID uuid.UUID) ([]models.EmailMessage, error) {
	var messages []models.EmailMessage
	err := r.db.Where("tenant_id = ? AND ticket_id = ?", tenantID, ticketID).
		Order("created_at ASC").Find(&messages).Error
	return messages, err
}

func (r *EmailMessageRepository) Update(msg *models.EmailMessage) error {
	return r.db.Save(msg).Error
}

// FindTicketByMessageID finds the ticket_id of the email with the given message_id.
func (r *EmailMessageRepository) FindTicketByMessageID(_ interface{}, messageID string) (uuid.UUID, error) {
	var msg models.EmailMessage
	if err := r.db.Where("message_id = ?", messageID).First(&msg).Error; err != nil {
		return uuid.Nil, err
	}
	return msg.TicketID, nil
}

// FindTicketByThreadID finds the ticket_id of the email with the given thread_id.
func (r *EmailMessageRepository) FindTicketByThreadID(_ interface{}, threadID string) (uuid.UUID, error) {
	var msg models.EmailMessage
	if err := r.db.Where("thread_id = ?", threadID).First(&msg).Error; err != nil {
		return uuid.Nil, err
	}
	return msg.TicketID, nil
}

// FindTicketBySubject finds the ticket_id of the most recent email matching subject and sender.
func (r *EmailMessageRepository) FindTicketBySubject(_ interface{}, subject, sender string) (uuid.UUID, error) {
	var msg models.EmailMessage
	if err := r.db.Where("subject = ? AND sender = ?", subject, sender).
		Order("created_at DESC").First(&msg).Error; err != nil {
		return uuid.Nil, err
	}
	return msg.TicketID, nil
}

// FindLatestInboundByTicket returns the most recent inbound message for a ticket.
func (r *EmailMessageRepository) FindLatestInboundByTicket(ticketID uuid.UUID) (*models.EmailMessage, error) {
	var msg models.EmailMessage
	err := r.db.Where("ticket_id = ? AND direction = 'INBOUND'", ticketID).
		Order("created_at DESC").First(&msg).Error
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// FindQueued returns outbound messages in QUEUED status, up to limit.
func (r *EmailMessageRepository) FindQueued(limit int) ([]models.EmailMessage, error) {
	var msgs []models.EmailMessage
	err := r.db.Where("direction = 'OUTBOUND' AND status = 'QUEUED'").
		Order("created_at ASC").Limit(limit).Find(&msgs).Error
	return msgs, err
}

// HasOutboundForTicket returns true if an outbound email already exists (QUEUED or SENT)
// for the given ticket. Used to prevent sending duplicate auto-replies.
func (r *EmailMessageRepository) HasOutboundForTicket(ticketID uuid.UUID) bool {
	var count int64
	r.db.Model(&models.EmailMessage{}).
		Where("ticket_id = ? AND direction = 'OUTBOUND' AND status IN ('QUEUED','SENT')", ticketID).
		Count(&count)
	return count > 0
}

// FindFailedRetryable returns outbound messages in FAILED status with retries below max.
func (r *EmailMessageRepository) FindFailedRetryable(maxRetries int) ([]models.EmailMessage, error) {
	var msgs []models.EmailMessage
	err := r.db.Where("direction = 'OUTBOUND' AND status = 'FAILED' AND retry_count < ?", maxRetries).
		Order("created_at ASC").Find(&msgs).Error
	return msgs, err
}

// UpdateStatus sets the status (and optional error) on a message.
func (r *EmailMessageRepository) UpdateStatus(id uint, status models.EmailStatus, errMsg string) error {
	updates := map[string]interface{}{"status": status}
	if errMsg != "" {
		updates["error_message"] = errMsg
	}
	return r.db.Model(&models.EmailMessage{}).Where("id = ?", id).Updates(updates).Error
}

// IncrementRetry increments the retry_count column.
func (r *EmailMessageRepository) IncrementRetry(id uint) error {
	return r.db.Model(&models.EmailMessage{}).Where("id = ?", id).
		UpdateColumn("retry_count", gorm.Expr("retry_count + 1")).Error
}

// FindByTicketID returns all messages for a ticket, ordered by creation time.
func (r *EmailMessageRepository) FindByTicketID(ticketID uuid.UUID) ([]models.EmailMessage, error) {
	var msgs []models.EmailMessage
	err := r.db.Where("ticket_id = ?", ticketID).Order("created_at ASC").Find(&msgs).Error
	return msgs, err
}

// CountByStatus returns total email messages with the given status.
func (r *EmailMessageRepository) CountByStatus(status models.EmailStatus) (int64, error) {
	var count int64
	err := r.db.Model(&models.EmailMessage{}).Where("status = ?", status).Count(&count).Error
	return count, err
}

// CountSentToday returns outbound SENT messages created today.
func (r *EmailMessageRepository) CountSentToday() (int64, error) {
	var count int64
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	err := r.db.Model(&models.EmailMessage{}).
		Where("direction = 'OUTBOUND' AND status = 'SENT' AND created_at >= ?", today).
		Count(&count).Error
	return count, err
}

// CountReceivedToday returns inbound RECEIVED messages created today.
func (r *EmailMessageRepository) CountReceivedToday() (int64, error) {
	var count int64
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	err := r.db.Model(&models.EmailMessage{}).
		Where("direction = 'INBOUND' AND status = 'RECEIVED' AND created_at >= ?", today).
		Count(&count).Error
	return count, err
}
