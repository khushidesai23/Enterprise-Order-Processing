package integration

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync/atomic"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/modules/payment"
)

type StubPaymentGateway struct {
	keyID         string
	keySecret     string
	webhookSecret string
	counter       uint64
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

func (g *StubPaymentGateway) SignWebhook(body []byte) string {
	hash := hmac.New(sha256.New, []byte(g.webhookSecret))
	hash.Write(body)
	return hex.EncodeToString(hash.Sum(nil))
}
