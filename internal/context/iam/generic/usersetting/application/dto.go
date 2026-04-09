package application

// UpsertSettingRequest is the DTO for creating or updating a user setting.
type UpsertSettingRequest struct {
	UserID string `json:"user_id"`
	Key    string `json:"key"`
	Value  string `json:"value"`
}
