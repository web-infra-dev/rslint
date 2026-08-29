package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/binder"
)

// IsInStrictMode checks whether a node is in strict mode code.
// Strict mode is active when:
//   - The file is an ES module (has import/export)
//   - The file or an enclosing function has a "use strict" directive
//   - The node is inside a class body (class bodies are implicitly strict in ES2015+)
//   - The node is inside a TypeScript namespace/module scope
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
		// @typescript-eslint/scope-manager treats TS module/namespace scopes as
		// strict scopes. Core rules that query sourceCode.getScope() therefore
		// observe declarations within them as strict code too.
		if current.Kind == ast.KindModuleDeclaration {
			return true
		}

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

// HasUseStrictDirective checks if a source-backed block or source file starts
// with an exact "use strict" directive. Escaped lookalikes do not enable strict
// mode, even when their decoded string value is "use strict".
func HasUseStrictDirective(block *ast.Node) bool {
	if block == nil {
		return false
	}
	sourceFile := ast.GetSourceFileOfNode(block)
	if sourceFile == nil {
		return false
	}
	return binder.FindUseStrictPrologue(sourceFile, block.Statements()) != nil
}
