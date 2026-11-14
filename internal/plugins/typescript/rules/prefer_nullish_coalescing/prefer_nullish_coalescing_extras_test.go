package prefer_nullish_coalescing

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Cases beyond the upstream suite. Except for rslintOnlyInvalid, every
// expectation below matches typescript-eslint v8.72.2 run with this package's
// fixture tsconfig: messages, complete ranges and suggestion outputs.
func TestPreferNullishCoalescingExtras(t *testing.T) {
	valid := append(append([]rule_tester.ValidTestCase{}, regressionValid...), earlierValid...)
	invalid := append(append(append([]rule_tester.InvalidTestCase{}, regressionInvalid...), earlierInvalid...), rslintOnlyInvalid...)
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferNullishCoalescingRule, valid, invalid)
}

// Regressions for tsgo/ESTree shape differences: parentheses, non-null
// assertions, `||=` versus `||`, literal kinds in member keys, Boolean
// arguments, generic constraints, comments kept by the if-statement
// suggestion and the precedence of the nullish branch.
var regressionValid = []rule_tester.ValidTestCase{
	// non-null in test
	{
		Code: `declare let a: { b: string | undefined };
a.b! !== undefined ? a.b : 'x';
`,
	},
	// non-null in branch
	{
		Code: `declare let a: { b: string | undefined };
a.b !== undefined ? a.b! : 'x';
`,
	},
	// non-null truthy
	{
		Code: `declare let a: { b: string | undefined };
a.b! ? a.b : 'x';
`,
	},
	// mixed parens
	{
		Code: `declare let a: string | null, b: string, c: boolean;
(a || b) && c;
`,
		Options: map[string]any{"ignoreMixedLogicalExpressions": true},
	},
	// mixed parens right
	{
		Code: `declare let a: string | null, b: string, c: boolean;
c && (a || b);
`,
		Options: map[string]any{"ignoreMixedLogicalExpressions": true},
	},
	// Boolean second arg
	{
		Code: `declare let a: string | null, b: string;
Boolean(0, a || b);
`,
		Options: map[string]any{"ignoreBooleanCoercion": true},
	},
	// template vs string key
	{
		Code: "declare let a: { x?: string };\na[`x`] !== undefined ? a['x'] : 'y';\n",
	},
	// template vs ident key
	{
		Code: "declare let a: { x?: string };\na[`x`] !== undefined ? a.x : 'y';\n",
	},
	// generic null check
	{
		Code: `function f<T extends string | undefined>(v: T | null) {
  return v !== null ? v : 'x';
}
`,
	},
	// unknown generic
	{
		Code: `function f<T>(v: T | undefined) {
  return v !== undefined ? v : 'x';
}
`,
	},
	// ternary call chain
	{
		Code: `declare let a: { b?: () => string | undefined };
a.b?.() ? a.b?.() : 'x';
`,
	},
	// ternary as cond
	{
		Code: `declare let a: string | null, b: string | null;
(a ? a : b) ? 1 : 2;
`,
	},
	// or in Boolean first
	{
		Code: `declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };
Boolean(a || b);
`,
		Options: map[string]any{"ignoreBooleanCoercion": true},
	},
	// mixed a || b || c && d
	{
		Code: `declare let a: string | null, b: string | null, c: string, d: string;
a || b || c && d;
`,
		Options: map[string]any{"ignoreMixedLogicalExpressions": true},
	},
	// mixed paren chain
	{
		Code: `declare let a: string | null, b: string | null, c: string, d: boolean;
((a || b) || c) && d;
`,
		Options: map[string]any{"ignoreMixedLogicalExpressions": true},
	},
	// cond test via ??
	{
		Code: `declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };
if ((a || b) ?? c) {}
`,
	},
	// cond test via seq last
	{
		Code: `declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };
if ((n, a || b)) {}
`,
	},
	// cond test for
	{
		Code: `declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };
for (; a || b; ) {}
`,
	},
	// cond test not
	{
		Code: `declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };
if (!(a || b)) {}
`,
	},
	// non-null truthy branch
	{
		Code: `declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };
o.x ? o.x! : b;
`,
	},
	// non-null member object
	{
		Code: `declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };
o!.x !== undefined ? o.x : b;
`,
	},
	// private field
	{
		Code: `class C { #x?: string; m() { return this.#x !== undefined ? this.#x : 'y'; } }
`,
	},
	// if optional test
	{
		Code: `declare let foo: { a?: string } | undefined;
if (foo?.a == null) { foo!.a = 'x'; }
`,
	},
	// generic T[K]
	{
		Code: `function f<T, K extends keyof T>(o: T, k: K) {
  const v = o[k];
  return v !== undefined ? v : 'x';
}
`,
	},
	// generic constrained null-only
	{
		Code: `function f<T extends string | null>(v: T | undefined) {
  return v !== undefined ? v : 'x';
}
`,
	},
	// unconstrained T with only a null check: earlier rslint expected a report; upstream v8.72.2 (#12994) treats T as possibly undefined
	{
		Code: `
function getOr<T>(v: T | null, fallback: T): T {
  return v !== null ? v : fallback;
}
`,
		Options: map[string]any{"ignoreTernaryTests": false},
	},
}

var regressionInvalid = []rule_tester.InvalidTestCase{
	// known positive
	{
		Code: `declare let x: string | null;
x || 'a';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 3, EndLine: 2, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let x: string | null;
x ?? 'a';
`},
				},
			},
		},
	},
	// mixed through ||=
	{
		Code: `declare let a: string | null, b: string | null, c: string, d: boolean;
d && (a ||= b || c);
`,
		Options: map[string]any{"ignoreMixedLogicalExpressions": true},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 15, EndLine: 2, EndColumn: 17,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string, d: boolean;
d && (a ||= b ?? c);
`},
				},
			},
		},
	},
	// or inside ||=
	{
		Code: `declare let a: string | null, b: string | null, c: string;
a ||= b || c;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      2, Column: 3, EndLine: 2, EndColumn: 6,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string;
a ??= b || c;
`},
				},
			},
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 9, EndLine: 2, EndColumn: 11,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string;
a ||= b ?? c;
`},
				},
			},
		},
	},
	// assign left inside or chain
	{
		Code: `declare let a: string | null, b: string | null, c: string, d: string;
(a = b) || c || d;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 9, EndLine: 2, EndColumn: 11,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string, d: string;
((a = b) ?? c) || d;
`},
				},
			},
		},
	},
	// sequence left inside or chain
	{
		Code: `declare let a: string | null, b: string | null, c: string, d: string;
(a, b) || c || d;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 8, EndLine: 2, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string, d: string;
((a, b) ?? c) || d;
`},
				},
			},
		},
	},
	// paren left in chain
	{
		Code: `declare let a: string | null, c: string, d: string;
( a ) || c || d;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 7, EndLine: 2, EndColumn: 9,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, c: string, d: string;
( (a ) ?? c) || d;
`},
				},
			},
		},
	},
	// computed member key vs nested
	{
		Code: `declare let a: Record<string, string | undefined>; declare let b: { c: string }; declare let c: 'c';
a[b.c] !== undefined ? a[b[c]] : 'y';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 37,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: Record<string, string | undefined>; declare let b: { c: string }; declare let c: 'c';
a[b.c] ?? 'y';
`},
				},
			},
		},
	},
	// numeric literal key
	{
		Code: `declare let a: (string | undefined)[];
a[1] !== undefined ? a[1.0] : 'y';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 34,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: (string | undefined)[];
a[1] ?? 'y';
`},
				},
			},
		},
	},
	// NonNullable generic
	{
		Code: `function f<T>(v: NonNullable<T> | undefined) {
  return v !== undefined ? v : 'x';
}
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 10, EndLine: 2, EndColumn: 35,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `function f<T>(v: NonNullable<T> | undefined) {
  return v ?? 'x';
}
`},
				},
			},
		},
	},
	// generic constrained
	{
		Code: `function f<T extends string>(v: T | undefined) {
  return v !== undefined ? v : 'x';
}
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 10, EndLine: 2, EndColumn: 35,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `function f<T extends string>(v: T | undefined) {
  return v ?? 'x';
}
`},
				},
			},
		},
	},
	// intersection branded
	{
		Code: `type B = string & { __b: 1 };
declare let v: B | undefined;
v !== undefined ? v : 'x';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 26,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `type B = string & { __b: 1 };
declare let v: B | undefined;
v ?? 'x';
`},
				},
			},
		},
	},
	// ternary in seq middle (cond test)
	{
		Code: `declare let a: string | null;
if ((a || 'x', 1)) {}
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 8, EndLine: 2, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null;
if ((a ?? 'x', 1)) {}
`},
				},
			},
		},
	},
	// nullish branch arrow
	{
		Code: `declare let a: (() => void) | null;
const f = a !== null ? a : () => {};
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 11, EndLine: 2, EndColumn: 36,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: (() => void) | null;
const f = a ?? (() => {});
`},
				},
			},
		},
	},
	// nullish branch ternary
	{
		Code: `declare let a: string | null, b: boolean;
a !== null ? a : b ? 'q' : 'r';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 31,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: boolean;
a ?? (b ? 'q' : 'r');
`},
				},
			},
		},
	},
	// nullish branch assignment
	{
		Code: `declare let a: string | null; let z: string;
a !== null ? a : z = 'r';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 25,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null; let z: string;
a ?? (z = 'r');
`},
				},
			},
		},
	},
	// nullish branch ??
	{
		Code: `declare let a: string | null, b: string | null;
a !== null ? a : b ?? 'r';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 26,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null;
a ?? (b ?? 'r');
`},
				},
			},
		},
	},
	// nullish branch yield
	{
		Code: `function* g(a: string | null) {
  const r = a !== null ? a : yield 1;
}
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 13, EndLine: 2, EndColumn: 37,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `function* g(a: string | null) {
  const r = a ?? (yield 1);
}
`},
				},
			},
		},
	},
	// nullish branch spread-ish comma
	{
		Code: `declare let a: string | null, b: string;
a !== null ? a : (b, 'r');
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 26,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string;
a ?? (b, 'r');
`},
				},
			},
		},
	},
	// if with comments
	{
		Code: `declare let foo: string | undefined;
if (!foo) {
  // a
  foo = 'x'; // b
  /* c */
}
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 6, EndColumn: 2,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let foo: string | undefined;
// a
foo ??= 'x'; // b
/* c */
`},
				},
			},
		},
	},
	// if single stmt comment
	{
		Code: `declare let foo: string | undefined;
if (!foo) /* a */ foo = 'x';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 29,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let foo: string | undefined;
/* a */ foo ??= 'x';
`},
				},
			},
		},
	},
	// if paren assignment
	{
		Code: `declare let foo: string | undefined;
if (foo == null) (foo = 'x');
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 30,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let foo: string | undefined;
foo ??= 'x';
`},
				},
			},
		},
	},
	// if compound assign
	{
		Code: `declare let foo: string | undefined;
if (foo == null) foo += 'x';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 29,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let foo: string | undefined;
foo ??= 'x';
`},
				},
			},
		},
	},
	// if this member
	{
		Code: `class C { v?: string; m() { if (!this.v) { this.v = 'x'; } } }
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      1, Column: 29, EndLine: 1, EndColumn: 59,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `class C { v?: string; m() { this.v ??= 'x'; } }
`},
				},
			},
		},
	},
	// unicode cols
	{
		Code: `declare let x: string | null;
const s = '日本'; x || '日本';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 19, EndLine: 2, EndColumn: 21,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let x: string | null;
const s = '日本'; x ?? '日本';
`},
				},
			},
		},
	},
	// paren truthy test
	{
		Code: `declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };
((a)) ? a : b;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 14,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };
(a) ?? b;
`},
				},
			},
		},
	},
	// paren negated test
	{
		Code: `declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };
!(a) ? b : a;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 13,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };
(a) ?? b;
`},
				},
			},
		},
	},
	// paren cmp operand
	{
		Code: `declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };
((a)) !== null ? a : b;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 23,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };
(a) ?? b;
`},
				},
			},
		},
	},
	// paren non-nullish branch
	{
		Code: `declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };
a !== null ? ((a)) : b;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 23,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };
a ?? b;
`},
				},
			},
		},
	},
	// paren nullish branch
	{
		Code: `declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };
a !== null ? a : ((b));
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 23,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };
a ?? (b);
`},
				},
			},
		},
	},
	// right ==
	{
		Code: `declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };
a !== null ? a : n == 1;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 24,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };
a ?? (n == 1);
`},
				},
			},
		},
	},
	// right ===
	{
		Code: `declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };
a !== null ? a : n === 1;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 25,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };
a ?? n === 1;
`},
				},
			},
		},
	},
	// right =
	{
		Code: `declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };
let z: string;
a !== null ? a : z = b;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 23,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };
let z: string;
a ?? (z = b);
`},
				},
			},
		},
	},
	// right +=
	{
		Code: `declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };
let z = '';
a !== null ? a : z += b;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 24,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };
let z = '';
a ?? (z += b);
`},
				},
			},
		},
	},
	// right satisfies
	{
		Code: `declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };
a !== null ? a : b satisfies string;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 36,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };
a ?? (b satisfies string);
`},
				},
			},
		},
	},
	// right as
	{
		Code: `declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };
a !== null ? a : b as string;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 29,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };
a ?? b as string;
`},
				},
			},
		},
	},
	// right import()
	{
		Code: `declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };
a !== null ? a : import('x');
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 29,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };
a ?? (import('x'));
`},
				},
			},
		},
	},
	// right await
	{
		Code: `async function f(a: string | null, b: Promise<string>) {
  return a !== null ? a : await b;
}
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 10, EndLine: 2, EndColumn: 34,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `async function f(a: string | null, b: Promise<string>) {
  return a ?? await b;
}
`},
				},
			},
		},
	},
	// right typeof
	{
		Code: `declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };
a !== null ? a : typeof b;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 26,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };
a ?? typeof b;
`},
				},
			},
		},
	},
	// right instantiation
	{
		Code: `declare function g<T>(x: T): T; declare let a: typeof g | null;
a !== null ? a : g<string>;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 27,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare function g<T>(x: T): T; declare let a: typeof g | null;
a ?? (g<string>);
`},
				},
			},
		},
	},
	// right +
	{
		Code: `declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };
a !== null ? a : b + 'z';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 25,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };
a ?? b + 'z';
`},
				},
			},
		},
	},
	// right template
	{
		Code: "declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };\na !== null ? a : `${b}`;\n",
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 24,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: "declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };\na ?? `${b}`;\n"},
				},
			},
		},
	},
	// right class
	{
		Code: `declare let a: object | null;
a !== null ? a : class {};
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 26,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: object | null;
a ?? class {};
`},
				},
			},
		},
	},
	// right comma paren
	{
		Code: `declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };
a !== null ? a : (b, 'z');
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 26,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };
a ?? (b, 'z');
`},
				},
			},
		},
	},
	// ternary in Boolean
	{
		Code: `declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };
Boolean(a ? a : b);
`,
		Options: map[string]any{"ignoreBooleanCoercion": true},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 9, EndLine: 2, EndColumn: 18,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };
Boolean(a ?? b);
`},
				},
			},
		},
	},
	// ternary in Boolean paren
	{
		Code: `declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };
Boolean((a ? a : b));
`,
		Options: map[string]any{"ignoreBooleanCoercion": true},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 10, EndLine: 2, EndColumn: 19,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };
Boolean((a ?? b));
`},
				},
			},
		},
	},
	// shadowed Boolean
	{
		Code: `declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };
function f(Boolean: (x: unknown) => boolean) { return Boolean(a || b); }
`,
		Options: map[string]any{"ignoreBooleanCoercion": true},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 65, EndLine: 2, EndColumn: 67,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };
function f(Boolean: (x: unknown) => boolean) { return Boolean(a ?? b); }
`},
				},
			},
		},
	},
	// chain && left
	{
		Code: `declare let a: boolean, b: string | null, c: string, d: string;
a && b || c || d;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 8, EndLine: 2, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: boolean, b: string | null, c: string, d: string;
a && (b ?? c) || d;
`},
				},
			},
		},
	},
	// chain || left
	{
		Code: `declare let a: string | null, b: string | null, c: string, d: string;
a || b || c || d;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 3, EndLine: 2, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string, d: string;
(a ?? b) || c || d;
`},
				},
			},
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 8, EndLine: 2, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string, d: string;
a || (b ?? c) || d;
`},
				},
			},
		},
	},
	// chain ?? left
	{
		Code: `declare let a: string | null, b: string | null, c: string, d: string;
(a ?? b) || c || d;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 10, EndLine: 2, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string, d: string;
(a ?? (b) ?? c) || d;
`},
				},
			},
		},
	},
	// chain paren && left
	{
		Code: `declare let a: boolean, b: string | null, c: string, d: string;
(a && b) || c || d;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 10, EndLine: 2, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: boolean, b: string | null, c: string, d: string;
(a && (b) ?? c) || d;
`},
				},
			},
		},
	},
	// optional chains similar
	{
		Code: `declare let q: { r?: { s?: string } } | undefined;
q?.r?.s !== undefined && q.r.s !== null ? q.r.s : 'z';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 54,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let q: { r?: { s?: string } } | undefined;
q?.r?.s ?? 'z';
`},
				},
			},
		},
	},
	// computed string vs ident
	{
		Code: `declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };
o['x'] !== undefined ? o.x : b;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 31,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };
o['x'] ?? b;
`},
				},
			},
		},
	},
	// this member
	{
		Code: `class C { x?: string; m() { return this.x !== undefined ? this.x : 'y'; } }
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 36, EndLine: 1, EndColumn: 71,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `class C { x?: string; m() { return this.x ?? 'y'; } }
`},
				},
			},
		},
	},
	// if paren left assignment
	{
		Code: `declare let foo: { a?: string };
if (foo.a == null) { (foo.a) = 'x'; }
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 38,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let foo: { a?: string };
(foo.a) ??= 'x';
`},
				},
			},
		},
	},
	// if comments around paren assignment
	{
		Code: `declare let foo: string | undefined;
if (!foo) {
  /* a */ ( /* b */ foo = 'x' /* c */ ); // d
}
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 4, EndColumn: 2,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let foo: string | undefined;
/* b */
foo ??= 'x'; // d
`},
				},
			},
		},
	},
	// if trailing line comment with spaces
	{
		Code: `declare let foo: string | undefined;
if (!foo) {
  foo = 'x'; // d   
}
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 4, EndColumn: 2,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let foo: string | undefined;
foo ??= 'x'; // d   
`},
				},
			},
		},
	},
	// if comment before semicolon-less statement
	{
		Code: `declare let foo: string | undefined;
if (!foo) /* a */ /* b */ foo = 'x'
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 36,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let foo: string | undefined;
/* a */ /* b */ foo ??= 'x';
`},
				},
			},
		},
	},
	// if comments in test are dropped
	{
		Code: `declare let foo: string | undefined;
if (/* t */ !foo) {
  foo = 'x';
}
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 4, EndColumn: 2,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let foo: string | undefined;
foo ??= 'x';
`},
				},
			},
		},
	},
	// if comment after statement on own lines
	{
		Code: `declare let foo: string | undefined;
if (!foo) {
  foo = 'x';
  // one
  // two
}
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 6, EndColumn: 2,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let foo: string | undefined;
foo ??= 'x'; // one
// two
`},
				},
			},
		},
	},
}

// Cases from rslint's earlier port that are not in the v8.72.2 upstream suite.
var earlierValid = []rule_tester.ValidTestCase{
	{
		Code: `
declare let x: string;
(x || 'foo');
`,
	},
	{
		Code: `
declare let x: number;
(x || 'foo');
`,
	},
	{
		Code: `
declare let x: boolean;
(x || 'foo');
`,
	},
	{
		Code: `
declare let x: object;
(x || 'foo');
`,
	},
	{
		Code: `
declare let x: string;
(x ||= 'foo');
`,
	},
	{
		Code: `
declare let x: number;
(x ||= 'foo');
`,
	},
	{
		Code: `
declare let x: boolean;
(x ||= 'foo');
`,
	},
	{
		Code: `
declare let x: object;
(x ||= 'foo');
`,
	},
	{
		Code: `
declare let x: string | null;
x ?? 'foo';
`,
	},
	{
		Code: `
declare let x: string | null;
x ??= 'foo';
`,
	},
	{
		Code: `
declare let x: number | null;
x ?? 'foo';
`,
	},
	{
		Code: `
declare let x: number | null;
x ??= 'foo';
`,
	},
	{
		Code: `
declare let x: boolean | null;
x ?? 'foo';
`,
	},
	{
		Code: `
declare let x: boolean | null;
x ??= 'foo';
`,
	},
	{
		Code: `
declare let x: object | null;
x ?? 'foo';
`,
	},
	{
		Code: `
declare let x: object | null;
x ??= 'foo';
`,
	},
	{
		Code: `
declare let x: string | undefined;
x ?? 'foo';
`,
	},
	{
		Code: `
declare let x: string | undefined;
x ??= 'foo';
`,
	},
	{
		Code: `
declare let x: number | undefined;
x ?? 'foo';
`,
	},
	{
		Code: `
declare let x: number | undefined;
x ??= 'foo';
`,
	},
	{
		Code: `
declare let x: boolean | undefined;
x ?? 'foo';
`,
	},
	{
		Code: `
declare let x: boolean | undefined;
x ??= 'foo';
`,
	},
	{
		Code: `
declare let x: object | undefined;
x ?? 'foo';
`,
	},
	{
		Code: `
declare let x: object | undefined;
x ??= 'foo';
`,
	},
	{
		Code: `
const a = 'b';
declare let x: { a: string; b: string } | null;
x?.a != null ? x[a] : 'foo';
`,
		Options: map[string]any{"ignoreTernaryTests": false},
	},
	{
		Code: `
const a = 'b';
declare let x: { a: string; b: string } | null;
x?.[a] != null ? x.a : 'foo';
`,
		Options: map[string]any{"ignoreTernaryTests": false},
	},
	{
		Code: `
declare let x: { a: string } | null;
declare let y: { a: string } | null;
x?.a ? y?.a : 'foo';
`,
		Options: map[string]any{"ignoreTernaryTests": false},
	},
	{
		Code: `
declare let x: string | null;
x || 'foo' ? null : null;
`,
	},
	{
		Code: `
declare let x: string | null;
(x ||= 'foo') ? null : null;
`,
	},
	{
		Code: `
declare let x: number | null;
x || 'foo' ? null : null;
`,
	},
	{
		Code: `
declare let x: number | null;
(x ||= 'foo') ? null : null;
`,
	},
	{
		Code: `
declare let x: boolean | null;
x || 'foo' ? null : null;
`,
	},
	{
		Code: `
declare let x: boolean | null;
(x ||= 'foo') ? null : null;
`,
	},
	{
		Code: `
declare let x: object | null;
x || 'foo' ? null : null;
`,
	},
	{
		Code: `
declare let x: object | null;
(x ||= 'foo') ? null : null;
`,
	},
	{
		Code: `
declare let x: string | undefined;
x || 'foo' ? null : null;
`,
	},
	{
		Code: `
declare let x: string | undefined;
(x ||= 'foo') ? null : null;
`,
	},
	{
		Code: `
declare let x: number | undefined;
x || 'foo' ? null : null;
`,
	},
	{
		Code: `
declare let x: number | undefined;
(x ||= 'foo') ? null : null;
`,
	},
	{
		Code: `
declare let x: boolean | undefined;
x || 'foo' ? null : null;
`,
	},
	{
		Code: `
declare let x: boolean | undefined;
(x ||= 'foo') ? null : null;
`,
	},
	{
		Code: `
declare let x: object | undefined;
x || 'foo' ? null : null;
`,
	},
	{
		Code: `
declare let x: object | undefined;
(x ||= 'foo') ? null : null;
`,
	},
	{
		Code: `
declare let x: string | null;
if (x || 'foo') {
}
`,
	},
	{
		Code: `
declare let x: string | null;
if (x ||= 'foo') {
}
`,
	},
	{
		Code: `
declare let x: number | null;
if (x || 'foo') {
}
`,
	},
	{
		Code: `
declare let x: number | null;
if (x ||= 'foo') {
}
`,
	},
	{
		Code: `
declare let x: boolean | null;
if (x || 'foo') {
}
`,
	},
	{
		Code: `
declare let x: boolean | null;
if (x ||= 'foo') {
}
`,
	},
	{
		Code: `
declare let x: object | null;
if (x || 'foo') {
}
`,
	},
	{
		Code: `
declare let x: object | null;
if (x ||= 'foo') {
}
`,
	},
	{
		Code: `
declare let x: string | undefined;
if (x || 'foo') {
}
`,
	},
	{
		Code: `
declare let x: string | undefined;
if (x ||= 'foo') {
}
`,
	},
	{
		Code: `
declare let x: number | undefined;
if (x || 'foo') {
}
`,
	},
	{
		Code: `
declare let x: number | undefined;
if (x ||= 'foo') {
}
`,
	},
	{
		Code: `
declare let x: boolean | undefined;
if (x || 'foo') {
}
`,
	},
	{
		Code: `
declare let x: boolean | undefined;
if (x ||= 'foo') {
}
`,
	},
	{
		Code: `
declare let x: object | undefined;
if (x || 'foo') {
}
`,
	},
	{
		Code: `
declare let x: object | undefined;
if (x ||= 'foo') {
}
`,
	},
	{
		Code: `
declare let x: string | null | undefined;
if (x ||= 'foo') {
}
`,
	},
	{
		Code: `
declare let x: number | null | undefined;
if (x ||= 'foo') {
}
`,
	},
	{
		Code: `
declare let x: boolean | null | undefined;
if (x ||= 'foo') {
}
`,
	},
	{
		Code: `
declare let x: object | null | undefined;
if (x ||= 'foo') {
}
`,
	},
	{
		Code: `
declare let x: string | null;
do {} while (x || 'foo');
`,
	},
	{
		Code: `
declare let x: string | null;
do {} while (x ||= 'foo');
`,
	},
	{
		Code: `
declare let x: number | null;
do {} while (x || 'foo');
`,
	},
	{
		Code: `
declare let x: number | null;
do {} while (x ||= 'foo');
`,
	},
	{
		Code: `
declare let x: boolean | null;
do {} while (x || 'foo');
`,
	},
	{
		Code: `
declare let x: boolean | null;
do {} while (x ||= 'foo');
`,
	},
	{
		Code: `
declare let x: object | null;
do {} while (x || 'foo');
`,
	},
	{
		Code: `
declare let x: object | null;
do {} while (x ||= 'foo');
`,
	},
	{
		Code: `
declare let x: string | undefined;
do {} while (x || 'foo');
`,
	},
	{
		Code: `
declare let x: string | undefined;
do {} while (x ||= 'foo');
`,
	},
	{
		Code: `
declare let x: number | undefined;
do {} while (x || 'foo');
`,
	},
	{
		Code: `
declare let x: number | undefined;
do {} while (x ||= 'foo');
`,
	},
	{
		Code: `
declare let x: boolean | undefined;
do {} while (x || 'foo');
`,
	},
	{
		Code: `
declare let x: boolean | undefined;
do {} while (x ||= 'foo');
`,
	},
	{
		Code: `
declare let x: object | undefined;
do {} while (x || 'foo');
`,
	},
	{
		Code: `
declare let x: object | undefined;
do {} while (x ||= 'foo');
`,
	},
	{
		Code: `
declare let x: string | null | undefined;
do {} while (x ||= 'foo');
`,
	},
	{
		Code: `
declare let x: number | null | undefined;
do {} while (x ||= 'foo');
`,
	},
	{
		Code: `
declare let x: boolean | null | undefined;
do {} while (x ||= 'foo');
`,
	},
	{
		Code: `
declare let x: object | null | undefined;
do {} while (x ||= 'foo');
`,
	},
	{
		Code: `
declare let x: string | null;
for (; x || 'foo'; ) {}
`,
	},
	{
		Code: `
declare let x: string | null;
for (; x ||= 'foo'; ) {}
`,
	},
	{
		Code: `
declare let x: number | null;
for (; x || 'foo'; ) {}
`,
	},
	{
		Code: `
declare let x: number | null;
for (; x ||= 'foo'; ) {}
`,
	},
	{
		Code: `
declare let x: boolean | null;
for (; x || 'foo'; ) {}
`,
	},
	{
		Code: `
declare let x: boolean | null;
for (; x ||= 'foo'; ) {}
`,
	},
	{
		Code: `
declare let x: object | null;
for (; x || 'foo'; ) {}
`,
	},
	{
		Code: `
declare let x: object | null;
for (; x ||= 'foo'; ) {}
`,
	},
	{
		Code: `
declare let x: string | undefined;
for (; x || 'foo'; ) {}
`,
	},
	{
		Code: `
declare let x: string | undefined;
for (; x ||= 'foo'; ) {}
`,
	},
	{
		Code: `
declare let x: number | undefined;
for (; x || 'foo'; ) {}
`,
	},
	{
		Code: `
declare let x: number | undefined;
for (; x ||= 'foo'; ) {}
`,
	},
	{
		Code: `
declare let x: boolean | undefined;
for (; x || 'foo'; ) {}
`,
	},
	{
		Code: `
declare let x: boolean | undefined;
for (; x ||= 'foo'; ) {}
`,
	},
	{
		Code: `
declare let x: object | undefined;
for (; x || 'foo'; ) {}
`,
	},
	{
		Code: `
declare let x: object | undefined;
for (; x ||= 'foo'; ) {}
`,
	},
	{
		Code: `
declare let x: string | null | undefined;
for (; x || 'foo'; ) {}
`,
	},
	{
		Code: `
declare let x: string | null | undefined;
for (; x ||= 'foo'; ) {}
`,
	},
	{
		Code: `
declare let x: number | null | undefined;
for (; x || 'foo'; ) {}
`,
	},
	{
		Code: `
declare let x: number | null | undefined;
for (; x ||= 'foo'; ) {}
`,
	},
	{
		Code: `
declare let x: boolean | null | undefined;
for (; x || 'foo'; ) {}
`,
	},
	{
		Code: `
declare let x: boolean | null | undefined;
for (; x ||= 'foo'; ) {}
`,
	},
	{
		Code: `
declare let x: object | null | undefined;
for (; x || 'foo'; ) {}
`,
	},
	{
		Code: `
declare let x: object | null | undefined;
for (; x ||= 'foo'; ) {}
`,
	},
	{
		Code: `
declare let x: string | null;
while (x || 'foo') {}
`,
	},
	{
		Code: `
declare let x: string | null;
while (x ||= 'foo') {}
`,
	},
	{
		Code: `
declare let x: number | null;
while (x || 'foo') {}
`,
	},
	{
		Code: `
declare let x: number | null;
while (x ||= 'foo') {}
`,
	},
	{
		Code: `
declare let x: boolean | null;
while (x || 'foo') {}
`,
	},
	{
		Code: `
declare let x: boolean | null;
while (x ||= 'foo') {}
`,
	},
	{
		Code: `
declare let x: object | null;
while (x || 'foo') {}
`,
	},
	{
		Code: `
declare let x: object | null;
while (x ||= 'foo') {}
`,
	},
	{
		Code: `
declare let x: string | undefined;
while (x || 'foo') {}
`,
	},
	{
		Code: `
declare let x: string | undefined;
while (x ||= 'foo') {}
`,
	},
	{
		Code: `
declare let x: number | undefined;
while (x || 'foo') {}
`,
	},
	{
		Code: `
declare let x: number | undefined;
while (x ||= 'foo') {}
`,
	},
	{
		Code: `
declare let x: boolean | undefined;
while (x || 'foo') {}
`,
	},
	{
		Code: `
declare let x: boolean | undefined;
while (x ||= 'foo') {}
`,
	},
	{
		Code: `
declare let x: object | undefined;
while (x || 'foo') {}
`,
	},
	{
		Code: `
declare let x: object | undefined;
while (x ||= 'foo') {}
`,
	},
	{
		Code: `
declare let x: string | null | undefined;
while (x ||= 'foo') {}
`,
	},
	{
		Code: `
declare let x: number | null | undefined;
while (x ||= 'foo') {}
`,
	},
	{
		Code: `
declare let x: boolean | null | undefined;
while (x ||= 'foo') {}
`,
	},
	{
		Code: `
declare let x: object | null | undefined;
while (x ||= 'foo') {}
`,
	},
	{
		Code: `
let a: string | undefined;
let b: string | undefined;
if (!(a || b)) {}
`,
		Options: map[string]any{"ignoreConditionalTests": true},
	},
	{
		Code: `
let a: string | undefined;
let b: string | undefined;
if (!!(a || b)) {}
`,
		Options: map[string]any{"ignoreConditionalTests": true},
	},
	{
		Code: `
let a: string | true | undefined;
let b: string | boolean | undefined;
if (a ? a : b) {}
`,
		Options: map[string]any{"ignoreConditionalTests": true},
	},
	{
		Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
if (!a ? b : a) {}
`,
		Options: map[string]any{"ignoreConditionalTests": true},
	},
	{
		Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;
if ((a ? a : b) || c) {}
`,
		Options: map[string]any{"ignoreConditionalTests": true},
	},
	{
		Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;
if (c || (!a ? b : a)) {}
`,
		Options: map[string]any{"ignoreConditionalTests": true},
	},
	{
		Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;
if ((a || b) ?? c) {}
`,
		Options: map[string]any{"ignoreConditionalTests": true},
	},
	{
		Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;
if (a ?? (b || c)) {}
`,
		Options: map[string]any{"ignoreConditionalTests": true},
	},
	{
		Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;
if (a ? b || c : 'fail') {}
`,
		Options: map[string]any{"ignoreConditionalTests": true},
	},
	{
		Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;
if (a ? 'success' : b || c) {}
`,
		Options: map[string]any{"ignoreConditionalTests": true},
	},
	{
		Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;
if (((a = b), b || c)) {}
`,
		Options: map[string]any{"ignoreConditionalTests": true},
	},
	{
		Code: `
declare let a: string | null;
declare let b: string | null;
declare let c: string | null;
a || (b && c);
`,
		Options: map[string]any{"ignoreMixedLogicalExpressions": true},
	},
	{
		Code: `
declare let a: string | null;
declare let b: string | null;
declare let c: string | null;
a ||= (b && c);
`,
		Options: map[string]any{"ignoreMixedLogicalExpressions": true},
	},
	{
		Code: `
declare let a: number | null;
declare let b: number | null;
declare let c: number | null;
a || (b && c);
`,
		Options: map[string]any{"ignoreMixedLogicalExpressions": true},
	},
	{
		Code: `
declare let a: number | null;
declare let b: number | null;
declare let c: number | null;
a ||= (b && c);
`,
		Options: map[string]any{"ignoreMixedLogicalExpressions": true},
	},
	{
		Code: `
declare let a: boolean | null;
declare let b: boolean | null;
declare let c: boolean | null;
a || (b && c);
`,
		Options: map[string]any{"ignoreMixedLogicalExpressions": true},
	},
	{
		Code: `
declare let a: boolean | null;
declare let b: boolean | null;
declare let c: boolean | null;
a ||= (b && c);
`,
		Options: map[string]any{"ignoreMixedLogicalExpressions": true},
	},
	{
		Code: `
declare let a: object | null;
declare let b: object | null;
declare let c: object | null;
a || (b && c);
`,
		Options: map[string]any{"ignoreMixedLogicalExpressions": true},
	},
	{
		Code: `
declare let a: object | null;
declare let b: object | null;
declare let c: object | null;
a ||= (b && c);
`,
		Options: map[string]any{"ignoreMixedLogicalExpressions": true},
	},
	{
		Code: `
declare let a: string | undefined;
declare let b: string | undefined;
declare let c: string | undefined;
a || (b && c);
`,
		Options: map[string]any{"ignoreMixedLogicalExpressions": true},
	},
	{
		Code: `
declare let a: string | undefined;
declare let b: string | undefined;
declare let c: string | undefined;
a ||= (b && c);
`,
		Options: map[string]any{"ignoreMixedLogicalExpressions": true},
	},
	{
		Code: `
declare let a: number | undefined;
declare let b: number | undefined;
declare let c: number | undefined;
a || (b && c);
`,
		Options: map[string]any{"ignoreMixedLogicalExpressions": true},
	},
	{
		Code: `
declare let a: number | undefined;
declare let b: number | undefined;
declare let c: number | undefined;
a ||= (b && c);
`,
		Options: map[string]any{"ignoreMixedLogicalExpressions": true},
	},
	{
		Code: `
declare let a: boolean | undefined;
declare let b: boolean | undefined;
declare let c: boolean | undefined;
a || (b && c);
`,
		Options: map[string]any{"ignoreMixedLogicalExpressions": true},
	},
	{
		Code: `
declare let a: boolean | undefined;
declare let b: boolean | undefined;
declare let c: boolean | undefined;
a ||= (b && c);
`,
		Options: map[string]any{"ignoreMixedLogicalExpressions": true},
	},
	{
		Code: `
declare let a: object | undefined;
declare let b: object | undefined;
declare let c: object | undefined;
a || (b && c);
`,
		Options: map[string]any{"ignoreMixedLogicalExpressions": true},
	},
	{
		Code: `
declare let a: object | undefined;
declare let b: object | undefined;
declare let c: object | undefined;
a ||= (b && c);
`,
		Options: map[string]any{"ignoreMixedLogicalExpressions": true},
	},
	{
		Code: `
declare let a: string | null | undefined;
declare let b: string | null | undefined;
declare let c: string | null | undefined;
a ||= (b && c);
`,
		Options: map[string]any{"ignoreMixedLogicalExpressions": true},
	},
	{
		Code: `
declare let a: number | null | undefined;
declare let b: number | null | undefined;
declare let c: number | null | undefined;
a ||= (b && c);
`,
		Options: map[string]any{"ignoreMixedLogicalExpressions": true},
	},
	{
		Code: `
declare let a: boolean | null | undefined;
declare let b: boolean | null | undefined;
declare let c: boolean | null | undefined;
a ||= (b && c);
`,
		Options: map[string]any{"ignoreMixedLogicalExpressions": true},
	},
	{
		Code: `
declare let a: object | null | undefined;
declare let b: object | null | undefined;
declare let c: object | null | undefined;
a ||= (b && c);
`,
		Options: map[string]any{"ignoreMixedLogicalExpressions": true},
	},
	{
		Code: `
declare let a: string | null | undefined;
declare let b: string | null | undefined;
declare let c: string | null | undefined;
declare let d: string | null | undefined;
a || (b && c) || d;
`,
		Options: map[string]any{"ignoreMixedLogicalExpressions": true},
	},
	{
		Code: `
declare let a: string | null | undefined;
declare let b: string | null | undefined;
declare let c: string | null | undefined;
declare let d: string | null | undefined;
a || (b && c && d);
`,
		Options: map[string]any{"ignoreMixedLogicalExpressions": true},
	},
	{
		Code: `
declare let x: number | null | undefined;
x || 1;
`,
		Options: map[string]any{"ignorePrimitives": map[string]any{"number": true}},
	},
	{
		Code: `
declare let x: bigint | null | undefined;
x || 1n;
`,
		Options: map[string]any{"ignorePrimitives": map[string]any{"bigint": true}},
	},
	{
		Code: `
declare let x: boolean | null | undefined;
x || true;
`,
		Options: map[string]any{"ignorePrimitives": map[string]any{"boolean": true}},
	},
	{
		Code: `
let a: string | true | undefined;
let b: string | boolean | undefined;
const x = Boolean(a || b);
`,
		Options: map[string]any{"ignoreBooleanCoercion": true},
	},
	{
		Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;
const test = Boolean(a || b || c);
`,
		Options: map[string]any{"ignoreBooleanCoercion": true},
	},
	{
		Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;
const test = Boolean(a || (b && c));
`,
		Options: map[string]any{"ignoreBooleanCoercion": true},
	},
	{
		Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;
const test = Boolean((a || b) ?? c);
`,
		Options: map[string]any{"ignoreBooleanCoercion": true},
	},
	{
		Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;
const test = Boolean(a ?? (b || c));
`,
		Options: map[string]any{"ignoreBooleanCoercion": true},
	},
	{
		Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;
const test = Boolean(a ? b || c : 'fail');
`,
		Options: map[string]any{"ignoreBooleanCoercion": true},
	},
	{
		Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;
const test = Boolean(a ? 'success' : b || c);
`,
		Options: map[string]any{"ignoreBooleanCoercion": true},
	},
	{
		Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;
const test = Boolean(((a = b), b || c));
`,
		Options: map[string]any{"ignoreBooleanCoercion": true},
	},
	{
		Code: `
function getName(user?: { name?: string }) {
  return user?.name ?? 'anonymous';
}
`,
	},
	{
		Code: `
type Config = { port?: number; host?: string };
declare const cfg: Config;
const port = cfg.port ?? 3000;
const host = cfg.host ?? 'localhost';
`,
	},
	{
		Code: `
function paginate(limit: number, offset: number = 0) {
  return [limit, offset];
}
`,
	},
	{
		Code: `
const role: 'admin' | 'user' | 'guest' = 'guest';
const permissions = role || 'guest';
`,
	},
	{
		Code: `
const cache: Map<string, string | null> = new Map();
declare let key: string;
let cached: string | null | undefined;
cached = cache.get(key) ?? null;
`,
	},
	{
		Code: `
declare let bigCount: bigint | null;
const total = bigCount ?? 0n;
`,
	},
	{
		Code: `
function search(q: string) { return q.length; }
function caller(query: string) {
  return search(query || '');
}
`,
	},
	{
		Code: `
const items: string[] = [];
const first = items.find(x => x.length > 0) ?? 'none';
`,
	},
	{
		Code: `
type Props = { value?: string };
declare const props: Props;
declare function Foo(p: any): any;
const el = <Foo value={props.value ?? 'default'} />;
`,
		Tsx: true,
	},
	{
		Code: `
declare function fetchUser(): Promise<{ name: string } | null>;
async function main() {
  const u = (await fetchUser()) ?? { name: 'anonymous' };
  return u.name;
}
`,
	},
	{
		Code: `
function* names(): Generator<string> {
  let names: string[] = ['a', 'b'];
  for (const n of names) yield n || 'default';
}
`,
	},
	{
		Code: `
class Counter {
  #value: number = 0;
  inc() { this.#value = (this.#value || 0) + 1; }
}
`,
	},
	{
		Code: "\nfunction tag(s: TemplateStringsArray, ...args: unknown[]): string { return s.join(''); }\nconst x = tag`hello ${1 || 2}`;\n",
	},
	{
		Code: `
declare let x: string | null;
(x as any) !== undefined ? (x as any) : y;
`,
		Options: map[string]any{"ignoreTernaryTests": false},
	},
	{
		Code: `
type T = (string | null) & {};
declare let x: T;
x ?? 'foo';
`,
	},
}

var earlierInvalid = []rule_tester.InvalidTestCase{
	{
		Code: `
declare let x: string | null;
x || 'foo';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | null;
x ?? 'foo';
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: string | null;
x ||= 'foo';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 6,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | null;
x ??= 'foo';
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: number | null;
x || 'foo';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | null;
x ?? 'foo';
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: number | null;
x ||= 'foo';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 6,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | null;
x ??= 'foo';
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: boolean | null;
x || 'foo';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | null;
x ?? 'foo';
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: boolean | null;
x ||= 'foo';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 6,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | null;
x ??= 'foo';
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: object | null;
x || 'foo';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | null;
x ?? 'foo';
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: object | null;
x ||= 'foo';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 6,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | null;
x ??= 'foo';
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: string | undefined;
x || 'foo';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | undefined;
x ?? 'foo';
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: string | undefined;
x ||= 'foo';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 6,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | undefined;
x ??= 'foo';
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: number | undefined;
x || 'foo';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | undefined;
x ?? 'foo';
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: number | undefined;
x ||= 'foo';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 6,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | undefined;
x ??= 'foo';
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: boolean | undefined;
x || 'foo';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | undefined;
x ?? 'foo';
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: boolean | undefined;
x ||= 'foo';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 6,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | undefined;
x ??= 'foo';
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: object | undefined;
x || 'foo';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | undefined;
x ?? 'foo';
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: object | undefined;
x ||= 'foo';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 6,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | undefined;
x ??= 'foo';
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: object | null | undefined;
x ||= {};
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 6,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | null | undefined;
x ??= {};
`},
				},
			},
		},
	},
	{
		Code:    `x != null && x != undefined ? x : y;`,
		Options: map[string]any{"ignoreTernaryTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 36,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `x ?? y;`},
				},
			},
		},
	},
	{
		Code: `
declare let b: string | undefined;
b !== undefined ? b : 'a string';
`,
		Options: map[string]any{"ignoreTernaryTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 33,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let b: string | undefined;
b ?? 'a string';
`},
				},
			},
		},
	},
	{
		Code: `
declare let b: string | undefined;
b === undefined ? 'a string' : b;
`,
		Options: map[string]any{"ignoreTernaryTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 33,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let b: string | undefined;
b ?? 'a string';
`},
				},
			},
		},
	},
	{
		Code: `
declare let c: string | null;
c !== null ? c : 'a string';
`,
		Options: map[string]any{"ignoreTernaryTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 28,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let c: string | null;
c ?? 'a string';
`},
				},
			},
		},
	},
	{
		Code: `
declare let c: string | null;
c === null ? 'a string' : c;
`,
		Options: map[string]any{"ignoreTernaryTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 28,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let c: string | null;
c ?? 'a string';
`},
				},
			},
		},
	},
	{
		Code: `
declare let b: string | undefined;
b ? b : 'a string';
`,
		Options: map[string]any{"ignoreTernaryTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 19,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let b: string | undefined;
b ?? 'a string';
`},
				},
			},
		},
	},
	{
		Code: `
declare let c: string | null;
!c ? 'a string' : c;
`,
		Options: map[string]any{"ignoreTernaryTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 20,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let c: string | null;
c ?? 'a string';
`},
				},
			},
		},
	},
	{
		Code: `
declare let a: any;
a ? a : 'a string';
`,
		Options: map[string]any{"ignoreTernaryTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 19,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let a: any;
a ?? 'a string';
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: { n?: { a?: string } };
x.n?.a !== undefined ? x.n?.a : y;
`,
		Options: map[string]any{"ignoreTernaryTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 34,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { n?: { a?: string } };
x.n?.a ?? y;
`},
				},
			},
		},
	},
	{
		Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitializeFoo1() {
  if (!foo) {
    foo = makeFoo();
  }
}
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      6, Column: 3, EndLine: 8, EndColumn: 4,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitializeFoo1() {
  foo ??= makeFoo();
}
`},
				},
			},
		},
	},
	{
		Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitializeFoo2() {
  if (!foo) foo = makeFoo();
}
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      6, Column: 3, EndLine: 6, EndColumn: 29,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitializeFoo2() {
  foo ??= makeFoo();
}
`},
				},
			},
		},
	},
	{
		Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) foo = makeFoo();
}
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      6, Column: 3, EndLine: 6, EndColumn: 36,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: string | null;
x || 'foo' ? null : null;
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | null;
x ?? 'foo' ? null : null;
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: string | null;
(x ||= 'foo') ? null : null;
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 4, EndLine: 3, EndColumn: 7,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | null;
(x ??= 'foo') ? null : null;
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: number | null;
x || 'foo' ? null : null;
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | null;
x ?? 'foo' ? null : null;
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: number | null;
(x ||= 'foo') ? null : null;
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 4, EndLine: 3, EndColumn: 7,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | null;
(x ??= 'foo') ? null : null;
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: boolean | null;
x || 'foo' ? null : null;
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | null;
x ?? 'foo' ? null : null;
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: boolean | null;
(x ||= 'foo') ? null : null;
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 4, EndLine: 3, EndColumn: 7,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | null;
(x ??= 'foo') ? null : null;
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: object | null;
x || 'foo' ? null : null;
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | null;
x ?? 'foo' ? null : null;
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: object | null;
(x ||= 'foo') ? null : null;
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 4, EndLine: 3, EndColumn: 7,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | null;
(x ??= 'foo') ? null : null;
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: string | undefined;
x || 'foo' ? null : null;
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | undefined;
x ?? 'foo' ? null : null;
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: string | undefined;
(x ||= 'foo') ? null : null;
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 4, EndLine: 3, EndColumn: 7,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | undefined;
(x ??= 'foo') ? null : null;
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: number | undefined;
x || 'foo' ? null : null;
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | undefined;
x ?? 'foo' ? null : null;
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: number | undefined;
(x ||= 'foo') ? null : null;
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 4, EndLine: 3, EndColumn: 7,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | undefined;
(x ??= 'foo') ? null : null;
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: boolean | undefined;
x || 'foo' ? null : null;
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | undefined;
x ?? 'foo' ? null : null;
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: boolean | undefined;
(x ||= 'foo') ? null : null;
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 4, EndLine: 3, EndColumn: 7,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | undefined;
(x ??= 'foo') ? null : null;
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: object | undefined;
x || 'foo' ? null : null;
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | undefined;
x ?? 'foo' ? null : null;
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: object | undefined;
(x ||= 'foo') ? null : null;
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 4, EndLine: 3, EndColumn: 7,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | undefined;
(x ??= 'foo') ? null : null;
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: string | null;
if (x || 'foo') {
}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 7, EndLine: 3, EndColumn: 9,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | null;
if (x ?? 'foo') {
}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: string | null;
if (x ||= 'foo') {
}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 7, EndLine: 3, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | null;
if (x ??= 'foo') {
}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: number | null;
if (x || 'foo') {
}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 7, EndLine: 3, EndColumn: 9,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | null;
if (x ?? 'foo') {
}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: number | null;
if (x ||= 'foo') {
}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 7, EndLine: 3, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | null;
if (x ??= 'foo') {
}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: boolean | null;
if (x || 'foo') {
}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 7, EndLine: 3, EndColumn: 9,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | null;
if (x ?? 'foo') {
}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: boolean | null;
if (x ||= 'foo') {
}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 7, EndLine: 3, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | null;
if (x ??= 'foo') {
}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: object | null;
if (x || 'foo') {
}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 7, EndLine: 3, EndColumn: 9,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | null;
if (x ?? 'foo') {
}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: object | null;
if (x ||= 'foo') {
}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 7, EndLine: 3, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | null;
if (x ??= 'foo') {
}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: string | undefined;
if (x || 'foo') {
}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 7, EndLine: 3, EndColumn: 9,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | undefined;
if (x ?? 'foo') {
}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: string | undefined;
if (x ||= 'foo') {
}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 7, EndLine: 3, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | undefined;
if (x ??= 'foo') {
}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: number | undefined;
if (x || 'foo') {
}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 7, EndLine: 3, EndColumn: 9,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | undefined;
if (x ?? 'foo') {
}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: number | undefined;
if (x ||= 'foo') {
}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 7, EndLine: 3, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | undefined;
if (x ??= 'foo') {
}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: boolean | undefined;
if (x || 'foo') {
}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 7, EndLine: 3, EndColumn: 9,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | undefined;
if (x ?? 'foo') {
}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: boolean | undefined;
if (x ||= 'foo') {
}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 7, EndLine: 3, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | undefined;
if (x ??= 'foo') {
}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: object | undefined;
if (x || 'foo') {
}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 7, EndLine: 3, EndColumn: 9,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | undefined;
if (x ?? 'foo') {
}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: object | undefined;
if (x ||= 'foo') {
}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 7, EndLine: 3, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | undefined;
if (x ??= 'foo') {
}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: string | null | undefined;
if (x ||= 'foo') {
}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 7, EndLine: 3, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | null | undefined;
if (x ??= 'foo') {
}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: number | null | undefined;
if (x ||= 'foo') {
}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 7, EndLine: 3, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | null | undefined;
if (x ??= 'foo') {
}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: boolean | null | undefined;
if (x ||= 'foo') {
}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 7, EndLine: 3, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | null | undefined;
if (x ??= 'foo') {
}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: object | null | undefined;
if (x ||= 'foo') {
}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 7, EndLine: 3, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | null | undefined;
if (x ??= 'foo') {
}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: string | null;
do {} while (x || 'foo');
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 16, EndLine: 3, EndColumn: 18,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | null;
do {} while (x ?? 'foo');
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: string | null;
do {} while (x ||= 'foo');
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 16, EndLine: 3, EndColumn: 19,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | null;
do {} while (x ??= 'foo');
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: number | null;
do {} while (x || 'foo');
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 16, EndLine: 3, EndColumn: 18,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | null;
do {} while (x ?? 'foo');
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: number | null;
do {} while (x ||= 'foo');
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 16, EndLine: 3, EndColumn: 19,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | null;
do {} while (x ??= 'foo');
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: boolean | null;
do {} while (x || 'foo');
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 16, EndLine: 3, EndColumn: 18,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | null;
do {} while (x ?? 'foo');
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: boolean | null;
do {} while (x ||= 'foo');
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 16, EndLine: 3, EndColumn: 19,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | null;
do {} while (x ??= 'foo');
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: object | null;
do {} while (x || 'foo');
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 16, EndLine: 3, EndColumn: 18,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | null;
do {} while (x ?? 'foo');
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: object | null;
do {} while (x ||= 'foo');
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 16, EndLine: 3, EndColumn: 19,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | null;
do {} while (x ??= 'foo');
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: string | undefined;
do {} while (x || 'foo');
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 16, EndLine: 3, EndColumn: 18,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | undefined;
do {} while (x ?? 'foo');
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: string | undefined;
do {} while (x ||= 'foo');
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 16, EndLine: 3, EndColumn: 19,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | undefined;
do {} while (x ??= 'foo');
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: number | undefined;
do {} while (x || 'foo');
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 16, EndLine: 3, EndColumn: 18,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | undefined;
do {} while (x ?? 'foo');
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: number | undefined;
do {} while (x ||= 'foo');
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 16, EndLine: 3, EndColumn: 19,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | undefined;
do {} while (x ??= 'foo');
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: boolean | undefined;
do {} while (x || 'foo');
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 16, EndLine: 3, EndColumn: 18,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | undefined;
do {} while (x ?? 'foo');
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: boolean | undefined;
do {} while (x ||= 'foo');
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 16, EndLine: 3, EndColumn: 19,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | undefined;
do {} while (x ??= 'foo');
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: object | undefined;
do {} while (x || 'foo');
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 16, EndLine: 3, EndColumn: 18,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | undefined;
do {} while (x ?? 'foo');
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: object | undefined;
do {} while (x ||= 'foo');
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 16, EndLine: 3, EndColumn: 19,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | undefined;
do {} while (x ??= 'foo');
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: string | null | undefined;
do {} while (x ||= 'foo');
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 16, EndLine: 3, EndColumn: 19,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | null | undefined;
do {} while (x ??= 'foo');
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: number | null | undefined;
do {} while (x ||= 'foo');
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 16, EndLine: 3, EndColumn: 19,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | null | undefined;
do {} while (x ??= 'foo');
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: boolean | null | undefined;
do {} while (x ||= 'foo');
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 16, EndLine: 3, EndColumn: 19,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | null | undefined;
do {} while (x ??= 'foo');
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: object | null | undefined;
do {} while (x ||= 'foo');
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 16, EndLine: 3, EndColumn: 19,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | null | undefined;
do {} while (x ??= 'foo');
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: string | null;
for (; x || 'foo'; ) {}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | null;
for (; x ?? 'foo'; ) {}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: string | null;
for (; x ||= 'foo'; ) {}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 13,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | null;
for (; x ??= 'foo'; ) {}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: number | null;
for (; x || 'foo'; ) {}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | null;
for (; x ?? 'foo'; ) {}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: number | null;
for (; x ||= 'foo'; ) {}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 13,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | null;
for (; x ??= 'foo'; ) {}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: boolean | null;
for (; x || 'foo'; ) {}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | null;
for (; x ?? 'foo'; ) {}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: boolean | null;
for (; x ||= 'foo'; ) {}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 13,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | null;
for (; x ??= 'foo'; ) {}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: object | null;
for (; x || 'foo'; ) {}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | null;
for (; x ?? 'foo'; ) {}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: object | null;
for (; x ||= 'foo'; ) {}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 13,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | null;
for (; x ??= 'foo'; ) {}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: string | undefined;
for (; x || 'foo'; ) {}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | undefined;
for (; x ?? 'foo'; ) {}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: string | undefined;
for (; x ||= 'foo'; ) {}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 13,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | undefined;
for (; x ??= 'foo'; ) {}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: number | undefined;
for (; x || 'foo'; ) {}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | undefined;
for (; x ?? 'foo'; ) {}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: number | undefined;
for (; x ||= 'foo'; ) {}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 13,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | undefined;
for (; x ??= 'foo'; ) {}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: boolean | undefined;
for (; x || 'foo'; ) {}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | undefined;
for (; x ?? 'foo'; ) {}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: boolean | undefined;
for (; x ||= 'foo'; ) {}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 13,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | undefined;
for (; x ??= 'foo'; ) {}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: object | undefined;
for (; x || 'foo'; ) {}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | undefined;
for (; x ?? 'foo'; ) {}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: object | undefined;
for (; x ||= 'foo'; ) {}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 13,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | undefined;
for (; x ??= 'foo'; ) {}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: string | null | undefined;
for (; x || 'foo'; ) {}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | null | undefined;
for (; x ?? 'foo'; ) {}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: string | null | undefined;
for (; x ||= 'foo'; ) {}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 13,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | null | undefined;
for (; x ??= 'foo'; ) {}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: number | null | undefined;
for (; x || 'foo'; ) {}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | null | undefined;
for (; x ?? 'foo'; ) {}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: number | null | undefined;
for (; x ||= 'foo'; ) {}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 13,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | null | undefined;
for (; x ??= 'foo'; ) {}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: boolean | null | undefined;
for (; x || 'foo'; ) {}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | null | undefined;
for (; x ?? 'foo'; ) {}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: boolean | null | undefined;
for (; x ||= 'foo'; ) {}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 13,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | null | undefined;
for (; x ??= 'foo'; ) {}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: object | null | undefined;
for (; x || 'foo'; ) {}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | null | undefined;
for (; x ?? 'foo'; ) {}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: object | null | undefined;
for (; x ||= 'foo'; ) {}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 13,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | null | undefined;
for (; x ??= 'foo'; ) {}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: string | null;
while (x || 'foo') {}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | null;
while (x ?? 'foo') {}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: string | null;
while (x ||= 'foo') {}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 13,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | null;
while (x ??= 'foo') {}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: number | null;
while (x || 'foo') {}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | null;
while (x ?? 'foo') {}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: number | null;
while (x ||= 'foo') {}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 13,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | null;
while (x ??= 'foo') {}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: boolean | null;
while (x || 'foo') {}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | null;
while (x ?? 'foo') {}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: boolean | null;
while (x ||= 'foo') {}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 13,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | null;
while (x ??= 'foo') {}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: object | null;
while (x || 'foo') {}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | null;
while (x ?? 'foo') {}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: object | null;
while (x ||= 'foo') {}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 13,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | null;
while (x ??= 'foo') {}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: string | undefined;
while (x || 'foo') {}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | undefined;
while (x ?? 'foo') {}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: string | undefined;
while (x ||= 'foo') {}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 13,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | undefined;
while (x ??= 'foo') {}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: number | undefined;
while (x || 'foo') {}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | undefined;
while (x ?? 'foo') {}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: number | undefined;
while (x ||= 'foo') {}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 13,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | undefined;
while (x ??= 'foo') {}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: boolean | undefined;
while (x || 'foo') {}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | undefined;
while (x ?? 'foo') {}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: boolean | undefined;
while (x ||= 'foo') {}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 13,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | undefined;
while (x ??= 'foo') {}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: object | undefined;
while (x || 'foo') {}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | undefined;
while (x ?? 'foo') {}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: object | undefined;
while (x ||= 'foo') {}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 13,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | undefined;
while (x ??= 'foo') {}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: string | null | undefined;
while (x ||= 'foo') {}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 13,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | null | undefined;
while (x ??= 'foo') {}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: number | null | undefined;
while (x ||= 'foo') {}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 13,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | null | undefined;
while (x ??= 'foo') {}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: boolean | null | undefined;
while (x ||= 'foo') {}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 13,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | null | undefined;
while (x ??= 'foo') {}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: object | null | undefined;
while (x ||= 'foo') {}
`,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 13,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | null | undefined;
while (x ??= 'foo') {}
`},
				},
			},
		},
	},
	{
		Code: `
declare const a: string | true | undefined;
declare const b: string | boolean | undefined;
const x = Boolean(a || b);
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      4, Column: 21, EndLine: 4, EndColumn: 23,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare const a: string | true | undefined;
declare const b: string | boolean | undefined;
const x = Boolean(a ?? b);
`},
				},
			},
		},
	},
	{
		Code: `
let a: string | true | undefined;
let b: string | boolean | undefined;
const x = Boolean(a ? a : b);
`,
		Options: map[string]any{"ignoreBooleanCoercion": true},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      4, Column: 19, EndLine: 4, EndColumn: 28,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
let a: string | true | undefined;
let b: string | boolean | undefined;
const x = Boolean(a ?? b);
`},
				},
			},
		},
	},
	{
		Code: `
let a: string | true | undefined;
let b: string | boolean | undefined;
const x = Boolean((a ? a : b));
`,
		Options: map[string]any{"ignoreBooleanCoercion": true},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      4, Column: 20, EndLine: 4, EndColumn: 29,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
let a: string | true | undefined;
let b: string | boolean | undefined;
const x = Boolean((a ?? b));
`},
				},
			},
		},
	},
	{
		Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
const test = Boolean(!a ? b : a);
`,
		Options: map[string]any{"ignoreBooleanCoercion": true},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      4, Column: 22, EndLine: 4, EndColumn: 32,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
const test = Boolean(a ?? b);
`},
				},
			},
		},
	},
	{
		Code: `
declare let a: string | null | undefined;
declare let b: string | null | undefined;
declare let c: string;

(a || b) || c;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      6, Column: 4, EndLine: 6, EndColumn: 6,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let a: string | null | undefined;
declare let b: string | null | undefined;
declare let c: string;

((a ?? b)) || c;
`},
				},
			},
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      6, Column: 10, EndLine: 6, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let a: string | null | undefined;
declare let b: string | null | undefined;
declare let c: string;

(a || b) ?? c;
`},
				},
			},
		},
	},
	{
		Code: `
declare let a: string | null | undefined;
declare let b: string;
declare let c: string;

a || b && c;
`,
		Options: map[string]any{"ignoreMixedLogicalExpressions": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      6, Column: 3, EndLine: 6, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let a: string | null | undefined;
declare let b: string;
declare let c: string;

a ?? b && c;
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: 0 | undefined;
!x ? y : x;
`,
		Options: map[string]any{"ignorePrimitives": map[string]any{"bigint": true, "boolean": true, "number": false, "string": true}},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 11,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: 0 | undefined;
x ?? y;
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: string | null;
if (() => x || 'foo') {
}
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 13, EndLine: 3, EndColumn: 15,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | null;
if (() => x ?? 'foo') {
}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: string | null;
if (() => (x ||= 'foo')) {
}
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 14, EndLine: 3, EndColumn: 17,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | null;
if (() => (x ??= 'foo')) {
}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: number | null;
if (() => x || 'foo') {
}
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 13, EndLine: 3, EndColumn: 15,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | null;
if (() => x ?? 'foo') {
}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: number | null;
if (() => (x ||= 'foo')) {
}
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 14, EndLine: 3, EndColumn: 17,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | null;
if (() => (x ??= 'foo')) {
}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: boolean | null;
if (() => x || 'foo') {
}
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 13, EndLine: 3, EndColumn: 15,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | null;
if (() => x ?? 'foo') {
}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: boolean | null;
if (() => (x ||= 'foo')) {
}
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 14, EndLine: 3, EndColumn: 17,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | null;
if (() => (x ??= 'foo')) {
}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: object | null;
if (() => x || 'foo') {
}
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 13, EndLine: 3, EndColumn: 15,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | null;
if (() => x ?? 'foo') {
}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: object | null;
if (() => (x ||= 'foo')) {
}
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 14, EndLine: 3, EndColumn: 17,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | null;
if (() => (x ??= 'foo')) {
}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: string | undefined;
if (() => x || 'foo') {
}
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 13, EndLine: 3, EndColumn: 15,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | undefined;
if (() => x ?? 'foo') {
}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: string | undefined;
if (() => (x ||= 'foo')) {
}
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 14, EndLine: 3, EndColumn: 17,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | undefined;
if (() => (x ??= 'foo')) {
}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: number | undefined;
if (() => x || 'foo') {
}
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 13, EndLine: 3, EndColumn: 15,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | undefined;
if (() => x ?? 'foo') {
}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: number | undefined;
if (() => (x ||= 'foo')) {
}
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 14, EndLine: 3, EndColumn: 17,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: number | undefined;
if (() => (x ??= 'foo')) {
}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: boolean | undefined;
if (() => x || 'foo') {
}
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 13, EndLine: 3, EndColumn: 15,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | undefined;
if (() => x ?? 'foo') {
}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: boolean | undefined;
if (() => (x ||= 'foo')) {
}
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 14, EndLine: 3, EndColumn: 17,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: boolean | undefined;
if (() => (x ??= 'foo')) {
}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: object | undefined;
if (() => x || 'foo') {
}
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 13, EndLine: 3, EndColumn: 15,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | undefined;
if (() => x ?? 'foo') {
}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: object | undefined;
if (() => (x ||= 'foo')) {
}
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      3, Column: 14, EndLine: 3, EndColumn: 17,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: object | undefined;
if (() => (x ??= 'foo')) {
}
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: { foo: string } | null;
x ? x : { foo: '' };
`,
		Options: map[string]any{"ignoreTernaryTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 20,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { foo: string } | null;
x ?? { foo: '' };
`},
				},
			},
		},
	},
	{
		Code: `
class Base {
  declare foo: string | null;
}
class Derived extends Base {
  use() {
    return super.foo || 'foo';
  }
}
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      7, Column: 22, EndLine: 7, EndColumn: 24,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
class Base {
  declare foo: string | null;
}
class Derived extends Base {
  use() {
    return super.foo ?? 'foo';
  }
}
`},
				},
			},
		},
	},
	{
		Code: `
class C {
  declare foo: string | null;
  get value() {
    return this.foo || 'foo';
  }
}
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      5, Column: 21, EndLine: 5, EndColumn: 23,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
class C {
  declare foo: string | null;
  get value() {
    return this.foo ?? 'foo';
  }
}
`},
				},
			},
		},
	},
	{
		Code: `
declare const arr: (string | null)[];
arr[0] || 'foo';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 8, EndLine: 3, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare const arr: (string | null)[];
arr[0] ?? 'foo';
`},
				},
			},
		},
	},
	{
		Code: `
declare const arr: ((string | null)[] | null);
arr?.[0] || 'foo';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare const arr: ((string | null)[] | null);
arr?.[0] ?? 'foo';
`},
				},
			},
		},
	},
	{
		Code: `
declare function getFoo(): string | null;
getFoo() || 'foo';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 10, EndLine: 3, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare function getFoo(): string | null;
getFoo() ?? 'foo';
`},
				},
			},
		},
	},
	{
		Code: `
declare let a: string | null;
declare let b: string | null;
declare let c: string;
a ?? b || c;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      5, Column: 8, EndLine: 5, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let a: string | null;
declare let b: string | null;
declare let c: string;
a ?? b ?? c;
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: string | null;
(x as string | null) || 'foo';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 22, EndLine: 3, EndColumn: 24,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: string | null;
(x as string | null) ?? 'foo';
`},
				},
			},
		},
	},
	{
		Code: `
class C {
  declare foo: string | null | undefined;
  use() {
    return this.foo || 'foo';
  }
}
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      5, Column: 21, EndLine: 5, EndColumn: 23,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
class C {
  declare foo: string | null | undefined;
  use() {
    return this.foo ?? 'foo';
  }
}
`},
				},
			},
		},
	},
	{
		Code: `
declare let foo: string | null;
declare function makeFoo(): string;

if (foo == null) /* c1 */ /* c2 */ foo = makeFoo();
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      5, Column: 1, EndLine: 5, EndColumn: 52,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let foo: string | null;
declare function makeFoo(): string;

/* c1 */ /* c2 */ foo ??= makeFoo();
`},
				},
			},
		},
	},
	{
		Code: `
declare let foo: string | null;
declare function makeFoo(): string;

function lazyInitialize() {
  if (foo == null) {
    // a
    // b
    // c
    foo = makeFoo();
  }
}
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      6, Column: 3, EndLine: 11, EndColumn: 4,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let foo: string | null;
declare function makeFoo(): string;

function lazyInitialize() {
  // a
// b
// c
foo ??= makeFoo();
}
`},
				},
			},
		},
	},
	{
		Code: `
declare let foo: string | null;
declare function makeFoo(): string;

function lazyInitialize() {
  if (foo == null) {
    foo = makeFoo();
    // trailing
  }
}
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      6, Column: 3, EndLine: 9, EndColumn: 4,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let foo: string | null;
declare function makeFoo(): string;

function lazyInitialize() {
  foo ??= makeFoo(); // trailing
}
`},
				},
			},
		},
	},
	{
		Code: `
declare let foo: string | null;
declare function makeFoo(): string;

function lazyInitialize() {
  if (foo == null) {
    /* lead */
    foo = makeFoo();
    /* tail */
  }
}
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      6, Column: 3, EndLine: 10, EndColumn: 4,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let foo: string | null;
declare function makeFoo(): string;

function lazyInitialize() {
  /* lead */
foo ??= makeFoo(); /* tail */
}
`},
				},
			},
		},
	},
	{
		Code: `
declare let foo: string | null;
declare function makeFoo(): string;

if (foo == null) // pre
foo = makeFoo();
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      5, Column: 1, EndLine: 6, EndColumn: 17,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let foo: string | null;
declare function makeFoo(): string;

// pre foo ??= makeFoo();
`},
				},
			},
		},
	},
	{
		Code: `
declare let foo: string | null;
declare function makeFoo(): string;

if (foo == null) /* a
b */ foo = makeFoo();
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      5, Column: 1, EndLine: 6, EndColumn: 22,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let foo: string | null;
declare function makeFoo(): string;

/* a
b */ foo ??= makeFoo();
`},
				},
			},
		},
	},
	{
		Code: `
declare let foo: string | null;
declare function makeFoo(): string;
function f() {
  if (foo == null) {
    foo = makeFoo();
  }
}
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      5, Column: 3, EndLine: 7, EndColumn: 4,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let foo: string | null;
declare function makeFoo(): string;
function f() {
  foo ??= makeFoo();
}
`},
				},
			},
		},
	},
	{
		Code: `
declare let foo: string | null;
declare function makeFoo(): string;
if (foo == null) foo = makeFoo();
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      4, Column: 1, EndLine: 4, EndColumn: 34,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let foo: string | null;
declare function makeFoo(): string;
foo ??= makeFoo();
`},
				},
			},
		},
	},
	{
		Code: `
declare let foo: string | null;
declare function makeFoo(): string;
function f() {
  if (foo == null) {
    foo = makeFoo();
    /* multi
       line
    */
  }
}
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      5, Column: 3, EndLine: 10, EndColumn: 4,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let foo: string | null;
declare function makeFoo(): string;
function f() {
  foo ??= makeFoo(); /* multi
       line
    */
}
`},
				},
			},
		},
	},
	{
		Code: `
function Boolean(x: any): boolean { return !!x; }
let a: string | boolean | undefined;
let b: string | boolean | undefined;
const x = Boolean(a || b);
`,
		Options: map[string]any{"ignoreBooleanCoercion": true},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      5, Column: 21, EndLine: 5, EndColumn: 23,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
function Boolean(x: any): boolean { return !!x; }
let a: string | boolean | undefined;
let b: string | boolean | undefined;
const x = Boolean(a ?? b);
`},
				},
			},
		},
	},
	{
		Code: `
type Req = { query: { search?: string } };
declare const req: Req;
const term = req.query.search || 'default';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      4, Column: 31, EndLine: 4, EndColumn: 33,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
type Req = { query: { search?: string } };
declare const req: Req;
const term = req.query.search ?? 'default';
`},
				},
			},
		},
	},
	{
		Code: `
const cache = new Map<string, string>();
const v = cache.get('k') || 'default';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 26, EndLine: 3, EndColumn: 28,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
const cache = new Map<string, string>();
const v = cache.get('k') ?? 'default';
`},
				},
			},
		},
	},
	{
		Code: `
type Props = { name?: string };
declare const props: Props;
declare function Foo(p: any): any;
const el = <Foo name={props.name || 'default'} />;
`,
		Tsx:     true,
		Options: map[string]any{"ignoreConditionalTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      5, Column: 34, EndLine: 5, EndColumn: 36,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
type Props = { name?: string };
declare const props: Props;
declare function Foo(p: any): any;
const el = <Foo name={props.name ?? 'default'} />;
`},
				},
			},
		},
	},
	{
		Code: `
declare function load(): Promise<string | null>;
async function main() {
  const value = (await load()) || 'default';
  return value;
}
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      4, Column: 32, EndLine: 4, EndColumn: 34,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare function load(): Promise<string | null>;
async function main() {
  const value = (await load()) ?? 'default';
  return value;
}
`},
				},
			},
		},
	},
	{
		Code: `
declare const cfg: { host?: string };
const host = cfg.host || 'localhost';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 23, EndLine: 3, EndColumn: 25,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare const cfg: { host?: string };
const host = cfg.host ?? 'localhost';
`},
				},
			},
		},
	},
	{
		Code: `
class Cache {
  #value: string | null = null;
  get(): string {
    return this.#value || 'fresh';
  }
}
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      5, Column: 24, EndLine: 5, EndColumn: 26,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
class Cache {
  #value: string | null = null;
  get(): string {
    return this.#value ?? 'fresh';
  }
}
`},
				},
			},
		},
	},
	{
		Code: `
class User {
  declare name?: string;
  display(): string {
    return this.name || 'anonymous';
  }
}
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      5, Column: 22, EndLine: 5, EndColumn: 24,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
class User {
  declare name?: string;
  display(): string {
    return this.name ?? 'anonymous';
  }
}
`},
				},
			},
		},
	},
	{
		Code: "\ndeclare function tag(s: TemplateStringsArray): string | null;\nconst x = tag`hello` || 'default';\n",
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 22, EndLine: 3, EndColumn: 24,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: "\ndeclare function tag(s: TemplateStringsArray): string | null;\nconst x = tag`hello` ?? 'default';\n"},
				},
			},
		},
	},
	{
		Code: `
const items: string[] = [];
const v = items.find(x => x.length > 0) || 'none';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 41, EndLine: 3, EndColumn: 43,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
const items: string[] = [];
const v = items.find(x => x.length > 0) ?? 'none';
`},
				},
			},
		},
	},
	{
		Code: `
type Dict = { [k: string]: string | undefined };
declare const d: Dict;
const v = d['key'] || 'fallback';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      4, Column: 20, EndLine: 4, EndColumn: 22,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
type Dict = { [k: string]: string | undefined };
declare const d: Dict;
const v = d['key'] ?? 'fallback';
`},
				},
			},
		},
	},
	{
		Code: `
declare const tup: [string | null, number | null];
const a = tup[0] || 'foo';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 18, EndLine: 3, EndColumn: 20,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare const tup: [string | null, number | null];
const a = tup[0] ?? 'foo';
`},
				},
			},
		},
	},
	{
		Code: `
declare const env: { FOO?: string };
const x = env.FOO || 'fallback';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 19, EndLine: 3, EndColumn: 21,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare const env: { FOO?: string };
const x = env.FOO ?? 'fallback';
`},
				},
			},
		},
	},
	{
		Code: `
declare const config: { db?: { host?: string } };
const host = (config.db?.host) || 'localhost';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 32, EndLine: 3, EndColumn: 34,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare const config: { db?: { host?: string } };
const host = (config.db?.host) ?? 'localhost';
`},
				},
			},
		},
	},
	{
		Code: `
declare function f(...args: string[]): void;
declare let opt: string | null;
f(opt || 'x');
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      4, Column: 7, EndLine: 4, EndColumn: 9,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare function f(...args: string[]): void;
declare let opt: string | null;
f(opt ?? 'x');
`},
				},
			},
		},
	},
	{
		Code: `
declare let a: string | null;
const arr = [a || 'foo'];
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 16, EndLine: 3, EndColumn: 18,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let a: string | null;
const arr = [a ?? 'foo'];
`},
				},
			},
		},
	},
	{
		Code: `
declare function load(): Promise<string | null>;
load().then(v => v || 'default');
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 20, EndLine: 3, EndColumn: 22,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare function load(): Promise<string | null>;
load().then(v => v ?? 'default');
`},
				},
			},
		},
	},
	{
		Code: `
declare let cache: Map<string, string | null>;
const get = (k: string) => cache.get(k) || 'fresh';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 41, EndLine: 3, EndColumn: 43,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let cache: Map<string, string | null>;
const get = (k: string) => cache.get(k) ?? 'fresh';
`},
				},
			},
		},
	},
	{
		Code: `
declare const o: { [k: string]: string | undefined };
const v = o['k'] || 'fallback';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 18, EndLine: 3, EndColumn: 20,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare const o: { [k: string]: string | undefined };
const v = o['k'] ?? 'fallback';
`},
				},
			},
		},
	},
	{
		Code: `
declare let x: { a?: { b?: { c?: string } } };
x.a?.b?.c !== undefined ? x.a?.b?.c : 'default';
`,
		Options: map[string]any{"ignoreTernaryTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 48,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let x: { a?: { b?: { c?: string } } };
x.a?.b?.c ?? 'default';
`},
				},
			},
		},
	},
	{
		Code: `
declare let a: string | undefined;
const x = Boolean(!a ? 'fallback' : a);
`,
		Options: map[string]any{"ignoreBooleanCoercion": true},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 19, EndLine: 3, EndColumn: 38,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let a: string | undefined;
const x = Boolean(a ?? 'fallback');
`},
				},
			},
		},
	},
	{
		Code: `
declare let foo: { bar?: string } | null;
foo?.bar !== null && foo?.bar !== undefined ? foo?.bar : 'fallback';
`,
		Options: map[string]any{"ignoreTernaryTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 68,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let foo: { bar?: string } | null;
foo?.bar ?? 'fallback';
`},
				},
			},
		},
	},
	{
		Code: `
declare let foo: void | string;
foo !== undefined ? foo : 'fallback';
`,
		Options: map[string]any{"ignoreTernaryTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: 37,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let foo: void | string;
foo ?? 'fallback';
`},
				},
			},
		},
	},
	{
		Code: `
declare const tup: [string | null, number];
const [a, b] = tup;
const display = a || 'unknown';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      4, Column: 19, EndLine: 4, EndColumn: 21,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare const tup: [string | null, number];
const [a, b] = tup;
const display = a ?? 'unknown';
`},
				},
			},
		},
	},
	{
		Code: `
type Props = { count?: number };
declare const props: Props;
declare function Counter(p: any): any;
const el = <Counter count={props.count !== undefined ? props.count : 0} />;
`,
		Tsx:     true,
		Options: map[string]any{"ignoreTernaryTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      5, Column: 28, EndLine: 5, EndColumn: 71,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
type Props = { count?: number };
declare const props: Props;
declare function Counter(p: any): any;
const el = <Counter count={props.count ?? 0} />;
`},
				},
			},
		},
	},
	{
		Code: `
type Cfg = { server?: { port?: number } };
declare const cfg: Cfg;
const port = cfg.server?.port || 3000;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      4, Column: 31, EndLine: 4, EndColumn: 33,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
type Cfg = { server?: { port?: number } };
declare const cfg: Cfg;
const port = cfg.server?.port ?? 3000;
`},
				},
			},
		},
	},
	{
		Code: `
declare let token: string | null;
const out = token != null ? token : '';
`,
		Options: map[string]any{"ignoreTernaryTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 13, EndLine: 3, EndColumn: 39,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let token: string | null;
const out = token ?? '';
`},
				},
			},
		},
	},
	{
		Code: `
declare const arr: string[];
const v = arr.at(0) || 'none';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 21, EndLine: 3, EndColumn: 23,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare const arr: string[];
const v = arr.at(0) ?? 'none';
`},
				},
			},
		},
	},
	{
		Code: `
declare function show(v: string): void;
declare let v: string | null;
show(v !== null ? v : 'fallback');
`,
		Options: map[string]any{"ignoreTernaryTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      4, Column: 6, EndLine: 4, EndColumn: 33,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare function show(v: string): void;
declare let v: string | null;
show(v ?? 'fallback');
`},
				},
			},
		},
	},
	{
		Code: `
class Thing {
  declare value: string | null;
  ensure() {
    if (this.value == null) {
      this.value = 'init';
    }
  }
}
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      5, Column: 5, EndLine: 7, EndColumn: 6,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
class Thing {
  declare value: string | null;
  ensure() {
    this.value ??= 'init';
  }
}
`},
				},
			},
		},
	},
	{
		Code: `
declare let v: string | null;
function f(arg: string = v !== null ? v : 'x') { return arg; }
`,
		Options: map[string]any{"ignoreTernaryTests": false},
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      3, Column: 26, EndLine: 3, EndColumn: 46,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare let v: string | null;
function f(arg: string = v ?? 'x') { return arg; }
`},
				},
			},
		},
	},
	{
		Code: `
declare const m: Map<string, string | null>;
const v = m.get('k') || 'fallback';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      3, Column: 22, EndLine: 3, EndColumn: 24,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare const m: Map<string, string | null>;
const v = m.get('k') ?? 'fallback';
`},
				},
			},
		},
	},
}

// Behavior that intentionally differs from upstream, or input upstream cannot
// parse. See "Differences from upstream" in prefer_nullish_coalescing.md.
var rslintOnlyInvalid = []rule_tester.InvalidTestCase{
	// nullish branch low prec: upstream suggests `a ?? b || 'z'`, a syntax error; rslint wraps the right operand.
	{
		Code: `declare let a: string | null, b: string;
a !== null ? a : b || 'z';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 26,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string;
a ?? (b || 'z');
`},
				},
			},
		},
	},
	// right &&: upstream suggests `a ?? b && 'z'`, a syntax error; rslint wraps the right operand.
	{
		Code: `declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };
a !== null ? a : b && 'z';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 26,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string, c: string | null, n: number, o: { x?: string };
a ?? (b && 'z');
`},
				},
			},
		},
	},
	// Upstream rejects this input with a parse error (a generator declaration in an ambient context); tsgo parses it, and the `||` on line 4 is reported.
	{
		Code: `
declare function* gen(): Generator<string | null>;
declare const it: Iterator<string | null>;
const r = it.next().value || 'fallback';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      4, Column: 27, EndLine: 4, EndColumn: 29,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `
declare function* gen(): Generator<string | null>;
declare const it: Iterator<string | null>;
const r = it.next().value ?? 'fallback';
`},
				},
			},
		},
	},
}

// TestPreferNullishCoalescingEditDemand checks that every edit demand reports
// the same diagnostics, and that only suggestion demands materialize the
// rule's suggestions. The rule has no autofixes.
func TestPreferNullishCoalescingEditDemand(t *testing.T) {
	t.Parallel()

	const source = `declare let a: string | null;
declare let foo: string | undefined;
a || 'b';
a !== null ? a : 'b';
if (!foo) {
  foo = 'x';
}
`
	wantSuggestions := []string{"??", "a ?? 'b'", "foo ??= 'x';"}

	rootDir := fixtures.GetRootDir()
	fileName := "edit-demand.ts"
	fs := utils.NewOverlayVFS(rootDir.FS, map[string]string{tspath.ResolvePath(rootDir.Dir, fileName): source})
	host := utils.CreateCompilerHost(rootDir.Dir, fs)
	program, err := utils.CreateProgram(true, fs, rootDir.Dir, "tsconfig.json", host)
	if err != nil {
		t.Fatalf("failed to create program: %v", err)
	}
	sourceFile := program.GetSourceFile(fileName)
	if sourceFile == nil {
		t.Fatalf("source file %q not found", fileName)
	}

	lintWithDemand := func(demand rule.EditDemand) []rule.RuleDiagnostic {
		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program:     lintprogram.NewFromCompiler(program),
			File:        sourceFile.FileName(),
			HasTypeInfo: true,
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{
					Name:     PreferNullishCoalescingRule.Name,
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return PreferNullishCoalescingRule.Run(ctx, nil)
					},
				}}
			},
			Consumer: rule.DiagnosticConsumer{
				Demand: demand,
				Report: func(diagnostic rule.RuleDiagnostic) {
					diagnostics = append(diagnostics, diagnostic)
				},
			},
		})
		return diagnostics
	}

	withoutEdits := func(diagnostics []rule.RuleDiagnostic) []rule.RuleDiagnostic {
		out := make([]rule.RuleDiagnostic, len(diagnostics))
		for i, diagnostic := range diagnostics {
			diagnostic.FixesPtr = nil
			diagnostic.Suggestions = nil
			out[i] = diagnostic
		}
		return out
	}

	byDemand := map[rule.EditDemand][]rule.RuleDiagnostic{}
	for _, demand := range []rule.EditDemand{
		rule.EditDemandNone,
		rule.EditDemandAutofix,
		rule.EditDemandSuggestion,
		rule.EditDemandAll,
	} {
		diagnostics := lintWithDemand(demand)
		if len(diagnostics) != len(wantSuggestions) {
			t.Fatalf("demand %d: diagnostics = %d, want %d", demand, len(diagnostics), len(wantSuggestions))
		}
		for _, diagnostic := range diagnostics {
			if diagnostic.FixesPtr != nil {
				t.Errorf("demand %d: unexpected autofix %#v", demand, *diagnostic.FixesPtr)
			}
		}
		byDemand[demand] = diagnostics
	}

	base := withoutEdits(byDemand[rule.EditDemandNone])
	for demand, diagnostics := range byDemand {
		if got := withoutEdits(diagnostics); !reflect.DeepEqual(got, base) {
			t.Errorf("demand %d changed diagnostic metadata:\ngot:  %#v\nwant: %#v", demand, got, base)
		}
	}
	for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix} {
		for _, diagnostic := range byDemand[demand] {
			if diagnostic.Suggestions != nil {
				t.Errorf("demand %d unexpectedly materialized suggestions", demand)
			}
		}
	}
	suggestionOnly := byDemand[rule.EditDemandSuggestion]
	all := byDemand[rule.EditDemandAll]
	for i, diagnostic := range suggestionOnly {
		if diagnostic.Suggestions == nil || all[i].Suggestions == nil {
			t.Fatalf("diagnostic %d: suggestions missing", i)
		}
		if !reflect.DeepEqual(*diagnostic.Suggestions, *all[i].Suggestions) {
			t.Errorf("diagnostic %d: suggestion-only and all-edits suggestions differ", i)
		}
		suggestions := *diagnostic.Suggestions
		if len(suggestions) != 1 {
			t.Fatalf("diagnostic %d: suggestions = %d, want 1", i, len(suggestions))
		}
		fixes := suggestions[0].Fixes()
		var replacement string
		for _, fix := range fixes {
			if fix.Range.Len() > 0 {
				replacement = fix.Text
			}
		}
		if replacement != wantSuggestions[i] {
			t.Errorf("diagnostic %d: replacement = %q, want %q", i, replacement, wantSuggestions[i])
		}
	}
}
