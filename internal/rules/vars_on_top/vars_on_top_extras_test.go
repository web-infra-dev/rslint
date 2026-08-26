package vars_on_top

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestVarsOnTopExtras covers tsgo AST shapes, upstream branch lock-ins, and
// real-world nesting cases not needed by the upstream migration.
func TestVarsOnTopExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &VarsOnTopRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: static block has no directive/import prefix ----
			{Code: "class C { static { var x; var y; } }", LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			// Locks in upstream isVarOnTop()'s directive/import skip arm.
			{Code: "'use strict'; import x from 'x'; var y;", LanguageOptions: rule.LanguageOptions{ECMAVersion: 6, SourceType: "module"}},
			// ---- Dimension 4: TypeScript annotations and declaration lists ----
			{Code: "function f() { var x: number = 1, y: string; return x; }"},
			// ---- Dimension 4: graceful degradation — malformed nesting is not special-cased ----
			{Code: "function f() { if (ok) { let x = 1; } }"},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: static block after executable statement ----
			{Code: "class C { static { foo(); var x; } }", LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "top", Message: varsOnTopMessage.Description, Line: 1, Column: 27}}},
			// ---- Dimension 4: loop and labeled statement boundaries ----
			{Code: "function f() { label: { var x; } }", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "top", Message: varsOnTopMessage.Description}}},
			// ---- Real-user: var nested in a conditional block ----
			{Code: "function render() { if (ready) { var value = getValue(); } return value; }", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "top", Message: varsOnTopMessage.Description}}},
			// ---- Real-user: var in a catch block ----
			{Code: "function load() { try { run(); } catch (error) { var message = error.message; } }", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "top", Message: varsOnTopMessage.Description}}},
			// Locks in blockScopeVarCheck()'s non-function-block arm.
			{Code: "{ var x; }", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "top", Message: varsOnTopMessage.Description}}},
		},
	)
}
