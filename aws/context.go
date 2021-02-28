package aws

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/yacchi/lambda-http-adaptor/utils"
)

func GetRawRequestContext(ctx context.Context) interface{} {
	rawReq, ok := utils.RawRequestValue(ctx)
	if !ok {
		return nil
	}
	switch req := rawReq.(type) {
	case *events.APIGatewayProxyRequest:
		return &req.RequestContext
	case *events.APIGatewayV2HTTPRequest:
		return &req.RequestContext
	case *events.ALBTargetGroupRequest:
		return &req.RequestContext
	case *events.APIGatewayWebsocketProxyRequest:
		return &req.RequestContext
	}
	return nil
}

func GetProxyRequestContext(ctx context.Context) (req *events.APIGatewayProxyRequestContext, ok bool) {
	if o := GetRawRequestContext(ctx); o != nil {
		req, ok = o.(*events.APIGatewayProxyRequestContext)
	}
	return
}

func GetV2HTTPRequestContext(ctx context.Context) (req *events.APIGatewayV2HTTPRequestContext, ok bool) {
	if o := GetRawRequestContext(ctx); o != nil {
		req, ok = o.(*events.APIGatewayV2HTTPRequestContext)
	}
	return
}

func GetALBTargetGroupRequestContext(ctx context.Context) (req *events.ALBTargetGroupRequestContext, ok bool) {
	if o := GetRawRequestContext(ctx); o != nil {
		req, ok = o.(*events.ALBTargetGroupRequestContext)
	}
	return
}

func GetWebsocketRequestContext(ctx context.Context) (req *events.APIGatewayWebsocketProxyRequestContext, ok bool) {
	if o := GetRawRequestContext(ctx); o != nil {
		req, ok = o.(*events.APIGatewayWebsocketProxyRequestContext)
	}
	return
}
