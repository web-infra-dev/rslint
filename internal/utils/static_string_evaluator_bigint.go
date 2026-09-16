package utils

import (
	"math"
	"math/big"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

// Bound constant folding independently of source-controlled exponents/shifts.
const maxStaticBigIntBits = 1 << 20

// Operands are immutable: cached initializers and aggregate members may share
// them. math/big supplies the arithmetic; this adapter enforces JS operators,
// throwing cases and mixed-type comparisons that tsgo does not fold.
func evalStaticBigIntBinary(operator ast.Kind, left, right any) (staticEvalResult, bool) {
	l, leftBig := left.(*big.Int)
	r, rightBig := right.(*big.Int)
	if !leftBig && !rightBig {
		return staticEvalResult{}, false
	}
	switch operator {
	case ast.KindEqualsEqualsToken, ast.KindExclamationEqualsToken,
		ast.KindLessThanToken, ast.KindLessThanEqualsToken,
		ast.KindGreaterThanToken, ast.KindGreaterThanEqualsToken:
		var order int
		var ordered, known bool
		equality := operator == ast.KindEqualsEqualsToken || operator == ast.KindExclamationEqualsToken
		if leftBig {
			order, ordered, known = compareStaticBigInt(l, right, equality)
		} else {
			order, ordered, known = compareStaticBigInt(r, left, equality)
			order = -order
		}
		if !known {
			return staticEvalResult{}, true
		}
		var result bool
		switch operator {
		case ast.KindEqualsEqualsToken:
			result = ordered && order == 0
		case ast.KindExclamationEqualsToken:
			result = !ordered || order != 0
		case ast.KindLessThanToken:
			result = ordered && order < 0
		case ast.KindLessThanEqualsToken:
			result = ordered && order <= 0
		case ast.KindGreaterThanToken:
			result = ordered && order > 0
		case ast.KindGreaterThanEqualsToken:
			result = ordered && order >= 0
		}
		return staticEvalResult{value: result, ok: true}, true
	}
	if !leftBig || !rightBig {
		// String concatenation was handled by the caller. Arithmetic cannot
		// mix Number/Boolean/etc. with BigInt, even when values are integral.
		return staticEvalResult{}, true
	}
	value := new(big.Int)
	switch operator {
	case ast.KindPlusToken:
		value.Add(l, r)
	case ast.KindMinusToken:
		value.Sub(l, r)
	case ast.KindAsteriskToken:
		if l.BitLen()+r.BitLen() > maxStaticBigIntBits {
			return staticEvalResult{}, true
		}
		value.Mul(l, r)
	case ast.KindSlashToken, ast.KindPercentToken:
		if r.Sign() == 0 {
			return staticEvalResult{}, true
		}
		if operator == ast.KindSlashToken {
			value.Quo(l, r) // JS truncates toward zero, not negative infinity.
		} else {
			value.Rem(l, r)
		}
	case ast.KindAsteriskAsteriskToken:
		if r.Sign() < 0 || !r.IsUint64() || r.Uint64() > maxStaticBigIntBits ||
			uint64(l.BitLen())*r.Uint64() > maxStaticBigIntBits {
			return staticEvalResult{}, true
		}
		value.Exp(l, r, nil)
	case ast.KindBarToken:
		value.Or(l, r)
	case ast.KindAmpersandToken:
		value.And(l, r)
	case ast.KindCaretToken:
		value.Xor(l, r)
	case ast.KindLessThanLessThanToken, ast.KindGreaterThanGreaterThanToken:
		shift := new(big.Int).Abs(r)
		leftShift := (operator == ast.KindLessThanLessThanToken) == (r.Sign() >= 0)
		if !shift.IsUint64() || shift.Uint64() > maxStaticBigIntBits {
			if leftShift && l.Sign() != 0 {
				return staticEvalResult{}, true
			}
			if !leftShift && l.Sign() < 0 {
				value.SetInt64(-1)
			}
		} else if leftShift {
			if uint64(l.BitLen())+shift.Uint64() > maxStaticBigIntBits {
				return staticEvalResult{}, true
			}
			value.Lsh(l, uint(shift.Uint64()))
		} else {
			value.Rsh(l, uint(shift.Uint64()))
		}
	default:
		// In particular, BigInt has no unsigned right shift operator.
		return staticEvalResult{}, true
	}
	return staticEvalResult{value: value, ok: value.BitLen() <= maxStaticBigIntBits}, true
}

func compareStaticBigInt(value *big.Int, other any, equality bool) (order int, ordered, known bool) {
	if integer, ok := other.(*big.Int); ok {
		return value.Cmp(integer), true, true
	}
	if equality && staticValueNullish(other) {
		return 0, false, true
	}
	if staticValueIsAggregate(other) {
		text, ok := staticValueToString(other)
		if !ok {
			return 0, false, false
		}
		other = text
	}
	if text, ok := staticValueAsString(other); ok {
		integer, valid := ecmascript.StringToBigInt(text)
		if !valid {
			return 0, false, true
		}
		return value.Cmp(integer), true, true
	}
	number, ok := staticValueToNumber(other)
	if !ok || math.IsNaN(number) {
		return 0, false, ok
	}
	if math.IsInf(number, 1) {
		return -1, true, true
	}
	if math.IsInf(number, -1) {
		return 1, true, true
	}
	// Compare exactly; converting the BigInt to float64 loses bits above 2^53.
	return new(big.Rat).SetInt(value).Cmp(new(big.Rat).SetFloat64(number)), true, true
}
