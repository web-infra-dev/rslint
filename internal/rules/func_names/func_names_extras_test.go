// TestFuncNamesExtras locks in branches and edge shapes that the upstream
// test suite doesn't exercise. Each case carries an inline comment pointing
// at the specific branch / Dimension 4 row / tsgo AST quirk it covers, so
// future refactors can't silently regress them without breaking a named
// lock-in.
package func_names

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestFuncNamesExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&FuncNamesRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: parenthesized receiver, single level ----
			// ESTree has no ParenthesizedExpression node, so ESLint's
			// hasInferredName sees straight through `(function(){})` to the
			// VariableDeclarator; tsgo must walk up through the explicit
			// ParenthesizedExpression node to match.
			{Code: "var foo = (function(){});", Options: []any{"as-needed"}},

			// ---- Dimension 4: parenthesized receiver, multi level ----
			{Code: "var foo = ((function(){}));", Options: []any{"as-needed"}},

			// ---- Dimension 4: parenthesized receiver through a compound
			// assignment operator (also covers the "quux ??= function(){}"
			// as-needed doc example directly, unparenthesized) ----
			{Code: "quux ??= function() {};", Options: []any{"as-needed"}},
			{Code: "quux ||= function() {};", Options: []any{"as-needed"}},
			{Code: "quux += (function() {});", Options: []any{"as-needed"}},

			// ---- Dimension 4: optional chain — N/A, func-names never
			// inspects a member/call expression's own object/callee shape;
			// nothing in its logic branches on `?.`. ----

			// ---- Dimension 4: string-literal key inferred name (as-needed) ----
			{Code: "({\"foo\": function(){}});", Options: []any{"as-needed"}},

			// ---- Real-user: https://github.com/eslint/eslint/issues/1699
			// "func-names incorrectly errors on get/set/shorthand methods" —
			// tsgo parses all three directly as MethodDeclaration /
			// GetAccessor / SetAccessor, never as a FunctionExpression whose
			// parent is a Property, so they're structurally exempt from this
			// rule's listeners regardless of mode. ----
			{Code: "var obj = { foo() {}, get bar() { return 1; }, set baz(v) {} };", Options: []any{"always"}},
			{Code: "class C { foo() {} get bar() { return 1; } set baz(v) {} }", Options: []any{"never"}},

			// ---- Real-user: https://github.com/eslint/eslint/issues/6616
			// "allow names for recursive use" (never) — nested distinct
			// recursive named function expressions, each guarded by its own
			// self-reference, including the outer one which calls itself too. ----
			{Code: "var a = function foo() { foo(); return function bar() { return bar(); }; };", Options: []any{"never"}},

			// ---- Dimension 4: computed key with a string-literal expression
			// (`['foo']`, as opposed to the bare string-literal-key form
			// `{"foo": ...}` already covered above) — GetStaticPropertyName
			// must unwrap the ComputedPropertyName to statically resolve it. ----
			{Code: "({['foo']: function(){}});", Options: []any{"as-needed"}},

			// ---- Dimension 4: class field with a definite-assignment `!`
			// postfix token on the property name — PostfixToken sits beside
			// Initializer, not between it and the FunctionExpression, so it
			// must not interfere with the PropertyDeclaration match. ----
			{Code: "class C { foo!: () => void = function() {}; }", Options: []any{"as-needed"}},

			// ---- Locks in upstream handleFunction() arm: the recursion
			// guard's reference lookup is scope-correct across a *non*-
			// function boundary — a nested arrow function body still resolves
			// `foo()` back to the enclosing named function expression (arrows
			// don't introduce a new lexical binding for it). ----
			{Code: "var a = function foo() { var g = () => foo(); return g; };", Options: []any{"never"}},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: numeric-literal key naming, "never" mode ----
			{
				Code:    "({0: function zero() {}});",
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "named", Message: "Unexpected named method '0'.", Line: 1, Column: 3, EndLine: 1, EndColumn: 19},
				},
			},

			// ---- Dimension 4: TS type-expression wrapper does not carry
			// inferred-name semantics through to the assignment target —
			// unlike a bare ParenthesizedExpression, `as`/`satisfies` are
			// real intermediate nodes upstream's ESTree-based check has no
			// equivalent concept for either, so this is a straightforward
			// "still needs a name" case under as-needed. ----
			{
				Code:    "const foo = (function(){}) as Function;",
				Options: []any{"as-needed"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed function.", Line: 1, Column: 14, EndLine: 1, EndColumn: 22},
				},
			},

			// ---- Dimension 4: computed non-literal key, always mode —
			// unnamed and un-quoted ("method", no property name available) ----
			{
				Code:    "({[x]: function() {}});",
				Options: []any{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed method.", Line: 1, Column: 3, EndLine: 1, EndColumn: 16},
				},
			},

			// ---- Dimension 4: static class field — "static method 'foo'" ----
			{
				Code:    "class C { static foo = function() {}; }",
				Options: []any{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed static method 'foo'.", Line: 1, Column: 11, EndLine: 1, EndColumn: 32},
				},
			},

			// ---- Dimension 4: async function expression — never exercised
			// by upstream's own test suite at all. ----
			{
				Code:    "var a = async function() {};",
				Options: []any{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed async function.", Line: 1, Column: 9, EndLine: 1, EndColumn: 23},
				},
			},

			// ---- Dimension 4: async generator, generators:"never" — combines
			// async + generator + named tokens in one message ----
			{
				Code:    "var a = bar(async function *baz() {});",
				Options: []any{"never", map[string]any{"generators": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "named", Message: "Unexpected named async generator function 'baz'.", Line: 1, Column: 13, EndLine: 1, EndColumn: 32},
				},
			},

			// ---- Locks in upstream handleFunction() arm: recursion guard
			// only shields the innermost binding — an outer named function
			// expression whose own name is fully shadowed (every call inside
			// resolves to an inner re-declaration of the same name) still
			// gets reported by "never", since upstream's recursion check
			// keys off eslint-scope references to *that* declaration only. ----
			{
				Code:    "var a = function outer() { var inner = function outer() { return outer(); }; return inner; };",
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "named", Message: "Unexpected named function 'outer'.", Line: 1, Column: 9, EndLine: 1, EndColumn: 23},
				},
			},

			// ---- Locks in upstream handleFunction() arm: nested function
			// expressions are checked independently — an outer named
			// function does not exempt an unnamed inner one from "always". ----
			{
				Code:    "var a = function foo() { return function() {}; };",
				Options: []any{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed function.", Line: 1, Column: 33, EndLine: 1, EndColumn: 41},
				},
			},

			// ---- Locks in upstream getConfigForNode() arm: `generators`
			// left unset (empty string sentinel) falls back to the base
			// mode for a generator, exactly like a non-generator function. ----
			{
				Code:    "var foo = bar(function *() {});",
				Options: []any{"always", map[string]any{}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed generator function.", Line: 1, Column: 15, EndLine: 1, EndColumn: 25},
				},
			},

			// ---- Dimension 4: multi-line source — the property key and the
			// reported unnamed function sit on line 2, proving line tracking
			// through GetFunctionHeadLoc isn't hardcoded to line 1. ----
			{
				Code:    "var obj = {\n  foo: function() {}\n};",
				Options: []any{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed method 'foo'.", Line: 2, Column: 3, EndLine: 2, EndColumn: 16},
				},
			},

			// ---- Dimension 4: TS non-null assertion wrapper — like `as` /
			// `satisfies`, `!` is a real intermediate node (NonNullExpression)
			// that WalkUpParenthesizedExpressions does not see through, so it
			// breaks the inferred-name walk the same way. ----
			{
				Code:    "var foo = (function(){})!;",
				Options: []any{"as-needed"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed function.", Line: 1, Column: 12, EndLine: 1, EndColumn: 20},
				},
			},

			// ---- Locks in upstream getConfigForNode() arm: the `generators`
			// override also applies to a FunctionDeclaration (the
			// `export default function*` form), not just FunctionExpression —
			// GetFunctionFlags must read the asterisk off both node kinds. ----
			{
				Code:    "export default function*() {}",
				Options: []any{"as-needed", map[string]any{"generators": "as-needed"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed generator function.", Line: 1, Column: 16, EndLine: 1, EndColumn: 25},
				},
			},

			// ---- Locks in upstream handleFunction() arm: the recursion
			// guard's reference lookup is scope-correct — an arrow function's
			// *own* parameter can shadow the enclosing named function
			// expression's recursion binding, same as a nested declaration
			// would, leaving the outer name unreferenced and reportable. ----
			{
				Code:    "var a = function foo() { var g = (foo) => foo(); return g; };",
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "named", Message: "Unexpected named function 'foo'.", Line: 1, Column: 9, EndLine: 1, EndColumn: 21},
				},
			},
		},
	)
}
