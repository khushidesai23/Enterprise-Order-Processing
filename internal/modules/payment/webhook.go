package payment

import "encoding/json"

// Razorpay Webhook Event
type RazorpayWebhook struct {
	Event string `json:"event"`

	Payload RazorpayWebhookPayload `json:"payload"`
}

type RazorpayWebhookPayload struct {
	Payment RazorpayPaymentPayload `json:"payment"`
}

type RazorpayPaymentPayload struct {
	Entity RazorpayPaymentEntity `json:"entity"`
}

type RazorpayPaymentEntity struct {
	ID string `json:"id"`

	OrderID string `json:"order_id"`

	Status string `json:"status"`

	Amount int64 `json:"amount"`

	Currency string `json:"currency"`
}

// Parse Webhook
func ParseWebhook(
	body []byte,
) (*RazorpayWebhook, error) {

	var webhook RazorpayWebhook

	if err := json.Unmarshal(
		body,
		&webhook,
	); err != nil {
		return nil, err
	}

	return &webhook, nil
}

// Helpers
func (w *RazorpayWebhook) IsPaymentCaptured() bool {
	return w.Event == "payment.captured"
}

func (w *RazorpayWebhook) IsPaymentAuthorized() bool {
	return w.Event == "payment.authorized"
}

func (w *RazorpayWebhook) IsPaymentFailed() bool {
	return w.Event == "payment.failed"
}

func (w *RazorpayWebhook) IsRefundCreated() bool {
	return w.Event == "refund.created"
}

func (w *RazorpayWebhook) ToPaymentStatus() (string, error) {

	switch w.Event {

	case "payment.authorized":
		return "AUTHORIZED", nil

	case "payment.captured":
		return "SUCCESS", nil

	case "payment.failed":
		return "FAILED", nil

	case "refund.created":
		return "REFUNDED", nil

	default:
		return "", ErrUnknownWebhookEvent
	}
}