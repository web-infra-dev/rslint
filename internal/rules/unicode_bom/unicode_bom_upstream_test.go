package unicode_bom

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestUnicodeBomUpstream migrates the `never` half of the valid/invalid suite
// from upstream eslint/tests/lib/rules/unicode-bom.js. Position assertions
// cover line/column for every invalid case. The `always` cases have no
// counterpart here — see the package comment. rslint-specific lock-in cases
// live in unicode_bom_extras_test.go.
//
// Upstream asserts `endLine: void 0, endColumn: void 0` on every error
// because ESLint reports a `loc` with no `end`. rslint diagnostics always
// carry a range, so the zero-width range [0, 0) surfaces as end line 1,
// end column 1 — the same point as the start. Those assertions are made
// explicitly here.
func TestUnicodeBomUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&UnicodeBomRule,
		[]rule_tester.ValidTestCase{
			{Code: "var a = 123;", Options: []interface{}{"never"}},
			{Code: "var a = 123; \uFEFF", Options: []interface{}{"never"}},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   "\uFEFF var a = 123;",
				Output: []string{" var a = 123;"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpected",
					Message:   "Unexpected Unicode BOM (Byte Order Mark).",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 1,
				}},
			},
			{
				Code:    "\uFEFF var a = 123;",
				Options: []interface{}{"never"},
				Output:  []string{" var a = 123;"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpected",
					Message:   "Unexpected Unicode BOM (Byte Order Mark).",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 1,
				}},
			},
		},
	)
}
