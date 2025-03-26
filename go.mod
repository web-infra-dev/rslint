module none.none/tsgolint

go 1.24.1

replace (
	github.com/microsoft/typescript-go/shim/ast => ./shim/ast
	github.com/microsoft/typescript-go/shim/bundled => ./shim/bundled
	github.com/microsoft/typescript-go/shim/checker => ./shim/checker
	github.com/microsoft/typescript-go/shim/compiler => ./shim/compiler
	github.com/microsoft/typescript-go/shim/core => ./shim/core
	github.com/microsoft/typescript-go/shim/scanner => ./shim/scanner
	github.com/microsoft/typescript-go/shim/tsoptions => ./shim/tsoptions
	github.com/microsoft/typescript-go/shim/tspath => ./shim/tspath
	github.com/microsoft/typescript-go/shim/vfs => ./shim/vfs
	github.com/microsoft/typescript-go/shim/vfs/osvfs => ./shim/vfs/osvfs
)

require (
	github.com/microsoft/typescript-go/shim/ast v0.0.0
	github.com/microsoft/typescript-go/shim/bundled v0.0.0
	github.com/microsoft/typescript-go/shim/checker v0.0.0
	github.com/microsoft/typescript-go/shim/compiler v0.0.0
	github.com/microsoft/typescript-go/shim/core v0.0.0
	github.com/microsoft/typescript-go/shim/scanner v0.0.0
	github.com/microsoft/typescript-go/shim/tsoptions v0.0.0
	github.com/microsoft/typescript-go/shim/tspath v0.0.0
	github.com/microsoft/typescript-go/shim/vfs v0.0.0
	github.com/microsoft/typescript-go/shim/vfs/osvfs v0.0.0
)

require (
	github.com/dlclark/regexp2 v1.11.5 // indirect
	github.com/go-json-experiment/json v0.0.0-20250223041408-d3c622f1b874 // indirect
	github.com/microsoft/typescript-go v0.0.0-20250320211407-c070823a32d8 // indirect
	golang.org/x/text v0.23.0
)
