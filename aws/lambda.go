package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambda/handlertrace"
	"github.com/yacchi/lambda-http-adaptor/registry"
	"github.com/yacchi/lambda-http-adaptor/types"
	"net/http"
	"os"
)

type LambdaIntegrationType int

const (
	UnknownLambdaIntegrationType LambdaIntegrationType = iota
	APIGatewayRESTIntegration
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
}

func (t integrationTypeChecker) IntegrationType() LambdaIntegrationType {
	if t.Resource != nil {
		return APIGatewayRESTIntegration
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

type LambdaHandler struct {
	httpHandler http.Handler
}

func NewLambdaHandler(h http.Handler) lambda.Handler {
	return &LambdaHandler{
		httpHandler: h,
	}
}

func (l LambdaHandler) InvokeRESTAPI(ctx context.Context, e *events.APIGatewayProxyRequest) (r *events.APIGatewayProxyResponse, err error) {
	req, multiValue, err := NewRESTAPIRequest(ctx, e)
	if err != nil {
		return nil, err
	}

	w := NewResponseWriter()
	l.httpHandler.ServeHTTP(w, req)
	return RESTAPITargetResponse(w, multiValue)
}

func (l LambdaHandler) InvokeHTTPAPI(ctx context.Context, e *events.APIGatewayV2HTTPRequest) (r *APIGatewayV2HTTPResponse, err error) {
	req, err := NewHTTPAPIRequest(ctx, e)
	if err != nil {
		return nil, err
	}

	w := NewResponseWriter()
	l.httpHandler.ServeHTTP(w, req)
	return HTTPAPIResponse(w)
}

func (l LambdaHandler) InvokeALBTargetGroup(ctx context.Context, request *events.ALBTargetGroupRequest) (r *events.ALBTargetGroupResponse, err error) {
	req, multiValue, err := NewALBTargetGroupRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	w := NewResponseWriter()
	l.httpHandler.ServeHTTP(w, req)
	return ALBTargetResponse(w, multiValue)
}

func (l LambdaHandler) Invoke(ctx context.Context, payload []byte) ([]byte, error) {
	trace := handlertrace.FromContext(ctx)

	var (
		checker integrationTypeChecker
		res     interface{}
		err     error
	)

	if err := json.Unmarshal(payload, &checker); err != nil {
		return nil, err
	}

	switch checker.IntegrationType() {
	case APIGatewayRESTIntegration:
		event := &events.APIGatewayProxyRequest{}
		if err := json.Unmarshal(payload, event); err != nil {
			return nil, err
		}
		if trace.RequestEvent != nil {
			trace.RequestEvent(ctx, payload)
		}
		res, err = l.InvokeRESTAPI(ctx, event)
	case APIGatewayHTTPIntegration:
		event := &events.APIGatewayV2HTTPRequest{}
		if err := json.Unmarshal(payload, event); err != nil {
			return nil, err
		}
		if trace.RequestEvent != nil {
			trace.RequestEvent(ctx, payload)
		}
		res, err = l.InvokeHTTPAPI(ctx, event)
	case ALBTargetGroupIntegration:
		event := &events.ALBTargetGroupRequest{}
		if err := json.Unmarshal(payload, event); err != nil {
			return nil, err
		}
		if trace.RequestEvent != nil {
			trace.RequestEvent(ctx, payload)
		}
		res, err = l.InvokeALBTargetGroup(ctx, event)
	default:
		return nil, fmt.Errorf("unknown lambda integration type")
	}

	if err != nil {
		return nil, err
	}

	if trace.ResponseEvent != nil {
		trace.ResponseEvent(ctx, res)
	}

	if responseBytes, err := json.Marshal(res); err != nil {
		return nil, err
	} else {
		return responseBytes, nil
	}
}

type LambdaAdaptor struct {
	h lambda.Handler
}

func (l LambdaAdaptor) ListenAndServe() error {
	lambda.StartHandler(l.h)
	return nil
}

func (l LambdaAdaptor) Shutdown(ctx context.Context) error {
	return fmt.Errorf("aws_lambda: unsupported shutdown")
}

func NewLambdaAdaptor(addr string, h http.Handler) types.Adaptor {
	return &LambdaAdaptor{
		h: NewLambdaHandler(h),
	}
}

func init() {
	registry.Registry.AddAdaptor("aws_lambda", LambdaDetector, NewLambdaAdaptor)
}
