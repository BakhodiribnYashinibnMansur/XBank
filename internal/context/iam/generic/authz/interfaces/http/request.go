package http

import "time"

type CreateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreatePermissionRequest struct {
	Resource    string `json:"resource"`
	Action      string `json:"action"`
	Description string `json:"description"`
}

type AssignPermissionRequest struct {
	RoleID       string `json:"role_id"`
	PermissionID string `json:"permission_id"`
	Scope        string `json:"scope"`
}

type RoleResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsSystem    bool      `json:"is_system"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PermissionResponse struct {
	ID          string    `json:"id"`
	Resource    string    `json:"resource"`
	Action      string    `json:"action"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type PolicyResponse struct {
	ID           string    `json:"id"`
	RoleID       string    `json:"role_id"`
	PermissionID string    `json:"permission_id"`
	Scope        string    `json:"scope"`
	CreatedAt    time.Time `json:"created_at"`
}

type AccessCheckResponse struct {
	Allowed bool   `json:"allowed"`
	Scope   string `json:"scope"`
}
