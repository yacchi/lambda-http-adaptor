package middlewares

import "github.com/yacchi/lambda-http-adaptor/aws"

var (
	APIGatewayStripStageVar   = aws.StripStageVar
	APIGatewayWebsocketRouter = aws.WebsocketRouter
)
