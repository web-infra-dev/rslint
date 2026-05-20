// TestStrictBooleanExpressionsExtras locks in branches and edge shapes that
// the upstream test suite doesn't exercise. Each case carries an inline
// comment pointing at the specific branch / Dimension 4 row / tsgo AST quirk
// it covers, so future refactors can't silently regress them without breaking
// a named lock-in.
package strict_boolean_expressions

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestStrictBooleanExpressionsExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &StrictBooleanExpressionsRule, []rule_tester.ValidTestCase{
		// ==== Dimension 4: receiver / expression wrappers ====
		// Single & multi-level parenthesized boolean stays valid.
		{Code: "declare const x: boolean;\nif ((x)) {}"},
		{Code: "declare const x: boolean;\nif (((x))) {}"},
		// Non-null assertion narrows `string | undefined` to `string`.
		{Code: "declare const x: string | undefined;\nif (x!) {}"},
		// `as boolean` is a transparent type assertion.
		{Code: "declare const x: unknown;\nif (x as boolean) {}"},
		// `satisfies boolean` preserves boolean.
		{Code: "declare const x: boolean;\nif (x satisfies boolean) {}"},
		// Optional chain producing boolean | undefined is OK once allowNullableBoolean is on.
		{
			Code:    "declare const x: { y?: boolean };\nif (x.y) {}",
			Options: map[string]interface{}{"allowNullableBoolean": true},
		},
		// Optional call producing boolean is OK once allowNullableBoolean is on.
		{
			Code:    "declare const f: (() => boolean) | undefined;\nif (f?.()) {}",
			Options: map[string]interface{}{"allowNullableBoolean": true},
		},

		// ==== Dimension 4: access / key forms (rule inspects member access via array.length etc.) ====
		// Element access with string literal key — same as dotted member.
		{
			Code:    "declare const a: number[];\nif (a['length']) {}",
			Options: map[string]interface{}{"allowNumber": true},
		},
		// Element access with template literal key.
		{
			Code:    "declare const a: number[];\nif (a[`length`]) {}",
			Options: map[string]interface{}{"allowNumber": true},
		},

		// ==== Dimension 4: declaration / container forms ====
		// Function declaration body, arrow body, function expression body.
		{Code: "function f() { if (true) return; }"},
		{Code: "const f = () => { if (true) return; };"},
		{Code: "const f = function() { if (true) return; };"},
		// Class method body — inner if-condition still checked.
		{Code: "class C { m() { if (true) return; } }"},
		// Class field arrow.
		{Code: "class C { f = () => { if (true) return; }; }"},
		// Async function body — inner condition still checked.
		{Code: "async function f() { if (true) return; }"},
		// Generator body.
		{Code: "function* g() { if (true) yield; }"},

		// ==== Dimension 4: nesting / traversal boundaries ====
		// Function-in-function: outer arrow body and inner arrow body each have
		// their own condition; both being boolean stays valid.
		{Code: "const f = (x: boolean) => (b: boolean) => (b ? 1 : 0);"},
		// IfStatement nested in IfStatement.
		{Code: "declare const a: boolean;\ndeclare const b: boolean;\nif (a) { if (b) {} }"},
		// Class static block — inner if-condition still checked, boolean stays valid.
		{Code: "class C { static { if (true) { /* ok */ } } }"},

		// ==== Dimension 4: graceful degradation ====
		// Empty `for` body / empty if body — no crash on null statement.
		{Code: "declare const x: boolean;\nif (x) {}"},
		{Code: "for (let i = 0; i < 1; i++) {}"},
		// Empty arrow body.
		{Code: "[true].some(x => x);"},

		// ==== Branch lock-ins: determineReportType arms ====
		// `never` exits without reporting (locks in early-exit arm).
		{Code: "declare const x: never;\nif (x) {}"},
		// `boolean` alone exits without reporting.
		{Code: "declare const x: boolean;\nif (x) {}"},
		// `true` literal alone (`truthy boolean` arm) exits without reporting.
		{Code: "if (true) {}"},
		// `nullish | truthy boolean` exits with no report (early "always-true" arm).
		{Code: "declare const x: true | null;\nif (x) {}"},

		// ==== Real-user shapes ====
		// Discriminated union switch — each case narrows away to a unique type.
		{Code: "type T = { kind: 'a'; v: number } | { kind: 'b'; v: string };\nfunction f(t: T) {\n  switch (t.kind) {\n    case 'a': if (t.v) return; break;\n    case 'b': if (t.v.length) return; break;\n  }\n}"},
		// Function component pattern: early-return on boolean.
		{Code: "declare function useLoading(): boolean;\nfunction Comp() { const loading = useLoading(); if (loading) return null; return 1; }"},
		// Nullish coalescing chain (`??`) is NOT a LogicalExpression we check.
		{Code: "declare const a: string | null;\nconst b = a ?? 'default';"},

		// ==== Real-user: array predicate type narrowing ====
		// Predicate returning boolean by `is`-narrowing.
		{Code: "declare function isString(x: unknown): x is string;\n['a', 1, null].filter(isString);"},
		// Arrow predicate already returning explicit boolean comparison.
		{Code: "declare const arr: (string | null)[];\narr.filter(x => x === 'a');"},

		// ==== Dimension 4: more wrapper combinations on the test condition ====
		// `as boolean` then negate.
		{Code: "declare const x: unknown;\n!(x as boolean);"},
		// `satisfies boolean` in conditional.
		{Code: "declare const x: boolean;\nconst y = (x satisfies boolean) ? 1 : 0;"},
		// Non-null + paren + as combo.
		{Code: "declare const x: boolean | undefined;\nif ((x! as boolean)) {}"},

		// ==== Dimension 4: access key forms producing the same array.length detection ====
		// Computed-property element access stays distinct (not array.length).
		{Code: "declare const a: number[];\nif (a['length']) {}", Options: map[string]interface{}{"allowNumber": true}},
		{Code: "declare const a: number[];\nif (a[`length`]) {}", Options: map[string]interface{}{"allowNumber": true}},

		// ==== Dimension 4: class expression vs declaration ====
		// Class declaration with method's condition.
		{Code: "class A { m() { if (true) {} } }"},
		// Class expression with method's condition.
		{Code: "const A = class { m() { if (true) {} } };"},
		// Constructor body.
		{Code: "class A { constructor() { if (true) {} } }"},
		// Get/set accessor bodies.
		{Code: "class A { get x() { if (true) return 1; return 0; } }"},
		{Code: "class A { set x(v: boolean) { if (v) {} } }"},

		// ==== Dimension 4: function-like variants ====
		// Async arrow with boolean body.
		{Code: "const f = async (x: boolean) => x;"},
		// Async generator function.
		{Code: "async function* g(x: boolean) { yield x; if (x) yield; }"},
		// Generator method.
		{Code: "class A { *g(x: boolean) { yield x; if (x) yield; } }"},

		// ==== Dimension 4: arrow function with paren-wrapped body returning a condition ====
		{Code: "const f = (x: boolean) => (x ? 1 : 0);"},

		// ==== Branch lock-ins for nullable-enum variants we haven't covered as valid ====
		// `nullish | string | enum` with allowNullableEnum:true (mixed enum).
		{
			Code:    "\nenum E { A = 'a', B = '' }\nfunction f(e: E | null) { if (e) {} }\n",
			Options: map[string]interface{}{"allowNullableEnum": true},
		},
		// `nullish | number | enum` with allowNullableEnum:true.
		{
			Code:    "\nenum E { A = 1, B = 0 }\nfunction f(e: E | null) { if (e) {} }\n",
			Options: map[string]interface{}{"allowNullableEnum": true},
		},
		// `nullish | string | number | enum` mixed enum w/ allowNullableEnum:true.
		{
			Code:    "\nenum E { A = 1, B = 'b', C = 0, D = '' }\nfunction f(e: E | null) { if (e) {} }\n",
			Options: map[string]interface{}{"allowNullableEnum": true},
		},
	}, []rule_tester.InvalidTestCase{
		// ==== Locks in determineReportType arms not exercised by upstream ====
		// `object` alone (not nullable) — distinct from `nullish | object`.
		{
			Code: "declare const x: { a: number };\nif (x) {}",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorObject", Line: 2, Column: 5,
			}},
		},
		// `null | undefined` literal union — locks `conditionErrorNullish` arm without other types.
		{
			Code: "declare const x: null | undefined;\nif (x) {}",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullish", Line: 2, Column: 5,
			}},
		},

		// ==== Dimension 4: optional chain test — boolean | undefined is nullable boolean by default ====
		{
			Code: "declare const x: { y?: boolean };\nif (x.y) {}",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableBoolean", Line: 2, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixDefaultFalse", Output: "declare const x: { y?: boolean };\nif (x.y ?? false) {}"},
					{MessageId: "conditionFixCompareTrue", Output: "declare const x: { y?: boolean };\nif (x.y === true) {}"},
				},
			}},
		},

		// ==== Dimension 4: optional call producing boolean | undefined ====
		{
			Code: "declare const f: (() => boolean) | undefined;\nif (f?.()) {}",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableBoolean", Line: 2, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixDefaultFalse", Output: "declare const f: (() => boolean) | undefined;\nif (f?.() ?? false) {}"},
					{MessageId: "conditionFixCompareTrue", Output: "declare const f: (() => boolean) | undefined;\nif (f?.() === true) {}"},
				},
			}},
		},

		// ==== Dimension 4: parenthesized test condition — tsgo retains the paren wrapper ====
		{
			Code: "declare const x: string | null;\nif ((x)) {}",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableString", Line: 2, Column: 6,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "declare const x: string | null;\nif ((x != null)) {}"},
					{MessageId: "conditionFixDefaultEmptyString", Output: "declare const x: string | null;\nif ((x ?? \"\")) {}"},
					{MessageId: "conditionFixCastBoolean", Output: "declare const x: string | null;\nif ((Boolean(x))) {}"},
				},
			}},
		},

		// ==== Dimension 4: multi-level paren ====
		{
			Code: "declare const x: string | null;\nif (((x))) {}",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableString", Line: 2, Column: 7,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "declare const x: string | null;\nif (((x != null))) {}"},
					{MessageId: "conditionFixDefaultEmptyString", Output: "declare const x: string | null;\nif (((x ?? \"\"))) {}"},
					{MessageId: "conditionFixCastBoolean", Output: "declare const x: string | null;\nif (((Boolean(x)))) {}"},
				},
			}},
		},

		// ==== Locks traverseLogical arm: right operand inside if-condition IS checked ====
		{
			Code: "declare const a: boolean;\ndeclare const b: number | null;\nif (a && b) {}",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableNumber", Line: 3, Column: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "declare const a: boolean;\ndeclare const b: number | null;\nif (a && (b != null)) {}"},
					{MessageId: "conditionFixDefaultZero", Output: "declare const a: boolean;\ndeclare const b: number | null;\nif (a && (b ?? 0)) {}"},
					{MessageId: "conditionFixCastBoolean", Output: "declare const a: boolean;\ndeclare const b: number | null;\nif (a && (Boolean(b))) {}"},
				},
			}},
		},

		// ==== Locks traverseLogical arm: right operand in top-level chain is NOT checked ====
		{
			Code: "declare const a: object;\ndeclare const b: number;\na || b;",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorObject", Line: 3, Column: 1,
			}},
		},

		// ==== Locks inspectVariantTypes branded-boolean negative path ====
		// `object & { __brand }` is NOT a branded boolean — must report as object.
		{
			Code: "declare const x: object & { __brand: 'Foo' };\nif (x) {}",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorObject", Line: 2, Column: 5,
			}},
		},

		// ==== Locks async function-expression predicate (predicateCannotBeAsync) ====
		{
			Code: "[1].some(async function (x) { return x > 0; });",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "predicateCannotBeAsync", Line: 1, Column: 10, EndLine: 1, EndColumn: 46,
			}},
		},

		// ==== Locks ConditionalExpression test position (separate from if-statement) ====
		{
			Code: "declare const x: string | null;\nconst y = x ? 1 : 0;",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableString", Line: 2, Column: 11,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "declare const x: string | null;\nconst y = (x != null) ? 1 : 0;"},
					{MessageId: "conditionFixDefaultEmptyString", Output: "declare const x: string | null;\nconst y = (x ?? \"\") ? 1 : 0;"},
					{MessageId: "conditionFixCastBoolean", Output: "declare const x: string | null;\nconst y = (Boolean(x)) ? 1 : 0;"},
				},
			}},
		},

		// ==== Locks DoStatement test position ====
		{
			Code: "declare const x: number | null;\ndo { /* */ } while (x);",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableNumber", Line: 2, Column: 21,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "declare const x: number | null;\ndo { /* */ } while (x != null);"},
					{MessageId: "conditionFixDefaultZero", Output: "declare const x: number | null;\ndo { /* */ } while (x ?? 0);"},
					{MessageId: "conditionFixCastBoolean", Output: "declare const x: number | null;\ndo { /* */ } while (Boolean(x));"},
				},
			}},
		},

		// ==== Locks WhileStatement test position ====
		{
			Code: "declare const x: string | null;\nwhile (x) {}",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableString", Line: 2, Column: 8,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "declare const x: string | null;\nwhile (x != null) {}"},
					{MessageId: "conditionFixDefaultEmptyString", Output: "declare const x: string | null;\nwhile (x ?? \"\") {}"},
					{MessageId: "conditionFixCastBoolean", Output: "declare const x: string | null;\nwhile (Boolean(x)) {}"},
				},
			}},
		},

		// ==== Locks ForStatement test position (condition optional in tsgo) ====
		{
			Code: "declare const x: number | null;\nfor (let i = 0; x; i++) {}",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableNumber", Line: 2, Column: 17,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "declare const x: number | null;\nfor (let i = 0; x != null; i++) {}"},
					{MessageId: "conditionFixDefaultZero", Output: "declare const x: number | null;\nfor (let i = 0; x ?? 0; i++) {}"},
					{MessageId: "conditionFixCastBoolean", Output: "declare const x: number | null;\nfor (let i = 0; Boolean(x); i++) {}"},
				},
			}},
		},

		// ==== Locks `!` operand in non-statement position ====
		{
			Code: "declare const x: number | null;\nconst y = !x;",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableNumber", Line: 2, Column: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "declare const x: number | null;\nconst y = x == null;"},
					{MessageId: "conditionFixDefaultZero", Output: "declare const x: number | null;\nconst y = !(x ?? 0);"},
					{MessageId: "conditionFixCastBoolean", Output: "declare const x: number | null;\nconst y = !Boolean(x);"},
				},
			}},
		},

		// ==== Locks parenthesized `!` operand path ====
		// tsgo preserves the paren around `(x)` but upstream's ESTree strips
		// it. `enclosingNegation` walks past the paren so the suggestion
		// fixer targets the outer `!(x)` exactly like upstream targets `!x`
		// — i.e. the conditionFixCompareNullish suggestion replaces the
		// whole negation with `x == null`, not the inner `x` with `x != null`.
		{
			Code: "declare const x: number | null;\n!(x);",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableNumber", Line: 2, Column: 3,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "declare const x: number | null;\nx == null;"},
					{MessageId: "conditionFixDefaultZero", Output: "declare const x: number | null;\n!(x ?? 0);"},
					{MessageId: "conditionFixCastBoolean", Output: "declare const x: number | null;\n!Boolean(x);"},
				},
			}},
		},

		// ==== Locks comparison: `(...nullableString, nullableString)` predicate index — when first arg is non-spread, should be evaluated ====
		// (covered indirectly by the spread invalid case in upstream; here we lock-in that spread doesn't break dispatch)

		// ==== Real-user-style: nested logical that reaches deep `nullableString` ====
		{
			Code: "declare const a: boolean;\ndeclare const s: string | null;\nif (a && (true || s)) {}",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableString", Line: 3, Column: 19,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "declare const a: boolean;\ndeclare const s: string | null;\nif (a && (true || (s != null))) {}"},
					{MessageId: "conditionFixDefaultEmptyString", Output: "declare const a: boolean;\ndeclare const s: string | null;\nif (a && (true || (s ?? \"\"))) {}"},
					{MessageId: "conditionFixCastBoolean", Output: "declare const a: boolean;\ndeclare const s: string | null;\nif (a && (true || (Boolean(s)))) {}"},
				},
			}},
		},

		// ==== Real-user: deeply nested array predicates ====
		// `.filter(x => x)` callback returns `number | null`; rule reports the
		// callback and emits the standard nullableNumber suggestion set plus
		// explicit-boolean-return-type.
		{
			Code: "declare const arr: { vals: (number | null)[] }[];\narr.forEach(o => o.vals.filter(x => x));",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableNumber",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "declare const arr: { vals: (number | null)[] }[];\narr.forEach(o => o.vals.filter(x => x != null));"},
					{MessageId: "conditionFixDefaultZero", Output: "declare const arr: { vals: (number | null)[] }[];\narr.forEach(o => o.vals.filter(x => x ?? 0));"},
					{MessageId: "conditionFixCastBoolean", Output: "declare const arr: { vals: (number | null)[] }[];\narr.forEach(o => o.vals.filter(x => Boolean(x)));"},
					{MessageId: "explicitBooleanReturnType", Output: "declare const arr: { vals: (number | null)[] }[];\narr.forEach(o => o.vals.filter((x): boolean => x));"},
				},
			}},
		},

		// ==== Real-user: BigInt comparison ====
		{
			Code:    "declare const x: bigint;\nif (x) {}",
			Options: map[string]interface{}{"allowNumber": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNumber", Line: 2, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareZero", Output: "declare const x: bigint;\nif (x !== 0) {}"},
					{MessageId: "conditionFixCompareNaN", Output: "declare const x: bigint;\nif (!Number.isNaN(x)) {}"},
					{MessageId: "conditionFixCastBoolean", Output: "declare const x: bigint;\nif (Boolean(x)) {}"},
				},
			}},
		},

		// ==== Real-user: unknown type without allowAny ====
		{
			Code: "declare const x: unknown;\nif (x) {}",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorAny", Line: 2, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCastBoolean", Output: "declare const x: unknown;\nif (Boolean(x)) {}"},
				},
			}},
		},

		// ==== TS wrappers: `as any` should still be conditionErrorAny ====
		// AsExpression is not strong-precedence, so wrappingFix wraps the inner
		// node in parens before applying `Boolean(...)` — exactly what upstream
		// `getWrappingFixer` does for the same reason.
		{
			Code: "declare const x: string;\nif (x as any) {}",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorAny", Line: 2, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCastBoolean", Output: "declare const x: string;\nif (Boolean((x as any))) {}"},
				},
			}},
		},

		// ==== TS wrapper: `(s | null)!` removes null, but `s | undefined!` removes undefined — verify on `(s | undefined)!` ====
		{
			Code:    "declare const x: number | undefined;\nif (x) {}",
			Options: map[string]interface{}{"allowNumber": true, "allowNullableNumber": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableNumber", Line: 2, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "declare const x: number | undefined;\nif (x != null) {}"},
					{MessageId: "conditionFixDefaultZero", Output: "declare const x: number | undefined;\nif (x ?? 0) {}"},
					{MessageId: "conditionFixCastBoolean", Output: "declare const x: number | undefined;\nif (Boolean(x)) {}"},
				},
			}},
		},

		// ==== void-returning function called in condition position ====
		{
			Code: "declare function f(): void;\nif (f()) {}",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullish", Line: 2, Column: 5,
			}},
		},

		// ==== Truthy primitive without nullable + allow* off (string only) ====
		{
			Code:    "if ('hi') {}",
			Options: map[string]interface{}{"allowString": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorString", Line: 1, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareStringLength", Output: "if ('hi'.length > 0) {}"},
					{MessageId: "conditionFixCompareEmptyString", Output: `if ('hi' !== "") {}`},
					{MessageId: "conditionFixCastBoolean", Output: "if (Boolean('hi')) {}"},
				},
			}},
		},

		// ==== Mixed `string | bigint` triggers Other ====
		{
			Code:    "declare const x: string | bigint;\nif (x) {}",
			Options: map[string]interface{}{"allowString": true, "allowNumber": true},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorOther", Line: 2, Column: 5,
			}},
		},

		// ==== Dimension 4 lock-in: TS NonNullAssertion on nullable input ====
		// `x!` removes null/undefined; verify `x! && rhs` continues to check rhs in condition position.
		{
			Code: "declare const x: string | null;\ndeclare const y: number | null;\nif (x! && y) {}",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableNumber", Line: 3, Column: 11,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "declare const x: string | null;\ndeclare const y: number | null;\nif (x! && (y != null)) {}"},
					{MessageId: "conditionFixDefaultZero", Output: "declare const x: string | null;\ndeclare const y: number | null;\nif (x! && (y ?? 0)) {}"},
					{MessageId: "conditionFixCastBoolean", Output: "declare const x: string | null;\ndeclare const y: number | null;\nif (x! && (Boolean(y))) {}"},
				},
			}},
		},

		// ==== Dimension 4: `as` wrapper on nullable string keeps the error ====
		{
			Code: "declare const x: unknown;\nif (x as string | null) {}",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableString", Line: 2, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "declare const x: unknown;\nif ((x as string | null) != null) {}"},
					{MessageId: "conditionFixDefaultEmptyString", Output: "declare const x: unknown;\nif ((x as string | null) ?? \"\") {}"},
					{MessageId: "conditionFixCastBoolean", Output: "declare const x: unknown;\nif (Boolean((x as string | null))) {}"},
				},
			}},
		},

		// ==== Branch lock-in: `is('nullish', 'truthy number', 'enum')` with allowNullableEnum:false ====
		{
			Code:    "\nenum E { A = 1, B = 2 }\nfunction f(e: E | null) { if (e) {} }\n",
			Options: map[string]interface{}{"allowNullableEnum": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableEnum", Line: 3, Column: 31,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "\nenum E { A = 1, B = 2 }\nfunction f(e: E | null) { if (e != null) {} }\n"},
				},
			}},
		},

		// ==== Branch lock-in: `is('nullish', 'truthy string', 'enum')` ====
		{
			Code:    "\nenum E { A = 'a', B = 'b' }\nfunction f(e: E | null) { if (e) {} }\n",
			Options: map[string]interface{}{"allowNullableEnum": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableEnum", Line: 3, Column: 31,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "\nenum E { A = 'a', B = 'b' }\nfunction f(e: E | null) { if (e != null) {} }\n"},
				},
			}},
		},

		// ==== Branch lock-in: `is('nullish', 'truthy number', 'string', 'enum')` mixed ====
		{
			Code:    "\nenum E { A = 1, B = 'b' }\nfunction f(e: E | null) { if (e) {} }\n",
			Options: map[string]interface{}{"allowNullableEnum": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableEnum", Line: 3, Column: 31,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "\nenum E { A = 1, B = 'b' }\nfunction f(e: E | null) { if (e != null) {} }\n"},
				},
			}},
		},

		// ==== Branch lock-in: `is('nullish', 'truthy string', 'number', 'enum')` mixed ====
		{
			Code:    "\nenum E { A = 'a', B = 0 }\nfunction f(e: E | null) { if (e) {} }\n",
			Options: map[string]interface{}{"allowNullableEnum": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableEnum", Line: 3, Column: 31,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "\nenum E { A = 'a', B = 0 }\nfunction f(e: E | null) { if (e != null) {} }\n"},
				},
			}},
		},

		// ==== Branch lock-in: `is('nullish', 'number', 'string', 'enum')` mixed (all general) ====
		{
			Code:    "\nenum E { A = 0, B = '' }\nfunction f(e: E | null) { if (e) {} }\n",
			Options: map[string]interface{}{"allowNullableEnum": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableEnum", Line: 3, Column: 31,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "\nenum E { A = 0, B = '' }\nfunction f(e: E | null) { if (e != null) {} }\n"},
				},
			}},
		},

		// ==== Branch lock-in: inspectVariantTypes `boolean` with single literal `false` (truthy boolean ARM false) ====
		// Single-literal boolean `false` is mapped to vt='boolean' (not 'truthy boolean'),
		// which lets determineReportType report normally on `false | null`.
		{
			Code:    "declare const x: false | null;\nif (x) {}",
			Options: map[string]interface{}{"allowNullableBoolean": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableBoolean", Line: 2, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixDefaultFalse", Output: "declare const x: false | null;\nif (x ?? false) {}"},
					{MessageId: "conditionFixCompareTrue", Output: "declare const x: false | null;\nif (x === true) {}"},
				},
			}},
		},

		// ==== Branch lock-in: `is('object')` arm when value is a class instance ====
		{
			Code: "class C {}\ndeclare const c: C;\nif (c) {}",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorObject", Line: 3, Column: 5,
			}},
		},

		// ==== Branch lock-in: void-returning method called in condition ====
		{
			Code: "class C { method(): void {} }\ndeclare const c: C;\nif (c.method()) {}",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullish", Line: 3, Column: 5,
			}},
		},

		// ==== Lock-in: ConditionalExpression nested in ConditionalExpression test ====
		{
			Code: "declare const a: string | null;\ndeclare const b: boolean;\nconst v = a ? 1 : (b ? 2 : 3);",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableString", Line: 3, Column: 11,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "declare const a: string | null;\ndeclare const b: boolean;\nconst v = (a != null) ? 1 : (b ? 2 : 3);"},
					{MessageId: "conditionFixDefaultEmptyString", Output: "declare const a: string | null;\ndeclare const b: boolean;\nconst v = (a ?? \"\") ? 1 : (b ? 2 : 3);"},
					{MessageId: "conditionFixCastBoolean", Output: "declare const a: string | null;\ndeclare const b: boolean;\nconst v = (Boolean(a)) ? 1 : (b ? 2 : 3);"},
				},
			}},
		},

		// ==== Lock-in: ConditionalExpression inside an if test position ====
		{
			Code: "declare const a: string | null;\ndeclare const b: boolean;\nif (a ? b : false) {}",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableString", Line: 3, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "declare const a: string | null;\ndeclare const b: boolean;\nif ((a != null) ? b : false) {}"},
					{MessageId: "conditionFixDefaultEmptyString", Output: "declare const a: string | null;\ndeclare const b: boolean;\nif ((a ?? \"\") ? b : false) {}"},
					{MessageId: "conditionFixCastBoolean", Output: "declare const a: string | null;\ndeclare const b: boolean;\nif ((Boolean(a)) ? b : false) {}"},
				},
			}},
		},

		// ==== Predicate body: nested array.filter with arr.length predicate ====
		{
			Code:    "declare const arr: number[][];\narr.filter(a => a.length);",
			Options: map[string]interface{}{"allowNumber": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNumber",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareArrayLengthNonzero", Output: "declare const arr: number[][];\narr.filter(a => a.length > 0);"},
					{MessageId: "explicitBooleanReturnType", Output: "declare const arr: number[][];\narr.filter((a): boolean => a.length);"},
				},
			}},
		},

		// ==== Lock-in: PrefixUnaryExpression with `!!` (double negation) does NOT fire — !x is boolean ====
		// `!!nullable` resolves to boolean. Inner `!nullable` becomes boolean → outer `!boolean` valid.
		// We assert as valid below in valid section already.

		// ==== Lock-in: Methods-as-method-shorthand in object literal ====
		{
			Code: "const o = { m() { if (true) return; } };\no.m;",
			// Inner `if (true)` is always-truthy boolean — valid; no diagnostic. Wrap as a no-op invalid (skip).
			Skip:   true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionErrorOther"}},
		},

		// ==== Lock-in: nested logical chains check correct operand ====
		{
			Code: "declare const a: object;\ndeclare const b: object;\ndeclare const c: number;\nif (a && b && c) {}",
			Errors: []rule_tester.InvalidTestCaseError{
				// `a` not checked? actually `a && (b && c)` ESTree-style — but tsgo BinaryExpression is left-assoc, so it's `(a && b) && c`.
				// IfStatement test = `(a && b) && c` → traverseLogical with isCondition=true:
				//   - left = `(a && b)` → traverseLogical(true) → check `a` & `b` as condition
				//   - right = `c` → condition (since isCondition=true)
				{MessageId: "conditionErrorObject", Line: 4, Column: 5},
				{MessageId: "conditionErrorObject", Line: 4, Column: 10},
				{
					MessageId: "conditionErrorNumber", Line: 4, Column: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "conditionFixCompareZero", Output: "declare const a: object;\ndeclare const b: object;\ndeclare const c: number;\nif (a && b && (c !== 0)) {}"},
						{MessageId: "conditionFixCompareNaN", Output: "declare const a: object;\ndeclare const b: object;\ndeclare const c: number;\nif (a && b && (!Number.isNaN(c))) {}"},
						{MessageId: "conditionFixCastBoolean", Output: "declare const a: object;\ndeclare const b: object;\ndeclare const c: number;\nif (a && b && (Boolean(c))) {}"},
					},
				},
			},
			Options: map[string]interface{}{"allowNumber": false},
		},
	})
}
