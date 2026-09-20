// TestPreferToHaveLengthUpstream migrates every valid and invalid case from
// @vitest/eslint-plugin@v1.6.27 tests/prefer-to-have-length.test.ts. Rstest
// provenance, framework boundaries and tsgo edit shapes live in the extras suite.
package prefer_to_have_length

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferToHaveLengthUpstream(t *testing.T) {
	invalid := func(code string, column int) rule_tester.InvalidTestCase {
		return rule_tester.InvalidTestCase{
			Code: code,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "useToHaveLength", Message: "Use `toHaveLength()` instead",
				Line: 1, Column: column,
			}},
		}
	}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferToHaveLengthRule,
		[]rule_tester.ValidTestCase{
			{Code: `expect.hasAssertions`},
			{Code: `expect.hasAssertions()`},
			{Code: `expect(files).toHaveLength(1);`},
			{Code: `expect(files).to.be.a("array");`},
			{Code: `expect(files.name).toBe('file');`},
			{Code: "expect(files[`name`]).toBe('file');"},
			{Code: `expect(users[0]?.permissions?.length).toBe(1);`},
			{Code: `expect(result).toBe(true);`},
			{Code: `expect(user.getUserName(5)).resolves.toEqual('Paul')`},
			{Code: `expect(user.getUserName(5)).rejects.toEqual('Paul')`},
			{Code: `expect(a);`},
			{Code: `expect().toBe();`},
		},
		[]rule_tester.InvalidTestCase{
			// ADAPT: all upstream diagnostics are retained, but identifier
			// receivers are not fixed because `files.length` may invoke a getter.
			invalid(`expect(files["length"]).toBe(1);`, 25),
			invalid(`expect(files["length"]).toBe(1,);`, 25),
			invalid(`expect(files["length"])["not"].toBe(1);`, 32),
			invalid(`expect(files["length"])["toBe"](1);`, 25),
			invalid(`expect(files["length"]).not["toBe"](1);`, 29),
			invalid(`expect(files["length"])["not"]["toBe"](1);`, 32),
			invalid(`expect(files.length).toBe(1);`, 22),
			invalid(`expect(files.length).toEqual(1);`, 22),
			invalid(`expect(files.length).toStrictEqual(1);`, 22),
			invalid(`expect(files.length).not.toStrictEqual(1);`, 26),
		},
	)
}
