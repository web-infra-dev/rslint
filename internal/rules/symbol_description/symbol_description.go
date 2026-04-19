package symbol_description

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/symbol-description
var SymbolDescriptionRule = rule.Rule{
	Name: "symbol-description",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				// Unwrap parentheses so `(Symbol)()` / `((Symbol))()` resolve
				// to the inner identifier — ESTree drops parens at parse time,
				// tsgo keeps them as explicit nodes.
				callee := ast.SkipParentheses(call.Expression)
				if callee == nil || callee.Kind != ast.KindIdentifier || callee.Text() != "Symbol" {
					return
				}
				if call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
					return
				}
				if isUserBoundSymbol(ctx, callee) {
					return
				}
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "expected",
					Description: "Expected Symbol to have a description.",
				})
			},
		}
	},
}

// isUserBoundSymbol returns true when the `Symbol` identifier at `callee`
// resolves to any user-site declaration — meaning it's shadowed/redeclared
// and the rule should not report.
//
// Primary signal is `utils.IsShadowed` (scope-based, mirrors ESLint's
// `getVariableByName` + `variable.defs.length === 0`). This is the authoritative
// semantics ESLint uses; the TypeChecker cannot fully replace it because TS's
// symbol resolution disagrees with ESLint's scope manager on hoisted function
// declarations, `let`/`const`/`class` shadowing, and `interface`+`const`
// declaration merging (same root cause noted in `no-new-symbol`).
//
// When TypeChecker is available, it's used as an *additional* shadow signal:
// any `Symbol` symbol whose declarations include a non-lib source file is
// treated as shadowed too. This catches rare TS-only bindings that the
// scope-based walker could miss (e.g. exotic declaration merging shapes),
// without regressing the ESLint-aligned behavior.
func isUserBoundSymbol(ctx rule.RuleContext, callee *ast.Node) bool {
	if utils.IsShadowed(callee, "Symbol") {
		return true
	}
	if ctx.TypeChecker == nil || ctx.Program == nil {
		return false
	}
	sym := ctx.TypeChecker.GetSymbolAtLocation(callee)
	if sym == nil {
		return false
	}
	for _, decl := range sym.Declarations {
		if decl == nil {
			continue
		}
		// Skip pure-type declarations — they share a symbol with the value via
		// TS namespace merging but don't shadow the runtime binding.
		switch decl.Kind {
		case ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration:
			continue
		}
		sf := ast.GetSourceFileOfNode(decl)
		if sf != nil && !utils.IsSourceFileDefaultLibrary(ctx.Program, sf) {
			return true
		}
	}
	return false
}
