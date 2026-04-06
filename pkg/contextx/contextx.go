// Package contextx provides type-safe context key management.
// Prevents string key collisions and ensures compile-time safety.
package contextx

import "context"

type ctxKey int

const (
	keyRequestID ctxKey = iota
	keyUserID
	keyUserRole
	keySessionID
	keyIPAddress
	keyUserAgent
)

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, keyRequestID, id)
}

func GetRequestID(ctx context.Context) string {
	v, _ := ctx.Value(keyRequestID).(string)
	return v
}

func WithUserID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, keyUserID, id)
}

func GetUserID(ctx context.Context) string {
	v, _ := ctx.Value(keyUserID).(string)
	return v
}

func WithUserRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, keyUserRole, role)
}

func GetUserRole(ctx context.Context) string {
	v, _ := ctx.Value(keyUserRole).(string)
	return v
}

func WithSessionID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, keySessionID, id)
}

func GetSessionID(ctx context.Context) string {
	v, _ := ctx.Value(keySessionID).(string)
	return v
}

func WithIPAddress(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, keyIPAddress, ip)
}

func GetIPAddress(ctx context.Context) string {
	v, _ := ctx.Value(keyIPAddress).(string)
	return v
}

func WithUserAgent(ctx context.Context, ua string) context.Context {
	return context.WithValue(ctx, keyUserAgent, ua)
}

func GetUserAgent(ctx context.Context) string {
	v, _ := ctx.Value(keyUserAgent).(string)
	return v
}
