package domain

import "testing"

func TestNewAppMetric(t *testing.T) {
	tests := []struct {
		name    string
		mName   string
		value   float64
		labels  map[string]string
		wantErr bool
	}{
		{"valid with labels", "http_requests_total", 1234.5, map[string]string{"method": "GET"}, false},
		{"valid nil labels", "cpu_usage", 85.3, nil, false},
		{"empty name", "", 100, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewAppMetric(tt.mName, tt.value, tt.labels)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if m.Name != tt.mName {
				t.Errorf("name = %q, want %q", m.Name, tt.mName)
			}
			if m.Value != tt.value {
				t.Errorf("value = %f, want %f", m.Value, tt.value)
			}
			if m.Labels == nil {
				t.Error("labels should not be nil (initialized to empty map)")
			}
			if m.CollectedAt.IsZero() {
				t.Error("collected_at should not be zero")
			}
		})
	}
}
