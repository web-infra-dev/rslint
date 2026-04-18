package no_label_var

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoLabelVarRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoLabelVarRule,

		[]rule_tester.ValidTestCase{
			// ---- Upstream ESLint suite ----
			{Code: `function bar() { q: for(;;) { break q; } } function foo () { var q = t; }`},
			{Code: `function bar() { var x = foo; q: for(;;) { break q; } }`},

			// ---- Top-level / sibling-scope ----
			{Code: `q: for(;;) { break q; }`},
			{Code: `function bar() { a: for(;;) { b: for(;;) { break b; } } }`},
			{Code: `function bar(y) { q: for(;;) { break q; } }`},
			{Code: `for (let i = 0; i < 1; i++) { q: for(;;) { break q; } }`},
			{Code: `for (const k in obj) { q: for(;;) { break q; } }`},
			{Code: `for (const v of arr) { q: for(;;) { break q; } }`},

			// ---- TS type-only declarations should NOT clash (only values count) ----
			{Code: `interface X {} X: for(;;) { break X; }`},
			{Code: `type X = number; X: for(;;) { break X; }`},
			// NOTE: `import type` is intentionally NOT in the valid list — utils.IsShadowed
			// does not distinguish type-only imports from value imports, so it triggers.
			// See the matching invalid case below.

			// ---- Nested label inside iteration; outer var has different name ----
			{Code: `var x = 1; function f() { y: for(;;) { break y; } }`},

			// ---- Catch parameter has different name ----
			{Code: `try {} catch (e) { q: for(;;) { break q; } }`},

			// ---- Method / arrow / generator scopes don't leak names ----
			{Code: `class C { m() { q: for(;;) { break q; } } }`},
			{Code: `const fn = (a) => { q: for(;;) { break q; } };`},
			{Code: `function* gen() { q: for(;;) { break q; } }`},
		},

		[]rule_tester.InvalidTestCase{
			// ---- Upstream ESLint suite (with explicit message + position) ----
			{
				Code: `var x = foo; function bar() { x: for(;;) { break x; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "identifierClashWithLabel", Message: "Found identifier with same name as label.", Line: 1, Column: 31},
				},
			},
			{
				Code: `function bar() { var x = foo; x: for(;;) { break x; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "identifierClashWithLabel", Line: 1, Column: 31},
				},
			},
			{
				Code: `function bar(x) { x: for(;;) { break x; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "identifierClashWithLabel", Line: 1, Column: 19},
				},
			},

			// ---- Local-binding clashes (strategy A path) ----
			// let / const in same block
			{
				Code: `function bar() { let x = 1; x: for(;;) { break x; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "identifierClashWithLabel", Line: 1, Column: 29},
				},
			},
			{
				Code: `function bar() { const x = 1; x: for(;;) { break x; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "identifierClashWithLabel", Line: 1, Column: 31},
				},
			},
			// Sibling function declaration in the same block
			{
				Code: `function bar() { function x() {} x: for(;;) { break x; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "identifierClashWithLabel", Line: 1, Column: 34},
				},
			},
			// Class declaration at module scope
			{
				Code: "class X {}\nX: for(;;) { break X; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "identifierClashWithLabel", Line: 2, Column: 1},
				},
			},
			// Destructuring parameter
			{
				Code: `function bar({ x }) { x: for(;;) { break x; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "identifierClashWithLabel", Line: 1, Column: 23},
				},
			},
			// Catch binding (simple)
			{
				Code: `try {} catch (e) { e: for(;;) { break e; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "identifierClashWithLabel", Line: 1, Column: 20},
				},
			},
			// Catch binding (destructured)
			{
				Code: `try {} catch ({ a }) { a: for(;;) { break a; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "identifierClashWithLabel", Line: 1, Column: 24},
				},
			},
			// for-statement let init
			{
				Code: `for (let i = 0; i < 1; i++) { i: for(;;) { break i; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "identifierClashWithLabel", Line: 1, Column: 31},
				},
			},
			// for-of let init
			{
				Code: `for (let v of arr) { v: for(;;) { break v; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "identifierClashWithLabel", Line: 1, Column: 22},
				},
			},
			// for-in const init
			{
				Code: `for (const k in obj) { k: for(;;) { break k; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "identifierClashWithLabel", Line: 1, Column: 24},
				},
			},
			// Hoisted var declared further down the same function body
			{
				Code: `function f() { x: for(;;) { break x; } { var x = 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "identifierClashWithLabel", Line: 1, Column: 16},
				},
			},
			// Label name equals enclosing function declaration name
			{
				Code: `function foo() { foo: for(;;) { break foo; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "identifierClashWithLabel", Line: 1, Column: 18},
				},
			},
			// Label name equals named function expression name
			{
				Code: `(function fee() { fee: for(;;) { break fee; } })()`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "identifierClashWithLabel", Line: 1, Column: 19},
				},
			},
			// Default import
			{
				Code: `import x from 'mod'; x: for(;;) { break x; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "identifierClashWithLabel", Line: 1, Column: 22},
				},
			},
			// Named import
			{
				Code: `import { x } from 'mod'; x: for(;;) { break x; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "identifierClashWithLabel", Line: 1, Column: 26},
				},
			},
			// Namespace import
			{
				Code: `import * as x from 'mod'; x: for(;;) { break x; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "identifierClashWithLabel", Line: 1, Column: 27},
				},
			},
			// Renamed named import
			{
				Code: `import { y as x } from 'mod'; x: for(;;) { break x; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "identifierClashWithLabel", Line: 1, Column: 31},
				},
			},
			// Type-only import: utils.IsShadowed does not differentiate type-only
			// from value imports, so this is reported. Lock current behavior.
			{
				Code: `import type { X } from 'mod'; X: for(;;) { break X; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "identifierClashWithLabel", Line: 1, Column: 31},
				},
			},
			// TS enum (runtime value)
			{
				Code: `enum X { A } X: for(;;) { break X; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "identifierClashWithLabel", Line: 1, Column: 14},
				},
			},
			// TS namespace with a value export (runtime value)
			{
				Code: `namespace N { export const x = 1; } N: for(;;) { break N; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "identifierClashWithLabel", Line: 1, Column: 37},
				},
			},
			// Nested labels: outer label clashes with module-level var
			{
				Code: `var a = 1; function f() { a: for(;;) { b: for(;;) { break b; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "identifierClashWithLabel", Line: 1, Column: 27},
				},
			},

			// ---- Globals from tsgo lib (strategy B path; requires TypeChecker) ----
			{
				Code: `window: for (;;) { break window; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "identifierClashWithLabel", Line: 1, Column: 1},
				},
			},
			{
				Code: `console: for (;;) { break console; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "identifierClashWithLabel", Line: 1, Column: 1},
				},
			},
			{
				Code: `Promise: for (;;) { break Promise; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "identifierClashWithLabel", Line: 1, Column: 1},
				},
			},
			{
				Code: `Array: for (;;) { break Array; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "identifierClashWithLabel", Line: 1, Column: 1},
				},
			},
		},
	)
}
