// TestPreferMockPromiseShorthandUpstream migrates every valid and invalid case
// from eslint-plugin-jest@v29.16.0
// src/rules/__tests__/prefer-mock-promise-shorthand.test.ts. tsgo edit shapes and
// branch lock-ins live in the Rstest rule's extras suite, which exercises the
// same shared engine.
package prefer_mock_promise_shorthand

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferMockPromiseShorthandUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferMockPromiseShorthandRule,
		[]rule_tester.ValidTestCase{
			{Code: `describe()`},
			{Code: `it()`},
			{Code: `describe.skip()`},
			{Code: `it.skip()`},
			{Code: `test()`},
			{Code: `test.skip()`},
			{Code: `var appliedOnly = describe.only; appliedOnly.apply(describe)`},
			{Code: `var calledOnly = it.only; calledOnly.call(it)`},
			{Code: `it.each()()`},
			{Code: "it.each`table`()"},
			{Code: `test.each()()`},
			{Code: "test.each`table`()"},
			{Code: `test.concurrent()`},
			{Code: `jest.fn().mockResolvedValue(42)`},
			{Code: `jest.fn(() => Promise.resolve(42))`},
			{Code: `jest.fn(() => Promise.reject(42))`},
			{Code: `aVariable.mockImplementation`},
			{Code: `aVariable.mockImplementation()`},
			{Code: `aVariable.mockImplementation([])`},
			{Code: `aVariable.mockImplementation(() => {})`},
			{Code: `aVariable.mockImplementation(() => [])`},
			{Code: `aVariable.mockReturnValue(() => Promise.resolve(1))`},
			{Code: `aVariable.mockReturnValue(Promise.resolve(1).then(() => 1))`},
			{Code: `aVariable.mockReturnValue(Promise.reject(1).then(() => 1))`},
			{Code: `aVariable.mockReturnValue(Promise.reject().then(() => 1))`},
			{Code: `aVariable.mockReturnValue(new Promise(resolve => resolve(1)))`},
			{Code: `aVariable.mockReturnValue(new Promise((_, reject) => reject(1)))`},
			{Code: `jest.spyOn(Thingy, 'method').mockImplementation(param => Promise.resolve(param));`},
			{Code: `aVariable.mockImplementation(() => {
  const value = new Date();

  return Promise.resolve(value);
});`},
			{Code: `aVariable.mockImplementation(() => {
  return Promise.resolve(value)
    .then(value => value + 1);
});`},
			{Code: `aVariable.mockImplementation(() => {
  return Promise.all([1, 2, 3]);
});`},
			{Code: `aVariable.mockImplementation(() => Promise.all([1, 2, 3]));`},
			{Code: `aVariable.mockReturnValue(Promise.all([1, 2, 3]));`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `jest.fn().mockImplementation(() => Promise.resolve(42))`,
				Output: []string{`jest.fn().mockResolvedValue(42)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockResolvedValue", Line: 1, Column: 11},
				},
			},
			{
				Code:   `jest.fn().mockImplementation(() => Promise.reject(42))`,
				Output: []string{`jest.fn().mockRejectedValue(42)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockRejectedValue", Line: 1, Column: 11},
				},
			},
			{
				Code:   `aVariable.mockImplementation(() => Promise.resolve(42))`,
				Output: []string{`aVariable.mockResolvedValue(42)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockResolvedValue", Line: 1, Column: 11},
				},
			},
			{
				Code: `aVariable.mockImplementation(() => {
  return Promise.resolve(42)
})`,
				Output: []string{`aVariable.mockResolvedValue(42)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockResolvedValue", Line: 1, Column: 11},
				},
			},
			{
				Code:   `aVariable.mockImplementation(() => Promise.reject(42))`,
				Output: []string{`aVariable.mockRejectedValue(42)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockRejectedValue", Line: 1, Column: 11},
				},
			},
			{
				Code:   `aVariable.mockImplementation(() => Promise.reject(42),)`,
				Output: []string{`aVariable.mockRejectedValue(42,)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockRejectedValue", Line: 1, Column: 11},
				},
			},
			{
				Code:   `aVariable.mockImplementationOnce(() => Promise.resolve(42))`,
				Output: []string{`aVariable.mockResolvedValueOnce(42)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockResolvedValueOnce", Line: 1, Column: 11},
				},
			},
			{
				Code:   `aVariable.mockImplementationOnce(() => Promise.reject(42))`,
				Output: []string{`aVariable.mockRejectedValueOnce(42)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockRejectedValueOnce", Line: 1, Column: 11},
				},
			},
			{
				Code:   `jest.fn().mockReturnValue(Promise.resolve(42))`,
				Output: []string{`jest.fn().mockResolvedValue(42)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockResolvedValue", Line: 1, Column: 11},
				},
			},
			{
				Code:   `jest.fn().mockReturnValue(Promise.reject(42))`,
				Output: []string{`jest.fn().mockRejectedValue(42)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockRejectedValue", Line: 1, Column: 11},
				},
			},
			{
				Code:   `aVariable.mockReturnValue(Promise.resolve(42))`,
				Output: []string{`aVariable.mockResolvedValue(42)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockResolvedValue", Line: 1, Column: 11},
				},
			},
			{
				Code:   `aVariable.mockReturnValue(Promise.reject(42))`,
				Output: []string{`aVariable.mockRejectedValue(42)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockRejectedValue", Line: 1, Column: 11},
				},
			},
			{
				Code:   `aVariable.mockReturnValueOnce(Promise.resolve(42))`,
				Output: []string{`aVariable.mockResolvedValueOnce(42)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockResolvedValueOnce", Line: 1, Column: 11},
				},
			},
			{
				Code:   `aVariable.mockReturnValueOnce(Promise.reject(42))`,
				Output: []string{`aVariable.mockRejectedValueOnce(42)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockRejectedValueOnce", Line: 1, Column: 11},
				},
			},
			{
				Code: `aVariable.mockReturnValue(Promise.resolve({
  target: 'world',
  message: 'hello'
}))`,
				Output: []string{`aVariable.mockResolvedValue({
  target: 'world',
  message: 'hello'
})`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockResolvedValue", Line: 1, Column: 11},
				},
			},
			// A chain is visited from its outermost call inward, so the diagnostics
			// arrive last link first.
			{
				Code: `aVariable
  .mockImplementation(() => Promise.reject(42))
  .mockImplementation(() => Promise.resolve(42))
  .mockReturnValue(Promise.reject(42))`,
				Output: []string{`aVariable
  .mockRejectedValue(42)
  .mockResolvedValue(42)
  .mockRejectedValue(42)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockRejectedValue", Line: 4, Column: 4},
					{MessageId: "useMockShorthand", Message: "Prefer mockResolvedValue", Line: 3, Column: 4},
					{MessageId: "useMockShorthand", Message: "Prefer mockRejectedValue", Line: 2, Column: 4},
				},
			},
			{
				Code: `aVariable
  .mockReturnValueOnce(Promise.reject(42))
  .mockImplementation(() => Promise.resolve(42))
  .mockReturnValueOnce(Promise.reject(42))`,
				Output: []string{`aVariable
  .mockRejectedValueOnce(42)
  .mockResolvedValue(42)
  .mockRejectedValueOnce(42)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockRejectedValueOnce", Line: 4, Column: 4},
					{MessageId: "useMockShorthand", Message: "Prefer mockResolvedValue", Line: 3, Column: 4},
					{MessageId: "useMockShorthand", Message: "Prefer mockRejectedValueOnce", Line: 2, Column: 4},
				},
			},
			{
				Code: `aVariable.mockReturnValueOnce(
  Promise.reject(
    new Error('oh noes!')
  )
)`,
				Output: []string{`aVariable.mockRejectedValueOnce(
  new Error('oh noes!')
)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockRejectedValueOnce", Line: 1, Column: 11},
				},
			},
			{
				Code:   `jest.fn().mockReturnValue(Promise.resolve(42), xyz)`,
				Output: []string{`jest.fn().mockResolvedValue(42, xyz)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockResolvedValue", Line: 1, Column: 11},
				},
			},
			{
				Code:   `jest.fn().mockImplementation(() => Promise.reject(42), xyz)`,
				Output: []string{`jest.fn().mockRejectedValue(42, xyz)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockRejectedValue", Line: 1, Column: 11},
				},
			},
			{
				Code:   `aVariable.mockReturnValueOnce(Promise.resolve(42, xyz))`,
				Output: []string{},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockResolvedValueOnce", Line: 1, Column: 11},
				},
			},
			{
				Code:   `aVariable.mockReturnValueOnce(Promise.resolve())`,
				Output: []string{`aVariable.mockResolvedValueOnce(undefined)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockResolvedValueOnce", Line: 1, Column: 11},
				},
			},
			{
				Code:   `jest.spyOn(fs, "readFile").mockReturnValue(Promise.reject(new Error("oh noes!")))`,
				Output: []string{`jest.spyOn(fs, "readFile").mockRejectedValue(new Error("oh noes!"))`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockRejectedValue", Line: 1, Column: 28},
				},
			},
		},
	)
}
