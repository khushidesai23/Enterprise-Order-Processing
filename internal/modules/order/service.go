package order

import (
	"bytes"
	"context"
	"errors"
	"slices"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/repository"
	"github.com/khushidesai23/Enterprise-Order-Processing/pkg/logger"
)

type Service struct {
	orderRepository     orderRepository
	orderItemRepository orderItemRepository
	userRepository      repository.UserRepository
	productRepository   orderProductRepository
	inventoryRepository orderInventoryRepository
	log                 *zap.Logger
}

type orderRepository interface {
	Begin(ctx context.Context) *gorm.DB
	Create(ctx context.Context, tx *gorm.DB, order *models.Order) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Order, error)
	GetByIDTx(ctx context.Context, tx *gorm.DB, id uuid.UUID) (*models.Order, error)
	GetAll(ctx context.Context, limit, offset int) ([]models.Order, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.Order, error)
	GetByUserIDPage(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Order, error)
	UpdateStatusTx(ctx context.Context, tx *gorm.DB, id uuid.UUID, status models.OrderStatus) error
}

type orderItemRepository interface {
	CreateMany(ctx context.Context, tx *gorm.DB, items []models.OrderItem) error
}

type orderProductRepository interface {
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]models.Product, error)
}

type orderInventoryRepository interface {
	ExistsByProductID(ctx context.Context, productID uuid.UUID) (bool, error)
	GetByProductIDTx(ctx context.Context, tx *gorm.DB, productID uuid.UUID) (*models.Inventory, error)
	ReserveStockTx(ctx context.Context, tx *gorm.DB, productID uuid.UUID, quantity int) error
	ReleaseReservedStockTx(ctx context.Context, tx *gorm.DB, productID uuid.UUID, quantity int) error
	ConfirmReservedStockTx(ctx context.Context, tx *gorm.DB, productID uuid.UUID, quantity int) error
}

var ErrInsufficientReserved = errors.New("insufficient reserved inventory")

func NewService(
	orderRepository orderRepository,
	orderItemRepository orderItemRepository,
	userRepository repository.UserRepository,
	productRepository orderProductRepository,
	inventoryRepository orderInventoryRepository,
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
//
// Reads that need no lock (user, product prices) happen before the
// transaction, and the response is built from the rows just written, so the
// transaction holds inventory row locks only for the writes themselves.
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

	// Lock inventory rows in a stable order so concurrent multi-item
	// orders cannot deadlock on each other.
	items := slices.Clone(req.Items)
	slices.SortFunc(items, func(a, b CreateOrderItemRequest) int {
		return bytes.Compare(a.ProductID[:], b.ProductID[:])
	})

	user, err := s.userRepository.GetByID(
		ctx,
		req.UserID,
	)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, ErrUserNotFound
	}

	productIDs := make([]uuid.UUID, len(items))
	for i, item := range items {
		productIDs[i] = item.ProductID
	}

	products, err := s.productRepository.GetByIDs(
		ctx,
		productIDs,
	)
	if err != nil {
		return nil, err
	}

	productsByID := make(map[uuid.UUID]models.Product, len(products))
	for _, product := range products {
		productsByID[product.ID] = product
	}

	order := ToOrderModel(req)
	orderItems := make([]models.OrderItem, 0, len(items))

	for _, requestItem := range items {

		product, ok := productsByID[requestItem.ProductID]
		if !ok {
			return nil, ErrProductNotFound
		}

		orderItem := BuildOrderItem(
			uuid.Nil,
			product.ID,
			requestItem.Quantity,
			product.Price,
		)

		orderItems = append(orderItems, *orderItem)

		order.TotalAmount += product.Price *
			float64(requestItem.Quantity)
	}

	tx := s.orderRepository.Begin(ctx)

	if tx.Error != nil {
		return nil, tx.Error
	}

	defer rollbackOnPanic(tx)

	if err := s.orderRepository.Create(
		ctx,
		tx,
		order,
	); err != nil {
		tx.Rollback()
		return nil, err
	}

	for i := range orderItems {

		orderItems[i].OrderID = order.ID

		// The guarded UPDATE locks the inventory row and checks stock in
		// one statement.
		err := s.inventoryRepository.ReserveStockTx(
			ctx,
			tx,
			orderItems[i].ProductID,
			orderItems[i].Quantity,
		)
		if err != nil {
			tx.Rollback()

			if errors.Is(err, repository.ErrInsufficientQuantity) {
				return nil, s.missingInventoryOrInsufficientStock(
					ctx,
					orderItems[i].ProductID,
				)
			}

			return nil, err
		}
	}

	if err := s.orderItemRepository.CreateMany(
		ctx,
		tx,
		orderItems,
	); err != nil {

		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Attach products only after the insert: GORM would otherwise upsert
	// the association and BaseModel.BeforeCreate would re-ID it.
	for i := range orderItems {
		orderItems[i].Product = productsByID[orderItems[i].ProductID]
	}
	order.Items = orderItems

	logger.InfoContext(
		ctx,
		s.log,
		"order created",
		zap.String("order_id", order.ID.String()),
		zap.String("user_id", order.UserID.String()),
		zap.Float64("total_amount", order.TotalAmount),
	)

	response := ToOrderResponse(order)

	return &response, nil
}

// missingInventoryOrInsufficientStock distinguishes the two reasons a
// guarded reservation can match no row. It only runs on the failure path.
func (s *Service) missingInventoryOrInsufficientStock(
	ctx context.Context,
	productID uuid.UUID,
) error {

	exists, err := s.inventoryRepository.ExistsByProductID(ctx, productID)
	if err != nil {
		return err
	}

	if !exists {
		return ErrInventoryNotFound
	}

	return ErrInsufficientStock
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
	page int,
	pageSize int,
) ([]OrderListResponse, error) {

	orders, err := s.orderRepository.GetAll(
		ctx,
		pageSize,
		(page-1)*pageSize,
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
	page int,
	pageSize int,
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

	orders, err := s.orderRepository.GetByUserIDPage(
		ctx,
		userID,
		pageSize,
		(page-1)*pageSize,
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

	for _, item := range itemsInLockOrder(order.Items) {

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
			if errors.Is(err, repository.ErrInsufficientQuantity) {
				return ErrInsufficientReserved
			}
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

	for _, item := range itemsInLockOrder(order.Items) {

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
			if errors.Is(err, repository.ErrInsufficientQuantity) {
				return ErrInsufficientReserved
			}
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

// itemsInLockOrder returns order items sorted by product ID, the same order
// CreateOrder uses, so every inventory-locking path acquires row locks
// consistently.
func itemsInLockOrder(items []models.OrderItem) []models.OrderItem {
	sorted := slices.Clone(items)
	slices.SortFunc(sorted, func(a, b models.OrderItem) int {
		return bytes.Compare(a.ProductID[:], b.ProductID[:])
	})
	return sorted
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
