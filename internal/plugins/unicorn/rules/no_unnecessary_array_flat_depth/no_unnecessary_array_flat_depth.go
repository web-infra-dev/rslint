package no_unnecessary_array_flat_depth

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
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

type staticPassThroughEvaluatorCacheKey struct{}

func staticPassThroughEvaluator(ctx rule.RuleContext) *utils.StaticStringEvaluator {
	return rule.CachedByFile(ctx, staticPassThroughEvaluatorCacheKey{}, func() *utils.StaticStringEvaluator {
		return utils.NewStaticStringEvaluatorWithReferenceResolver(ctx.TypeChecker, ctx.SourceFile, ctx.Refs)
	})
}

func shouldSkipFlatReceiver(ctx rule.RuleContext, receiver *ast.Node) bool {
	// For source-only JavaScript, Unicorn leaves CallExpression receivers unknown.
	// The exception does not cover safe Object pass-through calls, whose static
	// result Unicorn still classifies.
	runtimeReceiver := utils.ESTreeRuntimeExpression(receiver)
	if runtimeReceiver != nil && ast.IsInJSFile(runtimeReceiver) && ast.IsCallExpression(runtimeReceiver) {
		argument, isPassThrough := staticObjectPassThroughArgument(runtimeReceiver)
		if !isPassThrough || !isSafeStaticPassThroughValue(ctx, argument, map[*ast.Symbol]bool{}) {
			return false
		}
		isArray, known := staticPassThroughEvaluator(ctx).EvalArrayValue(runtimeReceiver)
		return known && !isArray
	}
	return unicornutil.ShouldSkipKnownNonArrayReceiver(ctx, receiver)
}

func staticObjectPassThroughArgument(node *ast.Node) (*ast.Node, bool) {
	oneArgument := 1
	call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
		ArgumentsLength:     &oneArgument,
		RejectSpreadElement: true,
	})
	if !ok {
		return nil, false
	}
	object := utils.ESTreeRuntimeExpression(call.Object)
	if object == nil || !ast.IsIdentifier(object) || object.AsIdentifier().Text != "Object" {
		return nil, false
	}
	switch call.Property.AsIdentifier().Text {
	case "freeze", "preventExtensions", "seal":
		return call.Call.Arguments()[0], true
	default:
		return nil, false
	}
}

func isSafeStaticPassThroughValue(
	ctx rule.RuleContext,
	node *ast.Node,
	visiting map[*ast.Symbol]bool,
) bool {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil {
		return false
	}
	if argument, ok := staticObjectPassThroughArgument(node); ok {
		return isSafeStaticPassThroughValue(ctx, argument, visiting)
	}
	if ast.IsCallExpression(node) || ast.IsNewExpression(node) ||
		node.Kind == ast.KindAwaitExpression || node.Kind == ast.KindYieldExpression ||
		node.Kind == ast.KindDeleteExpression {
		return false
	}
	if ast.IsBinaryExpression(node) && ast.IsAssignmentOperator(node.AsBinaryExpression().OperatorToken.Kind) {
		return false
	}
	if ast.IsIdentifier(node) && !utils.IsNonReferenceIdentifier(node) && ctx.Refs != nil {
		symbol := ctx.Refs.Resolve(node)
		if utils.IsValueSymbolDeclaredInFile(symbol, ctx.SourceFile) {
			if visiting[symbol] {
				return false
			}
			initializer := utils.GetConstVariableInitializer(node, ctx.TypeChecker)
			if initializer == nil {
				return false
			}
			visiting[symbol] = true
			safe := isSafeStaticPassThroughValue(ctx, initializer, visiting)
			delete(visiting, symbol)
			return safe
		}
	}

	safe := true
	node.ForEachChild(func(child *ast.Node) bool {
		if !isSafeStaticPassThroughValue(ctx, child, visiting) {
			safe = false
			return true
		}
		return false
	})
	return safe
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
