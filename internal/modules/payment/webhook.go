package payment

import (
	"encoding/json"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
)

// Razorpay Webhook
type RazorpayWebhook struct {
	// payment.captured
	// payment.failed
	// payment.authorized
	// refund.created
	Event string `json:"event"`
	// Merchant Account ID
	AccountID string `json:"account_id"`
	// Webhook creation timestamp (Unix)
	CreatedAt int64 `json:"created_at"`
	Payload RazorpayWebhookPayload `json:"payload"`
}

type RazorpayWebhookPayload struct {
	// Unique payload id
	ID string `json:"id"`
	Payment RazorpayPaymentPayload `json:"payment"`
}

type RazorpayPaymentPayload struct {
	Entity RazorpayPaymentEntity `json:"entity"`
}

type RazorpayPaymentEntity struct {
	// Razorpay Payment ID
	ID string `json:"id"`
	// Razorpay Order ID
	OrderID string `json:"order_id"`
	Status string `json:"status"`
	Amount int64 `json:"amount"`
	Currency string `json:"currency"`
	Method string `json:"method"`
	Description string `json:"description"`
	Email string `json:"email"`
	Contact string `json:"contact"`
}

// Parse Webhook
func ParseWebhook(
	body []byte,
) (*RazorpayWebhook, error) {

	var webhook RazorpayWebhook

	if err := json.Unmarshal(body, &webhook); err != nil {
		return nil, err
	}

	return &webhook, nil
}

// Domain Helpers
// Converts Razorpay event into our domain payment status.
func (w *RazorpayWebhook) PaymentStatus() (models.PaymentStatus, error) {

	switch w.Event {

	case "payment.authorized":
		return models.PaymentPending, nil

	case "payment.captured":
		return models.PaymentSuccess, nil

	case "payment.failed":
		return models.PaymentFailed, nil

	case "refund.created":
		return models.PaymentRefunded, nil

	default:
		return "", ErrUnknownWebhookEvent
	}
}

// Helper Getters
func (w *RazorpayWebhook) PayloadID() string {
	return w.Payload.ID
}

func (w *RazorpayWebhook) EventName() string {
	return w.Event
}

func (w *RazorpayWebhook) GatewayAccountID() string {
	return w.AccountID
}

func (w *RazorpayWebhook) WebhookCreatedAt() int64 {
	return w.CreatedAt
}

func (w *RazorpayWebhook) GatewayOrderID() string {
	return w.Payload.Payment.Entity.OrderID
}

func (w *RazorpayWebhook) TransactionID() string {
	return w.Payload.Payment.Entity.ID
}

func (w *RazorpayWebhook) Amount() int64 {
	return w.Payload.Payment.Entity.Amount
}

func (w *RazorpayWebhook) Currency() string {
	return w.Payload.Payment.Entity.Currency
}

func (w *RazorpayWebhook) Method() string {
	return w.Payload.Payment.Entity.Method
}

func (w *RazorpayWebhook) Email() string {
	return w.Payload.Payment.Entity.Email
}

func (w *RazorpayWebhook) Contact() string {
	return w.Payload.Payment.Entity.Contact
}

func (w *RazorpayWebhook) Description() string {
	return w.Payload.Payment.Entity.Description
}