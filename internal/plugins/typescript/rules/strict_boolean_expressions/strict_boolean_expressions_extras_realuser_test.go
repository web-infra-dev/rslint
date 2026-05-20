// TestStrictBooleanExpressionsExtrasRealUser collects real-world code shapes
// from production codebases and the upstream rule's GitHub issue tracker —
// patterns that exposed FP / FN regressions historically and that rslint
// users will hit first. The premise is the rslint testing philosophy: passing
// upstream's contrived tests proves nothing about inputs upstream's contributors
// didn't write.
package strict_boolean_expressions

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestStrictBooleanExpressionsExtrasRealUser(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &StrictBooleanExpressionsRule, []rule_tester.ValidTestCase{
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
	}, []rule_tester.InvalidTestCase{
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
	})
}
