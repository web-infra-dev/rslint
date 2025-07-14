init: 
	git submodule update --init
test:
	go test ./internal/...