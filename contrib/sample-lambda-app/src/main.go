package main

import (
	"context"
	"fmt"
	aws2 "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	adaptor "github.com/yacchi/lambda-http-adaptor"
	_ "github.com/yacchi/lambda-http-adaptor/all"
	"github.com/yacchi/lambda-http-adaptor/aws"
	"github.com/yacchi/lambda-http-adaptor/example/simple/api"
	"github.com/yacchi/lambda-http-adaptor/middlewares"
)

func main() {
	handler := api.ProvideAPI()
	h := middlewares.StripStageVar(handler.ServeHTTP)
	fmt.Println(adaptor.ListenAndServeWithOptions("", h,
		aws.WithAWSConfigProvider(func(ctx context.Context) (aws2.Config, error) {
			return config.LoadDefaultConfig(ctx)
		}),
	))
}
