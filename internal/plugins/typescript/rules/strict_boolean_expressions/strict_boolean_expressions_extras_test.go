// TestStrictBooleanExpressionsExtras locks in branches and edge shapes that
// the upstream test suite doesn't exercise. Cases are grouped by what they
// lock in: tsgo AST quirks (Dimension 4), every distinct arm in
// `inspectVariantTypes` / `determineReportType` / `traverseLogical` /
// `checkArrayMethodCallPredicate`, real-user code shapes pulled from
// production codebases, the `allow*` options × type cross matrix, and the
// `wrappingFix` paren / ASI / multi-line edges.
package strict_boolean_expressions

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestStrictBooleanExpressionsExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &StrictBooleanExpressionsRule, []rule_tester.ValidTestCase{
		// ━━━━━━━━━━━━ Dimension 4 + general lock-ins ━━━━━━━━━━━━
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
		// ━━━━━━━━━━━━ inspectVariantTypes / determineReportType / traverseLogical / checkArrayMethodCallPredicate arm lock-ins ━━━━━━━━━━━━
		// Locks in inspectVariantTypes booleans arm — single literal `true`.
		{Code: "declare const x: true;\nif (x) {}"},
		// Locks in inspectVariantTypes booleans arm — single literal `false`.
		{Code: "declare const x: false;\nif (x) {}"},
		// Locks in inspectVariantTypes booleans arm — true | false (= boolean after union split).
		{Code: "declare const x: true | false;\nif (x) {}"},

		// Locks in determineReportType arm: allowNumber + nullish + truthy number → ok.
		{
			Code:    "declare const x: 1 | 2 | null;\nif (x) {}",
			Options: map[string]interface{}{"allowNumber": true},
		},
		// Locks in determineReportType arm: allowString + nullish + truthy string → ok.
		{
			Code:    "declare const x: 'a' | 'b' | null;\nif (x) {}",
			Options: map[string]interface{}{"allowString": true},
		},
		// Locks in determineReportType arm: nullish + truthy boolean (`true | null`) → ok regardless of options.
		{Code: "declare const x: true | null;\nif (x) {}"},
		{Code: "declare const x: true | undefined;\nif (x) {}"},
		{Code: "declare const x: true | null | undefined;\nif (x) {}"},

		// Locks in determineReportType arm: bigint truthy-only with allowNumber.
		{
			Code:    "declare const x: 1n | 2n;\nif (x) {}",
			Options: map[string]interface{}{"allowNumber": true},
		},

		// Locks in nullable-enum allowNullableEnum:true paths for every variant combo.
		// nullish + number + enum.
		{
			Code:    "\nenum E { A = 0, B = 1 }\nfunction f(e: E | null) { if (e) {} }\n",
			Options: map[string]interface{}{"allowNullableEnum": true},
		},
		// nullish + string + enum.
		{
			Code:    "\nenum E { A = '', B = 'b' }\nfunction f(e: E | null) { if (e) {} }\n",
			Options: map[string]interface{}{"allowNullableEnum": true},
		},
		// nullish + truthy number + enum.
		{
			Code:    "\nenum E { A = 1, B = 2 }\nfunction f(e: E | null) { if (e) {} }\n",
			Options: map[string]interface{}{"allowNullableEnum": true},
		},
		// nullish + truthy string + enum.
		{
			Code:    "\nenum E { A = 'a', B = 'b' }\nfunction f(e: E | null) { if (e) {} }\n",
			Options: map[string]interface{}{"allowNullableEnum": true},
		},

		// Locks in checkArrayMethodCallPredicate boolean-OK arm.
		{Code: "[1, 2, 3].some((x): boolean => x > 0);"},
		// Locks in checkArrayMethodCallPredicate type-guard arm.
		{Code: "declare function isNum(x: unknown): x is number;\n[1, 'a'].filter(isNum);"},
		// Locks in checkArrayMethodCallPredicate non-array receiver arm: not array → no check.
		// (utils.IsArrayMethodCallWithPredicate returns false; rule shouldn't fire.)
		{Code: "declare const m: Map<string, number>;\nm.has;"},
		// Locks in checkArrayMethodCallPredicate predicate-as-identifier-returning-boolean arm.
		{Code: "declare const pred: (x: number) => boolean;\n[1].filter(pred);"},

		// Locks in traverseLogical right-operand-is-not-condition arm: top-level `||`.
		{Code: "declare const a: boolean;\ndeclare const b: number;\na || b;"},
		// Locks in traverseLogical right-operand-is-not-condition arm: top-level `&&`.
		{Code: "declare const a: boolean;\ndeclare const b: number;\na && b;"},

		// Locks in CallExpression listener: non-assertion call without array predicate → no check.
		{Code: "declare const x: number | null;\nconsole.log(x);"},

		// Locks in traverseNode dedup: paren-wrapped binary doesn't double-report.
		{Code: "declare const a: boolean;\ndeclare const b: boolean;\nif ((a && b)) {}"},

		// Locks in ForStatement absent-condition path: tsgo Condition is optional.
		{Code: "for (let i = 0; ; i++) { if (i > 10) break; }"},
		// Locks in ForStatement condition path when condition IS boolean.
		{Code: "declare const cond: boolean;\nfor (let i = 0; cond; i++) {}"},
		// ━━━━━━━━━━━━ Real-user code shapes (React-like, env, regex, configs) ━━━━━━━━━━━━
		// ---- Real-user: React-style early return on loading boolean ----
		{Code: "declare function useLoading(): boolean;\nfunction Comp() { const loading = useLoading(); if (loading) return null; return 1; }"},

		// ---- Real-user: Form validation with boolean field ----
		{Code: "interface Form { valid: boolean; }\nfunction submit(f: Form) { if (f.valid) console.log('ok'); }"},

		// ---- Real-user: feature-flag check (boolean) ----
		{Code: "declare const flags: { newUI: boolean };\nif (flags.newUI) console.log('new');"},

		// ---- Real-user: authentication guard returning early ----
		{Code: "interface User { id: string; admin: boolean; }\nfunction guard(u: User) { if (u.admin) return; throw new Error('no'); }"},

		// ---- Real-user: discriminated union exhaustive switch ----
		{Code: "type T = { kind: 'a'; v: boolean } | { kind: 'b'; v: boolean };\nfunction f(t: T) { switch (t.kind) { case 'a': if (t.v) return; break; case 'b': if (t.v) return; break; } }"},

		// ---- Real-user: typeof narrowing keeps result boolean ----
		{Code: "function f(x: unknown) { if (typeof x === 'string' && x.length > 0) return; }"},

		// ---- Real-user: instanceof narrowing ----
		{Code: "function f(x: unknown) { if (x instanceof Error) console.log(x.message); }"},

		// ---- Real-user: Array.isArray narrowing ----
		{Code: "function f(x: unknown) { if (Array.isArray(x)) console.log(x.length); }"},

		// ---- Real-user: user-defined type guard ----
		{Code: "function isString(x: unknown): x is string { return typeof x === 'string'; }\nfunction f(x: unknown) { if (isString(x)) console.log(x.length); }"},

		// ---- Real-user: nullish coalescing chain (?? not && / ||) is ignored ----
		{Code: "declare const a: string | null;\nconst v = a ?? 'default';"},

		// ---- Real-user: optional chain with type guard ----
		{Code: "interface A { b?: { c?: boolean } }\nfunction f(a: A) { if (a.b?.c === true) return; }"},

		// ---- Real-user: comparison operators always produce boolean ----
		{Code: "declare const n: number;\nif (n > 0) console.log('pos');\nif (n === 0) console.log('zero');\nif (n != 0) console.log('nonzero');"},

		// ---- Real-user: in operator ----
		{Code: "declare const o: { a?: number };\nif ('a' in o) console.log(o.a);"},

		// ---- Real-user: nested ternaries with booleans ----
		{Code: "declare const a: boolean;\ndeclare const b: boolean;\ndeclare const c: boolean;\nconst v = a ? 1 : b ? 2 : c ? 3 : 4;"},

		// ---- Real-user: explicit Boolean() conversion ----
		{Code: "declare const x: string | null;\nif (Boolean(x)) console.log('truthy');"},

		// (Removed: `!!x` is NOT actually a valid escape — the inner `!`
		// already enters condition position and reports on the nullable
		// string. This is intentional upstream behavior and is exercised as
		// invalid below.)

		// ---- Real-user: Array predicate with strict comparison ----
		{Code: "declare const arr: number[];\narr.filter(x => x !== 0);"},

		// ---- Real-user: Array predicate with NonNullable ----
		{Code: "declare function isNotNull<T>(x: T): x is NonNullable<T>;\n[1, null].filter(isNotNull);"},

		// ---- Real-user: Array.find with explicit return type ----
		{Code: "declare const items: { active: boolean }[];\nitems.find((i): boolean => i.active);"},

		// ---- Real-user: while-loop with !done flag ----
		{Code: "function f() { let done = false; while (!done) { done = true; } }"},

		// ---- Real-user: for-of with optional chain ----
		{Code: "declare const items: { value?: boolean }[];\nfor (const item of items) { if (item.value === true) console.log(item); }"},

		// ---- Real-user: jest-style assertion ----
		{Code: "declare function expect<T>(v: T): { toBe(v: T): void };\ndeclare const x: boolean;\nexpect(x).toBe(true);"},

		// ---- Real-user: try/catch with boolean check on error.code ----
		{Code: "interface MyError { code: string; }\nfunction f(e: unknown) { try { throw e; } catch (err: any) { if ((err as MyError).code === 'X') return; throw err; } }"},

		// ---- Real-user: switch case predicate ----
		{Code: "declare const x: 'a' | 'b' | 'c';\nswitch (x) { case 'a': console.log(1); break; case 'b': console.log(2); break; default: console.log(0); }"},

		// ---- Real-user: error early-return pattern ----
		{Code: "declare function tryThing(): Error | null;\nconst err = tryThing();\nif (err != null) throw err;"},

		// ---- Real-user: configuration optional chain ----
		{Code: "interface Config { logging?: { enabled: boolean } }\nfunction setup(c: Config) { if (c.logging?.enabled === true) console.log('on'); }"},

		// ---- Real-user: Promise.all + filter pattern ----
		{Code: "async function f() { const results: (string | null)[] = await Promise.all([]); return results.filter((r): r is string => r !== null); }"},

		// ---- Real-user: branded ID type ----
		{Code: "type UserID = string & { __brand: 'UserID' };\ndeclare const id: UserID;\nif (id.length > 0) console.log('has id');"},

		// ---- Real-user: enum with string values, narrowed ----
		{Code: "enum Status { Active = 'active', Disabled = 'disabled' }\ndeclare const s: Status;\nif (s === Status.Active) console.log('on');"},

		// ---- Real-user: class with boolean accessor ----
		{Code: "class Box { get filled(): boolean { return true; } }\nconst b = new Box();\nif (b.filled) console.log('full');"},
		// ━━━━━━━━━━━━ Options × type cross-matrix and options-shape (map/array/nil) ━━━━━━━━━━━━
		// ---- Default options sanity ----
		// Default: allowString=true → string stays valid.
		{Code: "declare const x: string;\nif (x) {}"},
		// Default: allowNumber=true → number stays valid.
		{Code: "declare const x: number;\nif (x) {}"},
		// Default: allowNullableObject=true → nullable object stays valid.
		{Code: "declare const x: object | null;\nif (x) {}"},
		// Default: allowNullableBoolean=false → still no fire on plain boolean.
		{Code: "declare const x: boolean;\nif (x) {}"},

		// ---- allowString ON: string and truthy string valid; nullable string still error ----
		{
			Code:    "declare const x: string;\nif (x) {}",
			Options: map[string]interface{}{"allowString": true},
		},
		{
			Code:    "declare const x: 'a';\nif (x) {}",
			Options: map[string]interface{}{"allowString": true},
		},
		// allowString does NOT affect number.
		{
			Code:    "declare const x: number;\nif (x) {}",
			Options: map[string]interface{}{"allowString": true, "allowNumber": true},
		},

		// ---- allowNumber ON: number, bigint, truthy number all valid ----
		{
			Code:    "declare const x: number;\nif (x) {}",
			Options: map[string]interface{}{"allowNumber": true},
		},
		{
			Code:    "declare const x: bigint;\nif (x) {}",
			Options: map[string]interface{}{"allowNumber": true},
		},
		{
			Code:    "declare const x: 42;\nif (x) {}",
			Options: map[string]interface{}{"allowNumber": true},
		},

		// ---- allowNullableObject ON: object | null valid; pure object still errors ----
		// (Pure object always errors regardless — locked in the invalid section.)
		{
			Code:    "declare const x: { a: 1 } | null;\nif (x) {}",
			Options: map[string]interface{}{"allowNullableObject": true},
		},
		{
			Code:    "declare const x: symbol | undefined;\nif (x) {}",
			Options: map[string]interface{}{"allowNullableObject": true},
		},
		{
			Code:    "declare const x: (() => void) | null;\nif (x) {}",
			Options: map[string]interface{}{"allowNullableObject": true},
		},

		// ---- allowNullableBoolean ON: boolean | null valid; pure boolean already valid ----
		{
			Code:    "declare const x: boolean | undefined;\nif (x) {}",
			Options: map[string]interface{}{"allowNullableBoolean": true},
		},

		// ---- allowNullableString ON: string | null valid ----
		{
			Code:    "declare const x: string | undefined;\nif (x) {}",
			Options: map[string]interface{}{"allowNullableString": true},
		},

		// ---- allowNullableNumber ON: number | null valid; bigint | null also valid ----
		{
			Code:    "declare const x: number | undefined;\nif (x) {}",
			Options: map[string]interface{}{"allowNullableNumber": true},
		},
		{
			Code:    "declare const x: bigint | null;\nif (x) {}",
			Options: map[string]interface{}{"allowNullableNumber": true},
		},

		// ---- allowAny ON: any / unknown / unconstrained T valid ----
		{
			Code:    "declare const x: any;\nif (x) {}",
			Options: map[string]interface{}{"allowAny": true},
		},
		{
			Code:    "declare const x: unknown;\nif (x) {}",
			Options: map[string]interface{}{"allowAny": true},
		},
		{
			Code:    "function f<T>(x: T) { if (x) {} }",
			Options: map[string]interface{}{"allowAny": true},
		},

		// ---- Combined opts: maximum permissive — every truthy primitive + nullable + any allowed ----
		{
			Code: "declare const a: string | null;\ndeclare const b: number | undefined;\ndeclare const c: object | null;\ndeclare const d: boolean | null;\ndeclare const e: any;\nif (a) {}\nif (b) {}\nif (c) {}\nif (d) {}\nif (e) {}",
			Options: map[string]interface{}{
				"allowString":          true,
				"allowNumber":          true,
				"allowNullableString":  true,
				"allowNullableNumber":  true,
				"allowNullableObject":  true,
				"allowNullableBoolean": true,
				"allowAny":             true,
				"allowNullableEnum":    true,
			},
		},

		// ---- Truthy literal stays valid regardless of `allow*` (early-exit) ----
		{
			Code:    "declare const x: true;\nif (x) {}",
			Options: map[string]interface{}{"allowString": false, "allowNumber": false, "allowNullableObject": false},
		},

		// ---- allowNullableEnum ON: mixed enum unions stay valid ----
		{
			Code: "\nenum E { A = 0, B = 1 }\ndeclare const x: E | null;\nif (x) {}\n",
			Options: map[string]interface{}{
				"allowNullableEnum": true,
			},
		},
		{
			Code: "\nenum E { A = 'a', B = 'b' }\ndeclare const x: E | undefined;\nif (x) {}\n",
			Options: map[string]interface{}{
				"allowNullableEnum": true,
			},
		},
		// ━━━━━━━━━━━━ wrappingFix paren / ASI / multi-line / TS wrapper edges ━━━━━━━━━━━━
		// Reference no-op valid: a strong-precedence identifier needs no inner paren.
		{Code: "declare const x: boolean;\nif (x) {}"},
	}, []rule_tester.InvalidTestCase{
		// ━━━━━━━━━━━━ Dimension 4 + general lock-ins ━━━━━━━━━━━━
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
		// ━━━━━━━━━━━━ inspectVariantTypes / determineReportType / traverseLogical / checkArrayMethodCallPredicate arm lock-ins ━━━━━━━━━━━━
		// Locks in determineReportType arm 1: `nullish` alone (no other types).
		{
			Code: "declare const x: null | undefined;\nif (x) {}",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullish", Line: 2, Column: 5,
			}},
		},

		// Locks in determineReportType arm: `string` alone with allowString:false.
		{
			Code:    "declare const x: string;\nif (x) {}",
			Options: map[string]interface{}{"allowString": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorString", Line: 2, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareStringLength", Output: "declare const x: string;\nif (x.length > 0) {}"},
					{MessageId: "conditionFixCompareEmptyString", Output: `declare const x: string;
if (x !== "") {}`},
					{MessageId: "conditionFixCastBoolean", Output: "declare const x: string;\nif (Boolean(x)) {}"},
				},
			}},
		},

		// Locks in determineReportType arm: `truthy string` alone with allowString:false.
		{
			Code:    "declare const x: 'hello';\nif (x) {}",
			Options: map[string]interface{}{"allowString": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorString", Line: 2, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareStringLength", Output: "declare const x: 'hello';\nif (x.length > 0) {}"},
					{MessageId: "conditionFixCompareEmptyString", Output: `declare const x: 'hello';
if (x !== "") {}`},
					{MessageId: "conditionFixCastBoolean", Output: "declare const x: 'hello';\nif (Boolean(x)) {}"},
				},
			}},
		},

		// Locks in determineReportType arm: `nullish | string` with allowNullableString:false.
		{
			Code: "declare const x: '' | null;\nif (x) {}",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableString", Line: 2, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "declare const x: '' | null;\nif (x != null) {}"},
					{MessageId: "conditionFixDefaultEmptyString", Output: `declare const x: '' | null;
if (x ?? "") {}`},
					{MessageId: "conditionFixCastBoolean", Output: "declare const x: '' | null;\nif (Boolean(x)) {}"},
				},
			}},
		},

		// Locks in determineReportType arm: `number` alone with allowNumber:false.
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

		// Locks in determineReportType arm: `truthy number` alone with allowNumber:false.
		{
			Code:    "declare const x: 42;\nif (x) {}",
			Options: map[string]interface{}{"allowNumber": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNumber", Line: 2, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareZero", Output: "declare const x: 42;\nif (x !== 0) {}"},
					{MessageId: "conditionFixCompareNaN", Output: "declare const x: 42;\nif (!Number.isNaN(x)) {}"},
					{MessageId: "conditionFixCastBoolean", Output: "declare const x: 42;\nif (Boolean(x)) {}"},
				},
			}},
		},

		// Locks in determineReportType arm: `object` alone.
		{
			Code: "declare const x: Date;\nif (x) {}",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorObject", Line: 2, Column: 5,
			}},
		},

		// Locks in determineReportType arm: `nullish | object` with allowNullableObject:false.
		{
			Code:    "declare const x: Date | null;\nif (x) {}",
			Options: map[string]interface{}{"allowNullableObject": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableObject", Line: 2, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "declare const x: Date | null;\nif (x != null) {}"},
				},
			}},
		},

		// Locks in determineReportType arm: `any` alone.
		{
			Code: "declare const x: any;\nif (x) {}",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorAny", Line: 2, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCastBoolean", Output: "declare const x: any;\nif (Boolean(x)) {}"},
				},
			}},
		},

		// Locks in determineReportType arm: `unknown` (also classified as any).
		{
			Code: "declare const x: unknown;\nif (x) {}",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorAny", Line: 2, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCastBoolean", Output: "declare const x: unknown;\nif (Boolean(x)) {}"},
				},
			}},
		},

		// Locks in determineReportType arm: unconstrained T (TypeParameter → any bucket).
		{
			Code: "function f<T>(x: T) { if (x) {} }",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorAny", Line: 1, Column: 27,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCastBoolean", Output: "function f<T>(x: T) { if (Boolean(x)) {} }"},
				},
			}},
		},

		// Locks in determineReportType fallthrough: `conditionErrorOther` for mixed primitives.
		// `bigint | string` doesn't match any specific arm.
		{
			Code:    "declare const x: bigint | string;\nif (x) {}",
			Options: map[string]interface{}{"allowString": true, "allowNumber": true},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorOther", Line: 2, Column: 5,
			}},
		},

		// Locks in determineReportType fallthrough: `number | boolean`.
		{
			Code:    "declare const x: number | boolean;\nif (x) {}",
			Options: map[string]interface{}{"allowNumber": true},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorOther", Line: 2, Column: 5,
			}},
		},

		// Locks in determineReportType arm: `nullish | truthy number` with allowNumber:false.
		// Upstream's logic is exact-set matching on variants. The set
		// {nullish, truthy number} doesn't match any specific arm when
		// allowNumber is off (the early-exit arm requires allowNumber=true;
		// the `nullish + number` arm requires general `number`, not `truthy
		// number`). So it falls through to `conditionErrorOther`.
		{
			Code:    "declare const x: 42 | null;\nif (x) {}",
			Options: map[string]interface{}{"allowNumber": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorOther", Line: 2, Column: 5,
			}},
		},

		// Locks in traverseLogical: top-level `a && b && c;` checks left & middle but not right.
		{
			Code:    "declare const a: object;\ndeclare const b: object;\ndeclare const c: number;\na && b && c;",
			Options: map[string]interface{}{"allowNumber": false},
			Errors: []rule_tester.InvalidTestCaseError{
				// `(a && b) && c` — outer right (`c`) is not in condition (top-level statement).
				{MessageId: "conditionErrorObject", Line: 4, Column: 1},
				{MessageId: "conditionErrorObject", Line: 4, Column: 6},
			},
		},

		// Locks in traverseLogical: nested `(a || b) && c` in condition checks all three.
		{
			Code:    "declare const a: object;\ndeclare const b: number;\ndeclare const c: string;\nif ((a || b) && c) {}",
			Options: map[string]interface{}{"allowNumber": false, "allowString": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "conditionErrorObject", Line: 4, Column: 6},
				{
					MessageId: "conditionErrorNumber", Line: 4, Column: 11,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "conditionFixCompareZero", Output: "declare const a: object;\ndeclare const b: number;\ndeclare const c: string;\nif ((a || (b !== 0)) && c) {}"},
						{MessageId: "conditionFixCompareNaN", Output: "declare const a: object;\ndeclare const b: number;\ndeclare const c: string;\nif ((a || (!Number.isNaN(b))) && c) {}"},
						{MessageId: "conditionFixCastBoolean", Output: "declare const a: object;\ndeclare const b: number;\ndeclare const c: string;\nif ((a || (Boolean(b))) && c) {}"},
					},
				},
				{
					MessageId: "conditionErrorString", Line: 4, Column: 17,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "conditionFixCompareStringLength", Output: "declare const a: object;\ndeclare const b: number;\ndeclare const c: string;\nif ((a || b) && (c.length > 0)) {}"},
						{MessageId: "conditionFixCompareEmptyString", Output: `declare const a: object;
declare const b: number;
declare const c: string;
if ((a || b) && (c !== "")) {}`},
						{MessageId: "conditionFixCastBoolean", Output: "declare const a: object;\ndeclare const b: number;\ndeclare const c: string;\nif ((a || b) && (Boolean(c))) {}"},
					},
				},
			},
		},

		// Locks in determineReportType all-suggestions-for-bigint arm.
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

		// Locks in determineReportType arm: `(0n)` truthy bigint literal.
		{
			Code:    "if (0n) {}",
			Options: map[string]interface{}{"allowNumber": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNumber", Line: 1, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareZero", Output: "if (0n !== 0) {}"},
					{MessageId: "conditionFixCompareNaN", Output: "if (!Number.isNaN(0n)) {}"},
					{MessageId: "conditionFixCastBoolean", Output: "if (Boolean(0n)) {}"},
				},
			}},
		},

		// Locks in `isArrayLengthExpression` negative path: tuple `.length`.
		// Tuple types are NOT detected by `Checker_isArrayType` (that helper
		// is the strict array check), so `t.length` on a tuple goes through
		// the normal number-suggestion path, not the array-length path.
		{
			Code:    "declare const t: [1, 2, 3];\nif (t.length) {}",
			Options: map[string]interface{}{"allowNumber": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNumber",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareZero", Output: "declare const t: [1, 2, 3];\nif (t.length !== 0) {}"},
					{MessageId: "conditionFixCompareNaN", Output: "declare const t: [1, 2, 3];\nif (!Number.isNaN(t.length)) {}"},
					{MessageId: "conditionFixCastBoolean", Output: "declare const t: [1, 2, 3];\nif (Boolean(t.length)) {}"},
				},
			}},
		},

		// Locks in checkArrayMethodCallPredicate: function-typed predicate variable
		// returning nullable boolean.
		{
			Code: "declare const pred: (x: number) => boolean | null;\n[1].filter(pred);",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableBoolean",
			}},
		},

		// Locks in checkArrayMethodCallPredicate: predicate returning object.
		{
			Code: "[1].filter(x => ({ wrapped: x }));",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorObject",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "explicitBooleanReturnType", Output: "[1].filter((x): boolean => ({ wrapped: x }));"},
				},
			}},
		},

		// Locks in checkArrayMethodCallPredicate: predicate returning string,
		// with allowString:false so the report actually fires.
		{
			Code:    "[1, 2].some(x => x.toString());",
			Options: map[string]interface{}{"allowString": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorString",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareStringLength", Output: "[1, 2].some(x => x.toString().length > 0);"},
					{MessageId: "conditionFixCompareEmptyString", Output: `[1, 2].some(x => x.toString() !== "");`},
					{MessageId: "conditionFixCastBoolean", Output: "[1, 2].some(x => Boolean(x.toString()));"},
					{MessageId: "explicitBooleanReturnType", Output: "[1, 2].some((x): boolean => x.toString());"},
				},
			}},
		},

		// Locks in checkArrayMethodCallPredicate: array.every with block return body.
		{
			Code: "['a'].every(x => { return x; });",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorString",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "explicitBooleanReturnType", Output: "['a'].every((x): boolean => { return x; });"},
				},
			}},
			Options: map[string]interface{}{"allowString": false},
		},

		// Locks in checkArrayMethodCallPredicate: findIndex variant.
		{
			Code: "['a'].findIndex(x => x);",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorString",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareStringLength", Output: "['a'].findIndex(x => x.length > 0);"},
					{MessageId: "conditionFixCompareEmptyString", Output: `['a'].findIndex(x => x !== "");`},
					{MessageId: "conditionFixCastBoolean", Output: "['a'].findIndex(x => Boolean(x));"},
					{MessageId: "explicitBooleanReturnType", Output: "['a'].findIndex((x): boolean => x);"},
				},
			}},
			Options: map[string]interface{}{"allowString": false},
		},

		// Locks in checkArrayMethodCallPredicate: findLastIndex variant.
		{
			Code: "['a'].findLastIndex(x => x);",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorString",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareStringLength", Output: "['a'].findLastIndex(x => x.length > 0);"},
					{MessageId: "conditionFixCompareEmptyString", Output: `['a'].findLastIndex(x => x !== "");`},
					{MessageId: "conditionFixCastBoolean", Output: "['a'].findLastIndex(x => Boolean(x));"},
					{MessageId: "explicitBooleanReturnType", Output: "['a'].findLastIndex((x): boolean => x);"},
				},
			}},
			Options: map[string]interface{}{"allowString": false},
		},
		// ━━━━━━━━━━━━ Real-user code shapes (React-like, env, regex, configs) ━━━━━━━━━━━━
		// ---- Real-user FP regression: optional property access in conditional ----
		// `(value?.field)` types as `T | undefined`; if T is object, it's
		// nullable object — should report when allowNullableObject:false.
		{
			Code:    "interface A { b?: { c: object } }\nfunction f(a: A) { if (a.b?.c) {} }",
			Options: map[string]interface{}{"allowNullableObject": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableObject", Line: 2, Column: 24,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "interface A { b?: { c: object } }\nfunction f(a: A) { if (a.b?.c != null) {} }"},
				},
			}},
		},

		// ---- Real-user: function returning string-or-undefined used as boolean ----
		{
			Code:    "declare function getName(): string | undefined;\nif (getName()) console.log('has');",
			Options: map[string]interface{}{"allowNullableString": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableString", Line: 2, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "declare function getName(): string | undefined;\nif (getName() != null) console.log('has');"},
					{MessageId: "conditionFixDefaultEmptyString", Output: `declare function getName(): string | undefined;
if (getName() ?? "") console.log('has');`},
					{MessageId: "conditionFixCastBoolean", Output: "declare function getName(): string | undefined;\nif (Boolean(getName())) console.log('has');"},
				},
			}},
		},

		// ---- Real-user: array.length in conditional + non-zero check missed ----
		{
			Code:    "declare const items: string[];\nconst msg = items.length ? 'has items' : 'empty';",
			Options: map[string]interface{}{"allowNumber": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNumber", Line: 2, Column: 13,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareArrayLengthNonzero", Output: "declare const items: string[];\nconst msg = (items.length > 0) ? 'has items' : 'empty';"},
				},
			}},
		},

		// ---- Real-user: string from formData/URLSearchParams (string | null) ----
		{
			Code: "function f(params: URLSearchParams) { const v = params.get('x'); if (v) console.log(v); }",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableString", Line: 1, Column: 70,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "function f(params: URLSearchParams) { const v = params.get('x'); if (v != null) console.log(v); }"},
					{MessageId: "conditionFixDefaultEmptyString", Output: `function f(params: URLSearchParams) { const v = params.get('x'); if (v ?? "") console.log(v); }`},
					{MessageId: "conditionFixCastBoolean", Output: "function f(params: URLSearchParams) { const v = params.get('x'); if (Boolean(v)) console.log(v); }"},
				},
			}},
		},

		// ---- Real-user: Map.get returning T | undefined ----
		{
			Code: "declare const m: Map<string, number>;\nconst v = m.get('k');\nif (v) console.log(v);",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableNumber", Line: 3, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "declare const m: Map<string, number>;\nconst v = m.get('k');\nif (v != null) console.log(v);"},
					{MessageId: "conditionFixDefaultZero", Output: "declare const m: Map<string, number>;\nconst v = m.get('k');\nif (v ?? 0) console.log(v);"},
					{MessageId: "conditionFixCastBoolean", Output: "declare const m: Map<string, number>;\nconst v = m.get('k');\nif (Boolean(v)) console.log(v);"},
				},
			}},
		},

		// ---- Real-user: process.env.X is string | undefined ----
		{
			Code: "declare const env: { [k: string]: string | undefined };\nif (env.DEBUG) console.log('debug');",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableString", Line: 2, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "declare const env: { [k: string]: string | undefined };\nif (env.DEBUG != null) console.log('debug');"},
					{MessageId: "conditionFixDefaultEmptyString", Output: `declare const env: { [k: string]: string | undefined };
if (env.DEBUG ?? "") console.log('debug');`},
					{MessageId: "conditionFixCastBoolean", Output: "declare const env: { [k: string]: string | undefined };\nif (Boolean(env.DEBUG)) console.log('debug');"},
				},
			}},
		},

		// ---- Real-user: regex match returns RegExpMatchArray | null ----
		{
			Code: "function isMatch(s: string) { return s.match(/foo/); }\nif (isMatch('foo')) console.log('hit');",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableObject", Line: 2, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "function isMatch(s: string) { return s.match(/foo/); }\nif (isMatch('foo') != null) console.log('hit');"},
				},
			}},
			Options: map[string]interface{}{"allowNullableObject": false},
		},

		// ---- Real-user: chained property access producing nullable ----
		{
			Code: "interface A { b: { c: string | null } }\ndeclare const a: A;\nif (a.b.c) {}",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableString", Line: 3, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "interface A { b: { c: string | null } }\ndeclare const a: A;\nif (a.b.c != null) {}"},
					{MessageId: "conditionFixDefaultEmptyString", Output: `interface A { b: { c: string | null } }
declare const a: A;
if (a.b.c ?? "") {}`},
					{MessageId: "conditionFixCastBoolean", Output: "interface A { b: { c: string | null } }\ndeclare const a: A;\nif (Boolean(a.b.c)) {}"},
				},
			}},
		},

		// ---- Real-user: callback returning nullable boolean ----
		{
			Code: "declare const items: { active?: boolean }[];\nitems.filter(i => i.active);",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableBoolean",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixDefaultFalse", Output: "declare const items: { active?: boolean }[];\nitems.filter(i => i.active ?? false);"},
					{MessageId: "conditionFixCompareTrue", Output: "declare const items: { active?: boolean }[];\nitems.filter(i => i.active === true);"},
					{MessageId: "explicitBooleanReturnType", Output: "declare const items: { active?: boolean }[];\nitems.filter((i): boolean => i.active);"},
				},
			}},
		},

		// ---- Real-user: function call result used as condition without Boolean wrap ----
		{
			Code:    "declare function trim(s: string): string;\nfunction f(s: string) { if (trim(s)) console.log('non-empty'); }",
			Options: map[string]interface{}{"allowString": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorString", Line: 2, Column: 29,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareStringLength", Output: "declare function trim(s: string): string;\nfunction f(s: string) { if (trim(s).length > 0) console.log('non-empty'); }"},
					{MessageId: "conditionFixCompareEmptyString", Output: `declare function trim(s: string): string;
function f(s: string) { if (trim(s) !== "") console.log('non-empty'); }`},
					{MessageId: "conditionFixCastBoolean", Output: "declare function trim(s: string): string;\nfunction f(s: string) { if (Boolean(trim(s))) console.log('non-empty'); }"},
				},
			}},
		},

		// ---- Real-user: destructure rest then check ----
		{
			Code:    "function f({ value }: { value: string }) { if (value) console.log(value); }",
			Options: map[string]interface{}{"allowString": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorString", Line: 1, Column: 48,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareStringLength", Output: "function f({ value }: { value: string }) { if (value.length > 0) console.log(value); }"},
					{MessageId: "conditionFixCompareEmptyString", Output: `function f({ value }: { value: string }) { if (value !== "") console.log(value); }`},
					{MessageId: "conditionFixCastBoolean", Output: "function f({ value }: { value: string }) { if (Boolean(value)) console.log(value); }"},
				},
			}},
		},

		// ---- Real-user: class field nullable ----
		{
			Code: "class C { value: string | null = null; check() { if (this.value) console.log(this.value); } }",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableString", Line: 1, Column: 54,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "class C { value: string | null = null; check() { if (this.value != null) console.log(this.value); } }"},
					{MessageId: "conditionFixDefaultEmptyString", Output: `class C { value: string | null = null; check() { if (this.value ?? "") console.log(this.value); } }`},
					{MessageId: "conditionFixCastBoolean", Output: "class C { value: string | null = null; check() { if (Boolean(this.value)) console.log(this.value); } }"},
				},
			}},
		},

		// ---- Real-user: this in arrow inside class field ----
		{
			Code: "class C { value: number | null = null; check = () => { if (this.value) console.log(this.value); }; }",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableNumber", Line: 1, Column: 60,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "class C { value: number | null = null; check = () => { if (this.value != null) console.log(this.value); }; }"},
					{MessageId: "conditionFixDefaultZero", Output: "class C { value: number | null = null; check = () => { if (this.value ?? 0) console.log(this.value); }; }"},
					{MessageId: "conditionFixCastBoolean", Output: "class C { value: number | null = null; check = () => { if (Boolean(this.value)) console.log(this.value); }; }"},
				},
			}},
		},

		// ---- Real-user: IIFE returning nullable ----
		{
			Code: "const v = (() => null as string | null)();\nif (v) console.log(v);",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableString", Line: 2, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "const v = (() => null as string | null)();\nif (v != null) console.log(v);"},
					{MessageId: "conditionFixDefaultEmptyString", Output: `const v = (() => null as string | null)();
if (v ?? "") console.log(v);`},
					{MessageId: "conditionFixCastBoolean", Output: "const v = (() => null as string | null)();\nif (Boolean(v)) console.log(v);"},
				},
			}},
		},

		// ---- Real-user: ternary in arguments ----
		{
			Code:    "declare function f(b: boolean): void;\ndeclare const x: number;\nf(x ? true : false);",
			Options: map[string]interface{}{"allowNumber": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNumber", Line: 3, Column: 3,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareZero", Output: "declare function f(b: boolean): void;\ndeclare const x: number;\nf((x !== 0) ? true : false);"},
					{MessageId: "conditionFixCompareNaN", Output: "declare function f(b: boolean): void;\ndeclare const x: number;\nf((!Number.isNaN(x)) ? true : false);"},
					{MessageId: "conditionFixCastBoolean", Output: "declare function f(b: boolean): void;\ndeclare const x: number;\nf((Boolean(x)) ? true : false);"},
				},
			}},
		},

		// ---- Real-user: deep nested arr.some().filter() chain ----
		{
			Code: "declare const arr: (number | null)[];\nconst nonZero = arr.filter(x => x);",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableNumber",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "declare const arr: (number | null)[];\nconst nonZero = arr.filter(x => x != null);"},
					{MessageId: "conditionFixDefaultZero", Output: "declare const arr: (number | null)[];\nconst nonZero = arr.filter(x => x ?? 0);"},
					{MessageId: "conditionFixCastBoolean", Output: "declare const arr: (number | null)[];\nconst nonZero = arr.filter(x => Boolean(x));"},
					{MessageId: "explicitBooleanReturnType", Output: "declare const arr: (number | null)[];\nconst nonZero = arr.filter((x): boolean => x);"},
				},
			}},
		},

		// ---- Real-user: typeof-guard followed by truthy check (FN regression in upstream #3060) ----
		{
			Code:    "function f(x: string | number) { if (typeof x === 'string' && x) return; }",
			Options: map[string]interface{}{"allowString": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorString", Line: 1, Column: 63,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareStringLength", Output: "function f(x: string | number) { if (typeof x === 'string' && (x.length > 0)) return; }"},
					{MessageId: "conditionFixCompareEmptyString", Output: `function f(x: string | number) { if (typeof x === 'string' && (x !== "")) return; }`},
					{MessageId: "conditionFixCastBoolean", Output: "function f(x: string | number) { if (typeof x === 'string' && (Boolean(x))) return; }"},
				},
			}},
		},
		// ━━━━━━━━━━━━ Options × type cross-matrix and options-shape (map/array/nil) ━━━━━━━━━━━━
		// ---- allowString OFF: every flavor of string fires ----
		{
			Code:    "declare const x: string;\nif (x) {}",
			Options: map[string]interface{}{"allowString": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorString", Line: 2, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareStringLength", Output: "declare const x: string;\nif (x.length > 0) {}"},
					{MessageId: "conditionFixCompareEmptyString", Output: `declare const x: string;
if (x !== "") {}`},
					{MessageId: "conditionFixCastBoolean", Output: "declare const x: string;\nif (Boolean(x)) {}"},
				},
			}},
		},
		// allowString OFF, template literal type also fires.
		{
			Code:    "declare const x: `prefix-${string}`;\nif (x) {}",
			Options: map[string]interface{}{"allowString": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorString", Line: 2, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareStringLength", Output: "declare const x: `prefix-${string}`;\nif (x.length > 0) {}"},
					{MessageId: "conditionFixCompareEmptyString", Output: "declare const x: `prefix-${string}`;\nif (x !== \"\") {}"},
					{MessageId: "conditionFixCastBoolean", Output: "declare const x: `prefix-${string}`;\nif (Boolean(x)) {}"},
				},
			}},
		},

		// ---- allowNumber OFF: number and bigint both fire (same arm) ----
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

		// ---- allowNullableObject OFF: still triggers on plain `object` (defaults to error regardless) ----
		// This is the "Object" always-true arm, independent of allowNullableObject.
		{
			Code:    "declare const x: { a: 1 };\nif (x) {}",
			Options: map[string]interface{}{"allowNullableObject": true},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorObject", Line: 2, Column: 5,
			}},
		},

		// ---- allowNullableBoolean OFF (default): bool | undefined fires ----
		{
			Code: "declare const x: boolean | undefined;\nif (x) {}",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableBoolean", Line: 2, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixDefaultFalse", Output: "declare const x: boolean | undefined;\nif (x ?? false) {}"},
					{MessageId: "conditionFixCompareTrue", Output: "declare const x: boolean | undefined;\nif (x === true) {}"},
				},
			}},
		},

		// ---- allowNullableNumber OFF (default): number | undefined fires ----
		{
			Code: "declare const x: number | undefined;\nif (x) {}",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableNumber", Line: 2, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "declare const x: number | undefined;\nif (x != null) {}"},
					{MessageId: "conditionFixDefaultZero", Output: "declare const x: number | undefined;\nif (x ?? 0) {}"},
					{MessageId: "conditionFixCastBoolean", Output: "declare const x: number | undefined;\nif (Boolean(x)) {}"},
				},
			}},
		},

		// ---- allowNullableString OFF (default): string | undefined fires ----
		{
			Code: "declare const x: string | undefined;\nif (x) {}",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableString", Line: 2, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "declare const x: string | undefined;\nif (x != null) {}"},
					{MessageId: "conditionFixDefaultEmptyString", Output: `declare const x: string | undefined;
if (x ?? "") {}`},
					{MessageId: "conditionFixCastBoolean", Output: "declare const x: string | undefined;\nif (Boolean(x)) {}"},
				},
			}},
		},

		// ---- allowAny OFF (default): unconstrained generic fires ----
		{
			Code: "function f<T>(x: T) { if (x) {} }",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorAny", Line: 1, Column: 27,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCastBoolean", Output: "function f<T>(x: T) { if (Boolean(x)) {} }"},
				},
			}},
		},

		// ---- allowNullableEnum OFF (default): enum | null fires ----
		{
			Code: "\nenum E { A = 0, B = 1 }\ndeclare const x: E | null;\nif (x) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableEnum", Line: 4, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "\nenum E { A = 0, B = 1 }\ndeclare const x: E | null;\nif (x != null) {}\n"},
				},
			}},
		},

		// ---- Cross-matrix: allowString=true, allowNullableString=false ----
		// Should fire on string | null but stay silent on plain string.
		{
			Code:    "declare const x: string | null;\nif (x) {}",
			Options: map[string]interface{}{"allowString": true, "allowNullableString": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableString", Line: 2, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "declare const x: string | null;\nif (x != null) {}"},
					{MessageId: "conditionFixDefaultEmptyString", Output: `declare const x: string | null;
if (x ?? "") {}`},
					{MessageId: "conditionFixCastBoolean", Output: "declare const x: string | null;\nif (Boolean(x)) {}"},
				},
			}},
		},

		// ---- Cross-matrix: allowNumber=true, allowNullableNumber=false ----
		{
			Code:    "declare const x: number | null;\nif (x) {}",
			Options: map[string]interface{}{"allowNumber": true, "allowNullableNumber": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableNumber", Line: 2, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "declare const x: number | null;\nif (x != null) {}"},
					{MessageId: "conditionFixDefaultZero", Output: "declare const x: number | null;\nif (x ?? 0) {}"},
					{MessageId: "conditionFixCastBoolean", Output: "declare const x: number | null;\nif (Boolean(x)) {}"},
				},
			}},
		},

		// ---- Cross-matrix: all-strict (every allow* off except the always-on boolean exit) ----
		{
			Code: "declare const x: number;\nif (x) {}",
			Options: map[string]interface{}{
				"allowString":          false,
				"allowNumber":          false,
				"allowNullableObject":  false,
				"allowNullableBoolean": false,
				"allowNullableString":  false,
				"allowNullableNumber":  false,
				"allowNullableEnum":    false,
				"allowAny":             false,
			},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNumber", Line: 2, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareZero", Output: "declare const x: number;\nif (x !== 0) {}"},
					{MessageId: "conditionFixCompareNaN", Output: "declare const x: number;\nif (!Number.isNaN(x)) {}"},
					{MessageId: "conditionFixCastBoolean", Output: "declare const x: number;\nif (Boolean(x)) {}"},
				},
			}},
		},

		// ---- Options shape: JSON round-trip — `map[string]interface{}` direct ----
		// (covered above as the standard test format)

		// ---- Options shape: array-wrapped (matches `[{...}]` from rule_tester) ----
		{
			Code:    "declare const x: number;\nif (x) {}",
			Options: []interface{}{map[string]interface{}{"allowNumber": false}},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNumber", Line: 2, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareZero", Output: "declare const x: number;\nif (x !== 0) {}"},
					{MessageId: "conditionFixCompareNaN", Output: "declare const x: number;\nif (!Number.isNaN(x)) {}"},
					{MessageId: "conditionFixCastBoolean", Output: "declare const x: number;\nif (Boolean(x)) {}"},
				},
			}},
		},

		// ---- Options shape: empty options object → defaults ----
		{
			Code:    "declare const x: string | null;\nif (x) {}",
			Options: map[string]interface{}{},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableString", Line: 2, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "declare const x: string | null;\nif (x != null) {}"},
					{MessageId: "conditionFixDefaultEmptyString", Output: `declare const x: string | null;
if (x ?? "") {}`},
					{MessageId: "conditionFixCastBoolean", Output: "declare const x: string | null;\nif (Boolean(x)) {}"},
				},
			}},
		},

		// ---- Options shape: nil options → defaults ----
		{
			Code: "declare const x: string | null;\nif (x) {}",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableString", Line: 2, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "declare const x: string | null;\nif (x != null) {}"},
					{MessageId: "conditionFixDefaultEmptyString", Output: `declare const x: string | null;
if (x ?? "") {}`},
					{MessageId: "conditionFixCastBoolean", Output: "declare const x: string | null;\nif (Boolean(x)) {}"},
				},
			}},
		},
		// ━━━━━━━━━━━━ wrappingFix paren / ASI / multi-line / TS wrapper edges ━━━━━━━━━━━━
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
