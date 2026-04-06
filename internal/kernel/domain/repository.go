package domain

import "context"

// WriteRepository defines the write-side port for aggregate persistence.
type WriteRepository[T any] interface {
	Save(ctx context.Context, entity *T) error
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*T, error)
}

// ReadRepository defines the read-side port for projections/queries.
type ReadRepository[T any] interface {
	FindByID(ctx context.Context, id string) (*T, error)
	List(ctx context.Context, filter Filter) ([]*T, int64, error)
}

// Filter is a generic filter for list queries.
type Filter struct {
	Limit  int
	Offset int
	Search string
	SortBy string
	Order  string // "ASC" or "DESC"
}

// DefaultFilter returns a filter with sensible defaults.
func DefaultFilter() Filter {
	return Filter{
		Limit:  20,
		Offset: 0,
		Order:  "DESC",
		SortBy: "created_at",
	}
}
