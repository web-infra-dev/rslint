package prefer_tag_over_role

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// expectedErr matches upstream's `expectedError(role, tag)` — the rule
// reports at the JsxOpeningElement / JsxSelfClosingElement; for top-level
// JSX expressions in these tests that's line 1, column 1.
func expectedErr(role, tag string) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "preferTagOverRole",
		Message:   errorMessage(tag, role),
		Line:      1,
		Column:    1,
	}
}

// TestPreferTagOverRoleUpstream mirrors the entire upstream test file
// (`__tests__/src/rules/prefer-tag-over-role-test.js` in
// eslint-plugin-jsx-a11y) — every `valid` and `invalid` entry is migrated
// 1:1. Lock-in tests for edge cases upstream doesn't cover (Dimensions 1–4,
// semantic-walk branches) live in
// `prefer_tag_over_role_extras_test.go` so this file stays trivially
// diff-comparable against future upstream updates.
func TestPreferTagOverRoleUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferTagOverRoleRule,
		[]rule_tester.ValidTestCase{
			// ---- valid ----
			{Code: `<div />;`, Tsx: true},
			{Code: `<div role="unknown" />;`, Tsx: true},
			{Code: `<div role="also unknown" />;`, Tsx: true},
			{Code: `<other />`, Tsx: true},
			{Code: `<img role="img" />`, Tsx: true},
			{Code: `<input role="checkbox" />`, Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			// ---- invalid ----
			{
				Code:   `<div role="checkbox" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("checkbox", `<input type="checkbox">`)},
			},
			{
				Code:   `<div role="button checkbox" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("checkbox", `<input type="checkbox">`)},
			},
			{
				Code:   `<div role="heading" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("heading", `<h1>, <h2>, <h3>, <h4>, <h5>, or <h6>`)},
			},
			{
				Code:   `<div role="link" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("link", `<a href=...>, or <area href=...>`)},
			},
			{
				Code:   `<div role="rowgroup" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("rowgroup", `<tbody>, <tfoot>, or <thead>`)},
			},
			{
				Code:   `<span role="checkbox" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("checkbox", `<input type="checkbox">`)},
			},
			{
				Code:   `<other role="checkbox" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("checkbox", `<input type="checkbox">`)},
			},
			{
				Code:   `<div role="banner" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("banner", `<header>`)},
			},
		},
	)
}
