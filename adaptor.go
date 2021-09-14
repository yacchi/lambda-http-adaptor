package adaptor

import (
	"fmt"
	"github.com/yacchi/lambda-http-adaptor/registry"
	"net/http"
)

func ListenAndServe(addr string, h http.Handler) error {
	if h == nil {
		h = http.DefaultServeMux
	}

	adaptor := registry.GetAdaptor(addr, h)
	if adaptor == nil {
		return fmt.Errorf("adaptor: not found")
	}

	return adaptor.ListenAndServe()
}

func ListenAndServeWithOptions(addr string, h http.Handler, opts ...interface{}) error {
	if h == nil {
		h = http.DefaultServeMux
	}

	adaptor := registry.GetAdaptor(addr, h, opts...)
	if adaptor == nil {
		return fmt.Errorf("adaptor: not found")
	}

	return adaptor.ListenAndServe()
}
