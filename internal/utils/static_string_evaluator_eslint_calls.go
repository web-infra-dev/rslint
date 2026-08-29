// cspell:ignore dgimsuvy fdlibm fontcolor fontsize Gurung Hrkt Khema Kirat libm Onal subarray Sunuwar Tigalari Todhri Tulu Unscopables unscopables

package utils

import (
	"math"
	"math/big"
	"math/bits"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	"golang.org/x/text/unicode/norm"
)

const (
	// The capture limit is stable across the Node versions supported by Unicorn
	// 73. V8 parses nested v-mode classes recursively, however, and its failure
	// point varies with the host stack and current call depth. Stop with ample
	// headroom below the tested default supported runtimes instead of pinning
	// the guard to one runtime's exact stack limit.
	maxStaticRegExpCaptures           = 1<<15 - 1
	maxStaticRegExpUnicodeSetsNesting = 1024
)

// evalESLintStaticCall mirrors the safety boundary in eslint-utils:
// every argument must be static, and only allowlisted native functions are
// invoked. Native identity is represented explicitly so stable aliases and
// locally composed objects such as `{abs: Number}` retain their semantics.
func (staticEvaluator *StaticStringEvaluator) evalESLintStaticCall(node *ast.Node) staticEvalResult {
	call := node.AsCallExpression()
	if call == nil || call.Expression == nil {
		return staticEvalResult{}
	}
	// eslint-utils resolves every argument before it inspects an optional
	// callee, so an unknown spread/argument keeps the whole call unknown even
	// when JavaScript itself would short-circuit it.
	arguments, ok := staticEvaluator.evalCallArguments(node)
	if !ok {
		return staticEvalResult{}
	}
	callee := ast.SkipOuterExpressions(call.Expression, ast.OEKParentheses|ast.OEKAssertions)
	if ast.IsAccessExpression(callee) {
		receiver := staticEvaluator.evalValue(AccessExpressionObject(callee))
		if !receiver.ok {
			return staticEvalResult{}
		}
		if _, shortCircuited := receiver.value.(staticOptionalChainShortCircuitValue); shortCircuited {
			if call.QuestionDotToken != nil {
				return receiver
			}
			return staticEvalResult{}
		}
		if staticValueNullish(receiver.value) {
			if call.QuestionDotToken != nil {
				return staticEvalResult{value: staticOptionalChainShortCircuitValue{}, ok: true}
			}
			return staticEvalResult{}
		}
		method, ok := staticEvaluator.evalAccessExpressionKey(callee)
		if !ok {
			return staticEvalResult{}
		}
		calleeValue := staticESLintMemberValue(receiver.value, method)
		if !calleeValue.ok {
			calleeValue = staticESLintBuiltinMember(receiver.value, method)
		}
		if !calleeValue.ok || staticValueNullish(calleeValue.value) {
			return staticEvalResult{}
		}
		if builtin, ok := calleeValue.value.(staticBuiltinValue); ok {
			if result := evalESLintBuiltinCall(staticEvaluator.stringWorkContext, string(builtin), arguments); result.ok {
				return result
			}
		}
		return evalESLintPrototypeCall(staticEvaluator.stringWorkContext, receiver.value, method, arguments)
	}

	calleeValue := staticEvaluator.evalValue(callee)
	if calleeValue.ok {
		if _, shortCircuited := calleeValue.value.(staticOptionalChainShortCircuitValue); shortCircuited {
			if ast.IsOptionalChain(node) && !ast.IsOuterExpression(call.Expression, ast.OEKParentheses) {
				return calleeValue
			}
			return staticEvalResult{}
		}
		if staticValueNullish(calleeValue.value) {
			if call.QuestionDotToken != nil {
				return staticEvalResult{value: staticOptionalChainShortCircuitValue{}, ok: true}
			}
			return staticEvalResult{}
		}
	}
	if calleeValue.ok {
		if builtin, ok := calleeValue.value.(staticBuiltinValue); ok {
			if result := evalESLintBuiltinCall(staticEvaluator.stringWorkContext, string(builtin), arguments); result.ok {
				return result
			}
		}
	}

	return staticEvalResult{}
}

func (staticEvaluator *StaticStringEvaluator) evalESLintStaticNew(node *ast.Node) staticEvalResult {
	newExpression := node.AsNewExpression()
	if newExpression == nil || newExpression.Expression == nil {
		return staticEvalResult{}
	}
	arguments, ok := staticEvaluator.evalCallArguments(node)
	if !ok {
		return staticEvalResult{}
	}
	callee := staticEvaluator.evalValue(newExpression.Expression)
	if !callee.ok {
		return staticEvalResult{}
	}
	builtin, ok := callee.value.(staticBuiltinValue)
	if !ok {
		return staticEvalResult{}
	}

	switch string(builtin) {
	case "Boolean", "Number", "String":
		if string(builtin) == "String" && len(arguments) > 0 {
			if _, symbol := arguments[0].(staticSymbolValue); symbol {
				return staticEvalResult{}
			}
		}
		primitive := evalESLintBuiltinCall(staticEvaluator.stringWorkContext, string(builtin), arguments)
		if !primitive.ok {
			return staticEvalResult{}
		}
		return staticEvalResult{value: staticBoxedValue{value: primitive.value, identity: &staticIdentity{}}, ok: true}
	case "Date":
		return staticDateConstructor(arguments)
	case "Map", "Set":
		return staticCollectionCall(staticEvaluator.stringWorkContext, string(builtin), arguments)
	case "Object":
		return staticObjectCall(arguments)
	case "RegExp":
		return staticRegExpCall(staticEvaluator.stringWorkContext, arguments, true)
	}
	return staticEvalResult{}
}

func evalESLintBuiltinCall(
	work *staticStringWorkContext,
	name string,
	arguments []any,
) staticEvalResult {
	if !work.reserve(max(len(arguments)*8, 1)) {
		return staticEvalResult{}
	}
	switch name {
	case "Array.isArray":
		isArray := staticValueIsArray(staticArgument(arguments, 0))
		return staticEvalResult{value: isArray, ok: true}
	case "Array.of":
		return staticEvalResult{value: staticArrayFromValues(arguments), ok: true}
	case "Boolean":
		if len(arguments) == 0 {
			return staticEvalResult{value: false, ok: true}
		}
		truthy, ok := staticValueTruthy(arguments[0])
		if !ok {
			return staticEvalResult{value: staticUnknownBooleanValue{}, ok: true}
		}
		return staticEvalResult{value: truthy, ok: true}
	case "Number":
		if len(arguments) == 0 {
			return staticEvalResult{value: staticNumberValue(0), ok: true}
		}
		return staticNumberCall(arguments[0])
	case "String":
		if len(arguments) == 0 {
			return staticEvalResult{value: "", ok: true}
		}
		return staticStringCall(arguments[0])
	case "BigInt":
		if len(arguments) == 0 {
			return staticEvalResult{}
		}
		return staticBigIntCall(arguments[0])
	case "Object":
		return staticObjectCall(arguments)
	case "Object.freeze", "Object.preventExtensions", "Object.seal":
		// eslint-utils treats these as pass-through calls. It does not invoke
		// the native method, so the evaluator-owned object is not mutated.
		return staticEvalResult{value: staticArgument(arguments, 0), ok: true}
	case "Object.entries", "Object.keys", "Object.values":
		if len(arguments) == 0 || staticValueNullish(arguments[0]) {
			return staticEvalResult{}
		}
		return staticObjectEntriesCall(work, name, arguments[0])
	case "Object.is":
		equal, ok := staticSameValue(staticArgument(arguments, 0), staticArgument(arguments, 1))
		if !ok {
			return staticEvalResult{value: staticUnknownBooleanValue{}, ok: true}
		}
		return staticEvalResult{value: equal, ok: true}
	case "Object.isExtensible", "Object.isFrozen", "Object.isSealed":
		argument := staticArgument(arguments, 0)
		_, builtin := argument.(staticBuiltinValue)
		if !staticValueIsAggregate(argument) && !builtin {
			return staticEvalResult{value: name != "Object.isExtensible", ok: true}
		}
		// All objects materialized by getStaticValue are fresh/extensible. Its
		// freeze/seal/preventExtensions allowlist is pass-through and never
		// mutates those evaluator-owned values.
		return staticEvalResult{value: name == "Object.isExtensible", ok: true}
	case "Number.isFinite", "Number.isNaN":
		return staticNumberPredicate(name, staticArgument(arguments, 0))
	case "Number.parseFloat", "parseFloat", "Number.parseInt", "parseInt":
		return staticParseNumber(name, arguments)
	case "String.fromCharCode":
		for _, argument := range arguments {
			if !staticCanConvertToNumber(argument) {
				return staticEvalResult{}
			}
			if _, exact := staticValueToNumber(argument); !exact {
				return staticEvalResult{value: staticUnknownStringValue{}, ok: true}
			}
		}
		return staticESLintStringFromCharCode(work, arguments)
	case "String.fromCodePoint":
		return staticStringFromCodePoint(work, arguments)
	case "String.raw":
		return staticStringRawCall(work, arguments)
	case "Date":
		// Date() always returns a non-empty string, independent of its ignored
		// arguments. The exact wall-clock value is deliberately not fabricated.
		return staticEvalResult{value: staticUnknownStringValue{truthy: true}, ok: true}
	case "Date.parse":
		return staticDateParse(staticArgument(arguments, 0))
	case "RegExp":
		return staticRegExpCall(work, arguments, false)
	case "isFinite", "isNaN":
		number, ok := staticValueToNumber(staticArgument(arguments, 0))
		if !ok {
			if staticCanConvertToNumber(staticArgument(arguments, 0)) {
				return staticEvalResult{value: staticUnknownBooleanValue{}, ok: true}
			}
			return staticEvalResult{}
		}
		if name == "isFinite" {
			return staticEvalResult{value: !math.IsNaN(number) && !math.IsInf(number, 0), ok: true}
		}
		return staticEvalResult{value: math.IsNaN(number), ok: true}
	case "isPrototypeOf":
		if _, isNull := staticArgument(arguments, 0).(staticNullValue); isNull {
			return staticEvalResult{value: false, ok: true}
		}
		return staticEvalResult{}
	case "decodeURI", "decodeURIComponent", "encodeURI", "encodeURIComponent", "escape", "unescape":
		return staticURICall(work, name, arguments)
	case "Symbol.for":
		key, ok := staticValueToString(staticArgument(arguments, 0))
		if !ok {
			return staticEvalResult{}
		}
		return staticEvalResult{value: staticSymbolValue{description: key, global: true}, ok: true}
	case "Symbol.keyFor":
		symbol, ok := staticArgument(arguments, 0).(staticSymbolValue)
		if !ok {
			return staticEvalResult{}
		}
		if symbol.hostDependent {
			return staticEvalResult{}
		}
		if !symbol.global {
			return staticEvalResult{value: staticUndefinedValue{}, ok: true}
		}
		return staticEvalResult{value: symbol.description, ok: true}
	}
	if strings.HasPrefix(name, "Math.") {
		return staticMathCall(strings.TrimPrefix(name, "Math."), arguments)
	}
	return staticEvalResult{}
}

func evalESLintPrototypeCall(
	work *staticStringWorkContext,
	receiver any,
	method string,
	arguments []any,
) staticEvalResult {
	if !work.reserve(max(len(arguments)*8, 1)) {
		return staticEvalResult{}
	}
	switch value := receiver.(type) {
	case *staticArrayValue:
		return staticArrayPrototypeCall(work, value, method, arguments)
	case string:
		return staticStringPrototypeCall(work, value, method, arguments)
	case *staticStringNode:
		text, _ := staticValueAsString(value)
		return staticStringPrototypeCall(work, text, method, arguments)
	case staticNumberValue, staticUnknownNumberValue:
		return staticNumberPrototypeCall(receiver, method, arguments)
	case staticCollectionValue:
		if !work.reserve(max(len(value.entries)*64, 1)) {
			return staticEvalResult{}
		}
		switch method {
		case "entries", "keys", "values":
			values := make([]any, 0, len(value.entries))
			for _, entry := range value.entries {
				switch method {
				case "entries":
					if value.kind == "Map" {
						values = append(values, staticArrayFromValues([]any{entry.key, entry.value}))
					} else {
						values = append(values, staticArrayFromValues([]any{entry.key, entry.key}))
					}
				case "keys":
					values = append(values, entry.key)
				case "values":
					if value.kind == "Map" {
						values = append(values, entry.value)
					} else {
						values = append(values, entry.key)
					}
				}
			}
			return staticEvalResult{value: staticIteratorValue{
				values: values, kind: value.kind, identity: &staticIdentity{},
			}, ok: true}
		case "has":
			return staticCollectionHas(value, staticArgument(arguments, 0))
		case "get":
			if value.kind == "Map" {
				return staticMapGet(value, staticArgument(arguments, 0))
			}
		}
	case staticBoxedValue:
		switch staticValueKindOf(value.value) {
		case staticKindString:
			text, ok := staticValueAsString(value.value)
			if ok {
				return staticStringPrototypeCall(work, text, method, arguments)
			}
		case staticKindNumber:
			return staticNumberPrototypeCall(value.value, method, arguments)
		}
	case staticBuiltinObjectValue:
		switch value {
		case "Array.prototype":
			return staticArrayPrototypeCall(work, staticArrayFromValues(nil), method, arguments)
		case "String.prototype":
			return staticStringPrototypeCall(work, "", method, arguments)
		case "Number.prototype":
			return staticNumberPrototypeCall(staticNumberValue(0), method, arguments)
		}
	}
	if staticValueKindOf(receiver) == staticKindNumber {
		return staticNumberPrototypeCall(receiver, method, arguments)
	}
	return staticEvalResult{}
}

func (staticEvaluator *StaticStringEvaluator) evalESLintTypeOfExpression(node *ast.Node) staticEvalResult {
	typeOf := node.AsTypeOfExpression()
	if typeOf == nil {
		return staticEvalResult{}
	}
	operand := staticEvaluator.evalValue(typeOf.Expression)
	if !operand.ok {
		return staticEvalResult{}
	}
	switch staticValueKindOf(operand.value) {
	case staticKindString:
		return staticEvalResult{value: "string", ok: true}
	case staticKindNumber:
		return staticEvalResult{value: "number", ok: true}
	case staticKindBoolean:
		return staticEvalResult{value: "boolean", ok: true}
	case staticKindBigInt:
		return staticEvalResult{value: "bigint", ok: true}
	case staticKindSymbol:
		return staticEvalResult{value: "symbol", ok: true}
	case staticKindUndefined:
		return staticEvalResult{value: "undefined", ok: true}
	case staticKindNull:
		return staticEvalResult{value: "object", ok: true}
	}
	if builtin, ok := operand.value.(staticBuiltinValue); ok && staticBuiltinIsFunction(string(builtin)) {
		return staticEvalResult{value: "function", ok: true}
	}
	if staticValueIsAggregate(operand.value) || staticValueTruthyObject(operand.value) {
		return staticEvalResult{value: "object", ok: true}
	}
	return staticEvalResult{}
}

func staticValueTruthyObject(value any) bool {
	switch value.(type) {
	case staticBuiltinValue, staticSymbolValue:
		return true
	}
	return false
}

func (staticEvaluator *StaticStringEvaluator) evalESLintBinaryExpression(binary *ast.BinaryExpression) staticEvalResult {
	left := staticEvaluator.evalValue(binary.Left)
	right := staticEvaluator.evalValue(binary.Right)
	if !left.ok || !right.ok {
		return staticEvalResult{}
	}
	op := binary.OperatorToken.Kind

	switch op {
	case ast.KindEqualsEqualsToken, ast.KindExclamationEqualsToken:
		equal, ok := staticValuesLooseEqual(left.value, right.value)
		if !ok {
			return staticEvalResult{}
		}
		if op == ast.KindExclamationEqualsToken {
			equal = !equal
		}
		return staticEvalResult{value: equal, ok: true}
	case ast.KindLessThanToken, ast.KindLessThanEqualsToken,
		ast.KindGreaterThanToken, ast.KindGreaterThanEqualsToken:
		comparison, ordered, ok := staticValuesCompare(left.value, right.value)
		if !ok {
			return staticEvalResult{}
		}
		var value bool
		if ordered {
			switch op {
			case ast.KindLessThanToken:
				value = comparison < 0
			case ast.KindLessThanEqualsToken:
				value = comparison <= 0
			case ast.KindGreaterThanToken:
				value = comparison > 0
			default:
				value = comparison >= 0
			}
		}
		return staticEvalResult{value: value, ok: true}
	}

	leftBigInt, leftIsBigInt := left.value.(staticBigIntValue)
	rightBigInt, rightIsBigInt := right.value.(staticBigIntValue)
	if leftIsBigInt || rightIsBigInt {
		if !leftIsBigInt || !rightIsBigInt {
			return staticEvalResult{}
		}
		result := staticBigIntBinary(op, leftBigInt, rightBigInt)
		result.cacheableBigInt = result.ok && left.cacheableBigInt && right.cacheableBigInt
		return result
	}
	if op == ast.KindPlusToken {
		leftPrimitive, leftPrimitiveOK := staticAddPrimitive(left.value)
		rightPrimitive, rightPrimitiveOK := staticAddPrimitive(right.value)
		if !leftPrimitiveOK || !rightPrimitiveOK {
			return staticEvalResult{}
		}
		left.value, right.value = leftPrimitive, rightPrimitive
		if _, ok := left.value.(staticUnknownStringValue); ok {
			if staticCanConvertToString(right.value) {
				return staticEvalResult{value: staticUnknownStringValue{}, ok: true}
			}
			return staticEvalResult{}
		}
		if _, ok := right.value.(staticUnknownStringValue); ok {
			if staticCanConvertToString(left.value) {
				return staticEvalResult{value: staticUnknownStringValue{}, ok: true}
			}
			return staticEvalResult{}
		}
		if staticValueIsString(left.value) || staticValueIsString(right.value) {
			return staticEvaluator.concatStaticValues(left, right)
		}
	}

	leftNumber, leftKnown := staticValueToNumber(left.value)
	rightNumber, rightKnown := staticValueToNumber(right.value)
	if !leftKnown || !rightKnown {
		if staticCanConvertToNumber(left.value) && staticCanConvertToNumber(right.value) {
			return staticEvalResult{value: staticUnknownNumberValue{}, ok: true}
		}
		return staticEvalResult{}
	}
	var result float64
	switch op {
	case ast.KindPlusToken:
		result = leftNumber + rightNumber
	case ast.KindMinusToken:
		result = leftNumber - rightNumber
	case ast.KindAsteriskToken:
		result = leftNumber * rightNumber
	case ast.KindSlashToken:
		result = leftNumber / rightNumber
	case ast.KindPercentToken:
		result = math.Mod(leftNumber, rightNumber)
	case ast.KindAsteriskAsteriskToken:
		var exact bool
		result, exact = staticMathExactSpecialCase("pow", []float64{leftNumber, rightNumber})
		if !exact {
			return staticEvalResult{value: staticUnknownNumberValue{}, ok: true}
		}
	case ast.KindBarToken:
		result = float64(toInt32(leftNumber) | toInt32(rightNumber))
	case ast.KindAmpersandToken:
		result = float64(toInt32(leftNumber) & toInt32(rightNumber))
	case ast.KindCaretToken:
		result = float64(toInt32(leftNumber) ^ toInt32(rightNumber))
	case ast.KindLessThanLessThanToken:
		result = float64(toInt32(leftNumber) << (toUint32(rightNumber) & 31))
	case ast.KindGreaterThanGreaterThanToken:
		result = float64(toInt32(leftNumber) >> (toUint32(rightNumber) & 31))
	case ast.KindGreaterThanGreaterThanGreaterThanToken:
		result = float64(toUint32(leftNumber) >> (toUint32(rightNumber) & 31))
	default:
		return staticEvalResult{}
	}
	return staticEvalResult{value: staticNumberValue(result), ok: true}
}

func staticBigIntBinary(operator ast.Kind, left, right staticBigIntValue) staticEvalResult {
	if left.value == nil || right.value == nil {
		return staticEvalResult{}
	}
	value := new(big.Int)
	switch operator {
	case ast.KindPlusToken:
		value.Add(left.value, right.value)
	case ast.KindMinusToken:
		value.Sub(left.value, right.value)
	case ast.KindAsteriskToken:
		value.Mul(left.value, right.value)
	case ast.KindSlashToken:
		if right.value.Sign() == 0 {
			return staticEvalResult{}
		}
		value.Quo(left.value, right.value)
	case ast.KindPercentToken:
		if right.value.Sign() == 0 {
			return staticEvalResult{}
		}
		value.Rem(left.value, right.value)
	case ast.KindAsteriskAsteriskToken:
		if right.value.Sign() < 0 || !right.value.IsUint64() {
			return staticEvalResult{}
		}
		exponent := right.value.Uint64()
		absolute := new(big.Int).Abs(left.value)
		if exponent > 0 && absolute.Cmp(big.NewInt(1)) > 0 {
			minimumBits, minimumOverflow := staticBigIntProductBits(uint64(absolute.BitLen()-1), exponent)
			maximumBits, maximumOverflow := staticBigIntProductBits(uint64(absolute.BitLen()), exponent)
			if minimumOverflow || minimumBits > maxJavaScriptBigIntBits {
				return staticEvalResult{}
			}
			if absolute.TrailingZeroBits() == uint(absolute.BitLen()-1) {
				maximumBits = minimumBits
				maximumOverflow = false
			}
			if maximumOverflow || maximumBits > maxStaticBigIntBits {
				if maximumOverflow || maximumBits > maxJavaScriptBigIntBits {
					return staticEvalResult{}
				}
				return staticEvalResult{value: staticAbstractBigInt(true), ok: true}
			}
		}
		value.Exp(left.value, right.value, nil)
	case ast.KindBarToken:
		value.Or(left.value, right.value)
	case ast.KindAmpersandToken:
		value.And(left.value, right.value)
	case ast.KindCaretToken:
		value.Xor(left.value, right.value)
	case ast.KindLessThanLessThanToken, ast.KindGreaterThanGreaterThanToken:
		if !right.value.IsInt64() {
			return staticEvalResult{}
		}
		shift := right.value.Int64()
		leftShift := operator == ast.KindLessThanLessThanToken
		magnitude := uint64(shift)
		if shift < 0 {
			magnitude = uint64(-(shift + 1)) + 1
			leftShift = !leftShift
		}
		if leftShift {
			if left.value.Sign() == 0 {
				return staticEvalResult{value: staticBigIntValue{value: new(big.Int)}, ok: true}
			}
			if magnitude > maxJavaScriptBigIntBits ||
				uint64(left.value.BitLen()) > maxJavaScriptBigIntBits-magnitude {
				return staticEvalResult{}
			}
			if magnitude > maxStaticBigIntBits ||
				uint64(left.value.BitLen()) > maxStaticBigIntBits-magnitude {
				return staticEvalResult{value: staticAbstractBigInt(true), ok: true}
			}
			value.Lsh(left.value, uint(magnitude))
		} else {
			if magnitude >= uint64(left.value.BitLen())+1 {
				if left.value.Sign() < 0 {
					return staticEvalResult{value: staticBigIntValue{value: big.NewInt(-1)}, ok: true}
				}
				return staticEvalResult{value: staticBigIntValue{value: new(big.Int)}, ok: true}
			}
			value.Rsh(left.value, uint(magnitude))
		}
	default:
		return staticEvalResult{}
	}
	if value.BitLen() > maxStaticBigIntBits {
		return staticEvalResult{
			value: staticAbstractBigInt(value.Sign() != 0), ok: true,
			bigIntWorkBytes: (value.BitLen() + 7) / 8,
		}
	}
	return staticEvalResult{
		value: staticBigIntValue{value: value}, ok: true,
		bigIntWorkBytes: (value.BitLen() + 7) / 8,
	}
}

func staticBigIntProductBits(factor, exponent uint64) (uint64, bool) {
	if factor == 0 || exponent == 0 {
		return 1, false
	}
	if factor > (^uint64(0)-1)/exponent {
		return 0, true
	}
	return factor*exponent + 1, false
}

func staticValuesCompare(left, right any) (comparison int, ordered bool, ok bool) {
	left, leftPrimitive := staticRelationalPrimitive(left)
	right, rightPrimitive := staticRelationalPrimitive(right)
	if !leftPrimitive || !rightPrimitive {
		return 0, false, false
	}
	if staticValueIsString(left) && staticValueIsString(right) {
		leftText, _ := staticValueAsString(left)
		rightText, _ := staticValueAsString(right)
		return ecmascript.CompareStrings(leftText, rightText), true, true
	}
	leftBigInt, leftIsBigInt := left.(staticBigIntValue)
	rightBigInt, rightIsBigInt := right.(staticBigIntValue)
	if leftIsBigInt && rightIsBigInt && leftBigInt.value != nil && rightBigInt.value != nil {
		if !staticReserveBigIntComparisonWork(leftBigInt, rightBigInt) {
			return 0, false, false
		}
		return leftBigInt.value.Cmp(rightBigInt.value), true, true
	}
	if leftIsBigInt || rightIsBigInt {
		bigint, other, reverse := leftBigInt, right, false
		if rightIsBigInt {
			bigint, other, reverse = rightBigInt, left, true
		}
		if bigint.value == nil {
			return 0, false, false
		}
		if text, isString := staticValueAsString(other); isString {
			parsed := staticBigIntFromString(text)
			otherBigInt, valid := parsed.value.(staticBigIntValue)
			if !parsed.ok || !valid || otherBigInt.value == nil {
				return 0, false, true
			}
			otherBigInt.bigIntWorkContext = bigint.bigIntWorkContext
			if !staticReserveBigIntComparisonWork(bigint, otherBigInt) {
				return 0, false, false
			}
			comparison = bigint.value.Cmp(otherBigInt.value)
		} else {
			number, numberOK := staticValueToNumber(other)
			if !numberOK {
				return 0, false, false
			}
			if math.IsNaN(number) {
				return 0, false, true
			}
			if math.IsInf(number, 1) {
				comparison = -1
			} else if math.IsInf(number, -1) {
				comparison = 1
			} else {
				integer, accuracy := new(big.Float).SetFloat64(number).Int(nil)
				comparison = bigint.value.Cmp(integer)
				if comparison == 0 {
					switch accuracy {
					case big.Below:
						comparison = -1
					case big.Above:
						comparison = 1
					}
				}
			}
		}
		if reverse {
			comparison = -comparison
		}
		return comparison, true, true
	}
	leftNumber, leftOK := staticValueToNumber(left)
	rightNumber, rightOK := staticValueToNumber(right)
	if !leftOK || !rightOK {
		return 0, false, false
	}
	if math.IsNaN(leftNumber) || math.IsNaN(rightNumber) {
		return 0, false, true
	}
	switch {
	case leftNumber < rightNumber:
		return -1, true, true
	case leftNumber > rightNumber:
		return 1, true, true
	default:
		return 0, true, true
	}
}

// BigInt comparisons can scan every machine word even though their result is
// only a boolean. Charge the largest magnitude once per comparison so cached
// values cannot turn a compact fanout into unbounded per-file work.
func staticReserveBigIntComparisonWork(left, right staticBigIntValue) bool {
	work := left.bigIntWorkContext
	if work == nil {
		work = right.bigIntWorkContext
	}
	cost := 32
	// big.Int comparison stops after the sign, magnitude length, or highest
	// differing word. Only equal bit lengths and signs can force a full scan.
	if left.value != nil && right.value != nil &&
		left.value.Sign() == right.value.Sign() && left.value.BitLen() == right.value.BitLen() {
		cost = max(cost, (left.value.BitLen()+7)/8)
	}
	return work.reserve(cost)
}

// staticRelationalPrimitive implements the number-hint ToPrimitive step used
// by Abstract Relational Comparison. It intentionally differs from addition's
// default hint: Date values prefer valueOf here, while arrays and ordinary
// objects still fall through to their string representation.
func staticRelationalPrimitive(value any) (any, bool) {
	switch value := value.(type) {
	case staticDateValue:
		if value.known {
			return value.milliseconds, true
		}
		return staticUnknownNumberValue{}, true
	case staticBoxedValue:
		return value.value, true
	case staticBuiltinObjectValue:
		switch value {
		case "Boolean.prototype":
			return false, true
		case "Number.prototype":
			return staticNumberValue(0), true
		case "BigInt.prototype":
			return staticBigIntValue{value: new(big.Int)}, true
		case "String.prototype", "Array.prototype":
			return "", true
		case "Date.prototype":
			return staticNumberValue(math.NaN()), true
		case "RegExp.prototype":
			return "/(?:)/", true
		case "Symbol.prototype", "Array.prototype[Symbol.unscopables]":
			return nil, false
		}
		text, ok := staticValueToString(value)
		return text, ok
	case staticBuiltinValue:
		if staticBuiltinIsFunction(string(value)) {
			return nil, false
		}
		return "[object " + string(value) + "]", true
	case *staticArrayValue, *staticObjectValue, staticRegExpValue, staticCollectionValue:
		text, ok := staticValueToString(value)
		return text, ok
	case staticOpaqueObjectValue, staticIteratorValue:
		return nil, false
	}
	return value, true
}

func staticAddPrimitive(value any) (any, bool) {
	if builtin, ok := value.(staticBuiltinValue); ok && staticBuiltinIsFunction(string(builtin)) {
		// A native function's source text is engine-specific. It is an object
		// for abstract equality, but its ToPrimitive result is not fabricated.
		return nil, false
	}
	if builtin, ok := value.(staticBuiltinValue); ok && !staticBuiltinIsFunction(string(builtin)) {
		return "[object " + string(builtin) + "]", true
	}
	if staticValueIsAggregate(value) {
		if prototype, ok := value.(staticBuiltinObjectValue); ok {
			switch prototype {
			case "Number.prototype":
				return staticNumberValue(0), true
			case "Boolean.prototype":
				return false, true
			case "String.prototype", "Array.prototype":
				return "", true
			case "BigInt.prototype":
				return staticBigIntValue{value: new(big.Int)}, true
			case "RegExp.prototype":
				return "/(?:)/", true
			}
		}
		if _, ok := value.(staticDateValue); ok {
			return staticUnknownStringValue{truthy: true, numericNaN: true}, true
		}
		if boxed, ok := value.(staticBoxedValue); ok {
			return boxed.value, true
		}
		text, ok := staticValueToString(value)
		return text, ok
	}
	return value, true
}

func staticValuesLooseEqual(left, right any) (bool, bool) {
	leftKind, rightKind := staticValueKindOf(left), staticValueKindOf(right)
	if leftKind == rightKind {
		return staticValuesStrictEqual(left, right)
	}
	if leftKind == staticKindNull && rightKind == staticKindUndefined ||
		leftKind == staticKindUndefined && rightKind == staticKindNull {
		return true, true
	}
	if leftKind == staticKindBoolean {
		leftNumber, ok := staticValueToNumber(left)
		if !ok {
			return false, false
		}
		return staticValuesLooseEqual(staticNumberValue(leftNumber), right)
	}
	if rightKind == staticKindBoolean {
		rightNumber, ok := staticValueToNumber(right)
		if !ok {
			return false, false
		}
		return staticValuesLooseEqual(left, staticNumberValue(rightNumber))
	}
	if leftKind == staticKindString && rightKind == staticKindNumber ||
		leftKind == staticKindNumber && rightKind == staticKindString {
		leftNumber, leftOK := staticValueToNumber(left)
		rightNumber, rightOK := staticValueToNumber(right)
		if !leftOK || !rightOK {
			return false, false
		}
		return leftNumber == rightNumber, true
	}
	if leftKind == staticKindBigInt && (rightKind == staticKindString || rightKind == staticKindNumber) ||
		rightKind == staticKindBigInt && (leftKind == staticKindString || leftKind == staticKindNumber) {
		bigint, other := left, right
		if rightKind == staticKindBigInt {
			bigint, other = right, left
		}
		bigintValue, bigintKnown := bigint.(staticBigIntValue)
		if !bigintKnown || bigintValue.value == nil {
			return false, false
		}
		if text, isString := staticValueAsString(other); isString {
			parsed := staticBigIntFromString(text)
			otherBigInt, valid := parsed.value.(staticBigIntValue)
			if !parsed.ok || !valid || otherBigInt.value == nil {
				return false, true
			}
			otherBigInt.bigIntWorkContext = bigintValue.bigIntWorkContext
			if !staticReserveBigIntComparisonWork(bigintValue, otherBigInt) {
				return false, false
			}
			return bigintValue.value.Cmp(otherBigInt.value) == 0, true
		}
		number, ok := staticValueToNumber(other)
		if !ok || math.IsNaN(number) || math.IsInf(number, 0) || math.Trunc(number) != number {
			return false, ok
		}
		integer, accuracy := new(big.Float).SetFloat64(number).Int(nil)
		return accuracy == big.Exact && bigintValue.value.Cmp(integer) == 0, true
	}
	if staticValueIsObjectForCoercion(left) && !staticValueIsObjectForCoercion(right) {
		primitive, ok := staticAddPrimitive(left)
		if !ok {
			return false, false
		}
		return staticValuesLooseEqual(primitive, right)
	}
	if staticValueIsObjectForCoercion(right) && !staticValueIsObjectForCoercion(left) {
		primitive, ok := staticAddPrimitive(right)
		if !ok {
			return false, false
		}
		return staticValuesLooseEqual(left, primitive)
	}
	return false, true
}

func staticValueIsObjectForCoercion(value any) bool {
	if builtin, ok := value.(staticBuiltinValue); ok {
		return staticBuiltinIsFunction(string(builtin))
	}
	return staticValueIsAggregate(value)
}

func staticArgument(arguments []any, index int) any {
	if index >= len(arguments) {
		return staticUndefinedValue{}
	}
	return arguments[index]
}

func staticNumberCall(value any) staticEvalResult {
	if boxed, ok := value.(staticBoxedValue); ok {
		return staticNumberCall(boxed.value)
	}
	if bigint, ok := value.(staticBigIntValue); ok {
		if bigint.value == nil {
			return staticEvalResult{}
		}
		if bigint.value.BitLen() > 1024 {
			converted := math.Inf(1)
			if bigint.value.Sign() < 0 {
				converted = math.Inf(-1)
			}
			return staticEvalResult{value: staticNumberValue(converted), ok: true}
		}
		converted, _ := new(big.Float).SetInt(bigint.value).Float64()
		return staticEvalResult{value: staticNumberValue(converted), ok: true}
	}
	if _, ok := value.(staticSymbolValue); ok {
		return staticEvalResult{}
	}
	switch value.(type) {
	case staticUnknownNumberValue:
		return staticEvalResult{value: value, ok: true}
	case staticUnknownStringValue, staticUnknownBooleanValue:
		return staticEvalResult{value: staticUnknownNumberValue{}, ok: true}
	}
	number, ok := staticValueToNumber(value)
	if !ok {
		return staticEvalResult{}
	}
	return staticEvalResult{value: staticNumberValue(number), ok: true}
}

func staticStringCall(value any) staticEvalResult {
	if symbol, ok := value.(staticSymbolValue); ok {
		description, known := staticSymbolDescription(symbol)
		if !known {
			return staticEvalResult{value: staticUnknownStringValue{truthy: true, numericNaN: true}, ok: true}
		}
		return staticEvalResult{value: "Symbol(" + description + ")", ok: true}
	}
	if _, ok := value.(staticUnknownStringValue); ok {
		return staticEvalResult{value: value, ok: true}
	}
	switch value.(type) {
	case staticUnknownNumberValue, staticUnknownBooleanValue:
		return staticEvalResult{value: staticUnknownStringValue{}, ok: true}
	}
	if _, ok := value.(staticDateValue); ok {
		return staticEvalResult{value: staticUnknownStringValue{truthy: true, numericNaN: true}, ok: true}
	}
	if builtin, ok := value.(staticBuiltinValue); ok && staticBuiltinIsFunction(string(builtin)) {
		return staticEvalResult{value: staticUnknownStringValue{truthy: true, numericNaN: true}, ok: true}
	}
	text, ok := staticValueToString(value)
	if !ok {
		return staticEvalResult{}
	}
	return staticEvalResult{value: text, ok: true}
}

func staticDateConstructor(arguments []any) staticEvalResult {
	if len(arguments) == 0 {
		return staticEvalResult{value: staticDateValue{identity: &staticIdentity{}}, ok: true}
	}
	for _, argument := range arguments {
		if _, symbol := argument.(staticSymbolValue); symbol {
			return staticEvalResult{}
		}
		if _, bigint := argument.(staticBigIntValue); bigint {
			return staticEvalResult{}
		}
	}
	if len(arguments) == 1 {
		argument := arguments[0]
		if date, ok := argument.(staticDateValue); ok {
			return staticEvalResult{value: staticDateValue{
				milliseconds: date.milliseconds, known: date.known, identity: &staticIdentity{},
			}, ok: true}
		}
		primitive, ok := staticAddPrimitive(argument)
		if !ok {
			// A failed ToPrimitive may represent an exception (for example an
			// invalid Symbol.toPrimitive property), not merely an unknown date.
			// Keep the whole call unknown so receiver classification cannot turn
			// that exception into a known non-array value.
			return staticEvalResult{}
		}
		if staticValueKindOf(primitive) == staticKindString {
			parsed := staticDateParse(primitive)
			milliseconds, known := parsed.value.(staticNumberValue)
			return staticEvalResult{value: staticDateValue{
				milliseconds: milliseconds, known: parsed.ok && known, identity: &staticIdentity{},
			}, ok: true}
		}
		if _, symbol := primitive.(staticSymbolValue); symbol {
			return staticEvalResult{}
		}
		if _, bigint := primitive.(staticBigIntValue); bigint {
			return staticEvalResult{}
		}
		number, ok := staticValueToNumber(primitive)
		if !ok {
			if staticCanConvertToNumber(primitive) {
				return staticEvalResult{value: staticDateValue{identity: &staticIdentity{}}, ok: true}
			}
			return staticEvalResult{}
		}
		return staticEvalResult{value: staticDateValue{milliseconds: staticNumberValue(number), known: true, identity: &staticIdentity{}}, ok: true}
	}
	if !staticNumericArguments(arguments, 0, len(arguments)) {
		return staticEvalResult{}
	}
	return staticEvalResult{value: staticDateValue{identity: &staticIdentity{}}, ok: true}
}

func staticBigIntCall(value any) staticEvalResult {
	if staticValueIsObjectForCoercion(value) {
		// BigInt applies ToPrimitive with the number hint. Date values therefore
		// use valueOf (unlike addition/default-hint coercion) before parsing.
		primitive, ok := staticRelationalPrimitive(value)
		if !ok {
			return staticEvalResult{}
		}
		return staticBigIntCall(primitive)
	}
	switch value := value.(type) {
	case staticBigIntValue:
		return staticEvalResult{value: value, ok: true}
	case bool:
		integer := int64(0)
		if value {
			integer = 1
		}
		return staticEvalResult{value: staticBigIntValue{value: big.NewInt(integer)}, ok: true}
	case staticNumberValue:
		number := float64(value)
		if math.IsNaN(number) || math.IsInf(number, 0) || math.Trunc(number) != number {
			return staticEvalResult{}
		}
		integer, _ := new(big.Float).SetFloat64(number).Int(nil)
		return staticEvalResult{value: staticBigIntValue{value: integer}, ok: true}
	case string:
		return staticBigIntFromString(value)
	case *staticStringNode:
		text, _ := staticValueAsString(value)
		return staticBigIntFromString(text)
	default:
		if staticValueKindOf(value) != staticKindNumber {
			return staticEvalResult{}
		}
		number, ok := staticValueToNumber(value)
		if !ok || math.IsNaN(number) || math.IsInf(number, 0) || math.Trunc(number) != number {
			return staticEvalResult{}
		}
		integer, _ := new(big.Float).SetFloat64(number).Int(nil)
		return staticEvalResult{value: staticBigIntValue{value: integer}, ok: true}
	}
}

func staticBigIntFromString(text string) staticEvalResult {
	text = ecmascript.StringTrim(text)
	if text == "" {
		return staticEvalResult{value: staticBigIntValue{value: new(big.Int)}, ok: true}
	}
	base := 10
	digits := text
	if len(text) > 2 && text[0] == '0' {
		switch text[1] {
		case 'b', 'B':
			base, digits = 2, text[2:]
		case 'o', 'O':
			base, digits = 8, text[2:]
		case 'x', 'X':
			base, digits = 16, text[2:]
		}
	}
	value := new(big.Int)
	if _, ok := value.SetString(digits, base); !ok {
		return staticEvalResult{}
	}
	if value.BitLen() > maxStaticBigIntBits {
		return staticEvalResult{
			value: staticAbstractBigInt(value.Sign() != 0), ok: true,
			bigIntWorkBytes: (value.BitLen() + 7) / 8,
		}
	}
	return staticEvalResult{
		value: staticBigIntValue{value: value}, ok: true,
		bigIntWorkBytes: (value.BitLen() + 7) / 8,
	}
}

func staticObjectCall(arguments []any) staticEvalResult {
	value := staticArgument(arguments, 0)
	if staticValueNullish(value) {
		return staticEvalResult{value: &staticObjectValue{enhancedCoercion: true}, ok: true}
	}
	switch value.(type) {
	case *staticObjectValue, *staticArrayValue, staticOpaqueObjectValue,
		staticRegExpValue, staticDateValue, staticBoxedValue, staticCollectionValue, staticIteratorValue,
		staticBuiltinValue, staticBuiltinObjectValue:
		return staticEvalResult{value: value, ok: true}
	default:
		return staticEvalResult{value: staticBoxedValue{value: value, identity: &staticIdentity{}}, ok: true}
	}
}

func staticCollectionCall(
	work *staticStringWorkContext,
	kind string,
	arguments []any,
) staticEvalResult {
	collection := staticCollectionValue{
		kind: kind, identity: &staticIdentity{}, stringWorkContext: work,
	}
	if len(arguments) == 0 || staticValueNullish(arguments[0]) {
		return staticEvalResult{value: collection, ok: true}
	}
	iterable, ok := staticIterableValues(work, arguments[0], maxStaticAggregateElements)
	if !ok {
		return staticEvalResult{}
	}
	if !work.reserve(max(len(iterable)*64, 1)) {
		return staticEvalResult{}
	}
	for _, element := range iterable {
		if kind == "Set" {
			if !staticCollectionStore(&collection, element, nil) {
				return staticEvalResult{}
			}
			continue
		}
		_, builtinObject := element.(staticBuiltinValue)
		if !staticValueIsAggregate(element) && !builtinObject {
			return staticEvalResult{}
		}
		key := staticESLintGetMemberValue(element, "0")
		value := staticESLintGetMemberValue(element, "1")
		if !key.ok || !value.ok {
			return staticEvalResult{}
		}
		if !staticCollectionStore(&collection, key.value, value.value) {
			return staticEvalResult{}
		}
	}
	return staticEvalResult{value: collection, ok: true}
}

func staticESLintGetMemberValue(object any, key string) staticEvalResult {
	if result := staticESLintMemberValue(object, key); result.ok {
		return result
	}
	return staticESLintBuiltinMember(object, key)
}

func staticIterableValues(work *staticStringWorkContext, value any, limit int) ([]any, bool) {
	if limit < 0 {
		return nil, false
	}
	switch value := value.(type) {
	case *staticArrayValue:
		if value.length > limit || !work.reserve(max(value.length*16, 1)) {
			return nil, false
		}
		values := make([]any, value.length)
		for index := range value.length {
			if value.omitted[index] {
				values[index] = staticUndefinedValue{}
			} else {
				values[index] = value.element(index)
			}
		}
		return values, true
	case string:
		return staticStringIteratorValues(work, value, limit)
	case *staticStringNode:
		text, _ := staticValueAsString(value)
		return staticStringIteratorValues(work, text, limit)
	case staticBoxedValue:
		if text, ok := staticValueAsString(value.value); ok {
			return staticStringIteratorValues(work, text, limit)
		}
	case staticIteratorValue:
		if len(value.values) > limit || !work.reserve(max(len(value.values)*16, 1)) {
			return nil, false
		}
		return value.values, true
	case staticCollectionValue:
		if len(value.entries) > limit || !work.reserve(max(len(value.entries)*48, 1)) {
			return nil, false
		}
		values := make([]any, 0, len(value.entries))
		for _, entry := range value.entries {
			if value.kind == "Map" {
				values = append(values, staticArrayFromValues([]any{entry.key, entry.value}))
			} else {
				values = append(values, entry.key)
			}
		}
		return values, true
	case staticBuiltinObjectValue:
		if value == "Array.prototype" {
			return []any{}, true
		}
	}
	return nil, false
}

func staticCollectionStore(collection *staticCollectionValue, key, value any) bool {
	if canonical, ok := staticCanonicalCollectionKey(key, collection.stringWorkContext); ok {
		if collection.index == nil {
			collection.index = make(map[staticCollectionKey]int, 8)
		}
		if index, exists := collection.index[canonical]; exists {
			if collection.kind == "Map" {
				collection.entries[index].value = value
			}
			return true
		}
		collection.index[canonical] = len(collection.entries)
		collection.entries = append(collection.entries, staticCollectionEntry{key: key, value: value})
		return true
	}
	if collection.stringWorkContext.exhausted() {
		return false
	}
	for index, entry := range collection.entries {
		equal, known := staticSameValueZero(entry.key, key)
		if !known {
			return false
		}
		if equal {
			if collection.kind == "Map" {
				collection.entries[index].value = value
			}
			return true
		}
	}
	collection.entries = append(collection.entries, staticCollectionEntry{key: key, value: value})
	return true
}

func staticCanonicalCollectionKey(
	value any,
	work *staticStringWorkContext,
) (staticCollectionKey, bool) {
	switch value := value.(type) {
	case staticNullValue:
		return staticCollectionKey{kind: 1}, true
	case staticUndefinedValue, staticOptionalChainShortCircuitValue:
		return staticCollectionKey{kind: 2}, true
	case bool:
		if value {
			return staticCollectionKey{kind: 3, numberBits: 1}, true
		}
		return staticCollectionKey{kind: 3}, true
	case string:
		if !work.reserve(max(len(value), 1)) {
			return staticCollectionKey{}, false
		}
		return staticCollectionKey{kind: 4, text: value}, true
	case *staticStringNode:
		text, ok := staticValueAsString(value)
		if ok && !work.reserve(max(len(text), 1)) {
			return staticCollectionKey{}, false
		}
		return staticCollectionKey{kind: 4, text: text}, ok
	case staticNumberValue:
		number := float64(value)
		bits := math.Float64bits(number)
		if number == 0 {
			bits = 0
		} else if math.IsNaN(number) {
			bits = math.Float64bits(math.NaN())
		}
		return staticCollectionKey{kind: 5, numberBits: bits}, true
	case staticBigIntValue:
		if value.value == nil {
			return staticCollectionKey{}, false
		}
		byteLength := (value.value.BitLen() + 7) / 8
		// big.Int.Bytes and the conversion to a comparable string each allocate
		// the magnitude once, and the map then hashes it. Charge all three passes
		// before materializing the key so a compact fanout cannot amplify them.
		if !work.reserve(max(byteLength*3, 1)) {
			return staticCollectionKey{}, false
		}
		sign := uint64(0)
		if value.value.Sign() < 0 {
			sign = 1
		}
		return staticCollectionKey{
			kind: 6, numberBits: sign, text: string(value.value.Bytes()),
		}, true
	case staticSymbolValue:
		if value.hostDependent || !value.global && !value.wellKnown {
			return staticCollectionKey{}, false
		}
		if !work.reserve(max(len(value.description), 1)) {
			return staticCollectionKey{}, false
		}
		kind := uint8(7)
		if value.wellKnown {
			kind = 8
		}
		return staticCollectionKey{kind: kind, text: value.description}, true
	case staticBuiltinValue:
		name := string(value)
		if staticBuiltinIsFunction(name) {
			name = staticCanonicalBuiltinFunctionName(name)
		}
		return staticCollectionKey{kind: 9, text: name}, true
	case staticBuiltinObjectValue:
		return staticCollectionKey{kind: 10, text: string(value)}, true
	case *staticObjectValue:
		return staticCollectionKey{kind: 11, identity: value}, true
	case *staticArrayValue:
		return staticCollectionKey{kind: 12, identity: value}, true
	case staticRegExpValue:
		if value.identity != nil {
			return staticCollectionKey{kind: 13, identity: value.identity}, true
		}
	case staticDateValue:
		if value.identity != nil {
			return staticCollectionKey{kind: 14, identity: value.identity}, true
		}
	case staticBoxedValue:
		if value.identity != nil {
			return staticCollectionKey{kind: 15, identity: value.identity}, true
		}
	case staticCollectionValue:
		if value.identity != nil {
			return staticCollectionKey{kind: 16, identity: value.identity}, true
		}
	case staticOpaqueObjectValue:
		if value.identity != nil {
			return staticCollectionKey{kind: 17, identity: value.identity}, true
		}
	case staticIteratorValue:
		if value.identity != nil {
			return staticCollectionKey{kind: 18, identity: value.identity}, true
		}
	}
	if staticValueKindOf(value) == staticKindNumber {
		number, ok := staticValueToNumber(value)
		if !ok {
			return staticCollectionKey{}, false
		}
		bits := math.Float64bits(number)
		if number == 0 {
			bits = 0
		} else if math.IsNaN(number) {
			bits = math.Float64bits(math.NaN())
		}
		return staticCollectionKey{kind: 5, numberBits: bits}, true
	}
	return staticCollectionKey{}, false
}

func staticStringIteratorValues(work *staticStringWorkContext, text string, limit int) ([]any, bool) {
	unitCount := ecmascript.StringCodeUnitCount(text)
	if unitCount > limit*2 ||
		!work.reserve(max(len(text)*2+min(unitCount, limit)*32, 1)) {
		return nil, false
	}
	units := ecmascript.StringCodeUnits(text)
	values := make([]any, 0, min(len(units), limit))
	for index := 0; index < len(units); index++ {
		if len(values) >= limit {
			return nil, false
		}
		end := index + 1
		if units[index] >= 0xD800 && units[index] <= 0xDBFF && end < len(units) &&
			units[end] >= 0xDC00 && units[end] <= 0xDFFF {
			end++
		}
		values = append(values, ecmascript.StringFromCodeUnits(units[index:end]))
		index = end - 1
	}
	return values, true
}

func staticCollectionHas(collection staticCollectionValue, key any) staticEvalResult {
	if canonical, ok := staticCanonicalCollectionKey(key, collection.stringWorkContext); ok {
		if collection.index != nil {
			_, exists := collection.index[canonical]
			return staticEvalResult{value: exists, ok: true}
		}
	} else if collection.stringWorkContext.exhausted() {
		return staticEvalResult{value: staticUnknownBooleanValue{}, ok: true}
	}
	for _, entry := range collection.entries {
		equal, known := staticSameValueZero(entry.key, key)
		if !known {
			return staticEvalResult{value: staticUnknownBooleanValue{}, ok: true}
		}
		if equal {
			return staticEvalResult{value: true, ok: true}
		}
	}
	return staticEvalResult{value: false, ok: true}
}

func staticMapGet(collection staticCollectionValue, key any) staticEvalResult {
	if canonical, ok := staticCanonicalCollectionKey(key, collection.stringWorkContext); ok {
		if collection.index != nil {
			if index, exists := collection.index[canonical]; exists {
				return staticEvalResult{value: collection.entries[index].value, ok: true}
			}
			return staticEvalResult{value: staticUndefinedValue{}, ok: true}
		}
	} else if collection.stringWorkContext.exhausted() {
		return staticEvalResult{}
	}
	for _, entry := range collection.entries {
		equal, known := staticSameValueZero(entry.key, key)
		if !known {
			return staticEvalResult{}
		}
		if equal {
			return staticEvalResult{value: entry.value, ok: true}
		}
	}
	return staticEvalResult{value: staticUndefinedValue{}, ok: true}
}

func staticRegExpCall(
	work *staticStringWorkContext,
	arguments []any,
	construct bool,
) staticEvalResult {
	pattern := ""
	flags := ""
	if len(arguments) > 0 && !staticValueUndefined(arguments[0]) {
		switch argument := arguments[0].(type) {
		case staticRegExpValue:
			if !construct && (len(arguments) < 2 || staticValueUndefined(arguments[1])) {
				return staticEvalResult{value: argument, ok: true}
			}
			pattern = argument.source
			if len(arguments) < 2 || staticValueUndefined(arguments[1]) {
				flags = argument.flags
			}
		case *staticObjectValue:
			isRegExp, known := staticRegExpObjectMatch(argument)
			if !known {
				return staticEvalResult{}
			}
			if !isRegExp {
				var ok bool
				pattern, ok = staticValueToString(argument)
				if !ok {
					return staticEvalResult{}
				}
				break
			}
			if !construct && (len(arguments) < 2 || staticValueUndefined(arguments[1])) {
				constructor := staticESLintGetMemberValue(argument, "constructor")
				if !constructor.ok {
					return staticEvalResult{}
				}
				equal, equalKnown := staticValuesStrictEqual(constructor.value, staticBuiltinValue("RegExp"))
				if !equalKnown {
					return staticEvalResult{}
				}
				if equal {
					return staticEvalResult{value: argument, ok: true}
				}
			}
			var ok bool
			pattern, ok = staticRegExpObjectTextProperty(argument, "source")
			if !ok {
				return staticEvalResult{}
			}
			if len(arguments) < 2 || staticValueUndefined(arguments[1]) {
				flags, ok = staticRegExpObjectTextProperty(argument, "flags")
				if !ok {
					return staticEvalResult{}
				}
			}
		case staticBuiltinObjectValue:
			if argument == "RegExp.prototype" {
				if !construct && (len(arguments) < 2 || staticValueUndefined(arguments[1])) {
					return staticEvalResult{value: argument, ok: true}
				}
				pattern = "(?:)"
				break
			}
			var ok bool
			pattern, ok = staticValueToString(argument)
			if !ok {
				return staticEvalResult{}
			}
		default:
			var ok bool
			pattern, ok = staticValueToString(argument)
			if !ok {
				return staticEvalResult{}
			}
		}
	}
	if len(arguments) > 1 && !staticValueUndefined(arguments[1]) {
		var ok bool
		flags, ok = staticValueToString(arguments[1])
		if !ok {
			return staticEvalResult{}
		}
	}
	if !work.reserve(max((len(pattern)+len(flags))*2, 1)) {
		return staticEvalResult{}
	}
	if !staticRegExpConstructorSourceValid(pattern, flags) {
		return staticEvalResult{}
	}
	source, ok := staticRegExpSource(work, pattern)
	if !ok || !staticRegExpConstructorPatternValid(source, flags) {
		return staticEvalResult{}
	}
	return staticEvalResult{value: staticRegExpValue{
		source: source, flags: canonicalRegExpFlags(flags), identity: &staticIdentity{},
	}, ok: true}
}

func staticRegExpObjectMatch(object *staticObjectValue) (bool, bool) {
	match := staticESLintSymbolMemberValue(object, staticSymbolValue{description: "match", wellKnown: true})
	if !match.ok {
		return false, false
	}
	if staticValueUndefined(match.value) {
		return false, true
	}
	return staticValueTruthy(match.value)
}

func staticRegExpObjectTextProperty(object *staticObjectValue, key string) (string, bool) {
	property := staticESLintGetMemberValue(object, key)
	if !property.ok {
		return "", false
	}
	if staticValueUndefined(property.value) {
		return "", true
	}
	return staticValueToString(property.value)
}

// A raw slash is ordinary constructor text in every grammar except a v-mode
// character class, where it is a reserved punctuator and must be escaped.
// Turning constructor text into a slash-delimited literal necessarily escapes
// the delimiter, so this early error has to be checked before that conversion.
// The two Hrkt script aliases are also a grammar-version boundary: tsgo's table
// accepts them while every Node version supported by Unicorn 73 rejects them.
func staticRegExpConstructorSourceValid(pattern, flags string) bool {
	unicodeSets := strings.Contains(flags, "v")
	unicodeMode := unicodeSets || strings.Contains(flags, "u")
	if unicodeMode && staticRegExpHasEscapedLineTerminator(pattern) {
		return false
	}
	inClass := 0
	inClassString := false
	captures := 0
	for index := 0; index < len(pattern); {
		if pattern[index] == '\\' {
			if unicodeMode && index+2 < len(pattern) &&
				(pattern[index+1] == 'p' || pattern[index+1] == 'P') &&
				pattern[index+2] == '{' {
				expressionStart := index + 3
				expressionEnd, ok := staticRegExpPropertyExpressionEnd(pattern, expressionStart)
				if !ok {
					return false
				}
				if staticRegExpNode22RejectsProperty(pattern[expressionStart:expressionEnd]) {
					return false
				}
				index = expressionEnd + 1
				continue
			}
			if unicodeSets && inClass > 0 && !inClassString && index+2 < len(pattern) &&
				pattern[index+1] == 'q' && pattern[index+2] == '{' {
				inClassString = true
				index += 3
				continue
			}
			index += 1 + staticRegExpNextCharacterSize(pattern[index+1:])
			continue
		}
		if inClassString {
			if pattern[index] == '/' {
				return false
			}
			if pattern[index] == '}' {
				inClassString = false
			}
			index += staticRegExpNextCharacterSize(pattern[index:])
			continue
		}
		switch pattern[index] {
		case '[':
			if unicodeSets {
				inClass++
				if inClass > maxStaticRegExpUnicodeSetsNesting {
					return false
				}
			} else if inClass == 0 {
				inClass = 1
			}
		case ']':
			if inClass > 0 {
				inClass--
			}
		case '/':
			if unicodeSets && inClass > 0 {
				return false
			}
		case '(':
			if inClass == 0 && (index+1 >= len(pattern) || pattern[index+1] != '?' ||
				index+3 < len(pattern) && pattern[index+2] == '<' &&
					pattern[index+3] != '=' && pattern[index+3] != '!') {
				captures++
				if captures > maxStaticRegExpCaptures {
					return false
				}
			}
		}
		index += staticRegExpNextCharacterSize(pattern[index:])
	}
	return true
}

// A Unicode property expression cannot contain another escape or opening
// brace. Rejecting those delimiters while looking for the closing brace keeps
// a compact sequence such as `\\p{\\p{...` linear instead of searching the
// same suffix once per escape. The regexp parser remains the authority for the
// property name and value grammar.
func staticRegExpPropertyExpressionEnd(source string, start int) (int, bool) {
	for index := start; index < len(source); index++ {
		switch source[index] {
		case '}':
			return index, index > start
		case '\\', '{':
			return 0, false
		}
	}
	return 0, false
}

func staticRegExpNode22RejectsProperty(expression string) bool {
	separator := strings.IndexByte(expression, '=')
	if separator < 0 {
		return false
	}
	name, value := expression[:separator], expression[separator+1:]
	if value != "Hrkt" && value != "Katakana_Or_Hiragana" {
		return false
	}
	return name == "Script" || name == "sc" || name == "Script_Extensions" || name == "scx"
}

// staticRegExpConstructorPatternValid asks tsgo's ECMAScript regexp parser to
// validate the constructor pattern after staticRegExpSource has made it safe
// to place between literal delimiters. The regexp package intentionally accepts
// a wider grammar in a few places and therefore cannot be used as a constructor
// validity oracle here: treating a native SyntaxError as a RegExp object could
// make receiver classification suppress a real diagnostic.
//
// tsgo implements the latest grammar. Node 22, the oldest runtime supported by
// Unicorn 73, predates inline modifiers and the relaxation for duplicate named
// captures in disjoint alternatives. Those two newer forms stay abstract so a
// result is safe for every supported upstream runtime.
func staticRegExpConstructorPatternValid(source, flags string) bool {
	unicodeSets := strings.Contains(flags, "v")
	source = staticRegExpNormalizeGroupNames(source, unicodeSets)
	namedGroups, compatible := staticRegExpNode22NamedGroupProbe(source, unicodeSets)
	if !compatible {
		return false
	}
	validatorSource := staticRegExpValidatorSource(source, flags, namedGroups != "")
	return ecmascript.IsValidRegexLiteral("/"+validatorSource+"/"+flags) &&
		(namedGroups == "" || ecmascript.IsValidRegexLiteral("/"+namedGroups+"/"+flags))
}

// tsgo's regexp parser validates decoded group names correctly but currently
// rejects their source-level Unicode escape spelling. Normalize only complete
// RegExpIdentifierNames in captures and backreferences; invalid escapes or
// decoded identifier characters are left for the validator to reject.
func staticRegExpNormalizeGroupNames(source string, unicodeSets bool) string {
	var builder strings.Builder
	lastWritten := 0
	changed := false
	replaceName := func(start, end int) {
		normalized, ok := staticRegExpNormalizeGroupName(source[start:end])
		if !ok || normalized == source[start:end] {
			return
		}
		if !changed {
			builder.Grow(len(source))
			changed = true
		}
		builder.WriteString(source[lastWritten:start])
		builder.WriteString(normalized)
		lastWritten = end
	}

	inClass := 0
	inClassString := false
	for index := 0; index < len(source); {
		if source[index] == '\\' {
			if inClass == 0 && index+3 < len(source) && source[index+1] == 'k' && source[index+2] == '<' {
				closingOffset := strings.IndexByte(source[index+3:], '>')
				if closingOffset < 0 {
					break
				}
				closing := index + 3 + closingOffset
				replaceName(index+3, closing)
				index = closing + 1
				continue
			}
			if unicodeSets && inClass > 0 && !inClassString && index+2 < len(source) &&
				source[index+1] == 'q' && source[index+2] == '{' {
				inClassString = true
				index += 3
				continue
			}
			index += 1 + staticRegExpNextCharacterSize(source[index+1:])
			continue
		}
		if inClassString {
			if source[index] == '}' {
				inClassString = false
			}
			index += staticRegExpNextCharacterSize(source[index:])
			continue
		}
		switch source[index] {
		case '[':
			if unicodeSets {
				inClass++
			} else if inClass == 0 {
				inClass = 1
			}
			index++
			continue
		case ']':
			if inClass > 0 {
				inClass--
			}
			index++
			continue
		}
		if inClass == 0 && index+3 < len(source) && source[index] == '(' && source[index+1] == '?' &&
			source[index+2] == '<' && source[index+3] != '=' && source[index+3] != '!' {
			closingOffset := strings.IndexByte(source[index+3:], '>')
			if closingOffset < 0 {
				break
			}
			closing := index + 3 + closingOffset
			replaceName(index+3, closing)
			index = closing + 1
			continue
		}
		index += staticRegExpNextCharacterSize(source[index:])
	}
	if !changed {
		return source
	}
	builder.WriteString(source[lastWritten:])
	return builder.String()
}

func staticRegExpNormalizeGroupName(name string) (string, bool) {
	if name == "" {
		return name, false
	}
	var builder strings.Builder
	builder.Grow(len(name))
	first := true
	for index := 0; index < len(name); {
		character, size := utf8.DecodeRuneInString(name[index:])
		if name[index] == '\\' {
			var ok bool
			character, size, ok = staticRegExpGroupNameEscape(name[index:])
			if !ok {
				return name, false
			}
		}
		valid := character == '$' || character == '_' || scanner.IsIdentifierPart(character)
		if first {
			valid = character == '$' || character == '_' || scanner.IsIdentifierStart(character)
		}
		if character == utf8.RuneError || !valid {
			return name, false
		}
		builder.WriteRune(character)
		first = false
		index += size
	}
	return builder.String(), true
}

func staticRegExpGroupNameEscape(text string) (rune, int, bool) {
	if len(text) < 2 || text[0] != '\\' || text[1] != 'u' {
		return 0, 0, false
	}
	if len(text) >= 4 && text[2] == '{' {
		closing := strings.IndexByte(text[3:], '}')
		if closing < 1 || closing > 6 {
			return 0, 0, false
		}
		value, err := strconv.ParseUint(text[3:3+closing], 16, 32)
		if err != nil || value > utf8.MaxRune {
			return 0, 0, false
		}
		return rune(value), closing + 4, true
	}
	if len(text) < 6 {
		return 0, 0, false
	}
	value, err := strconv.ParseUint(text[2:6], 16, 16)
	if err != nil {
		return 0, 0, false
	}
	return rune(value), 6, true
}

// staticRegExpValidatorSource bridges three conservative gaps in tsgo's
// otherwise-strict parser without changing the RegExp value we materialize:
// Annex B identity escapes, Unicode escapes in group names (handled above),
// and seven Unicode 16 script values supported by Node 22 but absent from
// tsgo's Unicode 15.1 property table. In a legacy character class, a malformed
// braced Unicode escape is also the identity escape "u" followed by ordinary
// class characters; spelling that identity escape without its backslash keeps
// tsgo from treating it as a Unicode escape when the pattern has named groups.
func staticRegExpValidatorSource(source, flags string, hasNamedGroups bool) string {
	unicodeMode := strings.ContainsAny(flags, "uv")
	unicodeSets := strings.Contains(flags, "v")
	var builder strings.Builder
	lastWritten := 0
	changed := false
	replace := func(start, end int, replacement string) {
		if !changed {
			builder.Grow(len(source))
			changed = true
		}
		builder.WriteString(source[lastWritten:start])
		builder.WriteString(replacement)
		lastWritten = end
	}

	inClass := 0
	inClassString := false
	for index := 0; index < len(source); {
		if source[index] == '\\' {
			if unicodeSets && inClass > 0 && !inClassString && index+2 < len(source) &&
				source[index+1] == 'q' && source[index+2] == '{' {
				inClassString = true
				index += 3
				continue
			}
			if !inClassString && index+2 < len(source) &&
				(source[index+1] == 'p' || source[index+1] == 'P') && source[index+2] == '{' {
				if !unicodeMode {
					replace(index, index+1, "")
					index += 2
					continue
				}
				expressionStart := index + 3
				expressionEnd, ok := staticRegExpPropertyExpressionEnd(source, expressionStart)
				if !ok {
					return source
				}
				if replacement, replaceProperty := staticRegExpUnicode16PropertyReplacement(
					source[expressionStart:expressionEnd],
				); replaceProperty {
					replace(expressionStart, expressionEnd, replacement)
				}
				index = expressionEnd + 1
				continue
			}
			if !unicodeMode && inClass > 0 && index+2 < len(source) &&
				source[index+1] == 'u' && source[index+2] == '{' {
				replace(index, index+1, "")
				index += 2
				continue
			}
			if !unicodeMode && !hasNamedGroups && inClass == 0 && index+2 < len(source) &&
				source[index+1] == 'k' && source[index+2] == '<' {
				replace(index, index+1, "")
				index += 2
				continue
			}
			index += 1 + staticRegExpNextCharacterSize(source[index+1:])
			continue
		}
		if inClassString {
			if source[index] == '}' {
				inClassString = false
			}
			index += staticRegExpNextCharacterSize(source[index:])
			continue
		}
		switch source[index] {
		case '[':
			if unicodeSets {
				inClass++
			} else if inClass == 0 {
				inClass = 1
			}
		case ']':
			if inClass > 0 {
				inClass--
			}
		}
		index += staticRegExpNextCharacterSize(source[index:])
	}
	if !changed {
		return source
	}
	builder.WriteString(source[lastWritten:])
	return builder.String()
}

func staticRegExpUnicode16PropertyReplacement(expression string) (string, bool) {
	separator := strings.IndexByte(expression, '=')
	if separator < 0 {
		return "", false
	}
	name, value := expression[:separator], expression[separator+1:]
	if name != "Script" && name != "sc" && name != "Script_Extensions" && name != "scx" {
		return "", false
	}
	switch value {
	case "Garay", "Gurung_Khema", "Kirat_Rai", "Ol_Onal", "Sunuwar", "Todhri", "Tulu_Tigalari":
		return name + "=Greek", true
	}
	return "", false
}

// staticRegExpNode22NamedGroupProbe extracts named-capture openers into one
// flat pattern. Validating that pattern makes tsgo reject every duplicate after
// decoding Unicode escapes in the names, including duplicates that its latest
// grammar otherwise permits in different alternatives or nested scopes.
func staticRegExpNode22NamedGroupProbe(source string, unicodeSets bool) (string, bool) {
	var namedGroups strings.Builder
	inClass := 0
	inClassString := false
	for index := 0; index < len(source); {
		if source[index] == '\\' {
			if unicodeSets && inClass > 0 && !inClassString && index+2 < len(source) &&
				source[index+1] == 'q' && source[index+2] == '{' {
				inClassString = true
				index += 3
				continue
			}
			index += 1 + staticRegExpNextCharacterSize(source[index+1:])
			continue
		}
		if inClassString {
			if source[index] == '}' {
				inClassString = false
			}
			index += staticRegExpNextCharacterSize(source[index:])
			continue
		}
		switch source[index] {
		case '[':
			if inClass == 0 || unicodeSets {
				inClass++
			}
			index++
			continue
		case ']':
			if inClass > 0 {
				inClass--
			}
			index++
			continue
		}
		if inClass == 0 && index+2 < len(source) && source[index] == '(' && source[index+1] == '?' {
			switch source[index+2] {
			case 'i', 'm', 's', '-':
				return "", false
			case '<':
				if index+3 < len(source) && source[index+3] != '=' && source[index+3] != '!' {
					closingOffset := strings.IndexByte(source[index+3:], '>')
					if closingOffset < 0 {
						return "", false
					}
					closing := index + 3 + closingOffset
					namedGroups.WriteString(source[index : closing+1])
					namedGroups.WriteByte(')')
					index = closing + 1
					continue
				}
			}
		}
		index += staticRegExpNextCharacterSize(source[index:])
	}
	return namedGroups.String(), true
}

func staticRegExpNextCharacterSize(text string) int {
	_, size := utf8.DecodeRuneInString(text)
	return max(size, 1)
}

func staticRegExpHasEscapedLineTerminator(pattern string) bool {
	for index := 0; index < len(pattern); {
		r, size := utf8.DecodeRuneInString(pattern[index:])
		if ecmascript.IsLineTerminator(r) && staticRegExpPrecedingBackslashes(pattern, index)%2 != 0 {
			return true
		}
		index += max(size, 1)
	}
	return false
}

// Without u/v, Annex B treats a backslash immediately before a raw line
// terminator as a legacy identity escape and RegExp.prototype.source omits that
// unmatched backslash. Constructor validation has already rejected this shape
// in Unicode modes before the source is canonicalized.
func staticRegExpNormalizeAnnexBLineEscapes(pattern string) string {
	var builder strings.Builder
	lastWritten := 0
	changed := false
	for index := 0; index < len(pattern); {
		r, size := utf8.DecodeRuneInString(pattern[index:])
		size = max(size, 1)
		if ecmascript.IsLineTerminator(r) && staticRegExpPrecedingBackslashes(pattern, index)%2 != 0 {
			if !changed {
				builder.Grow(len(pattern))
				changed = true
			}
			builder.WriteString(pattern[lastWritten : index-1])
			builder.WriteString(pattern[index : index+size])
			lastWritten = index + size
		}
		index += size
	}
	if !changed {
		return pattern
	}
	builder.WriteString(pattern[lastWritten:])
	return builder.String()
}

func staticRegExpSource(work *staticStringWorkContext, pattern string) (string, bool) {
	if pattern == "" {
		return "(?:)", true
	}
	pattern = staticRegExpNormalizeAnnexBLineEscapes(pattern)
	if !work.reserve(max(len(pattern)*6, 1)) {
		return "", false
	}
	var builder strings.Builder
	appendText := func(text string) bool {
		if len(text) > maxStaticStringLength-builder.Len() {
			work.exhaust()
			return false
		}
		builder.WriteString(text)
		return true
	}
	for index := 0; index < len(pattern); {
		switch pattern[index] {
		case '\n':
			if !appendText(`\n`) {
				return "", false
			}
			index++
		case '\r':
			if !appendText(`\r`) {
				return "", false
			}
			index++
		case '/':
			backslashes := 0
			for previous := index - 1; previous >= 0 && pattern[previous] == '\\'; previous-- {
				backslashes++
			}
			if backslashes%2 == 0 {
				if !appendText(`\`) {
					return "", false
				}
			}
			if !appendText("/") {
				return "", false
			}
			index++
		default:
			r, size := utf8.DecodeRuneInString(pattern[index:])
			switch r {
			case '\u2028':
				if !appendText(`\u2028`) {
					return "", false
				}
			case '\u2029':
				if !appendText(`\u2029`) {
					return "", false
				}
			default:
				if !appendText(pattern[index : index+size]) {
					return "", false
				}
			}
			index += size
		}
	}
	return builder.String(), true
}

func staticRegExpPrecedingBackslashes(pattern string, index int) int {
	count := 0
	for index--; index >= 0 && pattern[index] == '\\'; index-- {
		count++
	}
	return count
}

func staticParseNumber(name string, arguments []any) staticEvalResult {
	text, ok := staticValueToString(staticArgument(arguments, 0))
	if !ok {
		return staticEvalResult{}
	}
	if name == "Number.parseInt" || name == "parseInt" {
		radix := 0
		if len(arguments) > 1 && !staticValueUndefined(arguments[1]) {
			number, ok := staticValueToNumber(arguments[1])
			if !ok {
				if staticCanConvertToNumber(arguments[1]) {
					return staticEvalResult{value: staticUnknownNumberValue{}, ok: true}
				}
				return staticEvalResult{}
			}
			radix = int(toInt32(number))
		}
		return staticParseInt(text, radix)
	}
	return staticParseFloat(text)
}

func staticParseInt(text string, radix int) staticEvalResult {
	text = ecmascript.StringTrim(text)
	negative := false
	if text != "" && (text[0] == '+' || text[0] == '-') {
		negative = text[0] == '-'
		text = text[1:]
	}
	if radix != 0 && (radix < 2 || radix > 36) {
		return staticEvalResult{value: staticNumberValue(math.NaN()), ok: true}
	}
	stripHexPrefix := radix == 0 || radix == 16
	if radix == 0 {
		radix = 10
	}
	if stripHexPrefix && len(text) >= 2 && text[0] == '0' && (text[1] == 'x' || text[1] == 'X') {
		radix = 16
		text = text[2:]
	}
	digitCount := 0
	for digitCount < len(text) {
		digit := asciiRadixDigit(text[digitCount])
		if digit < 0 || digit >= radix {
			break
		}
		digitCount++
	}
	if digitCount == 0 {
		return staticEvalResult{value: staticNumberValue(math.NaN()), ok: true}
	}
	integer := new(big.Int)
	integer.SetString(text[:digitCount], radix)
	value, _ := new(big.Float).SetInt(integer).Float64()
	if negative {
		value = -value
		if value == 0 {
			value = math.Copysign(0, -1)
		}
	}
	return staticEvalResult{value: staticNumberValue(value), ok: true}
}

func asciiRadixDigit(value byte) int {
	switch {
	case value >= '0' && value <= '9':
		return int(value - '0')
	case value >= 'a' && value <= 'z':
		return int(value-'a') + 10
	case value >= 'A' && value <= 'Z':
		return int(value-'A') + 10
	default:
		return -1
	}
}

func staticParseFloat(text string) staticEvalResult {
	text = ecmascript.StringTrim(text)
	if strings.HasPrefix(text, "+Infinity") || strings.HasPrefix(text, "Infinity") {
		return staticEvalResult{value: staticNumberValue(math.Inf(1)), ok: true}
	}
	if strings.HasPrefix(text, "-Infinity") {
		return staticEvalResult{value: staticNumberValue(math.Inf(-1)), ok: true}
	}
	length := staticDecimalPrefixLength(text)
	if length == 0 || length == 1 && (text[0] == '+' || text[0] == '-') {
		return staticEvalResult{value: staticNumberValue(math.NaN()), ok: true}
	}
	value, err := strconv.ParseFloat(text[:length], 64)
	if err != nil && !strings.Contains(err.Error(), "value out of range") {
		return staticEvalResult{value: staticNumberValue(math.NaN()), ok: true}
	}
	return staticEvalResult{value: staticNumberValue(value), ok: true}
}

func staticDecimalPrefixLength(text string) int {
	index := 0
	if index < len(text) && (text[index] == '+' || text[index] == '-') {
		index++
	}
	integerStart := index
	for index < len(text) && text[index] >= '0' && text[index] <= '9' {
		index++
	}
	hasDigits := index > integerStart
	if index < len(text) && text[index] == '.' {
		index++
		fractionStart := index
		for index < len(text) && text[index] >= '0' && text[index] <= '9' {
			index++
		}
		hasDigits = hasDigits || index > fractionStart
	}
	if !hasDigits {
		return 0
	}
	if index < len(text) && (text[index] == 'e' || text[index] == 'E') {
		exponentMark := index
		index++
		if index < len(text) && (text[index] == '+' || text[index] == '-') {
			index++
		}
		exponentStart := index
		for index < len(text) && text[index] >= '0' && text[index] <= '9' {
			index++
		}
		if index == exponentStart {
			index = exponentMark
		}
	}
	return index
}

func staticDateParse(value any) staticEvalResult {
	text, ok := staticValueToString(value)
	if !ok {
		return staticEvalResult{}
	}
	if milliseconds, matched, exact := staticDateTimeStringMilliseconds(text); matched {
		if !exact {
			return staticEvalResult{value: staticUnknownNumberValue{}, ok: true}
		}
		return staticEvalResult{value: staticNumberValue(milliseconds), ok: true}
	}
	if staticDateHasImplementationDefinedZone(text) {
		return staticEvalResult{value: staticUnknownNumberValue{}, ok: true}
	}
	if staticDateTextDefinitelyInvalid(text) {
		return staticEvalResult{value: staticNumberValue(math.NaN()), ok: true}
	}
	// Date.parse accepts implementation-defined legacy formats and normalizes
	// some out-of-range dates. A failed Go layout match therefore does not prove
	// JavaScript would produce NaN; retain the known number type only.
	return staticEvalResult{value: staticUnknownNumberValue{}, ok: true}
}

// staticDateTimeStringMilliseconds parses ECMAScript's Date Time String
// Format before the implementation-defined legacy formats accepted by a
// particular JavaScript engine. A syntactically matching string with an
// out-of-bounds field is a known NaN; calendar overflow such as February 30 is
// normalized by MakeDay, just as the specification requires. Local date-time
// forms are recognized but deliberately not assigned a millisecond value:
// V8 and Go normalize DST gaps differently, so only date-only, UTC, and
// explicit-offset forms are exact across hosts.
func staticDateTimeStringMilliseconds(text string) (milliseconds float64, matched, exact bool) {
	index := 0
	year, expanded, ok := staticDateYear(text, &index)
	if !ok {
		return 0, false, false
	}
	if expanded && year == 0 && text[0] == '-' {
		return math.NaN(), true, true
	}

	month, day := 1, 1
	hour, minute, second, millisecond := 0, 0, 0, 0
	dateOnly := true
	location := time.UTC
	localTime := false
	if index < len(text) {
		if text[index] != '-' {
			return 0, false, false
		}
		index++
		if month, ok = staticDateTwoDigits(text, &index); !ok {
			return 0, false, false
		}
		if index < len(text) {
			if text[index] != '-' {
				return 0, false, false
			}
			index++
			if day, ok = staticDateTwoDigits(text, &index); !ok {
				return 0, false, false
			}
		}
	}

	if index < len(text) {
		if text[index] != 'T' {
			return 0, false, false
		}
		dateOnly = false
		localTime = true
		index++
		if hour, ok = staticDateTwoDigits(text, &index); !ok || index >= len(text) || text[index] != ':' {
			return 0, false, false
		}
		index++
		if minute, ok = staticDateTwoDigits(text, &index); !ok {
			return 0, false, false
		}
		if index < len(text) && text[index] == ':' {
			index++
			if second, ok = staticDateTwoDigits(text, &index); !ok {
				return 0, false, false
			}
			if index < len(text) && text[index] == '.' {
				index++
				if millisecond, ok = staticDateThreeDigits(text, &index); !ok {
					return 0, false, false
				}
			}
		}

		location = time.Local
		if index < len(text) && text[index] == 'Z' {
			location = time.UTC
			localTime = false
			index++
		} else if index < len(text) && (text[index] == '+' || text[index] == '-') {
			localTime = false
			sign := 1
			if text[index] == '-' {
				sign = -1
			}
			index++
			offsetHour, digitsOK := staticDateTwoDigits(text, &index)
			if !digitsOK || index >= len(text) || text[index] != ':' {
				return 0, false, false
			}
			index++
			offsetMinute, digitsOK := staticDateTwoDigits(text, &index)
			if !digitsOK {
				return 0, false, false
			}
			if offsetHour > 23 || offsetMinute > 59 {
				return math.NaN(), true, true
			}
			location = time.FixedZone("", sign*(offsetHour*60+offsetMinute)*60)
		}
	}

	if index != len(text) {
		return 0, false, false
	}
	if month < 1 || month > 12 || day < 1 || day > 31 ||
		hour < 0 || hour > 24 || minute < 0 || minute > 59 ||
		second < 0 || second > 59 || millisecond < 0 || millisecond > 999 ||
		hour == 24 && (minute != 0 || second != 0 || millisecond != 0) {
		return math.NaN(), true, true
	}
	if dateOnly {
		location = time.UTC
	}
	if localTime {
		return 0, true, false
	}
	parsed := time.Date(year, time.Month(month), day, hour, minute, second, millisecond*int(time.Millisecond), location)
	milliseconds = float64(parsed.UnixMilli())
	if math.Abs(milliseconds) > 8.64e15 {
		return math.NaN(), true, true
	}
	return milliseconds, true, true
}

func staticDateYear(text string, index *int) (year int, expanded bool, ok bool) {
	if *index >= len(text) {
		return 0, false, false
	}
	sign := 1
	digits := 4
	if text[*index] == '+' || text[*index] == '-' {
		expanded = true
		digits = 6
		if text[*index] == '-' {
			sign = -1
		}
		*index++
	}
	start := *index
	if start+digits > len(text) {
		return 0, expanded, false
	}
	for *index < start+digits {
		character := text[*index]
		if character < '0' || character > '9' {
			return 0, expanded, false
		}
		year = year*10 + int(character-'0')
		*index++
	}
	return sign * year, expanded, true
}

func staticDateTwoDigits(text string, index *int) (int, bool) {
	return staticDateDigits(text, index, 2)
}

func staticDateThreeDigits(text string, index *int) (int, bool) {
	return staticDateDigits(text, index, 3)
}

func staticDateDigits(text string, index *int, count int) (int, bool) {
	if *index+count > len(text) {
		return 0, false
	}
	value := 0
	for range count {
		character := text[*index]
		if character < '0' || character > '9' {
			return 0, false
		}
		value = value*10 + int(character-'0')
		*index++
	}
	return value, true
}

func staticDateHasImplementationDefinedZone(text string) bool {
	fields := strings.Fields(text)
	if len(fields) == 0 {
		return false
	}
	zone := fields[len(fields)-1]
	if zone == "GMT" || zone == "UTC" || len(zone) < 2 || len(zone) > 5 {
		return false
	}
	for index := range len(zone) {
		if zone[index] < 'A' || zone[index] > 'Z' {
			return false
		}
	}
	return true
}

func staticDateTextDefinitelyInvalid(text string) bool {
	if staticDateTextHasDigit(text) {
		return false
	}
	lower := ecmascript.StringToLowerCase(text)
	for _, month := range [...]string{
		"jan", "feb", "mar", "apr", "may", "jun", "jul", "aug", "sep", "oct", "nov", "dec",
	} {
		if strings.Contains(lower, month) {
			return false
		}
	}
	return true
}

func staticDateTextHasDigit(text string) bool {
	for index := range len(text) {
		if text[index] >= '0' && text[index] <= '9' {
			return true
		}
	}
	return false
}

func canonicalRegExpFlags(flags string) string {
	var builder strings.Builder
	for _, candidate := range "dgimsuvy" {
		if strings.ContainsRune(flags, candidate) {
			builder.WriteRune(candidate)
		}
	}
	return builder.String()
}

func staticObjectEntriesCall(
	work *staticStringWorkContext,
	name string,
	value any,
) staticEvalResult {
	properties, ok := staticEnumerableOwnProperties(work, value)
	if !ok {
		return staticEvalResult{}
	}
	values := make([]any, 0, len(properties))
	for _, property := range properties {
		switch name {
		case "Object.keys":
			values = append(values, property.name)
		case "Object.values":
			values = append(values, property.value)
		default:
			values = append(values, staticArrayFromValues([]any{property.name, property.value}))
		}
	}
	return staticEvalResult{value: staticArrayFromValues(values), ok: true}
}

func staticEnumerableOwnProperties(
	work *staticStringWorkContext,
	value any,
) ([]staticObjectProperty, bool) {
	switch value := value.(type) {
	case *staticObjectValue:
		if value.propertyCount > maxStaticAggregateElements {
			return nil, false
		}
		if !work.reserve(max(value.propertyCount*64, 1)) {
			return nil, false
		}
		properties := make([]staticObjectProperty, 0, value.propertyCount)
		if value.propertyCount > 0 && !value.property.symbolKey {
			properties = append(properties, value.property)
		}
		for _, property := range value.extraProperties {
			if !property.symbolKey {
				properties = append(properties, property)
			}
		}
		return staticUniqueObjectProperties(properties), true
	case *staticArrayValue:
		if value.length > maxStaticAggregateElements {
			return nil, false
		}
		if !work.reserve(max(value.length*48, 1)) {
			return nil, false
		}
		properties := make([]staticObjectProperty, 0, value.length)
		for index := range value.length {
			if value.omitted[index] {
				continue
			}
			properties = append(properties, staticObjectProperty{
				name: strconv.Itoa(index), value: value.element(index),
			})
		}
		return properties, true
	case string:
		return staticStringEnumerableProperties(work, value)
	case *staticStringNode:
		text, _ := staticValueAsString(value)
		return staticStringEnumerableProperties(work, text)
	case staticBoxedValue:
		if text, ok := staticValueAsString(value.value); ok {
			return staticStringEnumerableProperties(work, text)
		}
		return nil, true
	case staticBuiltinObjectValue:
		if value == "Array.prototype[Symbol.unscopables]" {
			properties := make([]staticObjectProperty, 0, len(staticArrayUnscopablesPropertyNames))
			for _, name := range staticArrayUnscopablesPropertyNames {
				properties = append(properties, staticObjectProperty{name: name, value: true})
			}
			return properties, true
		}
		return nil, true
	case bool, staticNumberValue, staticBigIntValue, staticSymbolValue,
		staticBuiltinValue, staticRegExpValue, staticDateValue,
		staticCollectionValue, staticOpaqueObjectValue:
		return nil, true
	}
	if staticValueKindOf(value) == staticKindNumber {
		return nil, true
	}
	return nil, false
}

var staticArrayUnscopablesPropertyNames = [...]string{
	"at", "copyWithin", "entries", "fill", "find", "findIndex", "findLast", "findLastIndex",
	"flat", "flatMap", "includes", "keys", "toReversed", "toSorted", "toSpliced", "values",
}

func staticUniqueObjectProperties(properties []staticObjectProperty) []staticObjectProperty {
	result := make([]staticObjectProperty, 0, len(properties))
	indices := make(map[string]int, len(properties))
	for _, property := range properties {
		if index, exists := indices[property.name]; exists {
			result[index].value = property.value
			continue
		}
		indices[property.name] = len(result)
		result = append(result, property)
	}
	sort.SliceStable(result, func(left, right int) bool {
		leftIndex, leftOK := staticObjectArrayIndex(result[left].name)
		rightIndex, rightOK := staticObjectArrayIndex(result[right].name)
		if leftOK != rightOK {
			return leftOK
		}
		return leftOK && leftIndex < rightIndex
	})
	return result
}

func staticObjectArrayIndex(value string) (uint32, bool) {
	if value == "" || value == "4294967295" {
		return 0, false
	}
	if len(value) > 1 && value[0] == '0' {
		return 0, false
	}
	parsed, err := strconv.ParseUint(value, 10, 32)
	if err != nil || parsed >= math.MaxUint32 {
		return 0, false
	}
	return uint32(parsed), true
}

func staticStringEnumerableProperties(
	work *staticStringWorkContext,
	text string,
) ([]staticObjectProperty, bool) {
	if ecmascript.StringCodeUnitCount(text) > maxStaticAggregateElements {
		return nil, false
	}
	if !work.reserve(max(len(text)*64, 1)) {
		return nil, false
	}
	units := ecmascript.StringCodeUnits(text)
	properties := make([]staticObjectProperty, len(units))
	for index, unit := range units {
		properties[index] = staticObjectProperty{
			name: strconv.Itoa(index), value: ecmascript.StringFromCodeUnits([]uint16{unit}),
		}
	}
	return properties, true
}

func staticNumberPredicate(name string, value any) staticEvalResult {
	number, ok := staticValueToNumber(value)
	if staticValueKindOf(value) != staticKindNumber {
		return staticEvalResult{value: false, ok: true}
	}
	if !ok {
		return staticEvalResult{value: staticUnknownBooleanValue{}, ok: true}
	}
	if name == "Number.isFinite" {
		return staticEvalResult{value: !math.IsNaN(number) && !math.IsInf(number, 0), ok: true}
	}
	return staticEvalResult{value: math.IsNaN(number), ok: true}
}

func staticStringFromCodePoint(
	work *staticStringWorkContext,
	arguments []any,
) staticEvalResult {
	if !work.reserve(max(len(arguments)*8, 1)) {
		return staticEvalResult{}
	}
	units := make([]uint16, 0, len(arguments))
	for _, argument := range arguments {
		number, ok := staticValueToNumber(argument)
		if !ok || math.Trunc(number) != number || number < 0 || number > utf8.MaxRune {
			return staticEvalResult{}
		}
		if number <= 0xFFFF {
			units = append(units, uint16(number))
			continue
		}
		codePoint := uint32(number) - 0x10000
		units = append(units, uint16(0xD800+codePoint>>10), uint16(0xDC00+codePoint&0x3FF))
	}
	return staticEvalResult{value: ecmascript.StringFromCodeUnits(units), ok: true}
}

func staticStringRawCall(
	work *staticStringWorkContext,
	arguments []any,
) staticEvalResult {
	if len(arguments) == 0 {
		return staticEvalResult{}
	}
	raw := staticESLintMemberValue(arguments[0], "raw")
	if !raw.ok {
		return staticEvalResult{}
	}
	lengthValue := staticESLintMemberValue(raw.value, "length")
	if !lengthValue.ok {
		return staticEvalResult{}
	}
	length, ok := staticToLength(lengthValue.value)
	if !ok || length > maxStaticAggregateElements {
		return staticEvalResult{}
	}
	var builder strings.Builder
	for index := range length {
		segmentValue := staticESLintMemberValue(raw.value, strconv.Itoa(index))
		if !segmentValue.ok {
			return staticEvalResult{}
		}
		segment, ok := staticValueToString(segmentValue.value)
		if !ok || len(segment) > maxStaticStringLength-builder.Len() {
			if ok {
				work.exhaust()
			}
			return staticEvalResult{}
		}
		if !work.reserve(len(segment) * 2) {
			return staticEvalResult{}
		}
		builder.WriteString(segment)
		if index+1 >= length {
			continue
		}
		substitution, ok := staticValueToString(staticArgument(arguments, index+1))
		if !ok || len(substitution) > maxStaticStringLength-builder.Len() {
			if ok {
				work.exhaust()
			}
			return staticEvalResult{}
		}
		if !work.reserve(len(substitution) * 2) {
			return staticEvalResult{}
		}
		builder.WriteString(substitution)
	}
	return staticEvalResult{value: builder.String(), ok: true}
}

func staticToLength(value any) (int, bool) {
	number, ok := staticValueToNumber(value)
	if !ok {
		return 0, false
	}
	if math.IsNaN(number) || number <= 0 {
		return 0, true
	}
	if math.IsInf(number, 1) || number > maxStaticAggregateElements {
		return 0, false
	}
	return int(math.Trunc(number)), true
}

func (staticEvaluator *StaticStringEvaluator) evalESLintStringRawTag(node *ast.Node) staticEvalResult {
	tagged := node.AsTaggedTemplateExpression()
	if tagged == nil || tagged.Tag == nil || tagged.Template == nil || staticEvaluator.sourceFile == nil {
		return staticEvalResult{}
	}
	tag := staticEvaluator.evalValue(tagged.Tag)
	builtin, ok := tag.value.(staticBuiltinValue)
	if !tag.ok || !ok || builtin != "String.raw" {
		return staticEvalResult{}
	}

	var builder strings.Builder
	appendPart := func(part string) bool {
		if !staticEvaluator.stringWorkContext.reserve(max(len(part)*6, 1)) {
			return false
		}
		part = strings.ReplaceAll(strings.ReplaceAll(part, "\r\n", "\n"), "\r", "\n")
		if len(part) > maxStaticStringLength-builder.Len() {
			staticEvaluator.stringWorkContext.exhaust()
			return false
		}
		builder.WriteString(part)
		return true
	}
	if tagged.Template.Kind == ast.KindNoSubstitutionTemplateLiteral {
		raw, ok := staticRawTemplateToken(staticEvaluator.sourceFile.Text(), tagged.Template)
		if !ok || !appendPart(raw) {
			return staticEvalResult{}
		}
		return staticEvalResult{value: builder.String(), ok: true}
	}
	if tagged.Template.Kind != ast.KindTemplateExpression {
		return staticEvalResult{}
	}
	template := tagged.Template.AsTemplateExpression()
	head, ok := staticRawTemplateToken(staticEvaluator.sourceFile.Text(), template.Head)
	if !ok || !appendPart(head) {
		return staticEvalResult{}
	}
	for _, spanNode := range template.TemplateSpans.Nodes {
		span := spanNode.AsTemplateSpan()
		value := staticEvaluator.evalValue(span.Expression)
		if !value.ok {
			return staticEvalResult{}
		}
		text, ok := staticValueToString(value.value)
		if !ok || !appendPart(text) {
			return staticEvalResult{}
		}
		raw, ok := staticRawTemplateToken(staticEvaluator.sourceFile.Text(), span.Literal)
		if !ok || !appendPart(raw) {
			return staticEvalResult{}
		}
	}
	return staticEvalResult{value: builder.String(), ok: true}
}

func staticRawTemplateToken(source string, node *ast.Node) (string, bool) {
	if node == nil || node.Pos() < 0 || node.End() > len(source) || node.Pos() >= node.End() {
		return "", false
	}
	text := source[node.Pos():node.End()]
	switch node.Kind {
	case ast.KindNoSubstitutionTemplateLiteral:
		start, end := strings.IndexByte(text, '`'), strings.LastIndexByte(text, '`')
		if start < 0 || end <= start {
			return "", false
		}
		return text[start+1 : end], true
	case ast.KindTemplateHead:
		start, end := strings.IndexByte(text, '`'), strings.LastIndex(text, "${")
		if start < 0 || end < start {
			return "", false
		}
		return text[start+1 : end], true
	case ast.KindTemplateMiddle:
		start, end := strings.IndexByte(text, '}'), strings.LastIndex(text, "${")
		if start < 0 || end < start {
			return "", false
		}
		return text[start+1 : end], true
	case ast.KindTemplateTail:
		start, end := strings.IndexByte(text, '}'), strings.LastIndexByte(text, '`')
		if start < 0 || end < start {
			return "", false
		}
		return text[start+1 : end], true
	default:
		return "", false
	}
}

func staticURICall(
	work *staticStringWorkContext,
	name string,
	arguments []any,
) staticEvalResult {
	text, ok := staticValueToString(staticArgument(arguments, 0))
	if !ok {
		return staticEvalResult{}
	}
	// URI helpers can build escaped output and, for the legacy escape pair,
	// full UTF-16 scratch buffers. Eight bytes of work per input byte is a
	// conservative upper bound for the implemented paths.
	if !work.reserve(max(len(text)*8, 1)) {
		return staticEvalResult{}
	}
	switch name {
	case "encodeURI", "encodeURIComponent":
		value, ok := staticEncodeURI(text, name == "encodeURIComponent")
		return staticEvalResult{value: value, ok: ok}
	case "decodeURI", "decodeURIComponent":
		value, ok := staticDecodeURI(text, name == "decodeURIComponent")
		return staticEvalResult{value: value, ok: ok}
	case "escape":
		value, ok := staticEscape(text)
		return staticEvalResult{value: value, ok: ok}
	case "unescape":
		return staticEvalResult{value: staticUnescape(text), ok: true}
	}
	return staticEvalResult{}
}

func isASCIIHex(value byte) bool {
	return value >= '0' && value <= '9' || value >= 'a' && value <= 'f' || value >= 'A' && value <= 'F'
}

func hasUnpairedSurrogate(text string) bool {
	units := ecmascript.StringCodeUnits(text)
	for index := 0; index < len(units); index++ {
		unit := units[index]
		switch {
		case unit >= 0xD800 && unit <= 0xDBFF:
			if index+1 >= len(units) || units[index+1] < 0xDC00 || units[index+1] > 0xDFFF {
				return true
			}
			index++
		case unit >= 0xDC00 && unit <= 0xDFFF:
			return true
		}
	}
	return false
}

const upperHex = "0123456789ABCDEF"

func staticEncodeURI(text string, component bool) (string, bool) {
	if hasUnpairedSurrogate(text) {
		return "", false
	}
	var builder strings.Builder
	for _, value := range []byte(text) {
		if value < utf8.RuneSelf && staticURIUnescaped(value, component) {
			if builder.Len() >= maxStaticStringLength {
				return "", false
			}
			builder.WriteByte(value)
			continue
		}
		if builder.Len() > maxStaticStringLength-3 {
			return "", false
		}
		builder.WriteByte('%')
		builder.WriteByte(upperHex[value>>4])
		builder.WriteByte(upperHex[value&0xF])
	}
	return builder.String(), true
}

func staticURIUnescaped(value byte, component bool) bool {
	if value >= 'a' && value <= 'z' || value >= 'A' && value <= 'Z' || value >= '0' && value <= '9' ||
		strings.ContainsRune("-_.!~*'()", rune(value)) {
		return true
	}
	return !component && strings.ContainsRune(";,/?:@&=+$#", rune(value))
}

func staticDecodeURI(text string, component bool) (string, bool) {
	var builder strings.Builder
	for index := 0; index < len(text); {
		if text[index] != '%' {
			if builder.Len() >= maxStaticStringLength*3 {
				return "", false
			}
			builder.WriteByte(text[index])
			index++
			continue
		}
		if index+2 >= len(text) || !isASCIIHex(text[index+1]) || !isASCIIHex(text[index+2]) {
			return "", false
		}
		first := asciiHexValue(text[index+1])<<4 | asciiHexValue(text[index+2])
		if first < utf8.RuneSelf {
			if !component && strings.ContainsRune(";,/?:@&=+$#", rune(first)) {
				if builder.Len() > maxStaticStringLength*3-3 {
					return "", false
				}
				builder.WriteString(text[index : index+3])
			} else {
				if builder.Len() >= maxStaticStringLength*3 {
					return "", false
				}
				builder.WriteByte(first)
			}
			index += 3
			continue
		}
		sequenceLength := 0
		switch {
		case first >= 0xC2 && first <= 0xDF:
			sequenceLength = 2
		case first >= 0xE0 && first <= 0xEF:
			sequenceLength = 3
		case first >= 0xF0 && first <= 0xF4:
			sequenceLength = 4
		default:
			return "", false
		}
		decoded := make([]byte, sequenceLength)
		for offset := range sequenceLength {
			position := index + offset*3
			if position+2 >= len(text) || text[position] != '%' ||
				!isASCIIHex(text[position+1]) || !isASCIIHex(text[position+2]) {
				return "", false
			}
			decoded[offset] = asciiHexValue(text[position+1])<<4 | asciiHexValue(text[position+2])
		}
		if !utf8.Valid(decoded) {
			return "", false
		}
		if len(decoded) > maxStaticStringLength*3-builder.Len() {
			return "", false
		}
		builder.Write(decoded)
		index += sequenceLength * 3
	}
	return builder.String(), true
}

func staticEscape(text string) (string, bool) {
	var builder strings.Builder
	for _, unit := range ecmascript.StringCodeUnits(text) {
		if unit < utf8.RuneSelf && (unit >= 'a' && unit <= 'z' || unit >= 'A' && unit <= 'Z' ||
			unit >= '0' && unit <= '9' || strings.ContainsRune("@*_+-./", rune(unit))) {
			if builder.Len() >= maxStaticStringLength {
				return "", false
			}
			builder.WriteByte(byte(unit))
			continue
		}
		if unit < 256 {
			if builder.Len() > maxStaticStringLength-3 {
				return "", false
			}
			builder.WriteByte('%')
			builder.WriteByte(upperHex[unit>>4])
			builder.WriteByte(upperHex[unit&0xF])
			continue
		}
		if builder.Len() > maxStaticStringLength-6 {
			return "", false
		}
		builder.WriteString("%u")
		builder.WriteByte(upperHex[unit>>12])
		builder.WriteByte(upperHex[unit>>8&0xF])
		builder.WriteByte(upperHex[unit>>4&0xF])
		builder.WriteByte(upperHex[unit&0xF])
	}
	return builder.String(), true
}

func staticUnescape(text string) string {
	units := ecmascript.StringCodeUnits(text)
	result := make([]uint16, 0, len(units))
	for index := 0; index < len(units); index++ {
		if units[index] != '%' {
			result = append(result, units[index])
			continue
		}
		if index+5 < len(units) && units[index+1] == 'u' &&
			staticHexUnits(units[index+2:index+6]) {
			result = append(result, uint16(asciiHexValue(byte(units[index+2])))<<12|
				uint16(asciiHexValue(byte(units[index+3])))<<8|
				uint16(asciiHexValue(byte(units[index+4])))<<4|
				uint16(asciiHexValue(byte(units[index+5]))))
			index += 5
			continue
		}
		if index+2 < len(units) && staticHexUnits(units[index+1:index+3]) {
			result = append(result, uint16(asciiHexValue(byte(units[index+1])))<<4|
				uint16(asciiHexValue(byte(units[index+2]))))
			index += 2
			continue
		}
		result = append(result, '%')
	}
	return ecmascript.StringFromCodeUnits(result)
}

func staticHexUnits(units []uint16) bool {
	for _, unit := range units {
		if unit > 0x7F || !isASCIIHex(byte(unit)) {
			return false
		}
	}
	return true
}

func asciiHexValue(value byte) byte {
	switch {
	case value >= '0' && value <= '9':
		return value - '0'
	case value >= 'a' && value <= 'f':
		return value - 'a' + 10
	default:
		return value - 'A' + 10
	}
}

func staticMathCall(method string, arguments []any) staticEvalResult {
	if method == "random" || !staticMathMethod(method) {
		return staticEvalResult{}
	}
	usedArguments := 1
	switch method {
	case "atan2", "imul", "pow":
		usedArguments = 2
	case "hypot", "max", "min":
		usedArguments = len(arguments)
	}
	numbers := make([]float64, min(len(arguments), usedArguments))
	exact := true
	for index, argument := range arguments[:len(numbers)] {
		number, ok := staticValueToNumber(argument)
		if !ok {
			if !staticCanConvertToNumber(argument) {
				return staticEvalResult{}
			}
			exact = false
		}
		numbers[index] = number
	}
	if !exact {
		return staticEvalResult{value: staticUnknownNumberValue{}, ok: true}
	}
	if method == "hypot" && len(numbers) == 0 {
		return staticEvalResult{value: staticNumberValue(0), ok: true}
	}
	if staticMathRequiresRuntimeExactness(method) {
		if result, ok := staticMathExactSpecialCase(method, numbers); ok {
			return staticEvalResult{value: staticNumberValue(result), ok: true}
		}
		// V8's fdlibm/LLVM-libc results can differ from Go's libm by one or
		// more ULPs. The return type is still certain, but using a Go result in
		// a strict comparison could choose the wrong receiver branch.
		return staticEvalResult{value: staticUnknownNumberValue{}, ok: true}
	}
	argument := func(index int) float64 {
		if index >= len(numbers) {
			return math.NaN()
		}
		return numbers[index]
	}
	var result float64
	switch method {
	case "abs":
		result = math.Abs(argument(0))
	case "acos":
		result = ecmascript.Acos(argument(0))
	case "acosh":
		result = math.Acosh(argument(0))
	case "asin":
		result = math.Asin(argument(0))
	case "asinh":
		result = math.Asinh(argument(0))
	case "atan":
		result = math.Atan(argument(0))
	case "atanh":
		result = math.Atanh(argument(0))
	case "atan2":
		result = math.Atan2(argument(0), argument(1))
	case "cbrt":
		result = math.Cbrt(argument(0))
	case "ceil":
		result = math.Ceil(argument(0))
	case "clz32":
		result = float64(bits.LeadingZeros32(toUint32(argument(0))))
	case "cos":
		result = math.Cos(argument(0))
	case "cosh":
		result = math.Cosh(argument(0))
	case "exp":
		result = math.Exp(argument(0))
	case "expm1":
		result = math.Expm1(argument(0))
	case "floor":
		result = math.Floor(argument(0))
	case "fround":
		result = float64(float32(argument(0)))
	case "hypot":
		result = 0
		for _, number := range numbers {
			result = math.Hypot(result, number)
		}
	case "imul":
		result = float64(int32(toUint32(argument(0))) * int32(toUint32(argument(1))))
	case "log":
		result = math.Log(argument(0))
	case "log1p":
		result = math.Log1p(argument(0))
	case "log2":
		result = math.Log2(argument(0))
	case "log10":
		result = math.Log10(argument(0))
	case "max":
		result = math.Inf(-1)
		for _, number := range numbers {
			if math.IsNaN(number) {
				result = math.NaN()
				break
			}
			result = math.Max(result, number)
		}
	case "min":
		result = math.Inf(1)
		for _, number := range numbers {
			if math.IsNaN(number) {
				result = math.NaN()
				break
			}
			result = math.Min(result, number)
		}
	case "pow":
		result = math.Pow(argument(0), argument(1))
	case "round":
		value := argument(0)
		if math.IsNaN(value) || math.IsInf(value, 0) || value == 0 || math.Abs(value) >= 1<<52 {
			result = value
			break
		}
		floor := math.Floor(value)
		result = floor
		if value-floor >= 0.5 {
			result = floor + 1
		}
		if result == 0 && (value < 0 || math.Signbit(value)) {
			result = math.Copysign(0, -1)
		}
	case "sign":
		value := argument(0)
		switch {
		case math.IsNaN(value), value == 0:
			result = value
		case value < 0:
			result = -1
		default:
			result = 1
		}
	case "sin":
		result = math.Sin(argument(0))
	case "sinh":
		result = math.Sinh(argument(0))
	case "sqrt":
		result = math.Sqrt(argument(0))
	case "tan":
		result = math.Tan(argument(0))
	case "tanh":
		result = math.Tanh(argument(0))
	case "trunc":
		result = math.Trunc(argument(0))
	default:
		return staticEvalResult{value: staticUnknownNumberValue{}, ok: true}
	}
	return staticEvalResult{value: staticNumberValue(result), ok: true}
}

// staticMathExactSpecialCase keeps results that ECMAScript defines exactly,
// without routing general transcendental calculations through Go's libm.
// The latter can differ from the JavaScript host by an ULP and must remain an
// abstract number because strict comparisons can select a receiver branch.
func staticMathExactSpecialCase(method string, numbers []float64) (float64, bool) {
	argument := math.NaN()
	if len(numbers) > 0 {
		argument = numbers[0]
	}
	if math.IsNaN(argument) && method != "hypot" && method != "pow" {
		return math.NaN(), true
	}
	if argument == 0 {
		switch method {
		case "asin", "asinh", "atan", "atanh", "cbrt", "expm1", "log1p", "sin", "sinh", "tan", "tanh":
			return argument, true
		case "cos", "cosh", "exp":
			return 1, true
		case "log", "log2", "log10":
			return math.Inf(-1), true
		}
	}
	switch method {
	case "acosh":
		if argument < 1 {
			return math.NaN(), true
		}
		if argument == 1 {
			return 0, true
		}
		if math.IsInf(argument, 1) {
			return argument, true
		}
	case "asin":
		if math.Abs(argument) > 1 {
			return math.NaN(), true
		}
	case "asinh", "sinh":
		if math.IsInf(argument, 0) {
			return argument, true
		}
	case "atanh":
		if argument == 1 {
			return math.Inf(1), true
		}
		if argument == -1 {
			return math.Inf(-1), true
		}
		if math.Abs(argument) > 1 {
			return math.NaN(), true
		}
	case "atan2":
		if len(numbers) < 2 || math.IsNaN(numbers[0]) || math.IsNaN(numbers[1]) {
			return math.NaN(), true
		}
		y, x := numbers[0], numbers[1]
		if x > 0 && y == 0 || math.IsInf(x, 1) && !math.IsInf(y, 0) {
			return math.Copysign(0, y), true
		}
	case "log", "log2", "log10":
		if argument < 0 {
			return math.NaN(), true
		}
		if argument == 1 {
			return 0, true
		}
		if math.IsInf(argument, 1) {
			return argument, true
		}
	case "log1p":
		if argument < -1 {
			return math.NaN(), true
		}
		if argument == -1 {
			return math.Inf(-1), true
		}
		if math.IsInf(argument, 1) {
			return argument, true
		}
	case "cbrt":
		if math.IsInf(argument, 0) {
			return argument, true
		}
		if math.Trunc(argument) == argument && math.Abs(argument) <= 1<<53 {
			root := math.Round(math.Cbrt(argument))
			if root*root*root == argument {
				return root, true
			}
		}
	case "hypot":
		allZero := true
		hasNaN := false
		for _, number := range numbers {
			if math.IsInf(number, 0) {
				return math.Inf(1), true
			}
			if math.IsNaN(number) {
				hasNaN = true
			}
			if number != 0 {
				allZero = false
			}
		}
		if hasNaN {
			return math.NaN(), true
		}
		if allZero {
			return 0, true
		}
		if len(numbers) == 1 {
			return math.Abs(numbers[0]), true
		}
	case "pow":
		if len(numbers) < 2 {
			return math.NaN(), true
		}
		base, exponent := numbers[0], numbers[1]
		if exponent == 0 {
			return 1, true
		}
		if math.IsNaN(exponent) || math.IsNaN(base) {
			return math.NaN(), true
		}
		absoluteBase := math.Abs(base)
		if math.IsInf(exponent, 0) {
			if absoluteBase == 1 {
				return math.NaN(), true
			}
			if (absoluteBase > 1) == math.IsInf(exponent, 1) {
				return math.Inf(1), true
			}
			return 0, true
		}
		if math.IsInf(base, 0) || base == 0 {
			negativeResult := math.Signbit(base) && staticOddInteger(exponent)
			if exponent > 0 {
				if math.IsInf(base, 0) {
					if negativeResult {
						return math.Inf(-1), true
					}
					return math.Inf(1), true
				}
				if negativeResult {
					return math.Copysign(0, -1), true
				}
				return 0, true
			}
			if math.IsInf(base, 0) {
				if negativeResult {
					return math.Copysign(0, -1), true
				}
				return 0, true
			}
			if negativeResult {
				return math.Inf(-1), true
			}
			return math.Inf(1), true
		}
		if base < 0 && math.Trunc(exponent) != exponent {
			return math.NaN(), true
		}
		if result, ok := staticExactSafeIntegerPower(base, exponent); ok {
			return result, true
		}
	case "cos":
		if math.IsInf(argument, 0) {
			return math.NaN(), true
		}
	case "cosh":
		if math.IsInf(argument, 0) {
			return math.Inf(1), true
		}
	case "exp":
		if math.IsInf(argument, 0) {
			if math.IsInf(argument, -1) {
				return 0, true
			}
			return argument, true
		}
	case "expm1":
		if math.IsInf(argument, -1) {
			return -1, true
		}
		if math.IsInf(argument, 1) {
			return argument, true
		}
	case "sin", "tan":
		if math.IsInf(argument, 0) {
			return math.NaN(), true
		}
	case "tanh":
		if math.IsInf(argument, 1) {
			return 1, true
		}
		if math.IsInf(argument, -1) {
			return -1, true
		}
	}
	return 0, false
}

func staticOddInteger(value float64) bool {
	absolute := math.Abs(value)
	return math.Trunc(value) == value && absolute < 1<<53 && int64(absolute)&1 != 0
}

func staticExactSafeIntegerPower(base, exponent float64) (float64, bool) {
	if math.Trunc(base) != base || math.Abs(base) > 1<<53 ||
		math.Trunc(exponent) != exponent || exponent < 0 || exponent > 1<<16 {
		return 0, false
	}
	if base == 0 {
		if math.Signbit(base) && staticOddInteger(exponent) {
			return math.Copysign(0, -1), true
		}
		return 0, true
	}
	absoluteBase := uint64(math.Abs(base))
	if absoluteBase > 1 {
		minimumBits := uint64(bits.Len64(absoluteBase)-1)*uint64(exponent) + 1
		if minimumBits > 53 {
			return 0, false
		}
	}
	baseInteger := new(big.Int)
	baseInteger.SetInt64(int64(base))
	exponentInteger := big.NewInt(int64(exponent))
	result := new(big.Int).Exp(baseInteger, exponentInteger, nil)
	if result.BitLen() > 53 || !result.IsInt64() {
		return 0, false
	}
	return float64(result.Int64()), true
}

func staticMathRequiresRuntimeExactness(method string) bool {
	switch method {
	case "acosh", "asin", "asinh", "atan", "atanh", "atan2", "cbrt", "cos", "cosh",
		"exp", "expm1", "hypot", "log", "log1p", "log2", "log10", "pow",
		"sin", "sinh", "tan", "tanh":
		return true
	}
	return false
}

func staticMathMethod(method string) bool {
	switch method {
	case "abs", "acos", "acosh", "asin", "asinh", "atan", "atanh", "atan2",
		"cbrt", "ceil", "clz32", "cos", "cosh", "exp", "expm1", "floor",
		"fround", "hypot", "imul", "log", "log1p", "log2", "log10", "max",
		"min", "pow", "round", "sign", "sin", "sinh", "sqrt", "tan", "tanh", "trunc":
		return true
	}
	return false
}

func staticBuiltinIsFunction(name string) bool {
	if strings.Contains(name, ".") {
		return true
	}
	switch name {
	case "Math", "JSON", "Reflect":
		return false
	case "Array", "ArrayBuffer", "BigInt", "BigInt64Array", "BigUint64Array",
		"Boolean", "DataView", "Date", "decodeURI", "decodeURIComponent",
		"encodeURI", "encodeURIComponent", "escape", "Float32Array", "Float64Array",
		"Function", "Int16Array", "Int32Array", "Int8Array", "isFinite", "isNaN",
		"isPrototypeOf", "Map", "Number", "Object", "parseFloat", "parseInt",
		"Promise", "Proxy", "RegExp", "Set", "String", "Symbol", "Uint16Array",
		"Uint32Array", "Uint8Array", "Uint8ClampedArray", "unescape", "WeakMap", "WeakSet":
		return true
	}
	return false
}

func staticValueIsArray(value any) bool {
	if _, ok := value.(*staticArrayValue); ok {
		return true
	}
	prototype, ok := value.(staticBuiltinObjectValue)
	return ok && prototype == "Array.prototype"
}

func staticBuiltinHasPrototype(name string) bool {
	switch name {
	case "Array", "ArrayBuffer", "BigInt", "BigInt64Array", "BigUint64Array", "Boolean",
		"DataView", "Date", "Float32Array", "Float64Array", "Function", "Int16Array",
		"Int32Array", "Int8Array", "Map", "Number", "Object", "Promise", "RegExp",
		"Set", "String", "Symbol", "Uint16Array", "Uint32Array", "Uint8Array",
		"Uint8ClampedArray", "WeakMap", "WeakSet":
		return true
	}
	return false
}

func staticBuiltinFunctionProperty(name, key string) bool {
	switch name {
	case "Array":
		return key == "from" || key == "fromAsync" || key == "isArray" || key == "of"
	case "ArrayBuffer":
		return key == "isView"
	case "BigInt":
		return key == "asIntN" || key == "asUintN"
	case "Number":
		switch key {
		case "isFinite", "isInteger", "isNaN", "isSafeInteger", "parseFloat", "parseInt":
			return true
		}
	case "Object":
		switch key {
		case "assign", "create", "defineProperties", "defineProperty", "entries", "freeze",
			"fromEntries", "getOwnPropertyDescriptor", "getOwnPropertyDescriptors",
			"getOwnPropertyNames", "getOwnPropertySymbols", "getPrototypeOf", "groupBy",
			"hasOwn", "is", "isExtensible", "isFrozen", "isSealed", "keys",
			"preventExtensions", "seal", "setPrototypeOf", "values":
			return true
		}
	case "String":
		return key == "fromCharCode" || key == "fromCodePoint" || key == "raw"
	case "Date":
		return key == "UTC" || key == "now" || key == "parse"
	case "Symbol":
		return key == "for" || key == "keyFor"
	case "JSON":
		return key == "isRawJSON" || key == "parse" || key == "rawJSON" || key == "stringify"
	case "Reflect":
		switch key {
		case "apply", "construct", "defineProperty", "deleteProperty", "get",
			"getOwnPropertyDescriptor", "getPrototypeOf", "has", "isExtensible",
			"ownKeys", "preventExtensions", "set", "setPrototypeOf":
			return true
		}
	case "Promise":
		switch key {
		case "all", "allSettled", "any", "race", "reject", "resolve", "withResolvers":
			return true
		}
	case "Map":
		return key == "groupBy"
	case "Proxy":
		return key == "revocable"
	case "Math":
		return key == "random" || staticMathMethod(key)
	case "BigInt64Array", "BigUint64Array", "Float32Array", "Float64Array", "Int16Array",
		"Int32Array", "Int8Array", "Uint16Array", "Uint32Array", "Uint8Array", "Uint8ClampedArray":
		return key == "from" || key == "of"
	}
	return false
}

func staticTypedArrayBytesPerElement(name string) (int, bool) {
	switch name {
	case "Int8Array", "Uint8Array", "Uint8ClampedArray":
		return 1, true
	case "Int16Array", "Uint16Array":
		return 2, true
	case "Int32Array", "Uint32Array", "Float32Array":
		return 4, true
	case "BigInt64Array", "BigUint64Array", "Float64Array":
		return 8, true
	default:
		return 0, false
	}
}

func staticTypedArrayPrototypeMethod(name, key string) bool {
	if _, ok := staticTypedArrayBytesPerElement(name); !ok {
		return false
	}
	switch key {
	case "at", "copyWithin", "entries", "every", "fill", "filter", "find", "findIndex",
		"findLast", "findLastIndex", "forEach", "includes", "indexOf", "join", "keys",
		"lastIndexOf", "map", "reduce", "reduceRight", "reverse", "set", "slice", "some",
		"sort", "subarray", "toLocaleString", "toReversed", "toSorted", "toString", "values", "with":
		return true
	default:
		return false
	}
}

func staticSymbolProperty(key string) bool {
	switch key {
	case "asyncDispose", "asyncIterator", "dispose", "hasInstance", "isConcatSpreadable",
		"iterator", "match", "matchAll", "replace", "search", "species", "split",
		"toPrimitive", "toStringTag", "unscopables":
		return true
	}
	return false
}

func staticArrayPrototypeMethod(key string) bool {
	switch key {
	case "at", "concat", "copyWithin", "entries", "every", "fill", "filter", "find",
		"findIndex", "findLast", "findLastIndex", "flat", "flatMap", "forEach", "includes",
		"indexOf", "join", "keys", "lastIndexOf", "map", "pop", "push", "reduce",
		"reduceRight", "reverse", "shift", "slice", "some", "sort", "splice", "toLocaleString",
		"toReversed", "toSorted", "toSpliced", "toString", "unshift", "values", "with":
		return true
	}
	return false
}

func staticIteratorPrototypeMethod(key string) bool {
	switch key {
	case "drop", "every", "filter", "find", "flatMap", "forEach", "map", "reduce", "some", "take", "toArray":
		return true
	}
	return false
}

func staticCollectionPrototypeMethod(kind, key string) bool {
	switch key {
	case "clear", "delete", "entries", "forEach", "has", "keys", "values":
		return true
	case "get", "set":
		return kind == "Map"
	case "add":
		return kind == "Set"
	case "difference", "intersection", "isDisjointFrom", "isSubsetOf", "isSupersetOf",
		"symmetricDifference", "union":
		return kind == "Set"
	}
	return false
}

func staticDataViewPrototypeMethod(key string) bool {
	switch key {
	case "getBigInt64", "getBigUint64", "getFloat32", "getFloat64", "getInt8", "getInt16",
		"getInt32", "getUint8", "getUint16", "getUint32", "setBigInt64", "setBigUint64",
		"setFloat32", "setFloat64", "setInt8", "setInt16", "setInt32", "setUint8",
		"setUint16", "setUint32":
		return true
	}
	return false
}

func staticStringPrototypeMethod(key string) bool {
	switch key {
	case "anchor", "at", "big", "blink", "bold", "charAt", "charCodeAt", "codePointAt",
		"concat", "endsWith", "fixed", "fontcolor", "fontsize", "includes", "indexOf",
		"isWellFormed", "italics", "lastIndexOf", "link", "localeCompare", "match", "matchAll",
		"normalize", "padEnd", "padStart", "repeat", "replace", "replaceAll", "search", "slice",
		"small", "split", "startsWith", "strike", "sub", "substr", "substring", "sup",
		"toLocaleLowerCase", "toLocaleUpperCase", "toLowerCase", "toString", "toUpperCase",
		"toWellFormed", "trim", "trimEnd", "trimLeft", "trimRight", "trimStart", "valueOf":
		return true
	}
	return false
}

func staticStringMemberValue(value, key string) staticEvalResult {
	if staticStringPrototypeMethod(key) {
		return staticEvalResult{value: staticBuiltinValue("String.prototype." + key), ok: true}
	}
	units := ecmascript.StringCodeUnits(value)
	if key == "length" {
		return staticEvalResult{value: staticNumberValue(len(units)), ok: true}
	}
	if index, ok := staticArrayIndex(key); ok {
		if index >= len(units) {
			return staticEvalResult{value: staticUndefinedValue{}, ok: true}
		}
		return staticEvalResult{value: ecmascript.StringFromCodeUnits(units[index : index+1]), ok: true}
	}
	if key == "__proto__" {
		return staticEvalResult{value: staticBuiltinObjectValue("String.prototype"), ok: true}
	}
	return staticObjectPrototypeMember(key)
}

func staticESLintMemberValue(object any, key string) staticEvalResult {
	switch object := object.(type) {
	case *staticObjectValue:
		if value, ok := staticObjectOwnProperty(object, key); ok {
			return staticEvalResult{value: value, ok: true}
		}
		if object.prototypeSet {
			if object.prototype == nil {
				return staticEvalResult{value: staticUndefinedValue{}, ok: true}
			}
			return staticESLintMemberValue(object.prototype, key)
		}
		return staticObjectPrototypeMember(key)
	case *staticArrayValue:
		if key == "length" {
			return staticEvalResult{value: staticNumberValue(object.length), ok: true}
		}
		if index, ok := staticArrayIndex(key); ok {
			if index >= object.length || object.omitted[index] {
				return staticEvalResult{value: staticUndefinedValue{}, ok: true}
			}
			return staticEvalResult{value: object.element(index), ok: true}
		}
		if staticArrayPrototypeMethod(key) {
			return staticEvalResult{value: staticBuiltinValue("Array.prototype." + key), ok: true}
		}
		if key == "constructor" {
			return staticEvalResult{value: staticBuiltinValue("Array"), ok: true}
		}
		if key == "__proto__" {
			return staticEvalResult{}
		}
		return staticObjectPrototypeMember(key)
	case string:
		return staticStringMemberValue(object, key)
	case *staticStringNode:
		text, _ := staticValueAsString(object)
		return staticStringMemberValue(text, key)
	case staticCollectionValue:
		if key == "size" {
			return staticEvalResult{value: staticNumberValue(len(object.entries)), ok: true}
		}
		if staticCollectionPrototypeMethod(object.kind, key) {
			return staticEvalResult{value: staticBuiltinValue(object.kind + ".prototype." + key), ok: true}
		}
		if key == "constructor" {
			return staticEvalResult{value: staticBuiltinValue(object.kind), ok: true}
		}
		if key == "__proto__" {
			return staticEvalResult{}
		}
		return staticObjectPrototypeMember(key)
	case staticRegExpValue:
		return staticRegExpMemberValue(object, key)
	case staticBoxedValue:
		if key == "__proto__" {
			return staticEvalResult{}
		}
		if _, symbol := object.value.(staticSymbolValue); symbol && key == "description" {
			return staticEvalResult{}
		}
		if key == "constructor" {
			switch staticValueKindOf(object.value) {
			case staticKindBoolean:
				return staticEvalResult{value: staticBuiltinValue("Boolean"), ok: true}
			case staticKindNumber:
				return staticEvalResult{value: staticBuiltinValue("Number"), ok: true}
			case staticKindString:
				return staticEvalResult{value: staticBuiltinValue("String"), ok: true}
			case staticKindBigInt:
				return staticEvalResult{value: staticBuiltinValue("BigInt"), ok: true}
			case staticKindSymbol:
				return staticEvalResult{value: staticBuiltinValue("Symbol"), ok: true}
			}
		}
		switch staticValueKindOf(object.value) {
		case staticKindString:
			if text, ok := staticValueAsString(object.value); ok {
				return staticStringMemberValue(text, key)
			}
		case staticKindNumber:
			if staticNumberPrototypeMethod(key) {
				return staticEvalResult{value: staticBuiltinValue("Number.prototype." + key), ok: true}
			}
		case staticKindBoolean:
			if key == "toString" || key == "valueOf" {
				return staticEvalResult{value: staticBuiltinValue("Boolean.prototype." + key), ok: true}
			}
		case staticKindBigInt:
			if key == "toString" || key == "valueOf" || key == "toLocaleString" {
				return staticEvalResult{value: staticBuiltinValue("BigInt.prototype." + key), ok: true}
			}
		case staticKindSymbol:
			if key == "toString" || key == "valueOf" {
				return staticEvalResult{value: staticBuiltinValue("Symbol.prototype." + key), ok: true}
			}
		}
		if key == "__proto__" {
			return staticEvalResult{}
		}
		return staticObjectPrototypeMember(key)
	case staticDateValue:
		if key == "constructor" {
			return staticEvalResult{value: staticBuiltinValue("Date"), ok: true}
		}
		if key == "__proto__" {
			return staticEvalResult{}
		}
		if staticDatePrototypeMethod(key) {
			return staticEvalResult{value: staticBuiltinValue("Date.prototype." + key), ok: true}
		}
		return staticObjectPrototypeMember(key)
	case staticIteratorValue:
		if key == "next" {
			return staticEvalResult{value: staticBuiltinValue(object.kind + "Iterator.prototype.next"), ok: true}
		}
		if staticIteratorPrototypeMethod(key) {
			return staticEvalResult{value: staticBuiltinValue("Iterator.prototype." + key), ok: true}
		}
		if key == "constructor" || key == "__proto__" {
			// Iterator.prototype.constructor and the derived iterator prototype
			// chain are native accessors/objects that eslint-utils does not read.
			return staticEvalResult{}
		}
		return staticObjectPrototypeMember(key)
	case staticOpaqueObjectValue:
		return staticObjectPrototypeMember(key)
	case staticBuiltinObjectValue:
		return staticESLintBuiltinPrototypeMember(string(object), key)
	}
	switch value := object.(type) {
	case bool:
		if key == "constructor" {
			return staticEvalResult{value: staticBuiltinValue("Boolean"), ok: true}
		}
		if key == "toString" || key == "valueOf" {
			return staticEvalResult{value: staticBuiltinValue("Boolean.prototype." + key), ok: true}
		}
		if key == "__proto__" {
			return staticEvalResult{value: staticBuiltinObjectValue("Boolean.prototype"), ok: true}
		}
		return staticObjectPrototypeMember(key)
	case staticBigIntValue:
		if key == "constructor" {
			return staticEvalResult{value: staticBuiltinValue("BigInt"), ok: true}
		}
		if key == "toString" || key == "valueOf" || key == "toLocaleString" {
			return staticEvalResult{value: staticBuiltinValue("BigInt.prototype." + key), ok: true}
		}
		if key == "__proto__" {
			return staticEvalResult{value: staticBuiltinObjectValue("BigInt.prototype"), ok: true}
		}
		return staticObjectPrototypeMember(key)
	case staticSymbolValue:
		if key == "description" {
			description, known := staticSymbolDescription(value)
			if !known {
				return staticEvalResult{value: staticUnknownStringValue{truthy: true, numericNaN: true}, ok: true}
			}
			return staticEvalResult{value: description, ok: true}
		}
		if key == "constructor" {
			return staticEvalResult{value: staticBuiltinValue("Symbol"), ok: true}
		}
		if key == "toString" || key == "valueOf" {
			return staticEvalResult{value: staticBuiltinValue("Symbol.prototype." + key), ok: true}
		}
		if key == "__proto__" {
			return staticEvalResult{value: staticBuiltinObjectValue("Symbol.prototype"), ok: true}
		}
		return staticObjectPrototypeMember(key)
	}
	if staticValueKindOf(object) == staticKindNumber {
		if staticNumberPrototypeMethod(key) {
			return staticEvalResult{value: staticBuiltinValue("Number.prototype." + key), ok: true}
		}
		if key == "constructor" {
			return staticEvalResult{value: staticBuiltinValue("Number"), ok: true}
		}
		if key == "__proto__" {
			return staticEvalResult{value: staticBuiltinObjectValue("Number.prototype"), ok: true}
		}
		return staticObjectPrototypeMember(key)
	}
	return staticEvalResult{}
}

func staticESLintSymbolMemberValue(object any, key staticSymbolValue) staticEvalResult {
	if value, ok := object.(*staticObjectValue); ok {
		if property, found := staticObjectOwnSymbolProperty(value, key); found {
			return staticEvalResult{value: property, ok: true}
		}
		if value.prototypeSet && value.prototype != nil {
			return staticESLintSymbolMemberValue(value.prototype, key)
		}
		return staticEvalResult{value: staticUndefinedValue{}, ok: true}
	}
	if builtin, ok := object.(staticBuiltinValue); ok && key.wellKnown && key.description == "hasInstance" &&
		staticBuiltinIsFunction(string(builtin)) {
		return staticEvalResult{value: staticBuiltinValue("Function.prototype.[Symbol.hasInstance]"), ok: true}
	}
	if builtin, ok := object.(staticBuiltinValue); ok && key.wellKnown && key.description == "toStringTag" {
		switch builtin {
		case "Math":
			return staticEvalResult{value: "Math", ok: true}
		case "JSON":
			return staticEvalResult{value: "JSON", ok: true}
		case "Reflect":
			return staticEvalResult{value: "Reflect", ok: true}
		}
	}
	if builtin, ok := object.(staticBuiltinValue); ok && key.wellKnown && key.description == "species" &&
		staticBuiltinHasSpeciesGetter(string(builtin)) {
		// eslint-utils refuses to invoke native getters other than its explicit
		// Map/Set/RegExp allowlist.
		return staticEvalResult{}
	}
	if key.wellKnown && key.description == "iterator" {
		switch value := object.(type) {
		case *staticArrayValue:
			return staticEvalResult{value: staticBuiltinValue("Array.prototype.values"), ok: true}
		case string, *staticStringNode:
			return staticEvalResult{value: staticBuiltinValue("String.prototype.[Symbol.iterator]"), ok: true}
		case staticBoxedValue:
			if staticValueKindOf(value.value) == staticKindString {
				return staticEvalResult{value: staticBuiltinValue("String.prototype.[Symbol.iterator]"), ok: true}
			}
		case staticCollectionValue:
			method := "values"
			if value.kind == "Map" {
				method = "entries"
			}
			return staticEvalResult{value: staticBuiltinValue(value.kind + ".prototype." + method), ok: true}
		case staticIteratorValue:
			return staticEvalResult{value: staticBuiltinValue("Iterator.prototype.[Symbol.iterator]"), ok: true}
		case staticBuiltinObjectValue:
			name := string(value)
			switch name {
			case "Array.prototype":
				return staticEvalResult{value: staticBuiltinValue("Array.prototype.values"), ok: true}
			case "String.prototype":
				return staticEvalResult{value: staticBuiltinValue("String.prototype.[Symbol.iterator]"), ok: true}
			case "Map.prototype":
				return staticEvalResult{value: staticBuiltinValue("Map.prototype.entries"), ok: true}
			case "Set.prototype":
				return staticEvalResult{value: staticBuiltinValue("Set.prototype.values"), ok: true}
			}
			constructor := strings.TrimSuffix(name, ".prototype")
			if _, ok := staticTypedArrayBytesPerElement(constructor); ok {
				return staticEvalResult{value: staticBuiltinValue(name + ".values"), ok: true}
			}
		}
	}
	if key.wellKnown && key.description == "unscopables" {
		if _, array := object.(*staticArrayValue); array {
			return staticEvalResult{value: staticBuiltinObjectValue("Array.prototype[Symbol.unscopables]"), ok: true}
		}
	}
	if key.wellKnown && key.description == "toStringTag" {
		switch value := object.(type) {
		case staticSymbolValue:
			return staticEvalResult{value: "Symbol", ok: true}
		case staticBigIntValue:
			return staticEvalResult{value: "BigInt", ok: true}
		case staticCollectionValue:
			return staticEvalResult{value: value.kind, ok: true}
		case staticBoxedValue:
			switch value.value.(type) {
			case staticSymbolValue:
				return staticEvalResult{value: "Symbol", ok: true}
			case staticBigIntValue:
				return staticEvalResult{value: "BigInt", ok: true}
			}
		case staticIteratorValue:
			// The derived iterator prototypes expose this through native getters;
			// eslint-utils deliberately refuses to invoke them.
			return staticEvalResult{}
		}
	}
	if key.wellKnown && key.description == "toPrimitive" {
		switch value := object.(type) {
		case staticDateValue:
			return staticEvalResult{value: staticBuiltinValue("Date.prototype.[Symbol.toPrimitive]"), ok: true}
		case staticSymbolValue:
			return staticEvalResult{value: staticBuiltinValue("Symbol.prototype.[Symbol.toPrimitive]"), ok: true}
		case staticBoxedValue:
			if _, symbol := value.value.(staticSymbolValue); symbol {
				return staticEvalResult{value: staticBuiltinValue("Symbol.prototype.[Symbol.toPrimitive]"), ok: true}
			}
		}
	}
	if _, ok := object.(staticRegExpValue); ok && key.wellKnown {
		switch key.description {
		case "match", "matchAll", "replace", "search", "split":
			return staticEvalResult{value: staticBuiltinValue("RegExp.prototype.[Symbol." + key.description + "]"), ok: true}
		}
	}
	if prototype, ok := object.(staticBuiltinObjectValue); ok && key.wellKnown {
		name := string(prototype)
		if key.description == "toStringTag" {
			if _, typedArray := staticTypedArrayBytesPerElement(strings.TrimSuffix(name, ".prototype")); typedArray {
				return staticEvalResult{}
			}
		}
		if name == "Array.prototype" && key.description == "unscopables" {
			return staticEvalResult{value: staticBuiltinObjectValue("Array.prototype[Symbol.unscopables]"), ok: true}
		}
		if (name == "Map.prototype" || name == "Set.prototype") && key.description == "toStringTag" {
			return staticEvalResult{value: strings.TrimSuffix(name, ".prototype"), ok: true}
		}
		if name == "RegExp.prototype" {
			switch key.description {
			case "match", "matchAll", "replace", "search", "split":
				return staticEvalResult{value: staticBuiltinValue(name + ".[Symbol." + key.description + "]"), ok: true}
			}
		}
		if name == "Symbol.prototype" && key.description == "toPrimitive" {
			return staticEvalResult{value: staticBuiltinValue(name + ".[Symbol.toPrimitive]"), ok: true}
		}
		if key.description == "toStringTag" {
			switch name {
			case "ArrayBuffer.prototype", "BigInt.prototype", "DataView.prototype", "Promise.prototype",
				"Symbol.prototype", "WeakMap.prototype", "WeakSet.prototype":
				return staticEvalResult{value: strings.TrimSuffix(name, ".prototype"), ok: true}
			}
		}
		if name == "Date.prototype" && key.description == "toPrimitive" {
			return staticEvalResult{value: staticBuiltinValue(name + ".[Symbol.toPrimitive]"), ok: true}
		}
	}
	return staticEvalResult{value: staticUndefinedValue{}, ok: true}
}

func staticBuiltinHasSpeciesGetter(name string) bool {
	switch name {
	case "Array", "ArrayBuffer", "Map", "Promise", "RegExp", "Set":
		return true
	}
	_, typedArray := staticTypedArrayBytesPerElement(name)
	return typedArray
}

func staticObjectOwnSymbolProperty(object *staticObjectValue, key staticSymbolValue) (any, bool) {
	for index := len(object.extraProperties) - 1; index >= 0; index-- {
		property := object.extraProperties[index]
		if property.symbolKey && staticSymbolsEqual(property.symbol, key) {
			return property.value, true
		}
	}
	if object.propertyCount > 0 && object.property.symbolKey && staticSymbolsEqual(object.property.symbol, key) {
		return object.property.value, true
	}
	return nil, false
}

func staticSymbolsEqual(left, right staticSymbolValue) bool {
	equal, known := staticSymbolIdentityEqual(left, right)
	return known && equal
}

func staticSymbolIdentityEqual(left, right staticSymbolValue) (bool, bool) {
	if left.hostDependent || right.hostDependent {
		if left.hostDependent && right.hostDependent {
			return left.description == right.description, true
		}
		return false, false
	}
	return (left.global && right.global || left.wellKnown && right.wellKnown) &&
		left.description == right.description, true
}

func staticSymbolDescription(value staticSymbolValue) (string, bool) {
	if value.hostDependent {
		return "", false
	}
	if value.wellKnown {
		return "Symbol." + value.description, true
	}
	return value.description, true
}

func staticObjectPrototypeMember(key string) staticEvalResult {
	switch key {
	case "__proto__":
		return staticEvalResult{}
	case "constructor":
		return staticEvalResult{value: staticBuiltinValue("Object"), ok: true}
	case "__defineGetter__", "__defineSetter__", "__lookupGetter__", "__lookupSetter__",
		"hasOwnProperty", "isPrototypeOf", "propertyIsEnumerable", "toLocaleString", "toString", "valueOf":
		return staticEvalResult{value: staticBuiltinValue("Object.prototype." + key), ok: true}
	}
	return staticEvalResult{value: staticUndefinedValue{}, ok: true}
}

func staticNumberPrototypeMethod(key string) bool {
	switch key {
	case "toExponential", "toFixed", "toLocaleString", "toPrecision", "toString", "valueOf":
		return true
	default:
		return false
	}
}

func staticDatePrototypeMethod(key string) bool {
	switch key {
	case "getDate", "getDay", "getFullYear", "getHours", "getMilliseconds", "getMinutes",
		"getMonth", "getSeconds", "getTime", "getTimezoneOffset", "getUTCDate", "getUTCDay",
		"getUTCFullYear", "getUTCHours", "getUTCMilliseconds", "getUTCMinutes", "getUTCMonth",
		"getUTCSeconds", "setDate", "setFullYear", "setHours", "setMilliseconds", "setMinutes",
		"setMonth", "setSeconds", "setTime", "setUTCDate", "setUTCFullYear", "setUTCHours",
		"setUTCMilliseconds", "setUTCMinutes", "setUTCMonth", "setUTCSeconds", "toDateString",
		"toISOString", "toJSON", "toLocaleDateString", "toLocaleString", "toLocaleTimeString",
		"toString", "toTimeString", "toUTCString", "toGMTString", "getYear", "setYear", "valueOf":
		return true
	default:
		return false
	}
}

func staticRegExpMemberValue(value staticRegExpValue, key string) staticEvalResult {
	switch key {
	case "constructor":
		return staticEvalResult{value: staticBuiltinValue("RegExp"), ok: true}
	case "__proto__":
		return staticEvalResult{}
	case "dotAll":
		return staticEvalResult{value: strings.Contains(value.flags, "s"), ok: true}
	case "flags":
		return staticEvalResult{value: value.flags, ok: true}
	case "global":
		return staticEvalResult{value: strings.Contains(value.flags, "g"), ok: true}
	case "hasIndices":
		return staticEvalResult{value: strings.Contains(value.flags, "d"), ok: true}
	case "ignoreCase":
		return staticEvalResult{value: strings.Contains(value.flags, "i"), ok: true}
	case "lastIndex":
		return staticEvalResult{value: staticNumberValue(0), ok: true}
	case "multiline":
		return staticEvalResult{value: strings.Contains(value.flags, "m"), ok: true}
	case "source":
		return staticEvalResult{value: value.source, ok: true}
	case "sticky":
		return staticEvalResult{value: strings.Contains(value.flags, "y"), ok: true}
	case "unicode":
		return staticEvalResult{value: strings.Contains(value.flags, "u"), ok: true}
	case "unicodeSets":
		// eslint-utils intentionally does not invoke this native getter.
		return staticEvalResult{}
	case "compile", "exec", "test", "toString":
		return staticEvalResult{value: staticBuiltinValue("RegExp.prototype." + key), ok: true}
	}
	return staticObjectPrototypeMember(key)
}

func staticArrayPrototypeCall(
	work *staticStringWorkContext,
	value *staticArrayValue,
	method string,
	arguments []any,
) staticEvalResult {
	if value != nil && value.stringWorkContext == nil {
		value.stringWorkContext = work
	}
	if value == nil || !work.reserve(max(value.length*32, 1)) {
		return staticEvalResult{}
	}
	switch method {
	case "concat":
		return staticArrayConcat(work, value, arguments)
	case "flat":
		depth, ok := staticArgumentInteger(arguments, 0, 1)
		if !ok {
			return staticEvalResult{}
		}
		return staticArrayFlat(work, value, max(depth, 0))
	case "slice":
		start, ok := staticArgumentInteger(arguments, 0, 0)
		if !ok {
			return staticEvalResult{}
		}
		end, ok := staticArgumentInteger(arguments, 1, value.length)
		if !ok {
			return staticEvalResult{}
		}
		return staticEvalResult{value: staticArraySlice(value, start, end), ok: true}
	case "entries", "keys", "values":
		values := make([]any, 0, value.length)
		for index := range value.length {
			element := any(staticUndefinedValue{})
			if !value.omitted[index] {
				element = value.element(index)
			}
			switch method {
			case "entries":
				values = append(values, staticArrayFromValues([]any{staticNumberValue(index), element}))
			case "keys":
				values = append(values, staticNumberValue(index))
			case "values":
				values = append(values, element)
			}
		}
		return staticEvalResult{value: staticIteratorValue{
			values: values, kind: "Array", identity: &staticIdentity{},
		}, ok: true}
	case "every", "filter", "find", "findIndex", "some":
		return staticArrayCallbackCall(work, value, method, arguments)
	case "includes", "indexOf", "lastIndexOf":
		return staticArraySearch(value, method, arguments)
	case "at":
		index, ok := staticArgumentInteger(arguments, 0, 0)
		if !ok {
			return staticEvalResult{}
		}
		if index < 0 {
			index += value.length
		}
		if index < 0 || index >= value.length || value.omitted[index] {
			return staticEvalResult{value: staticUndefinedValue{}, ok: true}
		}
		return staticEvalResult{value: value.element(index), ok: true}
	case "join", "toString":
		separator := ","
		if method == "join" && len(arguments) > 0 && !staticValueUndefined(arguments[0]) {
			var ok bool
			separator, ok = staticValueToString(arguments[0])
			if !ok {
				return staticEvalResult{}
			}
		}
		text, ok := staticArrayJoin(value, separator)
		return staticEvalResult{value: text, ok: ok}
	}
	return staticEvalResult{}
}

type staticArrayElement struct {
	value   any
	omitted bool
}

func staticArrayFromElements(elements []staticArrayElement) *staticArrayValue {
	array := &staticArrayValue{length: len(elements)}
	if array.length > len(array.inline) {
		array.overflow = make([]any, array.length-len(array.inline))
	}
	for index, element := range elements {
		array.set(index, element.value)
		if element.omitted {
			if array.omitted == nil {
				array.omitted = make(map[int]bool, 1)
			}
			array.omitted[index] = true
		}
	}
	return array
}

func staticArrayElements(value *staticArrayValue) []staticArrayElement {
	elements := make([]staticArrayElement, value.length)
	for index := range value.length {
		elements[index] = staticArrayElement{
			value: value.element(index), omitted: value.omitted[index],
		}
	}
	return elements
}

func staticArrayConcat(
	work *staticStringWorkContext,
	receiver *staticArrayValue,
	arguments []any,
) staticEvalResult {
	if receiver.length > maxStaticAggregateElements {
		return staticEvalResult{}
	}
	elements := staticArrayElements(receiver)
	for _, argument := range arguments {
		if array, ok := argument.(*staticArrayValue); ok {
			if array.length > maxStaticAggregateElements-len(elements) ||
				!work.reserve(max(array.length*32, 1)) {
				return staticEvalResult{}
			}
			elements = append(elements, staticArrayElements(array)...)
			continue
		}
		if object, ok := argument.(*staticObjectValue); ok {
			if object.prototypeSet {
				return staticEvalResult{}
			}
			spreadable, found := staticObjectOwnSymbolProperty(object, staticSymbolValue{
				description: "isConcatSpreadable", wellKnown: true,
			})
			if found {
				truthy, known := staticValueTruthy(spreadable)
				if !known {
					return staticEvalResult{}
				}
				if truthy {
					lengthValue, exists := staticObjectOwnProperty(object, "length")
					if !exists {
						lengthValue = staticUndefinedValue{}
					}
					length, ok := staticToLength(lengthValue)
					if !ok || length > maxStaticAggregateElements-len(elements) ||
						!work.reserve(max(length*32, 1)) {
						return staticEvalResult{}
					}
					for index := range length {
						value, exists := staticObjectOwnProperty(object, strconv.Itoa(index))
						elements = append(elements, staticArrayElement{value: value, omitted: !exists})
					}
					continue
				}
			}
		}
		if len(elements) >= maxStaticAggregateElements {
			return staticEvalResult{}
		}
		elements = append(elements, staticArrayElement{value: argument})
	}
	return staticEvalResult{value: staticArrayFromElements(elements), ok: true}
}

func staticArraySlice(value *staticArrayValue, start, end int) *staticArrayValue {
	from := normalizeSliceIndex(start, value.length)
	to := normalizeSliceIndex(end, value.length)
	if to < from {
		to = from
	}
	return staticArrayFromElements(staticArrayElements(value)[from:to])
}

func staticArrayFlat(
	work *staticStringWorkContext,
	value *staticArrayValue,
	depth int,
) staticEvalResult {
	elements := make([]staticArrayElement, 0, value.length)
	var appendArray func(array *staticArrayValue, currentDepth int) bool
	appendArray = func(array *staticArrayValue, currentDepth int) bool {
		for index := range array.length {
			if array.omitted[index] {
				continue
			}
			element := array.element(index)
			if nested, ok := element.(*staticArrayValue); ok && currentDepth > 0 {
				if !appendArray(nested, currentDepth-1) {
					return false
				}
				continue
			}
			if len(elements) >= maxStaticAggregateElements {
				return false
			}
			if !work.reserve(16) {
				return false
			}
			elements = append(elements, staticArrayElement{value: element})
		}
		return true
	}
	if !appendArray(value, depth) {
		return staticEvalResult{}
	}
	return staticEvalResult{value: staticArrayFromElements(elements), ok: true}
}

func staticArraySearch(value *staticArrayValue, method string, arguments []any) staticEvalResult {
	needle := staticArgument(arguments, 0)
	defaultIndex := 0
	if method == "lastIndexOf" {
		defaultIndex = value.length - 1
	}
	var fromIndex int
	var ok bool
	if method == "lastIndexOf" && len(arguments) > 1 && staticValueUndefined(arguments[1]) {
		// Unlike an omitted argument, an explicit undefined is converted to 0.
		fromIndex, ok = 0, true
	} else {
		fromIndex, ok = staticArgumentInteger(arguments, 1, defaultIndex)
	}
	if !ok {
		if len(arguments) > 1 && staticCanConvertToNumber(arguments[1]) {
			if method == "includes" {
				return staticEvalResult{value: staticUnknownBooleanValue{}, ok: true}
			}
			return staticEvalResult{value: staticUnknownNumberValue{}, ok: true}
		}
		return staticEvalResult{}
	}
	start, end, step := fromIndex, value.length, 1
	if method == "lastIndexOf" {
		if start < 0 {
			start += value.length
		}
		start = min(start, value.length-1)
		end, step = -1, -1
	} else {
		if start < 0 {
			start = max(value.length+start, 0)
		}
		if start >= value.length {
			if method == "includes" {
				return staticEvalResult{value: false, ok: true}
			}
			return staticEvalResult{value: staticNumberValue(-1), ok: true}
		}
	}
	for index := start; index != end; index += step {
		if index < 0 || index >= value.length {
			break
		}
		if value.omitted[index] && method != "includes" {
			continue
		}
		element := value.element(index)
		if value.omitted[index] {
			element = staticUndefinedValue{}
		}
		var equal, known bool
		if method == "includes" {
			equal, known = staticSameValueZero(element, needle)
		} else {
			equal, known = staticValuesStrictEqual(element, needle)
		}
		if !known {
			if method == "includes" {
				return staticEvalResult{value: staticUnknownBooleanValue{}, ok: true}
			}
			return staticEvalResult{value: staticUnknownNumberValue{}, ok: true}
		}
		if equal {
			if method == "includes" {
				return staticEvalResult{value: true, ok: true}
			}
			return staticEvalResult{value: staticNumberValue(index), ok: true}
		}
	}
	if method == "includes" {
		return staticEvalResult{value: false, ok: true}
	}
	return staticEvalResult{value: staticNumberValue(-1), ok: true}
}

func staticArrayCallbackCall(
	work *staticStringWorkContext,
	value *staticArrayValue,
	method string,
	arguments []any,
) staticEvalResult {
	if len(arguments) == 0 {
		return staticEvalResult{}
	}
	callback, ok := arguments[0].(staticBuiltinValue)
	if !ok {
		return staticEvalResult{}
	}
	hasUnknownPredicate := false
	filtered := make([]any, 0, value.length)
	for index := range value.length {
		if value.omitted[index] && method != "find" && method != "findIndex" {
			continue
		}
		element := value.element(index)
		if value.omitted[index] {
			element = staticUndefinedValue{}
		}
		predicate := evalESLintBuiltinCall(work, string(callback), []any{
			element, staticNumberValue(index), value,
		})
		if !predicate.ok {
			return staticEvalResult{}
		}
		truthy, known := staticValueTruthy(predicate.value)
		if !known {
			hasUnknownPredicate = true
			continue
		}
		switch method {
		case "filter":
			if truthy {
				filtered = append(filtered, element)
			}
		case "every":
			if !truthy {
				return staticEvalResult{value: false, ok: true}
			}
		case "some":
			if truthy {
				return staticEvalResult{value: true, ok: true}
			}
		case "find":
			if truthy {
				if hasUnknownPredicate {
					return staticEvalResult{}
				}
				return staticEvalResult{value: element, ok: true}
			}
		case "findIndex":
			if truthy {
				if hasUnknownPredicate {
					return staticEvalResult{value: staticUnknownNumberValue{}, ok: true}
				}
				return staticEvalResult{value: staticNumberValue(index), ok: true}
			}
		}
	}

	switch method {
	case "filter":
		if hasUnknownPredicate {
			return staticEvalResult{}
		}
		return staticEvalResult{value: staticArrayFromValues(filtered), ok: true}
	case "every":
		if hasUnknownPredicate {
			return staticEvalResult{value: staticUnknownBooleanValue{}, ok: true}
		}
		return staticEvalResult{value: true, ok: true}
	case "some":
		if hasUnknownPredicate {
			return staticEvalResult{value: staticUnknownBooleanValue{}, ok: true}
		}
		return staticEvalResult{value: false, ok: true}
	case "find":
		if hasUnknownPredicate {
			return staticEvalResult{}
		}
		return staticEvalResult{value: staticUndefinedValue{}, ok: true}
	case "findIndex":
		if hasUnknownPredicate {
			return staticEvalResult{value: staticUnknownNumberValue{}, ok: true}
		}
		return staticEvalResult{value: staticNumberValue(-1), ok: true}
	}
	return staticEvalResult{}
}

func staticStringPrototypeCall(
	work *staticStringWorkContext,
	value, method string,
	arguments []any,
) staticEvalResult {
	// String methods repeatedly materialize UTF-16 views and transformed
	// output. Reserve the receiver scan before any method-specific allocation;
	// builders below reserve their output separately.
	if !work.reserve(max(len(value)*3, 1)) {
		return staticEvalResult{}
	}
	for _, argument := range arguments {
		if text, ok := staticValueAsString(argument); ok &&
			!work.reserve(max(len(text)*8, 1)) {
			return staticEvalResult{}
		}
	}
	switch method {
	case "at":
		index, ok := staticArgumentInteger(arguments, 0, 0)
		if !ok {
			return staticEvalResult{}
		}
		units := ecmascript.StringCodeUnits(value)
		if index < 0 {
			index += len(units)
		}
		if index < 0 || index >= len(units) {
			return staticEvalResult{value: staticUndefinedValue{}, ok: true}
		}
		return staticEvalResult{value: ecmascript.StringFromCodeUnits(units[index : index+1]), ok: true}
	case "charAt":
		return staticESLintStringCharAt(value, arguments)
	case "charCodeAt", "codePointAt":
		index, ok := staticArgumentInteger(arguments, 0, 0)
		if !ok {
			return staticEvalResult{}
		}
		units := ecmascript.StringCodeUnits(value)
		if index < 0 || index >= len(units) {
			if method == "charCodeAt" {
				return staticEvalResult{value: staticNumberValue(math.NaN()), ok: true}
			}
			return staticEvalResult{value: staticUndefinedValue{}, ok: true}
		}
		codePoint := uint32(units[index])
		if method == "codePointAt" && units[index] >= 0xD800 && units[index] <= 0xDBFF &&
			index+1 < len(units) && units[index+1] >= 0xDC00 && units[index+1] <= 0xDFFF {
			codePoint = 0x10000 + uint32(units[index]-0xD800)*0x400 + uint32(units[index+1]-0xDC00)
		}
		return staticEvalResult{value: staticNumberValue(codePoint), ok: true}
	case "concat":
		return staticStringConcat(work, value, arguments)
	case "endsWith", "includes", "startsWith":
		return staticStringPredicate(value, method, arguments)
	case "indexOf", "lastIndexOf":
		return staticStringIndexOf(value, method, arguments)
	case "normalize":
		form := "NFC"
		if len(arguments) > 0 && !staticValueUndefined(arguments[0]) {
			var ok bool
			form, ok = staticValueToString(arguments[0])
			if !ok || form != "NFC" && form != "NFD" && form != "NFKC" && form != "NFKD" {
				return staticEvalResult{}
			}
		}
		if staticNormalizationNeedsNewerUnicodeTables(value, form, norm.Version) {
			// Go 1.26 selects x/text's Unicode 15 normalization tables, while
			// Node 22 (the Unicorn v73 oracle) uses Unicode 16. Preserve the
			// known non-empty string type without fabricating a spelling for
			// characters whose decomposition/composition was added in Unicode 16.
			return staticEvalResult{value: staticUnknownStringValue{truthy: true}, ok: true}
		}
		var normalForm norm.Form
		switch form {
		case "NFD":
			normalForm = norm.NFD
		case "NFKC":
			normalForm = norm.NFKC
		case "NFKD":
			normalForm = norm.NFKD
		default:
			normalForm = norm.NFC
		}
		return staticNormalizeString(work, normalForm, value)
	case "padEnd", "padStart":
		targetLength, ok := staticArgumentInteger(arguments, 0, 0)
		if !ok {
			return staticEvalResult{}
		}
		// StringPad returns before converting the fill string when no padding
		// is needed. This is observable for a Symbol filler.
		if targetLength <= ecmascript.StringCodeUnitCount(value) {
			return staticEvalResult{value: value, ok: true}
		}
		filler := " "
		if len(arguments) > 1 && !staticValueUndefined(arguments[1]) {
			filler, ok = staticValueToString(arguments[1])
			if !ok {
				return staticEvalResult{}
			}
		}
		return staticStringPad(work, value, method, targetLength, filler)
	case "slice", "substr", "substring":
		switch method {
		case "slice":
			return staticESLintStringSlice(value, arguments)
		case "substr":
			return staticESLintStringSubstr(value, arguments)
		default:
			return staticESLintStringSubstring(value, arguments)
		}
	case "toLowerCase":
		if ecmascript.StringCodeUnitCount(value) > maxStaticStringLength/3 {
			return staticEvalResult{}
		}
		if staticCaseConversionNeedsNewerRuntime(value, false) {
			return staticEvalResult{value: staticUnknownStringValue{truthy: true}, ok: true}
		}
		return staticEvalResult{value: ecmascript.StringToLowerCase(value), ok: true}
	case "toUpperCase":
		if ecmascript.StringCodeUnitCount(value) > maxStaticStringLength/3 {
			return staticEvalResult{}
		}
		if staticCaseConversionNeedsNewerRuntime(value, true) {
			return staticEvalResult{value: staticUnknownStringValue{truthy: true}, ok: true}
		}
		return staticEvalResult{value: ecmascript.StringToUpperCase(value), ok: true}
	case "toString":
		return staticEvalResult{value: value, ok: true}
	case "trim":
		return staticEvalResult{value: ecmascript.StringTrim(value), ok: true}
	case "trimEnd", "trimRight":
		return staticEvalResult{value: ecmascript.StringTrimEnd(value), ok: true}
	case "trimLeft", "trimStart":
		return staticEvalResult{value: ecmascript.StringTrimStart(value), ok: true}
	}
	return staticEvalResult{}
}

func staticNormalizationNeedsNewerUnicodeTables(value, form, tableVersion string) bool {
	if tableVersion != "15.0.0" {
		// The repository is pinned to Go 1.26, which selects x/text's Unicode
		// 15 tables. If a future toolchain selects different tables, retain exact
		// ASCII normalization and keep every other result abstract until that
		// table/runtime pair has been audited.
		for _, character := range value {
			if character >= utf8.RuneSelf {
				return true
			}
		}
		return false
	}
	compatibility := form == "NFKC" || form == "NFKD"
	for _, character := range value {
		if compatibility && character >= 0x1CCD6 && character <= 0x1CCF9 {
			return true
		}
		switch character {
		case 0x105C9, 0x105D2, 0x105DA, 0x105E4,
			0x0897,
			0x11382, 0x11383, 0x11384, 0x11385, 0x1138B, 0x1138E,
			0x11390, 0x11391, 0x113B8, 0x113BB, 0x113C2, 0x113C5,
			0x113C7, 0x113C8, 0x113C9, 0x113CE, 0x113CF, 0x113D0,
			0x1612F,
			0x16D63, 0x16D67, 0x16D68, 0x16D69, 0x16D6A:
			return true
		}
		if character >= 0x10D69 && character <= 0x10D6D ||
			character >= 0x1611E && character <= 0x16129 ||
			character >= 0x1E5EE && character <= 0x1E5EF {
			return true
		}
	}
	return false
}

// staticCaseConversionNeedsNewerRuntime keeps the enhanced evaluator aligned
// with Unicorn v73's Node 22 / Unicode 16 runtime. The shared ecmascript case
// helpers intentionally target Node 26 / Unicode 17 for production rules, so
// the Unicode 17-only mappings and Final_Sigma properties must stay abstract
// here rather than choosing a receiver branch with the newer spelling.
func staticCaseConversionNeedsNewerRuntime(value string, upper bool) bool {
	hasFinalSigma := !upper && strings.ContainsRune(value, '\u03A3')
	for _, character := range value {
		if staticUnicode17CaseMappingSource(character, upper) ||
			hasFinalSigma && staticUnicode17FinalSigmaPropertyChange(character) {
			return true
		}
	}
	return false
}

func staticUnicode17CaseMappingSource(character rune, upper bool) bool {
	if upper {
		return character == 0xA7CF || character == 0xA7D3 || character == 0xA7D5 ||
			character >= 0x16EBB && character <= 0x16ED3
	}
	return character == 0xA7CE || character == 0xA7D2 || character == 0xA7D4 ||
		character >= 0x16EA0 && character <= 0x16EB8
}

func staticUnicode17FinalSigmaPropertyChange(character rune) bool {
	switch character {
	case 0x0295,
		0xA7CE, 0xA7CF, 0xA7D2, 0xA7D4, 0xA7F1,
		0x10EC5, 0x11B60, 0x11B66, 0x11DD9,
		0x1E6E3, 0x1E6E6, 0x1E6F5, 0x1E6FF:
		return true
	}
	return character >= 0x1ACF && character <= 0x1ADD ||
		character >= 0x1AE0 && character <= 0x1AEB ||
		character >= 0x10EFA && character <= 0x10EFB ||
		character >= 0x11B62 && character <= 0x11B64 ||
		character >= 0x16EA0 && character <= 0x16EB8 ||
		character >= 0x16EBB && character <= 0x16ED3 ||
		character >= 0x16FF2 && character <= 0x16FF3 ||
		character >= 0x1E6EE && character <= 0x1E6EF
}

func staticNormalizeString(
	work *staticStringWorkContext,
	form norm.Form,
	value string,
) staticEvalResult {
	if !work.reserve(max(len(value)*20, 1)) {
		return staticEvalResult{}
	}
	var iterator norm.Iter
	var builder strings.Builder
	builder.Grow(min(len(value), maxStaticStringLength*3))
	chunkStart := 0
	for index := 0; index+2 < len(value); index++ {
		if value[index] != 0xED || value[index+1] < 0xA0 || value[index+1] > 0xBF ||
			value[index+2] < 0x80 || value[index+2] > 0xBF {
			continue
		}
		if !appendStaticNormalizedChunk(&builder, &iterator, form, value[chunkStart:index]) ||
			builder.Len() > maxStaticStringLength*3-3 {
			work.exhaust()
			return staticEvalResult{}
		}
		// Compiler strings carry an unpaired UTF-16 surrogate as its three WTF-8
		// bytes. It is a normalization barrier: preserve it verbatim, while still
		// normalizing the well-formed chunks on either side.
		builder.WriteString(value[index : index+3])
		index += 2
		chunkStart = index + 1
	}
	if !appendStaticNormalizedChunk(&builder, &iterator, form, value[chunkStart:]) {
		work.exhaust()
		return staticEvalResult{}
	}
	return staticEvalResult{value: builder.String(), ok: true}
}

func appendStaticNormalizedChunk(
	builder *strings.Builder,
	iterator *norm.Iter,
	form norm.Form,
	value string,
) bool {
	iterator.InitString(form, value)
	for !iterator.Done() {
		segment := iterator.Next()
		if len(segment) > maxStaticStringLength*3-builder.Len() {
			return false
		}
		builder.Write(segment)
	}
	return true
}

func staticESLintStringFromCharCode(
	work *staticStringWorkContext,
	arguments []any,
) staticEvalResult {
	if !work.reserve(max(len(arguments)*6, 1)) {
		return staticEvalResult{}
	}
	units := make([]uint16, 0, len(arguments))
	for _, argument := range arguments {
		number, ok := staticValueToNumber(argument)
		if !ok {
			return staticEvalResult{}
		}
		units = append(units, uint16(toUint32(number)))
	}
	return staticEvalResult{value: ecmascript.StringFromCodeUnits(units), ok: true}
}

func staticESLintStringSlice(text string, arguments []any) staticEvalResult {
	start, ok := staticArgumentInteger(arguments, 0, 0)
	if !ok {
		return staticEvalResult{}
	}
	end, ok := staticArgumentInteger(arguments, 1, math.MaxInt)
	if !ok {
		return staticEvalResult{}
	}
	units := ecmascript.StringCodeUnits(text)
	from, to := normalizeSliceIndex(start, len(units)), normalizeSliceIndex(end, len(units))
	if to < from {
		to = from
	}
	return staticEvalResult{value: ecmascript.StringFromCodeUnits(units[from:to]), ok: true}
}

func staticESLintStringSubstring(text string, arguments []any) staticEvalResult {
	start, ok := staticArgumentInteger(arguments, 0, 0)
	if !ok {
		return staticEvalResult{}
	}
	end, ok := staticArgumentInteger(arguments, 1, math.MaxInt)
	if !ok {
		return staticEvalResult{}
	}
	units := ecmascript.StringCodeUnits(text)
	from, to := clampSubstringIndex(start, len(units)), clampSubstringIndex(end, len(units))
	if from > to {
		from, to = to, from
	}
	return staticEvalResult{value: ecmascript.StringFromCodeUnits(units[from:to]), ok: true}
}

func staticESLintStringSubstr(text string, arguments []any) staticEvalResult {
	start, ok := staticArgumentInteger(arguments, 0, 0)
	if !ok {
		return staticEvalResult{}
	}
	count, ok := staticArgumentInteger(arguments, 1, math.MaxInt)
	if !ok {
		return staticEvalResult{}
	}
	units := ecmascript.StringCodeUnits(text)
	from := normalizeSliceIndex(start, len(units))
	if count <= 0 {
		return staticEvalResult{value: "", ok: true}
	}
	to := len(units)
	if count < len(units)-from {
		to = from + count
	}
	return staticEvalResult{value: ecmascript.StringFromCodeUnits(units[from:to]), ok: true}
}

func staticESLintStringCharAt(text string, arguments []any) staticEvalResult {
	index, ok := staticArgumentInteger(arguments, 0, 0)
	if !ok {
		return staticEvalResult{}
	}
	units := ecmascript.StringCodeUnits(text)
	if index < 0 || index >= len(units) {
		return staticEvalResult{value: "", ok: true}
	}
	return staticEvalResult{value: ecmascript.StringFromCodeUnits(units[index : index+1]), ok: true}
}

func staticStringPredicate(value, method string, arguments []any) staticEvalResult {
	searchValue := staticArgument(arguments, 0)
	if _, isRegExp := searchValue.(staticRegExpValue); isRegExp {
		return staticEvalResult{}
	}
	if staticValueIsObjectForCoercion(searchValue) {
		match := staticESLintSymbolMemberValue(searchValue, staticSymbolValue{
			description: "match", wellKnown: true,
		})
		if !match.ok {
			return staticEvalResult{}
		}
		if !staticValueUndefined(match.value) {
			truthy, known := staticValueTruthy(match.value)
			if !known || truthy {
				return staticEvalResult{}
			}
		}
	}
	search, ok := staticValueToString(searchValue)
	if !ok {
		return staticEvalResult{}
	}
	units := ecmascript.StringCodeUnits(value)
	searchUnits := ecmascript.StringCodeUnits(search)
	positionDefault := 0
	if method == "endsWith" {
		positionDefault = len(units)
	}
	position, ok := staticArgumentInteger(arguments, 1, positionDefault)
	if !ok {
		return staticEvalResult{}
	}
	position = min(max(position, 0), len(units))
	var matches bool
	switch method {
	case "startsWith":
		matches = staticCodeUnitsEqualAt(units, searchUnits, position)
	case "endsWith":
		matches = staticCodeUnitsEqualAt(units, searchUnits, position-len(searchUnits))
	default:
		matches = staticCodeUnitsIndex(units, searchUnits, position) >= 0
	}
	return staticEvalResult{value: matches, ok: true}
}

func staticStringIndexOf(value, method string, arguments []any) staticEvalResult {
	search, ok := staticValueToString(staticArgument(arguments, 0))
	if !ok {
		return staticEvalResult{}
	}
	units := ecmascript.StringCodeUnits(value)
	searchUnits := ecmascript.StringCodeUnits(search)
	defaultIndex := 0
	if method == "lastIndexOf" {
		defaultIndex = len(units)
	}
	var position int
	if method == "lastIndexOf" && len(arguments) > 1 {
		if number, numberOK := staticValueToNumber(arguments[1]); numberOK && math.IsNaN(number) {
			position, ok = len(units), true
		} else {
			position, ok = staticArgumentInteger(arguments, 1, defaultIndex)
		}
	} else {
		position, ok = staticArgumentInteger(arguments, 1, defaultIndex)
	}
	if !ok {
		return staticEvalResult{}
	}
	position = min(max(position, 0), len(units))
	var index int
	if method == "lastIndexOf" {
		index = staticCodeUnitsLastIndex(units, searchUnits, position)
	} else {
		index = staticCodeUnitsIndex(units, searchUnits, position)
	}
	return staticEvalResult{value: staticNumberValue(index), ok: true}
}

func staticCodeUnitsIndex(value, search []uint16, start int) int {
	start = max(start, 0)
	if len(search) == 0 {
		return min(start, len(value))
	}
	if start > len(value)-len(search) {
		return -1
	}
	prefix := staticCodeUnitsPrefix(search)
	matched := 0
	for index := start; index < len(value); index++ {
		for matched > 0 && value[index] != search[matched] {
			matched = prefix[matched-1]
		}
		if value[index] == search[matched] {
			matched++
		}
		if matched == len(search) {
			return index - len(search) + 1
		}
	}
	return -1
}

func staticCodeUnitsLastIndex(value, search []uint16, position int) int {
	limit := min(max(position, 0), len(value))
	if len(search) == 0 {
		return limit
	}
	limit = min(limit, len(value)-len(search))
	if limit < 0 {
		return -1
	}
	prefix := staticCodeUnitsPrefix(search)
	matched, last := 0, -1
	for index, unit := range value {
		for matched > 0 && unit != search[matched] {
			matched = prefix[matched-1]
		}
		if unit == search[matched] {
			matched++
		}
		if matched == len(search) {
			start := index - len(search) + 1
			if start > limit {
				break
			}
			last = start
			matched = prefix[matched-1]
		}
	}
	return last
}

func staticCodeUnitsPrefix(pattern []uint16) []int {
	prefix := make([]int, len(pattern))
	for index, matched := 1, 0; index < len(pattern); index++ {
		for matched > 0 && pattern[index] != pattern[matched] {
			matched = prefix[matched-1]
		}
		if pattern[index] == pattern[matched] {
			matched++
		}
		prefix[index] = matched
	}
	return prefix
}

func staticCodeUnitsEqualAt(value, search []uint16, index int) bool {
	if index < 0 || index+len(search) > len(value) {
		return false
	}
	for offset, unit := range search {
		if value[index+offset] != unit {
			return false
		}
	}
	return true
}

func staticStringPad(
	work *staticStringWorkContext,
	value, method string,
	targetLength int,
	filler string,
) staticEvalResult {
	units := ecmascript.StringCodeUnits(value)
	if targetLength <= len(units) || filler == "" {
		return staticEvalResult{value: value, ok: true}
	}
	if targetLength > maxStaticStringLength {
		if targetLength > maxJavaScriptStringLength {
			return staticEvalResult{}
		}
		return staticEvalResult{value: staticUnknownStringValue{truthy: true}, ok: true}
	}
	if !work.reserve(max(targetLength*8, 1)) {
		return staticEvalResult{}
	}
	fillerUnits := ecmascript.StringCodeUnits(filler)
	if len(fillerUnits) == 0 {
		return staticEvalResult{value: value, ok: true}
	}
	padding := make([]uint16, targetLength-len(units))
	for index := range padding {
		padding[index] = fillerUnits[index%len(fillerUnits)]
	}
	if len(units) == 0 {
		return staticEvalResult{value: ecmascript.StringFromCodeUnits(padding), ok: true}
	}
	if method == "padStart" {
		result := make([]uint16, len(padding)+len(units))
		copy(result, padding)
		copy(result[len(padding):], units)
		return staticEvalResult{value: ecmascript.StringFromCodeUnits(result), ok: true}
	}
	units = append(units, padding...)
	return staticEvalResult{value: ecmascript.StringFromCodeUnits(units), ok: true}
}

func staticNumberPrototypeCall(value any, method string, arguments []any) staticEvalResult {
	if method != "toExponential" && method != "toFixed" && method != "toPrecision" && method != "toString" {
		return staticEvalResult{}
	}
	receiver, receiverKnown := staticValueToNumber(value)
	if !receiverKnown {
		if staticValueKindOf(value) == staticKindNumber {
			return staticEvalResult{value: staticUnknownStringValue{truthy: true}, ok: true}
		}
		return staticEvalResult{}
	}
	digits, ok := staticArgumentInteger(arguments, 0, 0)
	if !ok {
		return staticEvalResult{}
	}
	switch method {
	case "toString":
		radix := 10
		if len(arguments) > 0 && !staticValueUndefined(arguments[0]) {
			radix = digits
		}
		if radix < 2 || radix > 36 {
			return staticEvalResult{}
		}
		if radix == 10 {
			return staticEvalResult{value: ecmascript.NumberToString(receiver), ok: true}
		}
		if formatted, ok := staticNumberToRadix(receiver, radix); ok {
			return staticEvalResult{value: formatted, ok: true}
		}
		return staticEvalResult{value: staticUnknownStringValue{truthy: true}, ok: true}
	case "toPrecision":
		if len(arguments) == 0 || staticValueUndefined(arguments[0]) {
			return staticEvalResult{value: ecmascript.NumberToString(receiver), ok: true}
		}
		if math.IsNaN(receiver) || math.IsInf(receiver, 0) {
			return staticEvalResult{value: ecmascript.NumberToString(receiver), ok: true}
		}
		if digits < 1 || digits > 100 {
			return staticEvalResult{}
		}
		return staticEvalResult{value: ecmascript.NumberToPrecision(receiver, digits), ok: true}
	case "toFixed":
		if digits < 0 || digits > 100 {
			return staticEvalResult{}
		}
		if math.IsNaN(receiver) || math.IsInf(receiver, 0) || math.Abs(receiver) >= 1e21 {
			return staticEvalResult{value: ecmascript.NumberToString(receiver), ok: true}
		}
		return staticEvalResult{value: staticNumberToFixed(receiver, digits), ok: true}
	case "toExponential":
		if math.IsNaN(receiver) || math.IsInf(receiver, 0) {
			return staticEvalResult{value: ecmascript.NumberToString(receiver), ok: true}
		}
		if len(arguments) > 0 && !staticValueUndefined(arguments[0]) && (digits < 0 || digits > 100) {
			return staticEvalResult{}
		}
		fractionDigits := -1
		if len(arguments) > 0 && !staticValueUndefined(arguments[0]) {
			fractionDigits = digits
		}
		return staticEvalResult{value: ecmascript.NumberToExponential(receiver, fractionDigits), ok: true}
	}
	return staticEvalResult{}
}

func staticNumberToFixed(number float64, digits int) string {
	negative := number < 0
	if negative {
		number = -number
	}
	rat := new(big.Rat).SetFloat64(number)
	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(digits)), nil)
	scaledNumerator := new(big.Int).Mul(new(big.Int).Set(rat.Num()), scale)
	integer, remainder := new(big.Int), new(big.Int)
	integer.QuoRem(scaledNumerator, rat.Denom(), remainder)
	if new(big.Int).Lsh(remainder, 1).Cmp(rat.Denom()) >= 0 {
		integer.Add(integer, big.NewInt(1))
	}
	text := integer.String()
	if digits > 0 {
		if len(text) <= digits {
			text = strings.Repeat("0", digits-len(text)+1) + text
		}
		text = text[:len(text)-digits] + "." + text[len(text)-digits:]
	}
	if negative {
		text = "-" + text
	}
	return text
}

func staticNumberToRadix(number float64, radix int) (string, bool) {
	if math.IsNaN(number) || math.IsInf(number, 0) {
		return ecmascript.NumberToString(number), true
	}
	if number == 0 {
		return "0", true
	}
	negative := number < 0
	if negative {
		number = -number
	}
	const maxSafeInteger = 9007199254740991
	if (math.Trunc(number) != number || number > maxSafeInteger) && radix&(radix-1) != 0 {
		// V8 chooses a shortest round-tripping representation for non-binary
		// radices, not the exact digits of the float's rational value. Exact
		// conversion is portable only for safe integers and power-of-two bases.
		return "", false
	}
	rat := new(big.Rat).SetFloat64(number)
	if rat == nil {
		return "", false
	}
	integer, remainder := new(big.Int), new(big.Int)
	integer.QuoRem(rat.Num(), rat.Denom(), remainder)
	text := integer.Text(radix)
	if remainder.Sign() != 0 {
		if radix%2 != 0 {
			return "", false
		}
		const digits = "0123456789abcdefghijklmnopqrstuvwxyz"
		var fraction strings.Builder
		base := big.NewInt(int64(radix))
		for remainder.Sign() != 0 && fraction.Len() <= 4096 {
			remainder.Mul(remainder, base)
			digit := new(big.Int)
			digit.QuoRem(remainder, rat.Denom(), remainder)
			fraction.WriteByte(digits[digit.Int64()])
		}
		if remainder.Sign() != 0 {
			return "", false
		}
		text += "." + fraction.String()
	}
	if negative {
		text = "-" + text
	}
	return text, true
}

func staticCanConvertToString(value any) bool {
	switch value.(type) {
	case staticSymbolValue:
		return false
	case staticUnknownStringValue, staticUnknownNumberValue, staticUnknownBooleanValue:
		return true
	}
	_, ok := staticValueToString(value)
	return ok
}

func staticCanConvertToNumber(value any) bool {
	switch value.(type) {
	case staticBigIntValue, staticSymbolValue:
		return false
	case staticUnknownStringValue, staticUnknownNumberValue, staticUnknownBooleanValue:
		return true
	}
	_, ok := staticValueToNumber(value)
	return ok
}

func staticNumericArguments(arguments []any, start, count int) bool {
	end := min(len(arguments), start+count)
	for index := start; index < end; index++ {
		if !staticValueUndefined(arguments[index]) && !staticCanConvertToNumber(arguments[index]) {
			return false
		}
	}
	return true
}

func staticSameValue(left, right any) (bool, bool) {
	if leftNumber, leftOK := staticValueToNumber(left); leftOK && staticValueKindOf(left) == staticKindNumber {
		rightNumber, rightOK := staticValueToNumber(right)
		if !rightOK || staticValueKindOf(right) != staticKindNumber {
			return false, true
		}
		if math.IsNaN(leftNumber) && math.IsNaN(rightNumber) {
			return true, true
		}
		if leftNumber == 0 && rightNumber == 0 {
			return math.Signbit(leftNumber) == math.Signbit(rightNumber), true
		}
		return leftNumber == rightNumber, true
	}
	return staticValuesStrictEqual(left, right)
}

func staticSameValueZero(left, right any) (bool, bool) {
	if leftNumber, leftOK := staticValueToNumber(left); leftOK && staticValueKindOf(left) == staticKindNumber {
		rightNumber, rightOK := staticValueToNumber(right)
		if !rightOK || staticValueKindOf(right) != staticKindNumber {
			return false, true
		}
		return leftNumber == rightNumber || math.IsNaN(leftNumber) && math.IsNaN(rightNumber), true
	}
	switch left := left.(type) {
	case *staticArrayValue:
		right, ok := right.(*staticArrayValue)
		if !ok {
			return false, true
		}
		return left == right, true
	case *staticObjectValue:
		right, ok := right.(*staticObjectValue)
		if !ok {
			return false, true
		}
		return left == right, true
	}
	return staticValuesStrictEqual(left, right)
}
