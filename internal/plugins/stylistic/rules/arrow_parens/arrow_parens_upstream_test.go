// TestArrowParensUpstream migrates the full valid/invalid suite from upstream
// packages/eslint-plugin/rules/arrow-parens/arrow-parens.test.ts 1:1. Position
// assertions cover line/column for every invalid case. rslint-specific
// lock-in cases live in arrow_parens_extras_test.go.
package arrow_parens_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/rules/arrow_parens"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func optsAlways() []any   { return []any{"always"} }
func optsAsNeeded() []any { return []any{"as-needed"} }
func optsAsNeededBlock() []any {
	return []any{"as-needed", map[string]any{"requireForBlockBody": true}}
}
func optsAsNeededFlag(b bool) []any {
	return []any{"as-needed", map[string]any{"requireForBlockBody": b}}
}

func TestArrowParensUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&arrow_parens.ArrowParensRule,
		[]rule_tester.ValidTestCase{
			// ---- "always" (by default) ----
			{Code: `() => {}`},
			{Code: `(a) => {}`},
			{Code: `(a) => a`},
			{Code: "(a) => {\n}"},
			{Code: `a.then((foo) => {});`},
			{Code: `a.then((foo) => { if (true) {}; });`},
			{Code: `const f = (/* */a) => a + a;`},
			{Code: `const f = (a/** */) => a + a;`},
			{Code: "const f = (a//\n) => a + a;"},
			{Code: "const f = (//\na) => a + a;"},
			{Code: "const f = (/*\n */a//\n) => a + a;"},
			{Code: `const f = (/** @type {number} */a/**hello*/) => a + a;`},
			{Code: `a.then(async (foo) => { if (true) {}; });`},

			// ---- "always" (explicit) ----
			{Code: `() => {}`, Options: optsAlways()},
			{Code: `(a) => {}`, Options: optsAlways()},
			{Code: `(a) => a`, Options: optsAlways()},
			{Code: "(a) => {\n}", Options: optsAlways()},
			{Code: `a.then((foo) => {});`, Options: optsAlways()},
			{Code: `a.then((foo) => { if (true) {}; });`, Options: optsAlways()},
			{Code: `a.then(async (foo) => { if (true) {}; });`, Options: optsAlways()},

			// ---- "as-needed" ----
			{Code: `() => {}`, Options: optsAsNeeded()},
			{Code: `a => {}`, Options: optsAsNeeded()},
			{Code: `a => a`, Options: optsAsNeeded()},
			{Code: `a => (a)`, Options: optsAsNeeded()},
			{Code: `(a => a)`, Options: optsAsNeeded()},
			{Code: `((a => a))`, Options: optsAsNeeded()},
			{Code: `([a, b]) => {}`, Options: optsAsNeeded()},
			{Code: `({ a, b }) => {}`, Options: optsAsNeeded()},
			{Code: `(a = 10) => {}`, Options: optsAsNeeded()},
			{Code: `(...a) => a[0]`, Options: optsAsNeeded()},
			{Code: `(a, b) => {}`, Options: optsAsNeeded()},
			{Code: `async a => a`, Options: optsAsNeeded()},
			{Code: `async ([a, b]) => {}`, Options: optsAsNeeded()},
			{Code: `async (a, b) => {}`, Options: optsAsNeeded()},

			// ---- "as-needed", { requireForBlockBody: true } ----
			{Code: `() => {}`, Options: optsAsNeededBlock()},
			{Code: `a => a`, Options: optsAsNeededBlock()},
			{Code: `a => (a)`, Options: optsAsNeededBlock()},
			{Code: `(a => a)`, Options: optsAsNeededBlock()},
			{Code: `((a => a))`, Options: optsAsNeededBlock()},
			{Code: `([a, b]) => {}`, Options: optsAsNeededBlock()},
			{Code: `([a, b]) => a`, Options: optsAsNeededBlock()},
			{Code: `({ a, b }) => {}`, Options: optsAsNeededBlock()},
			{Code: `({ a, b }) => a + b`, Options: optsAsNeededBlock()},
			{Code: `(a = 10) => {}`, Options: optsAsNeededBlock()},
			{Code: `(...a) => a[0]`, Options: optsAsNeededBlock()},
			{Code: `(a, b) => {}`, Options: optsAsNeededBlock()},
			{Code: `a => ({})`, Options: optsAsNeededBlock()},
			{Code: `async a => ({})`, Options: optsAsNeededBlock()},
			{Code: `async a => a`, Options: optsAsNeededBlock()},

			// ---- "as-needed" — comments inside parens keep them ----
			{Code: `const f = (/** @type {number} */a/**hello*/) => a + a;`, Options: optsAsNeeded()},
			{Code: `const f = (/* */a) => a + a;`, Options: optsAsNeeded()},
			{Code: `const f = (a/** */) => a + a;`, Options: optsAsNeeded()},
			{Code: "const f = (a//\n) => a + a;", Options: optsAsNeeded()},
			{Code: "const f = (//\na) => a + a;", Options: optsAsNeeded()},
			{Code: "const f = (/*\n */a//\n) => a + a;", Options: optsAsNeeded()},
			{Code: `var foo = (a,/**/) => b;`, Options: optsAsNeeded()},
			{Code: `var foo = (a , /**/) => b;`, Options: optsAsNeeded()},
			{Code: "var foo = (a\n,\n/**/) => b;", Options: optsAsNeeded()},
			{Code: "var foo = (a,//\n) => b;", Options: optsAsNeeded()},
			{Code: `const i = (a/**/,) => a + a;`, Options: optsAsNeeded()},
			{Code: "const i = (a \n /**/,) => a + a;", Options: optsAsNeeded()},
			{Code: `var bar = ({/*comment here*/a}) => a`, Options: optsAsNeeded()},
			{Code: `var bar = (/*comment here*/{a}) => a`, Options: optsAsNeeded()},

			// ---- generics — parens are forced by the `<T>` prefix ----
			{Code: `<T>(a) => b`, Options: optsAlways(), Tsx: false},
			{Code: `<T>(a) => b`, Options: optsAsNeeded(), Tsx: false},
			{Code: `<T>(a) => b`, Options: optsAsNeededBlock(), Tsx: false},
			{Code: `async <T>(a) => b`, Options: optsAlways(), Tsx: false},
			{Code: `async <T>(a) => b`, Options: optsAsNeeded(), Tsx: false},
			{Code: `async <T>(a) => b`, Options: optsAsNeededBlock(), Tsx: false},
			{Code: `<T>() => b`, Options: optsAlways(), Tsx: false},
			{Code: `<T>() => b`, Options: optsAsNeeded(), Tsx: false},
			{Code: `<T>() => b`, Options: optsAsNeededBlock(), Tsx: false},
			{Code: `<T extends A>(a) => b`, Options: optsAlways(), Tsx: false},
			{Code: `<T extends A>(a) => b`, Options: optsAsNeeded(), Tsx: false},
			{Code: `<T extends A>(a) => b`, Options: optsAsNeededBlock(), Tsx: false},
			{Code: `<T extends (A | B) & C>(a) => b`, Options: optsAlways(), Tsx: false},
			{Code: `<T extends (A | B) & C>(a) => b`, Options: optsAsNeeded(), Tsx: false},
			{Code: `<T extends (A | B) & C>(a) => b`, Options: optsAsNeededBlock(), Tsx: false},
		},
		[]rule_tester.InvalidTestCase{
			// ---- "always" (by default) ----
			{
				Code:   `a => {}`,
				Output: []string{`(a) => {}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedParens", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
				},
			},
			{
				Code:   `a => a`,
				Output: []string{`(a) => a`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedParens", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
				},
			},
			{
				Code:   "a => {\n}",
				Output: []string{"(a) => {\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedParens", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
				},
			},
			{
				Code:   `a.then(foo => {});`,
				Output: []string{`a.then((foo) => {});`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedParens", Line: 1, Column: 8, EndLine: 1, EndColumn: 11},
				},
			},
			{
				Code:   `a.then(foo => a);`,
				Output: []string{`a.then((foo) => a);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedParens", Line: 1, Column: 8, EndLine: 1, EndColumn: 11},
				},
			},
			{
				Code:   `a(foo => { if (true) {}; });`,
				Output: []string{`a((foo) => { if (true) {}; });`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedParens", Line: 1, Column: 3, EndLine: 1, EndColumn: 6},
				},
			},
			{
				Code:   `a(async foo => { if (true) {}; });`,
				Output: []string{`a(async (foo) => { if (true) {}; });`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedParens", Line: 1, Column: 9, EndLine: 1, EndColumn: 12},
				},
			},

			// ---- "as-needed" ----
			{
				Code:    `(a) => a`,
				Output:  []string{`a => a`},
				Options: optsAsNeeded(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedParens", Line: 1, Column: 2, EndLine: 1, EndColumn: 3},
				},
			},
			{
				Code:    `(  a  ) => b`,
				Output:  []string{`a => b`},
				Options: optsAsNeeded(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedParens", Line: 1, Column: 4, EndLine: 1, EndColumn: 5},
				},
			},
			{
				Code:    "(\na\n) => b",
				Output:  []string{`a => b`},
				Options: optsAsNeeded(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedParens", Line: 2, Column: 1, EndLine: 2, EndColumn: 2},
				},
			},
			{
				Code:    `(a,) => a`,
				Output:  []string{`a => a`},
				Options: optsAsNeeded(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedParens", Line: 1, Column: 2, EndLine: 1, EndColumn: 3},
				},
			},
			{
				Code:    `async (a) => a`,
				Output:  []string{`async a => a`},
				Options: optsAsNeeded(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedParens", Line: 1, Column: 8, EndLine: 1, EndColumn: 9},
				},
			},
			{
				Code:    `async(a) => a`,
				Output:  []string{`async a => a`},
				Options: optsAsNeeded(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedParens", Line: 1, Column: 7, EndLine: 1, EndColumn: 8},
				},
			},
			{
				Code:    `typeof((a) => {})`,
				Output:  []string{`typeof(a => {})`},
				Options: optsAsNeeded(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedParens", Line: 1, Column: 9, EndLine: 1, EndColumn: 10},
				},
			},
			{
				Code:    `function *f() { yield(a) => a; }`,
				Output:  []string{`function *f() { yield a => a; }`},
				Options: optsAsNeeded(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedParens", Line: 1, Column: 23, EndLine: 1, EndColumn: 24},
				},
			},

			// ---- "as-needed", { requireForBlockBody: true } ----
			{
				Code:    `a => {}`,
				Output:  []string{`(a) => {}`},
				Options: optsAsNeededBlock(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedParensBlock", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
				},
			},
			{
				Code:    `(a) => a`,
				Output:  []string{`a => a`},
				Options: optsAsNeededBlock(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedParensInline", Line: 1, Column: 2, EndLine: 1, EndColumn: 3},
				},
			},
			{
				Code:    `async a => {}`,
				Output:  []string{`async (a) => {}`},
				Options: optsAsNeededBlock(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedParensBlock", Line: 1, Column: 7, EndLine: 1, EndColumn: 8},
				},
			},
			{
				Code:    `async (a) => a`,
				Output:  []string{`async a => a`},
				Options: optsAsNeededBlock(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedParensInline", Line: 1, Column: 8, EndLine: 1, EndColumn: 9},
				},
			},
			{
				Code:    `async(a) => a`,
				Output:  []string{`async a => a`},
				Options: optsAsNeededBlock(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedParensInline", Line: 1, Column: 7, EndLine: 1, EndColumn: 8},
				},
			},
			{
				Code:    `const f = /** @type {number} */(a)/**hello*/ => a + a;`,
				Output:  []string{`const f = /** @type {number} */a/**hello*/ => a + a;`},
				Options: optsAsNeeded(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedParens", Line: 1, Column: 33, EndLine: 1, EndColumn: 34},
				},
			},
			{
				Code:    "const f = //\n(a) => a + a;",
				Output:  []string{"const f = //\na => a + a;"},
				Options: optsAsNeeded(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedParens", Line: 2, Column: 2, EndLine: 2, EndColumn: 3},
				},
			},
			{
				Code:   `var foo = /**/ a => b;`,
				Output: []string{`var foo = /**/ (a) => b;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedParens", Line: 1, Column: 16, EndLine: 1, EndColumn: 17},
				},
			},
			{
				Code:   `var bar = a /**/ =>  b;`,
				Output: []string{`var bar = (a) /**/ =>  b;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedParens", Line: 1, Column: 11, EndLine: 1, EndColumn: 12},
				},
			},
			{
				Code:   "const foo = a => {};\n\n// comment between 'a' and an unrelated closing paren\n\nbar();",
				Output: []string{"const foo = (a) => {};\n\n// comment between 'a' and an unrelated closing paren\n\nbar();"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedParens", Line: 1, Column: 13, EndLine: 1, EndColumn: 14},
				},
			},
		},
	)
}
