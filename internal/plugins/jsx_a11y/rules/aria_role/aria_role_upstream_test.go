// cspell:ignore datepicker fakeDOM foobar removalss

package aria_role

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/jsxa11yutil"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// allowedInvalidRolesOption mirrors upstream's `allowedInvalidRoles` option
// array (`['invalid-role', 'other-invalid-role']`). Reused across the
// valid / invalid suites so a future audit can diff against upstream
// `const allowedInvalidRoles = [{ allowedInvalidRoles: [...] }]`.
var allowedInvalidRolesOption = map[string]interface{}{
	"allowedInvalidRoles": []interface{}{"invalid-role", "other-invalid-role"},
}

// ignoreNonDOMOption mirrors upstream's `ignoreNonDOMSchema`.
var ignoreNonDOMOption = map[string]interface{}{"ignoreNonDOM": true}

// customDivSettings mirrors upstream's polymorphic + components config.
var customDivSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"polymorphicPropName": "asChild",
		"components": map[string]interface{}{
			"Div": "div",
		},
	},
}

// TestAriaRoleUpstream mirrors upstream's
// `__tests__/src/rules/aria-role-test.js` valid / invalid suite 1:1 and in
// upstream order. Anything NOT in upstream's file lives in
// aria_role_extras_test.go.
func TestAriaRoleUpstream(t *testing.T) {
	// Upstream uses `createTests(validRoles)` and `createTests(invalidRoles)`
	// to generate one `<div role="${role.toLowerCase()}" />` per
	// non-abstract / abstract role. The role lists are aria-query's
	// `roles.keys()` partitioned on `abstract === false` / `=== true`. We
	// source the same data from jsxa11yutil so a future aria-query refresh
	// updates both rule logic and tests at once.
	validRoleTests := make([]rule_tester.ValidTestCase, 0, len(jsxa11yutil.AriaRoleNonAbstract))
	for _, role := range jsxa11yutil.AriaRoleNonAbstract {
		validRoleTests = append(validRoleTests, rule_tester.ValidTestCase{
			Code: `<div role="` + role + `" />`,
			Tsx:  true,
		})
	}
	invalidRoleTests := make([]rule_tester.InvalidTestCase, 0, len(jsxa11yutil.AriaRoleAbstract))
	for _, role := range jsxa11yutil.AriaRoleAbstract {
		invalidRoleTests = append(invalidRoleTests, rule_tester.InvalidTestCase{
			Code: `<div role="` + role + `" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaRole",
				Message:   errorMessage,
			}},
		})
	}

	valid := []rule_tester.ValidTestCase{
		// No role / non-literal role — listener doesn't fire OR fires but
		// gets a NoLit value → skip.
		{Code: `<div />`, Tsx: true},
		{Code: `<div></div>`, Tsx: true},
		{Code: `<div role={role} />`, Tsx: true},
		{Code: `<div role={role || "button"} />`, Tsx: true},
		{Code: `<div role={role || "foobar"} />`, Tsx: true},

		// Literal valid roles — single and space-separated.
		{Code: `<div role="tabpanel row" />`, Tsx: true},
		{Code: `<div role="switch" />`, Tsx: true},
		{Code: `<div role="doc-abstract" />`, Tsx: true},
		{Code: `<div role="doc-appendix doc-bibliography" />`, Tsx: true},

		// Component with no role attr — listener fires on `baz` but
		// `name !== 'ROLE'` short-circuits.
		{Code: `<Bar baz />`, Tsx: true},

		// allowedInvalidRoles option.
		{Code: `<img role="invalid-role" />`, Tsx: true, Options: allowedInvalidRolesOption},
		{Code: `<img role="invalid-role tabpanel" />`, Tsx: true, Options: allowedInvalidRolesOption},
		{Code: `<img role="invalid-role other-invalid-role" />`, Tsx: true, Options: allowedInvalidRolesOption},

		// ignoreNonDOM option — custom components skip.
		{Code: `<Foo role="bar" />`, Tsx: true, Options: ignoreNonDOMOption},
		{Code: `<fakeDOM role="bar" />`, Tsx: true, Options: ignoreNonDOMOption},
		// DOM element with valid role — ignoreNonDOM doesn't skip it.
		{Code: `<img role="presentation" />`, Tsx: true, Options: ignoreNonDOMOption},

		// Settings: `components` map + `polymorphicPropName`. `<Div>` resolves
		// to `div` (components-map lookup); `<Box asChild="div">` resolves to
		// `div` via the polymorphic prop. The resolved name doesn't matter
		// for the valid-role check — only ignoreNonDOM consumes elementType
		// — but the cases exercise the settings plumbing path.
		{Code: `<Div role="button" />`, Tsx: true, Settings: customDivSettings},
		{Code: `<Box asChild="div" role="button" />`, Tsx: true, Settings: customDivSettings},

		// SVG-namespaced roles — graphics-* extension roles validate the
		// same as core / DPUB roles.
		{Code: `<svg role="graphics-document document" />`, Tsx: true},
		{Code: `<svg role="img" />`, Tsx: true},
	}
	valid = append(valid, validRoleTests...)

	invalid := []rule_tester.InvalidTestCase{
		// Unknown / abstract / case-sensitive / empty / boolean form / null.
		{
			Code: `<div role="foobar" />`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaRole", Message: errorMessage}},
		},
		{
			Code: `<div role="datepicker"></div>`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaRole", Message: errorMessage}},
		},
		{
			Code: `<div role="range"></div>`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaRole", Message: errorMessage}},
		},
		{
			Code: `<div role="Button"></div>`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaRole", Message: errorMessage}},
		},
		{
			Code: `<div role=""></div>`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaRole", Message: errorMessage}},
		},
		// Space-split with one bad token in the middle — every() short-circuits.
		{
			Code: `<div role="tabpanel row foobar"></div>`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaRole", Message: errorMessage}},
		},
		{
			Code: `<div role="tabpanel row range"></div>`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaRole", Message: errorMessage}},
		},
		{
			Code: `<div role="doc-endnotes range"></div>`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaRole", Message: errorMessage}},
		},
		// Boolean-attribute form — String(true) = "true", "true" not a role.
		{
			Code: `<div role />`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaRole", Message: errorMessage}},
		},
		// allowedInvalidRoles does NOT cover the unknown token here.
		{
			Code: `<div role="unknown-invalid-role" />`, Tsx: true,
			Options: allowedInvalidRolesOption,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaRole", Message: errorMessage}},
		},
		// `{null}` resolves to the literal STRING "null" via LITERAL_TYPES.Literal
		// override — NOT to the JS null sentinel. So it doesn't hit the
		// step-3 early return; it's validated and fails the role lookup.
		{
			Code: `<div role={null}></div>`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaRole", Message: errorMessage}},
		},
		// Custom component without ignoreNonDOM — rule still runs.
		{
			Code: `<Foo role="datepicker" />`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaRole", Message: errorMessage}},
		},
		{
			Code: `<Foo role="Button" />`, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaRole", Message: errorMessage}},
		},
		// Components-map: `Div` resolves to `div`; "Button" still fails case check.
		{
			Code: `<Div role="Button" />`, Tsx: true, Settings: customDivSettings,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaRole", Message: errorMessage}},
		},
		// ignoreNonDOM + components-map: `Div` → `div` IS in the DOM map, so
		// the rule still runs.
		{
			Code: `<Div role="Button" />`, Tsx: true,
			Options:  ignoreNonDOMOption,
			Settings: customDivSettings,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaRole", Message: errorMessage}},
		},
		// Polymorphic prop: `Box asChild="div"` → `div`; "Button" fails.
		{
			Code: `<Box asChild="div" role="Button" />`, Tsx: true, Settings: customDivSettings,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaRole", Message: errorMessage}},
		},
	}
	invalid = append(invalid, invalidRoleTests...)

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &AriaRoleRule, valid, invalid)
}
