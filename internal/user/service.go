package user

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/repository"
	"github.com/khushidesai23/Enterprise-Order-Processing/pkg/utils"
	"github.com/khushidesai23/Enterprise-Order-Processing/pkg/validator"
)

type Service interface {
	Create(ctx context.Context, req CreateUserRequest) (*UserResponse, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateUserRequest) (*UserResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*UserResponse, error)
	List(ctx context.Context) ([]UserResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type service struct {
	repository repository.UserRepository
}

func NewService(
	repository repository.UserRepository,
) Service {

	return &service{
		repository: repository,
	}
}

func (s *service) Create(
	ctx context.Context,
	req CreateUserRequest,
) (*UserResponse, error) {

	if err := validator.Validate(req); err != nil {
		return nil, err
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	existing, err := s.repository.GetByEmail(ctx, req.Email)

	if err != nil {
		return nil, err
	}

	if existing != nil {
		return nil, ErrUserAlreadyExists
	}

	hashedPassword, err := utils.HashPassword(req.Password)

	if err != nil {
		return nil, err
	}

	user := &models.User{
		FirstName: strings.TrimSpace(req.FirstName),
		LastName: strings.TrimSpace(req.LastName),
		Email: req.Email,
		Password: hashedPassword,
		IsActive: true,
	}

	if err := s.repository.Create(ctx, user); err != nil {
		return nil, err
	}

	resp := ToResponse(user)

	return &resp, nil
}

func (s *service) Update(
	ctx context.Context,
	id uuid.UUID,
	req UpdateUserRequest,
) (*UserResponse, error) {

	user, err := s.repository.GetByID(ctx, id)

	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, ErrUserNotFound
	}

	if strings.TrimSpace(req.FirstName) != "" {
		user.FirstName = strings.TrimSpace(req.FirstName)
	}

	if strings.TrimSpace(req.LastName) != "" {
		user.LastName = strings.TrimSpace(req.LastName)
	}

	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if err := s.repository.Update(ctx, user); err != nil {
		return nil, err
	}

	resp := ToResponse(user)

	return &resp, nil
}

func (s *service) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*UserResponse, error) {

	user, err := s.repository.GetByID(ctx, id)

	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, ErrUserNotFound
	}

	resp := ToResponse(user)

	return &resp, nil
}

func (s *service) List(
	ctx context.Context,
) ([]UserResponse, error) {

	users, err := s.repository.List(ctx)

	if err != nil {
		return nil, err
	}

	return ToResponseList(users), nil
}

func (s *service) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {

	user, err := s.repository.GetByID(ctx, id)

	if err != nil {
		return err
	}

	if user == nil {
		return ErrUserNotFound
	}

	return s.repository.Delete(ctx, id)
}

func IsBusinessError(err error) bool {

	return errors.Is(err, ErrUserAlreadyExists) ||
		errors.Is(err, ErrUserNotFound) ||
		errors.Is(err, ErrInvalidPassword)
}