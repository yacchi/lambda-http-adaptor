module sample_app

go 1.17

replace github.com/yacchi/lambda-http-adaptor => ./../../

require (
	github.com/aws/aws-cdk-go/awscdk v1.122.0-devpreview
	github.com/aws/aws-cdk-go/awscdk/v2 v2.0.0-rc.21
	github.com/aws/constructs-go/constructs/v10 v10.0.5
	github.com/aws/constructs-go/constructs/v3 v3.3.147
	github.com/aws/jsii-runtime-go v1.34.0
	github.com/stretchr/testify v1.7.0
	github.com/tidwall/gjson v1.9.0
	github.com/yacchi/lambda-http-adaptor v0.0.0-00010101000000-000000000000
)

require (
	github.com/Masterminds/semver/v3 v3.1.1 // indirect
	github.com/PaesslerAG/gval v1.1.1 // indirect
	github.com/PaesslerAG/jsonpath v0.1.1 // indirect
	github.com/aws/aws-lambda-go v1.26.0 // indirect
	github.com/aws/aws-sdk-go v1.40.42 // indirect
	github.com/aws/aws-sdk-go-v2 v1.9.0 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.8.1 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.4.1 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.5.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.2.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi v1.4.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.3.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.4.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.7.0 // indirect
	github.com/aws/smithy-go v1.8.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/tidwall/match v1.0.3 // indirect
	github.com/tidwall/pretty v1.1.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)
