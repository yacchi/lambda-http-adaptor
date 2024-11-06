package aws

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambda/handlertrace"
	"github.com/yacchi/lambda-http-adaptor/registry"
	"github.com/yacchi/lambda-http-adaptor/types"
	"net/http"
	"os"
)

var DEBUGDumpPayload = os.Getenv("DEBUG_DUMP_PAYLOAD")
var LambdaInvokeMode = os.Getenv("LAMBDA_INVOKE_MODE")

// WebsocketResponseMode
// * return - use lambda return value as response
// * post_to_connection - use PostToConnection API to send response
var WebsocketResponseMode = os.Getenv("WEBSOCKET_RESPONSE_MODE")

type LambdaIntegrationType int

const (
	UnknownLambdaIntegrationType LambdaIntegrationType = iota
	APIGatewayRESTIntegration
	APIGatewayWebsocketIntegration
	APIGatewayHTTPIntegration
	ALBTargetGroupIntegration
	LambdaFunctionURLIntegration
)

type integrationTypeChecker struct {
	// 'resource' parameter only has REST API event.
	Resource *string `json:"resource"`
	// 'version' only has API Gateway V2 payload (HTTP API mode or Lambda FunctionURL).
	Version *string `json:"version"`
	// 'pathParameters' parameter only has API Gateway V2 payload.
	// However, it is always nil for Function URLs.
	PathParameters map[string]string `json:"pathParameters"`
	// 'routeKey' parameter only has API Gateway V2 payload.
	// However, in the case of Function URLs, it is always $default.
	RouteKey string `json:"routeKey"`
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
		if t.RouteKey == "$default" && t.PathParameters == nil {
			return LambdaFunctionURLIntegration
		}
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
	h *LambdaHandler
}

func (l LambdaAdaptor) ListenAndServe() error {
	ctx := context.Background()
	if DEBUGDumpPayload != "" && (DEBUGDumpPayload == "1" || DEBUGDumpPayload == "true") {
		ctx = handlertrace.NewContext(ctx, handlertrace.HandlerTrace{
			RequestEvent: func(ctx context.Context, payload interface{}) {
				fmt.Printf("Request payload: %s\n", payload)
			},
			ResponseEvent: func(ctx context.Context, payload interface{}) {
				fmt.Printf("Response payload: %+v\n", payload)
			},
		})
	}
	lambda.StartHandlerFunc(l.h.Invoke, lambda.WithContext(ctx))
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
