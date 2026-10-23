//go:build e2e

package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
	ordermodule "github.com/khushidesai23/Enterprise-Order-Processing/internal/modules/order"
	paymentmodule "github.com/khushidesai23/Enterprise-Order-Processing/internal/modules/payment"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/repository"
	testintegration "github.com/khushidesai23/Enterprise-Order-Processing/internal/test/integration"
)

func TestCheckoutFlowEndToEnd(t *testing.T) {
	env := testintegration.NewAppEnv(t)

	type userResponse struct {
		ID        string `json:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		IsActive  bool   `json:"is_active"`
	}

	type loginResponse struct {
		Token     string `json:"token"`
		TokenType string `json:"token_type"`
		User      struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		} `json:"user"`
	}

	type categoryResponse struct {
		ID string `json:"id"`
	}

	type productResponse struct {
		ID  string `json:"id"`
		SKU string `json:"sku"`
	}

	type inventoryResponse struct {
		ProductID         string `json:"product_id"`
		AvailableQuantity int    `json:"available_quantity"`
		ReservedQuantity  int    `json:"reserved_quantity"`
	}

	type orderResponse struct {
		ID          string             `json:"id"`
		Status      models.OrderStatus `json:"status"`
		TotalAmount float64            `json:"total_amount"`
	}

	type paymentResponse struct {
		PaymentID      string  `json:"payment_id"`
		OrderID        string  `json:"order_id"`
		Amount         float64 `json:"amount"`
		GatewayOrderID string  `json:"gateway_order_id"`
	}

	createUserRecorder := testintegration.JSONRequest(
		t,
		env.Router,
		http.MethodPost,
		"/api/v1/users",
		map[string]interface{}{
			"first_name": "Khushi",
			"last_name":  "Desai",
			"email":      "e2e-user@example.com",
			"password":   "password123",
		},
		"",
		nil,
	)
	require.Equal(t, http.StatusCreated, createUserRecorder.Code, createUserRecorder.Body.String())

	createUserAPI := testintegration.DecodeAPIResponse(t, createUserRecorder)
	var createdUser userResponse
	testintegration.DecodeData(t, createUserAPI.Data, &createdUser)
	assert.Equal(t, "e2e-user@example.com", createdUser.Email)

	loginRecorder := testintegration.JSONRequest(
		t,
		env.Router,
		http.MethodPost,
		"/api/v1/auth/login",
		map[string]interface{}{
			"email":    "e2e-user@example.com",
			"password": "password123",
		},
		"",
		nil,
	)
	require.Equal(t, http.StatusOK, loginRecorder.Code, loginRecorder.Body.String())

	loginAPI := testintegration.DecodeAPIResponse(t, loginRecorder)
	var login loginResponse
	testintegration.DecodeData(t, loginAPI.Data, &login)
	require.NotEmpty(t, login.Token)

	meRecorder := testintegration.JSONRequest(
		t,
		env.Router,
		http.MethodGet,
		"/api/v1/auth/me",
		nil,
		login.Token,
		nil,
	)
	require.Equal(t, http.StatusOK, meRecorder.Code, meRecorder.Body.String())

	loginAs := func(email string) string {
		t.Helper()
		recorder := testintegration.JSONRequest(t, env.Router, http.MethodPost, "/api/v1/auth/login",
			map[string]interface{}{"email": email, "password": "password123"}, "", nil)
		require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
		var response loginResponse
		testintegration.DecodeData(t, testintegration.DecodeAPIResponse(t, recorder).Data, &response)
		return response.Token
	}

	statusOf := func(method string, path string, body interface{}, token string) int {
		t.Helper()
		return testintegration.JSONRequest(t, env.Router, method, path, body, token, nil).Code
	}

	creator := testintegration.GormCreator(env.DB.DB)
	testintegration.CreateAdminFixture(t, creator, "E2E", "Admin", "e2e-admin@example.com", "password123")
	adminToken := loginAs("e2e-admin@example.com")
	testintegration.CreateUserFixture(t, creator, "E2E", "Intruder", "e2e-intruder@example.com", "password123")
	intruderToken := loginAs("e2e-intruder@example.com")

	// Self-registration cannot grant privileges.
	assert.Equal(t, http.StatusForbidden, statusOf(http.MethodPost, "/api/v1/categories",
		map[string]interface{}{"name": "Nope", "description": "customer write"}, login.Token))
	assert.Equal(t, http.StatusForbidden, statusOf(http.MethodGet, "/api/v1/users", nil, login.Token))
	assert.Equal(t, http.StatusNotFound, statusOf(http.MethodGet, "/api/v1/users/"+createdUser.ID, nil, intruderToken))
	assert.Equal(t, http.StatusForbidden, statusOf(http.MethodPut, "/api/v1/users/"+createdUser.ID,
		map[string]interface{}{"is_active": false}, login.Token))

	createCategoryRecorder := testintegration.JSONRequest(
		t,
		env.Router,
		http.MethodPost,
		"/api/v1/categories",
		map[string]interface{}{
			"name":        "E2E Category",
			"description": "category for e2e",
		},
		adminToken,
		nil,
	)
	require.Equal(t, http.StatusCreated, createCategoryRecorder.Code, createCategoryRecorder.Body.String())

	createCategoryAPI := testintegration.DecodeAPIResponse(t, createCategoryRecorder)
	var category categoryResponse
	testintegration.DecodeData(t, createCategoryAPI.Data, &category)
	categoryID := uuid.MustParse(category.ID)

	createProductRecorder := testintegration.JSONRequest(
		t,
		env.Router,
		http.MethodPost,
		"/api/v1/products",
		map[string]interface{}{
			"name":        "E2E Laptop",
			"description": "product for e2e",
			"sku":         "E2E-SKU-1",
			"price":       1499.5,
			"category_id": categoryID,
		},
		adminToken,
		nil,
	)
	require.Equal(t, http.StatusCreated, createProductRecorder.Code, createProductRecorder.Body.String())

	createProductAPI := testintegration.DecodeAPIResponse(t, createProductRecorder)
	var product productResponse
	testintegration.DecodeData(t, createProductAPI.Data, &product)
	productID := uuid.MustParse(product.ID)

	assert.Equal(t, http.StatusForbidden, statusOf(http.MethodPost, "/api/v1/inventory",
		map[string]interface{}{"product_id": productID, "available_quantity": 1000}, login.Token))

	addInventoryRecorder := testintegration.JSONRequest(
		t,
		env.Router,
		http.MethodPost,
		"/api/v1/inventory",
		map[string]interface{}{
			"product_id":         productID,
			"available_quantity": 10,
		},
		adminToken,
		nil,
	)
	require.Equal(t, http.StatusCreated, addInventoryRecorder.Code, addInventoryRecorder.Body.String())

	addInventoryAPI := testintegration.DecodeAPIResponse(t, addInventoryRecorder)
	var inventory inventoryResponse
	testintegration.DecodeData(t, addInventoryAPI.Data, &inventory)
	assert.Equal(t, 10, inventory.AvailableQuantity)

	createOrderRecorder := testintegration.JSONRequest(
		t,
		env.Router,
		http.MethodPost,
		"/api/v1/orders",
		map[string]interface{}{
			"items": []map[string]interface{}{
				{
					"product_id": productID,
					"quantity":   2,
				},
			},
		},
		login.Token,
		nil,
	)
	require.Equal(t, http.StatusCreated, createOrderRecorder.Code, createOrderRecorder.Body.String())

	createOrderAPI := testintegration.DecodeAPIResponse(t, createOrderRecorder)
	var order orderResponse
	testintegration.DecodeData(t, createOrderAPI.Data, &order)
	orderID := uuid.MustParse(order.ID)
	assert.Equal(t, models.OrderCreated, order.Status)

	// Another customer can neither see, cancel nor pay for this order, and
	// gets the same 404 as for an order that does not exist.
	orderPath := fmt.Sprintf("/api/v1/orders/%s", orderID)
	assert.Equal(t, http.StatusNotFound, statusOf(http.MethodGet, orderPath, nil, intruderToken))
	assert.Equal(t, http.StatusNotFound, statusOf(http.MethodPatch, orderPath+"/cancel", nil, intruderToken))
	assert.Equal(t, http.StatusNotFound, statusOf(http.MethodPost, "/api/v1/payments",
		map[string]interface{}{"order_id": orderID}, intruderToken))

	intruderOrdersRecorder := testintegration.JSONRequest(t, env.Router, http.MethodGet, "/api/v1/orders", nil, intruderToken, nil)
	require.Equal(t, http.StatusOK, intruderOrdersRecorder.Code, intruderOrdersRecorder.Body.String())
	var intruderOrders []ordermodule.OrderListResponse
	testintegration.DecodeData(t, testintegration.DecodeAPIResponse(t, intruderOrdersRecorder).Data, &intruderOrders)
	assert.Empty(t, intruderOrders)

	createPaymentRecorder := testintegration.JSONRequest(
		t,
		env.Router,
		http.MethodPost,
		"/api/v1/payments",
		map[string]interface{}{
			"order_id": orderID,
		},
		login.Token,
		nil,
	)
	require.Equal(t, http.StatusCreated, createPaymentRecorder.Code, createPaymentRecorder.Body.String())

	createPaymentAPI := testintegration.DecodeAPIResponse(t, createPaymentRecorder)
	var checkout paymentResponse
	testintegration.DecodeData(t, createPaymentAPI.Data, &checkout)
	paymentID := uuid.MustParse(checkout.PaymentID)

	assert.Equal(t, http.StatusNotFound, statusOf(http.MethodGet, "/api/v1/payments/"+paymentID.String(), nil, intruderToken))
	assert.Equal(t, http.StatusNotFound, statusOf(http.MethodGet, "/api/v1/payments/order/"+orderID.String(), nil, intruderToken))
	assert.Equal(t, http.StatusOK, statusOf(http.MethodGet, "/api/v1/payments/"+paymentID.String(), nil, login.Token))
	assert.Equal(t, http.StatusForbidden, statusOf(http.MethodGet, "/api/v1/payments", nil, login.Token))
	assert.Equal(t, http.StatusForbidden, statusOf(http.MethodGet, "/api/v1/payments/summary", nil, login.Token))

	webhookBody := map[string]interface{}{
		"event":      "payment.captured",
		"account_id": "acc_test_1",
		"created_at": 1724000000,
		"payload": map[string]interface{}{
			"id": "payload_test_1",
			"payment": map[string]interface{}{
				"entity": map[string]interface{}{
					"id":          "pay_test_1",
					"order_id":    checkout.GatewayOrderID,
					"status":      "captured",
					"amount":      int(checkout.Amount * 100),
					"currency":    "INR",
					"method":      "card",
					"description": "test payment",
					"email":       createdUser.Email,
					"contact":     "9999999999",
				},
			},
		},
	}

	webhookPayload, err := json.Marshal(webhookBody)
	require.NoError(t, err)

	webhookRecorder := testintegration.JSONRequest(
		t,
		env.Router,
		http.MethodPost,
		"/api/v1/payments/webhook",
		json.RawMessage(webhookPayload),
		"",
		map[string]string{
			"X-Razorpay-Signature": env.Gateway.SignWebhook(webhookPayload),
			"X-Razorpay-Event-Id":  "event_test_1",
		},
	)
	require.Equal(t, http.StatusOK, webhookRecorder.Code, webhookRecorder.Body.String())

	// Customers cannot drive fulfilment or refund themselves, not even on
	// their own order.
	assert.Equal(t, http.StatusForbidden, statusOf(http.MethodPatch, orderPath+"/status",
		map[string]interface{}{"status": models.OrderPacked}, login.Token))
	assert.Equal(t, http.StatusForbidden, statusOf(http.MethodPost, "/api/v1/payments/"+paymentID.String()+"/refund", nil, login.Token))

	packRecorder := testintegration.JSONRequest(
		t,
		env.Router,
		http.MethodPatch,
		fmt.Sprintf("/api/v1/orders/%s/status", orderID),
		map[string]interface{}{
			"status": models.OrderPacked,
		},
		adminToken,
		nil,
	)
	require.Equal(t, http.StatusOK, packRecorder.Code, packRecorder.Body.String())

	shipRecorder := testintegration.JSONRequest(
		t,
		env.Router,
		http.MethodPatch,
		fmt.Sprintf("/api/v1/orders/%s/status", orderID),
		map[string]interface{}{
			"status": models.OrderShipped,
		},
		adminToken,
		nil,
	)
	require.Equal(t, http.StatusOK, shipRecorder.Code, shipRecorder.Body.String())

	orderRepo := repository.NewOrderRepository(env.DB.DB)
	paymentRepo := repository.NewPaymentRepository(env.DB.DB)
	inventoryRepo := repository.NewInventoryRepository(env.DB.DB)
	webhookRepo := repository.NewPaymentWebhookRepository(env.DB.DB)
	ctx := context.Background()

	finalOrder, err := orderRepo.GetByID(ctx, orderID)
	require.NoError(t, err)
	assert.Equal(t, models.OrderShipped, finalOrder.Status)
	assert.Equal(t, 1499.5*2, finalOrder.TotalAmount)

	finalPayment, err := paymentRepo.GetByID(ctx, paymentID)
	require.NoError(t, err)
	assert.Equal(t, models.PaymentSuccess, finalPayment.Status)
	require.NotNil(t, finalPayment.TransactionID)
	assert.Equal(t, "pay_test_1", *finalPayment.TransactionID)

	finalInventory, err := inventoryRepo.GetByProductID(ctx, productID)
	require.NoError(t, err)
	assert.Equal(t, 8, finalInventory.AvailableQuantity)
	assert.Equal(t, 0, finalInventory.ReservedQuantity)

	webhookRecord, err := webhookRepo.GetByPayloadID(ctx, "event_test_1")
	require.NoError(t, err)
	assert.Equal(t, models.WebhookProcessed, webhookRecord.Status)

	orderHTTPRecorder := testintegration.JSONRequest(
		t,
		env.Router,
		http.MethodGet,
		fmt.Sprintf("/api/v1/orders/%s", orderID),
		nil,
		login.Token,
		nil,
	)
	require.Equal(t, http.StatusOK, orderHTTPRecorder.Code, orderHTTPRecorder.Body.String())

	orderHTTPAPI := testintegration.DecodeAPIResponse(t, orderHTTPRecorder)
	var fetchedOrder ordermodule.OrderResponse
	testintegration.DecodeData(t, orderHTTPAPI.Data, &fetchedOrder)
	assert.Equal(t, models.OrderShipped, fetchedOrder.Status)

	paymentByOrderRecorder := testintegration.JSONRequest(
		t,
		env.Router,
		http.MethodGet,
		fmt.Sprintf("/api/v1/payments/order/%s", orderID),
		nil,
		login.Token,
		nil,
	)
	require.Equal(t, http.StatusOK, paymentByOrderRecorder.Code, paymentByOrderRecorder.Body.String())

	paymentByOrderAPI := testintegration.DecodeAPIResponse(t, paymentByOrderRecorder)
	var fetchedPayment paymentmodule.PaymentResponse
	testintegration.DecodeData(t, paymentByOrderAPI.Data, &fetchedPayment)
	assert.Equal(t, models.PaymentSuccess, fetchedPayment.Status)
}
