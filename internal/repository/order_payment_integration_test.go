//go:build integration

package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
	testintegration "github.com/khushidesai23/Enterprise-Order-Processing/internal/test/integration"
)

func TestOrderPaymentRepositoryTransactions(t *testing.T) {
	db := testintegration.NewTestDatabase(t)
	testintegration.CleanupDatabase(t, db)
	creator := testintegration.GormCreator(db.DB)
	ctx := context.Background()

	user := testintegration.CreateUserFixture(
		t,
		creator,
		"Order",
		"User",
		"order-user@example.com",
		"password123",
	)
	category := testintegration.CreateCategoryFixture(t, creator, "Orders")
	product := testintegration.CreateProductFixture(t, creator, category.ID, "Keyboard", "SKU-ORDER-1", 120)

	orderRepo := NewOrderRepository(db.DB)
	orderItemRepo := NewOrderItemRepository(db.DB)
	paymentRepo := NewPaymentRepository(db.DB)
	webhookRepo := NewPaymentWebhookRepository(db.DB)

	t.Run("rollback removes inserted order", func(t *testing.T) {
		tx := orderRepo.Begin(ctx)
		order := &models.Order{
			UserID:      user.ID,
			Status:      models.OrderCreated,
			TotalAmount: 120,
		}

		require.NoError(t, orderRepo.Create(ctx, tx, order))
		require.NoError(t, orderItemRepo.CreateMany(ctx, tx, []models.OrderItem{
			{
				OrderID:   order.ID,
				ProductID: product.ID,
				Quantity:  1,
				Price:     120,
			},
		}))
		require.NoError(t, tx.Rollback().Error)

		orders, err := orderRepo.GetByUserID(ctx, user.ID)
		require.NoError(t, err)
		assert.Empty(t, orders)
	})

	t.Run("commit persists order and payment state", func(t *testing.T) {
		tx := orderRepo.Begin(ctx)
		order := &models.Order{
			UserID:      user.ID,
			Status:      models.OrderCreated,
			TotalAmount: 120,
		}

		require.NoError(t, orderRepo.Create(ctx, tx, order))
		require.NoError(t, orderItemRepo.CreateMany(ctx, tx, []models.OrderItem{
			{
				OrderID:   order.ID,
				ProductID: product.ID,
				Quantity:  1,
				Price:     120,
			},
		}))
		require.NoError(t, orderRepo.UpdateTotalAmount(ctx, tx, order.ID, 120))

		paymentModel := &models.Payment{
			OrderID:        order.ID,
			GatewayOrderID: "gateway-order-1",
			Status:         models.PaymentPending,
			Amount:         120,
			Currency:       "INR",
			Gateway:        "RAZORPAY",
		}
		require.NoError(t, paymentRepo.Create(ctx, tx, paymentModel))
		require.NoError(t, tx.Commit().Error)

		fetchedOrder, err := orderRepo.GetByID(ctx, order.ID)
		require.NoError(t, err)
		require.Len(t, fetchedOrder.Items, 1)
		require.NotNil(t, fetchedOrder.Payment)

		transactionID := "txn-1"
		tx = paymentRepo.Begin(ctx)
		require.NoError(t, paymentRepo.CompletePayment(ctx, tx, paymentModel.ID, &transactionID, models.PaymentSuccess))
		require.NoError(t, webhookRepo.Create(ctx, tx, &models.PaymentWebhook{
			Gateway:        "RAZORPAY",
			PayloadID:      "payload-1",
			Event:          "payment.captured",
			GatewayOrderID: &paymentModel.GatewayOrderID,
			Status:         models.WebhookPending,
			RawPayload:     []byte(`{"event":"payment.captured"}`),
		}))
		require.NoError(t, webhookRepo.MarkProcessed(ctx, tx, "payload-1", fetchedOrder.CreatedAt))
		require.NoError(t, tx.Commit().Error)

		fetchedPayment, err := paymentRepo.GetByID(ctx, paymentModel.ID)
		require.NoError(t, err)
		require.NotNil(t, fetchedPayment.TransactionID)
		assert.Equal(t, models.PaymentSuccess, fetchedPayment.Status)
		assert.Equal(t, transactionID, *fetchedPayment.TransactionID)

		duplicatePayment := *paymentModel
		duplicatePayment.GatewayOrderID = "gateway-order-2"
		err = paymentRepo.Create(ctx, paymentRepo.Begin(ctx), &duplicatePayment)
		require.Error(t, err)
	})
}
