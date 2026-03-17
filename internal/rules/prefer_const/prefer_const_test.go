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
			{
				Code: `let x; x = 0;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 5},
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
					{MessageId: "useConst", Line: 1, Column: 5},
				},
			},
		},
	)
}
