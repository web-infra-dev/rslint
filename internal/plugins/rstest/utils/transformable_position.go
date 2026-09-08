package utils

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
)

// IsTransformablePosition reports whether the call stands on its own as a
// statement, which is the only place the mock transform can lift it out of.
// Rstest moves the call above the module's imports, so a call whose value is
// consumed — an argument to another call, a variable initializer, an operand
// of a comma expression, an awaited expression — is either left untransformed,
// and throws, or is lifted out of an expression that no longer parses without
// it. The wrappers the transform cannot see do not change the position.
func IsTransformablePosition(node *ast.Node) bool {
	// Climb while the parent is only a wrapper around what we came from, so
	// `(rs.mock('./dep'));` is judged by the statement, not by the parentheses.
	outermost := node
	for {
		parent := outermost.Parent
		if parent == nil {
			return false
		}
		if internalUtils.SkipAssertionsAndParens(parent) != outermost {
			return parent.Kind == ast.KindExpressionStatement
		}
		outermost = parent
	}
}
