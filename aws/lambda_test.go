package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"github.com/yacchi/lambda-http-adaptor/types"
	"github.com/yacchi/lambda-http-adaptor/utils"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type httpRequest struct {
	Method string
	URL    string
	Header http.Header
	Body   []byte
}

type httpResponse struct {
	Status int
	Header http.Header
	Body   []byte
}

type httpTest struct {
	Name     string
	Request  httpRequest
	Response httpResponse
	Handler  http.HandlerFunc
	Compare  func(t *testing.T)
}

const (
	textCtx    = "Testing"
	requestCtx = "Request"
)

func NewTestContext(t *testing.T, ht httpTest) context.Context {
	ctx := context.WithValue(context.Background(), textCtx, t)
	ctx = context.WithValue(ctx, requestCtx, &ht.Request)
	return ctx
}

func GetTestContext(ctx context.Context) (*testing.T, *httpRequest) {
	t := ctx.Value(textCtx).(*testing.T)
	req := ctx.Value(requestCtx).(*httpRequest)
	return t, req
}

var httpGetTests = []httpTest{
	{
		Name: "echo",
		Request: httpRequest{
			Method: http.MethodGet,
			URL:    "http://localhost/echo?message=hello",
		},
		Response: httpResponse{
			Status: http.StatusOK,
			Header: map[string][]string{
				"Content-Type": {"text/plain"},
			},
			Body: []byte("hello"),
		},
		Handler: func(writer http.ResponseWriter, request *http.Request) {
			t, req := GetTestContext(request.Context())

			m := request.URL.Query().Get("message")
			assert.Equal(t, request.Method, req.Method)
			assert.Equal(t, "hello", m)
			assert.Equal(t, request.URL.Path, "/echo")

			writer.Header().Set("content-type", "text/plain")
			_, err := writer.Write([]byte(m))
			assert.NoError(t, err)
		},
	},
	{
		Name: "echo-encoded",
		Request: httpRequest{
			Method: http.MethodGet,
			URL:    "http://localhost/encoded?message=hello%20world",
		},
		Response: httpResponse{
			Status: http.StatusOK,
			Header: map[string][]string{
				"Content-Type": {"text/plain"},
			},
			Body: []byte("hello world"),
		},
		Handler: func(writer http.ResponseWriter, request *http.Request) {
			t, req := GetTestContext(request.Context())

			m := request.URL.Query().Get("message")
			assert.Equal(t, request.Method, req.Method)
			assert.Equal(t, "hello world", m)
			assert.Equal(t, request.URL.Path, "/encoded")

			writer.Header().Set("content-type", "text/plain")
			_, err := writer.Write([]byte(m))
			assert.NoError(t, err)
		},
	},
}

var httpPostTests = []httpTest{
	{
		Name: "echo",
		Request: httpRequest{
			Method: http.MethodPost,
			URL:    "http://localhost/echo",
			Body:   []byte(`{"message": "hello"}`),
			Header: map[string][]string{
				"Content-Type": {"application/json"},
			},
		},
		Response: httpResponse{
			Status: http.StatusOK,
			Header: map[string][]string{
				"Content-Type": {"application/json"},
			},
			Body: []byte(`{"echo":"hello"}`),
		},
		Handler: func(writer http.ResponseWriter, request *http.Request) {
			t, req := GetTestContext(request.Context())

			assert.Equal(t, request.Method, req.Method)
			assert.Equal(t, request.URL.Path, "/echo")

			ct := request.Header.Get(types.HTTPHeaderContentType)
			assert.Equal(t, ct, req.Header.Get(types.HTTPHeaderContentType))

			body, err := ioutil.ReadAll(request.Body)
			assert.NoError(t, err)

			var data map[string]string
			err = json.Unmarshal(body, &data)
			assert.NoError(t, err)

			m, ok := data["message"]
			assert.True(t, ok)

			writer.Header().Set("content-type", "application/json")

			res := map[string]string{"echo": m}
			resBytes, err := json.Marshal(res)
			assert.NoError(t, err)

			_, err = writer.Write(resBytes)
			assert.NoError(t, err)
		},
	},
}

func TestLambdaHTTPModeGet(t *testing.T) {
	for _, test := range httpGetTests {
		t.Run(test.Name, func(t *testing.T) {
			h := NewLambdaHandler(test.Handler)

			u, _ := url.Parse(test.Request.URL)

			event := events.APIGatewayV2HTTPRequest{
				Version:         "2.0",
				RawPath:         u.Path + "?" + u.RawQuery,
				Cookies:         nil,
				Headers:         nil,
				PathParameters:  nil,
				StageVariables:  nil,
				Body:            string(test.Request.Body),
				IsBase64Encoded: false,
			}
			event.RequestContext.HTTP.Method = test.Request.Method

			b, err := json.Marshal(event)
			assert.NoError(t, err)

			ret, err := h.Invoke(NewTestContext(t, test), b)
			assert.NoError(t, err)

			var res events.APIGatewayV2HTTPResponse
			err = json.Unmarshal(ret, &res)
			assert.NoError(t, err)

			assert.Equal(t, test.Response.Header.Get(types.HTTPHeaderContentType), res.Headers[types.HTTPHeaderContentType])
			assert.Equal(t, string(test.Response.Body), res.Body)
		})
	}
}

func TestLambdaHTTPModePost(t *testing.T) {
	for _, test := range httpPostTests {
		t.Run(test.Name, func(t *testing.T) {
			h := NewLambdaHandler(test.Handler)

			u, _ := url.Parse(test.Request.URL)

			event := events.APIGatewayV2HTTPRequest{
				Version:         "2.0",
				RawPath:         u.Path,
				Headers:         utils.SemicolonSeparatedHeaderMap(test.Request.Header),
				RequestContext:  events.APIGatewayV2HTTPRequestContext{},
				Body:            string(test.Request.Body),
				IsBase64Encoded: false,
			}
			event.RequestContext.HTTP.Method = test.Request.Method

			b, err := json.Marshal(event)
			assert.NoError(t, err)

			ret, err := h.Invoke(NewTestContext(t, test), b)
			assert.NoError(t, err)

			var res events.APIGatewayV2HTTPResponse
			err = json.Unmarshal(ret, &res)
			assert.NoError(t, err)

			assert.Equal(t, test.Response.Header.Get(types.HTTPHeaderContentType), res.Headers[types.HTTPHeaderContentType])
			assert.Equal(t, string(test.Response.Body), res.Body)
		})
	}
}

func TestALBTargetGroupModeGet(t *testing.T) {
	for _, test := range httpGetTests {
		t.Run(test.Name, func(t *testing.T) {
			h := NewLambdaHandler(test.Handler)

			u, _ := url.Parse(test.Request.URL)

			event := events.ALBTargetGroupRequest{
				HTTPMethod:            test.Request.Method,
				Path:                  u.Path,
				QueryStringParameters: map[string]string{},
				Headers:               nil,
				Body:                  string(test.Request.Body),
				IsBase64Encoded:       false,
			}

			for k, v := range u.Query() {
				event.QueryStringParameters[k] = v[0]
			}

			b, err := json.Marshal(event)
			assert.NoError(t, err)

			ret, err := h.Invoke(NewTestContext(t, test), b)
			assert.NoError(t, err)

			var res events.ALBTargetGroupResponse
			err = json.Unmarshal(ret, &res)
			assert.NoError(t, err)

			assert.Equal(t, test.Response.Header.Get(types.HTTPHeaderContentType), res.Headers[types.HTTPHeaderContentType])
			assert.Equal(t, string(test.Response.Body), res.Body)
		})
	}
}

func TestALBTargetGroupModePost(t *testing.T) {
	for _, test := range httpPostTests {
		t.Run(test.Name, func(t *testing.T) {
			h := NewLambdaHandler(test.Handler)

			u, _ := url.Parse(test.Request.URL)

			event := events.ALBTargetGroupRequest{
				HTTPMethod:            test.Request.Method,
				Path:                  u.Path,
				QueryStringParameters: map[string]string{},
				Headers:               utils.SemicolonSeparatedHeaderMap(test.Request.Header),
				Body:                  string(test.Request.Body),
				IsBase64Encoded:       false,
			}

			for k, v := range u.Query() {
				event.QueryStringParameters[k] = v[0]
			}

			b, err := json.Marshal(event)
			assert.NoError(t, err)

			ret, err := h.Invoke(NewTestContext(t, test), b)
			assert.NoError(t, err)

			var res events.ALBTargetGroupResponse
			err = json.Unmarshal(ret, &res)
			assert.NoError(t, err)

			assert.Equal(t, test.Response.Header.Get(types.HTTPHeaderContentType), res.Headers[types.HTTPHeaderContentType])
			assert.Equal(t, string(test.Response.Body), res.Body)
		})
	}
}

func TestRESTAPIModeGet(t *testing.T) {
	for _, test := range httpGetTests {
		t.Run(test.Name, func(t *testing.T) {
			h := NewLambdaHandler(test.Handler)

			u, _ := url.Parse(test.Request.URL)

			event := events.APIGatewayProxyRequest{
				HTTPMethod:            test.Request.Method,
				Path:                  u.Path,
				QueryStringParameters: map[string]string{},
				Headers:               nil,
				Body:                  string(test.Request.Body),
				IsBase64Encoded:       false,
			}

			for k, v := range u.Query() {
				event.QueryStringParameters[k] = v[0]
			}

			b, err := json.Marshal(event)
			assert.NoError(t, err)

			ret, err := h.Invoke(NewTestContext(t, test), b)
			assert.NoError(t, err)

			var res events.ALBTargetGroupResponse
			err = json.Unmarshal(ret, &res)
			assert.NoError(t, err)

			assert.Equal(t, test.Response.Header.Get(types.HTTPHeaderContentType), res.Headers[types.HTTPHeaderContentType])
			assert.Equal(t, string(test.Response.Body), res.Body)
		})
	}
}

func TestRESTAPIModePost(t *testing.T) {
	for _, test := range httpGetTests {
		t.Run(test.Name, func(t *testing.T) {
			h := NewLambdaHandler(test.Handler)

			u, _ := url.Parse(test.Request.URL)

			event := events.APIGatewayProxyRequest{
				HTTPMethod:            test.Request.Method,
				Path:                  u.Path,
				QueryStringParameters: map[string]string{},
				Headers:               utils.SemicolonSeparatedHeaderMap(test.Request.Header),
				Body:                  string(test.Request.Body),
				IsBase64Encoded:       false,
			}

			for k, v := range u.Query() {
				event.QueryStringParameters[k] = v[0]
			}

			b, err := json.Marshal(event)
			assert.NoError(t, err)

			ret, err := h.Invoke(NewTestContext(t, test), b)
			assert.NoError(t, err)

			var res events.ALBTargetGroupResponse
			err = json.Unmarshal(ret, &res)
			assert.NoError(t, err)

			assert.Equal(t, test.Response.Header.Get(types.HTTPHeaderContentType), res.Headers[types.HTTPHeaderContentType])
			assert.Equal(t, string(test.Response.Body), res.Body)
		})
	}
}

func MockAPI() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/production/$connect", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("content-type", "application/json")
		writer.WriteHeader(http.StatusOK)
		err := json.NewEncoder(writer).Encode(utils.SemicolonSeparatedHeaderMap(request.Header))
		if err != nil {
			fmt.Println(err)
		}
	})

	mux.HandleFunc("/production/$default", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		io.Copy(writer, request.Body)
	})

	mux.HandleFunc("/production/$disconnect", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("OK " + request.URL.String()))
	})

	mux.HandleFunc("/ping", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("pong"))
	})

	mux.HandleFunc("/echo", func(writer http.ResponseWriter, request *http.Request) {
		m := request.URL.Query().Get("message")
		writer.Header().Set("Content-Type", "text/plain")
		writer.Write([]byte(m))
	})

	mux.HandleFunc("/headers", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		for k, v := range request.Header {
			writer.Write([]byte(fmt.Sprintf("%s=%s\n", k, strings.Join(v, ", "))))
		}
	})

	mux.HandleFunc("/request_context", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		json.NewEncoder(writer).Encode(GetRawRequestContext(request.Context()))
	})

	//return middlewares.StripStageVar(mux.ServeHTTP)
	return mux
}

func TestLambdaHandler_Invoke(t *testing.T) {
	h := NewLambdaHandler(MockAPI())

	cases := []struct {
		Request string
	}{
		{
			Request: filepath.Join("testdata", "websocket.connect.json"),
		},
		//{
		//	Request: filepath.Join("testdata", "websocket.default.json"),
		//},
	}

	for _, c := range cases {
		b, err := os.ReadFile(c.Request)
		if err != nil {
			log.Fatalln(err)
		}

		ctx := context.Background()
		res, err := h.Invoke(ctx, b)
		if err != nil {
			log.Fatalln(err)
		} else {
			_ = res
		}
	}
}
