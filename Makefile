init: 
	git submodule update --init
test:
	go test ./internal/...
build:
	go build -o rslint ./cmd/tsgolint