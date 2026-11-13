package no_unnecessary_splice

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestNoUnnecessarySpliceExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessarySpliceRule, []rule_tester.ValidTestCase{
		// A leading spread can supply the delete count or shift every argument.
		{Code: "const array = [1, 2]; array.splice(...[0, 1], 0);", FileName: "input.js"},
		{Code: "array.splice(0, ...counts, item);", FileName: "input.js"},
		{Code: "array.splice(-0, 1); array.splice(index, 1 - 1);", FileName: "input.js"},
		{Code: "array.splice(0, array['length']); array.splice(0, array?.length);", FileName: "input.js"},
		{Code: "array.splice(0, (array?.items).length); (array?.splice)(0);", FileName: "input.js"},
		{Code: "array.items.splice((array?.items).length - 1, 1);", FileName: "input.js"},
		{Code: "array[key.name].splice(array[key?.name].length - 1, 1);", FileName: "input.js"},
		{Code: "object[(getKey?.() as string)].splice(index, 0);", FileName: "input.ts"},
		{Code: "(array.splice as any)(0, 1);", FileName: "input.ts"},
		{Code: "new array.splice(0, 1); splice(0, 1);", FileName: "input.js"},
		{Code: "class C { #splice() {} f() { this.#splice(0, 1); } }", FileName: "input.js"},
		{Code: "const result = (value, (array.splice(0)));", FileName: "input.js"},
		{Code: "const result = (array.splice(0) as unknown, other);", FileName: "input.ts"},
		{Code: "for (array.splice(0); condition; array.splice(0)) {}", FileName: "input.js"},
		{Code: "const custom = { splice() {} }; custom.splice();", FileName: "input.js"},
	}, []rule_tester.InvalidTestCase{
		spliceCase("input.js", "(array.splice(index, -0));", "no-op", 1, 8, ""),
		spliceCase("input.js", "array.splice(index, -1e400);", "no-op", 1, 7, ""),
		spliceCase("input.js", "/* keep */ array.splice(index, 0);", "no-op", 1, 18, "/* keep */ "),
		spliceCase("input.js", "array.splice(index, 0), f();", "no-op", 1, 7),
		spliceCase("input.js", "class C { static { a.splice(); } }", "no-op", 1, 22),
		spliceCase("input.js", "consume()\narray.splice(index, 0);\n[1].map(String)", "no-op", 2, 7, "consume()\n;\n[1].map(String)"),
		spliceCase("input.js", "before();\r\n  array.splice();\r\nafter();", "no-op", 2, 9, "before();\r\nafter();"),
		spliceCase("input.js", "/* 😀 */ array\n  .splice((0x0), +0b1,);", "shift", 2, 4, "/* 😀 */ array.shift();"),
		spliceCase("input.js", "/* 😀 */ array.splice(0, 1);", "shift", 1, 16, "/* 😀 */ array.shift();"),
		spliceCase("input.js", `box["it" + "ems"].splice(box.items.length - 1, 1);`, "pop", 1, 19, `box["it" + "ems"].pop();`),
		spliceCase("input.js", "class C { #items = []; f() { this.#items.splice(this.#items.length - 1, 1); } }", "pop", 1, 42, "class C { #items = []; f() { this.#items.pop(); } }"),
		spliceCase("input.js", "const result = ((array.splice(0), value), other);", "empty", 1, 24, "const result = ((array.length = 0, value), other);"),
		spliceCase("input.tsx", "const element = <div>{(array.splice(0), value)}</div>;", "empty", 1, 30, "const element = <div>{(array.length = 0, value)}</div>;"),
		spliceCase("input.js", "do(array.splice)(0);while (ok);", "empty", 1, 10, "do array.length = 0;while (ok);"),
		spliceCase("input.js", "(/** @type {number[]} */ (a)).splice(0);", "empty", 1, 31),
		spliceCase("input.ts", "function f(a: any) { a.splice<number>(0, 1); }", "shift", 1, 24, "function f(a: any) { a.shift(); }"),
		spliceCase("input.js", "const list = [1, 2]; list.splice(0n, 0);", "no-op", 1, 27),
		spliceCase("input.js", "const list = [1, 2]; list.splice({valueOf() { list.push(1); return 0; }}, 0);", "no-op", 1, 27),
		spliceCase("input.js", "const array = [1, 2]; const obj = {get array() { sideEffect(); return array; }}; obj.array.splice(index, 0);", "no-op", 1, 92),
		spliceCase("input.js", "const array = [1, 2]; const item = {get value() { array.push(3); return 1; }}; array.splice(array.length, 0, item.value);", "push", 1, 86),
		spliceCase("input.js", "const box = {get items() { return []; }}; box.items.splice(box.items.length - 1, 1);", "pop", 1, 53),
	})
}

func TestNoUnnecessarySpliceEditsFollowDemand(t *testing.T) {
	const code = `array.splice();
array.splice(0, 1);
array.splice(array.length, 0, update());
const box = {get items() { return []; }};
box["items"].splice(box.items.length - 1, 1);`
	root := fixtures.GetRootDir()
	filename := tspath.ResolvePath(root.Dir, "input.js")
	fs := utils.NewOverlayVFS(root.FS, map[string]string{filename: code})
	sourceProgram, err := lintprogram.NewFromRoots(lintprogram.RootOptions{
		Host:            utils.CreateCompilerHost(root.Dir, fs),
		RootFileNames:   []string{filename},
		CompilerOptions: lintprogram.SourceOnlyCompilerOptions(),
	})
	if err != nil {
		t.Fatal(err)
	}
	configured := []rule.ConfiguredRule{{
		Name: NoUnnecessarySpliceRule.Name, Severity: rule.SeverityError,
		Run: func(ctx rule.RuleContext) rule.RuleListeners { return NoUnnecessarySpliceRule.Run(ctx, nil) },
	}}
	var allEdits []rule.RuleDiagnostic
	for _, demand := range []rule.EditDemand{rule.EditDemandAll, rule.EditDemandNone, rule.EditDemandSuggestion, rule.EditDemandAutofix} {
		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program: sourceProgram, File: filename,
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule { return configured },
			Consumer: rule.DiagnosticConsumer{
				Demand: demand,
				Report: func(diagnostic rule.RuleDiagnostic) { diagnostics = append(diagnostics, diagnostic) },
			},
		})
		if len(diagnostics) != 4 {
			t.Fatalf("demand %d: got %d diagnostics, want 4", demand, len(diagnostics))
		}
		if demand == rule.EditDemandAll {
			allEdits = diagnostics
		}
		for index, diagnostic := range diagnostics {
			expected := allEdits[index]
			if diagnostic.Range != expected.Range || diagnostic.Severity != expected.Severity ||
				!reflect.DeepEqual(diagnostic.Message, expected.Message) {
				t.Errorf("demand %d: changed diagnostic %d", demand, index)
			}
			wantFix := index < 2 && demand&rule.EditDemandAutofix != 0
			if (diagnostic.FixesPtr != nil) != wantFix ||
				wantFix && !reflect.DeepEqual(diagnostic.FixesPtr, expected.FixesPtr) {
				t.Errorf("demand %d: unexpected fix for diagnostic %d", demand, index)
			}
			if diagnostic.Suggestions != nil {
				t.Errorf("demand %d: unexpected suggestions for diagnostic %d", demand, index)
			}
		}
	}
}
