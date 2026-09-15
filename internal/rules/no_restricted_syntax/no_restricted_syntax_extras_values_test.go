package no_restricted_syntax

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Regression expectations were checked against ESLint 10.10.0, esquery 1.7.0,
// and @typescript-eslint/parser 8.69.0. These cases exercise the tsgo facade.
func TestNoRestrictedSyntaxValuesRegressions(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoRestrictedSyntaxRule,
		[]rule_tester.ValidTestCase{
			{
				// bigint regex
				Code:    "const n = 1n;",
				Options: []any{"Literal[value=/^1$/]"},
			},
			{
				// bigint
				Code:    "const x = [1n,9007199254740993n,0xFFn, 10000000000000000000000000000000000000000n];",
				Options: []any{"Literal[value=type(string)]"},
			},
			{
				// bigint
				Code:    "const x = [1n,9007199254740993n,0xFFn, 10000000000000000000000000000000000000000n];",
				Options: []any{"Literal[value=/^1$/]"},
			},
			{
				// bigint
				Code:    "const x = [1n,9007199254740993n,0xFFn, 10000000000000000000000000000000000000000n];",
				Options: []any{"Literal[value>\"1.5\"]"},
			},
			{
				// bigint
				Code:    "const x = [1n,9007199254740993n,0xFFn, 10000000000000000000000000000000000000000n];",
				Options: []any{"Literal[value<\"Infinity\"]"},
			},
			{
				// private name
				Code:    "class C { #x = 1; m(){ return this.#x; } }",
				Options: []any{"PrivateIdentifier[name=\"#x\"]"},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				// bigint type
				Code:    "const n = 1n;",
				Options: []any{"Literal[value=type(bigint)]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 11, EndLine: 1, EndColumn: 13},
				},
			},
			{
				// number coercion
				Code:    "const x = [1e21, 1e-7, 1e-6, 1e20, 1e999];",
				Options: []any{"Literal[value=\"1e+21\"]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 12, EndLine: 1, EndColumn: 16},
				},
			},
			{
				// number coercion
				Code:    "const x = [1e21, 1e-7, 1e-6, 1e20, 1e999];",
				Options: []any{"Literal[value=\"1e-7\"]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 18, EndLine: 1, EndColumn: 22},
				},
			},
			{
				// number coercion
				Code:    "const x = [1e21, 1e-7, 1e-6, 1e20, 1e999];",
				Options: []any{"Literal[value=\"0.000001\"]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 24, EndLine: 1, EndColumn: 28},
				},
			},
			{
				// number coercion
				Code:    "const x = [1e21, 1e-7, 1e-6, 1e20, 1e999];",
				Options: []any{"Literal[value=\"100000000000000000000\"]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 30, EndLine: 1, EndColumn: 34},
				},
			},
			{
				// number coercion
				Code:    "const x = [1e21, 1e-7, 1e-6, 1e20, 1e999];",
				Options: []any{"Literal[value=\"Infinity\"]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 36, EndLine: 1, EndColumn: 41},
				},
			},
			{
				// string coercion
				Code:    "const x = [\"0x100000000000000000000\", \"NaN\", \"\u0085\", \"😀\", \"\\uD800\", \"Infinity\", \"inf\"];",
				Options: []any{"Literal[value>0]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 12, EndLine: 1, EndColumn: 37},
					{MessageId: "restrictedSyntax", Line: 1, Column: 67, EndLine: 1, EndColumn: 77},
				},
			},
			{
				// string coercion
				Code:    "const x = [\"0x100000000000000000000\", \"NaN\", \"\u0085\", \"😀\", \"\\uD800\", \"Infinity\", \"inf\"];",
				Options: []any{"Literal[value.length=2]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 51, EndLine: 1, EndColumn: 55},
				},
			},
			{
				// string coercion
				Code:    "const x = [\"0x100000000000000000000\", \"NaN\", \"\u0085\", \"😀\", \"\\uD800\", \"Infinity\", \"inf\"];",
				Options: []any{"Literal[value.length=1]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 46, EndLine: 1, EndColumn: 49},
					{MessageId: "restrictedSyntax", Line: 1, Column: 57, EndLine: 1, EndColumn: 65},
				},
			},
			{
				// string coercion
				Code:    "const x = [\"0x100000000000000000000\", \"NaN\", \"\u0085\", \"😀\", \"\\uD800\", \"Infinity\", \"inf\"];",
				Options: []any{"Literal[value<\"�\"]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 12, EndLine: 1, EndColumn: 37},
					{MessageId: "restrictedSyntax", Line: 1, Column: 39, EndLine: 1, EndColumn: 44},
					{MessageId: "restrictedSyntax", Line: 1, Column: 46, EndLine: 1, EndColumn: 49},
					{MessageId: "restrictedSyntax", Line: 1, Column: 51, EndLine: 1, EndColumn: 55},
					{MessageId: "restrictedSyntax", Line: 1, Column: 57, EndLine: 1, EndColumn: 65},
					{MessageId: "restrictedSyntax", Line: 1, Column: 67, EndLine: 1, EndColumn: 77},
					{MessageId: "restrictedSyntax", Line: 1, Column: 79, EndLine: 1, EndColumn: 84},
				},
			},
			{
				// bigint
				Code:    "const x = [1n,9007199254740993n,0xFFn, 10000000000000000000000000000000000000000n];",
				Options: []any{"Literal[value=type(bigint)]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 12, EndLine: 1, EndColumn: 14},
					{MessageId: "restrictedSyntax", Line: 1, Column: 15, EndLine: 1, EndColumn: 32},
					{MessageId: "restrictedSyntax", Line: 1, Column: 33, EndLine: 1, EndColumn: 38},
					{MessageId: "restrictedSyntax", Line: 1, Column: 40, EndLine: 1, EndColumn: 82},
				},
			},
			{
				// bigint
				Code:    "const x = [1n,9007199254740993n,0xFFn, 10000000000000000000000000000000000000000n];",
				Options: []any{"Literal[value!=/^1$/]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 15, EndLine: 1, EndColumn: 32},
					{MessageId: "restrictedSyntax", Line: 1, Column: 33, EndLine: 1, EndColumn: 38},
					{MessageId: "restrictedSyntax", Line: 1, Column: 40, EndLine: 1, EndColumn: 82},
				},
			},
			{
				// bigint
				Code:    "const x = [1n,9007199254740993n,0xFFn, 10000000000000000000000000000000000000000n];",
				Options: []any{"Literal[value>9007199254740992]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 15, EndLine: 1, EndColumn: 32},
					{MessageId: "restrictedSyntax", Line: 1, Column: 40, EndLine: 1, EndColumn: 82},
				},
			},
			{
				// bigint
				Code:    "const x = [1n,9007199254740993n,0xFFn, 10000000000000000000000000000000000000000n];",
				Options: []any{"Literal[value<\"9007199254740994\"]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 12, EndLine: 1, EndColumn: 14},
					{MessageId: "restrictedSyntax", Line: 1, Column: 15, EndLine: 1, EndColumn: 32},
					{MessageId: "restrictedSyntax", Line: 1, Column: 33, EndLine: 1, EndColumn: 38},
				},
			},
			{
				// bigint
				Code:    "const x = [1n,9007199254740993n,0xFFn, 10000000000000000000000000000000000000000n];",
				Options: []any{"Literal[value<9999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 12, EndLine: 1, EndColumn: 14},
					{MessageId: "restrictedSyntax", Line: 1, Column: 15, EndLine: 1, EndColumn: 32},
					{MessageId: "restrictedSyntax", Line: 1, Column: 33, EndLine: 1, EndColumn: 38},
					{MessageId: "restrictedSyntax", Line: 1, Column: 40, EndLine: 1, EndColumn: 82},
				},
			},
			{
				// bigint
				Code:    "const x = [1n,9007199254740993n,0xFFn, 10000000000000000000000000000000000000000n];",
				Options: []any{"Literal[value>\"0xFE\"]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 15, EndLine: 1, EndColumn: 32},
					{MessageId: "restrictedSyntax", Line: 1, Column: 33, EndLine: 1, EndColumn: 38},
					{MessageId: "restrictedSyntax", Line: 1, Column: 40, EndLine: 1, EndColumn: 82},
				},
			},
			{
				// private name
				Code:    "class C { #x = 1; m(){ return this.#x; } }",
				Options: []any{"PrivateIdentifier[name=\"x\"]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 11, EndLine: 1, EndColumn: 13},
					{MessageId: "restrictedSyntax", Line: 1, Column: 36, EndLine: 1, EndColumn: 38},
				},
			},
			{
				// private name
				Code:    "class C { #x = 1; m(){ return this.#x; } }",
				Options: []any{"PrivateIdentifier[name=/^x$/]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 11, EndLine: 1, EndColumn: 13},
					{MessageId: "restrictedSyntax", Line: 1, Column: 36, EndLine: 1, EndColumn: 38},
				},
			},
		})
}
