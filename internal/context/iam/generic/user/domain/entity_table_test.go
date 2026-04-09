package domain

import "testing"

func TestNewUser_TableDriven(t *testing.T) {
	tests := []struct {
		name           string
		email          string
		hashedPassword string
		firstName      string
		lastName       string
		wantErr        error
	}{
		{
			name:           "valid user",
			email:          "test@example.com",
			hashedPassword: "hashed_password",
			firstName:      "Ali",
			lastName:       "Valiyev",
			wantErr:        nil,
		},
		{
			name:           "valid user with empty last name",
			email:          "ali@example.com",
			hashedPassword: "hashed_password",
			firstName:      "Ali",
			lastName:       "",
			wantErr:        nil,
		},
		{
			name:           "empty email",
			email:          "",
			hashedPassword: "hashed_password",
			firstName:      "Ali",
			lastName:       "Valiyev",
			wantErr:        ErrInvalidEmail,
		},
		{
			name:           "empty password",
			email:          "test@example.com",
			hashedPassword: "",
			firstName:      "Ali",
			lastName:       "Valiyev",
			wantErr:        ErrInvalidPassword,
		},
		{
			name:           "empty first name",
			email:          "test@example.com",
			hashedPassword: "hashed_password",
			firstName:      "",
			lastName:       "Valiyev",
			wantErr:        ErrInvalidName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := NewUser(tt.email, tt.hashedPassword, tt.firstName, tt.lastName)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if u.Email != tt.email {
				t.Errorf("expected email %s, got %s", tt.email, u.Email)
			}
			if u.Password != tt.hashedPassword {
				t.Errorf("expected password %s, got %s", tt.hashedPassword, u.Password)
			}
			if u.FirstName != tt.firstName {
				t.Errorf("expected firstName %s, got %s", tt.firstName, u.FirstName)
			}
			if u.LastName != tt.lastName {
				t.Errorf("expected lastName %s, got %s", tt.lastName, u.LastName)
			}
		})
	}
}

func TestNewUser_DefaultRole(t *testing.T) {
	u, err := NewUser("test@example.com", "hashed_password", "Ali", "Valiyev")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.Role != RoleCustomer {
		t.Errorf("default role should be CUSTOMER, got %s", u.Role)
	}
}

func TestNewUser_TimestampsAreSet(t *testing.T) {
	u, err := NewUser("test@example.com", "hashed_password", "Ali", "Valiyev")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
	if u.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero")
	}
	if !u.CreatedAt.Equal(u.UpdatedAt) {
		t.Error("CreatedAt and UpdatedAt should be equal for a new user")
	}
}

func TestNewUser_TOTPDefaultState(t *testing.T) {
	u, err := NewUser("test@example.com", "hashed_password", "Ali", "Valiyev")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.TOTPEnabled {
		t.Error("TOTP should be disabled by default")
	}
	if u.TOTPSecret != "" {
		t.Error("TOTP secret should be empty by default")
	}
	if u.TOTPVerifiedAt != nil {
		t.Error("TOTP verified_at should be nil by default")
	}
}

func TestRole_Constants(t *testing.T) {
	tests := []struct {
		role Role
		want string
	}{
		{RoleCustomer, "CUSTOMER"},
		{RoleTeller, "TELLER"},
		{RoleAdmin, "ADMIN"},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			if string(tt.role) != tt.want {
				t.Errorf("expected %s, got %s", tt.want, tt.role)
			}
		})
	}
}

func TestNewUser_IDNotSetByFactory(t *testing.T) {
	u, err := NewUser("test@example.com", "hashed_password", "Ali", "Valiyev")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// ID is set by the repository/DB, not the factory
	if u.ID != "" {
		t.Error("ID should be empty (set by repository)")
	}
}
