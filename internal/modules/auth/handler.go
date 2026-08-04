package auth

import (
	"errors"
	"net/http"
	"strings"

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

	loginResp, err := h.service.Login(
		c.Request.Context(),
		req,
	)

	if err != nil {

		switch {

		case errors.Is(err, ErrInvalidCredentials),
			errors.Is(err, ErrInactiveUser):

			response.Error(c, http.StatusUnauthorized, err.Error())
			return

		default:

			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}
	}

	response.OK(
		c,
		"login successful",
		loginResp,
	)
}

// Current User
//
// @Summary Current User
// @Description Returns the currently authenticated user
// @Tags Authentication
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Router /auth/me [get]
func (h *Handler) Me(c *gin.Context) {

	authHeader := c.GetHeader("Authorization")

	if authHeader == "" {

		response.Error(
			c,
			http.StatusUnauthorized,
			ErrMissingToken.Error(),
		)

		return
	}

	token := strings.TrimPrefix(
		authHeader,
		"Bearer ",
	)

	claims, err := h.service.VerifyToken(token)

	if err != nil {

		response.Error(
			c,
			http.StatusUnauthorized,
			err.Error(),
		)

		return
	}

	user, err := h.service.GetCurrentUser(
		c.Request.Context(),
		claims.UserID,
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