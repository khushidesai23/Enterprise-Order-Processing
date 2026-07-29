package payment

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
	ordermodule "github.com/khushidesai23/Enterprise-Order-Processing/internal/modules/order"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/repository"
)

type Service struct {
	paymentRepository *repository.PaymentRepository
	webhookRepository *repository.PaymentWebhookRepository
	orderService      *ordermodule.Service
	gateway           PaymentGateway
	keyID             string
}

func NewService(
	paymentRepository *repository.PaymentRepository,
	webhookRepository *repository.PaymentWebhookRepository,
	orderService *ordermodule.Service,
	gateway PaymentGateway,
	keyID string,
) *Service {

	return &Service{
		paymentRepository: paymentRepository,
		webhookRepository: webhookRepository,
		orderService:      orderService,
		gateway:           gateway,
		keyID:             keyID,
	}
}

// CreatePayment creates a payment and generates a Razorpay order.
func (s *Service) CreatePayment(
	ctx context.Context,
	req CreatePaymentRequest,
) (*CheckoutResponse, error) {

	tx := s.paymentRepository.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	defer rollbackOnPanic(tx)

	orderModel, err := s.orderService.GetOrderForPaymentTx(tx, req.OrderID)
	if err != nil {
		tx.Rollback()
		return nil, mapOrderError(err)
	}

	if err := s.validateOrderForPayment(orderModel.Status); err != nil {
		tx.Rollback()
		return nil, err
	}

	existingPayment, err := s.paymentRepository.GetByOrderIDTx(tx, req.OrderID)
	if err == nil && existingPayment != nil {
		tx.Rollback()

		if existingPayment.Status == models.PaymentPending {
			response := ToCheckoutResponse(existingPayment, s.keyID)
			return &response, nil
		}

		if existingPayment.Status == models.PaymentSuccess {
			return nil, ErrPaymentAlreadyCompleted
		}

		if existingPayment.Status == models.PaymentFailed {
			return nil, ErrPaymentAlreadyFailed
		}

		if existingPayment.Status == models.PaymentRefunded {
			return nil, ErrInvalidPaymentStatus
		}
	}

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return nil, err
	}

	payment := ToPaymentModel(req, orderModel.TotalAmount)

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

	if err := s.paymentRepository.Create(tx, payment); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := s.orderService.MarkOrderPaymentPending(tx, orderModel.ID); err != nil {
		tx.Rollback()
		return nil, mapOrderError(err)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	response := ToCheckoutResponse(payment, s.keyID)
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

	if _, err := s.orderService.GetOrderForPayment(orderID); err != nil {
		return nil, mapOrderError(err)
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

// ProcessWebhook verifies, persists, and processes Razorpay webhook events.
func (s *Service) ProcessWebhook(
	ctx context.Context,
	req ProcessWebhookRequest,
) error {

	if err := s.gateway.VerifyWebhookSignature(ctx, req.Body, req.Signature); err != nil {
		return err
	}

	webhook, err := ParseWebhook(req.Body)
	if err != nil {
		return err
	}

	status, err := webhook.PaymentStatus()
	if err != nil {
		return err
	}

	tx := s.paymentRepository.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer rollbackOnPanic(tx)

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

	if err := s.webhookRepository.Create(tx, webhookRecord); err != nil {
		tx.Rollback()

		if isDuplicateKey(err) {
			return nil
		}

		return err
	}

	gatewayOrderID := webhook.GatewayOrderID()
	if gatewayOrderID == nil {
		tx.Rollback()
		return ErrPaymentNotFound
	}

	payment, err := s.paymentRepository.GetByGatewayOrderIDTx(tx, *gatewayOrderID)
	if err != nil {
		tx.Rollback()

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPaymentNotFound
		}

		return err
	}

	if err := s.applyWebhookPaymentStatus(tx, payment, webhook.TransactionID(), status); err != nil {
		_ = s.webhookRepository.MarkFailed(tx, webhook.PayloadID(), time.Now().Unix())
		tx.Rollback()
		return err
	}

	if err := s.webhookRepository.MarkProcessed(tx, webhook.PayloadID(), time.Now().Unix()); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

// RefundPayment updates a successful payment as refunded and delegates order behavior.
func (s *Service) RefundPayment(
	id uuid.UUID,
) (*PaymentResponse, error) {

	tx := s.paymentRepository.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	defer rollbackOnPanic(tx)

	payment, err := s.paymentRepository.GetByIDTx(tx, id)
	if err != nil {
		tx.Rollback()

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPaymentNotFound
		}

		return nil, err
	}

	if payment.Status == models.PaymentRefunded {
		tx.Rollback()
		return nil, ErrInvalidPaymentStatus
	}

	if payment.Status != models.PaymentSuccess {
		tx.Rollback()
		return nil, ErrPaymentNotPending
	}

	if err := s.paymentRepository.UpdateStatus(tx, payment.ID, models.PaymentRefunded); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := s.orderService.RefundOrder(tx, payment.OrderID); err != nil {
		tx.Rollback()
		return nil, mapOrderError(err)
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

func (s *Service) applyWebhookPaymentStatus(
	tx *gorm.DB,
	payment *models.Payment,
	transactionID *string,
	status models.PaymentStatus,
) error {

	switch status {
	case models.PaymentPending:
		return nil

	case models.PaymentSuccess:
		if payment.Status == models.PaymentSuccess {
			return nil
		}

		if err := s.paymentRepository.CompletePayment(tx, payment.ID, transactionID, models.PaymentSuccess); err != nil {
			return err
		}

		return mapOrderError(s.orderService.MarkOrderPaid(tx, payment.OrderID))

	case models.PaymentFailed:
		if payment.Status == models.PaymentFailed {
			return nil
		}

		if err := s.paymentRepository.CompletePayment(tx, payment.ID, transactionID, models.PaymentFailed); err != nil {
			return err
		}

		return mapOrderError(s.orderService.CancelOrderByPaymentFailure(tx, payment.OrderID))

	case models.PaymentRefunded:
		if payment.Status == models.PaymentRefunded {
			return nil
		}

		if err := s.paymentRepository.UpdateStatus(tx, payment.ID, models.PaymentRefunded); err != nil {
			return err
		}

		return mapOrderError(s.orderService.RefundOrder(tx, payment.OrderID))

	default:
		return ErrUnknownWebhookEvent
	}
}

func (s *Service) validateOrderForPayment(status models.OrderStatus) error {
	switch status {
	case models.OrderPaid:
		return ErrOrderAlreadyPaid
	case models.OrderCancelled:
		return ErrOrderCancelled
	case models.OrderCreated, models.OrderPaymentPending:
		return nil
	default:
		return ErrOrderNotEligibleForPayment
	}
}

func mapOrderError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, ordermodule.ErrOrderNotFound):
		return ErrOrderNotFound
	case errors.Is(err, ordermodule.ErrOrderAlreadyCancelled):
		return ErrOrderCancelled
	case errors.Is(err, ordermodule.ErrOrderAlreadyCompleted),
		errors.Is(err, ordermodule.ErrOrderCannotBeCancelled),
		errors.Is(err, ordermodule.ErrInvalidOrderStatus):
		return ErrOrderNotEligibleForPayment
	default:
		return err
	}
}

func isDuplicateKey(err error) bool {
	message := strings.ToLower(err.Error())

	return strings.Contains(message, "duplicate key") ||
		strings.Contains(message, "duplicate entry") ||
		strings.Contains(message, "unique constraint") ||
		strings.Contains(message, "unique violation")
}

func rollbackOnPanic(tx *gorm.DB) {
	if r := recover(); r != nil {
		tx.Rollback()
		panic(r)
	}
}
