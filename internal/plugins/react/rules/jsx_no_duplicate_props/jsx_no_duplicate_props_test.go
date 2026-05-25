package jsx_no_duplicate_props

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxNoDuplicatePropsRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxNoDuplicatePropsRule, []rule_tester.ValidTestCase{
		// ---- Upstream valid cases ----
		{Code: `<App />;`, Tsx: true},
		{Code: `<App {...this.props} />;`, Tsx: true},
		{Code: `<App a b c />;`, Tsx: true},
		{Code: `<App a b c A />;`, Tsx: true},
		{Code: `<App {...this.props} a b c />;`, Tsx: true},
		{Code: `<App c {...this.props} a b />;`, Tsx: true},
		{Code: `<App a="c" b="b" c="a" />;`, Tsx: true},
		{Code: `<App {...this.props} a="c" b="b" c="a" />;`, Tsx: true},
		{Code: `<App c="a" {...this.props} a="c" b="b" />;`, Tsx: true},
		{Code: `<App A a />;`, Tsx: true},
		{Code: `<App A b a />;`, Tsx: true},
		{Code: `<App A="a" b="b" B="B" />;`, Tsx: true},
		{
			Code:    `<App a:b="c" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"ignoreCase": true},
		},

		// ---- Additional edge cases ----
		// Non-self-closing element without duplicates.
		{Code: `<App a b></App>;`, Tsx: true},
		// Spread between distinct props is fine.
		{Code: `<App a {...x} b />;`, Tsx: true},
		// Default (case-sensitive): different case counts as different.
		{Code: `<App foo Foo FOO />;`, Tsx: true},
		// Namespaced attributes skipped entirely (even if textually duplicated).
		{Code: `<App a:b="1" a:b="2" />;`, Tsx: true},
		// Namespaced attr next to Identifier attr of the same local name — still
		// valid because namespaced is skipped and Identifier appears only once.
		{Code: `<App a:b="1" a="2" a:b="3" />;`, Tsx: true},
		// Nested JSX, each element's attributes are independent.
		{Code: `<Outer a><Inner b /></Outer>;`, Tsx: true},
		// JSX inside JSX expression container — the inner element is independent.
		{Code: `<Outer a={<Inner b />} b />;`, Tsx: true},
		// JSX Fragment wrapping duplicates-free children.
		{Code: `<><A a /><B b /></>;`, Tsx: true},
		// Member-access tag name (e.g. <Foo.Bar>) — rule acts on attributes only.
		{Code: `<Foo.Bar a b />;`, Tsx: true},
		// Boolean-shorthand attrs interleaved with equality-valued attrs.
		{Code: `<App a b={1} c="x" />;`, Tsx: true},
		// Only spread attributes — always valid.
		{Code: `<App {...a} {...b} />;`, Tsx: true},
		// Two identical spreads — still valid (spreads never participate).
		{Code: `<App {...a} {...a} />;`, Tsx: true},
		// Namespaced attrs stay unreported even under ignoreCase.
		{
			Code:    `<App a:b="1" A:B="2" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"ignoreCase": true},
		},
	}, []rule_tester.InvalidTestCase{
		// ---- Upstream invalid cases ----
		{
			Code: `<App a a />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noDuplicateProps",
					Message:   "No duplicate props allowed",
					Line:      1, Column: 8, EndLine: 1, EndColumn: 9,
				},
			},
		},
		{
			Code: `<App A b c A />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDuplicateProps", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
			},
		},
		{
			Code: `<App a="a" b="b" a="a" />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDuplicateProps", Line: 1, Column: 18, EndLine: 1, EndColumn: 23},
			},
		},
		{
			Code:    `<App A a />;`,
			Tsx:     true,
			Options: map[string]interface{}{"ignoreCase": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDuplicateProps", Line: 1, Column: 8, EndLine: 1, EndColumn: 9},
			},
		},
		{
			Code:    `<App a b c A />;`,
			Tsx:     true,
			Options: map[string]interface{}{"ignoreCase": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDuplicateProps", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
			},
		},
		{
			Code:    `<App A="a" b="b" B="B" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"ignoreCase": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDuplicateProps", Line: 1, Column: 18, EndLine: 1, EndColumn: 23},
			},
		},

		// ---- Additional edge cases ----
		// Duplicate inside a non-self-closing element.
		{
			Code: `<App a a></App>;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDuplicateProps", Line: 1, Column: 8, EndLine: 1, EndColumn: 9},
			},
		},
		// Spread does not reset duplicate tracking.
		{
			Code: `<App a {...x} a />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDuplicateProps", Line: 1, Column: 15, EndLine: 1, EndColumn: 16},
			},
		},
		// Three duplicates — each subsequent dup reports.
		{
			Code: `<App a a a />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDuplicateProps", Line: 1, Column: 8, EndLine: 1, EndColumn: 9},
				{MessageId: "noDuplicateProps", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
			},
		},
		// Multi-line: the duplicate is on a later line.
		{
			Code: "<App\n  a\n  a\n/>;",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDuplicateProps", Line: 3, Column: 3, EndLine: 3, EndColumn: 4},
			},
		},
		// Namespaced attrs are skipped, but Identifier duplicates around them
		// are still detected.
		{
			Code: `<App a:b="1" c c />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDuplicateProps", Line: 1, Column: 16, EndLine: 1, EndColumn: 17},
			},
		},
		// Nested JSX: inner element has duplicates, outer is fine.
		{
			Code: `<Outer a><Inner b b /></Outer>;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDuplicateProps", Line: 1, Column: 19, EndLine: 1, EndColumn: 20},
			},
		},
		// JSX inside JSX expression container — both outer and inner can report
		// independently, proving per-element scoping.
		{
			Code: `<Outer a a={<Inner b b />} />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDuplicateProps", Line: 1, Column: 10, EndLine: 1, EndColumn: 27},
				{MessageId: "noDuplicateProps", Line: 1, Column: 22, EndLine: 1, EndColumn: 23},
			},
		},
		// Member-access tag with duplicate props.
		{
			Code: `<Foo.Bar x x />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDuplicateProps", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
			},
		},
		// Fragment wrapping multiple elements — each checked independently.
		{
			Code: `<><A a a /><B b b /></>;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDuplicateProps", Line: 1, Column: 8, EndLine: 1, EndColumn: 9},
				{MessageId: "noDuplicateProps", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
			},
		},
		// Two distinct duplicate pairs — each dup reports.
		{
			Code: `<App a a b b />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDuplicateProps", Line: 1, Column: 8, EndLine: 1, EndColumn: 9},
				{MessageId: "noDuplicateProps", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
			},
		},
		// Comment between dup and rest does not affect column of the dup itself.
		{
			Code: `<App a a /* comment */ />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDuplicateProps", Line: 1, Column: 8, EndLine: 1, EndColumn: 9},
			},
		},
		// Reserved word used as attribute name is still an Identifier and checked.
		{
			Code: `<App class class />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDuplicateProps", Line: 1, Column: 12, EndLine: 1, EndColumn: 17},
			},
		},
		// Prototype-like attribute names — Go map does not false-negative via
		// prototype inheritance; upstream uses an ownership check to reach the same answer.
		{
			Code: `<App hasOwnProperty hasOwnProperty />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDuplicateProps", Line: 1, Column: 21, EndLine: 1, EndColumn: 35},
			},
		},
		// Boolean-shorthand attribute duplicated by a valued attribute of the
		// same name — name-only comparison, initializer is irrelevant.
		{
			Code: `<App disabled disabled={true} />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDuplicateProps", Line: 1, Column: 15, EndLine: 1, EndColumn: 30},
			},
		},
		// Expression-initialized attributes (numeric literal).
		{
			Code: `<App a={1} a={2} />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDuplicateProps", Line: 1, Column: 12, EndLine: 1, EndColumn: 17},
			},
		},
		// ignoreCase: three case-insensitive duplicates — two reports.
		{
			Code:    `<App foo Foo FOO />;`,
			Tsx:     true,
			Options: map[string]interface{}{"ignoreCase": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDuplicateProps", Line: 1, Column: 10, EndLine: 1, EndColumn: 13},
				{MessageId: "noDuplicateProps", Line: 1, Column: 14, EndLine: 1, EndColumn: 17},
			},
		},
	})
}
