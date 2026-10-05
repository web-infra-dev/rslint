package no_new_array_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_new_array"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Additional AST, edit safety, and static evaluation cases checked against v77.0.0.
func TestNoNewArrayExpressions(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_new_array.NoNewArrayRule, []rule_tester.ValidTestCase{
		{Code: "new OtherArray('x');", FileName: "file.js"},
		{Code: "new globalThis.Array('x');", FileName: "file.js"},
		{Code: "new holder['Array']('x');", FileName: "file.js"},
		{Code: "class C { #Array; make() { return new this.#Array('x'); } }", FileName: "file.js"},
		{Code: "Array?.('x');", FileName: "file.js"},
		{Code: "new Array(...items, 'x');", FileName: "file.js"},
		{Code: "new (Array as ArrayConstructor)('x');", FileName: "file.ts"},
		{Code: "new Array!('x');", FileName: "file.ts"},
		{Code: "type T = Array<string>; const view = <Array value='x' />;", FileName: "file.tsx"},
	},
		[]rule_tester.InvalidTestCase{
			arrayCase("new \\u0041\\u0072\\u0072\\u0061\\u0079('x');", "file.js", 1, 1, 1, 40, "['x'];", ""),
			arrayCase("new ((Array))((('x')));", "file.js", 1, 1, 1, 23, "[('x')];", ""),
			arrayCase("new Array((0, 'x'));", "file.js", 1, 1, 1, 20, "[(0, 'x')];", ""),
			arrayCase("new Array(...(items));", "file.js", 1, 1, 1, 22, "", "[...(items)];"),
			arrayCase("new Array(true);", "file.js", 1, 1, 1, 16, "[true];", ""),
			arrayCase("new Array(1n);", "file.js", 1, 1, 1, 14, "[1n];", ""),
			arrayCase("new Array(undefined);", "file.js", 1, 1, 1, 21, "[undefined];", ""),
			arrayCase("new Array(void 0);", "file.js", 1, 1, 1, 18, "[void 0];", ""),
			arrayCase("new Array(/x/);", "file.js", 1, 1, 1, 15, "[/x/];", ""),
			arrayCase("new Array({});", "file.js", 1, 1, 1, 14, "[{}];", ""),
			arrayCase("new Array([]);", "file.js", 1, 1, 1, 14, "[[]];", ""),
			arrayCase("new Array(`text`);", "file.js", 1, 1, 1, 18, "[`text`];", ""),
			arrayCase("const value = 'x'; new Array(value);", "file.js", 1, 20, 1, 36, "const value = 'x'; [value];", ""),
			arrayCase("let value = 'x'; new Array(value);", "file.js", 1, 18, 1, 34, "", ""),
			arrayCase("new Array(value); const value = 'x';", "file.js", 1, 1, 1, 17, "", ""),
			arrayCase("const x = new Array(x);", "file.js", 1, 11, 1, 23, "", ""),
			arrayCase("const x = y; const y = 'x'; new Array(x);", "file.js", 1, 29, 1, 41, "", ""),
			arrayCase("const f = () => new Array(x); const x = 'x'; f();", "file.js", 1, 17, 1, 29, "", ""),
			arrayCase("const { x } = { x: 'x' }; new Array(x);", "file.js", 1, 27, 1, 39, "", ""),
			arrayCase("function f(Array, undefined) { return new Array(undefined); }", "file.js", 1, 39, 1, 59, "", ""),
			arrayCase("function f(Array) { return new Array('x'); }", "file.js", 1, 28, 1, 42, "function f(Array) { return ['x']; }", ""),
			arrayCase("new Array(true ? 'x' : unknown);", "file.js", 1, 1, 1, 32, "[true ? 'x' : unknown];", ""),
			arrayCase("new Array(false || 'x');", "file.js", 1, 1, 1, 24, "[false || 'x'];", ""),
			arrayCase("new Array(null ?? 'x');", "file.js", 1, 1, 1, 23, "[null ?? 'x'];", ""),
			arrayCase("new Array(({value: 'x'}).value);", "file.js", 1, 1, 1, 32, "[({value: 'x'}).value];", ""),
			arrayCase("new Array(({value: 'x'})['value']);", "file.js", 1, 1, 1, 35, "[({value: 'x'})['value']];", ""),
			arrayCase("const object = {value: 'x'}; new Array(object.value);", "file.js", 1, 30, 1, 53, "", ""),
			// The shared evaluator does not represent built-in objects, functions,
			// or symbols. Retain diagnostics without upstream's fixes for these
			// values and their typeof expressions, as documented for this rule.
			arrayCase("new Array(Symbol.iterator);", "file.js", 1, 1, 1, 27, "", ""),
			arrayCase("new Array(Array);", "file.js", 1, 1, 1, 17, "", ""),
			arrayCase("new Array(Math);", "file.js", 1, 1, 1, 16, "", ""),
			arrayCase("new Array(typeof Array);", "file.js", 1, 1, 1, 24, "", ""),
			arrayCase("new Array(typeof Math);", "file.js", 1, 1, 1, 23, "", ""),
			arrayCase("new Array(typeof Symbol.iterator);", "file.js", 1, 1, 1, 34, "", ""),
			arrayCase("new Array(Object.freeze(['x']));", "file.js", 1, 1, 1, 32, "[Object.freeze(['x'])];", ""),
			arrayCase("new Array({get x() { return 1; }});", "file.js", 1, 1, 1, 35, "", ""),
			arrayCase("new Array([unknown]);", "file.js", 1, 1, 1, 21, "", ""),
			arrayCase("new Array(void sideEffect());", "file.js", 1, 1, 1, 29, "", ""),
			arrayCase("new Array((sideEffect(), 'x'));", "file.js", 1, 1, 1, 31, "", ""),
			arrayCase("let value; new Array(value = 'x');", "file.js", 1, 12, 1, 34, "", ""),
			arrayCase("new Array(/* keep */ ('x'));", "file.js", 1, 1, 1, 28, "", ""),
			arrayCase("new Array('x' /* keep */);", "file.js", 1, 1, 1, 26, "", ""),
			arrayCase("new Array(...(/* keep */ items));", "file.js", 1, 1, 1, 33, "", ""),
			arrayCase("/* before */ new Array('/* text */'); // after", "file.js", 1, 14, 1, 37, "/* before */ ['/* text */']; // after", ""),
			arrayCase("'😀'; new Array(\n  '值',\n);", "file.js", 1, 7, 3, 2, "'😀'; ['值'];", ""),
			arrayCase("function f() { return new\n Array('x'); }", "file.js", 1, 23, 2, 12, "function f() { return ['x']; }", ""),
			arrayCase("const before = 1\nnew Array('x')[0];", "file.js", 2, 1, 2, 15, "const before = 1\n;['x'][0];", ""),
			arrayCase("const before = 1\n(new Array('x'))[0];", "file.js", 2, 2, 2, 16, "const before = 1\n(['x'])[0];", ""),
			arrayCase("if (ready) new Array('x');", "file.js", 1, 12, 1, 26, "if (ready) ['x'];", ""),
			arrayCase("do new Array('x'); while (ready);", "file.js", 1, 4, 1, 18, "do ['x']; while (ready);", ""),
			arrayCase("label: new Array('x');", "file.js", 1, 8, 1, 22, "label: ['x'];", ""),
			arrayCase("while (ready) new Array(...items);", "file.js", 1, 15, 1, 34, "", "while (ready) [...items];"),
			arrayCase("function f() {}\nnew Array('x');", "file.js", 2, 1, 2, 15, "function f() {}\n['x'];", ""),
			arrayCase("for (const item of new Array('x')) {}", "file.js", 1, 20, 1, 34, "for (const item of ['x']) {}", ""),
			arrayCase("new Array<string>('x');", "file.ts", 1, 1, 1, 23, "['x'];", ""),
			arrayCase("new Array(('x' as const));", "file.ts", 1, 1, 1, 26, "[('x' as const)];", ""),
			arrayCase("new Array(('x')!);", "file.ts", 1, 1, 1, 18, "[('x')!];", ""),
			arrayCase("new Array((1 as unknown as string));", "file.ts", 1, 1, 1, 36, "", ""),
			arrayCase("new Array('x' satisfies string);", "file.ts", 1, 1, 1, 32, "['x' satisfies string];", ""),
			arrayCase("new Array((<string>'x'));", "file.ts", 1, 1, 1, 25, "[(<string>'x')];", ""),
			arrayCase("const view = <div>{new Array('x')}</div>;", "file.tsx", 1, 20, 1, 34, "const view = <div>{['x']}</div>;", ""),
			arrayCase("new Array(/** @type {string} */ ('x'));", "file.js", 1, 1, 1, 39, "", ""),
			arrayCase("new (/** @type {ArrayConstructor} */ (Array))('x');", "file.js", 1, 1, 1, 51, "", ""),
			{
				Code: "new Array(new Array('x'));", FileName: "file.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: errorMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
					{MessageId: "error", Message: errorMessage, Line: 1, Column: 11, EndLine: 1, EndColumn: 25},
				},
				Output: []string{"new Array(['x']);", "[['x']];"},
			},
		})
}

func TestNoNewArrayGlobals(t *testing.T) {
	var invalid []rule_tester.InvalidTestCase
	for _, fixture := range []struct {
		code, name, access, output string
		column, endColumn          int
	}{
		{"new Array(undefined);", "undefined", "off", "", 1, 21},
		{"new Array(Object.freeze([1]));", "Object", "off", "", 1, 30},
		{"new Array(typeof NaN);", "NaN", "off", "", 1, 22},
		{"new Array(undefined);", "undefined", "writable", "[undefined];", 1, 21},
		{"new Array('x');", "Array", "off", "['x'];", 1, 15},
		{"/* global undefined: off */ new Array(undefined);", "", "", "", 29, 49},
		{"/* global Object: off */ new Array(Object.freeze([1]));", "", "", "", 26, 55},
	} {
		test := arrayCase(fixture.code, "file.js", 1, fixture.column, 1, fixture.endColumn, fixture.output, "")
		if fixture.name != "" {
			test.Globals = map[string]any{fixture.name: fixture.access}
		}
		invalid = append(invalid, test)
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_new_array.NoNewArrayRule, nil, invalid)
}

func TestNoNewArrayEditDemand(t *testing.T) {
	const source = "new Array('x');\nnew Array(...items);\nnew Array(3);"
	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(source, "edit-demand.js", "tsconfig.json")
	if err != nil {
		t.Fatal(err)
	}
	results := make(map[rule.EditDemand][]rule.RuleDiagnostic)
	for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program: lintprogram.NewFromCompiler(program),
			File:    sourceFile.FileName(),
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{
					Name:     no_new_array.NoNewArrayRule.Name,
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return no_new_array.NoNewArrayRule.Run(ctx, nil)
					},
				}}
			},
			Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(diagnostic rule.RuleDiagnostic) {
				results[demand] = append(results[demand], diagnostic)
			}},
		})
		if len(results[demand]) != 3 {
			t.Fatalf("demand %d: got %d diagnostics, want 3", demand, len(results[demand]))
		}
	}

	all := results[rule.EditDemandAll]
	for demand, diagnostics := range results {
		for i, diagnostic := range diagnostics {
			if diagnostic.Range != all[i].Range || !reflect.DeepEqual(diagnostic.Message, all[i].Message) {
				t.Errorf("demand %d changed diagnostic %d", demand, i)
			}
			if demand == rule.EditDemandAutofix || demand == rule.EditDemandAll {
				if !reflect.DeepEqual(diagnostic.FixesPtr, all[i].FixesPtr) {
					t.Errorf("demand %d changed fixes for diagnostic %d", demand, i)
				}
			} else if diagnostic.FixesPtr != nil {
				t.Errorf("demand %d produced unexpected fixes", demand)
			}
			if demand == rule.EditDemandSuggestion || demand == rule.EditDemandAll {
				if !reflect.DeepEqual(diagnostic.Suggestions, all[i].Suggestions) {
					t.Errorf("demand %d changed suggestions for diagnostic %d", demand, i)
				}
			} else if diagnostic.Suggestions != nil {
				t.Errorf("demand %d produced unexpected suggestions", demand)
			}
		}
	}
	output, _, fixed := linter.ApplyRuleFixes(source, all)
	if !fixed || output != "['x'];\nnew Array(...items);\nnew Array(3);" {
		t.Fatalf("unexpected autofix output: %q", output)
	}
	if all[0].Suggestions != nil || all[1].FixesPtr != nil || all[2].FixesPtr != nil || all[2].Suggestions != nil {
		t.Fatal("unexpected edit category")
	}
	if all[1].Suggestions == nil || len(*all[1].Suggestions) != 1 {
		t.Fatal("expected one spread suggestion")
	}
	suggestion := (*all[1].Suggestions)[0]
	if suggestion.Message.Id != "spread" || suggestion.Message.Description != "Spread the argument." {
		t.Fatalf("unexpected suggestion message: %+v", suggestion.Message)
	}
	output, _, fixed = linter.ApplyRuleFixes(source, []rule.RuleSuggestion{suggestion})
	if !fixed || output != "new Array('x');\n[...items];\nnew Array(3);" {
		t.Fatalf("unexpected suggestion output: %q", output)
	}
}
