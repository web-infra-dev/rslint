package no_unnecessary_condition

import (
	"fmt"
	"slices"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
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
	)
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, valid, nil)
}

func TestNoUnnecessaryConditionOptionalWithoutStrictNullChecks(t *testing.T) {
	options := map[string]any{"allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing": true}
	var valid []rule_tester.ValidTestCase
	var invalid []rule_tester.InvalidTestCase
	for _, access := range []string{"options.title", "options['title']", "options[(key)]"} {
		for _, operator := range []string{"??", "??="} {
			valid = append(valid, rule_tester.ValidTestCase{Code: fmt.Sprintf("function name(options: { title?: string }) { const key = 'title'; return %s %s 'Untitled'; }\n", access, operator), Options: options})
			invalid = append(invalid, rule_tester.InvalidTestCase{Code: fmt.Sprintf("function name(options: { title: string }) { const key = 'title'; return %s %s 'Untitled'; }\n", access, operator), Options: options, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverNullish"}}})
		}
	}
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
