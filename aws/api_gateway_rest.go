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

// https://docs.aws.amazon.com/apigateway/latest/developerguide/set-up-lambda-proxy-integrations.html#api-gateway-simple-proxy-for-lambda-input-format

func NewRESTAPIRequest(ctx context.Context, e *events.APIGatewayProxyRequest) (r *http.Request, multiValue bool, err error) {
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
	u, err := url.Parse(e.Path)
	if err != nil {
		return nil, false, fmt.Errorf("rest_api: parsing path: %w", err)
	}

	u.Scheme = "http"
	u.Host = header.Get(types.HTTPHeaderHost)

	if e.MultiValueQueryStringParameters != nil {
		query = e.MultiValueQueryStringParameters
	} else if e.QueryStringParameters != nil {
		for k, v := range e.QueryStringParameters {
			query.Set(k, v)
		}
	}

	u.RawQuery = query.Encode()

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

	r, err = http.NewRequestWithContext(ctx, e.HTTPMethod, u.String(), body)
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

func RESTAPITargetResponse(w *ResponseWriter, multiValue bool) (r *events.APIGatewayProxyResponse, err error) {
	r = &events.APIGatewayProxyResponse{
		StatusCode:      w.status,
		IsBase64Encoded: utils.IsBinaryContent(w.Header()),
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
