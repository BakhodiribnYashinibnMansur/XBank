package domain

import "testing"

func TestNewCurrency(t *testing.T) {
	tests := []struct {
		name          string
		code          string
		currName      string
		symbol        string
		decimalPlaces int
		wantErr       error
	}{
		{"valid USD", "USD", "US Dollar", "$", 2, nil},
		{"valid UZS", "UZS", "Uzbek Sum", "so'm", 0, nil},
		{"valid EUR", "EUR", "Euro", "€", 2, nil},
		{"empty code", "", "US Dollar", "$", 2, ErrMissingCode},
		{"short code", "US", "US Dollar", "$", 2, ErrInvalidCode},
		{"long code", "USDX", "US Dollar", "$", 2, ErrInvalidCode},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewCurrency(tt.code, tt.currName, tt.symbol, tt.decimalPlaces)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("got err %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if c.Code != tt.code {
				t.Errorf("code = %q, want %q", c.Code, tt.code)
			}
			if c.Status != StatusActive {
				t.Errorf("status = %q, want %q", c.Status, StatusActive)
			}
			if c.DecimalPlaces != tt.decimalPlaces {
				t.Errorf("decimal_places = %d, want %d", c.DecimalPlaces, tt.decimalPlaces)
			}
		})
	}
}

func TestCurrency_ActivateDeactivate(t *testing.T) {
	c, _ := NewCurrency("USD", "US Dollar", "$", 2)

	c.Deactivate()
	if c.Status != StatusInactive {
		t.Errorf("after Deactivate: status = %q, want %q", c.Status, StatusInactive)
	}

	c.Activate()
	if c.Status != StatusActive {
		t.Errorf("after Activate: status = %q, want %q", c.Status, StatusActive)
	}
}
