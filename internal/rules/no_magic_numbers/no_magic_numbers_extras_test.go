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

		// ---- Computed keys of object binding patterns (ESTree makes them the
		// key of a Property, like an object literal's) ----
		{Code: `const { [42]: value } = source;`},
		{Code: `const { [-42]: value } = source;`},
		{Code: `function f({ [42]: value }) {}`},
		{Code: `for (const { [42]: value } of sources) {}`},
		{Code: `const { a: { [42]: value } } = source;`},
		{Code: `const { [42]: [value] } = source;`},
		{Code: `let { [42]: value } = source;`, Options: map[string]interface{}{"enforceConst": true}},
		{Code: `const { 1: a } = source;`},
		{Code: `function f({ 1: a }) {}`},
		{Code: `const { 1: a, [2]: b } = source;`},
		{Code: `var one; ({one = 1} = {})`, Options: map[string]interface{}{"ignoreDefaultValues": true}},
		{Code: `var a, b; ({a = 1, b = 2} = {})`, Options: map[string]interface{}{"ignoreDefaultValues": true}},
		{Code: `var x; ({a: x = 42} = {})`, Options: map[string]interface{}{"ignoreDefaultValues": true}},

		// ---- Destructuring targets only (the iterable side still reports) ----
		{Code: `for ([a = 1] of foo) {}`, Options: map[string]interface{}{"ignoreDefaultValues": true}},
		{Code: `for ([a = 1] in foo) {}`, Options: map[string]interface{}{"ignoreDefaultValues": true}},
		{Code: `[a = 1] = arr;`, Options: map[string]interface{}{"ignoreDefaultValues": true}},
		{Code: `({ p: a = 1 } = o);`, Options: map[string]interface{}{"ignoreDefaultValues": true}},

		// ---- Computed keys of readonly class properties (ESTree makes them
		// direct children of the PropertyDefinition) ----
		{Code: `class C { readonly [1] = foo; }`, Options: map[string]interface{}{"ignoreReadonlyClassProperties": true}},
		{Code: `class C { readonly [-1] = foo; }`, Options: map[string]interface{}{"ignoreReadonlyClassProperties": true}},

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
			Code:    `for (const x of [a = 1]) {}`,
			Options: map[string]interface{}{"ignoreDefaultValues": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 1.", Line: 1, Column: 22, EndColumn: 23}},
		},
		{
			Code:    `for (const x in [a = 1]) {}`,
			Options: map[string]interface{}{"ignoreDefaultValues": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 1.", Line: 1, Column: 22, EndColumn: 23}},
		},
		{
			Code:    `[b] = [a = 1];`,
			Options: map[string]interface{}{"ignoreDefaultValues": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 1.", Line: 1, Column: 12, EndColumn: 13}},
		},
		{
			Code:    `x = [a = 1];`,
			Options: map[string]interface{}{"ignoreDefaultValues": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 1.", Line: 1, Column: 10, EndColumn: 11}},
		},
		{
			Code:    `class C { accessor x = 1; }`,
			Options: map[string]interface{}{"ignoreClassFieldInitialValues": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 1.", Line: 1, Column: 24, EndColumn: 25}},
		},
		{
			Code:    `class C { static accessor x = 1; }`,
			Options: map[string]interface{}{"ignoreClassFieldInitialValues": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 1.", Line: 1, Column: 31, EndColumn: 32}},
		},
		{
			Code:    `class C { [1] = foo; }`,
			Options: map[string]interface{}{"ignoreReadonlyClassProperties": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 1.", Line: 1, Column: 12, EndColumn: 13}},
		},
		{
			Code:   `class C { readonly [1] = foo; }`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 1.", Line: 1, Column: 21, EndColumn: 22}},
		},
		{
			Code:   `const { [42]: value = 7 } = source;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 7.", Line: 1, Column: 23, EndColumn: 24}},
		},
		{
			Code:    `const { [42]: value } = source;`,
			Options: map[string]interface{}{"detectObjects": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 42.", Line: 1, Column: 10, EndColumn: 12}},
		},
		{
			Code:    `const { 1: a } = source;`,
			Options: map[string]interface{}{"detectObjects": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 1.", Line: 1, Column: 9, EndColumn: 10}},
		},
		{
			Code:   `const { 1: a = 2 } = source;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 2.", Line: 1, Column: 16, EndColumn: 17}},
		},
		{
			Code:    `class C { readonly accessor x = 1; }`,
			Options: map[string]interface{}{"ignoreReadonlyClassProperties": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noMagic", Message: "No magic number: 1.", Line: 1, Column: 33, EndColumn: 34}},
		},
		{
			Code:    `class C { readonly accessor [1] = 2; }`,
			Options: map[string]interface{}{"ignoreReadonlyClassProperties": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noMagic", Message: "No magic number: 1.", Line: 1, Column: 30, EndColumn: 31},
				{MessageId: "noMagic", Message: "No magic number: 2.", Line: 1, Column: 35, EndColumn: 36},
			},
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
