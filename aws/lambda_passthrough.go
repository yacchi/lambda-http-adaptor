/*
Package aws provides an implementation using aws-sdk-go.

Lambda event type compatibility layer for AWS API Gateway with HTTP API mode.

See lambda event detail:
https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api-develop-integrations-lambda.html
*/
package aws

import (
	"bytes"
	"context"
	"fmt"
	"github.com/yacchi/lambda-http-adaptor/types"
	"net/http"
	"strconv"
	"strings"
)

// NewLambdaPassthroughRequest Raw lambda event type to http.Request converter.
func NewLambdaPassthroughRequest(ctx context.Context, payload []byte, eventPath string, contentType string) (r *http.Request, err error) {
	var (
		body   *bytes.Buffer
		header = make(http.Header)
	)

	// build raw request URL
	rawURL := "http://localhost/" + strings.TrimLeft(eventPath, "/")
	body = bytes.NewBuffer(payload)

	r, err = http.NewRequestWithContext(ctx, "POST", rawURL, body)
	if err != nil {
		return nil, fmt.Errorf("lambda_passthrough: new request: %w", err)
	}

	r.Header = header
	r.Header.Set(types.HTTPHeaderContentType, contentType)
	r.Header.Set(types.HTTPHeaderContentLength, strconv.Itoa(body.Len()))
	r.RemoteAddr = "127.0.0.1"
	r.RequestURI = r.URL.RequestURI()
	r.Host = r.URL.Host

	return
}
