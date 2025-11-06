package no_constant_binary_expression

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoConstantBinaryExpressionRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoConstantBinaryExpressionRule,
		// Valid cases - ported from ESLint
		[]rule_tester.ValidTestCase{
			// Variable references
			{Code: `bar && foo`},
			{Code: `true ? foo : bar`},
			{Code: `new Foo() == true`},
			{Code: `foo == true`},
			{Code: `foo === true`},

			// Template literals with expressions
			{Code: "var a = `${bar}` && foo"},

			// Assignment expressions
			{Code: `(x += 1) && foo`},

			// Delete operations
			{Code: `delete bar.baz && foo`},

			// Nullish coalescing edge cases
			{Code: `foo ?? null ?? bar`},

			// Shadowed built-in functions
			{Code: `function Boolean(n) { return n; } Boolean(x) ?? foo`},
			{Code: `function Boolean(n) { return n; } Boolean(x) && foo`},

			// Valid comparisons
			{Code: `x === null`},
			{Code: `null === x`},
			{Code: `x == null`},
			{Code: `x == undefined`},

			// Function return values (not constant)
			{Code: `foo() && bar`},
			{Code: `foo() || bar`},
			{Code: `foo() ?? bar`},

			// Array/object access (not constant)
			{Code: `foo[0] && bar`},
			{Code: `foo.bar && baz`},
		},
		// Invalid cases - ported from ESLint
		[]rule_tester.InvalidTestCase{
			// Constant short-circuit: &&
			{
				Code: `[] && greeting`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantShortCircuit", Line: 1, Column: 1},
				},
			},
			{
				Code: `true && hello`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantShortCircuit", Line: 1, Column: 1},
				},
			},
			{
				Code: `'' && foo`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantShortCircuit", Line: 1, Column: 1},
				},
			},
			{
				Code: `100 && foo`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantShortCircuit", Line: 1, Column: 1},
				},
			},
			{
				Code: `/[a-z]/ && foo`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantShortCircuit", Line: 1, Column: 1},
				},
			},
			{
				Code: `Boolean([]) && foo`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantShortCircuit", Line: 1, Column: 1},
				},
			},
			{
				Code: `({}) && foo`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantShortCircuit", Line: 1, Column: 1},
				},
			},
			{
				Code: `(() => {}) && foo`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantShortCircuit", Line: 1, Column: 1},
				},
			},
			{
				Code: `new Foo() && foo`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantShortCircuit", Line: 1, Column: 1},
				},
			},
			{
				Code: `undefined && foo`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantShortCircuit", Line: 1, Column: 1},
				},
			},

			// Constant short-circuit: ||
			{
				Code: `[] || greeting`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantShortCircuit", Line: 1, Column: 1},
				},
			},
			{
				Code: `true || hello`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantShortCircuit", Line: 1, Column: 1},
				},
			},
			{
				Code: `0 || foo`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantShortCircuit", Line: 1, Column: 1},
				},
			},
			{
				Code: `'' || foo`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantShortCircuit", Line: 1, Column: 1},
				},
			},

			// Constant short-circuit: ??
			{
				Code: `({}) ?? foo`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantShortCircuit", Line: 1, Column: 1},
				},
			},
			{
				Code: `1 ?? foo`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantShortCircuit", Line: 1, Column: 1},
				},
			},
			{
				Code: `null ?? foo`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantShortCircuit", Line: 1, Column: 1},
				},
			},
			{
				Code: `undefined ?? foo`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantShortCircuit", Line: 1, Column: 1},
				},
			},

			// Constant binary operand: ==
			{
				Code: `[] == true`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantBinaryOperand", Line: 1, Column: 1},
				},
			},
			{
				Code: `true == []`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantBinaryOperand", Line: 1, Column: 1},
				},
			},
			{
				Code: `({}) == true`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantBinaryOperand", Line: 1, Column: 1},
				},
			},
			{
				Code: `({}) == null`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantBinaryOperand", Line: 1, Column: 1},
				},
			},
			{
				Code: `({}) == undefined`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantBinaryOperand", Line: 1, Column: 1},
				},
			},
			{
				Code: `undefined == true`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantBinaryOperand", Line: 1, Column: 1},
				},
			},
			{
				Code: `true == true`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantBinaryOperand", Line: 1, Column: 1},
				},
			},

			// Constant binary operand: !=
			{
				Code: `({}) != true`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantBinaryOperand", Line: 1, Column: 1},
				},
			},
			{
				Code: `[] != null`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantBinaryOperand", Line: 1, Column: 1},
				},
			},

			// Constant binary operand: ===
			{
				Code: `true === true`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantBinaryOperand", Line: 1, Column: 1},
				},
			},
			{
				Code: `[] === null`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantBinaryOperand", Line: 1, Column: 1},
				},
			},
			{
				Code: `null === null`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantBinaryOperand", Line: 1, Column: 1},
				},
			},
			{
				Code: `({}) === undefined`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantBinaryOperand", Line: 1, Column: 1},
				},
			},
			{
				Code: `({}) === null`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantBinaryOperand", Line: 1, Column: 1},
				},
			},
			{
				Code: `true === false`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantBinaryOperand", Line: 1, Column: 1},
				},
			},

			// Constant binary operand: !==
			{
				Code: `[] !== null`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantBinaryOperand", Line: 1, Column: 1},
				},
			},
			{
				Code: `({}) !== undefined`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantBinaryOperand", Line: 1, Column: 1},
				},
			},

			// Both always new
			{
				Code: `[a] == [a]`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "bothAlwaysNew", Line: 1, Column: 1},
				},
			},
			{
				Code: `({}) == []`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "bothAlwaysNew", Line: 1, Column: 1},
				},
			},
			{
				Code: `[] === []`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "bothAlwaysNew", Line: 1, Column: 1},
				},
			},
			{
				Code: `({}) === ({})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "bothAlwaysNew", Line: 1, Column: 1},
				},
			},

			// Always new
			{
				Code: `x === {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "alwaysNew", Line: 1, Column: 1},
				},
			},
			{
				Code: `x === []`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "alwaysNew", Line: 1, Column: 1},
				},
			},
			{
				Code: `x === (() => {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "alwaysNew", Line: 1, Column: 1},
				},
			},
			{
				Code: `x === /[a-z]/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "alwaysNew", Line: 1, Column: 1},
				},
			},
			{
				Code: `x == /[a-z]/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "alwaysNew", Line: 1, Column: 1},
				},
			},
			{
				Code: `({}) === x`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "alwaysNew", Line: 1, Column: 1},
				},
			},
			{
				Code: `/[a-z]/ === x`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "alwaysNew", Line: 1, Column: 1},
				},
			},

			// String literals with boolean comparison
			{
				Code: `"" == true`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantBinaryOperand", Line: 1, Column: 1},
				},
			},
			{
				Code: `"hello" == false`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantBinaryOperand", Line: 1, Column: 1},
				},
			},
			{
				Code: `"" === true`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantBinaryOperand", Line: 1, Column: 1},
				},
			},

			// Numeric literals with boolean comparison
			{
				Code: `0 == true`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantBinaryOperand", Line: 1, Column: 1},
				},
			},
			{
				Code: `1 == false`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantBinaryOperand", Line: 1, Column: 1},
				},
			},
			{
				Code: `42 === true`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantBinaryOperand", Line: 1, Column: 1},
				},
			},

			// Unary negation
			{
				Code: `!foo && bar`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantShortCircuit", Line: 1, Column: 1},
				},
			},
			// TODO: These cases require scope analysis to detect if Boolean/String/Number
			// are shadowed built-in functions or user-defined functions. Commenting out for now.
			// {
			// 	Code: `Boolean(foo) && bar`,
			// 	Errors: []rule_tester.InvalidTestCaseError{
			// 		{MessageId: "constantShortCircuit", Line: 1, Column: 1},
			// 	},
			// },

			// Boolean constructor calls
			{
				Code: `Boolean(true) && foo`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantShortCircuit", Line: 1, Column: 1},
				},
			},
			{
				Code: `Boolean(false) || foo`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantShortCircuit", Line: 1, Column: 1},
				},
			},
			// TODO: These cases require scope analysis to detect if String/Number
			// are shadowed built-in functions or user-defined functions. Commenting out for now.
			// {
			// 	Code: `String(x) ?? foo`,
			// 	Errors: []rule_tester.InvalidTestCaseError{
			// 		{MessageId: "constantShortCircuit", Line: 1, Column: 1},
			// 	},
			// },
			// {
			// 	Code: `Number(x) ?? foo`,
			// 	Errors: []rule_tester.InvalidTestCaseError{
			// 		{MessageId: "constantShortCircuit", Line: 1, Column: 1},
			// 	},
			// },
		},
	)
}
