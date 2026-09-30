// Supplements the upstream cases in prefer_to_have_been_called_test.go.
package prefer_to_have_been_called_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/prefer_to_have_been_called"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferToHaveBeenCalledJestContract(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{Code: `expect(fn).toHaveBeenCalledTimes();`},
		{Code: `expect(fn).toHaveReturnedTimes(0);`},
		{Code: `function run(expect: any) { expect(fn).toHaveBeenCalledTimes(0); }`},
	}
	for _, count := range []string{"-0", "+0", "0n", "'0'", "false", "null", "1 - 1"} {
		valid = append(valid, rule_tester.ValidTestCase{Code: "expect(fn).toHaveBeenCalledTimes(" + count + ");"})
	}

	invalid := []rule_tester.InvalidTestCase{
		{
			Code:   `import { expect as check } from '@jest/globals'; check(fn).toHaveBeenCalledTimes(0);`,
			Output: []string{`import { expect as check } from '@jest/globals'; check(fn).not.toHaveBeenCalled();`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferMatcher", Line: 1, Column: 60}},
		},
		{
			Code:   `await expect(promise).resolves.toHaveBeenCalledTimes(0);`,
			Output: []string{`await expect(promise).resolves.not.toHaveBeenCalled();`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferMatcher", Line: 1, Column: 32}},
		},
		{
			Code:   `const result = expect(fn).toHaveBeenCalledTimes(0);`,
			Output: []string{`const result = expect(fn).not.toHaveBeenCalled();`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferMatcher", Line: 1, Column: 27}},
		},
		{
			Code:   `expect(fn).toHaveBeenCalledTimes(0)();`,
			Output: []string{`expect(fn).not.toHaveBeenCalled()();`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferMatcher", Line: 1, Column: 12}},
		},
		{
			Code:   `expect(fn).toHaveBeenCalledTimes(0).then(done);`,
			Output: []string{`expect(fn).not.toHaveBeenCalled().then(done);`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferMatcher", Line: 1, Column: 12}},
		},
		{
			Code:   `expect(fn).toHaveBeenCalledTimes<number>(0,);`,
			Output: []string{`expect(fn).not.toHaveBeenCalled();`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferMatcher", Line: 1, Column: 12}},
		},
		{
			Code:   `expect(fn).toHaveBeenCalledTimes /* keep */ (0);`,
			Output: []string{`expect(fn).not.toHaveBeenCalled /* keep */ ();`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferMatcher", Line: 1, Column: 12}},
		},
	}
	for _, count := range []string{"0.0", "0x0", "0b0", "0o0", "0e2", "0 as const", "<number>0"} {
		invalid = append(invalid, rule_tester.InvalidTestCase{
			Code:   "expect(fn).toHaveBeenCalledTimes(" + count + ");",
			Output: []string{"expect(fn).not.toHaveBeenCalled();"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferMatcher", Line: 1, Column: 12}},
		})
	}
	// Removing additional arguments or comments would change observable behavior or lose source text.
	for _, code := range []string{
		`expect(fn).toHaveBeenCalledTimes(0, notify());`,
		`expect(fn).not.toHaveBeenCalledTimes(0, notify());`,
		`expect(fn).toHaveBeenCalledTimes(0, ...counts);`,
		`expect(fn).toHaveBeenCalledTimes(/* keep */ 0);`,
		`expect(fn).toHaveBeenCalledTimes(0 /* keep */);`,
		`expect(fn).toHaveBeenCalledTimes</* keep */ number>(0);`,
	} {
		invalid = append(invalid, rule_tester.InvalidTestCase{
			Code:   code,
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferMatcher"}},
		})
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_to_have_been_called.PreferToHaveBeenCalledRule, valid, invalid)
}
