package payment

import (
	"strings"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/repository"
)

// package-level errors
var (
	ErrInsufficientReserved = errors.New("insufficient reserved stock")
)

type Service struct {

	paymentRepository *repository.PaymentRepository

	webhookRepository *repository.PaymentWebhookRepository

	orderRepository *repository.OrderRepository

	inventoryRepository *repository.InventoryRepository

	gateway PaymentGateway

	keyID string
}

func NewService(
	paymentRepository *repository.PaymentRepository,
	webhookRepository *repository.PaymentWebhookRepository,
	orderRepository *repository.OrderRepository,
	inventoryRepository *repository.InventoryRepository,
	gateway PaymentGateway,
	keyID string,
) *Service {

	return &Service{
		paymentRepository:   paymentRepository,
		webhookRepository:   webhookRepository,
		orderRepository:     orderRepository,
		inventoryRepository: inventoryRepository,
		gateway:             gateway,
		keyID:               keyID,
	}
}

// CreatePayment creates a payment and generates a Razorpay Order.
func (s *Service) CreatePayment(
	ctx context.Context,
	req CreatePaymentRequest,
) (*CheckoutResponse, error) {

	orderModel, err := s.orderRepository.GetByID(req.OrderID)
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}

		return nil, err
	}

	switch orderModel.Status {

	case models.OrderPaid:
		return nil, ErrOrderAlreadyPaid

	case models.OrderCancelled:
		return nil, ErrOrderCancelled
	}

	existingPayment, err := s.paymentRepository.GetByOrderID(
		req.OrderID,
	)

	if err == nil && existingPayment != nil {

		if existingPayment.Status == models.PaymentPending {

			response := ToCheckoutResponse(
				existingPayment,
				s.keyID,
			)

			return &response, nil
		}

		if existingPayment.Status == models.PaymentSuccess {
			return nil, ErrPaymentAlreadyCompleted
		}
	}

	tx := s.paymentRepository.Begin()

	if tx.Error != nil {
		return nil, tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	payment := ToPaymentModel(
		req,
		orderModel.TotalAmount,
	)

	// Create Razorpay Order
	gatewayOrder, err := s.gateway.CreateOrder(
		ctx,
		GatewayOrderRequest{
			Amount:   payment.Amount,
			Currency: payment.Currency,
			Receipt:  orderModel.ID.String(),
		},
	)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	payment.GatewayOrderID = gatewayOrder.OrderID

	// Save Payment
	if err := s.paymentRepository.Create(
		tx,
		payment,
	); err != nil {

		tx.Rollback()
		return nil, err
	}

	// Update Order Status
	if err := s.orderRepository.UpdateStatusTx(
		tx,
		orderModel.ID,
		models.OrderPaymentPending,
	); err != nil {

		tx.Rollback()
		return nil, err
	}

	// Commit
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	response := ToCheckoutResponse(
		payment,
		s.keyID,
	)

	return &response, nil
}

// GetPayment returns a payment by its ID.
func (s *Service) GetPayment(
	id uuid.UUID,
) (*PaymentResponse, error) {

	payment, err := s.paymentRepository.GetByID(id)
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPaymentNotFound
		}

		return nil, err
	}

	response := ToPaymentResponse(payment)

	return &response, nil
}

// GetPayments returns all payments.
func (s *Service) GetPayments() ([]PaymentListResponse, error) {

	payments, err := s.paymentRepository.GetAll()
	if err != nil {
		return nil, err
	}

	return ToPaymentList(payments), nil
}

// GetPaymentByOrder returns the payment for an order.
func (s *Service) GetPaymentByOrder(
	orderID uuid.UUID,
) (*PaymentResponse, error) {

	// Validate order exists
	_, err := s.orderRepository.GetByID(orderID)
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}

		return nil, err
	}

	payment, err := s.paymentRepository.GetByOrderID(orderID)
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPaymentNotFound
		}

		return nil, err
	}

	response := ToPaymentResponse(payment)

	return &response, nil
}

// GetPaymentSummary returns overall payment statistics.
func (s *Service) GetPaymentSummary() (*PaymentSummary, error) {

	payments, err := s.paymentRepository.GetAll()
	if err != nil {
		return nil, err
	}

	summary := ToPaymentSummary(payments)

	return &summary, nil
}

// ProcessWebhook processes Razorpay webhook events.
func (s *Service) ProcessWebhook(
	ctx context.Context,
	req ProcessWebhookRequest,
) error {

	//--------------------------------------------------
	// Verify Razorpay Signature
	//--------------------------------------------------

	if err := s.gateway.VerifyWebhookSignature(
		ctx,
		req.Body,
		req.Signature,
	); err != nil {
		return err
	}

	//--------------------------------------------------
	// Parse Webhook
	//--------------------------------------------------

	webhook, err := ParseWebhook(req.Body)
	if err != nil {
		return err
	}

	//--------------------------------------------------
	// Begin Transaction
	//--------------------------------------------------

	tx := s.paymentRepository.Begin()

	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	//--------------------------------------------------
	// Persist Webhook
	//--------------------------------------------------

	webhookRecord := &models.PaymentWebhook{
		Gateway:        "RAZORPAY",
		PayloadID:      webhook.PayloadID(),
		Event:          webhook.EventName(),
		AccountID:      webhook.GatewayAccountID(),
		TransactionID:  webhook.TransactionID(),
		GatewayOrderID: webhook.GatewayOrderID(),
		Status:         models.WebhookPending,
		RawPayload:     req.Body,
	}

	if err := s.webhookRepository.Create(
		tx,
		webhookRecord,
	); err != nil {

		// Duplicate webhook
		// payload_id has UNIQUE INDEX
		// Returning nil makes webhook
		// processing idempotent.

		if strings.Contains(
			err.Error(),
			"duplicate key",
		) {

			tx.Rollback()

			return nil
		}

		tx.Rollback()

		return err
	}

	//--------------------------------------------------
	// Convert Razorpay Event
	//--------------------------------------------------

	status, err := webhook.PaymentStatus()
	if err != nil {

		tx.Rollback()

		return err
	}

	//--------------------------------------------------
	// Find Payment
	//--------------------------------------------------

	payment, err := s.paymentRepository.GetByGatewayOrderID(
		webhook.GatewayOrderID(),
	)
	if err != nil {

		tx.Rollback()

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPaymentNotFound
		}

		return err
	}

		//--------------------------------------------------
	// Handle Business Event
	//--------------------------------------------------

	switch status {

	case models.PaymentPending:

		// Nothing to do.
		// Payment is authorized but not yet captured.

	case models.PaymentSuccess:

		//----------------------------------------------
		// Idempotency
		//----------------------------------------------

		if payment.Status == models.PaymentSuccess {

			processedAt := time.Now().Unix()

			if err := s.webhookRepository.MarkProcessed(
				tx,
				webhook.PayloadID(),
				processedAt,
			); err != nil {

				tx.Rollback()
				return err
			}

			if err := tx.Commit().Error; err != nil {
				tx.Rollback()
				return err
			}

			return nil
		}

		if err := s.paymentRepository.CompletePayment(
			tx,
			payment.ID,
			webhook.TransactionID(),
			models.PaymentSuccess,
		); err != nil {

			tx.Rollback()
			return err
		}

		if err := s.orderRepository.UpdateStatusTx(
			tx,
			payment.OrderID,
			models.OrderPaid,
		); err != nil {

			tx.Rollback()
			return err
		}

	case models.PaymentFailed:

		//----------------------------------------------
		// Idempotency
		//----------------------------------------------

		if payment.Status == models.PaymentFailed {

			processedAt := time.Now().Unix()

			if err := s.webhookRepository.MarkProcessed(
				tx,
				webhook.PayloadID(),
				processedAt,
			); err != nil {

				tx.Rollback()
				return err
			}

			if err := tx.Commit().Error; err != nil {
				tx.Rollback()
				return err
			}

			return nil
		}

		if err := s.paymentRepository.CompletePayment(
			tx,
			payment.ID,
			webhook.TransactionID(),
			models.PaymentFailed,
		); err != nil {

			tx.Rollback()
			return err
		}

		orderModel, err := s.orderRepository.GetByID(
			payment.OrderID,
		)
		if err != nil {
			tx.Rollback()
			return err
		}

		for _, item := range orderModel.Items {

			inventory, err := s.inventoryRepository.GetByProductIDTx(
				tx,
				item.ProductID,
			)
			if err != nil {
				tx.Rollback()
				return err
			}

			if inventory.ReservedQuantity < item.Quantity {
				tx.Rollback()

				_ = s.webhookRepository.MarkFailed(
					tx,
					webhook.PayloadID(),
					time.Now().Unix(),
				)

				return ErrInsufficientReserved
			}

			if err := s.inventoryRepository.ReleaseReservedStockTx(
				tx,
				item.ProductID,
				item.Quantity,
			); err != nil {

				tx.Rollback()

				_ = s.webhookRepository.MarkFailed(
					tx,
					webhook.PayloadID(),
					time.Now().Unix(),
				)

				return err
			}
		}

		if err := s.orderRepository.UpdateStatusTx(
			tx,
			payment.OrderID,
			models.OrderCancelled,
		); err != nil {

			tx.Rollback()

			_ = s.webhookRepository.MarkFailed(
				tx,
				webhook.PayloadID(),
				time.Now().Unix(),
			)

			return err
		}

	case models.PaymentRefunded:

		// Future implementation

	default:

		tx.Rollback()

		return ErrUnknownWebhookEvent
	}

	//--------------------------------------------------
	// Mark Webhook Processed
	//--------------------------------------------------

	processedAt := time.Now().Unix()

	if err := s.webhookRepository.MarkProcessed(
		tx,
		webhook.PayloadID(),
		processedAt,
	); err != nil {

		tx.Rollback()

		return err
	}

	//--------------------------------------------------
	// Commit
	//--------------------------------------------------

	if err := tx.Commit().Error; err != nil {

		tx.Rollback()

		return err
	}

	return nil
}

// RefundPayment updates a successful payment as refunded.
// Gateway refund integration will be added later.
func (s *Service) RefundPayment(
	id uuid.UUID,
) (*PaymentResponse, error) {

	payment, err := s.paymentRepository.GetByID(id)
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPaymentNotFound
		}

		return nil, err
	}

	if payment.Status == models.PaymentRefunded {
		return nil, ErrInvalidPaymentStatus
	}

	if payment.Status != models.PaymentSuccess {
		return nil, ErrPaymentNotPending
	}

	tx := s.paymentRepository.Begin()

	if tx.Error != nil {
		return nil, tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := s.paymentRepository.UpdateStatus(
		tx,
		payment.ID,
		models.PaymentRefunded,
	); err != nil {

		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	payment, err = s.paymentRepository.GetByID(payment.ID)
	if err != nil {
		return nil, err
	}

	response := ToPaymentResponse(payment)

	return &response, nil
}

//
// Helpers
//

func (s *Service) isPaymentCompleted(
	status models.PaymentStatus,
) bool {

	return status == models.PaymentSuccess
}

func (s *Service) isPaymentFailed(
	status models.PaymentStatus,
) bool {

	return status == models.PaymentFailed
}

func (s *Service) isPaymentPending(
	status models.PaymentStatus,
) bool {

	return status == models.PaymentPending
}

func (s *Service) isPaymentRefunded(
	status models.PaymentStatus,
) bool {

	return status == models.PaymentRefunded
}
