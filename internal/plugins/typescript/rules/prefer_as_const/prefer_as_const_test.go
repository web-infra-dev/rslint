package prefer_as_const

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferAsConstRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferAsConstRule, []rule_tester.ValidTestCase{
		{Code: "let foo = 'baz' as const;"},
		{Code: "let foo = 1 as const;"},
		{Code: "let foo = { bar: 'baz' as const };"},
		{Code: "let foo = { bar: 1 as const };"},
		{Code: "let foo = { bar: 'baz' };"},
		{Code: "let foo = { bar: 2 };"},
		{Code: "let foo = <bar>'bar';"},
		{Code: "let foo = <string>'bar';"},
		{Code: "let foo = 'bar' as string;"},
		{Code: "let foo = `bar` as `bar`;"},
		{Code: "let foo = `bar` as `foo`;"},
		{Code: "let foo = `bar` as 'bar';"},
		{Code: "let foo: string = 'bar';"},
		{Code: "let foo: number = 1;"},
		{Code: "let foo: 'bar' = baz;"},
		{Code: "let foo = 'bar';"},
		{Code: "let foo: 'bar';"},
		{Code: "let foo = { bar };"},
		{Code: "let foo: 'baz' = 'baz' as const;"},
		{Code: `
			class foo {
				bar = 'baz';
			}
		`},
		{Code: `
			class foo {
				bar: 'baz';
			}
		`},
		{Code: `
			class foo {
				bar;
			}
		`},
		{Code: `
			class foo {
				bar = <baz>'baz';
			}
		`},
		{Code: `
			class foo {
				bar: string = 'baz';
			}
		`},
		{Code: `
			class foo {
				bar: number = 1;
			}
		`},
		{Code: `
			class foo {
				bar = 'baz' as const;
			}
		`},
		{Code: `
			class foo {
				bar = 2 as const;
			}
		`},
		{Code: `
			class foo {
				get bar(): 'bar' {}
				set bar(bar: 'bar') {}
			}
		`},
		{Code: `
			class foo {
				bar = () => 'bar' as const;
			}
		`},
		{Code: `
			type BazFunction = () => 'baz';
			class foo {
				bar: BazFunction = () => 'bar';
			}
		`},
		{Code: `
			class foo {
				bar(): void {}
			}
		`},
	}, []rule_tester.InvalidTestCase{
		{
			Code: "let foo = { bar: 'baz' as 'baz' };",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferConstAssertion",
					Line:      1,
					Column:    27,
				},
			},
			Output: []string{"let foo = { bar: 'baz' as const };"},
		},
		{
			Code: "let foo = { bar: 1 as 1 };",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferConstAssertion",
					Line:      1,
					Column:    23,
				},
			},
			Output: []string{"let foo = { bar: 1 as const };"},
		},
		{
			Code: "let []: 'bar' = 'bar';",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "variableConstAssertion",
					Line:      1,
					Column:    9,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "variableSuggest",
							Output:    "let [] = 'bar' as const;",
						},
					},
				},
			},
		},
		{
			Code: "let foo: 'bar' = 'bar';",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "variableConstAssertion",
					Line:      1,
					Column:    10,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "variableSuggest",
							Output:    "let foo = 'bar' as const;",
						},
					},
				},
			},
		},
		{
			Code: "let foo: 2 = 2;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "variableConstAssertion",
					Line:      1,
					Column:    10,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "variableSuggest",
							Output:    "let foo = 2 as const;",
						},
					},
				},
			},
		},
		{
			Code: "let foo: 'bar' = 'bar' as 'bar';",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferConstAssertion",
					Line:      1,
					Column:    27,
				},
			},
			Output: []string{"let foo: 'bar' = 'bar' as const;"},
		},
		{
			Code: "let foo = <'bar'>'bar';",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferConstAssertion",
					Line:      1,
					Column:    12,
				},
			},
			Output: []string{"let foo = <const>'bar';"},
		},
		{
			Code: "let foo = <4>4;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferConstAssertion",
					Line:      1,
					Column:    12,
				},
			},
			Output: []string{"let foo = <const>4;"},
		},
		{
			Code: "let foo = 'bar' as 'bar';",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferConstAssertion",
					Line:      1,
					Column:    20,
				},
			},
			Output: []string{"let foo = 'bar' as const;"},
		},
		{
			Code: "let foo = 5 as 5;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferConstAssertion",
					Line:      1,
					Column:    16,
				},
			},
			Output: []string{"let foo = 5 as const;"},
		},
		{
			Code: `
class foo {
	bar: 'baz' = 'baz';
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "variableConstAssertion",
					Line:      3,
					Column:    7,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "variableSuggest",
							Output: `
class foo {
	bar = 'baz' as const;
}
			`,
						},
					},
				},
			},
		},
		{
			Code: `
class foo {
	bar: 2 = 2;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "variableConstAssertion",
					Line:      3,
					Column:    7,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "variableSuggest",
							Output: `
class foo {
	bar = 2 as const;
}
			`,
						},
					},
				},
			},
		},
		{
			Code: `
class foo {
	foo = <'bar'>'bar';
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferConstAssertion",
					Line:      3,
					Column:    9,
				},
			},
			Output: []string{`
class foo {
	foo = <const>'bar';
}
			`},
		},
		{
			Code: `
class foo {
	foo = 'bar' as 'bar';
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferConstAssertion",
					Line:      3,
					Column:    17,
				},
			},
			Output: []string{`
class foo {
	foo = 'bar' as const;
}
			`},
		},
		{
			Code: `
class foo {
	foo = 5 as 5;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferConstAssertion",
					Line:      3,
					Column:    13,
				},
			},
			Output: []string{`
class foo {
	foo = 5 as const;
}
			`},
		},
	})
}
