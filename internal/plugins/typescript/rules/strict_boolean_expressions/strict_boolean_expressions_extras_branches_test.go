// TestStrictBooleanExpressionsExtrasBranches locks in every branch of
// inspectVariantTypes / determineReportType / traverseLogical /
// checkArrayMethodCallPredicate so future refactors can't silently flip the
// classification for any single VariantType bucket combination. Each case
// names the arm it covers via the inline `// Locks in <fn>() arm <N>: ...`
// comment.
package strict_boolean_expressions

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestStrictBooleanExpressionsExtrasBranches(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &StrictBooleanExpressionsRule, []rule_tester.ValidTestCase{
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
	}, []rule_tester.InvalidTestCase{
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
	})
}
