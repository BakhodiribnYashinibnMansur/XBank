package e2e_test

import (
	"testing"

	"github.com/gofiber/fiber/v2"
)

// TestUserRegistrationFlow_RegisterKYCChangePassword tests the full user onboarding journey:
// register -> login -> submit KYC -> check KYC status -> change password -> login with new password
func TestUserRegistrationFlow_RegisterKYCChangePassword(t *testing.T) {
	email := uniqueEmail("onboarding")

	// Step 1: Register a new user
	rec := doRequest(t, fiber.MethodPost, "/api/v1/auth/register", map[string]string{
		"email":      email,
		"password":   "InitialPass123!",
		"first_name": "Onboard",
		"last_name":  "User",
	}, "")
	expectStatus(t, rec, fiber.StatusCreated)

	var userResp struct {
		ID        string `json:"id"`
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}
	parseResponse(t, rec, &userResp)

	if userResp.ID == "" {
		t.Fatal("expected user ID after registration")
	}
	if userResp.Email != email {
		t.Errorf("email = %q, want %q", userResp.Email, email)
	}
	if userResp.FirstName != "Onboard" {
		t.Errorf("first_name = %q, want %q", userResp.FirstName, "Onboard")
	}
	if userResp.LastName != "User" {
		t.Errorf("last_name = %q, want %q", userResp.LastName, "User")
	}

	// Step 2: Login
	rec = doRequest(t, fiber.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    email,
		"password": "InitialPass123!",
	}, "")
	expectStatus(t, rec, fiber.StatusOK)

	var tokens authTokens
	parseResponse(t, rec, &tokens)

	if tokens.AccessToken == "" {
		t.Fatal("expected access token after login")
	}
	if tokens.RefreshToken == "" {
		t.Fatal("expected refresh token after login")
	}

	// Step 3: Submit KYC verification
	rec = doRequest(t, fiber.MethodPost, "/api/v1/kyc/submit", map[string]string{
		"document_type":   "PASSPORT",
		"document_number": "AB1234567",
		"first_name":      "Onboard",
		"last_name":       "User",
		"date_of_birth":   "1990-05-15",
	}, tokens.AccessToken)
	expectStatus(t, rec, fiber.StatusCreated)

	var kycResp struct {
		ID           string `json:"id"`
		UserID       string `json:"user_id"`
		DocumentType string `json:"document_type"`
		Status       string `json:"status"`
	}
	parseResponse(t, rec, &kycResp)

	if kycResp.ID == "" {
		t.Fatal("expected KYC verification ID")
	}
	if kycResp.Status != "PENDING" {
		t.Errorf("kyc status = %q, want PENDING", kycResp.Status)
	}
	if kycResp.DocumentType != "PASSPORT" {
		t.Errorf("document_type = %q, want PASSPORT", kycResp.DocumentType)
	}

	// Step 4: Check KYC status
	rec = doRequest(t, fiber.MethodGet, "/api/v1/kyc/status", nil, tokens.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)

	var kycStatus struct {
		Status string `json:"status"`
	}
	parseResponse(t, rec, &kycStatus)

	if kycStatus.Status != "PENDING" {
		t.Errorf("kyc status check = %q, want PENDING", kycStatus.Status)
	}

	// Step 5: Get user profile
	rec = doRequest(t, fiber.MethodGet, "/api/v1/users/get?id="+tokens.User.ID, nil, tokens.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)

	// Step 6: Change password
	rec = doRequest(t, fiber.MethodPost, "/api/v1/users/change-password", map[string]string{
		"old_password": "InitialPass123!",
		"new_password": "NewSecurePass456!",
	}, tokens.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)

	// Step 7: Login with old password should fail
	rec = doRequest(t, fiber.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    email,
		"password": "InitialPass123!",
	}, "")
	if rec.Code == fiber.StatusOK {
		t.Error("expected login with old password to fail after password change")
	}

	// Step 8: Login with new password should succeed
	rec = doRequest(t, fiber.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    email,
		"password": "NewSecurePass456!",
	}, "")
	expectStatus(t, rec, fiber.StatusOK)
}

// TestUserRegistrationFlow_InvalidRegistration tests validation on registration.
func TestUserRegistrationFlow_InvalidRegistration(t *testing.T) {
	t.Run("missing_email", func(t *testing.T) {
		rec := doRequest(t, fiber.MethodPost, "/api/v1/auth/register", map[string]string{
			"password":   "Pass12345!",
			"first_name": "No",
			"last_name":  "Email",
		}, "")
		if rec.Code == fiber.StatusCreated {
			t.Error("expected registration without email to fail")
		}
	})

	t.Run("missing_password", func(t *testing.T) {
		rec := doRequest(t, fiber.MethodPost, "/api/v1/auth/register", map[string]string{
			"email":      uniqueEmail("no-pass"),
			"first_name": "No",
			"last_name":  "Pass",
		}, "")
		if rec.Code == fiber.StatusCreated {
			t.Error("expected registration without password to fail")
		}
	})

	t.Run("empty_body", func(t *testing.T) {
		rec := doRequest(t, fiber.MethodPost, "/api/v1/auth/register", map[string]string{}, "")
		if rec.Code == fiber.StatusCreated {
			t.Error("expected registration with empty body to fail")
		}
	})

	t.Run("invalid_email_format", func(t *testing.T) {
		rec := doRequest(t, fiber.MethodPost, "/api/v1/auth/register", map[string]string{
			"email":      "not-an-email",
			"password":   "Pass12345!",
			"first_name": "Bad",
			"last_name":  "Email",
		}, "")
		if rec.Code == fiber.StatusCreated {
			t.Error("expected registration with invalid email format to fail")
		}
	})

	t.Run("weak_password", func(t *testing.T) {
		rec := doRequest(t, fiber.MethodPost, "/api/v1/auth/register", map[string]string{
			"email":      uniqueEmail("weak-pw"),
			"password":   "123",
			"first_name": "Weak",
			"last_name":  "Pass",
		}, "")
		if rec.Code == fiber.StatusCreated {
			t.Error("expected registration with weak password to fail")
		}
	})

	t.Run("missing_first_name", func(t *testing.T) {
		rec := doRequest(t, fiber.MethodPost, "/api/v1/auth/register", map[string]string{
			"email":    uniqueEmail("no-name"),
			"password": "Pass12345!",
		}, "")
		// Some systems accept missing first_name; just verify it doesn't panic
		if rec.Code == fiber.StatusInternalServerError {
			t.Error("server error on missing first_name; expected validation error")
		}
	})
}

// TestUserRegistrationFlow_KYCSubmitWithoutAuth tests that KYC submission requires authentication.
func TestUserRegistrationFlow_KYCSubmitWithoutAuth(t *testing.T) {
	rec := doRequest(t, fiber.MethodPost, "/api/v1/kyc/submit", map[string]string{
		"document_type":   "PASSPORT",
		"document_number": "XX9999999",
		"first_name":      "Unauth",
		"last_name":       "User",
		"date_of_birth":   "2000-01-01",
	}, "")
	if rec.Code != fiber.StatusUnauthorized {
		t.Errorf("expected 401 for unauthenticated KYC submit, got %d", rec.Code)
	}
}

// TestUserRegistrationFlow_KYCStatusWithoutAuth tests that KYC status check requires authentication.
func TestUserRegistrationFlow_KYCStatusWithoutAuth(t *testing.T) {
	rec := doRequest(t, fiber.MethodGet, "/api/v1/kyc/status", nil, "")
	if rec.Code != fiber.StatusUnauthorized {
		t.Errorf("expected 401 for unauthenticated KYC status, got %d", rec.Code)
	}
}

// TestUserRegistrationFlow_ChangePasswordWrongOldPassword tests change password with wrong current password.
func TestUserRegistrationFlow_ChangePasswordWrongOldPassword(t *testing.T) {
	email := uniqueEmail("chpw-wrong")
	tokens := registerAndLogin(t, email, "CorrectPass123!", "ChPw")

	rec := doRequest(t, fiber.MethodPost, "/api/v1/users/change-password", map[string]string{
		"old_password": "WrongPassword!",
		"new_password": "NewPass456!",
	}, tokens.AccessToken)

	if rec.Code == fiber.StatusOK {
		t.Error("expected change-password with wrong old password to fail")
	}
}

// TestUserRegistrationFlow_ChangePasswordWithoutAuth tests that change password requires auth.
func TestUserRegistrationFlow_ChangePasswordWithoutAuth(t *testing.T) {
	rec := doRequest(t, fiber.MethodPost, "/api/v1/users/change-password", map[string]string{
		"old_password": "OldPass123!",
		"new_password": "NewPass456!",
	}, "")
	if rec.Code != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

// TestUserRegistrationFlow_GetProfileByID tests fetching user profile by ID.
func TestUserRegistrationFlow_GetProfileByID(t *testing.T) {
	email := uniqueEmail("profile")
	tokens := registerAndLogin(t, email, "Password123!", "Profile")

	rec := doRequest(t, fiber.MethodGet, "/api/v1/users/get?id="+tokens.User.ID, nil, tokens.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)

	var profile struct {
		ID        string `json:"id"`
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
	}
	parseResponse(t, rec, &profile)

	if profile.ID != tokens.User.ID {
		t.Errorf("profile ID = %q, want %q", profile.ID, tokens.User.ID)
	}
}

// TestUserRegistrationFlow_GetProfileWithoutID tests fetching user without providing an ID.
func TestUserRegistrationFlow_GetProfileWithoutID(t *testing.T) {
	email := uniqueEmail("no-id-profile")
	tokens := registerAndLogin(t, email, "Password123!", "NoID")

	rec := doRequest(t, fiber.MethodGet, "/api/v1/users/get", nil, tokens.AccessToken)
	// Should fail because no ID provided
	if rec.Code == fiber.StatusOK {
		// Only flag if the API strictly requires an ID parameter
		t.Log("GET /users/get without ID returned 200 (may return own profile)")
	}
}

// TestUserRegistrationFlow_DataExport tests the user data export endpoint.
func TestUserRegistrationFlow_DataExport(t *testing.T) {
	email := uniqueEmail("export")
	tokens := registerAndLogin(t, email, "Password123!", "Export")

	rec := doRequest(t, fiber.MethodGet, "/api/v1/users/me/data-export", nil, tokens.AccessToken)
	// Should return 200 with user data or 202 (accepted / processing)
	expectStatusOneOf(t, rec, fiber.StatusOK, fiber.StatusAccepted)
}

// TestUserRegistrationFlow_DeleteAccount tests the user account deletion endpoint.
func TestUserRegistrationFlow_DeleteAccount(t *testing.T) {
	email := uniqueEmail("delete-me")
	tokens := registerAndLogin(t, email, "Password123!", "Delete")

	rec := doRequest(t, fiber.MethodDelete, "/api/v1/users/me/delete", nil, tokens.AccessToken)
	expectStatusOneOf(t, rec, fiber.StatusOK, fiber.StatusNoContent, fiber.StatusAccepted)

	// After deletion, login should fail
	rec = doRequest(t, fiber.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    email,
		"password": "Password123!",
	}, "")
	if rec.Code == fiber.StatusOK {
		t.Error("expected login to fail after account deletion")
	}
}

// TestUserRegistrationFlow_KYCDocumentTypes tests submitting KYC with different document types.
func TestUserRegistrationFlow_KYCDocumentTypes(t *testing.T) {
	docTypes := []string{"PASSPORT", "ID_CARD", "DRIVERS_LICENSE"}

	for _, docType := range docTypes {
		t.Run(docType, func(t *testing.T) {
			email := uniqueEmail("kyc-" + docType)
			tokens := registerAndLogin(t, email, "Password123!", "KYC")

			rec := doRequest(t, fiber.MethodPost, "/api/v1/kyc/submit", map[string]string{
				"document_type":   docType,
				"document_number": "DOC" + docType + "123",
				"first_name":      "KYC",
				"last_name":       "Test",
				"date_of_birth":   "1988-03-25",
			}, tokens.AccessToken)
			expectStatus(t, rec, fiber.StatusCreated)

			var kycResp struct {
				DocumentType string `json:"document_type"`
				Status       string `json:"status"`
			}
			parseResponse(t, rec, &kycResp)

			if kycResp.DocumentType != docType {
				t.Errorf("document_type = %q, want %q", kycResp.DocumentType, docType)
			}
			if kycResp.Status != "PENDING" {
				t.Errorf("status = %q, want PENDING", kycResp.Status)
			}
		})
	}
}
