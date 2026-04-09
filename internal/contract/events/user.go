package events

import "time"

// UserCreated is published when a new user registers.
type UserCreated struct {
	UserID     string    `json:"user_id"`
	Email      string    `json:"email"`
	FullName   string    `json:"full_name"`
	OccurredAt time.Time `json:"occurred_at"`
}

// UserUpdated is published when user profile is modified.
type UserUpdated struct {
	UserID     string    `json:"user_id"`
	OccurredAt time.Time `json:"occurred_at"`
}

// UserDeleted is published when a user account is deactivated.
type UserDeleted struct {
	UserID     string    `json:"user_id"`
	OccurredAt time.Time `json:"occurred_at"`
}
