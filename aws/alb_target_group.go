/*
Package aws provides an implementation using aws-sdk-go.

Lambda event type compatibility layer for AWS Application Load Balancer target group mode.

See lambda event detail:
https://docs.aws.amazon.com/elasticloadbalancing/latest/application/lambda-functions.html#receive-event-from-load-balancer
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
	"net/url"
	"strconv"
)

// NewALBTargetGroupRequest Lambda event type to http.Request converter for Application Load Balancer target group mode.
func NewALBTargetGroupRequest(ctx context.Context, e *events.ALBTargetGroupRequest) (r *http.Request, multiValue bool, err error) {
	var (
		body   *bytes.Buffer
		header = make(http.Header)
	)

	if e.MultiValueHeaders != nil {
		multiValue = true
		for key, values := range e.MultiValueHeaders {
			for _, value := range values {
				header.Add(key, value)
			}
		}
	} else if e.Headers != nil {
		for k, v := range e.Headers {
			header.Set(k, v)
		}
	}

	// build raw request URL
	u, err := url.Parse(e.Path)
	if err != nil {
		return nil, false, fmt.Errorf("alb_target: parsing path: %w", err)
	}

	u.Scheme = "http"
	u.Host = header.Get(types.HTTPHeaderHost)

	if e.MultiValueQueryStringParameters != nil {
		u.RawQuery = utils.JoinMultiValueQueryParameters(e.MultiValueQueryStringParameters)
	} else if e.QueryStringParameters != nil {
		u.RawQuery = utils.JoinQueryParameters(e.QueryStringParameters)
	}

	// build body reader
	if e.IsBase64Encoded {
		b, err := base64.StdEncoding.DecodeString(e.Body)
		if err != nil {
			return nil, false, fmt.Errorf("alb_target: decode base64 body: %w", err)
		}
		body = bytes.NewBuffer(b)
	} else {
		body = bytes.NewBufferString(e.Body)
	}

	r, err = http.NewRequestWithContext(ctx, e.HTTPMethod, u.String(), body)
	if err != nil {
		return nil, false, fmt.Errorf("alb_target: new request: %w", err)
	}

	r.Header = header

	if r.Header.Get(types.HTTPHeaderContentLength) == "" && body != nil {
		r.Header.Set(types.HTTPHeaderContentLength, strconv.Itoa(body.Len()))
	}

	r.RemoteAddr = r.Header.Get(types.HTTPHeaderXForwardedFor)

	r.RequestURI = r.URL.RequestURI()

	r = r.WithContext(internal.NewRawRequestValueContext(r.Context(), e))

	return
}

// ALBTargetResponse Response writer for Application Load Balancer target group mode.
func ALBTargetResponse(w *ResponseWriter, multiValue bool) (r *events.ALBTargetGroupResponse, err error) {
	r = &events.ALBTargetGroupResponse{
		StatusCode:        w.status,
		StatusDescription: strconv.Itoa(w.status) + " " + http.StatusText(w.status),
		IsBase64Encoded:   utils.IsBinaryContent(w.Header()),
	}

	if multiValue {
		r.MultiValueHeaders = w.Header()
	} else {
		r.Headers = utils.SemicolonSeparatedHeaderMap(w.Header())
	}

	if r.IsBase64Encoded {
		r.Body = base64.StdEncoding.EncodeToString(w.buf.Bytes())
	} else {
		r.Body = w.buf.String()
	}

	w.Done()
	return
}
