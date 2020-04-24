.PHONY: run
run:
	go run main.go

.PHONY: build
build:
	go build ./...

.PHONY: test
test:
	go test ./...