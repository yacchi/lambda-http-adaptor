package main

import (
	"github.com/yacchi/lambda-http-adaptor"
	"log"
	"net/http"
)

func main() {
	log.Fatalln(adaptor.ListenAndServe("", http.HandlerFunc(echoReplyHandler)))
}

func echoReplyHandler(w http.ResponseWriter, r *http.Request) {
	m := r.URL.Query().Get("message")
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(m))
}
