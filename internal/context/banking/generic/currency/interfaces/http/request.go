package http

import "time"

// CreateCurrencyRequest represents the body for creating a currency.
type CreateCurrencyRequest struct {
	Code          string `json:"code"`
	Name          string `json:"name"`
	Symbol        string `json:"symbol"`
	DecimalPlaces int    `json:"decimal_places"`
}

// UpdateCurrencyRequest represents the body for updating a currency.
type UpdateCurrencyRequest struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Symbol        string `json:"symbol"`
	DecimalPlaces int    `json:"decimal_places"`
}

// ToggleStatusRequest represents the body for toggling currency status.
type ToggleStatusRequest struct {
	ID     string `json:"id"`
	Active bool   `json:"active"`
}

// CurrencyResponse represents a currency in API responses.
type CurrencyResponse struct {
	ID            string    `json:"id"`
	Code          string    `json:"code"`
	Name          string    `json:"name"`
	Symbol        string    `json:"symbol"`
	DecimalPlaces int       `json:"decimal_places"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
