// Supplements the unchanged upstream cases in prefer_called_with_test.go.
package prefer_called_with_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/prefer_called_with"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferCalledWithJestContract(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_called_with.PreferCalledWithRule,
		[]rule_tester.ValidTestCase{
			{Code: `expect(fn).toHaveBeenCalledOnce();`},
			{Code: `expect(fn).not.toHaveBeenCalled();`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `expect(fn).toBeCalled();`, Output: []string{},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledWith", Message: "Prefer toBeCalledWith(/* expected args */)", Line: 1, Column: 12, EndLine: 1, EndColumn: 22}},
			},
			{
				Code: `expect(fn)['toHaveBeenCalled']();`, Output: []string{},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledWith", Message: "Prefer toHaveBeenCalledWith(/* expected args */)", Line: 1, Column: 12, EndLine: 1, EndColumn: 30}},
			},
		})
}
