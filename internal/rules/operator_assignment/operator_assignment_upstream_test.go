package operator_assignment

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestOperatorAssignmentUpstream migrates the full valid/invalid suite from
// upstream tests/lib/rules/operator-assignment.js (eslint v10.8.1) 1:1.
// Position assertions cover line/column for every invalid case.
// rslint-specific lock-in cases live in operator_assignment_extras_test.go.
func TestOperatorAssignmentUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&OperatorAssignmentRule,
		[]rule_tester.ValidTestCase{
			{Code: `x = y`},
			{Code: `x = y + x`},
			{Code: `x += x + y`},
			{Code: `x = (x + y) - z`},
			{Code: `x -= y`},
			{Code: `x = y - x`},
			{Code: `x *= x`},
			{Code: `x = y * z`},
			{Code: `x = (x * y) * z`},
			{Code: `x = y / x`},
			{Code: `x /= y`},
			{Code: `x %= y`},
			{Code: `x <<= y`},
			{Code: `x >>= x >> y`},
			{Code: `x >>>= y`},
			{Code: `x &= y`},
			{Code: `x **= y`},
			{Code: `x ^= y ^ z`},
			{Code: `x |= x | y`},
			{Code: `x = x && y`},
			{Code: `x = x || y`},
			{Code: `x = x < y`},
			{Code: `x = x > y`},
			{Code: `x = x <= y`},
			{Code: `x = x >= y`},
			{Code: `x = x instanceof y`},
			{Code: `x = x in y`},
			{Code: `x = x == y`},
			{Code: `x = x != y`},
			{Code: `x = x === y`},
			{Code: `x = x !== y`},
			{Code: `x[y] = x['y'] + z`},
			{Code: `x.y = x['y'] / z`},
			{Code: `x.y = z + x.y`},
			{Code: `x[fn()] = x[fn()] + y`},
			{Code: `x += x + y`, Options: []any{"always"}},
			{Code: `x = x + y`, Options: []any{"never"}},
			{Code: `x = x ** y`, Options: []any{"never"}},
			{Code: `x = y ** x`},
			{Code: `x = x * y + z`},
			{Code: `this.x = this.y + z`, Options: []any{"always"}},
			{Code: `this.x = foo.x + y`, Options: []any{"always"}},
			{Code: `this.x = foo.this.x + y`, Options: []any{"always"}},
			{Code: `const foo = 0; class C { foo = foo + 1; }`},
			// does not check logical operators
			{Code: `x = x && y`, Options: []any{"always"}},
			{Code: `x = x || y`, Options: []any{"always"}},
			{Code: `x = x ?? y`, Options: []any{"always"}},
			{Code: `x &&= y`, Options: []any{"never"}},
			{Code: `x ||= y`, Options: []any{"never"}},
			{Code: `x ??= y`, Options: []any{"never"}},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `x = x + y`,
				Output: []string{`x += y`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaced", Message: "Assignment (=) can be replaced with operator assignment (+=).", Line: 1, Column: 1, EndLine: 1, EndColumn: 10},
				},
			},
			{
				Code:   `x = x - y`,
				Output: []string{`x -= y`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaced", Line: 1, Column: 1},
				},
			},
			{
				Code:   `x = x * y`,
				Output: []string{`x *= y`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaced", Line: 1, Column: 1},
				},
			},
			{
				Code: `x = y * x`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaced", Line: 1, Column: 1},
				},
			},
			{
				Code: `x = (y * z) * x`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaced", Line: 1, Column: 1},
				},
			},
			{
				Code:   `x = x / y`,
				Output: []string{`x /= y`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaced", Line: 1, Column: 1},
				},
			},
			{
				Code:   `x = x % y`,
				Output: []string{`x %= y`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaced", Line: 1, Column: 1},
				},
			},
			{
				Code:   `x = x << y`,
				Output: []string{`x <<= y`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaced", Line: 1, Column: 1},
				},
			},
			{
				Code:   `x = x >> y`,
				Output: []string{`x >>= y`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaced", Line: 1, Column: 1},
				},
			},
			{
				Code:   `x = x >>> y`,
				Output: []string{`x >>>= y`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaced", Line: 1, Column: 1},
				},
			},
			{
				Code:   `x = x & y`,
				Output: []string{`x &= y`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaced", Line: 1, Column: 1},
				},
			},
			{
				Code:   `x = x ^ y`,
				Output: []string{`x ^= y`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaced", Line: 1, Column: 1},
				},
			},
			{
				Code:   `x = x | y`,
				Output: []string{`x |= y`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaced", Line: 1, Column: 1},
				},
			},
			{
				Code:   `x[0] = x[0] - y`,
				Output: []string{`x[0] -= y`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaced", Line: 1, Column: 1},
				},
			},
			{
				Code: `x.y[z['a']][0].b = x.y[z['a']][0].b * 2`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaced", Line: 1, Column: 1},
				},
			},
			{
				Code:    `x = x + y`,
				Output:  []string{`x += y`},
				Options: []any{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaced", Line: 1, Column: 1},
				},
			},
			{
				Code:    `x = (x + y)`,
				Output:  []string{`x += y`},
				Options: []any{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaced", Line: 1, Column: 1},
				},
			},
			{
				Code:    `x = x + (y)`,
				Output:  []string{`x += (y)`},
				Options: []any{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaced", Line: 1, Column: 1},
				},
			},
			{
				Code:    `x += (y)`,
				Output:  []string{`x = x + (y)`},
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code:    `x += y`,
				Output:  []string{`x = x + y`},
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Message: "Unexpected operator assignment (+=) shorthand.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
				},
			},
			{
				Code:   `foo.bar = foo.bar + baz`,
				Output: []string{`foo.bar += baz`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaced", Line: 1, Column: 1},
				},
			},
			{
				Code:    `foo.bar += baz`,
				Output:  []string{`foo.bar = foo.bar + baz`},
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code:   `this.foo = this.foo + bar`,
				Output: []string{`this.foo += bar`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaced", Line: 1, Column: 1},
				},
			},
			{
				Code:    `this.foo += bar`,
				Output:  []string{`this.foo = this.foo + bar`},
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `foo.bar.baz = foo.bar.baz + qux`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaced", Line: 1, Column: 1},
				},
			},
			{
				Code:    `foo.bar.baz += qux`,
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `this.foo.bar = this.foo.bar + baz`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaced", Line: 1, Column: 1},
				},
			},
			{
				Code:    `this.foo.bar += baz`,
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `foo[bar] = foo[bar] + baz`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaced", Line: 1, Column: 1},
				},
			},
			{
				Code: `this[foo] = this[foo] + bar`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaced", Line: 1, Column: 1},
				},
			},
			{
				Code:    `foo[bar] >>>= baz`,
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code:    `this[foo] >>>= bar`,
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code:   `foo[5] = foo[5] / baz`,
				Output: []string{`foo[5] /= baz`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaced", Line: 1, Column: 1},
				},
			},
			{
				Code:   `this[5] = this[5] / foo`,
				Output: []string{`this[5] /= foo`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaced", Line: 1, Column: 1},
				},
			},
			{
				Code:    "/*1*/x/*2*/./*3*/y/*4*/= x.y +/*5*/z/*6*/./*7*/w/*8*/;",
				Output:  []string{"/*1*/x/*2*/./*3*/y/*4*/+=/*5*/z/*6*/./*7*/w/*8*/;"},
				Options: []any{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaced", Line: 1, Column: 6},
				},
			},
			{
				Code:    "x // 1\n . // 2\n y // 3\n = x.y + //4\n z //5\n . //6\n w;",
				Output:  []string{"x // 1\n . // 2\n y // 3\n += //4\n z //5\n . //6\n w;"},
				Options: []any{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaced", Line: 1, Column: 1, EndLine: 7, EndColumn: 3},
				},
			},
			{
				Code:    "x = /*1*/ x + y",
				Options: []any{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaced", Line: 1, Column: 1},
				},
			},
			{
				Code:    "x = //1\n x + y",
				Options: []any{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaced", Line: 1, Column: 1},
				},
			},
			{
				Code:    "x.y = x/*1*/.y + z",
				Options: []any{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaced", Line: 1, Column: 1},
				},
			},
			{
				Code:    "x.y = x. //1\n y + z",
				Options: []any{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaced", Line: 1, Column: 1},
				},
			},
			{
				Code:    "x = x /*1*/ + y",
				Options: []any{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaced", Line: 1, Column: 1},
				},
			},
			{
				Code:    "x = x //1\n + y",
				Options: []any{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaced", Line: 1, Column: 1},
				},
			},
			{
				Code:    "/*1*/x +=/*2*/y/*3*/;",
				Output:  []string{"/*1*/x = x +/*2*/y/*3*/;"},
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 6},
				},
			},
			{
				Code:    "x +=//1\n y",
				Output:  []string{"x = x +//1\n y"},
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1, EndLine: 2, EndColumn: 3},
				},
			},
			{
				Code:    "(/*1*/x += y)",
				Output:  []string{"(/*1*/x = x + y)"},
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 7},
				},
			},
			{
				Code:    "x/*1*/+=  y",
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code:    "x //1\n +=  y",
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code:    "(/*1*/x) +=  y",
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code:    "x/*1*/.y +=  z",
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code:    "x.//1\n y +=  z",
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code:    `(foo.bar) ^= ((((((((((((((((baz))))))))))))))))`,
				Output:  []string{`(foo.bar) = (foo.bar) ^ ((((((((((((((((baz))))))))))))))))`},
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code:   `foo = foo ** bar`,
				Output: []string{`foo **= bar`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaced", Line: 1, Column: 1},
				},
			},
			{
				Code:    `foo **= bar`,
				Output:  []string{`foo = foo ** bar`},
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code:    `foo *= bar + 1`,
				Output:  []string{`foo = foo * (bar + 1)`},
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code:    `foo -= bar - baz`,
				Output:  []string{`foo = foo - (bar - baz)`},
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code:    `foo += bar + baz`,
				Output:  []string{`foo = foo + (bar + baz)`},
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code:    `foo += bar = 1`,
				Output:  []string{`foo = foo + (bar = 1)`},
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code:    `foo *= (bar + 1)`,
				Output:  []string{`foo = foo * (bar + 1)`},
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code:    `foo+=-bar`,
				Output:  []string{`foo= foo+-bar`},
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code:    `foo/=bar`,
				Output:  []string{`foo= foo/bar`},
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code:    `foo/=/**/bar`,
				Output:  []string{`foo= foo/ /**/bar`},
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code:    "foo/=//\nbar",
				Output:  []string{"foo= foo/ //\nbar"},
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1, EndLine: 2, EndColumn: 4},
				},
			},
			{
				Code:    `foo/=/^bar$/`,
				Output:  []string{`foo= foo/ /^bar$/`},
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code:    `foo+=+bar`,
				Output:  []string{`foo= foo+ +bar`},
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code:    `foo+= +bar`,
				Output:  []string{`foo= foo+ +bar`},
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code:    `foo+=/**/+bar`,
				Output:  []string{`foo= foo+/**/+bar`},
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code:    `foo+=+bar===baz`,
				Output:  []string{`foo= foo+(+bar===baz)`},
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			// Optional chaining
			{
				Code: `(obj?.a).b = (obj?.a).b + y`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaced", Line: 1, Column: 1},
				},
			},
			{
				Code: `obj.a = obj?.a + b`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replaced", Line: 1, Column: 1},
				},
			},
		},
	)
}
