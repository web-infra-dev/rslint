package no_duplicate_enum_values

// cspell:ignore CCMWEB

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoDuplicateEnumValuesAdversarial(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoDuplicateEnumValuesRule,
		[]rule_tester.ValidTestCase{
			{Code: `enum E { A = +'not a number', B = +'not a number' }`},
			{Code: `enum E { A = +'-0x1', B = -1 }`},
			{Code: `enum E { A = +(1 + 1), B = 2 }`},
			{Code: `enum E { A = +'-0', B = 0 }`},
			{Code: `enum E { A = '0', B = 0 }`},
			{Code: `enum E { A = 'Infinity', B = 1e9999 }`},
			{Code: `enum E { A = 1n, B = 1n }`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `enum E {
  ORIGINAL = 'app_logo_24',
  CCMWEB_COMMON_H5_LOGO = 'app_logo_24',
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "duplicateValue",
						Message:   "Duplicate enum member value app_logo_24.",
						Line:      3,
						Column:    3,
						EndLine:   3,
						EndColumn: 40,
					},
				},
			},
			{
				Code: `enum E { A = ((1)), B = 1 }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "duplicateValue", Message: "Duplicate enum member value 1."},
				},
			},
			{
				Code: `enum E { A = +(('1')), B = 1 }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "duplicateValue", Message: "Duplicate enum member value 1."},
				},
			},
			{
				Code: `enum E {
  A = +'0x10',
  B = 16,
  C = +'  ',
  D = 0,
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "duplicateValue", Message: "Duplicate enum member value 16.", Line: 3},
					{MessageId: "duplicateValue", Message: "Duplicate enum member value 0.", Line: 5},
				},
			},
			{
				Code: `enum E {
  A = +'Infinity',
  B = 1e9999,
  C = -'Infinity',
  D = -1e9999,
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "duplicateValue", Message: "Duplicate enum member value Infinity.", Line: 3},
					{MessageId: "duplicateValue", Message: "Duplicate enum member value -Infinity.", Line: 5},
				},
			},
			{
				Code: `enum E {
  A = +'.0000001',
  B = 1e-7,
  C = +'.000001',
  D = 1e-6,
  E = +'1e21',
  F = 1e21,
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "duplicateValue", Message: "Duplicate enum member value 1e-7.", Line: 3},
					{MessageId: "duplicateValue", Message: "Duplicate enum member value 0.000001.", Line: 5},
					{MessageId: "duplicateValue", Message: "Duplicate enum member value 1e+21.", Line: 7},
				},
			},
			{
				Code: "enum E { A = '\\x41', B = 'A', C = `\\u0042`, D = 'B' }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "duplicateValue", Message: "Duplicate enum member value A."},
					{MessageId: "duplicateValue", Message: "Duplicate enum member value B."},
				},
			},
			{
				Code: `enum E {
  A = +'-0',
  B = -0,
  C = -(-0),
  D = 0,
  E = +-0,
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "duplicateValue", Message: "Duplicate enum member value 0.", Line: 3},
					{MessageId: "duplicateValue", Message: "Duplicate enum member value 0.", Line: 5},
					{MessageId: "duplicateValue", Message: "Duplicate enum member value 0.", Line: 6},
				},
			},
			{
				Code: `enum E { A = 1_000, B = 1000 }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "duplicateValue", Message: "Duplicate enum member value 1000."},
				},
			},
			{
				Code: `enum E { A = +'0xffffffffffffffff', B = 18446744073709552000 }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "duplicateValue", Message: "Duplicate enum member value 18446744073709552000."},
				},
			},
			{
				Code: `enum E {
  StringZero = '0',
  PositiveZero = 0,
  NegativeZero = -0,
  One = 1,
  Two = 2,
  Three = 3,
  Four = 4,
  Five = 5,
  Six = 6,
  Seven = 7,
  Eight = 8,
  DuplicateStringZero = '0',
  DuplicatePositiveZero = +0,
  DuplicateNegativeZero = -0,
  NegativeOne = -1,
  DuplicateNegativeOne = -1,
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "duplicateValue", Message: "Duplicate enum member value 0."},
					{MessageId: "duplicateValue", Message: "Duplicate enum member value 0."},
					{MessageId: "duplicateValue", Message: "Duplicate enum member value 0."},
					{MessageId: "duplicateValue", Message: "Duplicate enum member value -1."},
				},
			},
		},
	)
}

func TestGetEnumValueSkipsInvalidTemplateEscape(t *testing.T) {
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/invalid-template.ts",
		Path:     "/invalid-template.ts",
	}, "enum E { A = `\\unicode` }", core.ScriptKindTS)
	member := sourceFile.Statements.Nodes[0].AsEnumDeclaration().Members.Nodes[0].AsEnumMember()
	if _, ok := getEnumValue(member.Initializer); ok {
		t.Fatal("invalid template escape must not have a cooked enum value")
	}
}
