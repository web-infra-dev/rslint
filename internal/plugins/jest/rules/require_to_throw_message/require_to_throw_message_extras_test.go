package require_to_throw_message_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/require_to_throw_message"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestRequireToThrowMessageExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&require_to_throw_message.RequireToThrowMessageRule,
		[]rule_tester.ValidTestCase{
			{Code: "expect(() => { throw new Error('a'); }).toThrow(Error);"},
			{Code: "expect(() => { throw new Error('a'); }).toThrow(undefined);"},
			{Code: "expect(() => { throw new Error('a'); }).toThrow(null);"},
			{Code: "expect(() => { throw new Error('a'); })['not'].toThrow();"},
			{Code: "expect(() => { throw new Error('a'); }).not['toThrow']();"},
			{Code: "await expect(Promise.reject(new Error('a'))).rejects.not.toThrow();"},
			{Code: "(expect(() => { throw new Error('a'); })).toThrow('a');"},
			{Code: "((expect(() => { throw new Error('a'); }))).not.toThrow();"},
			{Code: "expect(() => { throw new Error('a'); })!.toThrow('a');"},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: "expect(() => { throw new Error('a'); })['toThrow']();",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "addErrorMessage"},
				},
			},
			{
				Code: "expect(() => { throw new Error('a'); })[`toThrow`]();",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "addErrorMessage"},
				},
			},
			{
				Code: "expect(() => { throw new Error('a'); })['toThrowError']();",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "addErrorMessage"},
				},
			},
			{
				Code: "expect(() => { throw new Error('a'); }).toThrow?.();",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "addErrorMessage"},
				},
			},
			{
				Code: "expect(() => { throw new Error('a'); }).toThrowError?.();",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "addErrorMessage"},
				},
			},
			{
				Code: "(expect(() => { throw new Error('a'); })).toThrow();",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "addErrorMessage"},
				},
			},
			{
				Code: "await expect(Promise.reject()).rejects.toThrow();",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "addErrorMessage"},
				},
			},
		},
	)
}
