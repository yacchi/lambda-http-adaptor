import * as core from "aws-cdk-lib/core"
import {Size} from "aws-cdk-lib/core"
import * as iam from "aws-cdk-lib/aws-iam"
import {Effect} from "aws-cdk-lib/aws-iam"
import * as lambda from "aws-cdk-lib/aws-lambda"
import * as apiGWv1 from "aws-cdk-lib/aws-apigateway"
import {AuthorizationType} from "aws-cdk-lib/aws-apigateway"
import * as apiGWv2 from "aws-cdk-lib/aws-apigatewayv2"
import {HttpLambdaIntegration, WebSocketLambdaIntegration} from "aws-cdk-lib/aws-apigatewayv2-integrations"
import {LogGroup, RetentionDays} from "aws-cdk-lib/aws-logs"

const app = new core.App()

const env = {
    region: app.node.tryGetContext('region') || process.env['CDK_DEFAULT_REGION'] || process.env['AWS_DEFAULT_REGION'],
    account: app.node.tryGetContext('account') || process.env['CDK_DEFAULT_ACCOUNT'] || process.env['AWS_ACCOUNT'],
}

const prefix = "SampleApp"

const stack = new core.Stack(app, "SampleLambdaApp", {env})

const handler = new lambda.DockerImageFunction(stack, `${prefix}Container`, {
    code: lambda.DockerImageCode.fromImageAsset("src"),
    memorySize: 128,
    timeout: core.Duration.minutes(1),
    architecture: lambda.Architecture.X86_64,
    logGroup: new LogGroup(stack, `${prefix}LogGroup`, {
        logGroupName: `/aws/lambda/${prefix}Container`,
        retention: RetentionDays.THREE_MONTHS,
        removalPolicy: core.RemovalPolicy.DESTROY,
    })
})

handler.addToRolePolicy(new iam.PolicyStatement({
    effect: Effect.ALLOW,
    actions: [
        "execute-api:ManageConnections",
    ],
    resources: [
        `arn:${core.Aws.PARTITION}:execute-api:*:${core.Aws.ACCOUNT_ID}:*/*/*/*`
    ],
}))

const integrationV1 = new apiGWv1.LambdaIntegration(handler)

const integrationV2 = new HttpLambdaIntegration("HTTPAPI", handler)

const restAPI = new apiGWv1.RestApi(stack, `${prefix}API-REST`, {
    restApiName: "sample-app-rest",
    cloudWatchRole: false,
    endpointTypes: [apiGWv1.EndpointType.REGIONAL],
    minCompressionSize: Size.kibibytes(100),
    policy: new iam.PolicyDocument({
        statements: [
            new iam.PolicyStatement({
                effect: Effect.ALLOW,
                principals: [
                    new iam.AnyPrincipal()
                ],
                actions: [
                    "execute-api:Invoke"
                ],
                resources: [
                    "execute-api:/*"
                ],
                conditions: {
                    StringEquals: {
                        "aws:PrincipalOrgID": [
                            "o-aq4agy4d07"  // dmgw
                        ],
                    }
                },
            })
        ]
    })
})

restAPI.root.addProxy({
    anyMethod: true,
    defaultIntegration: integrationV1,
    defaultMethodOptions: {
        authorizationType: AuthorizationType.IAM,
    }
})

const deploy = new apiGWv1.Deployment(stack, `${prefix}-API-REST-Deploy`, {
    api: restAPI,
})

const stages = ["dev"].map(stageName => {
    const stage = new apiGWv1.Stage(stack, `${prefix}-API-REST-Stage-${stageName}`, {
        stageName,
        deployment: deploy,
    })
    handler.addPermission(`${prefix}Func-Policy-API-REST-Stage-${stageName}`, {
        principal: new iam.ServicePrincipal("apigateway.amazonaws.com"),
        action: "lambda:InvokeFunction",
        sourceArn: restAPI.arnForExecuteApi("*", "/*", stageName)
    })
    return stage
})

handler.addPermission(`${prefix}Func-Policy-API-REST`, {
    principal: new iam.ServicePrincipal("apigateway.amazonaws.com"),
    action: "lambda:InvokeFunction",
    sourceArn: restAPI.arnForExecuteApi()
})

const httpAPI = new apiGWv2.HttpApi(stack, `${prefix}API-HTTP`, {
    apiName: "sample-app-http",
    createDefaultStage: true,
})

httpAPI.addRoutes({
    path: "/{proxy+}",
    methods: [
        apiGWv2.HttpMethod.ANY,
    ],
    integration: integrationV2,
})

new apiGWv2.HttpStage(stack, `${prefix}APIStage`, {
    httpApi: httpAPI,
    stageName: "test",
    autoDeploy: true,
})

const integrationWS = new WebSocketLambdaIntegration("WebsocketAPI", handler)

const webSocketApi = new apiGWv2.WebSocketApi(stack, `${prefix}API-WS`, {
    apiName: "websocket-api",
    routeSelectionExpression: "$request.body.action",
    connectRouteOptions: {
        integration: integrationWS,
    },
    disconnectRouteOptions: {
        integration: integrationWS,
    },
    defaultRouteOptions: {
        integration: integrationWS,
    },
})

new apiGWv2.WebSocketStage(stack, `${prefix}API-WS-Prod`, {
    stageName: "prod",
    webSocketApi,
    autoDeploy: true,
})

app.synth()
