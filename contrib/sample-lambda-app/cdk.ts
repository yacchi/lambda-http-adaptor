import * as core from "@aws-cdk/core"
import * as iam from "@aws-cdk/aws-iam"
import {Effect} from "@aws-cdk/aws-iam"
import * as lambda from "@aws-cdk/aws-lambda"
import * as apiGWv1 from "@aws-cdk/aws-apigateway"
import {AuthorizationType} from "@aws-cdk/aws-apigateway"
import * as apiGWv2 from "@aws-cdk/aws-apigatewayv2"
import {GoFunction} from "@aws-cdk/aws-lambda-go"
import {LambdaProxyIntegration, LambdaWebSocketIntegration} from "@aws-cdk/aws-apigatewayv2-integrations"
import {RetentionDays} from "@aws-cdk/aws-logs";
import * as path from "path";

(() => {
    const app = new core.App()

    const env = {
        region: app.node.tryGetContext('region') || process.env['CDK_DEFAULT_REGION'] || process.env['AWS_DEFAULT_REGION'],
        account: app.node.tryGetContext('account') || process.env['CDK_DEFAULT_ACCOUNT'] || process.env['AWS_ACCOUNT'],
    }

    const prefix = "SampleApp"

    const stack = new core.Stack(app, "SampleLambdaApp", {env})

    const role = new iam.Role(stack, `${prefix}Role`, {
        roleName: "sample-app",
        assumedBy: new iam.ServicePrincipal("lambda.amazonaws.com"),
        managedPolicies: [
            iam.ManagedPolicy.fromAwsManagedPolicyName("service-role/AWSLambdaBasicExecutionRole")
        ],
        inlinePolicies: {
            Websocket: new iam.PolicyDocument({
                statements: [
                    new iam.PolicyStatement({
                        effect: Effect.ALLOW,
                        actions: [
                            "execute-api:ManageConnections",
                        ],
                        resources: [
                            `arn:${core.Aws.PARTITION}:execute-api:*:${core.Aws.ACCOUNT_ID}:*/*/*/*`  
                        ],
                    })
                ]
            })
        }
    })

    const appDir = path.join(__dirname, "app")

    const handler = new GoFunction(stack, `${prefix}Func`, {
        functionName: "sample-app",
        role,
        logRetention: RetentionDays.THREE_MONTHS,
        runtime: lambda.Runtime.GO_1_X,
        entry: path.join(appDir, "main.go"),
        bundling: {
            goBuildFlags: [`-ldflags='-s -w'`],
            cgoEnabled: false,
        },
        memorySize: 128,
        timeout: core.Duration.minutes(1),
    })

    const integrationV1 = new apiGWv1.LambdaIntegration(handler)

    const integrationV2 = new LambdaProxyIntegration({
        handler,
    })

    const restAPI = new apiGWv1.RestApi(stack, `${prefix}API-REST`, {
        restApiName: "sample-app-rest",
        cloudWatchRole: false,
        endpointTypes: [apiGWv1.EndpointType.REGIONAL],
    })

    restAPI.root.addProxy({
        anyMethod: true,
        defaultIntegration: integrationV1,
        defaultMethodOptions: {
            authorizationType: AuthorizationType.IAM,
        }
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

    const integrationWS = new LambdaWebSocketIntegration({
        handler,
    })

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
})()
