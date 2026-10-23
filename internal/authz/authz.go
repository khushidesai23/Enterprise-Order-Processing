// Package authz carries the authenticated caller through a request and
// answers the two authorization questions handlers and services ask:
// "does the caller have this role?" and "may the caller touch this
// user's resource?".
package authz

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/api/response"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
)

var ErrForbidden = errors.New("insufficient permissions")

const principalKey = "principal"

// Principal is the authenticated caller of a request.
type Principal struct {
	UserID uuid.UUID
	Role   models.Role
}

func (p Principal) IsAdmin() bool {
	return p.Role == models.RoleAdmin
}

// CanAccess reports whether the caller may read or modify a resource
// owned by ownerID: admins may access everything, others only their own.
func (p Principal) CanAccess(ownerID uuid.UUID) bool {
	return p.IsAdmin() || (p.UserID != uuid.Nil && p.UserID == ownerID)
}

func SetPrincipal(c *gin.Context, principal Principal) {
	c.Set(principalKey, principal)
}

// FromContext returns the principal stored by the auth middleware.
func FromContext(c *gin.Context) (Principal, bool) {
	value, ok := c.Get(principalKey)
	if !ok {
		return Principal{}, false
	}

	principal, ok := value.(Principal)

	return principal, ok
}

// RequirePrincipal returns the principal, or writes a 401 and returns false
// when the request was not authenticated.
func RequirePrincipal(c *gin.Context) (Principal, bool) {
	principal, ok := FromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required")
		return Principal{}, false
	}

	return principal, true
}
