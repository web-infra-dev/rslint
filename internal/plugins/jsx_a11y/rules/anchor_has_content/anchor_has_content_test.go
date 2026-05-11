package anchor_has_content

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// linkSettings mirrors the upstream `{ 'jsx-a11y': { components: { Link: 'a' } } }`
// settings used by the sole settings-based test case (Link → a remap).
var linkSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"components": map[string]interface{}{
			"Link": "a",
		},
	},
}

// TestAnchorHasContentUpstream covers the full valid/invalid suite migrated
// 1:1 from upstream eslint-plugin-jsx-a11y's
// `__tests__/src/rules/anchor-has-content-test.js`. rslint-specific lock-ins
// (semantic-walk branches, Dimension 4 universal edge shapes, tsgo AST quirks)
// live in anchor_has_content_extras_test.go.
func TestAnchorHasContentUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &AnchorHasContentRule, []rule_tester.ValidTestCase{
		// ---- Empty / structural negatives that should NOT match ----
		{Code: `<div />;`, Tsx: true},

		// ---- Content via children ----
		{Code: `<a>Foo</a>`, Tsx: true},
		{Code: `<a><Bar /></a>`, Tsx: true},
		{Code: `<a>{foo}</a>`, Tsx: true},
		{Code: `<a>{foo.bar}</a>`, Tsx: true},

		// ---- Content via fallback props ----
		{Code: `<a dangerouslySetInnerHTML={{ __html: "foo" }} />`, Tsx: true},
		{Code: `<a children={children} />`, Tsx: true},

		// ---- Custom component without remap → not matched ----
		{Code: `<Link />`, Tsx: true},

		// ---- componentMap remap: Link → a, with content ----
		{Code: `<Link>foo</Link>`, Tsx: true, Settings: linkSettings},

		// ---- Title / aria-label as alternative to content ----
		{Code: `<a title={title} />`, Tsx: true},
		{Code: `<a aria-label={ariaLabel} />`, Tsx: true},
		{Code: `<a title={title} aria-label={ariaLabel} />`, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ---- Bare empty anchor ----
		{
			Code: `<a />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Single hidden child (Bar with aria-hidden) → no accessible content ----
		{
			Code: `<a><Bar aria-hidden /></a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- {undefined} expression child → upstream's switch hits the
		//      JSXExpressionContainer/Identifier === 'undefined' false-arm ----
		{
			Code: `<a>{undefined}</a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- componentMap remap: Link → a, no content → invalid ----
		{
			Code:     `<Link />`,
			Tsx:      true,
			Settings: linkSettings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},
	})
}
