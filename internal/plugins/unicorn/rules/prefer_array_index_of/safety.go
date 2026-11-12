package prefer_array_index_of

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/jsnum"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type arrayDensity uint8

const (
	unknownDensity arrayDensity = iota
	denseArray
	sparseArray
)

// findIndex/findLastIndex visit holes as undefined, but indexOf/lastIndexOf
// skip them. An array reached through a binding can have holes even when its
// initializer was dense, for example after delete or a length assignment.
func receiverDensity(ctx rule.RuleContext, receiver *ast.Node, getEvaluator func() *utils.StaticStringEvaluator) arrayDensity {
	seen := make(map[*ast.Node]bool)
	receiver = utils.SkipAssertionsAndParens(receiver)
	for receiver != nil && ast.IsIdentifier(receiver) {
		if seen[receiver] {
			return unknownDensity
		}
		seen[receiver] = true
		initializer, ok := getEvaluator().ResolveIdentifierInitializer(receiver)
		if !ok {
			return unknownDensity
		}
		receiver = utils.SkipAssertionsAndParens(initializer)
	}
	if receiver == nil {
		return unknownDensity
	}
	density := arrayCreationDensity(ctx, receiver, getEvaluator)
	if density == denseArray && len(seen) > 0 {
		return unknownDensity
	}
	return density
}

func arrayCreationDensity(ctx rule.RuleContext, receiver *ast.Node, getEvaluator func() *utils.StaticStringEvaluator) arrayDensity {
	switch receiver.Kind {
	case ast.KindArrayLiteralExpression:
		if slices.ContainsFunc(receiver.Elements(), ast.IsOmittedExpression) {
			return sparseArray
		}
		return denseArray
	case ast.KindCallExpression, ast.KindNewExpression:
		callee := utils.ESTreeRuntimeExpression(receiver.Expression())
		if callee == nil || !ast.IsIdentifier(callee) || callee.Text() != "Array" ||
			!ctx.Globals.Access("Array").IsDeclared() || !unicornutil.IsGlobalReference(ctx, callee) {
			return unknownDensity
		}
		args := receiver.Arguments()
		if slices.ContainsFunc(args, ast.IsSpreadElement) {
			return unknownDensity
		}
		if len(args) == 1 {
			value, known := getEvaluator().EvalValueIfNoSideEffects(args[0])
			length, numeric := value.(jsnum.Number)
			if !known || (numeric && length > 0) {
				return sparseArray
			}
		}
		return denseArray
	default:
		return unknownDensity
	}
}

func mayBeUndefined(ctx rule.RuleContext, evaluator *utils.StaticStringEvaluator, value *ast.Node) bool {
	value = utils.SkipAssertionsAndParens(value)
	if ast.IsLiteralExpression(value) || value.Kind == ast.KindNullKeyword {
		return false
	}
	if value.Kind == ast.KindMetaProperty && value.AsMetaProperty().KeywordToken == ast.KindNewKeyword {
		// An ordinary function call can leave new.target undefined.
		return true
	}
	if nullish, known := evaluator.EvalControlFlowNullish(value); known {
		return nullish
	}
	tc := ctx.TypeChecker
	if tc == nil {
		return true
	}
	valueType := utils.GetConstrainedTypeAtLocation(tc, value)
	// Keep null in the target type because only undefined can match a hole.
	defined := tc.GetUnionType([]*checker.Type{tc.GetNonNullableType(valueType), tc.GetNullType()})
	return tc.IsTypeAssignableTo(tc.GetUndefinedType(), defined) || !tc.IsTypeAssignableTo(valueType, defined)
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
