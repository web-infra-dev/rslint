// TestOneVarComputedAssignmentKey locks in rule behavior inside an evaluated
// computed property name of a destructuring assignment target. The shared
// traversal contract is covered by internal/linter; this test protects the
// rule's listener-to-diagnostic integration.
package one_var

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestOneVarComputedAssignmentKey(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&OneVarRule,
		nil,
		[]rule_tester.InvalidTestCase{
			{
				Code: `({ [(() => {
switch (value) { case 1: var first; var second; break; }
})()]: target } = source);`,
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combine", Line: 2, Column: 37, EndLine: 2, EndColumn: 48},
				},
			},
		},
	)
}
