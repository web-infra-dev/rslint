package no_array_constructor

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func useLiteralMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useLiteral",
		Description: "The array literal notation [] is preferable.",
	}
}

func buildArrayConstructorFixes(
	sourceText string,
	node *ast.Node,
	args *ast.NodeList,
	reportRange core.TextRange,
) []rule.RuleFix {
	if args == nil {
		return []rule.RuleFix{rule.RuleFixReplaceRange(reportRange, "[]")}
	}

	closeParenPos := node.End() - 1
	if args.Pos() < reportRange.Pos() ||
		closeParenPos < args.Pos() ||
		closeParenPos >= len(sourceText) ||
		sourceText[closeParenPos] != ')' {
		// Keep malformed/recovery ASTs safe and preserve the rule's previous
		// fallback for a missing closing parenthesis.
		return []rule.RuleFix{rule.RuleFixReplaceRange(reportRange, "[]")}
	}

	// Replace only the call boundaries. The argument text, including comments,
	// whitespace, and trailing commas, stays in place without another scan or a
	// copy into a replacement string.
	return []rule.RuleFix{
		rule.RuleFixReplaceRange(core.NewTextRange(reportRange.Pos(), args.Pos()), "["),
		rule.RuleFixReplaceRange(core.NewTextRange(closeParenPos, node.End()), "]"),
	}
}

var NoArrayConstructorRule = rule.CreateRule(rule.Rule{
	Name: "no-array-constructor",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		check := func(node *ast.Node) {
			var callee *ast.Node
			var args *ast.NodeList
			var typeArgs *ast.NodeList

			switch node.Kind {
			case ast.KindCallExpression:
				callExpr := node.AsCallExpression()
				callee = callExpr.Expression
				args = callExpr.Arguments
				typeArgs = callExpr.TypeArguments
			case ast.KindNewExpression:
				newExpr := node.AsNewExpression()
				callee = newExpr.Expression
				args = newExpr.Arguments
				typeArgs = newExpr.TypeArguments
			default:
				return
			}

			// ESTree does not expose grouping parentheses around a callee. Match
			// that behavior without unwrapping TypeScript-only outer expressions
			// such as non-null or `as` expressions.
			if callee == nil {
				return
			}
			callee = ast.SkipParentheses(callee)

			// Check if callee is an Identifier named "Array"
			if callee.Kind != ast.KindIdentifier {
				return
			}
			identifier := callee.AsIdentifier()
			if identifier.Text != "Array" {
				return
			}

			// Skip if there are type arguments (e.g., Array<Foo>())
			if typeArgs != nil && len(typeArgs.Nodes) > 0 {
				return
			}

			// Skip if there's exactly 1 argument (e.g., Array(5))
			argCount := 0
			if args != nil {
				argCount = len(args.Nodes)
			}
			if argCount == 1 {
				return
			}

			sourceText := ctx.SourceFile.Text()
			reportRange := core.NewTextRange(scanner.SkipTrivia(sourceText, node.Pos()), node.End())
			ctx.ReportRangeWithDeferredFixes(reportRange, useLiteralMessage(), func() []rule.RuleFix {
				return buildArrayConstructorFixes(sourceText, node, args, reportRange)
			})
		}

		return rule.RuleListeners{
			ast.KindCallExpression: check,
			ast.KindNewExpression:  check,
		}
	},
})
