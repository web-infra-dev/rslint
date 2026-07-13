package prefer_spy_on

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	rslintUtils "github.com/web-infra-dev/rslint/internal/utils"
)

func buildUseJestSpyOnMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useJestSpyOn",
		Description: "Use jest.spyOn() instead",
	}
}

func findNodeObject(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}

	if jestUtils.IsMemberAccessNode(node) {
		return rslintUtils.AccessExpressionObject(node)
	}

	if node.Kind == ast.KindCallExpression {
		callee := ast.SkipParentheses(node.AsCallExpression().Expression)
		if jestUtils.IsMemberAccessNode(callee) {
			return rslintUtils.AccessExpressionObject(callee)
		}
	}

	return nil
}

func getJestFnCall(node *ast.Node) *ast.Node {
	node = ast.SkipParentheses(node)
	if node == nil {
		return nil
	}

	if node.Kind != ast.KindCallExpression && !jestUtils.IsMemberAccessNode(node) {
		return nil
	}

	obj := findNodeObject(node)
	if obj == nil {
		return nil
	}

	if obj.Kind == ast.KindIdentifier && obj.AsIdentifier().Text == "jest" {
		if node.Kind == ast.KindCallExpression &&
			jestUtils.CalleeChainName(node.AsCallExpression().Expression) == "jest.fn" {
			return node
		}
		return nil
	}

	return getJestFnCall(obj)
}

func accessExpressionPropertyNode(left *ast.Node) *ast.Node {
	switch left.Kind {
	case ast.KindPropertyAccessExpression:
		return left.AsPropertyAccessExpression().Name()
	case ast.KindElementAccessExpression:
		return left.AsElementAccessExpression().ArgumentExpression
	default:
		return nil
	}
}

func leftPropQuote(left *ast.Node) string {
	if left.Kind == ast.KindPropertyAccessExpression {
		return "'"
	}
	return ""
}

func getAutoFixMockImplementation(jestFnCall *ast.Node, ctx rule.RuleContext) string {
	parent := jestFnCall.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	if parent != nil && parent.Kind == ast.KindPropertyAccessExpression {
		prop := parent.AsPropertyAccessExpression().Name()
		if jestUtils.IsNamedMember(prop, "mockImplementation") {
			return ""
		}
	}

	call := jestFnCall.AsCallExpression()
	if call != nil && call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
		arg := call.Arguments.Nodes[0]
		return ".mockImplementation(" + rslintUtils.TrimmedNodeText(ctx.SourceFile, arg) + ")"
	}

	return ".mockImplementation()"
}

func jestFnCallFixEnd(jestFnCall *ast.Node) int {
	node := jestFnCall
	for node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression {
		node = node.Parent
	}
	return node.End()
}

func wrappedJestFnChainClosingParenFixes(jestFnCall *ast.Node) []rule.RuleFix {
	node := jestFnCall
	for node.Parent != nil &&
		(node.Parent.Kind == ast.KindPropertyAccessExpression || node.Parent.Kind == ast.KindCallExpression) {
		node = node.Parent
	}
	if node == jestFnCall {
		return nil
	}

	var fixes []rule.RuleFix
	for node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression {
		node = node.Parent
		fixes = append(fixes, rule.RuleFixRemoveRange(core.NewTextRange(node.End()-1, node.End())))
	}
	return fixes
}

func buildSpyOnFixes(ctx rule.RuleContext, left *ast.Node, jestFnCall *ast.Node) []rule.RuleFix {
	obj := rslintUtils.AccessExpressionObject(left)
	prop := accessExpressionPropertyNode(left)
	quote := leftPropQuote(left)
	mockImplementation := getAutoFixMockImplementation(jestFnCall, ctx)
	jestFnEnd := jestFnCallFixEnd(jestFnCall)

	fixes := []rule.RuleFix{
		rule.RuleFixInsertBefore(ctx.SourceFile, left, "jest.spyOn("),
		rule.RuleFixReplaceRange(
			core.NewTextRange(obj.End(), prop.Pos()),
			", "+quote,
		),
		rule.RuleFixReplaceRange(
			core.NewTextRange(prop.End(), jestFnEnd),
			quote+")"+mockImplementation,
		),
	}

	return append(fixes, wrappedJestFnChainClosingParenFixes(jestFnCall)...)
}

var PreferSpyOnRule = rule.Rule{
	Name: "jest/prefer-spy-on",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				if !ast.IsAssignmentExpression(node, false) {
					return
				}

				bin := node.AsBinaryExpression()
				if bin == nil {
					return
				}

				left := ast.SkipParentheses(bin.Left)
				if !jestUtils.IsMemberAccessNode(left) {
					return
				}

				prop := accessExpressionPropertyNode(left)
				if prop != nil && prop.Kind == ast.KindPrivateIdentifier {
					return
				}

				jestFnCall := getJestFnCall(bin.Right)
				if jestFnCall == nil {
					return
				}

				ctx.ReportNodeWithFixes(
					node,
					buildUseJestSpyOnMessage(),
					buildSpyOnFixes(ctx, left, jestFnCall)...,
				)
			},
		}
	},
}
