package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"
)

// IsInStrictMode checks whether a node is in strict mode code.
// Strict mode is active when:
//   - The file is an ES module (has import/export)
//   - The file or an enclosing function has a "use strict" directive
//   - The node is inside a class body (class bodies are implicitly strict in ES2015+)
func IsInStrictMode(node *ast.Node, sourceFile *ast.SourceFile) bool {
	// ES modules are always strict
	if ast.IsExternalModule(sourceFile) {
		return true
	}

	// Check source file prologue for "use strict"
	if HasUseStrictDirective(sourceFile.AsNode()) {
		return true
	}

	// Walk up from node checking each scope boundary
	current := node.Parent
	for current != nil {
		// Class bodies are always strict in ES2015+
		if ast.IsClassLike(current) {
			return true
		}

		// Check function bodies for "use strict" directive
		if ast.IsFunctionLike(current) {
			body := current.Body()
			if body != nil && body.Kind == ast.KindBlock {
				if HasUseStrictDirective(body) {
					return true
				}
			}
		}

		current = current.Parent
	}

	return false
}

// HasUseStrictDirective checks if a block or source file starts with a "use strict" directive.
func HasUseStrictDirective(block *ast.Node) bool {
	for _, stmt := range block.Statements() {
		if !ast.IsPrologueDirective(stmt) {
			break
		}
		expr := stmt.Expression()
		if expr != nil && expr.Text() == "use strict" {
			return true
		}
	}
	return false
}
