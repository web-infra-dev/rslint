package no_unneeded_ternary

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnneededTernaryRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnneededTernaryRule,

		[]rule_tester.ValidTestCase{
			// ---- Upstream ESLint suite ----
			{Code: `config.newIsCap = config.newIsCap !== false`},
			{Code: `var a = x === 2 ? 'Yes' : 'No';`},
			{Code: `var a = x === 2 ? true : 'No';`},
			{Code: `var a = x === 2 ? 'Yes' : false;`},
			{Code: `var a = x === 2 ? 'true' : 'false';`},
			{Code: `var a = foo ? foo : bar;`},
			{Code: `var value = 'a';var canSet = true;var result = value || (canSet ? 'unset' : 'can not set')`},
			{Code: `var a = foo ? bar : foo;`},
			{Code: `foo ? bar : foo;`},
			{Code: `var a = f(x ? x : 1)`},
			{Code: `f(x ? x : 1);`},
			{Code: `foo ? foo : bar;`},
			{Code: `var a = foo ? 'Yes' : foo;`},
			{
				Code:    `var a = foo ? 'Yes' : foo;`,
				Options: map[string]interface{}{"defaultAssignment": false},
			},
			{
				Code:    `var a = foo ? bar : foo;`,
				Options: map[string]interface{}{"defaultAssignment": false},
			},
			{
				Code:    `foo ? bar : foo;`,
				Options: map[string]interface{}{"defaultAssignment": false},
			},

			// ---- Default-assignment pattern stays valid by default ----
			{Code: `var a = foo ? foo : 'No';`},
			{Code: `var a = foo ? foo : bar;`},

			// ---- Mixed boolean + non-boolean: not both literals → no report ----
			{Code: `var a = foo ? true : null;`},
			{Code: `var a = foo ? null : false;`},
			{Code: `var a = foo ? 1 : 0;`},

			// ---- Conditional that isn't a default-assignment pattern (test ≠ consequent) ----
			{
				Code:    `var a = foo ? bar : 1;`,
				Options: map[string]interface{}{"defaultAssignment": false},
			},
			// Property access doesn't match the simple-identifier pattern.
			{
				Code:    `var a = foo.bar ? foo.bar : baz;`,
				Options: map[string]interface{}{"defaultAssignment": false},
			},

			// ---- Explicit `defaultAssignment: true` matches the no-options default ----
			{
				Code:    `var a = foo ? foo : bar;`,
				Options: map[string]interface{}{"defaultAssignment": true},
			},
			// Empty options object: same as defaults.
			{
				Code:    `var a = foo ? foo : bar;`,
				Options: map[string]interface{}{},
			},
			// Identifier vs different identifier: not a default-assignment pattern.
			{
				Code:    `var a = foo ? bar : baz;`,
				Options: map[string]interface{}{"defaultAssignment": false},
			},
			// Test and consequent are SAME identifier text but different tokens
			// (different `Identifier` node references); both are "foo" so the rule
			// flags this — keep this in invalid.

			// ---- Test that *contains* a boolean literal but isn't itself one ----
			{Code: `var a = (x === true) ? "Y" : "N";`},
			{Code: `var a = !!x ? "Y" : "N";`},
		},

		[]rule_tester.InvalidTestCase{
			// ---- Upstream ESLint suite — boolean-literal ternary ----
			{
				Code:   `var a = x === 2 ? true : false;`,
				Output: []string{`var a = x === 2;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalExpression", Message: "Unnecessary use of boolean literals in conditional expression.", Line: 1, Column: 9, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code:   `var a = x >= 2 ? true : false;`,
				Output: []string{`var a = x >= 2;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalExpression", Line: 1, Column: 9, EndLine: 1, EndColumn: 30},
				},
			},
			{
				Code:   `var a = x ? true : false;`,
				Output: []string{`var a = !!x;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalExpression", Line: 1, Column: 9, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code:   `var a = x === 1 ? false : true;`,
				Output: []string{`var a = x !== 1;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalExpression", Line: 1, Column: 9, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code:   `var a = x != 1 ? false : true;`,
				Output: []string{`var a = x == 1;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalExpression", Line: 1, Column: 9, EndLine: 1, EndColumn: 30},
				},
			},
			{
				Code:   `var a = foo() ? false : true;`,
				Output: []string{`var a = !foo();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalExpression", Line: 1, Column: 9, EndLine: 1, EndColumn: 29},
				},
			},
			{
				Code:   `var a = !foo() ? false : true;`,
				Output: []string{`var a = !!foo();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalExpression", Line: 1, Column: 9, EndLine: 1, EndColumn: 30},
				},
			},
			{
				Code:   `var a = foo + bar ? false : true;`,
				Output: []string{`var a = !(foo + bar);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalExpression", Line: 1, Column: 9, EndLine: 1, EndColumn: 33},
				},
			},
			{
				Code:   `var a = x instanceof foo ? false : true;`,
				Output: []string{`var a = !(x instanceof foo);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalExpression", Line: 1, Column: 9, EndLine: 1, EndColumn: 40},
				},
			},
			{
				Code:   `var a = foo ? false : false;`,
				Output: []string{`var a = false;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalExpression", Line: 1, Column: 9, EndLine: 1, EndColumn: 28},
				},
			},
			{
				// Test has side effects → no autofix even though both
				// arms collapse to the same literal.
				Code: `var a = foo() ? false : false;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalExpression", Line: 1, Column: 9, EndLine: 1, EndColumn: 30},
				},
			},
			{
				Code:   `var a = x instanceof foo ? true : false;`,
				Output: []string{`var a = x instanceof foo;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalExpression", Line: 1, Column: 9, EndLine: 1, EndColumn: 40},
				},
			},
			{
				Code:   `var a = !foo ? true : false;`,
				Output: []string{`var a = !foo;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalExpression", Line: 1, Column: 9, EndLine: 1, EndColumn: 28},
				},
			},

			// ---- Upstream ESLint suite — defaultAssignment: false ----
			{
				Code: `
                var value = 'a'
                var canSet = true
                var result = value ? value : canSet ? 'unset' : 'can not set'
            `,
				Output: []string{`
                var value = 'a'
                var canSet = true
                var result = value || (canSet ? 'unset' : 'can not set')
            `},
				Options: map[string]interface{}{"defaultAssignment": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalAssignment", Message: "Unnecessary use of conditional expression for default assignment.", Line: 4, Column: 30, EndLine: 4, EndColumn: 78},
				},
			},
			{
				Code:    `foo ? foo : (bar ? baz : qux)`,
				Output:  []string{`foo || (bar ? baz : qux)`},
				Options: map[string]interface{}{"defaultAssignment": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalAssignment", Line: 1, Column: 1, EndLine: 1, EndColumn: 30},
				},
			},
			{
				Code:    `function* fn() { foo ? foo : yield bar }`,
				Output:  []string{`function* fn() { foo || (yield bar) }`},
				Options: map[string]interface{}{"defaultAssignment": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalAssignment", Line: 1, Column: 18, EndLine: 1, EndColumn: 39},
				},
			},
			{
				Code:    `var a = foo ? foo : 'No';`,
				Output:  []string{`var a = foo || 'No';`},
				Options: map[string]interface{}{"defaultAssignment": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalAssignment", Line: 1, Column: 9, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code:    `var a = ((foo)) ? (((((foo))))) : ((((((((((((((bar))))))))))))));`,
				Output:  []string{`var a = ((foo)) || ((((((((((((((bar))))))))))))));`},
				Options: map[string]interface{}{"defaultAssignment": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalAssignment", Line: 1, Column: 9, EndLine: 1, EndColumn: 66},
				},
			},
			{
				Code:    `var a = b ? b : c => c;`,
				Output:  []string{`var a = b || (c => c);`},
				Options: map[string]interface{}{"defaultAssignment": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalAssignment", Line: 1, Column: 9, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    `var a = b ? b : c = 0;`,
				Output:  []string{`var a = b || (c = 0);`},
				Options: map[string]interface{}{"defaultAssignment": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalAssignment", Line: 1, Column: 9, EndLine: 1, EndColumn: 22},
				},
			},
			{
				Code:    `var a = b ? b : (c => c);`,
				Output:  []string{`var a = b || (c => c);`},
				Options: map[string]interface{}{"defaultAssignment": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalAssignment", Line: 1, Column: 9, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code:    `var a = b ? b : (c = 0);`,
				Output:  []string{`var a = b || (c = 0);`},
				Options: map[string]interface{}{"defaultAssignment": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalAssignment", Line: 1, Column: 9, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    `var a = b ? b : (c) => (c);`,
				Output:  []string{`var a = b || ((c) => (c));`},
				Options: map[string]interface{}{"defaultAssignment": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalAssignment", Line: 1, Column: 9, EndLine: 1, EndColumn: 27},
				},
			},
			{
				Code:    `var a = b ? b : c, d; // this is ((b ? b : c), (d))`,
				Output:  []string{`var a = b || c, d; // this is ((b ? b : c), (d))`},
				Options: map[string]interface{}{"defaultAssignment": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalAssignment", Line: 1, Column: 9, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code:    `var a = b ? b : (c, d);`,
				Output:  []string{`var a = b || (c, d);`},
				Options: map[string]interface{}{"defaultAssignment": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalAssignment", Line: 1, Column: 9, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    `f(x ? x : 1);`,
				Output:  []string{`f(x || 1);`},
				Options: map[string]interface{}{"defaultAssignment": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalAssignment", Line: 1, Column: 3, EndLine: 1, EndColumn: 12},
				},
			},
			{
				Code:    `x ? x : 1;`,
				Output:  []string{`x || 1;`},
				Options: map[string]interface{}{"defaultAssignment": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalAssignment", Line: 1, Column: 1, EndLine: 1, EndColumn: 10},
				},
			},
			{
				Code:    `var a = foo ? foo : bar;`,
				Output:  []string{`var a = foo || bar;`},
				Options: map[string]interface{}{"defaultAssignment": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalAssignment", Line: 1, Column: 9, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    `var a = foo ? foo : a ?? b;`,
				Output:  []string{`var a = foo || (a ?? b);`},
				Options: map[string]interface{}{"defaultAssignment": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalAssignment", Line: 1, Column: 9, EndLine: 1, EndColumn: 27},
				},
			},

			// ---- TypeScript-specific (replaces upstream's typescript-parser fixtures) ----
			// TS-specific kinds (AsExpression, etc.) are unknown to ESLint's
			// precedence table and get -1 → wrapped defensively. We mirror
			// that behavior in eslintLikePrecedence so TS expressions get the
			// same parens ESLint would emit.
			{
				Code:   `foo as any ? false : true`,
				Output: []string{`!(foo as any)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalExpression", Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
				},
			},
			{
				Code:    `foo ? foo : bar as any`,
				Output:  []string{`foo || (bar as any)`},
				Options: map[string]interface{}{"defaultAssignment": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalAssignment", Line: 1, Column: 1, EndLine: 1, EndColumn: 23},
				},
			},

			// ---- Extra: tsgo-specific paren-shape coverage ----
			// Outer parens around the test still classified as Identifier
			// after SkipParentheses → both-equal fix path runs.
			{
				Code:   `var a = (foo) ? true : true;`,
				Output: []string{`var a = true;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalExpression", Line: 1, Column: 9, EndLine: 1, EndColumn: 28},
				},
			},
			// Parenthesized boolean literal in either arm — boolKind peels
			// parens, so this still reports.
			{
				Code:   `var a = x ? (true) : (false);`,
				Output: []string{`var a = !!x;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalExpression", Line: 1, Column: 9, EndLine: 1, EndColumn: 29},
				},
			},
			// Multi-line position assertion: the conditional spans 3 lines.
			{
				Code:   "var a = x\n  ? true\n  : false;",
				Output: []string{`var a = !!x;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalExpression", Line: 1, Column: 9, EndLine: 3, EndColumn: 10},
				},
			},

			// ---- Extra: unary-keyword tests (typeof / delete / void / await) ----
			// `typeof x === 'string'` is already a boolean expression: unwraps cleanly.
			{
				Code:   `var a = typeof x === 'string' ? true : false;`,
				Output: []string{`var a = typeof x === 'string';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalExpression", Line: 1, Column: 9, EndLine: 1, EndColumn: 45},
				},
			},
			// `typeof x` alone is not a boolean expression: needs `!!`. Token
			// fusion is safe — the next token is the `typeof` keyword.
			{
				Code:   `var a = typeof x ? true : false;`,
				Output: []string{`var a = !!typeof x;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalExpression", Line: 1, Column: 9, EndLine: 1, EndColumn: 32},
				},
			},
			// Operator-inversion path with `===` after `typeof`.
			{
				Code:   `var a = typeof x === 'string' ? false : true;`,
				Output: []string{`var a = typeof x !== 'string';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalExpression", Line: 1, Column: 9, EndLine: 1, EndColumn: 45},
				},
			},
			// `delete obj.x ? false : true` → `!delete obj.x`. Test has side
			// effects but the rule still emits a fix (matches ESLint).
			{
				Code:   `var a = delete obj.x ? false : true;`,
				Output: []string{`var a = !delete obj.x;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalExpression", Line: 1, Column: 9, EndLine: 1, EndColumn: 36},
				},
			},
			// `void` expression as test.
			{
				Code:   `var a = void x ? true : false;`,
				Output: []string{`var a = !!void x;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalExpression", Line: 1, Column: 9, EndLine: 1, EndColumn: 30},
				},
			},
			// `await` expression as test (inside async function).
			{
				Code:   `async function f() { return await x ? true : false; }`,
				Output: []string{`async function f() { return !!await x; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalExpression", Line: 1, Column: 29, EndLine: 1, EndColumn: 51},
				},
			},

			// ---- Update expressions (++ / --) as test ----
			// Prefix update — token fusion concern: `!` + `++x` should NOT
			// merge into anything bad. `!++x` parses as `!(++x)`.
			{
				Code:   `var a = ++x ? false : true;`,
				Output: []string{`var a = !++x;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalExpression", Line: 1, Column: 9, EndLine: 1, EndColumn: 27},
				},
			},
			// Postfix update.
			{
				Code:   `var a = x++ ? true : false;`,
				Output: []string{`var a = !!x++;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalExpression", Line: 1, Column: 9, EndLine: 1, EndColumn: 27},
				},
			},

			// ---- Paren-wrapped tests — verify TrimmedNodeText preserves them ----
			// `(foo + bar)` test: precedence still triggers wrap, so the
			// fixer emits `!((foo + bar))` (matches ESLint).
			{
				Code:   `var a = (foo + bar) ? false : true;`,
				Output: []string{`var a = !((foo + bar));`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalExpression", Line: 1, Column: 9, EndLine: 1, EndColumn: 35},
				},
			},
			// `(typeof x === 'string')` test — paren-wrapped boolean
			// expression: unwrap path keeps the parens.
			{
				Code:   `var a = (typeof x === 'string') ? true : false;`,
				Output: []string{`var a = (typeof x === 'string');`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalExpression", Line: 1, Column: 9, EndLine: 1, EndColumn: 47},
				},
			},

			// ---- Compact-spacing operator inversion ----
			// No spaces around `===` — inversion preserves the exact spacing.
			{
				Code:   `var a = x===1?false:true;`,
				Output: []string{`var a = x!==1;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalExpression", Line: 1, Column: 9, EndLine: 1, EndColumn: 25},
				},
			},
			// Wide spacing around `===` — same preservation rule.
			{
				Code:   `var a = x  ===   1 ? false : true;`,
				Output: []string{`var a = x  !==   1;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalExpression", Line: 1, Column: 9, EndLine: 1, EndColumn: 34},
				},
			},

			// ---- Multi-byte (CJK) identifier as test/consequent ----
			{
				Code:   `var a = 中 === 国 ? true : false;`,
				Output: []string{`var a = 中 === 国;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalExpression", Line: 1, Column: 9, EndLine: 1, EndColumn: 31},
				},
			},

			// ---- TypeScript: SatisfiesExpression ----
			// SatisfiesExpression is unknown to ESLint → -1 precedence → wrapped
			// defensively, mirroring the AsExpression behavior.
			{
				Code:   `var a = x satisfies number ? false : true;`,
				Output: []string{`var a = !(x satisfies number);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalExpression", Line: 1, Column: 9, EndLine: 1, EndColumn: 42},
				},
			},
			{
				Code:    `foo ? foo : bar satisfies number`,
				Output:  []string{`foo || (bar satisfies number)`},
				Options: map[string]interface{}{"defaultAssignment": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalAssignment", Line: 1, Column: 1, EndLine: 1, EndColumn: 33},
				},
			},

			// ---- Nested default-assignment: rule reports both, fixer
			// resolves overlap by re-running. Expect two iterations.
			{
				Code: `foo ? foo : (foo ? foo : c)`,
				Output: []string{
					`foo || (foo ? foo : c)`,
					`foo || (foo || c)`,
				},
				Options: map[string]interface{}{"defaultAssignment": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalAssignment", Line: 1, Column: 1, EndLine: 1, EndColumn: 28},
					{MessageId: "unnecessaryConditionalAssignment", Line: 1, Column: 14, EndLine: 1, EndColumn: 27},
				},
			},

			// ---- defaultAssignment: true explicit + no report ----
			// Already in valid suite. The default is true, so this is just a
			// regression guard for the option-parsing code path.

			// ---- Test that's a Boolean() call: NOT a boolean expression
			// because we don't recognize Boolean() as one (matches ESLint). ----
			{
				Code:   `var a = Boolean(x) ? true : false;`,
				Output: []string{`var a = !!Boolean(x);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalExpression", Line: 1, Column: 9, EndLine: 1, EndColumn: 34},
				},
			},

			// ---- A test that is a comparison wrapped in parens, both arms
			// equal (alternate fix path: emit literal text only when test is
			// an Identifier — here it's NOT, so no fix). ----
			{
				Code: `var a = x === 1 ? true : true;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalExpression", Line: 1, Column: 9, EndLine: 1, EndColumn: 30},
				},
			},

			// ---- Right-associative conditional as alternate — `?:` is
			// right-associative so `b ? b : c ? d : e` is `b ? b : (c?d:e)`.
			// The inner is NOT parenthesised → wrap on output.
			{
				Code:    `b ? b : c ? d : e;`,
				Output:  []string{`b || (c ? d : e);`},
				Options: map[string]interface{}{"defaultAssignment": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalAssignment", Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
				},
			},

			// ---- Function/class expressions as alternate ----
			{
				Code:    `b ? b : function() { return 1; };`,
				Output:  []string{`b || function() { return 1; };`},
				Options: map[string]interface{}{"defaultAssignment": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalAssignment", Line: 1, Column: 1, EndLine: 1, EndColumn: 33},
				},
			},
			{
				Code:    `b ? b : class { static foo() {} };`,
				Output:  []string{`b || class { static foo() {} };`},
				Options: map[string]interface{}{"defaultAssignment": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalAssignment", Line: 1, Column: 1, EndLine: 1, EndColumn: 34},
				},
			},

			// ---- `new Foo()` and template literal tests ----
			{
				Code:   `var a = new Foo() ? true : false;`,
				Output: []string{`var a = !!new Foo();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalExpression", Line: 1, Column: 9, EndLine: 1, EndColumn: 33},
				},
			},
			{
				Code:   "var a = `abc` ? true : false;",
				Output: []string{"var a = !!`abc`;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalExpression", Line: 1, Column: 9, EndLine: 1, EndColumn: 29},
				},
			},
			// Property access (computed and dotted) as test.
			{
				Code:   `var a = obj["foo"] ? true : false;`,
				Output: []string{`var a = !!obj["foo"];`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalExpression", Line: 1, Column: 9, EndLine: 1, EndColumn: 34},
				},
			},
			{
				Code:   `var a = obj.foo ? true : false;`,
				Output: []string{`var a = !!obj.foo;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalExpression", Line: 1, Column: 9, EndLine: 1, EndColumn: 31},
				},
			},

			// ---- Optional chain as test ----
			{
				Code:   `var a = foo?.bar ? true : false;`,
				Output: []string{`var a = !!foo?.bar;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryConditionalExpression", Line: 1, Column: 9, EndLine: 1, EndColumn: 32},
				},
			},
		},
	)
}
