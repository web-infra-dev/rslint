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
		// Leading trivia must not become part of the diagnostic, and the fix
		// still removes exactly the redundant assertion token.
		{
			Code: `const result = /* leading */ value!!;`,
			Output: []string{
				`const result = /* leading */ value!;`,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noExtraNonNullAssertion",
					Line:      1,
					Column:    30,
					EndLine:   1,
					EndColumn: 36,
				},
			},
		},
		// Comments between assertions are preserved verbatim by the fix.
		{
			Code: `const result = value /* first */ ! /* second */ !;`,
			Output: []string{
				`const result = value /* first */  /* second */ !;`,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noExtraNonNullAssertion"},
			},
		},
		// The token range remains byte-exact when the expression contains
		// multi-byte source text.
		{
			Code: `const result = 数据!!;`,
			Output: []string{
				`const result = 数据!;`,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noExtraNonNullAssertion"},
			},
		},
		// Every redundant token in a longer assertion chain gets its own
		// non-overlapping fix.
		{
			Code: `const result = value!!!;`,
			Output: []string{
				`const result = value!;`,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noExtraNonNullAssertion"},
				{MessageId: "noExtraNonNullAssertion"},
			},
		},
		// Exercise the optional element-access branch in addition to the
		// existing property-access and call cases.
		{
			Code: `
function foo(bar?: number[]) {
  return bar!?.[0];
}
`,
			Output: []string{`
function foo(bar?: number[]) {
  return bar?.[0];
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
	})
}
