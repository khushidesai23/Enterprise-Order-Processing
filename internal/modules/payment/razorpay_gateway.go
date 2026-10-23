package payment

import (
	"context"
	"fmt"

	"github.com/razorpay/razorpay-go"
)

type RazorpayGateway struct {
	client *razorpay.Client

	keyID string

	keySecret string

	webhookSecret string
}

func NewRazorpayGateway(
	keyID string,
	keySecret string,
	webhookSecret string,
) *RazorpayGateway {

	client := razorpay.NewClient(
		keyID,
		keySecret,
	)

	return &RazorpayGateway{
		client:        client,
		keyID:         keyID,
		keySecret:     keySecret,
		webhookSecret: webhookSecret,
	}
}

func (g *RazorpayGateway) KeyID() string {
	return g.keyID
}

func (g *RazorpayGateway) CreateOrder(
	ctx context.Context,
	req GatewayOrderRequest,
) (*GatewayOrderResponse, error) {

	// Razorpay expects amount in paise.
	amount := toPaise(req.Amount)

	data := map[string]interface{}{
		"amount":   amount,
		"currency": req.Currency,
		"receipt":  req.Receipt,
	}

	order, err := g.client.Order.Create(data, nil)
	if err != nil {
		return nil, fmt.Errorf(
			"razorpay create order failed: %w",
			err,
		)
	}

	orderID, ok := order["id"].(string)
	if !ok {
		return nil, fmt.Errorf("gateway returned invalid order id")
	}

	return &GatewayOrderResponse{
		OrderID:  orderID,
		Amount:   req.Amount,
		Currency: req.Currency,
	}, nil
}

func (g *RazorpayGateway) RefundPayment(
	ctx context.Context,
	req GatewayRefundRequest,
) (*GatewayRefundResponse, error) {

	refund, err := g.client.Payment.Refund(
		req.TransactionID,
		toPaise(req.Amount),
		map[string]interface{}{
			"receipt": req.Receipt,
		},
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"razorpay refund failed: %w",
			err,
		)
	}

	refundID, ok := refund["id"].(string)
	if !ok {
		return nil, fmt.Errorf("gateway returned invalid refund id")
	}

	return &GatewayRefundResponse{
		RefundID: refundID,
	}, nil
}

func (g *RazorpayGateway) VerifyCheckoutSignature(
	ctx context.Context,
	orderID string,
	paymentID string,
	signature string,
) error {

	verifier := NewSignatureVerifier(
		g.keySecret,
		g.webhookSecret,
	)

	return verifier.VerifyCheckoutSignature(
		orderID,
		paymentID,
		signature,
	)
}

func (g *RazorpayGateway) VerifyWebhookSignature(
	ctx context.Context,
	body []byte,
	signature string,
) error {

	verifier := NewSignatureVerifier(
		g.keySecret,
		g.webhookSecret,
	)

	return verifier.VerifyWebhookSignature(
		body,
		signature,
	)
}
