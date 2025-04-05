#!/usr/bin/env -S bash -euxo pipefail

pushd typescript-go
git switch main
git reset --hard origin/main
git pull --prune
popd
git add ./typescript-go
pushd typescript-go
git am --3way --no-gpg-sign ../patches/*.patch
popd

go work sync

find ./shim -type f -name 'go.mod' -execdir go get -u -x github.com/microsoft/typescript-go@latest \; -execdir go mod tidy -v \;
go mod tidy

go run ./tools/gen_shims

git add ./shim ./go.mod ./go.sum

go build ./cmd/tsgolint
