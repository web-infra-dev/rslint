package jsx_no_useless_fragment

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestJsxNoUselessFragmentExtrasThis covers the JSX positions where tsgo
// represents ESTree's JSXIdentifier("this") as KindThisKeyword. Other
// occurrences of `this`, including deeper members and namespace names, must
// keep their existing behavior. General rslint-added cases live in
// jsx_no_useless_fragment_extras_test.go; the upstream migration lives in
// jsx_no_useless_fragment_upstream_test.go.
func TestJsxNoUselessFragmentExtrasThis(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxNoUselessFragmentRule, []rule_tester.ValidTestCase{
		// ---- Dimension 4: `this` tag-name shapes that must not match ----
		// Only a single-level pragma.Fragment member can name a fragment.
		{Code: `<this.Namespace.Fragment>{foo}</this.Namespace.Fragment>`, Tsx: true, Settings: fragmentSettings("this", "")},
		{Code: `<this:path>{foo}</this:path>`, Tsx: true, Settings: fragmentSettings("this", "this")},
		// Locks in upstream isKeyedElement(): a key makes a custom fragment
		// useful, just as it does for the default React.Fragment spelling.
		{Code: `<this.Fragment key={item.id}>{item.value}</this.Fragment>`, Tsx: true, Settings: fragmentSettings("this", "")},
	}, []rule_tester.InvalidTestCase{
		// ---- Dimension 4: `this` in JSXIdentifier-equivalent positions ----
		{
			Code: "/** @jsx this.h */\n<this.Fragment>{value}</this.Fragment>;",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 2, Column: 1},
			},
		},
		{
			Code:     `<this.Fragment>{value}</this.Fragment>`,
			Tsx:      true,
			Settings: fragmentSettings("this", ""),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 1, Column: 1, EndLine: 1, EndColumn: 39},
			},
		},
		{
			Code:     `<this>{value}</this>`,
			Tsx:      true,
			Settings: fragmentSettings("", "this"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 1, Column: 1},
			},
		},
		// The member property remains an Identifier in tsgo; supporting the
		// object-side keyword must not change that existing match.
		{
			Code:     `<this.this>{value}</this.this>`,
			Tsx:      true,
			Settings: fragmentSettings("this", "this"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 1, Column: 1},
			},
		},
		{
			Code:     `<React.this>{value}</React.this>`,
			Tsx:      true,
			Settings: fragmentSettings("", "this"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 1, Column: 1},
			},
		},

		// ---- Regression: hashbang annotations precede comments and settings ----
		{
			Code: "#!/usr/bin/env node @jsx this.h\n<this.Fragment>{value}</this.Fragment>;",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 2, Column: 1},
			},
		},
		{
			Code:     "#!/usr/bin/env node @jsx Foo\n/* @jsx Bar */\n<Foo.Fragment>{value}</Foo.Fragment>;",
			Tsx:      true,
			Settings: fragmentSettings("Baz", ""),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 3, Column: 1},
			},
		},

		// ---- Locks in upstream isChildOfHtmlElement() for bare `this` ----
		// Member access is still a component parent: it gets no HTML-child
		// diagnostic and intentionally blocks the autofix.
		{
			Code: `<this.Widget><>foo</></this.Widget>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 1, Column: 14},
			},
		},
		{
			Code:   `<this><><A/><B/></></this>`,
			Tsx:    true,
			Output: []string{`<this><A/><B/></this>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "ChildOfHtmlElement", Message: childOfHtmlElementDescription, Line: 1, Column: 7, EndLine: 1, EndColumn: 20},
			},
		},
		{
			Code:   `<this><>foo</></this>`,
			Tsx:    true,
			Output: []string{`<this>foo</this>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 1, Column: 7},
				{MessageId: "ChildOfHtmlElement", Message: childOfHtmlElementDescription, Line: 1, Column: 7},
			},
		},
	})
}
