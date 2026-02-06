package no_extra_non_null_assertion

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// hasOwnQuestionDotToken checks if a node has its own ?. token (not inherited through chain).
func hasOwnQuestionDotToken(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		return node.AsPropertyAccessExpression().QuestionDotToken != nil
	case ast.KindElementAccessExpression:
		return node.AsElementAccessExpression().QuestionDotToken != nil
	case ast.KindCallExpression:
		return node.AsCallExpression().QuestionDotToken != nil
	}
	return false
}

// walkUpParens walks up through ParenthesizedExpression nodes to find the effective parent.
func walkUpParens(node *ast.Node) *ast.Node {
	current := node.Parent
	for current != nil && current.Kind == ast.KindParenthesizedExpression {
		current = current.Parent
	}
	return current
}

// walkUpParensFromNode walks up from a node through ParenthesizedExpression nodes,
// returning the outermost ParenthesizedExpression or the node itself.
func walkUpParensFromNode(node *ast.Node) *ast.Node {
	current := node
	for current.Parent != nil && current.Parent.Kind == ast.KindParenthesizedExpression {
		current = current.Parent
	}
	return current
}

var NoExtraNonNullAssertionRule = rule.CreateRule(rule.Rule{
	Name: "no-extra-non-null-assertion",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		msg := rule.RuleMessage{
			Id:          "noExtraNonNullAssertion",
			Description: "Forbidden extra non-null assertion.",
		}

		reportExtraNonNull := func(node *ast.Node) {
			expression := node.AsNonNullExpression().Expression
			s := scanner.GetScannerForSourceFile(ctx.SourceFile, expression.End())
			fix := rule.RuleFixRemoveRange(s.TokenRange())
			ctx.ReportNodeWithFixes(node, msg, fix)
		}

		return rule.RuleListeners{
			ast.KindNonNullExpression: func(node *ast.Node) {
				// Walk up through parens to find the effective parent
				effectiveParent := walkUpParens(node)
				if effectiveParent == nil {
					return
				}

				// The node wrapped in parens (what the effective parent sees as child)
				wrappedNode := walkUpParensFromNode(node)

				// Case 1: TSNonNullExpression > TSNonNullExpression (e.g., foo!! or (foo!)!)
				if effectiveParent.Kind == ast.KindNonNullExpression {
					reportExtraNonNull(node)
					return
				}

				// Case 2 & 3: Optional chain with non-null assertion on the object/callee
				// e.g. foo!?.bar, foo!?.(), foo!?.[0]
				// Check that the parent has its OWN ?. (not inherited through chain)
				isOptionalChainParent := effectiveParent.Kind == ast.KindPropertyAccessExpression ||
					effectiveParent.Kind == ast.KindElementAccessExpression ||
					effectiveParent.Kind == ast.KindCallExpression

				if isOptionalChainParent {
					if hasOwnQuestionDotToken(effectiveParent) && effectiveParent.Expression() == wrappedNode {
						reportExtraNonNull(node)
						return
					}
				}
			},
		}
	},
})
