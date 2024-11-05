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
import {Construct} from "constructs";

const app = new core.App()

const env = {
    region: app.node.tryGetContext('region') || process.env['CDK_DEFAULT_REGION'] || process.env['AWS_DEFAULT_REGION'],
    account: app.node.tryGetContext('account') || process.env['CDK_DEFAULT_ACCOUNT'] || process.env['AWS_ACCOUNT'],
}

class SampleLambdaApp extends core.Stack {
    constructor(scope: Construct, id: string, props?: core.StackProps) {
        super(scope, id, props)

        const code = lambda.DockerImageCode.fromImageAsset("src")
        const logGroup = new LogGroup(this, "LogGroup", {
            logGroupName: `/aws/lambda/${id}`,
            retention: RetentionDays.THREE_MONTHS,
            removalPolicy: core.RemovalPolicy.DESTROY,
        })

        const handler = new lambda.DockerImageFunction(this, "Container", {
            code,
            memorySize: 128,
            timeout: core.Duration.minutes(1),
            architecture: lambda.Architecture.X86_64,
            logGroup,
            environment: {
                DEBUG_DUMP_PAYLOAD: "1",
                WEBSOCKET_RESPONSE_MODE: "return",
            },
        })

        const bufferedUrl = handler.addFunctionUrl({
            invokeMode: lambda.InvokeMode.BUFFERED,
        })

        const streamHandler = new lambda.DockerImageFunction(this, "StreamHandler", {
            code,
            memorySize: 128,
            timeout: core.Duration.minutes(1),
            architecture: lambda.Architecture.X86_64,
            logGroup,
            environment: {
                DEBUG_DUMP_PAYLOAD: "1",
                LAMBDA_INVOKE_MODE: "response_stream",
                WEBSOCKET_RESPONSE_MODE: "post_to_connection",
            },
        })

        const streamUrl = streamHandler.addFunctionUrl({
            invokeMode: lambda.InvokeMode.RESPONSE_STREAM,
        })

        const integrationV1 = new apiGWv1.LambdaIntegration(handler)

        const integrationV2 = new HttpLambdaIntegration("HTTPAPI", handler)

        const restAPI = new apiGWv1.RestApi(this, "RESTAPI", {
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

        const deploy = new apiGWv1.Deployment(this, `RESTAPIDeployment`, {
            api: restAPI,
        })

        const stages = ["dev"].map(stageName => {
            const stage = new apiGWv1.Stage(this, `RESTAPIStage-${stageName}`, {
                stageName,
                deployment: deploy,
            })
            handler.addPermission(`FuncPolicyAPIRESTStage-${stageName}`, {
                principal: new iam.ServicePrincipal("apigateway.amazonaws.com"),
                action: "lambda:InvokeFunction",
                sourceArn: restAPI.arnForExecuteApi("*", "/*", stageName)
            })
            return stage
        })

        handler.addPermission(`Func-Policy-API-REST`, {
            principal: new iam.ServicePrincipal("apigateway.amazonaws.com"),
            action: "lambda:InvokeFunction",
            sourceArn: restAPI.arnForExecuteApi()
        })

        const httpAPI = new apiGWv2.HttpApi(this, "HTTPAPI", {
            createDefaultStage: true,
        })

        httpAPI.addRoutes({
            path: "/{proxy+}",
            methods: [
                apiGWv2.HttpMethod.ANY,
            ],
            integration: integrationV2,
        })

        new apiGWv2.HttpStage(this, "APIStage", {
            httpApi: httpAPI,
            stageName: "test",
            autoDeploy: true,
        })

        const webSocketApi1 = new apiGWv2.WebSocketApi(this, "WebsocketAPI1", {
            routeSelectionExpression: "$request.body.action",
            connectRouteOptions: {
                integration: new WebSocketLambdaIntegration("WebsocketAPIConnect", handler),
            },
            disconnectRouteOptions: {
                integration: new WebSocketLambdaIntegration("WebsocketAPIDisconnect", handler),
            },
            defaultRouteOptions: {
                integration: new WebSocketLambdaIntegration("WebsocketAPIDefault", handler),
                returnResponse: true,
            },
        })

        new apiGWv2.WebSocketStage(this, "WebsocketAPIProd1", {
            stageName: "prod",
            webSocketApi: webSocketApi1,
            autoDeploy: true,
        })

        const webSocketApi2 = new apiGWv2.WebSocketApi(this, "WebsocketAPI2", {
            routeSelectionExpression: "$request.body.action",
            connectRouteOptions: {
                integration: new WebSocketLambdaIntegration("WebsocketAPIConnect", streamHandler),
            },
            disconnectRouteOptions: {
                integration: new WebSocketLambdaIntegration("WebsocketAPIDisconnect", streamHandler),
            },
            defaultRouteOptions: {
                integration: new WebSocketLambdaIntegration("WebsocketAPIDefault", streamHandler),
            },
        })

        webSocketApi2.grantManageConnections(streamHandler)

        new apiGWv2.WebSocketStage(this, "WebsocketAPIProd2", {
            stageName: "prod",
            webSocketApi: webSocketApi2,
            autoDeploy: true,
        })

        new core.CfnOutput(this, "LambdaURL", {
            value: bufferedUrl.url,
        })

        new core.CfnOutput(this, "LambdaURLStream", {
            value: streamUrl.url,
        })

        new core.CfnOutput(this, "APIGatewayV1URL", {
            value: restAPI.url,
        })

        new core.CfnOutput(this, "APIGatewayV2URL", {
            value: httpAPI.apiEndpoint,
        })

        new core.CfnOutput(this, "WebSocketAPIURLReturn", {
            value: webSocketApi1.apiEndpoint,
        })

        new core.CfnOutput(this, "WebSocketAPIURLStream", {
            value: webSocketApi2.apiEndpoint,
        })
    }
}

new SampleLambdaApp(app, "SampleLambdaApp", {env})

app.synth()
