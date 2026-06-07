package prefer_called_with_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/prefer_called_with"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferCalledWithRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_called_with.PreferCalledWithRule,
		[]rule_tester.ValidTestCase{
			{Code: `expect(fn).toBeCalledWith();`},
			{Code: `expect(fn).toHaveBeenCalledWith();`},
			{Code: `expect(fn).toBeCalledWith(expect.anything());`},
			{Code: `expect(fn).toHaveBeenCalledWith(expect.anything());`},
			{Code: `expect(fn).not.toBeCalled();`},
			{Code: `expect(fn).rejects.not.toBeCalled();`},
			{Code: `expect(fn).not.toHaveBeenCalled();`},
			{Code: `expect(fn).not.toBeCalledWith();`},
			{Code: `expect(fn).not.toHaveBeenCalledWith();`},
			{Code: `expect(fn).resolves.not.toHaveBeenCalledWith();`},
			{Code: `expect(fn).toBeCalledTimes(0);`},
			{Code: `expect(fn).toHaveBeenCalledTimes(0);`},
			{Code: `expect(fn);`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `expect(fn).toBeCalled();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferCalledWith", Line: 1, Column: 12},
				},
			},
			{
				Code: `expect(fn).resolves.toBeCalled();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferCalledWith", Line: 1, Column: 21},
				},
			},
			{
				Code: `expect(fn).toHaveBeenCalled();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferCalledWith", Line: 1, Column: 12},
				},
			},
		},
	)
}
