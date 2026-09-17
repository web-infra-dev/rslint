// Additional AST, scope, and edit-demand coverage checked against Unicorn v75.0.0.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/test/explicit-timer-delay.js
package explicit_timer_delay_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/explicit_timer_delay"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestExplicitTimerDelayExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &explicit_timer_delay.ExplicitTimerDelayRule, []rule_tester.ValidTestCase{
		{Code: "(window?.setTimeout)(callback)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "window?.setTimeout?.(callback)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "function f(setTimeout) { setTimeout(callback); }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "function setInterval() {} setInterval(callback);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "const { setTimeout } = timers; setTimeout(callback);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "import { setTimeout } from \"node:timers\"; setTimeout(callback);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "{ let window = fake; window.setTimeout(callback); }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "{ const globalThis = fake; globalThis.setInterval(callback); }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "function f(self) { self.setTimeout(callback); }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "function f(global) { global.setTimeout(callback); }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "setTimeout.call(null,callback)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "obj.setTimeout(callback)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "window[\"setTimeout\"](callback)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "setInterval(...args)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "setTimeout(callback, 0n)", FileName: "case.js", Options: []any{"never"}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "setTimeout(callback, false)", FileName: "case.js", Options: []any{"never"}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "setTimeout(callback, \"0\")", FileName: "case.js", Options: []any{"never"}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "setTimeout(callback, undefined)", FileName: "case.js", Options: []any{"never"}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "setTimeout(callback, null)", FileName: "case.js", Options: []any{"never"}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "setTimeout(callback, -false)", FileName: "case.js", Options: []any{"never"}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "setTimeout(callback, 0 * 1)", FileName: "case.js", Options: []any{"never"}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "setTimeout(callback, 1 - 1)", FileName: "case.js", Options: []any{"never"}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "setTimeout(callback, !0)", FileName: "case.js", Options: []any{"never"}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "setTimeout(callback, --x)", FileName: "case.js", Options: []any{"never"}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "setTimeout(callback,0,argument)", FileName: "case.js", Options: []any{"never"}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "setTimeout?.(callback,0)", FileName: "case.js", Options: []any{"never"}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "setTimeout(callback);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "off", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "(setTimeout as Function)(callback)", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "namespace setTimeout {} setTimeout(callback)", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "interface setTimeout {} setTimeout(callback)", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "script"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "setTimeout(callback);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "off", "setInterval": "off", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "window.setTimeout(callback);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "off", "globalThis": "readonly", "global": "readonly", "self": "readonly"}},
	}, []rule_tester.InvalidTestCase{
		{Code: "setTimeout(callback,)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"setTimeout(callback, 0,)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missing-delay", Message: "`setTimeout` should have an explicit delay argument.", Line: 1, Column: 1, EndLine: 1, EndColumn: 22, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "setTimeout((callback /* keep */), /* after */)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"setTimeout((callback /* keep */), 0, /* after */)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missing-delay", Message: "`setTimeout` should have an explicit delay argument.", Line: 1, Column: 1, EndLine: 1, EndColumn: 47, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "globalThis?.setTimeout(callback)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"globalThis?.setTimeout(callback, 0)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missing-delay", Message: "`setTimeout` should have an explicit delay argument.", Line: 1, Column: 1, EndLine: 1, EndColumn: 33, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "(setTimeout)(callback)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"(setTimeout)(callback, 0)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missing-delay", Message: "`setTimeout` should have an explicit delay argument.", Line: 1, Column: 1, EndLine: 1, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "(window.setTimeout)(callback)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"(window.setTimeout)(callback, 0)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missing-delay", Message: "`setTimeout` should have an explicit delay argument.", Line: 1, Column: 1, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "\"😀\"; setTimeout(callback);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"\"😀\"; setTimeout(callback, 0);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missing-delay", Message: "`setTimeout` should have an explicit delay argument.", Line: 1, Column: 7, EndLine: 1, EndColumn: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "setTimeout(\n  callback\n);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"setTimeout(\n  callback, 0\n);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missing-delay", Message: "`setTimeout` should have an explicit delay argument.", Line: 1, Column: 1, EndLine: 3, EndColumn: 2, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "/** @type {Function} */ (setTimeout)(callback)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"/** @type {Function} */ (setTimeout)(callback, 0)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missing-delay", Message: "`setTimeout` should have an explicit delay argument.", Line: 1, Column: 25, EndLine: 1, EndColumn: 47, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "setInterval(callback)", FileName: "case.js", Options: []any{"always"}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"setInterval(callback, 0)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missing-delay", Message: "`setInterval` should have an explicit delay argument.", Line: 1, Column: 1, EndLine: 1, EndColumn: 22, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "setTimeout(callback, 0,)", FileName: "case.js", Options: []any{"never"}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"setTimeout(callback,)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "redundant-delay", Message: "`setTimeout` should not have an explicit delay of `0`.", Line: 1, Column: 22, EndLine: 1, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "setTimeout((callback), /* delay */ (0),)", FileName: "case.js", Options: []any{"never"}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"setTimeout((callback),)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "redundant-delay", Message: "`setTimeout` should not have an explicit delay of `0`.", Line: 1, Column: 37, EndLine: 1, EndColumn: 38, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "setTimeout(callback, +(-(+0)))", FileName: "case.js", Options: []any{"never"}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"setTimeout(callback)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "redundant-delay", Message: "`setTimeout` should not have an explicit delay of `0`.", Line: 1, Column: 22, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "setTimeout(callback, 0x0)", FileName: "case.js", Options: []any{"never"}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"setTimeout(callback)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "redundant-delay", Message: "`setTimeout` should not have an explicit delay of `0`.", Line: 1, Column: 22, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "setTimeout(callback, 0b0)", FileName: "case.js", Options: []any{"never"}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"setTimeout(callback)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "redundant-delay", Message: "`setTimeout` should not have an explicit delay of `0`.", Line: 1, Column: 22, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "setTimeout(callback, 0o0)", FileName: "case.js", Options: []any{"never"}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"setTimeout(callback)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "redundant-delay", Message: "`setTimeout` should not have an explicit delay of `0`.", Line: 1, Column: 22, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "setTimeout(callback, 0.000)", FileName: "case.js", Options: []any{"never"}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"setTimeout(callback)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "redundant-delay", Message: "`setTimeout` should not have an explicit delay of `0`.", Line: 1, Column: 22, EndLine: 1, EndColumn: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "setTimeout(callback, 0e3)", FileName: "case.js", Options: []any{"never"}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"setTimeout(callback)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "redundant-delay", Message: "`setTimeout` should not have an explicit delay of `0`.", Line: 1, Column: 22, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "setTimeout(...args,0)", FileName: "case.js", Options: []any{"never"}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"setTimeout(...args)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "redundant-delay", Message: "`setTimeout` should not have an explicit delay of `0`.", Line: 1, Column: 20, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "window?.setTimeout(callback,0)", FileName: "case.js", Options: []any{"never"}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"window?.setTimeout(callback)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "redundant-delay", Message: "`setTimeout` should not have an explicit delay of `0`.", Line: 1, Column: 29, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "\"😀\"; setTimeout(callback, -0);", FileName: "case.js", Options: []any{"never"}, LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"\"😀\"; setTimeout(callback);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "redundant-delay", Message: "`setTimeout` should not have an explicit delay of `0`.", Line: 1, Column: 28, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "setTimeout<string>(callback)", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"setTimeout<string>(callback, 0)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missing-delay", Message: "`setTimeout` should have an explicit delay argument.", Line: 1, Column: 1, EndLine: 1, EndColumn: 29, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "interface setTimeout {} setTimeout(callback)", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "readonly", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"interface setTimeout {} setTimeout(callback, 0)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missing-delay", Message: "`setTimeout` should have an explicit delay argument.", Line: 1, Column: 25, EndLine: 1, EndColumn: 45, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "setTimeout(callback);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"setTimeout": "writable", "setInterval": "readonly", "window": "readonly", "globalThis": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"setTimeout(callback, 0);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missing-delay", Message: "`setTimeout` should have an explicit delay argument.", Line: 1, Column: 1, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
	})
}

func TestExplicitTimerDelayArtifactsFollowDemand(t *testing.T) {
	for _, testCase := range []struct {
		source, output string
		options        []any
		suggestions    []struct{ messageID, description, output string }
	}{
		{source: "globalThis.setInterval(callback);", output: "globalThis.setInterval(callback, 0);", options: []any{}},
		{source: "globalThis.setInterval(callback, 0);", output: "globalThis.setInterval(callback);", options: []any{"never"}},
	} {
		t.Run(testCase.source, func(t *testing.T) {
			helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
			program, sourceFile, err := helper.CreateTestProgram(testCase.source, "edit-demand.js", "tsconfig.json")
			if err != nil {
				t.Fatal(err)
			}
			diagnostics := make(map[rule.EditDemand]rule.RuleDiagnostic, 4)
			for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
				var got []rule.RuleDiagnostic
				linter.LintSingleFile(linter.LintSingleFileOptions{
					Program: lintprogram.NewFromCompiler(program), File: sourceFile.FileName(),
					GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
						return []rule.ConfiguredRule{{Name: explicit_timer_delay.ExplicitTimerDelayRule.Name, Severity: rule.SeverityError,
							Run: func(ctx rule.RuleContext) rule.RuleListeners {
								return explicit_timer_delay.ExplicitTimerDelayRule.Run(ctx, testCase.options)
							},
						}}
					},
					Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) { got = append(got, d) }},
				})
				if len(got) != 1 {
					t.Fatalf("demand %d: got %d diagnostics", demand, len(got))
				}
				diagnostics[demand] = got[0]
			}
			baseline := diagnostics[rule.EditDemandNone]
			all := diagnostics[rule.EditDemandAll]
			for demand, diagnostic := range diagnostics {
				if diagnostic.Range != baseline.Range || !reflect.DeepEqual(diagnostic.Message, baseline.Message) || diagnostic.Severity != baseline.Severity {
					t.Errorf("demand %d changed diagnostic identity", demand)
				}
				wantFix := testCase.output != testCase.source && (demand == rule.EditDemandAutofix || demand == rule.EditDemandAll)
				if (diagnostic.FixesPtr != nil) != wantFix {
					t.Errorf("demand %d: unexpected autofix artifacts", demand)
				}
				if wantFix && !reflect.DeepEqual(diagnostic.FixesPtr, all.FixesPtr) {
					t.Errorf("demand %d changed autofix artifacts", demand)
				}
				expectedOutput := testCase.source
				if wantFix {
					expectedOutput = testCase.output
				}
				output, _, fixed := linter.ApplyRuleFixes(testCase.source, []rule.RuleDiagnostic{diagnostic})
				if output != expectedOutput || fixed != wantFix {
					t.Errorf("demand %d: unexpected autofix %q", demand, output)
				}
				wantSuggestions := len(testCase.suggestions) > 0 && (demand == rule.EditDemandSuggestion || demand == rule.EditDemandAll)
				if !wantSuggestions {
					if diagnostic.Suggestions != nil {
						t.Errorf("demand %d produced suggestions without demand", demand)
					}
					continue
				}
				if diagnostic.Suggestions == nil || len(*diagnostic.Suggestions) != len(testCase.suggestions) || !reflect.DeepEqual(diagnostic.Suggestions, all.Suggestions) {
					t.Fatalf("demand %d: inconsistent suggestions", demand)
				}
				for index, expected := range testCase.suggestions {
					suggestion := (*diagnostic.Suggestions)[index]
					if suggestion.Message.Id != expected.messageID || suggestion.Message.Description != expected.description {
						t.Errorf("demand %d: wrong suggestion message", demand)
					}
					output, _, fixed := linter.ApplyRuleFixes(testCase.source, (*diagnostic.Suggestions)[index:index+1])
					if !fixed || output != expected.output {
						t.Errorf("demand %d: unexpected suggestion %q", demand, output)
					}
				}
			}
		})
	}
}
