package prefer_unary_minus_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_unary_minus"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Additional cases checked against eslint-plugin-unicorn v79.0.0.
func TestPreferUnaryMinusExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_unary_minus.PreferUnaryMinusRule, []rule_tester.ValidTestCase{
		{Code: "value *= -1;", FileName: "case.js"},
		{Code: "value /= -1;", FileName: "case.js"},
		{Code: "(-1) * (-1);", FileName: "case.js"},
		{Code: "(-1) / (-1);", FileName: "case.js"},
		{Code: "value * -(+1);", FileName: "case.js"},
		{Code: "value * -(-1);", FileName: "case.js"},
		{Code: "value * -'1';", FileName: "case.js"},
		{Code: "value / -one;", FileName: "case.js"},
		{Code: "value * (-1 as number);", FileName: "case.ts"},
		{Code: "value / (-1)!;", FileName: "case.ts"},
		{Code: "value * (-1 satisfies number);", FileName: "case.ts"},
		// Numeric boundaries and rule disable directives.
		{Code: "value * -1.0000000000000002;", FileName: "case.js"},
		{Code: "value * -1e309;", FileName: "case.js"},
		{Code: "value / -1e-999;", FileName: "case.js"},
	}, []rule_tester.InvalidTestCase{
		// Numeric literal values and parentheses.
		{Code: "value * -(1);", FileName: "case.js", Output: []string{"-value;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
		}},
		{Code: "(-1) * (value);", FileName: "case.js", Output: []string{"-value;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 15},
		}},
		{Code: "value / -0x1;", FileName: "case.js", Output: []string{"-value;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over dividing by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
		}},
		{Code: "-0b1 * value;", FileName: "case.js", Output: []string{"-value;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
		}},
		{Code: "value * -0o1;", FileName: "case.js", Output: []string{"-value;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
		}},
		{Code: "value * -1e0;", FileName: "case.js", Output: []string{"-value;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
		}},
		{Code: "value * -1.0000000000000001;", FileName: "case.js", Output: []string{"-value;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 28},
		}},
		{Code: "value * -1_0e-1;", FileName: "case.js", Output: []string{"-value;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
		}},
		// Operand precedence and AST shapes.
		{Code: "(a ** b) * -1;", FileName: "case.js", Output: []string{"-(a ** b);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
		}},
		{Code: "(await getValue()) * -1;", FileName: "case.js", Output: []string{"-await getValue();"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
		}},
		{Code: "function* f() { return (yield value) / -1; }", FileName: "case.js", Output: []string{"function* f() { return -(yield value); }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over dividing by `-1`.", Line: 1, Column: 24, EndLine: 1, EndColumn: 42},
		}},
		{Code: "(a && b) * -1;", FileName: "case.js", Output: []string{"-(a && b);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
		}},
		{Code: "(a ?? b) / -1;", FileName: "case.js", Output: []string{"-(a ?? b);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over dividing by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
		}},
		{Code: "typeof value * -1;", FileName: "case.js", Output: []string{"-typeof value;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
		}},
		{Code: "++value * -1;", FileName: "case.js", Output: []string{"-(++value);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
		}},
		{Code: "value-- / -1;", FileName: "case.js", Output: []string{"-(value--);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over dividing by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
		}},
		{Code: "getObject?.()[key] * -1;", FileName: "case.js", Output: []string{"-getObject?.()[key];"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
		}},
		{Code: "class C { #n = 1; f() { return this.#n * -1; } }", FileName: "case.js", Output: []string{"class C { #n = 1; f() { return -this.#n; } }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 32, EndLine: 1, EndColumn: 44},
		}},
		{Code: "const view = <div>{value / -1}</div>;", FileName: "case.tsx", Output: []string{"const view = <div>{-value}</div>;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over dividing by `-1`.", Line: 1, Column: 20, EndLine: 1, EndColumn: 30},
		}},
		{Code: "(<number>value) * -1;", FileName: "case.ts", Output: []string{"-(<number>value);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
		}},
		{Code: "(value satisfies number) * -1;", FileName: "case.ts", Output: []string{"-(value satisfies number);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 30},
		}},
		{Code: "const getValue = () => /** @type {number} */(value) * -1;", FileName: "case.js", Output: []string{"const getValue = () => /** @type {number} */-value;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 45, EndLine: 1, EndColumn: 57},
		}},
		{Code: "/** @type {number} */(value * -1);", FileName: "case.js", Output: []string{"/** @type {number} */(-value);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 23, EndLine: 1, EndColumn: 33},
		}},
		{Code: "(/** @type {number} */ (value)) * -1;", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 37},
		}},
		// Ranges and comment boundaries.
		{Code: "\"😀\"; value * -1;", FileName: "case.js", Output: []string{"\"😀\"; -value;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 17},
		}},
		{Code: "(\n value\n + other\n) / -1;", FileName: "case.js", Output: []string{"-(value\n + other);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over dividing by `-1`.", Line: 1, Column: 1, EndLine: 4, EndColumn: 7},
		}},
		{Code: "value /* keep */ / -1;", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over dividing by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 22},
		}},
		{Code: "value * (/* keep */ -1);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
		}},
		{Code: "-/* keep */1 * value;", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
		}},
		{Code: "(value /* keep */) * -1;", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
		}},
		// Statement boundaries and keyword spacing.
		{Code: "value * (-1)as number;", FileName: "case.ts", Output: []string{"-value as number;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
		}},
		{Code: "(value * -1)satisfies number;", FileName: "case.ts", Output: []string{"(-value) satisfies number;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 2, EndLine: 1, EndColumn: 12},
		}},
		{Code: "foo()\nvalue / -1;", FileName: "case.js", Output: []string{"foo()\n;-value;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over dividing by `-1`.", Line: 2, Column: 1, EndLine: 2, EndColumn: 11},
		}},
		{Code: "if (ok) value * -1;", FileName: "case.js", Output: []string{"if (ok) -value;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 19},
		}},
		{Code: "for (const x of values) x * -1;", FileName: "case.js", Output: []string{"for (const x of values) -x;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 25, EndLine: 1, EndColumn: 31},
		}},
		{Code: "foo;\nvalue * -1;", FileName: "case.js", Output: []string{"foo;\n-value;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 2, Column: 1, EndLine: 2, EndColumn: 11},
		}},
		{Code: "a-/* keep */value * -1;", FileName: "case.js", Output: []string{"a-/* keep */-value;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 13, EndLine: 1, EndColumn: 23},
		}},
		{Code: "function f() {return(value * -1)}", FileName: "case.js", Output: []string{"function f() {return (-value)}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 22, EndLine: 1, EndColumn: 32},
		}},
		// Replacement context and erased wrappers.
		{Code: "(value * -1).toString();", FileName: "case.js", Output: []string{"(-value).toString();"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 2, EndLine: 1, EndColumn: 12},
		}},
		{Code: "(value * -1)();", FileName: "case.js", Output: []string{"(-value)();"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 2, EndLine: 1, EndColumn: 12},
		}},
		{Code: "new (value * -1)();", FileName: "case.js", Output: []string{"new (-value)();"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 16},
		}},
		{Code: "(value * -1)`x`;", FileName: "case.js", Output: []string{"(-value)`x`;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 2, EndLine: 1, EndColumn: 12},
		}},
		{Code: "class C extends (value * -1) {}", FileName: "case.js", Output: []string{"class C extends (-value) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 18, EndLine: 1, EndColumn: 28},
		}},
		{Code: "(value * -1 as any)! ** 2;", FileName: "case.ts", Output: []string{"((-value) as any)! ** 2;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 2, EndLine: 1, EndColumn: 12},
		}},
		{Code: "((value * -1) as any) ** 2;", FileName: "case.ts", Output: []string{"((-value) as any) ** 2;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 3, EndLine: 1, EndColumn: 13},
		}},
		{Code: "(<number>(value * -1)) ** 2;", FileName: "case.ts", Output: []string{"(<number>(-value)) ** 2;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 21},
		}},
		// Optional chains, type wrappers, lexical boundaries, and directives.
		{Code: "obj?.value! * -1;", FileName: "case.ts", Output: []string{"-obj?.value!;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 17},
		}},
		{Code: "obj?.value!! / -1;", FileName: "case.ts", Output: []string{"-obj?.value!!;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over dividing by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
		}},
		{Code: "obj?.method!() * -1;", FileName: "case.ts", Output: []string{"-obj?.method!();"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
		}},
		{Code: "(obj?.value)! * -1;", FileName: "case.ts", Output: []string{"-((obj?.value)!);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
		}},
		{Code: "(obj?.value!)! * -1;", FileName: "case.ts", Output: []string{"-((obj?.value!)!);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
		}},
		{Code: "-1 * (obj?.value)!;", FileName: "case.ts", Output: []string{"-((obj?.value)!);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
		}},
		{Code: "fn<number> * -1;", FileName: "case.ts", Output: []string{"-fn<number>;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
		}},
		{Code: "(fn<number>) * -1;", FileName: "case.ts", Output: []string{"-fn<number>;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
		}},
		{Code: "(value * -1)<number> ** 2;", FileName: "case.ts", Output: []string{"(-value)<number> ** 2;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 2, EndLine: 1, EndColumn: 12},
		}},
		{Code: "(value * -1 as any)!.foo;", FileName: "case.ts", Output: []string{"(-value as any)!.foo;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 2, EndLine: 1, EndColumn: 12},
		}},
		{Code: "(value * -1 satisfies number) ** 2;", FileName: "case.ts", Output: []string{"((-value) satisfies number) ** 2;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 2, EndLine: 1, EndColumn: 12},
		}},
		{Code: "((value * -1) satisfies number) ** 2;", FileName: "case.ts", Output: []string{"((-value) satisfies number) ** 2;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 3, EndLine: 1, EndColumn: 13},
		}},
		{Code: "(<Component />) * -1;", FileName: "case.tsx", Output: []string{"-<Component />;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
		}},
		{Code: "new Value * -1;", FileName: "case.js", Output: []string{"-new Value;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 15},
		}},
		{Code: "a-(-1)*value;", FileName: "case.js", Output: []string{"a- -value;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 3, EndLine: 1, EndColumn: 13},
		}},
		{Code: "function f(){throw(value * -1)}", FileName: "case.js", Output: []string{"function f(){throw (-value)}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 20, EndLine: 1, EndColumn: 30},
		}},
		{Code: "function* f(){yield(value * -1)}", FileName: "case.js", Output: []string{"function* f(){yield (-value)}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 21, EndLine: 1, EndColumn: 31},
		}},
		{Code: "value * -0x0_1;", FileName: "case.js", Output: []string{"-value;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 15},
		}},
		{Code: "value * -0b0_1;", FileName: "case.js", Output: []string{"-value;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 15},
		}},
		{Code: "-0o0_1 * value;", FileName: "case.js", Output: []string{"-value;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 15},
		}},
		{Code: "value * -0.99999999999999999;", FileName: "case.js", Output: []string{"-value;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-unary-minus", Message: "Prefer the unary minus operator over multiplying by `-1`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 29},
		}},
	})
}

func TestPreferUnaryMinusArtifactsFollowDemand(t *testing.T) {
	for _, tc := range []struct {
		source, output string
		diagnostics    int
	}{
		{"value / -1;", "-value;", 1},
		{"getValue(/* keep */) * -1;", "-getValue(/* keep */);", 1},
		{"(value * -1 as number) ** 2;", "((-value) as number) ** 2;", 1},
		{"value /* keep */ * -1;", "value /* keep */ * -1;", 1},
		// RuleTester uses the synthetic rule name "test"; exercise directives
		// through the native linter with the public rule name instead.
		{"// eslint-disable-next-line unicorn/prefer-unary-minus\nvalue * -1;", "// eslint-disable-next-line unicorn/prefer-unary-minus\nvalue * -1;", 0},
		{"value * -1; // eslint-disable-line unicorn/prefer-unary-minus", "value * -1; // eslint-disable-line unicorn/prefer-unary-minus", 0},
		{"/* eslint-disable unicorn/prefer-unary-minus */ value / -1;\n/* eslint-enable unicorn/prefer-unary-minus */ value * -1;", "/* eslint-disable unicorn/prefer-unary-minus */ value / -1;\n/* eslint-enable unicorn/prefer-unary-minus */ -value;", 1},
	} {
		t.Run(tc.source, func(t *testing.T) {
			helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
			program, sourceFile, err := helper.CreateTestProgram(tc.source, "edit-demand.ts", "tsconfig.json")
			if err != nil {
				t.Fatal(err)
			}
			diagnostics := make(map[rule.EditDemand]rule.RuleDiagnostic, 4)
			for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
				var got []rule.RuleDiagnostic
				linter.LintSingleFile(linter.LintSingleFileOptions{
					Program: lintprogram.NewFromCompiler(program), File: sourceFile.FileName(),
					GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
						return []rule.ConfiguredRule{{
							Name: prefer_unary_minus.PreferUnaryMinusRule.Name, Severity: rule.SeverityError,
							Run: func(ctx rule.RuleContext) rule.RuleListeners {
								return prefer_unary_minus.PreferUnaryMinusRule.Run(ctx, nil)
							},
						}}
					},
					Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) { got = append(got, d) }},
				})
				if len(got) != tc.diagnostics {
					t.Fatalf("demand %d: got %d diagnostics, want %d", demand, len(got), tc.diagnostics)
				}
				if tc.diagnostics != 0 {
					diagnostics[demand] = got[0]
				}
			}
			if tc.diagnostics == 0 {
				return
			}
			baseline, all := diagnostics[rule.EditDemandNone], diagnostics[rule.EditDemandAll]
			for demand, diagnostic := range diagnostics {
				if diagnostic.Range != baseline.Range || !reflect.DeepEqual(diagnostic.Message, baseline.Message) || diagnostic.Severity != baseline.Severity || diagnostic.RuleName != baseline.RuleName {
					t.Errorf("demand %d changed diagnostic identity", demand)
				}
				wantFix := tc.output != tc.source && (demand == rule.EditDemandAutofix || demand == rule.EditDemandAll)
				if (diagnostic.FixesPtr != nil) != wantFix {
					t.Errorf("demand %d: unexpected autofix artifacts", demand)
				}
				if wantFix && !reflect.DeepEqual(diagnostic.FixesPtr, all.FixesPtr) {
					t.Errorf("demand %d changed autofix artifacts", demand)
				}
				if diagnostic.Suggestions != nil {
					t.Errorf("demand %d produced suggestions", demand)
				}
				wantOutput := tc.source
				if wantFix {
					wantOutput = tc.output
				}
				output, _, fixed := linter.ApplyRuleFixes(tc.source, []rule.RuleDiagnostic{diagnostic})
				if output != wantOutput || fixed != wantFix {
					t.Errorf("demand %d: unexpected autofix %q", demand, output)
				}
			}
		})
	}
}
