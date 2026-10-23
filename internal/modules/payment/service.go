package payment

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/authz"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
	ordermodule "github.com/khushidesai23/Enterprise-Order-Processing/internal/modules/order"
	"github.com/khushidesai23/Enterprise-Order-Processing/pkg/logger"
)

type Service struct {
	paymentRepository paymentRepository
	webhookRepository webhookRepository
	orderService      orderService
	gateway           PaymentGateway
	keyID             string
	log               *zap.Logger
}

type paymentRepository interface {
	Begin(ctx context.Context) *gorm.DB
	Create(ctx context.Context, tx *gorm.DB, payment *models.Payment) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Payment, error)
	GetByIDTx(ctx context.Context, tx *gorm.DB, id uuid.UUID) (*models.Payment, error)
	GetByOrderID(ctx context.Context, orderID uuid.UUID) (*models.Payment, error)
	GetByOrderIDTx(ctx context.Context, tx *gorm.DB, orderID uuid.UUID) (*models.Payment, error)
	GetByGatewayOrderIDTx(ctx context.Context, tx *gorm.DB, gatewayOrderID string) (*models.Payment, error)
	GetAll(ctx context.Context) ([]models.Payment, error)
	UpdateStatus(ctx context.Context, tx *gorm.DB, id uuid.UUID, status models.PaymentStatus) error
	CompletePayment(ctx context.Context, tx *gorm.DB, id uuid.UUID, transactionID *string, status models.PaymentStatus) error
}

type webhookRepository interface {
	Create(ctx context.Context, tx *gorm.DB, webhook *models.PaymentWebhook) error
	MarkProcessed(ctx context.Context, tx *gorm.DB, payloadID string, processedAt time.Time) error
	MarkFailed(ctx context.Context, tx *gorm.DB, payloadID string, processedAt time.Time) error
}

type orderService interface {
	GetOrderForPaymentTx(tx *gorm.DB, id uuid.UUID) (*models.Order, error)
	GetOrderForPayment(ctx context.Context, id uuid.UUID) (*models.Order, error)
	MarkOrderPaymentPending(tx *gorm.DB, orderID uuid.UUID) error
	MarkOrderPaid(tx *gorm.DB, orderID uuid.UUID) error
	RefundOrder(tx *gorm.DB, orderID uuid.UUID) error
}

func NewService(
	paymentRepository paymentRepository,
	webhookRepository webhookRepository,
	orderService orderService,
	gateway PaymentGateway,
	keyID string,
	log *zap.Logger,
) *Service {

	return &Service{
		paymentRepository: paymentRepository,
		webhookRepository: webhookRepository,
		orderService:      orderService,
		gateway:           gateway,
		keyID:             keyID,
		log:               log,
	}
}

// CreatePayment creates a payment and generates a Razorpay order.
// Customers may pay only for their own orders.
func (s *Service) CreatePayment(
	ctx context.Context,
	req CreatePaymentRequest,
	principal authz.Principal,
) (*CheckoutResponse, error) {

	tx := s.paymentRepository.Begin(ctx)

	if tx.Error != nil {
		return nil, tx.Error
	}

	defer rollbackOnPanic(tx)

	orderModel, err := s.orderService.GetOrderForPaymentTx(
		tx,
		req.OrderID,
	)
	if err != nil {
		tx.Rollback()
		return nil, mapOrderError(err)
	}

	if !principal.CanAccess(orderModel.UserID) {
		tx.Rollback()
		return nil, ErrOrderNotFound
	}

	if err := s.validateOrderForPayment(orderModel.Status); err != nil {
		tx.Rollback()
		return nil, err
	}

	existingPayment, err := s.paymentRepository.GetByOrderIDTx(
		ctx,
		tx,
		req.OrderID,
	)

	if err == nil && existingPayment != nil {
		return s.resumePayment(ctx, tx, existingPayment)
	}

	if err != nil &&
		!errors.Is(err, gorm.ErrRecordNotFound) {

		tx.Rollback()
		return nil, err
	}

	payment := ToPaymentModel(
		req,
		orderModel.TotalAmount,
	)

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

	if err := s.paymentRepository.Create(
		ctx,
		tx,
		payment,
	); err != nil {

		tx.Rollback()
		return nil, err
	}

	if err := s.orderService.MarkOrderPaymentPending(
		tx,
		orderModel.ID,
	); err != nil {

		tx.Rollback()
		return nil, mapOrderError(err)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	response := ToCheckoutResponse(
		payment,
		s.keyID,
	)

	logger.InfoContext(
		ctx,
		s.log,
		"payment created",
		zap.String("payment_id", payment.ID.String()),
		zap.String("order_id", payment.OrderID.String()),
		zap.String("status", string(payment.Status)),
	)

	return &response, nil
}

// resumePayment answers a checkout request for an order that already has a
// payment, and owns tx from here on.
//
// A failed attempt does not end the payment: Razorpay accepts further
// attempts on the same gateway order until one is captured, so the same
// checkout is returned and the payment goes back to PENDING. The order
// stayed PAYMENT_PENDING with its stock reserved in the meantime.
func (s *Service) resumePayment(
	ctx context.Context,
	tx *gorm.DB,
	payment *models.Payment,
) (*CheckoutResponse, error) {

	switch payment.Status {

	case models.PaymentPending:
		tx.Rollback()

	case models.PaymentFailed:

		if err := s.paymentRepository.UpdateStatus(
			ctx,
			tx,
			payment.ID,
			models.PaymentPending,
		); err != nil {
			tx.Rollback()
			return nil, err
		}

		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		payment.Status = models.PaymentPending

		logger.InfoContext(
			ctx,
			s.log,
			"payment retry started",
			zap.String("payment_id", payment.ID.String()),
			zap.String("order_id", payment.OrderID.String()),
			zap.String("gateway_order_id", payment.GatewayOrderID),
		)

	case models.PaymentSuccess:
		tx.Rollback()
		return nil, ErrPaymentAlreadyCompleted

	default:
		tx.Rollback()
		return nil, ErrInvalidPaymentStatus
	}

	response := ToCheckoutResponse(
		payment,
		s.keyID,
	)

	return &response, nil
}

// GetPayment returns a payment by its ID. Payments for orders the caller
// may not access are reported as not found.
func (s *Service) GetPayment(
	ctx context.Context,
	id uuid.UUID,
	principal authz.Principal,
) (*PaymentResponse, error) {

	payment, err := s.paymentRepository.GetByID(
		ctx,
		id,
	)
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPaymentNotFound
		}

		return nil, err
	}

	// Ownership lives on the order, so only non-admins pay for the lookup.
	if !principal.IsAdmin() {

		order, err := s.orderService.GetOrderForPayment(
			ctx,
			payment.OrderID,
		)
		if err != nil {

			if errors.Is(err, ordermodule.ErrOrderNotFound) {
				return nil, ErrPaymentNotFound
			}

			return nil, err
		}

		if !principal.CanAccess(order.UserID) {
			return nil, ErrPaymentNotFound
		}
	}

	response := ToPaymentResponse(payment)

	return &response, nil
}

// GetPayments returns all payments. Callers must restrict it to admins.
func (s *Service) GetPayments(
	ctx context.Context,
) ([]PaymentListResponse, error) {

	payments, err := s.paymentRepository.GetAll(
		ctx,
	)
	if err != nil {
		return nil, err
	}

	return ToPaymentList(payments), nil
}

// GetPaymentByOrder returns the payment for an order the caller may access.
func (s *Service) GetPaymentByOrder(
	ctx context.Context,
	orderID uuid.UUID,
	principal authz.Principal,
) (*PaymentResponse, error) {

	order, err := s.orderService.GetOrderForPayment(
		ctx,
		orderID,
	)
	if err != nil {
		return nil, mapOrderError(err)
	}

	if !principal.CanAccess(order.UserID) {
		return nil, ErrOrderNotFound
	}

	payment, err := s.paymentRepository.GetByOrderID(
		ctx,
		orderID,
	)
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPaymentNotFound
		}

		return nil, err
	}

	response := ToPaymentResponse(payment)

	return &response, nil
}

// GetPaymentSummary returns overall payment statistics. Callers must
// restrict it to admins.
func (s *Service) GetPaymentSummary(
	ctx context.Context,
) (*PaymentSummary, error) {

	payments, err := s.paymentRepository.GetAll(
		ctx,
	)
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

	if err := s.gateway.VerifyWebhookSignature(
		ctx,
		req.Body,
		req.Signature,
	); err != nil {
		return err
	}

	// The event ID is the idempotency key. Without it every event would
	// share payload_id "" and all but the first would be dropped as
	// duplicates.
	if req.EventID == "" {
		return ErrMissingWebhookEvent
	}

	webhook, err := ParseWebhook(req.Body)
	if err != nil {
		return err
	}

	logger.InfoContext(
		ctx,
		s.log,
		"payment webhook received",
		zap.String("event_id", req.EventID),
		zap.String("event", webhook.EventName()),
	)

	status, err := webhook.PaymentStatus()

	// Acknowledge events that are not acted on (order.paid, refund.processed,
	// partial refunds, ...). Answering non-2xx would only make Razorpay retry
	// them for a day and eventually disable the webhook.
	if errors.Is(err, ErrUnknownWebhookEvent) || webhook.IsPartialRefund() {

		logger.InfoContext(
			ctx,
			s.log,
			"payment webhook ignored",
			zap.String("event_id", req.EventID),
			zap.String("event", webhook.EventName()),
		)

		return nil
	}

	if err != nil {
		return err
	}

	tx := s.paymentRepository.Begin(ctx)

	if tx.Error != nil {
		return tx.Error
	}

	defer rollbackOnPanic(tx)

	webhookRecord := &models.PaymentWebhook{
		Gateway:        "RAZORPAY",
		PayloadID:      req.EventID,
		Event:          webhook.EventName(),
		AccountID:      webhook.GatewayAccountID(),
		TransactionID:  webhook.TransactionID(),
		GatewayOrderID: webhook.GatewayOrderID(),
		Status:         models.WebhookPending,
		RawPayload:     req.Body,
	}

	if err := s.webhookRepository.Create(
		ctx,
		tx,
		webhookRecord,
	); err != nil {

		tx.Rollback()

		if isDuplicateKey(err) {
			return nil
		}

		return err
	}

	payment, err := s.findWebhookPayment(ctx, tx, webhook)
	if err != nil {

		if !errors.Is(err, ErrPaymentNotFound) {
			tx.Rollback()
			return err
		}

		// Not a payment this system created (another integration on the
		// same Razorpay account, a test event, ...). Redelivery cannot fix
		// that, so record it as failed and acknowledge it.
		if err := s.webhookRepository.MarkFailed(
			ctx,
			tx,
			req.EventID,
			time.Now(),
		); err != nil {
			tx.Rollback()
			return err
		}

		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			return err
		}

		logger.WarnContext(
			ctx,
			s.log,
			"payment webhook for unknown payment",
			zap.String("event_id", req.EventID),
			zap.String("event", webhook.EventName()),
			zap.String("gateway_order_id", stringValue(webhook.GatewayOrderID())),
		)

		return nil
	}

	if err := s.applyWebhookPaymentStatus(
		ctx,
		tx,
		payment,
		webhook.TransactionID(),
		status,
	); err != nil {

		// Roll back the webhook record as well so the gateway's retry
		// of this event is processed again rather than deduplicated.
		tx.Rollback()

		return err
	}

	if err := s.webhookRepository.MarkProcessed(
		ctx,
		tx,
		req.EventID,
		time.Now(),
	); err != nil {

		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}

	logger.InfoContext(
		ctx,
		s.log,
		"payment webhook processed",
		zap.String("event_id", req.EventID),
		zap.String("event", webhook.EventName()),
		zap.String("payment_id", payment.ID.String()),
		zap.String("order_id", payment.OrderID.String()),
		zap.String("status", string(status)),
	)

	return nil
}

// findWebhookPayment returns the payment a webhook refers to, read after
// its order is locked.
func (s *Service) findWebhookPayment(
	ctx context.Context,
	tx *gorm.DB,
	webhook *RazorpayWebhook,
) (*models.Payment, error) {

	gatewayOrderID := webhook.GatewayOrderID()

	if gatewayOrderID == nil {
		return nil, ErrPaymentNotFound
	}

	payment, err := s.paymentRepository.GetByGatewayOrderIDTx(
		ctx,
		tx,
		*gatewayOrderID,
	)
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPaymentNotFound
		}

		return nil, err
	}

	return s.lockPayment(ctx, tx, payment)
}

// lockPayment locks the payment's order and re-reads the payment.
//
// The order row lock is the single serialization point for payment state:
// checkout, webhooks, refunds and order cancellation all take it first, so
// they apply one at a time and always see each other's committed changes,
// and no two paths can deadlock by locking payment and order in opposite
// orders.
func (s *Service) lockPayment(
	ctx context.Context,
	tx *gorm.DB,
	payment *models.Payment,
) (*models.Payment, error) {

	if _, err := s.orderService.GetOrderForPaymentTx(
		tx,
		payment.OrderID,
	); err != nil {
		return nil, mapOrderError(err)
	}

	locked, err := s.paymentRepository.GetByIDTx(
		ctx,
		tx,
		payment.ID,
	)
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPaymentNotFound
		}

		return nil, err
	}

	return locked, nil
}

// RefundPayment refunds a captured payment in full through the gateway.
//
// Before shipment the order is cancelled and its reserved stock released;
// after shipment the order keeps its fulfilment status (returns are not
// modelled) and only the money is returned.
func (s *Service) RefundPayment(
	ctx context.Context,
	id uuid.UUID,
) (*PaymentResponse, error) {

	tx := s.paymentRepository.Begin(ctx)

	if tx.Error != nil {
		return nil, tx.Error
	}

	defer rollbackOnPanic(tx)

	payment, err := s.paymentRepository.GetByIDTx(
		ctx,
		tx,
		id,
	)
	if err != nil {

		tx.Rollback()

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPaymentNotFound
		}

		return nil, err
	}

	payment, err = s.lockPayment(ctx, tx, payment)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if payment.Status == models.PaymentRefunded {
		tx.Rollback()
		return nil, ErrInvalidPaymentStatus
	}

	if payment.Status != models.PaymentSuccess ||
		payment.TransactionID == nil {

		tx.Rollback()
		return nil, ErrPaymentNotRefundable
	}

	if err := s.paymentRepository.UpdateStatus(
		ctx,
		tx,
		payment.ID,
		models.PaymentRefunded,
	); err != nil {

		tx.Rollback()
		return nil, err
	}

	if err := s.orderService.RefundOrder(
		tx,
		payment.OrderID,
	); err != nil {

		tx.Rollback()
		return nil, mapOrderError(err)
	}

	// The gateway is called last. If it fails, nothing is committed. If it
	// succeeds and the commit then fails, Razorpay's refund.created webhook
	// applies the same change, so the two cannot stay out of step.
	gatewayRefund, err := s.gateway.RefundPayment(
		ctx,
		GatewayRefundRequest{
			TransactionID: *payment.TransactionID,
			Amount:        payment.Amount,
			Receipt:       payment.ID.String(),
		},
	)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("%w: %w", ErrGatewayRefund, err)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	payment, err = s.paymentRepository.GetByID(
		ctx,
		payment.ID,
	)
	if err != nil {
		return nil, err
	}

	logger.InfoContext(
		ctx,
		s.log,
		"payment refunded",
		zap.String("payment_id", payment.ID.String()),
		zap.String("order_id", payment.OrderID.String()),
		zap.String("refund_id", gatewayRefund.RefundID),
		zap.String("status", string(payment.Status)),
	)

	response := ToPaymentResponse(payment)

	return &response, nil
}

func (s *Service) applyWebhookPaymentStatus(
	ctx context.Context,
	tx *gorm.DB,
	payment *models.Payment,
	transactionID *string,
	status models.PaymentStatus,
) error {

	switch status {

	case models.PaymentPending:
		return nil

	case models.PaymentSuccess:

		switch payment.Status {
		case models.PaymentSuccess:
			return nil

		case models.PaymentRefunded:
			// Delivered after the refund; the refund stands.
			s.logIgnoredTransition(ctx, payment, status)
			return nil
		}

		if err := s.paymentRepository.CompletePayment(
			ctx,
			tx,
			payment.ID,
			transactionID,
			models.PaymentSuccess,
		); err != nil {
			return err
		}

		logger.InfoContext(
			ctx,
			s.log,
			"payment completed",
			zap.String("payment_id", payment.ID.String()),
			zap.String("order_id", payment.OrderID.String()),
			zap.String("transaction_id", stringValue(transactionID)),
		)

		err := s.orderService.MarkOrderPaid(
			tx,
			payment.OrderID,
		)

		if errors.Is(err, ordermodule.ErrOrderAlreadyCancelled) {
			// The order was cancelled while this attempt was in flight, so
			// its stock is gone. The capture is still recorded (the payment
			// stays SUCCESS) so that an admin can refund it, and the event
			// is acknowledged rather than retried forever.
			logger.ErrorContext(
				ctx,
				s.log,
				"payment captured for cancelled order; refund required",
				zap.String("payment_id", payment.ID.String()),
				zap.String("order_id", payment.OrderID.String()),
				zap.String("transaction_id", stringValue(transactionID)),
			)

			return nil
		}

		return mapOrderError(err)

	case models.PaymentFailed:

		switch payment.Status {
		case models.PaymentFailed:
			return nil

		case models.PaymentSuccess, models.PaymentRefunded:
			// A Razorpay order can have several attempts, and an earlier
			// attempt's failure may arrive after a later attempt's capture.
			// A failure never overrides a captured or refunded payment.
			s.logIgnoredTransition(ctx, payment, status)
			return nil
		}

		if err := s.paymentRepository.CompletePayment(
			ctx,
			tx,
			payment.ID,
			transactionID,
			models.PaymentFailed,
		); err != nil {
			return err
		}

		// The order is deliberately left PAYMENT_PENDING with its stock
		// reserved: the customer can retry (POST /payments returns the same
		// checkout) or cancel the order to release the stock.
		logger.WarnContext(
			ctx,
			s.log,
			"payment attempt failed",
			zap.String("payment_id", payment.ID.String()),
			zap.String("order_id", payment.OrderID.String()),
			zap.String("transaction_id", stringValue(transactionID)),
		)

		return nil

	case models.PaymentRefunded:

		if payment.Status == models.PaymentRefunded {
			return nil
		}

		if err := s.paymentRepository.UpdateStatus(
			ctx,
			tx,
			payment.ID,
			models.PaymentRefunded,
		); err != nil {
			return err
		}

		logger.InfoContext(
			ctx,
			s.log,
			"payment marked refunded",
			zap.String("payment_id", payment.ID.String()),
			zap.String("order_id", payment.OrderID.String()),
		)

		return mapOrderError(
			s.orderService.RefundOrder(
				tx,
				payment.OrderID,
			),
		)

	default:
		return ErrUnknownWebhookEvent
	}
}

func (s *Service) logIgnoredTransition(
	ctx context.Context,
	payment *models.Payment,
	status models.PaymentStatus,
) {
	logger.WarnContext(
		ctx,
		s.log,
		"stale payment webhook ignored",
		zap.String("payment_id", payment.ID.String()),
		zap.String("order_id", payment.OrderID.String()),
		zap.String("payment_status", string(payment.Status)),
		zap.String("webhook_status", string(status)),
	)
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

func stringValue(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}
