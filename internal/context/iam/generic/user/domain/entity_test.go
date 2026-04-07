package domain

import "testing"

func TestNewUser_Success(t *testing.T) {
	u, err := NewUser("test@example.com", "hashed_password", "Ali", "Valiyev")
	if err != nil {
		t.Fatalf("Kutilmagan xatolik: %v", err)
	}

	if u.Email != "test@example.com" {
		t.Errorf("Email kutilgan: test@example.com, kelgan: %s", u.Email)
	}
	if u.FirstName != "Ali" {
		t.Errorf("FirstName kutilgan: Ali, kelgan: %s", u.FirstName)
	}
	if u.CreatedAt.IsZero() {
		t.Error("CreatedAt bo'sh bo'lmasligi kerak")
	}
}

func TestNewUser_EmptyEmail(t *testing.T) {
	_, err := NewUser("", "hashed_password", "Ali", "Valiyev")
	if err != ErrInvalidEmail {
		t.Errorf("Kutilgan xatolik: %v, kelgan: %v", ErrInvalidEmail, err)
	}
}

func TestNewUser_EmptyPassword(t *testing.T) {
	_, err := NewUser("test@example.com", "", "Ali", "Valiyev")
	if err != ErrInvalidPassword {
		t.Errorf("Kutilgan xatolik: %v, kelgan: %v", ErrInvalidPassword, err)
	}
}

func TestNewUser_EmptyName(t *testing.T) {
	_, err := NewUser("test@example.com", "hashed_password", "", "Valiyev")
	if err != ErrInvalidName {
		t.Errorf("Kutilgan xatolik: %v, kelgan: %v", ErrInvalidName, err)
	}
}
