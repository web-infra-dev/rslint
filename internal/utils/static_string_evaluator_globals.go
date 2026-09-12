package utils

import (
	"math"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// Built-in constants use JavaScript Number values, not host integer types.
// As with other static built-in evaluation, globals must not be shadowed.
var staticGlobalNumbers = map[string]map[string]float64{
	"Math": {
		"E": math.E, "LN2": math.Ln2, "LN10": math.Ln10,
		"LOG2E": math.Log2E, "LOG10E": math.Log10E, "PI": math.Pi,
		"SQRT1_2": math.Sqrt2 / 2, "SQRT2": math.Sqrt2,
	},
	"Number": {
		"EPSILON": 0x1p-52, "MAX_SAFE_INTEGER": 1<<53 - 1,
		"MIN_SAFE_INTEGER": -(1<<53 - 1), "MAX_VALUE": math.MaxFloat64,
		"MIN_VALUE": math.SmallestNonzeroFloat64, "NaN": math.NaN(),
		"NEGATIVE_INFINITY": math.Inf(-1), "POSITIVE_INFINITY": math.Inf(1),
	},
}

func staticGlobalNumber(object *ast.Node, key string) (float64, bool) {
	object = SkipAssertionsAndParens(object)
	if object == nil || !ast.IsIdentifier(object) || IsShadowed(object, object.Text()) {
		return 0, false
	}
	number, ok := staticGlobalNumbers[object.Text()][key]
	return number, ok
}
