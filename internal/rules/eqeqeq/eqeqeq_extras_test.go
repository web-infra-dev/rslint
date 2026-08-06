// TestEqeqeqComputedAssignmentKey locks in rule behavior inside an evaluated
// computed property name of a destructuring assignment target. The shared
// traversal contract is covered by internal/linter; this test protects the
// rule's listener-to-diagnostic integration.
package eqeqeq

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestEqeqeqComputedAssignmentKey(t *testing.T) {
	const code = `({ [(() => { left == right; })()]: target } = source);`
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&EqeqeqRule,
		nil,
		[]rule_tester.InvalidTestCase{
			{
				Code: code,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Line:      1,
						Column:    19,
						EndLine:   1,
						EndColumn: 21,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "replaceOperator", Output: `({ [(() => { left === right; })()]: target } = source);`},
						},
					},
				},
			},
		},
	)
}
