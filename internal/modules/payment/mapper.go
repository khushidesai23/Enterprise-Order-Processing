package payment

import (
	"time"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
)

// Request -> Model
func ToPaymentModel(
	req CreatePaymentRequest,
	orderAmount float64,
) *models.Payment {

	return &models.Payment{
		OrderID: req.OrderID,

		Status: models.PaymentPending,

		Amount: orderAmount,

		Currency: "INR",

		Gateway: "RAZORPAY",
	}
}

// Model -> Checkout Response
func ToCheckoutResponse(
	payment *models.Payment,
	keyID string,
) CheckoutResponse {

	return CheckoutResponse{
		PaymentID: payment.ID,

		OrderID: payment.OrderID,

		Amount: payment.Amount,

		Currency: payment.Currency,

		KeyID: keyID,

		Gateway: payment.Gateway,

		GatewayOrderID: payment.GatewayOrderID,
	}
}

// Model -> Payment Response
func ToPaymentResponse(
	payment *models.Payment,
) PaymentResponse {

	return PaymentResponse{
		ID: payment.ID,

		OrderID: payment.OrderID,

		GatewayOrderID: payment.GatewayOrderID,

		TransactionID: payment.TransactionID,

		Status: payment.Status,

		Amount: payment.Amount,

		Currency: payment.Currency,

		Gateway: payment.Gateway,

		CreatedAt: payment.CreatedAt.Format(time.RFC3339),
	}
}

// Model -> Payment List Response
func ToPaymentListResponse(
	payment *models.Payment,
) PaymentListResponse {

	return PaymentListResponse{
		ID: payment.ID,

		OrderID: payment.OrderID,

		Status: payment.Status,

		Amount: payment.Amount,

		Currency: payment.Currency,

		Gateway: payment.Gateway,
	}
}

func ToPaymentList(
	payments []models.Payment,
) []PaymentListResponse {

	response := make(
		[]PaymentListResponse,
		0,
		len(payments),
	)

	for i := range payments {
		response = append(
			response,
			ToPaymentListResponse(&payments[i]),
		)
	}

	return response
}

// Payment Summary
func ToPaymentSummary(
	payments []models.Payment,
) PaymentSummary {

	var summary PaymentSummary

	summary.TotalPayments = int64(len(payments))

	for _, payment := range payments {

		switch payment.Status {

		case models.PaymentSuccess:
			summary.SuccessfulPayments++
			summary.Revenue += payment.Amount

		case models.PaymentFailed:
			summary.FailedPayments++

		case models.PaymentPending:
			summary.PendingPayments++

		case models.PaymentRefunded:
			summary.RefundedPayments++
		}
	}

	return summary
}