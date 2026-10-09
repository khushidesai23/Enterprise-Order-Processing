//go:build integration

package order_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
	ordermodule "github.com/khushidesai23/Enterprise-Order-Processing/internal/modules/order"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/repository"
	testintegration "github.com/khushidesai23/Enterprise-Order-Processing/internal/test/integration"
)

func TestOrderServiceIntegrationFlow(t *testing.T) {
	db := testintegration.NewTestDatabase(t)
	testintegration.CleanupDatabase(t, db)
	creator := testintegration.GormCreator(db.DB)
	ctx := context.Background()

	user := testintegration.CreateUserFixture(
		t,
		creator,
		"Flow",
		"User",
		"flow-user@example.com",
		"password123",
	)
	category := testintegration.CreateCategoryFixture(t, creator, "Flow Category")
	product := testintegration.CreateProductFixture(
		t,
		creator,
		category.ID,
		"Flow Product",
		"SKU-FLOW-1",
		99,
	)
	testintegration.CreateInventoryFixture(t, creator, product.ID, 10, 0)

	orderRepo := repository.NewOrderRepository(db.DB)
	orderItemRepo := repository.NewOrderItemRepository(db.DB)
	userRepo := repository.NewUserRepository(db.DB)
	productRepo := repository.NewProductRepository(db.DB)
	inventoryRepo := repository.NewInventoryRepository(db.DB)

	service := ordermodule.NewService(
		orderRepo,
		orderItemRepo,
		userRepo,
		productRepo,
		inventoryRepo,
		testintegration.NewTestLogger(t),
	)

	t.Run("creating an order reserves inventory", func(t *testing.T) {
		orderResp, err := service.CreateOrder(ctx, ordermodule.CreateOrderRequest{
			UserID: user.ID,
			Items: []ordermodule.CreateOrderItemRequest{
				{
					ProductID: product.ID,
					Quantity:  2,
				},
			},
		})

		require.NoError(t, err)
		assert.Equal(t, models.OrderCreated, orderResp.Status)
		assert.Equal(t, 198.0, orderResp.TotalAmount)

		inventory, err := inventoryRepo.GetByProductID(ctx, product.ID)
		require.NoError(t, err)
		assert.Equal(t, 8, inventory.AvailableQuantity)
		assert.Equal(t, 2, inventory.ReservedQuantity)
	})

	t.Run("cancelling an order releases reserved inventory", func(t *testing.T) {
		testintegration.CleanupDatabase(t, db)
		user = testintegration.CreateUserFixture(t, creator, "Cancel", "User", "cancel-user@example.com", "password123")
		category = testintegration.CreateCategoryFixture(t, creator, "Cancel Category")
		product = testintegration.CreateProductFixture(t, creator, category.ID, "Cancel Product", "SKU-CANCEL-1", 75)
		testintegration.CreateInventoryFixture(t, creator, product.ID, 6, 0)

		orderResp, err := service.CreateOrder(ctx, ordermodule.CreateOrderRequest{
			UserID: user.ID,
			Items: []ordermodule.CreateOrderItemRequest{
				{ProductID: product.ID, Quantity: 2},
			},
		})
		require.NoError(t, err)

		cancelled, err := service.CancelOrder(ctx, orderResp.ID)
		require.NoError(t, err)
		assert.Equal(t, models.OrderCancelled, cancelled.Status)

		inventory, err := inventoryRepo.GetByProductID(ctx, product.ID)
		require.NoError(t, err)
		assert.Equal(t, 6, inventory.AvailableQuantity)
		assert.Equal(t, 0, inventory.ReservedQuantity)
	})

	t.Run("shipping an order confirms reserved inventory", func(t *testing.T) {
		testintegration.CleanupDatabase(t, db)
		user = testintegration.CreateUserFixture(t, creator, "Ship", "User", "ship-user@example.com", "password123")
		category = testintegration.CreateCategoryFixture(t, creator, "Ship Category")
		product = testintegration.CreateProductFixture(t, creator, category.ID, "Ship Product", "SKU-SHIP-1", 110)
		testintegration.CreateInventoryFixture(t, creator, product.ID, 10, 0)

		orderResp, err := service.CreateOrder(ctx, ordermodule.CreateOrderRequest{
			UserID: user.ID,
			Items: []ordermodule.CreateOrderItemRequest{
				{ProductID: product.ID, Quantity: 2},
			},
		})
		require.NoError(t, err)

		err = db.DB.Transaction(func(tx *gorm.DB) error {
			if err := service.MarkOrderPaymentPending(tx, orderResp.ID); err != nil {
				return err
			}
			return service.MarkOrderPaid(tx, orderResp.ID)
		})
		require.NoError(t, err)

		_, err = service.UpdateOrderStatus(ctx, orderResp.ID, ordermodule.UpdateOrderStatusRequest{
			Status: models.OrderPacked,
		})
		require.NoError(t, err)

		shipped, err := service.UpdateOrderStatus(ctx, orderResp.ID, ordermodule.UpdateOrderStatusRequest{
			Status: models.OrderShipped,
		})
		require.NoError(t, err)
		assert.Equal(t, models.OrderShipped, shipped.Status)

		inventory, err := inventoryRepo.GetByProductID(ctx, product.ID)
		require.NoError(t, err)
		assert.Equal(t, 8, inventory.AvailableQuantity)
		assert.Equal(t, 0, inventory.ReservedQuantity)
	})

	t.Run("if a later item fails the transaction rolls back", func(t *testing.T) {
		testintegration.CleanupDatabase(t, db)
		user = testintegration.CreateUserFixture(t, creator, "Rollback", "User", "rollback-user@example.com", "password123")
		category = testintegration.CreateCategoryFixture(t, creator, "Rollback Category")
		firstProduct := testintegration.CreateProductFixture(t, creator, category.ID, "Rollback Product 1", "SKU-ROLL-1", 45)
		secondProduct := testintegration.CreateProductFixture(t, creator, category.ID, "Rollback Product 2", "SKU-ROLL-2", 55)
		testintegration.CreateInventoryFixture(t, creator, firstProduct.ID, 10, 0)
		testintegration.CreateInventoryFixture(t, creator, secondProduct.ID, 1, 0)

		orderResp, err := service.CreateOrder(ctx, ordermodule.CreateOrderRequest{
			UserID: user.ID,
			Items: []ordermodule.CreateOrderItemRequest{
				{ProductID: firstProduct.ID, Quantity: 2},
				{ProductID: secondProduct.ID, Quantity: 5},
			},
		})

		require.Error(t, err)
		assert.Nil(t, orderResp)
		assert.ErrorIs(t, err, ordermodule.ErrInsufficientStock)

		firstInventory, err := inventoryRepo.GetByProductID(ctx, firstProduct.ID)
		require.NoError(t, err)
		assert.Equal(t, 10, firstInventory.AvailableQuantity)
		assert.Equal(t, 0, firstInventory.ReservedQuantity)

		secondInventory, err := inventoryRepo.GetByProductID(ctx, secondProduct.ID)
		require.NoError(t, err)
		assert.Equal(t, 1, secondInventory.AvailableQuantity)
		assert.Equal(t, 0, secondInventory.ReservedQuantity)

		orders, err := orderRepo.GetByUserID(ctx, user.ID)
		require.NoError(t, err)
		assert.Empty(t, orders)
	})

	t.Run("concurrent reservations never oversell", func(t *testing.T) {
		testintegration.CleanupDatabase(t, db)
		loadUser := testintegration.CreateUserFixture(t, creator, "Concurrent", "User", "concurrent-user@example.com", "password123")
		loadCategory := testintegration.CreateCategoryFixture(t, creator, "Concurrent Category")
		loadProduct := testintegration.CreateProductFixture(t, creator, loadCategory.ID, "Concurrent Product", "SKU-CONCURRENT-1", 1)
		testintegration.CreateInventoryFixture(t, creator, loadProduct.ID, 10, 0)

		const requests = 50
		results := make(chan error, requests)
		var workers sync.WaitGroup
		workers.Add(requests)
		for range requests {
			go func() {
				defer workers.Done()
				_, err := service.CreateOrder(ctx, ordermodule.CreateOrderRequest{
					UserID: loadUser.ID,
					Items:  []ordermodule.CreateOrderItemRequest{{ProductID: loadProduct.ID, Quantity: 1}},
				})
				results <- err
			}()
		}
		workers.Wait()
		close(results)

		var created, rejected int
		for err := range results {
			switch {
			case err == nil:
				created++
			case errors.Is(err, ordermodule.ErrInsufficientStock):
				rejected++
			default:
				t.Errorf("unexpected order error: %v", err)
			}
		}
		assert.Equal(t, 10, created)
		assert.Equal(t, requests-10, rejected)

		inventory, err := inventoryRepo.GetByProductID(ctx, loadProduct.ID)
		require.NoError(t, err)
		assert.Equal(t, 0, inventory.AvailableQuantity)
		assert.Equal(t, 10, inventory.ReservedQuantity)
		orders, err := orderRepo.GetByUserID(ctx, loadUser.ID)
		require.NoError(t, err)
		assert.Len(t, orders, 10)
	})
}
