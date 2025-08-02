package no_non_null_assertion

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
)

func TestNoNonNullAssertionRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoNonNullAssertionRule, []rule_tester.ValidTestCase{
		{Code: "x;"},
		{Code: "x.y;"},
		{Code: "x.y.z;"},
		{Code: "x?.y.z;"},
		{Code: "x?.y?.z;"},
		{Code: "!x;"},
	}, []rule_tester.InvalidTestCase{
		{
			Code:   "x!;",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNonNull"}},
		},
		{
			Code:   "x!.y;",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noNonNull",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "suggestOptionalChain",
					Output:    "x?.y;",
				}},
			}},
		},
		{
			Code:   "x.y!;",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNonNull"}},
		},
		{
			Code:   "!x!.y;",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noNonNull",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "suggestOptionalChain",
					Output:    "!x?.y;",
				}},
			}},
		},
		{
			Code:   "x!.y?.z;",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noNonNull",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "suggestOptionalChain",
					Output:    "x?.y?.z;",
				}},
			}},
		},
		{
			Code:   "x![y];",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noNonNull",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "suggestOptionalChain",
					Output:    "x?.[y];",
				}},
			}},
		},
		{
			Code:   "x![y]?.z;",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noNonNull",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "suggestOptionalChain",
					Output:    "x?.[y]?.z;",
				}},
			}},
		},
		{
			Code:   "x.y.z!();",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noNonNull",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "suggestOptionalChain",
					Output:    "x.y.z?.();",
				}},
			}},
		},
		{
			Code:   "x.y?.z!();",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noNonNull",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "suggestOptionalChain",
					Output:    "x.y?.z?.();",
				}},
			}},
		},
		// Multiple non-null assertions
		{
			Code:   "x!!!;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNonNull"},
				{MessageId: "noNonNull"},
				{MessageId: "noNonNull"},
			},
		},
		{
			Code:   "x!!.y;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "suggestOptionalChain",
						Output:    "x!?.y;",
					}},
				},
				{MessageId: "noNonNull"},
			},
		},
		{
			Code:   "x.y!!;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNonNull"},
				{MessageId: "noNonNull"},
			},
		},
		{
			Code:   "x.y.z!!();",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "suggestOptionalChain",
						Output:    "x.y.z!?.();",
					}},
				},
				{MessageId: "noNonNull"},
			},
		},
		// Optional chaining with non-null assertion
		{
			Code:   "x!?.[y].z;",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noNonNull",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "suggestOptionalChain",
					Output:    "x?.[y].z;",
				}},
			}},
		},
		{
			Code:   "x!?.y.z;",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noNonNull",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "suggestOptionalChain",
					Output:    "x?.y.z;",
				}},
			}},
		},
		{
			Code:   "x.y.z!?.();",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noNonNull",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "suggestOptionalChain",
					Output:    "x.y.z?.();",
				}},
			}},
		},
	})
}