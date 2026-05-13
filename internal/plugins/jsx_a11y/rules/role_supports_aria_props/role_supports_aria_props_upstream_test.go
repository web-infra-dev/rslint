// cspell:ignore activedescendant atomic checkbox controls describedby disabled
// cspell:ignore dropeffect flowto grabbed haspopup hidden labelledby owns
// cspell:ignore posinset relevant setsize valuemax valuemin valuenow valuetext
// cspell:ignore searchbox combobox listbox menubar menuitem rowgroup rowheader
// cspell:ignore tablist tabpanel toolbar tooltip treegrid treeitem progressbar
// cspell:ignore radiogroup spinbutton scrollbar gridcell complementary contentinfo
// cspell:ignore alertdialog presentation directory definition feed group meter
// cspell:ignore log marquee navigation status switch term timer
// cspell:ignore braillelabel brailleroledescription colcount colindex colspan
// cspell:ignore errormessage keyshortcuts modal multiline multiselectable
// cspell:ignore orientation placeholder pressed readonly relevant required
// cspell:ignore roledescription rowcount rowindex rowspan selected
// cspell:ignore subscript superscript blockquote insertion deletion mark
// cspell:ignore generic emphasis strong paragraph caption columnheader
// cspell:ignore separator menubar tablist treeitem application banner
// cspell:ignore searchbox spinbutton textbox tabpanel toolbar tooltip alertdialog
// cspell:ignore code

package role_supports_aria_props

import (
	"sort"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/jsxa11yutil"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// componentsSettings mirrors upstream's `componentsSettings`:
//
//	const componentsSettings = {
//	  'jsx-a11y': { components: { Link: 'a' } },
//	};
//
// Used by the `<Link href="#" aria-checked />` invalid case so that the
// element resolves to "a" before the implicit-role lookup.
var componentsSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"components": map[string]interface{}{
			"Link": "a",
		},
	},
}

// generatedCases mirrors upstream's `createTests(nonAbstractRoles)` —
// for every non-abstract role and every ARIA attribute, synthesize a
// `<div role="<role>" <prop> />` case classified into valid / invalid
// based on whether the prop is in the role's supported props set.
//
// The output ordering matches upstream:
//
//	for each role in nonAbstractRoles (aria-query insertion order):
//	  emit valid cases for every supported prop (in supported-set
//	    insertion order, lower-cased per upstream `prop.toLowerCase()`)
//	  emit invalid cases for every aria-* prop NOT in the supported set
//	    (in `aria.keys()` order, lower-cased)
//
// We keep the iteration deterministic by sorting Go map keys before
// rendering. This is a slight observable divergence from upstream, where
// `Object.keys(propKeyValues)` follows V8's property-insertion order —
// in practice neither test ordering matters because each row is an
// independent assertion; sorting just guarantees reproducibility.
func generatedCases() ([]rule_tester.ValidTestCase, []rule_tester.InvalidTestCase) {
	var valid []rule_tester.ValidTestCase
	var invalid []rule_tester.InvalidTestCase
	for _, role := range jsxa11yutil.AriaRoleNonAbstract {
		supported, ok := jsxa11yutil.AriaRolePropsMap[role]
		if !ok {
			continue
		}
		// Sorted iteration so the generator is deterministic.
		supportedNames := make([]string, 0, len(supported))
		for k := range supported {
			supportedNames = append(supportedNames, k)
		}
		sort.Strings(supportedNames)
		for _, prop := range supportedNames {
			valid = append(valid, rule_tester.ValidTestCase{
				Code: `<div role="` + role + `" ` + prop + ` />`,
				Tsx:  true,
			})
		}
		// Invalid: every ARIA name NOT in supported. Walk
		// AriaPropertyNames so the order matches `aria.keys()`.
		for _, prop := range jsxa11yutil.AriaPropertyNames {
			if _, ok := supported[prop]; ok {
				continue
			}
			invalid = append(invalid, rule_tester.InvalidTestCase{
				Code: `<div role="` + role + `" ` + prop + ` />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-supports-aria-props",
					Message:   errorMessage(prop, role, "div", false),
					Line:      1, Column: 1,
				}},
			})
		}
	}
	return valid, invalid
}

// TestRoleSupportsAriaPropsUpstream mirrors upstream's
// `__tests__/src/rules/role-supports-aria-props-test.js` valid / invalid
// suite 1:1 and in upstream order. Anything NOT in upstream's file lives
// in role_supports_aria_props_extras_test.go.
func TestRoleSupportsAriaPropsUpstream(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// ---- Generic / non-DOM / explicit-role-no-aria cases. ----

		{Code: `<Foo bar />`, Tsx: true},
		{Code: `<div />`, Tsx: true},
		{Code: `<div id="main" />`, Tsx: true},
		// `<div role />` — boolean form → getLiteralPropValue returns true
		// → typeof !== 'string' → skip.
		{Code: `<div role />`, Tsx: true},
		{Code: `<div role="presentation" {...props} />`, Tsx: true},
		{Code: `<Foo.Bar baz={true} />`, Tsx: true},
		// `<Link href="#" aria-checked />` without componentsSettings —
		// elementType "Link" has no implicit role → skip.
		{Code: `<Link href="#" aria-checked />`, Tsx: true},

		// ---- IMPLICIT ROLE TESTS — A: implicit role is `link`. ----

		{Code: `<a href="#" aria-expanded />`, Tsx: true},
		{Code: `<a href="#" aria-atomic />`, Tsx: true},
		{Code: `<a href="#" aria-busy />`, Tsx: true},
		{Code: `<a href="#" aria-controls />`, Tsx: true},
		{Code: `<a href="#" aria-current />`, Tsx: true},
		{Code: `<a href="#" aria-describedby />`, Tsx: true},
		{Code: `<a href="#" aria-disabled />`, Tsx: true},
		{Code: `<a href="#" aria-dropeffect />`, Tsx: true},
		{Code: `<a href="#" aria-flowto />`, Tsx: true},
		{Code: `<a href="#" aria-haspopup />`, Tsx: true},
		{Code: `<a href="#" aria-grabbed />`, Tsx: true},
		{Code: `<a href="#" aria-hidden />`, Tsx: true},
		{Code: `<a href="#" aria-label />`, Tsx: true},
		{Code: `<a href="#" aria-labelledby />`, Tsx: true},
		{Code: `<a href="#" aria-live />`, Tsx: true},
		{Code: `<a href="#" aria-owns />`, Tsx: true},
		{Code: `<a href="#" aria-relevant />`, Tsx: true},
		// `<a aria-checked />` — without href, no implicit role → skip.
		{Code: `<a aria-checked />`, Tsx: true},

		// ---- IMPLICIT ROLE TESTS — AREA: implicit role is `link` (with href). ----

		{Code: `<area href="#" aria-expanded />`, Tsx: true},
		{Code: `<area href="#" aria-atomic />`, Tsx: true},
		{Code: `<area href="#" aria-busy />`, Tsx: true},
		{Code: `<area href="#" aria-controls />`, Tsx: true},
		{Code: `<area href="#" aria-describedby />`, Tsx: true},
		{Code: `<area href="#" aria-disabled />`, Tsx: true},
		{Code: `<area href="#" aria-dropeffect />`, Tsx: true},
		{Code: `<area href="#" aria-flowto />`, Tsx: true},
		{Code: `<area href="#" aria-grabbed />`, Tsx: true},
		{Code: `<area href="#" aria-haspopup />`, Tsx: true},
		{Code: `<area href="#" aria-hidden />`, Tsx: true},
		{Code: `<area href="#" aria-label />`, Tsx: true},
		{Code: `<area href="#" aria-labelledby />`, Tsx: true},
		{Code: `<area href="#" aria-live />`, Tsx: true},
		{Code: `<area href="#" aria-owns />`, Tsx: true},
		{Code: `<area href="#" aria-relevant />`, Tsx: true},
		{Code: `<area aria-checked />`, Tsx: true},

		// ---- IMPLICIT ROLE TESTS — LINK: implicit role is `link` (with href). ----

		{Code: `<link href="#" aria-expanded />`, Tsx: true},
		{Code: `<link href="#" aria-atomic />`, Tsx: true},
		{Code: `<link href="#" aria-busy />`, Tsx: true},
		{Code: `<link href="#" aria-controls />`, Tsx: true},
		{Code: `<link href="#" aria-describedby />`, Tsx: true},
		{Code: `<link href="#" aria-disabled />`, Tsx: true},
		{Code: `<link href="#" aria-dropeffect />`, Tsx: true},
		{Code: `<link href="#" aria-flowto />`, Tsx: true},
		{Code: `<link href="#" aria-grabbed />`, Tsx: true},
		{Code: `<link href="#" aria-hidden />`, Tsx: true},
		{Code: `<link href="#" aria-haspopup />`, Tsx: true},
		{Code: `<link href="#" aria-label />`, Tsx: true},
		{Code: `<link href="#" aria-labelledby />`, Tsx: true},
		{Code: `<link href="#" aria-live />`, Tsx: true},
		{Code: `<link href="#" aria-owns />`, Tsx: true},
		{Code: `<link href="#" aria-relevant />`, Tsx: true},
		{Code: `<link aria-checked />`, Tsx: true},

		// ---- IMPLICIT ROLE TESTS — IMG. ----

		// `<img alt="" aria-checked />` — empty alt suppresses the implicit
		// "img" role (returns "" from implicitRoleForImg), so the rule skips
		// (no implicit role → no validation).
		{Code: `<img alt="" aria-checked />`, Tsx: true},
		// `<img alt="foobar" aria-busy />` — non-empty alt → implicit role is
		// "img"; aria-busy IS in img's supported props.
		{Code: `<img alt="foobar" aria-busy />`, Tsx: true},

		// ---- IMPLICIT ROLE TESTS — MENU type="toolbar" → "toolbar". ----

		{Code: `<menu type="toolbar" aria-activedescendant />`, Tsx: true},
		{Code: `<menu type="toolbar" aria-atomic />`, Tsx: true},
		{Code: `<menu type="toolbar" aria-busy />`, Tsx: true},
		{Code: `<menu type="toolbar" aria-controls />`, Tsx: true},
		{Code: `<menu type="toolbar" aria-describedby />`, Tsx: true},
		{Code: `<menu type="toolbar" aria-disabled />`, Tsx: true},
		{Code: `<menu type="toolbar" aria-dropeffect />`, Tsx: true},
		{Code: `<menu type="toolbar" aria-flowto />`, Tsx: true},
		{Code: `<menu type="toolbar" aria-grabbed />`, Tsx: true},
		{Code: `<menu type="toolbar" aria-hidden />`, Tsx: true},
		{Code: `<menu type="toolbar" aria-label />`, Tsx: true},
		{Code: `<menu type="toolbar" aria-labelledby />`, Tsx: true},
		{Code: `<menu type="toolbar" aria-live />`, Tsx: true},
		{Code: `<menu type="toolbar" aria-owns />`, Tsx: true},
		{Code: `<menu type="toolbar" aria-relevant />`, Tsx: true},
		// `<menu aria-checked />` — no type → no implicit role → skip.
		{Code: `<menu aria-checked />`, Tsx: true},

		// ---- IMPLICIT ROLE TESTS — MENUITEM type="command" → "menuitem". ----

		{Code: `<menuitem type="command" aria-atomic />`, Tsx: true},
		{Code: `<menuitem type="command" aria-busy />`, Tsx: true},
		{Code: `<menuitem type="command" aria-controls />`, Tsx: true},
		{Code: `<menuitem type="command" aria-describedby />`, Tsx: true},
		{Code: `<menuitem type="command" aria-disabled />`, Tsx: true},
		{Code: `<menuitem type="command" aria-dropeffect />`, Tsx: true},
		{Code: `<menuitem type="command" aria-flowto />`, Tsx: true},
		{Code: `<menuitem type="command" aria-grabbed />`, Tsx: true},
		{Code: `<menuitem type="command" aria-haspopup />`, Tsx: true},
		{Code: `<menuitem type="command" aria-hidden />`, Tsx: true},
		{Code: `<menuitem type="command" aria-label />`, Tsx: true},
		{Code: `<menuitem type="command" aria-labelledby />`, Tsx: true},
		{Code: `<menuitem type="command" aria-live />`, Tsx: true},
		{Code: `<menuitem type="command" aria-owns />`, Tsx: true},
		{Code: `<menuitem type="command" aria-relevant />`, Tsx: true},

		// ---- IMPLICIT ROLE TESTS — MENUITEM type="checkbox" → "menuitemcheckbox". ----

		{Code: `<menuitem type="checkbox" aria-checked />`, Tsx: true},
		{Code: `<menuitem type="checkbox" aria-atomic />`, Tsx: true},
		{Code: `<menuitem type="checkbox" aria-busy />`, Tsx: true},
		{Code: `<menuitem type="checkbox" aria-controls />`, Tsx: true},
		{Code: `<menuitem type="checkbox" aria-describedby />`, Tsx: true},
		{Code: `<menuitem type="checkbox" aria-disabled />`, Tsx: true},
		{Code: `<menuitem type="checkbox" aria-dropeffect />`, Tsx: true},
		{Code: `<menuitem type="checkbox" aria-flowto />`, Tsx: true},
		{Code: `<menuitem type="checkbox" aria-grabbed />`, Tsx: true},
		{Code: `<menuitem type="checkbox" aria-haspopup />`, Tsx: true},
		{Code: `<menuitem type="checkbox" aria-hidden />`, Tsx: true},
		{Code: `<menuitem type="checkbox" aria-invalid />`, Tsx: true},
		{Code: `<menuitem type="checkbox" aria-label />`, Tsx: true},
		{Code: `<menuitem type="checkbox" aria-labelledby />`, Tsx: true},
		{Code: `<menuitem type="checkbox" aria-live />`, Tsx: true},
		{Code: `<menuitem type="checkbox" aria-owns />`, Tsx: true},
		{Code: `<menuitem type="checkbox" aria-relevant />`, Tsx: true},

		// ---- IMPLICIT ROLE TESTS — MENUITEM type="radio" → "menuitemradio". ----

		{Code: `<menuitem type="radio" aria-checked />`, Tsx: true},
		{Code: `<menuitem type="radio" aria-atomic />`, Tsx: true},
		{Code: `<menuitem type="radio" aria-busy />`, Tsx: true},
		{Code: `<menuitem type="radio" aria-controls />`, Tsx: true},
		{Code: `<menuitem type="radio" aria-describedby />`, Tsx: true},
		{Code: `<menuitem type="radio" aria-disabled />`, Tsx: true},
		{Code: `<menuitem type="radio" aria-dropeffect />`, Tsx: true},
		{Code: `<menuitem type="radio" aria-flowto />`, Tsx: true},
		{Code: `<menuitem type="radio" aria-grabbed />`, Tsx: true},
		{Code: `<menuitem type="radio" aria-haspopup />`, Tsx: true},
		{Code: `<menuitem type="radio" aria-hidden />`, Tsx: true},
		{Code: `<menuitem type="radio" aria-invalid />`, Tsx: true},
		{Code: `<menuitem type="radio" aria-label />`, Tsx: true},
		{Code: `<menuitem type="radio" aria-labelledby />`, Tsx: true},
		{Code: `<menuitem type="radio" aria-live />`, Tsx: true},
		{Code: `<menuitem type="radio" aria-owns />`, Tsx: true},
		{Code: `<menuitem type="radio" aria-relevant />`, Tsx: true},
		{Code: `<menuitem type="radio" aria-posinset />`, Tsx: true},
		{Code: `<menuitem type="radio" aria-setsize />`, Tsx: true},

		// `<menuitem aria-checked />` — no type → no implicit role → skip.
		{Code: `<menuitem aria-checked />`, Tsx: true},
		// `<menuitem type="foo" aria-checked />` — unknown type → no implicit role.
		{Code: `<menuitem type="foo" aria-checked />`, Tsx: true},

		// ---- IMPLICIT ROLE TESTS — INPUT type="button" → "button". ----

		{Code: `<input type="button" aria-expanded />`, Tsx: true},
		{Code: `<input type="button" aria-pressed />`, Tsx: true},
		{Code: `<input type="button" aria-atomic />`, Tsx: true},
		{Code: `<input type="button" aria-busy />`, Tsx: true},
		{Code: `<input type="button" aria-controls />`, Tsx: true},
		{Code: `<input type="button" aria-describedby />`, Tsx: true},
		{Code: `<input type="button" aria-disabled />`, Tsx: true},
		{Code: `<input type="button" aria-dropeffect />`, Tsx: true},
		{Code: `<input type="button" aria-flowto />`, Tsx: true},
		{Code: `<input type="button" aria-grabbed />`, Tsx: true},
		{Code: `<input type="button" aria-haspopup />`, Tsx: true},
		{Code: `<input type="button" aria-hidden />`, Tsx: true},
		{Code: `<input type="button" aria-label />`, Tsx: true},
		{Code: `<input type="button" aria-labelledby />`, Tsx: true},
		{Code: `<input type="button" aria-live />`, Tsx: true},
		{Code: `<input type="button" aria-owns />`, Tsx: true},
		{Code: `<input type="button" aria-relevant />`, Tsx: true},

		// ---- IMPLICIT ROLE TESTS — INPUT type="image" → "button". ----

		{Code: `<input type="image" aria-expanded />`, Tsx: true},
		{Code: `<input type="image" aria-pressed />`, Tsx: true},
		{Code: `<input type="image" aria-atomic />`, Tsx: true},
		{Code: `<input type="image" aria-busy />`, Tsx: true},
		{Code: `<input type="image" aria-controls />`, Tsx: true},
		{Code: `<input type="image" aria-describedby />`, Tsx: true},
		{Code: `<input type="image" aria-disabled />`, Tsx: true},
		{Code: `<input type="image" aria-dropeffect />`, Tsx: true},
		{Code: `<input type="image" aria-flowto />`, Tsx: true},
		{Code: `<input type="image" aria-grabbed />`, Tsx: true},
		{Code: `<input type="image" aria-haspopup />`, Tsx: true},
		{Code: `<input type="image" aria-hidden />`, Tsx: true},
		{Code: `<input type="image" aria-label />`, Tsx: true},
		{Code: `<input type="image" aria-labelledby />`, Tsx: true},
		{Code: `<input type="image" aria-live />`, Tsx: true},
		{Code: `<input type="image" aria-owns />`, Tsx: true},
		{Code: `<input type="image" aria-relevant />`, Tsx: true},

		// ---- IMPLICIT ROLE TESTS — INPUT type="reset" → "button". ----

		{Code: `<input type="reset" aria-expanded />`, Tsx: true},
		{Code: `<input type="reset" aria-pressed />`, Tsx: true},
		{Code: `<input type="reset" aria-atomic />`, Tsx: true},
		{Code: `<input type="reset" aria-busy />`, Tsx: true},
		{Code: `<input type="reset" aria-controls />`, Tsx: true},
		{Code: `<input type="reset" aria-describedby />`, Tsx: true},
		{Code: `<input type="reset" aria-disabled />`, Tsx: true},
		{Code: `<input type="reset" aria-dropeffect />`, Tsx: true},
		{Code: `<input type="reset" aria-flowto />`, Tsx: true},
		{Code: `<input type="reset" aria-grabbed />`, Tsx: true},
		{Code: `<input type="reset" aria-haspopup />`, Tsx: true},
		{Code: `<input type="reset" aria-hidden />`, Tsx: true},
		{Code: `<input type="reset" aria-label />`, Tsx: true},
		{Code: `<input type="reset" aria-labelledby />`, Tsx: true},
		{Code: `<input type="reset" aria-live />`, Tsx: true},
		{Code: `<input type="reset" aria-owns />`, Tsx: true},
		{Code: `<input type="reset" aria-relevant />`, Tsx: true},

		// ---- IMPLICIT ROLE TESTS — INPUT type="submit" → "button". ----

		{Code: `<input type="submit" aria-expanded />`, Tsx: true},
		{Code: `<input type="submit" aria-pressed />`, Tsx: true},
		{Code: `<input type="submit" aria-atomic />`, Tsx: true},
		{Code: `<input type="submit" aria-busy />`, Tsx: true},
		{Code: `<input type="submit" aria-controls />`, Tsx: true},
		{Code: `<input type="submit" aria-describedby />`, Tsx: true},
		{Code: `<input type="submit" aria-disabled />`, Tsx: true},
		{Code: `<input type="submit" aria-dropeffect />`, Tsx: true},
		{Code: `<input type="submit" aria-flowto />`, Tsx: true},
		{Code: `<input type="submit" aria-grabbed />`, Tsx: true},
		{Code: `<input type="submit" aria-haspopup />`, Tsx: true},
		{Code: `<input type="submit" aria-hidden />`, Tsx: true},
		{Code: `<input type="submit" aria-label />`, Tsx: true},
		{Code: `<input type="submit" aria-labelledby />`, Tsx: true},
		{Code: `<input type="submit" aria-live />`, Tsx: true},
		{Code: `<input type="submit" aria-owns />`, Tsx: true},
		{Code: `<input type="submit" aria-relevant />`, Tsx: true},

		// ---- IMPLICIT ROLE TESTS — INPUT type="checkbox" → "checkbox". ----

		{Code: `<input type="checkbox" aria-atomic />`, Tsx: true},
		{Code: `<input type="checkbox" aria-busy />`, Tsx: true},
		{Code: `<input type="checkbox" aria-checked />`, Tsx: true},
		{Code: `<input type="checkbox" aria-controls />`, Tsx: true},
		{Code: `<input type="checkbox" aria-describedby />`, Tsx: true},
		{Code: `<input type="checkbox" aria-disabled />`, Tsx: true},
		{Code: `<input type="checkbox" aria-dropeffect />`, Tsx: true},
		{Code: `<input type="checkbox" aria-flowto />`, Tsx: true},
		{Code: `<input type="checkbox" aria-grabbed />`, Tsx: true},
		{Code: `<input type="checkbox" aria-hidden />`, Tsx: true},
		{Code: `<input type="checkbox" aria-invalid />`, Tsx: true},
		{Code: `<input type="checkbox" aria-label />`, Tsx: true},
		{Code: `<input type="checkbox" aria-labelledby />`, Tsx: true},
		{Code: `<input type="checkbox" aria-live />`, Tsx: true},
		{Code: `<input type="checkbox" aria-owns />`, Tsx: true},
		{Code: `<input type="checkbox" aria-relevant />`, Tsx: true},

		// ---- IMPLICIT ROLE TESTS — INPUT type="radio" → "radio". ----

		{Code: `<input type="radio" aria-atomic />`, Tsx: true},
		{Code: `<input type="radio" aria-busy />`, Tsx: true},
		{Code: `<input type="radio" aria-checked />`, Tsx: true},
		{Code: `<input type="radio" aria-controls />`, Tsx: true},
		{Code: `<input type="radio" aria-describedby />`, Tsx: true},
		{Code: `<input type="radio" aria-disabled />`, Tsx: true},
		{Code: `<input type="radio" aria-dropeffect />`, Tsx: true},
		{Code: `<input type="radio" aria-flowto />`, Tsx: true},
		{Code: `<input type="radio" aria-grabbed />`, Tsx: true},
		{Code: `<input type="radio" aria-hidden />`, Tsx: true},
		{Code: `<input type="radio" aria-label />`, Tsx: true},
		{Code: `<input type="radio" aria-labelledby />`, Tsx: true},
		{Code: `<input type="radio" aria-live />`, Tsx: true},
		{Code: `<input type="radio" aria-owns />`, Tsx: true},
		{Code: `<input type="radio" aria-relevant />`, Tsx: true},
		{Code: `<input type="radio" aria-posinset />`, Tsx: true},
		{Code: `<input type="radio" aria-setsize />`, Tsx: true},

		// ---- IMPLICIT ROLE TESTS — INPUT type="range" → "slider". ----

		{Code: `<input type="range" aria-valuemax />`, Tsx: true},
		{Code: `<input type="range" aria-valuemin />`, Tsx: true},
		{Code: `<input type="range" aria-valuenow />`, Tsx: true},
		{Code: `<input type="range" aria-orientation />`, Tsx: true},
		{Code: `<input type="range" aria-atomic />`, Tsx: true},
		{Code: `<input type="range" aria-busy />`, Tsx: true},
		{Code: `<input type="range" aria-controls />`, Tsx: true},
		{Code: `<input type="range" aria-describedby />`, Tsx: true},
		{Code: `<input type="range" aria-disabled />`, Tsx: true},
		{Code: `<input type="range" aria-dropeffect />`, Tsx: true},
		{Code: `<input type="range" aria-flowto />`, Tsx: true},
		{Code: `<input type="range" aria-grabbed />`, Tsx: true},
		{Code: `<input type="range" aria-haspopup />`, Tsx: true},
		{Code: `<input type="range" aria-hidden />`, Tsx: true},
		{Code: `<input type="range" aria-invalid />`, Tsx: true},
		{Code: `<input type="range" aria-label />`, Tsx: true},
		{Code: `<input type="range" aria-labelledby />`, Tsx: true},
		{Code: `<input type="range" aria-live />`, Tsx: true},
		{Code: `<input type="range" aria-owns />`, Tsx: true},
		{Code: `<input type="range" aria-relevant />`, Tsx: true},
		{Code: `<input type="range" aria-valuetext />`, Tsx: true},

		// ---- INPUT default → "textbox" (any unknown / absent type). ----

		{Code: `<input type="email" aria-disabled />`, Tsx: true},
		{Code: `<input type="password" aria-disabled />`, Tsx: true},
		{Code: `<input type="search" aria-disabled />`, Tsx: true},
		{Code: `<input type="tel" aria-disabled />`, Tsx: true},
		{Code: `<input type="url" aria-disabled />`, Tsx: true},
		{Code: `<input aria-disabled />`, Tsx: true},

		// ---- Allow null/undefined values regardless of role. ----

		{Code: `<h2 role="presentation" aria-level={null} />`, Tsx: true},
		{Code: `<h2 role="presentation" aria-level={undefined} />`, Tsx: true},

		// ---- Other common implicit roles. ----

		{Code: `<button aria-pressed />`, Tsx: true},
		{Code: `<form aria-hidden />`, Tsx: true},
		{Code: `<h1 aria-hidden />`, Tsx: true},
		{Code: `<h2 aria-hidden />`, Tsx: true},
		{Code: `<h3 aria-hidden />`, Tsx: true},
		{Code: `<h4 aria-hidden />`, Tsx: true},
		{Code: `<h5 aria-hidden />`, Tsx: true},
		{Code: `<h6 aria-hidden />`, Tsx: true},
		{Code: `<hr aria-hidden />`, Tsx: true},
		{Code: `<li aria-current />`, Tsx: true},
		{Code: `<meter aria-atomic />`, Tsx: true},
		{Code: `<option aria-atomic />`, Tsx: true},
		{Code: `<progress aria-atomic />`, Tsx: true},
		{Code: `<textarea aria-hidden />`, Tsx: true},
		{Code: `<select aria-expanded />`, Tsx: true},
		{Code: `<datalist aria-expanded />`, Tsx: true},
		{Code: `<div role="heading" aria-level />`, Tsx: true},
		{Code: `<div role="heading" aria-level="1" />`, Tsx: true},

		// ---- Fragment-as-prop case from upstream (semver gated upstream;
		// always included here since rslint targets modern parsers). ----
		{
			Code: `
				const HelloThere = () => (
					<Hello
						role="searchbox"
						frag={
							<>
								<div>Hello</div>
								<div>There</div>
							</>
						}
					/>
				);

				const Hello = (props) => <div>{props.frag}</div>;
			`,
			Tsx: true,
		},
	}

	invalid := []rule_tester.InvalidTestCase{
		// ---- IMPLICIT BASIC CHECKS. ----

		{
			Code: `<a href="#" aria-checked />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-supports-aria-props",
				Message:   errorMessage("aria-checked", "link", "a", true),
				Line:      1, Column: 1,
			}},
		},
		{
			Code: `<area href="#" aria-checked />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-supports-aria-props",
				Message:   errorMessage("aria-checked", "link", "area", true),
				Line:      1, Column: 1,
			}},
		},
		{
			Code: `<link href="#" aria-checked />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-supports-aria-props",
				Message:   errorMessage("aria-checked", "link", "link", true),
				Line:      1, Column: 1,
			}},
		},
		{
			Code: `<img alt="foobar" aria-checked />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-supports-aria-props",
				Message:   errorMessage("aria-checked", "img", "img", true),
				Line:      1, Column: 1,
			}},
		},
		{
			Code: `<menu type="toolbar" aria-checked />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-supports-aria-props",
				Message:   errorMessage("aria-checked", "toolbar", "menu", true),
				Line:      1, Column: 1,
			}},
		},
		{
			Code: `<aside aria-checked />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-supports-aria-props",
				Message:   errorMessage("aria-checked", "complementary", "aside", true),
				Line:      1, Column: 1,
			}},
		},
		{
			Code: `<ul aria-expanded />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-supports-aria-props",
				Message:   errorMessage("aria-expanded", "list", "ul", true),
				Line:      1, Column: 1,
			}},
		},
		{
			Code: `<details aria-expanded />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-supports-aria-props",
				Message:   errorMessage("aria-expanded", "group", "details", true),
				Line:      1, Column: 1,
			}},
		},
		{
			Code: `<dialog aria-expanded />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-supports-aria-props",
				Message:   errorMessage("aria-expanded", "dialog", "dialog", true),
				Line:      1, Column: 1,
			}},
		},
		{
			Code: `<aside aria-expanded />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-supports-aria-props",
				Message:   errorMessage("aria-expanded", "complementary", "aside", true),
				Line:      1, Column: 1,
			}},
		},
		{
			Code: `<article aria-expanded />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-supports-aria-props",
				Message:   errorMessage("aria-expanded", "article", "article", true),
				Line:      1, Column: 1,
			}},
		},
		{
			Code: `<body aria-expanded />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-supports-aria-props",
				Message:   errorMessage("aria-expanded", "document", "body", true),
				Line:      1, Column: 1,
			}},
		},
		{
			Code: `<li aria-expanded />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-supports-aria-props",
				Message:   errorMessage("aria-expanded", "listitem", "li", true),
				Line:      1, Column: 1,
			}},
		},
		{
			Code: `<nav aria-expanded />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-supports-aria-props",
				Message:   errorMessage("aria-expanded", "navigation", "nav", true),
				Line:      1, Column: 1,
			}},
		},
		{
			Code: `<ol aria-expanded />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-supports-aria-props",
				Message:   errorMessage("aria-expanded", "list", "ol", true),
				Line:      1, Column: 1,
			}},
		},
		{
			Code: `<output aria-expanded />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-supports-aria-props",
				Message:   errorMessage("aria-expanded", "status", "output", true),
				Line:      1, Column: 1,
			}},
		},
		{
			Code: `<section aria-expanded />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-supports-aria-props",
				Message:   errorMessage("aria-expanded", "region", "section", true),
				Line:      1, Column: 1,
			}},
		},
		{
			Code: `<tbody aria-expanded />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-supports-aria-props",
				Message:   errorMessage("aria-expanded", "rowgroup", "tbody", true),
				Line:      1, Column: 1,
			}},
		},
		{
			Code: `<tfoot aria-expanded />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-supports-aria-props",
				Message:   errorMessage("aria-expanded", "rowgroup", "tfoot", true),
				Line:      1, Column: 1,
			}},
		},
		{
			Code: `<thead aria-expanded />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-supports-aria-props",
				Message:   errorMessage("aria-expanded", "rowgroup", "thead", true),
				Line:      1, Column: 1,
			}},
		},
		{
			Code: `<input type="radio" aria-invalid />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-supports-aria-props",
				Message:   errorMessage("aria-invalid", "radio", "input", true),
				Line:      1, Column: 1,
			}},
		},
		{
			Code: `<input type="radio" aria-selected />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-supports-aria-props",
				Message:   errorMessage("aria-selected", "radio", "input", true),
				Line:      1, Column: 1,
			}},
		},
		{
			Code: `<input type="radio" aria-haspopup />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-supports-aria-props",
				Message:   errorMessage("aria-haspopup", "radio", "input", true),
				Line:      1, Column: 1,
			}},
		},
		{
			Code: `<input type="checkbox" aria-haspopup />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-supports-aria-props",
				Message:   errorMessage("aria-haspopup", "checkbox", "input", true),
				Line:      1, Column: 1,
			}},
		},
		{
			Code: `<input type="reset" aria-invalid />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-supports-aria-props",
				Message:   errorMessage("aria-invalid", "button", "input", true),
				Line:      1, Column: 1,
			}},
		},
		{
			Code: `<input type="submit" aria-invalid />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-supports-aria-props",
				Message:   errorMessage("aria-invalid", "button", "input", true),
				Line:      1, Column: 1,
			}},
		},
		{
			Code: `<input type="image" aria-invalid />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-supports-aria-props",
				Message:   errorMessage("aria-invalid", "button", "input", true),
				Line:      1, Column: 1,
			}},
		},
		{
			Code: `<input type="button" aria-invalid />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-supports-aria-props",
				Message:   errorMessage("aria-invalid", "button", "input", true),
				Line:      1, Column: 1,
			}},
		},
		{
			Code: `<menuitem type="command" aria-invalid />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-supports-aria-props",
				Message:   errorMessage("aria-invalid", "menuitem", "menuitem", true),
				Line:      1, Column: 1,
			}},
		},
		{
			Code: `<menuitem type="radio" aria-selected />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-supports-aria-props",
				Message:   errorMessage("aria-selected", "menuitemradio", "menuitem", true),
				Line:      1, Column: 1,
			}},
		},
		{
			Code: `<menu type="toolbar" aria-haspopup />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-supports-aria-props",
				Message:   errorMessage("aria-haspopup", "toolbar", "menu", true),
				Line:      1, Column: 1,
			}},
		},
		{
			Code: `<menu type="toolbar" aria-invalid />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-supports-aria-props",
				Message:   errorMessage("aria-invalid", "toolbar", "menu", true),
				Line:      1, Column: 1,
			}},
		},
		{
			Code: `<menu type="toolbar" aria-expanded />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-supports-aria-props",
				Message:   errorMessage("aria-expanded", "toolbar", "menu", true),
				Line:      1, Column: 1,
			}},
		},
		{
			Code: `<link href="#" aria-invalid />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-supports-aria-props",
				Message:   errorMessage("aria-invalid", "link", "link", true),
				Line:      1, Column: 1,
			}},
		},
		{
			Code: `<area href="#" aria-invalid />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-supports-aria-props",
				Message:   errorMessage("aria-invalid", "link", "area", true),
				Line:      1, Column: 1,
			}},
		},
		{
			Code: `<a href="#" aria-invalid />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-supports-aria-props",
				Message:   errorMessage("aria-invalid", "link", "a", true),
				Line:      1, Column: 1,
			}},
		},
		// componentsSettings: Link → a; <Link href="#" aria-checked /> →
		// implicit role "link" on resolved element "a".
		{
			Code:     `<Link href="#" aria-checked />`,
			Tsx:      true,
			Settings: componentsSettings,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "role-supports-aria-props",
				Message:   errorMessage("aria-checked", "link", "a", true),
				Line:      1, Column: 1,
			}},
		},
	}

	// Append the auto-generated per-role × per-prop matrix. Mirrors
	// upstream's `validTests` / `invalidTests` from `createTests(...)`.
	gv, gi := generatedCases()
	valid = append(valid, gv...)
	invalid = append(invalid, gi...)

	rule_tester.RunRuleTesterBatched(fixtures.GetRootDir(), "tsconfig.json", t, &RoleSupportsAriaPropsRule, valid, invalid)
}
