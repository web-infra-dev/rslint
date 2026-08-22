package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/tspath"
)

// IsInStrictMode checks whether a node is in strict mode code.
// Strict mode is active when:
//   - The file is an ES module (has import/export)
//   - The file or an enclosing function has a "use strict" directive
//   - The node is inside a class body (class bodies are implicitly strict in ES2015+)
//   - The node is inside a TypeScript namespace/module or enum scope
//
// A class's own decorators are the one part of a class that is evaluated
// outside the class scope, so they take their strictness from whatever encloses
// the class.
func IsInStrictMode(node *ast.Node, sourceFile *ast.SourceFile) bool {
	// ES modules are always strict. ast.IsExternalModule also reports true for
	// a .cjs file with no import/export at all: TS forces its
	// ExternalModuleIndicator so it gets its own (non-global) top-level scope,
	// but it is still CommonJS — sloppy mode by default — not a real ES module.
	// Such a file is strict only once it uses module syntax itself.
	if ast.IsExternalModule(sourceFile) &&
		(HasModuleSyntax(sourceFile) || !IsCommonJSFileExtension(sourceFile.FileName())) {
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
		if ast.IsClassLike(current) && !(child.Kind == ast.KindDecorator && child.Parent == current) {
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

// HasModuleSyntax reports whether sourceFile is an external module because of
// what it contains — an import/export statement, `import.meta`, or a JSX tag
// under the JSX module detection — rather than because TypeScript forced an
// ExternalModuleIndicator on it for its file extension alone. TS marks the
// forced case by pointing the indicator at the source file itself, so an
// extension-driven module is exactly the one whose indicator is that sentinel.
func HasModuleSyntax(sourceFile *ast.SourceFile) bool {
	indicator := sourceFile.ExternalModuleIndicator
	return indicator != nil && indicator != sourceFile.AsNode()
}

// IsCommonJSFileExtension reports whether fileName's extension makes the file
// CommonJS rather than an ES module. ESLint's default language selection picks
// CommonJS for exactly one extension, .cjs; every other name, .cts included,
// falls back to the module default. A .cjs file still gets its own (non-global)
// top-level scope, so callers deriving module-ness from scoping need this to
// tell that wrapper scope apart from a real ES module's.
func IsCommonJSFileExtension(fileName string) bool {
	return tspath.GetAnyExtensionFromPath(fileName, nil, false) == tspath.ExtensionCjs
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
