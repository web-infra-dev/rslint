package no_meaningless_void_operator

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoMeaninglessVoidOperatorUnionTypes(t *testing.T) {
	for _, checkNever := range []bool{false, true} {
		name := "default"
		if checkNever {
			name = "checkNever"
		}
		t.Run(name, func(t *testing.T) {
			options := map[string]any{"checkNever": checkNever}
			valid := []rule_tester.ValidTestCase{
				{Code: "declare const audio: { pause(): number } | undefined; void audio?.pause();"},
				{Code: "declare const call: (() => Promise<void>) | undefined; void call?.();"},
				{Code: "declare function call(): void | number; void call();"},
				{Code: "declare function call(): void | undefined | number; void call();"},
				{Code: "declare function call(): void | null; void call();"},
				{Code: "declare function call(): void | unknown; void call();"},
				{Code: "declare function call(): void | any; void call();"},
				{Code: "declare function call(): void & { brand: true }; void call();"},
				{Code: "function test<T extends void | undefined>(call: () => T) { void call(); }"},
				{Code: "function test<T extends never>(call: () => T) { void call(); }"},
			}
			for i := range valid {
				valid[i].Options = options
			}
			invalid := []rule_tester.InvalidTestCase{
				{
					Code:   "declare const audio: { pause(): void } | undefined;\nvoid audio?.pause();",
					Output: []string{"declare const audio: { pause(): void } | undefined;\naudio?.pause();"},
					Errors: []rule_tester.InvalidTestCaseError{{
						MessageId: "meaninglessVoidOperator",
						Message:   "void operator shouldn't be used on void | undefined; it should convey that a return value is being ignored",
						Line:      2, Column: 1, EndLine: 2, EndColumn: 20,
					}},
				},
				{
					Code:   "declare const call: (() => void) | undefined;\nvoid call?.();",
					Output: []string{"declare const call: (() => void) | undefined;\ncall?.();"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "meaninglessVoidOperator", Line: 2, Column: 1}},
				},
				{
					Code:   "declare const audio: { pause?: () => void } | null;\nvoid audio?.pause?.();",
					Output: []string{"declare const audio: { pause?: () => void } | null;\naudio?.pause?.();"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "meaninglessVoidOperator"}},
				},
				{
					Code:   "declare function call(): void | undefined;\nvoid call();",
					Output: []string{"declare function call(): void | undefined;\ncall();"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "meaninglessVoidOperator"}},
				},
				{
					Code:   "type Result = void | undefined; declare function call(): Result;\nvoid call();",
					Output: []string{"type Result = void | undefined; declare function call(): Result;\ncall();"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "meaninglessVoidOperator"}},
				},
				{
					Code:   "declare function call(): void | undefined | never;\nvoid call();",
					Output: []string{"declare function call(): void | undefined | never;\ncall();"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "meaninglessVoidOperator"}},
				},
				{
					Code:   "declare const call: (() => never) | undefined;\nvoid call?.();",
					Output: []string{"declare const call: (() => never) | undefined;\ncall?.();"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "meaninglessVoidOperator"}},
				},
			}
			for i := range invalid {
				invalid[i].Options = options
			}
			rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoMeaninglessVoidOperatorRule, valid, invalid)
		})
	}
}

func TestNoMeaninglessVoidOperatorFixes(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoMeaninglessVoidOperatorRule, []rule_tester.ValidTestCase{
		{Code: "declare function fail(): never; void fail();"},
	}, []rule_tester.InvalidTestCase{
		{
			Code:   "declare function call(): void | undefined;\nvoid /* remove */ (/* keep */ call());",
			Output: []string{"declare function call(): void | undefined;\n(/* keep */ call());"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "meaninglessVoidOperator"}},
		},
		{
			Code:   "declare function call(): void | undefined;\nvoid // remove\n\tcall();",
			Output: []string{"declare function call(): void | undefined;\ncall();"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "meaninglessVoidOperator"}},
		},
		{
			Code:   "declare function call(): void | undefined;\nvoid\u00a0call();",
			Output: []string{"declare function call(): void | undefined;\ncall();"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "meaninglessVoidOperator"}},
		},
		{
			Code:   "declare function call(): void | undefined;\nconst result = () => void (call());",
			Output: []string{"declare function call(): void | undefined;\nconst result = () => (call());"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "meaninglessVoidOperator"}},
		},
		{
			Code:    "declare function fail(): never;\nvoid /* remove */ (fail());",
			Options: map[string]any{"checkNever": true},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "meaninglessVoidOperator",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "removeVoid",
					Output:    "declare function fail(): never;\n(fail());",
				}},
			}},
		},
	})
}
