package repository

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uuid.UUID) error

	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByIDTx(ctx context.Context, tx *gorm.DB, id uuid.UUID) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	List(ctx context.Context) ([]models.User, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) Create(
	ctx context.Context,
	user *models.User,
) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) Update(
	ctx context.Context,
	user *models.User,
) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {
	return r.db.WithContext(ctx).
		Delete(&models.User{}, "id = ?", id).
		Error
}

func (r *userRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*models.User, error) {

	var user models.User

	err := r.db.WithContext(ctx).
		First(&user, "id = ?", id).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) GetByIDTx(
	ctx context.Context,
	tx *gorm.DB,
	id uuid.UUID,
) (*models.User, error) {

	var user models.User

	err := tx.WithContext(ctx).
		First(&user, "id = ?", id).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) GetByEmail(
	ctx context.Context,
	email string,
) (*models.User, error) {

	var user models.User

	err := r.db.WithContext(ctx).
		Where("email = ?", email).
		First(&user).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) List(
	ctx context.Context,
) ([]models.User, error) {

	var users []models.User

	err := r.db.WithContext(ctx).
		Order("created_at desc").
		Find(&users).
		Error

	return users, err
}
