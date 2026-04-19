package no_useless_call

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// isCallOrNonVariadicApply mirrors ESLint's helper of the same name. It
// recognizes a CallExpression whose callee is a non-computed `.call` or
// `.apply` member access, and whose arguments shape matches the form the
// rule cares about (`.call(thisArg, ...)` or `.apply(thisArg, [args])`).
//
// Returns the unwrapped PropertyAccessExpression (`<callee>.call` /
// `<callee>.apply`) and the method name, or nil/"" if the call doesn't
// match.
func isCallOrNonVariadicApply(node *ast.Node) (*ast.Node, string) {
	call := node.AsCallExpression()
	if call == nil || call.Arguments == nil {
		return nil, ""
	}
	args := call.Arguments.Nodes

	// Skip parens around the callee. tsgo has no ChainExpression wrapper
	// (optional chains are flag-based on the access node itself), so the
	// only thing to unwrap here is parentheses — `(foo?.call)(...)`.
	callee := ast.SkipParentheses(call.Expression)
	if callee.Kind != ast.KindPropertyAccessExpression {
		return nil, ""
	}
	prop := callee.AsPropertyAccessExpression()
	name := prop.Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return nil, ""
	}
	methodName := name.AsIdentifier().Text

	switch methodName {
	case "call":
		if len(args) < 1 {
			return nil, ""
		}
	case "apply":
		if len(args) != 2 || args[1].Kind != ast.KindArrayLiteralExpression {
			return nil, ""
		}
	default:
		return nil, ""
	}

	return callee, methodName
}

// https://eslint.org/docs/latest/rules/no-useless-call
var NoUselessCallRule = rule.Rule{
	Name: "no-useless-call",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				callee, methodName := isCallOrNonVariadicApply(node)
				if callee == nil {
					return
				}

				// `applied` is the function being invoked via .call/.apply
				// (i.e. `<applied>.call(...)`). Unwrap parens so things like
				// `(obj?.foo).bar.call(obj?.foo, ...)` resolve correctly.
				applied := ast.SkipParentheses(callee.AsPropertyAccessExpression().Expression)

				// expectedThis is the receiver implied by `applied`. ESLint
				// treats only MemberExpression receivers as having an
				// implied `this`; tsgo splits that into PropertyAccessExpression
				// and ElementAccessExpression.
				var expectedThis *ast.Node
				switch applied.Kind {
				case ast.KindPropertyAccessExpression:
					expectedThis = applied.AsPropertyAccessExpression().Expression
				case ast.KindElementAccessExpression:
					expectedThis = applied.AsElementAccessExpression().Expression
				}

				thisArg := node.AsCallExpression().Arguments.Nodes[0]

				var matches bool
				if expectedThis == nil {
					matches = utils.IsNullOrUndefined(thisArg)
				} else {
					matches = utils.HasSameTokens(ctx.SourceFile, expectedThis, thisArg)
				}
				if !matches {
					return
				}

				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "unnecessaryCall",
					Description: "Unnecessary '." + methodName + "()'.",
				})
			},
		}
	},
}
