// Additional AST and scope regressions verified against Unicorn v75.0.0.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/test/no-unnecessary-array-splice-count.js
package no_unnecessary_array_splice_count_test

import (
	"path/filepath"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_unnecessary_array_splice_count"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
	"reflect"
	"testing"
)

func TestNoUnnecessaryArraySpliceCountExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_unnecessary_array_splice_count.NoUnnecessaryArraySpliceCountRule, []rule_tester.ValidTestCase{
		{Code: "const x = (a?.splice)(1, a.length);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "const x = a?.splice?.(1, a.length);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "const x = a[\"splice\"](1, a.length);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "const x = a.splice(1, a[\"length\"]);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "const x = getArray().splice(1, getArray().length);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "const x = a.splice(...start, Infinity);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "const x = a.splice(1, ...end);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "const x = a.splice(1, +Infinity);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "const x = a.splice(1, Number[\"POSITIVE_INFINITY\"]);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "const x = a.splice(1, Number?.POSITIVE_INFINITY);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "class C { #splice(start, end) {} f() { this.#splice(1, Infinity); } }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "function f(a: string) { return a.splice(1, a.length); }", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "function f(a: Set<number>) { return a.splice(1, Infinity); }", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "const x = (a?.toSpliced)(1, a.length);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "const x = a?.toSpliced?.(1, a.length);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "const x = a[\"toSpliced\"](1, a.length);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "const x = a.toSpliced(1, a[\"length\"]);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "const x = getArray().toSpliced(1, getArray().length);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "const x = a.toSpliced(...start, Infinity);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "const x = a.toSpliced(1, ...end);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "const x = a.toSpliced(1, +Infinity);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "const x = a.toSpliced(1, Number[\"POSITIVE_INFINITY\"]);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "const x = a.toSpliced(1, Number?.POSITIVE_INFINITY);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "class C { #toSpliced(start, end) {} f() { this.#toSpliced(1, Infinity); } }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "function f(a: string) { return a.toSpliced(1, a.length); }", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "function f(a: Set<number>) { return a.toSpliced(1, Infinity); }", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "function f(Infinity) { return a.splice(1, Infinity); }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "function f(Infinity) { return a.toSpliced(1, Infinity); }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "function f(Number) { return a.splice(1, Number.POSITIVE_INFINITY); }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "a.splice(1, Infinity);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly", "Infinity": "off"}},
		{Code: "function f(Number) { return a.toSpliced(1, Number.POSITIVE_INFINITY); }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "a.toSpliced(1, Infinity);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly", "Infinity": "off"}},
	}, []rule_tester.InvalidTestCase{
		{Code: "const x = a.splice(1, (a.length));", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"const x = a.splice(1);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `a.length` as the `deleteCount` argument is unnecessary.", Line: 1, Column: 24, EndLine: 1, EndColumn: 32, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const x = a.splice((1 /* keep */), /* remove */ ((a.length)), /* keep */);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"const x = a.splice((1 /* keep */), /* keep */);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `a.length` as the `deleteCount` argument is unnecessary.", Line: 1, Column: 51, EndLine: 1, EndColumn: 59, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const x = a.splice(1 /* keep */, a.length /* keep */);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"const x = a.splice(1 /* keep */ /* keep */);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `a.length` as the `deleteCount` argument is unnecessary.", Line: 1, Column: 34, EndLine: 1, EndColumn: 42, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const x = (a.splice)(1, a.length);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"const x = (a.splice)(1);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `a.length` as the `deleteCount` argument is unnecessary.", Line: 1, Column: 25, EndLine: 1, EndColumn: 33, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const x = a.splice(1, a?.length);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"const x = a.splice(1);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `a?.length` as the `deleteCount` argument is unnecessary.", Line: 1, Column: 23, EndLine: 1, EndColumn: 32, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const x = obj.a.splice(1, obj.a.length);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"const x = obj.a.splice(1);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `….length` as the `deleteCount` argument is unnecessary.", Line: 1, Column: 27, EndLine: 1, EndColumn: 39, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const x = obj[\"a\"].splice(1, obj.a.length);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"const x = obj[\"a\"].splice(1);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `….length` as the `deleteCount` argument is unnecessary.", Line: 1, Column: 30, EndLine: 1, EndColumn: 42, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "\"😀\"; a.splice(1, Infinity);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"\"😀\"; a.splice(1);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `Infinity` as the `deleteCount` argument is unnecessary.", Line: 1, Column: 19, EndLine: 1, EndColumn: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const x = /** @type {Array<number>} */ (a).splice(1, a.length);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"const x = /** @type {Array<number>} */ (a).splice(1);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `a.length` as the `deleteCount` argument is unnecessary.", Line: 1, Column: 54, EndLine: 1, EndColumn: 62, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "function f(a: readonly number[]) { return a.splice(1, a.length); }", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"function f(a: readonly number[]) { return a.splice(1); }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `a.length` as the `deleteCount` argument is unnecessary.", Line: 1, Column: 55, EndLine: 1, EndColumn: 63, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "(a satisfies number[]).splice(1, (a satisfies number[]).length);", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"(a satisfies number[]).splice(1);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `….length` as the `deleteCount` argument is unnecessary.", Line: 1, Column: 34, EndLine: 1, EndColumn: 63, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "a!.splice(1, a!.length);", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"a!.splice(1);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `….length` as the `deleteCount` argument is unnecessary.", Line: 1, Column: 14, EndLine: 1, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const x = a.toSpliced(1, (a.length));", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"const x = a.toSpliced(1);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `a.length` as the `skipCount` argument is unnecessary.", Line: 1, Column: 27, EndLine: 1, EndColumn: 35, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const x = a.toSpliced((1 /* keep */), /* remove */ ((a.length)), /* keep */);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"const x = a.toSpliced((1 /* keep */), /* keep */);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `a.length` as the `skipCount` argument is unnecessary.", Line: 1, Column: 54, EndLine: 1, EndColumn: 62, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const x = a.toSpliced(1 /* keep */, a.length /* keep */);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"const x = a.toSpliced(1 /* keep */ /* keep */);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `a.length` as the `skipCount` argument is unnecessary.", Line: 1, Column: 37, EndLine: 1, EndColumn: 45, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const x = (a.toSpliced)(1, a.length);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"const x = (a.toSpliced)(1);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `a.length` as the `skipCount` argument is unnecessary.", Line: 1, Column: 28, EndLine: 1, EndColumn: 36, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const x = a.toSpliced(1, a?.length);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"const x = a.toSpliced(1);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `a?.length` as the `skipCount` argument is unnecessary.", Line: 1, Column: 26, EndLine: 1, EndColumn: 35, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const x = obj.a.toSpliced(1, obj.a.length);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"const x = obj.a.toSpliced(1);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `….length` as the `skipCount` argument is unnecessary.", Line: 1, Column: 30, EndLine: 1, EndColumn: 42, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const x = obj[\"a\"].toSpliced(1, obj.a.length);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"const x = obj[\"a\"].toSpliced(1);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `….length` as the `skipCount` argument is unnecessary.", Line: 1, Column: 33, EndLine: 1, EndColumn: 45, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "\"😀\"; a.toSpliced(1, Infinity);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"\"😀\"; a.toSpliced(1);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `Infinity` as the `skipCount` argument is unnecessary.", Line: 1, Column: 22, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const x = /** @type {Array<number>} */ (a).toSpliced(1, a.length);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"const x = /** @type {Array<number>} */ (a).toSpliced(1);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `a.length` as the `skipCount` argument is unnecessary.", Line: 1, Column: 57, EndLine: 1, EndColumn: 65, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "function f(a: readonly number[]) { return a.toSpliced(1, a.length); }", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"function f(a: readonly number[]) { return a.toSpliced(1); }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `a.length` as the `skipCount` argument is unnecessary.", Line: 1, Column: 58, EndLine: 1, EndColumn: 66, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "(a satisfies number[]).toSpliced(1, (a satisfies number[]).length);", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"(a satisfies number[]).toSpliced(1);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `….length` as the `skipCount` argument is unnecessary.", Line: 1, Column: 37, EndLine: 1, EndColumn: 66, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "a!.toSpliced(1, a!.length);", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"a!.toSpliced(1);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-array-splice-count", Message: "Passing `….length` as the `skipCount` argument is unnecessary.", Line: 1, Column: 17, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
	})
}

func TestNoUnnecessaryArraySpliceCountArtifactsFollowDemand(t *testing.T) {
	for _, testCase := range []struct{ source, output string }{{source: "foo.splice(1, foo.length)", output: "foo.splice(1)"},
		{source: "// ❌\narray.splice(1, Number.POSITIVE_INFINITY);\n\n", output: "// ❌\narray.splice(1);\n\n"}} {
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
						return []rule.ConfiguredRule{{Name: no_unnecessary_array_splice_count.NoUnnecessaryArraySpliceCountRule.Name, Severity: rule.SeverityError, Run: func(ctx rule.RuleContext) rule.RuleListeners {
							return no_unnecessary_array_splice_count.NoUnnecessaryArraySpliceCountRule.Run(ctx, nil)
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

func TestNoUnnecessaryArraySpliceCountReviewRegressions(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_unnecessary_array_splice_count.NoUnnecessaryArraySpliceCountRule, []rule_tester.ValidTestCase{
		{
			Code:     "Number = {POSITIVE_INFINITY: 1}; const array = [0,1,2]; array.splice(1, Number.POSITIVE_INFINITY);",
			FileName: "review.js",
		},
		{
			Code:     "globalThis.Number = {POSITIVE_INFINITY: 1}; const array = [0,1,2]; array.splice(1, Number.POSITIVE_INFINITY);",
			FileName: "review.js",
		},
		{
			Code:     "window['Number'] = {POSITIVE_INFINITY: 1}; const array = [0,1,2]; array.splice(1, Number.POSITIVE_INFINITY);",
			FileName: "review.js",
			Globals:  map[string]any{"window": "readonly"},
		},
		{
			Code:     "globalThis['Num' + 'ber'] = {POSITIVE_INFINITY: 1}; const array = [0,1,2]; array.splice(1, Number.POSITIVE_INFINITY);",
			FileName: "review.js",
		},
		{
			Code:     "self.Number = {POSITIVE_INFINITY: 1}; const array = [0,1,2]; array.splice(1, Number.POSITIVE_INFINITY);",
			FileName: "review.js",
			Globals:  map[string]any{"self": "readonly"},
		},
		{
			Code:     "global.Number = {POSITIVE_INFINITY: 1}; const array = [0,1,2]; array.splice(1, Number.POSITIVE_INFINITY);",
			FileName: "review.js",
			Globals:  map[string]any{"global": "readonly"},
		},
		{
			Code:     "delete globalThis.Number; const array = [0,1,2]; array.splice(1, Number.POSITIVE_INFINITY);",
			FileName: "review.js",
		},
		{
			Code:     "let i = 0; const first = [0,1,2], second = [0]; const obj = { get a() { return i++ ? second : first; } }; obj.a.splice(1, obj.a.length);",
			FileName: "review.js",
		},
		{
			Code:     "let a = [0,1,2], b = [0]; a.splice((a = b, 1), a.length);",
			FileName: "review.js",
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code:     "const globalThis = {Number: {POSITIVE_INFINITY: 1}}; const array = [0,1,2]; array.splice(1, Number.POSITIVE_INFINITY);",
			FileName: "review.js",
			Output:   []string{"const globalThis = {Number: {POSITIVE_INFINITY: 1}}; const array = [0,1,2]; array.splice(1);"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "no-unnecessary-array-splice-count",
			}},
		},
		{
			Code:     "const array = [0,1,2]; array.splice(1, Number.POSITIVE_INFINITY); globalThis.Number = {POSITIVE_INFINITY: 1};",
			FileName: "review.js",
			Output:   []string{"const array = [0,1,2]; array.splice(1); globalThis.Number = {POSITIVE_INFINITY: 1};"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "no-unnecessary-array-splice-count",
			}},
		},
	})
}

func TestNoUnnecessaryArraySpliceCountProjectFalseSafety(t *testing.T) {
	for _, testCase := range []struct {
		name            string
		code            string
		wantDiagnostics int
	}{
		{
			name: "getter receiver",
			code: "let i = 0; const first = [0,1,2], second = [0]; const obj = { get a() { return i++ ? second : first; } }; obj.a.splice(1, obj.a.length);",
		},
		{
			name: "global object Number write",
			code: "globalThis.Number = {POSITIVE_INFINITY: 1}; const array = [0,1,2]; array.splice(1, Number.POSITIVE_INFINITY);",
		},
		{
			name: "computed global object Number write",
			code: "globalThis['Num' + 'ber'] = {POSITIVE_INFINITY: 1}; const array = [0,1,2]; array.splice(1, Number.POSITIVE_INFINITY);",
		},
		{
			name:            "identifier receiver",
			code:            "const array = [0,1,2]; array.splice(1, array.length);",
			wantDiagnostics: 1,
		},
		{
			name:            "later global object Number write",
			code:            "const array = [0,1,2]; array.splice(1, Number.POSITIVE_INFINITY); globalThis.Number = {POSITIVE_INFINITY: 1};",
			wantDiagnostics: 1,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			diagnostics := lintNoUnnecessaryArraySpliceCountSourceOnly(t, testCase.code)
			if len(diagnostics) != testCase.wantDiagnostics {
				t.Fatalf("project:false diagnostics = %d, want %d: %+v", len(diagnostics), testCase.wantDiagnostics, diagnostics)
			}
		})
	}
}

func lintNoUnnecessaryArraySpliceCountSourceOnly(t *testing.T, code string) []rule.RuleDiagnostic {
	t.Helper()
	dir := tspath.NormalizePath(t.TempDir())
	fileName := tspath.NormalizePath(filepath.Join(dir, "file.js"))
	fs := utils.NewOverlayVFS(bundled.WrapFS(osvfs.FS()), map[string]string{fileName: code})
	program, err := lintprogram.NewFromRoots(lintprogram.RootOptions{
		RootFileNames:   []string{fileName},
		Host:            utils.CreateCompilerHost(dir, fs),
		CompilerOptions: &core.CompilerOptions{Target: core.ScriptTargetESNext},
		SingleThreaded:  true,
	})
	if err != nil {
		t.Fatalf("create project:false program: %v", err)
	}

	if program.CanProvideTypeChecker(program.SourceFiles()[0]) {
		t.Fatal("project:false fixture unexpectedly received a TypeChecker")
	}

	diagnostics := []rule.RuleDiagnostic{}
	linter.LintSingleFile(linter.LintSingleFileOptions{
		Program: program,
		File:    fileName,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
				Name:     no_unnecessary_array_splice_count.NoUnnecessaryArraySpliceCountRule.Name,
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					if ctx.TypeChecker != nil {
						t.Fatal("project:false fixture unexpectedly received a TypeChecker")
					}
					return no_unnecessary_array_splice_count.NoUnnecessaryArraySpliceCountRule.Run(ctx, nil)
				},
			}}
		},
		Consumer: rule.DiagnosticConsumer{
			Demand: rule.EditDemandAutofix,
			Report: func(diagnostic rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, diagnostic)
			},
		},
	})
	return diagnostics
}
