// TestRequireToThrowMessageUpstream migrates the complete
// @vitest/eslint-plugin@v1.6.27 require-to-throw-message suite
// (tests/require-to-throw-message.test.ts) 1:1. Additional Rstest sources and
// tsgo edge shapes live in require_to_throw_message_extras_test.go.
package require_to_throw_message

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestRequireToThrowMessageUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&RequireToThrowMessageRule,
		[]rule_tester.ValidTestCase{
			// ---- String ----
			{Code: "expect(() => { throw new Error('a'); }).toThrow('a');"},
			{Code: "expect(() => { throw new Error('a'); }).toThrowError('a');"},
			{Code: `test('string', async () => {
  const throwErrorAsync = async () => { throw new Error('a') };
  await expect(throwErrorAsync()).rejects.toThrow('a');
  await expect(throwErrorAsync()).rejects.toThrowError('a');
})`},

			// ---- Template literal ----
			{Code: "const a = 'a'; expect(() => { throw new Error('a'); }).toThrow(`${a}`);"},
			{Code: "const a = 'a'; expect(() => { throw new Error('a'); }).toThrowError(`${a}`);"},
			{Code: `test('Template literal', async () => {
  const a = 'a';
  const throwErrorAsync = async () => { throw new Error('a') };
  await expect(throwErrorAsync()).rejects.toThrow(` + "`${a}`" + `);
  await expect(throwErrorAsync()).rejects.toThrowError(` + "`${a}`" + `);
})`},

			// ---- Regular expression ----
			{Code: "expect(() => { throw new Error('a'); }).toThrow(/^a$/);"},
			{Code: "expect(() => { throw new Error('a'); }).toThrowError(/^a$/);"},
			{Code: `test('Regex', async () => {
  const throwErrorAsync = async () => { throw new Error('a') };
  await expect(throwErrorAsync()).rejects.toThrow(/^a$/);
  await expect(throwErrorAsync()).rejects.toThrowError(/^a$/);
})`},

			// ---- Function ----
			{Code: "expect(() => { throw new Error('a'); }).toThrow((() => { return 'a'; })());"},
			{Code: "expect(() => { throw new Error('a'); }).toThrowError((() => { return 'a'; })());"},
			{Code: `test('Function', async () => {
  const throwErrorAsync = async () => { throw new Error('a') };
  const fn = () => { return 'a'; };
  await expect(throwErrorAsync()).rejects.toThrow(fn());
  await expect(throwErrorAsync()).rejects.toThrowError(fn());
})`},

			// ---- Allow no message for not ----
			{Code: "expect(() => { throw new Error('a'); }).not.toThrow();"},
			{Code: "expect(() => { throw new Error('a'); }).not.toThrowError();"},
			{Code: `test('Allow no message for not', async () => {
  const throwErrorAsync = async () => { throw new Error('a') };
  await expect(throwErrorAsync()).resolves.not.toThrow();
  await expect(throwErrorAsync()).resolves.not.toThrowError();
})`},
			{Code: "expect(a);"},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: "expect(() => { throw new Error('a'); }).toThrow();",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "addErrorMessage",
					Message:   "Add an error message to toThrow()",
					Line:      1,
					Column:    41,
					EndLine:   1,
					EndColumn: 48,
				}},
			},
			{
				Code: "expect(() => { throw new Error('a'); }).toThrowError();",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "addErrorMessage",
					Message:   "Add an error message to toThrowError()",
					Line:      1,
					Column:    41,
					EndLine:   1,
					EndColumn: 53,
				}},
			},
			{
				Code: `test('empty rejects.toThrow', async () => {
  const throwErrorAsync = async () => { throw new Error('a') };
  await expect(throwErrorAsync()).rejects.toThrow();
  await expect(throwErrorAsync()).rejects.toThrowError();
})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "addErrorMessage", Message: "Add an error message to toThrow()", Line: 3, Column: 43, EndLine: 3, EndColumn: 50},
					{MessageId: "addErrorMessage", Message: "Add an error message to toThrowError()", Line: 4, Column: 43, EndLine: 4, EndColumn: 55},
				},
			},
		},
	)
}
