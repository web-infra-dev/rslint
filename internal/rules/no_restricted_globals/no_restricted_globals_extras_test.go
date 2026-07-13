package no_restricted_globals

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoRestrictedGlobalsExtras covers branches and edge cases that the
// upstream ESLint test suite doesn't exercise. Upstream migration lives in
// no_restricted_globals_upstream_test.go.
func TestNoRestrictedGlobalsExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoRestrictedGlobalsRule,
		[]rule_tester.ValidTestCase{
			// ESLint's own astUtils.isSpecificMemberAccess / getStaticPropertyName
			// don't see through TSNonNullExpression or TSAsExpression (they aren't
			// MemberExpressions), so these aren't reported either.
			{
				Code:    `window!.foo()`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"foo"}, "checkGlobalObject": true}},
			},
			{
				Code:    `(window as any).foo()`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"foo"}, "checkGlobalObject": true}},
			},

			// A computed member access whose key can't be resolved statically must
			// not report, even when the runtime value could coincide with a
			// restricted name.
			{
				Code:    `let key = "foo"; window[key]();`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"foo"}, "checkGlobalObject": true}},
			},

			// var hoists to the enclosing function only, not to nested functions.
			{
				Code:    `function outer() { var foo; function inner() { foo; } }`,
				Options: []interface{}{"foo"},
			},

			// A rest element in a binding pattern declares its name, it doesn't reference it.
			{
				Code:    `const {...foo} = obj;`,
				Options: []interface{}{"foo"},
			},

			// Parameter shadowing in a forEach/arrow callback (the rule's canonical `event` motivation).
			{
				Code:    `elements.forEach((event) => { doSomething(event); });`,
				Options: []interface{}{"event"},
			},
			// Common restricted names collide constantly with plain object keys.
			{
				Code:    `const config = { name: 'foo', self: 1 };`,
				Options: []interface{}{"name", "self"},
			},
			// Enum members commonly collide with short/common restricted names.
			{
				Code:    `enum Test { foo, bar }`,
				Options: []interface{}{"foo"},
			},

			// No restricted globals configured is a no-op.
			{
				Code:    `foo`,
				Options: []interface{}{},
			},

			// The original (module export) name in an import specifier is not a scope reference.
			{
				Code:    `import { foo as bar } from 'mod'; bar;`,
				Options: []interface{}{"foo"},
			},
			// The original name in a re-export alias is not a scope reference (only real for `export ... from`).
			{
				Code:    `export { foo as bar } from 'other';`,
				Options: []interface{}{"foo"},
			},

			// Bare (unwrapped) string option, matching the CLI single-option shape.
			{Code: `bar`, Options: "foo"},
			// Bare (unwrapped) object option, matching the CLI single-option shape.
			{Code: `window.bar()`, Options: map[string]interface{}{"globals": []interface{}{"foo"}, "checkGlobalObject": true}},
			// nil options is a safe no-op.
			{Code: `foo`, Options: nil},

			// All case/default clauses of a switch share one CaseBlock scope, so a
			// `let` in a later clause shadows a reference to the same name in an
			// earlier clause.
			{
				Code:    `switch (x) { case 0: foo; break; case 1: let foo; }`,
				Options: []interface{}{"foo"},
			},

			// A lowercase/hyphenated JSX tag name (`<foo />`) names an intrinsic
			// element, not a variable reference — it must not be checked even when
			// "foo" is restricted.
			{
				Code:    `const element = <foo />;`,
				Options: []interface{}{"foo"},
				Tsx:     true,
			},
			{
				Code:    `const element = <foo-bar></foo-bar>;`,
				Options: []interface{}{"foo-bar"},
				Tsx:     true,
			},

			// An empty-string entry in `globals` is not a valid identifier and must
			// be filtered out rather than restricting the empty property name.
			{
				Code:    `window[""];`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{""}, "checkGlobalObject": true}},
			},
		},
		[]rule_tester.InvalidTestCase{
			// Parenthesized global-object receiver.
			{
				Code:    `(window).foo()`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"foo"}, "checkGlobalObject": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 10, EndLine: 1, EndColumn: 13}},
			},
			{
				Code:    `((window)).foo()`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"foo"}, "checkGlobalObject": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 12, EndLine: 1, EndColumn: 15}},
			},

			// Numeric / template-literal computed keys.
			{
				Code:    `window[0]()`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"0"}, "checkGlobalObject": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 8, EndLine: 1, EndColumn: 9}},
			},
			{
				Code:    "window[`foo`]()",
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"foo"}, "checkGlobalObject": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 8, EndLine: 1, EndColumn: 13}},
			},

			// ESLint's astUtils.getStaticStringValue resolves `null`/`true`/`false`
			// literals (via node.value) and BigInt literals (via node.bigint) to
			// their string form, so `window[null]` is equivalent to `window["null"]`.
			{
				Code:    `window[null]()`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"null"}, "checkGlobalObject": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 8, EndLine: 1, EndColumn: 12}},
			},
			{
				Code:    `window[true]()`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"true"}, "checkGlobalObject": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 8, EndLine: 1, EndColumn: 12}},
			},
			{
				Code:    `window[false]()`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"false"}, "checkGlobalObject": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 8, EndLine: 1, EndColumn: 13}},
			},
			{
				Code:    `window[123n]()`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"123"}, "checkGlobalObject": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 8, EndLine: 1, EndColumn: 12}},
			},

			// A reference outside every declaring function is not shadowed.
			{
				Code:    `function outer() { function inner() { var foo; } } foo;`,
				Options: []interface{}{"foo"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 52, EndLine: 1, EndColumn: 55}},
			},

			// Spread in an object literal reads the value.
			{
				Code:    `const obj = {...foo};`,
				Options: []interface{}{"foo"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 17, EndLine: 1, EndColumn: 20}},
			},

			// isInTypeContext's ExpressionWithTypeArguments case excludes class
			// `extends` (a value context) but not `implements` / `interface
			// extends` — upstream itself never tests the `extends` side of this
			// distinction.
			{
				Code:    `class Derived extends Test {}`,
				Options: []interface{}{"Test"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 23, EndLine: 1, EndColumn: 27}},
			},

			// Last entry wins when the same name is restricted twice.
			{
				Code:    `foo`,
				Options: []interface{}{"foo", map[string]interface{}{"name": "foo", "message": "second wins"}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "customMessage", Message: "Unexpected use of 'foo'. second wins", Line: 1, Column: 1, EndLine: 1, EndColumn: 4}},
			},

			// rslint doesn't model ESLint's environment/languageOptions.globals, so
			// globalThis/self/window (and configured globalObjects) are always
			// recognized as global-object roots when checkGlobalObject is enabled,
			// regardless of any declared environment. See the "SKIP" cases in the
			// upstream file's valid list and the rule doc's "Differences from
			// ESLint" section.
			{
				Code:    `window.foo()`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"foo"}, "checkGlobalObject": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 8, EndLine: 1, EndColumn: 11}},
			},
			{
				Code:    `self.foo()`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"foo"}, "checkGlobalObject": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 6, EndLine: 1, EndColumn: 9}},
			},
			{
				Code:    `globalThis.foo()`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"foo"}, "checkGlobalObject": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 12, EndLine: 1, EndColumn: 15}},
			},

			// Bare (unwrapped) string option, matching the CLI single-option shape.
			{
				Code:    `foo`,
				Options: "foo",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 1, EndLine: 1, EndColumn: 4}},
			},
			// Bare (unwrapped) object option, matching the CLI single-option shape.
			{
				Code:    `window.foo()`,
				Options: map[string]interface{}{"globals": []interface{}{"foo"}, "checkGlobalObject": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 8, EndLine: 1, EndColumn: 11}},
			},

			// A default parameter value is evaluated in the parameter environment,
			// which is a *parent* of the function body's variable environment — a
			// `var` in the body does not shadow a reference in a default value.
			{
				Code:    `function f(a = foo) { var foo; }`,
				Options: []interface{}{"foo"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 16, EndLine: 1, EndColumn: 19}},
			},

			// A capitalized JSX tag name is a real reference to a component
			// variable and must still be checked.
			{
				Code:    `const el = <Foo />;`,
				Options: []interface{}{"Foo"},
				Tsx:     true,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 13, EndLine: 1, EndColumn: 16}},
			},

			// BigInt computed keys beyond int64 range must still normalize to
			// their full decimal string, not silently degrade to raw literal text.
			{
				Code:    `window[0x10000000000000000n]()`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"18446744073709551616"}, "checkGlobalObject": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 8, EndLine: 1, EndColumn: 28}},
			},
		},
	)
}
