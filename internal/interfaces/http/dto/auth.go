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
	AccessToken  string       `json:"access_token,omitempty"`
	RefreshToken string       `json:"refresh_token,omitempty"`
	User         UserResponse `json:"user"`
	TOTPRequired bool         `json:"totp_required,omitempty"`
}

// --- TOTP DTOs ---

type TOTPVerifyLoginRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"` // 6-digit TOTP code
}

type TOTPSetupResponse struct {
	Secret string `json:"secret"` // base32 secret for manual entry
	URL    string `json:"url"`    // otpauth:// URL for QR code
}

type TOTPVerifySetupRequest struct {
	Code string `json:"code"` // 6-digit code to confirm setup
}

type TOTPDisableRequest struct {
	Password string `json:"password"` // confirm with password
}
