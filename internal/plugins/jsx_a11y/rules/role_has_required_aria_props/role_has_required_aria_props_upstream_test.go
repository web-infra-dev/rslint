// cspell:ignore chcked expandd valuemax valuemin valuenow

package role_has_required_aria_props

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/jsxa11yutil"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// componentsSettings mirrors upstream's
//
//	const componentsSettings = {
//	  'jsx-a11y': { components: { MyComponent: 'div' } },
//	};
//
// — used by the `<MyComponent role="..." />` cases to resolve
// MyComponent → div before the rule applies.
var componentsSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"components": map[string]interface{}{
			"MyComponent": "div",
		},
	},
}

// TestRoleHasRequiredAriaPropsUpstream mirrors upstream's
// `__tests__/src/rules/role-has-required-aria-props-test.js` valid / invalid
// suite 1:1 and in upstream order. Anything NOT in upstream's file lives
// in role_has_required_aria_props_extras_test.go.
func TestRoleHasRequiredAriaPropsUpstream(t *testing.T) {
	// ---- Generated valid cases — one per non-abstract role.
	//
	// Upstream's `basicValidityTests` synthesizes
	//   `<div role="${role.toLowerCase()}" ${requiredProps.join(' ')} />`
	// for every non-abstract ARIA role, asserting that EVERY role passes
	// when its required props are present (or trivially passes when no
	// required props apply). Sourced from `jsxa11yutil.AriaRoleNonAbstract`
	// so the role list stays in sync with upstream aria-query data.
	roleRequiredMap := make(map[string][]string, len(jsxa11yutil.AriaRoleRequiredProps))
	for _, e := range jsxa11yutil.AriaRoleRequiredProps {
		roleRequiredMap[e.Role] = e.Props
	}
	generatedValid := make([]rule_tester.ValidTestCase, 0, len(jsxa11yutil.AriaRoleNonAbstract))
	for _, role := range jsxa11yutil.AriaRoleNonAbstract {
		var propChain string
		if required := roleRequiredMap[role]; len(required) > 0 {
			propChain = " " + strings.Join(required, " ")
		}
		generatedValid = append(generatedValid, rule_tester.ValidTestCase{
			Code: `<div role="` + role + `"` + propChain + ` />`,
			Tsx:  true,
		})
	}

	valid := []rule_tester.ValidTestCase{
		// ---- Upstream test file — `valid` section, in upstream order. ----

		// Custom component without role → name === 'role' check fails on
		// `baz`; the rule never sees a role attribute on Bar.
		{Code: `<Bar baz />`, Tsx: true},
		// Custom component WITH role (no components-map setting) →
		// elementType resolves to "MyComponent", IsDOMElement("MyComponent")
		// false → skip.
		{Code: `<MyComponent role="combobox" />`, Tsx: true},
		// No role attribute at all.
		{Code: `<div />`, Tsx: true},
		{Code: `<div></div>`, Tsx: true},
		// Non-literal role value (Identifier / LogicalExpression /
		// CallExpression / ...) — LITERAL_TYPES noop → null → skip.
		{Code: `<div role={role} />`, Tsx: true},
		{Code: `<div role={role || "button"} />`, Tsx: true},
		{Code: `<div role={role || "foobar"} />`, Tsx: true},
		// Valid role with no required props (`row`) — validRoles forEach
		// runs but the role's requiredProps is empty, so no report.
		{Code: `<div role="row" />`, Tsx: true},
		// Valid role with all required props — passes the
		// FindAttributeByName check on every prop.
		{
			Code: `<span role="checkbox" aria-checked="false" aria-labelledby="foo" tabindex="0"></span>`,
			Tsx:  true,
		},
		// Spread-before-type input with role="checkbox" — the type attribute
		// is on the element regardless of spread ordering, so
		// isSemanticRoleElement("input", ...) returns true via the
		// type=checkbox concept → skip.
		{
			Code: `<input role="checkbox" aria-checked="false" aria-labelledby="foo" tabindex="0" {...props} type="checkbox" />`,
			Tsx:  true,
		},
		// <input type="checkbox" role="switch" /> — input/type=checkbox
		// concept maps to roles ["checkbox", "switch"]; "switch" is in the
		// set → isSemanticRoleElement true → skip (no aria-checked needed).
		{Code: `<input type="checkbox" role="switch" />`, Tsx: true},
		// MyComponent → div via components setting; role="checkbox" with
		// aria-checked present → all required props satisfied.
		{
			Code:     `<MyComponent role="checkbox" aria-checked="false" aria-labelledby="foo" tabindex="0" />`,
			Tsx:      true,
			Settings: componentsSettings,
		},
		// heading + aria-level via JsxExpression-wrapped numeric literal.
		{Code: `<div role="heading" aria-level={2} />`, Tsx: true},
		// heading + aria-level via direct string literal — FindAttributeByName
		// only checks presence, not value type, so both forms pass.
		{Code: `<div role="heading" aria-level="3" />`, Tsx: true},
	}
	valid = append(valid, generatedValid...)

	invalid := []rule_tester.InvalidTestCase{
		// ---- Upstream test file — `invalid` section, in upstream order. ----

		// SLIDER — requires aria-valuenow.
		{
			Code: `<div role="slider" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-has-required-aria-props",
				Message:   errorMessage("slider", []string{"aria-valuenow"}),
				Line:      1, Column: 6,
			}},
		},
		{
			Code: `<div role="slider" aria-valuemax />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-has-required-aria-props",
				Message:   errorMessage("slider", []string{"aria-valuenow"}),
				Line:      1, Column: 6,
			}},
		},
		{
			Code: `<div role="slider" aria-valuemax aria-valuemin />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-has-required-aria-props",
				Message:   errorMessage("slider", []string{"aria-valuenow"}),
				Line:      1, Column: 6,
			}},
		},

		// CHECKBOX — requires aria-checked.
		{
			Code: `<div role="checkbox" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-has-required-aria-props",
				Message:   errorMessage("checkbox", []string{"aria-checked"}),
				Line:      1, Column: 6,
			}},
		},
		// "checked" (HTML, not ARIA) does NOT count — FindAttributeByName
		// looks up "aria-checked" verbatim.
		{
			Code: `<div role="checkbox" checked />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-has-required-aria-props",
				Message:   errorMessage("checkbox", []string{"aria-checked"}),
				Line:      1, Column: 6,
			}},
		},
		// Misspelled aria attr (aria-chcked) does NOT satisfy aria-checked.
		{
			Code: `<div role="checkbox" aria-chcked />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-has-required-aria-props",
				Message:   errorMessage("checkbox", []string{"aria-checked"}),
				Line:      1, Column: 6,
			}},
		},
		// span (not input[type=checkbox]) → no semantic skip; missing
		// aria-checked fails.
		{
			Code: `<span role="checkbox" aria-labelledby="foo" tabindex="0"></span>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-has-required-aria-props",
				Message:   errorMessage("checkbox", []string{"aria-checked"}),
				Line:      1, Column: 7,
			}},
		},

		// COMBOBOX — requires aria-controls AND aria-expanded.
		{
			Code: `<div role="combobox" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-has-required-aria-props",
				Message:   errorMessage("combobox", []string{"aria-controls", "aria-expanded"}),
				Line:      1, Column: 6,
			}},
		},
		// "expanded" (HTML, not aria-expanded) does NOT count.
		{
			Code: `<div role="combobox" expanded />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-has-required-aria-props",
				Message:   errorMessage("combobox", []string{"aria-controls", "aria-expanded"}),
				Line:      1, Column: 6,
			}},
		},
		// Misspelled aria-expandd does NOT satisfy aria-expanded.
		{
			Code: `<div role="combobox" aria-expandd />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-has-required-aria-props",
				Message:   errorMessage("combobox", []string{"aria-controls", "aria-expanded"}),
				Line:      1, Column: 6,
			}},
		},

		// SCROLLBAR — requires aria-controls AND aria-valuenow.
		{
			Code: `<div role="scrollbar" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-has-required-aria-props",
				Message:   errorMessage("scrollbar", []string{"aria-controls", "aria-valuenow"}),
				Line:      1, Column: 6,
			}},
		},
		{
			Code: `<div role="scrollbar" aria-valuemax />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-has-required-aria-props",
				Message:   errorMessage("scrollbar", []string{"aria-controls", "aria-valuenow"}),
				Line:      1, Column: 6,
			}},
		},
		{
			Code: `<div role="scrollbar" aria-valuemax aria-valuemin />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-has-required-aria-props",
				Message:   errorMessage("scrollbar", []string{"aria-controls", "aria-valuenow"}),
				Line:      1, Column: 6,
			}},
		},
		// aria-valuemax + aria-valuenow → still missing aria-controls.
		{
			Code: `<div role="scrollbar" aria-valuemax aria-valuenow />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-has-required-aria-props",
				Message:   errorMessage("scrollbar", []string{"aria-controls", "aria-valuenow"}),
				Line:      1, Column: 6,
			}},
		},
		// aria-valuemin + aria-valuenow → still missing aria-controls.
		{
			Code: `<div role="scrollbar" aria-valuemin aria-valuenow />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-has-required-aria-props",
				Message:   errorMessage("scrollbar", []string{"aria-controls", "aria-valuenow"}),
				Line:      1, Column: 6,
			}},
		},

		// HEADING — requires aria-level.
		{
			Code: `<div role="heading" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-has-required-aria-props",
				Message:   errorMessage("heading", []string{"aria-level"}),
				Line:      1, Column: 6,
			}},
		},

		// OPTION — requires aria-selected. Note: the <option> HTML element
		// natively maps to role "option" via the semantic-role-element
		// check, BUT only when the tag is literally `<option>`. `<div
		// role="option" />` does not get the semantic skip.
		{
			Code: `<div role="option" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-has-required-aria-props",
				Message:   errorMessage("option", []string{"aria-selected"}),
				Line:      1, Column: 6,
			}},
		},

		// Custom element MyComponent → div via components setting; combobox
		// requires aria-controls + aria-expanded, neither present.
		{
			Code:     `<MyComponent role="combobox" />`,
			Tsx:      true,
			Settings: componentsSettings,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-has-required-aria-props",
				Message:   errorMessage("combobox", []string{"aria-controls", "aria-expanded"}),
				Line:      1, Column: 14,
			}},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &RoleHasRequiredAriaPropsRule, valid, invalid)
}
