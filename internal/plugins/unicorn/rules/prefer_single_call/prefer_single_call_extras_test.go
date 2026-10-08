// Ported from eslint-plugin-unicorn v77.0.0; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v77.0.0/test/prefer-single-call.js
// Additional statement, AST, option and fix-safety cases. Expected edits
// and complete diagnostic ranges were checked with ESLint 10.9.0 and Unicorn 77.0.0,
// except for the explicitly noted safety regressions where rslint avoids unsafe fixes.
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
	t.Run("Static object spreads", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_single_call.PreferSingleCallRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
			{Code: "const a=[]; a.push(1); a.push({...{x:1}});", FileName: "case.js", Output: []string{"const a=[]; a.push(1, {...{x:1}});"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 26, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[]; a.unshift(1); a.unshift({...{x:1}});", FileName: "case.js", Output: []string{"const a=[]; a.unshift({...{x:1}}, 1);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 1, Column: 29, EndLine: 1, EndColumn: 36, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[]; a.push(1); a.push({...null,...undefined,...1,...true,...1n});", FileName: "case.js", Output: []string{"const a=[]; a.push(1, {...null,...undefined,...1,...true,...1n});"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 26, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[]; a.push(1); a.push({...\"😀a\"});", FileName: "case.js", Output: []string{"const a=[]; a.push(1, {...\"😀a\"});"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 26, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[]; a.push(1); a.push({...\"\\uD800\"});", FileName: "case.js", Output: []string{"const a=[]; a.push(1, {...\"\\uD800\"});"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 26, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[]; a.push(1); a.push({...[,undefined,2]});", FileName: "case.js", Output: []string{"const a=[]; a.push(1, {...[,undefined,2]});"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 26, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[]; a.push(1); a.push([{...{x:1}, x:2, ...{x:3}}]);", FileName: "case.js", Output: []string{"const a=[]; a.push(1, [{...{x:1}, x:2, ...{x:3}}]);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 26, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[]; a.push(1); a.push(Object.freeze({...{x:1}}));", FileName: "case.js", Output: []string{"const a=[]; a.push(1, Object.freeze({...{x:1}}));"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 26, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[]; a.push(1); a.push({...{[\"__proto__\"]:1}});", FileName: "case.js", Output: []string{"const a=[]; a.push(1, {...{[\"__proto__\"]:1}});"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 26, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[]; a.push(1); a.push({...{__proto__:{x:1}}});", FileName: "case.js", Output: []string{"const a=[]; a.push(1, {...{__proto__:{x:1}}});"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 26, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[]; a.push(1); a.push({...unknown});", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 26, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; a.push(1, {...unknown});"}}},
			}},
			{Code: "const a=[]; a.push(1); a.push({...{get x(){return sideEffect();}}});", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 26, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; a.push(1, {...{get x(){return sideEffect();}}});"}}},
			}},
			{Code: "const a=[]; a.push(1); a.push({...getValues()});", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 26, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; a.push(1, {...getValues()});"}}},
			}},
			{Code: "const a=[]; const x={value:1}; x.value=2; a.push(1); a.push({...x});", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 56, EndLine: 1, EndColumn: 60, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; const x={value:1}; x.value=2; a.push(1, {...x});"}}},
			}},
		})
	})
	t.Run("TypeScript runtime wrappers", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_single_call.PreferSingleCallRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
			{Code: "element.classList.add(a); element.classList.add(class extends sideEffect() {});", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 1, Column: 45, EndLine: 1, EndColumn: 48, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "element.classList.add(a, class extends sideEffect() {});"}}},
			}},
			{Code: "element.classList.add(a); element.classList.add(class extends sideEffect()<T> {});", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 1, Column: 45, EndLine: 1, EndColumn: 48, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "element.classList.add(a, class extends sideEffect()<T> {});"}}},
			}},
			{Code: "element.classList.add(a); element.classList.add(class extends (sideEffect(), Base) {});", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 1, Column: 45, EndLine: 1, EndColumn: 48, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "element.classList.add(a, class extends (sideEffect(), Base) {});"}}},
			}},
			{Code: "element.classList.add(a); element.classList.add(class extends (Base = Other) {});", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 1, Column: 45, EndLine: 1, EndColumn: 48, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "element.classList.add(a, class extends (Base = Other) {});"}}},
			}},
			{Code: "element.classList.add(a); element.classList.add((sideEffect())<T>);", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 1, Column: 45, EndLine: 1, EndColumn: 48, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "element.classList.add(a, (sideEffect())<T>);"}}},
			}},
			{Code: "importScripts(a); importScripts((sideEffect(), fn)<T>);", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `importScripts()` multiple times.", Line: 1, Column: 19, EndLine: 1, EndColumn: 32, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "importScripts(a, (sideEffect(), fn)<T>);"}}},
			}},
			{Code: "element.classList.add(a); element.classList.add((fn = replacement)<T>);", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 1, Column: 45, EndLine: 1, EndColumn: 48, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "element.classList.add(a, (fn = replacement)<T>);"}}},
			}},
			{Code: "const a=[]; a.push(1); a.push(void ((sideEffect())<T>));", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 26, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; a.push(1, void ((sideEffect())<T>));"}}},
			}},
			{Code: "const a=[]; a.push(1); a.push(void (class extends sideEffect() {}));", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 26, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; a.push(1, void (class extends sideEffect() {}));"}}},
			}},
			{Code: "element.classList.add(a); element.classList.add(class extends Base<ReturnType<typeof sideEffect>> {});", FileName: "case.ts", Output: []string{"element.classList.add(a, class extends Base<ReturnType<typeof sideEffect>> {});"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 1, Column: 45, EndLine: 1, EndColumn: 48, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "element.classList.add(a); element.classList.add(class implements NS.Interface {});", FileName: "case.ts", Output: []string{"element.classList.add(a, class implements NS.Interface {});"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 1, Column: 45, EndLine: 1, EndColumn: 48, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "element.classList.add(a); element.classList.add(class { static { interface I extends NS.Type {} } });", FileName: "case.ts", Output: []string{"element.classList.add(a, class { static { interface I extends NS.Type {} } });"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 1, Column: 45, EndLine: 1, EndColumn: 48, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "element.classList.add(a); element.classList.add(fn<ReturnType<typeof sideEffect>>);", FileName: "case.ts", Output: []string{"element.classList.add(a, fn<ReturnType<typeof sideEffect>>);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 1, Column: 45, EndLine: 1, EndColumn: 48, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
		})
	})
	t.Run("Aliased iterator mutation", func(t *testing.T) {
		// Unlike upstream, never autofix a spread whose iterator is visibly replaced,
		// including writes through aliases and Object.defineProperty. Iteration can
		// mutate the receiver, so merging must remain an explicit suggestion.
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_single_call.PreferSingleCallRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
			{Code: "const a=[]; const values=[1]; const alias=values; alias[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1); a.push([...values]);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 120, EndLine: 1, EndColumn: 124, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; const values=[1]; const alias=values; alias[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1, [...values]);"}}},
			}},
			{Code: "const a=[]; const values=[1]; const alias=values; const alias2=alias; alias2[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1); a.push([...values]);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 141, EndLine: 1, EndColumn: 145, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; const values=[1]; const alias=values; const alias2=alias; alias2[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1, [...values]);"}}},
			}},
			{Code: "const a=[]; const values=[1]; const holder={values}; const {values:alias}=holder; alias[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1); a.push([...values]);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 152, EndLine: 1, EndColumn: 156, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; const values=[1]; const holder={values}; const {values:alias}=holder; alias[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1, [...values]);"}}},
			}},
			{Code: "const a=[]; const values=[1]; const [alias]=[values]; alias[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1); a.push([...values]);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 124, EndLine: 1, EndColumn: 128, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; const values=[1]; const [alias]=[values]; alias[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1, [...values]);"}}},
			}},
			{Code: "const a=[]; const values=[1]; const holder=Object.assign({}, {values}); holder.values[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1); a.push([...values]);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 150, EndLine: 1, EndColumn: 154, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; const values=[1]; const holder=Object.assign({}, {values}); holder.values[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1, [...values]);"}}},
			}},
			{Code: "const a=[]; const values=[1]; const holder=Object.assign({}, {values}); holder.values=[]; a.push(1); a.push([...values]);", FileName: "case.js", Output: []string{"const a=[]; const values=[1]; const holder=Object.assign({}, {values}); holder.values=[]; a.push(1, [...values]);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 104, EndLine: 1, EndColumn: 108, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[]; const values=[1]; const copy=Object.assign([], values); copy[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1); a.push([...values]);", FileName: "case.js", Output: []string{"const a=[]; const values=[1]; const copy=Object.assign([], values); copy[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1, [...values]);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 137, EndLine: 1, EndColumn: 141, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[]; const values=[1]; const copy=Object.assign({}, values); copy[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1); a.push([...values]);", FileName: "case.js", Output: []string{"const a=[]; const values=[1]; const copy=Object.assign({}, values); copy[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1, [...values]);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 137, EndLine: 1, EndColumn: 141, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[]; const values=[1]; const [...copy]=values; copy[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1); a.push([...values]);", FileName: "case.js", Output: []string{"const a=[]; const values=[1]; const [...copy]=values; copy[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1, [...values]);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 123, EndLine: 1, EndColumn: 127, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[]; const values=[1]; const {...copy}=values; copy[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1); a.push([...values]);", FileName: "case.js", Output: []string{"const a=[]; const values=[1]; const {...copy}=values; copy[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1, [...values]);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 123, EndLine: 1, EndColumn: 127, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[]; const values=[1]; Object.defineProperty(values,Symbol.iterator,{value:function*(){a.push(0);yield 1;}}); a.push(1); a.push([...values]);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 131, EndLine: 1, EndColumn: 135, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; const values=[1]; Object.defineProperty(values,Symbol.iterator,{value:function*(){a.push(0);yield 1;}}); a.push(1, [...values]);"}}},
			}},
			{Code: "const a=[]; const values=[1]; const alias=true?values:[]; alias[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1); a.push([...values]);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 128, EndLine: 1, EndColumn: 132, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; const values=[1]; const alias=true?values:[]; alias[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1, [...values]);"}}},
			}},
			{Code: "const a=[]; const values=[1]; const holder={values}; holder.values[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1); a.push([...values]);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 131, EndLine: 1, EndColumn: 135, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; const values=[1]; const holder={values}; holder.values[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1, [...values]);"}}},
			}},
			{Code: "const a=[]; const values=[1]; const holder=[values]; holder[0][Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1); a.push([...values]);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 127, EndLine: 1, EndColumn: 131, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; const values=[1]; const holder=[values]; holder[0][Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1, [...values]);"}}},
			}},
			{Code: "const a=[]; const values=[1]; const holder={values}; holder.other=[]; holder.values[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1); a.push([...values]);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 148, EndLine: 1, EndColumn: 152, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; const values=[1]; const holder={values}; holder.other=[]; holder.values[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1, [...values]);"}}},
			}},
			{Code: "const a=[]; const values=[1]; const alias=[...values]; alias[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1); a.push([...values]);", FileName: "case.js", Output: []string{"const a=[]; const values=[1]; const alias=[...values]; alias[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1, [...values]);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 125, EndLine: 1, EndColumn: 129, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[]; const values=[1]; const alias={...values}; alias[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1); a.push([...values]);", FileName: "case.js", Output: []string{"const a=[]; const values=[1]; const alias={...values}; alias[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1, [...values]);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 125, EndLine: 1, EndColumn: 129, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[]; const values=[1]; (true?values:[])[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1); a.push([...values]);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 111, EndLine: 1, EndColumn: 115, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; const values=[1]; (true?values:[])[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1, [...values]);"}}},
			}},
			{Code: "const a=[]; const values=[1]; ({values}).values[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1); a.push([...values]);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 112, EndLine: 1, EndColumn: 116, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; const values=[1]; ({values}).values[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1, [...values]);"}}},
			}},
			{Code: "const a=[]; const values=[1]; Reflect.set({},Symbol.iterator,function*(){a.push(0);yield 1;},values); a.push(1); a.push([...values]);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 116, EndLine: 1, EndColumn: 120, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; const values=[1]; Reflect.set({},Symbol.iterator,function*(){a.push(0);yield 1;},values); a.push(1, [...values]);"}}},
			}},
		})
	})
	t.Run("Built-in iterator mutation", func(t *testing.T) {
		// Replacing a built-in iterator can mutate the receiver even when the
		// spread reads a literal. Keep these unsafe upstream fixes as suggestions.
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_single_call.PreferSingleCallRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
			{Code: "const a=[]; Array.prototype[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1); a.push([...[1]]);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 92, EndLine: 1, EndColumn: 96, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; Array.prototype[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1, [...[1]]);"}}},
			}},
			{Code: "const a=[]; const values=[1]; Array.prototype[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1); a.push([...values]);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 110, EndLine: 1, EndColumn: 114, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; const values=[1]; Array.prototype[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1, [...values]);"}}},
			}},
			{Code: "const a=[]; Object.defineProperty(Array.prototype,Symbol.iterator,{value:function*(){a.push(0);yield 1;}}); a.push(1); a.push([...[1]]);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 122, EndLine: 1, EndColumn: 126, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; Object.defineProperty(Array.prototype,Symbol.iterator,{value:function*(){a.push(0);yield 1;}}); a.push(1, [...[1]]);"}}},
			}},
			{Code: "const a=[]; const values=[1]; Object.assign(Array.prototype,{[Symbol.iterator]:function*(){a.push(0);yield 1;}}); a.push(1); a.push([...values]);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 128, EndLine: 1, EndColumn: 132, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; const values=[1]; Object.assign(Array.prototype,{[Symbol.iterator]:function*(){a.push(0);yield 1;}}); a.push(1, [...values]);"}}},
			}},
			{Code: "const a=[]; Reflect.set(Array.prototype,Symbol.iterator,function*(){a.push(0);yield 1;}); a.push(1); a.push([...[1]]);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 104, EndLine: 1, EndColumn: 108, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; Reflect.set(Array.prototype,Symbol.iterator,function*(){a.push(0);yield 1;}); a.push(1, [...[1]]);"}}},
			}},
			{Code: "const a=[]; const values=[1]; Reflect.defineProperty(Array.prototype,Symbol.iterator,{value:function*(){a.push(0);yield 1;}}); a.push(1); a.push([...values]);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 141, EndLine: 1, EndColumn: 145, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; const values=[1]; Reflect.defineProperty(Array.prototype,Symbol.iterator,{value:function*(){a.push(0);yield 1;}}); a.push(1, [...values]);"}}},
			}},
			{Code: "const a=[]; String.prototype[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1); a.push([...\"ab\"]);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 93, EndLine: 1, EndColumn: 97, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; String.prototype[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1, [...\"ab\"]);"}}},
			}},
			{Code: "const a=[]; const value=\"ab\"; Object.defineProperties(String.prototype,{[Symbol.iterator]:{value:function*(){a.push(0);yield 1;}}}); a.push(1); a.push([...value]);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 147, EndLine: 1, EndColumn: 151, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; const value=\"ab\"; Object.defineProperties(String.prototype,{[Symbol.iterator]:{value:function*(){a.push(0);yield 1;}}}); a.push(1, [...value]);"}}},
			}},
			{Code: "const a=[]; Reflect.set(String.prototype,Symbol.iterator,function*(){a.push(0);yield 1;}); a.push(1); a.push([...\"ab\"]);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 105, EndLine: 1, EndColumn: 109, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; Reflect.set(String.prototype,Symbol.iterator,function*(){a.push(0);yield 1;}); a.push(1, [...\"ab\"]);"}}},
			}},
			{Code: "const a=[]; Reflect.set({},Symbol.iterator,function*(){a.push(0);yield 1;},String.prototype); a.push(1); a.push([...\"ab\"]);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 108, EndLine: 1, EndColumn: 112, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; Reflect.set({},Symbol.iterator,function*(){a.push(0);yield 1;},String.prototype); a.push(1, [...\"ab\"]);"}}},
			}},
			{Code: "Array.prototype[Symbol.iterator]=function*(){element.classList.add(\"changed\");yield \"b\";}; element.classList.add(\"a\"); element.classList.add(...[\"b\"]);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 1, Column: 138, EndLine: 1, EndColumn: 141, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Array.prototype[Symbol.iterator]=function*(){element.classList.add(\"changed\");yield \"b\";}; element.classList.add(\"a\", ...[\"b\"]);"}}},
			}},
			{Code: "String.prototype[Symbol.iterator]=function*(){importScripts(\"changed\");yield \"b\";}; importScripts(\"a\"); importScripts(...\"b\");", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `importScripts()` multiple times.", Line: 1, Column: 105, EndLine: 1, EndColumn: 118, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "String.prototype[Symbol.iterator]=function*(){importScripts(\"changed\");yield \"b\";}; importScripts(\"a\", ...\"b\");"}}},
			}},
			{Code: "const a=[]; const {prototype: proto}=Array; proto[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1); a.push([...[1]]);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 114, EndLine: 1, EndColumn: 118, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; const {prototype: proto}=Array; proto[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1, [...[1]]);"}}},
			}},
			{Code: "const a=[]; const {prototype: proto}=String; proto[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1); a.push([...\"ab\"]);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 115, EndLine: 1, EndColumn: 119, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; const {prototype: proto}=String; proto[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1, [...\"ab\"]);"}}},
			}},
			{Code: "const a=[]; Object.defineProperty(...[Array.prototype,Symbol.iterator,{value:function*(){a.push(0);yield 1;}}]); a.push(1); a.push([...[1]]);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 127, EndLine: 1, EndColumn: 131, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; Object.defineProperty(...[Array.prototype,Symbol.iterator,{value:function*(){a.push(0);yield 1;}}]); a.push(1, [...[1]]);"}}},
			}},
			{Code: "const a=[]; Reflect.set(...[{},Symbol.iterator,function*(){a.push(0);yield 1;},String.prototype]); a.push(1); a.push([...\"ab\"]);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 113, EndLine: 1, EndColumn: 117, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; Reflect.set(...[{},Symbol.iterator,function*(){a.push(0);yield 1;},String.prototype]); a.push(1, [...\"ab\"]);"}}},
			}},
			{Code: "const a=[]; const Array={prototype:{}}; Array.prototype[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1); a.push([...[1]]);", FileName: "case.js", Output: []string{"const a=[]; const Array={prototype:{}}; Array.prototype[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1, [...[1]]);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 120, EndLine: 1, EndColumn: 124, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[]; const String={prototype:{}}; String.prototype[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1); a.push([...\"ab\"]);", FileName: "case.js", Output: []string{"const a=[]; const String={prototype:{}}; String.prototype[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1, [...\"ab\"]);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 122, EndLine: 1, EndColumn: 126, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[]; const Object={defineProperty(){}}; Object.defineProperty(Array.prototype,Symbol.iterator,{value:function*(){a.push(0);yield 1;}}); a.push(1); a.push([...[1]]);", FileName: "case.js", Output: []string{"const a=[]; const Object={defineProperty(){}}; Object.defineProperty(Array.prototype,Symbol.iterator,{value:function*(){a.push(0);yield 1;}}); a.push(1, [...[1]]);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 157, EndLine: 1, EndColumn: 161, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[]; const Reflect={set(){}}; Reflect.set(String.prototype,Symbol.iterator,function*(){a.push(0);yield 1;}); a.push(1); a.push([...\"ab\"]);", FileName: "case.js", Output: []string{"const a=[]; const Reflect={set(){}}; Reflect.set(String.prototype,Symbol.iterator,function*(){a.push(0);yield 1;}); a.push(1, [...\"ab\"]);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 130, EndLine: 1, EndColumn: 134, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[]; Array.prototype[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1); a.push([...\"ab\"]);", FileName: "case.js", Output: []string{"const a=[]; Array.prototype[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1, [...\"ab\"]);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 92, EndLine: 1, EndColumn: 96, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[]; String.prototype[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1); a.push([...[1]]);", FileName: "case.js", Output: []string{"const a=[]; String.prototype[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push(1, [...[1]]);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 93, EndLine: 1, EndColumn: 97, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
		})
	})
	t.Run("Decorator evaluation", func(t *testing.T) {
		// Applying a decorator invokes it even when its expression is only an
		// identifier. Keep these merges as suggestions instead of upstream's
		// unsafe autofixes; decorated parameters run during class definition too.
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_single_call.PreferSingleCallRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
			{Code: "element.classList.add(a); element.classList.add(class { @decorate method() {} });", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 1, Column: 45, EndLine: 1, EndColumn: 48, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "element.classList.add(a, class { @decorate method() {} });"}}},
			}},
			{Code: "element.classList.add(a); element.classList.add(class { @decorate value; });", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 1, Column: 45, EndLine: 1, EndColumn: 48, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "element.classList.add(a, class { @decorate value; });"}}},
			}},
			{Code: "element.classList.add(a); element.classList.add(class { static { @decorate class C {} } });", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 1, Column: 45, EndLine: 1, EndColumn: 48, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "element.classList.add(a, class { static { @decorate class C {} } });"}}},
			}},
			{Code: "const a=[]; function decorate(){a.push(0);} a.push(1); a.push(void class { @decorate method() {} });", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 58, EndLine: 1, EndColumn: 62, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; function decorate(){a.push(0);} a.push(1, void class { @decorate method() {} });"}}},
			}},
			{Code: "const a=[]; function decorate(){a.push(0);} a.push(1); a.push(void class { static { class C { method(@decorate arg) {} } } });", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 58, EndLine: 1, EndColumn: 62, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; function decorate(){a.push(0);} a.push(1, void class { static { class C { method(@decorate arg) {} } } });"}}},
			}},
			{Code: "const a=[]; function decorate(){a.push(0);} a.push(1); a.push(void class { static { class C { constructor(@decorate arg) {} } } });", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 58, EndLine: 1, EndColumn: 62, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; function decorate(){a.push(0);} a.push(1, void class { static { class C { constructor(@decorate arg) {} } } });"}}},
			}},
			{Code: "element.classList.add(a); element.classList.add(class { constructor(value=sideEffect()) { sideEffect(); } method(value=sideEffect()) { sideEffect(); } });", FileName: "case.ts", Output: []string{"element.classList.add(a, class { constructor(value=sideEffect()) { sideEffect(); } method(value=sideEffect()) { sideEffect(); } });"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 1, Column: 45, EndLine: 1, EndColumn: 48, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
		})
	})
	t.Run("Tagged template evaluation", func(t *testing.T) {
		// A user tag is a function call, including inside a computed class key
		// or void expression. Upstream misses these effects. Preserve existing
		// fixes for the built-in String.raw while checking its substitutions.
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_single_call.PreferSingleCallRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
			{Code: "const a=[]; function tag(){a.push(0);} a.push(1); a.push(void tag``);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 53, EndLine: 1, EndColumn: 57, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; function tag(){a.push(0);} a.push(1, void tag``);"}}},
			}},
			{Code: "element.classList.add(a); element.classList.add(class { [tag``]() {} });", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 1, Column: 45, EndLine: 1, EndColumn: 48, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "element.classList.add(a, class { [tag``]() {} });"}}},
			}},
			{Code: "importScripts(a); importScripts(tag``);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `importScripts()` multiple times.", Line: 1, Column: 19, EndLine: 1, EndColumn: 32, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "importScripts(a, tag``);"}}},
			}},
			{Code: "const a=[]; a.push(1); a.push(String.raw`x`);", FileName: "case.js", Output: []string{"const a=[]; a.push(1, String.raw`x`);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 26, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[]; a.push(1); a.push([String.raw`x`]);", FileName: "case.js", Output: []string{"const a=[]; a.push(1, [String.raw`x`]);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 26, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "element.classList.add(a); element.classList.add(String.raw`x`);", FileName: "case.js", Output: []string{"element.classList.add(a, String.raw`x`);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 1, Column: 45, EndLine: 1, EndColumn: 48, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "importScripts(a); importScripts(String.raw`x`);", FileName: "case.js", Output: []string{"importScripts(a, String.raw`x`);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `importScripts()` multiple times.", Line: 1, Column: 19, EndLine: 1, EndColumn: 32, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[]; a.push(1); a.push(String.raw`x${sideEffect()}`);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 26, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; a.push(1, String.raw`x${sideEffect()}`);"}}},
			}},
			{Code: "element.classList.add(a); element.classList.add(String.raw`x${sideEffect()}`);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 1, Column: 45, EndLine: 1, EndColumn: 48, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "element.classList.add(a, String.raw`x${sideEffect()}`);"}}},
			}},
			{Code: "function f(String) { element.classList.add(a); element.classList.add(String.raw`x`); }", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 1, Column: 66, EndLine: 1, EndColumn: 69, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "function f(String) { element.classList.add(a, String.raw`x`); }"}}},
			}},
			{Code: "function f(String) { const a=[]; a.push(1); a.push(String.raw`x`); }", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 47, EndLine: 1, EndColumn: 51, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "function f(String) { const a=[]; a.push(1, String.raw`x`); }"}}},
			}},
			{Code: "const StringAlias=String; element.classList.add(a); element.classList.add(StringAlias.raw`x`);", FileName: "case.js", Output: []string{"const StringAlias=String; element.classList.add(a, StringAlias.raw`x`);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 1, Column: 71, EndLine: 1, EndColumn: 74, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
		})
	})
	t.Run("Argument safety", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_single_call.PreferSingleCallRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
			{Code: "const a=[]; a.push([...[1]]); a.push(2);", FileName: "case.js", Output: []string{"const a=[]; a.push([...[1]], 2);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 33, EndLine: 1, EndColumn: 37, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[]; a.unshift([0,...[1,...[2,3]],4]); a.unshift(5);", FileName: "case.js", Output: []string{"const a=[]; a.unshift(5, [0,...[1,...[2,3]],4]);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 1, Column: 49, EndLine: 1, EndColumn: 56, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[]; a.push([...[,,1]]); a.push(2);", FileName: "case.js", Output: []string{"const a=[]; a.push([...[,,1]], 2);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 35, EndLine: 1, EndColumn: 39, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[]; a.push([...[],...[],]); a.push(2);", FileName: "case.js", Output: []string{"const a=[]; a.push([...[],...[],], 2);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 39, EndLine: 1, EndColumn: 43, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[]; const values=[1,2]; a.push([...values]); a.push(2);", FileName: "case.js", Output: []string{"const a=[]; const values=[1,2]; a.push([...values], 2);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 56, EndLine: 1, EndColumn: 60, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[]; a.push([...\"😀a\"]); a.push(2);", FileName: "case.js", Output: []string{"const a=[]; a.push([...\"😀a\"], 2);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 35, EndLine: 1, EndColumn: 39, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[]; a.push([...\"\\uD800\"]); a.push(2);", FileName: "case.js", Output: []string{"const a=[]; a.push([...\"\\uD800\"], 2);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 38, EndLine: 1, EndColumn: 42, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const a=[]; a.push([...getValues()]); a.push(2);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 41, EndLine: 1, EndColumn: 45, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; a.push([...getValues()], 2);"}}},
			}},
			{Code: "const a=[]; a.push([...unknown]); a.push(2);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 37, EndLine: 1, EndColumn: 41, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; a.push([...unknown], 2);"}}},
			}},
			{Code: "const a=[]; a.push([...1]); a.push(2);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 31, EndLine: 1, EndColumn: 35, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; a.push([...1], 2);"}}},
			}},
			{Code: "const a=[]; const values=[1]; values[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push([...values]); a.push(2);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 111, EndLine: 1, EndColumn: 115, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; const values=[1]; values[Symbol.iterator]=function*(){a.push(0);yield 1;}; a.push([...values], 2);"}}},
			}},
			{Code: "const a=[]; a.push([...{[Symbol.iterator](){throw 1;}}]); a.push(2);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 61, EndLine: 1, EndColumn: 65, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; a.push([...{[Symbol.iterator](){throw 1;}}], 2);"}}},
			}},
			{Code: "const a=[]; a.push([...[1]].length); a.push(2);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 40, EndLine: 1, EndColumn: 44, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; a.push([...[1]].length, 2);"}}},
			}},
			{Code: "const a=[]; a.push([...[1]][0]); a.push(2);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 36, EndLine: 1, EndColumn: 40, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const a=[]; a.push([...[1]][0], 2);"}}},
			}},
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
