package no_non_null_asserted_optional_chain

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoNonNullAssertedOptionalChainRule(t *testing.T) {
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
		// Non-null assertion in middle of optional chain (not outermost)
		{Code: `foo?.bar!.baz;`},
		{Code: `foo?.bar!();`},
		{Code: `foo?.['bar']!.baz;`},
	}, []rule_tester.InvalidTestCase{
		// End of chain: foo?.bar!
		{
			Code: `foo?.bar!;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNullOptionalChain",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemovingNonNull",
							Output:    `foo?.bar;`,
						},
					},
				},
			},
		},
		// End of chain: foo?.['bar']!
		{
			Code: `foo?.['bar']!;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNullOptionalChain",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemovingNonNull",
							Output:    `foo?.['bar'];`,
						},
					},
				},
			},
		},
		// End of chain: foo?.bar()!
		{
			Code: `foo?.bar()!;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNullOptionalChain",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemovingNonNull",
							Output:    `foo?.bar();`,
						},
					},
				},
			},
		},
		// End of chain: foo.bar?.()!
		{
			Code: `foo.bar?.()!;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNullOptionalChain",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemovingNonNull",
							Output:    `foo.bar?.();`,
						},
					},
				},
			},
		},
		// Wrapping: (foo?.bar)!.baz
		{
			Code: `(foo?.bar)!.baz`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNullOptionalChain",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemovingNonNull",
							Output:    `(foo?.bar).baz`,
						},
					},
				},
			},
		},
		// Wrapping: (foo?.bar)!().baz
		{
			Code: `(foo?.bar)!().baz`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNullOptionalChain",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemovingNonNull",
							Output:    `(foo?.bar)().baz`,
						},
					},
				},
			},
		},
		// Wrapping: (foo?.bar)!
		{
			Code: `(foo?.bar)!`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNullOptionalChain",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemovingNonNull",
							Output:    `(foo?.bar)`,
						},
					},
				},
			},
		},
		// Wrapping: (foo?.bar)!()
		{
			Code: `(foo?.bar)!()`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNullOptionalChain",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemovingNonNull",
							Output:    `(foo?.bar)()`,
						},
					},
				},
			},
		},
		// Inside parens at end of chain: (foo?.bar!)
		{
			Code: `(foo?.bar!)`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNullOptionalChain",
					Line:      1,
					Column:    2,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemovingNonNull",
							Output:    `(foo?.bar)`,
						},
					},
				},
			},
		},
		// Inside parens at end of chain called: (foo?.bar!)()
		{
			Code: `(foo?.bar!)()`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNullOptionalChain",
					Line:      1,
					Column:    2,
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
