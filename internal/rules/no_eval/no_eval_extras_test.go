// TestNoEvalComputedAssignmentKey locks in rule behavior inside an evaluated
// computed property name of a destructuring assignment target. The shared
// traversal contract is covered by internal/linter; this test protects the
// rule's listener-to-diagnostic integration.
package no_eval

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoEvalComputedAssignmentKey(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoEvalRule,
		nil,
		[]rule_tester.InvalidTestCase{
			{
				Code: `({ [(() => { eval(code); })()]: target } = source);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 14, EndLine: 1, EndColumn: 18},
				},
			},
		},
	)
}

func TestNoEvalCommonJSTopLevelThis(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoEvalRule,
		nil,
		[]rule_tester.InvalidTestCase{
			{
				Code:            `this.eval("x")`,
				LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
				Errors:          []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			{
				Code:     `() => this.eval("x")`,
				FileName: "top-level-this.cjs",
				TSConfig: "tsconfig.allow-js.json",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
		},
	)
}
