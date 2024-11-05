import * as path from "path";
import {createSignedFetcher} from "aws-sigv4-fetch";
import {CloudFormationClient, DescribeStacksCommand} from "@aws-sdk/client-cloudformation";

const client = new CloudFormationClient();

type APIUrls = {
    APIGatewayV1URL: string;
    APIGatewayV2URL: string;
    WebSocketAPIURL: string
    LambdaURL: string;
    LambdaURLStream: string;
}

let urls: APIUrls;

beforeAll(async () => {
    const stack = await client.send(new DescribeStacksCommand({
        StackName: "SampleLambdaApp",
    }));

    const res = stack.Stacks![0].Outputs?.reduce((acc, output) => {
        acc[output.OutputKey as keyof APIUrls] = output.OutputValue!;
        return acc;
    }, {} as APIUrls);

    if (!res) {
        throw new Error("No outputs found");
    }

    urls = res;
})

describe("API Access Test", () => {
    const fetchExecuteAPI = createSignedFetcher({
        region: "ap-northeast-1",
        service: "execute-api",
    })

    const fetchLambdaURL = createSignedFetcher({
        region: "ap-northeast-1",
        service: "lambda",
    })

    it("APIGatewayV1 Get Test", async () => {
        const res = await fetchExecuteAPI(path.join(urls.APIGatewayV1URL, "ping"), {
            method: "GET",
        });
        expect(res.status).toBe(200);
        expect(res.headers.get("content-type")).toBe("application/json");
        const body = await res.json();
        expect(body).toEqual({message: "pong"});
    })

    it("APIGatewayV1 Post Test", async () => {
        const message = {message: "Hello, World!"};
        const res = await fetchExecuteAPI(path.join(urls.APIGatewayV1URL, "echo"), {
            method: "POST",
            body: JSON.stringify(message),
            headers: {
                "content-type": "application/json",
            },
        });
        expect(res.status).toBe(200);
        expect(res.headers.get("content-type")).toBe("application/json");
        const body = await res.json();
        expect(body).toEqual(message);
    })

    it("APIGatewayV2 Get Test", async () => {
        const res = await fetchExecuteAPI(path.join(urls.APIGatewayV2URL, "test", "ping"), {
            method: "GET",
        });
        expect(res.status).toBe(200);
        expect(res.headers.get("content-type")).toBe("application/json");
        const body = await res.json();
        expect(body).toEqual({message: "pong"});
    })

    it("APIGatewayV2 Post Test", async () => {
        const message = {message: "Hello, World!"};
        const res = await fetchExecuteAPI(path.join(urls.APIGatewayV2URL, "test", "echo"), {
            method: "POST",
            body: JSON.stringify(message),
            headers: {
                "content-type": "application/json",
            },
        });
        expect(res.status).toBe(200);
        expect(res.headers.get("content-type")).toBe("application/json");
        const body = await res.json();
        expect(body).toEqual(message);
    })

    it("LambdaURL(buffered) Get Test", async () => {
        const res = await fetchLambdaURL(path.join(urls.LambdaURL, "ping"), {
            method: "GET",
        });
        expect(res.status).toBe(200);
        expect(res.headers.get("content-type")).toBe("application/json");
        const body = await res.json();
        expect(body).toEqual({message: "pong"});
    })

    it("LambdaURL(buffered) Post Test", async () => {
        const message = {message: "Hello, World!"};
        const res = await fetchLambdaURL(path.join(urls.LambdaURL, "echo"), {
            method: "POST",
            body: JSON.stringify(message),
            headers: {
                "content-type": "application/json",
            },
        });
        expect(res.status).toBe(200);
        expect(res.headers.get("content-type")).toBe("application/json");
        const body = await res.json();
        expect(body).toEqual(message);
    })

    it("LambdaURL(response_stream) Get Test", async () => {
        const res = await fetchLambdaURL(path.join(urls.LambdaURLStream, "ping"), {
            method: "GET",
        });
        expect(res.status).toBe(200);
        expect(res.headers.get("content-type")).toBe("application/json");
        const body = await res.json();
        expect(body).toEqual({message: "pong"});
    })

    it("LambdaURL(response_stream) Post Test", async () => {
        const message = {message: "Hello, World!"};
        const res = await fetchLambdaURL(path.join(urls.LambdaURLStream, "echo"), {
            method: "POST",
            body: JSON.stringify(message),
            headers: {
                "content-type": "application/json",
            },
        });
        expect(res.status).toBe(200);
        expect(res.headers.get("content-type")).toBe("application/json");
        const body = await res.json();
        expect(body).toEqual(message);
    })
})