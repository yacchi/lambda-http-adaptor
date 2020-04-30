package registry

import (
	"github.com/yacchi/lambda-http-adaptor/types"
	"net/http"
)

type provider struct {
	Name        string
	EnvDetector types.EnvironmentDetector
	Init        types.AdaptorInitializer
}

type registry struct {
	providers []*provider
}

func (r registry) AddAdaptor(name string, detector types.EnvironmentDetector, adaptor types.AdaptorInitializer) {
	r.providers = append(r.providers, &provider{
		Name:        name,
		EnvDetector: detector,
		Init:        adaptor,
	})
}

func (r registry) GetAdaptor(addr string, h http.Handler) types.Adaptor {
	for _, d := range r.providers {
		if d.EnvDetector() {
			return d.Init(addr, h)
		}
	}
	return nil
}

var Registry registry

func GetAdaptor(addr string, h http.Handler) types.Adaptor {
	return Registry.GetAdaptor(addr, h)
}
