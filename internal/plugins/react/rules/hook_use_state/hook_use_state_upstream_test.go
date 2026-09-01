// This file migrates every test case from eslint-plugin-react v7.37.5's
// hook-use-state suite. rslint-specific shape and branch tests live in
// hook_use_state_extras_test.go.
package hook_use_state

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestHookUseStateUpstream(t *testing.T) {
	raw, err := rule_tester.LoadESLintTestSuiteFromJSON("testdata/hook_use_state_upstream_v7.37.5.json")
	if err != nil {
		t.Fatal(err)
	}
	suite := rule_tester.ConvertESLintTestSuite(raw)
	for i := range raw.Invalid {
		for j := range raw.Invalid[i].Errors {
			expected := &suite.Invalid[i].Errors[j]
			expected.Message = raw.Invalid[i].Errors[j].Message
			expected.MessageId = "useStateErrorMessage"
			if expected.Message == destructuredStateErrorText {
				expected.MessageId = "useStateErrorMessageOrAddOption"
			}
		}
	}
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&HookUseStateRule,
		suite.Valid,
		suite.Invalid,
	)
}
