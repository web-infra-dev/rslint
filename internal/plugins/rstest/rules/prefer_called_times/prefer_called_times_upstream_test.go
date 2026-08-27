// TestPreferCalledTimesUpstream migrates the complete
// @vitest/eslint-plugin@v1.6.27 prefer-called-times suite. The three upstream
// invalid cases written with `toBeCalledOnce` move to the valid list because
// Rstest's assertion library does not expose that matcher; everything else is
// carried over 1:1. Rstest API sources, accessor forms, fix boundaries and
// edit-demand coverage live in prefer_called_times_extras_test.go.
package prefer_called_times

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferCalledTimesUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferCalledTimesRule,
		[]rule_tester.ValidTestCase{
			{Code: `expect(fn).toBeCalledTimes(1);`},
			{Code: `expect(fn).toHaveBeenCalledTimes(1);`},
			{Code: `expect(fn).toBeCalledTimes(2);`},
			{Code: `expect(fn).toHaveBeenCalledTimes(2);`},
			{Code: `expect(fn).toBeCalledTimes(expect.anything());`},
			{Code: `expect(fn).toHaveBeenCalledTimes(expect.anything());`},
			{Code: `expect(fn).not.toBeCalledTimes(2);`},
			{Code: `expect(fn).rejects.not.toBeCalledTimes(1);`},
			{Code: `expect(fn).not.toHaveBeenCalledTimes(1);`},
			{Code: `expect(fn).resolves.not.toHaveBeenCalledTimes(1);`},
			{Code: `expect(fn).toBeCalledTimes(0);`},
			{Code: `expect(fn).toHaveBeenCalledTimes(0);`},
			{Code: `expect(fn);`},
			// Upstream reports the next three. `toBeCalledOnce` does not exist
			// in @vitest/expect@4.1.10, so Rstest code that calls it is broken
			// rather than improvable, and this rule stays out of it.
			{Code: `expect(fn).toBeCalledOnce();`},
			{Code: `expect(fn).not.toBeCalledOnce();`},
			{Code: `expect(fn).resolves.toBeCalledOnce();`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `expect(fn).toHaveBeenCalledOnce();`,
				Output: []string{`expect(fn).toHaveBeenCalledTimes(1);`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferCalledTimes",
					Message:   "Prefer toHaveBeenCalledTimes(1)",
					Line:      1,
					Column:    12,
					EndLine:   1,
					EndColumn: 32,
				}},
			},
			{
				Code:   `expect(fn).not.toHaveBeenCalledOnce();`,
				Output: []string{`expect(fn).not.toHaveBeenCalledTimes(1);`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferCalledTimes",
					Message:   "Prefer toHaveBeenCalledTimes(1)",
					Line:      1,
					Column:    16,
					EndLine:   1,
					EndColumn: 36,
				}},
			},
			{
				Code:   `expect(fn).resolves.toHaveBeenCalledOnce();`,
				Output: []string{`expect(fn).resolves.toHaveBeenCalledTimes(1);`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferCalledTimes",
					Message:   "Prefer toHaveBeenCalledTimes(1)",
					Line:      1,
					Column:    21,
					EndLine:   1,
					EndColumn: 41,
				}},
			},
		},
	)
}
