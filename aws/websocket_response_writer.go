package aws

import (
	"bytes"
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
	"net/http"
)

type WebsocketResponseWriter struct {
	ctx         context.Context
	client      *apigatewaymanagementapi.ApiGatewayManagementApi
	req         *events.APIGatewayWebsocketProxyRequest
	status      int
	headers     http.Header
	buf         bytes.Buffer
	wroteHeader bool
	closeCh     chan bool
}

func NewWebsocketResponseWriter(ctx context.Context, client *apigatewaymanagementapi.ApiGatewayManagementApi, request *events.APIGatewayWebsocketProxyRequest) *WebsocketResponseWriter {
	return &WebsocketResponseWriter{
		ctx:     ctx,
		client:  client,
		req:     request,
		headers: map[string][]string{},
		closeCh: make(chan bool, 1),
	}
}

func (w *WebsocketResponseWriter) Header() http.Header {
	return w.headers
}

func (w *WebsocketResponseWriter) Write(i []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	_, err := w.client.PostToConnectionWithContext(w.ctx, &apigatewaymanagementapi.PostToConnectionInput{
		ConnectionId: aws.String(w.req.RequestContext.ConnectionID),
		Data:         i,
	})
	if err != nil {
		return 0, err
	}
	return len(i), nil
}

func (w *WebsocketResponseWriter) WriteHeader(statusCode int) {
	if w.wroteHeader {
		return
	}

	w.status = statusCode

	if w.headers.Get("Content-Type") == "" {
		w.headers.Set("Content-Type", "text/plain; charset=utf8")
	}

	w.wroteHeader = true
}

func (w *WebsocketResponseWriter) CloseNotify() <-chan bool {
	return w.closeCh
}

func (w *WebsocketResponseWriter) Done() {
	w.closeCh <- true
}
