package order

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/repository"
	"github.com/khushidesai23/Enterprise-Order-Processing/pkg/logger"
)

type Service struct {
	orderRepository     *repository.OrderRepository
	orderItemRepository *repository.OrderItemRepository
	userRepository      repository.UserRepository
	productRepository   *repository.ProductRepository
	inventoryRepository *repository.InventoryRepository
	log                 *zap.Logger
}

var ErrInsufficientReserved = errors.New("insufficient reserved inventory")

func NewService(
	orderRepository *repository.OrderRepository,
	orderItemRepository *repository.OrderItemRepository,
	userRepository repository.UserRepository,
	productRepository *repository.ProductRepository,
	inventoryRepository *repository.InventoryRepository,
	log *zap.Logger,
) *Service {

	return &Service{
		orderRepository:     orderRepository,
		orderItemRepository: orderItemRepository,
		userRepository:      userRepository,
		productRepository:   productRepository,
		inventoryRepository: inventoryRepository,
		log:                 log,
	}
}

// CreateOrder creates a new order and reserves inventory in one transaction.
func (s *Service) CreateOrder(
	ctx context.Context,
	req CreateOrderRequest,
) (*OrderResponse, error) {

	if len(req.Items) == 0 {
		return nil, ErrOrderItemsRequired
	}

	if err := s.ensureUniqueProducts(req.Items); err != nil {
		return nil, err
	}

	tx := s.orderRepository.Begin(ctx)

	if tx.Error != nil {
		return nil, tx.Error
	}

	defer rollbackOnPanic(tx)

	user, err := s.userRepository.GetByIDTx(
		ctx,
		tx,
		req.UserID,
	)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if user == nil {
		tx.Rollback()
		return nil, ErrUserNotFound
	}

	order := ToOrderModel(req)

	if err := s.orderRepository.Create(
		ctx,
		tx,
		order,
	); err != nil {
		tx.Rollback()
		return nil, err
	}

	var (
		orderItems []models.OrderItem
		total      float64
	)

	for _, requestItem := range req.Items {

		product, err := s.productRepository.GetByID(
			ctx,
			requestItem.ProductID,
		)
		if err != nil {

			tx.Rollback()

			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrProductNotFound
			}

			return nil, err
		}

		inventory, err := s.inventoryRepository.GetByProductIDTx(
			ctx,
			tx,
			requestItem.ProductID,
		)
		if err != nil {

			tx.Rollback()

			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrInventoryNotFound
			}

			return nil, err
		}

		if inventory.AvailableQuantity < requestItem.Quantity {
			tx.Rollback()
			return nil, ErrInsufficientStock
		}

		if err := s.inventoryRepository.ReserveStockTx(
			ctx,
			tx,
			requestItem.ProductID,
			requestItem.Quantity,
		); err != nil {

			tx.Rollback()
			return nil, err
		}

		orderItem := BuildOrderItem(
			order.ID,
			product.ID,
			requestItem.Quantity,
			product.Price,
		)

		orderItems = append(
			orderItems,
			*orderItem,
		)

		total += product.Price *
			float64(requestItem.Quantity)
	}

	if err := s.orderItemRepository.CreateMany(
		ctx,
		tx,
		orderItems,
	); err != nil {

		tx.Rollback()
		return nil, err
	}

	if err := s.orderRepository.UpdateTotalAmount(
		ctx,
		tx,
		order.ID,
		total,
	); err != nil {

		tx.Rollback()
		return nil, err
	}

	order.TotalAmount = total

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	order, err = s.orderRepository.GetByID(
		ctx,
		order.ID,
	)
	if err != nil {
		return nil, err
	}

	response := ToOrderResponse(order)

	return &response, nil
}

// GetOrder returns a single order by its ID.
func (s *Service) GetOrder(
	ctx context.Context,
	id uuid.UUID,
) (*OrderResponse, error) {

	order, err := s.orderRepository.GetByID(
		ctx,
		id,
	)
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}

		return nil, err
	}

	response := ToOrderResponse(order)

	return &response, nil
}

// GetOrderForPaymentTx returns order data needed by the payment domain
// inside the caller's transaction.
func (s *Service) GetOrderForPaymentTx(
	tx *gorm.DB,
	id uuid.UUID,
) (*models.Order, error) {

	ctx := tx.Statement.Context

	order, err := s.orderRepository.GetByIDTx(
		ctx,
		tx,
		id,
	)
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}

		return nil, err
	}

	return order, nil
}

// GetOrderForPayment returns order data needed by the payment domain.
func (s *Service) GetOrderForPayment(
	ctx context.Context,
	id uuid.UUID,
) (*models.Order, error) {

	order, err := s.orderRepository.GetByID(
		ctx,
		id,
	)
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}

		return nil, err
	}

	return order, nil
}

// GetOrders returns all orders.
func (s *Service) GetOrders(
	ctx context.Context,
) ([]OrderListResponse, error) {

	orders, err := s.orderRepository.GetAll(
		ctx,
	)
	if err != nil {
		return nil, err
	}

	return ToOrderList(orders), nil
}

// GetOrdersByUser returns all orders for a specific user.
func (s *Service) GetOrdersByUser(
	ctx context.Context,
	userID uuid.UUID,
) ([]OrderListResponse, error) {

	user, err := s.userRepository.GetByID(
		ctx,
		userID,
	)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, ErrUserNotFound
	}

	orders, err := s.orderRepository.GetByUserID(
		ctx,
		userID,
	)
	if err != nil {
		return nil, err
	}

	return ToOrderList(orders), nil
}

// UpdateOrderStatus validates and applies a status transition.
func (s *Service) UpdateOrderStatus(
	ctx context.Context,
	id uuid.UUID,
	req UpdateOrderStatusRequest,
) (*OrderResponse, error) {

	tx := s.orderRepository.Begin(ctx)

	if tx.Error != nil {
		return nil, tx.Error
	}

	defer rollbackOnPanic(tx)

	order, err := s.orderRepository.GetByIDTx(
		ctx,
		tx,
		id,
	)
	if err != nil {

		tx.Rollback()

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}

		return nil, err
	}

	if err := s.transitionOrder(
		ctx,
		tx,
		order,
		req.Status,
	); err != nil {

		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	order, err = s.orderRepository.GetByID(
		ctx,
		order.ID,
	)
	if err != nil {
		return nil, err
	}

	response := ToOrderResponse(order)

	logger.InfoContext(
		ctx,
		s.log,
		"order created",
		zap.String("order_id", order.ID.String()),
		zap.String("user_id", order.UserID.String()),
		zap.Float64("total_amount", order.TotalAmount),
	)

	return &response, nil
}

// MarkOrderPaymentPending records that checkout was initialized.
//
// The transaction is owned by the Payment module.
// Its context is reused for OpenTelemetry propagation.
func (s *Service) MarkOrderPaymentPending(
	tx *gorm.DB,
	orderID uuid.UUID,
) error {

	ctx := tx.Statement.Context

	order, err := s.orderRepository.GetByIDTx(
		ctx,
		tx,
		orderID,
	)
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrOrderNotFound
		}

		return err
	}

	return s.transitionOrder(
		ctx,
		tx,
		order,
		models.OrderPaymentPending,
	)
}

// MarkOrderPaid owns the order-side business transition
// after payment succeeds.
func (s *Service) MarkOrderPaid(
	tx *gorm.DB,
	orderID uuid.UUID,
) error {

	ctx := tx.Statement.Context

	order, err := s.orderRepository.GetByIDTx(
		ctx,
		tx,
		orderID,
	)
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrOrderNotFound
		}

		return err
	}

	return s.transitionOrder(
		ctx,
		tx,
		order,
		models.OrderPaid,
	)
}

// CancelOrderByPaymentFailure cancels an order and releases
// reserved inventory.
func (s *Service) CancelOrderByPaymentFailure(
	tx *gorm.DB,
	orderID uuid.UUID,
) error {

	ctx := tx.Statement.Context

	order, err := s.orderRepository.GetByIDTx(
		ctx,
		tx,
		orderID,
	)
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrOrderNotFound
		}

		return err
	}

	return s.cancelOrderTx(
		ctx,
		tx,
		order,
	)
}

// RefundOrder owns order-side behavior for a refunded payment.
func (s *Service) RefundOrder(
	tx *gorm.DB,
	orderID uuid.UUID,
) error {

	ctx := tx.Statement.Context

	order, err := s.orderRepository.GetByIDTx(
		ctx,
		tx,
		orderID,
	)
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrOrderNotFound
		}

		return err
	}

	if order.Status == models.OrderDelivered {
		return nil
	}

	if order.Status == models.OrderCancelled {
		return nil
	}

	return s.cancelOrderTx(
		ctx,
		tx,
		order,
	)
}

// ConfirmShipment confirms reserved inventory consumption
// for shipped orders.
func (s *Service) ConfirmShipment(
	tx *gorm.DB,
	orderID uuid.UUID,
) error {

	ctx := tx.Statement.Context

	order, err := s.orderRepository.GetByIDTx(
		ctx,
		tx,
		orderID,
	)
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrOrderNotFound
		}

		return err
	}

	return s.transitionOrder(
		ctx,
		tx,
		order,
		models.OrderShipped,
	)
}

// CancelOrder cancels an order and releases any reserved inventory.
func (s *Service) CancelOrder(
	ctx context.Context,
	id uuid.UUID,
) (*OrderResponse, error) {

	tx := s.orderRepository.Begin(ctx)

	if tx.Error != nil {
		return nil, tx.Error
	}

	defer rollbackOnPanic(tx)

	order, err := s.orderRepository.GetByIDTx(
		ctx,
		tx,
		id,
	)
	if err != nil {

		tx.Rollback()

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}

		return nil, err
	}

	if err := s.cancelOrderTx(
		ctx,
		tx,
		order,
	); err != nil {

		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	order, err = s.orderRepository.GetByID(
		ctx,
		order.ID,
	)
	if err != nil {
		return nil, err
	}

	response := ToOrderResponse(order)

	return &response, nil
}

func (s *Service) transitionOrder(
	ctx context.Context,
	tx *gorm.DB,
	order *models.Order,
	next models.OrderStatus,
) error {

	if err := s.validateStatusTransition(
		order.Status,
		next,
	); err != nil {
		return err
	}

	if next == models.OrderShipped {

		if err := s.confirmReservedInventory(
			ctx,
			tx,
			order,
		); err != nil {
			return err
		}
	}

	if err := s.orderRepository.UpdateStatusTx(
		ctx,
		tx,
		order.ID,
		next,
	); err != nil {
		return err
	}

	logger.InfoContext(
		ctx,
		s.log,
		"order status updated",
		zap.String("order_id", order.ID.String()),
		zap.String("from_status", string(order.Status)),
		zap.String("to_status", string(next)),
	)

	return nil
}

func (s *Service) cancelOrderTx(
	ctx context.Context,
	tx *gorm.DB,
	order *models.Order,
) error {

	switch order.Status {
	case models.OrderCancelled:
		return ErrOrderAlreadyCancelled

	case models.OrderDelivered:
		return ErrOrderAlreadyCompleted

	case models.OrderShipped:
		return ErrOrderCannotBeCancelled
	}

	if err := s.releaseReservedInventory(
		ctx,
		tx,
		order,
	); err != nil {
		return err
	}

	return s.orderRepository.UpdateStatusTx(
		ctx,
		tx,
		order.ID,
		models.OrderCancelled,
	)
}

func (s *Service) confirmReservedInventory(
	ctx context.Context,
	tx *gorm.DB,
	order *models.Order,
) error {

	for _, item := range order.Items {

		inventory, err := s.inventoryRepository.GetByProductIDTx(
			ctx,
			tx,
			item.ProductID,
		)
		if err != nil {

			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrInventoryNotFound
			}

			return err
		}

		if inventory.ReservedQuantity < item.Quantity {
			return ErrInsufficientReserved
		}

		if err := s.inventoryRepository.ConfirmReservedStockTx(
			ctx,
			tx,
			item.ProductID,
			item.Quantity,
		); err != nil {
			return err
		}

		logger.InfoContext(
			ctx,
			s.log,
			"reserved inventory confirmed",
			zap.String("order_id", order.ID.String()),
			zap.String("product_id", item.ProductID.String()),
			zap.Int("quantity", item.Quantity),
		)
	}

	return nil
}

func (s *Service) releaseReservedInventory(
	ctx context.Context,
	tx *gorm.DB,
	order *models.Order,
) error {

	for _, item := range order.Items {

		inventory, err := s.inventoryRepository.GetByProductIDTx(
			ctx,
			tx,
			item.ProductID,
		)
		if err != nil {

			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrInventoryNotFound
			}

			return err
		}

		if inventory.ReservedQuantity < item.Quantity {
			return ErrInsufficientReserved
		}

		if err := s.inventoryRepository.ReleaseReservedStockTx(
			ctx,
			tx,
			item.ProductID,
			item.Quantity,
		); err != nil {
			return err
		}

		logger.InfoContext(
			ctx,
			s.log,
			"reserved inventory released",
			zap.String("order_id", order.ID.String()),
			zap.String("product_id", item.ProductID.String()),
			zap.Int("quantity", item.Quantity),
		)
	}

	return nil
}

func (s *Service) ensureUniqueProducts(
	items []CreateOrderItemRequest,
) error {

	productMap := make(
		map[uuid.UUID]struct{},
		len(items),
	)

	for _, item := range items {

		if _, exists := productMap[item.ProductID]; exists {
			return ErrDuplicateProduct
		}

		productMap[item.ProductID] = struct{}{}
	}

	return nil
}

// validateStatusTransition validates whether an order
// is allowed to move states.
func (s *Service) validateStatusTransition(
	current models.OrderStatus,
	next models.OrderStatus,
) error {

	if current == next {
		return nil
	}

	switch current {

	case models.OrderDelivered:
		return ErrOrderAlreadyCompleted

	case models.OrderCancelled:
		return ErrOrderAlreadyCancelled
	}

	allowedTransitions := map[models.OrderStatus][]models.OrderStatus{

		models.OrderCreated: {
			models.OrderPaymentPending,
			models.OrderCancelled,
		},

		models.OrderPaymentPending: {
			models.OrderPaid,
			models.OrderCancelled,
		},

		models.OrderPaid: {
			models.OrderPacked,
			models.OrderCancelled,
		},

		models.OrderPacked: {
			models.OrderShipped,
		},

		models.OrderShipped: {
			models.OrderDelivered,
		},
	}

	validStatuses, ok := allowedTransitions[current]
	if !ok {
		return ErrInvalidOrderStatus
	}

	for _, status := range validStatuses {

		if status == next {
			return nil
		}
	}

	return ErrInvalidOrderStatus
}

func rollbackOnPanic(tx *gorm.DB) {

	if r := recover(); r != nil {
		tx.Rollback()
		panic(r)
	}
}
