package no_alias_methods_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_alias_methods"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoAliasMethodsRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_alias_methods.NoAliasMethodsRule,
		[]rule_tester.ValidTestCase{
			{Code: "expect(a).toHaveBeenCalled()"},
			{Code: "expect(a).toHaveBeenCalledTimes()"},
			{Code: "expect(a).toHaveBeenCalledWith()"},
			{Code: "expect(a).toHaveBeenLastCalledWith()"},
			{Code: "expect(a).toHaveBeenNthCalledWith()"},
			{Code: "expect(a).toHaveReturned()"},
			{Code: "expect(a).toHaveReturnedTimes()"},
			{Code: "expect(a).toHaveReturnedWith()"},
			{Code: "expect(a).toHaveLastReturnedWith()"},
			{Code: "expect(a).toHaveNthReturnedWith()"},
			{Code: "expect(a).toThrow()"},
			{Code: "expect(a).rejects;"},
			{Code: "expect(a);"},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   "expect(a).toBeCalled()",
				Output: []string{"expect(a).toHaveBeenCalled()"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaceAlias", Column: 11, Line: 1},
				},
			},
			{
				Code:   "expect(a).toBeCalledTimes()",
				Output: []string{"expect(a).toHaveBeenCalledTimes()"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaceAlias", Column: 11, Line: 1},
				},
			},
			{
				Code:   "expect(a).toBeCalledWith()",
				Output: []string{"expect(a).toHaveBeenCalledWith()"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaceAlias", Column: 11, Line: 1},
				},
			},
			{
				Code:   "expect(a).lastCalledWith()",
				Output: []string{"expect(a).toHaveBeenLastCalledWith()"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaceAlias", Column: 11, Line: 1},
				},
			},
			{
				Code:   "expect(a).nthCalledWith()",
				Output: []string{"expect(a).toHaveBeenNthCalledWith()"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaceAlias", Column: 11, Line: 1},
				},
			},
			{
				Code:   "expect(a).toReturn()",
				Output: []string{"expect(a).toHaveReturned()"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaceAlias", Column: 11, Line: 1},
				},
			},
			{
				Code:   "expect(a).toReturnTimes()",
				Output: []string{"expect(a).toHaveReturnedTimes()"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaceAlias", Column: 11, Line: 1},
				},
			},
			{
				Code:   "expect(a).toReturnWith()",
				Output: []string{"expect(a).toHaveReturnedWith()"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaceAlias", Column: 11, Line: 1},
				},
			},
			{
				Code:   "expect(a).lastReturnedWith()",
				Output: []string{"expect(a).toHaveLastReturnedWith()"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaceAlias", Column: 11, Line: 1},
				},
			},
			{
				Code:   "expect(a).nthReturnedWith()",
				Output: []string{"expect(a).toHaveNthReturnedWith()"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaceAlias", Column: 11, Line: 1},
				},
			},
			{
				Code:   "expect(a).toThrowError()",
				Output: []string{"expect(a).toThrow()"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaceAlias", Column: 11, Line: 1},
				},
			},
			{
				Code:   "expect(a).resolves.toThrowError()",
				Output: []string{"expect(a).resolves.toThrow()"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaceAlias", Column: 20, Line: 1},
				},
			},
			{
				Code:   "expect(a).rejects.toThrowError()",
				Output: []string{"expect(a).rejects.toThrow()"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaceAlias", Column: 19, Line: 1},
				},
			},
			{
				Code:   "expect(a).not.toThrowError()",
				Output: []string{"expect(a).not.toThrow()"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaceAlias", Column: 15, Line: 1},
				},
			},
			{
				Code:   `expect(a).not["toThrowError"]()`,
				Output: []string{`expect(a).not["toThrow"]()`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaceAlias", Column: 15, Line: 1},
				},
			},
		},
	)
}
