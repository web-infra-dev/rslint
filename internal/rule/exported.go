package rule

import (
	"github.com/microsoft/typescript-go/shim/ast"
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
// holding a binding's raw binder symbol should ask
// [RuleContext.IsExportedGlobalBinding], which performs that lookup; Has is for
// rules that have no binding to resolve, such as no-implicit-globals'
// assignment to a read-only global.
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
// file lists name and symbol is bound in the file's global scope. Together
// those are the lookup ESLint performs before marking a variable: it resolves
// each listed name against the global scope, so a name bound anywhere else — a
// module's or a CommonJS wrapper's top level, a block, a function body, or a
// TypeScript type scope — is silently dropped.
//
// symbol must be the raw binder symbol attached to the declaration, not a
// checker-merged symbol or an alias target. A rule that wants the directive's
// effect asks this about the binding it holds, then applies whatever the flag
// means for it: no-unused-vars and no-useless-assignment read it as a use,
// prefer-const as "may be reassigned from another file", and
// no-implicit-globals as "this global is intentional".
func (ctx *RuleContext) IsExportedGlobalBinding(symbol *ast.Symbol, name string) bool {
	return ctx != nil && ctx.Exported.Has(name) && ctx.IsGlobalScopeBinding(symbol, name)
}

// IsGlobalScopeBinding reports whether the raw binder symbol is bound directly
// in the outer global scope selected for this file. A module program scope and a
// JavaScript CommonJS wrapper are deliberately non-global. TypeScript-flavoured
// CommonJS files retain the global program scope supplied by their parser.
//
// The symbol owner is authoritative. In particular, tsgo's generic block-scope
// helpers do not model TypeScript's type, mapped, conditional, enum, and
// signature scopes, so syntactic container fallbacks would turn their internal
// bindings into globals. A missing owner therefore fails closed.
func (ctx *RuleContext) IsGlobalScopeBinding(symbol *ast.Symbol, name string) bool {
	if ctx == nil || symbol == nil || ctx.SourceFile == nil {
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
	source := ctx.SourceFile.AsNode()
	// The overwhelmingly common case is a source-file local. Avoid walking
	// outward from its declaration; the fallback also covers declarations
	// represented by a source-file export symbol.
	if locals := source.Locals(); locals != nil {
		local := locals[name]
		// A named default export's export facet is stored under "default",
		// while its paired local retains name in this table.
		if local == symbol || (local != nil && local.ExportSymbol == symbol) {
			return true
		}
	}
	return bindingScope(symbol, name) == source
}
