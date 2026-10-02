package utils

import (
	"math"

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
	if !ok || method != "max" && method != "min" {
		return staticEvalResult{}
	}
	arguments, ok := staticEvaluator.evalCallArguments(node)
	if !ok {
		return staticEvalResult{}
	}
	value := math.Inf(-1)
	if method == "min" {
		value = math.Inf(1)
	}
	for _, argument := range arguments {
		number, ok := staticValueToNumber(argument)
		if !ok {
			return staticEvalResult{}
		}
		if math.IsNaN(value) || math.IsNaN(number) {
			value = math.NaN()
		} else if method == "max" {
			value = math.Max(value, number)
		} else {
			value = math.Min(value, number)
		}
	}
	return staticEvalResult{value: staticNumberValue(value), ok: true}
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
