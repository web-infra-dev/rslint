package no_unnecessary_condition

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnnecessaryCondition(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{
		// === Basic types ===
		{Code: "declare const b1: boolean;\ndeclare const b2: boolean;\nconst t1 = b1 && b2;\n"},
		{Code: "declare const b1: boolean;\ndeclare const b2: boolean;\nconst t1 = b1 || b2;\n"},
		{Code: "declare const b1: boolean;\nconst t1 = b1 ? 'yes' : 'no';\n"},
		{Code: "function foo(): boolean { return true; }\nif (foo()) {}\n"},
		{Code: "declare const n: number;\nif (n) {}\n"},
		{Code: "declare const s: string;\nif (s) {}\n"},
		{Code: "declare const b: bigint;\nif (b) {}\n"},
		{Code: "declare const x: boolean;\nif (x) {}\n"},

		// === any / unknown / type variables ===
		{Code: "declare const x: any;\nif (x) {}\n"},
		{Code: "declare const x: unknown;\nif (x) {}\n"},
		{Code: "declare const x: any;\nconst y = x ?? 'default';\n"},
		{Code: "declare const x: unknown;\nconst y = x ?? 'default';\n"},
		{Code: "function test<T>(a: T) { return a ?? 'default'; }\n"},
		{Code: "function test<T extends string | null>(a: T) { return a ?? 'default'; }\n"},

		// === Nullable types ===
		{Code: "declare const x: string | null;\nif (x) {}\n"},
		{Code: "declare const x: string | undefined;\nif (x) {}\n"},
		{Code: "declare const x: string | null | undefined;\nconst y = x ?? 'default';\n"},

		// === void type ===
		{Code: "declare function foo(): number | void;\nconst r = foo() === undefined;\n"},
		{Code: "declare function foo(): number | void;\nconst r = foo() == null;\n"},

		// === Generics and type parameters ===
		{Code: "function test<T extends string>(t: T) {\n  return t ? 'yes' : 'no';\n}\n"},
		{Code: "function test<T>(t: T) {\n  return t ? 'yes' : 'no';\n}\n"},
		{Code: "function test<T>(t: T | []) {\n  return t ? 'yes' : 'no';\n}\n"},

		// === Branded / intersection types ===
		{Code: "declare const b1: string & { __brand: string };\nif (b1) {}\n"},
		{Code: "declare const b1: number & { __brand: string };\nif (b1) {}\n"},
		{Code: "declare const b1: boolean & { __brand: string };\nif (b1) {}\n"},
		{Code: "declare const b1: bigint & { __brand: string };\nif (b1) {}\n"},
		{Code: "declare const b1: string & {};\nif (b1) {}\n"},
		{Code: "declare const b1: string & {} & { __brand: string };\nif (b1) {}\n"},
		{Code: "declare const b1: string & { __brandA: string } & { __brandB: string };\nif (b1) {}\n"},
		// Union of branded types
		{Code: "declare const b1: string & { __brand: string } | number;\ndeclare const b2: boolean;\nconst t1 = b1 && b2;\n"},
		{Code: "declare const b1: (string | number) & { __brand: string };\ndeclare const b2: boolean;\nconst t1 = b1 && b2;\n"},

		// === Unions mixing falsy and truthy ===
		{Code: "declare const x: false | 5;\ndeclare const b: boolean;\nconst t1 = x && b;\n"},
		{Code: "declare const x: boolean | 'foo';\ndeclare const b: boolean;\nconst t1 = x && b;\n"},
		{Code: "declare const x: 0 | boolean;\ndeclare const b: boolean;\nconst t1 = x && b;\n"},
		{Code: "declare const x: boolean | object;\ndeclare const b: boolean;\nconst t1 = x && b;\n"},
		{Code: "declare const x: null | object;\ndeclare const b: boolean;\nconst t1 = x && b;\n"},
		{Code: "declare const x: undefined | true;\ndeclare const b: boolean;\nconst t1 = x && b;\n"},
		{Code: "declare const x: void | true;\ndeclare const b: boolean;\nconst t1 = x && b;\n"},
		// BigInt union
		{Code: "declare const bigInt: 0n | 1n;\nif (bigInt) {}\n"},

		// === Boolean comparisons with non-literal types ===
		{Code: "declare const a: string;\ndeclare const b: string;\nif (a === b) {}\n"},
		{Code: "declare const a: number;\ndeclare const b: number;\nif (a !== b) {}\n"},
		{Code: "declare const a: string;\ndeclare const b: string;\nif (a == b) {}\n"},
		// Optional param null/undefined comparisons
		{Code: "function test(a?: string) { return a === undefined; }\n"},
		{Code: "function test(a?: string) { return a !== undefined; }\n"},
		{Code: "function test(a: null | string) { return a === null; }\n"},
		{Code: "function test(a: null | string) { return a !== null; }\n"},
		// Loose equality
		{Code: "function test(a?: string) { return a == null; }\n"},
		{Code: "function test(a?: string) { return a != null; }\n"},
		{Code: "function test(a: null | string) { return a == undefined; }\n"},
		{Code: "function test(a: null | string) { return a != undefined; }\n"},
		// any/unknown with all comparisons
		{Code: "function test(a: any) { return a === null; }\n"},
		{Code: "function test(a: unknown) { return a === undefined; }\n"},
		// Generic type parameter
		{Code: "function test<T>(a: T) { return a === undefined; }\n"},

		// === Predicate functions ===
		{Code: "[0, 1, 2, 3].filter(t => t);\n"},
		{Code: "['a', 'b', ''].filter(t => t);\n"},
		{Code: "[1, null].filter(1 as any);\n"},
		{Code: "[1, null].filter(1 as never);\n"},
		// Named predicate returning boolean
		{Code: "const length = (x: string) => x.length;\n['a', 'b', ''].filter(length);\n"},
		// Non-array object with filter method name
		{Code: "declare const notArray: { filter(fn: () => boolean): boolean };\nnotArray.filter(() => true);\n"},

		// === Nullish coalescing ===
		{Code: "declare const x: string | null;\nconst y = x ?? 'default';\n"},
		{Code: "declare const x: string | undefined;\nconst y = x ?? 'default';\n"},
		// testVal ?? true with optional boolean
		{Code: "function test(testVal?: boolean) { if (testVal ?? true) {} }\n"},

		// === Optional chaining ===
		{Code: "declare const x: { a?: { b: string } };\nx.a?.b;\n"},
		{Code: "declare const x: { a: { b: string } } | null;\nx?.a.b;\n"},
		{Code: "declare const x: { a: string } | null;\nx?.a;\n"},
		{Code: "let foo: undefined | { bar: true };\nfoo?.bar;\n"},
		{Code: "let foo: null | { bar: true };\nfoo?.bar;\n"},
		{Code: "let foo: undefined;\nfoo?.bar;\n"},
		{Code: "let foo: null;\nfoo?.bar;\n"},
		{Code: "let anyValue: any;\nanyValue?.foo;\n"},
		{Code: "let unknownValue: unknown;\nunknownValue?.foo;\n"},
		// Optional call
		{Code: "let foo: undefined | (() => {});\nfoo?.();\n"},
		{Code: "let foo: null | (() => {});\nfoo?.();\n"},
		{Code: "let anyValue: any;\nanyValue?.();\n"},
		// Deep optional chain
		{Code: "declare const foo: { bar?: { baz: { c: string } } } | null;\nfoo?.bar?.baz;\n"},
		// Optional chain with return type that may be nullish
		{Code: "type Foo = { bar: () => number | undefined } | null;\ndeclare const foo: Foo;\nfoo?.bar()?.toExponential();\n"},
		{Code: "type Foo = (() => number | undefined) | null;\ndeclare const foo: Foo;\nfoo?.()?.toExponential();\n"},
		// void return
		{Code: "declare function foo(): void | { key: string };\nconst bar = foo()?.key;\n"},
		{Code: "type fn = () => void;\ndeclare function foo(): void | fn;\nconst bar = foo()?.();\n"},

		// === Array index expressions ===
		{Code: "declare const arr: string[];\nif (arr[0]) {}\n"},
		{Code: "declare const arr: string[];\nconst x = arr[0] ?? 'default';\n"},
		{Code: "declare const arr: string[][];\nconst y = arr[0] ?? [];\n"},
		{Code: "declare const arr: string[];\nif (!arr[0]) {}\n"},
		// Array index with optional chain
		{Code: "declare const arr: string[];\narr[0]?.length;\n"},

		// === allowConstantLoopConditions ===
		{Code: "while (true) {}", Options: map[string]interface{}{"allowConstantLoopConditions": "always"}},
		{Code: "for (; true; ) {}", Options: map[string]interface{}{"allowConstantLoopConditions": "always"}},
		{Code: "do {} while (true);", Options: map[string]interface{}{"allowConstantLoopConditions": "always"}},
		{Code: "while (true) {}", Options: map[string]interface{}{"allowConstantLoopConditions": true}},
		{Code: "while (true) {}", Options: map[string]interface{}{"allowConstantLoopConditions": "only-allowed-literals"}},
		{Code: "while (false) {}", Options: map[string]interface{}{"allowConstantLoopConditions": "only-allowed-literals"}},
		{Code: "while (1) {}", Options: map[string]interface{}{"allowConstantLoopConditions": "only-allowed-literals"}},
		{Code: "while (0) {}", Options: map[string]interface{}{"allowConstantLoopConditions": "only-allowed-literals"}},
		{Code: "for (;;) {}"},

		// === Logical assignments ===
		{Code: "declare let x: string | null;\nx &&= 'hello';\n"},
		{Code: "declare let x: string | null;\nx ||= 'hello';\n"},
		{Code: "declare let x: string | null;\nx ??= 'hello';\n"},
		{Code: "declare let foo: number | null;\nfoo ??= 1;\n"},

		// === Enums ===
		{Code: "enum Fruit { Apple, Orange }\ndeclare const fruit: Fruit;\nif (fruit === Fruit.Apple) {}\n"},

		// === Parenthesized expressions ===
		{Code: "declare const b: boolean;\nif ((b)) {}\n"},
		{Code: "declare const b1: boolean;\ndeclare const b2: boolean;\nif ((b1 && b2)) {}\n"},

		// === Switch case ===
		{Code: "declare const x: string;\nswitch (x) { case 'a': break; case 'b': break; }\n"},

		// === checkTypePredicates valid cases ===
		{Code: "declare function isString(x: unknown): x is string;\ndeclare const u: unknown;\nisString(u);\n", Options: map[string]interface{}{"checkTypePredicates": true}},
		{Code: "declare function assert(x: unknown): asserts x;\nassert(Math.random() > 0.5);\n", Options: map[string]interface{}{"checkTypePredicates": true}},
		{Code: "declare function assert(x: unknown, y: unknown): asserts x;\nassert(Math.random() > 0.5, true);\n", Options: map[string]interface{}{"checkTypePredicates": true}},
		// Spread — parameter index mapping is unreliable
		{Code: "declare function assert(x: unknown): asserts x;\nassert(...[]);\n", Options: map[string]interface{}{"checkTypePredicates": true}},
		{Code: "declare function assert(x: unknown): asserts x;\nassert(...[], {});\n", Options: map[string]interface{}{"checkTypePredicates": true}},
		// checkTypePredicates disabled by default
		{Code: "declare function assert(x: unknown): asserts x;\nassert(true);\n"},
		{Code: "declare function isString(x: unknown): x is string;\ndeclare const a: string;\nisString(a);\n"},
		// Literal subtype (type is 'falafel' not string, so guard still narrows)
		{Code: "declare function assertString(x: unknown): asserts x is string;\nassertString('falafel');\n", Options: map[string]interface{}{"checkTypePredicates": true}},


		// === exactOptionalPropertyTypes: private optional field with ??= ===
		{
			Code:     "class C {\n  #rand?: number;\n  m() { this.#rand ??= Math.random(); }\n}\n",
			TSConfig: "tsconfig.exactOptionalPropertyTypes.json",
		},

		// === Nested logical expressions ===
		{Code: "declare const a: boolean;\ndeclare const b: boolean;\nif (a && b || a) {}\n"},
		{Code: "declare const x: string | null;\ndeclare const y: string;\nconst z = (x ?? y) || 'fallback';\n"},

		// === Negated boolean in logical context ===
		{Code: "declare const booleanTyped: boolean;\ndeclare const unknownTyped: unknown;\nif (!(booleanTyped || unknownTyped)) {}\n"},

		// === RHS of logical expression is not checked ===
		{Code: "declare const b1: boolean;\ndeclare const b2: true;\nconst x = b1 && b2;\n"},

		// === Generic indexed access ===
		{Code: "function foo<T extends object>(arg: T, key: keyof T): void { arg[key] == null; }\n"},
		{Code: "function foo<T extends object>(arg: T, key: keyof T): void { arg[key] ?? 'default'; }\n"},
	}, []rule_tester.InvalidTestCase{
		// === Always truthy ===
		{
			Code:   "declare const arr: number[];\nif (arr) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 5}},
		},
		{
			Code:   "declare const obj: { a: string };\nif (obj) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 5}},
		},
		{
			Code:   "declare const x: 'hello';\nif (x) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 5}},
		},
		{
			Code:   "declare const x: 1;\nif (x) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 5}},
		},
		// Truthy bigint literal
		{
			Code:   "declare const x: 1n;\nif (x) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 5}},
		},
		// Negative bigint (always truthy)
		{
			Code:   "declare const negBigInt: -2n;\nif (negBigInt) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 5}},
		},
		// object | true union always truthy
		{
			Code:   "declare const x: object | true;\ndeclare const b: boolean;\nconst t1 = x && b;\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 3, Column: 12}},
		},

		// === Always falsy ===
		{
			Code:   "declare const x: null;\nif (x) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy", Line: 2, Column: 5}},
		},
		{
			Code:   "declare const x: undefined;\nif (x) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy", Line: 2, Column: 5}},
		},
		{
			Code:   "declare const x: void;\nif (x) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy", Line: 2, Column: 5}},
		},
		{
			Code:   "declare const x: 0n;\nif (x) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy", Line: 2, Column: 5}},
		},
		{
			Code:   "declare const x: 0;\nif (x) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy", Line: 2, Column: 5}},
		},
		{
			Code:   "declare const x: '';\nif (x) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy", Line: 2, Column: 5}},
		},
		{
			Code:   "declare const x: '' | false;\nif (x) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy", Line: 2, Column: 5}},
		},
		// const null
		{
			Code:   "const a = null;\nif (!a) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 5}},
		},

		// === Ternary ===
		{
			Code:   "declare const obj: { a: string };\nconst x = obj ? 'yes' : 'no';\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 11}},
		},

		// === Logical operators ===
		{
			Code:   "declare const obj: { a: string };\ndeclare const b: boolean;\nconst x = obj && b;\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 3, Column: 11}},
		},
		{
			Code:   "declare const x: null;\ndeclare const b: boolean;\nconst y = x || b;\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy", Line: 3, Column: 11}},
		},
		// Nested logical with mixed truthy/falsy
		{
			Code: "declare const b1: boolean;\ndeclare const b2: boolean;\nif (true && b1 && b2) {}\nif (b1 && false && b2) {}\nif (b1 || b2 || true) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "alwaysTruthy", Line: 3, Column: 5},
				{MessageId: "alwaysFalsy", Line: 4, Column: 11},
				{MessageId: "alwaysTruthy", Line: 5, Column: 17},
			},
		},

		// === Comparisons between literal types ===
		{
			Code:   "declare const x: 1;\ndeclare const y: 2;\nif (x === y) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonBetweenLiteralTypes", Line: 3, Column: 5}},
		},
		{
			Code:   "declare const x: 1;\ndeclare const y: 1;\nif (x === y) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonBetweenLiteralTypes", Line: 3, Column: 5}},
		},
		{
			Code:   "declare const x: 'hello';\ndeclare const y: 'world';\nif (x === y) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonBetweenLiteralTypes", Line: 3, Column: 5}},
		},
		{
			Code:   "declare const x: true;\ndeclare const y: false;\nif (x === y) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonBetweenLiteralTypes", Line: 3, Column: 5}},
		},
		// Relational operators on literals
		{
			Code:   "declare const a: '34';\ndeclare const b: '56';\nconst r = a > b;\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonBetweenLiteralTypes", Line: 3, Column: 11}},
		},
		// Float comparisons
		{
			Code:   "const r = 2.3 >= 2.3;\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonBetweenLiteralTypes", Line: 1, Column: 11}},
		},
		// BigInt comparisons
		{
			Code:   "const r = 2n <= 2n;\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonBetweenLiteralTypes", Line: 1, Column: 11}},
		},
		// true vs false, true vs true, true vs undefined
		{
			Code:   "const r = true === false;\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonBetweenLiteralTypes", Line: 1, Column: 11}},
		},
		{
			Code:   "const r = true === true;\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonBetweenLiteralTypes", Line: 1, Column: 11}},
		},
		{
			Code:   "const r = true === undefined;\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonBetweenLiteralTypes", Line: 1, Column: 11}},
		},
		// Single-value function param
		{
			Code:   "function test(a: 'a') { return a === 'a'; }\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonBetweenLiteralTypes", Line: 1, Column: 32}},
		},
		// const narrowing in comparison
		{
			Code:   "const y = 1;\nif (y === 0) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonBetweenLiteralTypes", Line: 2, Column: 5}},
		},
		// Enum literal comparison
		{
			Code:   "enum Foo { a = 1, b = 2 }\nconst x = Foo.a;\nif (x === Foo.a) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonBetweenLiteralTypes", Line: 3, Column: 5}},
		},

		// === Nullish coalescing ===
		{
			Code:   "declare const x: string;\nconst y = x ?? 'default';\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish", Line: 2, Column: 11}},
		},
		{
			Code:   "declare const x: null;\nconst y = x ?? 'default';\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysNullish", Line: 2, Column: 11}},
		},
		{
			Code:   "declare const x: undefined;\nconst y = x ?? 'default';\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysNullish", Line: 2, Column: 11}},
		},
		// string | false is never nullish
		{
			Code:   "function test(a: string | false) { return a ?? 'default'; }\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish", Line: 1, Column: 43}},
		},
		// Generic extends string is never nullish
		{
			Code:   "function test<T extends string>(a: T) { return a ?? 'default'; }\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish", Line: 1, Column: 48}},
		},
		// Generic extends null is always nullish
		{
			Code:   "function test<T extends null>(a: T) { return a ?? 'default'; }\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysNullish", Line: 1, Column: 46}},
		},
		// never type
		{
			Code:   "function test(a: never) { return a ?? 'default'; }\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "never", Line: 1, Column: 34}},
		},

		// === Optional chain on non-nullish ===
		{
			Code: "declare const x: { a: string };\nx?.a;\n",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "neverOptionalChain",
				Line:      2,
				Column:    2,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestRemoveOptionalChain", Output: "declare const x: { a: string };\nx.a;\n"},
				},
			}},
		},
		// Nested optional chains all unnecessary
		{
			Code: "declare const x: { a: { b: string } };\nx?.a?.b;\n",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "neverOptionalChain", Line: 2, Column: 5,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestRemoveOptionalChain", Output: "declare const x: { a: { b: string } };\nx?.a.b;\n"},
					},
				},
				{
					MessageId: "neverOptionalChain", Line: 2, Column: 2,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestRemoveOptionalChain", Output: "declare const x: { a: { b: string } };\nx.a?.b;\n"},
					},
				},
			},
		},
		// Deep chain — only one unnecessary level
		{
			Code: "declare const x: { a?: { b: string } };\nx?.a?.b;\n",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "neverOptionalChain", Line: 2, Column: 2,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestRemoveOptionalChain", Output: "declare const x: { a?: { b: string } };\nx.a?.b;\n"},
				},
			}},
		},
		// Optional chain on call return — non-nullable return
		{
			Code: "type Foo = { bar: () => number } | null;\ndeclare const foo: Foo;\nfoo?.bar()?.toExponential();\n",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "neverOptionalChain", Line: 3, Column: 11,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestRemoveOptionalChain", Output: "type Foo = { bar: () => number } | null;\ndeclare const foo: Foo;\nfoo?.bar().toExponential();\n"},
				},
			}},
		},
		// Optional chain on callable return
		{
			Code: "type Foo = (() => number) | null;\ndeclare const foo: Foo;\nfoo?.()?.toExponential();\n",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "neverOptionalChain", Line: 3, Column: 8,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestRemoveOptionalChain", Output: "type Foo = (() => number) | null;\ndeclare const foo: Foo;\nfoo?.().toExponential();\n"},
				},
			}},
		},
		// Array literal optional access
		{
			Code: "const foo = [1, 2, 3]?.[0];\n",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "neverOptionalChain", Line: 1, Column: 22,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestRemoveOptionalChain", Output: "const foo = [1, 2, 3][0];\n"},
				},
			}},
		},

		// === Array predicate callbacks ===
		{
			Code:   "declare const arr: object[];\narr.filter(t => t);\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 17}},
		},
		{
			Code:   "function truthy() { return []; }\n[1, 3, 5].filter(truthy);\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthyFunc", Line: 2, Column: 18}},
		},
		{
			Code:   "function falsy() {}\n[1, 2, 3].find(falsy);\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsyFunc", Line: 2, Column: 16}},
		},
		{
			Code:   "[1, 3, 5].filter(() => true);\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 1, Column: 24}},
		},
		{
			Code:   "[1, 2, 3].find(() => { return false; });\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy", Line: 1, Column: 31}},
		},
		// Readonly array
		{
			Code:   "function nothing2(x: readonly string[]) { return x.filter(() => false); }\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy", Line: 1, Column: 65}},
		},
		// Tuple
		{
			Code:   "function nothing3(x: [string, string]) { return x.filter(() => false); }\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy", Line: 1, Column: 64}},
		},
		// Generic constrained to true
		{
			Code:   "declare const test: <T extends true>() => T;\n[1, null].filter(test);\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthyFunc", Line: 2, Column: 18}},
		},

		// === Loop conditions ===
		{
			Code:   "while (true) {}",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 1, Column: 8}},
		},
		{
			Code:   "do {} while (true);",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 1, Column: 14}},
		},
		{
			Code:   "for (; true; ) {}",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 1, Column: 8}},
		},
		// only-allowed-literals: for/do-while now also allow true/false/0/1
		// (moved to valid cases; non-0/1 still rejected)
		{
			Code:    "do {} while (0);",
			Options: map[string]interface{}{"allowConstantLoopConditions": "never"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy", Line: 1, Column: 14}},
		},
		// while(2) — not 0/1 literal
		{
			Code:    "while (2) {}",
			Options: map[string]interface{}{"allowConstantLoopConditions": "only-allowed-literals"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 1, Column: 8}},
		},
		// String literal in loop
		{
			Code:    "while ('truthy') {}",
			Options: map[string]interface{}{"allowConstantLoopConditions": "only-allowed-literals"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 1, Column: 8}},
		},
		// declare const test: true — type-level true, not literal true
		{
			Code:    "declare const test: true;\nwhile (test) {}\n",
			Options: map[string]interface{}{"allowConstantLoopConditions": "only-allowed-literals"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 8}},
		},
		// 'never' also rejects type-level true
		{
			Code:    "declare const test: true;\nwhile (test) {}\n",
			Options: map[string]interface{}{"allowConstantLoopConditions": "never"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 8}},
		},
		// declare const test: 1 with only-allowed-literals
		{
			Code:    "declare const test: 1;\nwhile (test) {}\n",
			Options: map[string]interface{}{"allowConstantLoopConditions": "only-allowed-literals"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 8}},
		},

		// === Unary negation ===
		{
			Code:   "declare const obj: { a: string };\nif (!obj) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy", Line: 2, Column: 5}},
		},
		{
			Code:   "declare const obj: { a: string };\nif (!!obj) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 5}},
		},

		// === Never type ===
		{
			Code:   "declare const x: never;\nif (x) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "never", Line: 2, Column: 5}},
		},

		// === No overlap boolean expressions ===
		{
			Code:   "declare const x: string;\nif (x === null) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noOverlapBooleanExpression", Line: 2, Column: 5}},
		},
		{
			Code:   "declare const x: string;\nif (x === undefined) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noOverlapBooleanExpression", Line: 2, Column: 5}},
		},
		// All 8 no-overlap variants on string
		{
			Code: "function test(a: string) {\n  a === undefined;\n  undefined === a;\n  a !== undefined;\n  undefined !== a;\n  a === null;\n  null === a;\n  a !== null;\n  null !== a;\n}\n",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noOverlapBooleanExpression", Line: 2, Column: 3},
				{MessageId: "noOverlapBooleanExpression", Line: 3, Column: 3},
				{MessageId: "noOverlapBooleanExpression", Line: 4, Column: 3},
				{MessageId: "noOverlapBooleanExpression", Line: 5, Column: 3},
				{MessageId: "noOverlapBooleanExpression", Line: 6, Column: 3},
				{MessageId: "noOverlapBooleanExpression", Line: 7, Column: 3},
				{MessageId: "noOverlapBooleanExpression", Line: 8, Column: 3},
				{MessageId: "noOverlapBooleanExpression", Line: 9, Column: 3},
			},
		},
		// Optional string — null comparisons are no-overlap
		{
			Code: "function test(a?: string) {\n  a === null;\n  null === a;\n  a !== null;\n  null !== a;\n}\n",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noOverlapBooleanExpression", Line: 2, Column: 3},
				{MessageId: "noOverlapBooleanExpression", Line: 3, Column: 3},
				{MessageId: "noOverlapBooleanExpression", Line: 4, Column: 3},
				{MessageId: "noOverlapBooleanExpression", Line: 5, Column: 3},
			},
		},
		// Nullable string — undefined strict comparisons are no-overlap
		{
			Code: "function test(a: null | string) {\n  a === undefined;\n  undefined === a;\n  a !== undefined;\n  undefined !== a;\n}\n",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noOverlapBooleanExpression", Line: 2, Column: 3},
				{MessageId: "noOverlapBooleanExpression", Line: 3, Column: 3},
				{MessageId: "noOverlapBooleanExpression", Line: 4, Column: 3},
				{MessageId: "noOverlapBooleanExpression", Line: 5, Column: 3},
			},
		},

		// === Logical assignments ===
		{
			Code:   "declare let x: object;\nx &&= { a: 1 };\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 1}},
		},
		{
			Code:   "declare let x: object;\nx ??= { a: 1 };\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish", Line: 2, Column: 1}},
		},
		// ??= on non-nullish
		{
			Code:   "declare let x: string;\nx ??= 'default';\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish", Line: 2, Column: 1}},
		},
		// &&= on always-truthy
		{
			Code:   "declare const obj: { a: number[] };\nobj.a &&= [1];\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 1}},
		},
		// ||= on always-falsy
		{
			Code:   "declare let x: undefined;\nx ||= true;\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy", Line: 2, Column: 1}},
		},
		// Empty object type with ??= / ||= / &&=
		{
			Code:   "declare let foo: {};\nfoo ??= 1;\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish", Line: 2, Column: 1}},
		},
		{
			Code:   "declare let foo: null;\nfoo ??= null;\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysNullish", Line: 2, Column: 1}},
		},

		// === noUncheckedIndexedAccess ===
		{
			Code:     "declare const arr: string[];\nif (arr[0]) {\n  arr[0] ?? 'foo';\n}\n",
			TSConfig: "tsconfig.noUncheckedIndexedAccess.json",
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish", Line: 3, Column: 3}},
		},
		{
			Code:     "declare const arr: object[];\nif (arr[42] && arr[42]) {}\n",
			TSConfig: "tsconfig.noUncheckedIndexedAccess.json",
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 16}},
		},

		// === noStrictNullCheck ===
		{
			Code:     "if (true) {}",
			TSConfig: "tsconfig.unstrict.json",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noStrictNullCheck", Line: 0, Column: 0},
				{MessageId: "alwaysTruthy", Line: 1, Column: 5},
			},
		},

		// === Parenthesized expressions ===
		{
			Code:   "declare const obj: { a: string };\nif ((obj)) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 6}},
		},

		// === Generic type constraints that are always truthy/falsy ===
		{
			Code:   "function test<T extends object>(t: T) {\n  return t ? 'yes' : 'no';\n}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 10}},
		},
		{
			Code:   "function test<T extends false>(t: T) {\n  return t ? 'yes' : 'no';\n}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy", Line: 2, Column: 10}},
		},
		// Generic extends 'a' | 'b' (non-empty string literals)
		{
			Code:   "function test<T extends 'a' | 'b'>(t: T) {\n  return t ? 'yes' : 'no';\n}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 10}},
		},
		// string & number = never
		{
			Code:   "declare const x: string & number;\ndeclare const b: boolean;\nconst t1 = x && b;\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "never", Line: 3, Column: 12}},
		},

		// === const narrowing ===
		{
			Code:   "const a = true;\nif (!a) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy", Line: 2, Column: 5}},
		},

		// === checkTypePredicates ===
		// type guard already matches
		{
			Code:    "declare function isString(x: unknown): x is string;\ndeclare const s: string;\nisString(s);\n",
			Options: map[string]interface{}{"checkTypePredicates": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "typeGuardAlreadyIsType", Line: 3, Column: 10}},
		},
		// asserts type guard
		{
			Code:    "declare function assertString(x: unknown): asserts x is string;\ndeclare const a: string;\nassertString(a);\n",
			Options: map[string]interface{}{"checkTypePredicates": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "typeGuardAlreadyIsType", Line: 3, Column: 14}},
		},
		// truthiness assertion on always-truthy
		{
			Code:    "declare function assert(x: unknown): asserts x;\nassert('hello');\n",
			Options: map[string]interface{}{"checkTypePredicates": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 8}},
		},
		// truthiness assertion on always-falsy
		{
			Code:    "declare function assert(x: unknown): asserts x;\nassert(false);\n",
			Options: map[string]interface{}{"checkTypePredicates": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy", Line: 2, Column: 8}},
		},
		// truthiness assertion on object
		{
			Code:    "declare function assert(x: unknown): asserts x;\nassert({});\n",
			Options: map[string]interface{}{"checkTypePredicates": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 8}},
		},
		// multi-param asserts — only first param checked
		{
			Code:    "declare function assert(x: unknown, y: unknown): asserts x;\nassert(true, Math.random() > 0.5);\n",
			Options: map[string]interface{}{"checkTypePredicates": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 8}},
		},

		// === switch/case with literal comparison ===
		{
			Code: "declare const x: 'a';\nswitch (x) {\n  case 'b':\n    break;\n}\n",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "comparisonBetweenLiteralTypes", Line: 3, Column: 8},
			},
		},

		// === allowConstantLoopConditions: 'always' still checks non-true conditions ===
		{
			Code:    "declare const x: object;\nwhile (x) {}\n",
			Options: map[string]interface{}{"allowConstantLoopConditions": "always"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 8}},
		},

		// === Branded types that are always truthy/falsy ===
		// Branded falsy string
		{
			Code:   "declare const x: '' & { __brand: string };\ndeclare const b: boolean;\nconst t1 = x && b;\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy", Line: 3, Column: 12}},
		},
		// Branded truthy string literals
		{
			Code:   "declare const x: ('foo' | 'bar') & { __brand: string };\ndeclare const b: boolean;\nconst t1 = x && b;\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 3, Column: 12}},
		},

		// === Record index without noUncheckedIndexedAccess ===
		{
			Code:   "declare const dict: Record<string, object>;\nif (dict['mightNotExist']) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 5}},
		},

		// === Tuple literal access ===
		{
			Code:   "const x = [{}] as [{ foo: string }];\nif (x[0]) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 5}},
		},

		// === Array method as property check ===
		{
			Code:   "declare const arr: object[];\nif (arr.filter) {}\n",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 5}},
		},
	})
}
