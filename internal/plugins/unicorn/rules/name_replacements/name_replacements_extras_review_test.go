package name_replacements_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/name_replacements"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNameReplacementsReviewRegressions(t *testing.T) {
	localeReplacement := "äth" + "er"
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&name_replacements.NameReplacementsRule,
		[]rule_tester.ValidTestCase{
			{Code: "const error = 1; function render() { return <err /> }", Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   "function f(err /* keep */ ?: string) { return err; }",
				Output: []string{"function f(error /* keep */ ?: string) { return error; }"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "replace",
					Message:   "The variable `err` should be named `error`. A more descriptive name will do too.",
					Line:      1, Column: 12, EndLine: 1, EndColumn: 36,
				}},
			},
			{
				Code:   "let err: string;",
				Output: []string{"let error: string;"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "replace", Line: 1, Column: 5, EndLine: 1, EndColumn: 16,
				}},
			},
			{
				Code:   "function f(err?: string) { return err; }",
				Output: []string{"function f(error?: string) { return error; }"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "replace", Line: 1, Column: 12, EndLine: 1, EndColumn: 24,
				}},
			},
			{
				Code:   "function f(...err: string[]) { return err; }",
				Output: []string{"function f(...error: string[]) { return error; }"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "replace", Line: 1, Column: 15, EndLine: 1, EndColumn: 18,
				}},
			},
			{
				Code:   "const outer = 1; function render() { const err = 2; use(err); return <err /> } use(outer);",
				Tsx:    true,
				Output: []string{"const outer = 1; function render() { const error = 2; use(error); return <err /> } use(outer);"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "replace", Line: 1, Column: 44, EndLine: 1, EndColumn: 47,
				}},
			},
			{
				Code:   "const Btn = 1; function render() { const Btn = component; return <Btn /> } use(Btn);",
				Tsx:    true,
				Output: []string{"const Button = 1; function render() { const Btn = component; return <Btn /> } use(Button);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replace", Line: 1, Column: 7, EndLine: 1, EndColumn: 10},
					{MessageId: "replace", Line: 1, Column: 42, EndLine: 1, EndColumn: 45},
				},
			},
			{
				Code:   "export default function err() { return err; }",
				Output: []string{"export default function error() { return error; }"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "replace", Line: 1, Column: 25, EndLine: 1, EndColumn: 28,
				}},
			},
			{
				Code:   "export default class Btn { value: Btn | undefined }",
				Output: []string{"export default class Button { value: Button | undefined }"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "replace", Line: 1, Column: 22, EndLine: 1, EndColumn: 25,
				}},
			},
			{
				Code: "export function err() { return err; }",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "replace", Line: 1, Column: 17, EndLine: 1, EndColumn: 20,
				}},
			},
			{
				Code:   "/** @parameter ctx */\nfunction f(ctx) { return ctx; }",
				Output: []string{"/** @parameter ctx */\nfunction f(context) { return context; }"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "replace", Line: 2, Column: 12, EndLine: 2, EndColumn: 15,
				}},
			},
			{
				Code:   "/** @param ctx */\n/* nearer */\nfunction f(ctx) { return ctx; }",
				Output: []string{"/** @param ctx */\n/* nearer */\nfunction f(context) { return context; }"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "replace", Line: 3, Column: 12, EndLine: 3, EndColumn: 15,
				}},
			},
			{
				Code: "/** @param ctx */\nfunction f(ctx) { return ctx; }",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "replace", Line: 2, Column: 12, EndLine: 2, EndColumn: 15,
				}},
			},
			{
				Code:   "/** @param ctx */\n\nfunction f(ctx) { return ctx; }",
				Output: []string{"/** @param ctx */\n\nfunction f(context) { return context; }"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "replace", Line: 3, Column: 12, EndLine: 3, EndColumn: 15,
				}},
			},
			{
				Code:   "// @param ctx\nfunction f(ctx) { return ctx; }",
				Output: []string{"// @param ctx\nfunction f(context) { return context; }"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "replace", Line: 2, Column: 12, EndLine: 2, EndColumn: 15,
				}},
			},
			{
				Code:   "/** @param_value ctx */\nfunction f(ctx) { return ctx; }",
				Output: []string{"/** @param_value ctx */\nfunction f(context) { return context; }"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "replace", Line: 2, Column: 12, EndLine: 2, EndColumn: 15,
				}},
			},
			{
				Code:    "object.foo = 1;",
				Options: map[string]any{"checkProperties": true, "extendDefaultReplacements": false, "replacements": map[string]any{"foo": map[string]any{"class": true, "default": true}}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "suggestion",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "rename", Output: "object.class = 1;"},
						{MessageId: "rename", Output: "object.default = 1;"},
					},
				}},
			},
			{
				Code:    "let x = 1;",
				Options: map[string]any{"extendDefaultReplacements": false, "replacements": map[string]any{"x": map[string]any{"alpha": true, localeReplacement: true, "zeta": true}}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "suggestion",
					Message:   "Please rename the variable `x`. Suggested names are: `alpha`, `" + localeReplacement + "`, `zeta`. A more descriptive name will do too.",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "rename", Output: "let alpha = 1;"},
						{MessageId: "rename", Output: "let " + localeReplacement + " = 1;"},
						{MessageId: "rename", Output: "let zeta = 1;"},
					},
				}},
			},
			{
				Code:    "const ǅoo = 1;",
				Options: map[string]any{"extendDefaultReplacements": false, "replacements": map[string]any{"ǆoo": map[string]any{"bar": true}}},
				Output:  []string{"const bar = 1;"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "replace"}},
			},
			{
				Code:    "const Ǆoo = 1;",
				Options: map[string]any{"extendDefaultReplacements": false, "replacements": map[string]any{"ǆoo": map[string]any{"bar": true}}},
				Output:  []string{"const Bar = 1;"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "replace"}},
			},
			{
				Code:    "const 𐐨oo = 1;",
				Options: map[string]any{"extendDefaultReplacements": false, "replacements": map[string]any{"𐐨oo": map[string]any{"bar": true}}},
				Output:  []string{"const Bar = 1;"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "replace"}},
			},
			{
				Code:    "export const {err} = source;",
				Options: map[string]any{"checkShorthandProperties": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "replace"}},
			},
			{
				Code:   "{ let errCb; use(errCb, errorCb); } let errorCb; use(errorCb);",
				Output: []string{"{ let errorCallback; use(errorCallback, errorCallback_); } let errorCallback_; use(errorCallback_);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "replace"},
					{MessageId: "replace"},
				},
			},
		},
	)
}
