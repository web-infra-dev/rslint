// Ported from eslint-plugin-unicorn v77.0.0; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v77.0.0/test/prefer-single-call.js
// Additional statement, AST, option and fix-safety cases. Expected edits
// and complete diagnostic ranges were checked with ESLint 10.9.0 and Unicorn 77.0.0.
package prefer_single_call_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_single_call"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferSingleCallEditDemand(t *testing.T) {
	for _, test := range []struct {
		code, output    string
		fix, suggestion bool
	}{
		{code: "const a=[]; a.push(1); a.push(2);", output: "const a=[]; a.push(1, 2);", fix: true},
		{code: "a.unshift(1); a.unshift(2);", output: "a.unshift(2, 1);", suggestion: true},
		{code: "importScripts(a); importScripts(load());", output: "importScripts(a, load());", suggestion: true},
		{code: "const a=[]; a.push(1); /*keep*/ a.push(2);"},
	} {
		t.Run(test.code, func(t *testing.T) {
			program, sourceFile, err := rule_tester.NewProgramHelper(fixtures.GetRootDir()).CreateTestProgram(test.code, "edit-demand.js", "tsconfig.json")
			if err != nil {
				t.Fatal(err)
			}
			diagnostics := map[rule.EditDemand]rule.RuleDiagnostic{}
			for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
				var found []rule.RuleDiagnostic
				linter.LintSingleFile(linter.LintSingleFileOptions{
					Program: lintprogram.NewFromCompiler(program), File: sourceFile.FileName(),
					GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
						return []rule.ConfiguredRule{{Name: prefer_single_call.PreferSingleCallRule.Name, Severity: rule.SeverityError, Run: func(ctx rule.RuleContext) rule.RuleListeners {
							return prefer_single_call.PreferSingleCallRule.Run(ctx, nil)
						}}}
					},
					Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(diagnostic rule.RuleDiagnostic) { found = append(found, diagnostic) }},
				})
				if len(found) != 1 {
					t.Fatalf("demand %d: expected one diagnostic, got %d", demand, len(found))
				}
				diagnostics[demand] = found[0]
			}
			all := diagnostics[rule.EditDemandAll]
			for demand, diagnostic := range diagnostics {
				if diagnostic.Range != all.Range || !reflect.DeepEqual(diagnostic.Message, all.Message) {
					t.Fatalf("demand %d changed diagnostic identity", demand)
				}
				wantFix := test.fix && demand&rule.EditDemandAutofix != 0
				wantSuggestion := test.suggestion && demand&rule.EditDemandSuggestion != 0
				if (diagnostic.FixesPtr != nil) != wantFix || (diagnostic.Suggestions != nil) != wantSuggestion {
					t.Fatalf("demand %d: unexpected edit artifacts", demand)
				}
				if wantFix && !reflect.DeepEqual(diagnostic.FixesPtr, all.FixesPtr) {
					t.Fatalf("demand %d changed fixes", demand)
				}
				if wantSuggestion && !reflect.DeepEqual(diagnostic.Suggestions, all.Suggestions) {
					t.Fatalf("demand %d changed suggestions", demand)
				}
			}
			if test.fix {
				output, _, fixed := linter.ApplyRuleFixes(test.code, []rule.RuleDiagnostic{all})
				if !fixed || output != test.output {
					t.Fatalf("fix: got %q, want %q", output, test.output)
				}
			}
			if test.suggestion {
				if len(*all.Suggestions) != 1 || (*all.Suggestions)[0].Message.Id != "suggestion" ||
					(*all.Suggestions)[0].Message.Description != "Merge with previous one." ||
					!reflect.DeepEqual((*all.Suggestions)[0].Message.Data, all.Message.Data) {
					t.Fatal("unexpected suggestion message")
				}
				output, _, fixed := linter.ApplyRuleFixes(test.code, *all.Suggestions)
				if !fixed || output != test.output {
					t.Fatalf("suggestion: got %q, want %q", output, test.output)
				}
			}
		})
	}
}

func TestPreferSingleCallExtras(t *testing.T) {
	t.Run("Statements and call shapes", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_single_call.PreferSingleCallRule, []rule_tester.ValidTestCase{
			{Code: "if (ready) values.push(1); else values.push(2);", FileName: "case.js"},
			{Code: "values.push(1); { values.push(2); }", FileName: "case.js"},
			{Code: "switch (x) { case 0: values.push(1); case 1: values.push(2); }", FileName: "case.js"},
			{Code: "values[\"push\"](1); values[\"push\"](2);", FileName: "case.js"},
			{Code: "class C { #push() {} m() { this.#push(1); this.#push(2); } }", FileName: "case.js"},
			{Code: "foo[\"classList\"].add(1); foo[\"classList\"].add(2);", FileName: "case.js"},
			{Code: "(foo?.classList).add(1); (foo?.classList).add(2);", FileName: "case.js"},
			{Code: "function f(values: number[]) { (values.push as Function)(1); (values.push as Function)(2); }", FileName: "case.ts"},
			{Code: "function f(values: number[]) { values.push!(1); values.push!(2); }", FileName: "case.ts"},
		}, []rule_tester.InvalidTestCase{
			{Code: "const values=[]; (values.push(1)); ((values.push(2)));", FileName: "case.js", Output: []string{"const values=[]; (values.push(1, 2));"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 45, EndLine: 1, EndColumn: 49, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const values=[]; /** @type {Array} */ (values).push(1); values.push(2);", FileName: "case.js", Output: []string{"const values=[]; /** @type {Array} */ (values).push(1, 2);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 64, EndLine: 1, EndColumn: 68, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "namespace N { const values: number[] = []; values.push(1); values.push(2); }", FileName: "case.ts", Output: []string{"namespace N { const values: number[] = []; values.push(1, 2); }"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 67, EndLine: 1, EndColumn: 71, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "class C { static { const values: number[] = []; values.push(1); values.push(2); } }", FileName: "case.ts", Output: []string{"class C { static { const values: number[] = []; values.push(1, 2); } }"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 72, EndLine: 1, EndColumn: 76, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "class C { #items: number[] = []; m() { this.#items.push(1); this.#items.push(2); } }", FileName: "case.ts", Output: []string{"class C { #items: number[] = []; m() { this.#items.push(1, 2); } }"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 73, EndLine: 1, EndColumn: 77, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "function f(values: number[]) { values.push<number>((1)); values.push<number>((2)); }", FileName: "case.ts", Output: []string{"function f(values: number[]) { values.push<number>((1), (2)); }"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 65, EndLine: 1, EndColumn: 69, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "function f(values: number[]) { (<number[]>values).push(1); (values satisfies number[]).push(2); }", FileName: "case.ts", Output: []string{"function f(values: number[]) { (<number[]>values).push(1, 2); }"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 88, EndLine: 1, EndColumn: 92, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "element.classList.add(\"a\"); element.classList.add(<i />);", FileName: "case.tsx", Output: []string{"element.classList.add(\"a\", <i />);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 1, Column: 47, EndLine: 1, EndColumn: 50, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "\"😀\"; const 值 = []; 值.push(1); 值.push(2);", FileName: "case.js", Output: []string{"\"😀\"; const 值 = []; 值.push(1, 2);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 34, EndLine: 1, EndColumn: 38, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
		})
	})
	t.Run("Comments, optional unshift and semicolons", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_single_call.PreferSingleCallRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
			{Code: "const a=[]; a.push(/*keep*/1); a.push(2);", FileName: "case.js", Output: []string{"const a=[]; a.push(/*keep*/1, 2);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 34, EndLine: 1, EndColumn: 38, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[]; a.push(1); a.push(/*keep*/2);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 26, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[]; a.push(1); /*keep*/ a.push(2);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 35, EndLine: 1, EndColumn: 39, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[]; a.unshift(/*keep*/1); a.unshift(2);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 1, Column: 37, EndLine: 1, EndColumn: 44, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[]; a.unshift(1); a.unshift(/*keep*/2,);", FileName: "case.js", Output: []string{"const a=[]; a.unshift(/*keep*/2, 1);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 1, Column: 29, EndLine: 1, EndColumn: 36, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[]; a.push(\"/*text*/\"); a.push(/a\\/\\/b/);", FileName: "case.js", Output: []string{"const a=[]; a.push(\"/*text*/\", /a\\/\\/b/);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 35, EndLine: 1, EndColumn: 39, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[]; a.unshift(); a?.unshift(2,);", FileName: "case.js", Output: []string{"const a=[]; a.unshift(2);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 1, Column: 29, EndLine: 1, EndColumn: 36, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[]; a.unshift(1); a?.unshift();", FileName: "case.js", Output: []string{"const a=[]; a.unshift(1);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 1, Column: 30, EndLine: 1, EndColumn: 37, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[]; a.unshift(); a?.unshift();", FileName: "case.js", Output: []string{"const a=[]; a.unshift();"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 1, Column: 29, EndLine: 1, EndColumn: 36, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[]; a.unshift(1,); a?.unshift(2,3,);", FileName: "case.js", Output: []string{"const a=[]; a.unshift(2,3, 1,);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 1, Column: 31, EndLine: 1, EndColumn: 38, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[];\nbar()\na.unshift(1);\n(a).unshift(2);", FileName: "case.js", Output: []string{"const a=[];\nbar()\n;(a).unshift(2, 1);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 4, Column: 5, EndLine: 4, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[];\nfunction f() {}\na.unshift(1);\n(a).unshift(2);", FileName: "case.js", Output: []string{"const a=[];\nfunction f() {}\n(a).unshift(2, 1);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 4, Column: 5, EndLine: 4, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[]; a.push(1)\na.push(2);\n[a].forEach(f);", FileName: "case.js", Output: []string{"const a=[]; a.push(1, 2);\n[a].forEach(f);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 2, Column: 3, EndLine: 2, EndColumn: 7, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[]; a.push(1);\r\na.\r\npush(2);", FileName: "case.js", Output: []string{"const a=[]; a.push(1, 2);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 3, Column: 1, EndLine: 3, EndColumn: 5, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
		})
	})
	t.Run("Options and reference identity", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_single_call.PreferSingleCallRule, []rule_tester.ValidTestCase{
			{Code: "(stream).push(1); (stream).push(2);", FileName: "case.js"},
			{Code: "foo.classList.remove(1); foo.classList.remove(2);", FileName: "case.js", Options: []any{map[string]any{"ignore": []any{"foo.classList.remove"}}}},
			{Code: "foo.bar.push(1); foo.bar.push(2);", FileName: "case.js", Options: []any{map[string]any{"ignore": []any{"  foo.bar.push  "}}}},
			{Code: "stream.push(1); stream.push(2); a.push(1); a.push(2);", FileName: "case.js", Options: []any{map[string]any{"ignore": []any{"a.push"}}}},
			{Code: "foo[a()].push(1); foo[a()].push(2);", FileName: "case.js"},
		}, []rule_tester.InvalidTestCase{
			{Code: "stream?.push(1); stream?.push(2);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 26, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "stream?.push(1, 2);"}}},
			}},
			{Code: "this.stream?.unshift(1); this.stream?.unshift(2);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 1, Column: 39, EndLine: 1, EndColumn: 46, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "this.stream?.unshift(2, 1);"}}},
			}},
			{Code: "process?.stdout.push(1); process?.stdout.push(2);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 42, EndLine: 1, EndColumn: 46, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "process?.stdout.push(1, 2);"}}},
			}},
			{Code: "const a=[]; a.push(1); a.push(2);", FileName: "case.js", Options: []any{map[string]any{}}, Output: []string{"const a=[]; a.push(1, 2);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 26, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[]; a.unshift(1); a.unshift(2);", FileName: "case.js", Options: []any{map[string]any{"ignore": []any{}}}, Output: []string{"const a=[]; a.unshift(2, 1);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 1, Column: 29, EndLine: 1, EndColumn: 36, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "foo[\"bar\"].push(1); foo[\"bar\"].push(2);", FileName: "case.js", Options: []any{map[string]any{"ignore": []any{"foo.bar.push"}}}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 32, EndLine: 1, EndColumn: 36, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo[\"bar\"].push(1, 2);"}}},
			}},
			{Code: "foo[\"a\"+\"b\"].push(1); foo.ab.push(2);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 30, EndLine: 1, EndColumn: 34, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo[\"a\"+\"b\"].push(1, 2);"}}},
			}},
			{Code: "foo[0x10].push(1); foo[16].push(2);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 28, EndLine: 1, EndColumn: 32, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo[0x10].push(1, 2);"}}},
			}},
		})
	})
	t.Run("Argument safety", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_single_call.PreferSingleCallRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
			{Code: "importScripts(a); importScripts(import(\"module.js\"));", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `importScripts()` multiple times.", Line: 1, Column: 19, EndLine: 1, EndColumn: 32, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "importScripts(a, import(\"module.js\"));"}}},
			}},
			{Code: "element.classList.add(a); element.classList.add(<i value={sideEffect()} />);", FileName: "case.tsx", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 1, Column: 45, EndLine: 1, EndColumn: 48, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "element.classList.add(a, <i value={sideEffect()} />);"}}},
			}},
			{Code: "element.classList.add(a); element.classList.add({ get [sideEffect()]() { return 1; } });", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 1, Column: 45, EndLine: 1, EndColumn: 48, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "element.classList.add(a, { get [sideEffect()]() { return 1; } });"}}},
			}},
			{Code: "const a=[]; let n=1; a.push(n); a.push(2);", FileName: "case.js", Output: []string{"const a=[]; let n=1; a.push(n, 2);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 35, EndLine: 1, EndColumn: 39, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[]; const n=1; a.push(n); a.push(2);", FileName: "case.js", Output: []string{"const a=[]; const n=1; a.push(n, 2);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 37, EndLine: 1, EndColumn: 41, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[]; const x={value:1}; a.push(x.value); a.push(2);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 51, EndLine: 1, EndColumn: 55, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; const x={value:1}; a.push(x.value, 2);"}}},
			}},
			{Code: "const a=[]; a.push(({value:1}).value); a.push([2][0]);", FileName: "case.js", Output: []string{"const a=[]; a.push(({value:1}).value, [2][0]);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 42, EndLine: 1, EndColumn: 46, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[]; a.push(Object.freeze([1])); a.push(Math.PI);", FileName: "case.js", Output: []string{"const a=[]; a.push(Object.freeze([1]), Math.PI);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 43, EndLine: 1, EndColumn: 47, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[]; a.push(...[1]); a.push(2);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 31, EndLine: 1, EndColumn: 35, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; a.push(...[1], 2);"}}},
			}},
			// The shared evaluator does not fold array spreads. Keep the merge as
			// a suggestion; see the documented narrow difference from upstream.
			{Code: "const a=[]; a.push([...[1]]); a.push(2);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 33, EndLine: 1, EndColumn: 37, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; a.push([...[1]], 2);"}}},
			}},
			{Code: "const a=[]; const values={*[Symbol.iterator](){a.push(0);yield 1;}}; a.push([...values]); a.push(2);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 93, EndLine: 1, EndColumn: 97, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; const values={*[Symbol.iterator](){a.push(0);yield 1;}}; a.push([...values], 2);"}}},
			}},
			{Code: "const a=[]; a.push(1); a.push(n++);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 26, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; a.push(1, n++);"}}},
			}},
			{Code: "const a=[]; a.push(1); a.push(void sideEffect());", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 26, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; a.push(1, void sideEffect());"}}},
			}},
			{Code: "element.classList.add(a()); element.classList.add(b);", FileName: "case.js", Output: []string{"element.classList.add(a(), b);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 1, Column: 47, EndLine: 1, EndColumn: 50, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "element.classList.add(a); element.classList.add(object.value);", FileName: "case.js", Output: []string{"element.classList.add(a, object.value);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 1, Column: 45, EndLine: 1, EndColumn: 48, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "element.classList.add(a); element.classList.add(() => sideEffect());", FileName: "case.js", Output: []string{"element.classList.add(a, () => sideEffect());"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 1, Column: 45, EndLine: 1, EndColumn: 48, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "element.classList.add(a); element.classList.add({m(v=sideEffect()) { sideEffect(); }});", FileName: "case.js", Output: []string{"element.classList.add(a, {m(v=sideEffect()) { sideEffect(); }});"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 1, Column: 45, EndLine: 1, EndColumn: 48, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "element.classList.add(a); element.classList.add({[sideEffect()]() {}});", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 1, Column: 45, EndLine: 1, EndColumn: 48, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "element.classList.add(a, {[sideEffect()]() {}});"}}},
			}},
			// A decorator expression runs while the class is created, even though
			// the method body is deferred. Verified with Unicorn v77.0.0.
			{Code: "element.classList.add(a); element.classList.add(class { @decorate() method() {} });", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 1, Column: 45, EndLine: 1, EndColumn: 48, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "element.classList.add(a, class { @decorate() method() {} });"}}},
			}},
			{Code: "element.classList.add(a); element.classList.add(class { value=sideEffect(); });", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 1, Column: 45, EndLine: 1, EndColumn: 48, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "element.classList.add(a, class { value=sideEffect(); });"}}},
			}},
			{Code: "element.classList.add(a); element.classList.add(class { static value=sideEffect(); });", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 1, Column: 45, EndLine: 1, EndColumn: 48, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "element.classList.add(a, class { static value=sideEffect(); });"}}},
			}},
			{Code: "element.classList.add(a); element.classList.add(class { static { sideEffect(); } });", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 1, Column: 45, EndLine: 1, EndColumn: 48, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "element.classList.add(a, class { static { sideEffect(); } });"}}},
			}},
			{Code: "importScripts(a); importScripts(delete object.key);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `importScripts()` multiple times.", Line: 1, Column: 19, EndLine: 1, EndColumn: 32, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "importScripts(a, delete object.key);"}}},
			}},
			{Code: "importScripts(a); importScripts(b = \"x\");", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `importScripts()` multiple times.", Line: 1, Column: 19, EndLine: 1, EndColumn: 32, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "importScripts(a, b = \"x\");"}}},
			}},
			{Code: "async function f() { importScripts(a); importScripts(await load()); }", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `importScripts()` multiple times.", Line: 1, Column: 40, EndLine: 1, EndColumn: 53, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "async function f() { importScripts(a, await load()); }"}}},
			}},
			{Code: "function* f() { importScripts(a); importScripts(yield next); }", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `importScripts()` multiple times.", Line: 1, Column: 35, EndLine: 1, EndColumn: 48, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "function* f() { importScripts(a, yield next); }"}}},
			}},
		})
	})
}
