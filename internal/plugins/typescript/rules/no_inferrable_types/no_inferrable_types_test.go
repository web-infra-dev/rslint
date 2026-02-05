package no_inferrable_types

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoInferrableTypesRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoInferrableTypesRule, []rule_tester.ValidTestCase{
		// No type annotation - valid
		{Code: `const a = 10;`},
		{Code: `const a = true;`},
		{Code: `const a = 'str';`},
		{Code: `const a = null;`},
		{Code: `const a = undefined;`},
		{Code: `const a = /a/;`},
		{Code: `const a = 10n;`},
		{Code: `const a = Symbol('a');`},

		// Type annotation with different type - valid
		{Code: `const a: unknown = 10;`},
		{Code: `const a: any = true;`},

		// Function parameters with ignoreParameters: true
		{
			Code:    `function fn(a: number = 5) {}`,
			Options: []interface{}{map[string]interface{}{"ignoreParameters": true}},
		},

		// Class properties with ignoreProperties: true
		{
			Code:    `class Foo { prop: number = 5; }`,
			Options: []interface{}{map[string]interface{}{"ignoreProperties": true}},
		},

		// Readonly class properties should be ignored even without option
		{Code: `class Foo { readonly prop: number = 5; }`},

		// Optional properties with initializers are not flagged by this rule.
		{
			Code: `class Foo {
  a?: number = 5;
}`,
		},
	}, []rule_tester.InvalidTestCase{
		// bigint
		{
			Code:   `const a: bigint = 10n;`,
			Output: []string{`const a = 10n;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    7,
				},
			},
		},
		{
			Code:   `const a: bigint = -10n;`,
			Output: []string{`const a = -10n;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    7,
				},
			},
		},
		{
			Code:   `const a: bigint = BigInt(10);`,
			Output: []string{`const a = BigInt(10);`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    7,
				},
			},
		},

		// boolean
		{
			Code:   `const a: boolean = true;`,
			Output: []string{`const a = true;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    7,
				},
			},
		},
		{
			Code:   `const a: boolean = false;`,
			Output: []string{`const a = false;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    7,
				},
			},
		},
		{
			Code:   `const a: boolean = Boolean(null);`,
			Output: []string{`const a = Boolean(null);`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    7,
				},
			},
		},
		{
			Code:   `const a: boolean = !0;`,
			Output: []string{`const a = !0;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    7,
				},
			},
		},

		// number
		{
			Code:   `const a: number = 10;`,
			Output: []string{`const a = 10;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    7,
				},
			},
		},
		{
			Code:   `const a: number = +10;`,
			Output: []string{`const a = +10;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    7,
				},
			},
		},
		{
			Code:   `const a: number = -10;`,
			Output: []string{`const a = -10;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    7,
				},
			},
		},
		{
			Code:   `const a: number = Number('1');`,
			Output: []string{`const a = Number('1');`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    7,
				},
			},
		},
		{
			Code:   `const a: number = Infinity;`,
			Output: []string{`const a = Infinity;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    7,
				},
			},
		},
		{
			Code:   `const a: number = NaN;`,
			Output: []string{`const a = NaN;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    7,
				},
			},
		},

		// null
		{
			Code:   `const a: null = null;`,
			Output: []string{`const a = null;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    7,
				},
			},
		},

		// RegExp
		{
			Code:   `const a: RegExp = /a/;`,
			Output: []string{`const a = /a/;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    7,
				},
			},
		},
		{
			Code:   `const a: RegExp = RegExp('a');`,
			Output: []string{`const a = RegExp('a');`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    7,
				},
			},
		},
		{
			Code:   `const a: RegExp = new RegExp('a');`,
			Output: []string{`const a = new RegExp('a');`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    7,
				},
			},
		},

		// string
		{
			Code:   `const a: string = 'str';`,
			Output: []string{`const a = 'str';`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    7,
				},
			},
		},
		{
			Code:   "const a: string = `str`;",
			Output: []string{"const a = `str`;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    7,
				},
			},
		},
		{
			Code:   `const a: string = String(1);`,
			Output: []string{`const a = String(1);`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    7,
				},
			},
		},

		// symbol
		{
			Code:   `const a: symbol = Symbol('a');`,
			Output: []string{`const a = Symbol('a');`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    7,
				},
			},
		},

		// undefined
		{
			Code:   `const a: undefined = undefined;`,
			Output: []string{`const a = undefined;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    7,
				},
			},
		},
		{
			Code:   `const a: undefined = void 0;`,
			Output: []string{`const a = void 0;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    7,
				},
			},
		},

		// Function parameters (without ignoreParameters option)
		{
			Code:   `function fn(a: number = 5) {}`,
			Output: []string{`function fn(a = 5) {}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    13,
				},
			},
		},
		{
			Code:   `const fn = (a: boolean = true) => {};`,
			Output: []string{`const fn = (a = true) => {};`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    13,
				},
			},
		},

		// Class properties (without ignoreProperties option)
		{
			Code:   `class Foo { prop: number = 5; }`,
			Output: []string{`class Foo { prop = 5; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      1,
					Column:    13,
				},
			},
		},

		// Class properties with definite assignment assertion (!)
		{
			Code: `class A {
  a!: number = 1;
}`,
			Output: []string{`class A {
  a = 1;
}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      2,
					Column:    3,
				},
			},
		},

		// Auto-accessor properties
		{
			Code: `class Foo {
  accessor a: number = 5;
}`,
			Output: []string{`class Foo {
  accessor a = 5;
}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noInferrableType",
					Line:      2,
					Column:    3,
				},
			},
		},
	})
}
