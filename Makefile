.PHONY: tidy test build vet

tidy:
	go mod tidy

test:
	go test ./...

vet:
	go vet ./...

build:
	go build ./...
