package no_debugger

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoDebuggerRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoDebuggerRule,
		// Valid cases - ported from ESLint
		[]rule_tester.ValidTestCase{
			{Code: `var test = { debugger: 1 }; test.debugger;`},
		},
		// Invalid cases - ported from ESLint
		[]rule_tester.InvalidTestCase{
			{
				Code: `if (foo) debugger`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-debugger", Line: 1, Column: 10},
				},
			},
		},
	)
}
