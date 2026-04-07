package http

type AddContactRequest struct {
	ContactID  string `json:"contact_id"`
	CustomName string `json:"custom_name"`
}
