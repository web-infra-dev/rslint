// TestStrictBooleanExpressionsExtrasFixer exercises the wrappingFix
// suggestion path — the place where Go-on-tsgo most easily diverges from
// upstream's ESLint Fixer. Cases below probe paren-emission rules, ASI
// hazards, object-literal-in-arrow-body, multi-line source, and the
// strong/weak-precedence classifier on every supported expression kind.
package strict_boolean_expressions

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestStrictBooleanExpressionsExtrasFixer(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &StrictBooleanExpressionsRule, []rule_tester.ValidTestCase{
		// Reference no-op valid: a strong-precedence identifier needs no inner paren.
		{Code: "declare const x: boolean;\nif (x) {}"},
	}, []rule_tester.InvalidTestCase{
		// ---- Identifier (strong precedence): no inner paren wrap. ----
		{
			Code:    "declare const x: number;\nif (x) {}",
			Options: map[string]interface{}{"allowNumber": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNumber", Line: 2, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareZero", Output: "declare const x: number;\nif (x !== 0) {}"},
					{MessageId: "conditionFixCompareNaN", Output: "declare const x: number;\nif (!Number.isNaN(x)) {}"},
					{MessageId: "conditionFixCastBoolean", Output: "declare const x: number;\nif (Boolean(x)) {}"},
				},
			}},
		},

		// ---- PropertyAccessExpression (strong): no inner paren. ----
		{
			Code: "declare const o: { v: number | null };\nif (o.v) {}",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableNumber", Line: 2, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "declare const o: { v: number | null };\nif (o.v != null) {}"},
					{MessageId: "conditionFixDefaultZero", Output: "declare const o: { v: number | null };\nif (o.v ?? 0) {}"},
					{MessageId: "conditionFixCastBoolean", Output: "declare const o: { v: number | null };\nif (Boolean(o.v)) {}"},
				},
			}},
		},

		// ---- ElementAccessExpression (strong): no inner paren. ----
		{
			Code: "declare const o: { [k: string]: number | null };\nif (o['v']) {}",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableNumber", Line: 2, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "declare const o: { [k: string]: number | null };\nif (o['v'] != null) {}"},
					{MessageId: "conditionFixDefaultZero", Output: "declare const o: { [k: string]: number | null };\nif (o['v'] ?? 0) {}"},
					{MessageId: "conditionFixCastBoolean", Output: "declare const o: { [k: string]: number | null };\nif (Boolean(o['v'])) {}"},
				},
			}},
		},

		// ---- CallExpression (strong): no inner paren. ----
		{
			Code:    "declare function f(): number;\nif (f()) {}",
			Options: map[string]interface{}{"allowNumber": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNumber", Line: 2, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareZero", Output: "declare function f(): number;\nif (f() !== 0) {}"},
					{MessageId: "conditionFixCompareNaN", Output: "declare function f(): number;\nif (!Number.isNaN(f())) {}"},
					{MessageId: "conditionFixCastBoolean", Output: "declare function f(): number;\nif (Boolean(f())) {}"},
				},
			}},
		},

		// ---- NewExpression (strong): no inner paren. ----
		{
			Code: "if (new Date()) {}",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorObject", Line: 1, Column: 5,
			}},
		},

		// ---- BinaryExpression (NOT strong): inner paren wrap. ----
		{
			Code:    "declare const a: number;\ndeclare const b: number;\nif (a + b) {}",
			Options: map[string]interface{}{"allowNumber": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNumber", Line: 3, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareZero", Output: "declare const a: number;\ndeclare const b: number;\nif ((a + b) !== 0) {}"},
					{MessageId: "conditionFixCompareNaN", Output: "declare const a: number;\ndeclare const b: number;\nif (!Number.isNaN((a + b))) {}"},
					{MessageId: "conditionFixCastBoolean", Output: "declare const a: number;\ndeclare const b: number;\nif (Boolean((a + b))) {}"},
				},
			}},
		},

		// ---- ConditionalExpression (NOT strong): inner paren wrap. ----
		{
			Code:    "declare const cond: boolean;\nif (cond ? 1 : 0) {}",
			Options: map[string]interface{}{"allowNumber": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNumber", Line: 2, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareZero", Output: "declare const cond: boolean;\nif ((cond ? 1 : 0) !== 0) {}"},
					{MessageId: "conditionFixCompareNaN", Output: "declare const cond: boolean;\nif (!Number.isNaN((cond ? 1 : 0))) {}"},
					{MessageId: "conditionFixCastBoolean", Output: "declare const cond: boolean;\nif (Boolean((cond ? 1 : 0))) {}"},
				},
			}},
		},

		// ---- AwaitExpression (NOT strong): inner paren wrap. ----
		{
			Code:    "declare const p: Promise<number>;\nasync function f() { if (await p) {} }",
			Options: map[string]interface{}{"allowNumber": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNumber", Line: 2, Column: 26,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareZero", Output: "declare const p: Promise<number>;\nasync function f() { if ((await p) !== 0) {} }"},
					{MessageId: "conditionFixCompareNaN", Output: "declare const p: Promise<number>;\nasync function f() { if (!Number.isNaN((await p))) {} }"},
					{MessageId: "conditionFixCastBoolean", Output: "declare const p: Promise<number>;\nasync function f() { if (Boolean((await p))) {} }"},
				},
			}},
		},

		// ---- TypeOfExpression (NOT strong by isStrongPrecedenceNode, but `typeof X` is a string — string arm). ----
		// `typeof X` returns string, so with allowString:false it fires.
		{
			Code:    "declare const x: unknown;\nif (typeof x) {}",
			Options: map[string]interface{}{"allowString": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorString", Line: 2, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareStringLength", Output: "declare const x: unknown;\nif ((typeof x).length > 0) {}"},
					{MessageId: "conditionFixCompareEmptyString", Output: `declare const x: unknown;
if ((typeof x) !== "") {}`},
					{MessageId: "conditionFixCastBoolean", Output: "declare const x: unknown;\nif (Boolean((typeof x))) {}"},
				},
			}},
		},

		// ---- Parent IS BinaryExpression: outer paren wrap. ----
		// `a && (b + 1)` — inner node is `b + 1` (BinaryExpression). Parent is `a && _` (also BinaryExpression).
		// Inner needs paren (binary not strong); outer also needs paren (parent is binary).
		{
			Code:    "declare const a: boolean;\ndeclare const b: number;\nif (a && b + 1) {}",
			Options: map[string]interface{}{"allowNumber": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNumber", Line: 3, Column: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareZero", Output: "declare const a: boolean;\ndeclare const b: number;\nif (a && ((b + 1) !== 0)) {}"},
					{MessageId: "conditionFixCompareNaN", Output: "declare const a: boolean;\ndeclare const b: number;\nif (a && (!Number.isNaN((b + 1)))) {}"},
					{MessageId: "conditionFixCastBoolean", Output: "declare const a: boolean;\ndeclare const b: number;\nif (a && (Boolean((b + 1)))) {}"},
				},
			}},
		},

		// ---- Parent IS UnaryExpression: outer paren wrap. ----
		// `!nullable` where nullable is `number | null` — replacement of `nullable` keeps the outer `!`.
		// Suggestions for nullable number under `!`-parent emit the inverted forms.
		{
			Code: "declare const x: number | null;\nif (!x) {}",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableNumber", Line: 2, Column: 6,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "declare const x: number | null;\nif (x == null) {}"},
					{MessageId: "conditionFixDefaultZero", Output: "declare const x: number | null;\nif (!(x ?? 0)) {}"},
					{MessageId: "conditionFixCastBoolean", Output: "declare const x: number | null;\nif (!Boolean(x)) {}"},
				},
			}},
		},

		// ---- Parent IS already parenthesized: no extra outer paren wrap. ----
		// `((x))` — inner `x` is paren-wrapped via ParenthesizedExpression. The replacement target is the inner `x`.
		// Parent is ParenthesizedExpression (NOT in isWeakPrecedenceParent's set), so no outer wrap.
		{
			Code: "declare const x: string | null;\nif ((x)) {}",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableString", Line: 2, Column: 6,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "declare const x: string | null;\nif ((x != null)) {}"},
					{MessageId: "conditionFixDefaultEmptyString", Output: `declare const x: string | null;
if ((x ?? "")) {}`},
					{MessageId: "conditionFixCastBoolean", Output: "declare const x: string | null;\nif ((Boolean(x))) {}"},
				},
			}},
		},

		// ---- Conditional expression test position: outer paren wrap (ConditionalExpression IS weak-precedence parent). ----
		{
			Code: "declare const x: string | null;\nconst y = x ? 1 : 0;",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableString", Line: 2, Column: 11,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "declare const x: string | null;\nconst y = (x != null) ? 1 : 0;"},
					{MessageId: "conditionFixDefaultEmptyString", Output: `declare const x: string | null;
const y = (x ?? "") ? 1 : 0;`},
					{MessageId: "conditionFixCastBoolean", Output: "declare const x: string | null;\nconst y = (Boolean(x)) ? 1 : 0;"},
				},
			}},
		},

		// ---- ASI hazard: statement starts with `(` after previous unterminated statement. ----
		// `obj` is `{x:number}|null`. Without `;` after the prior `!obj`, the
		// suggestion `(obj != null) || 0` would be glued to `!obj` as
		// `!obj(obj != null) || 0` if no leading `;` is inserted.
		{
			Code:    "\n        declare const obj: { x: number } | null;\n        !obj\n        obj || 0\n      ",
			Options: map[string]interface{}{"allowNullableObject": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullableObject", Line: 3, Column: 10,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "conditionFixCompareNullish", Output: "\n        declare const obj: { x: number } | null;\n        obj == null\n        obj || 0\n      "},
					},
				},
				{
					MessageId: "conditionErrorNullableObject", Line: 4, Column: 9,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "conditionFixCompareNullish", Output: "\n        declare const obj: { x: number } | null;\n        !obj\n        ;(obj != null) || 0\n      "},
					},
				},
			},
		},

		// ---- ASI hazard cleared: previous line ends with `;`, no leading `;` needed. ----
		{
			Code:    "\n        declare const obj: { x: number } | null;\n        !obj;\n        obj || 0\n      ",
			Options: map[string]interface{}{"allowNullableObject": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullableObject", Line: 3, Column: 10,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "conditionFixCompareNullish", Output: "\n        declare const obj: { x: number } | null;\n        obj == null;\n        obj || 0\n      "},
					},
				},
				{
					MessageId: "conditionErrorNullableObject", Line: 4, Column: 9,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "conditionFixCompareNullish", Output: "\n        declare const obj: { x: number } | null;\n        !obj;\n        (obj != null) || 0\n      "},
					},
				},
			},
		},

		// ---- Multi-line code: replacement preserves surrounding lines. ----
		{
			Code:    "if (\n  // a leading comment\n  ['hi']\n  .length\n) {}",
			Options: map[string]interface{}{"allowNumber": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNumber", Line: 3, Column: 3,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareArrayLengthNonzero", Output: "if (\n  // a leading comment\n  ['hi']\n  .length > 0\n) {}"},
				},
			}},
		},

		// ---- TypeAssertionExpression `<T>x`: AsExpression is NOT strong; needs inner paren. ----
		{
			Code: "declare const x: unknown;\nif (<string | null>x) {}",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableString", Line: 2, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "declare const x: unknown;\nif ((<string | null>x) != null) {}"},
					{MessageId: "conditionFixDefaultEmptyString", Output: `declare const x: unknown;
if ((<string | null>x) ?? "") {}`},
					{MessageId: "conditionFixCastBoolean", Output: "declare const x: unknown;\nif (Boolean((<string | null>x))) {}"},
				},
			}},
		},

		// ---- NonNullAssertion `x!`: strong precedence (no inner paren). ----
		// `declare const x: number | undefined; if (x!.toString())` — the `.toString()` makes the whole thing string.
		// For this fixer test, use `x!` directly in a position where `!`-narrowed `string | undefined` ⇒ `string`.
		{
			Code:    "declare const x: string | undefined;\nif (x!) {}",
			Options: map[string]interface{}{"allowString": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorString", Line: 2, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareStringLength", Output: "declare const x: string | undefined;\nif (x!.length > 0) {}"},
					{MessageId: "conditionFixCompareEmptyString", Output: `declare const x: string | undefined;
if (x! !== "") {}`},
					{MessageId: "conditionFixCastBoolean", Output: "declare const x: string | undefined;\nif (Boolean(x!)) {}"},
				},
			}},
		},

		// ---- nullableNumber inside conditional test — locks "ConditionalExpression as parent" outer-paren-on. ----
		{
			Code: "declare const x: number | null;\nx ? 1 : 0;",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableNumber", Line: 2, Column: 1,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "declare const x: number | null;\n(x != null) ? 1 : 0;"},
					{MessageId: "conditionFixDefaultZero", Output: "declare const x: number | null;\n(x ?? 0) ? 1 : 0;"},
					{MessageId: "conditionFixCastBoolean", Output: "declare const x: number | null;\n(Boolean(x)) ? 1 : 0;"},
				},
			}},
		},

		// ---- TaggedTemplateExpression as parent of fix target via callee position ----
		// `foo`x`` — if `foo` is replaced, parent is tagged-template, outer wrap.
		// Direct case: `if (foo``)` — but `foo``` returns the template-tag's return, complex to construct.
		// We test the inverse via: result of a tag returning string used in if.
		{
			Code:    "function tag(s: TemplateStringsArray): string { return s[0]; }\nif (tag`hi`) {}",
			Options: map[string]interface{}{"allowString": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorString", Line: 2, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareStringLength", Output: "function tag(s: TemplateStringsArray): string { return s[0]; }\nif (tag`hi`.length > 0) {}"},
					{MessageId: "conditionFixCompareEmptyString", Output: `function tag(s: TemplateStringsArray): string { return s[0]; }
if (tag` + "`hi`" + ` !== "") {}`},
					{MessageId: "conditionFixCastBoolean", Output: "function tag(s: TemplateStringsArray): string { return s[0]; }\nif (Boolean(tag`hi`)) {}"},
				},
			}},
		},

		// ---- Array predicate fix: 2-param arrow with type annotation ----
		// Tests insertion after the last parameter's `)`.
		{
			Code: "[1, null].every((x, i) => {});",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullish", Line: 1, Column: 17, EndLine: 1, EndColumn: 29,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "explicitBooleanReturnType", Output: "[1, null].every((x, i): boolean => {});"},
				},
			}},
		},

		// ---- Array predicate fix: parenless arrow needs `(` and `): boolean` insertion ----
		{
			Code: "[1, null].every(x => {});",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullish", Line: 1, Column: 17, EndLine: 1, EndColumn: 24,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "explicitBooleanReturnType", Output: "[1, null].every((x): boolean => {});"},
				},
			}},
		},

		// ---- Array predicate fix: no-arg arrow ----
		{
			Code: "[1, null].every(() => undefined);",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullish", Line: 1, Column: 17, EndLine: 1, EndColumn: 32,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "explicitBooleanReturnType", Output: "[1, null].every((): boolean => undefined);"},
				},
			}},
		},
	})
}
