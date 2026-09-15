#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."

# Resolve module metadata from the exact submodule revision, then regenerate
# the shims against that checkout. See CONTRIBUTING.md for the update workflow.
test -f typescript-go/tsc/go.mod
# cspell:ignore GOWORK modfile
export GOWORK="$PWD/go.work"

compiler_module=github.com/microsoft/TypeScript/tsc
compiler_revision=$(git -C typescript-go rev-parse HEAD)
compiler_version=$(go list -m -f '{{.Version}}' "$compiler_module@$compiler_revision")

go mod edit "-require=$compiler_module@$compiler_version"
while IFS= read -r -d '' module_file; do
  go mod edit -modfile="$module_file" "-require=$compiler_module@$compiler_version"
done < <(find ./shim -type f -name go.mod -print0)

go run ./tools/gen_shims

while IFS= read -r -d '' module_file; do
  (cd "$(dirname "$module_file")" && go mod tidy)
done < <(find ./shim -type f -name go.mod -print0)
go mod tidy

go build ./cmd/rslint ./cmd/tsgo
go test "$compiler_module/shim/checker"
