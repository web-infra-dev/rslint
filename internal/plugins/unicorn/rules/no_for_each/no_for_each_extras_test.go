package no_for_each_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_for_each"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestNoForEachExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_for_each.NoForEachRule,
		[]rule_tester.ValidTestCase{
			valid(`foo['forEach'](element => bar(element))`, "file.js"),
			valid("foo[`forEach`](element => bar(element))", "file.js"),
			valid(`foo.#forEach(element => bar(element))`, "file.ts"),
			valid(`R.forEach(values, callback)`, "file.js"),
		},
		[]rule_tester.InvalidTestCase{
			invalid(`foo?.bar.forEach(element => bar(element));`, "file.js", ""),
			invalid(`(foo || bar).forEach(element => log(element));`, "file.js", ""),
			invalid(`React.NotChildren.forEach(callback)`, "file.js", ""),
			invalid(`NotReact.Children.forEach(callback)`, "file.js", ""),
			invalid(`const array = []; array?.forEach(element => bar(element));`, "file.js", `const array = []; if (array) for (const element of array) bar(element);`),
			{
				Code:     `[foo()]?.forEach(element => bar(element));`,
				FileName: "file.js",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: errorMessageID,
					Line:      1,
					Column:    10,
					EndLine:   1,
					EndColumn: 17,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: suggestionMessageID,
						Output:    `if ([foo()]) for (const element of [foo()]) bar(element);`,
					}},
				}},
			},
			{
				Code:     `const arrays = [[1]]; arrays.forEach(array => array.forEach(value => use(value)));`,
				FileName: "file.js",
				Output:   []string{`const arrays = [[1]]; for (const array of arrays) array.forEach(value => use(value));`},
				Errors: []rule_tester.InvalidTestCaseError{
					forEachError(`const arrays = [[1]]; arrays.forEach(array => array.forEach(value => use(value)));`, 1),
					forEachError(`const arrays = [[1]]; arrays.forEach(array => array.forEach(value => use(value)));`, 0),
				},
			},
			invalid(`const array = [1]; array.forEach(element => { element++; });`, "file.js", `const array = [1]; for (let element of array) { element++; }`),
			invalid(`const array = [1]; array.forEach(({value}) => use(value));`, "file.js", `const array = [1]; for (const {value} of array) use(value);`),
			invalid(`const array = [1]; array.forEach((element, {index}) => use(element, index));`, "file.js", ""),
			invalid(`const array = [1]; array.forEach(function (element) { use(element); });`, "file.js", `const array = [1]; for (const element of array) { use(element); }`),
			invalid(`const array = [1]; array.forEach(function (element) { use(this, element); });`, "file.js", ""),
			invalid(`const array = [1]; array.forEach(function (element) { use(arguments, element); });`, "file.js", ""),
			invalid(`const array = [1]; array.forEach(function (element) { const capture = () => this; use(capture, element); });`, "file.js", ""),
			invalid(`const array = [1]; array.forEach(function named(element) { use(element); });`, "file.js", ""),
			invalid(`const array = [1]; array.forEach(element => { return; });`, "file.js", `const array = [1]; for (const element of array) { continue; }`),
			invalid(`const array = [1]; array.forEach(element => { const change = () => { element = 2; }; change(); });`, "file.js", `const array = [1]; for (let element of array) { const change = () => { element = 2; }; change(); }`),
			invalid(`const array = [1]; array.forEach(element => { function change(element) { element = 2; } change(element); });`, "file.js", `const array = [1]; for (const element of array) { function change(element) { element = 2; } change(element); }`),
		},
	)
}

func TestNoForEachEditDemand(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		wantFix        bool
		wantSuggestion bool
	}{
		{name: "autofix", code: `const array = []; array.forEach(value => use(value));`, wantFix: true},
		{name: "suggestion", code: `[getValue()]?.forEach(value => use(value));`, wantSuggestion: true},
		{name: "diagnostic only", code: `value.forEach(callback);`},
	}

	for index, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			program, sourceFile := createNoForEachProgram(t, string(rune('a'+index))+".js", test.code)
			diagnostics := map[rule.EditDemand]rule.RuleDiagnostic{}
			for _, demand := range []rule.EditDemand{
				rule.EditDemandNone,
				rule.EditDemandAutofix,
				rule.EditDemandSuggestion,
				rule.EditDemandAll,
			} {
				got := lintNoForEachWithDemand(program, sourceFile, demand)
				if len(got) != 1 {
					t.Fatalf("demand %d: diagnostics = %d, want 1", demand, len(got))
				}
				diagnostics[demand] = got[0]
			}

			base := diagnostics[rule.EditDemandNone]
			if base.FixesPtr != nil || base.Suggestions != nil {
				t.Fatal("EditDemandNone materialized edits")
			}
			for demand, diagnostic := range diagnostics {
				if diagnostic.Range != base.Range ||
					diagnostic.Message.Id != base.Message.Id ||
					diagnostic.Message.Description != base.Message.Description {
					t.Errorf("demand %d changed diagnostic identity", demand)
				}
			}

			if diagnostics[rule.EditDemandSuggestion].FixesPtr != nil {
				t.Error("suggestion-only demand materialized an autofix")
			}
			autofixOnly := diagnostics[rule.EditDemandAutofix].FixesPtr
			allFixes := diagnostics[rule.EditDemandAll].FixesPtr
			if test.wantFix {
				if autofixOnly == nil || allFixes == nil || !reflect.DeepEqual(*autofixOnly, *allFixes) {
					t.Fatal("autofix artifacts differ between autofix-only and all demand")
				}
			} else if autofixOnly != nil {
				t.Errorf("unexpected autofix: %#v", *autofixOnly)
			}

			if diagnostics[rule.EditDemandAutofix].Suggestions != nil {
				t.Error("autofix-only demand materialized a suggestion")
			}
			suggestionOnly := diagnostics[rule.EditDemandSuggestion].Suggestions
			allSuggestions := diagnostics[rule.EditDemandAll].Suggestions
			if test.wantSuggestion {
				if suggestionOnly == nil || allSuggestions == nil || !reflect.DeepEqual(*suggestionOnly, *allSuggestions) {
					t.Fatal("suggestion artifacts differ between suggestion-only and all demand")
				}
			} else if suggestionOnly != nil {
				t.Errorf("unexpected suggestion: %#v", *suggestionOnly)
			}
		})
	}
}

func createNoForEachProgram(t testing.TB, fileName, code string) (*compiler.Program, *ast.SourceFile) {
	t.Helper()
	rootDir := fixtures.GetRootDir()
	fs := utils.NewOverlayVFS(rootDir.FS, map[string]string{
		tspath.ResolvePath(rootDir.Dir, fileName): code,
	})
	host := utils.CreateCompilerHost(rootDir.Dir, fs)
	program, err := utils.CreateProgram(true, fs, rootDir.Dir, "tsconfig.json", host)
	if err != nil {
		t.Fatalf("failed to create program: %v", err)
	}
	sourceFile := program.GetSourceFile(fileName)
	if sourceFile == nil {
		t.Fatalf("source file %q not found", fileName)
	}
	return program, sourceFile
}

func lintNoForEachWithDemand(program *compiler.Program, sourceFile *ast.SourceFile, demand rule.EditDemand) []rule.RuleDiagnostic {
	var diagnostics []rule.RuleDiagnostic
	linter.LintSingleFile(linter.LintSingleFileOptions{
		Program:     lintprogram.NewFromCompiler(program),
		File:        sourceFile.FileName(),
		HasTypeInfo: true,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
				Name:     no_for_each.NoForEachRule.Name,
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return no_for_each.NoForEachRule.Run(ctx, nil)
				},
			}}
		},
		Consumer: rule.DiagnosticConsumer{
			Demand: demand,
			Report: func(diagnostic rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, diagnostic)
			},
		},
	})
	return diagnostics
}
