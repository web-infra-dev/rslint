// Additional AST and scope regressions verified against Unicorn v75.0.0.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/test/prefer-global-number-constants.js
package prefer_global_number_constants_test

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_global_number_constants"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_number_properties"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"reflect"
	"testing"
)

func TestPreferGlobalNumberConstantsExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_global_number_constants.PreferGlobalNumberConstantsRule, []rule_tester.ValidTestCase{
		{Code: "const {NaN} = Number;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "const key = \"NaN\"; const x = Number[key];", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "const x = Number[variable];", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "function f(Number) { return Number.NaN; }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "function f(NaN) { return Number.NaN; }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "function f(Infinity) { return Number.NEGATIVE_INFINITY; }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "Number.NaN++;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "Number.POSITIVE_INFINITY = value;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "({value: Number.NaN} = object);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "const x = Number.NaN;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly", "Number": "off"}},
		{Code: "interface NaN {} const x = Number.NaN;", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "delete Number.NaN;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "delete (Number.NaN);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "for (Number.NaN of values) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
	}, []rule_tester.InvalidTestCase{
		{Code: "const N = Number; const x = N.NaN;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"const N = Number; const x = NaN;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-number-constants", Message: "Prefer `NaN` over `Number.NaN`.", Line: 1, Column: 29, EndLine: 1, EndColumn: 34, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const g = globalThis; const x = g.Number.POSITIVE_INFINITY;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"const g = globalThis; const x = Infinity;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-number-constants", Message: "Prefer `Infinity` over `Number.POSITIVE_INFINITY`.", Line: 1, Column: 33, EndLine: 1, EndColumn: 59, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const {Number: N} = globalThis; const x = N.POSITIVE_INFINITY;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"const {Number: N} = globalThis; const x = Infinity;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-number-constants", Message: "Prefer `Infinity` over `Number.POSITIVE_INFINITY`.", Line: 1, Column: 43, EndLine: 1, EndColumn: 62, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const {NaN: value} = Number;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-number-constants", Message: "Prefer `NaN` over `Number.NaN`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const x = Number?.NaN;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"const x = NaN;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-number-constants", Message: "Prefer `NaN` over `Number.NaN`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 22, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const x = Number[\"POSITIVE_INFINITY\"];", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"const x = Infinity;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-number-constants", Message: "Prefer `Infinity` over `Number.POSITIVE_INFINITY`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 38, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const x = Number[`NaN`];", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"const x = NaN;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-number-constants", Message: "Prefer `NaN` over `Number.NaN`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const x = Number[\"N\"+\"aN\"];", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"const x = NaN;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-number-constants", Message: "Prefer `NaN` over `Number.NaN`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const x = Number /* keep */ . POSITIVE_INFINITY;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-number-constants", Message: "Prefer `Infinity` over `Number.POSITIVE_INFINITY`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 48, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const x = Number[/* keep */ \"NaN\"];", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-number-constants", Message: "Prefer `NaN` over `Number.NaN`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 35, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const x = -Number.NEGATIVE_INFINITY;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-number-constants", Message: "Prefer `-Infinity` over `Number.NEGATIVE_INFINITY`.", Line: 1, Column: 12, EndLine: 1, EndColumn: 36, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const x = Number.NEGATIVE_INFINITY ** 2;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-number-constants", Message: "Prefer `-Infinity` over `Number.NEGATIVE_INFINITY`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 35, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const x = globalThis[\"Number\"].NaN;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"const x = NaN;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-number-constants", Message: "Prefer `NaN` over `Number.NaN`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 35, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const x = self.Number.NaN;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"const x = NaN;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-number-constants", Message: "Prefer `NaN` over `Number.NaN`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "\"😀\"; const x = Number.NaN;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"\"😀\"; const x = NaN;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-number-constants", Message: "Prefer `NaN` over `Number.NaN`.", Line: 1, Column: 17, EndLine: 1, EndColumn: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "interface Number {} const x = Number.NaN;", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"interface Number {} const x = NaN;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-number-constants", Message: "Prefer `NaN` over `Number.NaN`.", Line: 1, Column: 31, EndLine: 1, EndColumn: 41, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const x = (Number as NumberConstructor).NaN;", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"const x = NaN;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-number-constants", Message: "Prefer `NaN` over `Number.NaN`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 44, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const {POSITIVE_INFINITY: value} = Number;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-number-constants", Message: "Prefer `Infinity` over `Number.POSITIVE_INFINITY`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 32, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const {NaN: value = 1} = Number;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-number-constants", Message: "Prefer `NaN` over `Number.NaN`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 22, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const value = Number.NaN; function f(){const NaN=1;}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"const value = NaN; function f(){const NaN=1;}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-number-constants", Message: "Prefer `NaN` over `Number.NaN`.", Line: 1, Column: 15, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const value = Number.POSITIVE_INFINITY; function f(){const Infinity=1;}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"const value = Infinity; function f(){const Infinity=1;}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-number-constants", Message: "Prefer `Infinity` over `Number.POSITIVE_INFINITY`.", Line: 1, Column: 15, EndLine: 1, EndColumn: 39, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
	})
}

func TestPreferGlobalNumberConstantsStableReferences(t *testing.T) {
	var valid []rule_tester.ValidTestCase
	for _, code := range []string{
		"let N = Number; N = {NaN: 42}; const x = N.NaN;",
		"let N = Number; const x = N.NaN; N = other;",
		"let N = Number; N ||= other; const x = N.NaN;",
		"let N = Number; ({N} = other); const x = N.NaN;",
		"let N = Number; for (N of values) {} const x = N.NaN;",
		"let N = Number; function change() { N = other; } const x = N.NaN;",
		"const x = (false ? Number : {NaN: 42}).NaN;",
		"const x = (flag && Number).NaN;",
		"const x = (other ?? Number).NaN;",
		"const N = flag ? Number : other; const x = N.NaN;",
		"function f(N = Number) { return N.NaN; }",
		"const {Number: N = other} = globalThis; const x = N.NaN;",
		"const {N = Number} = object; const x = N.NaN;",
		"let N; N = Number; const x = N.NaN;",
		"const x = (N = Number).NaN;",
		"const x = (sideEffect(), Number).NaN;",
		"const N = (sideEffect(), Number); const x = N.NaN;",
		"const x = N.NaN; var N = Number;",
		"var N = Number; var N = other; const x = N.NaN;",
		"const N = flag ? Number : Number; const x = N.NaN;",
		"const N = flag ? Number : globalThis.Number; const x = N.NaN;",
		"let N; N = Number; N = Number; const x = N.NaN;",
		"let {Number: N} = globalThis; N = other; const x = N.NaN;",
	} {
		valid = append(valid, rule_tester.ValidTestCase{Code: code, FileName: "case.js"})
	}
	var invalid []rule_tester.InvalidTestCase
	for _, test := range []struct{ code, output string }{
		{"const N = Number; N.NaN;", "const N = Number; NaN;"},
		{"let N = Number; N.NaN;", "let N = Number; NaN;"},
		{"var N = Number; N.NaN;", "var N = Number; NaN;"},
		{"const N = Number; const M = N; M.NaN;", "const N = Number; const M = N; NaN;"},
		{"const {Number: N} = globalThis; N.NaN;", "const {Number: N} = globalThis; NaN;"},
		{"const g = globalThis; const {Number: N} = g; N.NaN;", "const g = globalThis; const {Number: N} = g; NaN;"},
		{"const N = Number; function f(N) { N = other; } N.NaN;", "const N = Number; function f(N) { N = other; } NaN;"},
	} {
		invalid = append(invalid, rule_tester.InvalidTestCase{
			Code: test.code, FileName: "case.js", Output: []string{test.output},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-global-number-constants", Message: "Prefer `NaN` over `Number.NaN`."}},
		})
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_global_number_constants.PreferGlobalNumberConstantsRule, valid, invalid)
}

func TestPreferGlobalNumberConstantsArtifactsFollowDemand(t *testing.T) {
	for _, testCase := range []struct{ source, output string }{{source: "const foo = Number.NaN;", output: "const foo = NaN;"},
		{source: "const foo = Number.NEGATIVE_INFINITY;", output: "const foo = Number.NEGATIVE_INFINITY;"}} {
		t.Run(testCase.source, func(t *testing.T) {
			program, sourceFile, err := rule_tester.NewProgramHelper(fixtures.GetRootDir()).CreateTestProgram(testCase.source, "edit-demand.js", "tsconfig.json")
			if err != nil {
				t.Fatal(err)
			}
			demands := []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll}
			diagnostics := make([]rule.RuleDiagnostic, len(demands))
			for index, demand := range demands {
				var found []rule.RuleDiagnostic
				linter.LintSingleFile(linter.LintSingleFileOptions{
					Program: lintprogram.NewFromCompiler(program), File: sourceFile.FileName(),
					GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
						return []rule.ConfiguredRule{{Name: prefer_global_number_constants.PreferGlobalNumberConstantsRule.Name, Severity: rule.SeverityError, Run: func(ctx rule.RuleContext) rule.RuleListeners {
							return prefer_global_number_constants.PreferGlobalNumberConstantsRule.Run(ctx, nil)
						}}}
					},
					Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) { found = append(found, d) }},
				})
				if len(found) != 1 {
					t.Fatalf("demand %d: expected one diagnostic, got %d", demand, len(found))
				}
				diagnostics[index] = found[0]
			}
			all := diagnostics[len(diagnostics)-1]
			for index, demand := range demands {
				diagnostic := diagnostics[index]
				if diagnostic.Range != all.Range || !reflect.DeepEqual(diagnostic.Message, all.Message) {
					t.Errorf("demand %d changed diagnostic identity", demand)
				}
				wantFix := testCase.source != testCase.output && (demand == rule.EditDemandAutofix || demand == rule.EditDemandAll)
				if (diagnostic.FixesPtr != nil) != wantFix {
					t.Errorf("demand %d: unexpected fix artifacts", demand)
				}
				if wantFix && !reflect.DeepEqual(diagnostic.FixesPtr, all.FixesPtr) {
					t.Errorf("demand %d changed fix artifacts", demand)
				}
				if diagnostic.Suggestions != nil {
					t.Errorf("demand %d: autofix-only rule produced suggestions", demand)
				}
			}
			// Applying fixes can sort slices in place; identity checks are finished.
			for index, demand := range demands {
				wantFix := testCase.source != testCase.output && (demand == rule.EditDemandAutofix || demand == rule.EditDemandAll)
				expected := testCase.source
				if wantFix {
					expected = testCase.output
				}
				output, _, fixed := linter.ApplyRuleFixes(testCase.source, diagnostics[index:index+1])
				if output != expected || fixed != wantFix {
					t.Errorf("demand %d: got %q, want %q", demand, output, expected)
				}
			}
		})
	}
}

// The preset must retain parseInt/parseFloat checks without reversing the new
// preference for global numeric constants on the next autofix pass.
func TestPreferGlobalNumberConstantsPresetComposition(t *testing.T) {
	source := `const n = Number.NaN; const value = parseInt("10", 10);`
	expected := `const n = NaN; const value = Number.parseInt("10", 10);`
	for pass := range 2 {
		program, file, err := rule_tester.NewProgramHelper(fixtures.GetRootDir()).CreateTestProgram(source, "preset.js", "tsconfig.json")
		if err != nil {
			t.Fatal(err)
		}
		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program: lintprogram.NewFromCompiler(program), File: file.FileName(),
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{
					{Name: prefer_global_number_constants.PreferGlobalNumberConstantsRule.Name, Severity: rule.SeverityError, Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return prefer_global_number_constants.PreferGlobalNumberConstantsRule.Run(ctx, nil)
					}},
					{Name: prefer_number_properties.PreferNumberPropertiesRule.Name, Severity: rule.SeverityError, Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return prefer_number_properties.PreferNumberPropertiesRule.Run(ctx, []any{map[string]any{"checkNaN": false, "checkInfinity": false}})
					}},
				}
			},
			Consumer: rule.DiagnosticConsumer{Demand: rule.EditDemandAll, Report: func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }},
		})
		wanted := 2
		if pass == 1 {
			wanted = 0
		}
		if len(diagnostics) != wanted {
			t.Fatalf("pass %d: got %d diagnostics, want %d", pass, len(diagnostics), wanted)
		}
		output, _, fixed := linter.ApplyRuleFixes(source, diagnostics)
		if output != expected || fixed != (pass == 0) {
			t.Fatalf("pass %d: unexpected fix %q", pass, output)
		}
		source = output
	}
}
