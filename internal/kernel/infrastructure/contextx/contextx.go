package contextx

import "context"

// ctxKey is an unexported type to prevent collisions with context keys from other packages.
type ctxKey int

const (
	keyRequestID ctxKey = iota
	keyCorrelationID
	keyUserID
	keyUserRole
	keySessionID
	keyIPAddress
	keyUserAgent
	keyDeviceID
	keyTraceID
)

// --- RequestID ---

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, keyRequestID, id)
}

func GetRequestID(ctx context.Context) string {
	v, _ := ctx.Value(keyRequestID).(string)
	return v
}

// --- CorrelationID ---

func WithCorrelationID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, keyCorrelationID, id)
}

func GetCorrelationID(ctx context.Context) string {
	v, _ := ctx.Value(keyCorrelationID).(string)
	return v
}

// --- UserID ---

func WithUserID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, keyUserID, id)
}

func GetUserID(ctx context.Context) string {
	v, _ := ctx.Value(keyUserID).(string)
	return v
}

// --- UserRole ---

func WithUserRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, keyUserRole, role)
}

func GetUserRole(ctx context.Context) string {
	v, _ := ctx.Value(keyUserRole).(string)
	return v
}

// --- SessionID ---

func WithSessionID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, keySessionID, id)
}

func GetSessionID(ctx context.Context) string {
	v, _ := ctx.Value(keySessionID).(string)
	return v
}

// --- IPAddress ---

func WithIPAddress(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, keyIPAddress, ip)
}

func GetIPAddress(ctx context.Context) string {
	v, _ := ctx.Value(keyIPAddress).(string)
	return v
}

// --- UserAgent ---

func WithUserAgent(ctx context.Context, ua string) context.Context {
	return context.WithValue(ctx, keyUserAgent, ua)
}

func GetUserAgent(ctx context.Context) string {
	v, _ := ctx.Value(keyUserAgent).(string)
	return v
}

// --- DeviceID ---

func WithDeviceID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, keyDeviceID, id)
}

func GetDeviceID(ctx context.Context) string {
	v, _ := ctx.Value(keyDeviceID).(string)
	return v
}

// --- TraceID ---

func WithTraceID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, keyTraceID, id)
}

func GetTraceID(ctx context.Context) string {
	v, _ := ctx.Value(keyTraceID).(string)
	return v
}
