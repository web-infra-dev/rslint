package no_type_alias

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoTypeAliasRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoTypeAliasRule, []rule_tester.ValidTestCase{
		{Code: `interface Foo { bar: string }`},
		{Code: `type Foo = string`, Options: map[string]interface{}{"allowAliases": "always"}},
		{
			Code:    `type Foo = 'a' | 'b'`,
			Options: map[string]interface{}{"allowAliases": "in-unions"},
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code:   `type Foo = string`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noTypeAlias", Line: 1, Column: 12}},
		},
		{
			Code:    `type Foo = string | number`,
			Options: map[string]interface{}{"allowAliases": "never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noCompositionAlias", Line: 1, Column: 12},
				{MessageId: "noCompositionAlias", Line: 1, Column: 21},
			},
		},
	})
}
