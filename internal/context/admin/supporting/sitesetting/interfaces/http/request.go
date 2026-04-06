package http

// UpsertSettingRequest is the HTTP request body for creating/updating a setting.
type UpsertSettingRequest struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	SettingType string `json:"setting_type"` // general, email, security, payment
	Description string `json:"description"`
}
