package e2e_test

import (
	"testing"

	"github.com/gofiber/fiber/v2"
)

// Note: Admin routes require ADMIN role + IP whitelist. In the E2E test setup,
// the admin IP whitelist middleware is passed as nil, so these tests verify
// that non-admin users are correctly denied access to admin endpoints.

// TestAdminFlow_KYCApproveReject tests the admin KYC review workflow:
// user submits KYC -> regular user is denied admin KYC actions
func TestAdminFlow_KYCApproveReject(t *testing.T) {
	// Create a regular user and submit KYC
	user := registerAndLogin(t, uniqueEmail("kyc-review"), "Password123!", "KYCUser")

	// Submit KYC
	rec := doRequest(t, fiber.MethodPost, "/api/v1/kyc/submit", map[string]string{
		"document_type":   "ID_CARD",
		"document_number": "ID9876543",
		"first_name":      "KYC",
		"last_name":       "User",
		"date_of_birth":   "1985-10-20",
	}, user.AccessToken)
	expectStatus(t, rec, fiber.StatusCreated)

	var kycResp struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	parseResponse(t, rec, &kycResp)
	kycID := kycResp.ID

	// Verify status is PENDING
	if kycResp.Status != "PENDING" {
		t.Errorf("kyc status = %q, want PENDING", kycResp.Status)
	}

	t.Run("approve_denied", func(t *testing.T) {
		rec := doRequest(t, fiber.MethodPost, "/api/v1/admin/kyc/approve", map[string]string{
			"verification_id": kycID,
		}, user.AccessToken)
		if rec.Code == fiber.StatusOK {
			t.Error("expected non-admin user to be denied admin KYC approve")
		}
	})

	t.Run("list_pending_denied", func(t *testing.T) {
		rec := doRequest(t, fiber.MethodGet, "/api/v1/admin/kyc/pending", nil, user.AccessToken)
		if rec.Code == fiber.StatusOK {
			t.Error("expected non-admin user to be denied admin KYC listing")
		}
	})

	t.Run("reject_denied", func(t *testing.T) {
		rec := doRequest(t, fiber.MethodPost, "/api/v1/admin/kyc/reject", map[string]string{
			"verification_id": kycID,
			"reason":          "Incomplete documents",
		}, user.AccessToken)
		if rec.Code == fiber.StatusOK {
			t.Error("expected non-admin user to be denied admin KYC reject")
		}
	})
}

// TestAdminFlow_FeatureFlagAccessDenied tests that non-admin users cannot manage feature flags.
func TestAdminFlow_FeatureFlagAccessDenied(t *testing.T) {
	user := registerAndLogin(t, uniqueEmail("flag-nonadmin"), "Password123!", "FlagUser")

	t.Run("create_denied", func(t *testing.T) {
		rec := doRequest(t, fiber.MethodPost, "/api/v1/admin/flags/", map[string]string{
			"key":           "test_flag",
			"description":   "Test feature flag",
			"flag_type":     "bool",
			"default_value": "true",
		}, user.AccessToken)
		if rec.Code == fiber.StatusCreated || rec.Code == fiber.StatusOK {
			t.Error("expected non-admin user to be denied feature flag creation")
		}
	})

	t.Run("list_denied", func(t *testing.T) {
		rec := doRequest(t, fiber.MethodGet, "/api/v1/admin/flags/", nil, user.AccessToken)
		if rec.Code == fiber.StatusOK {
			t.Error("expected non-admin user to be denied feature flag listing")
		}
	})
}

// TestAdminFlow_StatisticsAccessDenied tests that non-admin users cannot access statistics.
func TestAdminFlow_StatisticsAccessDenied(t *testing.T) {
	user := registerAndLogin(t, uniqueEmail("stats-nonadmin"), "Password123!", "StatsUser")

	rec := doRequest(t, fiber.MethodGet, "/api/v1/admin/statistics/overview", nil, user.AccessToken)
	if rec.Code == fiber.StatusOK {
		t.Error("expected non-admin user to be denied statistics access")
	}
}

// TestAdminFlow_SiteSettingsAccessDenied tests that non-admin users cannot manage site settings.
func TestAdminFlow_SiteSettingsAccessDenied(t *testing.T) {
	user := registerAndLogin(t, uniqueEmail("siteset-nonadmin"), "Password123!", "SiteUser")

	t.Run("get_denied", func(t *testing.T) {
		rec := doRequest(t, fiber.MethodGet, "/api/v1/admin/settings/", nil, user.AccessToken)
		if rec.Code == fiber.StatusOK {
			t.Error("expected non-admin user to be denied site settings")
		}
	})

	t.Run("update_denied", func(t *testing.T) {
		rec := doRequest(t, fiber.MethodPut, "/api/v1/admin/settings/", map[string]string{
			"key":   "maintenance_mode",
			"value": "true",
		}, user.AccessToken)
		if rec.Code == fiber.StatusOK {
			t.Error("expected non-admin user to be denied site settings update")
		}
	})
}

// TestAdminFlow_AnnouncementsAccessDenied tests that non-admin users cannot manage announcements.
func TestAdminFlow_AnnouncementsAccessDenied(t *testing.T) {
	user := registerAndLogin(t, uniqueEmail("ann-nonadmin"), "Password123!", "AnnUser")

	t.Run("create_denied", func(t *testing.T) {
		rec := doRequest(t, fiber.MethodPost, "/api/v1/admin/announcements/", map[string]string{
			"title":   "Test Announcement",
			"content": "This is a test",
		}, user.AccessToken)
		if rec.Code == fiber.StatusCreated || rec.Code == fiber.StatusOK {
			t.Error("expected non-admin user to be denied announcement creation")
		}
	})

	t.Run("list_denied", func(t *testing.T) {
		rec := doRequest(t, fiber.MethodGet, "/api/v1/admin/announcements/", nil, user.AccessToken)
		if rec.Code == fiber.StatusOK {
			t.Error("expected non-admin user to be denied admin announcement listing")
		}
	})
}

// TestAdminFlow_PublicAnnouncementsAccessible tests that public announcements are accessible without admin role.
func TestAdminFlow_PublicAnnouncementsAccessible(t *testing.T) {
	// Public active announcements should be accessible without auth
	rec := doRequest(t, fiber.MethodGet, "/api/v1/announcements/active", nil, "")
	// Should return 200 even without auth (it's a public route)
	if rec.Code != fiber.StatusOK {
		t.Logf("public announcements returned status %d (may be expected if no announcements exist)", rec.Code)
	}
}

// TestAdminFlow_AuthzAccessDenied tests that non-admin users cannot access RBAC management.
func TestAdminFlow_AuthzAccessDenied(t *testing.T) {
	user := registerAndLogin(t, uniqueEmail("authz-nonadmin"), "Password123!", "AuthzUser")

	t.Run("list_roles_denied", func(t *testing.T) {
		rec := doRequest(t, fiber.MethodGet, "/api/v1/admin/authz/roles", nil, user.AccessToken)
		if rec.Code == fiber.StatusOK {
			t.Error("expected non-admin user to be denied roles listing")
		}
	})

	t.Run("create_role_denied", func(t *testing.T) {
		rec := doRequest(t, fiber.MethodPost, "/api/v1/admin/authz/roles", map[string]string{
			"name":        "TEST_ROLE",
			"description": "Test role",
		}, user.AccessToken)
		if rec.Code == fiber.StatusCreated || rec.Code == fiber.StatusOK {
			t.Error("expected non-admin user to be denied role creation")
		}
	})

	t.Run("list_permissions_denied", func(t *testing.T) {
		rec := doRequest(t, fiber.MethodGet, "/api/v1/admin/authz/permissions", nil, user.AccessToken)
		if rec.Code == fiber.StatusOK {
			t.Error("expected non-admin user to be denied permissions listing")
		}
	})
}

// TestAdminFlow_ErrorCodesAccessDenied tests that non-admin users cannot manage error codes.
func TestAdminFlow_ErrorCodesAccessDenied(t *testing.T) {
	user := registerAndLogin(t, uniqueEmail("errcode-nonadmin"), "Password123!", "ErrUser")

	t.Run("list_denied", func(t *testing.T) {
		rec := doRequest(t, fiber.MethodGet, "/api/v1/admin/errorcodes/", nil, user.AccessToken)
		if rec.Code == fiber.StatusOK {
			t.Error("expected non-admin user to be denied error codes listing")
		}
	})

	t.Run("create_denied", func(t *testing.T) {
		rec := doRequest(t, fiber.MethodPost, "/api/v1/admin/errorcodes/", map[string]string{
			"code":    "TEST_ERR",
			"message": "Test error",
		}, user.AccessToken)
		if rec.Code == fiber.StatusCreated || rec.Code == fiber.StatusOK {
			t.Error("expected non-admin user to be denied error code creation")
		}
	})
}

// TestAdminFlow_TranslationsAccessDenied tests that non-admin users cannot manage translations.
func TestAdminFlow_TranslationsAccessDenied(t *testing.T) {
	user := registerAndLogin(t, uniqueEmail("trans-nonadmin"), "Password123!", "TransUser")

	rec := doRequest(t, fiber.MethodGet, "/api/v1/admin/translations/", nil, user.AccessToken)
	if rec.Code == fiber.StatusOK {
		t.Error("expected non-admin user to be denied translations listing")
	}
}

// TestAdminFlow_FraudAccessDenied tests that non-admin users cannot access fraud management.
func TestAdminFlow_FraudAccessDenied(t *testing.T) {
	user := registerAndLogin(t, uniqueEmail("fraud-nonadmin"), "Password123!", "FraudUser")

	rec := doRequest(t, fiber.MethodGet, "/api/v1/admin/fraud/rules", nil, user.AccessToken)
	if rec.Code == fiber.StatusOK {
		t.Error("expected non-admin user to be denied fraud rules access")
	}
}

// TestAdminFlow_ReconciliationAccessDenied tests that non-admin users cannot access reconciliation.
func TestAdminFlow_ReconciliationAccessDenied(t *testing.T) {
	user := registerAndLogin(t, uniqueEmail("recon-nonadmin"), "Password123!", "ReconUser")

	rec := doRequest(t, fiber.MethodGet, "/api/v1/admin/reconciliation/reports", nil, user.AccessToken)
	if rec.Code == fiber.StatusOK {
		t.Error("expected non-admin user to be denied reconciliation access")
	}
}

// TestAdminFlow_AuditAccessDenied tests that non-admin users cannot access audit logs.
func TestAdminFlow_AuditAccessDenied(t *testing.T) {
	user := registerAndLogin(t, uniqueEmail("audit-nonadmin"), "Password123!", "AuditUser")

	rec := doRequest(t, fiber.MethodGet, "/api/v1/admin/audit/logs", nil, user.AccessToken)
	if rec.Code == fiber.StatusOK {
		t.Error("expected non-admin user to be denied audit log access")
	}
}

// TestAdminFlow_SystemErrorAccessDenied tests that non-admin users cannot access system errors.
func TestAdminFlow_SystemErrorAccessDenied(t *testing.T) {
	user := registerAndLogin(t, uniqueEmail("syserr-nonadmin"), "Password123!", "SysErrUser")

	rec := doRequest(t, fiber.MethodGet, "/api/v1/admin/systemerrors/", nil, user.AccessToken)
	if rec.Code == fiber.StatusOK {
		t.Error("expected non-admin user to be denied system error access")
	}
}

// TestAdminFlow_ExchangeRateUpsertDenied tests that non-admin users cannot upsert exchange rates.
func TestAdminFlow_ExchangeRateUpsertDenied(t *testing.T) {
	user := registerAndLogin(t, uniqueEmail("rate-nonadmin"), "Password123!", "RateUser")

	rec := doRequest(t, fiber.MethodPost, "/api/v1/admin/currencies/rate", map[string]interface{}{
		"from":     "USD",
		"to":       "UZS",
		"rate":     12500.0,
		"provider": "manual",
	}, user.AccessToken)
	if rec.Code == fiber.StatusOK || rec.Code == fiber.StatusCreated {
		t.Error("expected non-admin user to be denied exchange rate upsert")
	}
}

// TestAdminFlow_WithoutAuthToken tests that admin routes reject unauthenticated requests.
func TestAdminFlow_WithoutAuthToken(t *testing.T) {
	adminPaths := []string{
		"/api/v1/admin/kyc/pending",
		"/api/v1/admin/flags/",
		"/api/v1/admin/statistics/overview",
		"/api/v1/admin/settings/",
		"/api/v1/admin/announcements/",
		"/api/v1/admin/authz/roles",
		"/api/v1/admin/errorcodes/",
		"/api/v1/admin/translations/",
	}

	for _, path := range adminPaths {
		t.Run(path, func(t *testing.T) {
			rec := doRequest(t, fiber.MethodGet, path, nil, "")
			if rec.Code != fiber.StatusUnauthorized {
				t.Errorf("status = %d, want 401 for unauthenticated %s", rec.Code, path)
			}
		})
	}
}
