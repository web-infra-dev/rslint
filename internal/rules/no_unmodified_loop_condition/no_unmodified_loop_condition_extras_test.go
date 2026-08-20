// TestNoUnmodifiedLoopConditionExtras locks in branches and edge shapes that
// the upstream test suite doesn't exercise. The full pinned ESLint main suite
// lives in no_unmodified_loop_condition_upstream_test.go.
package no_unmodified_loop_condition

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnmodifiedLoopConditionExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnmodifiedLoopConditionRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// === Basic modifications ===
			{Code: `var foo = 0; while (foo) { ++foo; }`},
			{Code: `var foo = 0; while (foo) { foo += 1; }`},
			{Code: `var foo = 0; while (foo) { foo = bar(); }`},
			{Code: `var foo = 0; while (foo < 10) { foo++; }`},
			{Code: `var foo = 0; while (foo < 10) { --foo; }`},
			{Code: `var foo = 0; while (foo < 10) { foo -= 1; }`},

			// ---- Real-user: decrement/compound-assignment conditions from the supplied report ----
			{Code: `var foo = 0; while (foo++) { }`},
			{Code: `var foo = 0; while (--foo) { }`},
			{Code: `var foo = 0; while (foo = next()) { }`},
			{Code: `var foo = 0; while ((foo -= 1) > 0) { }`},
			{Code: `var foo = 0; do { } while (foo++);`},
			{Code: `for (var foo = 10; foo--; ) { }`},

			// === Do-while ===
			{Code: `var foo = 0; do { foo++; } while (foo < 10)`},
			{Code: `var foo = 0; do { foo = next(); } while (foo)`},

			// === For statement: incrementor ===
			{Code: `for (var i = 0; i < 10; i++) {}`},
			{Code: `for (var i = 0; i < 10; ++i) {}`},
			{Code: `for (var i = 10; i > 0; i--) {}`},
			{Code: `for (var i = 0; i < 10; i += 1) {}`},

			// === For statement: body modification ===
			{Code: `for (var i = 0; i < 10; ) { i++; }`},

			// === Dynamic expressions in condition — skip check ===
			{Code: `while (ok(foo)) { }`},
			{Code: `while (foo.ok) { }`},
			{Code: `while (foo[0]) { }`},
			{Code: `while (new Foo()) { }`},
			{Code: "while (tag`template`) { }"},
			{Code: `while (a.b.c) { }`},
			{Code: `for (var i = 0; f(i) < 10; ) { }`},
			{Code: `var x = "./module.js"; while (import(x)) { x = next(); }`},

			// === Comparison group semantics ===
			// a < b is one group: if a is modified, b is OK too
			{Code: `var a = 0, b = 0; while (a < b) { a++; }`},
			{Code: `var a = 0, b = 0; while (a > b) { a++; }`},
			{Code: `var a = 0, b = 0; while (a <= b) { a++; }`},
			{Code: `var a = 0, b = 0; while (a >= b) { a++; }`},
			{Code: `var a = 0, b = 0; while (a == b) { a++; }`},
			{Code: `var a = 0, b = 0; while (a === b) { a++; }`},
			{Code: `var a = 0, b = 0; while (a != b) { a++; }`},
			{Code: `var a = 0, b = 0; while (a !== b) { a++; }`},

			// === Logical literals / no identifiers ===
			{Code: `while (true) { break; }`},
			{Code: `do { break; } while (true)`},
			{Code: `for (;;) { break; }`},

			// === || groups: both sides modified ===
			{Code: `var a = 0, b = 0; while (a || b) { a++; b++; }`},

			// === Modification via function call ===
			{Code: `var x = 0; function inc() { x++; } while (x < 10) { inc(); }`},
			{Code: `var x = 0; function update() { x = next(); } while (x) { update(); }`},

			// === Function call in for incrementor ===
			{Code: `var x = 0; function inc() { x++; } for (var i = 0; x < 10; inc()) { i++; }`},

			// === Destructuring write ===
			{Code: `var x = 0; while (x < 10) { ({x} = {x: 1}); }`},
			{Code: `var x = 0; while (x < 10) { [x] = [1]; }`},

			// === Arrow/function expression in condition: not "dynamic" ===
			{Code: `var x = 0; while (x || (() => foo())) { x++; }`},

			// === Modification in nested function DOES count (ESLint range-based) ===
			{Code: `var foo = 0; while (foo) { function f() { foo = 1; } }`},
			{Code: `var foo = 0; while (foo) { var f = () => { foo = 1; }; }`},
			{Code: `var foo = 0; while (foo) { var f = function() { foo = 1; }; }`},

			// === Modification in nested non-function block ===
			{Code: `var x = 0; while (x < 10) { if (true) { x++; } }`},
			{Code: `var x = 0; while (x < 10) { { x++; } }`},
			{Code: `var x = 0; while (x < 10) { try { x++; } catch(e) {} }`},

			// === Modification in nested loop body (not function boundary) ===
			{Code: `var x = 0; while (x < 10) { for (var i = 0; i < 1; i++) { x++; } }`},
			{Code: `var x = 0; while (x < 10) { while (false) { x++; } }`},

			// === For-in/for-of as write target inside loop ===
			{Code: `var x = ""; while (x) { for (x in {a: 1}) {} }`},
			{Code: `var x = 0; while (x) { for (x of [1, 2]) {} }`},

			// === ConditionalExpression (ternary) as group ===
			{Code: `var a = 0, b = 0; while (a ? b : 0) { a++; }`},

			// === Complex nesting: (a < b) || c — group a<b OK, c independent ===
			{Code: `var a = 0, b = 10, c = 0; while ((a < b) || c) { a++; c++; }`},

			// === Compound assignment operators ===
			{Code: `var x = 0; while (x < 10) { x *= 2; }`},
			{Code: `var x = 0; while (x < 10) { x /= 2; }`},
			{Code: `var x = 0; while (x < 10) { x %= 2; }`},
			{Code: `var x = 0; while (x < 10) { x **= 2; }`},
			{Code: `var x = 0; while (x < 10) { x <<= 1; }`},
			{Code: `var x = 0; while (x < 10) { x >>= 1; }`},
			{Code: `var x = 0; while (x < 10) { x >>>= 1; }`},
			{Code: `var x = 0; while (x < 10) { x &= 1; }`},
			{Code: `var x = 0; while (x < 10) { x |= 1; }`},
			{Code: `var x = 0; while (x < 10) { x ^= 1; }`},
			{Code: `var x: any = 0; while (x < 10) { x ||= 1; }`},
			{Code: `var x: any = 0; while (x < 10) { x &&= 1; }`},
			{Code: `var x: any = 0; while (x < 10) { x ??= 1; }`},

			// === Ambient/lib global (declared in lib.dom.d.ts, not this file —
			// RefStore's per-file binder walk can't resolve it, only the
			// TypeChecker fallback can) modified in the body ===
			{Code: `while (window) { window = window; }`},

			// === A default-exported named function still resolves to the same
			// local binding when it is called from the loop. ===
			{Code: `export default function inc() { x++; } var x = 0; while (x < 10) { inc(); }`},

			// === Function references and writes matched by ESLint's loop range ===
			{Code: `var x = 0; function f(y = (x = 1)) { } while (x < 10) { f(2); }`},
			{Code: `var x = 0; function update() { x++; } while (x && update()) { }`},
			{Code: `var x = 0; while (x) { var x = 1; }`},
			{Code: `var x = 0; while (x) { for (var x of xs) { } }`},

			// === References inside function expressions are not loop conditions ===
			{Code: `var x = 0, guard = 1; while (x || (() => guard)) { x++; }`},
			{Code: `var x = 0, guard = 1; while (x || function () { return guard; }) { x++; }`},
			{Code: `var x = 0, guard = 1; while (x || class { field = guard }) { x++; }`},
			{Code: `var x = 0, guard = 1; while (x || ({ method() { return guard; } })) { x++; }`},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// === Basic: simple identifier not modified ===
			{
				Code: `var foo = 0; while (foo) { } foo = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loopConditionNotModified", Line: 1, Column: 21},
				},
			},

			// === Comparison group: both unmodified ===
			{
				Code: `var a = 0, b = 0; while (a < b) { }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loopConditionNotModified", Line: 1, Column: 26},
					{MessageId: "loopConditionNotModified", Line: 1, Column: 30},
				},
			},

			// === Do-while: identifier not modified ===
			{
				Code: `var foo = 0; do { } while (foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loopConditionNotModified", Line: 1, Column: 28},
				},
			},

			// === For: no incrementor, not modified in body ===
			{
				Code: `for (var i = 0; i < 10; ) { }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loopConditionNotModified", Line: 1, Column: 17},
				},
			},

			// === Modified outside loop but not inside ===
			{
				Code: `var foo = 0; while (foo) { } foo++;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loopConditionNotModified", Line: 1, Column: 21},
				},
			},

			// === || groups: both unmodified ===
			{
				Code: `var a = 0, b = 0; while (a || b) { }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loopConditionNotModified", Line: 1, Column: 26},
					{MessageId: "loopConditionNotModified", Line: 1, Column: 31},
				},
			},

			// === Variable shadowing: inner let shadows outer ===
			{
				Code: `var foo = 0; while (foo) { let foo = 1; foo++; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loopConditionNotModified", Line: 1, Column: 21},
				},
			},

			// === && operands are independent (not grouped) ===
			{
				Code: `var a = 0, b = 10; while (a && b) { a++; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loopConditionNotModified", Line: 1, Column: 32},
				},
			},

			// === ESLint issue #20844: main intentionally keeps LogicalExpression
			// operands independent; PR #21175 only added ternary checking. ===
			{
				Code: `function testFunction(filter) { let obj = { value: 1 }; do { obj = Object.getPrototypeOf(obj); } while (obj && !filter); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loopConditionNotModified", Line: 1, Column: 113, EndLine: 1, EndColumn: 119},
				},
			},

			// === || partial: only a modified, b should be reported ===
			{
				Code: `var a = 0, b = 0; while (a || b) { a++; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loopConditionNotModified", Line: 1, Column: 31},
				},
			},

			// === ?? (nullish coalescing) operands independent ===
			{
				Code: `var a: any = 0, b: any = 0; while (a ?? b) { a = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loopConditionNotModified", Line: 1, Column: 41},
				},
			},

			// === Nested logical: a || b && c — all independent ===
			{
				Code: `var a = 0, b = 0, c = 0; while (a || b && c) { a++; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loopConditionNotModified", Line: 1, Column: 38},
					{MessageId: "loopConditionNotModified", Line: 1, Column: 43},
				},
			},

			// === Complex: (a < b) && c — a<b group OK, c independent ===
			{
				Code: `var a = 0, b = 10, c = 0; while ((a < b) && c) { a++; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loopConditionNotModified", Line: 1, Column: 45},
				},
			},

			// === Shadowing via const ===
			{
				Code: `var x = 0; while (x) { const x = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loopConditionNotModified", Line: 1, Column: 19},
				},
			},

			// === Shadowing via function declaration ===
			{
				Code: `var x = 0; while (x) { function x() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loopConditionNotModified", Line: 1, Column: 19},
				},
			},

			// === Shadowing via catch clause ===
			{
				Code: `var e = 0; while (e) { try {} catch(e) { e = 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loopConditionNotModified", Line: 1, Column: 19},
				},
			},

			// === Function declared but NOT called in loop ===
			{
				Code: `var x = 0; function inc() { x++; } while (x < 10) { }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loopConditionNotModified", Line: 1, Column: 43},
				},
			},

			// === A write in a nested FunctionDeclaration belongs only to that
			// nearest function, not to a referenced outer function. ===
			{
				Code: `var x = 0; function outer() { function inner() { x++; } } while (x) { outer(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loopConditionNotModified", Line: 1, Column: 66},
				},
			},

			// === Ambient/lib global (declared in lib.dom.d.ts, not this file —
			// RefStore's per-file binder walk can't resolve it, only the
			// TypeChecker fallback can) never modified in the body ===
			{
				Code: `while (window) { }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loopConditionNotModified", Line: 1, Column: 8},
				},
			},

			// === AssignmentExpression is not a BinaryExpression group upstream ===
			{
				Code: `var x = 0, y = 1; while (x = y) { }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loopConditionNotModified", Line: 1, Column: 30},
				},
			},

			// === SequenceExpression operands are checked independently upstream ===
			{
				Code: `var x = 0, y = 1; while ((x++, y)) { }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loopConditionNotModified", Line: 1, Column: 32},
				},
			},

			// === Repeated references are reported individually ===
			{
				Code: `var x = 0; while (x < x) { }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loopConditionNotModified", Line: 1, Column: 19},
					{MessageId: "loopConditionNotModified", Line: 1, Column: 23},
				},
			},

			// === A condition reference to a modifying function modifies x, but the
			// function binding itself remains an independent condition reference. ===
			{
				Code: `var x = 0; function update() { x++; } while (x && update) { }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loopConditionNotModified", Line: 1, Column: 51},
				},
			},

			// === A standalone tagged template is not a dynamic group upstream ===
			{
				Code: "var tag = value; while (tag`value`) { }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loopConditionNotModified", Line: 1, Column: 25},
				},
			},

			// === ESTree ImportExpression is not a dynamic sentinel upstream ===
			{
				Code: `var x = "./module.js"; while (import(x)) { }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loopConditionNotModified", Line: 1, Column: 38},
				},
			},

			// === A var initializer does not count when the variable's first
			// definition is a function parameter. ===
			{
				Code: `function f(x) { while (x) { var x = 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loopConditionNotModified", Line: 1, Column: 24},
				},
			},
		},
	)
}
