package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambda/handlertrace"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go/aws/session"
	"net/http"
)

type LambdaHandlerOption func(handler *LambdaHandler)

type SDKSessionProvider func() (*session.Session, error)

func WithAWSSessionProvider(sp SDKSessionProvider) LambdaHandlerOption {
	return func(handler *LambdaHandler) {
		handler.sessProv = sp
	}
}

type SDKConfigProvider func(ctx context.Context) (aws.Config, error)

func WithAWSConfigProvider(sp SDKConfigProvider) LambdaHandlerOption {
	return func(handler *LambdaHandler) {
		handler.confProv = sp
	}
}

type LambdaHandler struct {
	httpHandler  http.Handler
	sessProv     SDKSessionProvider
	sess         *session.Session
	confProv     SDKConfigProvider
	conf         *aws.Config
	apiGW        APIGatewayManagementAPI
	wsPathPrefix string
}

func NewLambdaHandlerWithOption(h http.Handler, options []interface{}) lambda.Handler {
	handler := &LambdaHandler{
		httpHandler: h,
		confProv: func(ctx context.Context) (aws.Config, error) {
			return config.LoadDefaultConfig(ctx)
		},
		wsPathPrefix: DefaultWebsocketPathPrefix,
	}

	for _, opt := range options {
		if lambdaOpt, ok := opt.(LambdaHandlerOption); ok {
			lambdaOpt(handler)
		}
	}

	return handler
}

func NewLambdaHandler(h http.Handler) lambda.Handler {
	return NewLambdaHandlerWithOption(h, nil)
}

func (l *LambdaHandler) InvokeRESTAPI(ctx context.Context, e *events.APIGatewayProxyRequest) (r *events.APIGatewayProxyResponse, err error) {
	req, multiValue, err := NewRESTAPIRequest(ctx, e)
	if err != nil {
		return nil, err
	}

	w := NewResponseWriter()
	l.httpHandler.ServeHTTP(w, req)
	return RESTAPITargetResponse(w, multiValue)
}

func (l *LambdaHandler) InvokeHTTPAPI(ctx context.Context, e *events.APIGatewayV2HTTPRequest) (r *events.APIGatewayV2HTTPResponse, err error) {
	req, err := NewHTTPAPIRequest(ctx, e)
	if err != nil {
		return nil, err
	}

	w := NewResponseWriter()
	l.httpHandler.ServeHTTP(w, req)
	return HTTPAPIResponse(w)
}

func (l *LambdaHandler) InvokeALBTargetGroup(ctx context.Context, request *events.ALBTargetGroupRequest) (r *events.ALBTargetGroupResponse, err error) {
	req, multiValue, err := NewALBTargetGroupRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	w := NewResponseWriter()
	l.httpHandler.ServeHTTP(w, req)
	return ALBTargetResponse(w, multiValue)
}

func (l *LambdaHandler) ProvideAPIGatewayClient(ctx context.Context, request *events.APIGatewayWebsocketProxyRequest) (client APIGatewayManagementAPI, err error) {
	if l.apiGW != nil {
		return l.apiGW, nil
	}

	if l.sessProv != nil {
		if l.sess == nil {
			if l.sess, err = l.sessProv(); err != nil {
				return nil, err
			}
		}
		l.apiGW = NewAPIGatewayManagementClientV1(l.sess, request.RequestContext.DomainName, request.RequestContext.Stage)
	} else if l.confProv != nil {
		if l.conf == nil {
			if conf, err := l.confProv(ctx); err != nil {
				return nil, err
			} else {
				l.conf = &conf
			}
		}
		l.apiGW = NewAPIGatewayManagementClientV2(l.conf, request.RequestContext.DomainName, request.RequestContext.Stage)
	}

	if l.apiGW != nil {
		return l.apiGW, nil
	} else {
		return nil, fmt.Errorf("can not provide client for API Gateway management API")
	}
}

func (l *LambdaHandler) InvokeWebsocketAPI(ctx context.Context, request *events.APIGatewayWebsocketProxyRequest) (r *events.APIGatewayProxyResponse, err error) {
	req, multiValue, err := NewWebsocketRequest(ctx, request, l.wsPathPrefix)
	if err != nil {
		return nil, err
	}

	routeKey := request.RequestContext.RouteKey

	if routeKey == "$connect" || routeKey == "$disconnect" {
		w := NewResponseWriter()
		l.httpHandler.ServeHTTP(w, req)
		return RESTAPITargetResponse(w, multiValue)
	} else {
		if apiGW, err := l.ProvideAPIGatewayClient(ctx, request); err != nil {
			return nil, err
		} else {
			w := NewWebsocketResponseWriter(ctx, apiGW, request)
			l.httpHandler.ServeHTTP(w, req)
			return WebsocketResponse(w, multiValue)
		}
	}
}

func (l *LambdaHandler) Invoke(ctx context.Context, payload []byte) ([]byte, error) {
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
	case APIGatewayWebsocketIntegration:
		event := &events.APIGatewayWebsocketProxyRequest{}
		if err := json.Unmarshal(payload, event); err != nil {
			return nil, err
		}
		res, err = l.InvokeWebsocketAPI(ctx, event)
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
