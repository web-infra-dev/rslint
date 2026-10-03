package ecmascript

import (
	"math"
	"testing"
)

func TestMathBuiltinNumberEdges(t *testing.T) {
	negativeZero := math.Copysign(0, -1)
	for _, test := range []struct {
		method    string
		arguments []float64
		want      float64
	}{
		{"abs", []float64{negativeZero}, 0},
		{"ceil", []float64{-0.25}, negativeZero},
		{"floor", []float64{negativeZero}, negativeZero},
		{"round", []float64{-0.5}, negativeZero},
		{"round", []float64{-1.5}, -1},
		{"round", []float64{math.Nextafter(0.5, 0)}, 0},
		{"sign", []float64{negativeZero}, negativeZero},
		{"trunc", []float64{-0.25}, negativeZero},
		{"max", nil, math.Inf(-1)},
		{"max", []float64{negativeZero, 0}, 0},
		{"max", []float64{math.Inf(1), math.NaN()}, math.NaN()},
		{"min", nil, math.Inf(1)},
		{"min", []float64{negativeZero, 0}, negativeZero},
		{"min", []float64{math.Inf(-1), math.NaN()}, math.NaN()},
		{"hypot", nil, 0},
		{"hypot", []float64{negativeZero, negativeZero}, 0},
		{"hypot", []float64{math.NaN(), math.Inf(-1)}, math.Inf(1)},
		{"hypot", []float64{math.Inf(1), math.NaN()}, math.Inf(1)},
		{"hypot", []float64{math.MaxFloat64, math.MaxFloat64, math.NaN()}, math.NaN()},
		{"hypot", []float64{math.MaxFloat64, math.MaxFloat64, math.NaN(), math.Inf(1)}, math.Inf(1)},
		{"hypot", []float64{1e308, 1e308}, math.Sqrt2 * 1e308},
		{"hypot", []float64{3e-300, 4e-300}, 5e-300},
		{"clz32", []float64{math.NaN()}, 32},
		{"clz32", []float64{-1}, 0},
		{"imul", []float64{4294967295, 5}, -5},
		{"imul", []float64{2, math.NaN()}, 0},
		{"pow", []float64{1, math.NaN()}, math.NaN()},
		{"pow", []float64{-1, math.Inf(1)}, math.NaN()},
		{"pow", []float64{1, math.Inf(-1)}, math.NaN()},
		{"pow", []float64{math.NaN(), 0}, 1},
		{"pow", []float64{negativeZero, -3}, math.Inf(-1)},
		{"fround", []float64{negativeZero}, negativeZero},
		{"fround", []float64{1 + math.Ldexp(1, -24)}, 1},
		{"f16round", []float64{negativeZero}, negativeZero},
		{"f16round", []float64{1.00048828125}, 1},
		{"f16round", []float64{math.Nextafter(1.00048828125, 2)}, 1.0009765625},
		{"f16round", []float64{math.Ldexp(1, -25)}, 0},
		{"f16round", []float64{-math.Ldexp(1, -25)}, negativeZero},
		{"f16round", []float64{3 * math.Ldexp(1, -25)}, math.Ldexp(1, -23)},
		{"f16round", []float64{math.Ldexp(1, -14)}, math.Ldexp(1, -14)},
		{"f16round", []float64{65504}, 65504},
		{"f16round", []float64{65519}, 65504},
		{"f16round", []float64{65520}, math.Inf(1)},
		{"f16round", []float64{math.Inf(-1)}, math.Inf(-1)},
		{"f16round", []float64{math.NaN()}, math.NaN()},
	} {
		t.Run(test.method+"/"+NumberToString(test.want), func(t *testing.T) {
			builtin, ok := LookupMathBuiltin(test.method)
			if !ok {
				t.Fatal("missing Math builtin")
			}
			got := builtin.Evaluate(test.arguments)
			if math.IsNaN(test.want) {
				if !math.IsNaN(got) {
					t.Fatalf("got %v, want NaN", got)
				}
			} else if test.method == "hypot" && !math.IsInf(test.want, 0) && test.want != 0 {
				if math.Abs(got/test.want-1) > 1e-15 {
					t.Fatalf("got %v, want approximately %v", got, test.want)
				}
			} else if got != test.want || got == 0 && math.Signbit(got) != math.Signbit(test.want) {
				t.Fatalf("got %v, want %v", got, test.want)
			}
		})
	}
	if _, known := LookupMathBuiltin("random"); known {
		t.Fatal("Math.random must remain unknown")
	}
}
