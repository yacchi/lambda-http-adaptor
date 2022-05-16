module sample_app

go 1.17

replace github.com/yacchi/lambda-http-adaptor => ./../../

require (
	github.com/aws/aws-sdk-go-v2 v1.9.0
	github.com/aws/aws-sdk-go-v2/config v1.8.1
	github.com/yacchi/lambda-http-adaptor v0.0.0-00010101000000-000000000000
)

require (
	github.com/PaesslerAG/gval v1.1.1 // indirect
	github.com/PaesslerAG/jsonpath v0.1.1 // indirect
	github.com/aws/aws-lambda-go v1.26.0 // indirect
	github.com/aws/aws-sdk-go v1.40.42 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.4.1 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.5.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.2.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi v1.4.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.3.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.4.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.7.0 // indirect
	github.com/aws/smithy-go v1.8.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
)
