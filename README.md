# lambda-http-adaptor
[![Go Reference](https://pkg.go.dev/badge/github.com/yacchi/lambda-http-adaptor.svg)](https://pkg.go.dev/github.com/yacchi/lambda-http-adaptor) ![GitHub](https://img.shields.io/github/license/yacchi/lambda-http-adaptor)

lambda-http-adaptor is a compatible adaptor for Go `net/http` that can be used in multiple serverless environments.  

You can run your existing `http.HandlerFunc` compatible web application on AWS Lambda or Azure Functions.

## Example

lambda-http-adaptor provides the `adaptor.ListenAndServe` method, which can drop-in replacement for the `http.ListenAndServe`.

```go
package main

import (
	"github.com/yacchi/lambda-http-adaptor"
	_ "github.com/yacchi/lambda-http-adaptor/all"
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
```

## Pass through non-http events

Pass-through of non-HTTP Events has been added in v0.5.0.
The destination path is /events by default.
If you want to change the forwarding path or stop forwarding, please refer to the following sample for configuration.

```go
func main() {
  log.Fatalln(adaptor.ListenAndServeWithOptions(
    "",
    handler,
    // aws.WithNonHTTPEventPath("/api"),
	// aws.WithoutNonHTTPEventPassThrough(),
  )
}
```

## Features
- AWS Lambda support
  - [x] API Gateway REST API integration
    - [x] Multi-value headers
    - [x] Multi-value query string
    - [x] Get ProxyRequestContext value from Context 
  - [x] API Gateway HTTP API integration
    - [x] Multi-value headers
    - [x] Get V2 HTTPRequestContext value from Context 
  - [x] Application Load Balancer Lambda target
    - [x] Multi-value headers
    - [x] Get ALBTargetGroupRequestContext value from Context
  - [x] Get raw request value from Context
  - [x] Lambda container image function
  - [x] API Gateway Websocket API integration (Experimental)
  - [x] Non-HTTP event pass-through
- AWS API Gateway utilities
  - [x] Strip stage var middleware
  - [ ] Abstract interface of RequestContext  
- Azure Functions support
  - [ ] HTTP Trigger with custom handler
