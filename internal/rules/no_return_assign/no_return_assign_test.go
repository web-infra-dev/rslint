package no_return_assign

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoReturnAssignRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoReturnAssignRule,

		[]rule_tester.ValidTestCase{
			// ---- Upstream ESLint suite ----
			{Code: `module.exports = {'a': 1};`},
			{Code: `var result = a * b;`},
			{Code: `function x() { var result = a * b; return result; }`},
			{Code: `function x() { return (result = a * b); }`},
			{Code: `function x() { var result = a * b; return result; }`, Options: "except-parens"},
			{Code: `function x() { return (result = a * b); }`, Options: "except-parens"},
			{Code: `function x() { var result = a * b; return result; }`, Options: "always"},
			{Code: `function x() { return function y() { result = a * b }; }`, Options: "always"},
			{Code: `() => { return (result = a * b); }`, Options: "except-parens"},
			{Code: `() => (result = a * b)`, Options: "except-parens"},
			{Code: `const foo = (a,b,c) => ((a = b), c)`},
			{Code: `function foo(){
            return (a = b)
        }`},
			{Code: `function bar(){
            return function foo(){
                return (a = b) && c
            }
        }`},
			{Code: `const foo = (a) => (b) => (a = b)`},

			// ---- Non-assignment comparisons / declarations ----
			{Code: `function x() { return a == b; }`},
			{Code: `function x() { return a === b; }`},
			{Code: `function x() { var a = b; return a; }`},

			// ---- except-parens: nested parens directly wrapping the assignment ----
			{Code: `function x() { return ((a = b)); }`},
			{Code: `function x() { return (((a = b))); }`},
			{Code: `() => ((a = b))`},

			// ---- except-parens: assignment wrapped in parens inside a larger expression ----
			{Code: `function x() { return (a = b) && c; }`},
			{Code: `function x() { return c && (a = b); }`},
			{Code: `function x() { return (a = b) || c; }`},
			{Code: `function x() { return c || (a = b); }`},
			{Code: `function x() { return (a = b) ?? c; }`},
			{Code: `function x() { return (a = b) ? c : d; }`},
			{Code: `function x() { return c ? (a = b) : d; }`},
			{Code: `function x() { return c ? d : (a = b); }`},
			{Code: `function x() { return ((a = b), c); }`},
			{Code: `const foo = () => ((a = b), c)`},
			{Code: `function x() { return !(a = b); }`},
			{Code: `function x() { return typeof (a = b); }`},

			// ---- Sentinel blocks the walk-up (FunctionExpression / ClassExpression / nested arrow block) ----
			{Code: `function x() { return function y() { a = b; }; }`},
			{Code: `function x() { return function y() { a = b; }; }`, Options: "always"},
			{Code: `function x() { return class { m() { a = b; } }; }`},
			{Code: `function x() { return class { m() { a = b; } }; }`, Options: "always"},
			{Code: `function x() { return () => { a = b; }; }`},
			{Code: `function x() { return () => { a = b; }; }`, Options: "always"},
			{Code: `function x() { return class { static { a = b; } }; }`, Options: "always"},

			// ---- Assignment outside any return/arrow-body ----
			{Code: `a = b;`},
			{Code: `function f() { a = b; }`},
			{Code: `if (x) { a = b; }`},
			{Code: `while (x) { a = b; }`},
			{Code: `for (let i = 0; i < 10; i++) { a = b; }`},
			{Code: `function f() { for (a = 0; a < 10; a++) {} return a; }`},
			{Code: `switch (x) { case 1: a = b; break; }`},
			{Code: `try { a = b; } catch (e) {}`},
			{Code: `class C { m() { a = b; } }`},
			{Code: `function f(a = (b = 1)) { return a; }`},

			// ---- TypeScript: parens still exempt under type assertions / satisfies ----
			{Code: `function f(): number { return (a = b); }`},
			{Code: `function f(): number { return (a = b) as number; }`},
			{Code: `function f(): number { return (a = b) satisfies number; }`},
			{Code: `const f = (a: number): number => (a = 1)`},

			// ---- JSX: parenthesised assignment inside JSX is exempt under except-parens ----
			{Code: `const F = () => <div/>`, Tsx: true},
			{Code: `const F = () => <>{(a = b)}</>`, Tsx: true},
			{Code: `const F = () => <div id={(a = b)}/>`, Tsx: true},
			{Code: `function F() { return <div id={(a = b)}/>; }`, Tsx: true},
			// Assignment inside a nested function swallowed by FunctionExpression sentinel
			{Code: `const F = () => <div onClick={function(){ a = b; }}/>`, Tsx: true},
			// Assignment inside an arrow with block body is swallowed at ExpressionStatement
			{Code: `const F = () => <div onClick={() => { a = b; }}/>`, Tsx: true},
		},

		[]rule_tester.InvalidTestCase{
			// ---- Upstream ESLint suite ----
			{
				Code: `function x() { return result = a * b; };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnAssignment", Message: "Return statement should not contain assignment.", Line: 1, Column: 16},
				},
			},
			{
				Code: `function x() { return (result) = (a * b); };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnAssignment", Line: 1, Column: 16},
				},
			},
			{
				Code:    `function x() { return result = a * b; };`,
				Options: "except-parens",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnAssignment", Line: 1, Column: 16},
				},
			},
			{
				Code:    `function x() { return (result) = (a * b); };`,
				Options: "except-parens",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnAssignment", Line: 1, Column: 16},
				},
			},
			{
				Code: `() => { return result = a * b; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnAssignment", Line: 1, Column: 9},
				},
			},
			{
				Code: `() => result = a * b`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "arrowAssignment", Message: "Arrow function should not return assignment.", Line: 1, Column: 1},
				},
			},
			{
				Code:    `function x() { return result = a * b; };`,
				Options: "always",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnAssignment", Line: 1, Column: 16},
				},
			},
			{
				Code:    `function x() { return (result = a * b); };`,
				Options: "always",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnAssignment", Line: 1, Column: 16},
				},
			},
			{
				Code:    `function x() { return result || (result = a * b); };`,
				Options: "always",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnAssignment", Line: 1, Column: 16},
				},
			},
			{
				Code: `function foo(){
                return a = b
            }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnAssignment", Line: 2, Column: 17},
				},
			},
			{
				Code: `function doSomething() {
                return foo = bar && foo > 0;
            }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnAssignment", Line: 2, Column: 17},
				},
			},
			{
				Code: `function doSomething() {
                return foo = function(){
                    return (bar = bar1)
                }
            }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnAssignment", Line: 2, Column: 17},
				},
			},
			{
				Code: `function doSomething() {
                return foo = () => a
            }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnAssignment", Line: 2, Column: 17},
				},
			},
			{
				Code: `function doSomething() {
                return () => a = () => b
            }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "arrowAssignment", Line: 2, Column: 24},
				},
			},
			{
				Code: `function foo(a){
                return function bar(b){
                    return a = b
                }
            }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnAssignment", Line: 3, Column: 21},
				},
			},
			{
				Code: `const foo = (a) => (b) => a = b`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "arrowAssignment", Line: 1, Column: 20},
				},
			},

			// ---- Assignment operator coverage (returnAssignment) ----
			{Code: `function x() { return a -= b; }`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "returnAssignment", Line: 1, Column: 16}}},
			{Code: `function x() { return a *= b; }`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "returnAssignment", Line: 1, Column: 16}}},
			{Code: `function x() { return a /= b; }`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "returnAssignment", Line: 1, Column: 16}}},
			{Code: `function x() { return a %= b; }`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "returnAssignment", Line: 1, Column: 16}}},
			{Code: `function x() { return a **= b; }`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "returnAssignment", Line: 1, Column: 16}}},
			{Code: `function x() { return a <<= b; }`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "returnAssignment", Line: 1, Column: 16}}},
			{Code: `function x() { return a >>= b; }`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "returnAssignment", Line: 1, Column: 16}}},
			{Code: `function x() { return a >>>= b; }`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "returnAssignment", Line: 1, Column: 16}}},
			{Code: `function x() { return a &= b; }`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "returnAssignment", Line: 1, Column: 16}}},
			{Code: `function x() { return a |= b; }`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "returnAssignment", Line: 1, Column: 16}}},
			{Code: `function x() { return a ^= b; }`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "returnAssignment", Line: 1, Column: 16}}},
			{Code: `function x() { return a &&= b; }`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "returnAssignment", Line: 1, Column: 16}}},
			{Code: `function x() { return a ||= b; }`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "returnAssignment", Line: 1, Column: 16}}},
			{Code: `function x() { return a ??= b; }`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "returnAssignment", Line: 1, Column: 16}}},

			// ---- Assignment operator coverage (arrowAssignment) ----
			{Code: `() => a += b`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "arrowAssignment", Line: 1, Column: 1}}},
			{Code: `() => a ??= b`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "arrowAssignment", Line: 1, Column: 1}}},
			{Code: `(a, b) => a = b`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "arrowAssignment", Line: 1, Column: 1}}},

			// ---- Container coverage (returnAssignment) ----
			{Code: `async function f() { return a = b; }`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "returnAssignment", Line: 1, Column: 22}}},
			{Code: `function* g() { return a = b; }`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "returnAssignment", Line: 1, Column: 17}}},
			{Code: `async function* ag() { return a = b; }`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "returnAssignment", Line: 1, Column: 24}}},
			{Code: `class C { m() { return a = b; } }`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "returnAssignment", Line: 1, Column: 17}}},
			{Code: `class C { get x() { return a = b; } }`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "returnAssignment", Line: 1, Column: 21}}},
			{Code: `class C { set x(v) { return a = b; } }`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "returnAssignment", Line: 1, Column: 22}}},
			{Code: `({ m() { return a = b; } })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "returnAssignment", Line: 1, Column: 10}}},

			// ---- Container coverage (arrowAssignment) ----
			{Code: `async () => a = b`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "arrowAssignment", Line: 1, Column: 1}}},

			// ---- Walk-up wrappers (non-paren expressions) — still reports ----
			{Code: `function x() { return a = b, c; }`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "returnAssignment", Line: 1, Column: 16}}},
			{Code: `function x() { return a ? b = c : d; }`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "returnAssignment", Line: 1, Column: 16}}},
			{Code: `function x() { return { x: a = b }; }`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "returnAssignment", Line: 1, Column: 16}}},
			{Code: `function x() { return [a = b]; }`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "returnAssignment", Line: 1, Column: 16}}},
			{Code: `function x() { return a = b && c; }`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "returnAssignment", Line: 1, Column: 16}}},
			// UnaryExpression / TypeOf with a parenthesised assignment is VALID under except-parens
			// but INVALID under always (see valid cases above and always section below).
			{Code: `function x() { return !(a = b); }`, Options: "always", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "returnAssignment", Line: 1, Column: 16}}},
			{Code: `function x() { return typeof (a = b); }`, Options: "always", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "returnAssignment", Line: 1, Column: 16}}},

			// ---- always mode: parens don't exempt ----
			{Code: `function x() { return ((a = b)); }`, Options: "always", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "returnAssignment", Line: 1, Column: 16}}},
			{Code: `() => ((a = b))`, Options: "always", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "arrowAssignment", Line: 1, Column: 1}}},
			{Code: `function x() { return c || (a = b); }`, Options: "always", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "returnAssignment", Line: 1, Column: 16}}},
			{Code: `function x() { return a + (b = c); }`, Options: "always", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "returnAssignment", Line: 1, Column: 16}}},

			// ---- TypeScript syntax ----
			{Code: `function f(): number { return a = b; }`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "returnAssignment", Line: 1, Column: 24}}},
			{Code: `const f = (): number => a = b`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "arrowAssignment", Line: 1, Column: 11}}},
			{Code: `function f(): number { return a = b as number; }`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "returnAssignment", Line: 1, Column: 24}}},

			// ---- Inner container reports; outer unaffected (only one diagnostic) ----
			{
				Code: `function f() { return (function g() { return a = b; })(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnAssignment", Line: 1, Column: 39},
				},
			},
			// Outer `return a = (b = c)`: outer direct in return; inner wrapped in parens.
			// Only the outer should fire under except-parens.
			{
				Code: `function f() { return a = (b = c); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnAssignment", Line: 1, Column: 16},
				},
			},
			// With always, both fire.
			{
				Code:    `function f() { return a = (b = c); }`,
				Options: "always",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnAssignment", Line: 1, Column: 16},
					{MessageId: "returnAssignment", Line: 1, Column: 16},
				},
			},

			// ---- JSX: assignment flowing up through JSX nodes to arrow / return ----
			// JSX nodes (JsxElement / JsxExpression / JsxAttributes) are not sentinels,
			// so the walk continues until the enclosing arrow / return.
			{
				Code:   `const F = () => <div id={a = b}/>`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "arrowAssignment", Line: 1, Column: 11}},
			},
			{
				Code:   `function F() { return <div id={a = b}/>; }`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "returnAssignment", Line: 1, Column: 16}},
			},
			{
				Code:   `const F = () => <>{a = b}</>`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "arrowAssignment", Line: 1, Column: 11}},
			},
			{
				Code:   `function F() { return <>{a = b}</>; }`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "returnAssignment", Line: 1, Column: 16}},
			},
			// Nested arrow inside JSX attribute whose body IS an assignment
			{
				Code:   `const F = () => <div onClick={() => a = b}/>`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "arrowAssignment", Line: 1, Column: 31}},
			},
			// always mode: parenthesised assignment inside JSX is no longer exempt
			{
				Code:    `const F = () => <div id={(a = b)}/>`,
				Tsx:     true,
				Options: "always",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "arrowAssignment", Line: 1, Column: 11}},
			},

			// ---- Multi-line position assertions (Line + Column + EndLine + EndColumn) ----
			{
				Code: "function x() {\n  return a\n    = b;\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnAssignment", Line: 2, Column: 3, EndLine: 3, EndColumn: 9},
				},
			},
			{
				Code: "() =>\n  a = b",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "arrowAssignment", Line: 1, Column: 1, EndLine: 2, EndColumn: 8},
				},
			},
			// ---- Multi-byte character position assertions (UTF-16 code unit counting) ----
			// BMP-outside emoji (surrogate pair) in a string literal: 🍌 counts as
			// 2 UTF-16 units. Catches implementations that would use rune / code-point
			// counts instead. (Emojis are not valid identifier starts in JS.)
			{
				Code: "function x() {\n  return a\n    = \"🍌\";\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnAssignment", Line: 2, Column: 3, EndLine: 3, EndColumn: 12},
				},
			},
			// BMP CJK characters as identifiers: 中 / 文 are each 1 UTF-16 unit but
			// 3 bytes in UTF-8. Catches implementations that would use byte offsets.
			{
				Code: "function x() {\n  return 中\n    = 文;\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnAssignment", Line: 2, Column: 3, EndLine: 3, EndColumn: 9},
				},
			},
			// Emoji inside an arrow-body assignment (arrowAssignment path).
			{
				Code: "() =>\n  a = \"🍎\"",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "arrowAssignment", Line: 1, Column: 1, EndLine: 2, EndColumn: 11},
				},
			},

			// ---- Extra: compound assignment operators (legacy kept) ----
			{
				Code: `function x() { return foo += 2; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnAssignment", Line: 1, Column: 16},
				},
			},
			{
				Code: `function x() { return foo &&= bar; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnAssignment", Line: 1, Column: 16},
				},
			},
			{
				Code: `const foo = (a, b) => a = b`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "arrowAssignment", Line: 1, Column: 13},
				},
			},
		},
	)
}
