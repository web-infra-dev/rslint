package no_non_null_asserted_nullish_coalescing

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
)

func TestNoNonNullAssertedNullishCoalescingRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoNonNullAssertedNullishCoalescingRule, []rule_tester.ValidTestCase{
		{Code: `foo ?? bar;`},
		{Code: `foo ?? bar!;`},
		{Code: `foo.bazz ?? bar;`},
		{Code: `foo.bazz ?? bar!;`},
		{Code: `foo!.bazz ?? bar;`},
		{Code: `foo!.bazz ?? bar!;`},
		{Code: `foo() ?? bar;`},
		{Code: `foo() ?? bar!;`},
		{Code: `(foo ?? bar)!;`},
		{Code: `
      let x: string;
      x! ?? '';
    `},
		{Code: `
      let x: string;
      x ?? '';
    `},
		{Code: `
      let x!: string;
      x ?? '';
    `},
		{Code: `
      let x: string;
      foo(x);
      x! ?? '';
    `},
		{Code: `
      let x: string;
      x! ?? '';
      x = foo();
    `},
		{Code: `
      let x: string;
      foo(x);
      x! ?? '';
      x = foo();
    `},
		{Code: `
      let x = foo();
      x ?? '';
    `},
		{Code: `
      function foo() {
        let x: string;
        return x ?? '';
      }
    `},
		{Code: `
      let x: string;
      function foo() {
        return x ?? '';
      }
    `},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `foo! ?? bar;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNullAssertedNullishCoalescing",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 5,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemovingNonNull",
							Output:    `foo ?? bar;`,
						},
					},
				},
			},
		},
		{
			Code: `foo! ?? bar!;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNullAssertedNullishCoalescing",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 5,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemovingNonNull",
							Output:    `foo ?? bar!;`,
						},
					},
				},
			},
		},
		{
			Code: `foo.bazz! ?? bar;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNullAssertedNullishCoalescing",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 10,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemovingNonNull",
							Output:    `foo.bazz ?? bar;`,
						},
					},
				},
			},
		},
		{
			Code: `foo.bazz! ?? bar!;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNullAssertedNullishCoalescing",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 10,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemovingNonNull",
							Output:    `foo.bazz ?? bar!;`,
						},
					},
				},
			},
		},
		{
			Code: `foo!.bazz! ?? bar;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNullAssertedNullishCoalescing",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 11,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemovingNonNull",
							Output:    `foo!.bazz ?? bar;`,
						},
					},
				},
			},
		},
		{
			Code: `foo!.bazz! ?? bar!;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNullAssertedNullishCoalescing",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 11,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemovingNonNull",
							Output:    `foo!.bazz ?? bar!;`,
						},
					},
				},
			},
		},
		{
			Code: `foo()! ?? bar;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNullAssertedNullishCoalescing",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 7,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemovingNonNull",
							Output:    `foo() ?? bar;`,
						},
					},
				},
			},
		},
		{
			Code: `foo()! ?? bar!;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNullAssertedNullishCoalescing",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 7,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemovingNonNull",
							Output:    `foo() ?? bar!;`,
						},
					},
				},
			},
		},
		{
			Code: `
let x!: string;
x! ?? '';
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNullAssertedNullishCoalescing",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemovingNonNull",
							Output: `
let x!: string;
x ?? '';
      `,
						},
					},
				},
			},
		},
		{
			Code: `
let x: string;
x = foo();
x! ?? '';
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNullAssertedNullishCoalescing",
					Line:      4,
					Column:    1,
					EndLine:   4,
					EndColumn: 3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemovingNonNull",
							Output: `
let x: string;
x = foo();
x ?? '';
      `,
						},
					},
				},
			},
		},
		{
			Code: `
let x: string;
x = foo();
x! ?? '';
x = foo();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNullAssertedNullishCoalescing",
					Line:      4,
					Column:    1,
					EndLine:   4,
					EndColumn: 3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemovingNonNull",
							Output: `
let x: string;
x = foo();
x ?? '';
x = foo();
      `,
						},
					},
				},
			},
		},
		{
			Code: `
let x = foo();
x! ?? '';
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNullAssertedNullishCoalescing",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemovingNonNull",
							Output: `
let x = foo();
x ?? '';
      `,
						},
					},
				},
			},
		},
		{
			Code: `
function foo() {
  let x!: string;
  return x! ?? '';
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNullAssertedNullishCoalescing",
					Line:      4,
					Column:    10,
					EndLine:   4,
					EndColumn: 12,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemovingNonNull",
							Output: `
function foo() {
  let x!: string;
  return x ?? '';
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
let x!: string;
function foo() {
  return x! ?? '';
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNullAssertedNullishCoalescing",
					Line:      4,
					Column:    10,
					EndLine:   4,
					EndColumn: 12,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemovingNonNull",
							Output: `
let x!: string;
function foo() {
  return x ?? '';
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
let x = foo();
x  ! ?? '';
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNullAssertedNullishCoalescing",
					Line:      3,
					Column:    1,
					EndLine:   3,
					EndColumn: 5,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemovingNonNull",
							Output: `
let x = foo();
x   ?? '';
      `,
						},
					},
				},
			},
		},
	})
}
