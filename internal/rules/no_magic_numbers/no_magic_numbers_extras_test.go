// Edge-shape augmentation and rslint-specific lock-ins for no-magic-numbers,
// beyond what tests/lib/rules/no-magic-numbers.js covers. See
// no_magic_numbers_upstream_test.go for the migrated upstream suite.
package no_magic_numbers

import (
	"math"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoMagicNumbersExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoMagicNumbersRule, []rule_tester.ValidTestCase{
		// ---- ParenthesizedExpression (tsgo-specific, ESTree has no paren nodes) ----
		{Code: `const X = (42);`},
		{Code: `const X = ((42));`},
		{Code: `var x = { foo: (42) };`},
		{Code: `obj.prop = (1);`},
		{Code: `parseInt(y, (10));`},
		{Code: `foo[(0)]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},
		{Code: `enum E { A = (42) }`, Options: map[string]interface{}{"ignoreEnums": true}},
		{Code: `class C { readonly x = (42); }`, Options: map[string]interface{}{"ignoreReadonlyClassProperties": true}},
		{Code: `class C { foo = (2); }`, Options: map[string]interface{}{"ignoreClassFieldInitialValues": true}},
		{Code: `const func = (param = (123)) => {}`, Options: map[string]interface{}{"ignoreDefaultValues": true}},
		{Code: `var a = <input maxLength={(10)} />;`, FileName: "src/virtual.tsx"},
		{Code: `const X = -(42);`, Options: map[string]interface{}{"ignore": []interface{}{float64(-42)}}},

		// ---- Edge cases: tsgo AST shapes ----
		{Code: `type Foo = Bar[((((1))))];`, Options: map[string]interface{}{"ignoreTypeIndexes": true}},
		{Code: `enum E { A = 1 << 0, B = 1 << 1 }`, Options: map[string]interface{}{"ignoreEnums": true, "ignore": []interface{}{float64(0), float64(1)}}},
		{Code: `class C { protected readonly x = 42; }`, Options: map[string]interface{}{"ignoreReadonlyClassProperties": true}},
		{Code: `foo?.bar?.[0]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},
		{Code: `var x = { foo: 42 };`},
		{Code: `var colors = {}; colors.RED = 2;`},
		{Code: `const { a: { b = 1 } } = obj;`, Options: map[string]interface{}{"ignoreDefaultValues": true}},
		{Code: `class C { foo = -42; }`, Options: map[string]interface{}{"ignoreClassFieldInitialValues": true}},
		{Code: `class C { readonly x = 100n; }`, Options: map[string]interface{}{"ignoreReadonlyClassProperties": true}},
		{Code: `var x = {[42]: true}`},
		{Code: `var one; ({one = 1} = {})`, Options: map[string]interface{}{"ignoreDefaultValues": true}},
		{Code: `var a, b; ({a = 1, b = 2} = {})`, Options: map[string]interface{}{"ignoreDefaultValues": true}},
		{Code: `var x; ({a: x = 42} = {})`, Options: map[string]interface{}{"ignoreDefaultValues": true}},

		// ---- Signed zero (JS Set lookups use SameValueZero) ----
		{Code: `f(-0)`, Options: map[string]interface{}{"ignore": []interface{}{float64(0)}}},
		{Code: `f(0)`, Options: map[string]interface{}{"ignore": []interface{}{math.Copysign(0, -1)}}},
		{Code: `f(-0)`, Options: map[string]interface{}{"ignore": []interface{}{math.Copysign(0, -1)}}},
		{Code: `f(-0n)`, Options: map[string]interface{}{"ignore": []interface{}{"0n"}}},

		// ---- Object-literal methods and accessors (ESTree Property) ----
		{Code: `({ 42() {} })`},
		{Code: `({ [42]() {} })`},
		{Code: `({ get 42() {} })`},
		{Code: `({ set 42(v) {} })`},
		{Code: `({ async 42() {} })`},
		{Code: `({ *42() {} })`},

		// ---- Upstream semantic lock-in ----
		{Code: `var stats = {avg: 42};`},
		{Code: `({key: 90, another: 10})`},
		{Code: `colors.RED = 2;`},
		{Code: `const DAY = 86400;`},
		{Code: `var HOUR = 3600;`},
		{Code: `foo[0xABn]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},
		{Code: `foo[5.0000000000000001]`, Options: map[string]interface{}{"ignoreArrayIndexes": true}},
		{Code: "// eslint-disable-next-line test\nf(42);"},
		{Code: "/* eslint-disable test */\nf(42);\n/* eslint-enable test */"},

		// ---- Legacy octal literals are rejected by rslint's parser
		// regardless of module-ness, so ignoreArrayIndexes never needs to
		// special-case them the way ESLint's `sourceType: "script"` cases
		// (`foo[0123]`, `foo[-012]`) do upstream. Modern octal (`0o71`) is
		// covered by the upstream suite.
	}, []rule_tester.InvalidTestCase{
		{
			Code:   `function f() { return -(1); }`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic"}},
		},
		// The reported text is the operator plus the literal's own raw
		// text; the reported range still covers the parentheses.
		{
			Code:   `f(-(1));`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -1.", Line: 1, Column: 3, EndColumn: 7}},
		},
		{
			Code:   `f(-((1)));`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -1.", Line: 1, Column: 3, EndColumn: 9}},
		},
		{
			Code:   `f(+(1));`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: +1.", Line: 1, Column: 3, EndColumn: 7}},
		},
		{
			Code:   `f(- 1);`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -1.", Line: 1, Column: 3, EndColumn: 6}},
		},
		{
			Code:   `f(-(0x1F));`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -0x1F.", Line: 1, Column: 3, EndColumn: 10}},
		},
		{
			Code:   `f(-(1_000));`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: -1_000.", Line: 1, Column: 3, EndColumn: 11}},
		},
		{
			Code:   `a = (1);`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 1."}},
		},
		{
			Code:   `min = 1;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 1."}},
		},
		{
			Code:   `function f() { return 60; }`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 60."}},
		},
		{
			Code:   `class A { 42() {} }`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 42.", Line: 1, Column: 11}},
		},
		{
			Code:   `class A { [42]() {} }`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 42.", Line: 1, Column: 12}},
		},
		{
			Code:   `class A { get 42() {} }`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 42.", Line: 1, Column: 15}},
		},
		{
			Code:    `({ 42() {} })`,
			Options: map[string]interface{}{"detectObjects": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 42.", Line: 1, Column: 4}},
		},
		{
			Code: `f(/* leading trivia */ 42, /* unary */ -(7), /* bigint */ 9n);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noMagic", Message: "No magic number: 42.", Line: 1, Column: 24},
				{MessageId: "noMagic", Message: "No magic number: -7.", Line: 1, Column: 40},
				{MessageId: "noMagic", Message: "No magic number: 9n.", Line: 1, Column: 59},
			},
		},
	})
}
