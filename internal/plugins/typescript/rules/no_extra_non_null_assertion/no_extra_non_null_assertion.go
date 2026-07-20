package no_extra_non_null_assertion

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var NoExtraNonNullAssertionRule = rule.CreateRule(rule.Rule{
	Name: "no-extra-non-null-assertion",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
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
				// Keep the outermost parentheses as the child seen by the
				// containing expression, matching ESTree's transparent parens.
				wrappedNode := utils.OutermostParenthesizedExpression(node)
				if wrappedNode == nil || wrappedNode.Parent == nil {
					return
				}
				effectiveParent := wrappedNode.Parent

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
					if ast.IsOptionalChainRoot(effectiveParent) &&
						effectiveParent.Expression() == wrappedNode {
						reportExtraNonNull(node)
						return
					}
				}
			},
		}
	},
})
