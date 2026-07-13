package prefer_jest_mocked

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildUseJestMockedMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useJestMocked",
		Description: "Prefer `jest.mocked()` over type assertions",
	}
}

func isJestMockType(typeNode *ast.Node) bool {
	if typeNode == nil || typeNode.Kind != ast.KindTypeReference {
		return false
	}

	typeRef := typeNode.AsTypeReferenceNode()
	if typeRef == nil || typeRef.TypeName == nil || typeRef.TypeName.Kind != ast.KindQualifiedName {
		return false
	}

	qualifiedName := typeRef.TypeName.AsQualifiedName()
	if qualifiedName == nil ||
		qualifiedName.Left == nil ||
		qualifiedName.Left.Kind != ast.KindIdentifier ||
		qualifiedName.Right == nil ||
		qualifiedName.Right.Kind != ast.KindIdentifier {
		return false
	}

	if qualifiedName.Left.AsIdentifier().Text != "jest" {
		return false
	}

	switch qualifiedName.Right.AsIdentifier().Text {
	case "Mock", "MockedFunction", "MockedClass", "MockedObject":
		return true
	default:
		return false
	}
}

func reportJestMockedAssertion(ctx rule.RuleContext, assertionNode *ast.Node, expression *ast.Node) {
	innerExpression := ast.SkipOuterExpressions(expression, ast.OEKParentheses|ast.OEKTypeAssertions)
	fnName := utils.TrimmedNodeText(ctx.SourceFile, innerExpression)

	ctx.ReportNodeWithFixes(
		assertionNode,
		buildUseJestMockedMessage(),
		rule.RuleFixReplace(ctx.SourceFile, assertionNode, "jest.mocked("+fnName+")"),
	)
}

func checkJestMockAssertion(ctx rule.RuleContext, node *ast.Node, typeNode, expression *ast.Node) {
	if node.Kind == ast.KindAsExpression {
		if parent := ast.WalkUpParenthesizedExpressions(node.Parent); parent != nil && parent.Kind == ast.KindAsExpression {
			return
		}
	}
	if !isJestMockType(typeNode) {
		return
	}

	reportJestMockedAssertion(ctx, node, expression)
}

var PreferJestMockedRule = rule.Rule{
	Name: "jest/prefer-jest-mocked",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindAsExpression: func(node *ast.Node) {
				asExpression := node.AsAsExpression()
				if asExpression == nil {
					return
				}
				checkJestMockAssertion(ctx, node, asExpression.Type, asExpression.Expression)
			},
			ast.KindTypeAssertionExpression: func(node *ast.Node) {
				typeAssertion := node.AsTypeAssertion()
				if typeAssertion == nil {
					return
				}
				checkJestMockAssertion(ctx, node, typeAssertion.Type, typeAssertion.Expression)
			},
		}
	},
}
