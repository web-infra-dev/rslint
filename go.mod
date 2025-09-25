module github.com/web-infra-dev/rslint

go 1.25.0

replace (
	github.com/microsoft/typescript-go/shim/api => ./shim/api
	github.com/microsoft/typescript-go/shim/api/encoder => ./shim/api/encoder
	github.com/microsoft/typescript-go/shim/ast => ./shim/ast
	github.com/microsoft/typescript-go/shim/bundled => ./shim/bundled
	github.com/microsoft/typescript-go/shim/checker => ./shim/checker
	github.com/microsoft/typescript-go/shim/collections => ./shim/collections
	github.com/microsoft/typescript-go/shim/compiler => ./shim/compiler
	github.com/microsoft/typescript-go/shim/core => ./shim/core
	github.com/microsoft/typescript-go/shim/ls => ./shim/ls
	github.com/microsoft/typescript-go/shim/lsp/lsproto => ./shim/lsp/lsproto
	github.com/microsoft/typescript-go/shim/project => ./shim/project
	github.com/microsoft/typescript-go/shim/scanner => ./shim/scanner
	github.com/microsoft/typescript-go/shim/tsoptions => ./shim/tsoptions
	github.com/microsoft/typescript-go/shim/tspath => ./shim/tspath
	github.com/microsoft/typescript-go/shim/vfs => ./shim/vfs
	github.com/microsoft/typescript-go/shim/vfs/cachedvfs => ./shim/vfs/cachedvfs
	github.com/microsoft/typescript-go/shim/vfs/osvfs => ./shim/vfs/osvfs
)

require (
	github.com/bmatcuk/doublestar/v4 v4.9.1
	github.com/fatih/color v1.18.0
	github.com/microsoft/typescript-go/shim/api/encoder v0.0.0
	github.com/microsoft/typescript-go/shim/ast v0.0.0
	github.com/microsoft/typescript-go/shim/bundled v0.0.0
	github.com/microsoft/typescript-go/shim/checker v0.0.0
	github.com/microsoft/typescript-go/shim/collections v0.0.0
	github.com/microsoft/typescript-go/shim/compiler v0.0.0
	github.com/microsoft/typescript-go/shim/core v0.0.0
	github.com/microsoft/typescript-go/shim/ls v0.0.0
	github.com/microsoft/typescript-go/shim/lsp/lsproto v0.0.0
	github.com/microsoft/typescript-go/shim/project v0.0.0
	github.com/microsoft/typescript-go/shim/scanner v0.0.0
	github.com/microsoft/typescript-go/shim/tsoptions v0.0.0
	github.com/microsoft/typescript-go/shim/tspath v0.0.0
	github.com/microsoft/typescript-go/shim/vfs v0.0.0
	github.com/microsoft/typescript-go/shim/vfs/cachedvfs v0.0.0
	github.com/microsoft/typescript-go/shim/vfs/osvfs v0.0.0
	github.com/tailscale/hujson v0.0.0-20250605163823-992244df8c5a
	golang.org/x/sync v0.16.0
	golang.org/x/sys v0.35.0
	golang.org/x/tools v0.35.0
	gotest.tools/v3 v3.5.2
)

require (
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/klauspost/cpuid/v2 v2.0.9 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	golang.org/x/mod v0.26.0 // indirect
)

require (
	github.com/dlclark/regexp2 v1.11.5 // indirect
	github.com/go-json-experiment/json v0.0.0-20250811204210-4789234c3ea1
	github.com/microsoft/typescript-go v0.0.0-20250829050502-5d1d69a77a4c // indirect
	golang.org/x/text v0.28.0
)
