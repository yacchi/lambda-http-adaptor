CONTAINER_NAME := sampleapp

build:
	docker build -t $(CONTAINER_NAME) -f contrib/containerd-lambda/Dockerfile .

run-local: stop build
	docker run --rm -d --name $(CONTAINER_NAME) \
		-p 9000:8080 $(CONTAINER_NAME):latest /main

stop:
	docker kill $(CONTAINER_NAME) ; true