.PHONY: build
build: dist/main

.PHONY: dist/main
dist/main:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/main app/main.go