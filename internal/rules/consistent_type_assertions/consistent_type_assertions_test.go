package consistent_type_assertions

import (
	"testing"

	"github.com/typescript-eslint/rslint/internal/rule_tester"
	"github.com/typescript-eslint/rslint/internal/rules/fixtures"
)

func TestConsistentTypeAssertionsRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ConsistentTypeAssertionsRule,
		[]rule_tester.ValidTestCase{
			// Basic 'as' style tests
			{
				Code: "const x = new Generic<int>() as Foo;",
				Options: map[string]any{
					"assertionStyle":               "as",
					"objectLiteralTypeAssertions": "allow",
				},
			},
			{
				Code: "const x = b as A;",
				Options: map[string]any{
					"assertionStyle":               "as",
					"objectLiteralTypeAssertions": "allow",
				},
			},
			{
				Code: "const x = [1] as readonly number[];",
				Options: map[string]any{
					"assertionStyle":               "as",
					"objectLiteralTypeAssertions": "allow",
				},
			},
			{
				Code: "const x = 'string' as a | b;",
				Options: map[string]any{
					"assertionStyle":               "as",
					"objectLiteralTypeAssertions": "allow",
				},
			},
			{
				Code: "const x = !'string' as A;",
				Options: map[string]any{
					"assertionStyle":               "as",
					"objectLiteralTypeAssertions": "allow",
				},
			},
			{
				Code: "const x = (a as A) + b;",
				Options: map[string]any{
					"assertionStyle":               "as",
					"objectLiteralTypeAssertions": "allow",
				},
			},
			{
				Code: "const x = new Generic<string>() as Foo;",
				Options: map[string]any{
					"assertionStyle":               "as",
					"objectLiteralTypeAssertions": "allow",
				},
			},
			{
				Code: "const x = () => ({ bar: 5 }) as Foo;",
				Options: map[string]any{
					"assertionStyle":               "as",
					"objectLiteralTypeAssertions": "allow",
				},
			},
			{
				Code: "const x = () => bar as Foo;",
				Options: map[string]any{
					"assertionStyle":               "as",
					"objectLiteralTypeAssertions": "allow",
				},
			},
			{
				Code: "const x = { key: 'value' } as const;",
				Options: map[string]any{
					"assertionStyle":               "as",
					"objectLiteralTypeAssertions": "allow",
				},
			},

			// Basic 'angle-bracket' style tests
			{
				Code: "const x = <Foo>new Generic<int>();",
				Options: map[string]any{
					"assertionStyle":               "angle-bracket",
					"objectLiteralTypeAssertions": "allow",
				},
			},
			{
				Code: "const x = <A>b;",
				Options: map[string]any{
					"assertionStyle":               "angle-bracket",
					"objectLiteralTypeAssertions": "allow",
				},
			},
			{
				Code: "const x = <readonly number[]>[1];",
				Options: map[string]any{
					"assertionStyle":               "angle-bracket",
					"objectLiteralTypeAssertions": "allow",
				},
			},
			{
				Code: "const x = <a | b>'string';",
				Options: map[string]any{
					"assertionStyle":               "angle-bracket",
					"objectLiteralTypeAssertions": "allow",
				},
			},
			{
				Code: "const x = <A>!'string';",
				Options: map[string]any{
					"assertionStyle":               "angle-bracket",
					"objectLiteralTypeAssertions": "allow",
				},
			},
			{
				Code: "const x = <A>a + b;",
				Options: map[string]any{
					"assertionStyle":               "angle-bracket",
					"objectLiteralTypeAssertions": "allow",
				},
			},
			{
				Code: "const x = <Foo>new Generic<string>();",
				Options: map[string]any{
					"assertionStyle":               "angle-bracket",
					"objectLiteralTypeAssertions": "allow",
				},
			},
			{
				Code: "const x = () => <Foo>{ bar: 5 };",
				Options: map[string]any{
					"assertionStyle":               "angle-bracket",
					"objectLiteralTypeAssertions": "allow",
				},
			},
			{
				Code: "const x = () => <Foo>bar;",
				Options: map[string]any{
					"assertionStyle":               "angle-bracket",
					"objectLiteralTypeAssertions": "allow",
				},
			},
			{
				Code: "const x = <const>{ key: 'value' };",
				Options: map[string]any{
					"assertionStyle":               "angle-bracket",
					"objectLiteralTypeAssertions": "allow",
				},
			},

			// Object literal type assertions - allow
			{
				Code: "const x = {} as Foo<int>;",
				Options: map[string]any{
					"assertionStyle":               "as",
					"objectLiteralTypeAssertions": "allow",
				},
			},
			{
				Code: "const x = {} as a | b;",
				Options: map[string]any{
					"assertionStyle":               "as",
					"objectLiteralTypeAssertions": "allow",
				},
			},
			{
				Code: "print({ bar: 5 } as Foo);",
				Options: map[string]any{
					"assertionStyle":               "as",
					"objectLiteralTypeAssertions": "allow",
				},
			},
			{
				Code: "new print({ bar: 5 } as Foo);",
				Options: map[string]any{
					"assertionStyle":               "as",
					"objectLiteralTypeAssertions": "allow",
				},
			},

			// Object literal type assertions - allow-as-parameter
			{
				Code: "print({ bar: 5 } as Foo);",
				Options: map[string]any{
					"assertionStyle":                "as",
					"objectLiteralTypeAssertions": "allow-as-parameter",
				},
			},
			{
				Code: "new print({ bar: 5 } as Foo);",
				Options: map[string]any{
					"assertionStyle":                "as",
					"objectLiteralTypeAssertions": "allow-as-parameter",
				},
			},
			{
				Code: `
function foo() {
  throw { bar: 5 } as Foo;
}`,
				Options: map[string]any{
					"assertionStyle":                "as",
					"objectLiteralTypeAssertions": "allow-as-parameter",
				},
			},

			// Array literal type assertions - allow
			{
				Code: "const x = [] as string[];",
				Options: map[string]any{
					"assertionStyle": "as",
				},
			},
			{
				Code: "const x = ['a'] as Array<string>;",
				Options: map[string]any{
					"assertionStyle": "as",
				},
			},
			{
				Code: "const x = <string[]>[];",
				Options: map[string]any{
					"assertionStyle": "angle-bracket",
				},
			},
			{
				Code: "const x = <Array<string>>[];",
				Options: map[string]any{
					"assertionStyle": "angle-bracket",
				},
			},

			// Array literal type assertions - allow-as-parameter
			{
				Code: "print([5] as Foo);",
				Options: map[string]any{
					"arrayLiteralTypeAssertions": "allow-as-parameter",
					"assertionStyle":              "as",
				},
			},
			{
				Code: `
function foo() {
  throw [5] as Foo;
}`,
				Options: map[string]any{
					"arrayLiteralTypeAssertions": "allow-as-parameter",
					"assertionStyle":              "as",
				},
			},
			{
				Code: "new Print([5] as Foo);",
				Options: map[string]any{
					"arrayLiteralTypeAssertions": "allow-as-parameter",
					"assertionStyle":              "as",
				},
			},

			// Never style with const assertions
			{
				Code: "const x = <const>[1];",
				Options: map[string]any{
					"assertionStyle": "never",
				},
			},
			{
				Code: "const x = [1] as const;",
				Options: map[string]any{
					"assertionStyle": "never",
				},
			},

			// Any/unknown type assertions
			{
				Code: "const x = { key: 'value' } as any;",
				Options: map[string]any{
					"assertionStyle":               "as",
					"objectLiteralTypeAssertions": "never",
				},
			},
			{
				Code: "const x = { key: 'value' } as unknown;",
				Options: map[string]any{
					"assertionStyle":               "as",
					"objectLiteralTypeAssertions": "never",
				},
			},
		},
		[]rule_tester.InvalidTestCase{
			// Wrong assertion style - should be angle-bracket
			{
				Code: "const x = new Generic<int>() as Foo;",
				Options: map[string]any{
					"assertionStyle": "angle-bracket",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "as-assertion",
						Line:      1,
						Column:    11,
					},
				},
			},
			{
				Code: "const x = b as A;",
				Options: map[string]any{
					"assertionStyle": "angle-bracket",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "as-assertion",
						Line:      1,
						Column:    11,
					},
				},
			},

			// Wrong assertion style - should be as
			{
				Code: "const x = <Foo>new Generic<int>();",
				Options: map[string]any{
					"assertionStyle": "as",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "angle-bracket-assertion",
						Line:      1,
						Column:    11,
					},
				},
			},
			{
				Code: "const x = <A>b;",
				Options: map[string]any{
					"assertionStyle": "as",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "angle-bracket-assertion",
						Line:      1,
						Column:    11,
					},
				},
			},

			// Never style
			{
				Code: "const x = new Generic<int>() as Foo;",
				Options: map[string]any{
					"assertionStyle": "never",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "never",
						Line:      1,
						Column:    11,
					},
				},
			},
			{
				Code: "const x = b as A;",
				Options: map[string]any{
					"assertionStyle": "never",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "never",
						Line:      1,
						Column:    11,
					},
				},
			},
			{
				Code: "const x = <Foo>new Generic<int>();",
				Options: map[string]any{
					"assertionStyle": "never",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "never",
						Line:      1,
						Column:    11,
					},
				},
			},
			{
				Code: "const x = <A>b;",
				Options: map[string]any{
					"assertionStyle": "never",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "never",
						Line:      1,
						Column:    11,
					},
				},
			},

			// Object type assertions - never
			{
				Code: "const x = {} as Foo<int>;",
				Options: map[string]any{
					"assertionStyle":               "as",
					"objectLiteralTypeAssertions": "never",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "object-literal-assertion",
						Line:      1,
						Column:    11,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "object-literal-assertion-suggestion",
								Output:    "const x: Foo<int> = {};",
							},
							{
								MessageId: "object-literal-assertion-suggestion",
								Output:    "const x = {} satisfies Foo<int>;",
							},
						},
					},
				},
			},
			{
				Code: "const x = {} as a | b;",
				Options: map[string]any{
					"assertionStyle":               "as",
					"objectLiteralTypeAssertions": "never",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "object-literal-assertion",
						Line:      1,
						Column:    11,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "object-literal-assertion-suggestion",
								Output:    "const x: a | b = {};",
							},
							{
								MessageId: "object-literal-assertion-suggestion",
								Output:    "const x = {} satisfies a | b;",
							},
						},
					},
				},
			},
			{
				Code: "print({ bar: 5 } as Foo);",
				Options: map[string]any{
					"assertionStyle":               "as",
					"objectLiteralTypeAssertions": "never",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "object-literal-assertion",
						Line:      1,
						Column:    7,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "object-literal-assertion-suggestion",
								Output:    "print({ bar: 5 } satisfies Foo);",
							},
						},
					},
				},
			},

			// Object type assertions - allow-as-parameter (should fail when not parameter)
			{
				Code: "const x = {} as Foo<int>;",
				Options: map[string]any{
					"assertionStyle":                "as",
					"objectLiteralTypeAssertions": "allow-as-parameter",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "object-literal-assertion",
						Line:      1,
						Column:    11,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "object-literal-assertion-suggestion",
								Output:    "const x: Foo<int> = {};",
							},
							{
								MessageId: "object-literal-assertion-suggestion",
								Output:    "const x = {} satisfies Foo<int>;",
							},
						},
					},
				},
			},

			// Array type assertions - never
			{
				Code: "const x = [] as string[];",
				Options: map[string]any{
					"arrayLiteralTypeAssertions": "never",
					"assertionStyle":              "as",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "array-literal-assertion",
						Line:      1,
						Column:    11,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "array-literal-assertion-suggestion",
								Output:    "const x: string[] = [];",
							},
							{
								MessageId: "array-literal-assertion-suggestion",
								Output:    "const x = [] satisfies string[];",
							},
						},
					},
				},
			},
			{
				Code: "const x = <string[]>[];",
				Options: map[string]any{
					"arrayLiteralTypeAssertions": "never",
					"assertionStyle":              "angle-bracket",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "array-literal-assertion",
						Line:      1,
						Column:    11,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "array-literal-assertion-suggestion",
								Output:    "const x: string[] = [];",
							},
							{
								MessageId: "array-literal-assertion-suggestion",
								Output:    "const x = [] satisfies string[];",
							},
						},
					},
				},
			},

			// Array type assertions - allow-as-parameter (should fail when not parameter)
			{
				Code: "const foo = () => [5] as Foo;",
				Options: map[string]any{
					"arrayLiteralTypeAssertions": "allow-as-parameter",
					"assertionStyle":              "as",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "array-literal-assertion",
						Line:      1,
						Column:    19,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "array-literal-assertion-suggestion",
								Output:    "const foo = () => [5] satisfies Foo;",
							},
						},
					},
				},
			},
		},
	)
}