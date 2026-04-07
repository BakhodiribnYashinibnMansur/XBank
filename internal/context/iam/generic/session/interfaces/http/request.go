package http

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type AuthResponse struct {
	AccessToken  string       `json:"access_token,omitempty"`
	RefreshToken string       `json:"refresh_token,omitempty"`
	User         UserResponse `json:"user"`
	TOTPRequired bool         `json:"totp_required,omitempty"`
}

type UserResponse struct {
	ID        string      `json:"id"`
	Email     string      `json:"email"`
	FirstName string      `json:"first_name,omitempty"`
	LastName  string      `json:"last_name,omitempty"`
	CreatedAt interface{} `json:"created_at,omitempty"`
}

type TOTPVerifyLoginRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type TOTPSetupResponse struct {
	Secret string `json:"secret"`
	URL    string `json:"url"`
}

type TOTPVerifySetupRequest struct {
	Code string `json:"code"`
}

type TOTPDisableRequest struct {
	Password string `json:"password"`
}
