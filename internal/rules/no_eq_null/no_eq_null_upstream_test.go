// TestNoEqNullUpstream migrates the full valid/invalid suite from upstream
// eslint/tests/lib/rules/no-eq-null.js 1:1. Position assertions cover
// line/column/endLine/endColumn for every invalid case, mirroring upstream's
// own test file. rslint-specific lock-in cases live in
// no_eq_null_extras_test.go.
package no_eq_null

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoEqNullUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoEqNullRule,
		[]rule_tester.ValidTestCase{
			{Code: `if (x === null) { }`},
			{Code: `if (null === f()) { }`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `if (x == null) { }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 5, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code: `if (x != null) { }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 5, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code: `do {} while (null == x)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 14, EndLine: 1, EndColumn: 23},
				},
			},
		},
	)
}
