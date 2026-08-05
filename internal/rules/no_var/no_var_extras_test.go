// TestNoVarComputedAssignmentKey locks in rule behavior inside an evaluated
// computed property name of a destructuring assignment target. The shared
// traversal contract is covered by internal/linter; this test protects the
// rule's listener-to-diagnostic integration.
package no_var

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoVarComputedAssignmentKey(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoVarRule,
		nil,
		[]rule_tester.InvalidTestCase{
			{
				// The self-reference deliberately makes var→let unsafe, so this
				// regression test asserts only diagnostics rather than edit behavior.
				Code: `({ [(() => { var value = value; })()]: target } = source);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedVar", Line: 1, Column: 14, EndLine: 1, EndColumn: 32},
				},
			},
		},
	)
}
