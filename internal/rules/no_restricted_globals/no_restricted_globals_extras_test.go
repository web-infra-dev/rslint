package no_restricted_globals

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoRestrictedGlobalsExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it
// covers, so future refactors can't silently regress them without breaking a
// named lock-in. Upstream migration lives in
// no_restricted_globals_upstream_test.go.
func TestNoRestrictedGlobalsExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoRestrictedGlobalsRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: parenthesized / assertion-wrapped receiver ----
			// ESLint's own astUtils.isSpecificMemberAccess / getStaticPropertyName
			// also don't see through TSNonNullExpression or TSAsExpression (they
			// aren't MemberExpressions), so upstream doesn't report these either —
			// this matches upstream, not a divergence.
			{
				Code:    `window!.foo()`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"foo"}, "checkGlobalObject": true}},
			},
			{
				Code:    `(window as any).foo()`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"foo"}, "checkGlobalObject": true}},
			},

			// ---- Dimension 4: access/key forms — dynamic (non-static) computed key never matches ----
			// Locks in the `getStaticPropertyName` falsy branch: a computed member
			// access whose key can't be resolved statically must not report, even
			// when the runtime value could coincide with a restricted name.
			{
				Code:    `let key = "foo"; window[key]();`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"foo"}, "checkGlobalObject": true}},
			},

			// ---- Dimension 4: nesting — var hoists to the enclosing function only ----
			{
				Code:    `function outer() { var foo; function inner() { foo; } }`,
				Options: []interface{}{"foo"},
			},

			// ---- Dimension 4: graceful degradation — rest element in a binding pattern is a declaration ----
			{
				Code:    `const {...foo} = obj;`,
				Options: []interface{}{"foo"},
			},

			// ---- Real-user: parameter shadowing in a forEach/arrow callback (the rule's canonical `event` motivation) ----
			{
				Code:    `elements.forEach((event) => { doSomething(event); });`,
				Options: []interface{}{"event"},
			},
			// ---- Real-user: common restricted names collide constantly with plain object keys ----
			{
				Code:    `const config = { name: 'foo', self: 1 };`,
				Options: []interface{}{"name", "self"},
			},
			// ---- Real-user: enum members commonly collide with short/common restricted names ----
			{
				Code:    `enum Test { foo, bar }`,
				Options: []interface{}{"foo"},
			},

			// ---- Branch lock-in: restrictedGlobals.length === 0 short-circuits to a no-op ----
			{
				Code:    `foo`,
				Options: []interface{}{},
			},

			// ---- Branch lock-in: import specifier original (module export) name is not a scope reference ----
			{
				Code:    `import { foo as bar } from 'mod'; bar;`,
				Options: []interface{}{"foo"},
			},
			// ---- Branch lock-in: re-export original name is not a scope reference (only real for `export ... from`) ----
			{
				Code:    `export { foo as bar } from 'other';`,
				Options: []interface{}{"foo"},
			},

			// ---- Options coverage: bare (unwrapped) string, matching the CLI single-option shape ----
			{Code: `bar`, Options: "foo"},
			// ---- Options coverage: bare (unwrapped) object, matching the CLI single-option shape ----
			{Code: `window.bar()`, Options: map[string]interface{}{"globals": []interface{}{"foo"}, "checkGlobalObject": true}},
			// ---- Options coverage: nil options is a safe no-op ----
			{Code: `foo`, Options: nil},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: parenthesized global-object receiver ----
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

			// ---- Dimension 4: numeric / template-literal computed keys ----
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

			// ---- Dimension 4: null/boolean/BigInt computed keys ----
			// ESLint's astUtils.getStaticStringValue resolves `null`/`true`/`false`
			// literals (via node.value) and BigInt literals (via node.bigint) to
			// their string form, so `window[null]` is equivalent to `window["null"]`.
			// Regression test for a gap found in utils.GetStaticExpressionValue,
			// which originally only handled string/numeric/template/regex literals.
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

			// ---- Dimension 4: nesting — a reference outside every declaring function is not shadowed ----
			{
				Code:    `function outer() { function inner() { var foo; } } foo;`,
				Options: []interface{}{"foo"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 52, EndLine: 1, EndColumn: 55}},
			},

			// ---- Dimension 4: graceful degradation — spread in an object literal reads the value ----
			{
				Code:    `const obj = {...foo};`,
				Options: []interface{}{"foo"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 17, EndLine: 1, EndColumn: 20}},
			},

			// ---- Branch lock-in: isInTypeContext's ExpressionWithTypeArguments case
			// excludes class `extends` (a value context) but not `implements` /
			// `interface extends` — upstream itself never tests the `extends`
			// side of this distinction.
			{
				Code:    `class Derived extends Test {}`,
				Options: []interface{}{"Test"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 23, EndLine: 1, EndColumn: 27}},
			},

			// ---- Branch lock-in: last entry wins when the same name is restricted twice ----
			{
				Code:    `foo`,
				Options: []interface{}{"foo", map[string]interface{}{"name": "foo", "message": "second wins"}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "customMessage", Message: "Unexpected use of 'foo'. second wins", Line: 1, Column: 1, EndLine: 1, EndColumn: 4}},
			},

			// ---- Intentional divergence: rslint doesn't model ESLint's environment/
			// languageOptions.globals, so globalThis/self/window (and configured
			// globalObjects) are always recognized as global-object roots when
			// checkGlobalObject is enabled — regardless of any declared environment.
			// See the "SKIP" cases in the upstream file's valid list and the rule
			// doc's "Differences from ESLint" section.
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

			// ---- Options coverage: bare (unwrapped) string matching the CLI single-option shape ----
			{
				Code:    `foo`,
				Options: "foo",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 1, EndLine: 1, EndColumn: 4}},
			},
			// ---- Options coverage: bare (unwrapped) object matching the CLI single-option shape ----
			{
				Code:    `window.foo()`,
				Options: map[string]interface{}{"globals": []interface{}{"foo"}, "checkGlobalObject": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 8, EndLine: 1, EndColumn: 11}},
			},
		},
	)
}
