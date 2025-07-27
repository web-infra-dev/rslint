package no_non_null_asserted_optional_chain

import (
	"testing"

	"github.com/typescript-eslint/rslint/internal/rule_tester"
	"github.com/typescript-eslint/rslint/internal/rules/fixtures"
)

func TestNoNonNullAssertedOptionalChain(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoNonNullAssertedOptionalChainRule, []rule_tester.ValidTestCase{
		{Code: `foo.bar!;`},
		{Code: `foo.bar!.baz;`},
		{Code: `foo.bar!.baz();`},
		{Code: `foo.bar()!;`},
		{Code: `foo.bar()!();`},
		{Code: `foo.bar()!.baz;`},
		{Code: `foo?.bar;`},
		{Code: `foo?.bar();`},
		{Code: `(foo?.bar).baz!;`},
		{Code: `(foo?.bar()).baz!;`},
		{Code: `foo?.bar!.baz;`},
		{Code: `foo?.bar!();`},
		{Code: `foo?.['bar']!.baz;`},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `foo?.bar!;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNullOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemovingNonNull",
							Output:    `foo?.bar;`,
						},
					},
				},
			},
		},
		{
			Code: `foo?.['bar']!;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNullOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemovingNonNull",
							Output:    `foo?.['bar'];`,
						},
					},
				},
			},
		},
		{
			Code: `foo?.bar()!;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNullOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemovingNonNull",
							Output:    `foo?.bar();`,
						},
					},
				},
			},
		},
		{
			Code: `foo.bar?.()!;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNullOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemovingNonNull",
							Output:    `foo.bar?.();`,
						},
					},
				},
			},
		},
		{
			Code: `(foo?.bar)!.baz`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNullOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemovingNonNull",
							Output:    `(foo?.bar).baz`,
						},
					},
				},
			},
		},
		{
			Code: `(foo?.bar)!().baz`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNullOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemovingNonNull",
							Output:    `(foo?.bar)().baz`,
						},
					},
				},
			},
		},
		{
			Code: `(foo?.bar)!`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNullOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemovingNonNull",
							Output:    `(foo?.bar)`,
						},
					},
				},
			},
		},
		{
			Code: `(foo?.bar)!()`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNullOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemovingNonNull",
							Output:    `(foo?.bar)()`,
						},
					},
				},
			},
		},
		{
			Code: `(foo?.bar!)`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNullOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemovingNonNull",
							Output:    `(foo?.bar)`,
						},
					},
				},
			},
		},
		{
			Code: `(foo?.bar!)()`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNullOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemovingNonNull",
							Output:    `(foo?.bar)()`,
						},
					},
				},
			},
		},
	})
}