package no_new_func

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-new-func

// skipTransparent skips parentheses and TS type assertions (as, angle-bracket,
// non-null !, satisfies) so that e.g. (Function as any)("code") is still caught.
const skipTransparent = ast.OEKParentheses | ast.OEKAssertions

var callMethods = map[string]bool{
	"call":  true,
	"apply": true,
	"bind":  true,
}

var msg = rule.RuleMessage{
	Id:          "noFunctionConstructor",
	Description: "The Function constructor is eval.",
}

var NoNewFuncRule = rule.Rule{
	Name: "no-new-func",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// isGlobalFunction checks whether an identifier resolves to the
		// built-in Function (from lib.d.ts), not a user-declared one.
		//
		// Strategy: TypeChecker accurately resolves symbols inside function
		// scopes, but at the top level TypeScript's declaration merging can
		// cause GetSymbolAtLocation to return the global symbol even when a
		// local class/function with the same name exists. So we use both:
		//   1. TypeChecker (when available) — authoritative for non-top-level
		//   2. IsShadowed — catches top-level shadowing that TypeChecker misses
		isGlobalFunction := func(id *ast.Node) bool {
			if utils.IsShadowed(id, "Function") {
				return false
			}
			if ctx.TypeChecker != nil {
				symbol := ctx.TypeChecker.GetSymbolAtLocation(id)
				if symbol == nil {
					return false
				}
				return !utils.IsSymbolDeclaredInFile(symbol, ctx.SourceFile)
			}
			return true
		}

		check := func(node *ast.Node) {
			var callee *ast.Node
			if node.Kind == ast.KindNewExpression {
				callee = node.AsNewExpression().Expression
			} else {
				callee = node.AsCallExpression().Expression
			}

			if callee == nil {
				return
			}

			// Unwrap parentheses and TS assertions so that patterns like
			// (Function)(...), (Function as any)(...), Function!(...) are all caught.
			unwrapped := ast.SkipOuterExpressions(callee, skipTransparent)

			// Case 1: new Function(...), Function(...), (Function)(...)
			if unwrapped.Kind == ast.KindIdentifier && unwrapped.AsIdentifier().Text == "Function" {
				if !isGlobalFunction(unwrapped) {
					return
				}
				ctx.ReportNode(node, msg)
				return
			}

			// Case 2: Function.call(...), Function.apply(...), Function.bind(...)
			// Only applies to CallExpression (not NewExpression)
			if node.Kind != ast.KindCallExpression {
				return
			}

			propName, ok := utils.AccessExpressionStaticName(unwrapped)
			if !ok || !callMethods[propName] {
				return
			}

			obj := ast.SkipOuterExpressions(utils.AccessExpressionObject(unwrapped), skipTransparent)
			if obj == nil || obj.Kind != ast.KindIdentifier || obj.AsIdentifier().Text != "Function" {
				return
			}

			if !isGlobalFunction(obj) {
				return
			}

			ctx.ReportNode(node, msg)
		}

		return rule.RuleListeners{
			ast.KindNewExpression:  check,
			ast.KindCallExpression: check,
		}
	},
}
