package application

import "context"

// Query is a marker interface for read operations (no side effects).
// Queries are named descriptively: GetUser, ListAccounts, FindSession.
type Query interface{}

// QueryHandler executes a query and returns a result.
type QueryHandler[Q Query, R any] interface {
	Handle(ctx context.Context, q Q) (R, error)
}
