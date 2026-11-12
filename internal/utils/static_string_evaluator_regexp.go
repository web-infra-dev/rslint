package utils

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// EvalRegExp returns the pattern and flags of a statically known regular
// expression. Constructor arguments must be safe primitive values; folding
// must never assume that an object's coercion has no side effects.
func (staticEvaluator *StaticStringEvaluator) EvalRegExp(node *ast.Node) (pattern, flags string, known bool) {
	return staticEvaluator.evalRegExp(node, make(map[*ast.Node]bool))
}

func (staticEvaluator *StaticStringEvaluator) evalRegExp(node *ast.Node, visiting map[*ast.Node]bool) (string, string, bool) {
	if staticEvaluator == nil || node == nil || visiting[node] {
		return "", "", false
	}
	visiting[node] = true
	defer delete(visiting, node)
	node = SkipAssertionsAndParens(node)
	if value, ok := staticEvaluator.EvalValueIfNoSideEffects(node); ok {
		if literal, ok := value.(*staticRegExpNode); ok {
			pattern, flags := ExtractRegexPatternAndFlags((*ast.Node)(literal).Text())
			return pattern, flags, true
		}
	}
	if ast.IsIdentifier(node) {
		initializer, symbol, ok := staticEvaluator.resolveIdentifierInitializer(node)
		if ok && !staticEvaluator.hasPropertyMutation(symbol) {
			return staticEvaluator.evalRegExp(initializer, visiting)
		}
		return "", "", false
	}
	if node.Kind != ast.KindNewExpression || !staticEvaluator.isBuiltinIdentifier(ESTreeCallCallee(node.Expression()), "RegExp") {
		return "", "", false
	}
	args := node.Arguments()
	values := make([]any, len(args))
	for index, arg := range args {
		value, ok := staticEvaluator.EvalValueIfNoSideEffects(arg)
		_, regex := value.(*staticRegExpNode)
		if !ok || arg.Kind == ast.KindSpreadElement || (staticValueIsAggregate(value) && !regex) {
			return "", "", false
		}
		values[index] = value
	}
	pattern, flags := "", ""
	if len(values) > 0 && !staticValueUndefined(values[0]) {
		var ok bool
		if literal, regex := values[0].(*staticRegExpNode); regex {
			pattern, flags = ExtractRegexPatternAndFlags((*ast.Node)(literal).Text())
		} else if pattern, ok = staticValueToString(values[0]); !ok {
			return "", "", false
		}
	}
	if len(values) > 1 && !staticValueUndefined(values[1]) {
		var ok bool
		if flags, ok = staticValueToString(values[1]); !ok {
			return "", "", false
		}
	}
	seen := ""
	for _, flag := range flags {
		if !strings.ContainsRune("dgimsuvy", flag) || strings.ContainsRune(seen, flag) {
			return "", "", false
		}
		seen += string(flag)
	}
	if !IsValidRegexPattern(pattern, ParseRegexFlags(flags)) {
		return "", "", false
	}
	return pattern, flags, true
}
