// TestPreferCalledOnceUpstream migrates the complete
// @vitest/eslint-plugin@v1.6.27 prefer-called-once suite (tests/prefer-called-once.test.ts)
// 1:1. Position assertions cover line/column for every invalid case. The three
// upstream cases whose expected output is `toBeCalledOnce` keep their
// diagnostics but change output: Rstest's assertion library has no such
// matcher, so this port always writes `toHaveBeenCalledOnce`. Rstest API
// sources, accessor forms, fix boundaries and edit-demand coverage live in
// prefer_called_once_extras_test.go.
package prefer_called_once

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferCalledOnceUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferCalledOnceRule,
		[]rule_tester.ValidTestCase{
			{Code: `expect(fn).toBeCalledOnce();`},
			{Code: `expect(fn).toHaveBeenCalledOnce();`},
			{Code: `expect(fn).toBeCalledTimes(2);`},
			{Code: `expect(fn).toHaveBeenCalledTimes(2);`},
			{Code: `expect(fn).toBeCalledTimes(expect.anything());`},
			{Code: `expect(fn).toHaveBeenCalledTimes(expect.anything());`},
			{Code: `expect(fn).not.toBeCalledOnce();`},
			{Code: `expect(fn).rejects.not.toBeCalledOnce();`},
			{Code: `expect(fn).not.toHaveBeenCalledOnce();`},
			{Code: `expect(fn).resolves.not.toHaveBeenCalledOnce();`},
			{Code: `expect(fn).toBeCalledTimes(0);`},
			{Code: `expect(fn).toHaveBeenCalledTimes(0);`},
			{Code: `expect(fn);`},
		},
		[]rule_tester.InvalidTestCase{
			// Upstream's output here is `expect(fn).toBeCalledOnce();`, which
			// names a matcher @vitest/expect@4.1.10 does not define.
			{
				Code:   `expect(fn).toBeCalledTimes(1);`,
				Output: []string{`expect(fn).toHaveBeenCalledOnce();`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferCalledOnce",
					Message:   "Prefer toHaveBeenCalledOnce()",
					Line:      1,
					Column:    12,
					EndLine:   1,
					EndColumn: 27,
				}},
			},
			{
				Code:   `expect(fn).toHaveBeenCalledTimes(1);`,
				Output: []string{`expect(fn).toHaveBeenCalledOnce();`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferCalledOnce",
					Message:   "Prefer toHaveBeenCalledOnce()",
					Line:      1,
					Column:    12,
					EndLine:   1,
					EndColumn: 33,
				}},
			},
			{
				Code:   `expect(fn).not.toBeCalledTimes(1);`,
				Output: []string{`expect(fn).not.toHaveBeenCalledOnce();`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferCalledOnce",
					Message:   "Prefer toHaveBeenCalledOnce()",
					Line:      1,
					Column:    16,
					EndLine:   1,
					EndColumn: 31,
				}},
			},
			{
				Code:   `expect(fn).not.toHaveBeenCalledTimes(1);`,
				Output: []string{`expect(fn).not.toHaveBeenCalledOnce();`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferCalledOnce",
					Message:   "Prefer toHaveBeenCalledOnce()",
					Line:      1,
					Column:    16,
					EndLine:   1,
					EndColumn: 37,
				}},
			},
			{
				Code:   `expect(fn).resolves.toBeCalledTimes(1);`,
				Output: []string{`expect(fn).resolves.toHaveBeenCalledOnce();`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferCalledOnce",
					Message:   "Prefer toHaveBeenCalledOnce()",
					Line:      1,
					Column:    21,
					EndLine:   1,
					EndColumn: 36,
				}},
			},
			{
				Code:   `expect(fn).resolves.toHaveBeenCalledTimes(1);`,
				Output: []string{`expect(fn).resolves.toHaveBeenCalledOnce();`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferCalledOnce",
					Message:   "Prefer toHaveBeenCalledOnce()",
					Line:      1,
					Column:    21,
					EndLine:   1,
					EndColumn: 42,
				}},
			},
		},
	)
}
