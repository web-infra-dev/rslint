// Ported from eslint-plugin-unicorn v79.0.0 tests and documentation; see LICENSE.
package no_array_sort_for_min_max

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoArraySortForMinMaxUpstream(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// Upstream snapshot cases.
		{Code: "function f(set: Set<number>) { return set.toSorted((a, b) => a - b)[0]; }", FileName: "input.ts"},
		{Code: "const minimum = Math.min(...array);"},
		{Code: "const minimum = array.reduce((minimum, value) => Math.min(minimum, value));"},
		{Code: "const minimum = array.sort()[0];"},
		{Code: "const minimum = array.toSorted()[0];"},
		{Code: "const first = array.sort((a, b) => a.localeCompare(b))[0];"},
		{Code: "const first = array.sort(compare)[0];"},
		{Code: "const first = array.sort((a, b) => a - b || a.id - b.id)[0];"},
		{Code: "const first = array.sort((a, b) => a.value - b.value)[0];"},
		{Code: "const first = array.sort((a, b) => a - b)[1];"},
		{Code: "const first = array.sort((a, b) => a - b).at(1);"},
		{Code: "const first = array.sort((a, b) => a - b).at(index);"},
		{Code: "const first = array.sort?.((a, b) => a - b)[0];"},
		{Code: "const first = array?.sort((a, b) => a - b)[0];"},
		{Code: "const first = array.sort((a, b) => a - b)?.[0];"},
		{Code: "const first = object?.array.toSorted((a, b) => a - b)[0];"},
		{Code: "const first = object?.array.toSorted((a, b) => a - b).at(0);"},
		{Code: "const first = object?.array.sort((a, b) => a - b)[0];"},
		{Code: "const first = object?.getArray().toSorted((a, b) => a - b)[0];"},
		{Code: "const first = object?.getArray().toSorted((a, b) => a - b).at(0);"},
		{Code: "const first = array[\"sort\"]((a, b) => a - b)[0];"},
		{Code: "const first = array.sort((a, b) => { return a - b; return 0; })[0];"},
		{Code: "const first = array.sort((a, b) => {})[0];"},
		{Code: "const first = array.sort(async (a, b) => a - b)[0];"},
		{Code: "const first = array.sort(function * (a, b) { return a - b; })[0];"},
		{Code: "const first = ({sort() {}}).sort((a, b) => a - b)[0];"},
		{Code: "const first = new Set().sort((a, b) => a - b)[0];"},
		{Code: "array.sort((a, b) => a - b)[0] = value;"},
		{Code: "array.sort((a, b) => a - b)[0]++;"},
		{Code: "delete array.sort((a, b) => a - b)[0];"},
		{Code: "delete array.sort((a, b) => a - b).at(0);"},
		{Code: "for (array.sort((a, b) => a - b)[0] of values) {}"},
		{Code: "for (array.sort((a, b) => a - b)[0] in object) {}"},
		{Code: "const first = array.sort(function (a, a) { return a - a; })[0];", LanguageOptions: rule.LanguageOptions{SourceType: "script"}},
		{Code: "const first = object?.array!.toSorted((a: number, b: number) => a - b)[0];", FileName: "input.ts"},
		{Code: "const first = object?.getArray()!.toSorted((a: number, b: number) => a - b)[0];", FileName: "input.ts"},
		{Code: "const first = array.sort((a: number, b: number) => a - b).at(1);", FileName: "input.ts"},
		{Code: "new BigInt64Array().sort((a, b) => a - b)[0];"},
		{Code: "new BigUint64Array().sort((a, b) => a - b).at(-1);"},
		{Code: "BigInt64Array.from([]).toSorted((a, b) => a - b)[0];"},
		{Code: "BigUint64Array.of(1n).toSorted((a, b) => a - b).at(-1);"},
		{Code: "function f(values: BigInt64Array) { return values.sort((a, b) => a - b)[0]; }", FileName: "input.ts"},
		{Code: "function f(values: BigUint64Array) { return values.toSorted((a, b) => a - b).at(-1); }", FileName: "input.ts"},
		{Code: "function f(values: BigInt64Array | Int8Array) { return values.toSorted((a, b) => a - b)[0]; }", FileName: "input.ts"},
		// Additional examples from the upstream documentation.
		// The remaining examples appear in the snapshot cases above.
		{Code: "const maximum = Math.max(...array);"},
		{Code: "const sorted = array.toSorted((a, b) => a - b);"},
		{Code: "const first = array.toSorted((a, b) => a.value - b.value)[0];"},
	}
	invalid := []minMaxCase{
		// Keyword spacing.
		{
			code:       "function f() {return(getArray)(/* keep */).toSorted((a, b) => a - b).at(-1);}",
			method:     "max",
			position:   [4]int{1, 21, 1, 76},
			suggestion: "function f() {return Math.max(...(getArray)(/* keep */));}",
		},
		{
			code:       "function f() {throw(getArray)(/* keep */).toSorted((a, b) => a - b).at(-1);}",
			method:     "max",
			position:   [4]int{1, 20, 1, 75},
			suggestion: "function f() {throw Math.max(...(getArray)(/* keep */));}",
		},
		{code: "if (flag) {}else(array.sort((a, b) => a - b))[0];", method: "min", position: [4]int{1, 17, 1, 49}, suggestion: "if (flag) {}else Math.min(...array);"},
		{code: "do(array.sort((a, b) => a - b))[0];while (flag);", method: "min", position: [4]int{1, 3, 1, 35}, suggestion: "do Math.min(...array);while (flag);"},
		// Comments in the receiver.
		{code: "[3, /* keep */ 1, 2].sort((a, b) => a - b)[0];", method: "min", position: [4]int{1, 1, 1, 46}, suggestion: "Math.min(...[3, /* keep */ 1, 2]);"},
		{code: "getArray(/* keep */).toSorted((a, b) => a - b).at(-1);", method: "max", position: [4]int{1, 1, 1, 54}, suggestion: "Math.max(...getArray(/* keep */));"},
		// Comments and TypeScript.
		{code: "[3, // keep\r\n1, 2].toSorted((a, b) => b - a)[0];", method: "max", position: [4]int{1, 1, 2, 35}, suggestion: "Math.max(...[3, // keep\r\n1, 2]);"},
		{
			code:       "(array as /* keep */ number[]).sort((a, b) => a - b).at(0);",
			fileName:   "input.ts",
			method:     "min",
			position:   [4]int{1, 1, 1, 59},
			suggestion: "Math.min(...(array as /* keep */ number[]));",
		},
		{code: "(/* discard */ array).sort((a, b) => a - b)[0];", method: "min", position: [4]int{1, 1, 1, 47}},
		{code: "array.sort((a, b) => a /* discard */ - b)[0];", method: "min", position: [4]int{1, 1, 1, 45}},
		{code: "array.sort((a, b) => a - b)[/* discard */ 0];", method: "min", position: [4]int{1, 1, 1, 45}},
		{code: "array./* discard */toSorted((a, b) => a - b).at(-1);", method: "max", position: [4]int{1, 1, 1, 52}},
		{code: "getArray(/* keep */).sort((a, b) => a - b).at(/* discard */ 0);", method: "min", position: [4]int{1, 1, 1, 63}},
		// Comment-preserving suggestions with mutation and empty arrays.
		{
			code: `let calls = 0;
const getArray = () => {calls++; return values;};
function read() {return(getArray)(/* keep */).sort((a, b) => a - b)[0];}
const result = read();
({result, calls, values});`,
			method:   "min",
			position: [4]int{3, 24, 3, 71},
			suggestion: `let calls = 0;
const getArray = () => {calls++; return values;};
function read() {return Math.min(...(getArray)(/* keep */));}
const result = read();
({result, calls, values});`,
		},
		{
			code: `let calls = 0;
const getArray = () => {calls++; return values;};
function read() {return(getArray)(/* keep */).toSorted((a, b) => a - b).at(-1);}
const result = read();
({result, calls, values});`,
			method:   "max",
			position: [4]int{3, 24, 3, 79},
			suggestion: `let calls = 0;
const getArray = () => {calls++; return values;};
function read() {return Math.max(...(getArray)(/* keep */));}
const result = read();
({result, calls, values});`,
		},
		// Upstream snapshot cases.
		{code: "const minimum = array.sort((a, b) => a - b)[0];", method: "min", position: [4]int{1, 17, 1, 47}, suggestion: "const minimum = Math.min(...array);"},
		{code: "const minimum = array.toSorted((a, b) => a - b)[0];", method: "min", position: [4]int{1, 17, 1, 51}, suggestion: "const minimum = Math.min(...array);"},
		{
			code:       "const minimum = arrays[key?.name].toSorted((a, b) => a - b)[0];",
			method:     "min",
			position:   [4]int{1, 17, 1, 63},
			suggestion: "const minimum = Math.min(...arrays[key?.name]);",
		},
		{code: "const minimum = array.sort((a, b) => a - b).at(0);", method: "min", position: [4]int{1, 17, 1, 50}, suggestion: "const minimum = Math.min(...array);"},
		{code: "const minimum = array.toSorted((a, b) => a - b).at(0);", method: "min", position: [4]int{1, 17, 1, 54}, suggestion: "const minimum = Math.min(...array);"},
		{code: "const maximum = array.sort((a, b) => a - b).at(-1);", method: "max", position: [4]int{1, 17, 1, 51}, suggestion: "const maximum = Math.max(...array);"},
		{code: "const maximum = array.toSorted((a, b) => a - b).at(-1);", method: "max", position: [4]int{1, 17, 1, 55}, suggestion: "const maximum = Math.max(...array);"},
		{code: "const maximum = array.sort((a, b) => b - a)[0];", method: "max", position: [4]int{1, 17, 1, 47}, suggestion: "const maximum = Math.max(...array);"},
		{code: "const maximum = array.toSorted((a, b) => b - a)[0];", method: "max", position: [4]int{1, 17, 1, 51}, suggestion: "const maximum = Math.max(...array);"},
		{code: "const minimum = array.sort((a, b) => b - a).at(-1);", method: "min", position: [4]int{1, 17, 1, 51}, suggestion: "const minimum = Math.min(...array);"},
		{code: "const minimum = array.toSorted((a, b) => b - a).at(-1);", method: "min", position: [4]int{1, 17, 1, 55}, suggestion: "const minimum = Math.min(...array);"},
		{
			code: `const minimum = array.sort(function (a, b) {
	return a - b;
})[0];`,
			method:     "min",
			position:   [4]int{1, 17, 3, 6},
			suggestion: "const minimum = Math.min(...array);",
		},
		{
			code: `const maximum = array.toSorted((a, b) => {
	return a - b;
}).at(-1);`,
			method:     "max",
			position:   [4]int{1, 17, 3, 10},
			suggestion: "const maximum = Math.max(...array);",
		},
		{
			code:       "const minimum = (array ?? fallback).toSorted((a, b) => a - b)[0];",
			method:     "min",
			position:   [4]int{1, 17, 1, 65},
			suggestion: "const minimum = Math.min(...(array ?? fallback));",
		},
		{code: `const minimum = array.toSorted((a, b) => {
	// Compare numerically.
	return a - b;
})[0];`, method: "min", position: [4]int{1, 17, 4, 6}},
		{code: `class ArraySubclass extends Array {
	getMinimum() {
		return super.sort((a, b) => a - b)[0];
	}
}`, method: "min", position: [4]int{3, 10, 3, 40}},
		{
			code:       "const minimum = array.sort((a: number, b: number) => a - b)[0];",
			fileName:   "input.ts",
			method:     "min",
			position:   [4]int{1, 17, 1, 63},
			suggestion: "const minimum = Math.min(...array);",
		},
		{
			code:       "const minimum = array.toSorted((a, b) => (a - b) as number)[0];",
			fileName:   "input.ts",
			method:     "min",
			position:   [4]int{1, 17, 1, 63},
			suggestion: "const minimum = Math.min(...array);",
		},
		{
			code:       "const minimum = array.toSorted((a, b) => a - b satisfies number)[0];",
			fileName:   "input.ts",
			method:     "min",
			position:   [4]int{1, 17, 1, 68},
			suggestion: "const minimum = Math.min(...array);",
		},
		{
			code:       "function f(array: Int8Array) { return array.toSorted((a, b) => a - b)[0]; }",
			fileName:   "input.ts",
			method:     "min",
			position:   [4]int{1, 39, 1, 73},
			suggestion: "function f(array: Int8Array) { return Math.min(...array); }",
		},
		{
			code:       "for (const value of array.sort((a, b) => a - b)[0]) {}",
			method:     "min",
			position:   [4]int{1, 21, 1, 51},
			suggestion: "for (const value of Math.min(...array)) {}",
		},
		// Constructor callee.
		// The runtime constructor test uses the same first two inputs.
		{
			code:       "new ([3, 1, 2].toSorted((a, b) => a - b))[0].constructor(\"42\");",
			method:     "min",
			position:   [4]int{1, 5, 1, 45},
			suggestion: "new (Math.min(...[3, 1, 2])).constructor(\"42\");",
		},
		{
			code:       "new (([3, 1, 2].toSorted((a, b) => a - b))[0]).constructor(\"42\");",
			method:     "min",
			position:   [4]int{1, 6, 1, 46},
			suggestion: "new (Math.min(...[3, 1, 2])).constructor(\"42\");",
		},
		{
			code:       "const minimum = [3, 1, 2].toSorted((a, b) => a - b)[0];",
			method:     "min",
			position:   [4]int{1, 17, 1, 55},
			suggestion: "const minimum = Math.min(...[3, 1, 2]);",
		},
		// TypeScript constructor callee.
		{
			code:       "new ([3, 1, 2].toSorted((a, b) => a - b))[0]!.constructor(\"42\");",
			fileName:   "input.ts",
			method:     "min",
			position:   [4]int{1, 5, 1, 45},
			suggestion: "new (Math.min(...[3, 1, 2]))!.constructor(\"42\");",
		},
	}
	runMinMaxCases(t, valid, invalid)
}

type minMaxCase struct {
	code       string
	fileName   string
	method     string
	position   [4]int
	suggestion string
}

func runMinMaxCases(t *testing.T, valid []rule_tester.ValidTestCase, cases []minMaxCase) {
	t.Helper()
	for i := range valid {
		if valid[i].FileName == "" {
			valid[i].FileName = "input.js"
		}
	}
	invalid := make([]rule_tester.InvalidTestCase, 0, len(cases))
	for _, item := range cases {
		if item.fileName == "" {
			item.fileName = "input.js"
		}
		extreme := "minimum"
		if item.method == "max" {
			extreme = "maximum"
		}
		expected := rule_tester.InvalidTestCaseError{
			MessageId: "no-array-sort-for-min-max/error",
			Message:   "Use `Math." + item.method + "()` instead of sorting to get the " + extreme + " value.",
			Line:      item.position[0],
			Column:    item.position[1],
			EndLine:   item.position[2],
			EndColumn: item.position[3],
		}
		if item.suggestion != "" {
			expected.Suggestions = []rule_tester.InvalidTestCaseSuggestion{{
				MessageId: "no-array-sort-for-min-max/suggestion",
				Output:    item.suggestion,
			}}
		}
		invalid = append(invalid, rule_tester.InvalidTestCase{
			Code: item.code, FileName: item.fileName, Errors: []rule_tester.InvalidTestCaseError{expected},
		})
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoArraySortForMinMaxRule, valid, invalid)
}
