package dto

// RegisterRequest is the input body for POST /api/v1/auth/register.
type RegisterRequest struct {
	Name     string `json:"name"     binding:"required,min=2,max=100"`
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// LoginRequest is the input body for POST /api/v1/auth/login.
type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// UserResponse is the safe public representation of a user — no password hash.
type UserResponse struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Team     string `json:"team,omitempty"`
	IsActive bool   `json:"is_active"`
}

// AgentJoinRequest is the input body for POST /api/v1/auth/agent-join.
// Agents use this to register under an existing tenant by company slug.
type AgentJoinRequest struct {
	Name        string `json:"name"         binding:"required,min=2,max=100"`
	Email       string `json:"email"        binding:"required,email"`
	Password    string `json:"password"     binding:"required,min=8"`
	CompanySlug string `json:"company_slug" binding:"required"`
	Team        string `json:"team"         binding:"required"`
}

// AuthResponse is returned on successful login and registration.
type AuthResponse struct {
	AccessToken  string       `json:"accessToken"`
	RefreshToken string       `json:"refreshToken"`
	User         UserResponse `json:"user"`
}
