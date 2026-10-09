//go:build integration

package repository_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/repository"
	testintegration "github.com/khushidesai23/Enterprise-Order-Processing/internal/test/integration"
)

func TestCategoryProductInventoryIntegration(t *testing.T) {
	db := testintegration.NewTestDatabase(t)
	testintegration.CleanupDatabase(t, db)
	creator := testintegration.GormCreator(db.DB)
	ctx := context.Background()

	categoryRepo := repository.NewCategoryRepository(db.DB)
	productRepo := repository.NewProductRepository(db.DB)
	inventoryRepo := repository.NewInventoryRepository(db.DB)

	category := testintegration.CreateCategoryFixture(t, creator, "Electronics")
	product := testintegration.CreateProductFixture(
		t,
		creator,
		category.ID,
		"Laptop",
		"SKU-INT-1",
		1500,
	)
	testintegration.CreateInventoryFixture(t, creator, product.ID, 10, 0)

	exists, err := categoryRepo.ExistsByName(ctx, category.Name)
	require.NoError(t, err)
	assert.True(t, exists)

	hasProducts, err := categoryRepo.HasProducts(ctx, category.ID)
	require.NoError(t, err)
	assert.True(t, hasProducts)

	fetchedCategory, err := categoryRepo.GetByID(ctx, category.ID)
	require.NoError(t, err)
	require.Len(t, fetchedCategory.Products, 1)
	assert.Equal(t, product.ID, fetchedCategory.Products[0].ID)

	categoryDuplicate, err := categoryRepo.ExistsByNameExceptID(ctx, uuid.New(), category.Name)
	require.NoError(t, err)
	assert.True(t, categoryDuplicate)

	fetchedProduct, err := productRepo.GetByID(ctx, product.ID)
	require.NoError(t, err)
	require.NotNil(t, fetchedProduct.Inventory)
	assert.Equal(t, category.ID, fetchedProduct.Category.ID)

	inventory, err := inventoryRepo.GetByProductID(ctx, product.ID)
	require.NoError(t, err)
	assert.Equal(t, 10, inventory.AvailableQuantity)

	require.NoError(t, inventoryRepo.ReserveStock(ctx, product.ID, 3))
	require.NoError(t, inventoryRepo.ReleaseReservedStock(ctx, product.ID, 1))
	require.NoError(t, inventoryRepo.ConfirmReservedStock(ctx, product.ID, 2))

	inventory, err = inventoryRepo.GetByProductID(ctx, product.ID)
	require.NoError(t, err)
	assert.Equal(t, 8, inventory.AvailableQuantity)
	assert.Equal(t, 0, inventory.ReservedQuantity)

	invalidProduct := &models.Product{
		Name:       "Broken",
		SKU:        "SKU-INVALID",
		Price:      10,
		CategoryID: uuid.New(),
	}
	err = productRepo.Create(ctx, invalidProduct)
	require.Error(t, err)

	require.NoError(t, categoryRepo.Delete(ctx, category))

	_, err = categoryRepo.GetByID(ctx, category.ID)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestInventoryDecrementsNeverGoNegative(t *testing.T) {
	db := testintegration.NewTestDatabase(t)
	testintegration.CleanupDatabase(t, db)
	creator := testintegration.GormCreator(db.DB)
	ctx := context.Background()
	inventoryRepo := repository.NewInventoryRepository(db.DB)

	category := testintegration.CreateCategoryFixture(t, creator, "Decrement Category")
	product := testintegration.CreateProductFixture(t, creator, category.ID, "Decrement Product", "SKU-DECREMENT-1", 1)
	testintegration.CreateInventoryFixture(t, creator, product.ID, 10, 0)

	const requests = 50
	results := make(chan error, requests)
	var workers sync.WaitGroup
	workers.Add(requests)
	for range requests {
		go func() {
			defer workers.Done()
			results <- inventoryRepo.RemoveStock(ctx, product.ID, 1)
		}()
	}
	workers.Wait()
	close(results)

	var removed, rejected int
	for err := range results {
		switch {
		case err == nil:
			removed++
		case errors.Is(err, repository.ErrInsufficientQuantity):
			rejected++
		default:
			t.Errorf("unexpected error: %v", err)
		}
	}
	assert.Equal(t, 10, removed)
	assert.Equal(t, requests-10, rejected)

	inventory, err := inventoryRepo.GetByProductID(ctx, product.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, inventory.AvailableQuantity)

	assert.ErrorIs(t, inventoryRepo.ReleaseReservedStock(ctx, product.ID, 1), repository.ErrInsufficientQuantity)
	assert.ErrorIs(t, inventoryRepo.ConfirmReservedStock(ctx, product.ID, 1), repository.ErrInsufficientQuantity)
}
