package no_useless_continue_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_useless_continue"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUselessContinueExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t,
		&no_useless_continue.NoUselessContinueRule,
		[]rule_tester.ValidTestCase{
			{Code: "for (const value of values as number[]) { if (value < 0) { continue; } consume(value); }", FileName: "case.ts"},
			{Code: "for (const value of values) { try { continue; } finally { cleanup(); } }", FileName: "case.js"},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: "for (const value of values as number[]) { consume(value); continue; }", FileName: "case.ts",
				Output: []string{"for (const value of values as number[]) { consume(value);  }"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-useless-continue", Message: "Unnecessary `continue` statement.", Line: 1, Column: 59, EndLine: 1, EndColumn: 68}},
			},
			{
				Code: "for (const value of values) {\r\n\tconsume(value);\r\n\tcontinue;\r\n}\r\n", FileName: "case.js",
				Output: []string{"for (const value of values) {\r\n\tconsume(value);\r\n}\r\n"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-useless-continue", Message: "Unnecessary `continue` statement.", Line: 3, Column: 2, EndLine: 3, EndColumn: 11}},
			},
		},
	)
}
