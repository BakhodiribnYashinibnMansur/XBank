package domain

import (
	"testing"
)

func TestRate_Convert(t *testing.T) {
	tests := []struct {
		name     string
		sellRate int64
		amount   int64
		want     int64
	}{
		{
			name:     "1 USD to UZS at 12650.50",
			sellRate: 1265050,
			amount:   100, // 1 USD in minor units (cents)
			want:     1265050,
		},
		{
			name:     "10 USD to UZS",
			sellRate: 1265050,
			amount:   1000, // 10 USD
			want:     12650500,
		},
		{
			name:     "zero amount",
			sellRate: 1265050,
			amount:   0,
			want:     0,
		},
		{
			name:     "small amount 0.01 USD",
			sellRate: 1265050,
			amount:   1, // 0.01 USD
			want:     12650,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Rate{
				FromCurrency: "USD",
				ToCurrency:   "UZS",
				SellRate:     tt.sellRate,
			}
			got := r.Convert(tt.amount)
			if got != tt.want {
				t.Errorf("Convert(%d) = %d, want %d", tt.amount, got, tt.want)
			}
		})
	}
}

func TestRate_Fields(t *testing.T) {
	r := &Rate{
		ID:           "rate-1",
		FromCurrency: "USD",
		ToCurrency:   "UZS",
		BuyRate:      1260000,
		SellRate:     1265050,
	}

	if r.FromCurrency != "USD" {
		t.Errorf("FromCurrency expected USD, got: %s", r.FromCurrency)
	}
	if r.ToCurrency != "UZS" {
		t.Errorf("ToCurrency expected UZS, got: %s", r.ToCurrency)
	}
	if r.BuyRate >= r.SellRate {
		// BuyRate should typically be less than SellRate (bank spread)
		t.Log("Note: BuyRate >= SellRate (no spread)")
	}
}
