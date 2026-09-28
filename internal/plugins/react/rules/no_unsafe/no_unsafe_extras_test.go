package no_unsafe

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnsafeExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnsafeRule, nil, []rule_tester.InvalidTestCase{
		// A literal superclass key identifies a React component.
		{
			Code: `
        class Foo extends React['Component'] {
          UNSAFE_componentWillMount() {}
        }
      `,
			Tsx:      true,
			Settings: settingsReact("16.4.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillMount", "componentDidMount"), Line: 3, Column: 11},
			},
		},
	})
}
