package ports

import (
	"context"
	"time"
)

// UserView is the read-only projection of a user for cross-BC queries.
type UserView struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
	Active   bool   `json:"active"`
}

// UserReader provides read-only access to the User BC from other BCs.
type UserReader interface {
	GetByID(ctx context.Context, id string) (*UserView, error)
	GetByEmail(ctx context.Context, email string) (*UserView, error)
	ExistsByID(ctx context.Context, id string) (bool, error)
}

// UserAuthView contains authentication-related fields for cross-BC auth flows.
type UserAuthView struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	Password    string `json:"-"` // hashed, never serialized
	Role        string `json:"role"`
	TOTPSecret  string `json:"-"`
	TOTPEnabled bool   `json:"totp_enabled"`
}

// UserAuthReader provides auth-related access to the User BC.
// Used by Session and Challenge BCs for authentication flows.
type UserAuthReader interface {
	GetByID(ctx context.Context, id string) (*UserAuthView, error)
	GetByEmail(ctx context.Context, email string) (*UserAuthView, error)
	UpdateTOTP(ctx context.Context, userID, totpSecret string, enabled bool, verifiedAt *time.Time) error
}
