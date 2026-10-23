//go:build e2e

package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
	paymentmodule "github.com/khushidesai23/Enterprise-Order-Processing/internal/modules/payment"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/repository"
	testintegration "github.com/khushidesai23/Enterprise-Order-Processing/internal/test/integration"
)

const (
	lifecycleStock    = 10
	lifecycleQuantity = 2
	lifecyclePrice    = 19.99
)

// lifecycle drives one customer's order through the public API, with an
// admin for fulfilment and refunds, and signed webhooks from the stub
// gateway standing in for Razorpay.
type lifecycle struct {
	t             *testing.T
	env           *testintegration.AppEnv
	adminToken    string
	customerToken string
	productID     uuid.UUID
}

func newLifecycle(t *testing.T) *lifecycle {
	t.Helper()

	env := testintegration.NewAppEnv(t)
	creator := testintegration.GormCreator(env.DB.DB)
	testintegration.CreateAdminFixture(t, creator, "Life", "Admin", "life-admin@example.com", "password123")
	testintegration.CreateUserFixture(t, creator, "Life", "Customer", "life-customer@example.com", "password123")
	category := testintegration.CreateCategoryFixture(t, creator, "Lifecycle")
	product := testintegration.CreateProductFixture(t, creator, category.ID, "Lifecycle Product", "LIFE-1", lifecyclePrice)
	testintegration.CreateInventoryFixture(t, creator, product.ID, lifecycleStock, 0)

	l := &lifecycle{t: t, env: env, productID: product.ID}
	l.adminToken = l.login("life-admin@example.com")
	l.customerToken = l.login("life-customer@example.com")

	return l
}

func (l *lifecycle) login(email string) string {
	l.t.Helper()

	code, api := l.call(http.MethodPost, "/api/v1/auth/login",
		map[string]interface{}{"email": email, "password": "password123"}, "")
	require.Equal(l.t, http.StatusOK, code, api.Error)

	var response struct {
		Token string `json:"token"`
	}
	testintegration.DecodeData(l.t, api.Data, &response)

	return response.Token
}

func (l *lifecycle) call(method, path string, body interface{}, token string) (int, testintegration.APIResponse) {
	l.t.Helper()

	recorder := testintegration.JSONRequest(l.t, l.env.Router, method, path, body, token, nil)

	return recorder.Code, testintegration.DecodeAPIResponse(l.t, recorder)
}

func (l *lifecycle) createOrder() uuid.UUID {
	l.t.Helper()

	code, api := l.call(http.MethodPost, "/api/v1/orders", map[string]interface{}{
		"items": []map[string]interface{}{{"product_id": l.productID, "quantity": lifecycleQuantity}},
	}, l.customerToken)
	require.Equal(l.t, http.StatusCreated, code, api.Error)

	var order struct {
		ID uuid.UUID `json:"id"`
	}
	testintegration.DecodeData(l.t, api.Data, &order)

	return order.ID
}

func (l *lifecycle) checkout(orderID uuid.UUID) paymentmodule.CheckoutResponse {
	l.t.Helper()

	code, api := l.call(http.MethodPost, "/api/v1/payments",
		map[string]interface{}{"order_id": orderID}, l.customerToken)
	require.Equal(l.t, http.StatusCreated, code, api.Error)

	var checkout paymentmodule.CheckoutResponse
	testintegration.DecodeData(l.t, api.Data, &checkout)

	return checkout
}

// webhookBody builds a Razorpay-shaped payment event for one attempt
// (paymentID) on a gateway order.
func webhookBody(event, gatewayOrderID, paymentID string, entity map[string]interface{}) []byte {
	fields := map[string]interface{}{
		"id":       paymentID,
		"order_id": gatewayOrderID,
		"amount":   3998, // paise: lifecyclePrice * lifecycleQuantity
		"currency": "INR",
	}
	for key, value := range entity {
		fields[key] = value
	}

	body, _ := json.Marshal(map[string]interface{}{
		"event":      event,
		"account_id": "acc_test",
		"payload":    map[string]interface{}{"payment": map[string]interface{}{"entity": fields}},
	})

	return body
}

func (l *lifecycle) deliver(eventID string, body []byte) int {
	l.t.Helper()

	return testintegration.JSONRequest(l.t, l.env.Router, http.MethodPost, "/api/v1/payments/webhook",
		json.RawMessage(body), "", map[string]string{
			"X-Razorpay-Signature": l.env.Gateway.SignWebhook(body),
			"X-Razorpay-Event-Id":  eventID,
		}).Code
}

func (l *lifecycle) webhook(eventID, event, gatewayOrderID, paymentID string) {
	l.t.Helper()

	code := l.deliver(eventID, webhookBody(event, gatewayOrderID, paymentID, nil))
	require.Equal(l.t, http.StatusOK, code, "%s %s", event, eventID)
}

func (l *lifecycle) setStatus(orderID uuid.UUID, status models.OrderStatus) int {
	l.t.Helper()

	code, _ := l.call(http.MethodPatch, fmt.Sprintf("/api/v1/orders/%s/status", orderID),
		map[string]interface{}{"status": status}, l.adminToken)

	return code
}

func (l *lifecycle) refund(paymentID uuid.UUID) (int, testintegration.APIResponse) {
	l.t.Helper()

	return l.call(http.MethodPost, fmt.Sprintf("/api/v1/payments/%s/refund", paymentID), nil, l.adminToken)
}

func (l *lifecycle) cancel(orderID uuid.UUID) (int, testintegration.APIResponse) {
	l.t.Helper()

	return l.call(http.MethodPatch, fmt.Sprintf("/api/v1/orders/%s/cancel", orderID), nil, l.customerToken)
}

// assertState checks the order, its payment (if paymentID is set) and the
// product's stock as stored in the database.
func (l *lifecycle) assertState(
	orderID uuid.UUID,
	orderStatus models.OrderStatus,
	paymentID uuid.UUID,
	paymentStatus models.PaymentStatus,
	available int,
	reserved int,
) {
	l.t.Helper()
	ctx := context.Background()

	order, err := repository.NewOrderRepository(l.env.DB.DB).GetByID(ctx, orderID)
	require.NoError(l.t, err)
	assert.Equal(l.t, orderStatus, order.Status, "order status")

	if paymentID != uuid.Nil {
		payment, err := repository.NewPaymentRepository(l.env.DB.DB).GetByID(ctx, paymentID)
		require.NoError(l.t, err)
		assert.Equal(l.t, paymentStatus, payment.Status, "payment status")
	}

	inventory, err := repository.NewInventoryRepository(l.env.DB.DB).GetByProductID(ctx, l.productID)
	require.NoError(l.t, err)
	assert.Equal(l.t, available, inventory.AvailableQuantity, "available stock")
	assert.Equal(l.t, reserved, inventory.ReservedQuantity, "reserved stock")
}

func (l *lifecycle) webhookRows(eventID string) (int64, models.WebhookStatus) {
	l.t.Helper()

	var rows []models.PaymentWebhook
	require.NoError(l.t, l.env.DB.DB.Where("payload_id = ?", eventID).Find(&rows).Error)
	if len(rows) == 0 {
		return 0, ""
	}

	return int64(len(rows)), rows[0].Status
}

func (l *lifecycle) transactionID(paymentID uuid.UUID) string {
	l.t.Helper()

	payment, err := repository.NewPaymentRepository(l.env.DB.DB).GetByID(context.Background(), paymentID)
	require.NoError(l.t, err)
	require.NotNil(l.t, payment.TransactionID)

	return *payment.TransactionID
}

const (
	// Stock levels: a reservation moves units from available to reserved,
	// shipping consumes the reserved units and a release returns them.
	inStock   = lifecycleStock
	remaining = lifecycleStock - lifecycleQuantity
)

func TestLifecycleFulfilmentThenRefundAfterDelivery(t *testing.T) {
	l := newLifecycle(t)

	orderID := l.createOrder()
	l.assertState(orderID, models.OrderCreated, uuid.Nil, "", remaining, lifecycleQuantity)

	checkout := l.checkout(orderID)
	assert.InDelta(t, lifecyclePrice*lifecycleQuantity, checkout.Amount, 0.001)
	l.assertState(orderID, models.OrderPaymentPending, checkout.PaymentID, models.PaymentPending, remaining, lifecycleQuantity)

	l.webhook("evt_capture", "payment.captured", checkout.GatewayOrderID, "pay_1")
	l.assertState(orderID, models.OrderPaid, checkout.PaymentID, models.PaymentSuccess, remaining, lifecycleQuantity)

	require.Equal(t, http.StatusOK, l.setStatus(orderID, models.OrderPacked))
	require.Equal(t, http.StatusOK, l.setStatus(orderID, models.OrderShipped))
	l.assertState(orderID, models.OrderShipped, checkout.PaymentID, models.PaymentSuccess, remaining, 0)
	require.Equal(t, http.StatusOK, l.setStatus(orderID, models.OrderDelivered))

	code, api := l.refund(checkout.PaymentID)
	require.Equal(t, http.StatusOK, code, api.Error)
	// Money goes back; the delivered order and the consumed stock stay.
	l.assertState(orderID, models.OrderDelivered, checkout.PaymentID, models.PaymentRefunded, remaining, 0)
	require.Len(t, l.env.Gateway.Refunds(), 1)
	assert.Equal(t, "pay_1", l.env.Gateway.Refunds()[0].TransactionID)
	assert.InDelta(t, lifecyclePrice*lifecycleQuantity, l.env.Gateway.Refunds()[0].Amount, 0.001)

	// Razorpay then reports the refund we made: nothing changes.
	l.webhook("evt_refund", "refund.created", checkout.GatewayOrderID, "pay_1")
	l.assertState(orderID, models.OrderDelivered, checkout.PaymentID, models.PaymentRefunded, remaining, 0)

	code, _ = l.refund(checkout.PaymentID)
	assert.Equal(t, http.StatusBadRequest, code, "a payment is refunded once")
	assert.Len(t, l.env.Gateway.Refunds(), 1)
}

func TestLifecycleRefundAfterShipmentKeepsShippedOrder(t *testing.T) {
	l := newLifecycle(t)
	orderID := l.createOrder()
	checkout := l.checkout(orderID)
	l.webhook("evt_capture", "payment.captured", checkout.GatewayOrderID, "pay_1")
	require.Equal(t, http.StatusOK, l.setStatus(orderID, models.OrderPacked))
	require.Equal(t, http.StatusOK, l.setStatus(orderID, models.OrderShipped))

	code, api := l.refund(checkout.PaymentID)

	require.Equal(t, http.StatusOK, code, api.Error)
	l.assertState(orderID, models.OrderShipped, checkout.PaymentID, models.PaymentRefunded, remaining, 0)
}

func TestLifecycleFailedPaymentCanBeRetried(t *testing.T) {
	l := newLifecycle(t)
	orderID := l.createOrder()
	first := l.checkout(orderID)

	l.webhook("evt_fail_1", "payment.failed", first.GatewayOrderID, "pay_attempt_1")

	// The order stays open with its stock held for the retry.
	l.assertState(orderID, models.OrderPaymentPending, first.PaymentID, models.PaymentFailed, remaining, lifecycleQuantity)

	retry := l.checkout(orderID)
	assert.Equal(t, first.PaymentID, retry.PaymentID)
	assert.Equal(t, first.GatewayOrderID, retry.GatewayOrderID, "retry reuses the Razorpay order")
	l.assertState(orderID, models.OrderPaymentPending, first.PaymentID, models.PaymentPending, remaining, lifecycleQuantity)

	l.webhook("evt_capture_2", "payment.captured", first.GatewayOrderID, "pay_attempt_2")
	l.assertState(orderID, models.OrderPaid, first.PaymentID, models.PaymentSuccess, remaining, lifecycleQuantity)
	assert.Equal(t, "pay_attempt_2", l.transactionID(first.PaymentID))

	// The first attempt's failure delivered again, late, under a new event
	// ID must not undo the capture.
	l.webhook("evt_fail_1_late", "payment.failed", first.GatewayOrderID, "pay_attempt_1")
	l.assertState(orderID, models.OrderPaid, first.PaymentID, models.PaymentSuccess, remaining, lifecycleQuantity)

	code, _ := l.call(http.MethodPost, "/api/v1/payments", map[string]interface{}{"order_id": orderID}, l.customerToken)
	assert.Equal(t, http.StatusBadRequest, code, "a paid order cannot be paid again")
}

func TestLifecycleAbandonAfterFailedPayment(t *testing.T) {
	l := newLifecycle(t)
	orderID := l.createOrder()
	checkout := l.checkout(orderID)
	l.webhook("evt_fail", "payment.failed", checkout.GatewayOrderID, "pay_attempt_1")

	code, api := l.cancel(orderID)

	require.Equal(t, http.StatusOK, code, api.Error)
	l.assertState(orderID, models.OrderCancelled, checkout.PaymentID, models.PaymentFailed, inStock, 0)
	code, _ = l.call(http.MethodPost, "/api/v1/payments", map[string]interface{}{"order_id": orderID}, l.customerToken)
	assert.Equal(t, http.StatusBadRequest, code)
}

func TestLifecycleCancelBeforePayment(t *testing.T) {
	l := newLifecycle(t)
	orderID := l.createOrder()

	code, api := l.cancel(orderID)

	require.Equal(t, http.StatusOK, code, api.Error)
	l.assertState(orderID, models.OrderCancelled, uuid.Nil, "", inStock, 0)

	code, _ = l.cancel(orderID)
	assert.Equal(t, http.StatusBadRequest, code, "already cancelled")
	code, _ = l.call(http.MethodPost, "/api/v1/payments", map[string]interface{}{"order_id": orderID}, l.customerToken)
	assert.Equal(t, http.StatusBadRequest, code, "cancelled orders cannot be paid")
	l.assertState(orderID, models.OrderCancelled, uuid.Nil, "", inStock, 0)
}

func TestLifecycleAdminCancelReleasesStock(t *testing.T) {
	l := newLifecycle(t)
	orderID := l.createOrder()

	require.Equal(t, http.StatusOK, l.setStatus(orderID, models.OrderCancelled))

	// Previously a status change to CANCELLED left the stock reserved.
	l.assertState(orderID, models.OrderCancelled, uuid.Nil, "", inStock, 0)
}

func TestLifecyclePaidOrderIsCancelledByRefund(t *testing.T) {
	l := newLifecycle(t)
	orderID := l.createOrder()
	checkout := l.checkout(orderID)
	l.webhook("evt_capture", "payment.captured", checkout.GatewayOrderID, "pay_1")

	code, api := l.cancel(orderID)
	assert.Equal(t, http.StatusBadRequest, code, "cancelling would keep the customer's money")
	assert.Contains(t, api.Error, "refund")
	assert.Equal(t, http.StatusBadRequest, l.setStatus(orderID, models.OrderCancelled))
	l.assertState(orderID, models.OrderPaid, checkout.PaymentID, models.PaymentSuccess, remaining, lifecycleQuantity)

	code, api = l.refund(checkout.PaymentID)

	require.Equal(t, http.StatusOK, code, api.Error)
	l.assertState(orderID, models.OrderCancelled, checkout.PaymentID, models.PaymentRefunded, inStock, 0)

	// The refund webhook must not release the stock a second time.
	l.webhook("evt_refund", "refund.created", checkout.GatewayOrderID, "pay_1")
	l.assertState(orderID, models.OrderCancelled, checkout.PaymentID, models.PaymentRefunded, inStock, 0)
}

func TestLifecycleRefundIsRolledBackWhenGatewayFails(t *testing.T) {
	l := newLifecycle(t)
	orderID := l.createOrder()
	checkout := l.checkout(orderID)
	l.webhook("evt_capture", "payment.captured", checkout.GatewayOrderID, "pay_1")
	l.env.Gateway.RefundErr = errors.New("razorpay unavailable")

	code, _ := l.refund(checkout.PaymentID)

	assert.Equal(t, http.StatusBadGateway, code)
	l.assertState(orderID, models.OrderPaid, checkout.PaymentID, models.PaymentSuccess, remaining, lifecycleQuantity)

	l.env.Gateway.RefundErr = nil
	code, api := l.refund(checkout.PaymentID)
	require.Equal(t, http.StatusOK, code, api.Error)
	l.assertState(orderID, models.OrderCancelled, checkout.PaymentID, models.PaymentRefunded, inStock, 0)
}

func TestLifecycleRefundWebhookWithoutAdminRefund(t *testing.T) {
	// A refund issued from the Razorpay dashboard arrives only as a webhook.
	l := newLifecycle(t)
	orderID := l.createOrder()
	checkout := l.checkout(orderID)
	l.webhook("evt_capture", "payment.captured", checkout.GatewayOrderID, "pay_1")

	partial := webhookBody("refund.created", checkout.GatewayOrderID, "pay_1", map[string]interface{}{"refund_status": "partial"})
	require.Equal(t, http.StatusOK, l.deliver("evt_partial", partial))
	l.assertState(orderID, models.OrderPaid, checkout.PaymentID, models.PaymentSuccess, remaining, lifecycleQuantity)

	fullRefund := webhookBody("refund.created", checkout.GatewayOrderID, "pay_1", map[string]interface{}{"refund_status": "full"})
	require.Equal(t, http.StatusOK, l.deliver("evt_full", fullRefund))
	l.assertState(orderID, models.OrderCancelled, checkout.PaymentID, models.PaymentRefunded, inStock, 0)
	assert.Empty(t, l.env.Gateway.Refunds(), "the gateway already refunded it")
}

func TestLifecycleCaptureAfterCustomerCancelled(t *testing.T) {
	l := newLifecycle(t)
	orderID := l.createOrder()
	checkout := l.checkout(orderID)
	code, api := l.cancel(orderID)
	require.Equal(t, http.StatusOK, code, api.Error)

	// The customer completed the Razorpay checkout anyway.
	l.webhook("evt_late_capture", "payment.captured", checkout.GatewayOrderID, "pay_1")

	// Recorded and acknowledged; the stock stays released.
	l.assertState(orderID, models.OrderCancelled, checkout.PaymentID, models.PaymentSuccess, inStock, 0)
	_, status := l.webhookRows("evt_late_capture")
	assert.Equal(t, models.WebhookProcessed, status)

	code, api = l.refund(checkout.PaymentID)
	require.Equal(t, http.StatusOK, code, api.Error)
	l.assertState(orderID, models.OrderCancelled, checkout.PaymentID, models.PaymentRefunded, inStock, 0)
}

func TestLifecycleOrderStatusCannotBeSetToPaid(t *testing.T) {
	l := newLifecycle(t)
	orderID := l.createOrder()
	l.checkout(orderID)

	assert.Equal(t, http.StatusBadRequest, l.setStatus(orderID, models.OrderPaid))
	assert.Equal(t, http.StatusBadRequest, l.setStatus(orderID, models.OrderPaymentPending))
	l.assertState(orderID, models.OrderPaymentPending, uuid.Nil, "", remaining, lifecycleQuantity)
}

func TestWebhookRedeliveryIsDeduplicatedByEventID(t *testing.T) {
	l := newLifecycle(t)
	orderID := l.createOrder()
	checkout := l.checkout(orderID)
	failure := webhookBody("payment.failed", checkout.GatewayOrderID, "pay_attempt_1", nil)

	require.Equal(t, http.StatusOK, l.deliver("evt_fail", failure))
	l.checkout(orderID) // retry: payment back to PENDING

	// Replaying the failure would mark the retried payment FAILED again;
	// the event ID makes the redelivery a no-op.
	require.Equal(t, http.StatusOK, l.deliver("evt_fail", failure))

	l.assertState(orderID, models.OrderPaymentPending, checkout.PaymentID, models.PaymentPending, remaining, lifecycleQuantity)
	rows, status := l.webhookRows("evt_fail")
	assert.EqualValues(t, 1, rows)
	assert.Equal(t, models.WebhookProcessed, status)
}

func TestWebhookConcurrentDeliveries(t *testing.T) {
	l := newLifecycle(t)
	orderID := l.createOrder()
	checkout := l.checkout(orderID)
	capture := webhookBody("payment.captured", checkout.GatewayOrderID, "pay_1", nil)

	const deliveries = 10
	codes := make(chan int, 2*deliveries)
	var wg sync.WaitGroup
	for i := range deliveries {
		wg.Add(2)
		// The same event delivered in parallel...
		go func() { defer wg.Done(); codes <- l.deliver("evt_same", capture) }()
		// ...and distinct events with the same effect.
		go func() { defer wg.Done(); codes <- l.deliver(fmt.Sprintf("evt_distinct_%d", i), capture) }()
	}
	wg.Wait()
	close(codes)

	for code := range codes {
		assert.Equal(t, http.StatusOK, code)
	}
	rows, _ := l.webhookRows("evt_same")
	assert.EqualValues(t, 1, rows)
	l.assertState(orderID, models.OrderPaid, checkout.PaymentID, models.PaymentSuccess, remaining, lifecycleQuantity)
}

func TestWebhookFailedProcessingIsRetried(t *testing.T) {
	l := newLifecycle(t)
	orderID := l.createOrder()
	checkout := l.checkout(orderID)
	l.webhook("evt_capture", "payment.captured", checkout.GatewayOrderID, "pay_1")
	refund := webhookBody("refund.created", checkout.GatewayOrderID, "pay_1", nil)

	// Corrupt the reservation so releasing stock fails mid-processing.
	inventory := l.env.DB.DB.Model(&models.Inventory{}).Where("product_id = ?", l.productID)
	require.NoError(t, inventory.Update("reserved_quantity", 0).Error)

	assert.Equal(t, http.StatusInternalServerError, l.deliver("evt_refund", refund))
	rows, _ := l.webhookRows("evt_refund")
	assert.EqualValues(t, 0, rows, "a failed event is not recorded, so its retry is processed")
	l.assertState(orderID, models.OrderPaid, checkout.PaymentID, models.PaymentSuccess, remaining, 0)

	require.NoError(t, l.env.DB.DB.Model(&models.Inventory{}).Where("product_id = ?", l.productID).
		Update("reserved_quantity", lifecycleQuantity).Error)

	require.Equal(t, http.StatusOK, l.deliver("evt_refund", refund))
	l.assertState(orderID, models.OrderCancelled, checkout.PaymentID, models.PaymentRefunded, inStock, 0)
}

func TestWebhookEdgeCases(t *testing.T) {
	l := newLifecycle(t)
	orderID := l.createOrder()
	checkout := l.checkout(orderID)
	capture := webhookBody("payment.captured", checkout.GatewayOrderID, "pay_1", nil)

	t.Run("bad signature is rejected and not recorded", func(t *testing.T) {
		code := testintegration.JSONRequest(t, l.env.Router, http.MethodPost, "/api/v1/payments/webhook",
			json.RawMessage(capture), "", map[string]string{
				"X-Razorpay-Signature": "forged",
				"X-Razorpay-Event-Id":  "evt_forged",
			}).Code

		assert.Equal(t, http.StatusUnauthorized, code)
		rows, _ := l.webhookRows("evt_forged")
		assert.EqualValues(t, 0, rows)
		l.assertState(orderID, models.OrderPaymentPending, checkout.PaymentID, models.PaymentPending, remaining, lifecycleQuantity)
	})

	t.Run("unknown gateway order is acknowledged and recorded as failed", func(t *testing.T) {
		stranger := webhookBody("payment.captured", "order_not_ours", "pay_x", nil)

		require.Equal(t, http.StatusOK, l.deliver("evt_stranger", stranger))
		require.Equal(t, http.StatusOK, l.deliver("evt_stranger", stranger))

		rows, status := l.webhookRows("evt_stranger")
		assert.EqualValues(t, 1, rows)
		assert.Equal(t, models.WebhookFailed, status)
	})

	t.Run("unhandled events are acknowledged", func(t *testing.T) {
		for _, event := range []string{"order.paid", "refund.processed", "payment.dispute.created"} {
			body := webhookBody(event, checkout.GatewayOrderID, "pay_1", nil)
			assert.Equal(t, http.StatusOK, l.deliver("evt_"+event, body), event)
		}
		l.assertState(orderID, models.OrderPaymentPending, checkout.PaymentID, models.PaymentPending, remaining, lifecycleQuantity)
	})
}
