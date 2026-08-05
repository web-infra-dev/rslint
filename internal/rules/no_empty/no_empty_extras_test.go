// TestNoEmptyComputedAssignmentKey locks in rule behavior inside an evaluated
// computed property name of a destructuring assignment target. The shared
// traversal contract is covered by internal/linter; this test protects the
// rule's listener-to-diagnostic integration.
package no_empty

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoEmptyComputedAssignmentKey(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoEmptyRule,
		nil,
		[]rule_tester.InvalidTestCase{
			{
				Code: `({ [(() => { if (condition) {} })()]: target } = source);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 29, EndLine: 1, EndColumn: 31},
				},
			},
		},
	)
}
