package no_unnecessary_condition

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNesting_ConditionalTypeAndMappedType tests conditional types, mapped types,
// template literal types, and other advanced type constructs in conditions.
func TestNesting_ConditionalTypeAndMappedType(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{
		// Conditional type — TypeVariable, always necessary
		{Code: "function f<T>(x: T extends string ? number : boolean) { if (x) {} }\n"},
		// Mapped type property access — string value, could be empty
		{Code: "type M = { [K in 'a' | 'b']: string };\ndeclare const m: M;\nif (m.a) {}\n"},
		// Template literal type with possible empty — `${string}` can be empty
		{Code: "declare const x: `${string}`;\nif (x) {}\n"},
		// Recursive type alias
		{Code: "type Tree = { value: number; children: Tree[] };\ndeclare const t: Tree | null;\nif (t) { t.children[0]?.value; }\n"},
		// keyof — string | number | symbol, could be falsy
		{Code: "function f<T>(k: keyof T) { if (k) {} }\n"},
		// Index access type — T[K] is type variable
		{Code: "function f<T, K extends keyof T>(val: T[K]) { if (val) {} }\n"},
		// Discriminated union in condition
		{Code: "type A = { kind: 'a'; value: string };\ntype B = { kind: 'b'; value: number };\ndeclare const x: A | B;\nif (x.value) {}\n"},
	}, []rule_tester.InvalidTestCase{
		// Mapped type with object values — always truthy
		{
			Code:   "type M = { [K in 'a' | 'b']: { v: number } };\ndeclare const m: M;\nif (m.a) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 3, Column: 5}},
		},
	})
}

// TestNesting_AsConstAndReadonly tests `as const`, readonly arrays/tuples,
// const assertions, and readonly modifiers.
func TestNesting_AsConstAndReadonly(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{
		// as const array with possible falsy — 0 is falsy
		{Code: "const arr = [0, 1, 2] as const;\ndeclare const i: 0 | 1 | 2;\nif (arr[i]) {}\n"},
		// Readonly array — same as regular for truthiness
		{Code: "declare const arr: readonly string[];\nif (arr[0]) {}\n"},
		// const enum — could be 0
		{Code: "const enum E { A = 0, B = 1 }\ndeclare const e: E;\nif (e) {}\n"},
	}, []rule_tester.InvalidTestCase{
		// as const object — always truthy
		{
			Code:   "const obj = { a: 1 } as const;\nif (obj) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 5}},
		},
		// Readonly array itself — always truthy (array is object)
		{
			Code:   "declare const arr: readonly string[];\nif (arr) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 5}},
		},
	})
}

// TestNesting_ClassAndInheritance tests class instances, inheritance,
// abstract classes, and method return types.
func TestNesting_ClassAndInheritance(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{
		// Class instance — nullable
		{Code: "class Foo { bar() { return 1; } }\ndeclare const f: Foo | null;\nif (f) { f.bar(); }\n"},
		// Abstract class method return — could be any value
		{Code: "abstract class Base { abstract getValue(): string | null; }\ndeclare const b: Base;\nif (b.getValue()) {}\n"},
		// Method returning boolean
		{Code: "class Checker { check(): boolean { return true; } }\ndeclare const c: Checker;\nif (c.check()) {}\n"},
	}, []rule_tester.InvalidTestCase{
		// Class instance — always truthy
		{
			Code:   "class Foo {}\ndeclare const f: Foo;\nif (f) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 3, Column: 5}},
		},
		// Optional chain on non-nullable class instance
		{
			Code: "class Foo { bar = 'hello'; }\ndeclare const f: Foo;\nf?.bar;\n",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "neverOptionalChain", Line: 3, Column: 2,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestRemoveOptionalChain", Output: "class Foo { bar = 'hello'; }\ndeclare const f: Foo;\nf.bar;\n"},
				},
			}},
		},
	})
}

// TestNesting_PromiseAndAsync tests async/await patterns, Promise types.
func TestNesting_PromiseAndAsync(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{
		// Promise is always truthy but awaited value might not be
		{Code: "async function f() {\n  const x = await Promise.resolve(Math.random() > 0.5);\n  if (x) {}\n}\n"},
		// Awaited nullable
		{Code: "async function f(p: Promise<string | null>) {\n  const x = await p;\n  if (x) {}\n}\n"},
	}, []rule_tester.InvalidTestCase{
		// Promise itself is always truthy (object)
		{
			Code:   "declare const p: Promise<string>;\nif (p) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 5}},
		},
	})
}

// TestNesting_DeepOptionalChainCompositions tests complex optional chain
// patterns: mixed ?.prop, ?.[], ?.() in deep chains with various nullable positions.
func TestNesting_DeepOptionalChainCompositions(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{
		// Chain: nullable root → method call → property access
		{Code: "type API = { fetch(): { data: string } };\ndeclare const api: API | null;\napi?.fetch().data;\n"},
		// Chain: property → nullable method return → property
		{Code: "type Obj = { get(): string | null };\ndeclare const o: Obj;\no.get()?.length;\n"},
		// Chain: nullable root → computed → method
		{Code: "type Dict = { [k: string]: { run(): void } | null };\ndeclare const d: Dict | null;\nd?.['key']?.run();\n"},
		// Array index → optional chain → method (without noUncheckedIndexedAccess)
		{Code: "declare const callbacks: Array<(() => string) | null>;\ncallbacks[0]?.();\n"},
		// Nested ternary in optional chain context
		{Code: "declare const x: { a?: { b: string } } | null;\nconst y = x?.a?.b ?? 'default';\n"},
	}, []rule_tester.InvalidTestCase{
		// Non-nullable root → method with non-nullable return → unnecessary ?.
		{
			Code: "type API = { fetch(): { data: string } };\ndeclare const api: API;\napi.fetch()?.data;\n",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "neverOptionalChain", Line: 3, Column: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestRemoveOptionalChain", Output: "type API = { fetch(): { data: string } };\ndeclare const api: API;\napi.fetch().data;\n"},
				},
			}},
		},
	})
}

// TestNesting_NullishCoalescingCompositions tests ?? with various left-hand
// side expressions: member access, call, element access, nested ??.
func TestNesting_NullishCoalescingCompositions(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{
		// ?? with function call returning nullable
		{Code: "function get(): string | null { return null; }\nconst x = get() ?? 'fallback';\n"},
		// ?? with optional property
		{Code: "declare const obj: { a?: string };\nconst x = obj.a ?? 'default';\n"},
		// ?? with computed nullable property
		{Code: "declare const obj: { [k: string]: string | undefined };\nconst x = obj['key'] ?? 'default';\n"},
		// Chained ?? — first is nullable, second catches
		{Code: "declare const a: string | null;\ndeclare const b: string | null;\ndeclare const c: string;\nconst x = a ?? b ?? c;\n"},
		// ?? with ternary
		{Code: "declare const a: string | null;\nconst x = (a ?? 'default') ? 'yes' : 'no';\n"},
	}, []rule_tester.InvalidTestCase{
		// ?? on function returning non-nullable
		{
			Code:   "function get(): string { return 'hi'; }\nconst x = get() ?? 'fallback';\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish", Line: 2, Column: 11}},
		},
		// ?? on non-nullable member access
		{
			Code:   "declare const obj: { a: string };\nconst x = obj.a ?? 'default';\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish", Line: 2, Column: 11}},
		},
	})
}

// TestNesting_LogicalChainCompositions tests complex logical expression chains
// with multiple levels: if (a && b || c && d), nested negation, mixed types.
func TestNesting_LogicalChainCompositions(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{
		// Complex chain — all necessary
		{Code: "declare const a: boolean;\ndeclare const b: string;\ndeclare const c: number;\nif (a && b && c) {}\n"},
		// Negated chain
		{Code: "declare const a: boolean;\ndeclare const b: boolean;\nif (!(a && b)) {}\n"},
		// Mixed && and || — right sides checked in nested context
		{Code: "declare const a: boolean;\ndeclare const b: boolean;\ndeclare const c: boolean;\nif ((a && b) || c) {}\n"},
		// Ternary with logical
		{Code: "declare const a: boolean;\ndeclare const b: string;\nconst x = a && b ? 'yes' : 'no';\n"},
		// Short circuit assignment chain
		{Code: "declare let a: string | null;\ndeclare let b: string | null;\na ||= b || 'default';\n"},
	}, []rule_tester.InvalidTestCase{
		// Always-truthy at start of && chain
		{
			Code:   "declare const b: boolean;\nif ({} && b) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 5}},
		},
		// Always-falsy at end of || chain
		{
			Code:   "declare const b: boolean;\nif (b || undefined) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy", Line: 2, Column: 10}},
		},
		// Nested: [] is always-truthy — b && [] simple case
		{
			Code:   "declare const b: boolean;\nif (b && []) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 10}},
		},
		// Three-part chain: [] reported twice (RHS of left &&, LHS of right &&)
		{
			Code: "declare const b: boolean;\nif (b && [] && b) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "alwaysTruthy", Line: 2, Column: 16},
				{MessageId: "alwaysTruthy", Line: 2, Column: 10},
			},
		},
	})
}

// TestNesting_SwitchCaseCompositions tests switch/case with various
// discriminant and case types.
func TestNesting_SwitchCaseCompositions(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{
		// Discriminated union switch
		{Code: "type A = { kind: 'a' };\ntype B = { kind: 'b' };\ndeclare const x: A | B;\nswitch (x.kind) { case 'a': break; case 'b': break; }\n"},
		// Enum switch
		{Code: "enum E { A, B, C }\ndeclare const e: E;\nswitch (e) { case E.A: break; case E.B: break; }\n"},
		// String switch with multiple cases
		{Code: "declare const s: string;\nswitch (s) { case 'a': break; case 'b': break; default: break; }\n"},
	}, []rule_tester.InvalidTestCase{
		// Literal discriminant vs impossible case
		{
			Code:   "const x = 'a' as const;\nswitch (x) { case 'b': break; }\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonBetweenLiteralTypes", Line: 2, Column: 19}},
		},
		// Boolean switch
		{
			Code:   "switch (true) { case false: break; }\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonBetweenLiteralTypes", Line: 1, Column: 22}},
		},
	})
}

// TestNesting_AssignmentInCondition tests assignment expressions used
// as conditions or in logical/nullish context.
func TestNesting_AssignmentInCondition(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{
		// Assignment in while — result type is boolean, necessary
		{Code: "declare let x: boolean;\nwhile (x = Math.random() > 0.5) {}\n"},
		// Nullish assign on nullable
		{Code: "declare let cache: Map<string, string> | null;\ncache ??= new Map();\n"},
		// &&= on string (could be empty)
		{Code: "declare let s: string;\ns &&= s.trim();\n"},
		// ||= on number (could be 0)
		{Code: "declare let n: number;\nn ||= 1;\n"},
	}, []rule_tester.InvalidTestCase{
		// ??= on Map (non-nullable)
		{
			Code:   "declare let cache: Map<string, string>;\ncache ??= new Map();\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish", Line: 2, Column: 1}},
		},
	})
}

// TestNesting_MultipleErrorsInSingleExpression tests expressions that
// produce multiple diagnostics.
func TestNesting_MultipleErrorsInSingleExpression(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		// Deep chain: 4 unnecessary optional chains
		{
			Code: "declare const x: { a: { b: { c: { d: string } } } };\nx?.a?.b?.c?.d;\n",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "neverOptionalChain", Line: 2, Column: 11,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveOptionalChain", Output: "declare const x: { a: { b: { c: { d: string } } } };\nx?.a?.b?.c.d;\n"}}},
				{MessageId: "neverOptionalChain", Line: 2, Column: 8,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveOptionalChain", Output: "declare const x: { a: { b: { c: { d: string } } } };\nx?.a?.b.c?.d;\n"}}},
				{MessageId: "neverOptionalChain", Line: 2, Column: 5,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveOptionalChain", Output: "declare const x: { a: { b: { c: { d: string } } } };\nx?.a.b?.c?.d;\n"}}},
				{MessageId: "neverOptionalChain", Line: 2, Column: 2,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveOptionalChain", Output: "declare const x: { a: { b: { c: { d: string } } } };\nx.a?.b?.c?.d;\n"}}},
			},
		},
		// Mixed: always-truthy in if + noOverlap in comparison
		{
			Code: "declare const obj: { a: string };\nif (obj) {}\nif (obj.a === null) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "alwaysTruthy", Line: 2, Column: 5},
				{MessageId: "noOverlapBooleanExpression", Line: 3, Column: 5},
			},
		},
		// && chain with multiple truthy — order depends on visitor traversal
		{
			Code: "if ([] && {} && true) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "alwaysTruthy", Line: 1, Column: 17},
				{MessageId: "alwaysTruthy", Line: 1, Column: 11},
				{MessageId: "alwaysTruthy", Line: 1, Column: 5},
			},
		},
		// noStrictNullCheck produces error at line 0 + actual errors
		{
			Code:     "declare const obj: object;\nif (obj) {}\n",
			TSConfig: "tsconfig.unstrict.json",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noStrictNullCheck", Line: 0, Column: 0},
				{MessageId: "alwaysTruthy", Line: 2, Column: 5},
			},
		},
	})
}

// TestNesting_GenericConstraintCombinations tests complex generic type
// parameters with various constraint combinations.
func TestNesting_GenericConstraintCombinations(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{
		// Unconstrained generic — always necessary
		{Code: "function f<T>(x: T) { if (x) {} }\n"},
		// Constrained to union of truthy/falsy
		{Code: "function f<T extends string | number>(x: T) { if (x) {} }\n"},
		// Constrained to nullable
		{Code: "function f<T extends string | null>(x: T) { if (x) {} }\n"},
		// Multiple type params
		{Code: "function f<T, U extends T>(x: U) { if (x) {} }\n"},
		// Generic with default
		{Code: "function f<T = string>(x: T) { if (x) {} }\n"},
		// Generic method
		{Code: "class C<T> { check(x: T) { if (x) {} } }\n"},
		// Constrained to nullable + ??
		{Code: "function f<T extends string | null>(x: T) { const y = x ?? 'default'; }\n"},
	}, []rule_tester.InvalidTestCase{
		// Constrained to always-truthy
		{
			Code:   "function f<T extends object>(x: T) { if (x) {} }\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 1, Column: 42}},
		},
		// Constrained to non-empty string literals
		{
			Code:   "function f<T extends 'a' | 'b' | 'c'>(x: T) { if (x) {} }\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 1, Column: 51}},
		},
		// Constrained to non-nullish + ??
		{
			Code:   "function f<T extends string>(x: T) { const y = x ?? 'default'; }\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish", Line: 1, Column: 48}},
		},
	})
}

// TestNesting_ArrayPredicateCompositions tests complex array method
// predicate patterns: chained methods, generic callbacks, overloads.
func TestNesting_ArrayPredicateCompositions(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{
		// Chained filter().find() — different callbacks
		{Code: "[1, 2, 3].filter(x => x > 0).find(x => x < 3);\n"},
		// Generic array with boolean predicate
		{Code: "function f<T>(arr: T[]) { arr.filter(x => !!x); }\n"},
		// Array of union — predicate is necessary
		{Code: "const arr: (string | null)[] = [];\narr.filter(x => x !== null);\n"},
		// every() with boolean callback
		{Code: "declare const arr: number[];\narr.every(x => x > 0);\n"},
		// some() with boolean callback
		{Code: "declare const arr: string[];\narr.some(x => x.length > 0);\n"},
	}, []rule_tester.InvalidTestCase{
		// filter with always-truthy arrow — objects are always truthy
		{
			Code:   "const arr: object[] = [];\narr.filter(x => x);\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 17}},
		},
		// some() with always-falsy block body
		{
			Code:   "[1, 2].some(() => { return null; });\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy", Line: 1, Column: 28}},
		},
	})
}

// TestNesting_TypePredicateCombinations tests checkTypePredicates with
// complex type hierarchies and assertion patterns.
func TestNesting_TypePredicateCombinations(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{
		// Type guard on unknown — necessary
		{Code: "function isString(x: unknown): x is string { return typeof x === 'string'; }\ndeclare const u: unknown;\nif (isString(u)) {}\n", Options: map[string]interface{}{"checkTypePredicates": true}},
		// Type guard narrowing wider → narrower (not subtype)
		{Code: "interface A { a: string }\ninterface B { a: string; b: number }\nfunction isB(x: A): x is B { return 'b' in x; }\ndeclare const a: A;\nif (isB(a)) {}\n", Options: map[string]interface{}{"checkTypePredicates": true}},
		// Assertion on boolean condition — necessary
		{Code: "declare function assert(x: unknown): asserts x;\nconst b = Math.random() > 0.5;\nassert(b);\n", Options: map[string]interface{}{"checkTypePredicates": true}},
		// Type guard with union predicate — necessary when arg is wider
		{Code: "function isStringOrNum(x: unknown): x is string | number { return true; }\ndeclare const x: string | number | boolean;\nisStringOrNum(x);\n", Options: map[string]interface{}{"checkTypePredicates": true}},
	}, []rule_tester.InvalidTestCase{
		// Type guard: arg already matches predicate type
		{
			Code:    "function isNum(x: unknown): x is number { return typeof x === 'number'; }\ndeclare const n: number;\nisNum(n);\n",
			Options: map[string]interface{}{"checkTypePredicates": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "typeGuardAlreadyIsType", Line: 3, Column: 7}},
		},
		// Assertion: arg is always truthy object
		{
			Code:    "declare function assert(x: unknown): asserts x;\nassert({});\n",
			Options: map[string]interface{}{"checkTypePredicates": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 8}},
		},
	})
}
