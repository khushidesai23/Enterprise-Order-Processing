//go:build integration

package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	testintegration "github.com/khushidesai23/Enterprise-Order-Processing/internal/test/integration"
)

func TestUserRepositoryIntegrationCRUDAndConstraints(t *testing.T) {
	db := testintegration.NewTestDatabase(t)
	testintegration.CleanupDatabase(t, db)

	repo := NewUserRepository(db.DB)
	ctx := context.Background()

	user := testintegration.CreateUserFixture(
		t,
		testintegration.GormCreator(db.DB),
		"Khushi",
		"Desai",
		"khushi.repository@example.com",
		"password123",
	)

	fetchedByEmail, err := repo.GetByEmail(ctx, user.Email)
	require.NoError(t, err)
	require.NotNil(t, fetchedByEmail)
	assert.Equal(t, user.ID, fetchedByEmail.ID)

	fetchedByID, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	require.NotNil(t, fetchedByID)
	assert.Equal(t, user.Email, fetchedByID.Email)

	users, err := repo.List(ctx)
	require.NoError(t, err)
	require.Len(t, users, 1)

	fetchedByID.FirstName = "Updated"
	err = repo.Update(ctx, fetchedByID)
	require.NoError(t, err)

	updated, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, "Updated", updated.FirstName)

	duplicate := *fetchedByID
	duplicate.ID = user.ID
	duplicate.Email = user.Email

	err = repo.Create(ctx, &duplicate)
	require.Error(t, err)

	err = repo.Delete(ctx, user.ID)
	require.NoError(t, err)

	deleted, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Nil(t, deleted)
}
