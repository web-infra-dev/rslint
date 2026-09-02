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

func shouldSkipFlatReceiver(ctx rule.RuleContext, receiver *ast.Node) bool {
	// For source-only JavaScript, Unicorn leaves CallExpression receivers unknown.
	// The exception does not cover safe Object pass-through calls, whose static
	// result Unicorn still classifies.
	runtimeReceiver := utils.ESTreeRuntimeExpression(receiver)
	if runtimeReceiver != nil && ast.IsInJSFile(runtimeReceiver) && ast.IsCallExpression(runtimeReceiver) &&
		!isPotentialStaticObjectPassThroughCall(runtimeReceiver) {
		return false
	}
	return unicornutil.ShouldSkipKnownNonArrayReceiver(ctx, receiver)
}

func isPotentialStaticObjectPassThroughCall(node *ast.Node) bool {
	oneArgument := 1
	call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
		ArgumentsLength:     &oneArgument,
		RejectSpreadElement: true,
	})
	if !ok {
		return false
	}
	object := utils.ESTreeRuntimeExpression(call.Object)
	if object == nil || !ast.IsIdentifier(object) || object.AsIdentifier().Text != "Object" {
		return false
	}
	switch call.Property.AsIdentifier().Text {
	case "freeze", "preventExtensions", "seal":
		return true
	default:
		return false
	}
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
				depth := utils.ESTreeRuntimeExpression(rawArgument)
				if depth == nil || depth.Kind != ast.KindNumericLiteral ||
					utils.NormalizeNumericLiteral(depth.AsNumericLiteral().Text) != "1" {
					return
				}

				if shouldSkipFlatReceiver(ctx, call.Object) {
					return
				}

				ctx.ReportNodeWithDeferredFixes(depth, unnecessaryDepthMessage, func() []rule.RuleFix {
					return removeDepthArgumentFix(ctx.SourceFile, rawArgument)
				})
			},
		}
	},
}
