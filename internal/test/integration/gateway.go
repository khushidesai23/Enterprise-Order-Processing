//go:build e2e

package integration

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/modules/payment"
)

type StubPaymentGateway struct {
	keyID         string
	keySecret     string
	webhookSecret string
	counter       uint64

	mu      sync.Mutex
	refunds []payment.GatewayRefundRequest
	// RefundErr, when set, makes RefundPayment fail like a gateway outage.
	RefundErr error
}

func NewStubPaymentGateway(keyID, keySecret, webhookSecret string) *StubPaymentGateway {
	return &StubPaymentGateway{
		keyID:         keyID,
		keySecret:     keySecret,
		webhookSecret: webhookSecret,
	}
}

func (g *StubPaymentGateway) CreateOrder(
	ctx context.Context,
	req payment.GatewayOrderRequest,
) (*payment.GatewayOrderResponse, error) {
	_ = ctx

	orderID := fmt.Sprintf("test_gateway_order_%d", atomic.AddUint64(&g.counter, 1))

	return &payment.GatewayOrderResponse{
		OrderID:  orderID,
		Amount:   req.Amount,
		Currency: req.Currency,
	}, nil
}

func (g *StubPaymentGateway) VerifyCheckoutSignature(
	ctx context.Context,
	orderID string,
	paymentID string,
	signature string,
) error {
	_ = ctx

	verifier := payment.NewSignatureVerifier(
		g.keySecret,
		g.webhookSecret,
	)

	return verifier.VerifyCheckoutSignature(orderID, paymentID, signature)
}

func (g *StubPaymentGateway) VerifyWebhookSignature(
	ctx context.Context,
	body []byte,
	signature string,
) error {
	_ = ctx

	verifier := payment.NewSignatureVerifier(
		g.keySecret,
		g.webhookSecret,
	)

	return verifier.VerifyWebhookSignature(body, signature)
}

func (g *StubPaymentGateway) RefundPayment(
	ctx context.Context,
	req payment.GatewayRefundRequest,
) (*payment.GatewayRefundResponse, error) {
	_ = ctx

	g.mu.Lock()
	defer g.mu.Unlock()

	if g.RefundErr != nil {
		return nil, g.RefundErr
	}

	g.refunds = append(g.refunds, req)

	return &payment.GatewayRefundResponse{
		RefundID: fmt.Sprintf("rfnd_test_%d", len(g.refunds)),
	}, nil
}

// Refunds returns the refunds the gateway was asked to make.
func (g *StubPaymentGateway) Refunds() []payment.GatewayRefundRequest {
	g.mu.Lock()
	defer g.mu.Unlock()

	return append([]payment.GatewayRefundRequest(nil), g.refunds...)
}

func (g *StubPaymentGateway) SignWebhook(body []byte) string {
	hash := hmac.New(sha256.New, []byte(g.webhookSecret))
	hash.Write(body)
	return hex.EncodeToString(hash.Sum(nil))
}

// SignCheckout returns the signature Razorpay Checkout hands the browser
// for a successful payment attempt.
func (g *StubPaymentGateway) SignCheckout(gatewayOrderID, paymentID string) string {
	hash := hmac.New(sha256.New, []byte(g.keySecret))
	hash.Write([]byte(gatewayOrderID + "|" + paymentID))
	return hex.EncodeToString(hash.Sum(nil))
}
