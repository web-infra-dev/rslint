module github.com/web-infra-dev/rslint

go 1.26.0

replace (
	github.com/microsoft/TypeScript/tsc/shim/api => ./shim/api
	github.com/microsoft/TypeScript/tsc/shim/api/encoder => ./shim/api/encoder
	github.com/microsoft/TypeScript/tsc/shim/ast => ./shim/ast
	github.com/microsoft/TypeScript/tsc/shim/binder => ./shim/binder
	github.com/microsoft/TypeScript/tsc/shim/bundled => ./shim/bundled
	github.com/microsoft/TypeScript/tsc/shim/checker => ./shim/checker
	github.com/microsoft/TypeScript/tsc/shim/collections => ./shim/collections
	github.com/microsoft/TypeScript/tsc/shim/compiler => ./shim/compiler
	github.com/microsoft/TypeScript/tsc/shim/core => ./shim/core
	github.com/microsoft/TypeScript/tsc/shim/diagnostics => ./shim/diagnostics
	github.com/microsoft/TypeScript/tsc/shim/evaluator => ./shim/evaluator
	github.com/microsoft/TypeScript/tsc/shim/jsonrpc => ./shim/jsonrpc
	github.com/microsoft/TypeScript/tsc/shim/locale => ./shim/locale
	github.com/microsoft/TypeScript/tsc/shim/ls => ./shim/ls
	github.com/microsoft/TypeScript/tsc/shim/lsp/lsproto => ./shim/lsp/lsproto
	github.com/microsoft/TypeScript/tsc/shim/module => ./shim/module
	github.com/microsoft/TypeScript/tsc/shim/parser => ./shim/parser
	github.com/microsoft/TypeScript/tsc/shim/project => ./shim/project
	github.com/microsoft/TypeScript/tsc/shim/scanner => ./shim/scanner
	github.com/microsoft/TypeScript/tsc/shim/transformers/jsxtransforms => ./shim/transformers/jsxtransforms
	github.com/microsoft/TypeScript/tsc/shim/tsoptions => ./shim/tsoptions
	github.com/microsoft/TypeScript/tsc/shim/tspath => ./shim/tspath
	github.com/microsoft/TypeScript/tsc/shim/vfs => ./shim/vfs
	github.com/microsoft/TypeScript/tsc/shim/vfs/cachedvfs => ./shim/vfs/cachedvfs
	github.com/microsoft/TypeScript/tsc/shim/vfs/iovfs => ./shim/vfs/iovfs
	github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs => ./shim/vfs/osvfs
	github.com/microsoft/TypeScript/tsc/shim/vfs/trackingvfs => ./shim/vfs/trackingvfs
	github.com/microsoft/TypeScript/tsc/shim/vfs/vfsmatch => ./shim/vfs/vfsmatch
)

require (
	github.com/bmatcuk/doublestar/v4 v4.10.0
	github.com/fatih/color v1.19.0
	github.com/microsoft/TypeScript/tsc/shim/api/encoder v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/ast v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/binder v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/bundled v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/checker v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/collections v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/compiler v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/core v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/diagnostics v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/evaluator v0.0.0-00010101000000-000000000000
	github.com/microsoft/TypeScript/tsc/shim/jsonrpc v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/locale v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/lsp/lsproto v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/module v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/parser v0.0.0-00010101000000-000000000000
	github.com/microsoft/TypeScript/tsc/shim/project v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/scanner v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/transformers/jsxtransforms v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/tsoptions v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/tspath v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/vfs v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/vfs/cachedvfs v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/vfs/iovfs v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/vfs/trackingvfs v0.0.0
	github.com/rivo/uniseg v0.4.7
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.2
	github.com/tailscale/hujson v0.0.0-20250605163823-992244df8c5a
	github.com/zeebo/xxh3 v1.1.0
	golang.org/x/sync v0.22.0
	golang.org/x/sys v0.47.0
	golang.org/x/tools v0.49.0
	gotest.tools/v3 v3.5.2
)

require (
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/klauspost/compress v1.19.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.10 // indirect
	github.com/mackerelio/go-osstat v0.2.7 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	golang.org/x/mod v0.39.0 // indirect
)

require (
	github.com/dlclark/regexp2 v1.12.0
	github.com/fxamacker/cbor/v2 v2.9.0
	github.com/go-json-experiment/json v0.0.0-20260623181947-01eb4420fa68
	github.com/microsoft/TypeScript/tsc v0.0.0-20260904213532-1f70213d4922 // indirect
	golang.org/x/text v0.41.0
)
