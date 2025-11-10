package no_unnecessary_condition

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnnecessaryConditionRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{
		// Valid cases - conditions that can be either truthy or falsy
		{Code: `
declare const b1: boolean;
declare const b2: boolean;
if (b1 && b2) {}
if (b1 || b2) {}
		`},
		{Code: `
declare const b1: boolean | undefined;
declare const b2: boolean;
if (b1 && b2) {}
if (b1 || b2) {}
		`},
		{Code: `
declare const str: string;
if (str) {}
		`},
		{Code: `
declare const num: number;
if (num) {}
		`},
		{Code: `
declare const arr: any[];
if (arr.length) {}
		`},
		{Code: `
declare const strOrNull: string | null;
if (strOrNull) {}
		`},
		{Code: `
declare const strOrUndef: string | undefined;
if (strOrUndef) {}
		`},
		{Code: `
declare const numOrNull: number | null;
if (numOrNull) {}
		`},
		{Code: `
declare const obj: { prop?: string };
if (obj.prop) {}
		`},
		// Nullish coalescing with nullable types
		{Code: `
declare const strOrNull: string | null;
const result = strOrNull ?? 'default';
		`},
		{Code: `
declare const numOrUndef: number | undefined;
const result = numOrUndef ?? 0;
		`},
		// Optional chaining with nullable types
		{Code: `
declare const obj: { prop?: { nested: string } };
const result = obj.prop?.nested;
		`},
		{Code: `
declare const obj: { method?: () => void } | undefined;
obj?.method?.();
		`},
		// Generic types
		{Code: `
function test<T>(value: T) {
  if (value) {}
}
		`},
		{Code: `
function test<T extends string | number>(value: T) {
  if (value) {}
}
		`},
		// any and unknown
		{Code: `
declare const anyValue: any;
if (anyValue) {}
		`},
		{Code: `
declare const unknownValue: unknown;
if (unknownValue) {}
		`},
		// Comparison operators
		{Code: `
declare const a: string;
declare const b: string;
if (a === b) {}
		`},
		{Code: `
declare const num: number;
if (num === 0) {}
		`},
		// Loop conditions with allowConstantLoopConditions
		{Code: `while (true) {}`, Options: map[string]any{"allowConstantLoopConditions": true}},
		{Code: `do {} while (true);`, Options: map[string]any{"allowConstantLoopConditions": true}},
		{Code: `for (; true; ) {}`, Options: map[string]any{"allowConstantLoopConditions": true}},
		{Code: `while (1) {}`, Options: map[string]any{"allowConstantLoopConditions": "only-allowed-literals"}},
		{Code: `while (0) {}`, Options: map[string]any{"allowConstantLoopConditions": "only-allowed-literals"}},
		// Array predicate methods with proper return types
		{Code: `
declare const arr: number[];
arr.filter(x => x > 0);
		`},
		{Code: `
declare const arr: boolean[];
arr.some(x => x);
		`},
		// Branded types
		{Code: `
type Brand = string & { __brand: 'Brand' };
declare const branded: Brand;
if (branded) {}
		`},
		// Union types with falsy values
		{Code: `
declare const val: string | false;
if (val) {}
		`},
		{Code: `
declare const val: number | 0;
if (val) {}
		`},
	}, []rule_tester.InvalidTestCase{
		// Always truthy object
		{
			Code: `
declare const obj: { prop: string };
if (obj) {}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "alwaysTruthy", Line: 3},
			},
		},
		// Always truthy non-nullable string
		{
			Code: `
declare const str: string;
const result = str ?? 'default';
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "neverNullish", Line: 3},
			},
		},
		// Always truthy number literal
		{
			Code: `
if (42) {}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "alwaysTruthy", Line: 2},
			},
		},
		// Always falsy null
		{
			Code: `
if (null) {}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "alwaysFalsy", Line: 2},
			},
		},
		// Always falsy undefined
		{
			Code: `
if (undefined) {}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "alwaysFalsy", Line: 2},
			},
		},
		// Always falsy void
		{
			Code: `
declare function voidFunc(): void;
if (voidFunc()) {}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "alwaysFalsy", Line: 3},
			},
		},
		// Always falsy false literal
		{
			Code: `
if (false) {}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "alwaysFalsy", Line: 2},
			},
		},
		// Always falsy empty string literal
		{
			Code: `
if ("") {}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "alwaysFalsy", Line: 2},
			},
		},
		// Always falsy 0
		{
			Code: `
if (0) {}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "alwaysFalsy", Line: 2},
			},
		},
		// Always truthy true literal
		{
			Code: `
if (true) {}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "alwaysTruthy", Line: 2},
			},
		},
		// Always truthy non-empty string literal
		{
			Code: `
if ("hello") {}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "alwaysTruthy", Line: 2},
			},
		},
		// Ternary with always truthy condition
		{
			Code: `
const result = true ? "yes" : "no";
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "alwaysTruthy", Line: 2},
			},
		},
		// Logical AND with always truthy left side
		{
			Code: `
declare const b: boolean;
if (true && b) {}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "alwaysTruthy", Line: 3},
			},
		},
		// Logical OR with always truthy left side
		{
			Code: `
declare const b: boolean;
if (true || b) {}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "alwaysTruthy", Line: 3},
			},
		},
		// NOT operator with always truthy
		{
			Code: `
if (!true) {}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "alwaysTruthy", Line: 2},
			},
		},
		// Unnecessary optional chaining on non-nullable
		{
			Code: `
declare const obj: { prop: string };
const result = obj?.prop;
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "neverOptionalChain", Line: 3},
			},
		},
		// Literal comparison
		{
			Code: `
if (true === true) {}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noOverlap", Line: 2},
			},
		},
		// Loop conditions
		{
			Code: `while (true) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "alwaysTruthy", Line: 1},
			},
		},
		{
			Code: `do {} while (false);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "alwaysFalsy", Line: 1},
			},
		},
		{
			Code: `for (; 1; ) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "alwaysTruthy", Line: 1},
			},
		},
		// Array predicate with always truthy
		{
			Code: `
declare const arr: number[];
arr.filter(() => true);
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "alwaysTruthyFunc", Line: 3},
			},
		},
		// Array predicate with always falsy
		{
			Code: `
declare const arr: number[];
arr.filter(() => false);
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "alwaysFalsyFunc", Line: 3},
			},
		},
		// Union of only truthy types
		{
			Code: `
type OnlyTruthy = { a: string } | { b: number };
declare const val: OnlyTruthy;
if (val) {}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "alwaysTruthy", Line: 4},
			},
		},
		// Union of only falsy types
		{
			Code: `
type OnlyFalsy = null | undefined;
declare const val: OnlyFalsy;
if (val) {}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "alwaysFalsy", Line: 4},
			},
		},
		// Constrained generic that's always truthy
		{
			Code: `
function test<T extends object>(value: T) {
  if (value) {}
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "alwaysTruthy", Line: 3},
			},
		},
		// Nullish coalescing on non-nullable
		{
			Code: `
declare const num: number;
const result = num ?? 0;
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "neverNullish", Line: 3},
			},
		},
	})
}
