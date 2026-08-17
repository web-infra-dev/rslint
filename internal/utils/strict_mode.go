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
func IsInStrictMode(node *ast.Node, sourceFile *ast.SourceFile) bool {
	// ES modules are always strict. ast.IsExternalModule also reports true for
	// .cjs/.cts files with no import/export at all: TS forces their
	// ExternalModuleIndicator so they get their own (non-global) top-level
	// scope, but they're still CommonJS — sloppy mode by default — not real ES
	// modules. Such a file is strict only once it uses module syntax itself.
	if ast.IsExternalModule(sourceFile) &&
		(HasModuleSyntax(sourceFile) || !IsCommonJSFileExtension(sourceFile.FileName())) {
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
// CommonJS rather than an ES module: .cjs and .cts. Both still get their own
// (non-global) top-level scope — .cjs from ESLint's default language selection,
// .cts from TypeScript forcing an ExternalModuleIndicator — so callers deriving
// module-ness from scoping need this to tell the two apart. .mjs/.mts are
// excluded: those are genuine ESM.
func IsCommonJSFileExtension(fileName string) bool {
	switch tspath.GetAnyExtensionFromPath(fileName, nil, false) {
	case tspath.ExtensionCjs, tspath.ExtensionCts:
		return true
	default:
		return false
	}
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
