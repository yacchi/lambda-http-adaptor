package azure

import (
	"context"
	"github.com/yacchi/lambda-http-adaptor/registry"
	"github.com/yacchi/lambda-http-adaptor/types"
	"net/http"
	"os"
)

const (
	FunctionsHTTPWorkerPortEnvKey = "FUNCTIONS_HTTPWORKER_PORT"
)

type FunctionsAdaptor struct {
	s *http.Server
}

func (f FunctionsAdaptor) ListenAndServe() error {
	return f.s.ListenAndServe()
}

func (f FunctionsAdaptor) Shutdown(ctx context.Context) error {
	return f.s.Shutdown(ctx)
}

func FunctionsDetector() bool {
	_, ok := os.LookupEnv(FunctionsHTTPWorkerPortEnvKey)
	return ok
}

func NewAzureFunctionsAdaptor(addr string, h http.Handler, opts []interface{}) types.Adaptor {
	port := os.Getenv(FunctionsHTTPWorkerPortEnvKey)
	return &FunctionsAdaptor{
		s: &http.Server{
			Addr:    ":" + port,
			Handler: h,
		},
	}
}

func init() {
	registry.Registry.AddAdaptor("azure_functions", FunctionsDetector, NewAzureFunctionsAdaptor)
}
