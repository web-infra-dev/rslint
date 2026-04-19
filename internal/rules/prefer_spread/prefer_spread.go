package prefer_spread

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/prefer-spread
var PreferSpreadRule = rule.Rule{
	Name: "prefer-spread",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()

				callee := ast.SkipParentheses(call.Expression)
				if !utils.IsSpecificMemberAccess(callee, "", "apply") {
					return
				}

				if call.Arguments == nil {
					return
				}
				args := call.Arguments.Nodes
				if len(args) != 2 {
					return
				}
				if args[1].Kind == ast.KindArrayLiteralExpression ||
					args[1].Kind == ast.KindSpreadElement {
					return
				}

				var memberObject *ast.Node
				switch callee.Kind {
				case ast.KindPropertyAccessExpression:
					memberObject = callee.AsPropertyAccessExpression().Expression
				case ast.KindElementAccessExpression:
					memberObject = callee.AsElementAccessExpression().Expression
				default:
					return
				}

				applied := ast.SkipParentheses(memberObject)
				var expectedThis *ast.Node
				switch applied.Kind {
				case ast.KindPropertyAccessExpression:
					expectedThis = applied.AsPropertyAccessExpression().Expression
				case ast.KindElementAccessExpression:
					expectedThis = applied.AsElementAccessExpression().Expression
				}

				thisArg := args[0]
				if !isValidThisArg(ctx.SourceFile, expectedThis, thisArg) {
					return
				}

				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "preferSpread",
					Description: "Use the spread operator instead of '.apply()'.",
				})
			},
		}
	},
}

// isValidThisArg reports whether the `thisArg` passed to `.apply()` preserves
// the `this` binding of the applied function. When the function is not
// accessed via a member expression (no implicit `this`), only `null` /
// `undefined` / `void 0` are safe. Otherwise the `thisArg` must produce the
// same token stream as the member's object — ESLint's `equalTokens` oracle.
func isValidThisArg(sf *ast.SourceFile, expectedThis, thisArg *ast.Node) bool {
	if expectedThis == nil {
		return utils.IsNullOrUndefined(thisArg)
	}
	return utils.HasSameTokens(sf, expectedThis, thisArg)
}
