CONTAINER_NAME := sampleapp

build:
	docker build -t $(CONTAINER_NAME) -f contrib/containerd-lambda/Dockerfile .

run-local: stop build
	docker run --rm -d --name $(CONTAINER_NAME) \
		-p 9000:8080 $(CONTAINER_NAME):latest /main

stop:
	docker kill $(CONTAINER_NAME) ; true

lint:
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck@latest ./...

SIMPLE_EXAMPLE_DIR := example/simple

sample-image:
	docker build -t sampleapp -f contrib/containerd-lambda/Dockerfile .

simple-example:
	env GOOS=linux GOARCH=amd64 go build -o $(SIMPLE_EXAMPLE_DIR)/main $(SIMPLE_EXAMPLE_DIR)/main.go
	cd $(SIMPLE_EXAMPLE_DIR) && zip -o main.zip main