# Referenced by https://docs.aws.amazon.com/lambda/latest/dg/go-image.html
FROM golang:1.22.8 AS build

WORKDIR /go/src
COPY . ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache \
    CGO_ENABLED=0 go build -tags lambda.norpc -o /main main.go

# copy artifacts to a clean image
FROM public.ecr.aws/lambda/provided:al2023
COPY --from=build /main /var/runtime/bootstrap
CMD ["dummy"]