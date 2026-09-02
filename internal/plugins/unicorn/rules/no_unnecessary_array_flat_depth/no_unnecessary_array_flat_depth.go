package no_unnecessary_array_flat_depth

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const messageID = "no-unnecessary-array-flat-depth"

var unnecessaryDepthMessage = rule.RuleMessage{
	Id:          messageID,
	Description: "Passing `1` as the `depth` argument is unnecessary.",
}

func removeDepthArgumentFix(sourceFile *ast.SourceFile, argument *ast.Node) []rule.RuleFix {
	removalRange := utils.TrimNodeTextRange(sourceFile, argument)
	tokenScanner := scanner.GetScannerForSourceFile(sourceFile, argument.End())
	if tokenScanner.Token() == ast.KindCommaToken {
		removalRange = removalRange.WithEnd(tokenScanner.TokenEnd())
	}
	return []rule.RuleFix{rule.RuleFixRemoveRange(removalRange)}
}

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/docs/rules/no-unnecessary-array-flat-depth.md
var NoUnnecessaryArrayFlatDepthRule = rule.Rule{
	Name:   "unicorn/no-unnecessary-array-flat-depth",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		oneArgument := 1
		return rule.RuleListeners{
			// Exit order keeps nested diagnostics in source order, matching ESLint's
			// sorted public output while preserving ordinary sibling order.
			rule.ListenerOnExit(ast.KindCallExpression): func(node *ast.Node) {
				call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
					Method:              "flat",
					ArgumentsLength:     &oneArgument,
					AllowOptionalMember: true,
				})
				if !ok {
					return
				}

				rawArgument := call.Call.Arguments()[0]
				depth := ast.SkipParentheses(rawArgument)
				if depth == nil || depth.Kind != ast.KindNumericLiteral ||
					utils.NormalizeNumericLiteral(depth.AsNumericLiteral().Text) != "1" {
					return
				}

				if unicornutil.ShouldSkipKnownNonArrayReceiver(ctx, call.Object) {
					return
				}

				ctx.ReportNodeWithDeferredFixes(depth, unnecessaryDepthMessage, func() []rule.RuleFix {
					return removeDepthArgumentFix(ctx.SourceFile, rawArgument)
				})
			},
		}
	},
}
