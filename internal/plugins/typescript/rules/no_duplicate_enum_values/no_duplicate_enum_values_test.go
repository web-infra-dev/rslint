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
			{Code: `enum E { A = -1, B = -2 }`},
			{Code: `enum E { A = +1, B = +2 }`},
			{Code: `enum E { A = +1, B = -1 }`},
			{Code: `enum E { A = 1, B = -1 }`},
			{Code: `enum E { A = 'A', B = 'B' }`},
			{Code: `enum E { A = 'A', B }`},
			{Code: `enum E { A = 'A', B = 1 + 1 }`},
			{Code: `enum E { A = 1, B = 2, C = 3 }`},
			{Code: `enum E { A = 'foo', B = 'bar' }`},
			{Code: "enum E { A = '', B = 0 }"},
			{Code: `enum E { A = 1, B = '1' }`},
			{Code: `enum E { A = -1, B = '-1' }`},
			{Code: `enum E { A = 0, B = -0 }`},
			{Code: `enum E { A = -0, B = +0 }`},
			{Code: `enum E { A = -0, B = 0 }`},
			{Code: `enum E { A = NaN }`},
			{Code: `enum E { A = NaN, B = NaN }`},
			{Code: `enum E { A = NaN, B = -NaN }`},
			{Code: `enum E { A = 'NaN', B = NaN }`},
			{Code: `enum E { A = -+-0, B = +-+0 }`},
			{Code: `enum E { A = -'', B = 0 }`},
			{Code: `enum E { A = Infinity, B = Infinity }`},
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
				Code: `enum E { A = -1, B = -1 }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "duplicateValue",
						Message:   "Duplicate enum member value -1.",
						Line:      1,
						Column:    18,
					},
				},
			},
			{
				Code: `enum E {
  A = +1,
  B = +1,
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "duplicateValue", Message: "Duplicate enum member value 1.", Line: 3, Column: 3, EndLine: 3, EndColumn: 9},
				},
			},
			{
				Code: `enum E {
  A = +0,
  B = 0,
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "duplicateValue", Message: "Duplicate enum member value 0.", Line: 3, Column: 3, EndLine: 3, EndColumn: 8},
				},
			},
			{
				Code: `enum E {
  A = -0,
  B = -0,
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "duplicateValue", Message: "Duplicate enum member value 0.", Line: 3, Column: 3, EndLine: 3, EndColumn: 9},
				},
			},
			{
				Code: `enum E {
  A = +'0',
  B = 0,
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "duplicateValue", Message: "Duplicate enum member value 0.", Line: 3, Column: 3, EndLine: 3, EndColumn: 8},
				},
			},
			{
				Code: `enum E {
  A = 0x10,
  B = 16,
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "duplicateValue", Message: "Duplicate enum member value 16.", Line: 3, Column: 3, EndLine: 3, EndColumn: 9},
				},
			},
			{
				Code: `enum E {
  A = +'1e2',
  B = 100,
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "duplicateValue", Message: "Duplicate enum member value 100.", Line: 3, Column: 3, EndLine: 3, EndColumn: 10},
				},
			},
			{
				Code: `enum E {
  A = +'',
  B = 0,
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "duplicateValue", Message: "Duplicate enum member value 0.", Line: 3, Column: 3, EndLine: 3, EndColumn: 8},
				},
			},
			{
				Code: `enum E {
  A = -+1,
  B = +-1,
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "duplicateValue", Message: "Duplicate enum member value -1.", Line: 3, Column: 3, EndLine: 3, EndColumn: 10},
				},
			},
			{
				Code: "enum E {\n  A = -`0`,\n  B = -0,\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "duplicateValue", Message: "Duplicate enum member value 0.", Line: 3, Column: 3, EndLine: 3, EndColumn: 9},
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
			{
				Code: `enum E {
  A = 1,
  B = '1',
  C = 1,
  D = '1',
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "duplicateValue",
						Line:      4,
						Column:    3,
					},
					{
						MessageId: "duplicateValue",
						Line:      5,
						Column:    3,
					},
				},
			},
			{
				Code: `enum E {
  A0 = 'same',
  A1 = 'same',
  A2 = '2',
  A3 = '3',
  A4 = '4',
  A5 = '5',
  A6 = '6',
  A7 = '7',
  A8 = '8',
  A9 = '9',
  A10 = 'same',
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "duplicateValue",
						Line:      3,
						Column:    3,
					},
					{
						MessageId: "duplicateValue",
						Line:      12,
						Column:    3,
					},
				},
			},
		},
	)
}
