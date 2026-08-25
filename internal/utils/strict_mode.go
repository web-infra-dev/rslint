package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/tspath"
)

// IsInStrictMode checks whether a node is in strict mode code.
// Strict mode is active when:
//   - The file is an ES module
//   - The file or an enclosing function has a "use strict" directive
//   - The node is inside a class body (class bodies are implicitly strict in ES2015+)
//   - The node is inside a TypeScript namespace/module or enum scope
//
// A class's own decorators are the one part of a class that is evaluated
// outside the class scope, so they take their strictness from whatever encloses
// the class.
func IsInStrictMode(node *ast.Node, sourceFile *ast.SourceFile) bool {
	sourceType := "script"
	if tspath.GetAnyExtensionFromPath(sourceFile.FileName(), nil, false) == tspath.ExtensionCjs {
		sourceType = "commonjs"
	} else if ast.IsExternalModule(sourceFile) {
		sourceType = "module"
	}
	return IsInStrictModeWithSourceType(node, sourceFile, sourceType)
}

// IsInStrictModeWithSourceType is IsInStrictMode with the normalized source
// goal selected by the caller. CommonJS stays sloppy even when the parser
// accepts import/export syntax, while an explicitly selected module is strict
// regardless of syntax or file extension.
func IsInStrictModeWithSourceType(node *ast.Node, sourceFile *ast.SourceFile, sourceType string) bool {
	if sourceType == "module" {
		return true
	}

	// Check source file prologue for "use strict"
	if HasUseStrictDirective(sourceFile.AsNode()) {
		return true
	}

	// Walk up from node checking each scope boundary
	child := node
	current := node.Parent
	for current != nil {
		// @typescript-eslint/scope-manager treats TS module/namespace and enum
		// scopes as strict scopes. Core rules that query sourceCode.getScope()
		// therefore observe declarations within them as strict code too.
		if current.Kind == ast.KindModuleDeclaration || current.Kind == ast.KindEnumDeclaration {
			return true
		}

		// Class bodies are always strict in ES2015+. A decorator applied to the
		// class itself is evaluated before the class scope exists, so it is not
		// part of that body; a decorator on a member is.
		if ast.IsClassLike(current) && (child.Kind != ast.KindDecorator || child.Parent != current) {
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

		child = current
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
