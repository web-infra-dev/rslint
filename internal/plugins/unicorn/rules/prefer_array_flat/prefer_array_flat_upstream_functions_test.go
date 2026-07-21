package prefer_array_flat_test

import "testing"

// TestPreferArrayFlatUpstreamFunctions migrates the Lodash/custom-function,
// option, ASI, parentheses, numeric, and comment groups from upstream
// test/prefer-array-flat.js 1:1. Position assertions cover line and column for
// every invalid case. Other upstream groups live in the sibling
// prefer_array_flat_upstream_*_test.go files; rslint-specific cases live in
// prefer_array_flat_extras_test.go.
func TestPreferArrayFlatUpstreamFunctions(t *testing.T) {
	var suite upstreamSuite

	// ---- `_.flatten(array)` ----
	suite.addValid(nil,
		`new _.flatten(array)`,
		`_.flatten`,
		`_.flatten(array, EXTRA_ARGUMENT)`,
		`_.flatten(...array)`,
		`_[flatten](array)`,
		`_.notFlatten(array)`,
		`NOT_LODASH.flatten(array)`,
		`_.flatten?.(array)`,
		`_?.flatten(array)`,
		`object._.flatten(array)`,
	)
	for _, testCase := range []struct {
		code        string
		description string
	}{
		{`_.flatten(array)`, `_.flatten()`},
		{`lodash.flatten(array)`, `lodash.flatten()`},
		{`underscore.flatten(array)`, `underscore.flatten()`},
	} {
		suite.addFixed(
			testCase.code,
			testCase.code,
			`array.flat()`,
			testCase.description,
			nil,
		)
	}

	// ---- `options.functions` ----
	options := map[string]interface{}{
		"functions": []interface{}{
			"flat",
			"utils.flat",
			"globalThis.lodash.flatten",
		},
	}
	suite.addValid(options,
		`flat`,
		`new flat(array)`,
		`flat?.(array)`,
		`object.flat?.(array)`,
		`utils.flat`,
		`new utils.flat(array)`,
		`utils.flat?.(array)`,
		`utils?.flat(array)`,
		`utils.flat2(array)`,
		`utils2.flat(array)`,
		`object.utils.flat(array)`,
		`globalThis.lodash.flatten`,
		`new globalThis.lodash.flatten(array)`,
		`globalThis.lodash.flatten?.(array)`,
		`globalThis.lodash?.flatten(array)`,
		`globalThis?.lodash.flatten(array)`,
		`object.globalThis.lodash.flatten(array)`,
		// cspell:disable
		`GLOBALTHIS.lodash.flatten(array)`,
		`globalthis.lodash.flatten(array)`,
		`GLOBALTHIS.LODASH.FLATTEN(array)`,
		// cspell:enable
		`flat(array, EXTRA_ARGUMENT)`,
		`flat(...array)`,
	)
	for _, testCase := range []struct {
		code        string
		target      string
		replacement string
		description string
		output      string
	}{
		{
			code:        `flat(array)`,
			target:      `flat(array)`,
			replacement: `array.flat()`,
			description: `flat()`,
		},
		{
			code:        `flat(array,)`,
			target:      `flat(array,)`,
			replacement: `array.flat()`,
			description: `flat()`,
		},
		{
			code:        `utils.flat(array)`,
			target:      `utils.flat(array)`,
			replacement: `array.flat()`,
			description: `utils.flat()`,
		},
		{
			code:        `globalThis.lodash.flatten(array)`,
			target:      `globalThis.lodash.flatten(array)`,
			replacement: `array.flat()`,
			description: `globalThis.lodash.flatten()`,
		},
		{
			code: `import {flatten as flat} from 'lodash-es';
			const foo = flat(bar);`,
			target:      `flat(bar)`,
			replacement: `bar.flat()`,
			description: `flat()`,
		},
		{
			code:        `_.flatten(array).length`,
			target:      `_.flatten(array)`,
			replacement: `array.flat()`,
			description: `_.flatten()`,
		},
		{
			code:        `Array.prototype.concat.apply([], array)`,
			target:      `Array.prototype.concat.apply([], array)`,
			replacement: `array.flat()`,
			description: `Array.prototype.concat()`,
		},
	} {
		suite.addFixed(
			testCase.code,
			testCase.target,
			testCase.replacement,
			testCase.description,
			options,
		)
	}
	suite.addFixedOutput(
		`flat(array).map(array => utils.flat(array))`,
		`array.flat().map(array => array.flat())`,
		options,
		expectedDiagnostic{target: `flat(array)`, description: `flat()`},
		expectedDiagnostic{target: `utils.flat(array)`, description: `utils.flat()`},
	)

	// ---- Whitespace in `options.functions` ----
	spacesInFunctions := []interface{}{map[string]interface{}{
		"functions": []interface{}{
			"",
			" ",
			" flat1 ",
			"utils..flat2",
			"utils . flat3",
			"utils.fl at4",
			"utils.flat5  ",
			"  utils.flat6",
		},
	}}
	suite.addValid(spacesInFunctions,
		`utils.flat2(x)`,
		`utils.flat3(x)`,
		`utils.flat4(x)`,
	)
	for _, testCase := range []struct {
		code        string
		description string
	}{
		{`flat1(x)`, `flat1()`},
		{`utils.flat5(x)`, `utils.flat5()`},
		{`utils.flat6(x)`, `utils.flat6()`},
	} {
		suite.addFixed(
			testCase.code,
			testCase.code,
			`x.flat()`,
			testCase.description,
			spacesInFunctions,
		)
	}

	// ---- Existing `.flat()` calls ----
	suite.addValid(nil,
		`array.flat()`,
		`array.flat(1)`,
	)

	// ---- ASI ----
	suite.addFixedOutput(
		`before()
			Array.prototype.concat.apply([], [array].concat(array))`,
		`before()
			;[array].concat(array).flat()`,
		nil,
		expectedDiagnostic{
			target:      `Array.prototype.concat.apply([], [array].concat(array))`,
			description: `Array.prototype.concat()`,
		},
	)
	suite.addFixedOutput(
		`before()
			Array.prototype.concat.apply([], +1)`,
		`before()
			;(+1).flat()`,
		nil,
		expectedDiagnostic{
			target:      `Array.prototype.concat.apply([], +1)`,
			description: `Array.prototype.concat()`,
		},
	)
	suite.addFixedOutput(
		`before()
			Array.prototype.concat.call([], +1)`,
		`before()
			;[+1].flat()`,
		nil,
		expectedDiagnostic{
			target:      `Array.prototype.concat.call([], +1)`,
			description: `Array.prototype.concat()`,
		},
	)

	// ---- Parentheses and await ----
	for _, testCase := range []struct {
		code        string
		target      string
		replacement string
		description string
	}{
		{
			`Array.prototype.concat.apply([], (0, array))`,
			`Array.prototype.concat.apply([], (0, array))`,
			`(0, array).flat()`,
			`Array.prototype.concat()`,
		},
		{
			`Array.prototype.concat.call([], (0, array))`,
			`Array.prototype.concat.call([], (0, array))`,
			`[(0, array)].flat()`,
			`Array.prototype.concat()`,
		},
		{
			`async function a() { return [].concat(await getArray()); }`,
			`[].concat(await getArray())`,
			`[await getArray()].flat()`,
			`[].concat()`,
		},
		{
			`_.flatten((0, array))`,
			`_.flatten((0, array))`,
			`(0, array).flat()`,
			`_.flatten()`,
		},
		{
			`async function a() { return _.flatten(await getArray()); }`,
			`_.flatten(await getArray())`,
			`(await getArray()).flat()`,
			`_.flatten()`,
		},
		{
			`async function a() { return _.flatten((await getArray())); }`,
			`_.flatten((await getArray()))`,
			`(await getArray()).flat()`,
			`_.flatten()`,
		},
	} {
		suite.addFixed(
			testCase.code,
			testCase.target,
			testCase.replacement,
			testCase.description,
			nil,
		)
	}

	// ---- Numeric member-expression objects ----
	for _, testCase := range []struct {
		code   string
		output string
	}{
		{
			`before()
			Array.prototype.concat.apply([], 1)`,
			`before()
			;(1).flat()`,
		},
		{
			`before()
			Array.prototype.concat.call([], 1)`,
			`before()
			;[1].flat()`,
		},
		{
			`before()
			Array.prototype.concat.apply([], 1.)`,
			`before()
			1..flat()`,
		},
		{
			`before()
			Array.prototype.concat.call([], 1.)`,
			`before()
			;[1.].flat()`,
		},
		{
			`before()
			Array.prototype.concat.apply([], .1)`,
			`before()
			;.1.flat()`,
		},
		{
			`before()
			Array.prototype.concat.call([], .1)`,
			`before()
			;[.1].flat()`,
		},
		{
			`before()
			Array.prototype.concat.apply([], 1.0)`,
			`before()
			1.0.flat()`,
		},
		{
			`before()
			Array.prototype.concat.call([], 1.0)`,
			`before()
			;[1.0].flat()`,
		},
	} {
		target := testCase.code[len("before()\n\t\t\t"):]
		suite.addFixedOutput(
			testCase.code,
			testCase.output,
			nil,
			expectedDiagnostic{
				target:      target,
				description: `Array.prototype.concat()`,
			},
		)
	}

	// ---- Comments ----
	suite.addFixed(
		`[].concat(some./**/array)`,
		`[].concat(some./**/array)`,
		`[some./**/array].flat()`,
		`[].concat()`,
		nil,
	)
	suite.addNoFix(
		`[/**/].concat(some./**/array)`,
		`[/**/].concat(some./**/array)`,
		`[].concat()`,
		nil,
	)
	suite.addNoFix(
		`[/**/].concat(some.array)`,
		`[/**/].concat(some.array)`,
		`[].concat()`,
		nil,
	)

	suite.run(t)
}
