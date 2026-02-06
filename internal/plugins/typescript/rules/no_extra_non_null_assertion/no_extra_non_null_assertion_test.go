package no_extra_non_null_assertion

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoExtraNonNullAssertionRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoExtraNonNullAssertionRule, []rule_tester.ValidTestCase{
		// Single non-null assertion is fine
		{Code: `
const foo: { bar: number } | null = null;
const bar = foo!.bar;
`},
		// Single non-null assertion on parameter
		{Code: `
function foo(bar: number | undefined) {
  const bar: number = bar!;
}
`},
		// Optional chaining without non-null assertion
		{Code: `
function foo(bar?: { n: number }) {
  return bar?.n;
}
`},
		// Non-null assertion before optional chaining (valid: not the object of optional)
		{Code: `checksCounter?.textContent!.trim();`},
		// Non-null assertion inside element access of optional chain
		{Code: `
function foo(key: string | null) {
  const obj = {};
  return obj?.[key!];
}
`},
	}, []rule_tester.InvalidTestCase{
		// Double non-null assertion: foo!!
		{
			Code: `
const foo: { bar: number } | null = null;
const bar = foo!!.bar;
`,
			Output: []string{`
const foo: { bar: number } | null = null;
const bar = foo!.bar;
`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noExtraNonNullAssertion",
					Line:      3,
					Column:    13,
					EndLine:   3,
					EndColumn: 17,
				},
			},
		},
		// Double non-null on parameter
		{
			Code: `
function foo(bar: number | undefined) {
  const bar: number = bar!!;
}
`,
			Output: []string{`
function foo(bar: number | undefined) {
  const bar: number = bar!;
}
`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noExtraNonNullAssertion",
					Line:      3,
					Column:    23,
					EndLine:   3,
					EndColumn: 27,
				},
			},
		},
		// Non-null before optional property access: bar!?.n
		{
			Code: `
function foo(bar?: { n: number }) {
  return bar!?.n;
}
`,
			Output: []string{`
function foo(bar?: { n: number }) {
  return bar?.n;
}
`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noExtraNonNullAssertion",
					Line:      3,
					Column:    10,
					EndLine:   3,
					EndColumn: 14,
				},
			},
		},
		// Non-null before optional call: bar!?.()
		{
			Code: `
function foo(bar?: { n: number }) {
  return bar!?.();
}
`,
			Output: []string{`
function foo(bar?: { n: number }) {
  return bar?.();
}
`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noExtraNonNullAssertion",
					Line:      3,
					Column:    10,
					EndLine:   3,
					EndColumn: 14,
				},
			},
		},
		// Parenthesized: (foo!)!
		{
			Code: `
const foo: { bar: number } | null = null;
const bar = (foo!)!.bar;
`,
			Output: []string{`
const foo: { bar: number } | null = null;
const bar = (foo)!.bar;
`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noExtraNonNullAssertion",
					Line:      3,
					Column:    14,
					EndLine:   3,
					EndColumn: 18,
				},
			},
		},
		// Parenthesized: (bar!)?.n
		{
			Code: `
function foo(bar?: { n: number }) {
  return (bar!)?.n;
}
`,
			Output: []string{`
function foo(bar?: { n: number }) {
  return (bar)?.n;
}
`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noExtraNonNullAssertion",
					Line:      3,
					Column:    11,
					EndLine:   3,
					EndColumn: 15,
				},
			},
		},
		// Parenthesized: (bar)!?.n
		{
			Code: `
function foo(bar?: { n: number }) {
  return (bar)!?.n;
}
`,
			Output: []string{`
function foo(bar?: { n: number }) {
  return (bar)?.n;
}
`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noExtraNonNullAssertion",
					Line:      3,
					Column:    10,
					EndLine:   3,
					EndColumn: 16,
				},
			},
		},
		// Parenthesized: (bar!)?.()
		{
			Code: `
function foo(bar?: { n: number }) {
  return (bar!)?.();
}
`,
			Output: []string{`
function foo(bar?: { n: number }) {
  return (bar)?.();
}
`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noExtraNonNullAssertion",
					Line:      3,
					Column:    11,
					EndLine:   3,
					EndColumn: 15,
				},
			},
		},
	})
}
