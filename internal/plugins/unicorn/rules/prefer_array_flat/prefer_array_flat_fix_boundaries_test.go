package prefer_array_flat_test

import "testing"

// TestPreferArrayFlatFixBoundaries covers comment ownership, expression kinds
// that do or do not need member-object parentheses, and ASI-sensitive
// unbraced statement bodies.
func TestPreferArrayFlatFixBoundaries(t *testing.T) {
	var suite upstreamSuite

	suite.addNoFix(
		`array.flatMap(value /* keep */ => value)`,
		`array.flatMap(value /* keep */ => value)`,
		`Array#flatMap()`,
		nil,
	)
	suite.addFixed(
		`foo./* keep */bar.flatMap(value => value)`,
		`foo./* keep */bar.flatMap(value => value)`,
		`foo./* keep */bar.flat()`,
		`Array#flatMap()`,
		nil,
	)

	expressionCases := []struct {
		code        string
		replacement string
	}{
		{code: `_.flatten({value: 1})`, replacement: `({value: 1}).flat()`},
		{code: `_.flatten(class {})`, replacement: `(class {}).flat()`},
		{code: `_.flatten(value => value)`, replacement: `(value => value).flat()`},
		{code: `_.flatten(flag ? left : right)`, replacement: `(flag ? left : right).flat()`},
		{code: `_.flatten(value = input)`, replacement: `(value = input).flat()`},
		{code: `_.flatten(++value)`, replacement: `(++value).flat()`},
		{code: `_.flatten("value")`, replacement: `"value".flat()`},
		{code: `_.flatten(/value/)`, replacement: `/value/.flat()`},
		{code: `_.flatten(1n)`, replacement: `1n.flat()`},
		{code: `_.flatten(true)`, replacement: `true.flat()`},
		{code: `_.flatten(null)`, replacement: `null.flat()`},
		{code: "_.flatten(`value`)", replacement: "`value`.flat()"},
	}
	for _, testCase := range expressionCases {
		suite.addFixed(
			testCase.code,
			testCase.code,
			testCase.replacement,
			`_.flatten()`,
			nil,
		)
	}
	suite.addFixed(
		`function* values() { return _.flatten(yield input); }`,
		`_.flatten(yield input)`,
		`(yield input).flat()`,
		`_.flatten()`,
		nil,
	)

	embeddedCases := []string{
		`while (condition)
	Array.prototype.concat.call([], value)`,
		`for (;;)
	Array.prototype.concat.call([], value)`,
		`for (const item of items)
	Array.prototype.concat.call([], value)`,
		`do
	Array.prototype.concat.call([], value)
while (condition)`,
		`with (object)
	Array.prototype.concat.call([], value)`,
	}
	for _, code := range embeddedCases {
		suite.addFixed(
			code,
			`Array.prototype.concat.call([], value)`,
			`[value].flat()`,
			`Array.prototype.concat()`,
			nil,
		)
	}
	suite.addFixed(
		`before()
Array.prototype.concat.call([], value)`,
		`Array.prototype.concat.call([], value)`,
		`;[value].flat()`,
		`Array.prototype.concat()`,
		nil,
	)
	suite.addFixed(
		`value = {}
Array.prototype.concat.call([], item)`,
		`Array.prototype.concat.call([], item)`,
		`;[item].flat()`,
		`Array.prototype.concat()`,
		nil,
	)

	suite.run(t)
}
