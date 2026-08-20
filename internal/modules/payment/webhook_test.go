package payment

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
)

func TestParseWebhookAndHelpers(t *testing.T) {
	body := []byte(`{
		"event":"payment.captured",
		"account_id":"acc_1",
		"created_at":123456,
		"payload":{
			"id":"payload_1",
			"payment":{
				"entity":{
					"id":"pay_1",
					"order_id":"order_1",
					"status":"captured",
					"amount":1000,
					"currency":"INR",
					"method":"card",
					"description":"test",
					"email":"user@example.com",
					"contact":"9999999999"
				}
			}
		}
	}`)

	webhook, err := ParseWebhook(body)
	require.NoError(t, err)

	status, err := webhook.PaymentStatus()
	require.NoError(t, err)

	assert.Equal(t, models.PaymentSuccess, status)
	assert.Equal(t, "payment.captured", webhook.EventName())
	assert.Equal(t, int64(123456), webhook.WebhookCreatedAt())
	assert.Equal(t, "order_1", *webhook.GatewayOrderID())
	assert.Equal(t, "pay_1", *webhook.TransactionID())
	assert.Equal(t, "acc_1", *webhook.GatewayAccountID())
	assert.Equal(t, int64(1000), webhook.Amount())
	assert.Equal(t, "INR", webhook.Currency())
	assert.Equal(t, "card", webhook.Method())
	assert.Equal(t, "user@example.com", webhook.Email())
	assert.Equal(t, "9999999999", webhook.Contact())
	assert.Equal(t, "test", webhook.Description())
}

func TestWebhookPaymentStatusRejectsUnknownEvent(t *testing.T) {
	webhook := &RazorpayWebhook{Event: "unknown"}

	_, err := webhook.PaymentStatus()

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnknownWebhookEvent)
}
