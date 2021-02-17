# Referenced by https://docs.aws.amazon.com/lambda/latest/dg/go-image.html
FROM public.ecr.aws/lambda/provided:al2 as build

# install compiler
RUN yum install -y golang
RUN go env -w GOPROXY=direct

# cache dependencies
ADD go.mod go.sum ./
RUN go mod download

# build
COPY . .
RUN go build -o /main example/simple/main.go

# copy artifacts to a clean image
FROM public.ecr.aws/lambda/provided:al2
COPY --from=build /main /var/runtime/bootstrap
CMD [ "/var/runtime/bootstrap" ]