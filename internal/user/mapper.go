package user

import (
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
)

func ToResponse(user *models.User) UserResponse {

	return UserResponse{

		ID: user.ID.String(),

		FirstName: user.FirstName,

		LastName: user.LastName,

		Email: user.Email,

		IsActive: user.IsActive,

		CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),

		UpdatedAt: user.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func ToResponseList(users []models.User) []UserResponse {

	resp := make([]UserResponse, 0, len(users))

	for _, user := range users {

		resp = append(resp, ToResponse(&user))
	}

	return resp
}