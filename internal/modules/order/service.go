package order

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/repository"
)

type Service struct {
	orderRepository     *repository.OrderRepository
	orderItemRepository *repository.OrderItemRepository

	userRepository repository.UserRepository

	productRepository   *repository.ProductRepository
	inventoryRepository *repository.InventoryRepository
}

var ErrInsufficientReserved = errors.New("insufficient reserved inventory")

func NewService(
	orderRepository *repository.OrderRepository,
	orderItemRepository *repository.OrderItemRepository,
	userRepository repository.UserRepository,
	productRepository *repository.ProductRepository,
	inventoryRepository *repository.InventoryRepository,
) *Service {

	return &Service{
		orderRepository:     orderRepository,
		orderItemRepository: orderItemRepository,
		userRepository:      userRepository,
		productRepository:   productRepository,
		inventoryRepository: inventoryRepository,
	}
}

// CreateOrder creates a new order.
//
// Flow:
//
// 1. Validate User
// 2. Validate Products
// 3. Validate Inventory
// 4. Reserve Inventory
// 5. Create Order
// 6. Create Order Items
// 7. Calculate Total
// 8. Commit Transaction
func (s *Service) CreateOrder(
	ctx context.Context,
	req CreateOrderRequest,
) (*OrderResponse, error) {

	// -----------------------------
	// Validate Request
	// -----------------------------

	if len(req.Items) == 0 {
		return nil, ErrOrderItemsRequired
	}

	// -----------------------------
	// Validate User
	// -----------------------------

	user, err := s.userRepository.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, ErrUserNotFound
	}

	// -----------------------------
	// Prevent duplicate products
	// -----------------------------

	productMap := make(map[uuid.UUID]bool)

	for _, item := range req.Items {

		if productMap[item.ProductID] {
			return nil, ErrDuplicateProduct
		}

		productMap[item.ProductID] = true
	}

	// -----------------------------
	// Begin Transaction
	// -----------------------------

	tx := s.orderRepository.Begin()

	if tx.Error != nil {
		return nil, tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	order := ToOrderModel(req)

	if err := s.orderRepository.Create(tx, order); err != nil {
		tx.Rollback()
		return nil, err
	}

	var (
		orderItems []models.OrderItem
		total      float64
	)

		// ----------------------------------
	// Validate Products & Reserve Stock
	// ----------------------------------

	for _, requestItem := range req.Items {

		product, err := s.productRepository.GetByID(requestItem.ProductID)
		if err != nil {

			if errors.Is(err, gorm.ErrRecordNotFound) {
				tx.Rollback()
				return nil, ErrProductNotFound
			}

			tx.Rollback()
			return nil, err
		}

		if product == nil {
			tx.Rollback()
			return nil, ErrProductNotFound
		}

		inventory, err := s.inventoryRepository.GetByProductID(requestItem.ProductID)
		if err != nil {

			if errors.Is(err, gorm.ErrRecordNotFound) {
				tx.Rollback()
				return nil, ErrInventoryNotFound
			}

			tx.Rollback()
			return nil, err
		}

		if inventory == nil {
			tx.Rollback()
			return nil, ErrInventoryNotFound
		}

		// -----------------------------
		// Check Available Stock
		// -----------------------------

		if inventory.AvailableQuantity < requestItem.Quantity {
			tx.Rollback()
			return nil, ErrInsufficientStock
		}

		// -----------------------------
		// Reserve Inventory
		// -----------------------------

		if err := s.inventoryRepository.ReserveStock(
			requestItem.ProductID,
			requestItem.Quantity,
		); err != nil {
			tx.Rollback()
			return nil, err
		}

		// -----------------------------
		// Build Order Item
		// -----------------------------

		orderItem := BuildOrderItem(
			order.ID,
			product.ID,
			requestItem.Quantity,
			product.Price,
		)

		orderItems = append(orderItems, *orderItem)

		total += product.Price * float64(requestItem.Quantity)
	}

		// ----------------------------------
	// Persist Order Items
	// ----------------------------------

	if err := s.orderItemRepository.CreateMany(
		tx,
		orderItems,
	); err != nil {
		tx.Rollback()
		return nil, err
	}

	// ----------------------------------
	// Update Order Total
	// ----------------------------------

	if err := s.orderRepository.UpdateTotalAmount(
		tx,
		order.ID,
		total,
	); err != nil {
		tx.Rollback()
		return nil, err
	}

	// Keep local object in sync
	order.TotalAmount = total

	// ----------------------------------
	// Commit Transaction
	// ----------------------------------

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// ----------------------------------
	// Reload Complete Order
	// ----------------------------------

	order, err = s.orderRepository.GetByID(order.ID)
	if err != nil {
		return nil, err
	}

	response := ToOrderResponse(order)

	return &response, nil
}

// GetOrder returns a single order by its ID.
func (s *Service) GetOrder(
	id uuid.UUID,
) (*OrderResponse, error) {

	order, err := s.orderRepository.GetByID(id)
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}

		return nil, err
	}

	response := ToOrderResponse(order)

	return &response, nil
}

// GetOrders returns all orders.
func (s *Service) GetOrders() ([]OrderListResponse, error) {

	orders, err := s.orderRepository.GetAll()
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

	user, err := s.userRepository.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, ErrUserNotFound
	}

	orders, err := s.orderRepository.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	return ToOrderList(orders), nil
}

// UpdateOrderStatus updates the status of an order.
func (s *Service) UpdateOrderStatus(
	id uuid.UUID,
	req UpdateOrderStatusRequest,
) (*OrderResponse, error) {

	order, err := s.orderRepository.GetByID(id)
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}

		return nil, err
	}

	if err := s.validateStatusTransition(
		order.Status,
		req.Status,
	); err != nil {
		return nil, err
	}

	if err := s.orderRepository.UpdateStatus(
		order.ID,
		req.Status,
	); err != nil {
		return nil, err
	}

	order, err = s.orderRepository.GetByID(order.ID)
	if err != nil {
		return nil, err
	}

	response := ToOrderResponse(order)

	return &response, nil
}

// validateStatusTransition validates whether an order
// is allowed to move from one state to another.
func (s *Service) validateStatusTransition(
	current models.OrderStatus,
	next models.OrderStatus,
) error {

	// Same status
	if current == next {
		return nil
	}

	// Terminal states
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

// CancelOrder cancels an order and releases any reserved inventory.
func (s *Service) CancelOrder(
	id uuid.UUID,
) (*OrderResponse, error) {

	order, err := s.orderRepository.GetByID(id)
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}

		return nil, err
	}

	// Only orders before shipping can be cancelled.
	switch order.Status {

	case models.OrderCancelled:
		return nil, ErrOrderAlreadyCancelled

	case models.OrderDelivered:
		return nil, ErrOrderAlreadyCompleted

	case models.OrderShipped:
		return nil, ErrOrderCannotBeCancelled
	}

	tx := s.orderRepository.Begin()

	if tx.Error != nil {
		return nil, tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// -----------------------------------------
	// Release Reserved Inventory
	// -----------------------------------------

	for _, item := range order.Items {

		inventory, err := s.inventoryRepository.GetByProductID(
			item.ProductID,
		)
		if err != nil {

			if errors.Is(err, gorm.ErrRecordNotFound) {
				tx.Rollback()
				return nil, ErrInventoryNotFound
			}

			tx.Rollback()
			return nil, err
		}

		if inventory.ReservedQuantity < item.Quantity {
			tx.Rollback()
			return nil, ErrInsufficientReserved
		}

		if err := s.inventoryRepository.ReleaseReservedStock(
			item.ProductID,
			item.Quantity,
		); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// -----------------------------------------
	// Update Order Status
	// -----------------------------------------

	if err := tx.
		Model(&models.Order{}).
		Where("id = ?", order.ID).
		Update("status", models.OrderCancelled).
		Error; err != nil {

		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	order, err = s.orderRepository.GetByID(order.ID)
	if err != nil {
		return nil, err
	}

	response := ToOrderResponse(order)

	return &response, nil
}