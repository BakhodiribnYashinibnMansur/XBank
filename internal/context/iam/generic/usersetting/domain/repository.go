package domain

import "context"

type WriteRepository interface {
	Upsert(ctx context.Context, s interface{}) error
	Delete(ctx context.Context, id string) error
	FindByUserIDAndKey(ctx context.Context, userID, key string) (interface{}, error)
}

type UserSettingView struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
	Key    string `json:"key"`
	Value  string `json:"value"`
}

type UserSettingFilter struct {
	UserID string
	Key    string
	Limit  int
	Offset int
}

type ReadRepository interface {
	List(ctx context.Context, filter UserSettingFilter) ([]*UserSettingView, int64, error)
}
