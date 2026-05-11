package scope

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// expectedError mirrors upstream's expected error shape. Centralized so a
// future text tweak lives in one place. Shared with scope_extras_test.go.
var expectedError = rule_tester.InvalidTestCaseError{
	MessageId: "scopeOnTh",
	Message:   errorMessage,
}

// componentsSettings mirrors upstream's
//
//	settings: { 'jsx-a11y': { components: { Foo: 'div', TableHeader: 'th' } } }
//
// — used to verify GetElementType honors the components map for both the
// "should report" direction (Foo → div, in dom set, not th) and the "should
// skip" direction (TableHeader → th, in dom set, exempt).
var componentsSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"components": map[string]interface{}{
			"Foo":         "div",
			"TableHeader": "th",
		},
	},
}

// TestScopeUpstream covers the full valid/invalid suite migrated 1:1 from
// upstream eslint-plugin-jsx-a11y's `__tests__/src/rules/scope-test.js`.
// Order and grouping mirror the upstream file so a future audit can grep
// across both side-by-side.
//
// rslint-specific lock-ins (case-sensitivity matrix, namespaced names,
// polymorphicPropName, listener boundary repeats, position assertions, the
// upstream branches that upstream itself never tests directly) live in
// scope_extras_test.go.
func TestScopeUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ScopeRule, []rule_tester.ValidTestCase{
		// ---- Upstream valid ----
		// No scope attribute — listener never reports.
		{Code: `<div />;`, Tsx: true},
		// Non-scope attribute — propName "foo", "FOO" !== "SCOPE" → early exit.
		{Code: `<div foo />;`, Tsx: true},
		// Boolean-form scope on th — th is in dom set, "TH" === "TH" → exempt.
		{Code: `<th scope />`, Tsx: true},
		// String-literal scope on th — exempt.
		{Code: `<th scope="row" />`, Tsx: true},
		// Identifier-valued scope on th — exempt.
		{Code: `<th scope={foo} />`, Tsx: true},
		// JsxExpression with string-literal value plus a spread attribute.
		// SpreadAttribute is its own AST kind — listener never fires on it.
		{Code: `<th scope={"col"} {...props} />`, Tsx: true},
		// Custom component without a components-map mapping — "Foo" is not in
		// the dom set → skipped at the IsDOMElement gate.
		{Code: `<Foo scope="bar" {...props} />`, Tsx: true},
		// components map remaps TableHeader → th → in dom set, exempt.
		{Code: `<TableHeader scope="row" />`, Tsx: true, Settings: componentsSettings},
	}, []rule_tester.InvalidTestCase{
		// ---- Upstream invalid ----
		// Boolean-form scope on a non-th DOM element — div is in dom set,
		// "DIV" !== "TH" → reports.
		{Code: `<div scope />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// components map remaps Foo → div → in dom set, "DIV" !== "TH" →
		// reports. Locks in that the components map can flip a custom
		// component INTO scope-rule coverage.
		{Code: `<Foo scope="bar" />`, Tsx: true, Settings: componentsSettings, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
	})
}
