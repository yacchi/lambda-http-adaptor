/*
Lambda event type compatibility layer for AWS API Gateway with HTTP API mode.

See lambda event detail:
https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api-develop-integrations-lambda.html
*/
package aws

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/yacchi/lambda-http-adaptor/internal"
	"github.com/yacchi/lambda-http-adaptor/types"
	"github.com/yacchi/lambda-http-adaptor/utils"
	"net/http"
	"strconv"
)

// Lambda event type to http.Request converter for API Gateway with HTTP API mode.
func NewHTTPAPIRequest(ctx context.Context, e *events.APIGatewayV2HTTPRequest) (r *http.Request, err error) {
	var (
		body   *bytes.Buffer
		header = make(http.Header)
	)

	reqCtx := e.RequestContext

	for k, v := range e.Headers {
		header.Set(k, v)
	}

	if header.Get(types.HTTPHeaderCookie) == "" {
		header[types.HTTPHeaderCookie] = e.Cookies
	}

	// build raw request URL
	rawURL := "http://" + reqCtx.DomainName + e.RawPath

	if e.RawQueryString != "" {
		rawURL += "?" + e.RawQueryString
	}

	// build body reader
	if e.IsBase64Encoded {
		b, err := base64.StdEncoding.DecodeString(e.Body)
		if err != nil {
			return nil, fmt.Errorf("http_api: decode base64 body: %w", err)
		}
		body = bytes.NewBuffer(b)
	} else {
		body = bytes.NewBufferString(e.Body)
	}

	r, err = http.NewRequestWithContext(ctx, e.RequestContext.HTTP.Method, rawURL, body)
	if err != nil {
		return nil, fmt.Errorf("http_api: new request: %w", err)
	}

	r.Header = header

	if r.Header.Get(types.HTTPHeaderContentLength) == "" && body != nil {
		r.Header.Set(types.HTTPHeaderContentLength, strconv.Itoa(body.Len()))
	}

	header.Set(types.HTTPHeaderXForwardedFor, reqCtx.HTTP.SourceIP)
	// HTTP API only HTTPS
	header.Set(types.HTTPHeaderXForwardedPort, "443")
	header.Set(types.HTTPHeaderXForwardedProto, "https")

	r.RemoteAddr = reqCtx.HTTP.SourceIP

	r.RequestURI = r.URL.RequestURI()

	r = r.WithContext(internal.NewRawRequestValueContext(r.Context(), e))

	if r.Header.Get(types.HTTPHeaderXRayTraceIDKey) == "" {
		if traceID := ctx.Value(types.AWSXRayTraceIDContextKey); traceID != nil {
			r.Header.Set(types.HTTPHeaderXRayTraceIDKey, fmt.Sprintf("%v", traceID))
		}
	}

	return

}

// Response writer for API Gateway with HTTP API mode.
func HTTPAPIResponse(w *ResponseWriter) (r *events.APIGatewayV2HTTPResponse, err error) {
	r = &events.APIGatewayV2HTTPResponse{
		StatusCode:      w.status,
		IsBase64Encoded: utils.IsBinaryContent(w.Header()),
		Headers:         utils.SemicolonSeparatedHeaderMap(w.headers),
	}

	if r.IsBase64Encoded {
		r.Body = base64.StdEncoding.EncodeToString(w.buf.Bytes())
	} else {
		r.Body = w.buf.String()
	}

	w.Done()
	return
}
