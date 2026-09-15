// TestNoRestrictedMatchersUpstream migrates the complete
// @vitest/eslint-plugin@v1.6.27 no-restricted-matchers suite. Rstest source,
// parser, ordering, and range coverage lives in no_restricted_matchers_extras_test.go.
package no_restricted_matchers

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoRestrictedMatchersUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoRestrictedMatchersRule,
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
			{Code: `expect(a).resolves`, Options: restrictions("not", nil)},
			{Code: `expect(a).toBe(b)`, Options: restrictions("not.toBe", nil)},
			{Code: `expect(a).toBeUndefined(b)`, Options: restrictions("toBe", nil)},
			{Code: `expect(a)["toBe"](b)`, Options: restrictions("not.toBe", nil)},
			{Code: `expect(a).resolves.not.toBe(b)`, Options: restrictions("not", nil)},
			{Code: `expect(a).resolves.not.toBe(b)`, Options: restrictions("not.toBe", nil)},
			{Code: `expect(a).to.be.a("string").and.contain("hell")`, Options: restrictions("contain", nil)},
		},
		[]rule_tester.InvalidTestCase{
			invalid(`expect(a).not.toBe(b)`, restrictions("not", nil), "restrictedChain", "Use of `not` is disallowed", 1, 11, 1, 19),
			invalid(`expect(a).not.to.be.a("string")`, restrictions("not", nil), "restrictedChain", "", 1, 11, 1, 22),
			invalid(`expect(a).to.be.a("string")`, restrictions("to.be.a", nil), "restrictedChain", "", 1, 11, 1, 18),
			invalid(`expect(a).to.be.a("string").and.contain("hell")`, restrictions("to.be.a", nil), "restrictedChain", "", 1, 11, 1, 18),
			invalid(`expect(a).resolves.toBe(b)`, restrictions("resolves", nil), "restrictedChain", "", 1, 11, 1, 24),
			invalid(`expect(a).resolves.not.toBe(b)`, restrictions("resolves", nil), "restrictedChain", "", 1, 11, 1, 28),
			invalid(`expect(a).resolves.not.toBe(b)`, restrictions("resolves.not", nil), "restrictedChain", "", 1, 11, 1, 28),
			invalid(`expect(a).not.toBe(b)`, restrictions("not.toBe", nil), "restrictedChain", "", 1, 11, 1, 19),
			invalid(`expect(a).resolves.not.toBe(b)`, restrictions("resolves.not.toBe", nil), "restrictedChain", "", 1, 11, 1, 28),
			invalid(`expect(a).toBe(b)`, restrictions("toBe", "Prefer `toStrictEqual` instead"), "restrictedChainWithMessage", "Prefer `toStrictEqual` instead", 1, 11, 1, 15),
			invalid(`
       test('some test', async () => {
           await expect(Promise.resolve(1)).resolves.toBe(1);
        });
     `, restrictions("resolves", "Use `expect(await promise)` instead."), "restrictedChainWithMessage", "Use `expect(await promise)` instead.", 3, 45, 3, 58),
			invalid(`expect(Promise.resolve({})).rejects.toBeFalsy()`, restrictions("rejects.toBeFalsy", nil), "restrictedChain", "", 1, 29, 1, 46),
			invalid(`expect(uploadFileMock).not.toHaveBeenCalledWith('file.name')`, restrictions("not.toHaveBeenCalledWith", "Use not.toHaveBeenCalled instead"), "restrictedChainWithMessage", "Use not.toHaveBeenCalled instead", 1, 24, 1, 48),
		},
	)
}

func restrictions(chain string, message any) []any {
	return []any{map[string]any{chain: message}}
}

func invalid(
	code string,
	options []any,
	messageID string,
	message string,
	line int,
	column int,
	endLine int,
	endColumn int,
) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:    code,
		Options: options,
		Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: messageID,
			Message:   message,
			Line:      line,
			Column:    column,
			EndLine:   endLine,
			EndColumn: endColumn,
		}},
	}
}
