package no_duplicate_enum_values

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoDuplicateEnumValuesRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoDuplicateEnumValuesRule,
		[]rule_tester.ValidTestCase{
			{Code: `enum E { A }`},
			{Code: `enum E { A = 1, B }`},
			{Code: `enum E { A = 1, B = 2 }`},
			{Code: `enum E { A = 'A', B = 'B' }`},
			{Code: `enum E { A = 'A', B }`},
			{Code: `enum E { A = 'A', B = 1 + 1 }`},
			{Code: `enum E { A = 1, B = 2, C = 3 }`},
			{Code: `enum E { A = 'foo', B = 'bar' }`},
			{Code: "enum E { A = '', B = 0 }"},
			{Code: `enum E { A = 0, B = -0 }`},
			{Code: `enum E { A = NaN }`},
			{Code: "const x = 'A'; enum E { A = `${x}` }"},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `enum E { A = 1, B = 1 }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "duplicateValue",
						Line:      1,
						Column:    17,
					},
				},
			},
			{
				Code: `enum E { A = 'A', B = 'A' }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "duplicateValue",
						Line:      1,
						Column:    19,
					},
				},
			},
			{
				Code: `enum E { A = 'A', B = 'A', C = 1, D = 1 }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "duplicateValue",
						Line:      1,
						Column:    19,
					},
					{
						MessageId: "duplicateValue",
						Line:      1,
						Column:    35,
					},
				},
			},
			{
				Code: "enum E { A = 'A', B = `A` }",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "duplicateValue",
						Line:      1,
						Column:    19,
					},
				},
			},
			{
				Code: "enum E { A = `A`, B = `A` }",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "duplicateValue",
						Line:      1,
						Column:    19,
					},
				},
			},
		},
	)
}
