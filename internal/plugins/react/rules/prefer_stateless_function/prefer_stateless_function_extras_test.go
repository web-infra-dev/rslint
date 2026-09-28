package prefer_stateless_function

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferStatelessFunctionExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferStatelessFunctionRule, nil, []rule_tester.InvalidTestCase{
		// A literal superclass key identifies a React component.
		{
			Code: `
        class Foo extends React['Component'] {
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "componentShouldBePure", Line: 2, Column: 9},
			},
		},
	})
}
