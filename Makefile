.PHONY: test build run

test:
	CGO_ENABLED=0 go test ./...

build:
	go build server/main.go

run:
	go run server/main.go