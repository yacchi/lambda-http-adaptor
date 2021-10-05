package aws

import (
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/yacchi/lambda-http-adaptor/log"
	"github.com/yacchi/lambda-http-adaptor/utils"
	"net/http"
	"strings"
)

// StripBasePath Strip base path from Request path under the AWS API Gateway environment with deployment stage and API Mapping feature.
func StripBasePath(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if rawReq, ok := utils.RawRequestValue(request.Context()); ok {
			var stage string

			switch req := rawReq.(type) {
			case *events.APIGatewayProxyRequest:
				// If CustomDomain API Mapping was used, BasePath will be inserted instead of StageName.
				// Recreate the original path from ResourcePath and PathParameters.
				request.URL.Path = ExpandPathParameters(req.RequestContext.ResourcePath, req.PathParameters)
			case *events.APIGatewayV2HTTPRequest:
				stage = req.RequestContext.Stage
			case *events.APIGatewayWebsocketProxyRequest:
				stage = req.RequestContext.Stage
			}

			if stage != "" {
				stage = "/" + stage
				pos := strings.Index(request.URL.Path, stage)
				if pos == 0 {
					request.URL.Path = request.URL.Path[len(stage):]
				}
			}
		}
		next.ServeHTTP(writer, request)
	})
}

func WebsocketRouter(next http.Handler, routeExpression string) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var event *events.APIGatewayWebsocketProxyRequest
		if raw, ok := utils.RawRequestValue(request.Context()); !ok {
			next.ServeHTTP(writer, request)
			return
		} else if event, ok = raw.(*events.APIGatewayWebsocketProxyRequest); !ok {
			next.ServeHTTP(writer, request)
			return
		}

		routeKey := event.RequestContext.RouteKey

		// extract path from route key and expression
		if routeKey == "$default" && routeExpression != "" {
			if route, err := RouteSelector(event, routeExpression); err != nil {
				log.Warning(fmt.Errorf("can not extract route: %w", err))
			} else {
				request.URL.Path = strings.Replace(request.URL.Path, routeKey, route, 1)
			}
		}

		next.ServeHTTP(writer, request)
	})
}
