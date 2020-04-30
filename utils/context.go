package utils

import (
	"context"
	"github.com/yacchi/lambda-http-adaptor/internal"
)

func RawRequestValue(ctx context.Context) (interface{}, bool) {
	v := ctx.Value(internal.RawRequestValueContextKey)
	return v, v != nil
}
