package prefer_const

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferConstRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferConstRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// Already const
			{Code: `const x = 1;`},

			// Reassigned variable
			{Code: `let x = 1; x = 2;`},

			// Reassigned via +=
			{Code: `let x = 1; x += 2;`},

			// Reassigned via ++
			{Code: `let x = 1; x++;`},

			// Reassigned via prefix ++
			{Code: `let x = 1; ++x;`},

			// Reassigned via --
			{Code: `let x = 1; x--;`},

			// Reassigned via prefix --
			{Code: `let x = 1; --x;`},

			// No initializer, never assigned - don't report
			{Code: `let x;`},

			// No initializer, multiple assignments - don't report
			{Code: `let x; x = 0; x = 1;`},

			// var declaration (not let)
			{Code: `var x = 1;`},

			// Reassigned in function
			{Code: `let x = 1; function f() { x = 2; }`},

			// Reassigned in arrow function
			{Code: `let x = 1; const f = () => { x = 2; };`},

			// Reassigned in nested block
			{Code: `let x = 1; { x = 2; }`},

			// Reassigned in if
			{Code: `let x = 1; if (true) { x = 2; }`},

			// Reassigned via array destructuring
			{Code: `let x = 1; [x] = [2];`},

			// Reassigned via object destructuring
			{Code: `let x = 1; ({x} = {x: 2});`},

			// Reassigned via nested destructuring
			{Code: `let a = 1; [{a}] = [{a: 2}];`},

			// For loop counter
			{Code: `for (let i = 0; i < 10; i++) {}`},

			// For loop with reassignment in body
			{Code: `for (let i = 0; i < 10; i++) { i = 5; }`},

			// For loop variable never reassigned - ESLint skips regular for-loop initializers
			{Code: `for (let x = 10; x > 0; ) { break; }`},

			// For loop with multiple declarators, none reassigned
			{Code: `for (let x = 0, y = 10; x < y; ) { break; }`},

			// for-in with reassignment inside loop
			{Code: `for (let x in obj) { x = 'modified'; }`},

			// for-of with reassignment inside loop
			{Code: `for (let x of arr) { x = 'modified'; }`},

			// destructuring: "all" - not all can be const (b is reassigned)
			{
				Code:    `let {a, b} = {a: 1, b: 2}; b = 3;`,
				Options: map[string]interface{}{"destructuring": "all"},
			},

			// destructuring: "all" - array destructuring, one reassigned
			{
				Code:    `let [x, y] = [1, 2]; y = 3;`,
				Options: map[string]interface{}{"destructuring": "all"},
			},

			// ignoreReadBeforeAssign: true - variable read before first assignment
			{
				Code:    `let x; console.log(x); x = 0;`,
				Options: map[string]interface{}{"ignoreReadBeforeAssign": true},
			},

			// Uninitialized, assigned inside if block - can't be safely converted to const
			{Code: `let x: number; if (true) { x = 1; }`},

			// Uninitialized, assigned inside try block
			{Code: `let x: number; try { x = 1; } catch { x = 2; }`},

			// Uninitialized, single assignment inside try block (can't merge into declaration)
			{Code: `let x: number; try { x = 1; } catch {}`},

			// Uninitialized, assigned inside nested block
			{Code: `let x: number; { x = 1; }`},

			// Uninitialized, assigned inside for loop body
			{Code: `let x: number; for (let i = 0; i < 1; i++) { x = i; }`},

			// Uninitialized, assigned inside arrow function
			{Code: `let x: number; const fn = () => { x = 1; };`},

			// Uninitialized, assignment in while condition (not standalone ExpressionStatement)
			{Code: `function f() { let x: string | null; while (x = g()) { void x; } } function g(): string | null { return null; }`},

			// Uninitialized, assignment in if condition
			{Code: `function f() { let x: number; if (x = g()) { return x; } return 0; } function g(): number { return 1; }`},

			// Uninitialized, assignment in for condition
			{Code: `function f() { let x: number; for (; x = g(); ) { void x; } } function g(): number { return 0; }`},

			// Chained assignment: b's write is inside another assignment, not standalone
			{Code: `let a = 0; let b: number; a = b = 1;`},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// Simple let that should be const
			{
				Code: `let x = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 5},
				},
			},

			// String value
			{
				Code: `let x = 'hello';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 5},
				},
			},

			// Object value (not reassigned, only properties modified)
			{
				Code: `let obj = {key: 0}; obj.key = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 5},
				},
			},

			// Array value
			{
				Code: `let arr = [1, 2, 3];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 5},
				},
			},

			// Only read, never reassigned
			{
				Code: `let x = 1; console.log(x);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 5},
				},
			},

			// Used in expression but never reassigned (both x and y)
			{
				Code: `let x = 1; let y = x + 2;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 5},
					{MessageId: "useConst", Line: 1, Column: 16},
				},
			},

			// for-in without reassignment
			{
				Code: `for (let x in {a: 1}) { console.log(x); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 10},
				},
			},

			// for-of without reassignment
			{
				Code: `for (let x of [1, 2, 3]) { console.log(x); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 10},
				},
			},

			// Function expression never reassigned
			{
				Code: `let fn = function() {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 5},
				},
			},

			// Arrow function never reassigned
			{
				Code: `let fn = () => {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 5},
				},
			},

			// Multiple declarations, all never reassigned
			{
				Code: `let x = 1, y = 2;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 5},
					{MessageId: "useConst", Line: 1, Column: 12},
				},
			},

			// Destructuring: none reassigned (default destructuring: "any")
			{
				Code: `let {a, b} = {a: 1, b: 2};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 6},
					{MessageId: "useConst", Line: 1, Column: 9},
				},
			},

			// Array destructuring: none reassigned
			{
				Code: `let [x, y] = [1, 2];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 6},
					{MessageId: "useConst", Line: 1, Column: 9},
				},
			},

			// Uninitialized let with single assignment - should be const
			// ESLint reports at the write location (column 8), not the declaration
			{
				Code: `let x; x = 0;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 8},
				},
			},

			// Uninitialized let, parenthesized assignment
			{
				Code: `let x: number; (x = 1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 17},
				},
			},

			// Uninitialized let, compound assignment (standalone)
			{
				Code: `let x: any; x += 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 13},
				},
			},

			// Uninitialized let, logical assignment (standalone)
			{
				Code: `let x: any; x ||= 'hi';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 13},
				},
			},

			// Uninitialized let, array destructuring assignment
			{
				Code: `let x: number; [x] = [1];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 17},
				},
			},

			// Uninitialized let, object destructuring assignment (shorthand)
			{
				Code: `let x: number; ({x} = {x: 1});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 18},
				},
			},

			// Uninitialized let, object destructuring assignment (renamed)
			{
				Code: `let x: number; ({val: x} = {val: 1});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 23},
				},
			},

			// Uninitialized let, array destructuring with default value
			{
				Code: `let x: number; [x = 5] = [1];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 17},
				},
			},

			// Uninitialized let, object destructuring rename with default
			{
				Code: `let x: number; ({val: x = 5} = {val: 1});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 23},
				},
			},

			// Uninitialized, multiple via array destructuring assignment
			{
				Code: `let a: number, b: number; [a, b] = [1, 2];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 28},
					{MessageId: "useConst", Line: 1, Column: 31},
				},
			},

			// destructuring: "any" - both reported (explicit option)
			{
				Code:    `let {a, b} = {a: 1, b: 2};`,
				Options: map[string]interface{}{"destructuring": "any"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 6},
					{MessageId: "useConst", Line: 1, Column: 9},
				},
			},

			// destructuring: "all" - all can be const, so report all
			{
				Code:    `let {a, b} = {a: 1, b: 2};`,
				Options: map[string]interface{}{"destructuring": "all"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 6},
					{MessageId: "useConst", Line: 1, Column: 9},
				},
			},

			// destructuring: "any" - only a reported when b is reassigned
			{
				Code:    `let {a, b} = {a: 1, b: 2}; b = 3;`,
				Options: map[string]interface{}{"destructuring": "any"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 6},
				},
			},

			// Array destructuring: both reported with destructuring: "any"
			{
				Code:    `let [x, y] = [1, 2];`,
				Options: map[string]interface{}{"destructuring": "any"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 6},
					{MessageId: "useConst", Line: 1, Column: 9},
				},
			},

			// Separate declarations - only a reported (b is reassigned)
			{
				Code: `let {a} = {a: 1}; let {b} = {b: 2}; b = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 6},
				},
			},

			// ignoreReadBeforeAssign: false - uninitialized with single assignment still reported
			{
				Code:    `let x; x = 0;`,
				Options: map[string]interface{}{"ignoreReadBeforeAssign": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 8},
				},
			},
		},
	)
}
