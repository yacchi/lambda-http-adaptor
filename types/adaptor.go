package types

import (
	"context"
	"net/http"
)

type Adaptor interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}

type AdaptorInitializer func(addr string, h http.Handler) Adaptor
