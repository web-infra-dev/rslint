// TestNoTernaryUpstream migrates the full valid/invalid suite from ESLint
// v10.8.1 tests/lib/rules/no-ternary.js 1:1. Position assertions cover
// line/column for every invalid case. rslint-specific lock-in cases live in
// no_ternary_extras_test.go.
package no_ternary

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoTernaryUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoTernaryRule,
		[]rule_tester.ValidTestCase{
			// ---- upstream valid ----
			{Code: `"x ? y";`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- upstream invalid ----
			{
				Code: "var foo = true ? thing : stuff;",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noTernaryOperator",
						Message:   "Ternary operator used.",
						Line:      1,
						Column:    11,
						EndLine:   1,
						EndColumn: 31,
					},
				},
			},
			{
				Code: "true ? thing() : stuff();",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noTernaryOperator",
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 25,
					},
				},
			},
			{
				Code: "function foo(bar) { return bar ? baz : qux; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noTernaryOperator",
						Line:      1,
						Column:    28,
						EndLine:   1,
						EndColumn: 43,
					},
				},
			},
		},
	)
}
