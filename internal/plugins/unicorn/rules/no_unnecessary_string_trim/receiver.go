package no_unnecessary_string_trim

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// hasNonStringReceiver follows values that need the rule's Zod or JSDoc checks.
func hasNonStringReceiver(ctx rule.RuleContext, evaluator *utils.StaticStringEvaluator, receiver *ast.Node) bool {
	// Bound the work even for cyclic aliases or repeatedly shared branches.
	remaining := 256
	var visit func(*ast.Node) bool
	knownAlternative := func(node *ast.Node) bool {
		if node == nil || remaining == 0 {
			return false
		}
		return unicornutil.IsKnownNonStringWithAnnotations(ctx, node) ||
			isStaticNonString(ctx.SourceFile, evaluator, ast.SkipOuterExpressions(node, ast.OEKAll)) ||
			visit(node)
	}
	visit = func(node *ast.Node) bool {
		for node != nil && remaining > 0 {
			remaining--
			node = ast.SkipOuterExpressions(node, ast.OEKAll|ast.OEKExpressionTypePassthrough)
			switch node.Kind {
			case ast.KindIdentifier:
				if hasNonStringJSDocReceiver(ctx, node) {
					return true
				}
				initializer, stable := evaluator.ResolveIdentifierInitializer(node)
				if !stable || ast.GetSourceFileOfNode(initializer) != ctx.SourceFile {
					return false
				}
				node = initializer
			case ast.KindCallExpression:
				return isZodStringCall(ctx, node) || hasNonStringJSDocReceiver(ctx, node)
			case ast.KindConditionalExpression:
				conditional := node.AsConditionalExpression()
				return knownAlternative(conditional.WhenTrue) && knownAlternative(conditional.WhenFalse)
			case ast.KindBinaryExpression:
				binary := node.AsBinaryExpression()
				return ast.IsLogicalOrCoalescingBinaryOperator(binary.OperatorToken.Kind) &&
					knownAlternative(binary.Left) && knownAlternative(binary.Right)
			default:
				return false
			}
		}
		return false
	}
	return visit(receiver)
}
