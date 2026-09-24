// TestPreferMockPromiseShorthandUpstream migrates every valid and invalid case
// from @vitest/eslint-plugin@v1.6.27 tests/prefer-mock-promise-shorthand.test.ts,
// with the mock utilities object written as Rstest spells it. Rstest call
// shapes, tsgo edit shapes and branch lock-ins live in the extras suite.
//
// One upstream case is deliberately reversed: a resolved object literal is
// reported without a fix. See the rule source for why.
package prefer_mock_promise_shorthand

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
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
			{Code: `rs.fn().mockResolvedValue(42)`},
			{Code: `rs.fn(() => Promise.resolve(42))`},
			{Code: `rs.fn(() => Promise.reject(42))`},
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
			{Code: `rs.spyOn(Thingy, 'method').mockImplementation(param => Promise.resolve(param));`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `rs.fn().mockImplementation(() => Promise.resolve(42))`,
				Output: []string{`rs.fn().mockResolvedValue(42)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockResolvedValue", Line: 1, Column: 9},
				},
			},
			{
				Code:   `rs.fn().mockImplementation(() => Promise.reject(42))`,
				Output: []string{`rs.fn().mockRejectedValue(42)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockRejectedValue", Line: 1, Column: 9},
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
				Code:   `aVariable.mockImplementation(() => { return Promise.resolve(42) })`,
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
				Code:   `rs.fn().mockReturnValue(Promise.resolve(42))`,
				Output: []string{`rs.fn().mockResolvedValue(42)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockResolvedValue", Line: 1, Column: 9},
				},
			},
			{
				Code:   `rs.fn().mockReturnValue(Promise.reject(42))`,
				Output: []string{`rs.fn().mockRejectedValue(42)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockRejectedValue", Line: 1, Column: 9},
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
			// Reversed: upstream rewrites this to `mockResolvedValue({ ... })`, but a
			// fresh object literal checked directly against the settled type can fail
			// an excess-property check the promise passed, so only a primitive
			// literal is rewritten.
			{
				Code:   `aVariable.mockReturnValue(Promise.resolve({ target: 'world', message: 'hello' }))`,
				Output: []string{},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockResolvedValue", Line: 1, Column: 11},
				},
			},
			// A chain is visited from its outermost call inward, so the diagnostics
			// arrive last link first.
			{
				Code:   `aVariable.mockImplementation(() => Promise.reject(42)).mockImplementation(() => Promise.resolve(42)).mockReturnValue(Promise.reject(42))`,
				Output: []string{`aVariable.mockRejectedValue(42).mockResolvedValue(42).mockRejectedValue(42)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockRejectedValue", Line: 1, Column: 102},
					{MessageId: "useMockShorthand", Message: "Prefer mockResolvedValue", Line: 1, Column: 56},
					{MessageId: "useMockShorthand", Message: "Prefer mockRejectedValue", Line: 1, Column: 11},
				},
			},
			{
				Code:   `aVariable.mockReturnValueOnce(Promise.reject(42)).mockImplementation(() => Promise.resolve(42)).mockReturnValueOnce(Promise.reject(42))`,
				Output: []string{`aVariable.mockRejectedValueOnce(42).mockResolvedValue(42).mockRejectedValueOnce(42)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockRejectedValueOnce", Line: 1, Column: 97},
					{MessageId: "useMockShorthand", Message: "Prefer mockResolvedValue", Line: 1, Column: 51},
					{MessageId: "useMockShorthand", Message: "Prefer mockRejectedValueOnce", Line: 1, Column: 11},
				},
			},
			{
				Code:   `aVariable.mockReturnValueOnce(Promise.reject(new Error('oh noes!')))`,
				Output: []string{`aVariable.mockRejectedValueOnce(new Error('oh noes!'))`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockRejectedValueOnce", Line: 1, Column: 11},
				},
			},
			{
				Code:   `rs.fn().mockReturnValue(Promise.resolve(42), xyz)`,
				Output: []string{`rs.fn().mockResolvedValue(42, xyz)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockResolvedValue", Line: 1, Column: 9},
				},
			},
			{
				Code:   `rs.fn().mockImplementation(() => Promise.reject(42), xyz)`,
				Output: []string{`rs.fn().mockRejectedValue(42, xyz)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useMockShorthand", Message: "Prefer mockRejectedValue", Line: 1, Column: 9},
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
		},
	)
}
