// Mirrors @vitest/eslint-plugin@1.6.27; Rstest-specific coverage is in prefer_called_with_extras_test.go.
package prefer_called_with

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferCalledWithUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferCalledWithRule,
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
			{Code: `expect(fn).toHaveBeenCalledExactlyOnceWith()`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `expect(fn).toBeCalled();`, Output: []string{`expect(fn).toBeCalledWith();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledWith", Message: "Prefer toBeCalledWith(/* expected args */)", Line: 1, Column: 12, EndLine: 1, EndColumn: 22}},
			},
			{
				Code: `expect(fn).resolves.toBeCalled();`, Output: []string{`expect(fn).resolves.toBeCalledWith();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledWith", Message: "Prefer toBeCalledWith(/* expected args */)", Line: 1, Column: 21, EndLine: 1, EndColumn: 31}},
			},
			{
				Code: `expect(fn).toHaveBeenCalled();`, Output: []string{`expect(fn).toHaveBeenCalledWith();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledWith", Message: "Prefer toHaveBeenCalledWith(/* expected args */)", Line: 1, Column: 12, EndLine: 1, EndColumn: 28}},
			},
			{
				Code: `it("some test", () => {expect(mockApi).toHaveBeenCalledOnce();});`, Output: []string{`it("some test", () => {expect(mockApi).toHaveBeenCalledExactlyOnceWith();});`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledWith", Message: "Prefer toHaveBeenCalledExactlyOnceWith(/* expected args */)", Line: 1, Column: 40, EndLine: 1, EndColumn: 60}},
			},
		})
}
