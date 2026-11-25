package no_unnecessary_condition

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestNoUnnecessaryConditionLiteralIndex(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{
		// === noUncheckedIndexedAccess: `?.` after an index access inside an optional chain ===
		{
			Code:     "declare const list: { items: { id: string }[] } | undefined;\nlist?.items[0]?.id;\n",
			TSConfig: "tsconfig.noUncheckedIndexedAccess.json",
		},
		{
			Code:     "declare const items: { id: string }[];\nitems[0]?.id;\n",
			TSConfig: "tsconfig.noUncheckedIndexedAccess.json",
		},
		{
			Code:     "declare const maybe: { id: string }[] | undefined;\nmaybe?.[0]?.id;\n",
			TSConfig: "tsconfig.noUncheckedIndexedAccess.json",
		},
		{
			Code:     "declare const maybe: { id: string }[] | undefined;\nmaybe?.['0']?.id;\n",
			TSConfig: "tsconfig.noUncheckedIndexedAccess.json",
		},
		{
			Code:     "declare const maybe: Record<string, { id: string }> | undefined;\nmaybe?.['key']?.id;\n",
			TSConfig: "tsconfig.noUncheckedIndexedAccess.json",
		},
	}, nil)
}

func TestNoUnnecessaryConditionParentheses(t *testing.T) {
	t.Run("required-property", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, nil, []rule_tester.InvalidTestCase{{Code: `declare const obj: {item: { id: string }} | undefined;
declare const key: 'item';
obj?.[(key)]?.id;
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "neverOptionalChain", Line: 3, Column: 13, EndLine: 3, EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestRemoveOptionalChain", Output: `declare const obj: {item: { id: string }} | undefined;
declare const key: 'item';
obj?.[(key)].id;
`},
					},
				},
			}}})
	})
	t.Run("nested-key-parentheses", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, nil, []rule_tester.InvalidTestCase{{Code: `declare const obj: {item: { id: string }} | undefined;
declare const key: 'item';
obj?.[(((key)))]?.id;
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "neverOptionalChain", Line: 3, Column: 17, EndLine: 3, EndColumn: 19,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestRemoveOptionalChain", Output: `declare const obj: {item: { id: string }} | undefined;
declare const key: 'item';
obj?.[(((key)))].id;
`},
					},
				},
			}}})
	})
	t.Run("optional-property-value", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{{Code: `declare const obj: {item: { id: string } | undefined} | undefined;
declare const key: 'item';
obj?.[(key)]?.id;
`}}, nil)
	})
	t.Run("nullable-property-value", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{{Code: `declare const obj: {item: { id: string } | null} | undefined;
declare const key: 'item';
obj?.[(key)]?.id;
`}}, nil)
	})
	t.Run("any-property-value", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{{Code: `declare const obj: {item: any} | undefined;
declare const key: 'item';
obj?.[(key)]?.id;
`}}, nil)
	})
	t.Run("unknown-property-value", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{{Code: `declare const obj: {item: unknown} | undefined;
declare const key: 'item';
obj?.[(key)]?.id;
`}}, nil)
	})
	t.Run("template-tuple-key", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{{Code: "declare const arr: [{id:string}] | undefined;\ndeclare const key: 0;\narr?.[`0`]?.id;\n"}}, nil)
	})
	t.Run("optional-tuple-key", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{{Code: `declare const arr: [{id:string}?] | undefined;
declare const key: 0;
arr?.[(0)]?.id;
`}}, nil)
	})
	t.Run("numeric-record-key", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, nil, []rule_tester.InvalidTestCase{{Code: `declare const arr: Record<number,{id:string}> | undefined;
declare const key: 0;
arr?.[(key)]?.id;
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "neverOptionalChain", Line: 3, Column: 13, EndLine: 3, EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestRemoveOptionalChain", Output: `declare const arr: Record<number,{id:string}> | undefined;
declare const key: 0;
arr?.[(key)].id;
`},
					},
				},
			}}})
	})
	t.Run("string-record-key", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, nil, []rule_tester.InvalidTestCase{{Code: `declare const arr: Record<string,{id:string}> | undefined;
declare const key: 0;
arr?.[(key)]?.id;
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "neverOptionalChain", Line: 3, Column: 13, EndLine: 3, EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestRemoveOptionalChain", Output: `declare const arr: Record<string,{id:string}> | undefined;
declare const key: 0;
arr?.[(key)].id;
`},
					},
				},
			}}})
	})
	t.Run("tuple-element", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, nil, []rule_tester.InvalidTestCase{{Code: `declare const tuple: [{id:string}];
tuple[(0)]?.id;
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "neverOptionalChain", Line: 2, Column: 11, EndLine: 2, EndColumn: 13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestRemoveOptionalChain", Output: `declare const tuple: [{id:string}];
tuple[(0)].id;
`},
					},
				},
			}}})
	})
	t.Run("grouped-chain", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{{Code: `declare const obj: {item:{id:string}, method:()=>string, get:()=>{id:string}} | undefined;
(obj?.item)?.id;
`}}, nil)
	})
	t.Run("grouped-computed-chain", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{{Code: `declare const obj: {item:{id:string}, method:()=>string, get:()=>{id:string}} | undefined;
(obj?.item)?.["id"];
`}}, nil)
	})
	t.Run("grouped-optional-call", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{{Code: `declare const obj: {item:{id:string}, method:()=>string, get:()=>{id:string}} | undefined;
(obj?.method)?.();
`}}, nil)
	})

	t.Run("grouped-array-access", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{{Code: `declare const arr:{id:string}[] | undefined;
(arr[0])?.id;
`}}, nil)
	})
	t.Run("tuple-condition", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, nil, []rule_tester.InvalidTestCase{{Code: `declare const tuple: [object];
if (tuple[(0)]) {}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "alwaysTruthy", Line: 2, Column: 5, EndLine: 2, EndColumn: 15},
			}}})
	})
	t.Run("tuple-coalescing", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, nil, []rule_tester.InvalidTestCase{{Code: `declare const tuple: [object];
tuple[(0)] ?? {};
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "neverNullish", Line: 2, Column: 1, EndLine: 2, EndColumn: 11},
			}}})
	})
	t.Run("template-tuple-condition", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{{Code: "declare const tuple: [object];\nif (tuple[`0`]) {}\n"}}, nil)
	})
	t.Run("template-tuple-coalescing", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{{Code: "declare const tuple: [object];\ntuple[`0`] ?? {};\n"}}, nil)
	})
	t.Run("grouped-tuple-condition", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, nil, []rule_tester.InvalidTestCase{{Code: `declare const tuple: [object];
if ((tuple[(0)])) {}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "alwaysTruthy", Line: 2, Column: 6, EndLine: 2, EndColumn: 16},
			}}})
	})
	t.Run("grouped-tuple-coalescing", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, nil, []rule_tester.InvalidTestCase{{Code: `declare const tuple: [object];
(tuple[(0)]) ?? {};
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "neverNullish", Line: 2, Column: 2, EndLine: 2, EndColumn: 12},
			}}})
	})
	t.Run("negated-tuple-condition", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, nil, []rule_tester.InvalidTestCase{{Code: `declare const tuple: [object];
if (!tuple[(0)]) {}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "alwaysFalsy", Line: 2, Column: 5, EndLine: 2, EndColumn: 16},
			}}})
	})
	t.Run("negated-grouped-tuple", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, nil, []rule_tester.InvalidTestCase{{Code: `declare const tuple: [object];
if (!(tuple[(0)])) {}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "alwaysFalsy", Line: 2, Column: 5, EndLine: 2, EndColumn: 18},
			}}})
	})
	t.Run("double-negated-tuple", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, nil, []rule_tester.InvalidTestCase{{Code: `declare const tuple: [object];
if (!!tuple[(0)]) {}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "alwaysTruthy", Line: 2, Column: 5, EndLine: 2, EndColumn: 17},
			}}})
	})
	t.Run("double-negated-grouped-tuple", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, nil, []rule_tester.InvalidTestCase{{Code: `declare const tuple: [object];
if (!!(tuple[(0)])) {}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "alwaysTruthy", Line: 2, Column: 5, EndLine: 2, EndColumn: 19},
			}}})
	})
	t.Run("checked-required-property", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.noUncheckedIndexedAccess.json", t, &NoUnnecessaryConditionRule, nil, []rule_tester.InvalidTestCase{{Code: `declare const obj: {item: { id: string }} | undefined;
declare const key: 'item';
obj?.[(key)]?.id;
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "neverOptionalChain", Line: 3, Column: 13, EndLine: 3, EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestRemoveOptionalChain", Output: `declare const obj: {item: { id: string }} | undefined;
declare const key: 'item';
obj?.[(key)].id;
`},
					},
				},
			}}})
	})
	t.Run("checked-nested-key", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.noUncheckedIndexedAccess.json", t, &NoUnnecessaryConditionRule, nil, []rule_tester.InvalidTestCase{{Code: `declare const obj: {item: { id: string }} | undefined;
declare const key: 'item';
obj?.[(((key)))]?.id;
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "neverOptionalChain", Line: 3, Column: 17, EndLine: 3, EndColumn: 19,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestRemoveOptionalChain", Output: `declare const obj: {item: { id: string }} | undefined;
declare const key: 'item';
obj?.[(((key)))].id;
`},
					},
				},
			}}})
	})
	t.Run("checked-optional-value", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.noUncheckedIndexedAccess.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{{Code: `declare const obj: {item: { id: string } | undefined} | undefined;
declare const key: 'item';
obj?.[(key)]?.id;
`}}, nil)
	})
	t.Run("checked-nullable-value", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.noUncheckedIndexedAccess.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{{Code: `declare const obj: {item: { id: string } | null} | undefined;
declare const key: 'item';
obj?.[(key)]?.id;
`}}, nil)
	})
	t.Run("checked-any-value", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.noUncheckedIndexedAccess.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{{Code: `declare const obj: {item: any} | undefined;
declare const key: 'item';
obj?.[(key)]?.id;
`}}, nil)
	})
	t.Run("checked-unknown-value", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.noUncheckedIndexedAccess.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{{Code: `declare const obj: {item: unknown} | undefined;
declare const key: 'item';
obj?.[(key)]?.id;
`}}, nil)
	})
	t.Run("checked-tuple-key", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.noUncheckedIndexedAccess.json", t, &NoUnnecessaryConditionRule, nil, []rule_tester.InvalidTestCase{{Code: `declare const arr: [{id:string}] | undefined;
declare const key: 0;
arr?.[(key)]?.id;
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "neverOptionalChain", Line: 3, Column: 13, EndLine: 3, EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestRemoveOptionalChain", Output: `declare const arr: [{id:string}] | undefined;
declare const key: 0;
arr?.[(key)].id;
`},
					},
				},
			}}})
	})
	t.Run("checked-optional-tuple-key", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.noUncheckedIndexedAccess.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{{Code: `declare const arr: [{id:string}?] | undefined;
declare const key: 0;
arr?.[(key)]?.id;
`}}, nil)
	})
	t.Run("checked-readonly-tuple", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.noUncheckedIndexedAccess.json", t, &NoUnnecessaryConditionRule, nil, []rule_tester.InvalidTestCase{{Code: `declare const arr: readonly [{id:string}] | undefined;
declare const key: 0;
arr?.[(key)]?.id;
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "neverOptionalChain", Line: 3, Column: 13, EndLine: 3, EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestRemoveOptionalChain", Output: `declare const arr: readonly [{id:string}] | undefined;
declare const key: 0;
arr?.[(key)].id;
`},
					},
				},
			}}})
	})
	t.Run("checked-numeric-record", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.noUncheckedIndexedAccess.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{{Code: `declare const arr: Record<number,{id:string}> | undefined;
declare const key: 0;
arr?.[(key)]?.id;
`}}, nil)
	})
	t.Run("checked-string-record", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.noUncheckedIndexedAccess.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{{Code: `declare const arr: Record<string,{id:string}> | undefined;
declare const key: 0;
arr?.[(key)]?.id;
`}}, nil)
	})

}

func TestNoUnnecessaryConditionLiteralComparisons(t *testing.T) {
	operators := []string{"==", "!=", "===", "!==", "<", "<=", ">", ">="}
	var invalid []rule_tester.InvalidTestCase
	for _, test := range []struct {
		left, right   string
		trueOperators []string
	}{
		{"true", "1", []string{"==", "!==", "<=", ">="}},
		{"false", "0", []string{"==", "!==", "<=", ">="}},
		{"\"1\"", "1", []string{"==", "!==", "<=", ">="}},
		{"\"2\"", "10", []string{"!=", "!==", "<", "<="}},
		{"null", "0", []string{"!=", "!==", "<=", ">="}},
		{"undefined", "0", []string{"!=", "!=="}},
		{"1n", "1", []string{"==", "!==", "<=", ">="}},
		{"9007199254740993n", "9007199254740992n", []string{"!=", "!==", ">", ">="}},
		{"\"\U00010000\"", "\"\uE000\"", []string{"!=", "!==", "<", "<="}},
	} {
		for _, operator := range operators {
			expected := slices.Contains(test.trueOperators, operator)
			code := fmt.Sprintf("declare const left: %s;\ndeclare const right: %s;\nleft %s right;\n", test.left, test.right, operator)
			message := fmt.Sprintf("Unnecessary conditional, comparison is always %t, since `%s %s %s` is %t.", expected, test.left, operator, test.right, expected)
			invalid = append(invalid, rule_tester.InvalidTestCase{Code: code, Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "comparisonBetweenLiteralTypes", Message: message, Line: 3, Column: 1,
			}}})
		}
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, nil, invalid)
}

func TestNoUnnecessaryConditionUncheckedKeyVariables(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.noUncheckedIndexedAccess.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{
		{Code: "declare const arr: { id: string }[] | undefined;\ndeclare const key: 0;\narr?.[key]?.id;\narr?.[(key)]?.id;\n"},
		{Code: "declare const arr: { id: string }[] | undefined;\ndeclare const key: '0';\narr?.[key]?.id;\narr?.[(key)]?.id;\n"},
		{Code: "declare const arr: { id: string }[] | undefined;\ndeclare const key: 0 | 1;\narr?.[key]?.id;\narr?.[(key)]?.id;\n"},
		{Code: "declare const record: Record<number, { id: string }> | undefined;\ndeclare const key: '0';\nrecord?.[(key)]?.id;\n"},
	}, nil)
}

func TestNoUnnecessaryConditionParenthesizedLoops(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{
		{Code: "while ((true)) {}\n", Options: map[string]any{"allowConstantLoopConditions": "only-allowed-literals"}},
		{Code: "while (((false))) {}\n", Options: map[string]any{"allowConstantLoopConditions": "only-allowed-literals"}},
		{Code: "while ((0)) {}\n", Options: map[string]any{"allowConstantLoopConditions": "only-allowed-literals"}},
		{Code: "while ((1)) {}\n", Options: map[string]any{"allowConstantLoopConditions": "only-allowed-literals"}},
	}, []rule_tester.InvalidTestCase{{
		Code:    "while ((-1)) {}\n",
		Options: map[string]any{"allowConstantLoopConditions": "only-allowed-literals"},
		Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 1, Column: 9}},
	}})
}

func TestNoUnnecessaryConditionNumericStringIndex(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{
		{Code: "declare const obj: Record<number, { id: string } | undefined> | undefined;\ndeclare const key: '0';\nobj?.[(key)]?.id;\n"},
		{Code: "declare const obj: Record<number, { id: string } | undefined> | undefined;\ndeclare const key: 'NaN';\nobj?.[(key)]?.id;\n"},
	}, []rule_tester.InvalidTestCase{{
		Code: "declare const obj: Record<number, { id: string }> | undefined;\ndeclare const key: '0';\nobj?.[(key)]?.id;\n",
		Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "neverOptionalChain", Line: 3, Column: 13,
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
				MessageId: "suggestRemoveOptionalChain",
				Output:    "declare const obj: Record<number, { id: string }> | undefined;\ndeclare const key: '0';\nobj?.[(key)].id;\n",
			}},
		}},
	}})
}

func TestNoUnnecessaryConditionIndexSignatureSafety(t *testing.T) {
	var valid []rule_tester.ValidTestCase
	var invalid []rule_tester.InvalidTestCase
	for _, config := range []string{"tsconfig.json", "tsconfig.noUncheckedIndexedAccess.json", "tsconfig.exactOptionalPropertyTypes.json"} {
		for _, keyExpression := range []string{"key", "((key))"} {
			for _, test := range []struct{ key, object string }{
				{"0", "Record<string, Item | undefined>"},
				{"number", "Record<string, Item | null>"},
				{"`${number}`", "Record<number, Item | undefined>"},
				{"unique symbol", "{ [key]?: Item }"},
				{"unique symbol", "{ [key]: Item | undefined }"},
				{"symbol", "Record<symbol, Item | undefined>"},
				{"`data-${number}`", "{ [key: `data-${string}`]: Item | undefined }"},
			} {
				code := fmt.Sprintf("type Item = { id: number };\ndeclare const key: %s;\ndeclare const rows: (%s) | undefined;\nrows?.[%s]?.id;\n", test.key, test.object, keyExpression)
				valid = append(valid, rule_tester.ValidTestCase{Code: code, TSConfig: config})
			}
			for _, test := range []struct{ key, object string }{
				{"number", "Record<string, Item>"},
				{"`${number}`", "Record<number, Item>"},
				{"`data-${number}`", "{ [key: `data-${string}`]: Item }"},
			} {
				prefix := fmt.Sprintf("type Item = { id: number };\ndeclare const key: %s;\ndeclare const rows: (%s) | undefined;\n", test.key, test.object)
				code := prefix + fmt.Sprintf("rows?.[%s]?.id;\n", keyExpression)
				if config == "tsconfig.noUncheckedIndexedAccess.json" || test.key == "`data-${number}`" {
					valid = append(valid, rule_tester.ValidTestCase{Code: code, TSConfig: config})
				} else {
					invalid = append(invalid, rule_tester.InvalidTestCase{Code: code, TSConfig: config, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "suggestRemoveOptionalChain", Output: prefix + fmt.Sprintf("rows?.[%s].id;\n", keyExpression),
					}}}}})
				}
			}
			prefix := "type Item = { id: number };\ndeclare const key: unique symbol;\ndeclare const rows: { [key]: Item } | undefined;\n"
			invalid = append(invalid, rule_tester.InvalidTestCase{Code: prefix + fmt.Sprintf("rows?.[%s]?.id;\n", keyExpression), TSConfig: config, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
				MessageId: "suggestRemoveOptionalChain", Output: prefix + fmt.Sprintf("rows?.[%s].id;\n", keyExpression),
			}}}}})
			prefix = "type Item = { id: number };\ndeclare const key: `data-${number}`;\ndeclare const rows: { [p: `data-${string}`]: Item; [p: `${string}-${number}`]: Item | undefined } | undefined;\n"
			code := prefix + fmt.Sprintf("rows?.[%s]?.id;\n", keyExpression)
			if config == "tsconfig.noUncheckedIndexedAccess.json" {
				valid = append(valid, rule_tester.ValidTestCase{Code: code, TSConfig: config})
			} else {
				invalid = append(invalid, rule_tester.InvalidTestCase{Code: code, TSConfig: config, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "suggestRemoveOptionalChain", Output: prefix + fmt.Sprintf("rows?.[%s].id;\n", keyExpression),
				}}}}})
			}
		}
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, valid, invalid)
}

func TestNoUnnecessaryConditionTruthinessContexts(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, nil, []rule_tester.InvalidTestCase{
		{Code: "declare const flags: boolean[];\nflags.filter(flag => flag || true);\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}}},
		{Code: "declare const flags: boolean[];\nflags.filter(flag => flag && true);\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}}},
		{Code: "declare function assert(x: unknown): asserts x;\ndeclare const flag: boolean;\nassert(flag && true);\n", Options: map[string]any{"checkTypePredicates": true}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}}},
		{Code: "declare const name: string | undefined;\nconst label = name || 'Untitled' || 'Fallback';\nexport {};\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}}},
		{Code: "declare const flags: boolean[];\ndeclare const other: boolean;\nflags.filter(flag => (flag && true) || other);\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 3, Column: 31}}},
		{Code: "declare const flags: boolean[];\ndeclare const other: boolean;\nflags.some(function(flag) { return (flag && true) || other; });\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}}},
		{Code: "declare const flags: boolean[];\nflags.every((flag => flag || true));\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}}},
		{Code: "declare function assert(x: unknown): asserts x;\ndeclare const a: boolean;\ndeclare const b: boolean;\nassert((a && true) || b);\n", Options: map[string]any{"checkTypePredicates": true}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}}},
		{Code: "declare const a: boolean;\ndeclare const b: boolean;\nconst result = a && 0 && b;\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy"}}},
	})
}

func TestNoUnnecessaryConditionLogicalValues(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{
		{Code: "declare const a: boolean;\ndeclare const b: boolean;\nconst result = (a && 'a') || (b && 'b');\n"},
		{Code: "declare const a: boolean;\ndeclare const b: boolean;\nconst result = (a || (b && true)) || b;\n"},
		{Code: "declare const flags: boolean[];\ndeclare const other: boolean;\nflags.map(flag => (flag && true) || other);\n"},
		{Code: "declare const object: { filter(callback: (flag: boolean) => unknown): void };\ndeclare const other: boolean;\nobject.filter(flag => (flag && true) || other);\n"},
		{Code: "declare function assert(x: unknown): asserts x;\ndeclare const a: boolean;\ndeclare const b: boolean;\nassert((a && true) || b);\n"},
		{Code: "declare const items: object[];\ndeclare const b: boolean;\nconst result = (items[0] && true) || b;\n"},
		{Code: "declare const items: object[];\ndeclare const a: boolean;\ndeclare const b: boolean;\nconst result = (a || (items[0] && true)) || b;\n"},
	}, nil)
}

func TestNoUnnecessaryConditionNarrowedComputedReads(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, nil, []rule_tester.InvalidTestCase{
		{Code: "function read(o: {x?: number}) {\n  if (o.x === undefined) return o['x'] ?? 42;\n  return o.x;\n}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysNullish", Line: 2, Column: 33}}},
		{Code: "declare const key: unique symbol;\nfunction read(o: {[key]?: number}) {\n  if (o[key] === undefined) return o[key] ?? 42;\n}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysNullish", Line: 3, Column: 36}}},
	})
}

func TestNoUnnecessaryConditionOmittedForCondition(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{
		{Code: "declare function next(): boolean;\nfor (;; next() || next() || next()) { break; }\n"},
		{Code: "declare function next(): boolean;\nfor (next() || next() || next();;) { break; }\n"},
	}, nil)
}

func TestNoUnnecessaryConditionWrappedCallbacks(t *testing.T) {
	var invalid []rule_tester.InvalidTestCase
	for _, callback := range []string{
		"async () => false",
		"async flag => flag",
		"async flag => (flag && true) || other",
		"async function(flag) { return (flag && true) || other; }",
		"function* () { return false; }",
		"async function* () { return false; }",
	} {
		for _, wrapped := range []string{callback, "((" + callback + "))"} {
			invalid = append(invalid, rule_tester.InvalidTestCase{
				Code:   "declare const flags: boolean[];\ndeclare const other: boolean;\nflags.filter(" + wrapped + ");\n",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthyFunc"}},
			})
		}
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, nil, invalid)
}

func TestNoUnnecessaryConditionNarrowedIndexedReads(t *testing.T) {
	var invalid []rule_tester.InvalidTestCase
	for _, config := range []string{"tsconfig.json", "tsconfig.noUncheckedIndexedAccess.json", "tsconfig.exactOptionalPropertyTypes.json"} {
		for _, test := range []struct{ declaration, object, key string }{
			{"", "Record<string, number | undefined>", "0"},
			{"declare const key: 0;\n", "Record<string, number | undefined>", "key"},
			{"declare const key: unique symbol;\n", "Record<symbol, number | undefined>", "key"},
			{"", "{ value?: number }", "'value'"},
		} {
			for _, body := range []string{
				fmt.Sprintf("table[%s] = 1; return table[%s] ?? 42;", test.key, test.key),
				fmt.Sprintf("if (table[%s] !== undefined) return table[%s] ?? 42;", test.key, test.key),
				fmt.Sprintf("table[%s] = 1; table[%s] ??= 42;", test.key, test.key),
			} {
				invalid = append(invalid, rule_tester.InvalidTestCase{
					Code:     test.declaration + fmt.Sprintf("function read(table: %s) { %s }\n", test.object, body),
					TSConfig: config,
					Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish"}},
				})
			}
		}
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, nil, invalid)
}

func TestNoUnnecessaryConditionAccessorAssignments(t *testing.T) {
	var valid []rule_tester.ValidTestCase
	var invalid []rule_tester.InvalidTestCase
	for _, config := range []string{"tsconfig.json", "tsconfig.noUncheckedIndexedAccess.json", "tsconfig.exactOptionalPropertyTypes.json"} {
		for _, access := range []string{"box.value", "(box['value'])", "box[(key)]"} {
			for _, test := range []struct{ read, write, assignment, message string }{
				{"number | undefined", "null", "??= null", ""},
				{"boolean", "undefined", "&&= undefined", ""},
				{"number", "undefined", "??= undefined", "neverNullish"},
				{"undefined", "number", "??= 1", "alwaysNullish"},
				{"0", "1", "||= 1", "alwaysFalsy"},
				{"true", "false", "&&= false", "alwaysTruthy"},
			} {
				code := fmt.Sprintf("interface Base { get value(): %s; set value(reset: %s); }\ninterface Box extends Base {}\nfunction reset(box: Box) { const key = 'value'; %s %s; }\n", test.read, test.write, access, test.assignment)
				if test.message == "" {
					valid = append(valid, rule_tester.ValidTestCase{Code: code, TSConfig: config})
				} else {
					invalid = append(invalid, rule_tester.InvalidTestCase{Code: code, TSConfig: config, Errors: []rule_tester.InvalidTestCaseError{{MessageId: test.message}}})
				}
			}
		}
		valid = append(valid, rule_tester.ValidTestCase{Code: "class Box { get #value(): number | undefined { return undefined; } set #value(reset: null) {} reset() { this.#value ??= null; } }\n", TSConfig: config})
		invalid = append(invalid,
			rule_tester.InvalidTestCase{Code: "class Box { get #value(): number { return 7; } set #value(reset: undefined) {} reset() { this.#value ??= undefined; } }\n", TSConfig: config, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish"}}},
			rule_tester.InvalidTestCase{Code: "function read(o: {x?: number}) {\n  if (o.x === undefined) return o['x'] ??= 42;\n}\n", TSConfig: config, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysNullish", Line: 2, Column: 33}}},
			rule_tester.InvalidTestCase{Code: "interface Box { get value(): 0; set value(reset: 1); }\nfunction reset(box: Box) { (box['value'])! ||= 1; }\n", TSConfig: config, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy"}}},
			rule_tester.InvalidTestCase{Code: "interface Box { get value(): true; set value(reset: false); }\nfunction reset(box: Box) { box['value']! &&= false; }\n", TSConfig: config, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}}},
		)
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, valid, invalid)
}

func TestNoUnnecessaryConditionAccessorObservations(t *testing.T) {
	var valid []rule_tester.ValidTestCase
	for _, access := range []string{"store.value", "store['value']", "store[(key)]"} {
		for _, statement := range []string{access + " ?? 42;", access + " ??= 42;"} {
			valid = append(valid, rule_tester.ValidTestCase{Code: "class Store {\n  private stored?: number;\n  get value(): number | undefined { return this.stored; }\n  set value(value: number | undefined) { this.stored = value === 0 ? undefined : value; }\n}\nconst store = new Store();\nconst key = 'value';\nstore.value = 0;\n" + statement})
		}
	}
	valid = append(valid,
		rule_tester.ValidTestCase{Code: "class Switch { attempts = 0; private on = false; get enabled(): boolean { return this.on; } set enabled(value: boolean) { this.on = value && ++this.attempts > 1; } }\nconst store = new Switch();\nstore.enabled = true;\nstore.enabled ||= true;\n"},
		rule_tester.ValidTestCase{Code: "const queue = [7];\nconst feed = { get first(): number | undefined { return queue.shift(); } };\nconst value = feed['first'] !== undefined ? feed['first'] ?? 42 : null;\n"},
		rule_tester.ValidTestCase{Code: "interface Store { get value(): number | undefined; set value(value: number); }\nfunction reset(store: Store) { if (store.value !== undefined) store['value'] ??= 42; }\n"},
		rule_tester.ValidTestCase{Code: "interface Store { get value(): number | undefined; set value(value: number); }\nfunction reset(store: Store | { value?: number }) { store.value = 0; store['value'] ??= 42; }\n"},
		rule_tester.ValidTestCase{Code: "interface Store { get value(): { id: number } | undefined; }\nfunction read(store: Store) { if (store.value) return store.value?.id; }\n"},
		rule_tester.ValidTestCase{Code: "interface Store { get ready(): true; }\nfunction read(store: Store | undefined) { if (store?.ready) {} }\n"},
		rule_tester.ValidTestCase{Code: "interface Store { get callback(): (() => number) | undefined; }\nfunction read(store: Store | undefined) { return store?.callback?.(); }\n"},
		rule_tester.ValidTestCase{Code: "declare const key: unique symbol;\ninterface Store { get [key](): boolean; set [key](value: boolean); }\nfunction reset(store: Store) { store[key] = true; store[key] ||= true; }\n"},
		rule_tester.ValidTestCase{Code: "interface Store { get ready(): boolean; set ready(value: boolean); }\nfunction read(store: Store) { store.ready = true; if (store.ready satisfies boolean) {} }\n"},
		rule_tester.ValidTestCase{Code: "interface Store { get value(): number | undefined; }\nfunction read<K extends 'value'>(store: Store, key: K) { if (store[key] !== undefined) return store[key] ?? 42; }\n"},
		rule_tester.ValidTestCase{Code: "export {};\ninterface Store {get value():{ready:true}|undefined;set value(v:{ready:true});} function read(store:Store) {store.value={ready:true}; if(store.value?.ready) {}}\n"},
		rule_tester.ValidTestCase{Code: "export {};\ninterface Store {get value():{ready:true}|undefined;set value(v:{ready:true});} function read(store:Store) {store.value={ready:true}; return store.value?.ready ?? false;}\n"},
		rule_tester.ValidTestCase{Code: "export {};\ninterface Store {get value():{ready:true}|undefined;set value(v:{ready:true});} function read(store:Store) {store.value={ready:true}; return store.value?.ready === true;}\n"},
		rule_tester.ValidTestCase{Code: "export {};\ninterface Store {get callback():(()=>true)|undefined;set callback(v:()=>true);} function read(store:Store) {store.callback=()=>true; if(store.callback?.()) {}}\n"},
		rule_tester.ValidTestCase{Code: "interface Leaf {get item():{id:number};} function f(root:{middle:Leaf}|undefined) {return (root?.middle!.item)?.id;}\n"},
		rule_tester.ValidTestCase{Code: "declare const source: { extra: number };\nconst store = { ...source, get value(): number | undefined { return undefined; }, set value(value: number) {} };\nstore.value = 1;\nstore.value ?? 42;\n"},
		rule_tester.ValidTestCase{Code: "function empty(value: unknown, context: unknown) { return { get(): number | undefined { return undefined; }, set(value: number | undefined) {} }; }\nclass Box { @empty accessor #value: number | undefined; read() { this.#value = 1; return this.#value ?? 42; } }\n"},
		rule_tester.ValidTestCase{Code: "interface Store { get child(): { value?: number }; }\nfunction read(store: Store) { store.child.value = 1; return store.child.value ?? 42; }\n"},
		rule_tester.ValidTestCase{Code: "interface Store { get child(): { nested: { value?: number } }; }\nfunction read(store: Store) { store.child.nested.value = 1; return store.child.nested.value ?? 42; }\n"},
		rule_tester.ValidTestCase{Code: "interface Store { get child(): { value: true } | { value: false }; set child(value: { value: true }); }\nfunction read(store: Store) { store.child = { value: true }; if (store.child.value) {} }\n"},
		rule_tester.ValidTestCase{Code: "interface Store { get callback(): (() => true) | (() => false); set callback(value: () => true); }\nfunction read(store: Store) { store.callback = () => true; if (store.callback()) {} }\n"},
		rule_tester.ValidTestCase{Code: "interface Store { get callback(): (() => { ready: true }) | (() => { ready: false }); set callback(value: () => { ready: true }); }\nfunction read(store: Store) { store.callback = () => ({ ready: true }); if (store.callback().ready) {} }\n"},
		rule_tester.ValidTestCase{Code: "interface Store { get callback(): (() => { ready: true }) | (() => { ready: false }) | undefined; set callback(value: () => { ready: true }); }\nfunction read(store: Store) { store.callback = () => ({ ready: true }); return store.callback?.().ready ?? false; }\n"},
		rule_tester.ValidTestCase{Code: "interface Store { get callback(): (<T>(value: T) => T) | undefined; set callback(value: <T>(value: T) => T); }\nfunction read(store: Store) { store.callback = value => value; if (store.callback?.({})) {} }\n"},
	)
	invalid := []rule_tester.InvalidTestCase{
		{Code: "declare const source: { get value(): number | undefined };\nconst copy = { ...source };\ncopy.value = 1;\ncopy.value ?? 42;\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish"}}},
		{Code: "declare const source: { get value(): number | undefined; set value(value: number) };\nconst copy = { ...source };\ncopy.value = 1;\ncopy['value'] ?? 42;\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish"}}},
		{Code: "class Box { accessor #value: number | undefined; read() { this.#value = 1; return this.#value ?? 42; } }\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish"}}},
		{Code: "class Box { private accessor value: number | undefined; read() { this.value = 1; return this.value ?? 42; } }\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish"}}},
		{Code: "declare const source: { get value(): number | undefined };\nconst copy = { ...source };\nfunction read(value: Pick<typeof copy, 'value'>) { value.value = 1; return value.value ?? 42; }\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish"}}},
		{Code: "declare const source: { get value(): number | undefined; set value(value: number) };\nconst copy = { ...source };\nfunction read(value: Pick<typeof copy, 'value'>) { value.value = 1; return value.value ?? 42; }\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish"}}},
		{Code: "declare const source: { get value(): number | undefined };\nconst copy = { ...source };\nconst key = 'value';\ncopy.value = 1;\ncopy[(key)] ?? 42;\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish"}}},
		{Code: "interface Store { get callback(): (() => true) | ((value?: string) => true); set callback(value: () => true); }\nfunction read(store: Store) { store.callback = () => true; if (store.callback()) {} }\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}}},
	}
	for _, code := range []string{
		"type Read = (value: { value: true }) => true; type Write = (value: { other: number }) => true; interface Store { get callback(): Read | Write; set callback(value: Write); } function read(store: Store, callback: Write) { store.callback = callback; if (store.callback({ other: 1 })) {} }",
		"type Item = { id: number }; type Full = (key: string) => Item; type PartialRead = { (key: 'missing'): undefined; (key: string): Item }; interface Store { get callback(): Full | PartialRead; set callback(value: Full); } function read(store: Store) { store.callback = () => ({ id: 1 }); return store.callback('missing')?.id; }",
		"interface Store { get ready(): boolean; } const store: Store = { get ready(): true { return true; } }; Object.defineProperty(store, 'ready', { get: () => false }); if (store.ready) {}",
		"type Read = <T extends { value: true }>(value: T) => true; type Write = <T extends { other: number }>(value: T) => true; interface Store { get callback(): Read | Write; set callback(value: Write); } function read(store: Store, callback: Write) { store.callback = callback; if (store.callback<{ other: number }>({ other: 1 })) {} }",
		"interface Tree { next: Tree; ready: true; } interface Store { get head(): Tree | undefined; set head(value: Tree); } function read(store: Store, value: Tree) { store.head = value; if (store.head?." + strings.Repeat("next.", 24) + "ready) {} }",
	} {
		valid = append(valid, rule_tester.ValidTestCase{Code: code})
	}
	for _, test := range []struct{ target, access string }{
		{"new Store()", "copy.value"},
		{"target", "copy['value']"},
		{"create()", "copy[(key)]"},
		{"Object.assign(new Store(), {})", "copy.value"},
	} {
		code := "class Store { get value(): number | undefined { return undefined; } set value(value: number | undefined) {} } declare function create(): Store; function read<T extends Store>(target: T) { const key = 'value'; const copy = Object.assign(" + test.target + ", { value: 1 }); copy.value = 1; return " + test.access + " ?? 42; }"
		valid = append(valid, rule_tester.ValidTestCase{Code: code})
	}
	for _, code := range []string{
		"type Item = { id: number }; interface Store { get item(): Item | undefined; set item(value: Item | undefined); } function read(target: object, source: Store) { const copy = Object.assign(target, source); copy.item = { id: 1 }; return copy.item?.id; }",
		"interface Store { get value(): number | undefined; set value(value: number); } function read(source: Store) { const copy = Object.assign(new Proxy({}, { get() { return undefined; }, set() { return true; } }), source); copy.value = 1; return copy.value ?? 42; }",
		"interface Store { get value(): number | undefined; set value(value: number); } function read(source: Store) { const copy = Object.assign({ __proto__: source }, source); copy.value = 1; return copy.value ?? 42; }",
		"interface Store { get value(): number | undefined; set value(value: number); } function read<T extends object>(target: T, source: Store) { const copy = Object.assign(target, source); copy.value = 1; return copy.value ?? 42; }",
		"interface Store { get value(): number | undefined; set value(value: number); } function read(target: { value?: number }, source: Store) { const alias = target; const copy = Object.assign(alias, source); copy.value = 1; return copy.value ?? 42; }",
		"type Item = { id: number }; type Full = (key: string) => Item; type PartialRead = { (key: 'missing'): undefined; (key: string): Item }; interface Store { get callback(): Full | PartialRead; set callback(value: Full); } function read(store: Store) { return store.callback('missing')?.id; }",
	} {
		valid = append(valid, rule_tester.ValidTestCase{Code: code})
	}
	for _, code := range []string{
		"interface Store { get callback(): (<T>(value: T) => T) | ((value: true) => true); } function read(store: Store) { if (store.callback(true)) {} }",
		"type Item = { id: number }; function read(target: object, source: { item: Item | undefined }) { const copy = Object.assign(target, source); copy.item = { id: 1 }; return copy.item?.id; }",
		"type Item = { id: number }; function read(target: object, source: Record<string, Item | undefined>) { const copy = Object.assign(target, source); copy.item = { id: 1 }; return copy.item?.id; }",
		"type Item = { id: number }; function read(target: object, source: Record<string, Item | undefined>, key: string) { const copy = Object.assign(target, source); copy[key] = { id: 1 }; return copy[key]?.id; }",
		"type Item = { ready: true }; function read(target: object, source: Record<string, Item>, key: string) { const copy = Object.assign(target, source); copy[key] = { ready: true }; if (copy[key]?.ready) {} }",
		"type Item = { id: number }; function read(source: { item: Item | undefined }) { const copy = Object.assign({ __proto__: source }, source); copy.item = { id: 1 }; return copy.item?.id; }",
		"interface Store { get item(): { id: number } | undefined; set item(value: { id: number } | undefined); } const target: Store = { get item() { return undefined; }, set item(value) {} }; const copy = Object.assign(target, { item: { id: 1 } }); copy.item = { id: 1 }; copy.item?.id;",
		"function read(target: object, source: { value: true }) { const copy = Object.assign(target, source); copy.value = true; if (copy.value) {} }",
		"type Item = { ready: true }; function read(target: object, source: { item: Item }) { const copy = Object.assign(target, source); copy.item = { ready: true }; if (copy.item?.ready) {} }",
		"type Item = { id: number }; interface Store { get item(): Item | undefined; set item(value: Item | undefined); } class Target { declare item: Item | undefined; constructor() { return new Proxy(this, { get() { return undefined; }, set() { return true; } }); } } function read(source: Store) { const target: Store = new Target(); const copy = Object.assign(target, source); copy.item = { id: 1 }; return copy.item?.id; }",
		"type Item = { id: number }; interface Store { get item(): Item | undefined; set item(value: Item | undefined); } class Target { declare item: Item | undefined; constructor() { return new Proxy(this, { get() { return undefined; }, set() { return true; } }); } } function read(source: Store) { const target: Store = Object.assign(new Target(), {}); const copy = Object.assign(target, source); copy.item = { id: 1 }; return copy.item?.id; }",
		"type Item = { id: number }; interface Store { [key: string]: Item | undefined; } class Target { [key: string]: Item; constructor() { return new Proxy(this, { get() { return undefined; }, set() { return true; } }); } } function read(source: Record<string, Item>, key: string) { const target: Store = Object.assign(new Target(), {}); const copy = Object.assign(target, source); copy[key] = { id: 1 }; return copy[key]?.id; }",
		"type Item = { id: number }; function read(target: object, source: { item: Item }) { const copy = Object.assign(target, source); const item = copy.item; return item?.id; }",
		"type Item = { ready: true }; function read(target: object, source: { item: Item }) { const copy = Object.assign(target, source); const first = copy.item; const item = first; if (item?.ready) {} }",
		"type Item = { id: number }; function read(target: object, source: { make: () => Item }) { const copy = Object.assign(target, source); const make = copy.make; const item = make(); return item?.id; }",
		"type Item = { id: number }; function read(target: object, source: { item: Item }) { const copy = Object.assign(target, source); const { item: first = { id: 0 } } = copy; const item = first; return item?.id; }",
		"type Item = { id: number }; interface Store { get item(): Item | undefined; set item(value: Item); } function read(store: Store) { store.item = { id: 1 }; const item = store.item; return item?.id; }",
		"interface Store { get value(): boolean; set value(value: boolean); } function read(store: Store) { store.value = true; const value = store.value; if (value) {} }",
		"type Item = { id: number }; interface Store { get item(): Item | undefined; set item(value: Item | undefined); } function read(source: Store) { const target: Store = Object.assign(() => 0, { item: undefined }); Object.defineProperty(target, 'item', { get() { return undefined; }, set(value) {} }); const copy = Object.assign(target, source); copy.item = { id: 1 }; return copy.item?.id; }",
	} {
		valid = append(valid, rule_tester.ValidTestCase{Code: code})
	}
	for _, config := range []string{"tsconfig.json", "tsconfig.noUncheckedIndexedAccess.json", "tsconfig.exactOptionalPropertyTypes.json"} {
		for _, code := range []string{
			"function read(extra: { value?: { id: number } }) { const store = { get value(): { id: number } | undefined { return undefined; }, set value(value: { id: number } | undefined) {}, ...extra }; store.value = { id: 1 }; return store.value?.id; }",
			"type Item = { id: number }; interface Store { get item(): Item | undefined; } function read(extra: { item?: Item | undefined }) { const store: Store = { get item(): Item { return { id: 1 }; }, ...extra }; return store.item?.id; }",
			"type Item = { id: number }; interface Store { get item(): Item | undefined; } function read(extra: Record<string, Item | undefined>) { const original = { get item(): Item { return { id: 1 }; }, ...extra }; const store: Store = original; return store.item?.id; }",
			"class Store { get value(): { id: number } | undefined { return undefined; } set value(value: { id: number } | undefined) {} } const copy = Object.assign(new Store(), { value: { id: 1 } }); copy.value = { id: 1 }; copy.value?.id;",
			"declare const source: { value: number; other: number }; const copy = Object.assign({ get value(): number | undefined { return undefined; }, set value(value: number | undefined) {}, other: 1 }, source); function read(key: 'value' | 'other') { copy[key] = 1; return copy[key] ?? 42; }",
			"interface Store { get value(): number | undefined; set value(value: number); } function read<T extends Store>(source: T) { let copy = { ...source }; copy = source; copy.value = 1; return copy.value ?? 42; }",
			"interface Store { get value(): number | undefined; set value(value: number); } function read(source: Store) { const Object = { assign(target: object, value: Store) { return value; } }; const copy = Object.assign({}, source); copy.value = 1; return copy.value ?? 42; }",
			"interface Store { get value(): number | undefined; set value(value: number); } function read(source: Store) { const copy = { __proto__: source } as unknown as Store; copy.value = 1; return copy.value ?? 42; }",
		} {
			valid = append(valid, rule_tester.ValidTestCase{Code: code, TSConfig: config})
		}
		valid = append(valid, rule_tester.ValidTestCase{Code: "class Store { get value(): number | undefined { return undefined; } set value(value: number | undefined) {} other = 1; } const copy = Object.assign(new Store(), { value: 1 }); function read(key: 'value' | 'other') { copy[key] = 1; return copy[key] ?? 42; }", TSConfig: config})
		invalid = append(invalid, rule_tester.InvalidTestCase{Code: "class Store { value: number | undefined; } const copy = Object.assign(new Store(), { value: 1 }); copy.value = 1; copy.value ?? 42;", TSConfig: config, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish"}}})
		for _, access := range []string{"store.value", "store['value']", "store[(key)]"} {
			code := "function read(extra: { value?: number }) { const key = 'value'; const store = { get value(): number | undefined { return undefined; }, set value(value: number | undefined) {}, ...extra }; store.value = 1; return " + access + " ?? 42; }"
			valid = append(valid, rule_tester.ValidTestCase{Code: code, TSConfig: config})
		}
		for _, creation := range []string{"const copy = { ...source };", "const first = { ...source }; const copy = first;", "const copy = Object.assign({}, source);", "const assign = Object.assign; const copy = assign({}, source);", "const Obj = Object; const copy = Obj.assign({}, source);", "const { tag, ...copy } = source;"} {
			code := "interface Store { tag: string; get value(): number | undefined; set value(value: number); } function read<T extends Store>(source: T) { " + creation + " copy.value = 1; return copy.value ?? 42; }"
			invalid = append(invalid, rule_tester.InvalidTestCase{Code: code, TSConfig: config, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish"}}})
		}

	}
	invalid = append(invalid, rule_tester.InvalidTestCase{Code: "interface Store { get value(): number | undefined; set value(value: number); } class Target { value: number | undefined; } function read(source: Store) { const target = new Target(); const copy = Object.assign(target, source); copy.value = 1; return copy.value ?? 42; }", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish"}}})
	invalid = append(invalid, rule_tester.InvalidTestCase{Code: "interface Store { get value(): number | undefined; set value(value: number); } function read<T extends Store>(source: T) { const first = { ...source }; const copy = Object.assign(first, {}); copy.value = 1; return copy.value ?? 42; }", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish"}}})
	invalid = append(invalid, rule_tester.InvalidTestCase{Code: "type Item = { id: number }; function read(source: { item: Item | undefined }) { const copy = Object.assign({ __proto__: source, item: undefined as Item | undefined }, source); copy.item = { id: 1 }; return copy.item ?? { id: 2 }; }", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish"}}})
	for _, code := range []string{
		"const copy = Object.assign(() => 0, { version: 1 }); function read() { return copy.version ?? 42; }",
		"function target() {} const copy = Object.assign(target, { version: 1 }); function read() { return copy.version ?? 42; }",
		"interface Store { get value(): number | undefined; set value(value: number); } function read(source: Store) { const copy = Object.assign([], source); copy.value = 1; return copy.value ?? 42; }",
		"interface Store { get value(): number | undefined; set value(value: number); } function read(source: Store) { const copy = Object.assign({ __proto__: null }, source); copy.value = 1; return copy.value ?? 42; }",
		"type Item = { id: number }; function read(target: object, source: { item: Item }) { const copy = Object.assign(target, source); const item = copy.item; if (item != null) { return item ?? { id: 0 }; } }",
		"type Item = { id: number }; interface Store { get item(): Item | undefined; set item(value: Item); } function read(store: Store) { store.item = { id: 1 }; const item = store.item; item?.id; if (item) { item ?? { id: 0 }; } return item?.id; }",
		"type Item = { id: number }; interface Store { get item(): Item | undefined; set item(value: Item); } function read(store: Store) { store.item = { id: 1 }; const { item = { id: 0 } } = store; return item ?? { id: 0 }; }",
	} {
		invalid = append(invalid, rule_tester.InvalidTestCase{Code: code, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish"}}})
	}
	for _, test := range []struct{ code, message string }{
		{"interface Store { get callback(): (() => true) | ((value: number) => true); } function read(store: Store) { if (store.callback(1)) {} }", "alwaysTruthy"},
		{"interface Store { get callback(): (() => true) | { (value: number): true; (value: string): true }; } function read(store: Store) { if (store.callback(1)) {} }", "alwaysTruthy"},
		{"type Read = { (key: 'yes'): true; (key: 'no'): false }; type Write = (key: 'yes' | 'no') => true; interface Store { get callback(): Read | Write; set callback(value: Write); } function read(store: Store) { if (store.callback('yes')) {} }", "alwaysTruthy"},
		{"type Read = { (key: 'yes'): true; (key: 'no'): false }; type Write = (key: 'yes' | 'no') => true; interface Store { get callback(): Read | Write; set callback(value: Write); } function read(store: Store) { store.callback = () => true; if (store.callback('yes')) {} }", "alwaysTruthy"},
		{"type Read = <T>(value: T) => T; type Write = (value: true) => true; interface Store { get callback(): Read | Write; set callback(value: Write); } function read(store: Store) { store.callback = () => true; if (store.callback(true)) {} }", "alwaysTruthy"},
		{"type Read = <T>(value: T) => T; type Write = (value: 0) => 0; interface Store { get callback(): Read | Write; set callback(value: Write); } function read(store: Store) { store.callback = () => 0; if (store.callback(0)) {} }", "alwaysFalsy"},
	} {
		invalid = append(invalid, rule_tester.InvalidTestCase{Code: test.code, Errors: []rule_tester.InvalidTestCaseError{{MessageId: test.message}}})
	}
	valid = append(valid, rule_tester.ValidTestCase{Code: "export {};\nfunction read(source:{['__proto__']:object}|{unused:true}){const copy=Object.assign(()=>1,source,{item:{id:1}});return copy.item?.id;}\n"})
	valid = append(valid, rule_tester.ValidTestCase{Code: "export {};\nfunction read(source:{['__proto__']:object}|undefined){const copy=Object.assign(()=>1,source,{item:{id:1}});return copy.item?.id;}\n"})
	valid = append(valid, rule_tester.ValidTestCase{Code: "export {};\nfunction read(source:Record<string,object>|{unused:true}){const copy=Object.assign(()=>1,source,{item:{id:1}});return copy.item?.id;}\n"})
	invalid = append(invalid, rule_tester.InvalidTestCase{Code: "export {};\ninterface Wide{get ready():boolean;}interface Narrow{get ready():true;}const target:Wide={get ready():true{return true;}};const narrow=Object.assign(target as Narrow,{}); if(narrow.ready){}const wide=Object.assign(target,{});if(wide.ready){}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 175, EndLine: 2, EndColumn: 187}}})
	invalid = append(invalid, rule_tester.InvalidTestCase{Code: "export {};\ninterface Wide{get ready():boolean;}interface Narrow{get ready():true;}const target:Wide={get ready():true{return true;}};const wide=Object.assign(target,{});if(wide.ready){}const narrow=Object.assign(target as Narrow,{}); if(narrow.ready){}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 227, EndLine: 2, EndColumn: 239}}})
	invalid = append(invalid, rule_tester.InvalidTestCase{Code: "export {};\ninterface Wide{get ready():boolean;}interface Narrow{get ready():true;}const target:Wide={get ready():true{return true;}};const alias=target;const narrow=Object.assign(alias as Narrow,{}); if(narrow.ready){} const wide=Object.assign(alias,{});if(wide.ready){}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 193, EndLine: 2, EndColumn: 205}}})
	invalid = append(invalid, rule_tester.InvalidTestCase{Code: "export {};\ninterface Wide{get ready():boolean;}interface Narrow{get ready():true;}const target:Wide={get ready():true{return true;}};const alias=target as Narrow; const narrow=Object.assign(alias,{}); if(narrow.ready){} const wide=Object.assign(target,{});if(wide.ready){}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 194, EndLine: 2, EndColumn: 206}}})
	invalid = append(invalid, rule_tester.InvalidTestCase{Code: "export {};\ninterface Wide{kind:'wide';get ready():boolean;}interface Narrow{kind:'narrow';get ready():true;}declare function create():Wide|Narrow;const target:Wide|Narrow=create();function read(){if(target.kind==='narrow'){const copy=Object.assign(target,{}); if(copy.ready){}}else{const copy=Object.assign(target,{});if(copy.ready){}}}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 253, EndLine: 2, EndColumn: 263}}})
	invalid = append(invalid, rule_tester.InvalidTestCase{Code: "export {};\ninterface Wide{kind:'wide';get ready():boolean;}interface Narrow{kind:'narrow';get ready():true;}declare function create():Wide|Narrow;const target=create();function read(){if(target.kind==='narrow'){const copy=Object.assign(target,{}); if(copy.ready){}}else{const copy=Object.assign(target,{});if(copy.ready){}}}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 241, EndLine: 2, EndColumn: 251}}})
	// Getter reads remain independent of result annotations and descriptor changes.
	for _, code := range []string{
		"export {};\ntype Item={id:number};interface Store{get item():Item|undefined;set item(value:Item);}   function f(){const copy:{item:Item}=Object.assign({get item():Item|undefined{return undefined;},set item(v:Item){}},{item:{id:1}}); return copy.item?.id;}\n",
		"export {};\ntype Item={id:number};interface Store{get item():Item|undefined;set item(value:Item);}  const target={get item():Item|undefined{return undefined;},set item(v:Item){}}; function f(){const copy:{get item():Item}=Object.assign(target,{item:{id:1}}); return copy.item?.id;}\n",
		"export {};\ntype Item={id:number};interface Store{get item():Item|undefined;set item(value:Item);}  class Box {get item():Item|undefined{return undefined;} set item(v:Item){}} function f(){const copy:{item:Item}=Object.assign(new Box(),{item:{id:1}}); return copy.item?.id;}\n",
		"export {};\ntype Item={id:number};interface Store{get item():Item|undefined;set item(value:Item);}   function f(target:Store){const copy:{item:Item}=Object.assign(target,{item:{id:1}}); return copy.item?.id;}\n",
		"export {};\ntype Item={id:number};interface Store{get item():Item|undefined;set item(value:Item);}  const target={get item():Item|undefined{return undefined;},set item(v:Item){}}; function f(){const copy:{item:Item|undefined}=Object.assign(target,{item:{id:1}}); return copy.item?.id;}\n",
		"export {};\ntype Item={id:number}; function fn(){} function f(){Object.defineProperty(fn,'item',{get(){return undefined;},set(v:unknown){}});const out=Object.assign(fn,{item:{id:1}});return out.item?.id;}\n",
		"export {};\ntype Item={id:number}; const fn=()=>1; function f(){const out=Object.assign(fn,{item:{id:1} as Item|undefined});Object.defineProperty(out,'item',{get(){return undefined;},set(v:unknown){}});out.item={id:1};return out.item?.id;}\n",
		"export {};\ntype Item={id:number}; const xs:number[]=[]; function f(){Object.defineProperty(xs,'item',{get(){return undefined;},set(v:unknown){}});const out=Object.assign(xs,{item:{id:1}});return out.item?.id;}\n",
		"export {};\ntype Item={id:number}; function f(target:object, source:{item:Item|undefined},flag:boolean){const copy=Object.assign(target,source);copy.item={id:1};const saved:Item|undefined=copy.item;return saved?.id;}\n",
		"export {};\ntype Item={id:number}; function f(target:object, source:{item:Item|undefined},flag:boolean){const copy=Object.assign(target,source);copy.item={id:1};const {item:saved}:{item:Item|undefined}=copy;return saved?.id;}\n",
		"export {};\ntype Item={id:number};interface Store{get item():Item|undefined;set item(value:Item);}  function f(a:Store,b:Store){ a.item={id:1}; b.item={id:2}; const {item=b.item}=a;return item?.id;}\n",
		"export {};\ntype Item={id:number};interface Store{get item():Item|undefined;set item(value:Item);}  function f(a:Store,b:Store){a.item={id:1};b.item={id:2};const fallback=b.item;const {item=fallback}=a;return item?.id;}\n",
		"export {};\ntype Item={id:number}; function target() {} function read(){const alias=target; Object.assign(alias,{}); Object.defineProperty(alias,'item',{get(){return undefined},set(value:unknown){},configurable:true});const copy=Object.assign(target,{item:{id:1} as Item|undefined});return copy.item?.id;}\n",
		"export {};\ntype Item={id:number}; const target=()=>1; function read(){const copy=Object.assign(target,{item:{id:1} as Item|undefined});const alias=copy; Object.assign(alias,{}); Object.defineProperty(alias,'item',{get(){return undefined},set(value:unknown){},configurable:true});copy.item={id:1};return copy.item?.id;}\n",
		"export {};\ntype Item={id:number}; const target:number[]=[]; function read(){const define=Object.defineProperty; define(target,'item',{get(){return undefined},set(value:unknown){},configurable:true});const copy=Object.assign(target,{item:{id:1} as Item|undefined});return copy.item?.id;}\n",
		"export {};\ntype Item={id:number}; const target:number[]=[]; function read(){const first=Object.assign(target,{});Object.defineProperty(first,'item',{get(){return undefined},set(value:unknown){},configurable:true});const copy=Object.assign(first,{item:{id:1}});return copy.item?.id;}\n",
		"export {};\ntype Item={id:number}; function target() {} function read(){Object.setPrototypeOf(target,{get item(){return undefined},set item(value:unknown){}});const copy=Object.assign(target,{item:{id:1} as Item|undefined});return copy.item?.id;}\n",
		"export {};\ntype Item={id:number}; const target=()=>1; function read(){const copy=Object.assign(target,{item:{id:1} as Item|undefined});const holder={copy};const mutate=(value:{copy:object})=>Object.defineProperty(value.copy,'item',{get(){return undefined},set(value:unknown){},configurable:true});mutate(holder);copy.item={id:1};return copy.item?.id;}\n",
		"export {};\ntype Item={id:number}; const target:number[]=[]; function read(){const copy=Object.assign(target,{item:{id:1} as Item|undefined});delete (copy as Partial<typeof copy>).item;(copy as unknown as {__proto__:object}).__proto__={get item(){return undefined},set item(value:unknown){}};copy.item={id:1};return copy.item?.id;}\n",
		"export {};\nfunction read(){const first=Object.assign(([]),{item:{id:1}});Object.defineProperty(first,'item',{get(){return undefined},set(v:unknown){}});const copy=Object.assign(first,{});return copy.item?.id;}\n",
		"export {};\nconst target=function self(){Object.defineProperty(self,'item',{get(){return undefined},set(v:unknown){}});};target();const copy=Object.assign(target,{item:{id:1}});function read(){return copy.item?.id;}\n",
		"export {};\nfunction read(){const target:number[]=[];Object.defineProperty(target,'0',{get(){return undefined},set(v:unknown){}});const copy=Object.assign(target,{0:{id:1}});return copy[0]?.id;}\n",
		"export {};\nfunction read(){const copy=Object.assign(()=>1,{['__proto__']:{get item(){return undefined},set item(v:unknown){}}},{item:{id:1}});return copy.item?.id;}\n",
		"export {};\nconst proto:Record<string,unknown>={['__proto__']:{get item(){return undefined},set item(v:unknown){}}};function read(){const copy=Object.assign([],proto,{item:{id:1}});return copy.item?.id;}\n",
	} {
		valid = append(valid, rule_tester.ValidTestCase{Code: code})
	}
	invalid = append(invalid,
		rule_tester.InvalidTestCase{Code: "export {};\n type Item={id:number}; type Obj={item:Item|undefined};interface Holder{get obj():Obj|undefined;set obj(x:Obj);} function f(h:Holder){h.obj={item:undefined};const v=h.obj;if(v){v.item={id:1};return v.item?.id;}}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain", Line: 2, Column: 205, EndLine: 2, EndColumn: 207, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveOptionalChain", Output: "export {};\n type Item={id:number}; type Obj={item:Item|undefined};interface Holder{get obj():Obj|undefined;set obj(x:Obj);} function f(h:Holder){h.obj={item:undefined};const v=h.obj;if(v){v.item={id:1};return v.item.id;}}\n"}}}}},
		rule_tester.InvalidTestCase{Code: "export {};\n type Item={id:number}; type Obj={item:Item|undefined};interface Holder{get obj():Obj|undefined;set obj(x:Obj);} function f(h:Holder){h.obj={item:undefined};const v=h.obj;if(v){if(v.item!=null){return v.item?.id;}}}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain", Line: 2, Column: 208, EndLine: 2, EndColumn: 210, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveOptionalChain", Output: "export {};\n type Item={id:number}; type Obj={item:Item|undefined};interface Holder{get obj():Obj|undefined;set obj(x:Obj);} function f(h:Holder){h.obj={item:undefined};const v=h.obj;if(v){if(v.item!=null){return v.item.id;}}}\n"}}}}},
		rule_tester.InvalidTestCase{Code: "export {};\ntype Item={id:number}; type Obj={item:Item|undefined};interface Holder{get obj():Obj|undefined;set obj(x:Obj);} function f(h:Holder){h.obj={item:undefined};const v=h.obj;if(v){v.item={id:1}; return v.item ?? {id:0};}}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish"}}},
		rule_tester.InvalidTestCase{Code: "export {};\n type Item={id:number};interface Store{get item():Item|undefined;set item(value:Item);}  function f(s:Store){s.item={id:1};const item=s.item;if(item!=null){return item?.id}}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain", Line: 2, Column: 168, EndLine: 2, EndColumn: 170, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveOptionalChain", Output: "export {};\n type Item={id:number};interface Store{get item():Item|undefined;set item(value:Item);}  function f(s:Store){s.item={id:1};const item=s.item;if(item!=null){return item.id}}\n"}}}}},
		rule_tester.InvalidTestCase{Code: "export {};\n type Item={id:number};interface Store{get item():Item|undefined;set item(value:Item);}  function f(a:Store){a.item={id:1};const {item={id:2}}=a;return item?.id;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain", Line: 2, Column: 157, EndLine: 2, EndColumn: 159, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveOptionalChain", Output: "export {};\n type Item={id:number};interface Store{get item():Item|undefined;set item(value:Item);}  function f(a:Store){a.item={id:1};const {item={id:2}}=a;return item.id;}\n"}}}}},
		rule_tester.InvalidTestCase{Code: "export {};\ninterface S{get run():((x:number)=>true) | { (x:number,y:string):true; (x:number,y:boolean):true; };} function f(s:S){ if(s.run(1,'x')) {}}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy"}}},
		rule_tester.InvalidTestCase{Code: "export {};\ntype Item={id:number}; const target:number[]=[]; function read(){const copy=Object.assign(target,{item:{id:1} as Item|undefined});copy.item={id:2};return copy.item?.id;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain", Line: 2, Column: 164, EndLine: 2, EndColumn: 166, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveOptionalChain", Output: "export {};\ntype Item={id:number}; const target:number[]=[]; function read(){const copy=Object.assign(target,{item:{id:1} as Item|undefined});copy.item={id:2};return copy.item.id;}\n"}}}}},
		rule_tester.InvalidTestCase{Code: "export {};\nfunction read(){const copy=Object.assign([] as number[],{item:{id:1}});copy[0]=1;return copy.item?.id;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain", Line: 2, Column: 98, EndLine: 2, EndColumn: 100, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveOptionalChain", Output: "export {};\nfunction read(){const copy=Object.assign([] as number[],{item:{id:1}});copy[0]=1;return copy.item.id;}\n"}}}}},
		rule_tester.InvalidTestCase{Code: "export {};\ntype Item={id:number}; const target=()=>1; function read(){const copy=Object.assign(target,{item:{id:1} as Item|undefined});const alias=copy;alias.item={id:2};return alias.item?.id;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain", Line: 2, Column: 177, EndLine: 2, EndColumn: 179, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveOptionalChain", Output: "export {};\ntype Item={id:number}; const target=()=>1; function read(){const copy=Object.assign(target,{item:{id:1} as Item|undefined});const alias=copy;alias.item={id:2};return alias.item.id;}\n"}}}}},
	)
	// Public field contracts and descriptor changes survive captured reads.
	for _, code := range []string{
		"export {};\ntype Item={id:number};type Narrow={item:Item};type Wide={item:Item|undefined};interface Outer{get box():Narrow|undefined;set box(v:Narrow);}function clear(x:Wide){x.item=undefined;}function f(h:Outer){h.box={item:{id:1}};const v:Wide=h.box;clear(v);return v.item?.id;}\n",
		"export {};\ntype Item={id:number}; type Narrow={item:Item};type Wide={item:Item|undefined};interface Outer{get box():Narrow|undefined;set box(v:Wide);}function clear(x:Wide){x.item=undefined;}function f(h:Outer){h.box={item:{id:1}};const v:any=h.box;if(!v)return;clear(v);return v.item?.id;}\n",
		"export {};\ntype Item={id:number}; type Narrow={item:Item};type Wide={item:Item|undefined};interface Outer{get box():Narrow|undefined;set box(v:Wide);}function clear(x:Wide){x.item=undefined;}function f(h:Outer){h.box={item:{id:1}};const v:Wide|undefined=h.box;if(!v)return;clear(v);return v.item?.id;}\n",
		"export {};\ntype Item={id:number}; type Data={item:Item|undefined};interface Public{get item():Item|undefined;set item(v:Item|undefined);} interface Outer{get box():Data|undefined;set box(v:Data);}function f(h:Outer){h.box={item:undefined};const v:Public|undefined=h.box;if(v){v.item={id:1};return v.item?.id;}}\n",
		"export {};\ntype Item={id:number}; type Wide={item:Item|undefined}; type Narrow={item:Item}; interface Store{get box():Narrow|undefined;set box(v:Narrow);} function f(s:Store){s.box={item:{id:1}};const v:Wide=s.box;Object.assign(v,{item:undefined});return v.item?.id;}\n",
		"export {};\ntype Item={id:number}; type Wide={item:Item|undefined}; type Narrow={item:Item}; interface Store{get box():Narrow|undefined;set box(v:Narrow);} function f(s:Store){s.box={item:{id:1}};const v:Wide=s.box;const alias=v;alias.item=undefined;return v.item?.id;}\n",
		"export {};\ntype Item={id:number}; type Wide={item:Item|undefined}; type Narrow={item:Item}; interface Store{get box():Narrow|undefined;set box(v:Narrow);} function f(s:Store){s.box={item:{id:1}};const v:Wide=s.box;function clear(x:Wide){x.item=undefined}clear(v);return v.item?.id;}\n",
		"export {};\ntype Item={id:number}; function f(){const target=()=>0;Object.defineProperty(target,'item',{get(){return undefined;},set(_v:unknown){},configurable:true});const copy=Object.assign(target,{item:{id:1}});return copy.item?.id;}\n",
		"export {};\ntype Item={id:number}; function f(){const target=[] as number[];Object.defineProperty(target,'item',{get(){return undefined;},set(_v:unknown){},configurable:true});const copy=Object.assign(target,{item:{id:1}});return copy.item?.id;}\n",
		"export {};\ntype Item={id:number}; function f(){const target={};Object.defineProperty(target,'item',{get(){return undefined;},set(_v:unknown){},configurable:true});const copy=Object.assign(target,{item:{id:1}});return copy.item?.id;}\n",
		"export {};\ntype Item={id:number}; function f(){const target={item:undefined as Item|undefined};Object.defineProperty(target,'item',{get(){return undefined;},set(_v:unknown){},configurable:true});const copy=Object.assign(target,{item:{id:1}});return copy.item?.id;}\n",
		"export {};\ntype Item={id:number}; function f(){const target={__proto__:null};Object.defineProperty(target,'item',{get(){return undefined;},set(_v:unknown){},configurable:true});const copy=Object.assign(target,{item:{id:1}});return copy.item?.id;}\n",
		"export {};\ntype Item={id:number}; function f(){const target={...{}};Object.defineProperty(target,'item',{get(){return undefined;},set(_v:unknown){},configurable:true});const copy=Object.assign(target,{item:{id:1}});return copy.item?.id;}\n",
		"export {};\ntype Item={id:number}; function f(target:Record<string,Item|undefined>,source:Record<string,Item>,key:string){const copy:Record<string,Item>=Object.assign(target,source);return copy[key]?.id;}\n",
		"export {};\ntype Item={id:number}; function f(target:Record<string,Item|undefined>,source:Record<string,Item>,key:string){const copy=Object.assign(target,source);return copy[key]?.id;}\n",
		"export {};\ntype Item={id:number}; function f(target:Record<number,Item|undefined>,source:Record<number,Item>,key:number){const copy:Record<number,Item>=Object.assign(target,source);return copy[key]?.id;}\n",
		"export {};\ntype Item={id:number}; function f(target:Record<number,Item|undefined>,source:Record<number,Item>,key:number){const copy=Object.assign(target,source);return copy[key]?.id;}\n",
		"export {};\ntype Item={id:number}; function f(target:Record<symbol,Item|undefined>,source:Record<symbol,Item>,key:symbol){const copy:Record<symbol,Item>=Object.assign(target,source);return copy[key]?.id;}\n",
		"export {};\ntype Item={id:number}; function f(target:Record<symbol,Item|undefined>,source:Record<symbol,Item>,key:symbol){const copy=Object.assign(target,source);return copy[key]?.id;}\n",
		"export {};\ntype Narrow={ready:true};type Wide={ready:boolean};interface Outer{get box():Narrow|undefined;set box(v:Wide);}function set(v:Wide){v.ready=false;}function f(s:Outer){s.box={ready:true};const v:Wide|undefined=s.box;if(v!=null){set(v);if(v.ready) {}}}\n",
		"export {};\nfunction f(){const copy=Object.assign({},{version:1});Object.defineProperty(copy,\"version\",{get(){return undefined;}});/*probe*/ return copy.version ?? 0;}\n",
		"export {};\nfunction f(){const copy=Object.assign({},{version:1});const key=String(Math.random());Object.defineProperty(copy,key,{value:undefined});/*probe*/ return copy.version ?? 0;}\n",
		"export {};\nfunction f(){const copy=Object.assign({},{version:1});Object.defineProperty(copy,\"tag\",{get value(){Object.defineProperty(copy,\"version\",{value:undefined});return 2;}});/*probe*/ return copy.version ?? 0;}\n",
		"export {};\nfunction f(){const copy=Object.assign({},{version:1});Object.defineProperty(copy,\"tag\",{__proto__:{get value(){Object.defineProperty(copy,\"version\",{value:undefined});return 2;}}} as PropertyDescriptor);/*probe*/ return copy.version ?? 0;}\n",
		"export {};\nfunction f(){const copy=Object.assign({},{version:1});Object.defineProperty(copy,\"tag\",{get(){Object.defineProperty(copy,\"version\",{value:undefined});return 2;},enumerable:true});Object.values(copy);/*probe*/ return copy.version ?? 0;}\n",
		"export {};\nfunction f(){const copy=Object.assign({},{version:1});function change(x:object){Object.defineProperty(x,\"version\",{value:undefined})}change(copy);/*probe*/ return copy.version ?? 0;}\n",
		"export {};\nfunction f(){const copy=globalThis.Object.assign({},{version:1});const Object={keys(x:object){globalThis.Object.defineProperty(x,\"version\",{value:undefined})}};Object.keys(copy);/*probe*/ return copy.version ?? 0;}\n",
		"export {};\nfunction f(){const copy=Object.assign({},{version:1});for(let i=0;i<2;i++){/*probe*/ const result=copy.version ?? 0;Object.defineProperty(copy,\"version\",{value:undefined});}}\n",
		"export {};\nconst copy=Object.assign({},{version:1});function f(){/*probe*/ const result=copy.version ?? 0;Object.defineProperty(copy,\"version\",{value:undefined});return result;}\n",
		"export {};\nfunction f(){const copy=Object.assign({},{version:1});function change(){Object.defineProperty(copy,\"version\",{value:undefined});}change();/*probe*/ return copy.version ?? 0;}\n",
		"export {};\nfunction f(){const target:number[]=[];const copy=Object.assign(target,{version:1});copy.push=function(){Object.defineProperty(copy,\"version\",{value:undefined});return 1;};copy.push(1);/*probe*/ return copy.version ?? 0;}\n",
		"export {};\nfunction f(){const target:number[]=[];const copy=Object.assign(target,{version:1});target.push=function(){Object.defineProperty(this,\"version\",{value:undefined});return 1;};copy.push(1);/*probe*/ return copy.version ?? 0;}\n",
		"export {};\nfunction f(){const target:number[]=[];const copy=Object.assign(target,{version:1});Object.defineProperty(target,\"push\",{value:function(){Object.defineProperty(this,\"version\",{value:undefined});return 1;}});copy.push(1);/*probe*/ return copy.version ?? 0;}\n",
		"export {};\nfunction f(){const target:number[]=[];const copy=Object.assign(target,{version:1});Array.prototype.push=function(){Object.defineProperty(this,\"version\",{value:undefined});return 1;};copy.push(1);/*probe*/ return copy.version ?? 0;}\n",
		"export {};\nfunction f(){const target:number[]=[];const copy=Object.assign(target,{version:1});function compare(){Object.defineProperty(copy,\"version\",{value:undefined});return 0;}copy.sort(compare);/*probe*/ return copy.version ?? 0;}\n",
		"export {};\nfunction f(){const copy=Object.assign([] as number[],{version:1,push(){Object.defineProperty(this,\"version\",{value:undefined});return 1;}});copy.push(1);/*probe*/ return copy.version ?? 0;}\n",
		"export {};\nfunction f(){const copy=Object.assign([] as number[],{version:1});copy.push(1);/*probe*/ return copy.length ?? 0;}\n",
		"export {};\nfunction f(){const copy=Object.assign(()=>1,{version:1});Object.keys=(value:object)=>{Object.defineProperty(value,\"version\",{value:undefined});return [];};Object.keys(copy);/*probe*/ return copy.version ?? 0;}\n",
	} {
		valid = append(valid, rule_tester.ValidTestCase{Code: code})
	}
	invalid = append(invalid,
		rule_tester.InvalidTestCase{Code: "export {};\ntype Item={id:number}; type Data={item:Item|undefined};interface Public{get item():Item|undefined;set item(v:Item|undefined);} interface Outer{get box():Data|undefined;set box(v:Data);}function f(h:Outer){h.box={item:undefined};const v:Data|undefined=h.box;if(v){v.item={id:1};return v.item?.id;}}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain", Line: 2, Column: 291, EndLine: 2, EndColumn: 293, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveOptionalChain", Output: "export {};\ntype Item={id:number}; type Data={item:Item|undefined};interface Public{get item():Item|undefined;set item(v:Item|undefined);} interface Outer{get box():Data|undefined;set box(v:Data);}function f(h:Outer){h.box={item:undefined};const v:Data|undefined=h.box;if(v){v.item={id:1};return v.item.id;}}\n"}}}}},
		rule_tester.InvalidTestCase{Code: "export {};\ntype Narrow={ready:true};type Wide={ready:boolean};interface Outer{get box():Narrow|undefined;set box(v:Wide);}function set(v:Wide){v.ready=false;}function f(s:Outer){s.box={ready:true};const v:Wide|undefined=s.box;if(v!=null){v.ready=false;/*probe*/ if(v.ready) {}}}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysFalsy", Line: 2, Column: 255, EndLine: 2, EndColumn: 262}}},
		rule_tester.InvalidTestCase{Code: "export {};\ntype Item={id:number};function f(){const xs:(Item|undefined)[]=[];const copy=Object.assign(xs,{version:1});copy.push({id:0});/*probe*/ return copy.version ?? 0;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish", Line: 2, Column: 143, EndLine: 2, EndColumn: 155}}},
		rule_tester.InvalidTestCase{Code: "export {};\ntype Item={id:number};function f(){const xs=Object.assign([] as number[],{version:1});/*probe*/ const v=xs.version ?? 0;xs.push(2);return v;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish", Line: 2, Column: 105, EndLine: 2, EndColumn: 115}}},
		rule_tester.InvalidTestCase{Code: "export {};\ntype Item={id:number};function f(){const fn=Object.assign(()=>1,{version:1});/*probe*/ const v=fn.version ?? 0;console.log(fn);return v;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish", Line: 2, Column: 96, EndLine: 2, EndColumn: 106}}},
		rule_tester.InvalidTestCase{Code: "export {};\ntype Item={id:number};function f(){const fn=Object.assign(()=>1,{version:1});/*probe*/ const v=fn.version ?? 0;fn();return v;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish", Line: 2, Column: 96, EndLine: 2, EndColumn: 106}}},
		rule_tester.InvalidTestCase{Code: "export {};\ndeclare function publish(v:unknown):void;function f(){const copy=Object.assign(()=>0,{version:1});/*probe*/ const result=copy.version ?? 0;return result;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish", Line: 2, Column: 122, EndLine: 2, EndColumn: 134}}},
		rule_tester.InvalidTestCase{Code: "export {};\ndeclare function publish(v:unknown):void;function f(){const copy=Object.assign({},{version:1});/*probe*/ const result=copy.version ?? 0;publish(copy);return result;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish", Line: 2, Column: 119, EndLine: 2, EndColumn: 131}}},
		rule_tester.InvalidTestCase{Code: "export {};\ndeclare function publish(v:unknown):void;function f(){const copy=Object.assign(()=>0,{version:1});Object.freeze(copy);/*probe*/ const result=copy.version ?? 0;return result;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish", Line: 2, Column: 142, EndLine: 2, EndColumn: 154}}},
		rule_tester.InvalidTestCase{Code: "export {};\ndeclare function publish(v:unknown):void;function f(){const copy=Object.assign(()=>0,{version:1});Object.keys(copy);/*probe*/ const result=copy.version ?? 0;return result;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish", Line: 2, Column: 140, EndLine: 2, EndColumn: 152}}},
		rule_tester.InvalidTestCase{Code: "export {};\ndeclare function publish(v:unknown):void;function f(){const copy=Object.assign([] as number[],{version:1});/*probe*/ const result=copy.version ?? 0;return result;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish", Line: 2, Column: 131, EndLine: 2, EndColumn: 143}}},
		rule_tester.InvalidTestCase{Code: "export {};\ndeclare function publish(v:unknown):void;function f(){const copy=Object.assign([] as number[],{version:1});/*probe*/ const result=copy.version ?? 0;Object.defineProperty(copy,'version',{value:2});return result;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish", Line: 2, Column: 131, EndLine: 2, EndColumn: 143}}},
		rule_tester.InvalidTestCase{Code: "export {};\ndeclare function publish(v:unknown):void;function f(){const copy=Object.assign([] as number[],{version:1});Object.assign(copy,{});/*probe*/ const result=copy.version ?? 0;return result;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish", Line: 2, Column: 154, EndLine: 2, EndColumn: 166}}},
		rule_tester.InvalidTestCase{Code: "export {};\ndeclare function publish(v:unknown):void;function f(){const copy=Object.assign({},{version:1});/*probe*/ const result=copy.version ?? 0;Object.defineProperty(copy,'version',{value:2});return result;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish", Line: 2, Column: 119, EndLine: 2, EndColumn: 131}}},
		rule_tester.InvalidTestCase{Code: "export {};\ndeclare function publish(v:unknown):void;function f(){const copy=Object.assign([] as number[],{version:1});Object.freeze(copy);/*probe*/ const result=copy.version ?? 0;return result;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish", Line: 2, Column: 151, EndLine: 2, EndColumn: 163}}},
		rule_tester.InvalidTestCase{Code: "export {};\ndeclare function publish(v:unknown):void;function f(){const copy=Object.assign({},{version:1});Object.freeze(copy);/*probe*/ const result=copy.version ?? 0;return result;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish", Line: 2, Column: 139, EndLine: 2, EndColumn: 151}}},
		rule_tester.InvalidTestCase{Code: "export {};\ndeclare function publish(v:unknown):void;function f(){const copy=Object.assign(()=>0,{version:1});Object.defineProperty(copy,'tag',{value:2});/*probe*/ const result=copy.version ?? 0;return result;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish", Line: 2, Column: 166, EndLine: 2, EndColumn: 178}}},
		rule_tester.InvalidTestCase{Code: "export {};\ndeclare function publish(v:unknown):void;function f(){const copy=Object.assign([] as number[],{version:1});Object.defineProperty(copy,'tag',{value:2});/*probe*/ const result=copy.version ?? 0;return result;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish", Line: 2, Column: 175, EndLine: 2, EndColumn: 187}}},
		rule_tester.InvalidTestCase{Code: "export {};\ndeclare function publish(v:unknown):void;function f(){const copy=Object.assign(()=>0,{version:1});/*probe*/ const result=copy.version ?? 0;Object.defineProperty(copy,'version',{value:2});return result;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish", Line: 2, Column: 122, EndLine: 2, EndColumn: 134}}},
		rule_tester.InvalidTestCase{Code: "export {};\ndeclare function publish(v:unknown):void;function f(){const copy=Object.assign({},{version:1});Object.defineProperty(copy,'tag',{value:2});/*probe*/ const result=copy.version ?? 0;return result;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish", Line: 2, Column: 163, EndLine: 2, EndColumn: 175}}},
		rule_tester.InvalidTestCase{Code: "export {};\ndeclare function publish(v:unknown):void;function f(){const copy=Object.assign([] as number[],{version:1});/*probe*/ const result=copy.version ?? 0;publish(copy);return result;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish", Line: 2, Column: 131, EndLine: 2, EndColumn: 143}}},
		rule_tester.InvalidTestCase{Code: "export {};\ndeclare function publish(v:unknown):void;function f(){const copy=Object.assign(()=>0,{version:1});Object.assign(copy,{});/*probe*/ const result=copy.version ?? 0;return result;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish", Line: 2, Column: 145, EndLine: 2, EndColumn: 157}}},
		rule_tester.InvalidTestCase{Code: "export {};\ndeclare function publish(v:unknown):void;function f(){const copy=Object.assign({},{version:1});Object.keys(copy);/*probe*/ const result=copy.version ?? 0;return result;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish", Line: 2, Column: 137, EndLine: 2, EndColumn: 149}}},
		rule_tester.InvalidTestCase{Code: "export {};\ndeclare function publish(v:unknown):void;function f(){const copy=Object.assign({},{version:1});/*probe*/ const result=copy.version ?? 0;return result;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish", Line: 2, Column: 119, EndLine: 2, EndColumn: 131}}},
		rule_tester.InvalidTestCase{Code: "export {};\ndeclare function publish(v:unknown):void;function f(){const copy=Object.assign([] as number[],{version:1});Object.keys(copy);/*probe*/ const result=copy.version ?? 0;return result;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish", Line: 2, Column: 149, EndLine: 2, EndColumn: 161}}},
		rule_tester.InvalidTestCase{Code: "export {};\ndeclare function publish(v:unknown):void;function f(){const copy=Object.assign(()=>0,{version:1});/*probe*/ const result=copy.version ?? 0;publish(copy);return result;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish", Line: 2, Column: 122, EndLine: 2, EndColumn: 134}}},
		rule_tester.InvalidTestCase{Code: "export {};\ndeclare function publish(v:unknown):void;function f(){const copy=Object.assign({},{version:1});Object.assign(copy,{});/*probe*/ const result=copy.version ?? 0;return result;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish", Line: 2, Column: 142, EndLine: 2, EndColumn: 154}}},
		rule_tester.InvalidTestCase{Code: "export {};\nfunction f(){const copy=Object.assign({},{version:1});const keys=Object.keys;keys(copy);/*probe*/ return copy.version ?? 0;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish", Line: 2, Column: 106, EndLine: 2, EndColumn: 118}}},
		rule_tester.InvalidTestCase{Code: "export {};\nfunction f(){const copy=Object.assign({},{version:1});Object.getOwnPropertyNames(copy);/*probe*/ return copy.version ?? 0;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish", Line: 2, Column: 105, EndLine: 2, EndColumn: 117}}},
		rule_tester.InvalidTestCase{Code: "export {};\nfunction f(){const copy=Object.assign({},{version:1});Object.getOwnPropertyDescriptors(copy);/*probe*/ return copy.version ?? 0;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish", Line: 2, Column: 111, EndLine: 2, EndColumn: 123}}},
		rule_tester.InvalidTestCase{Code: "export {};\nfunction f(){const copy=Object.assign({},{version:1});Object.assign(copy,null,undefined,{});/*probe*/ return copy.version ?? 0;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish", Line: 2, Column: 110, EndLine: 2, EndColumn: 122}}},
		rule_tester.InvalidTestCase{Code: "export {};\nfunction f(){const copy=Object.assign({},{version:1});Object.defineProperty(copy,\"tag\",{value:2});/*probe*/ return copy.version ?? 0;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish", Line: 2, Column: 116, EndLine: 2, EndColumn: 128}}},
		rule_tester.InvalidTestCase{Code: "export {};\nfunction f(){const target:number[]=[];const copy=Object.assign(target,{version:1});const alias=copy;alias.push(1);/*probe*/ return copy.version ?? 0;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish", Line: 2, Column: 132, EndLine: 2, EndColumn: 144}}},
	)
	valid = append(valid, rule_tester.ValidTestCase{Code: "export {};\ntype Item={id:number};type Wide={item:Item|undefined};type Narrow={item:Item};interface Store{get box():Narrow|undefined;set box(v:Narrow);}function f(s:Store){s.box={item:{id:1}};const v:Wide & { tag?: string }=s.box;Object.assign(v,{item:undefined});return v.item?.id;}\n"})
	valid = append(valid, rule_tester.ValidTestCase{Code: "export {};\ntype Item={id:number};type Wide={item:Item|undefined};type Narrow={item:Item};interface Store{get box():Narrow|undefined;set box(v:Narrow);}function f(s:Store){s.box={item:{id:1}};const v:(Wide & { tag?: string }) | undefined=s.box;Object.assign(v,{item:undefined});return v.item?.id;}\n"})
	valid = append(valid, rule_tester.ValidTestCase{Code: "export {};\ntype Item={id:number};type Wide={item:Item|undefined};type Narrow={item:Item};interface Store{get box():Narrow|undefined;set box(v:Narrow);}function f(s:Store){s.box={item:{id:1}};const v:Readonly<Wide> & { tag?: string }=s.box;Object.assign(v,{item:undefined});return v.item?.id;}\n"})
	invalid = append(invalid, rule_tester.InvalidTestCase{Code: "export {};\ninterface Store{get value():true;set value(value:boolean);}function f(store:Store){store.value=true;const value:boolean=store.value;/*probe*/ if(value){}}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 146, EndLine: 2, EndColumn: 151}}})
	valid = append(valid, rule_tester.ValidTestCase{Code: "export {};\ntype Item={id:number};type Wide={item:Item|undefined}|({item:Item|null}&{tag?:true});interface Store{get box():{item:Item}|undefined;set box(value:{item:Item});}function f(store:Store){store.box={item:{id:1}};const value:Wide=store.box;Object.assign(value,{item:undefined});return value.item?.id;}\n"})
	valid = append(valid, rule_tester.ValidTestCase{Code: "export {};\nfunction f(){Object.defineProperty(Object.prototype,'item',{get(){return undefined;},set(_value:unknown){},configurable:true});const copy=Object.assign({},{item:{id:1}});return copy.item?.id;}\n"})
	valid = append(valid, rule_tester.ValidTestCase{Code: "export {};\nfunction f(){Object.defineProperty(Object.prototype,'item',{get(){return undefined;},set(_value:unknown){},configurable:true});const copy=Object.assign([] as number[],{item:{id:1}});return copy.item?.id;}\n"})
	valid = append(valid, rule_tester.ValidTestCase{Code: "export {};\nfunction f(){Object.defineProperty(Object.prototype,'item',{get(){return undefined;},set(_value:unknown){},configurable:true});const copy=Object.assign(()=>1,{item:{id:1}});return copy.item?.id;}\n"})
	valid = append(valid, rule_tester.ValidTestCase{Code: "export {};\nfunction f(){Object.defineProperty(Array.prototype,'item',{get(){return undefined;},set(_value:unknown){},configurable:true});const copy=Object.assign([] as number[],{item:{id:1}});return copy.item?.id;}\n"})
	valid = append(valid, rule_tester.ValidTestCase{Code: "export {};\nfunction f(){Object.defineProperty(Function.prototype,'item',{get(){return undefined;},set(_value:unknown){},configurable:true});const copy=Object.assign(()=>1,{item:{id:1}});return copy.item?.id;}\n"})
	valid = append(valid, rule_tester.ValidTestCase{Code: "export {};\nfunction f(){const {...target}={};Object.defineProperty(target,'item',{get(){return undefined;},set(_value:unknown){}});const copy=Object.assign(target,{item:{id:1}});return copy.item?.id;}\n"})
	valid = append(valid, rule_tester.ValidTestCase{Code: "export {};\nfunction f(){const {...target}={};Object.defineProperty(Object.prototype,'item',{get(){return undefined;},set(_value:unknown){},configurable:true});const copy=Object.assign(target,{item:{id:1}});return copy.item?.id;}\n"})
	invalid = append(invalid, rule_tester.InvalidTestCase{Code: "export {};\nfunction f(){const {...target}={};const copy=Object.assign(target,{item:{id:1}});return copy.item?.id;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain", Line: 2, Column: 98, EndLine: 2, EndColumn: 100, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveOptionalChain", Output: "export {};\nfunction f(){const {...target}={};const copy=Object.assign(target,{item:{id:1}});return copy.item.id;}\n"}}}}})
	invalid = append(invalid, rule_tester.InvalidTestCase{Code: "export {};\nfunction f(){const copy=Object.assign({}, {version:1});const keys=Object.keys;keys(copy);keys(copy);/*probe*/ return copy.version ?? 0;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish", Line: 2, Column: 118, EndLine: 2, EndColumn: 130}}})
	valid = append(valid, rule_tester.ValidTestCase{Code: "export {};\nfunction f(){const target={};const {item}=Object.assign(target,(Object.defineProperty(target,'item',{get(){return undefined;},set(_value:unknown){},configurable:true}),{item:{id:1}}));return item?.id;}\n"})
	invalid = append(invalid, rule_tester.InvalidTestCase{Code: "export {};\nfunction f(){const target={};const {item,last=(Object.defineProperty(target,'item',{get(){return undefined;},configurable:true}),0)}=Object.assign(target,{item:{id:1},last:undefined});return item?.id;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain", Line: 2, Column: 196, EndLine: 2, EndColumn: 198, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveOptionalChain", Output: "export {};\nfunction f(){const target={};const {item,last=(Object.defineProperty(target,'item',{get(){return undefined;},configurable:true}),0)}=Object.assign(target,{item:{id:1},last:undefined});return item.id;}\n"}}}}})
	invalid = append(invalid, rule_tester.InvalidTestCase{Code: "export {};\ndeclare function consume(value:unknown):void;function f(){const copy=Object.assign({},{version:1});/*probe*/ consume(copy.version ?? 0);}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish", Line: 2, Column: 118, EndLine: 2, EndColumn: 130}}})
	valid = append(valid, rule_tester.ValidTestCase{Code: "export {};\nfunction f(){const target=[] as number[];const {item}=Object.assign(target,(Object.defineProperty(target,'item',{get(){return undefined;},set(_value:unknown){},configurable:true}),{item:{id:1}}));return item?.id;}\n"})
	invalid = append(invalid, rule_tester.InvalidTestCase{Code: "export {};\nfunction f(){const target=[] as number[];const {item,last=(Object.defineProperty(target,'item',{get(){return undefined;},configurable:true}),0)}=Object.assign(target,{item:{id:1},last:undefined});return item?.id;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain", Line: 2, Column: 208, EndLine: 2, EndColumn: 210, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveOptionalChain", Output: "export {};\nfunction f(){const target=[] as number[];const {item,last=(Object.defineProperty(target,'item',{get(){return undefined;},configurable:true}),0)}=Object.assign(target,{item:{id:1},last:undefined});return item.id;}\n"}}}}})
	invalid = append(invalid, rule_tester.InvalidTestCase{Code: "export {};\ndeclare function consume(value:unknown):void;function f(){const copy=Object.assign([] as number[],{version:1});/*probe*/ consume(copy.version ?? 0);}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish", Line: 2, Column: 130, EndLine: 2, EndColumn: 142}}})
	valid = append(valid, rule_tester.ValidTestCase{Code: "export {};\nfunction f(){const target=()=>1;const {item}=Object.assign(target,(Object.defineProperty(target,'item',{get(){return undefined;},set(_value:unknown){},configurable:true}),{item:{id:1}}));return item?.id;}\n"})
	invalid = append(invalid, rule_tester.InvalidTestCase{Code: "export {};\nfunction f(){const target=()=>1;const {item,last=(Object.defineProperty(target,'item',{get(){return undefined;},configurable:true}),0)}=Object.assign(target,{item:{id:1},last:undefined});return item?.id;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain", Line: 2, Column: 199, EndLine: 2, EndColumn: 201, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveOptionalChain", Output: "export {};\nfunction f(){const target=()=>1;const {item,last=(Object.defineProperty(target,'item',{get(){return undefined;},configurable:true}),0)}=Object.assign(target,{item:{id:1},last:undefined});return item.id;}\n"}}}}})
	invalid = append(invalid, rule_tester.InvalidTestCase{Code: "export {};\ndeclare function consume(value:unknown):void;function f(){const copy=Object.assign(()=>1,{version:1});/*probe*/ consume(copy.version ?? 0);}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish", Line: 2, Column: 121, EndLine: 2, EndColumn: 133}}})
	valid = append(valid, rule_tester.ValidTestCase{Code: "export {};\ntype Item={id:number};function f(){const key=Symbol('item');const target={get [key]():Item|undefined{return undefined;},set [key](value:Item){}};const {[key]:item}=Object.assign(target,{[key]:{id:1}});return item?.id;}\n"})
	valid = append(valid, rule_tester.ValidTestCase{Code: "export {};\ntype Item={id:number};function f(target:Record<string,Item|undefined>,source:Record<string,Item>,key:string){const {[key]:item}=Object.assign(target,source);return item?.id;}\n"})
	valid = append(valid, rule_tester.ValidTestCase{Code: "export {};\ntype Item={id:number};function f(){const target:Item[]=[];Object.defineProperty(target,'0',{get(){return undefined;},set(_v:unknown){}});const [item]=Object.assign(target,{0:{id:1}});return item?.id;}\n"})
	invalid = append(invalid, rule_tester.InvalidTestCase{Code: "export {};\ntype Item={id:number};function f(){const [item]=Object.assign([] as Item[],{0:{id:1}});return item?.id;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain", Line: 2, Column: 99, EndLine: 2, EndColumn: 101, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveOptionalChain", Output: "export {};\ntype Item={id:number};function f(){const [item]=Object.assign([] as Item[],{0:{id:1}});return item.id;}\n"}}}}})
	valid = append(valid, rule_tester.ValidTestCase{Code: "export {};\ntype Item={id:number};type Inner={item:Item|undefined};function f(){const target={get inner():Inner{return {item:undefined};},set inner(value:{item:Item}){}};const {inner:{item}}=Object.assign(target,{inner:{item:{id:1}}});return item?.id;}\n"})
	invalid = append(invalid, rule_tester.InvalidTestCase{Code: "export {};\nfunction f(){const {inner:{item}}=Object.assign({},{inner:{item:{id:1}}});return item?.id;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain", Line: 2, Column: 86, EndLine: 2, EndColumn: 88, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveOptionalChain", Output: "export {};\nfunction f(){const {inner:{item}}=Object.assign({},{inner:{item:{id:1}}});return item.id;}\n"}}}}})
	valid = append(valid, rule_tester.ValidTestCase{Code: "export {};\ntype Item={id:number};declare function make():{get item():Item|undefined;set item(value:Item);};function f(){const {item}=make();return item?.id;}\n"})
	invalid = append(invalid, rule_tester.InvalidTestCase{Code: "export {};\nfunction make(){return {ready:true as const};}function f(){const {ready}=make();/*probe*/ if(ready){}}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "alwaysTruthy", Line: 2, Column: 94, EndLine: 2, EndColumn: 99}}})
	// Destructuring defaults, iterator reads, and runtime mutation order.
	valid = append(valid, rule_tester.ValidTestCase{Code: "export {};\nfunction f(){const copy=Object.assign({},{item:{id:1}});class View{id=copy.item?.id;}Object.defineProperty(copy,'item',{value:undefined});return new View().id;}\n"})
	valid = append(valid, rule_tester.ValidTestCase{Code: "export {};\nfunction f(){const copy=Object.assign({},{item:{id:1}});switch(0 as number){default:return copy.item?.id;case (Object.defineProperty(copy,'item',{value:undefined}),1):return 1;}}\n"})
	valid = append(valid, rule_tester.ValidTestCase{Code: "export {};\nfunction f(){const copy=Object.assign({},{item:{id:1}});class View{static id=copy.item?.id;[(Object.defineProperty(copy,'item',{value:undefined}),'method')]() {}}return View.id;}\n"})
	invalid = append(invalid, rule_tester.InvalidTestCase{Code: "export {};\nfunction f(){const copy=Object.assign({},{item:{id:1}});Object.hasOwn(copy,'item');return copy.item?.id;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain", Line: 2, Column: 100, EndLine: 2, EndColumn: 102, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveOptionalChain", Output: "export {};\nfunction f(){const copy=Object.assign({},{item:{id:1}});Object.hasOwn(copy,'item');return copy.item.id;}\n"}}}}})
	valid = append(valid, rule_tester.ValidTestCase{Code: "export {};\nfunction f(){const copy=Object.assign({},{item:{id:1}});const key={ [Symbol.toPrimitive](){Object.defineProperty(copy,'item',{value:undefined});return 'item';}};Object.hasOwn(copy,key as unknown as string);return copy.item?.id;}\n"})
	valid = append(valid, rule_tester.ValidTestCase{Code: "export {};\nfunction f(){const copy=Object.assign(()=>0,{item:{id:1}});class View{id=copy.item?.id;}Object.defineProperty(copy,'item',{value:undefined});return new View().id;}\n"})
	invalid = append(invalid, rule_tester.InvalidTestCase{Code: "export {};\nfunction f(){const copy=Object.assign(()=>0,{item:{id:1}});const own=Object.hasOwn;own(copy,'item');return copy.item?.id;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain", Line: 2, Column: 117, EndLine: 2, EndColumn: 119, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveOptionalChain", Output: "export {};\nfunction f(){const copy=Object.assign(()=>0,{item:{id:1}});const own=Object.hasOwn;own(copy,'item');return copy.item.id;}\n"}}}}})
	valid = append(valid, rule_tester.ValidTestCase{Code: "export {};\nfunction f(){const copy=Object.assign(()=>0,{item:{id:1}});const key={ [Symbol.toPrimitive](){Object.defineProperty(copy,'item',{value:undefined});return 'item';}};Object.getOwnPropertyDescriptor(copy,key as unknown as string);return copy.item?.id;}\n"})
	valid = append(valid, rule_tester.ValidTestCase{Code: "export {};\nfunction f(){const copy=Object.assign([] as {id:number}[],{item:{id:1}});class View{id=copy.item?.id;}Object.defineProperty(copy,'item',{value:undefined});return new View().id;}\n"})
	valid = append(valid, rule_tester.ValidTestCase{Code: "export {};\nfunction f(){const xs=Object.assign([] as {id:number}[],{0:{id:1}});Object.defineProperty(xs,'length',{value:0});return xs[0]?.id;}\n"})
	valid = append(valid, rule_tester.ValidTestCase{Code: "export {};\nfunction f(){const xs=Object.assign([] as {id:number}[],{0:{id:1}});Object.defineProperty(xs,'length',{value:0});const item=xs['0'];return item?.id;}\n"})
	valid = append(valid, rule_tester.ValidTestCase{Code: "export {};\nfunction f(){const xs=Object.assign([] as {id:number}[],{0:{id:1}});Object.defineProperty(xs,'length',{value:0});const {[(0)]:item}=xs;return item?.id;}\n"})
	valid = append(valid, rule_tester.ValidTestCase{Code: "export {};\nfunction f(){const xs=Object.assign([] as {id:number}[],{0:{id:1}});Object.defineProperty(xs,'length',{value:0});const [item]=xs;return item?.id;}\n"})
	invalid = append(invalid, rule_tester.InvalidTestCase{Code: "export {};\nfunction f(){const xs=Object.assign([] as {id:number}[],{\"4294967295\":{id:1}});Object.defineProperty(xs,'length',{value:0});return xs[\"4294967295\"]?.id;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain", Line: 2, Column: 148, EndLine: 2, EndColumn: 150, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveOptionalChain", Output: "export {};\nfunction f(){const xs=Object.assign([] as {id:number}[],{\"4294967295\":{id:1}});Object.defineProperty(xs,'length',{value:0});return xs[\"4294967295\"].id;}\n"}}}}})
	valid = append(valid, rule_tester.ValidTestCase{Code: "export {};\ntype Item={id:number};function f(){const target={get box():{item:{id:string}|Item}|undefined{return undefined;},set box(v:{item:Item}){}};const fallback:{item:Item|undefined}={item:undefined};function makeFallback(){return fallback;}const {box:{item}=fallback}=Object.assign(target,{box:{item:{id:1}}});return item?.id;}\n"})
	valid = append(valid, rule_tester.ValidTestCase{Code: "export {};\ntype Item={id:number};function f(){const target={get box():{item:{id:string}|Item}|undefined{return undefined;},set box(v:{item:Item}){}};const fallback:{item:Item|undefined}={item:undefined};function makeFallback(){return fallback;}const {box:{item}={item:undefined}}=Object.assign(target,{box:{item:{id:1}}});return item?.id;}\n"})
	valid = append(valid, rule_tester.ValidTestCase{Code: "export {};\ntype Item={id:number};function f(){const target={get box():{item:{id:string}|Item}|undefined{return undefined;},set box(v:{item:Item}){}};const fallback:{item:Item|undefined}={item:undefined};function makeFallback(){return fallback;}const copy=Object.assign(target,{box:{item:{id:1}}});const {box:{item}=makeFallback()}=copy;return item?.id;}\n"})
	valid = append(valid, rule_tester.ValidTestCase{Code: "export {};\nfunction f(){const [item]=Object.assign({get 0():{id:number}|{id:string}{return {id:1};},set 0(v:{id:number}){},*[Symbol.iterator]():IterableIterator<{id:number}|undefined>{yield undefined;}},{0:{id:1}});return item?.id;}\n"})
	valid = append(valid, rule_tester.ValidTestCase{Code: "export {};\nfunction f(){function make(){return Object.assign({get 0():{id:number}|{id:string}{return {id:1};},set 0(v:{id:number}){},*[Symbol.iterator]():IterableIterator<{id:number}|undefined>{yield undefined;}},{0:{id:1}});}const [item]=make();return item?.id;}\n"})
	valid = append(valid, rule_tester.ValidTestCase{Code: "export {};\nfunction f(){const xs=Object.assign([] as {id:number}[],{2147483648:{id:1}});Object.defineProperty(xs,'length',{value:0});return xs['2147483648']?.id;}\n"})
	invalid = append(invalid, rule_tester.InvalidTestCase{Code: "export {};\nfunction f(){const copy=Object.assign({},{item:{id:1}});class View{static id=copy.item?.id;}Object.defineProperty(copy,'item',{value:undefined});return View.id;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain", Line: 2, Column: 87, EndLine: 2, EndColumn: 89, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveOptionalChain", Output: "export {};\nfunction f(){const copy=Object.assign({},{item:{id:1}});class View{static id=copy.item.id;}Object.defineProperty(copy,'item',{value:undefined});return View.id;}\n"}}}}})
	invalid = append(invalid, rule_tester.InvalidTestCase{Code: "export {};\nfunction f(){const copy=Object.assign({},{item:{id:1}});switch(0 as number){default:const value=copy.item?.id;Object.defineProperty(copy,'item',{value:undefined});return value;}}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain", Line: 2, Column: 106, EndLine: 2, EndColumn: 108, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveOptionalChain", Output: "export {};\nfunction f(){const copy=Object.assign({},{item:{id:1}});switch(0 as number){default:const value=copy.item.id;Object.defineProperty(copy,'item',{value:undefined});return value;}}\n"}}}}})
	invalid = append(invalid, rule_tester.InvalidTestCase{Code: "export {};\nfunction f(key:string='item'){const copy=Object.assign([] as {id:number}[],{item:{id:1}});Object.hasOwn(copy,key);return copy.item?.id;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain", Line: 2, Column: 131, EndLine: 2, EndColumn: 133, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveOptionalChain", Output: "export {};\nfunction f(key:string='item'){const copy=Object.assign([] as {id:number}[],{item:{id:1}});Object.hasOwn(copy,key);return copy.item.id;}\n"}}}}})
	valid = append(valid, rule_tester.ValidTestCase{Code: "export {};\ntype Item={id:number};function f(){const target={get box():{item:Item}|undefined{return undefined;},set box(value:{item:Item}){}};target.box={item:{id:1}};const fallback:{item:Item|undefined}={item:undefined};const {box:{item}=fallback}=target;return item?.id;}\n"})
	invalid = append(invalid, rule_tester.InvalidTestCase{Code: "export {};\ntype Item={id:number};function f(){const target={get box():{item:Item}|undefined{return undefined;},set box(value:{item:Item}){}};target.box={item:{id:1}};const fallback:{item:Item|undefined}={item:undefined};const {box:{item}=fallback}=target;if(item===undefined)return;return item?.id;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain", Line: 2, Column: 283, EndLine: 2, EndColumn: 285, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveOptionalChain", Output: "export {};\ntype Item={id:number};function f(){const target={get box():{item:Item}|undefined{return undefined;},set box(value:{item:Item}){}};target.box={item:{id:1}};const fallback:{item:Item|undefined}={item:undefined};const {box:{item}=fallback}=target;if(item===undefined)return;return item.id;}\n"}}}}})
	invalid = append(invalid, rule_tester.InvalidTestCase{Code: "export {};\ntype Item={id:number};function f(){const target={get box():{item:Item}|undefined{return undefined;},set box(value:{item:Item}){}};target.box={item:{id:1}};const fallback:{item:Item|undefined}={item:undefined};const {box:{item}={item:{id:1}}}=target;return item?.id;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain", Line: 2, Column: 261, EndLine: 2, EndColumn: 263, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveOptionalChain", Output: "export {};\ntype Item={id:number};function f(){const target={get box():{item:Item}|undefined{return undefined;},set box(value:{item:Item}){}};target.box={item:{id:1}};const fallback:{item:Item|undefined}={item:undefined};const {box:{item}={item:{id:1}}}=target;return item.id;}\n"}}}}})
	valid = append(valid, rule_tester.ValidTestCase{Code: "export {};\ntype Item={id:number};function f(){const target={get box():{inner:{item:Item}}|undefined{return undefined;},set box(value:{inner:{item:Item}}){}};const fallback:{inner:{item:Item|undefined}|undefined}={inner:undefined};const {box:{inner:{item}={item:undefined}}=fallback}=Object.assign(target,{box:{inner:{item:{id:1}}}});return item?.id;}\n"})
	valid = append(valid, rule_tester.ValidTestCase{Code: "export {};\nfunction f(){const [item]=Object.assign([] as {id:number}[],{0:{id:1},*[Symbol.iterator]():IterableIterator<{id:number}|undefined>{yield undefined;}});return item?.id;}\n"})
	invalid = append(invalid, rule_tester.InvalidTestCase{Code: "export {};\nfunction f(){const [item={id:1}]=Object.assign([] as {id:number}[],{0:{id:1},*[Symbol.iterator]():IterableIterator<{id:number}|undefined>{yield undefined;}});return item?.id;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain", Line: 2, Column: 170, EndLine: 2, EndColumn: 172, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveOptionalChain", Output: "export {};\nfunction f(){const [item={id:1}]=Object.assign([] as {id:number}[],{0:{id:1},*[Symbol.iterator]():IterableIterator<{id:number}|undefined>{yield undefined;}});return item.id;}\n"}}}}})
	invalid = append(invalid, rule_tester.InvalidTestCase{Code: "export {};\nfunction f(){const [item={id:1}]=Object.assign({get 0():{id:number}|{id:string}{return {id:1};},set 0(v:{id:number}){},*[Symbol.iterator]():IterableIterator<{id:number}|undefined>{yield undefined;}},{0:{id:1}});return item?.id;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain", Line: 2, Column: 223, EndLine: 2, EndColumn: 225, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveOptionalChain", Output: "export {};\nfunction f(){const [item={id:1}]=Object.assign({get 0():{id:number}|{id:string}{return {id:1};},set 0(v:{id:number}){},*[Symbol.iterator]():IterableIterator<{id:number}|undefined>{yield undefined;}},{0:{id:1}});return item.id;}\n"}}}}})
	invalid = append(invalid, rule_tester.InvalidTestCase{Code: "export {};\nfunction f(){const target={get box():{item:{id:number}|{id:string}}{return {item:{id:1}};},set box(value:{item:{id:number}}){}};const {box:{item}={item:undefined}}=Object.assign(target,{box:{item:{id:1}}});return item?.id;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain", Line: 2, Column: 218, EndLine: 2, EndColumn: 220, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveOptionalChain", Output: "export {};\nfunction f(){const target={get box():{item:{id:number}|{id:string}}{return {item:{id:1}};},set box(value:{item:{id:number}}){}};const {box:{item}={item:undefined}}=Object.assign(target,{box:{item:{id:1}}});return item.id;}\n"}}}}})
	invalid = append(invalid, rule_tester.InvalidTestCase{Code: "export {};\nfunction f(){const [item]=Object.assign([] as ({id:number}|undefined)[],{0:undefined as {id:number}|undefined,*[Symbol.iterator]():IterableIterator<{id:number}>{yield {id:1};}});return item?.id;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain", Line: 2, Column: 190, EndLine: 2, EndColumn: 192, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveOptionalChain", Output: "export {};\nfunction f(){const [item]=Object.assign([] as ({id:number}|undefined)[],{0:undefined as {id:number}|undefined,*[Symbol.iterator]():IterableIterator<{id:number}>{yield {id:1};}});return item.id;}\n"}}}}})
	valid = append(valid, rule_tester.ValidTestCase{Code: "export {};\nfunction f(){const source:{0:{id:number};[Symbol.iterator]?:()=>IterableIterator<{id:number}|undefined>}={0:{id:1},*[Symbol.iterator](){yield undefined;}};const [item]=Object.assign([] as {id:number}[],source);return item?.id;}\n"})
	valid = append(valid, rule_tester.ValidTestCase{Code: "export {};\nfunction f(native:boolean=false){const custom=Object.assign([] as {id:number}[],{0:{id:1},*[Symbol.iterator]():IterableIterator<{id:number}|undefined>{yield undefined;}});const [item]=native?Object.assign([] as {id:number}[],{0:{id:1}}):custom;return item?.id;}\n"})
	valid = append(valid, rule_tester.ValidTestCase{Code: "export {};\nfunction f(){const copy=Object.assign([] as {id:number}[],{0:{id:1},*[Symbol.iterator]():IterableIterator<{id:number}>{yield {id:1};}});Object.defineProperty(copy,Symbol.iterator,{value:function*(){yield undefined;}});const [item]=copy;return item?.id;}\n"})
	valid = append(valid, rule_tester.ValidTestCase{Code: "export {};\nfunction f(){const copy=Object.assign({},{0:{id:1},*[Symbol.iterator]():IterableIterator<{id:number}>{yield {id:1};}});Object.defineProperty(copy,Symbol.iterator,{value:function*(){yield undefined;}});const [item]=copy;return item?.id;}\n"})
	invalid = append(invalid, rule_tester.InvalidTestCase{Code: "export {};\nfunction f(){const [item]=Object.assign({get 0():unknown{return undefined;},set 0(value:{id:number}){},*[Symbol.iterator]():IterableIterator<{id:number}>{yield {id:1};}},{0:{id:1}});return item?.id;}\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverOptionalChain", Line: 2, Column: 194, EndLine: 2, EndColumn: 196, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveOptionalChain", Output: "export {};\nfunction f(){const [item]=Object.assign({get 0():unknown{return undefined;},set 0(value:{id:number}){},*[Symbol.iterator]():IterableIterator<{id:number}>{yield {id:1};}},{0:{id:1}});return item.id;}\n"}}}}})
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, valid, invalid)
}

func TestNoUnnecessaryConditionOptionalWithoutStrictNullChecks(t *testing.T) {
	options := map[string]any{"allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing": true}
	var valid []rule_tester.ValidTestCase
	var invalid []rule_tester.InvalidTestCase
	for _, access := range []string{"options.title", "options['title']", "options[(key)]"} {
		for _, operator := range []string{"??", "??="} {
			valid = append(valid, rule_tester.ValidTestCase{Code: fmt.Sprintf("function name(options: { title?: string }) { const key = 'title'; return %s %s 'Untitled'; }\n", access, operator), Options: options})
			invalid = append(invalid, rule_tester.InvalidTestCase{Code: fmt.Sprintf("function name(options: { title: string }) { const key = 'title'; return %s %s 'Untitled'; }\n", access, operator), Options: options, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish"}}})
			invalid = append(invalid, rule_tester.InvalidTestCase{Code: fmt.Sprintf("function name(options: { title?: string }) { const key = 'title'; options.title = 'set'; return %s %s 'Untitled'; }\n", access, operator), Options: options, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish"}}})
			valid = append(valid, rule_tester.ValidTestCase{Code: fmt.Sprintf("function name(options: { title?: string }, write: boolean) { const key = 'title'; if (write) options.title = 'set'; return %s %s 'Untitled'; }\n", access, operator), Options: options})
			for _, source := range []string{"title", "source.title", "use ? 'yes' : undefined", "use ? null : 'yes'", "title ?? 'set'"} {
				code := fmt.Sprintf("function name(options: { title?: string }, title?: string, source: { title?: string } = {}, use = false) { const key = 'title'; options.title = %s; return %s %s 'Untitled'; }\n", source, access, operator)
				valid = append(valid, rule_tester.ValidTestCase{Code: code, Options: options})
			}
		}
	}
	valid = append(valid, rule_tester.ValidTestCase{Code: "function name(title?: string) { return title ?? 'Untitled'; }", Options: options})
	for _, declaration := range []string{"const count = 1;", "let count = 1;", "var count = 1;", "const count = null;", "const count = -1;", "const alias = options;", "const value = {};", "const value = [];", "const value = () => 42;", "const value = function() {};"} {
		code := "function name(options: { title?: string }) { options.title = 'set'; " + declaration + " return options['title'] ?? 'Untitled'; }"
		invalid = append(invalid, rule_tester.InvalidTestCase{Code: code, Options: options, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish"}}})
	}
	for _, statement := range []string{"var options: { title?: string } = {};", "const count = (delete options.title, 1);", "const alias = options; alias.title = undefined;", "const count = clear(options);"} {
		code := "declare function clear(options: { title?: string }): number; function name(options: { title?: string }) { options.title = 'set'; " + statement + " return options.title ?? 'Untitled'; }"
		valid = append(valid, rule_tester.ValidTestCase{Code: code, Options: options})
	}
	for _, test := range []struct{ source, declaration string }{
		{"{ get value() { delete options.title; return 1; } }", "const snapshot = { ...source };"},
		{"{ get value() { delete options.title; return 1; } }", "const snapshot = void { ...source };"},
		{"{ get value() { delete options.title; return 1; } }", "const { value } = source;"},
		{"{ *[Symbol.iterator]() { delete options.title; yield 1; } }", "const snapshot = [...source];"},
		{"{ *[Symbol.iterator]() { delete options.title; yield 1; } }", "const [value] = source;"},
		{"{ valueOf() { delete options.title; return 1; } }", "const snapshot = +source;"},
		{"{ get [Symbol.dispose]() { delete options.title; return () => {}; } }", "using snapshot = source;"},
	} {
		for _, operator := range []string{"??", "??="} {
			code := "function name(options: { title?: string }) { const source = " + test.source + "; options.title = 'set'; " + test.declaration + " return options.title " + operator + " 'Untitled'; }"
			valid = append(valid, rule_tester.ValidTestCase{Code: code, Options: options})
		}
	}
	for _, loop := range []string{"for (const value of values)", "for (const value in values)"} {
		for _, access := range []string{"options.title", "title"} {
			for _, operator := range []string{"??", "??="} {
				code := "function name(options: { title?: string }, values: number[], title?: string) { " + loop + " { return " + access + " " + operator + " 'Untitled'; } }"
				valid = append(valid, rule_tester.ValidTestCase{Code: code, Options: options})
			}
		}
	}
	invalid = append(invalid, rule_tester.InvalidTestCase{Code: "function name(title?: string) { title = 'set'; return title ?? 'Untitled'; }", Options: options, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish"}}})
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.unstrict.json", t, &NoUnnecessaryConditionRule, valid, invalid)
}

func TestNoUnnecessaryConditionForInIndex(t *testing.T) {
	var valid []rule_tester.ValidTestCase
	var invalid []rule_tester.InvalidTestCase
	for _, key := range []string{"key", "((key))"} {
		for _, test := range []struct{ operator, message string }{{"&&=", "alwaysFalsy"}, {"||=", "alwaysFalsy"}, {"??=", "alwaysNullish"}} {
			prefix := "type Cells = { [index: number]: null; [name: string]: string | null };\ndeclare const cells: Cells;\ndeclare function update(): null;\n"
			invalid = append(invalid, rule_tester.InvalidTestCase{Code: prefix + fmt.Sprintf("for (const key in [123]) cells[%s] %s update();\n", key, test.operator), Errors: []rule_tester.InvalidTestCaseError{{MessageId: test.message}}})
			valid = append(valid, rule_tester.ValidTestCase{Code: prefix + fmt.Sprintf("for (const key in { name: 123 }) cells[%s] %s update();\n", key, test.operator)})
		}
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, valid, invalid)
}

func TestNoUnnecessaryConditionImportedCopyTargets(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/copied-class-targets.txtar")
	root := fixtures.GetRootDir()
	names, err := archive.FileNames("")
	if err != nil {
		t.Fatal(err)
	}
	files := make(map[string]string, len(names))
	for _, name := range names {
		data, err := archive.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		files[tspath.ResolvePath(root.Dir, name)] = string(data)
	}
	root.FS = utils.NewOverlayVFS(root.FS, files)
	names, err = archive.FileNames("cases")
	if err != nil {
		t.Fatal(err)
	}
	var invalid []rule_tester.InvalidTestCase
	for _, name := range names {
		fileName := tspath.ResolvePath(root.Dir, "cases/"+name)
		invalid = append(invalid, rule_tester.InvalidTestCase{Code: files[fileName], FileName: fileName, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish"}}})
	}
	rule_tester.RunRuleTester(root, "tsconfig.json", t, &NoUnnecessaryConditionRule, nil, invalid)
}
