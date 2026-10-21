package models

// Role is a user's authorization role.
type Role string

const (
	// RoleCustomer can place orders and manage only its own orders,
	// payments and profile. Every self-registered user is a customer.
	RoleCustomer Role = "customer"

	// RoleAdmin can manage the catalog, inventory, users and every order.
	RoleAdmin Role = "admin"
)

func (r Role) Valid() bool {
	return r == RoleCustomer || r == RoleAdmin
}
