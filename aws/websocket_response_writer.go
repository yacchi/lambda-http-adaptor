package aws

import (
	"bytes"
	"context"
	"github.com/aws/aws-lambda-go/events"
	"net/http"
)

type WebsocketResponseWriter struct {
	ctx         context.Context
	client      APIGatewayManagementAPI
	req         *events.APIGatewayWebsocketProxyRequest
	status      int
	headers     http.Header
	wroteHeader bool
	closeCh     chan bool
	buf         *bytes.Buffer
}

func NewWebsocketResponseWriter(ctx context.Context, client APIGatewayManagementAPI, request *events.APIGatewayWebsocketProxyRequest) *WebsocketResponseWriter {
	return &WebsocketResponseWriter{
		ctx:     ctx,
		client:  client,
		req:     request,
		headers: map[string][]string{},
		closeCh: make(chan bool, 1),
		buf:     bytes.NewBuffer(nil),
	}
}

func (w *WebsocketResponseWriter) Header() http.Header {
	return w.headers
}

func (w *WebsocketResponseWriter) Write(i []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	w.buf.Write(i)

	err := w.client.PostToConnection(w.ctx, w.req.RequestContext.ConnectionID, i)
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
