package prefer_nullish_coalescing

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// These expectations preserve expression grouping, distinguish dynamic keys,
// and retain comments. They intentionally reject unsafe upstream suggestions.
// Every suggestion was independently checked with a JavaScript/TypeScript
// syntax checker; representative cases also have runtime counterexamples.
func TestPreferNullishCoalescingSafety(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferNullishCoalescingRule, safetyValid, safetyInvalid)
}

var safetyValid = []rule_tester.ValidTestCase{
	// key b.c / b[c]
	{
		Code: `declare const a: Record<string,string | undefined>; declare const b: { c: string; [x:string]: string }; declare const c: string;
a[b.c] !== undefined ? a[b[c]] : 'fallback';
`,
	},
	// if key b.c / b[c]
	{
		Code: `declare const a: Record<string,string | undefined>; declare const b: { c: string; [x:string]: string }; declare const c: string;
if (a[b.c] === undefined) a[b[c]] = 'fallback';
`,
	},
	// key b[c] / b.c
	{
		Code: `declare const a: Record<string,string | undefined>; declare const b: { c: string; [x:string]: string }; declare const c: string;
a[b[c]] !== undefined ? a[b.c] : 'fallback';
`,
	},
	// if key b[c] / b.c
	{
		Code: `declare const a: Record<string,string | undefined>; declare const b: { c: string; [x:string]: string }; declare const c: string;
if (a[b[c]] === undefined) a[b.c] = 'fallback';
`,
	},
	// key b.c / b['c']
	{
		Code: `declare const a: Record<string,string | undefined>; declare const b: { c: string; [x:string]: string }; declare const c: string;
a[b.c] !== undefined ? a[b['c']] : 'fallback';
`,
	},
	// if key b.c / b['c']
	{
		Code: `declare const a: Record<string,string | undefined>; declare const b: { c: string; [x:string]: string }; declare const c: string;
if (a[b.c] === undefined) a[b['c']] = 'fallback';
`,
	},
	// key b?.c / b.c
	{
		Code: `declare const a: Record<string,string | undefined>; declare const b: { c: string; [x:string]: string }; declare const c: string;
a[b?.c] !== undefined ? a[b.c] : 'fallback';
`,
	},
	// if key b?.c / b.c
	{
		Code: `declare const a: Record<string,string | undefined>; declare const b: { c: string; [x:string]: string }; declare const c: string;
if (a[b?.c] === undefined) a[b.c] = 'fallback';
`,
	},
	// key (b?.c) / (b.c)
	{
		Code: `declare const a: Record<string,string | undefined>; declare const b: { c: string; [x:string]: string }; declare const c: string;
a[(b?.c)] !== undefined ? a[(b.c)] : 'fallback';
`,
	},
	// if key (b?.c) / (b.c)
	{
		Code: `declare const a: Record<string,string | undefined>; declare const b: { c: string; [x:string]: string }; declare const c: string;
if (a[(b?.c)] === undefined) a[(b.c)] = 'fallback';
`,
	},
	// if +=
	{
		Code: `declare let a: string | undefined;
if (a === undefined) a += 'x';
`,
	},
	// if -=
	{
		Code: `declare let a: string | undefined;
if (a === undefined) a -= 'x';
`,
	},
	// if *=
	{
		Code: `declare let a: string | undefined;
if (a === undefined) a *= 'x';
`,
	},
	// if /=
	{
		Code: `declare let a: string | undefined;
if (a === undefined) a /= 'x';
`,
	},
	// if &&=
	{
		Code: `declare let a: string | undefined;
if (a === undefined) a &&= 'x';
`,
	},
	// shadow undefined
	{
		Code: `function f(a: number | undefined, undefined: number) { return a === undefined ? 10 : a; }`,
	},
	// optional key
	{
		Code: `declare let a: Record<string,string|undefined>, b: {c:string} | undefined;
a[b?.c] !== undefined ? a[b.c] : 'x';
`,
	},
	// nullable optional key b?.c / b.c
	{
		Code: `declare let a: Record<string, string | undefined>, b: any;
a[b?.c] != null ? a[b.c] : 'x';
`,
	},
	// nullable optional key b.c / b?.c
	{
		Code: `declare let a: Record<string, string | undefined>, b: any;
a[b.c] != null ? a[b?.c] : 'x';
`,
	},
	// nullable optional key (b?.c).d / b?.c.d
	{
		Code: `declare let a: Record<string, string | undefined>, b: any;
a[(b?.c).d] != null ? a[b?.c.d] : 'x';
`,
	},
	// shadow block
	{
		Code: `declare let a: number | undefined;
{ const undefined = 5; a === undefined ? 10 : a; }
`,
	},
	// shadow catch
	{
		Code: `declare let a: number | undefined;
try {} catch (undefined) { a === undefined ? 10 : a; }
`,
	},
	// shadow if
	{
		Code: `function f(a: number | undefined, undefined: number) { if (a === undefined) a = 10; }
`,
	},
	// shadow hoisted
	{
		Code: `function f(a: number | undefined) { const r = a === undefined ? 10 : a; var undefined = 5; return r; }
`,
	},
	// same optional keys conservative
	{
		Code: `declare let a: Record<string,string|undefined>, b: any;
a[b?.c] != null ? a[b?.c] : 'x';
`,
	},
	// nested dynamic key
	{
		Code: `declare let a: Record<string,string|undefined>, b: any, c: any;
a[b.c.d] != null ? a[b[c].d] : 'x';
`,
	},
}

var safetyInvalid = []rule_tester.InvalidTestCase{
	// a ||= b || c || d
	{
		Code: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
a ||= b || c || d;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      2, Column: 3, EndLine: 2, EndColumn: 6,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
a ??= b || c || d;
`},
				},
			},
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 9, EndLine: 2, EndColumn: 11,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
a ||= (b ?? c) || d;
`},
				},
			},
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 14, EndLine: 2, EndColumn: 16,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
a ||= (b || c) ?? d;
`},
				},
			},
		},
	},
	// a = b || c || d
	{
		Code: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
a = b || c || d;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 7, EndLine: 2, EndColumn: 9,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
a = (b ?? c) || d;
`},
				},
			},
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 12, EndLine: 2, EndColumn: 14,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
a = (b || c) ?? d;
`},
				},
			},
		},
	},
	// b || c || d
	{
		Code: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
b || c || d;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 3, EndLine: 2, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
(b ?? c) || d;
`},
				},
			},
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 8, EndLine: 2, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
(b || c) ?? d;
`},
				},
			},
		},
	},
	// a && (b || c)
	{
		Code: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
a && (b || c);
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 9, EndLine: 2, EndColumn: 11,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
a && (b ?? c);
`},
				},
			},
		},
	},
	// (a || b) && c
	{
		Code: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
(a || b) && c;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 4, EndLine: 2, EndColumn: 6,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
(a ?? b) && c;
`},
				},
			},
		},
	},
	// (a || b) ?? c
	{
		Code: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
(a || b) ?? c;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 4, EndLine: 2, EndColumn: 6,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
(a ?? b) ?? c;
`},
				},
			},
		},
	},
	// a ?? (b || c)
	{
		Code: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
a ?? (b || c);
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 9, EndLine: 2, EndColumn: 11,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
a ?? (b ?? c);
`},
				},
			},
		},
	},
	// a || b && c
	{
		Code: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
a || b && c;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 3, EndLine: 2, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
a ?? (b && c);
`},
				},
			},
		},
	},
	// a && b || c
	{
		Code: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
a && b || c;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 8, EndLine: 2, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
(a && b) ?? c;
`},
				},
			},
		},
	},
	// a || b || c || d
	{
		Code: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
a || b || c || d;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 3, EndLine: 2, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
(a ?? b) || c || d;
`},
				},
			},
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 8, EndLine: 2, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
((a || b) ?? c) || d;
`},
				},
			},
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 13, EndLine: 2, EndColumn: 15,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
(a || b || c) ?? d;
`},
				},
			},
		},
	},
	// a || (b || c)
	{
		Code: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
a || (b || c);
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 3, EndLine: 2, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
a ?? (b || c);
`},
				},
			},
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 9, EndLine: 2, EndColumn: 11,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
a || (b ?? c);
`},
				},
			},
		},
	},
	// (a || b) || c
	{
		Code: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
(a || b) || c;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 4, EndLine: 2, EndColumn: 6,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
(a ?? b) || c;
`},
				},
			},
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 10, EndLine: 2, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
(a || b) ?? c;
`},
				},
			},
		},
	},
	// a ||= (b || c) || d
	{
		Code: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
a ||= (b || c) || d;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      2, Column: 3, EndLine: 2, EndColumn: 6,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
a ??= (b || c) || d;
`},
				},
			},
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 10, EndLine: 2, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
a ||= (b ?? c) || d;
`},
				},
			},
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 16, EndLine: 2, EndColumn: 18,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
a ||= (b || c) ?? d;
`},
				},
			},
		},
	},
	// a ||= b || (c || d)
	{
		Code: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
a ||= b || (c || d);
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      2, Column: 3, EndLine: 2, EndColumn: 6,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
a ??= b || (c || d);
`},
				},
			},
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 9, EndLine: 2, EndColumn: 11,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
a ||= b ?? (c || d);
`},
				},
			},
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 15, EndLine: 2, EndColumn: 17,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
a ||= b || (c ?? d);
`},
				},
			},
		},
	},
	// a ||= (b || c || d)
	{
		Code: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
a ||= (b || c || d);
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      2, Column: 3, EndLine: 2, EndColumn: 6,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
a ??= (b || c || d);
`},
				},
			},
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 10, EndLine: 2, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
a ||= ((b ?? c) || d);
`},
				},
			},
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 15, EndLine: 2, EndColumn: 17,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
a ||= ((b || c) ?? d);
`},
				},
			},
		},
	},
	// a ||= b || c
	{
		Code: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
a ||= b || c;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      2, Column: 3, EndLine: 2, EndColumn: 6,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
a ??= b || c;
`},
				},
			},
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 9, EndLine: 2, EndColumn: 11,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
a ||= b ?? c;
`},
				},
			},
		},
	},
	// a &&= b || c || d
	{
		Code: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
a &&= b || c || d;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 9, EndLine: 2, EndColumn: 11,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
a &&= (b ?? c) || d;
`},
				},
			},
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 14, EndLine: 2, EndColumn: 16,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
a &&= (b || c) ?? d;
`},
				},
			},
		},
	},
	// a ??= b || c || d
	{
		Code: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
a ??= b || c || d;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 9, EndLine: 2, EndColumn: 11,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
a ??= (b ?? c) || d;
`},
				},
			},
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 14, EndLine: 2, EndColumn: 16,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
a ??= (b || c) ?? d;
`},
				},
			},
		},
	},
	// a || (b && c) || d
	{
		Code: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
a || (b && c) || d;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 3, EndLine: 2, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
(a ?? (b && c)) || d;
`},
				},
			},
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 15, EndLine: 2, EndColumn: 17,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
(a || (b && c)) ?? d;
`},
				},
			},
		},
	},
	// (a = b) || c || d
	{
		Code: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
(a = b) || c || d;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 9, EndLine: 2, EndColumn: 11,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
((a = b) ?? c) || d;
`},
				},
			},
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 14, EndLine: 2, EndColumn: 16,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
((a = b) || c) ?? d;
`},
				},
			},
		},
	},
	// key b.c / b.c
	{
		Code: `declare const a: Record<string,string | undefined>; declare const b: { c: string; [x:string]: string }; declare const c: string;
a[b.c] !== undefined ? a[b.c] : 'fallback';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 43,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare const a: Record<string,string | undefined>; declare const b: { c: string; [x:string]: string }; declare const c: string;
a[b.c] ?? 'fallback';
`},
				},
			},
		},
	},
	// if key b.c / b.c
	{
		Code: `declare const a: Record<string,string | undefined>; declare const b: { c: string; [x:string]: string }; declare const c: string;
if (a[b.c] === undefined) a[b.c] = 'fallback';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 47,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare const a: Record<string,string | undefined>; declare const b: { c: string; [x:string]: string }; declare const c: string;
a[b.c] ??= 'fallback';
`},
				},
			},
		},
	},
	// if =
	{
		Code: `declare let a: string | undefined;
if (a === undefined) a = 'x';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 30,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | undefined;
a ??= 'x';
`},
				},
			},
		},
	},
	// if ||=
	{
		Code: `declare let a: string | undefined;
if (a === undefined) a ||= 'x';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 32,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | undefined;
a ??= 'x';
`},
				},
			},
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of a logical assignment (`||=`), as it is a safer operator.",
				Line:      2, Column: 24, EndLine: 2, EndColumn: 27,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | undefined;
if (a === undefined) a ??= 'x';
`},
				},
			},
		},
	},
	// if ??=
	{
		Code: `declare let a: string | undefined;
if (a === undefined) a ??= 'x';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 32,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | undefined;
a ??= 'x';
`},
				},
			},
		},
	},
	// optional target
	{
		Code: `declare let a: { x?: string } | undefined;
if (a?.x === undefined) a.x = "x";
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 35,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: { x?: string } | undefined;
a.x ??= "x";
`},
				},
			},
		},
	},
	// line comment if
	{
		Code: `declare let a: string | undefined;
if (a === undefined) // keep
 a = "x";
a.toString();
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 3, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | undefined;
// keep
a ??= "x";
a.toString();
`},
				},
			},
		},
	},
	// line comment nested if
	{
		Code: `declare let a: string | undefined; declare let flag: boolean;
if (flag) if (a === undefined) // keep
 a = "x";
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      2, Column: 11, EndLine: 3, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | undefined; declare let flag: boolean;
if (flag) // keep
a ??= "x";
`},
				},
			},
		},
	},
	// comment ternary
	{
		Code: `declare let a: string | undefined;
a !== undefined /* keep */ ? a : 'x';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 37,
			},
		},
	},
	// optional reverse
	{
		Code: `declare let a: { b?: string } | undefined;
a.b !== undefined ? a?.b : 'x';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 31,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: { b?: string } | undefined;
a.b ?? 'x';
`},
				},
			},
		},
	},
	// (a || b) || c || d
	{
		Code: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
(a || b) || c || d;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 4, EndLine: 2, EndColumn: 6,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
(a ?? b) || c || d;
`},
				},
			},
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 10, EndLine: 2, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
((a || b) ?? c) || d;
`},
				},
			},
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 15, EndLine: 2, EndColumn: 17,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
((a || b) || c) ?? d;
`},
				},
			},
		},
	},
	// (a && b) || c || d
	{
		Code: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
(a && b) || c || d;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 10, EndLine: 2, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
((a && b) ?? c) || d;
`},
				},
			},
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 15, EndLine: 2, EndColumn: 17,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
((a && b) || c) ?? d;
`},
				},
			},
		},
	},
	// (a ?? b) || c || d
	{
		Code: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
(a ?? b) || c || d;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 10, EndLine: 2, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
((a ?? b) ?? c) || d;
`},
				},
			},
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 15, EndLine: 2, EndColumn: 17,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
((a ?? b) || c) ?? d;
`},
				},
			},
		},
	},
	// a || b && c || d
	{
		Code: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
a || b && c || d;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 3, EndLine: 2, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
(a ?? (b && c)) || d;
`},
				},
			},
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 13, EndLine: 2, EndColumn: 15,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
(a || b && c) ?? d;
`},
				},
			},
		},
	},
	// a || ((b || c)) || d
	{
		Code: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
a || ((b || c)) || d;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 3, EndLine: 2, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
(a ?? ((b || c))) || d;
`},
				},
			},
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 10, EndLine: 2, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
a || ((b ?? c)) || d;
`},
				},
			},
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 17, EndLine: 2, EndColumn: 19,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
(a || ((b || c))) ?? d;
`},
				},
			},
		},
	},
	// (a || b || c) || d
	{
		Code: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
(a || b || c) || d;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 4, EndLine: 2, EndColumn: 6,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
((a ?? b) || c) || d;
`},
				},
			},
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 9, EndLine: 2, EndColumn: 11,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
((a || b) ?? c) || d;
`},
				},
			},
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 15, EndLine: 2, EndColumn: 17,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
(a || b || c) ?? d;
`},
				},
			},
		},
	},
	// a || (b = c) || d
	{
		Code: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
a || (b = c) || d;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 3, EndLine: 2, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
(a ?? (b = c)) || d;
`},
				},
			},
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 14, EndLine: 2, EndColumn: 16,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
(a || (b = c)) ?? d;
`},
				},
			},
		},
	},
	// a || (b, c) || d
	{
		Code: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
a || (b, c) || d;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 3, EndLine: 2, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
(a ?? (b, c)) || d;
`},
				},
			},
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 13, EndLine: 2, EndColumn: 15,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
(a || (b, c)) ?? d;
`},
				},
			},
		},
	},
	// (a as string | null) || b || c
	{
		Code: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
(a as string | null) || b || c;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 22, EndLine: 2, EndColumn: 24,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
((a as string | null) ?? b) || c;
`},
				},
			},
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 27, EndLine: 2, EndColumn: 29,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
((a as string | null) || b) ?? c;
`},
				},
			},
		},
	},
	// a || (b satisfies string | null) || c
	{
		Code: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
a || (b satisfies string | null) || c;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 3, EndLine: 2, EndColumn: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
(a ?? (b satisfies string | null)) || c;
`},
				},
			},
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 34, EndLine: 2, EndColumn: 36,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null, b: string | null, c: string | null, d: string, e: string;
(a || (b satisfies string | null)) ?? c;
`},
				},
			},
		},
	},
	// direct static equivalent
	{
		Code: `declare let a: { x?: string };
a.x !== undefined ? a['x'] : 'z';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 33,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: { x?: string };
a.x ?? 'z';
`},
				},
			},
		},
	},
	// if trailing line before next statement
	{
		Code: `declare let a: string | undefined;
if (a === undefined) { a = 'x'; // keep
} a.toString();
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 3, EndColumn: 2,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | undefined;
a ??= 'x'; // keep
 a.toString();
`},
				},
			},
		},
	},
	// if condition comment
	{
		Code: `declare let a: string | undefined;
if (a === undefined /* keep */) a = 'x';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 41,
			},
		},
	},
	// if assignment comment
	{
		Code: `declare let a: string | undefined;
if (a === undefined) a /* keep */ = 'x';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 41,
			},
		},
	},
	// if RHS leading comment
	{
		Code: `declare let a: string | undefined;
if (a === undefined) a = /* keep */ 'x';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 41,
			},
		},
	},
	// if retained RHS comment
	{
		Code: `declare let a: string | undefined;
if (a === undefined) a = (/* keep */ 'x');
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 43,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | undefined;
a ??= (/* keep */ 'x');
`},
				},
			},
		},
	},
	// ternary retained RHS comment
	{
		Code: `declare let a: string | undefined;
a !== undefined ? a : (/* keep */ 'x');
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 39,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | undefined;
a ?? (/* keep */ 'x');
`},
				},
			},
		},
	},
	// ternary outer parentheses comment
	{
		Code: `declare let a: string | undefined;
a !== undefined ? a : (/* keep */ ('x'));
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 41,
			},
		},
	},
	// line comment CRLF
	{
		Code: "declare let a: string | undefined;\r\nif (a === undefined) // keep\r\n a = \"x\";\r\n",
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverAssignment",
				Message:   "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 3, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: "declare let a: string | undefined;\r\n// keep\na ??= \"x\";\r\n"},
				},
			},
		},
	},
	// global undefined control
	{
		Code: `declare let a: number | undefined;
a === undefined ? 10 : a;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 25,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: number | undefined;
a ?? 10;
`},
				},
			},
		},
	},
	// outer optional chain control
	{
		Code: `declare let a: { b?: string } | undefined;
a?.b !== undefined ? a.b : 'x';
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverTernary",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 31,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: { b?: string } | undefined;
a?.b ?? 'x';
`},
				},
			},
		},
	},
	// operator adjacent comments
	{
		Code: `declare let a: string | null,b: string | null,c: string;
a /* left */ || b /* middle */ || c;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 14, EndLine: 2, EndColumn: 16,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null,b: string | null,c: string;
(a /* left */ ?? b) /* middle */ || c;
`},
				},
			},
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 32, EndLine: 2, EndColumn: 34,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null,b: string | null,c: string;
(a /* left */ || b) /* middle */ ?? c;
`},
				},
			},
		},
	},
	// assignment parenthesized left
	{
		Code: `declare let a: string | null,b: string | null,c: string,d: string;
((a = b)) || c || d;
`,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "preferNullishOverOr",
				Message:   "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
				Line:      2, Column: 11, EndLine: 2, EndColumn: 13,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestNullish", Output: `declare let a: string | null,b: string | null,c: string,d: string;
(((a = b)) ?? c) || d;
`},
				},
			},
		},
	},
}
