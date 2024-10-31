/*
Package aws provides an implementation using aws-sdk-go.

Lambda event type compatibility layer for AWS API Gateway with REST API mode.

See lambda event detail:
https://docs.aws.amazon.com/apigateway/latest/developerguide/set-up-lambda-proxy-integrations.html#api-gateway-simple-proxy-for-lambda-input-format
*/
package aws

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"
	"net/url"
)

type v2api struct {
	*apigatewaymanagementapi.Client
}

func (v v2api) PostToConnection(ctx context.Context, connectionID string, data []byte) (err error) {
	_, err = v.Client.PostToConnection(ctx, &apigatewaymanagementapi.PostToConnectionInput{
		ConnectionId: aws.String(connectionID),
		Data:         data,
	})
	return
}

// NewAPIGatewayManagementClientV2 creates a new API Gateway Management Client instance from the provided parameters. The
// new client will have a custom endpoint that resolves to the application's deployed API.
func NewAPIGatewayManagementClientV2(conf *aws.Config, domain, stage string) APIGatewayManagementAPI {
	var endpoint url.URL
	endpoint.Path = stage
	endpoint.Host = domain
	endpoint.Scheme = "https"
	return &v2api{apigatewaymanagementapi.NewFromConfig(conf.Copy(), func(o *apigatewaymanagementapi.Options) {
		o.BaseEndpoint = aws.String(endpoint.String())
	})}
}
