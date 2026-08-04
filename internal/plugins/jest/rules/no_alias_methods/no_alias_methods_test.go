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
			// Leading-trivia cases. The matcher's Pos() sits at the end of the
			// preceding token, so a fix range built from it swallows whatever
			// whitespace or comment separates the dot from the matcher name.
			{
				Code:   "expect(a).\n  toBeCalled()",
				Output: []string{"expect(a).\n  toHaveBeenCalled()"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaceAlias", Column: 3, Line: 2},
				},
			},
			{
				// Control: the newline belongs to the dot rather than to the
				// matcher, so this shape was already correct.
				Code:   "expect(a)\n  .toBeCalled()",
				Output: []string{"expect(a)\n  .toHaveBeenCalled()"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaceAlias", Column: 4, Line: 2},
				},
			},
			{
				Code:   "expect(a) . /* c */ toBeCalled()",
				Output: []string{"expect(a) . /* c */ toHaveBeenCalled()"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaceAlias", Column: 21, Line: 1},
				},
			},
			{
				Code:   "expect(a). // c\n  toBeCalled()",
				Output: []string{"expect(a). // c\n  toHaveBeenCalled()"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaceAlias", Column: 3, Line: 2},
				},
			},
			{
				// The hand-written `Pos()+1 / End()-1` offsets used to land
				// inside the comment here, stripping `*` instead of the opening
				// quote and emitting `expect(a)[/toHaveBeenCalled']()`, which
				// does not parse.
				Code:   "expect(a)[/* c */ 'toBeCalled']()",
				Output: []string{"expect(a)[/* c */ 'toHaveBeenCalled']()"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaceAlias", Column: 19, Line: 1},
				},
			},
			{
				Code:   "expect(a)[\n  'toBeCalled'\n]()",
				Output: []string{"expect(a)[\n  'toHaveBeenCalled'\n]()"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaceAlias", Column: 3, Line: 2},
				},
			},
			{
				Code:   "expect(a)[`toBeCalled`]()",
				Output: []string{"expect(a)[`toHaveBeenCalled`]()"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaceAlias", Column: 11, Line: 1},
				},
			},
			{
				// Multiple aliases in one chain are each fixed in place.
				Code:   "expect(a).\n  resolves.\n  toBeCalledTimes(1)",
				Output: []string{"expect(a).\n  resolves.\n  toHaveBeenCalledTimes(1)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaceAlias", Column: 3, Line: 3},
				},
			},
		},
	)
}
