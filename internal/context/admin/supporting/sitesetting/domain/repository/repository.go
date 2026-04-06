package repository

import (
	"context"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/sitesetting/domain/entity"
)

// WriteRepository defines write operations for SiteSetting aggregate.
type WriteRepository interface {
	Save(ctx context.Context, setting *entity.SiteSetting) error
	Update(ctx context.Context, setting *entity.SiteSetting) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*entity.SiteSetting, error)
	FindByKey(ctx context.Context, key string) (*entity.SiteSetting, error)
}

// SiteSettingView is the read projection for list queries.
type SiteSettingView struct {
	ID          string             `json:"id"`
	Key         string             `json:"key"`
	Value       string             `json:"value"`
	SettingType entity.SettingType `json:"setting_type"`
	Description string             `json:"description"`
	CreatedAt   string             `json:"created_at"`
	UpdatedAt   string             `json:"updated_at"`
}

// SiteSettingFilter for list queries.
type SiteSettingFilter struct {
	Key         string
	SettingType string
	Limit       int
	Offset      int
}

// ReadRepository defines read operations for SiteSetting projections.
type ReadRepository interface {
	FindByID(ctx context.Context, id string) (*SiteSettingView, error)
	List(ctx context.Context, filter SiteSettingFilter) ([]*SiteSettingView, int64, error)
}
