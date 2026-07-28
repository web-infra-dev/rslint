package require_to_throw_message_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/require_to_throw_message"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestRequireToThrowMessageUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&require_to_throw_message.RequireToThrowMessageRule,
		[]rule_tester.ValidTestCase{
			// ---- String ----
			{Code: "expect(() => { throw new Error('a'); }).toThrow('a');"},
			{Code: "expect(() => { throw new Error('a'); }).toThrowError('a');"},
			{Code: `
      test('string', async () => {
        const throwErrorAsync = async () => { throw new Error('a') };
        await expect(throwErrorAsync()).rejects.toThrow('a');
        await expect(throwErrorAsync()).rejects.toThrowError('a');
      })
    `},

			// ---- Template literal ----
			{Code: "const a = 'a'; expect(() => { throw new Error('a'); }).toThrow(`${a}`);"},
			{Code: "const a = 'a'; expect(() => { throw new Error('a'); }).toThrowError(`${a}`);"},
			{Code: `
      test('Template literal', async () => {
        const a = 'a';
        const throwErrorAsync = async () => { throw new Error('a') };
        await expect(throwErrorAsync()).rejects.toThrow(` + "`${a}`" + `);
        await expect(throwErrorAsync()).rejects.toThrowError(` + "`${a}`" + `);
      })
    `},

			// ---- Regex ----
			{Code: "expect(() => { throw new Error('a'); }).toThrow(/^a$/);"},
			{Code: "expect(() => { throw new Error('a'); }).toThrowError(/^a$/);"},
			{Code: `
      test('Regex', async () => {
        const throwErrorAsync = async () => { throw new Error('a') };
        await expect(throwErrorAsync()).rejects.toThrow(/^a$/);
        await expect(throwErrorAsync()).rejects.toThrowError(/^a$/);
      })
    `},

			// ---- Function ----
			{Code: "expect(() => { throw new Error('a'); }).toThrow((() => { return 'a'; })());"},
			{Code: "expect(() => { throw new Error('a'); }).toThrowError((() => { return 'a'; })());"},
			{Code: `
      test('Function', async () => {
        const throwErrorAsync = async () => { throw new Error('a') };
        const fn = () => { return 'a'; };
        await expect(throwErrorAsync()).rejects.toThrow(fn());
        await expect(throwErrorAsync()).rejects.toThrowError(fn());
      })
    `},

			// ---- Allow no message for `not`. ----
			{Code: "expect(() => { throw new Error('a'); }).not.toThrow();"},
			{Code: "expect(() => { throw new Error('a'); }).not.toThrowError();"},
			{Code: `
      test('Allow no message for "not"', async () => {
        const throwErrorAsync = async () => { throw new Error('a') };
        await expect(throwErrorAsync()).resolves.not.toThrow();
        await expect(throwErrorAsync()).resolves.not.toThrowError();
      })
    `},
			{Code: "expect(a);"},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Empty toThrow ----
			{
				Code: "expect(() => { throw new Error('a'); }).toThrow();",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "addErrorMessage", Line: 1, Column: 41},
				},
			},
			// ---- Empty toThrowError ----
			{
				Code: "expect(() => { throw new Error('a'); }).toThrowError();",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "addErrorMessage", Line: 1, Column: 41},
				},
			},
			// ---- Empty rejects.toThrow / rejects.toThrowError ----
			{
				Code: `test('empty rejects.toThrow', async () => {
  const throwErrorAsync = async () => { throw new Error('a') };
  await expect(throwErrorAsync()).rejects.toThrow();
  await expect(throwErrorAsync()).rejects.toThrowError();
})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "addErrorMessage", Line: 3, Column: 43},
					{MessageId: "addErrorMessage", Line: 4, Column: 43},
				},
			},
		},
	)
}
