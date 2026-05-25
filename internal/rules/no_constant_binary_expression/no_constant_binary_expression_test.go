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
		// ============================================
		// Valid cases
		// ============================================
		[]rule_tester.ValidTestCase{
			// --- Variable references (not constant) ---
			{Code: `bar && foo`},
			{Code: `bar || foo`},
			{Code: `bar ?? foo`},
			{Code: `foo == true`},
			{Code: `foo === true`},
			{Code: `true ? foo : bar`},

			// --- Function return values (not constant) ---
			{Code: `foo() && bar`},
			{Code: `foo() || bar`},
			{Code: `foo() ?? bar`},

			// --- Property access (not constant) ---
			{Code: `foo[0] && bar`},
			{Code: `foo.bar && baz`},

			// --- Template literals with expressions (not constant in non-boolean) ---
			{Code: "var a = `${bar}` && foo"},

			// --- Compound assignment (not constant) ---
			{Code: `(x += 1) && foo`},
			{Code: `(x -= 1) || bar`},

			// --- Delete operations (not constant) ---
			{Code: `delete bar.baz && foo`},

			// --- Nullish coalescing edge cases ---
			{Code: `foo ?? null ?? bar`},

			// --- Shadowed built-in functions ---
			{Code: `function Boolean(n: any) { return n; } Boolean(x) ?? foo`},
			{Code: `function Boolean(n: any) { return n; } Boolean(x) && foo`},
			{Code: `var Boolean = (n: any) => n; Boolean(x) ?? foo`},

			// --- Valid comparisons ---
			{Code: `x === null`},
			{Code: `null === x`},
			{Code: `x == null`},
			{Code: `x == undefined`},
			{Code: `x !== null`},
			{Code: `x != undefined`},

			// --- Logical NOT of non-constant is not constant (#552) ---
			{Code: `!foo && bar`},
			{Code: `!foo || bar`},
			{Code: `!module || !module[pluginName]`},
			{Code: `!!foo && bar`},

			// --- For ==, alwaysNew only applies when BOTH sides are always new ---
			{Code: `x == /[a-z]/`},
			{Code: `x == []`},

			// --- new with user-defined constructors (not guaranteed always new) ---
			{Code: `new Foo() == true`},

			// --- PostfixUnary (not constant) ---
			{Code: `x++ && bar`},
			{Code: `x-- || bar`},

			// --- PrefixUnary ++ / -- (not constant, modifies variable) ---
			{Code: `++x && bar`},

			// --- Boolean(variable) is not constant (result depends on variable) ---
			{Code: `Boolean(foo) && bar`},

			// --- delete is not constant in isConstant ---
			{Code: `delete bar.baz && foo`},

			// --- Unary +/- of variable is not constant ---
			{Code: `+foo && bar`},
			{Code: `-foo || bar`},

			// --- Comma expression with variable last ---
			{Code: `(1, x) && foo`},

			// --- Logical assignment with non-identity rhs (not constant) ---
			{Code: `(x ||= foo) && bar`},
			{Code: `(x &&= foo) || bar`},

			// --- Single-element array has variable loose boolean comparison ---
			{Code: `[x] == true`},

			// --- new user-defined constructor is not always new for === ---
			{Code: `new Foo() === x`},

			// --- delete returns boolean, comparison to boolean varies ---
			{Code: `delete x.y === true`},
		},
		// ============================================
		// Invalid cases
		// ============================================
		[]rule_tester.InvalidTestCase{
			// ============================
			// Constant short-circuit: &&
			// ============================
			// 2-char string is truthy (regression: quote stripping bug)
			{
				Code: `"ab" && foo`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantShortCircuit", Line: 1, Column: 1},
				},
			},
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
			// Negation of constant
			{
				Code: `!true && bar`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantShortCircuit", Line: 1, Column: 1},
				},
			},
			{
				Code: `!undefined && bar`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantShortCircuit", Line: 1, Column: 1},
				},
			},
			{
				Code: `![] && bar`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantShortCircuit", Line: 1, Column: 1},
				},
			},
			// void is always constant
			{
				Code: `void 0 && bar`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantShortCircuit", Line: 1, Column: 1},
				},
			},
			// typeof in boolean position is always constant (non-empty string)
			{
				Code: `typeof foo && bar`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantShortCircuit", Line: 1, Column: 1},
				},
			},
			// Constant binary expressions (e.g. 1+2) are constant
			{
				Code: `(1 + 2) && bar`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantShortCircuit", Line: 1, Column: 1},
				},
			},
			// Assignment with constant right side
			{
				Code: `(x = true) && bar`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantShortCircuit", Line: 1, Column: 1},
				},
			},
			// null is constant
			{
				Code: `null && bar`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantShortCircuit", Line: 1, Column: 1},
				},
			},
			// Class expression is constant
			{
				Code: `(class {}) && bar`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantShortCircuit", Line: 1, Column: 1},
				},
			},

			// ============================
			// Constant short-circuit: ||
			// ============================
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
			{
				Code: `!true || bar`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantShortCircuit", Line: 1, Column: 1},
				},
			},

			// ============================
			// Constant short-circuit: ??
			// ============================
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
			// Comparison operators always produce non-nullish boolean
			{
				Code: `(x > 0) ?? fallback`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantShortCircuit", Line: 1, Column: 1},
				},
			},
			{
				Code: `(x === y) ?? fallback`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantShortCircuit", Line: 1, Column: 1},
				},
			},
			// Unary expressions always have constant nullishness
			{
				Code: `!foo ?? bar`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantShortCircuit", Line: 1, Column: 1},
				},
			},
			{
				Code: `typeof foo ?? bar`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantShortCircuit", Line: 1, Column: 1},
				},
			},
			// String(), Number() always return non-nullish
			{
				Code: `String(x) ?? foo`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantShortCircuit", Line: 1, Column: 1},
				},
			},
			{
				Code: `Number(x) ?? foo`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantShortCircuit", Line: 1, Column: 1},
				},
			},

			// ============================
			// Constant binary operand: ==
			// ============================
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
			// String/numeric comparisons with booleans
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

			// ============================
			// Constant binary operand: !=
			// ============================
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

			// ============================
			// Constant binary operand: ===
			// ============================
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
			{
				Code: `"" === true`,
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

			// ============================
			// Constant binary operand: !==
			// ============================
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

			// ============================
			// Both always new (== only)
			// ============================
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

			// ============================
			// Always new (=== / !==)
			// ============================
			{
				Code: `[] === []`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "alwaysNew", Line: 1, Column: 1},
				},
			},
			{
				Code: `({}) === ({})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "alwaysNew", Line: 1, Column: 1},
				},
			},
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

			// ============================
			// Boolean constructor calls
			// ============================
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

			// ============================
			// Parenthesized expressions
			// ============================
			{
				Code: `(!true) && bar`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantShortCircuit", Line: 1, Column: 1},
				},
			},
			{
				Code: `(null) ?? bar`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantShortCircuit", Line: 1, Column: 1},
				},
			},

			// ============================
			// typeof / void / delete edge cases
			// ============================
			// typeof is always a string, strict comparison with boolean is constant
			{
				Code: `typeof x === true`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantBinaryOperand", Line: 1, Column: 1},
				},
			},
			// void 0 is undefined, comparing with undefined is constant
			{
				Code: `void 0 == undefined`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantBinaryOperand", Line: 1, Column: 1},
				},
			},
			// void produces undefined (always nullish), constant for ??
			{
				Code: `void 0 ?? bar`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantShortCircuit", Line: 1, Column: 1},
				},
			},

			// ============================
			// Logical assignment with identity rhs
			// ============================
			{
				Code: `(x ||= 1) && foo`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantShortCircuit", Line: 1, Column: 1},
				},
			},
			{
				Code: `(x &&= 0) || bar`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantShortCircuit", Line: 1, Column: 1},
				},
			},

			// ============================
			// Comma (sequence) expression
			// ============================
			{
				Code: `(1, 2) && bar`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantShortCircuit", Line: 1, Column: 1},
				},
			},

			// ============================
			// Assignment with constant rhs
			// ============================
			{
				Code: `(x = []) && bar`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantShortCircuit", Line: 1, Column: 1},
				},
			},

			// ============================
			// Conditional expression in isAlwaysNew
			// ============================
			{
				Code: `x === (true ? [] : {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "alwaysNew", Line: 1, Column: 1},
				},
			},

			// ============================
			// Boolean() always non-nullish for ??
			// ============================
			{
				Code: `Boolean(x) ?? foo`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantShortCircuit", Line: 1, Column: 1},
				},
			},

			// ============================
			// Negation of constant in equality
			// ============================
			{
				Code: `!true == 42`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantBinaryOperand", Line: 1, Column: 1},
				},
			},

			// ============================
			// Multi-element array in loose comparison
			// ============================
			{
				Code: `[1, 2] == true`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "constantBinaryOperand", Line: 1, Column: 1},
				},
			},

			// ============================
			// Class expression is always new
			// ============================
			{
				Code: `(class {}) === x`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "alwaysNew", Line: 1, Column: 1},
				},
			},
		},
	)
}

