package consistent_type_assertions

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestConsistentTypeAssertionsRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ConsistentTypeAssertionsRule, []rule_tester.ValidTestCase{
		// Default options (assertionStyle: 'as')
		{Code: `const x = b as A;`},
		{Code: `const x = [1] as readonly number[];`},
		{Code: `const x = 'string' as a | b;`},
		{Code: `const x = () => ({ bar: 5 }) as Foo;`},
		{Code: `const x = { key: 'value' } as const;`},
		{Code: `const x = <const>{ key: 'value' };`},
		{Code: `const x = [1] as const;`},
		{Code: `const x = <const>[1];`},
		{Code: `const x = { foo: 1 } as Foo;`},
		{Code: `const x = [] as Foo[];`},
		{Code: `const x = new Generic<int>() as Foo;`},
		{Code: `const x = (foo as Bar).baz;`},
		{Code: `const x = [1, 2, 3] as number[];`},

		// assertionStyle: 'angle-bracket'
		{Code: `const x = <A>b;`, Options: []interface{}{map[string]interface{}{"assertionStyle": "angle-bracket"}}},
		{Code: `const x = <readonly number[]>[1];`, Options: []interface{}{map[string]interface{}{"assertionStyle": "angle-bracket"}}},
		{Code: `const x = <a | b>'string';`, Options: []interface{}{map[string]interface{}{"assertionStyle": "angle-bracket"}}},
		{Code: `const x = <Foo>(() => ({ bar: 5 }));`, Options: []interface{}{map[string]interface{}{"assertionStyle": "angle-bracket"}}},
		{Code: `const x = <const>{ key: 'value' };`, Options: []interface{}{map[string]interface{}{"assertionStyle": "angle-bracket"}}},
		{Code: `const x = { key: 'value' } as const;`, Options: []interface{}{map[string]interface{}{"assertionStyle": "angle-bracket"}}},
		{Code: `const x = <const>[1];`, Options: []interface{}{map[string]interface{}{"assertionStyle": "angle-bracket"}}},
		{Code: `const x = [1] as const;`, Options: []interface{}{map[string]interface{}{"assertionStyle": "angle-bracket"}}},
		{Code: `const x = <Foo>{ foo: 1 };`, Options: []interface{}{map[string]interface{}{"assertionStyle": "angle-bracket"}}},
		{Code: `const x = <Foo[]>[];`, Options: []interface{}{map[string]interface{}{"assertionStyle": "angle-bracket"}}},

		// assertionStyle: 'never'
		{Code: `const x = { key: 'value' } as const;`, Options: []interface{}{map[string]interface{}{"assertionStyle": "never"}}},
		{Code: `const x = <const>{ key: 'value' };`, Options: []interface{}{map[string]interface{}{"assertionStyle": "never"}}},
		{Code: `const x = [1] as const;`, Options: []interface{}{map[string]interface{}{"assertionStyle": "never"}}},
		{Code: `const x = <const>[1];`, Options: []interface{}{map[string]interface{}{"assertionStyle": "never"}}},

		// objectLiteralTypeAssertions: 'never'
		{Code: `const x: Foo = { bar: 5 };`, Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "objectLiteralTypeAssertions": "never"}}},
		{Code: `const x = { bar: 5 } as any;`, Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "objectLiteralTypeAssertions": "never"}}},
		{Code: `const x = { bar: 5 } as unknown;`, Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "objectLiteralTypeAssertions": "never"}}},
		{Code: `const x = { bar: 5 } as const;`, Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "objectLiteralTypeAssertions": "never"}}},
		{Code: `const x = 'string' as Foo;`, Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "objectLiteralTypeAssertions": "never"}}},
		{Code: `const x = 123 as Foo;`, Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "objectLiteralTypeAssertions": "never"}}},
		{Code: `const x = true as Foo;`, Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "objectLiteralTypeAssertions": "never"}}},

		// objectLiteralTypeAssertions: 'allow-as-parameter'
		{Code: `const x: Foo = { bar: 5 };`, Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "objectLiteralTypeAssertions": "allow-as-parameter"}}},
		{Code: `foo({ bar: 5 } as Foo);`, Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "objectLiteralTypeAssertions": "allow-as-parameter"}}},
		{Code: `new Foo({ bar: 5 } as Foo);`, Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "objectLiteralTypeAssertions": "allow-as-parameter"}}},
		{Code: `throw { bar: 5 } as Foo;`, Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "objectLiteralTypeAssertions": "allow-as-parameter"}}},
		{Code: `const x = { bar: 5 } as any;`, Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "objectLiteralTypeAssertions": "allow-as-parameter"}}},
		{Code: `const x = { bar: 5 } as unknown;`, Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "objectLiteralTypeAssertions": "allow-as-parameter"}}},
		{Code: `const x = { bar: 5 } as const;`, Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "objectLiteralTypeAssertions": "allow-as-parameter"}}},
		{Code: `function foo() { throw { bar: 5 } as Foo; }`, Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "objectLiteralTypeAssertions": "allow-as-parameter"}}},
		{Code: `const foo = (x = { bar: 5 } as Foo) => {};`, Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "objectLiteralTypeAssertions": "allow-as-parameter"}}},

		// arrayLiteralTypeAssertions: 'never'
		{Code: `const x: string[] = [];`, Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "arrayLiteralTypeAssertions": "never"}}},
		{Code: `const x = [] as any;`, Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "arrayLiteralTypeAssertions": "never"}}},
		{Code: `const x = [] as unknown;`, Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "arrayLiteralTypeAssertions": "never"}}},
		{Code: `const x = [] as const;`, Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "arrayLiteralTypeAssertions": "never"}}},
		{Code: `const x = 'string' as Foo;`, Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "arrayLiteralTypeAssertions": "never"}}},

		// arrayLiteralTypeAssertions: 'allow-as-parameter'
		{Code: `const x: string[] = [];`, Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "arrayLiteralTypeAssertions": "allow-as-parameter"}}},
		{Code: `foo([] as string[]);`, Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "arrayLiteralTypeAssertions": "allow-as-parameter"}}},
		{Code: `new Foo([] as string[]);`, Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "arrayLiteralTypeAssertions": "allow-as-parameter"}}},
		{Code: `throw [] as string[];`, Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "arrayLiteralTypeAssertions": "allow-as-parameter"}}},
		{Code: `const x = [] as any;`, Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "arrayLiteralTypeAssertions": "allow-as-parameter"}}},
		{Code: `const x = [] as unknown;`, Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "arrayLiteralTypeAssertions": "allow-as-parameter"}}},
		{Code: `const x = [] as const;`, Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "arrayLiteralTypeAssertions": "allow-as-parameter"}}},
		{Code: `function foo() { throw [] as string[]; }`, Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "arrayLiteralTypeAssertions": "allow-as-parameter"}}},

		// Both objectLiteralTypeAssertions and arrayLiteralTypeAssertions: 'allow-as-parameter'
		{Code: `const x: Foo = { bar: 5 };`, Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "objectLiteralTypeAssertions": "allow-as-parameter", "arrayLiteralTypeAssertions": "allow-as-parameter"}}},
		{Code: `const x: string[] = [];`, Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "objectLiteralTypeAssertions": "allow-as-parameter", "arrayLiteralTypeAssertions": "allow-as-parameter"}}},
		{Code: `foo({ bar: 5 } as Foo);`, Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "objectLiteralTypeAssertions": "allow-as-parameter", "arrayLiteralTypeAssertions": "allow-as-parameter"}}},
		{Code: `foo([] as string[]);`, Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "objectLiteralTypeAssertions": "allow-as-parameter", "arrayLiteralTypeAssertions": "allow-as-parameter"}}},
		{Code: `foo({ bar: 5 } as Foo, [] as string[]);`, Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "objectLiteralTypeAssertions": "allow-as-parameter", "arrayLiteralTypeAssertions": "allow-as-parameter"}}},

		// angle-bracket with objectLiteralTypeAssertions: 'never'
		{Code: `const x: Foo = { bar: 5 };`, Options: []interface{}{map[string]interface{}{"assertionStyle": "angle-bracket", "objectLiteralTypeAssertions": "never"}}},
		{Code: `const x = <any>{ bar: 5 };`, Options: []interface{}{map[string]interface{}{"assertionStyle": "angle-bracket", "objectLiteralTypeAssertions": "never"}}},
		{Code: `const x = <unknown>{ bar: 5 };`, Options: []interface{}{map[string]interface{}{"assertionStyle": "angle-bracket", "objectLiteralTypeAssertions": "never"}}},
		{Code: `const x = <const>{ bar: 5 };`, Options: []interface{}{map[string]interface{}{"assertionStyle": "angle-bracket", "objectLiteralTypeAssertions": "never"}}},

		// angle-bracket with arrayLiteralTypeAssertions: 'never'
		{Code: `const x: string[] = [];`, Options: []interface{}{map[string]interface{}{"assertionStyle": "angle-bracket", "arrayLiteralTypeAssertions": "never"}}},
		{Code: `const x = <any>[];`, Options: []interface{}{map[string]interface{}{"assertionStyle": "angle-bracket", "arrayLiteralTypeAssertions": "never"}}},
		{Code: `const x = <unknown>[];`, Options: []interface{}{map[string]interface{}{"assertionStyle": "angle-bracket", "arrayLiteralTypeAssertions": "never"}}},
		{Code: `const x = <const>[];`, Options: []interface{}{map[string]interface{}{"assertionStyle": "angle-bracket", "arrayLiteralTypeAssertions": "never"}}},

		// Additional edge cases
		{Code: `const x = value as string | number;`},
		{Code: `const x = value as (string | number)[];`},
		{Code: `const x = (value as Foo) as Bar;`},
		{Code: `const x = (value as Foo).bar;`},
		{Code: `const x = (value as Foo)();`},
		{Code: "const x = `template ${value as string}`;"},

		// Union types containing any/unknown
		{Code: `const x = {} as any | string;`, Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "objectLiteralTypeAssertions": "never"}}},
		{Code: `const x = {} as unknown | string;`, Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "objectLiteralTypeAssertions": "never"}}},
		{Code: `const x = [] as any | string[];`, Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "arrayLiteralTypeAssertions": "never"}}},
		{Code: `const x = [] as unknown | string[];`, Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "arrayLiteralTypeAssertions": "never"}}},
	}, []rule_tester.InvalidTestCase{
		// Default options - using angle-bracket when 'as' is required
		{
			Code: `const x = <A>b;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "angle-bracket"},
			},
		},
		{
			Code: `const x = <readonly number[]>[1];`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "angle-bracket"},
			},
		},
		{
			Code: `const x = <Foo>{ bar: 5 };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "angle-bracket"},
			},
		},
		{
			Code: `const x = <Foo>[];`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "angle-bracket"},
			},
		},

		// assertionStyle: 'angle-bracket' - using 'as' when angle-bracket is required
		{
			Code:    `const x = b as A;`,
			Options: []interface{}{map[string]interface{}{"assertionStyle": "angle-bracket"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "as"},
			},
		},
		{
			Code:    `const x = [1] as readonly number[];`,
			Options: []interface{}{map[string]interface{}{"assertionStyle": "angle-bracket"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "as"},
			},
		},
		{
			Code:    `const x = { bar: 5 } as Foo;`,
			Options: []interface{}{map[string]interface{}{"assertionStyle": "angle-bracket"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "as"},
			},
		},

		// assertionStyle: 'never'
		{
			Code:    `const x = b as A;`,
			Options: []interface{}{map[string]interface{}{"assertionStyle": "never"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "never"},
			},
		},
		{
			Code:    `const x = <A>b;`,
			Options: []interface{}{map[string]interface{}{"assertionStyle": "never"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "never"},
			},
		},
		{
			Code:    `const x = { bar: 5 } as Foo;`,
			Options: []interface{}{map[string]interface{}{"assertionStyle": "never"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "never"},
			},
		},
		{
			Code:    `const x = <Foo>{ bar: 5 };`,
			Options: []interface{}{map[string]interface{}{"assertionStyle": "never"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "never"},
			},
		},

		// objectLiteralTypeAssertions: 'never'
		{
			Code:    `const x = {} as Foo;`,
			Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "objectLiteralTypeAssertions": "never"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "object-literal-with-type-annotation"},
			},
		},
		{
			Code:    `const x = { bar: 5 } as Foo;`,
			Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "objectLiteralTypeAssertions": "never"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "object-literal-with-type-annotation"},
			},
		},
		{
			Code:    `const x = { bar: 5 } as Foo<int>;`,
			Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "objectLiteralTypeAssertions": "never"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "object-literal-with-type-annotation"},
			},
		},
		{
			Code:    `const x = { bar: 5 } as a | b;`,
			Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "objectLiteralTypeAssertions": "never"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "object-literal-with-type-annotation"},
			},
		},
		{
			Code:    `const x = ({} as A) + b;`,
			Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "objectLiteralTypeAssertions": "never"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "never-object-literal"},
			},
		},
		{
			Code:    `const x = <Foo>{ bar: 5 };`,
			Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "objectLiteralTypeAssertions": "never"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "object-literal-with-type-annotation"},
			},
		},

		// objectLiteralTypeAssertions: 'allow-as-parameter'
		{
			Code:    `const x = { bar: 5 } as Foo;`,
			Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "objectLiteralTypeAssertions": "allow-as-parameter"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "object-literal-with-type-annotation"},
			},
		},
		{
			Code:    `const x = {} as Foo;`,
			Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "objectLiteralTypeAssertions": "allow-as-parameter"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "object-literal-with-type-annotation"},
			},
		},
		{
			Code:    `const x = ({ bar: 5 } as Foo).baz;`,
			Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "objectLiteralTypeAssertions": "allow-as-parameter"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "never-object-literal"},
			},
		},
		{
			Code:    `const x = <Foo>{ bar: 5 };`,
			Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "objectLiteralTypeAssertions": "allow-as-parameter"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "object-literal-with-type-annotation"},
			},
		},

		// arrayLiteralTypeAssertions: 'never'
		{
			Code:    `const x = [] as string[];`,
			Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "arrayLiteralTypeAssertions": "never"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "array-literal-with-type-annotation"},
			},
		},
		{
			Code:    `const x = [1, 2] as number[];`,
			Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "arrayLiteralTypeAssertions": "never"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "array-literal-with-type-annotation"},
			},
		},
		{
			Code:    `const x = <string[]>[];`,
			Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "arrayLiteralTypeAssertions": "never"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "array-literal-with-type-annotation"},
			},
		},
		{
			Code:    `const x = ([] as A) + b;`,
			Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "arrayLiteralTypeAssertions": "never"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "never-array-literal"},
			},
		},

		// arrayLiteralTypeAssertions: 'allow-as-parameter'
		{
			Code:    `const x = [] as string[];`,
			Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "arrayLiteralTypeAssertions": "allow-as-parameter"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "array-literal-with-type-annotation"},
			},
		},
		{
			Code:    `const x = [1, 2] as number[];`,
			Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "arrayLiteralTypeAssertions": "allow-as-parameter"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "array-literal-with-type-annotation"},
			},
		},
		{
			Code:    `const x = ([] as string[]).length;`,
			Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "arrayLiteralTypeAssertions": "allow-as-parameter"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "never-array-literal"},
			},
		},

		// Both objectLiteralTypeAssertions and arrayLiteralTypeAssertions: 'allow-as-parameter'
		{
			Code:    `const x = { bar: 5 } as Foo;`,
			Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "objectLiteralTypeAssertions": "allow-as-parameter", "arrayLiteralTypeAssertions": "allow-as-parameter"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "object-literal-with-type-annotation"},
			},
		},
		{
			Code:    `const x = [] as string[];`,
			Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "objectLiteralTypeAssertions": "allow-as-parameter", "arrayLiteralTypeAssertions": "allow-as-parameter"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "array-literal-with-type-annotation"},
			},
		},

		// angle-bracket with objectLiteralTypeAssertions: 'never'
		{
			Code:    `const x = <Foo>{ bar: 5 };`,
			Options: []interface{}{map[string]interface{}{"assertionStyle": "angle-bracket", "objectLiteralTypeAssertions": "never"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "object-literal-with-type-annotation"},
			},
		},
		{
			Code:    `const x = <Foo>{};`,
			Options: []interface{}{map[string]interface{}{"assertionStyle": "angle-bracket", "objectLiteralTypeAssertions": "never"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "object-literal-with-type-annotation"},
			},
		},

		// angle-bracket with arrayLiteralTypeAssertions: 'never'
		{
			Code:    `const x = <string[]>[];`,
			Options: []interface{}{map[string]interface{}{"assertionStyle": "angle-bracket", "arrayLiteralTypeAssertions": "never"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "array-literal-with-type-annotation"},
			},
		},
		{
			Code:    `const x = <number[]>[1, 2];`,
			Options: []interface{}{map[string]interface{}{"assertionStyle": "angle-bracket", "arrayLiteralTypeAssertions": "never"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "array-literal-with-type-annotation"},
			},
		},

		// Additional edge cases
		{
			Code: `const x = <string | number>value;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "angle-bracket"},
			},
		},

		// Nested object/array literals
		{
			Code:    `const x = { bar: { baz: 5 } } as Foo;`,
			Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "objectLiteralTypeAssertions": "never"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "object-literal-with-type-annotation"},
			},
		},
		{
			Code:    `const x = [[]] as string[][];`,
			Options: []interface{}{map[string]interface{}{"assertionStyle": "as", "arrayLiteralTypeAssertions": "never"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "array-literal-with-type-annotation"},
			},
		},
	})
}
