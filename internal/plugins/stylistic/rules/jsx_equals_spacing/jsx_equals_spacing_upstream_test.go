// TestJsxEqualsSpacingUpstream migrates the full valid/invalid suite from
// upstream packages/eslint-plugin/rules/jsx-equals-spacing/
// jsx-equals-spacing.test.ts 1:1. Upstream asserts only messageId on its
// invalid cases; line/column/endLine/endColumn here are computed from the exact
// source the case carries (the report range is the `=` token, so column points
// at `=` and endColumn is one past it). rslint-specific lock-in cases live in
// jsx_equals_spacing_extras_test.go.
package jsx_equals_spacing

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxEqualsSpacingUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxEqualsSpacingRule, []rule_tester.ValidTestCase{
		// ---- default (no option == 'never') ----
		{Code: "<App />", Tsx: true},
		{Code: "<App foo />", Tsx: true},
		{Code: "<App foo=\"bar\" />", Tsx: true},
		{Code: "<App foo={e => bar(e)} />", Tsx: true},
		{Code: "<App {...props} />", Tsx: true},
		// ---- 'never' ----
		{Code: "<App />", Tsx: true, Options: []interface{}{"never"}},
		{Code: "<App foo />", Tsx: true, Options: []interface{}{"never"}},
		{Code: "<App foo=\"bar\" />", Tsx: true, Options: []interface{}{"never"}},
		{Code: "<App foo={e => bar(e)} />", Tsx: true, Options: []interface{}{"never"}},
		{Code: "<App {...props} />", Tsx: true, Options: []interface{}{"never"}},
		// ---- 'always' ----
		{Code: "<App />", Tsx: true, Options: []interface{}{"always"}},
		{Code: "<App foo />", Tsx: true, Options: []interface{}{"always"}},
		{Code: "<App foo = \"bar\" />", Tsx: true, Options: []interface{}{"always"}},
		{Code: "<App foo = {e => bar(e)} />", Tsx: true, Options: []interface{}{"always"}},
		{Code: "<App {...props} />", Tsx: true, Options: []interface{}{"always"}},
	}, []rule_tester.InvalidTestCase{
		// ---- 1. default: space on both sides ----
		{
			Code:   "<App foo = {bar} />",
			Tsx:    true,
			Output: []string{"<App foo={bar} />"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceBefore", Message: "There should be no space before '='", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
				{MessageId: "noSpaceAfter", Message: "There should be no space after '='", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
			},
		},
		// ---- 2. 'never': space on both sides ----
		{
			Code:    "<App foo = {bar} />",
			Tsx:     true,
			Options: []interface{}{"never"},
			Output:  []string{"<App foo={bar} />"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceBefore", Message: "There should be no space before '='", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
				{MessageId: "noSpaceAfter", Message: "There should be no space after '='", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
			},
		},
		// ---- 3. 'never': space before only ----
		{
			Code:    "<App foo ={bar} />",
			Tsx:     true,
			Options: []interface{}{"never"},
			Output:  []string{"<App foo={bar} />"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceBefore", Message: "There should be no space before '='", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
			},
		},
		// ---- 4. 'never': space after only ----
		{
			Code:    "<App foo= {bar} />",
			Tsx:     true,
			Options: []interface{}{"never"},
			Output:  []string{"<App foo={bar} />"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Message: "There should be no space after '='", Line: 1, Column: 9, EndLine: 1, EndColumn: 10},
			},
		},
		// ---- 5. 'never': two attributes, mixed ----
		{
			Code:    "<App foo= {bar} bar = {baz} />",
			Tsx:     true,
			Options: []interface{}{"never"},
			Output:  []string{"<App foo={bar} bar={baz} />"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Message: "There should be no space after '='", Line: 1, Column: 9, EndLine: 1, EndColumn: 10},
				{MessageId: "noSpaceBefore", Message: "There should be no space before '='", Line: 1, Column: 21, EndLine: 1, EndColumn: 22},
				{MessageId: "noSpaceAfter", Message: "There should be no space after '='", Line: 1, Column: 21, EndLine: 1, EndColumn: 22},
			},
		},
		// ---- 6. 'always': no space on either side ----
		{
			Code:    "<App foo={bar} />",
			Tsx:     true,
			Options: []interface{}{"always"},
			Output:  []string{"<App foo = {bar} />"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "needSpaceBefore", Message: "A space is required before '='", Line: 1, Column: 9, EndLine: 1, EndColumn: 10},
				{MessageId: "needSpaceAfter", Message: "A space is required after '='", Line: 1, Column: 9, EndLine: 1, EndColumn: 10},
			},
		},
		// ---- 7. 'always': space before only (missing after) ----
		{
			Code:    "<App foo ={bar} />",
			Tsx:     true,
			Options: []interface{}{"always"},
			Output:  []string{"<App foo = {bar} />"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "needSpaceAfter", Message: "A space is required after '='", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
			},
		},
		// ---- 8. 'always': space after only (missing before) ----
		{
			Code:    "<App foo= {bar} />",
			Tsx:     true,
			Options: []interface{}{"always"},
			Output:  []string{"<App foo = {bar} />"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "needSpaceBefore", Message: "A space is required before '='", Line: 1, Column: 9, EndLine: 1, EndColumn: 10},
			},
		},
		// ---- 9. 'always': two attributes, mixed ----
		{
			Code:    "<App foo={bar} bar ={baz} />",
			Tsx:     true,
			Options: []interface{}{"always"},
			Output:  []string{"<App foo = {bar} bar = {baz} />"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "needSpaceBefore", Message: "A space is required before '='", Line: 1, Column: 9, EndLine: 1, EndColumn: 10},
				{MessageId: "needSpaceAfter", Message: "A space is required after '='", Line: 1, Column: 9, EndLine: 1, EndColumn: 10},
				{MessageId: "needSpaceAfter", Message: "A space is required after '='", Line: 1, Column: 20, EndLine: 1, EndColumn: 21},
			},
		},
	})
}
