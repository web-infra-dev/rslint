// TestNoAliasMethodsUpstream migrates the complete
// eslint-plugin-vitest@v1.6.27 no-alias-methods suite. Rstest-only edge cases,
// real-user shapes, computed-key non-reporting policy, and edit-demand coverage
// live in no_alias_methods_extras_test.go.
package no_alias_methods

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoAliasMethodsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoAliasMethodsRule,
		[]rule_tester.ValidTestCase{
			{Code: `expect(a).toHaveBeenCalled()`},
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
			{Code: `expect(a).to.be.a("string");`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `expect(a).toBeCalled()`,
				Output: []string{`expect(a).toHaveBeenCalled()`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "replaceAlias",
					Message:   "Replace toBeCalled() with its canonical name of toHaveBeenCalled()",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 21,
				}},
			},
			{
				Code:   `expect(a).toBeCalledTimes()`,
				Output: []string{`expect(a).toHaveBeenCalledTimes()`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "replaceAlias",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 26,
				}},
			},
			{
				Code:   `expect(a).not["toThrowError"]()`,
				Output: []string{`expect(a).not["toThrow"]()`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "replaceAlias",
					Line:      1,
					Column:    15,
					EndLine:   1,
					EndColumn: 29,
				}},
			},
		},
	)
}
