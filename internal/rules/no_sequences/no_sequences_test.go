package no_sequences

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoSequencesRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoSequencesRule,

		[]rule_tester.ValidTestCase{
			// ---- Upstream ESLint suite ----
			{Code: `var arr = [1, 2];`},
			{Code: `var obj = {a: 1, b: 2};`},
			{Code: `var a = 1, b = 2;`},
			{Code: `var foo = (1, 2);`},
			{Code: `(0,eval)("foo()");`},
			{Code: `for (i = 1, j = 2;; i++, j++);`},
			{Code: `foo(a, (b, c), d);`},
			{Code: `do {} while ((doSomething(), !!test));`},
			{Code: `for ((doSomething(), somethingElse()); (doSomething(), !!test); );`},
			{Code: `if ((doSomething(), !!test));`},
			{Code: `switch ((doSomething(), val)) {}`},
			{Code: `while ((doSomething(), !!test));`},
			{Code: `with ((doSomething(), val)) {}`},
			{Code: `a => ((doSomething(), a))`},

			// Options object without the "allowInParentheses" property
			{Code: `var foo = (1, 2);`, Options: map[string]interface{}{}},

			// Explicitly set option "allowInParentheses" to default value
			{Code: `var foo = (1, 2);`, Options: map[string]interface{}{"allowInParentheses": true}},

			// allowInParentheses: false — for-init / for-update are always allowed
			{Code: `for ((i = 0, j = 0); test; );`, Options: map[string]interface{}{"allowInParentheses": false}},
			{Code: `for (; test; (i++, j++));`, Options: map[string]interface{}{"allowInParentheses": false}},

			// https://github.com/eslint/eslint/issues/14572 — return of a parenthesised sequence
			{Code: `const foo = () => { return ((bar = 123), 10) }`},
			{Code: `const foo = () => (((bar = 123), 10));`},

			// ---- Extra: chain element that is itself a (parenthesised) sequence
			// stays valid — only the outer chain is considered, and it's parenthesised.
			{Code: `var foo = ((1, 2), 3);`},

			// ---- Extra containers: none appear in the grammar-paren list, so a
			// single pair of parens around the sequence is enough to exempt.
			// for-in RHS (Expression allows comma), for-of RHS (AssignmentExpression
			// — here the paren wrap makes the sequence a PrimaryExpression).
			{Code: `for (x in (a, b)) {}`},
			{Code: `for (x of (a, b)) {}`},

			// throw / return with parenthesised sequence
			{Code: `function f() { throw (a, b); }`},
			{Code: `function f() { return (a, b); }`},

			// TypeScript: `as` / `satisfies` bind tighter than `,`, so the
			// sequence must be wrapped for the AssignmentExpression position.
			{Code: `const x = (a, b) as number;`},
			{Code: `const x = (a, b) satisfies number;`},
			{Code: `const x = ((a, b)) as number;`},

			// Template literal / tagged template substitution
			{Code: "const x = `${(a, b)}`;"},
			{Code: "const x = tag`${(a, b)}`;"},

			// Function / arrow parameter default value
			{Code: `function f(x = (a, b)) {}`},
			{Code: `const f = (x = (a, b)) => x;`},

			// Optional-chain / element access — parenthesised sequence is fine
			{Code: `foo?.((a, b));`},
			{Code: `foo?.[(a, b)];`},
			{Code: `foo[(a, b)];`},

			// Class field initializer
			{Code: `class C { x = (a, b); }`},
			{Code: `class C { static x = (a, b); }`},

			// Object computed key
			{Code: `const obj = { [(a, b)]: 1 };`},

			// Conditional expression slots — each slot is AssignmentExpression,
			// so a bare sequence there requires parens to parse anyway.
			{Code: `const x = (a, b) ? c : d;`},
			{Code: `const x = a ? (b, c) : d;`},
			{Code: `const x = a ? b : (c, d);`},
		},

		[]rule_tester.InvalidTestCase{
			// ---- Upstream ESLint suite ----
			{
				Code: `1, 2;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCommaExpression", Message: "Unexpected use of comma operator.", Line: 1, Column: 2, EndLine: 1, EndColumn: 3},
				},
			},
			{
				Code: `a = 1, 2`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCommaExpression", Line: 1, Column: 6},
				},
			},
			{
				Code: `do {} while (doSomething(), !!test);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCommaExpression", Line: 1, Column: 27},
				},
			},
			{
				Code: `for (; doSomething(), !!test; );`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCommaExpression", Line: 1, Column: 21},
				},
			},
			{
				Code: `if (doSomething(), !!test);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCommaExpression", Line: 1, Column: 18},
				},
			},
			{
				Code: `switch (doSomething(), val) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCommaExpression", Line: 1, Column: 22},
				},
			},
			{
				Code: `while (doSomething(), !!test);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCommaExpression", Line: 1, Column: 21},
				},
			},
			{
				Code: `with (doSomething(), val) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCommaExpression", Line: 1, Column: 20},
				},
			},
			{
				Code: `a => (doSomething(), a)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCommaExpression", Line: 1, Column: 20},
				},
			},
			{
				Code: `(1), 2`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCommaExpression", Line: 1, Column: 4},
				},
			},
			{
				Code: `((1)) , (2)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCommaExpression", Line: 1, Column: 7},
				},
			},
			{
				Code: `while((1) , 2);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCommaExpression", Line: 1, Column: 11},
				},
			},

			// ---- allowInParentheses: false — sequences are flagged even inside parens
			{
				Code:    `var foo = (1, 2);`,
				Options: map[string]interface{}{"allowInParentheses": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCommaExpression", Line: 1, Column: 13},
				},
			},
			{
				Code:    `(0,eval)("foo()");`,
				Options: map[string]interface{}{"allowInParentheses": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCommaExpression", Line: 1, Column: 3},
				},
			},
			{
				Code:    `foo(a, (b, c), d);`,
				Options: map[string]interface{}{"allowInParentheses": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCommaExpression", Line: 1, Column: 10},
				},
			},
			{
				Code:    `do {} while ((doSomething(), !!test));`,
				Options: map[string]interface{}{"allowInParentheses": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCommaExpression", Line: 1, Column: 28},
				},
			},
			{
				Code:    `for (; (doSomething(), !!test); );`,
				Options: map[string]interface{}{"allowInParentheses": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCommaExpression", Line: 1, Column: 22},
				},
			},
			{
				Code:    `if ((doSomething(), !!test));`,
				Options: map[string]interface{}{"allowInParentheses": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCommaExpression", Line: 1, Column: 19},
				},
			},
			{
				Code:    `switch ((doSomething(), val)) {}`,
				Options: map[string]interface{}{"allowInParentheses": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCommaExpression", Line: 1, Column: 23},
				},
			},
			{
				Code:    `while ((doSomething(), !!test));`,
				Options: map[string]interface{}{"allowInParentheses": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCommaExpression", Line: 1, Column: 22},
				},
			},
			{
				Code:    `with ((doSomething(), val)) {}`,
				Options: map[string]interface{}{"allowInParentheses": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCommaExpression", Line: 1, Column: 21},
				},
			},
			{
				Code:    `a => ((doSomething(), a))`,
				Options: map[string]interface{}{"allowInParentheses": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCommaExpression", Line: 1, Column: 21},
				},
			},

			// ---- Extra edge cases ----

			// Multi-line — `Line` / `EndLine` / `Column` / `EndColumn` assertion on the comma token
			{
				Code: "foo(\n  a,\n  b\n),\n2;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCommaExpression", Line: 4, Column: 2, EndLine: 4, EndColumn: 3},
				},
			},

			// Chain of 3: should report once, at the FIRST comma (leftmost).
			{
				Code: `a, b, c;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCommaExpression", Line: 1, Column: 2, EndLine: 1, EndColumn: 3},
				},
			},

			// ---- for-in / for-of RHS — Expression / AssignmentExpression slot
			// respectively. Neither is in ESLint's grammar-paren list; a bare
			// sequence there must be flagged (for-of RHS even requires parens
			// to parse, so the bare form is only reachable via for-in).
			{
				Code: `for (x in a, b) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCommaExpression", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
				},
			},
			// allowInParentheses:false still reports when wrapped
			{
				Code:    `for (x in (a, b)) {}`,
				Options: map[string]interface{}{"allowInParentheses": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCommaExpression", Line: 1, Column: 13},
				},
			},
			{
				Code:    `for (x of (a, b)) {}`,
				Options: map[string]interface{}{"allowInParentheses": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCommaExpression", Line: 1, Column: 13},
				},
			},

			// ---- throw / return — ThrowStatement / ReturnStatement are not
			// in ESLint's grammar-paren list.
			{
				Code: `function f() { throw a, b; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCommaExpression", Line: 1, Column: 23},
				},
			},
			{
				Code: `function f() { return a, b; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCommaExpression", Line: 1, Column: 24},
				},
			},
			{
				Code: `() => { throw a, b; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCommaExpression", Line: 1, Column: 16},
				},
			},

			// ---- TypeScript: `as` / `satisfies` bind tighter than comma, so
			// `a, b as T` is `a, (b as T)` — outer sequence at bare
			// ExpressionStatement scope, must be flagged.
			{
				Code: `a, b as number;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCommaExpression", Line: 1, Column: 2},
				},
			},
			{
				Code: `a, b satisfies number;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCommaExpression", Line: 1, Column: 2},
				},
			},

			// ---- Template literal / tagged template substitution
			{
				Code: "`${a, b}`;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCommaExpression", Line: 1, Column: 5},
				},
			},
			{
				Code: "tag`${a, b}`;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCommaExpression", Line: 1, Column: 8},
				},
			},

			// ---- Element access with bracket-separated sequence
			// Brackets (`[...]`) are NOT parens, so a bare sequence inside
			// computed-access brackets must be flagged.
			{
				Code: `foo[a, b];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCommaExpression", Line: 1, Column: 6},
				},
			},
			{
				Code: `foo?.[a, b];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCommaExpression", Line: 1, Column: 8},
				},
			},

			// ---- JSX: JsxExpression slot — not in grammar-paren list.
			{
				Code:   `const x = <div>{a, b}</div>;`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedCommaExpression", Line: 1, Column: 18}},
			},
			{
				Code:   `const x = <div id={a, b}/>;`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedCommaExpression", Line: 1, Column: 21}},
			},
			// allowInParentheses:false — even parenthesised JSX expression
			// slot should be flagged.
			{
				Code:    `const x = <div id={(a, b)}/>;`,
				Tsx:     true,
				Options: map[string]interface{}{"allowInParentheses": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedCommaExpression", Line: 1, Column: 22}},
			},

			// ---- Conditional expression slots: bare sequence where a slot
			// expects an AssignmentExpression is a parse error, but a comma
			// at the ConditionalExpression's outer boundary is still flagged.
			// Example: `a ? b : c, d` parses as `(a ? b : c), d`.
			{
				Code: `a ? b : c, d;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCommaExpression", Line: 1, Column: 10},
				},
			},

			// ---- Multi-byte position assertions (UTF-16 code units).
			// Surrogate pair in a string literal — each emoji counts as 2
			// UTF-16 units. Catches byte-offset / code-point counting bugs.
			{
				Code: `'🍎', 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCommaExpression", Line: 1, Column: 5, EndLine: 1, EndColumn: 6},
				},
			},
			// BMP CJK identifiers — each is 1 UTF-16 unit but 3 UTF-8 bytes.
			{
				Code: `中, 文;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCommaExpression", Line: 1, Column: 2, EndLine: 1, EndColumn: 3},
				},
			},
			// Multi-line with emoji on a prior column
			{
				Code: "function f() {\n  return \"🍎\", 1;\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCommaExpression", Line: 2, Column: 14, EndLine: 2, EndColumn: 15},
				},
			},

			// ---- Longer chain (4 elements) — still reports once, at the
			// leftmost comma.
			{
				Code: `a, b, c, d;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCommaExpression", Line: 1, Column: 2, EndLine: 1, EndColumn: 3},
				},
			},
		},
	)
}
