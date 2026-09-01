package rule

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Exported is the file's complete `/* exported */` view: the names an inline
// directive marks as intentionally global, plus the declaration metadata behind
// them. Its zero value declares no names.
//
// The directive is ESLint's script-only escape hatch for code that shares
// globals across separately loaded files: a name listed here is declared in one
// file and consumed by another, so the rules that would otherwise call it
// unused, or call declaring it a global-scope pollution, leave it alone.
//
// ESLint applies it by looking each name up in the global scope and setting
// `eslintUsed` and `eslintExported` on the variable it finds, which silently
// drops names bound anywhere else — a module's top-level bindings live in the
// module scope, so the directive does nothing there. rslint has no such
// per-variable flag, so this view answers for the raw name alone. A rule
// holding a declaration should ask [RuleContext.IsExportedGlobalBinding],
// which performs that lookup; Has is for the rules that have no declaration to
// resolve, such as no-implicit-globals' assignment to a read-only global.
type Exported struct {
	names        map[string]bool
	declarations []InlineExported
}

// NewExported constructs the per-file exported view. The map and slice are
// treated as read-only and may be shared with the comment store that produced
// them.
func NewExported(names map[string]bool, declarations []InlineExported) Exported {
	return Exported{names: names, declarations: declarations}
}

// Has reports whether an `/* exported */` comment in this file lists name.
func (e Exported) Has(name string) bool {
	return e.names[name]
}

// Declarations returns every parsed inline exported declaration in source
// order, including repeated names and their source ranges. Treat it as
// read-only.
func (e Exported) Declarations() []InlineExported {
	return e.declarations
}

// IsExportedGlobalBinding reports whether an `/* exported */` comment in this
// file lists name and decl binds it in the file's global scope. Together those
// are the lookup ESLint performs before marking a variable: it resolves each
// listed name against the global scope, so a name bound anywhere else — a
// module's or a CommonJS wrapper's top level, a block, a function body — is
// silently dropped.
//
// A rule that wants the directive's effect asks this about the declaration it
// holds, then applies whatever the flag means for it: no-unused-vars and
// no-useless-assignment read it as a use, prefer-const as "may be reassigned
// from another file", no-implicit-globals as "this global is intentional".
func (ctx *RuleContext) IsExportedGlobalBinding(decl *ast.Node, name string) bool {
	if ctx == nil || decl == nil || ctx.SourceFile == nil || !ctx.Exported.Has(name) {
		return false
	}
	if ctx.Refs != nil {
		// RefStore carries the effective source-goal decision. In particular,
		// sourceType: "script" keeps TypeScript's `import x = require()` in the
		// global program scope even though ts-go marks it as an external module
		// indicator based on syntax alone.
		if ctx.Refs.HasNonGlobalProgramScope() {
			return false
		}
	} else {
		// Manually assembled contexts may omit RefStore. Use the normalized
		// source goal as the fallback, including the JS-only CommonJS wrapper.
		sourceType := ctx.LanguageOptions.EffectiveSourceType()
		if sourceType == "module" ||
			(sourceType == "commonjs" && isJavaScriptSourceExtension(ctx.SourceFile.FileName())) {
			return false
		}
	}
	return bindingScopeContainer(decl) == ctx.SourceFile.AsNode()
}

// bindingScopeContainer returns the node owning the scope decl introduces its
// name into: a `var` binding belongs to the nearest function or file scope and
// reaches the global scope from inside any block, while every other
// declaration — `let`/`const`/`class`, a function declaration, a catch
// parameter — belongs to the nearest block scope.
func bindingScopeContainer(decl *ast.Node) *ast.Node {
	root := ast.GetRootDeclaration(decl)
	if root == nil {
		root = decl
	}
	if root.Kind == ast.KindVariableDeclaration && root.Parent != nil &&
		root.Parent.Kind == ast.KindVariableDeclarationList && utils.IsVarKeyword(root.Parent) {
		return utils.FindEnclosingScope(root)
	}
	return ast.GetEnclosingBlockScopeContainer(root)
}
