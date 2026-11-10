package prefer_array_index_of

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/jsnum"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// findIndex/findLastIndex visit holes as undefined, but indexOf/lastIndexOf
// skip them. Follow stable initializers to find literals or constructors that
// can create holes.
func hasSparseReceiver(ctx rule.RuleContext, receiver *ast.Node, getEvaluator func() *utils.StaticStringEvaluator) bool {
	seen := make(map[*ast.Node]bool)
	for receiver = utils.SkipAssertionsAndParens(receiver); receiver != nil; receiver = utils.SkipAssertionsAndParens(receiver) {
		switch receiver.Kind {
		case ast.KindArrayLiteralExpression:
			return slices.ContainsFunc(receiver.Elements(), ast.IsOmittedExpression)
		case ast.KindIdentifier:
			if seen[receiver] {
				return false
			}
			seen[receiver] = true
			initializer, ok := getEvaluator().ResolveIdentifierInitializer(receiver)
			if !ok {
				return false
			}
			receiver = initializer
		case ast.KindCallExpression, ast.KindNewExpression:
			callee := utils.ESTreeRuntimeExpression(receiver.Expression())
			args := receiver.Arguments()
			if len(args) != 1 || callee == nil || !ast.IsIdentifier(callee) || callee.Text() != "Array" ||
				!ctx.Globals.Access("Array").IsDeclared() || !unicornutil.IsGlobalReference(ctx, callee) {
				return false
			}
			value, known := getEvaluator().EvalValueIfNoSideEffects(args[0])
			length, numeric := value.(jsnum.Number)
			return !known || (numeric && length > 0)
		default:
			return false
		}
	}
	return false
}

func requiresSuggestion(evaluator *utils.StaticStringEvaluator, value *ast.Node) bool {
	if !evaluator.HasSideEffect(value, true) {
		return false
	}
	if evaluator.HasSideEffect(value, false) {
		return true
	}
	// Keep an autofix only if a potential property read can be folded safely.
	_, known := evaluator.EvalValueIfNoSideEffects(value)
	return !known
}
