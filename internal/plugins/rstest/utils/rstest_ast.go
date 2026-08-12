package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

// IsPromiseChainCall remains as a compatibility wrapper for Rstest utilities.
// The framework-neutral implementation lives in test_framework now that both
// Jest and Rstest consume it.
func IsPromiseChainCall(node *ast.Node) bool {
	return testFramework.IsPromiseChainCall(node)
}
