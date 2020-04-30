package internal

import "context"

type contextKey int

const (
	RawRequestValueContextKey contextKey = iota
)

func NewRawRequestValueContext(ctx context.Context, v interface{}) context.Context {
	return context.WithValue(ctx, RawRequestValueContextKey, v)
}
