package aws

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/yacchi/lambda-http-adaptor/registry"
	"github.com/yacchi/lambda-http-adaptor/types"
	"net/http"
	"os"
)

type LambdaIntegrationType int

const (
	UnknownLambdaIntegrationType LambdaIntegrationType = iota
	APIGatewayRESTIntegration
	APIGatewayWebsocketIntegration
	APIGatewayHTTPIntegration
	ALBTargetGroupIntegration
)

type integrationTypeChecker struct {
	// 'resource' parameter only has REST API event.
	Resource *string `json:"resource"`
	// 'version' parameter only has HTTP API mode event.
	Version *string `json:"version"`
	// 'http_method' parameter has event of API Gateway REST API mode and ALB target group mode.
	// However, ALB target group mode has not 'resource' parameter.
	HTTPMethod *string `json:"httpMethod"`

	RequestContext struct {
		// 'connectionID' parameter nly has API Gateway Websocket mode event.
		ConnectionID *string `json:"connectionId"`
	} `json:"requestContext"`
}

func (t integrationTypeChecker) IntegrationType() LambdaIntegrationType {
	if t.Resource != nil {
		if t.RequestContext.ConnectionID == nil {
			return APIGatewayRESTIntegration
		} else {
			return APIGatewayWebsocketIntegration
		}
	}
	if t.RequestContext.ConnectionID != nil {
		return APIGatewayWebsocketIntegration
	}
	if t.Version != nil {
		return APIGatewayHTTPIntegration
	}
	if t.HTTPMethod != nil && t.Resource == nil {
		return ALBTargetGroupIntegration
	}
	return UnknownLambdaIntegrationType
}

func LambdaDetector() bool {
	if os.Getenv("AWS_EXECUTION_ENV") != "" {
		return true
	}
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		return true
	}
	if os.Getenv("AWS_LAMBDA_FUNCTION_VERSION") != "" {
		return true
	}
	return false
}

type LambdaAdaptor struct {
	h lambda.Handler
}

func (l LambdaAdaptor) ListenAndServe() error {
	lambda.Start(l.h)
	return nil
}

func (l LambdaAdaptor) Shutdown(ctx context.Context) error {
	return fmt.Errorf("aws_lambda: unsupported shutdown")
}

func NewLambdaAdaptor(addr string, h http.Handler, options []interface{}) types.Adaptor {
	return &LambdaAdaptor{
		h: NewLambdaHandlerWithOption(h, options),
	}
}

func init() {
	registry.Registry.AddAdaptor("aws_lambda", LambdaDetector, NewLambdaAdaptor)
}
