/*
Package aws provides an implementation using aws-sdk-go.

Lambda event type compatibility layer for AWS API Gateway with REST API mode.

See lambda event detail:
https://docs.aws.amazon.com/apigateway/latest/developerguide/set-up-lambda-proxy-integrations.html#api-gateway-simple-proxy-for-lambda-input-format
*/
package aws

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/PaesslerAG/jsonpath"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
	"github.com/yacchi/lambda-http-adaptor/internal"
	"github.com/yacchi/lambda-http-adaptor/types"
	"github.com/yacchi/lambda-http-adaptor/utils"
	"net/http"
	"net/url"
	"path"
	"strconv"
)

var (
	DefaultWebsocketPathPrefix      = "websocket"
	DefaultRouteSelectionExpression = "$request.body.action"
)

// NewWebsocketRequest Lambda event type to http.Request converter for API Gateway with REST API mode.
func NewWebsocketRequest(ctx context.Context, e *events.APIGatewayWebsocketProxyRequest, pathPrefix string) (r *http.Request, multiValue bool, err error) {
	var (
		body   *bytes.Buffer
		header = make(http.Header)
		query  = make(url.Values)
	)

	if e.MultiValueHeaders != nil {
		multiValue = true
		header = e.MultiValueHeaders
	} else if e.Headers != nil {
		for k, v := range e.Headers {
			header.Set(k, v)
		}
	}

	// build raw request URL
	u, err := url.Parse(path.Join("/", e.RequestContext.Stage, pathPrefix, e.RequestContext.RouteKey))
	if err != nil {
		return nil, false, fmt.Errorf("rest_api: parsing path: %w", err)
	}

	u.Scheme = "https"
	u.Host = header.Get(types.HTTPHeaderHost)
	if u.Host == "" {
		u.Host = e.RequestContext.DomainName
	}

	if e.MultiValueQueryStringParameters != nil {
		query = e.MultiValueQueryStringParameters
	} else if e.QueryStringParameters != nil {
		for k, v := range e.QueryStringParameters {
			query.Set(k, v)
		}
	}

	if 0 < len(query) {
		u.RawQuery = query.Encode()
	}

	// build body reader
	if e.IsBase64Encoded {
		b, err := base64.StdEncoding.DecodeString(e.Body)
		if err != nil {
			return nil, false, fmt.Errorf("rest_api: decode base64 body: %w", err)
		}
		body = bytes.NewBuffer(b)
	} else {
		body = bytes.NewBufferString(e.Body)
	}

	method := "POST"

	r, err = http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, false, fmt.Errorf("alb_target: new request: %w", err)
	}

	r.Header = header

	if r.Header.Get(types.HTTPHeaderContentLength) == "" && body != nil {
		r.Header.Set(types.HTTPHeaderContentLength, strconv.Itoa(body.Len()))
	}

	r.RemoteAddr = e.RequestContext.Identity.SourceIP

	r.RequestURI = r.URL.RequestURI()

	if r.Header.Get(types.HTTPHeaderXRayTraceIDKey) == "" {
		if traceID := ctx.Value(types.AWSXRayTraceIDContextKey); traceID != nil {
			r.Header.Set(types.HTTPHeaderXRayTraceIDKey, fmt.Sprintf("%v", traceID))
		}
	}

	r = r.WithContext(internal.NewRawRequestValueContext(r.Context(), e))

	return
}

type APIGatewayManagementAPI interface {
	PostToConnection(ctx context.Context, connectionID string, data []byte) (err error)
}

type v1api struct {
	*apigatewaymanagementapi.ApiGatewayManagementApi
}

func (v *v1api) PostToConnection(ctx context.Context, connectionID string, data []byte) (err error) {
	_, err = v.ApiGatewayManagementApi.PostToConnectionWithContext(ctx, &apigatewaymanagementapi.PostToConnectionInput{
		ConnectionId: aws.String(connectionID),
		Data:         data,
	})
	return
}

// NewAPIGatewayManagementClientV1 creates a new API Gateway Management Client instance from the provided parameters. The
// new client will have a custom endpoint that resolves to the application's deployed API.
func NewAPIGatewayManagementClientV1(sess *session.Session, domain, stage string) APIGatewayManagementAPI {
	conf := aws.NewConfig()
	conf.WithEndpointResolver(endpoints.ResolverFunc(func(service, region string, opts ...func(*endpoints.Options)) (endpoints.ResolvedEndpoint, error) {
		if service != apigatewaymanagementapi.EndpointsID {
			return endpoints.ResolvedEndpoint{}, &endpoints.EndpointNotFoundError{}
		} else {
			var endpoint url.URL
			endpoint.Path = stage
			endpoint.Host = domain
			endpoint.Scheme = "https"
			return endpoints.ResolvedEndpoint{
				SigningRegion: region,
				URL:           endpoint.String(),
			}, nil
		}
	}))
	return &v1api{apigatewaymanagementapi.New(sess, conf)}
}

// WebsocketResponse Response writer for API Gateway with REST API mode.
func WebsocketResponse(w *WebsocketResponseWriter, multiValue bool) (r *events.APIGatewayProxyResponse, err error) {
	r = &events.APIGatewayProxyResponse{
		StatusCode:      w.status,
		IsBase64Encoded: utils.IsBinaryContent(w.Header()),
	}

	if multiValue {
		r.MultiValueHeaders = w.Header()
	} else {
		r.Headers = utils.SemicolonSeparatedHeaderMap(w.Header())
	}

	w.Done()
	return
}

func ExpandJSONPath(s string, obj interface{}) string {
	var buf []byte
	// ${} is all ASCII, so bytes are fine for this operation.
	i := 0
	for j := 0; j < len(s); j++ {
		if s[j] == '$' && j+1 < len(s) {
			if buf == nil {
				buf = make([]byte, 0, 2*len(s))
			}
			buf = append(buf, s[i:j]...)
			expression, w := expandBrace(s[j+1:])
			if expression == "" && w > 0 {
				// Encountered invalid syntax; eat the
				// characters.
			} else if expression == "" {
				// Valid syntax, but $ was not followed by a
				// expression. Leave the dollar character untouched.
				buf = append(buf, s[j])
			} else {
				if extend, err := jsonpath.Get(expression, obj); err != nil {
					buf = append(buf, ""...)
				} else if chars, ok := extend.(string); ok {
					buf = append(buf, chars...)
				} else {
					buf = append(buf, ""...)
				}
			}
			j += w
			i = j + 1
		}
	}
	if buf == nil {
		return s
	}
	return string(buf) + s[i:]
}

func expandBrace(s string) (string, int) {
	if s[0] == '{' {
		for i := 1; i < len(s); i++ {
			if s[i] == '}' {
				if i == 1 {
					return "", 2 // Bad syntax; eat "${}"
				}
				return s[1:i], i + 1
			}
		}
		return "", 1 // Bad syntax; eat "${"
	}
	return s, len(s)
}

// RouteSelector
// https://docs.aws.amazon.com/apigateway/latest/developerguide/websocket-api-develop-routes.html#apigateway-websocket-api-route-selection-expressions
func RouteSelector(req *events.APIGatewayWebsocketProxyRequest, expression string) (route string, err error) {
	var (
		rawReq  map[string]interface{}
		rawBody interface{}
	)

	if b, err := json.Marshal(req); err != nil {
		return "", err
	} else if err := json.Unmarshal(b, &rawReq); err != nil {
		return "", err
	}

	if err := json.Unmarshal([]byte(req.Body), &rawBody); err != nil {
		return "", err
	}

	rawReq["body"] = rawBody

	raw := map[string]interface{}{
		"request": rawReq,
	}

	route = ExpandJSONPath(expression, raw)
	return
}
