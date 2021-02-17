package middlewares

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/yacchi/lambda-http-adaptor/utils"
	"net/http"
	"strings"
)

// Strip stage var from Request path under the AWS API Gateway environment with deployment stage feature.
func StripStageVar(next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if rawReq, ok := utils.RawRequestValue(request.Context()); ok {
			var stage string

			switch req := rawReq.(type) {
			case *events.APIGatewayProxyRequest:
				stage = req.RequestContext.Stage
			case *events.APIGatewayV2HTTPRequest:
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
	}
}
