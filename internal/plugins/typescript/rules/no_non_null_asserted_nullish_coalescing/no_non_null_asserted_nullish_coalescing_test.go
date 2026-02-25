package no_non_null_asserted_nullish_coalescing

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
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
		// Variable declared without initializer or definite assignment — no prior assignment
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
		// foo(x) is not an assignment to x
		{Code: `
let x: string;
foo(x);
x! ?? '';
`},
		// Assignment is after the node
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
		// Initialized but no non-null assertion
		{Code: `
let x = foo();
x ?? '';
`},
		// Function scope
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestRemovingNonNull",
							Output:    `foo() ?? bar!;`,
						},
					},
				},
			},
		},
		// Definite assignment
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
		// Assignment before node
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
		// Assignment before and after
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
		// Initialized variable
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
		// Definite assignment in function scope
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
		// Definite assignment in outer scope, used in function
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
	})
}
