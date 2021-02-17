package main

import (
	"fmt"
	adaptor "github.com/yacchi/lambda-http-adaptor"
	_ "github.com/yacchi/lambda-http-adaptor/all"
	api2 "github.com/yacchi/lambda-http-adaptor/example/simple/api"
	"github.com/yacchi/lambda-http-adaptor/middlewares"
)

func main() {
	handler := api2.ProvideAPI()
	fmt.Println(adaptor.ListenAndServe("", middlewares.StripStageVar(handler.ServeHTTP)))
}
