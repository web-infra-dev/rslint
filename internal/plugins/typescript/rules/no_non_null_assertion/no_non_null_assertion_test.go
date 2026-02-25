package no_non_null_assertion

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoNonNullAssertionRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoNonNullAssertionRule, []rule_tester.ValidTestCase{
		{Code: `x;`},
		{Code: `x.y;`},
		{Code: `x.y.z;`},
		{Code: `x?.y.z;`},
		{Code: `x?.y?.z;`},
		{Code: `!x;`},
	}, []rule_tester.InvalidTestCase{
		// Simple non-null assertion — no suggestion
		{
			Code: `x!;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 3,
				},
			},
		},
		// Non-null before property access — suggest optional chain
		{
			Code: `x!.y;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestOptionalChain",
							Output:    `x?.y;`,
						},
					},
				},
			},
		},
		// Non-null at end — no suggestion
		{
			Code: `x.y!;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 5,
				},
			},
		},
		// Prefix ! with non-null before property access
		{
			Code: `!x!.y;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
					Line:      1,
					Column:    2,
					EndLine:   1,
					EndColumn: 4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestOptionalChain",
							Output:    `!x?.y;`,
						},
					},
				},
			},
		},
		// Non-null before optional chain
		{
			Code: `x!.y?.z;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestOptionalChain",
							Output:    `x?.y?.z;`,
						},
					},
				},
			},
		},
		// Non-null before element access
		{
			Code: `x![y];`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestOptionalChain",
							Output:    `x?.[y];`,
						},
					},
				},
			},
		},
		// Non-null before element access with optional chain
		{
			Code: `x![y]?.z;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestOptionalChain",
							Output:    `x?.[y]?.z;`,
						},
					},
				},
			},
		},
		// Non-null before call
		{
			Code: `x.y.z!();`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 7,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestOptionalChain",
							Output:    `x.y.z?.();`,
						},
					},
				},
			},
		},
		// Non-null before optional call
		{
			Code: `x.y?.z!();`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 8,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestOptionalChain",
							Output:    `x.y?.z?.();`,
						},
					},
				},
			},
		},
		// Triple non-null — no suggestions
		{
			Code: `x!!!;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 5,
				},
				{
					MessageId: "noNonNull",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 4,
				},
				{
					MessageId: "noNonNull",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 3,
				},
			},
		},
		// Double non-null before property access
		{
			Code: `x!!.y;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestOptionalChain",
							Output:    `x!?.y;`,
						},
					},
				},
				{
					MessageId: "noNonNull",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 3,
				},
			},
		},
		// Double non-null at end — no suggestions
		{
			Code: `x.y!!;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 6,
				},
				{
					MessageId: "noNonNull",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 5,
				},
			},
		},
		// Double non-null before call
		{
			Code: `x.y.z!!();`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 8,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestOptionalChain",
							Output:    `x.y.z!?.();`,
						},
					},
				},
				{
					MessageId: "noNonNull",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 7,
				},
			},
		},
		// Already optional element access with non-null
		{
			Code: `x!?.[y].z;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestOptionalChain",
							Output:    `x?.[y].z;`,
						},
					},
				},
			},
		},
		// Already optional property access with non-null
		{
			Code: `x!?.y.z;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestOptionalChain",
							Output:    `x?.y.z;`,
						},
					},
				},
			},
		},
		// Already optional call with non-null
		{
			Code: `x.y.z!?.();`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 7,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestOptionalChain",
							Output:    `x.y.z?.();`,
						},
					},
				},
			},
		},
	})
}
