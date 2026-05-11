// cspell:ignore fooaria klajsd labeledby skldjfaria

package aria_props

import (
	"fmt"
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/jsxa11yutil"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// bareMessage / withSuggestions mirror the upstream `errorMessage(name)`
// output exactly. Defined at package scope so aria_props_extras_test.go
// can reuse them without duplication.
//
// `${suggestions}` in upstream's template literal hits
// `Array.prototype.toString`, which joins with a bare comma (no space).
// Mirror byte-for-byte.
func bareMessage(name string) string {
	return name + ": This attribute is an invalid ARIA attribute."
}

func withSuggestions(name string, suggestions ...string) string {
	return bareMessage(name) + " Did you mean to use " + strings.Join(suggestions, ",") + "?"
}

// TestAriaPropsUpstream mirrors the full valid/invalid suite from upstream's
// `__tests__/src/rules/aria-props-test.js`, 1:1 and in upstream order so a
// future audit can grep across both side-by-side.
//
// Anything NOT in upstream's test file — case-sensitivity probes, AST-shape
// lockdowns, real-user typo / confusion patterns, library wrapper patterns,
// algorithm-extreme inputs, multi-element nesting — lives in
// aria_props_extras_test.go.
func TestAriaPropsUpstream(t *testing.T) {
	var validCases []rule_tester.ValidTestCase

	// Upstream's `basicValidityTests`:
	//   const basicValidityTests = ariaAttributes.map((prop) => ({
	//     code: `<div ${prop.toLowerCase()}="foobar" />`,
	//   }));
	// `prop` is already lowercase in aria-query's map; `.toLowerCase()` is
	// defensive. Iteration order matches upstream's `aria.keys()`.
	for _, prop := range jsxa11yutil.AriaPropertyNames {
		validCases = append(validCases, rule_tester.ValidTestCase{
			Code: fmt.Sprintf(`<div %s="foobar" />`, strings.ToLower(prop)),
			Tsx:  true,
		})
	}

	// Upstream's literal valid array — preserved in upstream order with the
	// same inline comments upstream maintainers wrote.
	validCases = append(validCases,
		// "Variables should pass, as we are only testing literals."
		rule_tester.ValidTestCase{Code: `<div />`, Tsx: true},
		rule_tester.ValidTestCase{Code: `<div></div>`, Tsx: true},
		// "Needs aria-*"
		rule_tester.ValidTestCase{Code: `<div aria="wee"></div>`, Tsx: true},
		rule_tester.ValidTestCase{Code: `<div abcARIAdef="true"></div>`, Tsx: true},
		rule_tester.ValidTestCase{Code: `<div fooaria-foobar="true"></div>`, Tsx: true},
		rule_tester.ValidTestCase{Code: `<div fooaria-hidden="true"></div>`, Tsx: true},
		rule_tester.ValidTestCase{Code: `<Bar baz />`, Tsx: true},
		rule_tester.ValidTestCase{Code: `<input type="text" aria-errormessage="foobar" />`, Tsx: true},
	)

	invalidCases := []rule_tester.InvalidTestCase{
		// `aria-` alone — passes the prefix check, fails `aria.has`, no
		// candidate within distance ≤ 2, so no suggestion suffix.
		{
			Code: `<div aria-="foobar" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaProp",
				Message:   bareMessage("aria-"),
			}},
		},
		// `aria-labeledby` for `aria-labelledby` — distance 1 (single
		// insertion), so the suggestion suffix lists `aria-labelledby`.
		{
			Code: `<div aria-labeledby="foobar" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaProp",
				Message:   withSuggestions("aria-labeledby", "aria-labelledby"),
			}},
		},
		// Long random garbage — passes the prefix check, fails the lookup,
		// no candidate within distance 2 (length-only delta already exceeds
		// the threshold), so no suggestion suffix.
		{
			Code: `<div aria-skldjfaria-klajsd="foobar" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidAriaProp",
				Message:   bareMessage("aria-skldjfaria-klajsd"),
			}},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &AriaPropsRule, validCases, invalidCases)
}
