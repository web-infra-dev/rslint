package consistent_return

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestConsistentReturnExtras locks in branches and edge shapes that the upstream
// test suite doesn't exercise. Each case carries an inline comment pointing at
// the specific branch / Dimension 4 row / tsgo AST quirk it covers, so future
// refactors can't silently regress them without breaking a named lock-in. Every
// verdict below was read off ESLint itself.
//
// N/A rows of the Dimension 4 checklist:
//   - Autofix boundaries: this rule emits no fix and no suggestion.
//   - Element access (`X['y']`, `X[0]`) on a receiver: the rule never reads a
//     member access, only a member's key and a `return` argument.
func TestConsistentReturnExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ConsistentReturnRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: parenthesized return argument ----
			{Code: `function foo() { if (a) return (undefined); return; }`, Options: map[string]interface{}{"treatUndefinedAsUnspecified": true}},
			{Code: `function foo() { if (a) return ((undefined)); return; }`, Options: map[string]interface{}{"treatUndefinedAsUnspecified": true}},
			{Code: `function foo() { if (a) return (void 0); return; }`, Options: map[string]interface{}{"treatUndefinedAsUnspecified": true}},
			{Code: `function foo() { if (a) return ((void 0)); return; }`, Options: map[string]interface{}{"treatUndefinedAsUnspecified": true}},

			// ---- Dimension 4: graceful degradation on body-less and empty forms ----
			{Code: `declare function foo(): number;`},
			{Code: `abstract class C { abstract foo(): number; }`},
			{Code: `function foo() {}`},
			{Code: `const f = () => a;`},
			{Code: `const f = () => {};`},
			{Code: `class C {}`},
			{Code: `class C { static { } }`},
			{Code: `const o = { ...spread };`},

			// ---- Locks in upstream checkLastSegment() arm 1: hasReturnValue false leaves the end alone ----
			{Code: `function foo() { if (a) return; }`},

			// ---- Locks in upstream checkLastSegment() arm 2: the end of the code path is unreachable ----
			{Code: `function foo() { if (a) return 1; throw new Error(); }`},
			{Code: `function foo() { if (a) return 1; while (true) { } }`},
			{Code: `function foo() { for (;;) { return 1; } }`},
			{Code: `function foo() { do { return 1; } while (a); }`},
			{Code: `function foo() { if (a) { return 1; } else { return 2; } }`},
			{Code: `function foo() { if (a) { return 1; } else { throw 1; } }`},
			{Code: `function foo() { switch (a) { case 1: return 1; default: return 2; } }`},
			{Code: `function foo() { try { return 1; } catch (e) { return 2; } }`},
			{Code: `function foo() { if (a) return 1; try { foo(); return 2; } catch (e) { return 3; } }`},
			{Code: `function foo() { if (a) return 1; try { return 2; } catch (e) { } }`},
			{Code: `function foo() { if (a) return 1; try { } finally { return 2; } }`},

			// ---- Locks in upstream checkLastSegment() arm 3: isES5Constructor ----
			{Code: `function Foo() { if (a) return 1; }`},
			{Code: `const f = function Foo() { if (a) return 1; };`},
			{Code: `const o = { foo: function Bar() { if (a) return 1; } };`},
			{Code: `class C { foo = function Bar() { if (a) return 1; } }`},
			{Code: `function Ábc() { if (a) return 1; }`},
			{Code: `function ǅbc() { if (a) return 1; }`},
			{Code: `function Ⅰ() { if (a) return 1; }`},
			{Code: `async function Foo() { if (a) return 1; }`},
			{Code: `function* Foo() { if (a) return 1; }`},

			// ---- Locks in upstream checkLastSegment() arm 4: isClassConstructor ----
			{Code: `class C { constructor() { if (a) return 1; } }`},
			{Code: `const C = class { constructor() { if (a) return 1; } };`},

			// ---- Locks in upstream ReturnStatement() arm: treatUndefinedAsUnspecified off by default ----
			{Code: `function foo() { if (a) return undefined; return; }`, Options: map[string]interface{}{"treatUndefinedAsUnspecified": true}},

			// ---- Locks in upstream ReturnStatement() arm: `void` is read off the top-level operator only ----
			{Code: `function foo() { if (a) return void foo(); return; }`, Options: map[string]interface{}{"treatUndefinedAsUnspecified": true}},

			// ---- Locks in upstream ReturnStatement() arm: isSpecificId is syntactic, not a lookup ----
			{Code: `function foo() { var undefined = 1; if (a) return undefined; return; }`, Options: map[string]interface{}{"treatUndefinedAsUnspecified": true}},

			// ---- Real-user: eslint#8590 a switch with a default counts as returning ----
			{Code: `function tes(a, b) { if (a) return 'Hello!'; switch (b) { case 1: return 'tada'; default: return 'boo'; } }`},

			// ---- Options plumbing: every JSON shape a config can hand the rule ----
			// The bare object is the single-option CLI shape; the one-element
			// array is the rule_tester / multi-element shape. Both must reach
			// the same option value.
			{Code: `function foo() { if (a) return undefined; return; }`, Options: map[string]interface{}{"treatUndefinedAsUnspecified": true}},
			{Code: `function foo() { if (a) return undefined; return; }`, Options: []interface{}{map[string]interface{}{"treatUndefinedAsUnspecified": true}}},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: TS type-expression wrappers around the return argument ----
			{
				Code:    `function foo() { if (a) return undefined as any; return; }`,
				Options: map[string]interface{}{"treatUndefinedAsUnspecified": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturnValue",
					Message:   "Function 'foo' expected a return value.",
					Line:      1,
					Column:    50,
					EndLine:   1,
					EndColumn: 57,
				}},
			},
			{
				Code:    `function foo() { if (a) return undefined satisfies unknown; return; }`,
				Options: map[string]interface{}{"treatUndefinedAsUnspecified": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturnValue",
					Message:   "Function 'foo' expected a return value.",
					Line:      1,
					Column:    61,
					EndLine:   1,
					EndColumn: 68,
				}},
			},
			{
				Code:    `function foo() { if (a) return undefined!; return; }`,
				Options: map[string]interface{}{"treatUndefinedAsUnspecified": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturnValue",
					Message:   "Function 'foo' expected a return value.",
					Line:      1,
					Column:    44,
					EndLine:   1,
					EndColumn: 51,
				}},
			},
			{
				Code:    `function foo() { if (a) return (void 0) as any; return; }`,
				Options: map[string]interface{}{"treatUndefinedAsUnspecified": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturnValue",
					Message:   "Function 'foo' expected a return value.",
					Line:      1,
					Column:    49,
					EndLine:   1,
					EndColumn: 56,
				}},
			},

			// ---- Dimension 4: optional chain as the return argument ----
			{
				Code: `function foo() { if (a) return obj?.x; return; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturnValue",
					Message:   "Function 'foo' expected a return value.",
					Line:      1,
					Column:    40,
					EndLine:   1,
					EndColumn: 47,
				}},
			},
			{
				Code: `function foo() { if (a) return obj?.(); return; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturnValue",
					Message:   "Function 'foo' expected a return value.",
					Line:      1,
					Column:    41,
					EndLine:   1,
					EndColumn: 48,
				}},
			},

			// ---- Dimension 4: key forms on a class member ----
			{
				Code: `class C { foo() { if (a) return 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of method 'foo'.",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 14,
				}},
			},
			{
				Code: `class C { 'foo'() { if (a) return 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of method 'foo'.",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 16,
				}},
			},
			{
				Code: `class C { 0() { if (a) return 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of method '0'.",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 12,
				}},
			},
			{
				Code: `class C { #foo() { if (a) return 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of private method '#foo'.",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 15,
				}},
			},
			{
				Code: `class C { ['foo']() { if (a) return 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of method 'foo'.",
					Line:      1,
					Column:    12,
					EndLine:   1,
					EndColumn: 17,
				}},
			},
			{
				Code: "class C { [`foo`]() { if (a) return 1; } }",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of method 'foo'.",
					Line:      1,
					Column:    12,
					EndLine:   1,
					EndColumn: 17,
				}},
			},
			{
				Code: `class C { [k]() { if (a) return 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of method.",
					Line:      1,
					Column:    12,
					EndLine:   1,
					EndColumn: 13,
				}},
			},
			{
				Code: `class C { [(k)]() { if (a) return 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of method.",
					Line:      1,
					Column:    13,
					EndLine:   1,
					EndColumn: 14,
				}},
			},

			// ---- Dimension 4: key forms on an object literal member ----
			{
				Code: `const o = { foo() { if (a) return 1; } };`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of method 'foo'.",
					Line:      1,
					Column:    13,
					EndLine:   1,
					EndColumn: 16,
				}},
			},
			{
				Code: `const o = { 'foo'() { if (a) return 1; } };`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of method 'foo'.",
					Line:      1,
					Column:    13,
					EndLine:   1,
					EndColumn: 18,
				}},
			},
			{
				Code: `const o = { 0() { if (a) return 1; } };`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of method '0'.",
					Line:      1,
					Column:    13,
					EndLine:   1,
					EndColumn: 14,
				}},
			},
			{
				Code: `const o = { ['foo']() { if (a) return 1; } };`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of method 'foo'.",
					Line:      1,
					Column:    14,
					EndLine:   1,
					EndColumn: 19,
				}},
			},
			{
				Code: "const o = { [`foo`]() { if (a) return 1; } };",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of method 'foo'.",
					Line:      1,
					Column:    14,
					EndLine:   1,
					EndColumn: 19,
				}},
			},
			{
				Code: `const o = { [k]() { if (a) return 1; } };`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of method.",
					Line:      1,
					Column:    14,
					EndLine:   1,
					EndColumn: 15,
				}},
			},
			{
				Code: `const o = { [(k)]() { if (a) return 1; } };`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of method.",
					Line:      1,
					Column:    15,
					EndLine:   1,
					EndColumn: 16,
				}},
			},

			// ---- Dimension 4: accessors, in a class and in an object literal ----
			{
				Code: `class C { get foo() { if (a) return 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of getter 'foo'.",
					Line:      1,
					Column:    15,
					EndLine:   1,
					EndColumn: 18,
				}},
			},
			{
				Code: `class C { set foo(v) { if (a) return 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of setter 'foo'.",
					Line:      1,
					Column:    15,
					EndLine:   1,
					EndColumn: 18,
				}},
			},
			{
				Code: `class C { get #foo() { if (a) return 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of private getter '#foo'.",
					Line:      1,
					Column:    15,
					EndLine:   1,
					EndColumn: 19,
				}},
			},
			{
				Code: `class C { get [k]() { if (a) return 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of getter.",
					Line:      1,
					Column:    16,
					EndLine:   1,
					EndColumn: 17,
				}},
			},
			{
				Code: `const o = { get foo() { if (a) return 1; } };`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of getter 'foo'.",
					Line:      1,
					Column:    20,
					EndLine:   1,
					EndColumn: 21,
				}},
			},
			{
				Code: `const o = { set foo(v) { if (a) return 1; } };`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of setter 'foo'.",
					Line:      1,
					Column:    20,
					EndLine:   1,
					EndColumn: 21,
				}},
			},
			{
				Code: `const o = { get [k]() { if (a) return 1; } };`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of getter.",
					Line:      1,
					Column:    20,
					EndLine:   1,
					EndColumn: 21,
				}},
			},
			{
				Code: `const o = { get [(k)]() { if (a) return 1; } };`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of getter.",
					Line:      1,
					Column:    22,
					EndLine:   1,
					EndColumn: 23,
				}},
			},

			// ---- Dimension 4: static and private modifiers ----
			{
				Code: `class C { static foo() { if (a) return 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of static method 'foo'.",
					Line:      1,
					Column:    18,
					EndLine:   1,
					EndColumn: 21,
				}},
			},
			{
				Code: `class C { static #foo() { if (a) return 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of static private method '#foo'.",
					Line:      1,
					Column:    18,
					EndLine:   1,
					EndColumn: 22,
				}},
			},
			{
				Code: `class C { static get #foo() { if (a) return 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of static private getter '#foo'.",
					Line:      1,
					Column:    22,
					EndLine:   1,
					EndColumn: 26,
				}},
			},
			{
				Code: `class C { static async #foo() { if (a) return 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of static private async method '#foo'.",
					Line:      1,
					Column:    24,
					EndLine:   1,
					EndColumn: 28,
				}},
			},

			// ---- Dimension 4: declaration and container forms ----
			{
				Code: `const f = function () { if (a) return 1; };`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of function.",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 19,
				}},
			},
			{
				Code: `const f = function bar() { if (a) return 1; };`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of function 'bar'.",
					Line:      1,
					Column:    20,
					EndLine:   1,
					EndColumn: 23,
				}},
			},
			{
				Code: `const f = () => { if (a) return 1; };`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of arrow function.",
					Line:      1,
					Column:    14,
					EndLine:   1,
					EndColumn: 16,
				}},
			},
			{
				Code: `class C { f = () => { if (a) return 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of method 'f'.",
					Line:      1,
					Column:    18,
					EndLine:   1,
					EndColumn: 20,
				}},
			},
			{
				Code: `class C { f = function () { if (a) return 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of method 'f'.",
					Line:      1,
					Column:    15,
					EndLine:   1,
					EndColumn: 23,
				}},
			},
			{
				Code: `class C { static #f = () => { if (a) return 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of static private method '#f'.",
					Line:      1,
					Column:    26,
					EndLine:   1,
					EndColumn: 28,
				}},
			},
			{
				Code: `const o = { foo: function () { if (a) return 1; } };`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of method 'foo'.",
					Line:      1,
					Column:    18,
					EndLine:   1,
					EndColumn: 26,
				}},
			},
			{
				Code: `const o = { foo: function bar() { if (a) return 1; } };`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of method 'foo'.",
					Line:      1,
					Column:    27,
					EndLine:   1,
					EndColumn: 30,
				}},
			},
			{
				Code: `const o = { foo: () => { if (a) return 1; } };`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of method 'foo'.",
					Line:      1,
					Column:    21,
					EndLine:   1,
					EndColumn: 23,
				}},
			},
			{
				Code: `const C = class { foo() { if (a) return 1; } };`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of method 'foo'.",
					Line:      1,
					Column:    19,
					EndLine:   1,
					EndColumn: 22,
				}},
			},
			{
				Code: `export default function () { if (a) return 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of function.",
					Line:      1,
					Column:    16,
					EndLine:   1,
					EndColumn: 24,
				}},
			},
			{
				Code: `export default async function () { if (a) return 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of async function.",
					Line:      1,
					Column:    16,
					EndLine:   1,
					EndColumn: 21,
				}},
			},
			{
				Code: `export default function foo() { if (a) return 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of function 'foo'.",
					Line:      1,
					Column:    25,
					EndLine:   1,
					EndColumn: 28,
				}},
			},
			{
				Code: `export function foo() { if (a) return 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of function 'foo'.",
					Line:      1,
					Column:    17,
					EndLine:   1,
					EndColumn: 20,
				}},
			},

			// ---- Dimension 4: async / generator / async generator variants ----
			{
				Code: `async function foo() { if (a) return 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of async function 'foo'.",
					Line:      1,
					Column:    16,
					EndLine:   1,
					EndColumn: 19,
				}},
			},
			{
				Code: `function* foo() { if (a) return 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of generator function 'foo'.",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 14,
				}},
			},
			{
				Code: `async function* foo() { if (a) return 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of async generator function 'foo'.",
					Line:      1,
					Column:    17,
					EndLine:   1,
					EndColumn: 20,
				}},
			},
			{
				Code: `const f = async () => { if (a) return 1; };`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of async arrow function.",
					Line:      1,
					Column:    20,
					EndLine:   1,
					EndColumn: 22,
				}},
			},
			{
				Code: `const f = async function () { if (a) return 1; };`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of async function.",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 16,
				}},
			},
			{
				Code: `const f = function* () { if (a) return 1; };`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of generator function.",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 19,
				}},
			},
			{
				Code: `const o = { async foo() { if (a) return 1; } };`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of async method 'foo'.",
					Line:      1,
					Column:    19,
					EndLine:   1,
					EndColumn: 22,
				}},
			},
			{
				Code: `const o = { *foo() { if (a) return 1; } };`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of generator method 'foo'.",
					Line:      1,
					Column:    14,
					EndLine:   1,
					EndColumn: 17,
				}},
			},
			{
				Code: `const o = { async *foo() { if (a) return 1; } };`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of async generator method 'foo'.",
					Line:      1,
					Column:    20,
					EndLine:   1,
					EndColumn: 23,
				}},
			},
			{
				Code: `class C { async *foo() { if (a) return 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of async generator method 'foo'.",
					Line:      1,
					Column:    18,
					EndLine:   1,
					EndColumn: 21,
				}},
			},

			// ---- Dimension 4: same-kind nesting, only the intended function matches ----
			{
				Code: `function outer() { if (a) return 1; function inner() { if (b) return 2; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of function 'outer'.",
					Line:      1,
					Column:    10,
					EndLine:   1,
					EndColumn: 15,
				}, {
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of function 'inner'.",
					Line:      1,
					Column:    46,
					EndLine:   1,
					EndColumn: 51,
				}},
			},
			{
				Code: `function outer() { return 1; function inner() { if (b) return 2; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of function 'inner'.",
					Line:      1,
					Column:    39,
					EndLine:   1,
					EndColumn: 44,
				}},
			},
			{
				Code: `function outer() { if (a) return 1; function inner() { return; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of function 'outer'.",
					Line:      1,
					Column:    10,
					EndLine:   1,
					EndColumn: 15,
				}},
			},
			{
				Code: `class C { foo() { const f = () => { if (a) return 1; }; if (b) return 2; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of method 'foo'.",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 14,
				}, {
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of arrow function.",
					Line:      1,
					Column:    32,
					EndLine:   1,
					EndColumn: 34,
				}},
			},
			{
				Code: `class C { foo() { if (a) return 1; class D { bar() { if (b) return 2; } } } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of method 'foo'.",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 14,
				}, {
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of method 'bar'.",
					Line:      1,
					Column:    46,
					EndLine:   1,
					EndColumn: 49,
				}},
			},
			{
				Code: `class C { static { const f = function () { if (a) return 1; }; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of function.",
					Line:      1,
					Column:    30,
					EndLine:   1,
					EndColumn: 38,
				}},
			},
			{
				Code: `class C { p = { m() { if (a) return 1; } }; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of method 'm'.",
					Line:      1,
					Column:    17,
					EndLine:   1,
					EndColumn: 18,
				}},
			},
			{
				Code: `function foo() { return 1; function g() { return; return 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedReturnValue",
					Message:   "Function 'g' expected no return value.",
					Line:      1,
					Column:    51,
					EndLine:   1,
					EndColumn: 60,
				}},
			},

			// ---- Dimension 4: graceful degradation on body-less and empty forms ----
			{
				Code: `function foo(); function foo() { if (a) return 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of function 'foo'.",
					Line:      1,
					Column:    26,
					EndLine:   1,
					EndColumn: 29,
				}},
			},
			{
				Code: `class C { foo(): void; foo() { if (a) return 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of method 'foo'.",
					Line:      1,
					Column:    24,
					EndLine:   1,
					EndColumn: 27,
				}},
			},
			{
				Code: `function foo({ a, ...rest }) { if (a) return 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of function 'foo'.",
					Line:      1,
					Column:    10,
					EndLine:   1,
					EndColumn: 13,
				}},
			},

			// ---- Locks in upstream checkLastSegment() arm 1: hasReturnValue false leaves the end alone ----
			{
				Code: `function foo() { if (a) return; if (b) return 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedReturnValue",
					Message:   "Function 'foo' expected no return value.",
					Line:      1,
					Column:    40,
					EndLine:   1,
					EndColumn: 49,
				}},
			},

			// ---- Locks in upstream checkLastSegment() arm 2: the end of the code path is unreachable ----
			{
				Code: `function foo() { if (a) return 1; while (true) { break; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of function 'foo'.",
					Line:      1,
					Column:    10,
					EndLine:   1,
					EndColumn: 13,
				}},
			},
			{
				Code: `function foo() { if (a) return 1; for (;;) { if (b) break; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of function 'foo'.",
					Line:      1,
					Column:    10,
					EndLine:   1,
					EndColumn: 13,
				}},
			},
			{
				Code: `function foo() { if (a) return 1; switch (a) { case 1: return 2; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of function 'foo'.",
					Line:      1,
					Column:    10,
					EndLine:   1,
					EndColumn: 13,
				}},
			},
			{
				Code: `function foo() { if (a) return 1; try { } catch (e) { return 2; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of function 'foo'.",
					Line:      1,
					Column:    10,
					EndLine:   1,
					EndColumn: 13,
				}},
			},
			{
				Code: `function foo() { if (a) return 1; try { foo(); } finally { } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of function 'foo'.",
					Line:      1,
					Column:    10,
					EndLine:   1,
					EndColumn: 13,
				}},
			},
			{
				Code: `function foo() { if (a) return 1; label: { break label; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of function 'foo'.",
					Line:      1,
					Column:    10,
					EndLine:   1,
					EndColumn: 13,
				}},
			},
			{
				Code: `function foo() { if (a) return 1; debugger; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of function 'foo'.",
					Line:      1,
					Column:    10,
					EndLine:   1,
					EndColumn: 13,
				}},
			},

			// ---- Locks in upstream checkLastSegment() arm 3: isES5Constructor ----
			{
				Code: `function ábc() { if (a) return 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of function 'ábc'.",
					Line:      1,
					Column:    10,
					EndLine:   1,
					EndColumn: 13,
				}},
			},
			{
				// A name outside the BMP is read as a lone surrogate, whose
				// lowercase form is itself, so it is not an ES5 constructor.
				Code: `function 𐐀() { if (a) return 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of function '𐐀'.",
				}},
			},
			{
				Code: `function $foo() { if (a) return 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of function '$foo'.",
					Line:      1,
					Column:    10,
					EndLine:   1,
					EndColumn: 14,
				}},
			},
			{
				Code: `const Foo = function () { if (a) return 1; };`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of function.",
					Line:      1,
					Column:    13,
					EndLine:   1,
					EndColumn: 21,
				}},
			},
			{
				Code: `class C { Foo() { if (a) return 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of method 'Foo'.",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 14,
				}},
			},

			// ---- Locks in upstream checkLastSegment() arm 4: isClassConstructor ----
			{
				Code: `class C { constructor() { if (a) return 1; return; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturnValue",
					Message:   "Constructor expected a return value.",
					Line:      1,
					Column:    44,
					EndLine:   1,
					EndColumn: 51,
				}},
			},

			// ---- Locks in upstream ReturnStatement() arm: only the first return sets the expectation ----
			{
				Code: `function foo() { if (a) return 1; if (b) return; if (c) return 2; if (d) return; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of function 'foo'.",
					Line:      1,
					Column:    10,
					EndLine:   1,
					EndColumn: 13,
				}, {
					MessageId: "missingReturnValue",
					Message:   "Function 'foo' expected a return value.",
					Line:      1,
					Column:    42,
					EndLine:   1,
					EndColumn: 49,
				}, {
					MessageId: "missingReturnValue",
					Message:   "Function 'foo' expected a return value.",
					Line:      1,
					Column:    74,
					EndLine:   1,
					EndColumn: 81,
				}},
			},
			{
				Code: `function foo() { return; if (a) return 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedReturnValue",
					Message:   "Function 'foo' expected no return value.",
					Line:      1,
					Column:    33,
					EndLine:   1,
					EndColumn: 42,
				}},
			},
			{
				Code: `function foo() { if (a) return 1; else return 2; return; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturnValue",
					Message:   "Function 'foo' expected a return value.",
					Line:      1,
					Column:    50,
					EndLine:   1,
					EndColumn: 57,
				}},
			},

			// ---- Locks in upstream ReturnStatement() arm: treatUndefinedAsUnspecified off by default ----
			{
				Code: `function foo() { if (a) return undefined; return; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturnValue",
					Message:   "Function 'foo' expected a return value.",
					Line:      1,
					Column:    43,
					EndLine:   1,
					EndColumn: 50,
				}},
			},
			{
				Code: `function foo() { if (a) return void 0; return; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturnValue",
					Message:   "Function 'foo' expected a return value.",
					Line:      1,
					Column:    40,
					EndLine:   1,
					EndColumn: 47,
				}},
			},
			{
				Code:    `function foo() { if (a) return undefined; return; }`,
				Options: map[string]interface{}{"treatUndefinedAsUnspecified": false},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturnValue",
					Message:   "Function 'foo' expected a return value.",
					Line:      1,
					Column:    43,
					EndLine:   1,
					EndColumn: 50,
				}},
			},

			// ---- Locks in upstream ReturnStatement() arm: `void` is read off the top-level operator only ----
			{
				Code:    `function foo() { if (a) return -void 0; return; }`,
				Options: map[string]interface{}{"treatUndefinedAsUnspecified": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturnValue",
					Message:   "Function 'foo' expected a return value.",
					Line:      1,
					Column:    41,
					EndLine:   1,
					EndColumn: 48,
				}},
			},
			{
				Code:    `function foo() { if (a) return typeof undefined; return; }`,
				Options: map[string]interface{}{"treatUndefinedAsUnspecified": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturnValue",
					Message:   "Function 'foo' expected a return value.",
					Line:      1,
					Column:    50,
					EndLine:   1,
					EndColumn: 57,
				}},
			},

			// ---- Locks in upstream ReturnStatement() arm: isSpecificId is syntactic, not a lookup ----
			{
				Code:    `function foo() { if (a) return undefined2; return; }`,
				Options: map[string]interface{}{"treatUndefinedAsUnspecified": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturnValue",
					Message:   "Function 'foo' expected a return value.",
					Line:      1,
					Column:    44,
					EndLine:   1,
					EndColumn: 51,
				}},
			},

			// ---- Locks in upstream ReturnStatement() arm: the Program code path names itself `Program` ----
			{
				Code: `if (a) { return 1; } return;`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturnValue",
					Message:   "Program expected a return value.",
					Line:      1,
					Column:    22,
					EndLine:   1,
					EndColumn: 29,
				}},
			},
			{
				Code: `if (a) { return; } return 1;`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedReturnValue",
					Message:   "Program expected no return value.",
					Line:      1,
					Column:    20,
					EndLine:   1,
					EndColumn: 29,
				}},
			},

			// ---- Real-user: eslint#8179 void inside a ternary is still a value ----
			{
				Code:    `function test(input) { if (input) { return input === true ? void 0 : void 1; } return void false; }`,
				Options: map[string]interface{}{"treatUndefinedAsUnspecified": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturnValue",
					Message:   "Function 'test' expected a return value.",
					Line:      1,
					Column:    80,
					EndLine:   1,
					EndColumn: 98,
				}},
			},

			// ---- Real-user: eslint#12865 an async function returning a promise on one path only ----
			{
				Code: `async function example() { if (something) { someSyncSideEffect(); return; } else { return new Promise(resolve => { someAsyncSideEffect(resolve); }); } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedReturnValue",
					Message:   "Async function 'example' expected no return value.",
					Line:      1,
					Column:    84,
					EndLine:   1,
					EndColumn: 149,
				}},
			},

			// ---- Real-user: eslint#11371 a generator with a bare return alongside a value return ----
			{
				Code: `function* gen() { yield 1; if (a) return 1; return; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturnValue",
					Message:   "Generator function 'gen' expected a return value.",
					Line:      1,
					Column:    45,
					EndLine:   1,
					EndColumn: 52,
				}},
			},

			// ---- Real-user: express-style middleware that guards and falls through ----
			{
				Code: `function middleware(req, res, next) { if (!req.user) return res.status(401).send(); next(); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of function 'middleware'.",
					Line:      1,
					Column:    10,
					EndLine:   1,
					EndColumn: 20,
				}},
			},

			// ---- Real-user: a helper that ends by exiting the process still falls off the end ----
			{
				Code: `function foo() { if (a) return 1; process.exit(1); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of function 'foo'.",
					Line:      1,
					Column:    10,
					EndLine:   1,
					EndColumn: 13,
				}},
			},

			// ---- Options plumbing: every JSON shape a config can hand the rule ----
			// `treatUndefinedAsUnspecified` defaults to false, so an omitted
			// option, an empty option object, and an explicit false all report
			// the same diagnostic.
			{
				Code:    `function foo() { if (a) return undefined; return; }`,
				Options: map[string]interface{}{"treatUndefinedAsUnspecified": false},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturnValue",
					Message:   "Function 'foo' expected a return value.",
					Line:      1,
					Column:    43,
					EndLine:   1,
					EndColumn: 50,
				}},
			},
			{
				Code:    `function foo() { if (a) return undefined; return; }`,
				Options: []interface{}{map[string]interface{}{"treatUndefinedAsUnspecified": false}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturnValue",
					Message:   "Function 'foo' expected a return value.",
					Line:      1,
					Column:    43,
					EndLine:   1,
					EndColumn: 50,
				}},
			},
			{
				Code:    `function foo() { if (a) return undefined; return; }`,
				Options: map[string]interface{}{},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturnValue",
					Message:   "Function 'foo' expected a return value.",
					Line:      1,
					Column:    43,
					EndLine:   1,
					EndColumn: 50,
				}},
			},
			{
				Code: `function foo() { if (a) return undefined; return; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturnValue",
					Message:   "Function 'foo' expected a return value.",
					Line:      1,
					Column:    43,
					EndLine:   1,
					EndColumn: 50,
				}},
			},

			// ---- Dimension 4: positions on multi-line sources ----
			{
				Code: `function foo(a) {
  if (a) {
    return 1;
  }
}
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of function 'foo'.",
					Line:      1,
					Column:    10,
					EndLine:   1,
					EndColumn: 13,
				}},
			},
			{
				Code: `class C {
  bar() {
    if (a) {
      return 1;
    }
    return;
  }
}
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturnValue",
					Message:   "Method 'bar' expected a return value.",
					Line:      6,
					Column:    5,
					EndLine:   6,
					EndColumn: 12,
				}},
			},
			{
				Code: `const o = {
  get thing() {
    if (a) {
      return 1;
    }
  },
};
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of getter 'thing'.",
					Line:      2,
					Column:    12,
					EndLine:   2,
					EndColumn: 13,
				}},
			},
			{
				Code: `const f = (
  a,
  b,
) => {
  if (a) return 1;
};
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of arrow function.",
					Line:      4,
					Column:    3,
					EndLine:   4,
					EndColumn: 5,
				}},
			},
			{
				Code: `function outer() {
  if (a) return 1;
  const inner = function () {
    if (b) return 2;
  };
}
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of function 'outer'.",
					Line:      1,
					Column:    10,
					EndLine:   1,
					EndColumn: 15,
				}, {
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of function.",
					Line:      3,
					Column:    17,
					EndLine:   3,
					EndColumn: 25,
				}},
			},
		},
	)
}
