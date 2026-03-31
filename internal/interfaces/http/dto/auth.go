package dto

// LoginRequest - incoming data for login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RefreshRequest - for token refresh
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// LogoutRequest - for logout
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// AuthResponse - login/refresh response
type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         UserResponse `json:"user"`
}
