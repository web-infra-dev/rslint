// cspell:ignore tems
package prefer_array_flat_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_array_flat"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestPreferArrayFlatExtrasRegressions locks in expressions, token boundaries,
// nested calls, and scope behavior found by comparing the Go implementation
// with unicorn v64. The complete upstream suite lives in the
// prefer_array_flat_upstream_*_test.go files; the general Dimension 4,
// real-user, branch, and fix-boundary cases live in the sibling extras file.
func TestPreferArrayFlatExtrasRegressions(t *testing.T) {
	var suite upstreamSuite

	// ---- Differential: ESTree ImportExpression vs tsgo CallExpression ----
	// Unicorn conservatively parenthesizes dynamic import before adding
	// `.flat()`, even though tsgo represents it as a call.
	suite.addFixed(
		`_.flatten(import("pkg"))`,
		`_.flatten(import("pkg"))`,
		`(import("pkg")).flat()`,
		`_.flatten()`,
		nil,
	)

	// ---- Differential: a standalone scanner must recover the parser's regex token ----
	suite.addFixedOutput(
		`/before/giu
Array.prototype.concat.call([], value)`,
		`/before/giu
;[value].flat()`,
		nil,
		expectedDiagnostic{
			target:      `Array.prototype.concat.call([], value)`,
			description: `Array.prototype.concat()`,
		},
	)

	// ---- Differential: TypeScript-only contextual keywords are identifier tokens ----
	// The closing `)` already separates these tokens, so upstream does not add
	// whitespace around `as` or `satisfies`.
	suite.addFixed(
		`const result = _.flatten(value)as unknown[];`,
		`_.flatten(value)`,
		`value.flat()`,
		`_.flatten()`,
		nil,
	)
	suite.addFixed(
		`const result = _.flatten(value)satisfies unknown[];`,
		`_.flatten(value)`,
		`value.flat()`,
		`_.flatten()`,
		nil,
	)

	// ---- Differential: every untested member-object expression class ----
	for _, testCase := range []struct {
		code        string
		replacement string
	}{
		{
			`const result = _.flatten(function () {});`,
			`function () {}.flat()`,
		},
		{
			`const result = _.flatten(async function () {});`,
			`async function () {}.flat()`,
		},
		{
			`const result = _.flatten(function * () {});`,
			`function * () {}.flat()`,
		},
		{
			`const result = _.flatten(async function * () {});`,
			`async function * () {}.flat()`,
		},
		{
			`const result = _.flatten(import.meta);`,
			`(import.meta).flat()`,
		},
		{
			"const result = _.flatten(tag`value`);",
			"(tag`value`).flat()",
		},
		{
			"const result = _.flatten(`value${part}`);",
			"`value${part}`.flat()",
		},
		{
			`const result = _.flatten(typeof value);`,
			`(typeof value).flat()`,
		},
		{
			`const result = _.flatten(void value);`,
			`(void value).flat()`,
		},
		{
			`const result = _.flatten(delete object.value);`,
			`(delete object.value).flat()`,
		},
		{
			`const result = _.flatten(value++);`,
			`(value++).flat()`,
		},
		{
			`const result = _.flatten(value && other);`,
			`(value && other).flat()`,
		},
		{
			`const result = _.flatten(value ?? other);`,
			`(value ?? other).flat()`,
		},
		{
			`const result = _.flatten(object?.value);`,
			`object?.value.flat()`,
		},
		{
			`const result = _.flatten(getValue?.());`,
			`getValue?.().flat()`,
		},
	} {
		suite.addFixed(
			testCase.code,
			testCase.code[len("const result = "):len(testCase.code)-1],
			testCase.replacement,
			`_.flatten()`,
			nil,
		)
	}

	// ---- Differential: raw numeric spelling controls lexical parentheses ----
	for _, testCase := range []struct {
		code        string
		replacement string
	}{
		{`_.flatten(0)`, `(0).flat()`},
		{`_.flatten(1_000)`, `(1_000).flat()`},
		{`_.flatten(0x1)`, `0x1.flat()`},
		{`_.flatten(0b1)`, `0b1.flat()`},
		{`_.flatten(0o1)`, `0o1.flat()`},
		{`_.flatten(1e2)`, `1e2.flat()`},
	} {
		suite.addFixed(
			testCase.code,
			testCase.code,
			testCase.replacement,
			`_.flatten()`,
			nil,
		)
	}

	// ---- Differential: every definitely-non-array initializer kind ----
	suite.addValid(nil,
		`const value = 1; value.flatMap(item => item);`,
		`const value = 1n; value.flatMap(item => item);`,
		`const value = /value/u; value.flatMap(item => item);`,
		"const value = `value${part}`; value.flatMap(item => item);",
		`const value = false; value.flatMap(item => item);`,
		`const value = null; value.flatMap(item => item);`,
		`const value = function () {}; value.flatMap(item => item);`,
		`const value = async function () {}; value.flatMap(item => item);`,
	)

	// A namespaced constructor is not one of upstream's syntactically known
	// non-array constructions, so the lower-case const remains reportable.
	suite.addFixed(
		`const value = new namespace.Collection(); value.flatMap(item => item);`,
		`value.flatMap(item => item)`,
		`value.flat()`,
		`Array#flatMap()`,
		nil,
	)

	// Parentheses are transparent at every segment of the recognized
	// Array.prototype.concat path.
	suite.addFixed(
		`((Array)).prototype.concat.apply(([]), ((value)))`,
		`((Array)).prototype.concat.apply(([]), ((value)))`,
		`((value)).flat()`,
		`Array.prototype.concat()`,
		nil,
	)
	suite.addValid(nil,
		`Array.prototype.concat.apply(...[[]], value)`,
		`Array.prototype.concat.call(...[[]], value)`,
		`array.reduce(...callbacks, [])`,
	)

	// Escaped identifier spelling is retained in the fix while symbol
	// classification uses the identifier's decoded name.
	suite.addFixed(
		`const Items = []; Item\u0073.flatMap(item => item);`,
		`Item\u0073.flatMap(item => item)`,
		`Item\u0073.flat()`,
		`Array#flatMap()`,
		nil,
	)

	// ---- Differential: all lowercase ECMAScript/contextual keyword boundaries ----
	for _, testCase := range []struct {
		code   string
		target string
		output string
	}{
		{
			code:   `function flatten() { throw[].concat(value); }`,
			target: `[].concat(value)`,
			output: `function flatten() { throw [value].flat(); }`,
		},
		{
			code:   `const result = typeof[].concat(value);`,
			target: `[].concat(value)`,
			output: `const result = typeof [value].flat();`,
		},
		{
			code:   `const result = void[].concat(value);`,
			target: `[].concat(value)`,
			output: `const result = void [value].flat();`,
		},
		{
			code:   `const result = delete[].concat(value);`,
			target: `[].concat(value)`,
			output: `const result = delete [value].flat();`,
		},
		{
			code:   `const result = [].concat(value)in object;`,
			target: `[].concat(value)`,
			output: `const result = [value].flat() in object;`,
		},
		{
			code:   `for (const item of[].concat(value)) {}`,
			target: `[].concat(value)`,
			output: `for (const item of [value].flat()) {}`,
		},
		{
			code:   `async function flatten() { await[].concat(value); }`,
			target: `[].concat(value)`,
			output: `async function flatten() { await [value].flat(); }`,
		},
		{
			code:   `function* flatten() { yield[].concat(value); }`,
			target: `[].concat(value)`,
			output: `function* flatten() { yield [value].flat(); }`,
		},
	} {
		suite.addFixedOutput(
			testCase.code,
			testCase.output,
			nil,
			expectedDiagnostic{
				target:      testCase.target,
				description: `[].concat()`,
			},
		)
	}

	// ---- Regressions: nested callees, lexical scope, and overlapping fixes ----
	suite.addFixed(
		`(array.flatMap)(x => x)`,
		`(array.flatMap)(x => x)`,
		`array.flat()`,
		`Array#flatMap()`,
		nil,
	)
	suite.addFixed(
		`((array.flatMap(x => x)))`,
		`array.flatMap(x => x)`,
		`array.flat()`,
		`Array#flatMap()`,
		nil,
	)
	suite.addFixed(
		`array.flatMap((value?: unknown) => value)`,
		`array.flatMap((value?: unknown) => value)`,
		`array.flat()`,
		`Array#flatMap()`,
		nil,
	)
	suite.addValid(nil,
		`array.flatMap((value = []) => value)`,
		`array.flatMap((...values) => values)`,
		`array.flatMap(([value]) => value)`,
	)

	// Receiver classification follows the exact symbol in the current scope
	// and only unwraps parentheses around a const initializer.
	suite.addValid(nil,
		`const collection = ({}); collection.flatMap(value => value);`,
		`const Items = [] as unknown[]; Items.flatMap(value => value);`,
	)
	suite.addFixed(
		`const collection = {} as unknown; collection.flatMap(value => value);`,
		`collection.flatMap(value => value)`,
		`collection.flat()`,
		`Array#flatMap()`,
		nil,
	)
	suite.addFixed(
		`let collection = {}; collection.flatMap(value => value);`,
		`collection.flatMap(value => value)`,
		`collection.flat()`,
		`Array#flatMap()`,
		nil,
	)
	// U+2160 is uppercase-like but belongs to Unicode category Nl, not Lu.
	// Upstream's /^\p{Lu}/u therefore does not suppress this receiver.
	suite.addFixed(
		`const Ⅰtems = getValues(); Ⅰtems.flatMap(value => value);`,
		`Ⅰtems.flatMap(value => value)`,
		`Ⅰtems.flat()`,
		`Array#flatMap()`,
		nil,
	)
	shadowedCode := `{ const Items = {}; Items.flatMap(value => value); }
const Items = [];
Items.flatMap(value => value);`
	suite.addFixedOutput(
		shadowedCode,
		`{ const Items = {}; Items.flatMap(value => value); }
const Items = [];
Items.flat();`,
		nil,
		expectedDiagnostic{
			target:      `Items.flatMap(value => value)`,
			description: `Array#flatMap()`,
			occurrence:  1,
		},
	)
	suite.addFixed(
		`Items.flatMap(value => value); const Items = [];`,
		`Items.flatMap(value => value)`,
		`Items.flat()`,
		`Array#flatMap()`,
		nil,
	)
	// ESLint's variable definition points each destructured binding at the
	// enclosing declarator, so classification uses that declarator's complete
	// initializer rather than the extracted property's runtime value.
	suite.addValid(nil,
		`const {items} = {}; items.flatMap(value => value);`,
	)
	suite.addFixed(
		`const [items] = []; items.flatMap(value => value);`,
		`items.flatMap(value => value)`,
		`items.flat()`,
		`Array#flatMap()`,
		nil,
	)
	suite.addFixed(
		`const {Items} = []; Items.flatMap(value => value);`,
		`Items.flatMap(value => value)`,
		`Items.flat()`,
		`Array#flatMap()`,
		nil,
	)

	pathOptions := map[string]interface{}{
		"functions": []interface{}{"utils.deep.flat", "utils.flat", "flat"},
	}
	suite.addFixed(
		`(utils.deep).flat(array)`,
		`(utils.deep).flat(array)`,
		`array.flat()`,
		`utils.deep.flat()`,
		pathOptions,
	)
	suite.addFixed(
		`((utils.flat))(array)`,
		`((utils.flat))(array)`,
		`array.flat()`,
		`utils.flat()`,
		pathOptions,
	)
	suite.addValid(pathOptions,
		`utils?.deep.flat(array)`,
		`utils.deep?.flat(array)`,
		`utils["deep"].flat(array)`,
	)
	suite.addFixed(
		`flat(array.flatMap(value => other))`,
		`flat(array.flatMap(value => other))`,
		`array.flatMap(value => other).flat()`,
		`flat()`,
		pathOptions,
	)

	// A configured function can intentionally overlap a built-in matcher.
	// Both diagnostics are retained, while the first case's overlapping fix
	// wins just as it does in ESLint.
	overlapCode := `array.flatMap(value => value)`
	suite.invalid = append(suite.invalid, rule_tester.InvalidTestCase{
		Code:    overlapCode,
		Output:  []string{`array.flat()`},
		Options: map[string]interface{}{"functions": []interface{}{"array.flatMap"}},
		Errors: []rule_tester.InvalidTestCaseError{
			upstreamError(overlapCode, overlapCode, `Array#flatMap()`, 0),
			upstreamError(overlapCode, overlapCode, `array.flatMap()`, 0),
		},
	})

	// All three calls report on the first pass. Their overlapping replacements
	// are then applied one enclosing call at a time.
	nestedCode := `flat(flat(_.flatten(array)))`
	suite.invalid = append(suite.invalid, rule_tester.InvalidTestCase{
		Code: nestedCode,
		Output: []string{
			`flat(_.flatten(array)).flat()`,
			`_.flatten(array).flat().flat()`,
			`array.flat().flat().flat()`,
		},
		Options: pathOptions,
		Errors: []rule_tester.InvalidTestCaseError{
			upstreamError(nestedCode, nestedCode, `flat()`, 0),
			upstreamError(nestedCode, `flat(_.flatten(array))`, `flat()`, 0),
			upstreamError(nestedCode, `_.flatten(array)`, `_.flatten()`, 0),
		},
	})

	// Nested member receivers, switch-to-array fixes, and direct member-object
	// fixes all settle one non-overlapping replacement per pass, in the same
	// range order as ESLint's SourceCodeFixer.
	nestedMemberCode := `array.flatMap(value => value).flatMap(value => value)`
	suite.invalid = append(suite.invalid, rule_tester.InvalidTestCase{
		Code: nestedMemberCode,
		Output: []string{
			`array.flat().flatMap(value => value)`,
			`array.flat().flat()`,
		},
		Errors: []rule_tester.InvalidTestCaseError{
			upstreamError(nestedMemberCode, nestedMemberCode, `Array#flatMap()`, 0),
			upstreamError(
				nestedMemberCode,
				`array.flatMap(value => value)`,
				`Array#flatMap()`,
				0,
			),
		},
	})

	nestedSwitchCode := `[].concat(_.flatten(value))`
	suite.invalid = append(suite.invalid, rule_tester.InvalidTestCase{
		Code: nestedSwitchCode,
		Output: []string{
			`[_.flatten(value)].flat()`,
			`[value.flat()].flat()`,
		},
		Errors: []rule_tester.InvalidTestCaseError{
			upstreamError(nestedSwitchCode, nestedSwitchCode, `[].concat()`, 0),
			upstreamError(nestedSwitchCode, `_.flatten(value)`, `_.flatten()`, 0),
		},
	})

	nestedDirectCode := `_.flatten([].concat(value))`
	suite.invalid = append(suite.invalid, rule_tester.InvalidTestCase{
		Code: nestedDirectCode,
		Output: []string{
			`[].concat(value).flat()`,
			`[value].flat().flat()`,
		},
		Errors: []rule_tester.InvalidTestCaseError{
			upstreamError(nestedDirectCode, nestedDirectCode, `_.flatten()`, 0),
			upstreamError(nestedDirectCode, `[].concat(value)`, `[].concat()`, 0),
		},
	})

	suite.run(t)
}

// TestPreferArrayFlatSchemaParity locks the public option schema to unicorn
// v64. In particular, upstream leaves array item types unconstrained while
// still enforcing tuple length, uniqueness, and additionalProperties.
func TestPreferArrayFlatSchemaParity(t *testing.T) {
	tests := []struct {
		name    string
		options []any
		wantErr bool
	}{
		{name: "no options"},
		{name: "empty object", options: []any{map[string]any{}}},
		{
			name: "string functions",
			options: []any{map[string]any{
				"functions": []any{"flat", "utils.flat"},
			}},
		},
		{
			name: "unconstrained item types",
			options: []any{map[string]any{
				"functions": []any{1.0, nil, true, map[string]any{}},
			}},
		},
		{
			name: "duplicate functions",
			options: []any{map[string]any{
				"functions": []any{"flat", "flat"},
			}},
			wantErr: true,
		},
		{
			name: "functions must be array",
			options: []any{map[string]any{
				"functions": "flat",
			}},
			wantErr: true,
		},
		{
			name: "unknown property",
			options: []any{map[string]any{
				"unknown": true,
			}},
			wantErr: true,
		},
		{
			name:    "only one option object",
			options: []any{map[string]any{}, map[string]any{}},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := prefer_array_flat.PreferArrayFlatRule.Schema.Validate(test.options)
			if (err != nil) != test.wantErr {
				t.Fatalf("Schema.Validate(%#v) error = %v, wantErr %v",
					test.options, err, test.wantErr)
			}
		})
	}
}
