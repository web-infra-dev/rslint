package no_const_assign

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoConstAssignRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoConstAssignRule,
		// Valid cases - ported from ESLint
		[]rule_tester.ValidTestCase{
			// Reassignment in different scope
			{Code: `const x = 0; { let x; x = 1; }`},

			// Parameter shadows constant
			{Code: `const x = 0; function a(x) { x = 1; }`},

			// Reading constant value
			{Code: `const x = 0; foo(x);`},

			// For-in loop - x is redeclared on each iteration
			{Code: `for (const x in [1,2,3]) { foo(x); }`},

			// For-of loop - x is redeclared on each iteration
			{Code: `for (const x of [1,2,3]) { foo(x); }`},

			// Explicit resource management bindings are constant but may be read.
			{Code: `function f() { using x = foo(); bar(x); }`},
			{Code: `async function f() { await using x = foo(); bar(x); }`},

			// Modifying property, not reassigning the constant
			{Code: `const x = {key: 0}; x.key = 1;`},

			// var can be reassigned
			{Code: `var x = 0; x = 1;`},

			// let can be reassigned
			{Code: `let x = 0; x = 1;`},

			// Keyword filter false positives must not change symbol semantics.
			{Code: `let constable = 0; constable = 1;`},

			// The DOM global `name` is declared with const in lib.dom.d.ts, but
			// this rule only owns constant declarations in the linted file.
			{Code: `const local = 0; name = "value";`},

			// A later mutable declaration shadows the outer constant for the
			// entire block, including its temporal dead zone.
			{Code: `const x = 0; { x = 1; let x; }`},

			// Function declaration can be reassigned
			{Code: `function x() {} x = 1;`},

			// Catch parameter can be reassigned
			{Code: `try {} catch (x) { x = 1; }`},

			// Const in initializer is valid
			{Code: `const x = x;`},

			// Multiple const declarations
			{Code: `const x = 0, y = 1;`},

			// Const with destructuring
			{Code: `const {a} = {a: 0};`},

			// Const with array destructuring
			{Code: `const [a] = [0];`},

			// Const object with method calls
			{Code: `const x = {}; x.method();`},

			// A using binding's object may be mutated without reassigning it.
			{Code: `function f() { using x = foo(); x.value = 1; }`},

			// Reading const in conditional
			{Code: `const x = 0; if (x === 0) {}`},

			// Const in arrow function parameter
			{Code: `const x = 0; const f = (y = x) => y;`},

			// Scope shadowing with destructuring - different variable
			{Code: `const {x} = {x: 0}; { let x; x = 1; }`},

			// Scope shadowing with destructuring in block
			{Code: `const {x} = {x: 0}; { const x = 1; }`},

			// Try-catch with const - different scopes
			{Code: `try { const x = 1; } catch (e) { const x = 2; }`},

			// Catch variable shadows const
			{Code: `const e = 0; try {} catch (e) { e = 1; }`},

			// Function parameter shadows const in nested function
			{Code: `const x = 0; function outer() { function inner(x) { x = 1; } }`},

			// Arrow function parameter shadows const
			{Code: `const x = 0; const f = (x) => { x = 1; };`},

			// Block scope shadows const
			{Code: `const x = 0; { const x = 1; }`},

			// Nested block scopes
			{Code: `const x = 0; { { let x; x = 1; } }`},

			// Const read inside a nested object literal used as a
			// destructuring default value — the default value is a plain
			// expression, not a pattern, so the shorthand read must not be
			// mistaken for a write to the destructuring target.
			{Code: `const x = 0; let y; ({a: y = {x}} = {});`},

			// Same case with array destructuring default value
			{Code: `const x = 0; let y; ([y = [x]] = []);`},

			// Const read inside a shorthand-property default value
			{Code: `const x = 0; let y; ({y = {x}} = {});`},

			// Const read as a plain (non-object) default value
			{Code: `const x = 0; let y; ({y = x} = {});`},

			// Const read inside a nested object literal wrapped in a call
			// expression that's used as a destructuring default value
			{Code: `const x = 0; let y; ({a: y = foo({x})} = {});`},

			// Const read inside a nested object literal wrapped in a
			// conditional expression used as a destructuring default value
			{Code: `const x = 0; let y; ({a: y = true ? {x} : {}} = {});`},

			// Const read inside a computed property key of a destructuring
			// pattern — the key is evaluated as a value, not a pattern
			// element, so the shorthand read must not be mistaken for a
			// write to the destructuring target.
			{Code: `const x = 0; let y; ({[{x}]: y} = {});`},

			// Same case with the key wrapped in a call expression
			{Code: `const x = 0; let y; ({[foo({x})]: y} = {});`},
		},
		// Invalid cases - ported from ESLint
		[]rule_tester.InvalidTestCase{
			// References are resolved against declarations that occur later in
			// the same scope as well as declarations already visited.
			{
				Code: "x = 1;\nconst x = 0;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "const", Line: 1, Column: 1},
				},
			},

			// Direct reassignment
			{
				Code: `const x = 0; x = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "const", Line: 1, Column: 14},
				},
			},

			// `using` and `await using` declarations are constant bindings too.
			{
				Code: "function f() {\n  using x = foo();\n  x = 1;\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "const", Line: 3, Column: 3},
				},
			},
			{
				Code: "async function f() {\n  await using x = foo();\n  x ||= bar();\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "const", Line: 3, Column: 3},
				},
			},

			// Destructured constant reassignment
			{
				Code: `const {a: x} = {a: 0}; x = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "const", Line: 1, Column: 24},
				},
			},

			// Assignment via destructuring
			{
				Code: `const x = 0; ({x} = {x: 1});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "const", Line: 1, Column: 16},
				},
			},

			// Compound assignment +=
			{
				Code: `const x = 0; x += 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "const", Line: 1, Column: 14},
				},
			},

			// A pre-existing constant is written by a for-of initializer.
			{
				Code: `const x = 0; for (x of values) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "const", Line: 1, Column: 19},
				},
			},

			// TypeScript wrappers do not hide an assignment target.
			{
				Code: `const x = 0; (x as any) = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "const", Line: 1, Column: 15},
				},
			},

			// Defaults inside an assignment pattern still write the target.
			{
				Code: `const x = 0; [x = 1] = values;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "const", Line: 1, Column: 15},
				},
			},

			// Resource bindings remain constant inside destructuring writes.
			{
				Code: "function f() {\n  using x = foo();\n  [x, y] = bar();\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "const", Line: 3, Column: 4},
				},
			},

			// Prefix increment operator
			{
				Code: `const x = 0; ++x;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "const", Line: 1, Column: 16},
				},
			},

			// Postfix increment operator
			{
				Code: `const x = 0; x++;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "const", Line: 1, Column: 14},
				},
			},

			// For loop counter increment
			{
				Code: `for (const i = 0; i < 10; ++i) { foo(i); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "const", Line: 1, Column: 29},
				},
			},

			// Multiple reassignments
			{
				Code: `const x = 0; x = 1; x = 2;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "const", Line: 1, Column: 14},
					{MessageId: "const", Line: 1, Column: 21},
				},
			},

			// Compound assignment -=
			{
				Code: `const x = 0; x -= 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "const", Line: 1, Column: 14},
				},
			},

			// Compound assignment *=
			{
				Code: `const x = 0; x *= 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "const", Line: 1, Column: 14},
				},
			},

			// Compound assignment /=
			{
				Code: `const x = 0; x /= 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "const", Line: 1, Column: 14},
				},
			},

			// Compound assignment %=
			{
				Code: `const x = 0; x %= 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "const", Line: 1, Column: 14},
				},
			},

			// Compound assignment <<=
			{
				Code: `const x = 0; x <<= 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "const", Line: 1, Column: 14},
				},
			},

			// Compound assignment >>=
			{
				Code: `const x = 0; x >>= 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "const", Line: 1, Column: 14},
				},
			},

			// Compound assignment >>>=
			{
				Code: `const x = 0; x >>>= 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "const", Line: 1, Column: 14},
				},
			},

			// Compound assignment &=
			{
				Code: `const x = 0; x &= 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "const", Line: 1, Column: 14},
				},
			},

			// Compound assignment |=
			{
				Code: `const x = 0; x |= 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "const", Line: 1, Column: 14},
				},
			},

			// Compound assignment ^=
			{
				Code: `const x = 0; x ^= 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "const", Line: 1, Column: 14},
				},
			},

			// Prefix decrement operator
			{
				Code: `const x = 0; --x;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "const", Line: 1, Column: 16},
				},
			},

			// Postfix decrement operator
			{
				Code: `const x = 0; x--;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "const", Line: 1, Column: 14},
				},
			},

			// Array destructuring reassignment
			{
				Code: `const [x] = [0]; x = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "const", Line: 1, Column: 18},
				},
			},

			// Nested object destructuring
			{
				Code: `const {a: {b: x}} = {a: {b: 0}}; x = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "const", Line: 1, Column: 34},
				},
			},

			// Assignment in conditional
			{
				Code: `const x = 0; if (true) { x = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "const", Line: 1, Column: 26},
				},
			},

			// Assignment in function
			{
				Code: `const x = 0; function f() { x = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "const", Line: 1, Column: 29},
				},
			},

			// Assignment in arrow function
			{
				Code: `const x = 0; const f = () => { x = 1; };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "const", Line: 1, Column: 32},
				},
			},

			// Nullish coalescing assignment ??=
			{
				Code: `const x = 0; x ??= 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "const", Line: 1, Column: 14},
				},
			},

			// Logical AND assignment &&=
			{
				Code: `const x = 1; x &&= 2;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "const", Line: 1, Column: 14},
				},
			},

			// Logical OR assignment ||=
			{
				Code: `const x = 0; x ||= 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "const", Line: 1, Column: 14},
				},
			},

			// Exponentiation assignment **=
			{
				Code: `const x = 2; x **= 3;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "const", Line: 1, Column: 14},
				},
			},

			// Const in different scopes - outer scope
			{
				Code: `const x = 1; function foo() { const x = 2; x = 3; } x = 4;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "const", Line: 1, Column: 44},
					{MessageId: "const", Line: 1, Column: 53},
				},
			},

			// Try-catch with const reassignment
			{
				Code: `try { const x = 1; x = 2; } catch (e) { const x = 3; x = 4; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "const", Line: 1, Column: 20},
					{MessageId: "const", Line: 1, Column: 54},
				},
			},

			// Nested function reassigns outer const
			{
				Code: `const x = 1; function outer() { function inner() { x = 2; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "const", Line: 1, Column: 52},
				},
			},

			// Reassignment in nested block
			{
				Code: `const x = 1; { { x = 2; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "const", Line: 1, Column: 18},
				},
			},

			// Const is the actual write target behind a default value —
			// distinct from `{a: y = {x}}`, where `x` is only read.
			{
				Code: `const x = 0; ({a: x = 1} = {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "const", Line: 1, Column: 19},
				},
			},
		},
	)
}
