package orders

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Order represents one row in orders.json.
type Order struct {
	OrderID          string  `json:"order_id"`
	CustomerName     string  `json:"customer_name"`
	CustomerEmail    string  `json:"customer_email"`
	Product          string  `json:"product"`
	Amount           float64 `json:"amount"`
	Currency         string  `json:"currency"`
	Status           string  `json:"status"`
	OrderedAt        string  `json:"ordered_at"`
	ExpectedDelivery string  `json:"expected_delivery"`
	RevisedDelivery  string  `json:"revised_delivery,omitempty"`
	DeliveredAt      string  `json:"delivered_at,omitempty"`
	TrackingNumber   string  `json:"tracking_number"`
	DelayReason      string  `json:"delay_reason,omitempty"`
	CancelReason     string  `json:"cancel_reason,omitempty"`
	CancelledAt      string  `json:"cancelled_at,omitempty"`
	RefundAmount     float64 `json:"refund_amount,omitempty"`
	CurrentLocation  string  `json:"current_location,omitempty"`
	Note             string  `json:"note"`
}

// Loader reads orders from a JSON file once and caches them.
type Loader struct {
	orders []Order
}

var orderIDRe = regexp.MustCompile(`(?i)ORD[-_]?\d{4,6}`)

// NewLoader loads orders from the given file path.
// Returns an empty loader (no-op) if the file doesn't exist.
func NewLoader(path string) *Loader {
	data, err := os.ReadFile(path)
	if err != nil {
		return &Loader{}
	}
	var payload struct {
		Orders []Order `json:"orders"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return &Loader{}
	}
	return &Loader{orders: payload.Orders}
}

// FindByText extracts an order ID from free text and returns the matching order.
// Returns nil if no order number is found or no match in the dataset.
func (l *Loader) FindByText(text string) *Order {
	match := orderIDRe.FindString(text)
	if match == "" {
		return nil
	}
	match = strings.ToUpper(match)
	if !strings.Contains(match, "-") {
		match = "ORD-" + match[3:]
	}
	for i := range l.orders {
		if strings.EqualFold(l.orders[i].OrderID, match) {
			return &l.orders[i]
		}
	}
	return nil
}

// FindOrderContext implements services.OrderLookup — returns the order context
// snippet ready for injection into an AI prompt, or "" if not found.
func (l *Loader) FindOrderContext(text string) string {
	order := l.FindByText(text)
	if order == nil {
		return ""
	}
	return order.ContextSnippet()
}

// FindByEmail returns all orders for a given customer email.
func (l *Loader) FindByEmail(email string) []Order {
	var result []Order
	for _, o := range l.orders {
		if strings.EqualFold(o.CustomerEmail, email) {
			result = append(result, o)
		}
	}
	return result
}

// ContextSnippet returns a compact string describing the order status,
// ready to inject into an AI prompt.
func (o *Order) ContextSnippet() string {
	var sb strings.Builder
	sb.WriteString("ORDER STATUS LOOKUP RESULT:\n")
	sb.WriteString("Order ID: " + o.OrderID + "\n")
	sb.WriteString("Product: " + o.Product + "\n")
	sb.WriteString("Status: " + o.Status + "\n")
	sb.WriteString("Amount: " + o.Currency + " " + fmt.Sprintf("%.0f", o.Amount) + "\n")
	sb.WriteString("Ordered: " + o.OrderedAt + "\n")
	if o.DeliveredAt != "" {
		sb.WriteString("Delivered: " + o.DeliveredAt + "\n")
	} else if o.RevisedDelivery != "" {
		sb.WriteString("Revised Delivery: " + o.RevisedDelivery + "\n")
		sb.WriteString("Delay Reason: " + o.DelayReason + "\n")
	} else {
		sb.WriteString("Expected Delivery: " + o.ExpectedDelivery + "\n")
	}
	if o.CurrentLocation != "" {
		sb.WriteString("Current Location: " + o.CurrentLocation + "\n")
	}
	if o.TrackingNumber != "" {
		sb.WriteString("Tracking: " + o.TrackingNumber + "\n")
	}
	if o.RefundAmount > 0 {
		sb.WriteString("Refund Amount: " + o.Currency + " " + fmt.Sprintf("%.0f", o.RefundAmount) + "\n")
	}
	sb.WriteString("Note: " + o.Note + "\n")
	return sb.String()
}
