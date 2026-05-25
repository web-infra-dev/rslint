package no_self_compare

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoSelfCompareRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoSelfCompareRule,

		[]rule_tester.ValidTestCase{
			// ---- Upstream ESLint suite ----
			{Code: `if (x === y) { }`},
			{Code: `if (1 === 2) { }`},
			{Code: `y=x*x`},
			{Code: `foo.bar.baz === foo.bar.qux`},
			{Code: `class C { #field; foo() { this.#field === this['#field']; } }`},
			{Code: `class C { #field; foo() { this['#field'] === this.#field; } }`},

			// ---- Non-comparison binary operators: rule only triggers on the 8 comparison ops ----
			{Code: `x + x`},
			{Code: `x - x`},
			{Code: `x * x`},
			{Code: `x / x`},
			{Code: `x % x`},
			{Code: `x ** x`},
			{Code: `x & x`},
			{Code: `x | x`},
			{Code: `x ^ x`},
			{Code: `x << x`},
			{Code: `x >> x`},
			{Code: `x >>> x`},
			{Code: `x && x`},
			{Code: `x || x`},
			{Code: `x ?? x`},
			{Code: `x, x`},
			{Code: `x = x`},

			// ---- Structurally different operands ----
			{Code: `foo() === bar()`},
			{Code: `a.b === a.c`},
			{Code: `a[0] === a[1]`},
			{Code: `(a + b) === (a - b)`},
			{Code: `foo.bar() === foo.bar`},
			{Code: `a?.b === a.b`}, // optional chain preserved

			// ---- Unary / update differences (different Kind on one side) ----
			{Code: `+x === x`},
			{Code: `-x === x`},
			{Code: `!x === x`},
			{Code: `typeof x === x`},
			{Code: `++x === x`},

			// ---- Same Kind, different operator token ----
			// Regression guard: PrefixUnary / PostfixUnary store their operator
			// as a Kind enum that tsgo's ForEachChild does NOT visit, so these
			// would collapse under a naive children-only compare.
			{Code: `+x === -x`},
			{Code: `++x === --x`},
			{Code: `x++ === x--`},
			{Code: `~x === !x`},

			// ---- Different literal kinds or values ----
			{Code: `1 === 1n`},
			{Code: `1 === 2`},
			{Code: `'a' === 'b'`},
			{Code: "`a` === `b`"},

			// ---- Template with differing quasi ----
			{Code: "`a${x}` === `b${x}`"},
			{Code: "`${x}` === `${y}`"},

			// ---- TypeScript syntax shouldn't confuse structural compare ----
			{Code: `(x as number) === (y as number)`},
			{Code: `(x!) === (y!)`},

			// ---- Token-form-sensitive (matches ESLint: different source tokens) ----
			// HasSameTokens compares raw source at each leaf, so these are
			// distinguished just like ESLint's token-level getTokens() would.
			{Code: `0x1 === 1`},
			{Code: `1n === 0x1n`},
			{Code: `'a' === "a"`},
			{Code: `1e2 === 100`},
			{Code: `1.0 === 1`},
		},

		[]rule_tester.InvalidTestCase{
			// ---- Upstream ESLint suite ----
			{
				Code: `if (x === x) { }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparingToSelf", Message: "Comparing to itself is potentially pointless.", Line: 1, Column: 5, EndLine: 1, EndColumn: 12},
				},
			},
			{
				Code: `if (x !== x) { }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparingToSelf", Line: 1, Column: 5, EndLine: 1, EndColumn: 12},
				},
			},
			{
				Code: `if (x > x) { }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparingToSelf", Line: 1, Column: 5, EndLine: 1, EndColumn: 10},
				},
			},
			{
				Code: `if ('x' > 'x') { }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparingToSelf", Line: 1, Column: 5, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code: `do {} while (x === x)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparingToSelf", Line: 1, Column: 14, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code: `x === x`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparingToSelf", Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
				},
			},
			{
				Code: `x !== x`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparingToSelf", Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
				},
			},
			{
				Code: `x == x`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparingToSelf", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
				},
			},
			{
				Code: `x != x`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparingToSelf", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
				},
			},
			{
				Code: `x > x`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparingToSelf", Line: 1, Column: 1, EndLine: 1, EndColumn: 6},
				},
			},
			{
				Code: `x < x`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparingToSelf", Line: 1, Column: 1, EndLine: 1, EndColumn: 6},
				},
			},
			{
				Code: `x >= x`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparingToSelf", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
				},
			},
			{
				Code: `x <= x`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparingToSelf", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
				},
			},
			{
				Code: `foo.bar().baz.qux >= foo.bar ().baz .qux`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparingToSelf", Line: 1, Column: 1, EndLine: 1, EndColumn: 41},
				},
			},
			{
				Code: `class C { #field; foo() { this.#field === this.#field; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparingToSelf", Line: 1, Column: 27, EndLine: 1, EndColumn: 54},
				},
			},

			// ---- Extra: full operator coverage on compound / complex expressions ----
			{
				Code: `a.b.c === a.b.c`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparingToSelf", Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code: `a[0] === a[0]`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparingToSelf", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code: `foo() === foo()`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparingToSelf", Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
				},
			},
			// tsgo strips parens structurally; self-compare still detected.
			{
				Code: `(x) === x`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparingToSelf", Line: 1, Column: 1, EndLine: 1, EndColumn: 10},
				},
			},
			// ---- Multi-line position assertion ----
			{
				Code: "if (\n  x\n  ===\n  x\n) {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparingToSelf", Line: 2, Column: 3, EndLine: 4, EndColumn: 4},
				},
			},
			// ---- Multi-byte / CJK identifier ----
			{
				Code: `中 === 中`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparingToSelf", Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
				},
			},

			// ---- Same literal value, same source form → flagged (aligns with ESLint) ----
			{
				Code: `0x1 === 0x1`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparingToSelf", Line: 1, Column: 1, EndLine: 1, EndColumn: 12},
				},
			},
			{
				Code: `1n === 1n`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparingToSelf", Line: 1, Column: 1, EndLine: 1, EndColumn: 10},
				},
			},

			// ---- Same-operator unary/update — regression guard that the operator
			//      gate doesn't over-filter the valid self-compare cases. ----
			{
				Code: `+x === +x`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparingToSelf", Line: 1, Column: 1, EndLine: 1, EndColumn: 10},
				},
			},
			{
				Code: `x++ === x++`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparingToSelf", Line: 1, Column: 1, EndLine: 1, EndColumn: 12},
				},
			},

			// ---- Multi-byte (UTF-16 surrogate pair, BMP-outside emoji in string) ----
			// LSP ranges are UTF-16-based; an emoji is 2 code units, not 1 rune.
			{
				Code: `'🍎' === '🍎'`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparingToSelf", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
				},
			},
		},
	)
}
