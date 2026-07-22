package payment

import (
	"context"
	"encoding/json"
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
	amount := int(req.Amount * 100)

	data := map[string]interface{}{
		"amount":   amount,
		"currency": req.Currency,
		"receipt":  req.Receipt,
	}

	order, err := g.client.Order.Create(data, nil)
	if err != nil {
		return nil, ErrGatewayOrderCreation
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

func (g *RazorpayGateway) VerifySignature(
	ctx context.Context,
	req GatewayVerificationRequest,
) error {

	verifier := NewSignatureVerifier(
		g.keySecret,
		g.webhookSecret,
	)

	return verifier.VerifyCheckoutSignature(
		req.OrderID,
		req.PaymentID,
		req.Signature,
	)
}

func (g *RazorpayGateway) ParseWebhook(
	ctx context.Context,
	body []byte,
	signature string,
) (*GatewayCallback, error) {

	var payload struct {
		Event string `json:"event"`

		Payload struct {
			Payment struct {
				Entity struct {
					OrderID string `json:"order_id"`

					ID string `json:"id"`

					Status string `json:"status"`
				} `json:"entity"`
			} `json:"payment"`
		} `json:"payload"`
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	return &GatewayCallback{
		OrderID:       payload.Payload.Payment.Entity.OrderID,
		PaymentID:     payload.Payload.Payment.Entity.ID,
		PaymentStatus: payload.Payload.Payment.Entity.Status,
		Signature:     signature,
	}, nil
}
