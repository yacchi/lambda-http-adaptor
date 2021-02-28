package aws

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDefaultRouteSelector(t *testing.T) {
	// https://docs.aws.amazon.com/apigateway/latest/developerguide/websocket-api-develop-routes.html#apigateway-websocket-api-route-selection-expressions
	sampleBody := `{
    "service" : "chat",
    "action" : "join",
    "data" : {
        "room" : "room1234"
   }
}`
	reqCtx := events.APIGatewayWebsocketProxyRequestContext{
		RouteKey:   "$default",
		DomainName: "u44by84te2.execute-api.ap-northeast-1.amazonaws.com",
		Stage:      "production",
	}

	cases := []struct {
		req   events.APIGatewayWebsocketProxyRequest
		exp   string
		route string
	}{
		{
			req: events.APIGatewayWebsocketProxyRequest{
				Body:           sampleBody,
				RequestContext: reqCtx,
			},
			exp:   DefaultRouteSelectionExpression,
			route: "join",
		},
		{
			req: events.APIGatewayWebsocketProxyRequest{
				Body:           sampleBody,
				RequestContext: reqCtx,
			},
			exp:   "$request.body.action",
			route: "join",
		},
		{
			req: events.APIGatewayWebsocketProxyRequest{
				Body:           sampleBody,
				RequestContext: reqCtx,
			},
			exp:   "${request.body.action}",
			route: "join",
		},
		{
			req: events.APIGatewayWebsocketProxyRequest{
				Body:           sampleBody,
				RequestContext: reqCtx,
			},
			exp:   "${request.body.service}/${request.body.action}",
			route: "chat/join",
		},
		{
			req: events.APIGatewayWebsocketProxyRequest{
				Body:           sampleBody,
				RequestContext: reqCtx,
			},
			exp:   "${request.body.action}-${request.body.invalidPath}",
			route: "join-",
		},
		{
			req: events.APIGatewayWebsocketProxyRequest{
				Body:           `{"action": "echo?message=1"}`,
				RequestContext: reqCtx,
			},
			exp:   "${request.body.action}",
			route: "echo?message=1",
		},
	}

	for _, c := range cases {
		route, err := RouteSelector(&c.req, c.exp)
		assert.NoError(t, err)
		assert.Equal(t, route, c.route)
	}
}
