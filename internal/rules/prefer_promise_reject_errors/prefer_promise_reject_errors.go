package prefer_promise_reject_errors

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// skipTransparent unwraps parentheses, type assertions, non-null assertions,
// and `satisfies` so that TypeScript-only syntax does not perturb the
// AST-based shape checks ESLint performs at the source level.
const skipTransparent = ast.OEKParentheses | ast.OEKAssertions

func buildRejectAnErrorMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "rejectAnError",
		Description: "Expected the Promise rejection reason to be an Error.",
	}
}

type Options struct {
	AllowEmptyReject bool
}

func parseOptions(options any) Options {
	opts := Options{AllowEmptyReject: false}
	optsMap := utils.GetOptionsMap(options)
	if optsMap != nil {
		if v, ok := optsMap["allowEmptyReject"].(bool); ok {
			opts.AllowEmptyReject = v
		}
	}
	return opts
}

func checkRejectCall(ctx rule.RuleContext, callExpression *ast.Node, allowEmptyReject bool) {
	args := callExpression.Arguments()
	if len(args) == 0 {
		if allowEmptyReject {
			return
		}
		ctx.ReportNode(callExpression, buildRejectAnErrorMessage())
		return
	}
	first := args[0]
	if !utils.CouldBeError(first) || utils.IsUndefinedIdentifier(first) {
		ctx.ReportNode(callExpression, buildRejectAnErrorMessage())
	}
}

var PreferPromiseRejectErrorsRule = rule.Rule{
	Name: "prefer-promise-reject-errors",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				if utils.IsSpecificMemberAccess(node.AsCallExpression().Expression, "Promise", "reject") {
					checkRejectCall(ctx, node, opts.AllowEmptyReject)
				}
			},
			ast.KindNewExpression: func(node *ast.Node) {
				// ESTree drops parentheses, so ESLint's `node.callee.type === "Identifier"`
				// already succeeds for `new (Promise)(...)`. tsgo retains parens (and TS
				// assertions), so unwrap them here to keep behavior aligned.
				callee := ast.SkipOuterExpressions(node.AsNewExpression().Expression, skipTransparent)
				if callee == nil || !ast.IsIdentifier(callee) || callee.AsIdentifier().Text != "Promise" {
					return
				}
				args := node.Arguments()
				if len(args) == 0 {
					return
				}
				executor := ast.SkipOuterExpressions(args[0], skipTransparent)
				if executor == nil || !ast.IsFunctionExpressionOrArrowFunction(executor) {
					return
				}
				params := executor.Parameters()
				if len(params) < 2 {
					return
				}
				// ESLint requires the second parameter to be a plain Identifier:
				// `params[1].type === "Identifier"` excludes destructuring
				// (ObjectPattern / ArrayPattern), defaults (AssignmentPattern),
				// and rest (RestElement). Mirror that here on the tsgo Parameter.
				rejectParam := params[1].AsParameterDeclaration()
				if rejectParam == nil || rejectParam.Initializer != nil || rejectParam.DotDotDotToken != nil {
					return
				}
				rejectName := rejectParam.Name()
				if rejectName == nil || !ast.IsIdentifier(rejectName) {
					return
				}
				name := rejectName.AsIdentifier().Text

				findRejectReferences(ctx, executor, name, func(call *ast.Node) {
					checkRejectCall(ctx, call, opts.AllowEmptyReject)
				})
			},
		}
	},
}

// findRejectReferences walks the executor function's parameters and body and
// invokes `visit` for every CallExpression whose callee identifier resolves to
// a parameter declared on `executor` itself (i.e., not shadowed by an inner
// scope). When the TypeChecker is available, symbol resolution drives the
// decision — this transparently handles function-scope `var reject = …`
// re-binding (which merges with the parameter symbol) and inner block-scoped
// `let` / `const` / nested function shadowing. Otherwise, falls back to a
// name-based scope walk.
func findRejectReferences(ctx rule.RuleContext, executor *ast.Node, name string, visit func(call *ast.Node)) {
	var walk func(n *ast.Node)
	walk = func(n *ast.Node) {
		if n == nil {
			return
		}
		if ast.IsCallExpression(n) {
			callee := ast.SkipOuterExpressions(n.AsCallExpression().Expression, skipTransparent)
			if callee != nil && ast.IsIdentifier(callee) && callee.AsIdentifier().Text == name {
				if isExecutorParameterReference(ctx, callee, executor, name) {
					visit(n)
				}
			}
		}
		n.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	for _, param := range executor.Parameters() {
		walk(param)
	}
	if body := executor.Body(); body != nil {
		walk(body)
	}
}

// isExecutorParameterReference reports whether `callee` resolves to a
// parameter declared on `executor`. With type info we look at the callee's
// symbol declarations; when the parser deduplicates the symbol (e.g.,
// `function(reject, reject)` produces two distinct symbols, one per parameter),
// any of them will still have `executor` as its parent, which is enough to
// confirm the reference targets the executor's binding.
func isExecutorParameterReference(ctx rule.RuleContext, callee *ast.Node, executor *ast.Node, name string) bool {
	if ctx.TypeChecker != nil {
		symbol := ctx.TypeChecker.GetSymbolAtLocation(callee)
		if symbol != nil {
			for _, decl := range symbol.Declarations {
				if ast.IsParameter(decl) && decl.Parent == executor {
					return true
				}
			}
			return false
		}
	}
	return !utils.IsNameShadowedBetween(callee, executor, name)
}
