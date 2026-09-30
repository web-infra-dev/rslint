// TestNoAliasMethodsExtras pins which matcher names count as an alias. Only a
// matcher whose name is written statically is checked, and it is reported
// once, at the call that invokes it.
package no_alias_methods_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_alias_methods"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoAliasMethodsExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_alias_methods.NoAliasMethodsRule,
		[]rule_tester.ValidTestCase{
			// A computed identifier key names the matcher by the variable's
			// value, not its name, so none of these are alias uses.
			{Code: "const toBeCalled = 'toHaveBeenCalled';\nexpect(a)[toBeCalled]();"},
			{Code: "const toBeCalled = 'toBeCalled';\nexpect(a)[toBeCalled]();"},
			{Code: "const toBeCalled = 'toBeCalled';\nexpect(a)[(toBeCalled)]();"},
			{Code: "declare const toThrowError: string;\nexpect(a).not[toThrowError]();"},
			{Code: "declare const toReturn: string;\nexpect(a).resolves[toReturn]();"},
			{Code: "function check(toReturn: string) {\n  expect(a)[toReturn]();\n}"},
			{Code: "expect(a)[`${'toBeCalled'}`]();"},
		},
		[]rule_tester.InvalidTestCase{
			// A static bracket key is the matcher name itself.
			{
				Code:   "expect(a)['toBeCalled']();",
				Output: []string{"expect(a)['toHaveBeenCalled']();"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaceAlias", Line: 1, Column: 11},
				},
			},
			// Calls further along the chain do not report the matcher again.
			{
				Code:   "expect(a).toBeCalled().then(done);",
				Output: []string{"expect(a).toHaveBeenCalled().then(done);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaceAlias", Line: 1, Column: 11},
				},
			},
			{
				Code:   "expect(a).toBeCalled().foo;",
				Output: []string{"expect(a).toHaveBeenCalled().foo;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaceAlias", Line: 1, Column: 11},
				},
			},
			// Only the matcher is checked, not members read from its result.
			{
				Code:   "expect(a).toBeCalled().toReturn();",
				Output: []string{"expect(a).toHaveBeenCalled().toReturn();"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaceAlias", Line: 1, Column: 11},
				},
			},
		},
	)
}
