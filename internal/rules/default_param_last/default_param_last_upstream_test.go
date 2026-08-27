package default_param_last

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestDefaultParamLastUpstream migrates the full valid/invalid suite from upstream eslint/tests/lib/rules/default-param-last.js 1:1.
// Position assertions cover line/column for every invalid case. rslint-specific lock-in cases live in default_param_last_extras_test.go.
func TestDefaultParamLastUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&DefaultParamLastRule,
		upstreamValidCases(),
		upstreamInvalidCases(),
	)
}

func upstreamValidCases() []rule_tester.ValidTestCase {
	codes := []string{
		// ---- ESLint core JavaScript valid ----
		"function f() {}",
		"function f(a) {}",
		"function f(a = 5) {}",
		"function f(a, b) {}",
		"function f(a, b = 5) {}",
		"function f(a, b = 5, c = 5) {}",
		"function f(a, b = 5, ...c) {}",
		"const f = () => {}",
		"const f = (a) => {}",
		"const f = (a = 5) => {}",
		"const f = function f() {}",
		"const f = function f(a) {}",
		"const f = function f(a = 5) {}",

		// ---- ESLint core TypeScript valid: function declarations ----
		"function foo() {}",
		"function foo(a: number) {}",
		"function foo(a = 1) {}",
		"function foo(a?: number) {}",
		"function foo(a: number, b: number) {}",
		"function foo(a: number, b: number, c?: number) {}",
		"function foo(a: number, b = 1) {}",
		"function foo(a: number, b = 1, c = 1) {}",
		"function foo(a: number, b = 1, c?: number) {}",
		"function foo(a: number, b?: number, c = 1) {}",
		"function foo(a: number, b = 1, ...c) {}",

		// ---- ESLint core TypeScript valid: function expressions ----
		"const foo = function () {};",
		"const foo = function (a: number) {};",
		"const foo = function (a = 1) {};",
		"const foo = function (a?: number) {};",
		"const foo = function (a: number, b: number) {};",
		"const foo = function (a: number, b: number, c?: number) {};",
		"const foo = function (a: number, b = 1) {};",
		"const foo = function (a: number, b = 1, c = 1) {};",
		"const foo = function (a: number, b = 1, c?: number) {};",
		"const foo = function (a: number, b?: number, c = 1) {};",
		"const foo = function (a: number, b = 1, ...c) {};",

		// ---- ESLint core TypeScript valid: arrows ----
		"const foo = () => {};",
		"const foo = (a: number) => {};",
		"const foo = (a = 1) => {};",
		"const foo = (a?: number) => {};",
		"const foo = (a: number, b: number) => {};",
		"const foo = (a: number, b: number, c?: number) => {};",
		"const foo = (a: number, b = 1) => {};",
		"const foo = (a: number, b = 1, c = 1) => {};",
		"const foo = (a: number, b = 1, c?: number) => {};",
		"const foo = (a: number, b?: number, c = 1) => {};",
		"const foo = (a: number, b = 1, ...c) => {};",

		// ---- ESLint core TypeScript valid: constructors and parameter properties ----
		`class Foo { constructor(a: number, b: number, c: number) {} }`,
		`class Foo { constructor(a: number, b?: number, c = 1) {} }`,
		`class Foo { constructor(a: number, b = 1, c?: number) {} }`,
		`class Foo { constructor(public a: number, protected b: number, private c: number) {} }`,
		`class Foo { constructor(public a: number, protected b?: number, private c = 10) {} }`,
		`class Foo { constructor(public a: number, protected b = 10, private c?: number) {} }`,
		`class Foo { constructor(a: number, protected b?: number, private c = 0) {} }`,
		`class Foo { constructor(a: number, b?: number, private c = 0) {} }`,
		`class Foo { constructor(a: number, private b?: number, c = 0) {} }`,
	}

	valid := make([]rule_tester.ValidTestCase, 0, len(codes))
	for _, code := range codes {
		valid = append(valid, rule_tester.ValidTestCase{Code: code})
	}
	return valid
}

func upstreamInvalidCases() []rule_tester.InvalidTestCase {
	return []rule_tester.InvalidTestCase{
		// ---- ESLint core JavaScript invalid ----
		invalidAt("function f(a = 5, b) {}", positions(1, 12, 17)),
		invalidAt("function f(a = 5, b = 6, c) {}", positions(1, 12, 17), positions(1, 19, 24)),
		invalidAt("function f (a = 5, b, c = 6, d) {}", positions(1, 13, 18), positions(1, 23, 28)),
		invalidAt("function f(a = 5, b, c = 5) {}", positions(1, 12, 17)),
		invalidAt("const f = (a = 5, b, ...c) => {}", positions(1, 12, 17)),
		invalidAt("const f = function f (a, b = 5, c) {}", positions(1, 26, 31)),
		invalidAt("const f = (a = 5, { b }) => {}", positions(1, 12, 17)),
		invalidAt("const f = ({ a } = {}, b) => {}", positions(1, 12, 22)),
		invalidAt("const f = ({ a, b } = { a: 1, b: 2 }, c) => {}", positions(1, 12, 37)),
		invalidAt("const f = ([a] = [], b) => {}", positions(1, 12, 20)),
		invalidAt("const f = ([a, b] = [1, 2], c) => {}", positions(1, 12, 27)),

		// ---- ESLint core TypeScript invalid: function declarations ----
		invalidAt("function foo(a = 1, b: number) {}", positions(1, 14, 19)),
		invalidAt("function foo(a = 1, b = 2, c: number) {}", positions(1, 14, 19), positions(1, 21, 26)),
		invalidAt("function foo(a = 1, b: number, c = 2, d: number) {}", positions(1, 14, 19), positions(1, 32, 37)),
		invalidAt("function foo(a = 1, b: number, c = 2) {}", positions(1, 14, 19)),
		invalidAt("function foo(a = 1, b: number, ...c) {}", positions(1, 14, 19)),
		invalidAt("function foo(a?: number, b: number) {}", positions(1, 14, 24)),
		invalidAt("function foo(a: number, b?: number, c: number) {}", positions(1, 25, 35)),
		invalidAt("function foo(a = 1, b?: number, c: number) {}", positions(1, 14, 19), positions(1, 21, 31)),

		// ---- ESLint core TypeScript invalid: function expressions ----
		invalidAt("const foo = function (a = 1, b: number) {};", positions(1, 23, 28)),
		invalidAt("const foo = function (a = 1, b = 2, c: number) {};", positions(1, 23, 28), positions(1, 30, 35)),
		invalidAt("const foo = function (a = 1, b: number, c = 2, d: number) {};", positions(1, 23, 28), positions(1, 41, 46)),
		invalidAt("const foo = function (a = 1, b: number, c = 2) {};", positions(1, 23, 28)),
		invalidAt("const foo = function (a = 1, b: number, ...c) {};", positions(1, 23, 28)),
		invalidAt("const foo = function (a?: number, b: number) {};", positions(1, 23, 33)),
		invalidAt("const foo = function (a: number, b?: number, c: number) {};", positions(1, 34, 44)),
		invalidAt("const foo = function (a = 1, b?: number, c: number) {};", positions(1, 23, 28), positions(1, 30, 40)),

		// ---- ESLint core TypeScript invalid: arrows ----
		invalidAt("const foo = (a = 1, b: number) => {};", positions(1, 14, 19)),
		invalidAt("const foo = (a = 1, b = 2, c: number) => {};", positions(1, 14, 19), positions(1, 21, 26)),
		invalidAt("const foo = (a = 1, b: number, c = 2, d: number) => {};", positions(1, 14, 19), positions(1, 32, 37)),
		invalidAt("const foo = (a = 1, b: number, c = 2) => {};", positions(1, 14, 19)),
		invalidAt("const foo = (a = 1, b: number, ...c) => {};", positions(1, 14, 19)),
		invalidAt("const foo = (a?: number, b: number) => {};", positions(1, 14, 24)),
		invalidAt("const foo = (a: number, b?: number, c: number) => {};", positions(1, 25, 35)),
		invalidAt("const foo = (a = 1, b?: number, c: number) => {};", positions(1, 14, 19), positions(1, 21, 31)),

		// ---- ESLint core TypeScript invalid: constructors and parameter properties ----
		invalidAt("\nclass Foo {\n  constructor(\n    public a: number,\n    protected b?: number,\n    private c: number,\n  ) {}\n}", positions(5, 5, 25)),
		invalidAt("\nclass Foo {\n  constructor(\n    public a: number,\n    protected b = 0,\n    private c: number,\n  ) {}\n}", positions(5, 5, 20)),
		invalidAt("\nclass Foo {\n  constructor(\n    public a?: number,\n    private b: number,\n  ) {}\n}", positions(4, 5, 22)),
		invalidAt("\nclass Foo {\n  constructor(\n    public a = 0,\n    private b: number,\n  ) {}\n}", positions(4, 5, 17)),
		invalidAt("\nclass Foo {\n  constructor(a = 0, b: number) {}\n}", positions(3, 15, 20)),
		invalidAt("\nclass Foo {\n  constructor(a?: number, b: number) {}\n}", positions(3, 15, 25)),
	}
}

type expectedPosition struct {
	line      int
	column    int
	endColumn int
}

func positions(line, column, endColumn int) expectedPosition {
	return expectedPosition{line: line, column: column, endColumn: endColumn}
}

func invalidAt(code string, positions ...expectedPosition) rule_tester.InvalidTestCase {
	errors := make([]rule_tester.InvalidTestCaseError, 0, len(positions))
	for i, pos := range positions {
		err := rule_tester.InvalidTestCaseError{
			MessageId: "shouldBeLast",
			Line:      pos.line,
			Column:    pos.column,
			EndLine:   pos.line,
			EndColumn: pos.endColumn,
		}
		if i == 0 {
			err.Message = "Default parameters should be last."
		}
		errors = append(errors, err)
	}
	return rule_tester.InvalidTestCase{Code: code, Errors: errors}
}
