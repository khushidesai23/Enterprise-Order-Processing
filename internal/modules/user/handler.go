package user

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/api/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// CreateUser
//
// @Summary Create User
// @Description Register a new user
// @Tags Users
// @Accept json
// @Produce json
// @Param request body CreateUserRequest true "User"
// @Success 201 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 409 {object} response.APIResponse
// @Router /users [post]
func (h *Handler) Create(c *gin.Context) {

	var req CreateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.service.Create(c.Request.Context(), req)

	if err != nil {

		switch {

		case errors.Is(err, ErrUserAlreadyExists):
			response.Error(c, http.StatusConflict, err.Error())

		default:
			response.Error(c, http.StatusBadRequest, err.Error())
		}

		return
	}

	response.Created(c, "user created successfully", user)
}

// GetUserByID
//
// @Summary Get User by ID
// @Description Get user details by ID
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /users/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {

	id, err := uuid.Parse(c.Param("id"))

	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid user id")
		return
	}

	user, err := h.service.GetByID(c.Request.Context(), id)

	if err != nil {

		if errors.Is(err, ErrUserNotFound) {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}

		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.OK(c, "user fetched successfully", user)
}

// ListUsers
//
// @Summary List Users
// @Description Get a list of all users
// @Tags Users
// @Accept json
// @Produce json
// @Success 200 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /users [get]
func (h *Handler) List(c *gin.Context) {

	users, err := h.service.List(c.Request.Context())

	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.OK(c, "users fetched successfully", users)
}

// UpdateUser
//
// @Summary Update User
// @Description Update user details by ID
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body UpdateUserRequest true "User"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /users/{id} [put]
func (h *Handler) Update(c *gin.Context) {

	id, err := uuid.Parse(c.Param("id"))

	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid user id")
		return
	}

	var req UpdateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.service.Update(
		c.Request.Context(),
		id,
		req,
	)

	if err != nil {

		if errors.Is(err, ErrUserNotFound) {

			response.Error(c, http.StatusNotFound, err.Error())
			return
		}

		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.OK(c, "user updated successfully", user)
}

// DeleteUser
//
// @Summary Delete User
// @Description Delete user by ID
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /users/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {

	id, err := uuid.Parse(c.Param("id"))

	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid user id")
		return
	}

	err = h.service.Delete(
		c.Request.Context(),
		id,
	)

	if err != nil {

		if errors.Is(err, ErrUserNotFound) {

			response.Error(c, http.StatusNotFound, err.Error())
			return
		}

		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.OK(c, "user deleted successfully", nil)
}