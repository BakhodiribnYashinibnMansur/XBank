package dto

import "time"

type ConvertRequest struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Amount int64  `json:"amount"`
}

type UpsertRateRequest struct {
	From     string `json:"from"`
	To       string `json:"to"`
	BuyRate  int64  `json:"buy_rate"`
	SellRate int64  `json:"sell_rate"`
}

type RateResponse struct {
	ID           string    `json:"id"`
	FromCurrency string    `json:"from_currency"`
	ToCurrency   string    `json:"to_currency"`
	BuyRate      int64     `json:"buy_rate"`
	SellRate     int64     `json:"sell_rate"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ConvertResponse struct {
	From           string `json:"from"`
	To             string `json:"to"`
	OriginalAmount int64  `json:"original_amount"`
	ConvertedAmount int64 `json:"converted_amount"`
	RateUsed       int64  `json:"rate_used"`
}
