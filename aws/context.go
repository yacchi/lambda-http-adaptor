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
	}
	return nil
}

func GetProxyRequestContext(ctx context.Context) *events.APIGatewayProxyRequestContext {
	if o := GetRawRequestContext(ctx); o == nil {
		return nil
	} else {
		return o.(*events.APIGatewayProxyRequestContext)
	}
}

func GetV2HTTPRequestContext(ctx context.Context) *events.APIGatewayV2HTTPRequestContext {
	if o := GetRawRequestContext(ctx); o == nil {
		return nil
	} else {
		return o.(*events.APIGatewayV2HTTPRequestContext)
	}
}

func GetALBTargetGroupRequestContext(ctx context.Context) *events.ALBTargetGroupRequestContext {
	if o := GetRawRequestContext(ctx); o == nil {
		return nil
	} else {
		return o.(*events.ALBTargetGroupRequestContext)
	}
}
