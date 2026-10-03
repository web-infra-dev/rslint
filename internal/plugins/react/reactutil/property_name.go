package reactutil

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// IsAuthoredPropertyName matches an identifier or private name, including a
// computed identifier. String literals are deliberately not identifier names.
func IsAuthoredPropertyName(name *ast.Node, expected string) bool {
	if name == nil {
		return false
	}
	if IdentifierOrPrivateName(name) == expected {
		return true
	}
	if name.Kind == ast.KindComputedPropertyName {
		return IdentifierOrPrivateName(utils.ESTreeRuntimeExpression(name.AsComputedPropertyName().Expression)) == expected
	}
	return false
}
