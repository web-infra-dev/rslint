package no_extra_non_null_assertion

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoExtraNonNullAssertionRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoExtraNonNullAssertionRule, []rule_tester.ValidTestCase{
		// Valid: Single non-null assertion on nullable object property access
		{Code: `const foo: { bar: number } | null = null; const bar = foo!.bar;`},
		// Valid: Single non-null assertion on parameter
		{Code: `function foo(bar: number | undefined) { const bar: number = bar!; }`},
		// Valid: Optional chaining without non-null assertion
		{Code: `function foo(bar?: { n: number }) { return bar?.n; }`},
		// Valid: Non-null assertion with optional chaining (GitHub issue #2166)
		{Code: `checksCounter?.textContent!.trim();`},
		// Valid: Non-null assertion in computed property access (GitHub issue #2732)
		{Code: `function foo(key: string | null) { const obj = {}; return obj?.[key!]; }`},
	}, []rule_tester.InvalidTestCase{
		// Invalid: Double non-null assertion on object
		{
			Code: `const foo: { bar: number } | null = null;
const bar = foo!!.bar;`,
			Output: []string{`const foo: { bar: number } | null = null;
const bar = foo!.bar;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noExtraNonNullAssertion",
					Line:      2,
					Column:    13,
					EndLine:   2,
					EndColumn: 18,
				},
			},
		},
		// Invalid: Double non-null assertion on parameter
		{
			Code:   `function foo(bar: number | undefined) { const bar: number = bar!!; }`,
			Output: []string{`function foo(bar: number | undefined) { const bar: number = bar!; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noExtraNonNullAssertion",
					Line:      1,
					Column:    61,
					EndLine:   1,
					EndColumn: 66,
				},
			},
		},
		// Invalid: Non-null with optional chaining on property access
		{
			Code:   `function foo(bar?: { n: number }) { return bar!?.n; }`,
			Output: []string{`function foo(bar?: { n: number }) { return bar?.n; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noExtraNonNullAssertion",
					Line:      1,
					Column:    44,
					EndLine:   1,
					EndColumn: 48,
				},
			},
		},
		// Invalid: Non-null with optional call
		{
			Code:   `function foo(bar?: { n: number }) { return bar!?.(); }`,
			Output: []string{`function foo(bar?: { n: number }) { return bar?.(); }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noExtraNonNullAssertion",
					Line:      1,
					Column:    44,
					EndLine:   1,
					EndColumn: 48,
				},
			},
		},
		// Invalid: Parenthesized double non-null assertion
		{
			Code:   `const foo: { bar: number } | null = null; const bar = (foo)!!.bar;`,
			Output: []string{`const foo: { bar: number } | null = null; const bar = (foo)!.bar;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noExtraNonNullAssertion",
					Line:      1,
					Column:    55,
					EndLine:   1,
					EndColumn: 62,
				},
			},
		},
		// Invalid: Parenthesized non-null with optional chaining
		{
			Code:   `function foo(bar?: { n: number }) { return (bar)!?.n; }`,
			Output: []string{`function foo(bar?: { n: number }) { return (bar)?.n; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noExtraNonNullAssertion",
					Line:      1,
					Column:    44,
					EndLine:   1,
					EndColumn: 50,
				},
			},
		},
		// Invalid: Parenthesized non-null with optional call
		{
			Code:   `function foo(bar?: { n: number }) { return (bar!)!?.n; }`,
			Output: []string{`function foo(bar?: { n: number }) { return (bar!)?.n; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noExtraNonNullAssertion",
					Line:      1,
					Column:    44,
					EndLine:   1,
					EndColumn: 51,
				},
			},
		},
		// Invalid: Nested parenthesized non-null with optional chaining
		{
			Code:   `function foo(bar?: { n: number }) { return ((bar))!?.n; }`,
			Output: []string{`function foo(bar?: { n: number }) { return ((bar))?.n; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noExtraNonNullAssertion",
					Line:      1,
					Column:    44,
					EndLine:   1,
					EndColumn: 52,
				},
			},
		},
	})
}
