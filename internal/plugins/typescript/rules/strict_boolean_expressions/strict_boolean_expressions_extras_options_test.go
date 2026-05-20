// TestStrictBooleanExpressionsExtrasOptions exercises the options × type
// product space. The motivation is that determineReportType has 8+ option
// gates and 16+ variant arms; only a cross matrix proves no option flip
// silently degrades a previously-flagged arm into a pass (or vice versa).
//
// Coverage tactic: for each `allow*` option, lock in (1) the rule firing
// with the option OFF and (2) the rule staying silent with the option ON,
// against the canonical type for that option AND every neighboring type
// that should NOT be affected by that option flip. Also lock in the default
// option set so a change to defaults breaks loudly.
package strict_boolean_expressions

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestStrictBooleanExpressionsExtrasOptions(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &StrictBooleanExpressionsRule, []rule_tester.ValidTestCase{
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
	}, []rule_tester.InvalidTestCase{
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
	})
}
