package no_type_alias

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
)

func Test_NoTypeAlias(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoTypeAliasRule, []rule_tester.ValidTestCase{
		{Code: `
			interface Example {
				name: string;
			}
		`},
		{Code: `
			class Example {
				name: string;
			}
		`},
		{Code: `
			enum Example {
				A = 1,
				B = 2,
			}
		`},
		{Code: `
			function example(): string {
				return 'test';
			}
		`},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
				type Example = string;
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noTypeAlias",
					Line:      2,
					Column:    20,
					EndColumn: 26,
				},
			},
		},
		{
			Code: `
				type Example = number | string;
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noCompositionAlias",
					Line:      2,
					Column:    20,
					EndColumn: 26,
				},
				{
					MessageId: "noCompositionAlias",
					Line:      2,
					Column:    29,
					EndColumn: 35,
				},
			},
		},
		{
			Code: `
				type Example = {
					name: string;
				};
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noTypeAlias",
					Line:      2,
					Column:    20,
					EndColumn: 6,
				},
			},
		},
	})
}
