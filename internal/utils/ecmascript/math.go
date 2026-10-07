package ecmascript

import (
	"math"
	"math/bits"
)

// MathBuiltin describes a deterministic Math method after argument coercion.
// Arity is -1 for methods consuming every argument. Evaluate receives exactly
// Arity numbers for fixed-arity methods, with missing arguments converted to NaN.
type MathBuiltin struct {
	Arity    int
	Evaluate func([]float64) float64
}

// LookupMathBuiltin excludes Math.random. Transcendental results use Go's math
// implementation, within ECMAScript's implementation-approximated contract.
// https://tc39.es/ecma262/multipage/numbers-and-dates.html#sec-math-object
func LookupMathBuiltin(name string) (MathBuiltin, bool) {
	builtin, ok := mathBuiltins[name]
	return builtin, ok
}

var mathBuiltins = map[string]MathBuiltin{
	"abs":      unaryMathBuiltin(math.Abs),
	"acos":     unaryMathBuiltin(math.Acos),
	"acosh":    unaryMathBuiltin(math.Acosh),
	"asin":     unaryMathBuiltin(math.Asin),
	"asinh":    unaryMathBuiltin(math.Asinh),
	"atan":     unaryMathBuiltin(math.Atan),
	"atanh":    unaryMathBuiltin(math.Atanh),
	"atan2":    binaryMathBuiltin(math.Atan2),
	"cbrt":     unaryMathBuiltin(math.Cbrt),
	"ceil":     unaryMathBuiltin(math.Ceil),
	"clz32":    unaryMathBuiltin(func(value float64) float64 { return float64(bits.LeadingZeros32(NumberToUint32(value))) }),
	"cos":      unaryMathBuiltin(math.Cos),
	"cosh":     unaryMathBuiltin(math.Cosh),
	"exp":      unaryMathBuiltin(math.Exp),
	"expm1":    unaryMathBuiltin(math.Expm1),
	"f16round": unaryMathBuiltin(mathF16Round),
	"floor":    unaryMathBuiltin(math.Floor),
	"fround":   unaryMathBuiltin(func(value float64) float64 { return float64(float32(value)) }),
	"hypot":    {Arity: -1, Evaluate: mathHypot},
	"imul": binaryMathBuiltin(func(left, right float64) float64 {
		return float64(int32(NumberToUint32(left) * NumberToUint32(right)))
	}),
	"log":   unaryMathBuiltin(math.Log),
	"log10": unaryMathBuiltin(math.Log10),
	"log1p": unaryMathBuiltin(math.Log1p),
	"log2":  unaryMathBuiltin(math.Log2),
	"max":   {Arity: -1, Evaluate: mathMax},
	"min":   {Arity: -1, Evaluate: mathMin},
	"pow":   binaryMathBuiltin(mathPow),
	"round": unaryMathBuiltin(mathRound),
	"sign":  unaryMathBuiltin(mathSign),
	"sin":   unaryMathBuiltin(math.Sin),
	"sinh":  unaryMathBuiltin(math.Sinh),
	"sqrt":  unaryMathBuiltin(math.Sqrt),
	"tan":   unaryMathBuiltin(math.Tan),
	"tanh":  unaryMathBuiltin(math.Tanh),
	"trunc": unaryMathBuiltin(math.Trunc),
}

func unaryMathBuiltin(evaluate func(float64) float64) MathBuiltin {
	return MathBuiltin{Arity: 1, Evaluate: func(arguments []float64) float64 { return evaluate(arguments[0]) }}
}

func binaryMathBuiltin(evaluate func(float64, float64) float64) MathBuiltin {
	return MathBuiltin{Arity: 2, Evaluate: func(arguments []float64) float64 { return evaluate(arguments[0], arguments[1]) }}
}

func mathMax(arguments []float64) float64 {
	value := math.Inf(-1)
	for _, argument := range arguments {
		if math.IsNaN(argument) {
			return math.NaN()
		}
		value = math.Max(value, argument)
	}
	return value
}

func mathMin(arguments []float64) float64 {
	value := math.Inf(1)
	for _, argument := range arguments {
		if math.IsNaN(argument) {
			return math.NaN()
		}
		value = math.Min(value, argument)
	}
	return value
}

func mathHypot(arguments []float64) float64 {
	for _, argument := range arguments {
		if math.IsInf(argument, 0) {
			return math.Inf(1)
		}
	}
	value := 0.0
	for _, argument := range arguments {
		if math.IsNaN(argument) {
			return math.NaN()
		}
		// Hypot scales to avoid intermediate overflow/underflow.
		value = math.Hypot(value, argument)
	}
	return value
}

func mathPow(base, exponent float64) float64 {
	if math.IsNaN(exponent) || math.IsInf(exponent, 0) && math.Abs(base) == 1 {
		return math.NaN()
	}
	return math.Pow(base, exponent)
}

func mathRound(value float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) || value == math.Trunc(value) {
		return value
	}
	if value < 0 && value >= -0.5 {
		return math.Copysign(0, -1)
	}
	lower := math.Floor(value)
	if value-lower < 0.5 {
		return lower
	}
	return lower + 1
}

func mathSign(value float64) float64 {
	if math.IsNaN(value) || value == 0 {
		return value
	}
	return math.Copysign(1, value)
}

func mathF16Round(value float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) || value == 0 {
		return value
	}
	magnitude := math.Abs(value)
	if magnitude >= 65520 {
		return math.Copysign(math.Inf(1), value)
	}
	// Binary16 has 11 significant bits and a smallest subnormal of 2^-24.
	// Power-of-two scaling is exact, so round directly from binary64 without
	// the double rounding introduced by an intermediate float32 conversion.
	_, exponent := math.Frexp(magnitude)
	spacingExponent := max(exponent-11, -24)
	rounded := math.Ldexp(math.RoundToEven(math.Ldexp(magnitude, -spacingExponent)), spacingExponent)
	return math.Copysign(rounded, value)
}
