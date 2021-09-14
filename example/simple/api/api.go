package api

import (
	"encoding/json"
	"fmt"
	"github.com/yacchi/lambda-http-adaptor/aws"
	"github.com/yacchi/lambda-http-adaptor/utils"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type writerWrapper struct {
	http.ResponseWriter
	code int
}

type logEntry struct {
	Time         time.Time `json:"time"`
	Method       string    `json:"method"`
	Server       string    `json:"server"`
	Path         string    `json:"path"`
	Code         int       `json:"code"`
	Duration     string    `json:"duration"`
	ConnectionID string    `json:"connection_id,omitempty"`
}

func (w *writerWrapper) WriteHeader(code int) {
	w.code = code
	w.ResponseWriter.WriteHeader(code)
}

func ProvideAPI() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/websocket/$connect", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		err := json.NewEncoder(writer).Encode(map[string]interface{}{
			"message": "hello",
			"headers": utils.SemicolonSeparatedHeaderMap(request.Header),
		})
		if err != nil {
			fmt.Println(err)
		}
	})

	mux.HandleFunc("/websocket/$default", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		_, err := io.Copy(writer, request.Body)
		if err != nil {
			fmt.Println(err)
		}
	})

	mux.HandleFunc("/websocket/$disconnect", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		err := json.NewEncoder(writer).Encode(map[string]string{
			"message": "bye",
		})
		if err != nil {
			fmt.Println(err)
		}
	})

	mux.HandleFunc("/ping", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		_, err := writer.Write([]byte("pong"))
		if err != nil {
			fmt.Println(err)
		}
	})

	mux.HandleFunc("/echo", func(writer http.ResponseWriter, request *http.Request) {
		m := request.URL.Query().Get("message")
		writer.Header().Set("Content-Type", "text/plain")
		_, err := writer.Write([]byte(m))
		if err != nil {
			fmt.Println(err)
		}
	})

	mux.HandleFunc("/headers", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		for k, v := range request.Header {
			_, err := writer.Write([]byte(fmt.Sprintf("%s=%s\n", k, strings.Join(v, ", "))))
			if err != nil {
				fmt.Println(err)
			}
		}
	})

	mux.HandleFunc("/request_context", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		err := json.NewEncoder(writer).Encode(aws.GetRawRequestContext(request.Context()))
		if err != nil {
			fmt.Println(err)
		}
	})

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		start := time.Now()
		rWriter := &writerWrapper{ResponseWriter: writer}

		raw, ok := utils.RawRequestValue(request.Context())
		if ok {
			rawReq, _ := json.Marshal(raw)
			fmt.Println(string(rawReq))
		}
		bReq, _ := json.Marshal(request)
		fmt.Println(string(bReq))

		mux.ServeHTTP(rWriter, request)
		duration := time.Now().Sub(start)

		entry := logEntry{
			Time:     start,
			Method:   request.Method,
			Server:   request.URL.Host,
			Path:     request.URL.Path,
			Code:     rWriter.code,
			Duration: duration.String(),
		}

		if rawCtx, ok := aws.GetWebsocketRequestContext(request.Context()); ok {
			entry.ConnectionID = rawCtx.ConnectionID
		}

		b, err := json.Marshal(entry)
		if err != nil {
			log.Printf("log entry marshal error: %s\n", err)
		} else {
			fmt.Println(string(b))
		}
	})
}
