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
	t.Run("computed-chain", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, nil, []rule_tester.InvalidTestCase{{Code: `declare const obj: {item:{id:string}, method:()=>string, get:()=>{id:string}} | undefined;
declare const key: 'item';
obj?.[(key)]?.id;
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "neverOptionalChain", Line: 3, Column: 13, EndLine: 3, EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestRemoveOptionalChain", Output: `declare const obj: {item:{id:string}, method:()=>string, get:()=>{id:string}} | undefined;
declare const key: 'item';
obj?.[(key)].id;
`},
					},
				},
			}}})
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
	t.Run("checked-computed-chain", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.noUncheckedIndexedAccess.json", t, &NoUnnecessaryConditionRule, nil, []rule_tester.InvalidTestCase{{Code: `declare const obj: {item:{id:string}, method:()=>string, get:()=>{id:string}} | undefined;
declare const key: 'item';
obj?.[(key)]?.id;
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "neverOptionalChain", Line: 3, Column: 13, EndLine: 3, EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestRemoveOptionalChain", Output: `declare const obj: {item:{id:string}, method:()=>string, get:()=>{id:string}} | undefined;
declare const key: 'item';
obj?.[(key)].id;
`},
					},
				},
			}}})
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
