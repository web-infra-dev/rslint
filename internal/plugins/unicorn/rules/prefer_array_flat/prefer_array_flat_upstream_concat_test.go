package prefer_array_flat_test

import "testing"

// TestPreferArrayFlatUpstreamConcat migrates every direct-concat and
// Array.prototype.concat case from upstream test/prefer-array-flat.js 1:1.
// Position assertions cover line and column for every invalid case. The
// flatMap/reduce groups live in prefer_array_flat_upstream_test.go and
// rslint-specific cases live in prefer_array_flat_extras_test.go.
func TestPreferArrayFlatUpstreamConcat(t *testing.T) {
	var suite upstreamSuite

	// ---- `[].concat(array)` ----
	suite.addValid(nil,
		`[].concat`,
		`new [].concat(array)`,
		`[][concat](array)`,
		`[].notConcat(array)`,
		`[,].concat(array)`,
		`({}).concat(array)`,
		`[].concat()`,
		`[].concat(array, EXTRA_ARGUMENT)`,
		`[]?.concat(array)`,
		`[].concat?.(array)`,
	)
	for _, testCase := range []struct {
		code        string
		target      string
		replacement string
		output      string
	}{
		{
			code:        `[].concat(maybeArray)`,
			target:      `[].concat(maybeArray)`,
			replacement: `[maybeArray].flat()`,
		},
		{
			code:        `[].concat( ((0, maybeArray)) )`,
			target:      `[].concat( ((0, maybeArray)) )`,
			replacement: `[((0, maybeArray))].flat()`,
		},
		{
			code:        `[].concat( ((maybeArray)) )`,
			target:      `[].concat( ((maybeArray)) )`,
			replacement: `[((maybeArray))].flat()`,
		},
		{
			code:        `[].concat( [foo] )`,
			target:      `[].concat( [foo] )`,
			replacement: `[[foo]].flat()`,
		},
		{
			code:        `[].concat( [[foo]] )`,
			target:      `[].concat( [[foo]] )`,
			replacement: `[[[foo]]].flat()`,
		},
		{
			code:        `function foo(){return[].concat(maybeArray)}`,
			target:      `[].concat(maybeArray)`,
			replacement: `[maybeArray].flat()`,
			output:      `function foo(){return [maybeArray].flat()}`,
		},
	} {
		if testCase.output == "" {
			suite.addFixed(
				testCase.code,
				testCase.target,
				testCase.replacement,
				`[].concat()`,
				nil,
			)
			continue
		}
		suite.addFixedOutput(
			testCase.code,
			testCase.output,
			nil,
			expectedDiagnostic{target: testCase.target, description: `[].concat()`},
		)
	}

	// ---- `[].concat(...array)` ----
	suite.addValid(nil,
		`new [].concat(...array)`,
		`[][concat](...array)`,
		`[].notConcat(...array)`,
		`[,].concat(...array)`,
		`({}).concat(...array)`,
		`[].concat()`,
		`[].concat(...array, EXTRA_ARGUMENT)`,
		`[]?.concat(...array)`,
		`[].concat?.(...array)`,
	)
	for _, testCase := range []struct {
		code        string
		target      string
		replacement string
		output      string
	}{
		{
			code:        `[].concat(...array)`,
			target:      `[].concat(...array)`,
			replacement: `array.flat()`,
		},
		{
			code:        `[].concat(...(( array )))`,
			target:      `[].concat(...(( array )))`,
			replacement: `(( array )).flat()`,
		},
		{
			code:        `[].concat(...(( [foo] )))`,
			target:      `[].concat(...(( [foo] )))`,
			replacement: `(( [foo] )).flat()`,
		},
		{
			code:        `[].concat(...(( [[foo]] )))`,
			target:      `[].concat(...(( [[foo]] )))`,
			replacement: `(( [[foo]] )).flat()`,
		},
		{
			code:        `function foo(){return[].concat(...array)}`,
			target:      `[].concat(...array)`,
			replacement: `array.flat()`,
			output:      `function foo(){return array.flat()}`,
		},
		{
			code:        `class A extends[].concat(...array){}`,
			target:      `[].concat(...array)`,
			replacement: `array.flat()`,
			output:      `class A extends array.flat(){}`,
		},
		{
			code:        `const A = class extends[].concat(...array){}`,
			target:      `[].concat(...array)`,
			replacement: `array.flat()`,
			output:      `const A = class extends array.flat(){}`,
		},
	} {
		if testCase.output == "" {
			suite.addFixed(
				testCase.code,
				testCase.target,
				testCase.replacement,
				`[].concat()`,
				nil,
			)
			continue
		}
		suite.addFixedOutput(
			testCase.code,
			testCase.output,
			nil,
			expectedDiagnostic{target: testCase.target, description: `[].concat()`},
		)
	}

	// ---- `[].concat.{apply,call}([], array)` ----
	suite.addValid(nil,
		`new [].concat.apply([], array)`,
		`[].concat.apply`,
		`[].concat.apply([], ...array)`,
		`[].concat.apply([], array, EXTRA_ARGUMENT)`,
		`[].concat.apply([])`,
		`[].concat.apply(NOT_EMPTY_ARRAY, array)`,
		`[].concat.apply([,], array)`,
		`[,].concat.apply([], array)`,
		`[].concat[apply]([], array)`,
		`[][concat].apply([], array)`,
		`[].concat.notApply([], array)`,
		`[].notConcat.apply([], array)`,
		`[].concat.apply?.([], array)`,
		`[].concat?.apply([], array)`,
		`[]?.concat.apply([], array)`,
	)
	for _, testCase := range []struct {
		code        string
		replacement string
	}{
		{`[].concat.apply([], array)`, `array.flat()`},
		{`[].concat.apply([], ((0, array)))`, `((0, array)).flat()`},
		{`[].concat.apply([], ((array)))`, `((array)).flat()`},
		{`[].concat.apply([], [foo])`, `[foo].flat()`},
		{`[].concat.apply([], [[foo]])`, `[[foo]].flat()`},
		{`[].concat.call([], maybeArray)`, `[maybeArray].flat()`},
		{`[].concat.call([], ((0, maybeArray)))`, `[((0, maybeArray))].flat()`},
		{`[].concat.call([], ((maybeArray)))`, `[((maybeArray))].flat()`},
		{`[].concat.call([], [foo])`, `[[foo]].flat()`},
		{`[].concat.call([], [[foo]])`, `[[[foo]]].flat()`},
		{`[].concat.call([], ...array)`, `array.flat()`},
		{`[].concat.call([], ...((0, array)))`, `((0, array)).flat()`},
		{`[].concat.call([], ...((array)))`, `((array)).flat()`},
		{`[].concat.call([], ...[foo])`, `[foo].flat()`},
		{`[].concat.call([], ...[[foo]])`, `[[foo]].flat()`},
	} {
		suite.addFixed(
			testCase.code,
			testCase.code,
			testCase.replacement,
			`Array.prototype.concat()`,
			nil,
		)
	}
	suite.addFixedOutput(
		`function foo(){return[].concat.call([], ...array)}`,
		`function foo(){return array.flat()}`,
		nil,
		expectedDiagnostic{
			target:      `[].concat.call([], ...array)`,
			description: `Array.prototype.concat()`,
		},
	)

	// ---- `Array.prototype.concat.{apply,call}([], array)` ----
	suite.addValid(nil,
		`new Array.prototype.concat.apply([], array)`,
		`Array.prototype.concat.apply`,
		`Array.prototype.concat.apply([], ...array)`,
		`Array.prototype.concat.apply([], array, EXTRA_ARGUMENT)`,
		`Array.prototype.concat.apply([])`,
		`Array.prototype.concat.apply(NOT_EMPTY_ARRAY, array)`,
		`Array.prototype.concat.apply([,], array)`,
		`Array.prototype.concat[apply]([], array)`,
		`Array.prototype[concat].apply([], array)`,
		`Array[prototype].concat.apply([], array)`,
		`Array.prototype.concat.notApply([], array)`,
		`Array.prototype.notConcat.apply([], array)`,
		`Array.notPrototype.concat.apply([], array)`,
		`NotArray.prototype.concat.apply([], array)`,
		`Array.prototype.concat.apply?.([], array)`,
		`Array.prototype.concat?.apply([], array)`,
		`Array.prototype?.concat.apply([], array)`,
		`Array?.prototype.concat.apply([], array)`,
		`object.Array.prototype.concat.apply([], array)`,
	)
	for _, testCase := range []struct {
		code        string
		replacement string
	}{
		{`Array.prototype.concat.apply([], array)`, `array.flat()`},
		{`Array.prototype.concat.apply([], ((0, array)))`, `((0, array)).flat()`},
		{`Array.prototype.concat.apply([], ((array)))`, `((array)).flat()`},
		{`Array.prototype.concat.apply([], [foo])`, `[foo].flat()`},
		{`Array.prototype.concat.apply([], [[foo]])`, `[[foo]].flat()`},
		{`Array.prototype.concat.call([], maybeArray)`, `[maybeArray].flat()`},
		{`Array.prototype.concat.call([], ((0, maybeArray)))`, `[((0, maybeArray))].flat()`},
		{`Array.prototype.concat.call([], ((maybeArray)))`, `[((maybeArray))].flat()`},
		{`Array.prototype.concat.call([], [foo])`, `[[foo]].flat()`},
		{`Array.prototype.concat.call([], [[foo]])`, `[[[foo]]].flat()`},
		{`Array.prototype.concat.call([], ...array)`, `array.flat()`},
		{`Array.prototype.concat.call([], ...((0, array)))`, `((0, array)).flat()`},
		{`Array.prototype.concat.call([], ...((array)))`, `((array)).flat()`},
		{`Array.prototype.concat.call([], ...[foo])`, `[foo].flat()`},
		{`Array.prototype.concat.call([], ...[[foo]])`, `[[foo]].flat()`},
	} {
		suite.addFixed(
			testCase.code,
			testCase.code,
			testCase.replacement,
			`Array.prototype.concat()`,
			nil,
		)
	}

	// ---- #1146: comments before the call do not block the fix ----
	suite.addFixed(
		`/**/[].concat.apply([], array)`,
		`[].concat.apply([], array)`,
		`array.flat()`,
		`Array.prototype.concat()`,
		nil,
	)
	suite.addFixed(
		`Array.prototype.concat.apply([], array)`,
		`Array.prototype.concat.apply([], array)`,
		`array.flat()`,
		`Array.prototype.concat()`,
		nil,
	)

	suite.run(t)
}
