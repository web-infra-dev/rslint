package max_statements

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestMaxStatementsExtras locks in branches and edge shapes that the upstream
// test suite doesn't exercise. Each case carries an inline comment pointing at
// the specific branch / Dimension 4 row / tsgo AST quirk it covers, so future
// refactors can't silently regress them without breaking a named lock-in.
//
// N/A: Dimension 4 "receiver / expression wrappers" (parenthesized receiver,
// `!`/`as`/`satisfies`, optional chain) doesn't apply — this rule never
// inspects an expression receiver, only block-statement counts and
// function-like node kinds.
func TestMaxStatementsExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&MaxStatementsRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: overload signature (body-absent FunctionDeclaration) has no body → not pushed/popped ----
			{Code: "function foo(x: number): void;\nfunction foo(x: string): void;\nfunction foo(x: any) { 1; 2; }", Options: option(2)},

			// ---- Dimension 4: abstract method (body-absent MethodDeclaration) is skipped ----
			{Code: "abstract class C {\n  abstract foo(): void;\n}", Options: option(map[string]interface{}{"max": 0})},

			// ---- Dimension 4: empty function body never exceeds ----
			{Code: "function foo() {}", Options: option(0)},

			// ---- Dimension 4: empty static block never exceeds and never reports ----
			{Code: "class C { static {} }", Options: option(0)},

			// ---- Dimension 1: TS namespace/module body (tsgo KindModuleBlock, distinct
			// from KindBlock) is invisible to statement counting, matching upstream where
			// TSModuleBlock isn't a BlockStatement either — its statements don't leak into
			// the enclosing function's count ----
			{Code: "function foo() { namespace N { one; two; three; four; five; six; seven; eight; nine; ten; } }", Options: option(1)},

			// Locks in upstream reportIfTooManyStatements() arm: `option.maximum || option.max`
			// when maximum is explicitly 0 (falsy) falls through to `max`, not "no limit".
			{Code: "function foo() { 1; 2; 3; }", Options: option(map[string]interface{}{"maximum": 0, "max": 3})},

			// Locks in the reverse: `maximum: 0` with no `max` key leaves maxStatements
			// unresolved, which never exceeds — effectively unlimited.
			{Code: "function foo() { 1; 2; 3; 4; 5; 6; 7; 8; 9; 10; 11; 12; }", Options: option(map[string]interface{}{"maximum": 0})},

			// Locks in countStatements() when functionStack is empty: a block statement
			// that isn't nested in any function must not crash and must not report.
			{Code: "{ one; two; three; four; five; }", Options: option(1)},

			// ---- Dimension 1: a decorated member is a single top-level function
			// (the decorator holds no function of its own), so it stays exempt ----
			{Code: "class C { @dec(1) get foo() { one; two; } }", Options: optionsWithTopLevel(1, true)},

			// Locks in endFunction(): overload signatures never enter topLevelFunctions,
			// so the lone implementation is still "the" top-level function and is
			// exempted by the topLevelFunctions.length === 1 check, despite exceeding max.
			{Code: "function foo(x: number): void;\nfunction foo(x: string): void;\nfunction foo(x: any) { 1; 2; 3; 4; 5; }", Options: optionsWithTopLevel(1, true)},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: object literal method, string-literal key ----
			{
				Code:    `var foo = { "th ing": function() { 1; 2; 3; } }`,
				Options: option(2),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Method 'th ing'", 3, 2), Line: 1, Column: 13, EndLine: 1, EndColumn: 31}},
			},
			// ---- Dimension 4: object literal method, numeric-literal key ----
			{
				Code:    "var foo = { 0: function() { 1; 2; 3; } }",
				Options: option(2),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Method '0'", 3, 2), Line: 1, Column: 13, EndLine: 1, EndColumn: 24}},
			},
			// ---- Dimension 4: object literal method, computed key that isn't statically resolvable → unnamed "Method" ----
			{
				Code:    "var foo = { [Symbol.iterator]: function() { 1; 2; 3; } }",
				Options: option(2),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Method", 3, 2), Line: 1, Column: 13, EndLine: 1, EndColumn: 40}},
			},

			// ---- Dimension 1: class private method (tsgo KindMethodDeclaration; ESTree wraps this in MethodDefinition) ----
			{
				Code:    "class C { #foo() { 1; 2; 3; } }",
				Options: option(2),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Private method #foo", 3, 2), Line: 1, Column: 11, EndLine: 1, EndColumn: 15}},
			},
			// ---- Dimension 1: class static method ----
			{
				Code:    "class C { static foo() { 1; 2; 3; } }",
				Options: option(2),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Static method 'foo'", 3, 2), Line: 1, Column: 11, EndLine: 1, EndColumn: 21}},
			},
			// ---- Dimension 1: class getter (tsgo KindGetAccessor) ----
			{
				Code:    "class C { get foo() { 1; 2; 3; } }",
				Options: option(2),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Getter 'foo'", 3, 2), Line: 1, Column: 11, EndLine: 1, EndColumn: 18}},
			},
			// ---- Dimension 1: class setter (tsgo KindSetAccessor) ----
			{
				Code:    "class C { set foo(v) { 1; 2; 3; } }",
				Options: option(2),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Setter 'foo'", 3, 2), Line: 1, Column: 11, EndLine: 1, EndColumn: 18}},
			},
			// ---- Dimension 1: class constructor (tsgo KindConstructor) ----
			{
				Code:    "class C { constructor() { 1; 2; 3; } }",
				Options: option(2),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Constructor", 3, 2), Line: 1, Column: 11, EndLine: 1, EndColumn: 22}},
			},
			// ---- Dimension 1: async generator method (flag composition) ----
			{
				Code:    "class C { async *foo() { 1; 2; 3; } }",
				Options: option(2),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Async generator method 'foo'", 3, 2), Line: 1, Column: 11, EndLine: 1, EndColumn: 21}},
			},
			// ---- Dimension 1: class field arrow-function initializer (tsgo has no MethodDefinition/PropertyDefinition wrapper) ----
			{
				Code:    "class C { method = () => { 1; 2; 3; } }",
				Options: option(2),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Method 'method'", 3, 2), Line: 1, Column: 11, EndLine: 1, EndColumn: 20}},
			},

			// ---- Dimension 2: 3+ level nesting (function > method > arrow) — only the innermost violates, counts don't leak to ancestors ----
			{
				Code: `function outer() {
  const obj = {
    method() {
      const inner = () => { 1; 2; 3; };
    },
  };
}`,
				Options: option(2),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Arrow function", 3, 2), Line: 4, Column: 24, EndLine: 4, EndColumn: 26}},
			},

			// ---- Dimension 4: abstract method is skipped; sibling concrete method still checked ----
			{
				Code:    "abstract class C {\n  abstract foo(): void;\n  bar() { 1; 2; 3; }\n}",
				Options: option(2),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Method 'bar'", 3, 2), Line: 3, Column: 3, EndLine: 3, EndColumn: 6}},
			},

			// Locks in reportIfTooManyStatements() called from endFunction()'s object-option arm.
			{
				Code:    "function foo() { 1; 2; 3; 4; }",
				Options: option(map[string]interface{}{"maximum": 0, "max": 3}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Function 'foo'", 4, 3), Line: 1, Column: 1, EndLine: 1, EndColumn: 13}},
			},

			// Locks in "Program:exit" forEach() when topLevelFunctions.length !== 1 (3 top-level
			// functions, none of them exempted — proves the check is strict equality, not "> 0").
			{
				Code:    "function a() { 1; 2; }\nfunction b() { 1; 2; }\nfunction c() { 1; 2; }",
				Options: optionsWithTopLevel(1, true),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Message: exceedMessage("Function 'a'", 2, 1), Line: 1, Column: 1, EndLine: 1, EndColumn: 11},
					{MessageId: "exceed", Message: exceedMessage("Function 'b'", 2, 1), Line: 2, Column: 1, EndLine: 2, EndColumn: 11},
					{MessageId: "exceed", Message: exceedMessage("Function 'c'", 2, 1), Line: 3, Column: 1, EndLine: 3, EndColumn: 11},
				},
			},

			// Locks in the `node.Body() == nil` guard: overload signatures are skipped
			// entirely, only the implementation's own count is reported.
			{
				Code:    "function foo(x: number): void;\nfunction foo(x: string): void;\nfunction foo(x: any) { 1; 2; 3; }",
				Options: option(2),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Function 'foo'", 3, 2), Line: 3, Column: 1, EndLine: 3, EndColumn: 13}},
			},

			// ---- Dimension 1: a function in a member's computed name is a sibling of the
			// member in ESTree, so both count as top-level and neither is exempted ----
			{
				Code:    "class C { [(() => {})()]() { one; two; } }",
				Options: optionsWithTopLevel(1, true),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Method", 2, 1), Line: 1, Column: 11, EndLine: 1, EndColumn: 25}},
			},
			// ---- Dimension 1: same for an object-literal method's computed key ----
			{
				Code:    "var o = { [(() => {})()]() { one; two; } };",
				Options: optionsWithTopLevel(1, true),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Method", 2, 1), Line: 1, Column: 11, EndLine: 1, EndColumn: 25}},
			},
			// ---- Dimension 1: a function in a decorator is likewise a sibling of the member ----
			{
				Code:    "class C { @dec(() => {}) foo() { one; two; } }",
				Options: optionsWithTopLevel(1, true),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Method 'foo'", 2, 1), Line: 1, Column: 11, EndLine: 1, EndColumn: 29}},
			},
			// ---- Dimension 2: statements in a decorator's function belong to that function
			// only — the member's own count is unaffected by hiding its frame. Listed in
			// emission order (inner function exits first); the API and CLI sort a file's
			// diagnostics by start position before handing them out. ----
			{
				Code:    "class C { @dec(function () { a; b; }) foo() { one; two; } }",
				Options: option(1),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exceed", Message: exceedMessage("Function", 2, 1), Line: 1, Column: 16, EndLine: 1, EndColumn: 25},
					{MessageId: "exceed", Message: exceedMessage("Method 'foo'", 2, 1), Line: 1, Column: 11, EndLine: 1, EndColumn: 42},
				},
			},
			// ---- Dimension 2: a computed name's function keeps its own count, and it is
			// nested when the enclosing class sits inside a function ----
			{
				Code:    "function outer() { class C { [(() => { a; b; })()]() { one; } } }",
				Options: option(1),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Arrow function", 2, 1), Line: 1, Column: 35, EndLine: 1, EndColumn: 37}},
			},
			// ---- Dimension 2: a parameter default *is* inside the member's function in
			// ESTree, so it stays nested and is reported directly rather than deferred ----
			{
				Code:    "class C { m(a = () => { x; y; }) { one; } }",
				Options: optionsWithTopLevel(1, true),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Arrow function", 2, 1), Line: 1, Column: 20, EndLine: 1, EndColumn: 22}},
			},

			// ---- Dimension 3: core ESLint's getFunctionHeadLoc starts a class member's
			// report at the member itself, decorators included (unlike the
			// typescript-eslint variant the shared helper mirrors) ----
			{
				Code:    "class C { @dec(1) foo() { one; two; } }",
				Options: option(1),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Method 'foo'", 2, 1), Line: 1, Column: 11, EndLine: 1, EndColumn: 22}},
			},
			{
				Code:    "class C { @dec(1) get foo() { one; two; } }",
				Options: option(1),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Getter 'foo'", 2, 1), Line: 1, Column: 11, EndLine: 1, EndColumn: 26}},
			},
			{
				Code:    "class C { @dec(1)\n  @dec2\n  static async foo() { one; two; } }",
				Options: option(1),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Static async method 'foo'", 2, 1), Line: 1, Column: 11, EndLine: 3, EndColumn: 19}},
			},
			// ---- Dimension 3: the same start applies to a decorated field's initializer,
			// whose ESTree parent is the PropertyDefinition ----
			{
				Code:    "class C { @dec(1) foo = function () { one; two; } }",
				Options: option(1),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Method 'foo'", 2, 1), Line: 1, Column: 11, EndLine: 1, EndColumn: 34}},
			},
			{
				Code:    "class C { @dec(1) foo = () => { one; two; } }",
				Options: option(1),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Method 'foo'", 2, 1), Line: 1, Column: 11, EndLine: 1, EndColumn: 25}},
			},

			// ---- Real-user: eslint/eslint#12950 — ignoreTopLevelFunctions only exempts a
			// file with exactly one top-level function; a sibling IIFE keeps `foo` from
			// being exempted even though the option is enabled ----
			{
				Code: `(function bar () {
  console.log("hello");
})();

function foo() {
  const foo1 = 1;
  const foo2 = 2;
}`,
				Options: optionsWithTopLevel(1, true),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Function 'foo'", 2, 1), Line: 5, Column: 1, EndLine: 5, EndColumn: 13}},
			},

			// ---- Real-user: eslint/eslint#12373 — UMD-style wrapper where the factory is
			// passed as a call argument to the outer IIFE rather than nested in its body;
			// both wrapper and factory are "top-level", so the factory isn't exempted ----
			{
				Code: `(function (root, factory) {
  if (typeof define === "function") define(factory);
  else root.foo = factory();
})(this, function () {
  var a = 1;
  var b = 2;
  var c = 3;
});`,
				Options: optionsWithTopLevel(2, true),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exceed", Message: exceedMessage("Function", 3, 2), Line: 4, Column: 10, EndLine: 4, EndColumn: 19}},
			},
		},
	)
}
