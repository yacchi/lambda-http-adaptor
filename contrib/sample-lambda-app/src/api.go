package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/yacchi/lambda-http-adaptor/aws"
	"github.com/yacchi/lambda-http-adaptor/utils"
	"io"
	"log"
	"net/http"
	"os"
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

	mux.HandleFunc("/events", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		body, _ := io.ReadAll(request.Body)
		println(request.Header.Get("Content-Type"))
		println(string(body))
		_, err := writer.Write(body)
		if err != nil {
			fmt.Println(err)
		}
	})

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
		body, _ := io.ReadAll(request.Body)
		if _, err := writer.Write(body); err != nil {
			fmt.Println(err)
		}
		if _, err := writer.Write(body); err != nil {
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
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)
		_, err := writer.Write([]byte(`{"message": "pong"}`))
		if err != nil {
			fmt.Println(err)
		}
	})

	mux.HandleFunc("/echo", func(writer http.ResponseWriter, request *http.Request) {
		ct := request.Header.Get("Content-Type")
		if ct == "" {
			ct = "text/plain"
		}
		writer.Header().Set("Content-Type", ct)
		writer.WriteHeader(http.StatusOK)
		if request.Method == http.MethodGet {
			m := request.URL.Query().Get("message")
			_, err := writer.Write([]byte(m))
			if err != nil {
				fmt.Println(err)
			}
		}
		if request.Method == http.MethodPost {
			_, err := io.Copy(writer, request.Body)
			if err != nil {
				fmt.Println(err)
			}
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

	mux.HandleFunc("/envs", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		envs := map[string]string{}
		for _, e := range os.Environ() {
			pair := strings.SplitN(e, "=", 2)
			envs[pair[0]] = pair[1]
		}
		lc, _ := lambdacontext.FromContext(request.Context())
		json.NewEncoder(os.Stdout).Encode(map[string]any{
			"envs":    envs,
			"context": lc,
		})
	})

	mux.HandleFunc("/request_context", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
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
