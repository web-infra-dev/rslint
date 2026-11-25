// cspell:ignore Exponentiate

package utils

import (
	"math/big"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/jsnum"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

// CompareLiteralTypes evaluates a JavaScript comparison between singleton
// primitive types using the same coercions as static expression evaluation.
func CompareLiteralTypes(left, right *checker.Type, operator ast.Kind) (value, known bool) {
	l, leftOK := staticValueFromLiteralType(left)
	r, rightOK := staticValueFromLiteralType(right)
	if !leftOK || !rightOK {
		return false, false
	}
	if operator == ast.KindEqualsEqualsEqualsToken || operator == ast.KindExclamationEqualsEqualsToken {
		equal, ok := staticValuesStrictEqual(l, r)
		if operator == ast.KindExclamationEqualsEqualsToken {
			equal = !equal
		}
		return equal, ok
	}
	var evaluator StaticStringEvaluator
	result := evaluator.evalBinaryValues(operator, l, r)
	value, isBoolean := result.value.(bool)
	return value, result.ok && isBoolean
}

func staticValueFromLiteralType(t *checker.Type) (any, bool) {
	if IsTrueLiteralType(t) {
		return true, true
	}
	if IsFalseLiteralType(t) {
		return false, true
	}
	switch checker.Type_flags(t) {
	case checker.TypeFlagsNull:
		return staticNullValue{}, true
	case checker.TypeFlagsUndefined:
		return staticUndefinedValue{}, true
	}
	if t.IsBigIntLiteral() {
		value, ok := t.AsLiteralType().Value().(jsnum.PseudoBigInt)
		if !ok {
			return nil, false
		}
		integer, ok := new(big.Int).SetString(value.String(), 10)
		return integer, ok
	}
	if t.IsStringLiteral() || t.IsNumberLiteral() {
		return t.AsLiteralType().Value(), true
	}
	return nil, false
}

// Supplement tsgo's literal arithmetic with JavaScript coercions and comparisons.
// Number operators still use tsgo's jsnum methods; BigInt keeps its own type rules.
func (staticEvaluator *StaticStringEvaluator) evalBinaryValues(operator ast.Kind, left, right any) staticEvalResult {
	switch operator {
	case ast.KindEqualsEqualsToken, ast.KindExclamationEqualsToken:
		equal, known := staticValuesLooseEqual(left, right)
		if operator == ast.KindExclamationEqualsToken {
			equal = !equal
		}
		return staticEvalResult{value: equal, ok: known}
	case ast.KindPlusToken, ast.KindMinusToken, ast.KindAsteriskToken, ast.KindSlashToken,
		ast.KindPercentToken, ast.KindAsteriskAsteriskToken, ast.KindBarToken, ast.KindAmpersandToken,
		ast.KindCaretToken, ast.KindLessThanLessThanToken, ast.KindGreaterThanGreaterThanToken,
		ast.KindGreaterThanGreaterThanGreaterThanToken, ast.KindLessThanToken, ast.KindLessThanEqualsToken,
		ast.KindGreaterThanToken, ast.KindGreaterThanEqualsToken:
	default:
		return staticEvalResult{}
	}
	// Supported aggregate values use their ordinary string primitive. Reuse the
	// existing conversion so custom conversion methods remain conservative.
	if staticValueIsAggregate(left) {
		value, ok := staticValueToString(left)
		if !ok {
			return staticEvalResult{}
		}
		left = value
	}
	if staticValueIsAggregate(right) {
		value, ok := staticValueToString(right)
		if !ok {
			return staticEvalResult{}
		}
		right = value
	}
	if operator == ast.KindPlusToken && (staticValueIsString(left) || staticValueIsString(right)) {
		return staticEvaluator.concatStaticValues(left, right)
	}
	if result, handled := evalStaticBigIntBinary(operator, left, right); handled {
		return result
	}

	// Two string primitives compare UTF-16 code units instead of numeric values.
	comparison := operator == ast.KindLessThanToken || operator == ast.KindLessThanEqualsToken ||
		operator == ast.KindGreaterThanToken || operator == ast.KindGreaterThanEqualsToken
	leftText, leftString := staticValueAsString(left)
	rightText, rightString := staticValueAsString(right)
	var l, r staticNumberValue
	if comparison && leftString && rightString {
		l = staticNumberValue(ecmascript.CompareStrings(leftText, rightText))
	} else {
		leftNumber, leftOK := staticValueToNumber(left)
		rightNumber, rightOK := staticValueToNumber(right)
		if !leftOK || !rightOK {
			return staticEvalResult{}
		}
		l, r = staticNumberValue(leftNumber), staticNumberValue(rightNumber)
	}
	var value staticNumberValue
	switch operator {
	case ast.KindPlusToken:
		value = l + r
	case ast.KindMinusToken:
		value = l - r
	case ast.KindAsteriskToken:
		value = l * r
	case ast.KindSlashToken:
		value = l / r
	case ast.KindPercentToken:
		value = l.Remainder(r)
	case ast.KindAsteriskAsteriskToken:
		value = l.Exponentiate(r)
	case ast.KindBarToken:
		value = l.BitwiseOR(r)
	case ast.KindAmpersandToken:
		value = l.BitwiseAND(r)
	case ast.KindCaretToken:
		value = l.BitwiseXOR(r)
	case ast.KindLessThanLessThanToken:
		value = l.LeftShift(r)
	case ast.KindGreaterThanGreaterThanToken:
		value = l.SignedRightShift(r)
	case ast.KindGreaterThanGreaterThanGreaterThanToken:
		value = l.UnsignedRightShift(r)
	case ast.KindLessThanToken:
		return staticEvalResult{value: l < r, ok: true}
	case ast.KindLessThanEqualsToken:
		return staticEvalResult{value: l <= r, ok: true}
	case ast.KindGreaterThanToken:
		return staticEvalResult{value: l > r, ok: true}
	case ast.KindGreaterThanEqualsToken:
		return staticEvalResult{value: l >= r, ok: true}
	}
	return staticEvalResult{value: value, ok: true}
}

func staticValuesLooseEqual(left, right any) (bool, bool) {
	if staticValueKindOf(left) == staticValueKindOf(right) {
		return staticValuesStrictEqual(left, right)
	}
	if staticValueNullish(left) || staticValueNullish(right) {
		return staticValueNullish(left) && staticValueNullish(right), true
	}
	if result, handled := evalStaticBigIntBinary(ast.KindEqualsEqualsToken, left, right); handled {
		value, _ := result.value.(bool)
		return value, result.ok
	}
	if staticValueIsAggregate(left) {
		value, ok := staticValueToString(left)
		if !ok {
			return false, false
		}
		return staticValuesLooseEqual(value, right)
	}
	if staticValueIsAggregate(right) {
		return staticValuesLooseEqual(right, left)
	}
	leftNumber, leftOK := staticValueToNumber(left)
	rightNumber, rightOK := staticValueToNumber(right)
	return leftNumber == rightNumber, leftOK && rightOK
}
