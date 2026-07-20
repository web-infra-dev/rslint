package prefer_array_flat_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestPreferArrayFlatExtras locks in branches and edge shapes that the
// upstream test suite does not exercise. Each case carries an inline comment
// pointing at the specific branch, Dimension 4 row, or upstream issue it
// covers, so future refactors cannot silently regress it without breaking a
// named lock-in. The 1:1 upstream migration lives in the
// prefer_array_flat_upstream_*_test.go files.
func TestPreferArrayFlatExtras(t *testing.T) {
	var suite upstreamSuite

	// ---- Dimension 4: single- and multi-level parenthesized receivers/body ----
	suite.addFixed(
		`((array)).flatMap(x => (x))`,
		`((array)).flatMap(x => (x))`,
		`((array)).flat()`,
		`Array#flatMap()`,
		nil,
	)
	suite.addFixed(
		`((array)).reduce((a, b) => ((a.concat(b))), (([])))`,
		`((array)).reduce((a, b) => ((a.concat(b))), (([])))`,
		`((array)).flat()`,
		`Array#reduce()`,
		nil,
	)

	// ---- Dimension 4: TS receiver/expression wrappers ----
	suite.addFixed(
		`array!.flatMap(x => x)`,
		`array!.flatMap(x => x)`,
		`(array!).flat()`,
		`Array#flatMap()`,
		nil,
	)
	suite.addFixed(
		`(array as unknown[]).flatMap(x => x)`,
		`(array as unknown[]).flatMap(x => x)`,
		`(array as unknown[]).flat()`,
		`Array#flatMap()`,
		nil,
	)
	suite.addFixed(
		`(array satisfies unknown[]).flatMap(x => x)`,
		`(array satisfies unknown[]).flatMap(x => x)`,
		`(array satisfies unknown[]).flat()`,
		`Array#flatMap()`,
		nil,
	)
	suite.addFixed(
		`[].concat((maybeArray as unknown))`,
		`[].concat((maybeArray as unknown))`,
		`[(maybeArray as unknown)].flat()`,
		`[].concat()`,
		nil,
	)
	suite.addFixed(
		`_.flatten(array!)`,
		`_.flatten(array!)`,
		`(array!).flat()`,
		`_.flatten()`,
		nil,
	)
	suite.addValid(nil,
		`array.flatMap(x => x!)`,
		`array.reduce((a, b) => (a.concat(b) as unknown[]), [])`,
	)

	// ---- Dimension 4: type annotations are transparent on plain parameters ----
	suite.addFixed(
		`array.reduce((a: unknown[], b: unknown[]) => (a.concat(b)), [])`,
		`array.reduce((a: unknown[], b: unknown[]) => (a.concat(b)), [])`,
		`array.flat()`,
		`Array#reduce()`,
		nil,
	)

	// ---- Dimension 4: optional call/member boundaries ----
	suite.addValid(nil,
		`array.flatMap?.(x => x)`,
		`(array?.flatMap)(x => x)`,
		`array.reduce((a, b) => a?.concat(b), [])`,
		`array.reduce((a, b) => a.concat?.(b), [])`,
		`_.flatten?.(array)`,
		`_?.flatten(array)`,
	)
	suite.addFixed(
		`array?.items.flatMap(x => x)`,
		`array?.items.flatMap(x => x)`,
		`array?.items.flat()`,
		`Array#flatMap()`,
		nil,
	)

	// ---- Dimension 4: computed/static/numeric/Symbol access stays unmatched ----
	suite.addValid(nil,
		`array["flatMap"](x => x)`,
		"array[`flatMap`](x => x)",
		`array[0](x => x)`,
		`array[Symbol.iterator](x => x)`,
		`array["reduce"]((a, b) => a.concat(b), [])`,
		`[]["concat"](array)`,
		`_["flatten"](array)`,
	)

	// ---- Dimension 4: TypeScript type arguments are removed with the legacy call ----
	suite.addFixed(
		`array.flatMap<unknown>(x => x)`,
		`array.flatMap<unknown>(x => x)`,
		`array.flat()`,
		`Array#flatMap()`,
		nil,
	)
	suite.addFixed(
		`_.flatten<Array<unknown>>(array)`,
		`_.flatten<Array<unknown>>(array)`,
		`array.flat()`,
		`_.flatten()`,
		nil,
	)

	// ---- Dimension 4: comments around parens suppress fixes like ESTree ----
	suite.addNoFix(
		`_.flatten((foo /* keep */))`,
		`_.flatten((foo /* keep */))`,
		`_.flatten()`,
		nil,
	)
	suite.addNoFix(
		`[].concat(((foo /* keep */)))`,
		`[].concat(((foo /* keep */)))`,
		`[].concat()`,
		nil,
	)
	suite.addFixed(
		`_.flatten((foo./* keep */bar))`,
		`_.flatten((foo./* keep */bar))`,
		`(foo./* keep */bar).flat()`,
		`_.flatten()`,
		nil,
	)

	// ---- Dimension 4: side-effecting arguments remain single-evaluation fixes ----
	suite.addFixed(
		`_.flatten(getArray())`,
		`_.flatten(getArray())`,
		`getArray().flat()`,
		`_.flatten()`,
		nil,
	)
	suite.addFixed(
		`[].concat(getMaybeArray())`,
		`[].concat(getMaybeArray())`,
		`[getMaybeArray()].flat()`,
		`[].concat()`,
		nil,
	)

	// ---- Dimension 4: ASI handling must not detach an embedded statement ----
	suite.addFixed(
		`if (condition)
			Array.prototype.concat.call([], value)`,
		`Array.prototype.concat.call([], value)`,
		`[value].flat()`,
		`Array.prototype.concat()`,
		nil,
	)

	// ---- Dimension 4: multi-line report and fix range ----
	suite.addFixed(
		`const result = array.reduce(
			(a, b) => [...a, ...b],
			[],
		);`,
		`array.reduce(
			(a, b) => [...a, ...b],
			[],
		)`,
		`array.flat()`,
		`Array#reduce()`,
		nil,
	)

	// ---- Dimension 4: graceful degradation for empty/spread/holey shapes ----
	suite.addValid(nil,
		`array.flatMap()`,
		`array.flatMap(...callbacks)`,
		`array.reduce()`,
		`array.reduce((a, b) => [...a, ...b,,], [])`,
		`[].concat()`,
		`_.flatten()`,
		`_.flatten(...arrays)`,
	)

	// ---- Real-user: #2660 parenthesized non-array object receiver ----
	suite.addValid(nil, `
		const randomObject = {
			flatMap(function_: (value: unknown) => void) {
				function_(1);
			},
		};
		(randomObject).flatMap(value => value);
	`)

	// ---- Real-user: #1316 union input must be wrapped before `.flat()` ----
	suite.addFixed(
		`declare const subAppLoads: SubAppLoad | SubAppLoad[];
		const loads: SubAppLoad[] = [].concat(subAppLoads);`,
		`[].concat(subAppLoads)`,
		`[subAppLoads].flat()`,
		`[].concat()`,
		nil,
	)

	// ---- Real-user: #2470 v64 still reports the first concat in a chain ----
	suite.addFixed(
		`const values = [].concat([1]).concat([2]);`,
		`[].concat([1])`,
		`[[1]].flat()`,
		`[].concat()`,
		nil,
	)

	// Locks in upstream arrayFlatMap.testFunction() arm 1: PascalCase unknowns
	// are treated as obvious non-array receivers.
	suite.addValid(nil, `function consume(Items: unknown) { return Items.flatMap(x => x); }`)

	// Locks in upstream arrayFlatMap.testFunction() arm 2: a PascalCase const
	// initialized with an array remains eligible.
	suite.addFixed(
		`const Items = new Array(); Items.flatMap(item => item);`,
		`Items.flatMap(item => item)`,
		`Items.flat()`,
		`Array#flatMap()`,
		nil,
	)

	// Locks in upstream arrayFlatMap.testFunction() arm 3: a lower-case const
	// initialized with a known non-array remains ignored.
	suite.addValid(nil, `const collection = class {}; collection.flatMap(x => x);`)

	// Locks in upstream arrayReduce.testFunction() arm 1: concat reducer.
	suite.addFixed(
		`items.reduce((left, right) => left.concat(right), [])`,
		`items.reduce((left, right) => left.concat(right), [])`,
		`items.flat()`,
		`Array#reduce()`,
		nil,
	)

	// Locks in upstream arrayReduce.testFunction() arm 2: spread reducer.
	suite.addFixed(
		`items.reduce((left, right) => [...left, ...right], [])`,
		`items.reduce((left, right) => [...left, ...right], [])`,
		`items.flat()`,
		`Array#reduce()`,
		nil,
	)

	// Locks in upstream arrayReduce.testFunction() arm 3: default/rest
	// parameters are not plain identifiers.
	suite.addValid(nil,
		`items.reduce((left = [], right) => left.concat(right), [])`,
		`items.reduce((...values) => values[0].concat(values[1]), [])`,
	)

	// Locks in upstream emptyArrayConcat.getArrayNode(): non-spread arguments
	// switch to `[argument].flat()`, spread arguments do not.
	suite.addFixed(
		`[].concat(value)`,
		`[].concat(value)`,
		`[value].flat()`,
		`[].concat()`,
		nil,
	)
	suite.addFixed(
		`[].concat(...values)`,
		`[].concat(...values)`,
		`values.flat()`,
		`[].concat()`,
		nil,
	)

	// Locks in upstream arrayPrototypeConcat.testFunction() arms: apply rejects
	// a spread second argument; call accepts both spread and non-spread forms.
	suite.addValid(nil, `Array.prototype.concat.apply([], ...values)`)
	suite.addFixed(
		`Array.prototype.concat.call([], value)`,
		`Array.prototype.concat.call([], value)`,
		`[value].flat()`,
		`Array.prototype.concat()`,
		nil,
	)
	suite.addFixed(
		`Array.prototype.concat.call([], ...values)`,
		`Array.prototype.concat.call([], ...values)`,
		`values.flat()`,
		`Array.prototype.concat()`,
		nil,
	)

	// Locks in upstream isNodeMatchesNameOrPath() roots: this, super, and meta
	// properties all participate in configured dotted paths.
	pathOptions := map[string]interface{}{
		"functions": []interface{}{
			"this.flatten",
			"super.flatten",
			"import.meta.flatten",
			"new.target.flatten",
		},
	}
	suite.addFixed(
		`class A { method() { return this.flatten(values); } }`,
		`this.flatten(values)`,
		`values.flat()`,
		`this.flatten()`,
		pathOptions,
	)
	suite.addFixed(
		`class A extends B { method() { return super.flatten(values); } }`,
		`super.flatten(values)`,
		`values.flat()`,
		`super.flatten()`,
		pathOptions,
	)
	suite.addFixed(
		`const values = import.meta.flatten(input);`,
		`import.meta.flatten(input)`,
		`input.flat()`,
		`import.meta.flatten()`,
		pathOptions,
	)
	suite.addFixed(
		`function flatten() { return new.target.flatten(input); }`,
		`new.target.flatten(input)`,
		`input.flat()`,
		`new.target.flatten()`,
		pathOptions,
	)

	// Locks in upstream create() comment branch: comments outside the selected
	// array suppress the fix, but the diagnostic remains.
	suite.addNoFix(
		`[] /* receiver */.concat(value)`,
		`[] /* receiver */.concat(value)`,
		`[].concat()`,
		nil,
	)

	// Locks in upstream fix() member-object branches for `new` with and without
	// constructor parentheses.
	suite.addFixed(
		`_.flatten(new Collection)`,
		`_.flatten(new Collection)`,
		`(new Collection).flat()`,
		`_.flatten()`,
		nil,
	)
	suite.addFixed(
		`_.flatten(new Collection())`,
		`_.flatten(new Collection())`,
		`new Collection().flat()`,
		`_.flatten()`,
		nil,
	)

	// Locks in upstream create() traversal: nested matching calls each report,
	// and overlapping fixes settle over two passes.
	nestedOptions := map[string]interface{}{
		"functions": []interface{}{"flat"},
	}
	nestedCode := `flat(_.flatten(array))`
	suite.invalid = append(suite.invalid, rule_tester.InvalidTestCase{
		Code: nestedCode,
		Output: []string{
			`_.flatten(array).flat()`,
			`array.flat().flat()`,
		},
		Options: nestedOptions,
		Errors: []rule_tester.InvalidTestCaseError{
			upstreamError(nestedCode, nestedCode, `flat()`, 0),
			upstreamError(nestedCode, `_.flatten(array)`, `_.flatten()`, 0),
		},
	})

	// Locks in option defaults: no options and an explicit empty object produce
	// the same default Lodash diagnostic and fix.
	suite.addFixed(
		`_.flatten(values)`,
		`_.flatten(values)`,
		`values.flat()`,
		`_.flatten()`,
		map[string]interface{}{},
	)

	// ---- Dimension 3: comment ownership and fix suppression ----
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

	// ---- Dimension 3: member-object parentheses ----
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

	// ---- Dimension 3: ASI-sensitive embedded statement bodies ----
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

	// N/A: string/numeric/private object or class property keys are not rule
	// inputs; only call callee member accesses are inspected.
	// N/A: class declaration/expression and function declaration/expression
	// variants are containers only; the rule independently visits CallExpression.
	// N/A: async/generator declaration variants do not change call matching.
	// N/A: ancestor scope walks, this-binding boundaries, and static blocks do
	// not apply; only const-variable receiver classification consults a symbol.
	// N/A: object SpreadAssignment, binding RestElement, empty destructuring,
	// overload signatures, abstract members, and declare members contain no
	// candidate call by themselves and are ignored by the CallExpression listener.

	suite.run(t)
}
