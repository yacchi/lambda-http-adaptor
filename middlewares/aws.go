package middlewares

import (
	"github.com/yacchi/lambda-http-adaptor/aws"
	"net/http"
)

var (
	APIGatewayStripStageVar   = aws.StripStageVar
	APIGatewayWebsocketRouter = aws.WebsocketRouter
)

func StripStageVar(next http.HandlerFunc) http.HandlerFunc {
	return aws.StripStageVar(next).ServeHTTP
}
