// Additional AST and scope regressions verified against Unicorn v75.0.0.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/test/prefer-identifier-import-export-specifiers.js
package prefer_identifier_import_export_specifiers_test

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_identifier_import_export_specifiers"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"reflect"
	"testing"
)

func TestPreferIdentifierImportExportSpecifiersExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_identifier_import_export_specifiers.PreferIdentifierImportExportSpecifiersRule, []rule_tester.ValidTestCase{
		{Code: "import {\"a-b\" as x, \"1x\" as y} from \"module\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "import {\"\\u200Cx\" as x} from \"module\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "import {\"💡\" as x} from \"module\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "export * as\"x-y\" from \"module\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "const obj = {\"type\": \"json\"};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "const value = \"foo\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
	}, []rule_tester.InvalidTestCase{
		{Code: "import {\"NaN\" as x, \"Infinity\" as y} from \"module\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"import {NaN as x, Infinity as y} from \"module\";"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `NaN` over string literal `\"NaN\"`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `Infinity` over string literal `\"Infinity\"`.", Line: 1, Column: 21, EndLine: 1, EndColumn: 31, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "import {\"class\" as x} from \"module\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"import {class as x} from \"module\";"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `class` over string literal `\"class\"`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "import {\"éclair\" as x, \"$x\" as y, \"_x\" as z} from \"module\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"import {éclair as x, $x as y, _x as z} from \"module\";"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `éclair` over string literal `\"éclair\"`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `$x` over string literal `\"$x\"`.", Line: 1, Column: 24, EndLine: 1, EndColumn: 28, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `_x` over string literal `\"_x\"`.", Line: 1, Column: 35, EndLine: 1, EndColumn: 39, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "import {\"\U00010400name\" as x} from \"module\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"import {\U00010400name as x} from \"module\";"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `\U00010400name` over string literal `\"\U00010400name\"`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "import {\"x\\u200Cy\" as x} from \"module\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"import {x\u200cy as x} from \"module\";"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `x\u200cy` over string literal `\"x\\u200Cy\"`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "import {\"\\u0066oo\"as x} from \"module\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"import {foo as x} from \"module\";"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `foo` over string literal `\"\\u0066oo\"`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "import {\"x\" /* keep */ as y} from \"module\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"import {x /* keep */ as y} from \"module\";"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `x` over string literal `\"x\"`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "export {\"x\"as\"y\"} from \"module\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"export {x as y} from \"module\";"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `x` over string literal `\"x\"`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `y` over string literal `\"y\"`.", Line: 1, Column: 14, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "export {\"x\" as \"x\"} from \"module\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"export {x as x} from \"module\";"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `x` over string literal `\"x\"`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `x` over string literal `\"x\"`.", Line: 1, Column: 16, EndLine: 1, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "export {\"x\"} from \"module\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"export {x} from \"module\";"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `x` over string literal `\"x\"`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "export * as\"default\" from \"module\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"export * as default from \"module\";"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `default` over string literal `\"default\"`.", Line: 1, Column: 12, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "import x from \"module\" with {\"type\": \"json\", \"default\": \"x\"};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"import x from \"module\" with {type: \"json\", default: \"x\"};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `type` over string literal `\"type\"`.", Line: 1, Column: 30, EndLine: 1, EndColumn: 36, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `default` over string literal `\"default\"`.", Line: 1, Column: 46, EndLine: 1, EndColumn: 55, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "export {x} from \"module\" with {\"type\": \"json\"};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"export {x} from \"module\" with {type: \"json\"};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `type` over string literal `\"type\"`.", Line: 1, Column: 32, EndLine: 1, EndColumn: 38, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "\"😀\"; export * as \"foo\" from \"module\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"\"😀\"; export * as foo from \"module\";"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `foo` over string literal `\"foo\"`.", Line: 1, Column: 19, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "import type {\"foo\"as Foo} from \"module\";", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"import type {foo as Foo} from \"module\";"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `foo` over string literal `\"foo\"`.", Line: 1, Column: 14, EndLine: 1, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "export type {\"foo\" as \"bar\"} from \"module\";", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"export type {foo as bar} from \"module\";"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `foo` over string literal `\"foo\"`.", Line: 1, Column: 14, EndLine: 1, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `bar` over string literal `\"bar\"`.", Line: 1, Column: 23, EndLine: 1, EndColumn: 28, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "export type * as \"types\" from \"module\";", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"export type * as types from \"module\";"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-identifier-import-export-specifiers", Message: "Prefer identifier `types` over string literal `\"types\"`.", Line: 1, Column: 18, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
	})
}

func TestPreferIdentifierImportExportSpecifiersArtifactsFollowDemand(t *testing.T) {
	for _, testCase := range []struct{ source, output string }{{source: "import {\"foo\" as foo} from \"foo\";", output: "import {foo as foo} from \"foo\";"},
		{source: "// ❌\nimport foo from 'foo' with {'type': 'json'};\n\n", output: "// ❌\nimport foo from 'foo' with {type: 'json'};\n\n"}} {
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
						return []rule.ConfiguredRule{{Name: prefer_identifier_import_export_specifiers.PreferIdentifierImportExportSpecifiersRule.Name, Severity: rule.SeverityError, Run: func(ctx rule.RuleContext) rule.RuleListeners {
							return prefer_identifier_import_export_specifiers.PreferIdentifierImportExportSpecifiersRule.Run(ctx, nil)
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
