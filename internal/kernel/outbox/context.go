package outbox

import "context"

type outboxMetaKey struct{}

// OutboxMeta carries aggregate info for outbox entries.
type OutboxMeta struct {
	AggregateType string
	AggregateID   string
}

// WithOutboxMeta injects aggregate metadata into context.
// The outbox publisher uses this to populate aggregate_type and aggregate_id columns.
func WithOutboxMeta(ctx context.Context, aggregateType, aggregateID string) context.Context {
	return context.WithValue(ctx, outboxMetaKey{}, OutboxMeta{
		AggregateType: aggregateType,
		AggregateID:   aggregateID,
	})
}

// ExtractOutboxMeta retrieves aggregate metadata from context.
func ExtractOutboxMeta(ctx context.Context) (OutboxMeta, bool) {
	meta, ok := ctx.Value(outboxMetaKey{}).(OutboxMeta)
	return meta, ok
}
