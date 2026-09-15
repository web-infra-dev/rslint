package prefer_to_have_been_called

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferToHaveBeenCalledUpstream(t *testing.T) {
	// eslint-plugin-jest v29.16.6; Rstest accepts not.resolves and retains extra arguments.
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferToHaveBeenCalledRule,
		[]rule_tester.ValidTestCase{
			{Code: `expect(method.mock.calls).toHaveLength;`},
			{Code: `expect(method).toBeCalledTimes();`},
			{Code: `expect(method.mock.calls).toHaveLength(0);`},
			{Code: `expect(method).toHaveBeenCalledTimes(1)`},
			{Code: `expect(method).not.toHaveBeenCalledTimes(x)`},
			{Code: `expect(method).not.toHaveBeenCalledTimes(1)`},
			{Code: `expect(method).not.toHaveBeenCalledTimes(...x)`},
			{Code: `expect(a);`},
			{Code: `expect(method).toBe([])`},
			{Code: `expect(fn.mock.calls).toEqual([])`},
			{Code: `expect(fn.mock.calls).toContain(1, 2, 3)`},
		},
		[]rule_tester.InvalidTestCase{
			{Code: `expect(method).toBeCalledTimes(0);`, Output: []string{`expect(method).not.toHaveBeenCalled();`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferMatcher", Line: 1, Column: 16, EndLine: 1, EndColumn: 31}}},
			{Code: `expect(method).not.toBeCalledTimes(0);`, Output: []string{`expect(method).toHaveBeenCalled();`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferMatcher", Line: 1, Column: 20, EndLine: 1, EndColumn: 35}}},
			{Code: `expect(method).toHaveBeenCalledTimes(0);`, Output: []string{`expect(method).not.toHaveBeenCalled();`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferMatcher", Line: 1, Column: 16, EndLine: 1, EndColumn: 37}}},
			{Code: `expect(method).not.toHaveBeenCalledTimes(0);`, Output: []string{`expect(method).toHaveBeenCalled();`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferMatcher", Line: 1, Column: 20, EndLine: 1, EndColumn: 41}}},
			{Code: `expect(method).not.toHaveBeenCalledTimes(0, 1, 2);`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferMatcher", Line: 1, Column: 20, EndLine: 1, EndColumn: 41}}},
			{Code: `expect(method).resolves.toHaveBeenCalledTimes(0);`, Output: []string{`expect(method).resolves.not.toHaveBeenCalled();`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferMatcher", Line: 1, Column: 25, EndLine: 1, EndColumn: 46}}},
			{Code: `expect(method).rejects.not.toHaveBeenCalledTimes(0);`, Output: []string{`expect(method).rejects.toHaveBeenCalled();`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferMatcher", Line: 1, Column: 28, EndLine: 1, EndColumn: 49}}},
			{Code: `expect(method).toBeCalledTimes(0 as number);`, Output: []string{`expect(method).not.toHaveBeenCalled();`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferMatcher", Line: 1, Column: 16, EndLine: 1, EndColumn: 31}}},
			{Code: `expect(method).not.resolves.toHaveBeenCalledTimes(0);`, Output: []string{`expect(method).resolves.toHaveBeenCalled();`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferMatcher", Line: 1, Column: 29, EndLine: 1, EndColumn: 50}}},
		},
	)
}
