// TestStrictComputedAssignmentKey locks in rule behavior inside an evaluated
// computed property name of a destructuring assignment target. The shared
// traversal contract is covered by internal/linter; this test protects the
// rule's listener-to-diagnostic integration.
package strict

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestStrictComputedAssignmentKey(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&StrictRule,
		nil,
		[]rule_tester.InvalidTestCase{
			{
				Code:            `({ [(() => { 'use strict'; run(); })()]: target } = source);`,
				Options:         "never",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "never", Line: 1, Column: 14, EndLine: 1, EndColumn: 27},
				},
			},
		},
	)
}
