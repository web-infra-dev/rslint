// TestJsxSortPropsExtras locks in tsgo shapes and upstream branches not fully
// represented by the upstream migration in jsx_sort_props_upstream_test.go.
package jsx_sort_props

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxSortPropsExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxSortPropsRule, []rule_tester.ValidTestCase{
		// ---- Dimension 4: JSX namespaced names are read as their complete name. ----
		{Code: `<App aria:a aria:b />`, Tsx: true},
		// ---- Dimension 4: a non-DOM component does not reserve dangerouslySetInnerHTML. ----
		{Code: `<App a dangerouslySetInnerHTML={{ __html: "x" }} />`, Tsx: true, Options: map[string]any{"reservedFirst": true}},
		// ---- Locks in upstream spread arm: a trailing spread does not report or panic. ----
		{Code: `<App b {...props} a />`, Tsx: true},
		// ---- Real-user: issue #3612 comment-bearing attribute blocks preserve comments. ----
		{Code: `<App a /* explanation */ b />`, Tsx: true},
		// ---- Real-user: issue #1632 gives reserved props precedence over callbacks. ----
		{Code: `<App key={1} a onClick={fn} />`, Tsx: true, Options: map[string]any{"reservedFirst": true, "callbacksLast": true}},
	}, []rule_tester.InvalidTestCase{
		// Locks in upstream alphabetic arm: errors are evaluated independently for each inversion.
		{Code: `<App c a b />`, Tsx: true, Output: []string{`<App a b c />`}, Errors: []rule_tester.InvalidTestCaseError{jsxSortError("sortPropsByAlpha", "Props should be sorted alphabetically", 1, 8), jsxSortError("sortPropsByAlpha", "Props should be sorted alphabetically", 1, 10)}},
		// Locks in upstream custom-reserved-list branch and the user-visible data substitution.
		{Code: `<App ref={ref} key={key} />`, Tsx: true, Options: map[string]any{"reservedFirst": []any{"key"}}, Output: []string{`<App key={key} ref={ref} />`}, Errors: []rule_tester.InvalidTestCaseError{jsxSortError("listReservedPropsFirst", "Reserved props must be listed before all other props", 1, 16)}},
		// ---- Dimension 4: comments move with the preceding prop in an autofix. ----
		{Code: `<App b /* b comment */ a />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{jsxSortError("sortPropsByAlpha", "Props should be sorted alphabetically", 1, 24)}},
		// ---- Dimension 4: member and intrinsic tag forms use the same attribute order. ----
		{Code: `<UI.Button b a />`, Tsx: true, Output: []string{`<UI.Button a b />`}, Errors: []rule_tester.InvalidTestCaseError{jsxSortError("sortPropsByAlpha", "Props should be sorted alphabetically", 1, 14)}},
		{Code: `<svg:path b a />`, Tsx: true, Output: []string{`<svg:path a b />`}, Errors: []rule_tester.InvalidTestCaseError{jsxSortError("sortPropsByAlpha", "Props should be sorted alphabetically", 1, 13)}},
		// ---- Real-user: issue #3936 trailing comments are included in the moved attribute. ----
		{Code: "<App\n  b // b comment\n  a\n/>", Tsx: true, Output: []string{"<App\n  a\n  b // b comment\n/>"}, Errors: []rule_tester.InvalidTestCaseError{jsxSortError("sortPropsByAlpha", "Props should be sorted alphabetically", 3, 3)}},
		// A comment before a spread must not let the fixer move either side across the spread boundary.
		{Code: `<App b /* b */ {...p} d c />`, Tsx: true, Output: []string{`<App b /* b */ {...p} c d />`}, Errors: []rule_tester.InvalidTestCaseError{jsxSortError("sortPropsByAlpha", "Props should be sorted alphabetically", 1, 25)}},
		// A trailing comment on the attribute absorbed by a preceding comment stays with that attribute.
		{Code: "<App\n b /* b */\n // leading a\n a // a\n c\n/>", Tsx: true, Output: []string{"<App\n c\n b /* b */\n // leading a\n a // a\n/>"}, Errors: []rule_tester.InvalidTestCaseError{jsxSortError("sortPropsByAlpha", "Props should be sorted alphabetically", 4, 2)}},
	})
}
