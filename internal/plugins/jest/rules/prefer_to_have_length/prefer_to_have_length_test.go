package prefer_to_have_length_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/prefer_to_have_length"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferToHaveLengthRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_to_have_length.PreferToHaveLengthRule,
		[]rule_tester.ValidTestCase{
			{Code: "expect.hasAssertions"},
			{Code: "expect.hasAssertions()"},
			{Code: "expect(files).toHaveLength(1);"},
			{Code: "expect(files.name).toBe('file');"},
			{Code: "expect(files[`name`]).toBe('file');"},
			{Code: "expect(users[0]?.permissions?.length).toBe(1);"},
			{Code: "expect(result).toBe(true);"},
			{Code: "expect(user.getUserName(5)).resolves.toEqual('Paul')"},
			{Code: "expect(user.getUserName(5)).rejects.toEqual('Paul')"},
			{Code: "expect(a);"},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `expect(files["length"]).toBe(1);`,
				Output: []string{
					`expect(files).toHaveLength(1);`,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToHaveLength", Column: 25, Line: 1},
				},
			},
			{
				Code: `expect(files["length"]).toBe(1,);`,
				Output: []string{
					`expect(files).toHaveLength(1,);`,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToHaveLength", Column: 25, Line: 1},
				},
			},
			{
				Code: `expect(files["length"])["not"].toBe(1);`,
				Output: []string{
					`expect(files)["not"].toHaveLength(1);`,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToHaveLength", Column: 32, Line: 1},
				},
			},
			{
				Code: `expect(files["length"])["toBe"](1);`,
				Output: []string{
					`expect(files).toHaveLength(1);`,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToHaveLength", Column: 25, Line: 1},
				},
			},
			{
				Code: `expect(files["length"]).not["toBe"](1);`,
				Output: []string{
					`expect(files).not.toHaveLength(1);`,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToHaveLength", Column: 29, Line: 1},
				},
			},
			{
				Code: `expect(files["length"])["not"]["toBe"](1);`,
				Output: []string{
					`expect(files)["not"].toHaveLength(1);`,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToHaveLength", Column: 32, Line: 1},
				},
			},
			{
				Code: `expect(files.length).toBe(1);`,
				Output: []string{
					`expect(files).toHaveLength(1);`,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToHaveLength", Column: 22, Line: 1},
				},
			},
			{
				Code: `expect(files.length).toEqual(1);`,
				Output: []string{
					`expect(files).toHaveLength(1);`,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToHaveLength", Column: 22, Line: 1},
				},
			},
			{
				Code: `expect(files.length).toStrictEqual(1);`,
				Output: []string{
					`expect(files).toHaveLength(1);`,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToHaveLength", Column: 22, Line: 1},
				},
			},
			{
				Code: `expect(files.length).not.toStrictEqual(1);`,
				Output: []string{
					`expect(files).not.toHaveLength(1);`,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToHaveLength", Column: 26, Line: 1},
				},
			},
		},
	)
}
