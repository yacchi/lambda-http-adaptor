package main

import (
	adaptor "github.com/yacchi/lambda-http-adaptor"
	_ "github.com/yacchi/lambda-http-adaptor/all"
	api2 "github.com/yacchi/lambda-http-adaptor/example/simple/api"
	"github.com/yacchi/lambda-http-adaptor/middlewares"
	"log"
)

func main() {
	api := api2.ProvideAPI()
	log.Fatalln(adaptor.ListenAndServe("0.0.0.0:8888", middlewares.APIGatewayStripStageVar(api)))
}
