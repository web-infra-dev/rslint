package no_unnecessary_condition

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestEdgeCases_CheckNode exercises every branch in checkNode:
// - SkipParentheses
// - PrefixUnaryExpression (!, !!, !!!)
// - ArrayIndexExpression skip
// - BinaryExpression (&&, ||) recursive right-only check
// - isConditionalAlwaysNecessary (any, unknown, TypeVariable)
// - never type
// - IsPossiblyTruthy / IsPossiblyFalsy
func TestEdgeCases_CheckNode(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{
		// Triple parenthesized
		{Code: "declare const b: boolean;\nif (((b))) {}\n"},
		// Triple negation on boolean — necessary since boolean can be truthy or falsy
		{Code: "declare const b: boolean;\nif (!!!b) {}\n"},
		// && nested in if — right side of && is checked
		{Code: "declare const a: boolean;\ndeclare const b: boolean;\nif (a && b) {}\n"},
		// || nested in ternary — right side of || is checked
		{Code: "declare const a: boolean;\ndeclare const b: boolean;\nconst x = (a || b) ? 1 : 2;\n"},
		// Conditional type (TypeVariable) — always necessary
		{Code: "function test<T extends boolean>(a: T extends true ? string : number) { if (a) {} }\n"},
		// Index signature access on arrays — unsound, skip
		{Code: "declare const arr: object[];\nif (arr[0]) {}\n"},
		// Deeply nested parens + unary
		{Code: "declare const b: boolean;\nif (((!((b))))) {}\n"},
	}, []rule_tester.InvalidTestCase{
		// Triple negation on always-truthy
		{
			Code:   "declare const obj: object;\nif (!!!obj) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy", Line: 2, Column: 5}},
		},
		// && with always-truthy right — right is checked by parent checkNode
		{
			Code:   "declare const b: boolean;\nif (b && []) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 10}},
		},
		// || with always-falsy right
		{
			Code:   "declare const b: boolean;\nif (b || null) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy", Line: 2, Column: 10}},
		},
		// Nested ternary with truthy
		{
			Code:   "declare const b: boolean;\nconst x = [] ? 'a' : 'b';\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 11}},
		},
	})
}

// TestEdgeCases_CheckNodeForNullish exercises:
// - any/unknown/TypeParameter early return
// - never type
// - isPossiblyNullish + isNullableMemberExpression combined check
// - isAlwaysNullish
// - ArrayIndexExpression skip for ??
// - isChainExpressionWithOptionalArrayIndex
func TestEdgeCases_CheckNodeForNullish(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{
		// any/unknown/TypeVariable — always necessary
		{Code: "declare const x: any;\nconst y = x ?? 1;\n"},
		{Code: "declare const x: unknown;\nconst y = x ?? 1;\n"},
		{Code: "function f<T>(x: T) { const y = x ?? 1; }\n"},
		// Optional property access — nullish because property is optional
		{Code: "declare const obj: { a?: string };\nconst x = obj.a ?? 'default';\n"},
		// Array index ?? — unsound without noUncheckedIndexedAccess
		{Code: "declare const arr: string[];\nconst x = arr[0] ?? '';\n"},
		// Chain with optional array index — should skip
		{Code: "declare const arr: string[];\narr[0]?.length ?? 0;\n"},
		// Nullable union with ??
		{Code: "declare const x: string | null | undefined;\nconst y = x ?? '';\n"},
	}, []rule_tester.InvalidTestCase{
		// never type in ??
		{
			Code:   "declare const x: never;\nconst y = x ?? 1;\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "never", Line: 2, Column: 11}},
		},
		// Non-nullish object in ??
		{
			Code:   "declare const x: { a: string };\nconst y = x ?? {};\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish", Line: 2, Column: 11}},
		},
		// null | undefined — always nullish
		{
			Code:   "declare const x: null | undefined;\nconst y = x ?? 'fallback';\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysNullish", Line: 2, Column: 11}},
		},
	})
}

// TestEdgeCases_BooleanComparison exercises checkIfBoolExpressionIsNecessaryConditional:
// - toStaticValue for each literal kind (bool, string, number, bigint, null, undefined)
// - booleanComparison for ===, !==, ==, !=, <, <=, >, >=
// - noOverlapBooleanExpression (TS #37160 workaround)
// - switch/case comparison (uses ===)
func TestEdgeCases_BooleanComparison(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{
		// Non-literal types — no static comparison possible
		{Code: "declare const a: string;\ndeclare const b: string;\nif (a === b) {}\n"},
		{Code: "declare const a: number;\ndeclare const b: number;\nif (a < b) {}\n"},
		// Mixed literal and non-literal
		{Code: "declare const a: string;\nif (a === 'hello') {}\n"},
		// null == undefined is valid when both sides have nullable types
		{Code: "declare const a: string | null;\nif (a == null) {}\n"},
		{Code: "declare const a: string | undefined;\nif (a == undefined) {}\n"},
		// any/unknown comparisons — always necessary
		{Code: "declare const a: any;\nif (a === null) {}\n"},
		{Code: "declare const a: unknown;\nif (a === null) {}\n"},
		// Type parameter comparisons
		{Code: "function f<T>(a: T) { if (a === null) {} }\n"},
		// Loose equality with nullable types
		{Code: "declare const a: string | null;\nif (a != null) {}\n"},
	}, []rule_tester.InvalidTestCase{
		// === between same literals
		{
			Code:   "if (1 === 1) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonBetweenLiteralTypes", Line: 1, Column: 5}},
		},
		// !== between same literals
		{
			Code:   "if ('a' !== 'a') {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonBetweenLiteralTypes", Line: 1, Column: 5}},
		},
		// == cross-type: null == undefined
		{
			Code:   "if (null == undefined) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonBetweenLiteralTypes", Line: 1, Column: 5}},
		},
		// != null vs undefined
		{
			Code:   "if (null != undefined) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonBetweenLiteralTypes", Line: 1, Column: 5}},
		},
		// String relational: lexicographic order
		{
			Code:   "declare const a: 'abc';\ndeclare const b: 'abd';\nif (a < b) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonBetweenLiteralTypes", Line: 3, Column: 5}},
		},
		// BigInt comparison
		{
			Code:   "if (-2n !== 2n) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonBetweenLiteralTypes", Line: 1, Column: 5}},
		},
		// Number float comparison
		{
			Code:   "if (2.3 > 2.3) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonBetweenLiteralTypes", Line: 1, Column: 5}},
		},
		// noOverlapBooleanExpression: string vs null
		{
			Code:   "function test(a: string) { a === null; }\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noOverlapBooleanExpression", Line: 1, Column: 28}},
		},
		// noOverlapBooleanExpression: string vs undefined (strict ===)
		{
			Code:   "function test(a: string) { a === undefined; }\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noOverlapBooleanExpression", Line: 1, Column: 28}},
		},
		// noOverlapBooleanExpression: reversed order
		{
			Code:   "function test(a: string) { null === a; }\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noOverlapBooleanExpression", Line: 1, Column: 28}},
		},
		// !== also triggers noOverlap
		{
			Code:   "function test(a: string) { a !== null; }\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noOverlapBooleanExpression", Line: 1, Column: 28}},
		},
		// Switch/case literal comparison
		{
			Code:   "switch (true as const) {\n  case false:\n    break;\n}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonBetweenLiteralTypes", Line: 2, Column: 8}},
		},
	})
}

// TestEdgeCases_OptionalChain exercises checkOptionalChain:
// - isOptionableExpression (member vs call origin)
// - isMemberExpressionNullableOriginFromObject
// - isCallExpressionNullableOriginFromCallee
// - optionChainContainsUnsoundIndexAccess (array + object index)
// - Multi-level chains with mixed nullable positions
// - Call chain return type nullable origin
func TestEdgeCases_OptionalChain(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{
		// Nullable object — optional chain necessary
		{Code: "declare const x: { a: string } | null;\nx?.a;\n"},
		// Optional property — necessary
		{Code: "declare const x: { a?: string };\nx.a?.length;\n"},
		// Union of types with different property shapes — nullable origin from type
		{Code: "type A = { foo: string };\ntype B = { foo: string | undefined };\ndeclare const x: A | B;\nx.foo?.length;\n"},
		// Callable union with nullable return
		{Code: "type Fn = (() => string | null) | (() => number);\ndeclare const fn: Fn;\nfn()?.toString();\n"},
		// Array index optional chain — unsound without noUncheckedIndexedAccess
		{Code: "declare const arr: { a: string }[];\narr[0]?.a;\n"},
		// Deep chain: nullable at root only
		{Code: "declare const x: { a: { b: { c: string } } } | null;\nx?.a.b.c;\n"},
		// Chain with method call having nullable return
		{Code: "declare const x: { getValue(): string | null };\nx.getValue()?.length;\n"},
		// any/unknown in chain — always necessary
		{Code: "declare const x: any;\nx?.a?.b;\n"},
		{Code: "declare const x: unknown;\nx?.toString();\n"},
	}, []rule_tester.InvalidTestCase{
		// Non-nullable object property access with ?.
		{
			Code: "declare const x: { a: { b: string } };\nx.a?.b;\n",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "neverOptionalChain", Line: 2, Column: 4,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestRemoveOptionalChain", Output: "declare const x: { a: { b: string } };\nx.a.b;\n"},
				},
			}},
		},
		// Non-nullable callable with ?.()
		{
			Code: "declare const fn: () => number;\nfn?.();\n",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "neverOptionalChain", Line: 2, Column: 3,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestRemoveOptionalChain", Output: "declare const fn: () => number;\nfn();\n"},
				},
			}},
		},
		// Three-level chain all unnecessary
		{
			Code: "declare const x: { a: { b: { c: string } } };\nx?.a?.b?.c;\n",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "neverOptionalChain", Line: 2, Column: 8,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestRemoveOptionalChain", Output: "declare const x: { a: { b: { c: string } } };\nx?.a?.b.c;\n"},
					},
				},
				{
					MessageId: "neverOptionalChain", Line: 2, Column: 5,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestRemoveOptionalChain", Output: "declare const x: { a: { b: { c: string } } };\nx?.a.b?.c;\n"},
					},
				},
				{
					MessageId: "neverOptionalChain", Line: 2, Column: 2,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestRemoveOptionalChain", Output: "declare const x: { a: { b: { c: string } } };\nx.a?.b?.c;\n"},
					},
				},
			},
		},
		// Nullable at root, but inner chain unnecessary
		{
			Code: "declare const x: { a: { b: string } } | null;\nx?.a?.b;\n",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "neverOptionalChain", Line: 2, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestRemoveOptionalChain", Output: "declare const x: { a: { b: string } } | null;\nx?.a.b;\n"},
				},
			}},
		},
		// Computed element access unnecessary
		{
			Code: "declare const x: { [k: string]: string };\nx?.['foo'];\n",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "neverOptionalChain", Line: 2, Column: 2,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestRemoveOptionalChain", Output: "declare const x: { [k: string]: string };\nx['foo'];\n"},
				},
			}},
		},
	})
}

// TestEdgeCases_LoopConditions exercises checkIfLoopIsNecessaryConditional:
// - While / do-while / for
// - allowConstantLoopConditions: 'always' / 'never' / 'only-allowed-literals' / true / false
// - only-allowed-literals allows only while(true/false/0/1)
// - for(;;) with no condition — no check
func TestEdgeCases_LoopConditions(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{
		// for(;;) — no condition, no check
		{Code: "for (;;) {}\n"},
		// while with boolean condition
		{Code: "declare const b: boolean;\nwhile (b) {}\n"},
		// allowConstantLoopConditions: 'always' allows while(true)
		{Code: "while (true) {}", Options: map[string]interface{}{"allowConstantLoopConditions": "always"}},
		// only-allowed-literals allows while(true/false/0/1)
		{Code: "while (true) {}", Options: map[string]interface{}{"allowConstantLoopConditions": "only-allowed-literals"}},
		{Code: "while (false) {}", Options: map[string]interface{}{"allowConstantLoopConditions": "only-allowed-literals"}},
		{Code: "while (1) {}", Options: map[string]interface{}{"allowConstantLoopConditions": "only-allowed-literals"}},
		{Code: "while (0) {}", Options: map[string]interface{}{"allowConstantLoopConditions": "only-allowed-literals"}},
		// only-allowed-literals: do-while and for also accept true/false/0/1
		{Code: "do {} while (true);", Options: map[string]interface{}{"allowConstantLoopConditions": "only-allowed-literals"}},
		{Code: "for (; true; ) {}", Options: map[string]interface{}{"allowConstantLoopConditions": "only-allowed-literals"}},
		{Code: "do {} while (0);", Options: map[string]interface{}{"allowConstantLoopConditions": "only-allowed-literals"}},
		{Code: "for (; 0; ) {}", Options: map[string]interface{}{"allowConstantLoopConditions": "only-allowed-literals"}},
		// Boolean legacy option: true = always
		{Code: "while (true) {}", Options: map[string]interface{}{"allowConstantLoopConditions": true}},
	}, []rule_tester.InvalidTestCase{
		// do-while with always-falsy
		{
			Code:   "do {} while (false);\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy", Line: 1, Column: 14}},
		},
		// for with always-truthy condition
		{
			Code:   "for (let i = 0; []; i++) { break; }\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 1, Column: 17}},
		},
		// only-allowed-literals: do-while and for also allowed with true/false/0/1
		// (these are valid now, moved to valid cases)
		// only-allowed-literals rejects non-0/1 numeric literal
		{
			Code:    "while (2) {}",
			Options: map[string]interface{}{"allowConstantLoopConditions": "only-allowed-literals"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 1, Column: 8}},
		},
		// 'always' still checks non-literal true type
		{
			Code:    "declare const obj: object;\nwhile (obj) {}\n",
			Options: map[string]interface{}{"allowConstantLoopConditions": "always"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 8}},
		},
	})
}

// TestEdgeCases_AssignmentExpressions exercises checkAssignmentExpression:
// - &&= checks left for truthy/falsy
// - ||= checks left for truthy/falsy
// - ??= checks left for nullish
func TestEdgeCases_AssignmentExpressions(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{
		// Nullable with ??=
		{Code: "declare let x: string | null;\nx ??= 'hi';\n"},
		// Boolean with &&=
		{Code: "declare let x: boolean;\nx &&= true;\n"},
		// Boolean with ||=
		{Code: "declare let x: boolean;\nx ||= false;\n"},
		// String (can be falsy) with &&= and ||=
		{Code: "declare let x: string;\nx &&= 'hi';\nx ||= 'hi';\n"},
	}, []rule_tester.InvalidTestCase{
		// &&= on always-truthy
		{
			Code:   "declare let x: object;\nx &&= {};\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 1}},
		},
		// ||= on always-falsy
		{
			Code:   "declare let x: null;\nx ||= null;\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy", Line: 2, Column: 1}},
		},
		// ??= on non-nullish
		{
			Code:   "declare let x: number;\nx ??= 0;\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish", Line: 2, Column: 1}},
		},
		// ??= on always-nullish
		{
			Code:   "declare let x: null;\nx ??= null;\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysNullish", Line: 2, Column: 1}},
		},
	})
}

// TestEdgeCases_CallExpression exercises checkCallExpression:
// - Array predicate with arrow expression body
// - Array predicate with arrow block body (single return)
// - Array predicate with arrow block body (multiple returns → return type check)
// - Array predicate with named function reference
// - findIndex, some, every, findLast, findLastIndex methods
// - Generic callback constraint
// - checkTypePredicates: asserts x, x is Type, asserts x is Type
// - Spread arguments bail out
func TestEdgeCases_CallExpression(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{
		// Arrow expression body with boolean return
		{Code: "[1, 2, 3].filter(x => x > 1);\n"},
		// Arrow block body with conditional return
		{Code: "[1, 2, 3].filter(x => { if (x > 1) return true; return false; });\n"},
		// some/every/find/findIndex all valid with boolean predicate
		{Code: "[1, 2, 3].some(x => x > 0);\n"},
		{Code: "[1, 2, 3].every(x => x > 0);\n"},
		{Code: "[1, 2, 3].find(x => x > 0);\n"},
		{Code: "[1, 2, 3].findIndex(x => x > 0);\n"},
		// Named function reference with boolean return type
		{Code: "function isPositive(x: number) { return x > 0; }\n[1, 2, 3].filter(isPositive);\n"},
		// checkTypePredicates: disabled, so type guard doesn't report
		{Code: "declare function isString(x: unknown): x is string;\ndeclare const s: string;\nisString(s);\n"},
		// Spread args — skip because parameter index is unreliable
		{Code: "declare function assert(x: unknown): asserts x;\nassert(...[]);\n", Options: map[string]interface{}{"checkTypePredicates": true}},
	}, []rule_tester.InvalidTestCase{
		// Arrow expression body always truthy
		{
			Code:   "[1, 2, 3].filter(x => [x]);\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 1, Column: 23}},
		},
		// Arrow block body with single always-falsy return
		{
			Code:   "[1, 2, 3].filter(x => { return null; });\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy", Line: 1, Column: 32}},
		},
		// Named function with always-truthy return
		{
			Code:   "function getTruthy() { return {}; }\n[1, 2, 3].some(getTruthy);\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthyFunc", Line: 2, Column: 16}},
		},
		// findIndex with always-falsy callback
		{
			Code:   "function getFalsy() { return; }\n[1, 2, 3].findIndex(getFalsy);\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsyFunc", Line: 2, Column: 21}},
		},
		// checkTypePredicates: asserts x on always truthy
		{
			Code:    "declare function assert(x: unknown): asserts x;\nassert([]);\n",
			Options: map[string]interface{}{"checkTypePredicates": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 8}},
		},
		// checkTypePredicates: x is Type already matches
		{
			Code:    "declare function isNumber(x: unknown): x is number;\ndeclare const n: number;\nisNumber(n);\n",
			Options: map[string]interface{}{"checkTypePredicates": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "typeGuardAlreadyIsType", Line: 3, Column: 10}},
		},
	})
}

// TestEdgeCases_IsPossiblyFalsyTruthy exercises all branches in
// isConstituentPossiblyFalsy and isConstituentPossiblyTruthy:
// - any, unknown, TypeVariable → always both
// - never → neither
// - null, undefined, void → falsy only
// - BooleanLike: true → truthy only, false → falsy only, boolean → both
// - StringLiteral: "" → falsy, "x" → truthy
// - NumberLiteral: 0 → falsy, 1 → truthy
// - BigIntLiteral: 0n → falsy, 1n → truthy
// - String, Number, BigInt (non-literal) → both
// - EnumLiteral → both (conservative)
// - Object → truthy only
// - Union → recurse
// - Intersection → Some(falsy), Every(truthy)
func TestEdgeCases_IsPossiblyFalsyTruthy(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{
		// Enum with possible 0 value — both truthy and falsy
		{Code: "enum E { A = 0, B = 1 }\ndeclare const e: E;\nif (e) {}\n"},
		// boolean (non-literal) — both
		{Code: "declare const b: boolean;\nif (b) {}\n"},
		// String (non-literal) — both
		{Code: "declare const s: string;\nif (s) {}\n"},
		// BigInt (non-literal) — both
		{Code: "declare const bi: bigint;\nif (bi) {}\n"},
		// Number (non-literal) — both
		{Code: "declare const n: number;\nif (n) {}\n"},
		// Intersection with both truthy and falsy constituent
		{Code: "declare const x: string & { __brand: string };\nif (x) {}\n"},
	}, []rule_tester.InvalidTestCase{
		// null — falsy only
		{
			Code:   "if (null) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy", Line: 1, Column: 5}},
		},
		// undefined — falsy only
		{
			Code:   "declare const x: undefined;\nif (x) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy", Line: 2, Column: 5}},
		},
		// void — falsy only
		{
			Code:   "declare const x: void;\nif (x) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy", Line: 2, Column: 5}},
		},
		// "" — falsy only
		{
			Code:   "declare const x: '';\nif (x) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy", Line: 2, Column: 5}},
		},
		// 0 — falsy only
		{
			Code:   "declare const x: 0;\nif (x) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy", Line: 2, Column: 5}},
		},
		// 0n — falsy only
		{
			Code:   "declare const x: 0n;\nif (x) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy", Line: 2, Column: 5}},
		},
		// false — falsy only
		{
			Code:   "declare const x: false;\nif (x) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy", Line: 2, Column: 5}},
		},
		// true — truthy only
		{
			Code:   "declare const x: true;\nif (x) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 5}},
		},
		// "hello" — truthy only
		{
			Code:   "declare const x: 'hello';\nif (x) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 5}},
		},
		// 42 — truthy only
		{
			Code:   "declare const x: 42;\nif (x) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 5}},
		},
		// 1n — truthy only
		{
			Code:   "declare const x: 1n;\nif (x) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 5}},
		},
		// object — truthy only
		{
			Code:   "declare const x: object;\nif (x) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 5}},
		},
		// Intersection of falsy literals → falsy
		{
			Code:   "declare const x: '' & { __brand: string };\nif (x) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy", Line: 2, Column: 5}},
		},
		// Union of only falsy types
		{
			Code:   "declare const x: null | undefined | 0 | '' | false;\nif (x) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy", Line: 2, Column: 5}},
		},
		// Union of only truthy types
		{
			Code:   "declare const x: object | true | 'hello' | 42;\nif (x) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 5}},
		},
	})
}

// TestEdgeCases_DeepNesting tests deeply nested/composed patterns
func TestEdgeCases_DeepNesting(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{
		// Nested ternaries with boolean conditions
		{Code: "declare const a: boolean;\ndeclare const b: boolean;\nconst x = a ? (b ? 1 : 2) : 3;\n"},
		// Deep logical chain
		{Code: "declare const a: boolean;\ndeclare const b: boolean;\ndeclare const c: boolean;\nif (a && b && c) {}\n"},
		// Nested optional chain with mixed nullable levels
		{Code: "type A = { b?: { c: { d: string } | null } };\ndeclare const a: A | null;\na?.b?.c?.d;\n"},
		// Deep ?? chain
		{Code: "declare const a: string | null;\ndeclare const b: string | null;\ndeclare const c: string;\nconst x = a ?? b ?? c;\n"},
		// if inside if
		{Code: "declare const a: boolean;\ndeclare const b: boolean;\nif (a) { if (b) {} }\n"},
		// Nested function with generic
		{Code: "function outer<T>(t: T) {\n  function inner<U>(u: U) {\n    if (u) {}\n  }\n  if (t) {}\n}\n"},
	}, []rule_tester.InvalidTestCase{
		// Deep logical chain with always-truthy at left
		{
			Code: "declare const b: boolean;\ndeclare const c: boolean;\nif ([] && b && c) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "alwaysTruthy", Line: 3, Column: 5},
			},
		},
		// Nested ?? — inner is never nullish
		{
			Code: "declare const a: string;\ndeclare const b: string | null;\nconst x = (a ?? 'x') ?? b;\n",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "neverNullish", Line: 3, Column: 11},
				{MessageId: "neverNullish", Line: 3, Column: 12},
			},
		},
	})
}
