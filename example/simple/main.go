package main

import (
	"fmt"
	adaptor "github.com/yacchi/lambda-http-adaptor"
	_ "github.com/yacchi/lambda-http-adaptor/all"
	"github.com/yacchi/lambda-http-adaptor/middlewares"
	"log"
	"net/http"
	"strings"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
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
			writer.Write([]byte(fmt.Sprintf("%s=%s", k, strings.Join(v, ", "))))
		}
	})

	log.Fatalln(adaptor.ListenAndServe("", middlewares.StripStageVar(mux.ServeHTTP)))
}
