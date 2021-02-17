# lambda-http-adaptor

[![GoDoc Reference](https://pkg.go.dev/github.com/yacchi/lambda-http-adaptor?status.svg)](https://pkg.go.dev/github.com/yacchi/lambda-http-adaptor)  
![GitHub](https://img.shields.io/github/license/yacchi/lambda-http-adaptor)

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

## Features
- AWS Lambda support
  - [x] API Gateway REST API integration
    - [x] Multi-value headers
    - [x] Multi-value query string
    - [ ] Get ProxyRequestContext value from Context 
  - [x] API Gateway HTTP API integration
    - [x] Multi-value headers
    - [ ] Get V2 HTTPRequestContext value from Context 
  - [x] Application Load Balancer Lambda target
    - [x] Multi-value headers
    - [ ] Get ALBTargetGroupRequestContext value from Context
  - [x] Get raw request value from Context
  - [ ] Abstract interface of RequestContext for API-GW REST and HTTP modes  
  - [x] Lambda container image function
- Azure Functions support
  - [ ] HTTP Trigger with custom handler
