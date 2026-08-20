package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/api/response"
)

type Handler struct {
	service *Service
}

func NewHandler(
	service *Service,
) *Handler {

	return &Handler{
		service: service,
	}
}

// Login
//
// @Summary Login
// @Description Authenticate user and generate JWT token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login Request"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Router /auth/login [post]
func (h *Handler) Login(c *gin.Context) {

	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	loginResponse, err := h.service.Login(
		c.Request.Context(),
		req,
	)

	if err != nil {

		switch {

		case errors.Is(err, ErrInvalidCredentials),
			errors.Is(err, ErrInactiveUser):

			response.Error(
				c,
				http.StatusUnauthorized,
				err.Error(),
			)
			return

		default:

			response.Error(
				c,
				http.StatusInternalServerError,
				err.Error(),
			)
			return
		}
	}

	response.OK(
		c,
		"login successful",
		loginResponse,
	)
}

// Me
//
// @Summary Current User
// @Description Get currently authenticated user
// @Tags Authentication
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /auth/me [get]
func (h *Handler) Me(c *gin.Context) {

	userID := c.GetString("user_id")

	if userID == "" {

		response.Error(
			c,
			http.StatusUnauthorized,
			ErrInvalidToken.Error(),
		)

		return
	}

	user, err := h.service.GetCurrentUser(
		c.Request.Context(),
		userID,
	)

	if err != nil {

		response.Error(
			c,
			http.StatusInternalServerError,
			err.Error(),
		)

		return
	}

	response.OK(
		c,
		"user retrieved successfully",
		user,
	)
}
