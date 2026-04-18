package no_prototype_builtins

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var disallowedProps = map[string]struct{}{
	"hasOwnProperty":       {},
	"isPrototypeOf":        {},
	"propertyIsEnumerable": {},
}

// isAfterOptional walks the member/call chain leftward from node, returning
// true if any link owns a `?.` token. Parentheses are unwrapped on each step
// so a chain hidden inside `(...)` (tsgo's stand-in for ESTree's
// ChainExpression, e.g. `(foo?.hasOwnProperty)('bar')`) is also caught —
// ESLint handles that case via a separate ChainExpression check, but on the
// tsgo AST the same short-circuiting concern folds into one walk.
func isAfterOptional(node *ast.Node) bool {
	for node != nil {
		node = ast.SkipParentheses(node)
		if node == nil {
			return false
		}
		if ast.IsOptionalChainRoot(node) {
			return true
		}
		switch node.Kind {
		case ast.KindCallExpression:
			node = node.AsCallExpression().Expression
		case ast.KindPropertyAccessExpression:
			node = node.AsPropertyAccessExpression().Expression
		case ast.KindElementAccessExpression:
			node = node.AsElementAccessExpression().Expression
		default:
			return false
		}
	}
	return false
}

// isCommaBinaryExpression reports whether node is a BinaryExpression with the
// comma operator — tsgo's encoding of ESTree's SequenceExpression, which is the
// only operand that ESLint's precedence check wraps in `(...)`.
func isCommaBinaryExpression(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindBinaryExpression {
		return false
	}
	bin := node.AsBinaryExpression()
	return bin != nil && bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindCommaToken
}

// https://eslint.org/docs/latest/rules/no-prototype-builtins
var NoPrototypeBuiltinsRule = rule.Rule{
	Name: "no-prototype-builtins",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				callExpr := node.AsCallExpression()
				if callExpr == nil || callExpr.Expression == nil {
					return
				}

				callee := ast.SkipParentheses(callExpr.Expression)
				if callee == nil {
					return
				}

				var reportNode *ast.Node
				switch callee.Kind {
				case ast.KindPropertyAccessExpression:
					reportNode = callee.AsPropertyAccessExpression().Name()
				case ast.KindElementAccessExpression:
					reportNode = callee.AsElementAccessExpression().ArgumentExpression
				default:
					return
				}
				if reportNode == nil {
					return
				}

				propName, ok := utils.AccessExpressionStaticName(callee)
				if !ok {
					return
				}
				if _, forbidden := disallowedProps[propName]; !forbidden {
					return
				}

				msg := rule.RuleMessage{
					Id:          "prototypeBuildIn",
					Description: fmt.Sprintf("Do not access Object.prototype method '%s' from target object.", propName),
				}

				// No suggestion when the chain may short-circuit, or when the
				// global `Object` is shadowed and can't be referenced as-is.
				if isAfterOptional(node) || utils.IsShadowed(node, "Object") {
					ctx.ReportNode(reportNode, msg)
					return
				}

				obj := utils.AccessExpressionObject(callee)
				if obj == nil {
					ctx.ReportNode(reportNode, msg)
					return
				}

				unwrappedObj := ast.SkipParentheses(obj)
				if unwrappedObj == nil {
					ctx.ReportNode(reportNode, msg)
					return
				}

				objText := utils.TrimmedNodeText(ctx.SourceFile, unwrappedObj)
				if isCommaBinaryExpression(unwrappedObj) {
					objText = "(" + objText + ")"
				}

				text := ctx.SourceFile.Text()
				openParenPos := scanner.SkipTrivia(text, callExpr.Expression.End())
				if openParenPos >= len(text) || text[openParenPos] != '(' {
					ctx.ReportNode(reportNode, msg)
					return
				}

				delim := ", "
				if callExpr.Arguments == nil || len(callExpr.Arguments.Nodes) == 0 {
					delim = ""
				}

				suggestion := rule.RuleSuggestion{
					Message: rule.RuleMessage{
						Id:          "callObjectPrototype",
						Description: fmt.Sprintf("Call Object.prototype.%s explicitly.", propName),
					},
					FixesArr: []rule.RuleFix{
						rule.RuleFixReplace(ctx.SourceFile, callee, "Object.prototype."+propName+".call"),
						rule.RuleFixReplaceRange(
							core.NewTextRange(openParenPos+1, openParenPos+1),
							objText+delim,
						),
					},
				}

				ctx.ReportNodeWithSuggestions(reportNode, msg, suggestion)
			},
		}
	},
}
