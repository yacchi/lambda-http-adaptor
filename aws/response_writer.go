package aws

import (
	"bytes"
	"net/http"
)

type ResponseWriter struct {
	status      int
	headers     http.Header
	buf         bytes.Buffer
	wroteHeader bool
	closeCh     chan bool
}

func NewResponseWriter() *ResponseWriter {
	return &ResponseWriter{
		headers: map[string][]string{},
		closeCh: make(chan bool, 1),
	}
}

func (r *ResponseWriter) Header() http.Header {
	return r.headers
}

func (r *ResponseWriter) Write(i []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	return r.buf.Write(i)
}

func (r *ResponseWriter) WriteHeader(statusCode int) {
	if r.wroteHeader {
		return
	}

	r.status = statusCode

	if r.headers.Get("Content-Type") == "" {
		r.headers.Set("Content-Type", "text/plain; charset=utf8")
	}

	r.wroteHeader = true
}

func (r *ResponseWriter) CloseNotify() <-chan bool {
	return r.closeCh
}

func (r *ResponseWriter) Done() {
	r.closeCh <- true
}
