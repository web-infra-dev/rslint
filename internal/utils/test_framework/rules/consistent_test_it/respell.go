package consistent_test_it

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

// Respell describes how to rename the root identifier of a test registration.
type Respell struct {
	// Root is the identifier the call is made through.
	Root *ast.Node
	// Symbol is Root's resolved symbol, or nil.
	Symbol *ast.Symbol
	// Global reports whether Root resolved to the framework's global API.
	Global bool
	// Text replaces Root, such as `test` or `test.only`.
	Text string
	// Name is the binding Text is called through, such as `test`.
	Name string
	// IsModule reports whether an import specifier names the framework's module.
	IsModule func(string) bool
}

// RespellFixes returns the edits that call the registration through Name, or
// nil when Name cannot be proven to reach the framework API at Root. It reuses
// a plain value import of Name, keeps a global call global, and otherwise adds
// Name to the named import Root came from. The original import stays, since
// other references may still use it.
func RespellFixes(ctx rule.RuleContext, respell Respell) []rule.RuleFix {
	root, name := respell.Root, respell.Name
	fix := rule.RuleFixReplace(ctx.SourceFile, root, respell.Text)
	if !isNameFreeBelowFile(ctx.SourceFile, root, name) {
		return nil
	}
	// Reuse a plain, value import. Aliased imports do not bind the spelling
	// being introduced, even if they import the preferred export.
	for _, statement := range ctx.SourceFile.Statements.Nodes {
		if statement.Kind != ast.KindImportDeclaration {
			continue
		}
		declaration := statement.AsImportDeclaration()
		if !respell.IsModule(declaration.ModuleSpecifier.Text()) {
			continue
		}
		for _, element := range testFramework.NamedImportElements(declaration) {
			if testFramework.ImportedSpecifierName(element) == name && element.Name().Text() == name {
				return []rule.RuleFix{fix}
			}
		}
	}
	if utils.IsShadowed(root, name) {
		return nil
	}
	if respell.Global {
		return []rule.RuleFix{fix}
	}
	if respell.Symbol == nil {
		return nil
	}
	for _, declaration := range respell.Symbol.Declarations {
		if declaration.Kind != ast.KindImportSpecifier || ast.IsTypeOnlyImportDeclaration(declaration) {
			continue
		}
		// Keep the original import for exports, fixture factories, and calls
		// governed by the other option. Insertion preserves comments/trivia.
		return []rule.RuleFix{rule.RuleFixInsertBefore(ctx.SourceFile, declaration, name+", "), fix}
	}
	return nil
}

// isNameFreeBelowFile reports whether no declaration of name sits between root
// and the file scope, so that calling name at root reaches a file-level
// binding or a global.
func isNameFreeBelowFile(sourceFile *ast.SourceFile, root *ast.Node, name string) bool {
	if utils.IsNameShadowedBetween(root, sourceFile.AsNode(), name) {
		return false
	}
	for ancestor := root.Parent; ancestor != nil; ancestor = ancestor.Parent {
		if ancestor.Kind == ast.KindWithStatement {
			return false
		}
		if ancestor.Kind == ast.KindModuleBlock &&
			(utils.HasLocalDeclarationInStatements(ancestor.AsModuleBlock().Statements.Nodes, name) ||
				utils.HasHoistedVarDeclaration(ancestor, name)) {
			return false
		}
		// A class static block is a variable environment of its own, and is not
		// function-like, so a `var` nested inside it is reported by neither the
		// walk above nor `IsNameShadowedBetween`.
		if ancestor.Kind == ast.KindClassStaticBlockDeclaration {
			if body := ancestor.AsClassStaticBlockDeclaration().Body; body != nil &&
				utils.HasHoistedVarDeclaration(body, name) {
				return false
			}
		}
	}
	return true
}
