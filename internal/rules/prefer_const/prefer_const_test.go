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

			// destructuring: "all" - uninitialized, destructuring write, one has extra reassignment
			{
				Code:    `let a: any, b: any; ({a, b} = ({} as any)); b = 1;`,
				Options: map[string]interface{}{"destructuring": "all"},
			},

			// destructuring: "all" - separate let statements, destructuring write, one reassigned
			{
				Code:    `function f() { let a: any; let b: any; ({a, b} = ({} as any)); b = 1; void a; }`,
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

			// Cross-declaration destructuring in different scope — unmergeable
			{Code: `function f() { let a: any; { let b: any; ({a, b} = ({} as any)); void b; } return a; }`},

			// Cross-declaration array destructuring in different scope
			{Code: `function f() { let a: any; { let b: any; ([a, b] = ([] as any)); void b; } return a; }`},

			// Cross-declaration with member expression in destructuring — unmergeable
			{Code: `function f() { let v: any; [({} as any).prop, v] = [1, 2]; return v; }`},

			// Cross-declaration with renamed property and member expression
			{Code: `let a: any; const b: any = {}; ({ a, c: (b as any).c } = ({} as any));`},

			// Member expression in nested object destructuring → unmergeable, don't report
			{Code: `function f() { let v: any; ({a: ({} as any).prop, b: v} = ({} as any)); return v; }`},

			// Shadowed variable: inner shorthand write should NOT affect outer count
			// (both outer and inner have 2+ writes → neither reported)
			{Code: `function f() { let x: any; x = 0; x = 1; { let x: any; ({x} = ({} as any)); x = 2; } }`},

			// === Dimension 1: ALL compound assignment operators as writes ===
			{Code: `let x = 1; x -= 2;`},
			{Code: `let x = 1; x *= 2;`},
			{Code: `let x = 1; x /= 2;`},
			{Code: `let x = 1; x %= 2;`},
			{Code: `let x = 1; x **= 2;`},
			{Code: `let x = 1; x <<= 2;`},
			{Code: `let x = 1; x >>= 2;`},
			{Code: `let x = 1; x >>>= 2;`},
			{Code: `let x = 1; x &= 2;`},
			{Code: `let x = 1; x |= 2;`},
			{Code: `let x = 1; x ^= 2;`},
			{Code: `let x = 1; x ||= 2;`},
			{Code: `let x = 1; x &&= 2;`},
			{Code: `let x = 1; x ??= 2;`},

			// === Dimension 1: Destructuring writes ===
			{Code: `let x = 1; [x] = [2];`},
			{Code: `let x = 1; ({x} = {x: 2} as any);`},
			{Code: `let x = 1; ({key: x} = {key: 2} as any);`},
			{Code: `let x = 1; [...x] = [2] as any;`},
			{Code: `let x = 1; ({...x} = {x: 2} as any);`},

			// === Dimension 1: Nested destructuring writes ===
			{Code: `let x = 1; [[x]] = [[2]] as any;`},
			{Code: `let x = 1; ({a: {b: x}} = {a: {b: 2}} as any);`},
			{Code: `let x = 1; [{x}] = [{x: 2}] as any;`},
			{Code: `let x = 1; ({a: [x]} = {a: [2]} as any);`},

			// === Dimension 1: Destructuring write with defaults ===
			{Code: `let x = 1; [x = 99] = [2] as any;`},
			{Code: `let x = 1; ({x = 99} = {} as any);`},
			{Code: `let x = 1; ({key: x = 99} = {} as any);`},

			// === Dimension 1: for-in / for-of reassignment ===
			{Code: `let x = 1; for (x in ({} as any)) {}`},
			{Code: `let x = 1; for (x of ([] as any)) {}`},

			// === Dimension 1: Type assertion / non-null writes ===
			{Code: `let x: any = 1; (x as any) = 2;`},
			{Code: `let x: any = 1; x! += 1;`},

			// === Dimension 2: Scope boundary reassignments ===
			{Code: `let x = 1; (function() { x = 2; })();`},
			{Code: `let x = 1; (() => { x = 2; })();`},
			{Code: `let x = 1; while (false) { x = 2; }`},
			{Code: `let x = 1; do { x = 2; } while (false);`},
			{Code: `let x = 1; try { x = 2; } catch {}`},
			{Code: `let x = 1; try {} catch { x = 2; }`},
			{Code: `let x = 1; try {} finally { x = 2; }`},

			// === Dimension 2: Class static block reassignment ===
			{Code: `class C { static { let x = 1; x = 2; } }`},
			{Code: `let x = 1; class C { static { x = 2; } }`},

			// === Dimension 2: Deep nesting ===
			{Code: `let x = 1; function f() { if (true) { for (let i = 0; i < 1; i++) { try { x = 2; } catch {} } } }`},

			// === Dimension 3: Uninitialized in nested blocks ===
			{Code: `function f() { let x: any; if (true) x = 0; }`},
			{Code: `function f() { let x: any; if (true) { x = 0; } }`},
			{Code: `function f() { let x: any; { x = 0; } }`},
			{Code: `function f() { let x: any; try { x = 0; } catch {} }`},
			{Code: `function f() { let x: any; try {} catch { x = 0; } }`},
			{Code: `function f() { let x: any; try {} finally { x = 0; } }`},

			// === Dimension 3: Uninitialized, non-standalone writes ===
			{Code: `function f() { let x: any; const y = true ? (x = 1) : 2; void x; void y; }`},
			{Code: `function f() { let x: any; do {} while (x = (1 as any)); }`},
			{Code: `function f() { let x: any; for (;; ++x); }`},

			// === Dimension 5: Member expression variants in destructuring ===
			{Code: `function f() { let v: any; [[({} as any).prop], v] = [[1], 2] as any; return v; }`},
			{Code: `function f() { let v: any; ({a: ({} as any).prop, b: v} = ({} as any)); return v; }`},

			// === Dimension 6: for-in/of with inner reassignment ===
			{Code: `for (let x in ({} as any)) { x = 'modified'; }`},
			{Code: `for (let x of ([] as any)) { x = 0; }`},

			// === Dimension 6: for-of iterator variable (let in outer scope) ===
			{Code: `let x: any; for (x of ([1, 2] as any)) { x; }`},

			// === Dimension 7: TypeScript constructs that count as writes ===
			{Code: `let x: any = 1; x! = 2;`},

			// === Dimension 8: Closure writes ===
			{Code: `let x = 1; const f = () => { x = 2; };`},
			{Code: `let x = 0; [1, 2].forEach(v => { x += v; });`},
			{Code: `let x: any = null; Promise.resolve().then(() => { x = 1; });`},

			// === Dimension 8: Lazy init pattern ===
			{Code: `let x: any; function init() { if (typeof x !== 'undefined') return; x = 1; }`},

			// === Dimension 8: Deep nested closure write ===
			{Code: `let x = 1; function a() { function b() { function c() { x = 2; } } }`},

			// === Dimension 8: Shadowing — outer reassigned, inner independently checked ===
			// (inner x is INVALID — tested in invalid section)

			// === Dimension 8: Async/generator writes ===
			{Code: `async function f() { let x = await Promise.resolve(1); x = await Promise.resolve(2); return x; }`},
			{Code: `function* g() { let x: any = yield 1; x = yield 2; return x; }`},

			// === destructuring: "all" valid: one reassigned, with destructuring assign ===
			{
				Code:    `let a: any, b: any; ({a, b} = ({} as any)); b = 1;`,
				Options: map[string]interface{}{"destructuring": "all"},
			},
			{
				Code:    `function f() { let a: any; let b: any; ({a, b} = ({} as any)); b = 1; void a; }`,
				Options: map[string]interface{}{"destructuring": "all"},
			},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// Simple let that should be const
			{
				Code:   `let x = 1;`,
				Output: []string{`const x = 1;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 5},
				},
			},

			// String value
			{
				Code:   `let x = 'hello';`,
				Output: []string{`const x = 'hello';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 5},
				},
			},

			// Object value (not reassigned, only properties modified)
			{
				Code:   `let obj = {key: 0}; obj.key = 1;`,
				Output: []string{`const obj = {key: 0}; obj.key = 1;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 5},
				},
			},

			// Array value
			{
				Code:   `let arr = [1, 2, 3];`,
				Output: []string{`const arr = [1, 2, 3];`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 5},
				},
			},

			// Only read, never reassigned
			{
				Code:   `let x = 1; console.log(x);`,
				Output: []string{`const x = 1; console.log(x);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 5},
				},
			},

			// Used in expression but never reassigned (both x and y)
			{
				Code:   `let x = 1; let y = x + 2;`,
				Output: []string{`const x = 1; const y = x + 2;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 5},
					{MessageId: "useConst", Line: 1, Column: 16},
				},
			},

			// for-in without reassignment
			{
				Code:   `for (let x in {a: 1}) { console.log(x); }`,
				Output: []string{`for (const x in {a: 1}) { console.log(x); }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 10},
				},
			},

			// for-of without reassignment
			{
				Code:   `for (let x of [1, 2, 3]) { console.log(x); }`,
				Output: []string{`for (const x of [1, 2, 3]) { console.log(x); }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 10},
				},
			},

			// Function expression never reassigned
			{
				Code:   `let fn = function() {};`,
				Output: []string{`const fn = function() {};`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 5},
				},
			},

			// Arrow function never reassigned
			{
				Code:   `let fn = () => {};`,
				Output: []string{`const fn = () => {};`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 5},
				},
			},

			// Multiple declarations, all never reassigned
			{
				Code:   `let x = 1, y = 2;`,
				Output: []string{`const x = 1, y = 2;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 5},
					{MessageId: "useConst", Line: 1, Column: 12},
				},
			},

			// Destructuring: none reassigned (default destructuring: "any")
			{
				Code:   `let {a, b} = {a: 1, b: 2};`,
				Output: []string{`const {a, b} = {a: 1, b: 2};`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 6},
					{MessageId: "useConst", Line: 1, Column: 9},
				},
			},

			// Array destructuring: none reassigned
			{
				Code:   `let [x, y] = [1, 2];`,
				Output: []string{`const [x, y] = [1, 2];`},
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
				Output:  []string{`const {a, b} = {a: 1, b: 2};`},
				Options: map[string]interface{}{"destructuring": "any"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 6},
					{MessageId: "useConst", Line: 1, Column: 9},
				},
			},

			// destructuring: "all" - all can be const, so report all
			{
				Code:    `let {a, b} = {a: 1, b: 2};`,
				Output:  []string{`const {a, b} = {a: 1, b: 2};`},
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
				Output:  []string{`const [x, y] = [1, 2];`},
				Options: map[string]interface{}{"destructuring": "any"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 6},
					{MessageId: "useConst", Line: 1, Column: 9},
				},
			},

			// Separate declarations - only a reported (b is reassigned)
			{
				Code:   `let {a} = {a: 1}; let {b} = {b: 2}; b = 1;`,
				Output: []string{`const {a} = {a: 1}; let {b} = {b: 2}; b = 1;`},
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

			// Cross-declaration destructuring in same scope — should report both
			{
				Code: `function f() { let a: any; let b: any; ({a, b} = ({} as any)); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 42},
					{MessageId: "useConst", Line: 1, Column: 45},
				},
			},

			// destructuring: "all" — uninitialized, all targets have single write → report all
			{
				Code:    `let a: any, b: any; ({a, b} = ({} as any));`,
				Options: map[string]interface{}{"destructuring": "all"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 23},
					{MessageId: "useConst", Line: 1, Column: 26},
				},
			},

			// Cross-declaration array destructuring in same scope
			{
				Code: `function f() { let a: any; let b: any; ([a, b] = ([] as any)); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 42},
					{MessageId: "useConst", Line: 1, Column: 45},
				},
			},

			// Cross-declaration renamed object destructuring in same scope
			{
				Code: `function f() { let a: any; let b: any; ({x: a, y: b} = ({} as any)); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 45},
					{MessageId: "useConst", Line: 1, Column: 51},
				},
			},

			// Cross-declaration in class static block — should report both
			{
				Code: `class C { static { let a: any; let b: any; ({a, b} = ({} as any)); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 46},
					{MessageId: "useConst", Line: 1, Column: 49},
				},
			},

			// === Cross-static-block: same-named vars, shorthand destructuring ===
			// A's x has 1 write → INVALID; B's x has 2 writes → should NOT appear here
			{
				Code: `class A { static { let x: any; ({x} = ({} as any)); } } class B { static { let x: any; ({x} = ({} as any)); x = 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 34},
				},
			},

			// === Dimension 1: Property mutation ≠ reassignment ===
			{
				Code:   `let obj = {a: 1}; obj.a = 2;`,
				Output: []string{`const obj = {a: 1}; obj.a = 2;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 5},
				},
			},
			{
				Code:   `let arr = [1]; arr[0] = 2;`,
				Output: []string{`const arr = [1]; arr[0] = 2;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 5},
				},
			},

			// === Dimension 2: Scope boundary — only read in nested scope ===
			{
				Code:   `let x = 1; function f() { void x; }`,
				Output: []string{`const x = 1; function f() { void x; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 5},
				},
			},
			{
				Code:   `let x = 1; const f = function() { void x; };`,
				Output: []string{`const x = 1; const f = function() { void x; };`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 5},
				},
			},
			{
				Code:   `let x = 1; const f = () => void x;`,
				Output: []string{`const x = 1; const f = () => void x;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 5},
				},
			},
			{
				Code:   `let x = 1; { void x; }`,
				Output: []string{`const x = 1; { void x; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 5},
				},
			},
			{
				Code:   `let x = 1; if (true) { void x; }`,
				Output: []string{`const x = 1; if (true) { void x; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 5},
				},
			},

			// === Dimension 2: Class static block — initialized ===
			{
				Code:   `class C { static { let x = 1; void x; } }`,
				Output: []string{`class C { static { const x = 1; void x; } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 24},
				},
			},

			// === Dimension 2: Class static block — uninitialized + simple assign ===
			{
				Code: `class C { static { let x: any; x = 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 32},
				},
			},

			// === Dimension 2: Class static block — if inside ===
			{
				Code:   `class C { static { if (true) { let x = 1; void x; } } }`,
				Output: []string{`class C { static { if (true) { const x = 1; void x; } } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 36},
				},
			},

			// === Dimension 2: Class static block — read in nested if ===
			{
				Code:   `class C { static { let x = 1; if (true) { void x; } } }`,
				Output: []string{`class C { static { const x = 1; if (true) { void x; } } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 24},
				},
			},

			// === Dimension 2: Class method ===
			{
				Code:   `class C { m() { let x = 1; void x; } }`,
				Output: []string{`class C { m() { const x = 1; void x; } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 21},
				},
			},

			// === Dimension 2: Getter ===
			{
				Code:   `class C { get g() { let x = 1; return x; } }`,
				Output: []string{`class C { get g() { const x = 1; return x; } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 25},
				},
			},

			// === Dimension 2: Switch case ===
			{
				Code:   `function f(n: number) { switch (n) { case 0: { let x = 1; void x; } } }`,
				Output: []string{`function f(n: number) { switch (n) { case 0: { const x = 1; void x; } } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 52},
				},
			},

			// === Dimension 3: Uninitialized — double parenthesized ===
			{
				Code: `function f() { let x: any; ((x = 1)); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 30},
				},
			},

			// === Dimension 3: Uninitialized in for-of body ===
			{
				Code: `function f() { for (const b of ([1] as any)) { let a: any; a = 1; void b; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 60},
				},
			},

			// === Dimension 3: Uninitialized in for-of body with destructuring write ===
			{
				Code: `function f() { for (const b of ([1] as any)) { let a: any; ({a} = {a: 1}); void b; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 62},
				},
			},

			// === Dimension 3: Read before assign — reported at declaration ===
			{
				Code: `let x: any; function f() { void x; } x = 0;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 5},
				},
			},

			// === Dimension 4: Nested object destructuring ===
			{
				Code:   `let {a: {b, c}} = {a: {b: 1, c: 2}};`,
				Output: []string{`const {a: {b, c}} = {a: {b: 1, c: 2}};`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 10},
					{MessageId: "useConst", Line: 1, Column: 13},
				},
			},

			// === Dimension 4: Array with holes ===
			{
				Code:   `const x = [1,2]; let [,y] = x;`,
				Output: []string{`const x = [1,2]; const [,y] = x;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 24},
				},
			},
			{
				Code:   `const x = [1,2,3]; let [y,,z] = x;`,
				Output: []string{`const x = [1,2,3]; const [y,,z] = x;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 25},
					{MessageId: "useConst", Line: 1, Column: 28},
				},
			},

			// === Dimension 4: Rest element, rest reassigned ===
			{
				Code:    `let { name, ...rest } = ({} as any); rest = {};`,
				Options: map[string]interface{}{"destructuring": "any"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 7},
				},
			},

			// === Dimension 4: Mixed destructuring + non-destructuring ===
			{
				Code: `let {a, b} = ({} as any), c: any;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 6},
					{MessageId: "useConst", Line: 1, Column: 9},
				},
			},

			// === Dimension 4: for-of destructuring → INVALID ===
			{
				Code:   `for (let {a, b} of ([{a:1, b:2}] as any)) { void a; void b; }`,
				Output: []string{`for (const {a, b} of ([{a:1, b:2}] as any)) { void a; void b; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 11},
					{MessageId: "useConst", Line: 1, Column: 14},
				},
			},

			// === Dimension 4: for-of array destructuring → INVALID ===
			{
				Code:   `for (let [a, b] of ([[1, 2]] as any)) { void a; void b; }`,
				Output: []string{`for (const [a, b] of ([[1, 2]] as any)) { void a; void b; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 11},
					{MessageId: "useConst", Line: 1, Column: 14},
				},
			},

			// === Dimension 4: Deeply nested destructuring → INVALID ===
			{
				Code:   `let {a: {b: {c}}} = {a: {b: {c: 1}}};`,
				Output: []string{`const {a: {b: {c}}} = {a: {b: {c: 1}}};`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 14},
				},
			},

			// === Dimension 5: Cross-decl across 3 let statements ===
			{
				Code: `function f() { let a: any; let b: any; let c: any; ({a, b, c} = ({} as any)); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 54},
					{MessageId: "useConst", Line: 1, Column: 57},
					{MessageId: "useConst", Line: 1, Column: 60},
				},
			},

			// === Dimension 5: Nested destructuring assignment same scope ===
			{
				Code: `function f() { let a: any, b: any; ({x: {y: a}, z: b} = ({} as any)); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 45},
					{MessageId: "useConst", Line: 1, Column: 52},
				},
			},

			// === Dimension 5: Deep nesting same scope ===
			{
				Code: `function f() { let a: any; ({x: {y: {z: a}}} = ({} as any)); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 41},
				},
			},

			// === Dimension 5: Rest + named, cross-decl same scope ===
			{
				Code: `function f() { let a: any; let b: any; [a, ...b] = [1, 2]; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 41},
					{MessageId: "useConst", Line: 1, Column: 47},
				},
			},

			// === Dimension 6: for-in: both loop var and inner let ===
			{
				Code:   `for (let k in ({} as any)) { let v = 1; void k; void v; }`,
				Output: []string{`for (const k in ({} as any)) { const v = 1; void k; void v; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 10},
					{MessageId: "useConst", Line: 1, Column: 34},
				},
			},

			// === Dimension 6: Deeply nested for loops ===
			{
				Code:   `for (let x of [1]) { for (let y of [2]) { for (let z of [3]) { let w = x + y + z; void w; } } }`,
				Output: []string{`for (const x of [1]) { for (const y of [2]) { for (const z of [3]) { const w = x + y + z; void w; } } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 10},
					{MessageId: "useConst", Line: 1, Column: 31},
					{MessageId: "useConst", Line: 1, Column: 52},
					{MessageId: "useConst", Line: 1, Column: 68},
				},
			},

			// === Dimension 7: as / satisfies / typeof not writes ===
			{
				Code:   `let x = 1; void (x as number);`,
				Output: []string{`const x = 1; void (x as number);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 5},
				},
			},
			{
				Code:   `let x = {a: 1} satisfies Record<string, number>;`,
				Output: []string{`const x = {a: 1} satisfies Record<string, number>;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 5},
				},
			},
			{
				Code:   `let x = 'key'; const obj = { [x]: 1 }; void obj;`,
				Output: []string{`const x = 'key'; const obj = { [x]: 1 }; void obj;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 5},
				},
			},

			// === Dimension 7: export let ===
			{
				Code:   `export let x = 1;`,
				Output: []string{`export const x = 1;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 12},
				},
			},

			// === Dimension 7: let in namespace ===
			{
				Code:   `namespace NS { export let x = 1; }`,
				Output: []string{`namespace NS { export const x = 1; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 27},
				},
			},

			// === Dimension 7: let in catch ===
			{
				Code:   `try { throw new Error(); } catch (e) { let x = e; void x; }`,
				Output: []string{`try { throw new Error(); } catch (e) { const x = e; void x; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 44},
				},
			},

			// === Dimension 8: Closure, only reads ===
			{
				Code:   `let x = 1; const f = () => x; void f;`,
				Output: []string{`const x = 1; const f = () => x; void f;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 5},
				},
			},

			// === Dimension 8: IIFE with inner const candidate ===
			{
				Code:   `(function() { let x = 1; void x; })();`,
				Output: []string{`(function() { const x = 1; void x; })();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 19},
				},
			},

			// === Dimension 8: Shadowing — inner should be independent ===
			{
				Code:   `let x = 1; { let x = 2; void x; } x = 3;`,
				Output: []string{`let x = 1; { const x = 2; void x; } x = 3;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 18},
				},
			},

			// === Shadowed shorthand: no double-counting ===
			// Inner ({x}) should not be counted for outer x; outer x has 2 writes → VALID
			// But inner x has 1 write → INVALID
			{
				Code: `function f() { let x: any; x = 1; x = 2; { let x: any; ({x} = ({} as any)); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 58},
				},
			},

			// === Dimension 8: Async / Generator ===
			{
				Code:   `async function f() { let x = await Promise.resolve(1); void x; }`,
				Output: []string{`async function f() { const x = await Promise.resolve(1); void x; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 26},
				},
			},
			{
				Code:   `function* g() { let x = yield 1; void x; }`,
				Output: []string{`function* g() { const x = yield 1; void x; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 21},
				},
			},

			// === Dimension 8: Multiple declarations, middle one reassigned ===
			{
				Code: `let a = 1, b = 2, c = 3; b = 10; void a; void c;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 5},
					{MessageId: "useConst", Line: 1, Column: 19},
				},
			},

			// === Dimension 8: Template literal (not a write) ===
			{
				Code:   "let x = 'world'; void `hello ${x}`;",
				Output: []string{"const x = 'world'; void `hello ${x}`;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 5},
				},
			},

			// === Shadowed: inner shorthand destructuring does NOT affect outer uninitialized ===
			// Outer x reported at write location (x = 1), inner x at shorthand location
			{
				Code: `function f() { let x: any; { let x: any; ({x} = ({} as any)); } x = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 65},
					{MessageId: "useConst", Line: 1, Column: 44},
				},
			},

			// === Same-scope var + let in destructuring: reported, no fix ===
			{
				Code: `function f() { var x: any; let y: any; ({y, ...x} = ({} as any)); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 42},
				},
			},

			// === Same-scope import + let in destructuring: reported, no fix ===
			{
				Code: `import { readFileSync } from 'fs'; let y: any; ({readFileSync: y} = ({} as any));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useConst", Line: 1, Column: 64},
				},
			},

			// === Same-scope function param + let in destructuring: not reported ===
			// (parameters are in the function scope, but findContainingBlock for the
			// parameter declaration returns the function body block, matching the let)
		},
	)
}
