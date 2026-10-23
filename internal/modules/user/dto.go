package user

import "github.com/khushidesai23/Enterprise-Order-Processing/internal/models"

type CreateUserRequest struct {
	FirstName string `json:"first_name" validate:"required,min=2,max=100"`
	LastName  string `json:"last_name" validate:"max=100"`

	Email string `json:"email" validate:"required,email"`

	Password string `json:"password" validate:"required,min=8"`
}

type UpdateUserRequest struct {
	FirstName string `json:"first_name"`

	LastName string `json:"last_name"`

	IsActive *bool `json:"is_active"`
}

// UpdateUserRoleRequest changes a user's role (admin only).
type UpdateUserRoleRequest struct {
	Role models.Role `json:"role" binding:"required"`
}

type UserResponse struct {
	ID string `json:"id"`

	FirstName string `json:"first_name"`

	LastName string `json:"last_name"`

	Email string `json:"email"`

	IsActive bool `json:"is_active"`

	Role models.Role `json:"role"`

	CreatedAt string `json:"created_at"`

	UpdatedAt string `json:"updated_at"`
}
