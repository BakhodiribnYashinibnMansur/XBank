package ptrutil

import "time"

// Ptr returns a pointer to the given value.
func Ptr[T any](v T) *T {
	return &v
}

// Value dereferences a pointer, returning the zero value if nil.
func Value[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

// ValueOr dereferences a pointer, returning fallback if nil.
func ValueOr[T any](p *T, fallback T) T {
	if p == nil {
		return fallback
	}
	return *p
}

// String is a convenience alias for Ptr(s) with type string.
func String(s string) *string { return Ptr(s) }

// Int is a convenience alias for Ptr(n) with type int.
func Int(n int) *int { return Ptr(n) }

// Int64 is a convenience alias for Ptr(n) with type int64.
func Int64(n int64) *int64 { return Ptr(n) }

// Bool is a convenience alias for Ptr(b) with type bool.
func Bool(b bool) *bool { return Ptr(b) }

// Time is a convenience alias for Ptr(t) with type time.Time.
func Time(t time.Time) *time.Time { return Ptr(t) }

// TimeNow returns a pointer to the current UTC time.
func TimeNow() *time.Time { return Ptr(time.Now().UTC()) }

// IsNil returns true if the pointer is nil.
func IsNil[T any](p *T) bool { return p == nil }

// Equal compares two pointers by value. Both nil → true.
func Equal[T comparable](a, b *T) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
