// TestStrictBooleanExpressionsUpstream migrates the valid/invalid suite from
// typescript-eslint's strict-boolean-expressions test file
// (packages/eslint-plugin/tests/rules/strict-boolean-expressions.test.ts).
// Position assertions cover line/column for every invalid case. rslint-specific
// lock-in cases live in the strict_boolean_expressions_extras_test.go file.
package strict_boolean_expressions

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestStrictBooleanExpressionsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &StrictBooleanExpressionsRule, []rule_tester.ValidTestCase{
		// boolean in boolean context
		{Code: "true ? 'a' : 'b';"},
		{Code: "\nif (false) {\n}\n"},
		{Code: "while (true) {}"},
		{Code: "for (; false; ) {}"},
		{Code: "!true;"},
		{Code: "false || 123;"},
		{Code: "true && 'foo';"},
		{Code: "!(false || true);"},
		{Code: "true && false ? true : false;"},
		{Code: "(false && true) || false;"},
		{Code: "(false && true) || [];"},
		{Code: "(false && 1) || (true && 2);"},
		{Code: "\ndeclare const x: boolean;\nif (x) {\n}\n"},
		{Code: "(x: boolean) => !x;"},
		{Code: "<T extends boolean>(x: T) => (x ? 1 : 0);"},
		{Code: "\ndeclare const x: never;\nif (x) {\n}\n"},

		// string in boolean context (default allowString=true)
		{Code: "\nif ('') {\n}\n"},
		{Code: "while ('x') {}"},
		{Code: "for (; ''; ) {}"},
		{Code: "('' && '1') || x;"},
		{Code: "\ndeclare const x: string;\nif (x) {\n}\n"},
		{Code: "(x: string) => !x;"},
		{Code: "<T extends string>(x: T) => (x ? 1 : 0);"},

		// number in boolean context (default allowNumber=true)
		{Code: "\nif (0) {\n}\n"},
		{Code: "while (1n) {}"},
		{Code: "for (; Infinity; ) {}"},
		{Code: "(0 / 0 && 1 + 2) || x;"},
		{Code: "\ndeclare const x: number;\nif (x) {\n}\n"},
		{Code: "(x: bigint) => !x;"},
		{Code: "<T extends number>(x: T) => (x ? 1 : 0);"},

		// nullable object in boolean context (default allowNullableObject=true)
		{Code: "\ndeclare const x: null | object;\nif (x) {\n}\n"},
		{Code: "(x?: { a: any }) => !x;"},
		{Code: "<T extends {} | null | undefined>(x: T) => (x ? 1 : 0);"},

		// nullable boolean — allowed when opted in
		{Code: "\n        declare const x: boolean | null;\n        if (x) {\n        }\n      ", Options: map[string]interface{}{"allowNullableBoolean": true}},
		{Code: "\n        (x?: boolean) => !x;\n      ", Options: map[string]interface{}{"allowNullableBoolean": true}},
		{Code: "\n        <T extends boolean | null | undefined>(x: T) => (x ? 1 : 0);\n      ", Options: map[string]interface{}{"allowNullableBoolean": true}},

		// nullable string — allowed when opted in
		{Code: "\n        declare const x: string | null;\n        if (x) {\n        }\n      ", Options: map[string]interface{}{"allowNullableString": true}},
		{Code: "\n        (x?: string) => !x;\n      ", Options: map[string]interface{}{"allowNullableString": true}},
		{Code: "\n        <T extends string | null | undefined>(x: T) => (x ? 1 : 0);\n      ", Options: map[string]interface{}{"allowNullableString": true}},

		// nullable number — allowed when opted in
		{Code: "\n        declare const x: number | null;\n        if (x) {\n        }\n      ", Options: map[string]interface{}{"allowNullableNumber": true}},
		{Code: "\n        (x?: number) => !x;\n      ", Options: map[string]interface{}{"allowNullableNumber": true}},
		{Code: "\n        <T extends number | null | undefined>(x: T) => (x ? 1 : 0);\n      ", Options: map[string]interface{}{"allowNullableNumber": true}},

		// any — allowed when opted in
		{Code: "\n        declare const x: any;\n        if (x) {\n        }\n      ", Options: map[string]interface{}{"allowAny": true}},
		{Code: "\n        x => !x;\n      ", Options: map[string]interface{}{"allowAny": true}},
		{Code: "\n        <T extends any>(x: T) => (x ? 1 : 0);\n      ", Options: map[string]interface{}{"allowAny": true}},

		// logical operator: only outermost operands are checked when not in a condition
		{Code: "\n        1 && true && 'x' && {};\n      ", Options: map[string]interface{}{"allowNumber": true, "allowString": true}},
		{Code: "\n        let x = 0 || false || '' || null;\n      ", Options: map[string]interface{}{"allowNumber": true, "allowString": true}},
		{Code: "\n        if (1 && true && 'x') void 0;\n      ", Options: map[string]interface{}{"allowNumber": true, "allowString": true}},
		{Code: "\n        if (0 || false || '') void 0;\n      ", Options: map[string]interface{}{"allowNumber": true, "allowString": true}},

		// nullable enum — allowed when opted in
		{
			Code: "\n        enum ExampleEnum {\n          This = 0,\n          That = 1,\n        }\n        const rand = Math.random();\n        let theEnum: ExampleEnum | null = null;\n        if (rand < 0.3) {\n          theEnum = ExampleEnum.This;\n        }\n        if (theEnum) {\n        }\n      ",
			Options: map[string]interface{}{"allowNullableEnum": true},
		},
		{
			Code: "\n        enum ExampleEnum {\n          This = 'one',\n          That = 'two',\n        }\n        const rand = Math.random();\n        let theEnum: ExampleEnum | null = null;\n        if (rand < 0.3) {\n          theEnum = ExampleEnum.This;\n        }\n        if (!theEnum) {\n        }\n      ",
			Options: map[string]interface{}{"allowNullableEnum": true},
		},

		// "truthy + nullable" early-exits
		{Code: "\nfunction f(arg: 'a' | null) {\n  if (arg) console.log(arg);\n}\n    "},
		{Code: "\nfunction f(arg: 1 | null) {\n  if (arg) console.log(arg);\n}\n    "},
		{Code: "\ndeclare const x: true | null;\nif (x) {\n}\n    "},

		// branded boolean
		{Code: "\ndeclare const foo: boolean & { __BRAND: 'Foo' };\nif (foo) {\n}\n    "},
		{Code: "\ndeclare const foo: true & { __BRAND: 'Foo' };\nif (foo) {\n}\n    "},
		{Code: "\ndeclare const foo: false & { __BRAND: 'Foo' };\nif (foo) {\n}\n    "},

		// assertion functions: traversed argument is OK when boolean
		{Code: "\ndeclare function assert(a: number, b: unknown): asserts a;\ndeclare const nullableString: string | null;\ndeclare const boo: boolean;\nassert(boo, nullableString);\n    "},
		{Code: "\ndeclare function assert(a: boolean, b: unknown): asserts b is string;\ndeclare const nullableString: string | null;\ndeclare const boo: boolean;\nassert(boo, nullableString);\n    "},
		{Code: "\ndeclare function assert(x: unknown): x is string;\ndeclare const nullableString: string | null;\nassert(nullableString);\n      "},

		// array predicate: boolean / type guard / nullish-checking predicate
		{Code: "\n[true, false].some(function (x) {\n  return x;\n});\n    "},
		{Code: "\n[true, false].some(function check(x) {\n  return x;\n});\n    "},
		{Code: "\n[true, false].some(x => {\n  return x;\n});\n    "},
		{Code: "\n[1, null].filter(function (x) {\n  return x != null;\n});\n    "},
		{Code: "\n['one', 'two', ''].filter(function (x) {\n  return !!x;\n});\n    "},
		{Code: "\n['one', 'two', ''].filter(function (x): boolean {\n  return !!x;\n});\n    "},
		{Code: "\ndeclare const predicate: (x: string) => boolean;\n['one', 'two', ''].filter(predicate);\n    "},
		{Code: "\ndeclare function notNullish<T>(x: T): x is NonNullable<T>;\n['one', null].filter(notNullish);\n    "},
		{Code: "\ndeclare function predicate(x: string | null): x is string;\n['one', null].filter(predicate);\n    "},

		// for-without-condition shouldn't crash
		{Code: "\n      for (let x = 0; ; x++) {\n        break;\n      }\n    "},

		// ---- missing upstream valid: nullable boolean .some() with allowNullableBoolean:true ----
		{
			Code:    "\n        const a: (undefined | boolean | null)[] = [true, undefined, null];\n        a.some(x => x);\n      ",
			Options: map[string]interface{}{"allowNullableBoolean": true},
		},

		// ---- missing upstream valid: nullable number array predicate with allowNullableNumber:true ----
		{
			Code:    "\n        declare const arrayOfArrays: (null | unknown[])[];\n        const isAnyNonEmptyArray1 = arrayOfArrays.some(array => array?.length);\n      ",
			Options: map[string]interface{}{"allowNullableNumber": true},
		},

		// ---- missing upstream valid: any array predicate with allowAny:true ----
		{
			Code:    "\n        declare const arrayOfArrays: any[];\n        const isAnyNonEmptyArray1 = arrayOfArrays.some(array => array);\n      ",
			Options: map[string]interface{}{"allowAny": true},
		},

		// ---- missing upstream valid: logical chain in conditional expression ----
		{
			Code:    "\n        1 && true && 'x' ? {} : null;\n      ",
			Options: map[string]interface{}{"allowNumber": true, "allowString": true},
		},
		{
			Code:    "\n        0 || false || '' ? null : {};\n      ",
			Options: map[string]interface{}{"allowNumber": true, "allowString": true},
		},

		// ---- missing upstream valid: array predicate body returning primitive with matching allow* ----
		{
			Code:    "\n        declare const arrayOfArrays: string[];\n        const isAnyNonEmptyArray1 = arrayOfArrays.some(array => array);\n      ",
			Options: map[string]interface{}{"allowString": true},
		},
		{
			Code:    "\n        declare const arrayOfArrays: number[];\n        const isAnyNonEmptyArray1 = arrayOfArrays.some(array => array);\n      ",
			Options: map[string]interface{}{"allowNumber": true},
		},
		{
			Code:    "\n        declare const arrayOfArrays: (null | object)[];\n        const isAnyNonEmptyArray1 = arrayOfArrays.some(array => array);\n      ",
			Options: map[string]interface{}{"allowNullableObject": true},
		},

		// ---- missing upstream valid: nullable enum truthy-only with allowNullableEnum:true ----
		{
			Code:    "\n        enum ExampleEnum {\n          This = 1,\n          That = 2,\n        }\n        const rand = Math.random();\n        let theEnum: ExampleEnum | null = null;\n        if (rand < 0.3) {\n          theEnum = ExampleEnum.This;\n        }\n        if (!theEnum) {\n        }\n      ",
			Options: map[string]interface{}{"allowNullableEnum": true},
		},

		// ---- missing upstream valid: nullable mixed enum variants ----
		{
			// falsy number + truthy string
			Code:    "\n        enum ExampleEnum {\n          This = 0,\n          That = 'one',\n        }\n        (value?: ExampleEnum) => (value ? 1 : 0);\n      ",
			Options: map[string]interface{}{"allowNullableEnum": true},
		},
		{
			// falsy string + truthy number
			Code:    "\n        enum ExampleEnum {\n          This = '',\n          That = 1,\n        }\n        (value?: ExampleEnum) => (!value ? 1 : 0);\n      ",
			Options: map[string]interface{}{"allowNullableEnum": true},
		},
		{
			// truthy string + truthy number
			Code:    "\n        enum ExampleEnum {\n          This = 'this',\n          That = 1,\n        }\n        (value?: ExampleEnum) => (!value ? 1 : 0);\n      ",
			Options: map[string]interface{}{"allowNullableEnum": true},
		},
		{
			// falsy string + falsy number
			Code:    "\n        enum ExampleEnum {\n          This = '',\n          That = 0,\n        }\n        (value?: ExampleEnum) => (!value ? 1 : 0);\n      ",
			Options: map[string]interface{}{"allowNullableEnum": true},
		},
		{
			// (ExampleEnum | null)[] array predicate with allowNullableEnum:true
			Code:    "\n        enum ExampleEnum {\n          This = '',\n          That = 0,\n        }\n        declare const arrayOfArrays: (ExampleEnum | null)[];\n        const isAnyNonEmptyArray1 = arrayOfArrays.some(array => array);\n      ",
			Options: map[string]interface{}{"allowNullableEnum": true},
		},

		// ---- missing upstream valid: unstrict tsconfig opt-out ----
		{
			Code:     "\ndeclare const x: string[] | null;\n// eslint-disable-next-line\nif (x) {\n}\n      ",
			TSConfig: "tsconfig.unstrict.json",
			Options:  map[string]interface{}{"allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing": true},
		},

		// ---- missing upstream valid: 'a' | 'b' | null with default allowString ----
		{Code: "\nfunction f(arg: 'a' | 'b' | null) {\n  if (arg) console.log(arg);\n}\n    "},
		{
			Code:    "\ndeclare const x: 1 | null;\ndeclare const y: 1;\nif (x) {\n}\nif (y) {\n}\n      ",
			Options: map[string]interface{}{"allowNumber": true},
		},
		{Code: "\nfunction f(arg: 1 | 2 | null) {\n  if (arg) console.log(arg);\n}\n    "},
		{Code: "\ninterface Options {\n  readonly enableSomething?: true;\n}\n\nfunction f(opts: Options): void {\n  if (opts.enableSomething) console.log('Do something');\n}\n    "},
		{
			Code:    "\ndeclare const x: 'a' | null;\ndeclare const y: 'a';\nif (x) {\n}\nif (y) {\n}\n      ",
			Options: map[string]interface{}{"allowString": true},
		},

		// ---- missing upstream valid: assertion function variants ----
		// asserts asserts b refers to b which is `boo: boolean` (param at index 1) → valid
		{Code: "\ndeclare function assert(a: number, b: unknown): asserts b;\ndeclare const nullableString: string | null;\ndeclare const boo: boolean;\nassert(nullableString, boo);\n    "},
		// spread argument — predicate parameter index can't be reliably mapped
		{Code: "\ndeclare function assert(a: number, b: unknown): asserts b;\ndeclare const nullableString: string | null;\ndeclare const boo: boolean;\nassert(...nullableString, nullableString);\n    "},
		// method-call assertion via `o.assert(...)`
		{Code: "\ndeclare function assert(\n  this: object,\n  a: number,\n  b?: unknown,\n  c?: unknown,\n): asserts c;\ndeclare const nullableString: string | null;\ndeclare const foo: number;\nconst o: { assert: typeof assert } = {\n  assert,\n};\no.assert(foo, nullableString);\n    "},
		// asserts this
		{Code: "\nclass ThisAsserter {\n  assertThis(this: unknown, arg2: unknown): asserts this {}\n}\n\ndeclare const lol: string | number | unknown | null;\n\nconst thisAsserter: ThisAsserter = new ThisAsserter();\nthisAsserter.assertThis(lol);\n      "},
		// multi-overload — the void overload matches and there's no asserts predicate, so no traversal of nullableString
		{Code: "\nfunction assert(this: object, a: number, b: unknown): asserts b;\nfunction assert(a: bigint, b: unknown): asserts b;\nfunction assert(this: object, a: string, two: string): asserts two;\nfunction assert(\n  this: object,\n  a: string,\n  assertee: string,\n  c: bigint,\n  d: object,\n): asserts assertee;\nfunction assert(...args: any[]): void;\n\nfunction assert(...args: any[]) {\n  throw new Error('lol');\n}\n\ndeclare const nullableString: string | null;\nassert(3 as any, nullableString);\n      "},
		// any-typed `assert` call has no call signature info → no truthiness assertion detected
		{Code: "\ndeclare const assert: any;\ndeclare const nullableString: string | null;\nassert(nullableString);\n    "},

		// ---- missing upstream valid: array predicate returning boolean (conditional return path) ----
		{Code: "\n['one', 'two', ''].filter(function (x): boolean {\n  if (x) {\n    return true;\n  }\n});\n    "},
		{Code: "\n['one', 'two', ''].filter(function (x): boolean {\n  if (x) {\n    return true;\n  }\n\n  throw new Error('oops');\n});\n    "},
		// predicate function whose param is parameter-less (no type annotation)
		{Code: "\ndeclare const predicate: (string) => boolean;\n['one', 'two', ''].filter(predicate);\n    "},
		// type-guard return T inferred to boolean
		{Code: "\ndeclare function predicate<T extends boolean>(x: string | null): T;\n['one', null].filter(predicate);\n    "},
		// overloaded predicate returning boolean
		{Code: "\ndeclare function f(x: number): boolean;\ndeclare function f(x: string | null): boolean;\n\n[35].filter(f);\n    "},
	}, []rule_tester.InvalidTestCase{
		// ---- non-boolean in RHS of test expression ----
		{
			Code:    "\nif (true && 1 + 1) {\n}\n      ",
			Options: map[string]interface{}{"allowNullableObject": false, "allowNumber": false, "allowString": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNumber", Line: 2, Column: 13,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareZero", Output: "\nif (true && ((1 + 1) !== 0)) {\n}\n      "},
					{MessageId: "conditionFixCompareNaN", Output: "\nif (true && (!Number.isNaN((1 + 1)))) {\n}\n      "},
					{MessageId: "conditionFixCastBoolean", Output: "\nif (true && (Boolean((1 + 1)))) {\n}\n      "},
				},
			}},
		},
		{
			Code:    "while (false || 'a' + 'b') {}",
			Options: map[string]interface{}{"allowNullableObject": false, "allowNumber": false, "allowString": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorString", Line: 1, Column: 17,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareStringLength", Output: "while (false || (('a' + 'b').length > 0)) {}"},
					{MessageId: "conditionFixCompareEmptyString", Output: `while (false || (('a' + 'b') !== "")) {}`},
					{MessageId: "conditionFixCastBoolean", Output: "while (false || (Boolean(('a' + 'b')))) {}"},
				},
			}},
		},
		{
			// Object error path has no suggestions.
			Code:    "(x: object) => (true || false || x ? true : false);",
			Options: map[string]interface{}{"allowNullableObject": false, "allowNumber": false, "allowString": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorObject", Line: 1, Column: 34,
			}},
		},

		// ---- last logical operand skipped when used for control flow ----
		{
			Code:    "'asd' && 123 && [] && null;",
			Options: map[string]interface{}{"allowNumber": false, "allowString": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorString", Line: 1, Column: 1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "conditionFixCompareStringLength", Output: "('asd'.length > 0) && 123 && [] && null;"},
						{MessageId: "conditionFixCompareEmptyString", Output: `('asd' !== "") && 123 && [] && null;`},
						{MessageId: "conditionFixCastBoolean", Output: "(Boolean('asd')) && 123 && [] && null;"},
					},
				},
				{
					MessageId: "conditionErrorNumber", Line: 1, Column: 10,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "conditionFixCompareZero", Output: "'asd' && (123 !== 0) && [] && null;"},
						{MessageId: "conditionFixCompareNaN", Output: "'asd' && (!Number.isNaN(123)) && [] && null;"},
						{MessageId: "conditionFixCastBoolean", Output: "'asd' && (Boolean(123)) && [] && null;"},
					},
				},
				{MessageId: "conditionErrorObject", Line: 1, Column: 17},
			},
		},
		{
			Code:    "'asd' || 123 || [] || null;",
			Options: map[string]interface{}{"allowNumber": false, "allowString": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorString", Line: 1, Column: 1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "conditionFixCompareStringLength", Output: "('asd'.length > 0) || 123 || [] || null;"},
						{MessageId: "conditionFixCompareEmptyString", Output: `('asd' !== "") || 123 || [] || null;`},
						{MessageId: "conditionFixCastBoolean", Output: "(Boolean('asd')) || 123 || [] || null;"},
					},
				},
				{
					MessageId: "conditionErrorNumber", Line: 1, Column: 10,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "conditionFixCompareZero", Output: "'asd' || (123 !== 0) || [] || null;"},
						{MessageId: "conditionFixCompareNaN", Output: "'asd' || (!Number.isNaN(123)) || [] || null;"},
						{MessageId: "conditionFixCastBoolean", Output: "'asd' || (Boolean(123)) || [] || null;"},
					},
				},
				{MessageId: "conditionErrorObject", Line: 1, Column: 17},
			},
		},

		// ---- nullish in boolean context (no suggestions) ----
		{Code: "null || {};", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionErrorNullish", Line: 1, Column: 1}}},
		{Code: "undefined && [];", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionErrorNullish", Line: 1, Column: 1}}},
		{Code: "\ndeclare const x: null;\nif (x) {\n}\n      ", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionErrorNullish", Line: 3, Column: 5}}},
		{Code: "(x: undefined) => !x;", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionErrorNullish", Line: 1, Column: 20}}},
		{Code: "<T extends null | undefined>(x: T) => (x ? 1 : 0);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionErrorNullish", Line: 1, Column: 40}}},
		{Code: "<T extends null>(x: T) => (x ? 1 : 0);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionErrorNullish", Line: 1, Column: 28}}},
		{Code: "<T extends undefined>(x: T) => (x ? 1 : 0);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionErrorNullish", Line: 1, Column: 33}}},

		// ---- object in boolean context (no suggestions) ----
		{Code: "[] || 1;", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionErrorObject", Line: 1, Column: 1}}},
		{Code: "({}) && 'a';", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionErrorObject", Line: 1, Column: 2}}},
		{Code: "\ndeclare const x: symbol;\nif (x) {\n}\n      ", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionErrorObject", Line: 3, Column: 5}}},
		{Code: "(x: () => void) => !x;", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionErrorObject", Line: 1, Column: 21}}},
		{Code: "<T extends object>(x: T) => (x ? 1 : 0);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionErrorObject", Line: 1, Column: 30}}},
		{Code: "<T extends Object | Function>(x: T) => (x ? 1 : 0);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionErrorObject", Line: 1, Column: 41}}},
		{Code: "<T extends { a: number }>(x: T) => (x ? 1 : 0);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionErrorObject", Line: 1, Column: 37}}},
		{Code: "<T extends () => void>(x: T) => (x ? 1 : 0);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionErrorObject", Line: 1, Column: 34}}},

		// ---- string in boolean context with allowString=false ----
		{
			Code:    "while ('') {}",
			Options: map[string]interface{}{"allowString": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorString", Line: 1, Column: 8,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareStringLength", Output: "while (''.length > 0) {}"},
					{MessageId: "conditionFixCompareEmptyString", Output: `while ('' !== "") {}`},
					{MessageId: "conditionFixCastBoolean", Output: "while (Boolean('')) {}"},
				},
			}},
		},
		{
			Code:    "for (; 'foo'; ) {}",
			Options: map[string]interface{}{"allowString": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorString", Line: 1, Column: 8,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareStringLength", Output: "for (; 'foo'.length > 0; ) {}"},
					{MessageId: "conditionFixCompareEmptyString", Output: `for (; 'foo' !== ""; ) {}`},
					{MessageId: "conditionFixCastBoolean", Output: "for (; Boolean('foo'); ) {}"},
				},
			}},
		},
		{
			Code:    "\ndeclare const x: string;\nif (x) {\n}\n      ",
			Options: map[string]interface{}{"allowString": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorString", Line: 3, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareStringLength", Output: "\ndeclare const x: string;\nif (x.length > 0) {\n}\n      "},
					{MessageId: "conditionFixCompareEmptyString", Output: "\ndeclare const x: string;\nif (x !== \"\") {\n}\n      "},
					{MessageId: "conditionFixCastBoolean", Output: "\ndeclare const x: string;\nif (Boolean(x)) {\n}\n      "},
				},
			}},
		},
		{
			Code:    "(x: string) => !x;",
			Options: map[string]interface{}{"allowString": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorString", Line: 1, Column: 17,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareStringLength", Output: "(x: string) => x.length === 0;"},
					{MessageId: "conditionFixCompareEmptyString", Output: `(x: string) => x === "";`},
					{MessageId: "conditionFixCastBoolean", Output: "(x: string) => !Boolean(x);"},
				},
			}},
		},
		{
			Code:    "<T extends string>(x: T) => (x ? 1 : 0);",
			Options: map[string]interface{}{"allowString": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorString", Line: 1, Column: 30,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareStringLength", Output: "<T extends string>(x: T) => ((x.length > 0) ? 1 : 0);"},
					{MessageId: "conditionFixCompareEmptyString", Output: `<T extends string>(x: T) => ((x !== "") ? 1 : 0);`},
					{MessageId: "conditionFixCastBoolean", Output: "<T extends string>(x: T) => ((Boolean(x)) ? 1 : 0);"},
				},
			}},
		},

		// ---- number with allowNumber=false ----
		{
			Code:    "while (0n) {}",
			Options: map[string]interface{}{"allowNumber": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNumber", Line: 1, Column: 8,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareZero", Output: "while (0n !== 0) {}"},
					{MessageId: "conditionFixCompareNaN", Output: "while (!Number.isNaN(0n)) {}"},
					{MessageId: "conditionFixCastBoolean", Output: "while (Boolean(0n)) {}"},
				},
			}},
		},
		{
			Code:    "for (; 123; ) {}",
			Options: map[string]interface{}{"allowNumber": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNumber", Line: 1, Column: 8,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareZero", Output: "for (; 123 !== 0; ) {}"},
					{MessageId: "conditionFixCompareNaN", Output: "for (; !Number.isNaN(123); ) {}"},
					{MessageId: "conditionFixCastBoolean", Output: "for (; Boolean(123); ) {}"},
				},
			}},
		},
		{
			Code:    "\ndeclare const x: number;\nif (x) {\n}\n      ",
			Options: map[string]interface{}{"allowNumber": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNumber", Line: 3, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareZero", Output: "\ndeclare const x: number;\nif (x !== 0) {\n}\n      "},
					{MessageId: "conditionFixCompareNaN", Output: "\ndeclare const x: number;\nif (!Number.isNaN(x)) {\n}\n      "},
					{MessageId: "conditionFixCastBoolean", Output: "\ndeclare const x: number;\nif (Boolean(x)) {\n}\n      "},
				},
			}},
		},
		{
			Code:    "(x: bigint) => !x;",
			Options: map[string]interface{}{"allowNumber": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNumber", Line: 1, Column: 17,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareZero", Output: "(x: bigint) => x === 0;"},
					{MessageId: "conditionFixCompareNaN", Output: "(x: bigint) => Number.isNaN(x);"},
					{MessageId: "conditionFixCastBoolean", Output: "(x: bigint) => !Boolean(x);"},
				},
			}},
		},
		{
			Code:    "<T extends number>(x: T) => (x ? 1 : 0);",
			Options: map[string]interface{}{"allowNumber": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNumber", Line: 1, Column: 30,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareZero", Output: "<T extends number>(x: T) => ((x !== 0) ? 1 : 0);"},
					{MessageId: "conditionFixCompareNaN", Output: "<T extends number>(x: T) => ((!Number.isNaN(x)) ? 1 : 0);"},
					{MessageId: "conditionFixCastBoolean", Output: "<T extends number>(x: T) => ((Boolean(x)) ? 1 : 0);"},
				},
			}},
		},
		{
			Code:    "![]['length']; // doesn't count as array.length when computed",
			Options: map[string]interface{}{"allowNumber": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNumber", Line: 1, Column: 2,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareZero", Output: "[]['length'] === 0; // doesn't count as array.length when computed"},
					{MessageId: "conditionFixCompareNaN", Output: "Number.isNaN([]['length']); // doesn't count as array.length when computed"},
					{MessageId: "conditionFixCastBoolean", Output: "!Boolean([]['length']); // doesn't count as array.length when computed"},
				},
			}},
		},
		{
			Code:    "\ndeclare const a: any[] & { notLength: number };\nif (a.notLength) {\n}\n      ",
			Options: map[string]interface{}{"allowNumber": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNumber", Line: 3, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareZero", Output: "\ndeclare const a: any[] & { notLength: number };\nif (a.notLength !== 0) {\n}\n      "},
					{MessageId: "conditionFixCompareNaN", Output: "\ndeclare const a: any[] & { notLength: number };\nif (!Number.isNaN(a.notLength)) {\n}\n      "},
					{MessageId: "conditionFixCastBoolean", Output: "\ndeclare const a: any[] & { notLength: number };\nif (Boolean(a.notLength)) {\n}\n      "},
				},
			}},
		},

		// ---- number (array.length) variant ----
		{
			Code:    "\nif (![].length) {\n}\n      ",
			Options: map[string]interface{}{"allowNumber": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNumber", Line: 2, Column: 6,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareArrayLengthZero", Output: "\nif ([].length === 0) {\n}\n      "},
				},
			}},
		},
		{
			Code:    "\n(a: number[]) => a.length && '...';\n      ",
			Options: map[string]interface{}{"allowNumber": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNumber", Line: 2, Column: 18,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareArrayLengthNonzero", Output: "\n(a: number[]) => (a.length > 0) && '...';\n      "},
				},
			}},
		},
		{
			Code:    "\n<T extends unknown[]>(...a: T) => a.length || 'empty';\n      ",
			Options: map[string]interface{}{"allowNumber": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNumber", Line: 2, Column: 35,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareArrayLengthNonzero", Output: "\n<T extends unknown[]>(...a: T) => (a.length > 0) || 'empty';\n      "},
				},
			}},
		},

		// ---- mixed string|number (conditionErrorOther — no suggestions) ----
		{
			Code:    "\ndeclare const x: string | number;\nif (x) {\n}\n      ",
			Options: map[string]interface{}{"allowNumber": true, "allowString": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "conditionErrorOther", Line: 3, Column: 5}},
		},
		{
			Code:    "(x: bigint | string) => !x;",
			Options: map[string]interface{}{"allowNumber": true, "allowString": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "conditionErrorOther", Line: 1, Column: 26}},
		},
		{
			Code:    "<T extends number | bigint | string>(x: T) => (x ? 1 : 0);",
			Options: map[string]interface{}{"allowNumber": true, "allowString": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "conditionErrorOther", Line: 1, Column: 48}},
		},

		// ---- nullable boolean ----
		{
			Code:    "\ndeclare const x: boolean | null;\nif (x) {\n}\n      ",
			Options: map[string]interface{}{"allowNullableBoolean": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableBoolean", Line: 3, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixDefaultFalse", Output: "\ndeclare const x: boolean | null;\nif (x ?? false) {\n}\n      "},
					{MessageId: "conditionFixCompareTrue", Output: "\ndeclare const x: boolean | null;\nif (x === true) {\n}\n      "},
				},
			}},
		},
		{
			Code:    "(x?: boolean) => !x;",
			Options: map[string]interface{}{"allowNullableBoolean": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableBoolean", Line: 1, Column: 19,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixDefaultFalse", Output: "(x?: boolean) => !(x ?? false);"},
					{MessageId: "conditionFixCompareFalse", Output: "(x?: boolean) => x === false;"},
				},
			}},
		},
		{
			Code:    "<T extends boolean | null | undefined>(x: T) => (x ? 1 : 0);",
			Options: map[string]interface{}{"allowNullableBoolean": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableBoolean", Line: 1, Column: 50,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixDefaultFalse", Output: "<T extends boolean | null | undefined>(x: T) => ((x ?? false) ? 1 : 0);"},
					{MessageId: "conditionFixCompareTrue", Output: "<T extends boolean | null | undefined>(x: T) => ((x === true) ? 1 : 0);"},
				},
			}},
		},

		// ---- nullable object ----
		{
			Code:    "\ndeclare const x: object | null;\nif (x) {\n}\n      ",
			Options: map[string]interface{}{"allowNullableObject": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableObject", Line: 3, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "\ndeclare const x: object | null;\nif (x != null) {\n}\n      "},
				},
			}},
		},
		{
			Code:    "(x?: { a: number }) => !x;",
			Options: map[string]interface{}{"allowNullableObject": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableObject", Line: 1, Column: 25,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "(x?: { a: number }) => x == null;"},
				},
			}},
		},
		{
			Code:    "<T extends {} | null | undefined>(x: T) => (x ? 1 : 0);",
			Options: map[string]interface{}{"allowNullableObject": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableObject", Line: 1, Column: 45,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "<T extends {} | null | undefined>(x: T) => ((x != null) ? 1 : 0);"},
				},
			}},
		},

		// ---- nullable string (default — allowNullableString=false) ----
		{
			Code: "\ndeclare const x: string | null;\nif (x) {\n}\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableString", Line: 3, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "\ndeclare const x: string | null;\nif (x != null) {\n}\n      "},
					{MessageId: "conditionFixDefaultEmptyString", Output: "\ndeclare const x: string | null;\nif (x ?? \"\") {\n}\n      "},
					{MessageId: "conditionFixCastBoolean", Output: "\ndeclare const x: string | null;\nif (Boolean(x)) {\n}\n      "},
				},
			}},
		},
		{
			Code: "(x?: string) => !x;",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableString", Line: 1, Column: 18,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "(x?: string) => x == null;"},
					{MessageId: "conditionFixDefaultEmptyString", Output: `(x?: string) => !(x ?? "");`},
					{MessageId: "conditionFixCastBoolean", Output: "(x?: string) => !Boolean(x);"},
				},
			}},
		},
		{
			Code: "<T extends string | null | undefined>(x: T) => (x ? 1 : 0);",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableString", Line: 1, Column: 49,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "<T extends string | null | undefined>(x: T) => ((x != null) ? 1 : 0);"},
					{MessageId: "conditionFixDefaultEmptyString", Output: `<T extends string | null | undefined>(x: T) => ((x ?? "") ? 1 : 0);`},
					{MessageId: "conditionFixCastBoolean", Output: "<T extends string | null | undefined>(x: T) => ((Boolean(x)) ? 1 : 0);"},
				},
			}},
		},
		{
			Code: "\nfunction foo(x: '' | 'bar' | null) {\n  if (!x) {\n  }\n}\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableString", Line: 3, Column: 8,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "\nfunction foo(x: '' | 'bar' | null) {\n  if (x == null) {\n  }\n}\n      "},
					{MessageId: "conditionFixDefaultEmptyString", Output: "\nfunction foo(x: '' | 'bar' | null) {\n  if (!(x ?? \"\")) {\n  }\n}\n      "},
					{MessageId: "conditionFixCastBoolean", Output: "\nfunction foo(x: '' | 'bar' | null) {\n  if (!Boolean(x)) {\n  }\n}\n      "},
				},
			}},
		},

		// ---- nullable number (default) ----
		{
			Code: "\ndeclare const x: number | null;\nif (x) {\n}\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableNumber", Line: 3, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "\ndeclare const x: number | null;\nif (x != null) {\n}\n      "},
					{MessageId: "conditionFixDefaultZero", Output: "\ndeclare const x: number | null;\nif (x ?? 0) {\n}\n      "},
					{MessageId: "conditionFixCastBoolean", Output: "\ndeclare const x: number | null;\nif (Boolean(x)) {\n}\n      "},
				},
			}},
		},
		{
			Code: "(x?: number) => !x;",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableNumber", Line: 1, Column: 18,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "(x?: number) => x == null;"},
					{MessageId: "conditionFixDefaultZero", Output: "(x?: number) => !(x ?? 0);"},
					{MessageId: "conditionFixCastBoolean", Output: "(x?: number) => !Boolean(x);"},
				},
			}},
		},
		{
			Code: "<T extends number | null | undefined>(x: T) => (x ? 1 : 0);",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableNumber", Line: 1, Column: 49,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "<T extends number | null | undefined>(x: T) => ((x != null) ? 1 : 0);"},
					{MessageId: "conditionFixDefaultZero", Output: "<T extends number | null | undefined>(x: T) => ((x ?? 0) ? 1 : 0);"},
					{MessageId: "conditionFixCastBoolean", Output: "<T extends number | null | undefined>(x: T) => ((Boolean(x)) ? 1 : 0);"},
				},
			}},
		},
		{
			Code: "\nfunction foo(x: 0 | 1 | null) {\n  if (!x) {\n  }\n}\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableNumber", Line: 3, Column: 8,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "\nfunction foo(x: 0 | 1 | null) {\n  if (x == null) {\n  }\n}\n      "},
					{MessageId: "conditionFixDefaultZero", Output: "\nfunction foo(x: 0 | 1 | null) {\n  if (!(x ?? 0)) {\n  }\n}\n      "},
					{MessageId: "conditionFixCastBoolean", Output: "\nfunction foo(x: 0 | 1 | null) {\n  if (!Boolean(x)) {\n  }\n}\n      "},
				},
			}},
		},

		// ---- nullable enum ----
		{
			Code:    "\n        enum ExampleEnum {\n          This = 0,\n          That = 1,\n        }\n        const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;\n        if (theEnum) {\n        }\n      ",
			Options: map[string]interface{}{"allowNullableEnum": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableEnum", Line: 7, Column: 13, EndLine: 7, EndColumn: 20,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "\n        enum ExampleEnum {\n          This = 0,\n          That = 1,\n        }\n        const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;\n        if (theEnum != null) {\n        }\n      "},
				},
			}},
		},
		{
			Code:    "\n        enum ExampleEnum {\n          This = 0,\n          That = 1,\n        }\n        const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;\n        if (!theEnum) {\n        }\n      ",
			Options: map[string]interface{}{"allowNullableEnum": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableEnum", Line: 7, Column: 14, EndLine: 7, EndColumn: 21,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "\n        enum ExampleEnum {\n          This = 0,\n          That = 1,\n        }\n        const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;\n        if (theEnum == null) {\n        }\n      "},
				},
			}},
		},
		{
			Code:    "\n        enum ExampleEnum {\n          This = 'one',\n          That = 'two',\n        }\n        const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;\n        if (!theEnum) {\n        }\n      ",
			Options: map[string]interface{}{"allowNullableEnum": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableEnum", Line: 7, Column: 14, EndLine: 7, EndColumn: 21,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "\n        enum ExampleEnum {\n          This = 'one',\n          That = 'two',\n        }\n        const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;\n        if (theEnum == null) {\n        }\n      "},
				},
			}},
		},

		// ---- nullable mixed enum ----
		{
			Code:    "\n        enum ExampleEnum {\n          This = 0,\n          That = 'one',\n        }\n        (value?: ExampleEnum) => (value ? 1 : 0);\n      ",
			Options: map[string]interface{}{"allowNullableEnum": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableEnum", Line: 6, Column: 35, EndLine: 6, EndColumn: 40,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "\n        enum ExampleEnum {\n          This = 0,\n          That = 'one',\n        }\n        (value?: ExampleEnum) => ((value != null) ? 1 : 0);\n      "},
				},
			}},
		},

		// ---- any ----
		{
			Code: "\nif (x) {\n}\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorAny", Line: 2, Column: 5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCastBoolean", Output: "\nif (Boolean(x)) {\n}\n      "},
				},
			}},
		},
		{
			Code: "x => !x;",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorAny", Line: 1, Column: 7,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCastBoolean", Output: "x => !(Boolean(x));"},
				},
			}},
		},
		{
			Code: "<T extends any>(x: T) => (x ? 1 : 0);",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorAny", Line: 1, Column: 27,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCastBoolean", Output: "<T extends any>(x: T) => ((Boolean(x)) ? 1 : 0);"},
				},
			}},
		},
		{
			Code: "<T,>(x: T) => (x ? 1 : 0);",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorAny", Line: 1, Column: 16,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCastBoolean", Output: "<T,>(x: T) => ((Boolean(x)) ? 1 : 0);"},
				},
			}},
		},

		// ---- noStrictNullCheck ----
		{
			Code:     "\ndeclare const x: string[] | null;\nif (x) {\n}\n      ",
			TSConfig: "tsconfig.unstrict.json",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noStrictNullCheck", Line: 0, Column: 0},
				{MessageId: "conditionErrorObject", Line: 3, Column: 5},
			},
		},

		// ---- assertion function argument ----
		{
			Code: "\ndeclare function assert(x: unknown): asserts x;\ndeclare const nullableString: string | null;\nassert(nullableString);\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableString", Line: 4, Column: 8,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "\ndeclare function assert(x: unknown): asserts x;\ndeclare const nullableString: string | null;\nassert(nullableString != null);\n      "},
					{MessageId: "conditionFixDefaultEmptyString", Output: "\ndeclare function assert(x: unknown): asserts x;\ndeclare const nullableString: string | null;\nassert(nullableString ?? \"\");\n      "},
					{MessageId: "conditionFixCastBoolean", Output: "\ndeclare function assert(x: unknown): asserts x;\ndeclare const nullableString: string | null;\nassert(Boolean(nullableString));\n      "},
				},
			}},
		},
		{
			Code: "\ndeclare function assert(a: number, b: unknown): asserts b;\ndeclare const nullableString: string | null;\nassert(foo, nullableString);\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableString", Line: 4, Column: 13,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "\ndeclare function assert(a: number, b: unknown): asserts b;\ndeclare const nullableString: string | null;\nassert(foo, nullableString != null);\n      "},
					{MessageId: "conditionFixDefaultEmptyString", Output: "\ndeclare function assert(a: number, b: unknown): asserts b;\ndeclare const nullableString: string | null;\nassert(foo, nullableString ?? \"\");\n      "},
					{MessageId: "conditionFixCastBoolean", Output: "\ndeclare function assert(a: number, b: unknown): asserts b;\ndeclare const nullableString: string | null;\nassert(foo, Boolean(nullableString));\n      "},
				},
			}},
		},

		// ---- array predicate: explicit-boolean-return suggestion alongside the conditional fixes ----
		{
			Code:    "\n        declare const array: string[];\n        array.some(x => x);\n      ",
			Options: map[string]interface{}{"allowNullableBoolean": true, "allowString": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorString",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareStringLength", Output: "\n        declare const array: string[];\n        array.some(x => x.length > 0);\n      "},
					{MessageId: "conditionFixCompareEmptyString", Output: "\n        declare const array: string[];\n        array.some(x => x !== \"\");\n      "},
					{MessageId: "conditionFixCastBoolean", Output: "\n        declare const array: string[];\n        array.some(x => Boolean(x));\n      "},
					{MessageId: "explicitBooleanReturnType", Output: "\n        declare const array: string[];\n        array.some((x): boolean => x);\n      "},
				},
			}},
		},

		// ---- predicate cannot be async (no suggestions per upstream) ----
		{
			Code: "\n[1, null].every(async x => {\n  return x != null;\n});\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "predicateCannotBeAsync", Line: 2, Column: 17, EndLine: 4, EndColumn: 2,
			}},
		},

		// ---- predicate returning non-boolean ----
		{
			Code: "\n[1, null].every((x): boolean | number => {\n  return x != null;\n});\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorOther", Line: 2, Column: 17, EndLine: 4, EndColumn: 2,
			}},
		},
		{
			Code: "\n[1, null].every((x): boolean | undefined => {\n  return x != null;\n});\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableBoolean", Line: 2, Column: 17, EndLine: 4, EndColumn: 2,
			}},
		},

		// ---- predicate suggestion fix: explicitBooleanReturnType across function shapes ----
		{
			Code: "\n[1, null].every((x, i) => {});\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullish", Line: 2, Column: 17, EndLine: 2, EndColumn: 29,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "explicitBooleanReturnType", Output: "\n[1, null].every((x, i): boolean => {});\n      "},
				},
			}},
		},
		{
			Code: "\n[() => {}, null].every((x: () => void) => {});\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullish", Line: 2, Column: 24, EndLine: 2, EndColumn: 45,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "explicitBooleanReturnType", Output: "\n[() => {}, null].every((x: () => void): boolean => {});\n      "},
				},
			}},
		},
		{
			Code: "\n[() => {}, null].every(function (x: () => void) {});\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullish", Line: 2, Column: 24, EndLine: 2, EndColumn: 51,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "explicitBooleanReturnType", Output: "\n[() => {}, null].every(function (x: () => void): boolean {});\n      "},
				},
			}},
		},
		{
			Code: "\n[() => {}, null].every(() => {});\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullish", Line: 2, Column: 24, EndLine: 2, EndColumn: 32,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "explicitBooleanReturnType", Output: "\n[() => {}, null].every((): boolean => {});\n      "},
				},
			}},
		},

		// ---- function overloading ----
		{
			Code: "\ndeclare function f(x: number): string;\ndeclare function f(x: string | null): boolean;\n\n[35].filter(f);\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorOther", Line: 5, Column: 13, EndLine: 5, EndColumn: 14,
			}},
		},

		// ---- type constraint on predicate ----
		{
			Code: "\ndeclare function foo<T>(x: number): T;\n[1, null].every(foo);\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorAny", Line: 3, Column: 17, EndLine: 3, EndColumn: 20,
			}},
		},

		// ---- non-boolean RHS: 4-error logical chain (upstream invalid #4) ----
		{
			Code:    `if (('' && {}) || (0 && void 0)) { }`,
			Options: map[string]interface{}{"allowNullableObject": false, "allowNumber": false, "allowString": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorString", Line: 1, Column: 6,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "conditionFixCompareStringLength", Output: `if (((''.length > 0) && {}) || (0 && void 0)) { }`},
						{MessageId: "conditionFixCompareEmptyString", Output: `if ((('' !== "") && {}) || (0 && void 0)) { }`},
						{MessageId: "conditionFixCastBoolean", Output: `if (((Boolean('')) && {}) || (0 && void 0)) { }`},
					},
				},
				{MessageId: "conditionErrorObject", Line: 1, Column: 12},
				{
					MessageId: "conditionErrorNumber", Line: 1, Column: 20,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "conditionFixCompareZero", Output: `if (('' && {}) || ((0 !== 0) && void 0)) { }`},
						{MessageId: "conditionFixCompareNaN", Output: `if (('' && {}) || ((!Number.isNaN(0)) && void 0)) { }`},
						{MessageId: "conditionFixCastBoolean", Output: `if (('' && {}) || ((Boolean(0)) && void 0)) { }`},
					},
				},
				{MessageId: "conditionErrorNullish", Line: 1, Column: 25},
			},
		},

		// ---- branded boolean in logical chain: true brand (upstream invalid #6) ----
		{
			Code:    "\ndeclare const foo: true & { __BRAND: 'Foo' };\nif (('' && foo) || (0 && void 0)) { }\n      ",
			Options: map[string]interface{}{"allowNullableObject": false, "allowNumber": false, "allowString": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorString", Line: 3, Column: 6,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "conditionFixCompareStringLength", Output: "\ndeclare const foo: true & { __BRAND: 'Foo' };\nif (((''.length > 0) && foo) || (0 && void 0)) { }\n      "},
						{MessageId: "conditionFixCompareEmptyString", Output: "\ndeclare const foo: true & { __BRAND: 'Foo' };\nif ((('' !== \"\") && foo) || (0 && void 0)) { }\n      "},
						{MessageId: "conditionFixCastBoolean", Output: "\ndeclare const foo: true & { __BRAND: 'Foo' };\nif (((Boolean('')) && foo) || (0 && void 0)) { }\n      "},
					},
				},
				{
					MessageId: "conditionErrorNumber", Line: 3, Column: 21,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "conditionFixCompareZero", Output: "\ndeclare const foo: true & { __BRAND: 'Foo' };\nif (('' && foo) || ((0 !== 0) && void 0)) { }\n      "},
						{MessageId: "conditionFixCompareNaN", Output: "\ndeclare const foo: true & { __BRAND: 'Foo' };\nif (('' && foo) || ((!Number.isNaN(0)) && void 0)) { }\n      "},
						{MessageId: "conditionFixCastBoolean", Output: "\ndeclare const foo: true & { __BRAND: 'Foo' };\nif (('' && foo) || ((Boolean(0)) && void 0)) { }\n      "},
					},
				},
				{MessageId: "conditionErrorNullish", Line: 3, Column: 26},
			},
		},

		// ---- branded boolean in logical chain: false brand (upstream invalid #7) ----
		{
			Code:    "\ndeclare const foo: false & { __BRAND: 'Foo' };\nif (('' && {}) || (foo && void 0)) { }\n      ",
			Options: map[string]interface{}{"allowNullableObject": false, "allowNumber": false, "allowString": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorString", Line: 3, Column: 6,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "conditionFixCompareStringLength", Output: "\ndeclare const foo: false & { __BRAND: 'Foo' };\nif (((''.length > 0) && {}) || (foo && void 0)) { }\n      "},
						{MessageId: "conditionFixCompareEmptyString", Output: "\ndeclare const foo: false & { __BRAND: 'Foo' };\nif ((('' !== \"\") && {}) || (foo && void 0)) { }\n      "},
						{MessageId: "conditionFixCastBoolean", Output: "\ndeclare const foo: false & { __BRAND: 'Foo' };\nif (((Boolean('')) && {}) || (foo && void 0)) { }\n      "},
					},
				},
				{MessageId: "conditionErrorObject", Line: 3, Column: 12},
				{MessageId: "conditionErrorNullish", Line: 3, Column: 27},
			},
		},

		// ---- nested logical chain in control flow position (upstream invalid #10) ----
		{
			Code:    `let x = (1 && 'a' && null) || 0 || '' || {};`,
			Options: map[string]interface{}{"allowNumber": false, "allowString": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNumber", Line: 1, Column: 10,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "conditionFixCompareZero", Output: `let x = ((1 !== 0) && 'a' && null) || 0 || '' || {};`},
						{MessageId: "conditionFixCompareNaN", Output: `let x = ((!Number.isNaN(1)) && 'a' && null) || 0 || '' || {};`},
						{MessageId: "conditionFixCastBoolean", Output: `let x = ((Boolean(1)) && 'a' && null) || 0 || '' || {};`},
					},
				},
				{
					MessageId: "conditionErrorString", Line: 1, Column: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "conditionFixCompareStringLength", Output: `let x = (1 && ('a'.length > 0) && null) || 0 || '' || {};`},
						{MessageId: "conditionFixCompareEmptyString", Output: `let x = (1 && ('a' !== "") && null) || 0 || '' || {};`},
						{MessageId: "conditionFixCastBoolean", Output: `let x = (1 && (Boolean('a')) && null) || 0 || '' || {};`},
					},
				},
				{MessageId: "conditionErrorNullish", Line: 1, Column: 22},
				{
					MessageId: "conditionErrorNumber", Line: 1, Column: 31,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "conditionFixCompareZero", Output: `let x = (1 && 'a' && null) || (0 !== 0) || '' || {};`},
						{MessageId: "conditionFixCompareNaN", Output: `let x = (1 && 'a' && null) || (!Number.isNaN(0)) || '' || {};`},
						{MessageId: "conditionFixCastBoolean", Output: `let x = (1 && 'a' && null) || (Boolean(0)) || '' || {};`},
					},
				},
				{
					MessageId: "conditionErrorString", Line: 1, Column: 36,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "conditionFixCompareStringLength", Output: `let x = (1 && 'a' && null) || 0 || (''.length > 0) || {};`},
						{MessageId: "conditionFixCompareEmptyString", Output: `let x = (1 && 'a' && null) || 0 || ('' !== "") || {};`},
						{MessageId: "conditionFixCastBoolean", Output: `let x = (1 && 'a' && null) || 0 || (Boolean('')) || {};`},
					},
				},
			},
		},

		// ---- nested logical chain with return (upstream invalid #11) ----
		{
			Code:    `return (1 || 'a' || null) && 0 && '' && {};`,
			Options: map[string]interface{}{"allowNumber": false, "allowString": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNumber", Line: 1, Column: 9,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "conditionFixCompareZero", Output: `return ((1 !== 0) || 'a' || null) && 0 && '' && {};`},
						{MessageId: "conditionFixCompareNaN", Output: `return ((!Number.isNaN(1)) || 'a' || null) && 0 && '' && {};`},
						{MessageId: "conditionFixCastBoolean", Output: `return ((Boolean(1)) || 'a' || null) && 0 && '' && {};`},
					},
				},
				{
					MessageId: "conditionErrorString", Line: 1, Column: 14,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "conditionFixCompareStringLength", Output: `return (1 || ('a'.length > 0) || null) && 0 && '' && {};`},
						{MessageId: "conditionFixCompareEmptyString", Output: `return (1 || ('a' !== "") || null) && 0 && '' && {};`},
						{MessageId: "conditionFixCastBoolean", Output: `return (1 || (Boolean('a')) || null) && 0 && '' && {};`},
					},
				},
				{MessageId: "conditionErrorNullish", Line: 1, Column: 21},
				{
					MessageId: "conditionErrorNumber", Line: 1, Column: 30,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "conditionFixCompareZero", Output: `return (1 || 'a' || null) && (0 !== 0) && '' && {};`},
						{MessageId: "conditionFixCompareNaN", Output: `return (1 || 'a' || null) && (!Number.isNaN(0)) && '' && {};`},
						{MessageId: "conditionFixCastBoolean", Output: `return (1 || 'a' || null) && (Boolean(0)) && '' && {};`},
					},
				},
				{
					MessageId: "conditionErrorString", Line: 1, Column: 35,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "conditionFixCompareStringLength", Output: `return (1 || 'a' || null) && 0 && (''.length > 0) && {};`},
						{MessageId: "conditionFixCompareEmptyString", Output: `return (1 || 'a' || null) && 0 && ('' !== "") && {};`},
						{MessageId: "conditionFixCastBoolean", Output: `return (1 || 'a' || null) && 0 && (Boolean('')) && {};`},
					},
				},
			},
		},

		// ---- nested logical inside console.log argument (upstream invalid #12) ----
		{
			Code:    `console.log((1 && []) || ('a' && {}));`,
			Options: map[string]interface{}{"allowNumber": false, "allowString": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNumber", Line: 1, Column: 14,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "conditionFixCompareZero", Output: `console.log(((1 !== 0) && []) || ('a' && {}));`},
						{MessageId: "conditionFixCompareNaN", Output: `console.log(((!Number.isNaN(1)) && []) || ('a' && {}));`},
						{MessageId: "conditionFixCastBoolean", Output: `console.log(((Boolean(1)) && []) || ('a' && {}));`},
					},
				},
				{MessageId: "conditionErrorObject", Line: 1, Column: 19},
				{
					MessageId: "conditionErrorString", Line: 1, Column: 27,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "conditionFixCompareStringLength", Output: `console.log((1 && []) || (('a'.length > 0) && {}));`},
						{MessageId: "conditionFixCompareEmptyString", Output: `console.log((1 && []) || (('a' !== "") && {}));`},
						{MessageId: "conditionFixCastBoolean", Output: `console.log((1 && []) || ((Boolean('a')) && {}));`},
					},
				},
			},
		},

		// ---- all-operands-in-condition (upstream invalid #13) ----
		{
			Code:    `if ((1 && []) || ('a' && {})) void 0;`,
			Options: map[string]interface{}{"allowNumber": false, "allowString": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNumber", Line: 1, Column: 6,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "conditionFixCompareZero", Output: `if (((1 !== 0) && []) || ('a' && {})) void 0;`},
						{MessageId: "conditionFixCompareNaN", Output: `if (((!Number.isNaN(1)) && []) || ('a' && {})) void 0;`},
						{MessageId: "conditionFixCastBoolean", Output: `if (((Boolean(1)) && []) || ('a' && {})) void 0;`},
					},
				},
				{MessageId: "conditionErrorObject", Line: 1, Column: 11},
				{
					MessageId: "conditionErrorString", Line: 1, Column: 19,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "conditionFixCompareStringLength", Output: `if ((1 && []) || (('a'.length > 0) && {})) void 0;`},
						{MessageId: "conditionFixCompareEmptyString", Output: `if ((1 && []) || (('a' !== "") && {})) void 0;`},
						{MessageId: "conditionFixCastBoolean", Output: `if ((1 && []) || ((Boolean('a')) && {})) void 0;`},
					},
				},
				{MessageId: "conditionErrorObject", Line: 1, Column: 26},
			},
		},

		// ---- conditional expression test (upstream invalid #14) ----
		{
			Code:    `let x = null || 0 || 'a' || [] ? {} : undefined;`,
			Options: map[string]interface{}{"allowNumber": false, "allowString": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "conditionErrorNullish", Line: 1, Column: 9},
				{
					MessageId: "conditionErrorNumber", Line: 1, Column: 17,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "conditionFixCompareZero", Output: `let x = null || (0 !== 0) || 'a' || [] ? {} : undefined;`},
						{MessageId: "conditionFixCompareNaN", Output: `let x = null || (!Number.isNaN(0)) || 'a' || [] ? {} : undefined;`},
						{MessageId: "conditionFixCastBoolean", Output: `let x = null || (Boolean(0)) || 'a' || [] ? {} : undefined;`},
					},
				},
				{
					MessageId: "conditionErrorString", Line: 1, Column: 22,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "conditionFixCompareStringLength", Output: `let x = null || 0 || ('a'.length > 0) || [] ? {} : undefined;`},
						{MessageId: "conditionFixCompareEmptyString", Output: `let x = null || 0 || ('a' !== "") || [] ? {} : undefined;`},
						{MessageId: "conditionFixCastBoolean", Output: `let x = null || 0 || (Boolean('a')) || [] ? {} : undefined;`},
					},
				},
				{MessageId: "conditionErrorObject", Line: 1, Column: 29},
			},
		},

		// ---- return !(...) (upstream invalid #15) ----
		{
			Code:    `return !(null || 0 || 'a' || []);`,
			Options: map[string]interface{}{"allowNumber": false, "allowString": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "conditionErrorNullish", Line: 1, Column: 10},
				{
					MessageId: "conditionErrorNumber", Line: 1, Column: 18,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "conditionFixCompareZero", Output: `return !(null || (0 !== 0) || 'a' || []);`},
						{MessageId: "conditionFixCompareNaN", Output: `return !(null || (!Number.isNaN(0)) || 'a' || []);`},
						{MessageId: "conditionFixCastBoolean", Output: `return !(null || (Boolean(0)) || 'a' || []);`},
					},
				},
				{
					MessageId: "conditionErrorString", Line: 1, Column: 23,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "conditionFixCompareStringLength", Output: `return !(null || 0 || ('a'.length > 0) || []);`},
						{MessageId: "conditionFixCompareEmptyString", Output: `return !(null || 0 || ('a' !== "") || []);`},
						{MessageId: "conditionFixCastBoolean", Output: `return !(null || 0 || (Boolean('a')) || []);`},
					},
				},
				{MessageId: "conditionErrorObject", Line: 1, Column: 30},
			},
		},

		// ---- nullable enum: no initializers (upstream invalid #65) ----
		{
			Code:    "\n        enum ExampleEnum {\n          This,\n          That,\n        }\n        const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;\n        if (!theEnum) {\n        }\n      ",
			Options: map[string]interface{}{"allowNullableEnum": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableEnum", Line: 7, Column: 14, EndLine: 7, EndColumn: 21,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "\n        enum ExampleEnum {\n          This,\n          That,\n        }\n        const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;\n        if (theEnum == null) {\n        }\n      "},
				},
			}},
		},

		// ---- nullable enum: ''/'a' (upstream invalid #66) ----
		{
			Code:    "\n        enum ExampleEnum {\n          This = '',\n          That = 'a',\n        }\n        const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;\n        if (!theEnum) {\n        }\n      ",
			Options: map[string]interface{}{"allowNullableEnum": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableEnum", Line: 7, Column: 14, EndLine: 7, EndColumn: 21,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "\n        enum ExampleEnum {\n          This = '',\n          That = 'a',\n        }\n        const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;\n        if (theEnum == null) {\n        }\n      "},
				},
			}},
		},

		// ---- nullable mixed-enum: ''/0 (upstream invalid #67) ----
		{
			Code:    "\n        enum ExampleEnum {\n          This = '',\n          That = 0,\n        }\n        const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;\n        if (!theEnum) {\n        }\n      ",
			Options: map[string]interface{}{"allowNullableEnum": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableEnum", Line: 7, Column: 14, EndLine: 7, EndColumn: 21,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "\n        enum ExampleEnum {\n          This = '',\n          That = 0,\n        }\n        const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;\n        if (theEnum == null) {\n        }\n      "},
				},
			}},
		},

		// ---- nullable enum: 1/2 truthy-only (upstream invalid #69) ----
		{
			Code:    "\n        enum ExampleEnum {\n          This = 1,\n          That = 2,\n        }\n        const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;\n        if (!theEnum) {\n        }\n      ",
			Options: map[string]interface{}{"allowNullableEnum": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableEnum", Line: 7, Column: 14, EndLine: 7, EndColumn: 21,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "\n        enum ExampleEnum {\n          This = 1,\n          That = 2,\n        }\n        const theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;\n        if (theEnum == null) {\n        }\n      "},
				},
			}},
		},

		// ---- nullable mixed enum: falsy string + truthy number (upstream invalid #71) ----
		{
			Code:    "\n        enum ExampleEnum {\n          This = '',\n          That = 1,\n        }\n        (value?: ExampleEnum) => (!value ? 1 : 0);\n      ",
			Options: map[string]interface{}{"allowNullableEnum": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableEnum", Line: 6, Column: 36, EndLine: 6, EndColumn: 41,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "\n        enum ExampleEnum {\n          This = '',\n          That = 1,\n        }\n        (value?: ExampleEnum) => ((value == null) ? 1 : 0);\n      "},
				},
			}},
		},

		// ---- nullable mixed enum: truthy string + truthy number (upstream invalid #72) ----
		{
			Code:    "\n        enum ExampleEnum {\n          This = 'this',\n          That = 1,\n        }\n        (value?: ExampleEnum) => (!value ? 1 : 0);\n      ",
			Options: map[string]interface{}{"allowNullableEnum": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableEnum", Line: 6, Column: 36, EndLine: 6, EndColumn: 41,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "\n        enum ExampleEnum {\n          This = 'this',\n          That = 1,\n        }\n        (value?: ExampleEnum) => ((value == null) ? 1 : 0);\n      "},
				},
			}},
		},

		// ---- nullable mixed enum: falsy string + falsy number (upstream invalid #73) ----
		{
			Code:    "\n        enum ExampleEnum {\n          This = '',\n          That = 0,\n        }\n        (value?: ExampleEnum) => (!value ? 1 : 0);\n      ",
			Options: map[string]interface{}{"allowNullableEnum": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableEnum", Line: 6, Column: 36, EndLine: 6, EndColumn: 41,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "\n        enum ExampleEnum {\n          This = '',\n          That = 0,\n        }\n        (value?: ExampleEnum) => ((value == null) ? 1 : 0);\n      "},
				},
			}},
		},

		// ---- ASI: semicolon-insertion safety (upstream invalid #79) ----
		{
			Code:    "\n        declare const obj: { x: number } | null;\n        !obj ? 1 : 0\n        !obj\n        obj || 0\n        obj && 1 || 0\n      ",
			Options: map[string]interface{}{"allowNullableObject": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "conditionErrorNullableObject", Line: 3, Column: 10,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "conditionFixCompareNullish", Output: "\n        declare const obj: { x: number } | null;\n        (obj == null) ? 1 : 0\n        !obj\n        obj || 0\n        obj && 1 || 0\n      "},
					},
				},
				{
					MessageId: "conditionErrorNullableObject", Line: 4, Column: 10,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "conditionFixCompareNullish", Output: "\n        declare const obj: { x: number } | null;\n        !obj ? 1 : 0\n        obj == null\n        obj || 0\n        obj && 1 || 0\n      "},
					},
				},
				{
					MessageId: "conditionErrorNullableObject", Line: 5, Column: 9,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "conditionFixCompareNullish", Output: "\n        declare const obj: { x: number } | null;\n        !obj ? 1 : 0\n        !obj\n        ;(obj != null) || 0\n        obj && 1 || 0\n      "},
					},
				},
				{
					MessageId: "conditionErrorNullableObject", Line: 6, Column: 9,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "conditionFixCompareNullish", Output: "\n        declare const obj: { x: number } | null;\n        !obj ? 1 : 0\n        !obj\n        obj || 0\n        ;(obj != null) && 1 || 0\n      "},
					},
				},
			},
		},

		// ---- assertion function: 2 overloads (upstream invalid #82) ----
		{
			Code: "\ndeclare function assert(a: number, b: unknown): asserts b;\ndeclare function assert(one: number, two: unknown): asserts two;\ndeclare const nullableString: string | null;\nassert(foo, nullableString);\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableString", Line: 5, Column: 13,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "\ndeclare function assert(a: number, b: unknown): asserts b;\ndeclare function assert(one: number, two: unknown): asserts two;\ndeclare const nullableString: string | null;\nassert(foo, nullableString != null);\n      "},
					{MessageId: "conditionFixDefaultEmptyString", Output: "\ndeclare function assert(a: number, b: unknown): asserts b;\ndeclare function assert(one: number, two: unknown): asserts two;\ndeclare const nullableString: string | null;\nassert(foo, nullableString ?? \"\");\n      "},
					{MessageId: "conditionFixCastBoolean", Output: "\ndeclare function assert(a: number, b: unknown): asserts b;\ndeclare function assert(one: number, two: unknown): asserts two;\ndeclare const nullableString: string | null;\nassert(foo, Boolean(nullableString));\n      "},
				},
			}},
		},

		// ---- assertion function: this:object overload (upstream invalid #83) ----
		{
			Code: "\ndeclare function assert(this: object, a: number, b: unknown): asserts b;\ndeclare const nullableString: string | null;\nassert(foo, nullableString);\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableString", Line: 4, Column: 13,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "\ndeclare function assert(this: object, a: number, b: unknown): asserts b;\ndeclare const nullableString: string | null;\nassert(foo, nullableString != null);\n      "},
					{MessageId: "conditionFixDefaultEmptyString", Output: "\ndeclare function assert(this: object, a: number, b: unknown): asserts b;\ndeclare const nullableString: string | null;\nassert(foo, nullableString ?? \"\");\n      "},
					{MessageId: "conditionFixCastBoolean", Output: "\ndeclare function assert(this: object, a: number, b: unknown): asserts b;\ndeclare const nullableString: string | null;\nassert(foo, Boolean(nullableString));\n      "},
				},
			}},
		},

		// ---- assertion function: skipped TS#59707 (upstream invalid #84) ----
		// SKIP: tsgo, like the TS API the upstream test references, does not
		// report `someAssert(maybeString)` as a type predicate call when
		// `someAssert` is a union of assertion overloads. Carried as Skip
		// to preserve the upstream order.
		{
			Skip: true,
			Code: "\nfunction asserts1(x: string | number | undefined): asserts x {}\nfunction asserts2(x: string | number | undefined): asserts x {}\n\nconst maybeString = Math.random() ? 'string'.slice() : undefined;\n\nconst someAssert: typeof asserts1 | typeof asserts2 =\n  Math.random() > 0.5 ? asserts1 : asserts2;\n\nsomeAssert(maybeString);\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableString",
			}},
		},

		// ---- implementation overload, no rest match (upstream invalid #85) ----
		{
			Code: "\nfunction assert(this: object, a: number, b: unknown): asserts b;\nfunction assert(a: bigint, b: unknown): asserts b;\nfunction assert(this: object, a: string, two: string): asserts two;\nfunction assert(\n  this: object,\n  a: string,\n  assertee: string,\n  c: bigint,\n  d: object,\n): asserts assertee;\n\nfunction assert(...args: any[]) {\n  throw new Error('lol');\n}\n\ndeclare const nullableString: string | null;\nassert(3 as any, nullableString);\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableString", Line: 18, Column: 18,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "\nfunction assert(this: object, a: number, b: unknown): asserts b;\nfunction assert(a: bigint, b: unknown): asserts b;\nfunction assert(this: object, a: string, two: string): asserts two;\nfunction assert(\n  this: object,\n  a: string,\n  assertee: string,\n  c: bigint,\n  d: object,\n): asserts assertee;\n\nfunction assert(...args: any[]) {\n  throw new Error('lol');\n}\n\ndeclare const nullableString: string | null;\nassert(3 as any, nullableString != null);\n      "},
					{MessageId: "conditionFixDefaultEmptyString", Output: "\nfunction assert(this: object, a: number, b: unknown): asserts b;\nfunction assert(a: bigint, b: unknown): asserts b;\nfunction assert(this: object, a: string, two: string): asserts two;\nfunction assert(\n  this: object,\n  a: string,\n  assertee: string,\n  c: bigint,\n  d: object,\n): asserts assertee;\n\nfunction assert(...args: any[]) {\n  throw new Error('lol');\n}\n\ndeclare const nullableString: string | null;\nassert(3 as any, nullableString ?? \"\");\n      "},
					{MessageId: "conditionFixCastBoolean", Output: "\nfunction assert(this: object, a: number, b: unknown): asserts b;\nfunction assert(a: bigint, b: unknown): asserts b;\nfunction assert(this: object, a: string, two: string): asserts two;\nfunction assert(\n  this: object,\n  a: string,\n  assertee: string,\n  c: bigint,\n  d: object,\n): asserts assertee;\n\nfunction assert(...args: any[]) {\n  throw new Error('lol');\n}\n\ndeclare const nullableString: string | null;\nassert(3 as any, Boolean(nullableString));\n      "},
				},
			}},
		},

		// ---- implementation overload + more trailing args (upstream invalid #86) ----
		{
			Code: "\nfunction assert(this: object, a: number, b: unknown): asserts b;\nfunction assert(a: bigint, b: unknown): asserts b;\nfunction assert(this: object, a: string, two: string): asserts two;\nfunction assert(\n  this: object,\n  a: string,\n  assertee: string,\n  c: bigint,\n  d: object,\n): asserts assertee;\nfunction assert(a: any, two: unknown, ...rest: any[]): asserts two;\n\nfunction assert(...args: any[]) {\n  throw new Error('lol');\n}\n\ndeclare const nullableString: string | null;\nassert(3 as any, nullableString, 'more', 'args', 'afterwards');\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableString", Line: 19, Column: 18,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "\nfunction assert(this: object, a: number, b: unknown): asserts b;\nfunction assert(a: bigint, b: unknown): asserts b;\nfunction assert(this: object, a: string, two: string): asserts two;\nfunction assert(\n  this: object,\n  a: string,\n  assertee: string,\n  c: bigint,\n  d: object,\n): asserts assertee;\nfunction assert(a: any, two: unknown, ...rest: any[]): asserts two;\n\nfunction assert(...args: any[]) {\n  throw new Error('lol');\n}\n\ndeclare const nullableString: string | null;\nassert(3 as any, nullableString != null, 'more', 'args', 'afterwards');\n      "},
					{MessageId: "conditionFixDefaultEmptyString", Output: "\nfunction assert(this: object, a: number, b: unknown): asserts b;\nfunction assert(a: bigint, b: unknown): asserts b;\nfunction assert(this: object, a: string, two: string): asserts two;\nfunction assert(\n  this: object,\n  a: string,\n  assertee: string,\n  c: bigint,\n  d: object,\n): asserts assertee;\nfunction assert(a: any, two: unknown, ...rest: any[]): asserts two;\n\nfunction assert(...args: any[]) {\n  throw new Error('lol');\n}\n\ndeclare const nullableString: string | null;\nassert(3 as any, nullableString ?? \"\", 'more', 'args', 'afterwards');\n      "},
					{MessageId: "conditionFixCastBoolean", Output: "\nfunction assert(this: object, a: number, b: unknown): asserts b;\nfunction assert(a: bigint, b: unknown): asserts b;\nfunction assert(this: object, a: string, two: string): asserts two;\nfunction assert(\n  this: object,\n  a: string,\n  assertee: string,\n  c: bigint,\n  d: object,\n): asserts assertee;\nfunction assert(a: any, two: unknown, ...rest: any[]): asserts two;\n\nfunction assert(...args: any[]) {\n  throw new Error('lol');\n}\n\ndeclare const nullableString: string | null;\nassert(3 as any, Boolean(nullableString), 'more', 'args', 'afterwards');\n      "},
				},
			}},
		},

		// ---- assertion with destructured-param overload (upstream invalid #87) ----
		{
			Code: "\ndeclare function assert(a: boolean, b: unknown): asserts b;\ndeclare function assert({ a }: { a: boolean }, b: unknown): asserts b;\ndeclare const nullableString: string | null;\ndeclare const boo: boolean;\nassert(boo, nullableString);\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableString", Line: 6,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "\ndeclare function assert(a: boolean, b: unknown): asserts b;\ndeclare function assert({ a }: { a: boolean }, b: unknown): asserts b;\ndeclare const nullableString: string | null;\ndeclare const boo: boolean;\nassert(boo, nullableString != null);\n      "},
					{MessageId: "conditionFixDefaultEmptyString", Output: "\ndeclare function assert(a: boolean, b: unknown): asserts b;\ndeclare function assert({ a }: { a: boolean }, b: unknown): asserts b;\ndeclare const nullableString: string | null;\ndeclare const boo: boolean;\nassert(boo, nullableString ?? \"\");\n      "},
					{MessageId: "conditionFixCastBoolean", Output: "\ndeclare function assert(a: boolean, b: unknown): asserts b;\ndeclare function assert({ a }: { a: boolean }, b: unknown): asserts b;\ndeclare const nullableString: string | null;\ndeclare const boo: boolean;\nassert(boo, Boolean(nullableString));\n      "},
				},
			}},
		},

		// ---- assertion picks overload matching TS analysis (upstream invalid #88) ----
		{
			Code: "\nfunction assert(one: unknown): asserts one;\nfunction assert(one: unknown, two: unknown): asserts two;\nfunction assert(...args: unknown[]) {\n  throw new Error('not implemented');\n}\ndeclare const nullableString: string | null;\nassert(nullableString);\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableString", Line: 8,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "\nfunction assert(one: unknown): asserts one;\nfunction assert(one: unknown, two: unknown): asserts two;\nfunction assert(...args: unknown[]) {\n  throw new Error('not implemented');\n}\ndeclare const nullableString: string | null;\nassert(nullableString != null);\n      "},
					{MessageId: "conditionFixDefaultEmptyString", Output: "\nfunction assert(one: unknown): asserts one;\nfunction assert(one: unknown, two: unknown): asserts two;\nfunction assert(...args: unknown[]) {\n  throw new Error('not implemented');\n}\ndeclare const nullableString: string | null;\nassert(nullableString ?? \"\");\n      "},
					{MessageId: "conditionFixCastBoolean", Output: "\nfunction assert(one: unknown): asserts one;\nfunction assert(one: unknown, two: unknown): asserts two;\nfunction assert(...args: unknown[]) {\n  throw new Error('not implemented');\n}\ndeclare const nullableString: string | null;\nassert(Boolean(nullableString));\n      "},
				},
			}},
		},

		// ---- find with allowString:false (upstream invalid #89) ----
		{
			Code:    "\n['one', 'two', ''].find(x => {\n  return x;\n});\n      ",
			Options: map[string]interface{}{"allowString": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorString", Line: 2, Column: 25, EndLine: 4, EndColumn: 2,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "explicitBooleanReturnType", Output: "\n['one', 'two', ''].find((x): boolean => {\n  return x;\n});\n      "},
				},
			}},
		},

		// ---- find returning bare `return;` (upstream invalid #90) ----
		{
			Code: "\n['one', 'two', ''].find(x => {\n  return;\n});\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullish", Line: 2, Column: 25, EndLine: 4, EndColumn: 2,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "explicitBooleanReturnType", Output: "\n['one', 'two', ''].find((x): boolean => {\n  return;\n});\n      "},
				},
			}},
		},

		// ---- findLast returning undefined (upstream invalid #91) ----
		{
			Code: "\n['one', 'two', ''].findLast(x => {\n  return undefined;\n});\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullish", Line: 2, Column: 29, EndLine: 4, EndColumn: 2,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "explicitBooleanReturnType", Output: "\n['one', 'two', ''].findLast((x): boolean => {\n  return undefined;\n});\n      "},
				},
			}},
		},

		// ---- find with conditional return (upstream invalid #92) ----
		{
			Code: "\n['one', 'two', ''].find(x => {\n  if (x) {\n    return Math.random() > 0.5;\n  }\n});\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableBoolean", Line: 2, Column: 25, EndLine: 6, EndColumn: 2,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "explicitBooleanReturnType", Output: "\n['one', 'two', ''].find((x): boolean => {\n  if (x) {\n    return Math.random() > 0.5;\n  }\n});\n      "},
				},
			}},
		},

		// ---- predicate variable nullableBoolean (upstream invalid #93) ----
		{
			Code:    "\nconst predicate = (x: string) => {\n  if (x) {\n    return Math.random() > 0.5;\n  }\n};\n\n['one', 'two', ''].find(predicate);\n      ",
			Options: map[string]interface{}{"allowNullableBoolean": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableBoolean", Line: 8, Column: 25, EndLine: 8, EndColumn: 34,
			}},
		},

		// ---- async predicate variable (upstream invalid #95) ----
		{
			Code: "\nconst predicate = async x => {\n  return x != null;\n};\n\n[1, null].every(predicate);\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorObject", Line: 6, Column: 17, EndLine: 6, EndColumn: 26,
			}},
		},

		// ---- 3-overload function (upstream invalid #103) ----
		{
			Code: "\ndeclare function f(x: number): string;\ndeclare function f(x: number | boolean): boolean;\ndeclare function f(x: string | null): boolean;\n\n[35].filter(f);\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorOther", Line: 6, Column: 13, EndLine: 6, EndColumn: 14,
			}},
		},

		// ---- type-constrained predicate: T extends number (upstream invalid #105) ----
		{
			Code:    "\nfunction foo<T extends number>(x: number): T {}\n[1, null].every(foo);\n      ",
			Options: map[string]interface{}{"allowNumber": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNumber", Line: 3, Column: 17, EndLine: 3, EndColumn: 20,
			}},
		},

		// ---- predicate body: nullable string in body (upstream invalid #106) ----
		{
			Code: "\ndeclare const nullOrString: string | null;\n['one', null].filter(x => nullOrString);\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableString", Line: 3, Column: 22, EndLine: 3, EndColumn: 39,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "\ndeclare const nullOrString: string | null;\n['one', null].filter(x => nullOrString != null);\n      "},
					{MessageId: "conditionFixDefaultEmptyString", Output: "\ndeclare const nullOrString: string | null;\n['one', null].filter(x => nullOrString ?? \"\");\n      "},
					{MessageId: "conditionFixCastBoolean", Output: "\ndeclare const nullOrString: string | null;\n['one', null].filter(x => Boolean(nullOrString));\n      "},
					{MessageId: "explicitBooleanReturnType", Output: "\ndeclare const nullOrString: string | null;\n['one', null].filter((x): boolean => nullOrString);\n      "},
				},
			}},
		},

		// ---- predicate body: !nullable string in body (upstream invalid #107) ----
		{
			Code: "\ndeclare const nullOrString: string | null;\n['one', null].filter(x => !nullOrString);\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableString", Line: 3, Column: 28, EndLine: 3, EndColumn: 40,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "\ndeclare const nullOrString: string | null;\n['one', null].filter(x => nullOrString == null);\n      "},
					{MessageId: "conditionFixDefaultEmptyString", Output: "\ndeclare const nullOrString: string | null;\n['one', null].filter(x => !(nullOrString ?? \"\"));\n      "},
					{MessageId: "conditionFixCastBoolean", Output: "\ndeclare const nullOrString: string | null;\n['one', null].filter(x => !Boolean(nullOrString));\n      "},
				},
			}},
		},

		// ---- predicate body: any value (upstream invalid #108) ----
		{
			Code: "\ndeclare const anyValue: any;\n['one', null].filter(x => anyValue);\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorAny", Line: 3, Column: 22, EndLine: 3, EndColumn: 35,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCastBoolean", Output: "\ndeclare const anyValue: any;\n['one', null].filter(x => Boolean(anyValue));\n      "},
					{MessageId: "explicitBooleanReturnType", Output: "\ndeclare const anyValue: any;\n['one', null].filter((x): boolean => anyValue);\n      "},
				},
			}},
		},

		// ---- predicate body: nullable boolean (upstream invalid #109) ----
		{
			Code: "\ndeclare const nullOrBoolean: boolean | null;\n[true, null].filter(x => nullOrBoolean);\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableBoolean", Line: 3, Column: 21, EndLine: 3, EndColumn: 39,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixDefaultFalse", Output: "\ndeclare const nullOrBoolean: boolean | null;\n[true, null].filter(x => nullOrBoolean ?? false);\n      "},
					{MessageId: "conditionFixCompareTrue", Output: "\ndeclare const nullOrBoolean: boolean | null;\n[true, null].filter(x => nullOrBoolean === true);\n      "},
					{MessageId: "explicitBooleanReturnType", Output: "\ndeclare const nullOrBoolean: boolean | null;\n[true, null].filter((x): boolean => nullOrBoolean);\n      "},
				},
			}},
		},

		// ---- predicate body: nullable enum (upstream invalid #110) ----
		{
			Code: "\nenum ExampleEnum {\n  This = 0,\n  That = 1,\n}\nconst theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;\n[0, 1].filter(x => theEnum);\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableEnum", Line: 7, Column: 15, EndLine: 7, EndColumn: 27,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "\nenum ExampleEnum {\n  This = 0,\n  That = 1,\n}\nconst theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;\n[0, 1].filter(x => theEnum != null);\n      "},
					{MessageId: "explicitBooleanReturnType", Output: "\nenum ExampleEnum {\n  This = 0,\n  That = 1,\n}\nconst theEnum = Math.random() < 0.3 ? ExampleEnum.This : null;\n[0, 1].filter((x): boolean => theEnum);\n      "},
				},
			}},
		},

		// ---- predicate body: nullable number (upstream invalid #111) ----
		{
			Code: "\ndeclare const nullOrNumber: number | null;\n[0, null].filter(x => nullOrNumber);\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableNumber", Line: 3, Column: 18, EndLine: 3, EndColumn: 35,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "\ndeclare const nullOrNumber: number | null;\n[0, null].filter(x => nullOrNumber != null);\n      "},
					{MessageId: "conditionFixDefaultZero", Output: "\ndeclare const nullOrNumber: number | null;\n[0, null].filter(x => nullOrNumber ?? 0);\n      "},
					{MessageId: "conditionFixCastBoolean", Output: "\ndeclare const nullOrNumber: number | null;\n[0, null].filter(x => Boolean(nullOrNumber));\n      "},
					{MessageId: "explicitBooleanReturnType", Output: "\ndeclare const nullOrNumber: number | null;\n[0, null].filter((x): boolean => nullOrNumber);\n      "},
				},
			}},
		},

		// ---- predicate body: object value (upstream invalid #112) ----
		{
			Code: "\nconst objectValue: object = {};\n[{ a: 0 }, {}].filter(x => objectValue);\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorObject", Line: 3, Column: 23, EndLine: 3, EndColumn: 39,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "explicitBooleanReturnType", Output: "\nconst objectValue: object = {};\n[{ a: 0 }, {}].filter((x): boolean => objectValue);\n      "},
				},
			}},
		},

		// ---- predicate body: object in block return (upstream invalid #113) ----
		{
			Code: "\nconst objectValue: object = {};\n[{ a: 0 }, {}].filter(x => {\n  return objectValue;\n});\n      ",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorObject", Line: 3, Column: 23, EndLine: 5, EndColumn: 2,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "explicitBooleanReturnType", Output: "\nconst objectValue: object = {};\n[{ a: 0 }, {}].filter((x): boolean => {\n  return objectValue;\n});\n      "},
				},
			}},
		},

		// ---- predicate body: nullable object (upstream invalid #114) ----
		{
			Code:    "\ndeclare const nullOrObject: object | null;\n[{ a: 0 }, null].filter(x => nullOrObject);\n      ",
			Options: map[string]interface{}{"allowNullableObject": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNullableObject", Line: 3, Column: 25, EndLine: 3, EndColumn: 42,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareNullish", Output: "\ndeclare const nullOrObject: object | null;\n[{ a: 0 }, null].filter(x => nullOrObject != null);\n      "},
					{MessageId: "explicitBooleanReturnType", Output: "\ndeclare const nullOrObject: object | null;\n[{ a: 0 }, null].filter((x): boolean => nullOrObject);\n      "},
				},
			}},
		},

		// ---- predicate body: array.length (upstream invalid #115) ----
		{
			Code:    "\nconst numbers: number[] = [1];\n[1, 2].filter(x => numbers.length);\n      ",
			Options: map[string]interface{}{"allowNumber": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNumber", Line: 3, Column: 15, EndLine: 3, EndColumn: 34,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareArrayLengthNonzero", Output: "\nconst numbers: number[] = [1];\n[1, 2].filter(x => numbers.length > 0);\n      "},
					{MessageId: "explicitBooleanReturnType", Output: "\nconst numbers: number[] = [1];\n[1, 2].filter((x): boolean => numbers.length);\n      "},
				},
			}},
		},

		// ---- predicate body: number value (upstream invalid #116) ----
		{
			Code:    "\nconst numberValue: number = 1;\n[1, 2].filter(x => numberValue);\n      ",
			Options: map[string]interface{}{"allowNumber": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorNumber", Line: 3, Column: 15, EndLine: 3, EndColumn: 31,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareZero", Output: "\nconst numberValue: number = 1;\n[1, 2].filter(x => numberValue !== 0);\n      "},
					{MessageId: "conditionFixCompareNaN", Output: "\nconst numberValue: number = 1;\n[1, 2].filter(x => !Number.isNaN(numberValue));\n      "},
					{MessageId: "conditionFixCastBoolean", Output: "\nconst numberValue: number = 1;\n[1, 2].filter(x => Boolean(numberValue));\n      "},
					{MessageId: "explicitBooleanReturnType", Output: "\nconst numberValue: number = 1;\n[1, 2].filter((x): boolean => numberValue);\n      "},
				},
			}},
		},

		// ---- predicate body: string value (upstream invalid #117) ----
		{
			Code:    "\nconst stringValue: string = 'hoge';\n['hoge', 'foo'].filter(x => stringValue);\n      ",
			Options: map[string]interface{}{"allowString": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionErrorString", Line: 3, Column: 24, EndLine: 3, EndColumn: 40,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "conditionFixCompareStringLength", Output: "\nconst stringValue: string = 'hoge';\n['hoge', 'foo'].filter(x => stringValue.length > 0);\n      "},
					{MessageId: "conditionFixCompareEmptyString", Output: "\nconst stringValue: string = 'hoge';\n['hoge', 'foo'].filter(x => stringValue !== \"\");\n      "},
					{MessageId: "conditionFixCastBoolean", Output: "\nconst stringValue: string = 'hoge';\n['hoge', 'foo'].filter(x => Boolean(stringValue));\n      "},
					{MessageId: "explicitBooleanReturnType", Output: "\nconst stringValue: string = 'hoge';\n['hoge', 'foo'].filter((x): boolean => stringValue);\n      "},
				},
			}},
		},
	})
}
