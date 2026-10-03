package utils

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

func (staticEvaluator *StaticStringEvaluator) evalNumericBuiltinCall(node *ast.Node) staticEvalResult {
	callee := node.AsCallExpression().Expression
	if staticEvaluator.isBuiltinValue(callee, "parseInt", map[*ast.Symbol]bool{}) ||
		staticEvaluator.isBuiltinMethodValue(callee, "Number", "parseInt", map[*ast.Symbol]bool{}) {
		arguments, ok := staticEvaluator.evalCallArguments(node)
		if !ok {
			return staticEvalResult{}
		}
		return staticParseInt(arguments)
	}
	method, ok := staticEvaluator.builtinMethodName(callee, "Math", map[*ast.Symbol]bool{})
	if !ok {
		return staticEvalResult{}
	}
	builtin, ok := ecmascript.LookupMathBuiltin(method)
	if !ok {
		return staticEvalResult{}
	}
	arguments, ok := staticEvaluator.evalCallArguments(node)
	if !ok {
		return staticEvalResult{}
	}
	count := builtin.Arity
	if count < 0 {
		count = len(arguments)
	}
	numbers := make([]float64, count)
	for index := range count {
		argument := any(staticUndefinedValue{})
		if index < len(arguments) {
			argument = arguments[index]
		}
		number, ok := staticValueToNumber(argument)
		if !ok {
			return staticEvalResult{}
		}
		numbers[index] = number
	}
	return staticEvalResult{value: staticNumberValue(builtin.Evaluate(numbers)), ok: true}
}

func staticParseInt(arguments []any) staticEvalResult {
	text := "undefined"
	if len(arguments) > 0 {
		var ok bool
		text, ok = staticValueToString(arguments[0])
		if !ok || len(text) > maxStaticStringLength {
			return staticEvalResult{}
		}
	}
	var radix int32
	if len(arguments) > 1 {
		number, ok := staticValueToNumber(arguments[1])
		if !ok {
			return staticEvalResult{}
		}
		radix = toInt32(number)
	}
	return staticEvalResult{value: staticNumberValue(ecmascript.NumberParseInt(text, radix)), ok: true}
}
