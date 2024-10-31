TAG ?= lambda-http-adaptor

image:
	docker build -t $(TAG) src

run: image
	docker run --rm \
		--name lambda-http-adaptor \
		-p 9000:8080 \
		$(TAG)
