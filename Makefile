test: 
	git submodule update --init
	cd typescript-go
	git am --3way --no-gpg-sign ../patches/*.patch
	cd ..
	go test ./internal/...