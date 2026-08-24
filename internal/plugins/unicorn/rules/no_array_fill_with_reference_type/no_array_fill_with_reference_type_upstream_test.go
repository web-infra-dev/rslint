// TestNoArrayFillWithReferenceTypeUpstream migrates the full valid/invalid
// suite from upstream test/no-array-fill-with-reference-type.js 1:1. Position
// assertions cover line/column for every invalid case. rslint-specific lock-in
// cases live in no_array_fill_with_reference_type_extras_test.go.
package no_array_fill_with_reference_type_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	no_array_fill_with_reference_type "github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_array_fill_with_reference_type"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	messageID = "no-array-fill-with-reference-type"
	message   = "Do not use a reference value as the fill value."
)

func upstreamError(line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: messageID,
		Message:   message,
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
}

func TestNoArrayFillWithReferenceTypeUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_array_fill_with_reference_type.NoArrayFillWithReferenceTypeRule,
		[]rule_tester.ValidTestCase{
			// ---- JavaScript ----
			{Code: `array.fill(0)`, FileName: "file.js"},
			{Code: `array.fill("x")`, FileName: "file.js"},
			{Code: `array.fill(false)`, FileName: "file.js"},
			{Code: `array.fill(null)`, FileName: "file.js"},
			{Code: `array.fill(undefined)`, FileName: "file.js"},
			{Code: `array.fill(10n)`, FileName: "file.js"},
			{Code: `array.fill(Symbol("x"))`, FileName: "file.js"},
			{Code: `array.fill(Symbol.for("x"))`, FileName: "file.js"},
			{Code: `array.fill(Symbol.iterator)`, FileName: "file.js"},
			{Code: "array.fill(`x`)", FileName: "file.js"},
			{Code: `array.fill(() => {})`, FileName: "file.js"},
			{Code: `array.fill(function () {})`, FileName: "file.js"},
			{Code: `array.fill(/x/)`, FileName: "file.js"},
			{Code: `array.fill(new RegExp("x"))`, FileName: "file.js"},
			{Code: `const value = new RegExp("x"); array.fill(value)`, FileName: "file.js"},
			{Code: `let value = {}; value = 1; array.fill(value)`, FileName: "file.js"},
			{Code: `var value = {}; array.fill(value)`, FileName: "file.js"},
			{Code: `const value = {}; const alias = value; array.fill(alias)`, FileName: "file.js"},
			{Code: `array.fill(object.value)`, FileName: "file.js"},
			{Code: `array.fill(this.value)`, FileName: "file.js"},
			{Code: `array?.fill({})`, FileName: "file.js"},
			{Code: `array.fill?.({})`, FileName: "file.js"},
			{Code: `array[fill]({})`, FileName: "file.js"},
			{Code: `Array.from({length: 3}, () => value)`, FileName: "file.js"},
			{Code: `Array.from({length: 3}).map(() => value)`, FileName: "file.js"},
			{Code: `const {value = {}} = object; array.fill(value)`, FileName: "file.js"},
			{Code: `function foo(value = {}) { array.fill(value); }`, FileName: "file.js"},

			// ---- TypeScript ----
			{Code: `array.fill(/x/ as RegExp)`, FileName: "file.ts", Tsx: false},
			{Code: `array.fill((() => {}) as Function)`, FileName: "file.ts", Tsx: false},
			{Code: `array.fill(object.value as Foo)`, FileName: "file.ts", Tsx: false},
			{Code: `const value = {}; const alias = value as Foo; array.fill(alias)`, FileName: "file.ts", Tsx: false},
			{Code: `function f(foo: Uint8Array) { foo.fill({}); }`, FileName: "file.ts", Tsx: false},
			{Code: `const foo = new Uint8Array(3); foo.fill({});`, FileName: "file.ts", Tsx: false},
			{Code: `function f(foo: Set<object>) { foo.fill({}); }`, FileName: "file.ts", Tsx: false},
		},
		[]rule_tester.InvalidTestCase{
			// ---- JavaScript ----
			{Code: `new Array(3).fill({})`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{upstreamError(1, 19, 1, 21)}},
			{Code: `Array(3).fill([])`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{upstreamError(1, 15, 1, 17)}},
			{Code: `Array.from({length: 3}).fill(new Map())`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{upstreamError(1, 30, 1, 39)}},
			{Code: `[1, 2, 3].fill(new Set())`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{upstreamError(1, 16, 1, 25)}},
			{Code: `const value = {}; array.fill(value)`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{upstreamError(1, 30, 1, 35)}},
			{Code: `const value = []; array.fill(value)`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{upstreamError(1, 30, 1, 35)}},
			{Code: `const value = new Map(); array.fill(value)`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{upstreamError(1, 37, 1, 42)}},
			{Code: `const value = new class {}; array.fill(value)`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{upstreamError(1, 40, 1, 45)}},
			{Code: `const RegExp = class {}; array.fill(new RegExp())`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{upstreamError(1, 37, 1, 49)}},
			{Code: `array.fill(class {})`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{upstreamError(1, 12, 1, 20)}},
			{
				Code:     "const value = {};\narray.fill(value, 1);",
				FileName: "file.js",
				Errors:   []rule_tester.InvalidTestCaseError{upstreamError(2, 12, 2, 17)},
			},

			// ---- TypeScript ----
			{Code: `array.fill({} as Foo)`, FileName: "file.ts", Tsx: false, Errors: []rule_tester.InvalidTestCaseError{upstreamError(1, 12, 1, 21)}},
			{Code: `array.fill(<Foo>{})`, FileName: "file.ts", Tsx: false, Errors: []rule_tester.InvalidTestCaseError{upstreamError(1, 12, 1, 19)}},
			{Code: `array.fill({} satisfies Foo)`, FileName: "file.ts", Tsx: false, Errors: []rule_tester.InvalidTestCaseError{upstreamError(1, 12, 1, 28)}},
			{Code: `array.fill({}!)`, FileName: "file.ts", Tsx: false, Errors: []rule_tester.InvalidTestCaseError{upstreamError(1, 12, 1, 15)}},
			{Code: `const value = {} as Foo; array.fill(value)`, FileName: "file.ts", Tsx: false, Errors: []rule_tester.InvalidTestCaseError{upstreamError(1, 37, 1, 42)}},
			{Code: `const value = {}; array.fill(value!)`, FileName: "file.ts", Tsx: false, Errors: []rule_tester.InvalidTestCaseError{upstreamError(1, 30, 1, 36)}},
			{Code: `function f(foo: object[]) { foo.fill({}); }`, FileName: "file.ts", Tsx: false, Errors: []rule_tester.InvalidTestCaseError{upstreamError(1, 38, 1, 40)}},
			{Code: `function f(foo) { foo.fill({}); }`, FileName: "file.ts", Tsx: false, Errors: []rule_tester.InvalidTestCaseError{upstreamError(1, 28, 1, 30)}},
		},
	)
}
