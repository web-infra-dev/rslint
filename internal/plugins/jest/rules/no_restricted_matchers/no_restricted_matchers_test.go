package no_restricted_matchers_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_restricted_matchers"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoRestrictedMatchersRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_restricted_matchers.NoRestrictedMatchersRule,
		[]rule_tester.ValidTestCase{
			{Code: `expect(a).toHaveBeenCalled()`},
			{Code: `expect(a).not.toHaveBeenCalled()`},
			{Code: `expect(a).toHaveBeenCalledTimes()`},
			{Code: `expect(a).toHaveBeenCalledWith()`},
			{Code: `expect(a).toHaveBeenLastCalledWith()`},
			{Code: `expect(a).toHaveBeenNthCalledWith()`},
			{Code: `expect(a).toHaveReturned()`},
			{Code: `expect(a).toHaveReturnedTimes()`},
			{Code: `expect(a).toHaveReturnedWith()`},
			{Code: `expect(a).toHaveLastReturnedWith()`},
			{Code: `expect(a).toHaveNthReturnedWith()`},
			{Code: `expect(a).toThrow()`},
			{Code: `expect(a).rejects;`},
			{Code: `expect(a);`},
			{
				Code: `expect(a).resolves`,
				Options: []interface{}{
					map[string]interface{}{"not": nil},
				},
			},
			{
				Code: `expect(a).toBe(b)`,
				Options: []interface{}{
					map[string]interface{}{"not.toBe": nil},
				},
			},
			{
				Code: `expect(a).toBeUndefined(b)`,
				Options: []interface{}{
					map[string]interface{}{"toBe": nil},
				},
			},
			{
				Code: `expect(a)["toBe"](b)`,
				Options: []interface{}{
					map[string]interface{}{"not.toBe": nil},
				},
			},
			{
				Code: `expect(a).resolves.not.toBe(b)`,
				Options: []interface{}{
					map[string]interface{}{"not": nil},
				},
			},
			{
				Code: `expect(a).resolves.not.toBe(b)`,
				Options: []interface{}{
					map[string]interface{}{"not.toBe": nil},
				},
			},
			{
				Code: `expect(uploadFileMock).resolves.toHaveBeenCalledWith('file.name')`,
				Options: []interface{}{
					map[string]interface{}{"not.toHaveBeenCalledWith": "Use not.toHaveBeenCalled instead"},
				},
			},
			{
				Code: `expect(uploadFileMock).resolves.not.toHaveBeenCalledWith('file.name')`,
				Options: []interface{}{
					map[string]interface{}{"not.toHaveBeenCalledWith": "Use not.toHaveBeenCalled instead"},
				},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `expect(a).toBe(b)`,
				Options: []interface{}{
					map[string]interface{}{"toBe": nil},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "restrictedChain",
						Message:   "Use of `toBe` is disallowed",
						Line:      1,
						Column:    11,
					},
				},
			},
			{
				Code: `expect(a)["toBe"](b)`,
				Options: []interface{}{
					map[string]interface{}{"toBe": nil},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedChain", Line: 1, Column: 11},
				},
			},
			{
				Code: `expect(a).not[x]()`,
				Options: []interface{}{
					map[string]interface{}{"not": nil},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedChain", Line: 1, Column: 11},
				},
			},
			{
				Code: `expect(a).not.toBe(b)`,
				Options: []interface{}{
					map[string]interface{}{"not": nil},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedChain", Line: 1, Column: 11},
				},
			},
			{
				Code: `expect(a).resolves.toBe(b)`,
				Options: []interface{}{
					map[string]interface{}{"resolves": nil},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedChain", Line: 1, Column: 11},
				},
			},
			{
				Code: `expect(a).resolves.not.toBe(b)`,
				Options: []interface{}{
					map[string]interface{}{"resolves": nil},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedChain", Line: 1, Column: 11},
				},
			},
			{
				Code: `expect(a).resolves.not.toBe(b)`,
				Options: []interface{}{
					map[string]interface{}{"resolves.not": nil},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedChain", Line: 1, Column: 11},
				},
			},
			{
				Code: `expect(a).not.toBe(b)`,
				Options: []interface{}{
					map[string]interface{}{"not.toBe": nil},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedChain", Line: 1, Column: 11, EndColumn: 19},
				},
			},
			{
				Code: `expect(a).resolves.not.toBe(b)`,
				Options: []interface{}{
					map[string]interface{}{"resolves.not.toBe": nil},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedChain", Line: 1, Column: 11, EndColumn: 28},
				},
			},
			{
				Code: `expect(a).resolves.not.toBe(b)`,
				Options: []interface{}{
					map[string]interface{}{
						"resolves":     "Do not use resolves",
						"resolves.not": "Do not use resolves.not",
					},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "restrictedChainWithMessage",
						Message:   "Do not use resolves.not",
						Line:      1,
						Column:    11,
						EndColumn: 28,
					},
				},
			},
			{
				Code: `expect(a).toBe(b)`,
				Options: []interface{}{
					map[string]interface{}{"toBe": "Prefer `toStrictEqual` instead"},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "restrictedChainWithMessage",
						Message:   "Prefer `toStrictEqual` instead",
						Line:      1,
						Column:    11,
					},
				},
			},
			{
				Code: `
        test('some test', async () => {
          await expect(Promise.resolve(1)).resolves.toBe(1);
         });
      `,
				Options: []interface{}{
					map[string]interface{}{"resolves": "Use `expect(await promise)` instead."},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "restrictedChainWithMessage",
						Message:   "Use `expect(await promise)` instead.",
						Column:    44,
						EndColumn: 57,
					},
				},
			},
			{
				Code: `expect(Promise.resolve({})).rejects.toBeFalsy()`,
				Options: []interface{}{
					map[string]interface{}{"rejects.toBeFalsy": nil},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedChain", Column: 29, EndColumn: 46},
				},
			},
			{
				Code: `expect(uploadFileMock).not.toHaveBeenCalledWith('file.name')`,
				Options: []interface{}{
					map[string]interface{}{"not.toHaveBeenCalledWith": "Use not.toHaveBeenCalled instead"},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "restrictedChainWithMessage",
						Message:   "Use not.toHaveBeenCalled instead",
						Column:    24,
						EndColumn: 48,
					},
				},
			},
		},
	)
}
