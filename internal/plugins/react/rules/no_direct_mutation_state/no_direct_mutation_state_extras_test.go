package no_direct_mutation_state

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoDirectMutationStateExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoDirectMutationStateRule, nil, []rule_tester.InvalidTestCase{
		// Computed object methods returning JSX are components upstream.
		// Verified with eslint-plugin-react v7.37.5 using both parsers.
		{Code: `
        const obj = {
          ['Hello']() {
            this.state.x = 1;
            return <div/>;
          },
        };
      `, Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Message: "Do not mutate state directly. Use setState().", Line: 4, Column: 13, EndLine: 4, EndColumn: 23},
			}},
	})
}
