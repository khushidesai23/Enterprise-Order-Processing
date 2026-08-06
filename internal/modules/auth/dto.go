package auth

// Login Request
type LoginRequest struct {
	Email string `json:"email" binding:"required,email"`

	Password string `json:"password" binding:"required"`
}

// Login Response
type LoginResponse struct {
	Token string `json:"token"`

	TokenType string `json:"token_type"`

	ExpiresIn string `json:"expires_in"`

	User UserInfo `json:"user"`
}

// Logged-in User Information
type UserInfo struct {
	ID string `json:"id"`

	FirstName string `json:"first_name"`

	LastName string `json:"last_name"`

	Email string `json:"email"`

	IsActive bool `json:"is_active"`
}