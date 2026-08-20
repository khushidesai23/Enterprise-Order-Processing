package order

import (
	"time"

	"github.com/google/uuid"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
)

func ToOrderModel(req CreateOrderRequest) *models.Order {
	return &models.Order{
		UserID:      req.UserID,
		Status:      models.OrderCreated,
		TotalAmount: 0,
	}
}

func BuildOrderItem(
	orderID uuid.UUID,
	productID uuid.UUID,
	quantity int,
	price float64,
) *models.OrderItem {
	return &models.OrderItem{
		OrderID:   orderID,
		ProductID: productID,
		Quantity:  quantity,
		Price:     price,
	}
}

func ToOrderResponse(order *models.Order) OrderResponse {

	response := OrderResponse{
		ID:          order.ID,
		UserID:      order.UserID,
		Status:      order.Status,
		TotalAmount: order.TotalAmount,
		CreatedAt:   order.CreatedAt.Format(time.RFC3339),
	}

	items := make([]OrderItemResponse, 0, len(order.Items))

	for _, item := range order.Items {

		orderItem := OrderItemResponse{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
			SubTotal:  item.Price * float64(item.Quantity),
		}

		orderItem.ProductName = item.Product.Name
		orderItem.SKU = item.Product.SKU

		items = append(items, orderItem)
	}

	response.Items = items

	return response
}

func ToOrderListResponse(order *models.Order) OrderListResponse {
	return OrderListResponse{
		ID:          order.ID,
		UserID:      order.UserID,
		Status:      order.Status,
		TotalAmount: order.TotalAmount,
		TotalItems:  len(order.Items),
	}
}

func ToOrderList(orders []models.Order) []OrderListResponse {

	response := make([]OrderListResponse, 0, len(orders))

	for i := range orders {
		response = append(response, ToOrderListResponse(&orders[i]))
	}

	return response
}

func CalculateOrderTotal(items []models.OrderItem) float64 {

	var total float64

	for _, item := range items {
		total += item.Price * float64(item.Quantity)
	}

	return total
}
