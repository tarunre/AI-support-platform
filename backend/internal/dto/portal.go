package dto

import "time"

// PortalTicketInfo is the minimal ticket info shown in the customer portal.
type PortalTicketInfo struct {
	TicketNumber string    `json:"ticket_number"`
	Subject      string    `json:"subject"`
	Status       string    `json:"status"`
	CustomerName string    `json:"customer_name"`
	CreatedAt    time.Time `json:"created_at"`
}

// PortalMessage represents one message in the customer portal conversation.
type PortalMessage struct {
	ID        uint      `json:"id"`
	Direction string    `json:"direction"` // "INBOUND" = customer, "OUTBOUND" = agent
	Body      string    `json:"body"`
	Sender    string    `json:"sender"`
	CreatedAt time.Time `json:"created_at"`
}

// PortalConversationResponse is the full payload returned for the portal view.
type PortalConversationResponse struct {
	Ticket   PortalTicketInfo `json:"ticket"`
	Messages []PortalMessage  `json:"messages"`
}

// PortalReplyRequest is sent by the customer to add a message.
type PortalReplyRequest struct {
	Token   string `json:"token"   binding:"required"`
	Message string `json:"message" binding:"required,min=1,max=5000"`
}
