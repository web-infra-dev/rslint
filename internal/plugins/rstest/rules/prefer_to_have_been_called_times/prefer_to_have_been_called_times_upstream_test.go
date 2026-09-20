package prefer_to_have_been_called_times

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferToHaveBeenCalledTimesUpstream(t *testing.T) {
	// eslint-plugin-jest v29.16.1 and @vitest/eslint-plugin v1.6.27 share this suite verbatim.
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferToHaveBeenCalledTimesRule,
		[]rule_tester.ValidTestCase{
			{Code: `expect.assertions(1)`},
			{Code: `expect(fn).toHaveBeenCalledTimes`},
			{Code: `expect(fn.mock.calls).toHaveLength`},
			{Code: `expect(fn.mock.values).toHaveLength(0)`},
			{Code: `expect(fn.values.calls).toHaveLength(0)`},
			{Code: `expect(fn).toHaveBeenCalledTimes(0)`},
			{Code: `expect(fn).resolves.toHaveBeenCalledTimes(10)`},
			{Code: `expect(fn).not.toHaveBeenCalledTimes(10)`},
			{Code: `expect(fn).toHaveBeenCalledTimes(1)`},
			{Code: `expect(fn).toBeCalledTimes(0);`},
			{Code: `expect(fn).toHaveBeenCalledTimes(0);`},
			{Code: `expect(fn);`},
			{Code: `expect(method.mock.calls[0][0]).toStrictEqual(value);`},
			{Code: `expect(fn.mock.length).toEqual(1);`},
			{Code: `expect(fn.mock.calls).toEqual([]);`},
			{Code: `expect(fn.mock.calls).toContain(1, 2, 3);`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `expect(method.mock.calls).toHaveLength(1);`,
				Output: []string{`expect(method).toHaveBeenCalledTimes(1);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferMatcher", Line: 1, Column: 27, EndLine: 1, EndColumn: 39}},
			},
			{
				Code:   `expect(method.mock.calls).resolves.toHaveLength(x);`,
				Output: []string{`expect(method).resolves.toHaveBeenCalledTimes(x);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferMatcher", Line: 1, Column: 36, EndLine: 1, EndColumn: 48}},
			},
			{
				Code:   `expect(method["mock"].calls).toHaveLength(0);`,
				Output: []string{`expect(method).toHaveBeenCalledTimes(0);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferMatcher", Line: 1, Column: 30, EndLine: 1, EndColumn: 42}},
			},
			{
				Code:   `expect(my.method.mock.calls).not.toHaveLength(0);`,
				Output: []string{`expect(my.method).not.toHaveBeenCalledTimes(0);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferMatcher", Line: 1, Column: 34, EndLine: 1, EndColumn: 46}},
			},
		},
	)
}
