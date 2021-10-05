package middlewares

import (
	"github.com/yacchi/lambda-http-adaptor/aws"
	"net/http"
)

var (
	APIGatewayStripStageVar   = aws.StripBasePath
	APIGatewayStripBasePath   = aws.StripBasePath
	APIGatewayWebsocketRouter = aws.WebsocketRouter
)

func StripStageVar(next http.HandlerFunc) http.HandlerFunc {
	return aws.StripBasePath(next).ServeHTTP
}

func StripBasePath(next http.HandlerFunc) http.HandlerFunc {
	return aws.StripBasePath(next).ServeHTTP
}
