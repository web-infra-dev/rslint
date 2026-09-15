package no_is_mounted

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoIsMountedExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoIsMountedRule, []rule_tester.ValidTestCase{
		{Code: `abstract class C { abstract [this.isMounted()](): void; }`, Tsx: true},
		{Code: `abstract class C { abstract get [this.isMounted()](): string; }`, Tsx: true},
		{Code: `abstract class C { abstract set [this.isMounted()](value: string); }`, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `declare class C { [this.isMounted()](): void; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noIsMounted", Line: 1, Column: 20, EndLine: 1, EndColumn: 34},
			},
		},
		{
			Code: `class C { [this.isMounted()](): void; [this.isMounted()]() {} }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noIsMounted", Line: 1, Column: 12, EndLine: 1, EndColumn: 26},
				{MessageId: "noIsMounted", Line: 1, Column: 40, EndLine: 1, EndColumn: 54},
			},
		},
	})
}
