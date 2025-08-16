package rules_of_hooks_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/react-hooks/rules/rules_of_hooks"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestRulesOfHooks(t *testing.T) {
	errors := make([]rule_tester.InvalidTestCaseError, 6)
	for i, err := range errors {
		err.MessageId = "import/no-self-import"
		err.Line = i + 2
		err.Column = 1
		errors[i] = err
	}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&rules_of_hooks.RulesOfHooks,
		[]rule_tester.ValidTestCase{
			{
				Code: `
// Valid because components can use hooks.
function ComponentWithHook() {
	useHook();
}
				`,
				FileName: "foo.ts",
			},
		},
		[]rule_tester.InvalidTestCase{},
	)
}
