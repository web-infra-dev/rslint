package prefer_promise_reject_errors

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

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
				// ESTree drops parentheses at parse time, so ESLint's
				// `node.callee.type === "Identifier"` succeeds for `new (Promise)(...)`.
				// tsgo retains the ParenthesizedExpression wrapper — unwrap it here.
				// TS assertion wrappers (`(Promise as any)`, `Promise!`, `<any>Promise`)
				// are NOT unwrapped: ESLint's identifier check fails on them, so
				// `new (Promise as any)(...)` is not recognized as a Promise constructor.
				callee := ast.SkipParentheses(node.AsNewExpression().Expression)
				if callee == nil || !ast.IsIdentifier(callee) || callee.AsIdentifier().Text != "Promise" {
					return
				}
				args := node.Arguments()
				if len(args) == 0 {
					return
				}
				// Same reasoning as the callee above: ESLint requires
				// `executor.type === "FunctionExpression" || "ArrowFunctionExpression"`
				// at the AST level — assertion-wrapped executors fail that check.
				executor := ast.SkipParentheses(args[0])
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
			// Parens are transparent in ESLint AST; TS assertions are not —
			// `(reject as any)(5)` and `reject!(5)` are NOT flagged by upstream
			// because their callee is TSAsExpression / TSNonNullExpression, not
			// Identifier "reject".
			callee := ast.SkipParentheses(n.AsCallExpression().Expression)
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
				if ast.IsParameterDeclaration(decl) && decl.Parent == executor {
					return true
				}
			}
			return false
		}
	}
	return !utils.IsNameShadowedBetween(callee, executor, name)
}
