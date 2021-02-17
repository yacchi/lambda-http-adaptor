import * as core from "@aws-cdk/core"
import * as iam from "@aws-cdk/aws-iam"
import * as lambda from "@aws-cdk/aws-lambda"
import * as apiGw from "@aws-cdk/aws-apigatewayv2"
import {LambdaProxyIntegration} from "@aws-cdk/aws-apigatewayv2-integrations"
import {RetentionDays} from "@aws-cdk/aws-logs";
import * as path from "path";
import * as child_process from "child_process";

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
    })

    const distDir = path.join(__dirname, "dist")

    child_process.execSync("make", {
        cwd: __dirname,
    })
    
    const handler = new lambda.Function(stack, `${prefix}Func`, {
        functionName: "sample-app",
        role,
        logRetention: RetentionDays.THREE_MONTHS,
        runtime: lambda.Runtime.GO_1_X,
        handler: "main",
        code: new lambda.AssetCode(distDir),
        memorySize: 128,
        timeout: core.Duration.minutes(1),
    })
    
    const integration = new LambdaProxyIntegration({
        handler,
    })
    
    const httpAPI = new apiGw.HttpApi(stack, `${prefix}API`, {
        apiName: "sample-app",
        createDefaultStage: true,
    })
    
    httpAPI.addRoutes({
        path: "/{proxy+}",
        methods: [
            apiGw.HttpMethod.ANY,
        ],
        integration,
    })
    
    new apiGw.HttpStage(stack, `${prefix}APIStage`, {
        httpApi: httpAPI,
        stageName: "test",
        autoDeploy: true,
    })

})()
