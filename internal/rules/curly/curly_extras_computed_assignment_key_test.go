// TestCurlyComputedAssignmentKey locks in rule behavior inside an evaluated
// computed property name of a destructuring assignment target. The shared
// traversal contract is covered by internal/linter; this test protects the
// rule's listener-to-diagnostic integration.
package curly

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestCurlyComputedAssignmentKey(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&CurlyRule,
		nil,
		[]rule_tester.InvalidTestCase{
			{
				Code:   `({ [(() => { if (condition) action(); })()]: target } = source);`,
				Output: []string{`({ [(() => { if (condition) {action();} })()]: target } = source);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfterCondition", Line: 1, Column: 29, EndLine: 1, EndColumn: 38},
				},
			},
		},
	)
}
