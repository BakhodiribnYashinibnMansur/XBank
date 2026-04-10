package domain

import "testing"

func TestNewTemplate(t *testing.T) {
	tests := []struct {
		name    string
		slug    string
		channel Channel
		subject string
		body    string
		locale  string
		wantErr error
	}{
		{"valid email", "welcome_email", ChannelEmail, "Welcome!", "Hello {{.Name}}", "en", nil},
		{"valid sms", "otp_sms", ChannelSMS, "", "Your OTP: {{.Code}}", "uz", nil},
		{"valid push", "promo_push", ChannelPush, "", "New offer!", "ru", nil},
		{"empty slug", "", ChannelEmail, "Sub", "Body", "en", ErrMissingSlug},
		{"empty body", "test", ChannelEmail, "Sub", "", "en", ErrMissingBody},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := NewTemplate(tt.slug, tt.channel, tt.subject, tt.body, tt.locale)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("got err %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tmpl.Slug != tt.slug {
				t.Errorf("slug = %q, want %q", tmpl.Slug, tt.slug)
			}
			if tmpl.Status != StatusDraft {
				t.Errorf("status = %q, want %q", tmpl.Status, StatusDraft)
			}
			if tmpl.Version != 1 {
				t.Errorf("version = %d, want 1", tmpl.Version)
			}
		})
	}
}

func TestTemplate_Lifecycle(t *testing.T) {
	tmpl, _ := NewTemplate("test", ChannelEmail, "Sub", "Body", "en")

	tmpl.Activate()
	if tmpl.Status != StatusActive {
		t.Errorf("after Activate: status = %q, want %q", tmpl.Status, StatusActive)
	}

	tmpl.Archive()
	if tmpl.Status != StatusArchived {
		t.Errorf("after Archive: status = %q, want %q", tmpl.Status, StatusArchived)
	}
}

func TestTemplate_UpdateBody(t *testing.T) {
	tmpl, _ := NewTemplate("test", ChannelEmail, "Sub", "Old body", "en")

	if err := tmpl.UpdateBody("New Sub", "New body"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tmpl.Body != "New body" {
		t.Errorf("body = %q, want %q", tmpl.Body, "New body")
	}
	if tmpl.Subject != "New Sub" {
		t.Errorf("subject = %q, want %q", tmpl.Subject, "New Sub")
	}
	if tmpl.Version != 2 {
		t.Errorf("version = %d, want 2", tmpl.Version)
	}

	// empty body should fail
	if err := tmpl.UpdateBody("Sub", ""); err != ErrMissingBody {
		t.Errorf("empty body: got err %v, want %v", err, ErrMissingBody)
	}
	// version should not change on error
	if tmpl.Version != 2 {
		t.Errorf("version after error = %d, want 2", tmpl.Version)
	}
}
