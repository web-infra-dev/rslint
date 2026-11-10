package prefer_array_index_of_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_array_index_of"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestPreferArrayIndexOfExtras(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// Optional calls and a parenthesized optional-chain callee are not direct calls.
		{Code: `values.findIndex?.(x => x === needle); values?.findLastIndex?.(x => x === needle);`, FileName: "case.js"},
		{Code: `(values?.findIndex)(x => x === needle);`, FileName: "case.js"},
		{Code: `class List { #findIndex(cb) {} search() { return this.#findIndex(x => x === needle); } }`, FileName: "case.js"},
		// Default and rest parameters are not plain identifier parameters in ESTree.
		{Code: `values.findIndex((x = fallback) => x === needle); values.findLastIndex((...x) => x === needle);`, FileName: "case.js"},
		{Code: `values.findIndex(x => { return x === needle; unreachable(); });`, FileName: "case.js"},
		{Code: `values.findIndex((x => x === needle) as any); values.findLastIndex!(x => x === needle);`, FileName: "case.ts"},
		{Code: `values.findIndex(x => (x as number) === needle);`, FileName: "case.ts"},
		{Code: `values.findIndex(x => x === ({x}));`, FileName: "case.js"},
		{Code: `[, 1].findIndex(x => x === undefined);`, FileName: "case.js"},
		{Code: `([1, , 2] as (number | undefined)[]).findLastIndex(x => x === undefined);`, FileName: "case.ts"},
		{Code: `const values = [, 1]; const alias = values; values.findIndex(x => x === undefined); alias.findLastIndex(x => x === undefined);`, FileName: "case.js"},
		{Code: `new Array(2).findIndex(x => x === undefined); const values = Array(2); values.findLastIndex(x => x === undefined);`, FileName: "case.js"},
		{Code: `function last(length: number) { return new Array(length).findLastIndex(x => x === undefined); }`, FileName: "case.ts"},
		// These bindings would change if moved out of the ordinary callback.
		{Code: `values.findIndex(function (x) { return x === new.target; });`, FileName: "case.js"},
		{Code: `values.findIndex(function (x) { return x === (() => this.value)(); });`, FileName: "case.js"},
		{Code: `values.findLastIndex(function (x) { return x === (() => arguments[0])(); });`, FileName: "case.js"},
		{Code: `declare const values: unknown[], needle: unknown; values.findIndex(<T,>(x: T) => x === (needle as T));`, FileName: "case.tsx"},
		{Code: `declare const values: unknown[], needle: unknown; values.findLastIndex(x => x === (needle as typeof x));`, FileName: "case.ts"},
	}
	first, last := indexSearch{"findIndex", "indexOf"}, indexSearch{"findLastIndex", "lastIndexOf"}
	invalid := []rule_tester.InvalidTestCase{
		first.invalid("case.js", `const label = '🧪'; values.FIND((x) => ((x) === ((label))));`,
			`const label = '🧪'; values.INDEX((label));`, ""),
		first.invalid("case.js", `(values.FIND)(x => x === needle);`, `(values.INDEX)(needle);`, ""),
		last.invalid("case.js", `values?.items.FIND(x => x === needle);`, `values?.items.INDEX(needle);`, ""),
		last.invalid("case.js", `(values?.items).FIND(x => x === needle);`, `(values?.items).INDEX(needle);`, ""),
		first.invalid("case.jsx", `<List index={values.FIND(x => x === needle)} />;`, `<List index={values.INDEX(needle)} />;`, ""),
		first.invalid("case.js", `values.FIND(x => x === item.x);`, "", `values.INDEX(item.x);`),
		first.invalid("case.js", `[0, 2].FIND(x => x === ({ value: 2 }).value);`, `[0, 2].INDEX(({ value: 2 }).value);`, ""),
		first.invalid("case.js", `let reads = 0; const target = { get value() { return ++reads; } }; [0, 2].FIND(x => x === target.value);`, "",
			`let reads = 0; const target = { get value() { return ++reads; } }; [0, 2].INDEX(target.value);`),
		first.invalid("case.js", `Array(0).FIND(x => x === undefined);`, `Array(0).INDEX(undefined);`, ""),
		last.invalid("case.js", `new Array("item").FIND(x => x === "item");`, `new Array("item").INDEX("item");`, ""),
		first.invalid("case.js", `function lookup(Array) { return Array(2).FIND(x => x === needle); }`,
			`function lookup(Array) { return Array(2).INDEX(needle); }`, ""),
		first.invalid("case.js", `values.FIND(x => x === ((x) => x)(needle));`, "", `values.INDEX(((x) => x)(needle));`),
		last.invalid("case.js", `function outer() { return values.FIND(x => x === new.target); }`,
			`function outer() { return values.INDEX(new.target); }`, ""),
		first.invalid("case.js", "values.FIND(x => x === makeValue`ready`);", "", "values.INDEX(makeValue`ready`);"),
		last.invalid("case.js", `values.FIND(x => x === /* keep */ next());`, "", ""),
		first.invalid("case.js", `values.FIND(/* before */ x => x === needle /* after */);`,
			`values.INDEX(/* before */ needle /* after */);`, ""),
		first.invalid("case.js", `values.FIND(x => (/** @type {boolean} */ (x === needle)));`, "", ""),
		first.invalid("case.js", `const values = ["foo"]; const needle = "foo"; values.FIND((/** @type {(x: string) => boolean} */ (x => x === needle)));`, "", ""),
		first.invalid("case.js", `const values = ["foo"]; const needle = "foo"; values.FIND(/** @this {{name: string}} */ function (x) { return x === needle; });`,
			`const values = ["foo"]; const needle = "foo"; values.INDEX(/** @this {{name: string}} */ needle);`, ""),
		first.invalid("case.ts", `declare const values: unknown[], needle: unknown; values.FIND(<T>(x: T) => x === needle);`,
			`declare const values: unknown[], needle: unknown; values.INDEX(needle);`, ""),
		last.invalid("case.js", `class Search { #id = "foo"; index(values) { return values.FIND(x => x === this.#id); } }`, "",
			`class Search { #id = "foo"; index(values) { return values.INDEX(this.#id); } }`),
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_array_index_of.PreferArrayIndexOfRule, valid, invalid)
}

func TestPreferArrayIndexOfEditDemand(t *testing.T) {
	const code = "values.findIndex(x => x === needle);\n" +
		"values.findLastIndex(x => x === next());\n" +
		"values.findIndex(x => x === /* keep */ next());"
	root := fixtures.GetRootDir()
	file := tspath.ResolvePath(root.Dir, "edit-demand.js")
	fs := utils.NewOverlayVFS(root.FS, map[string]string{file: code})
	prog, err := program.NewFromRoots(program.RootOptions{
		Host:            utils.CreateCompilerHost(root.Dir, fs),
		RootFileNames:   []string{file},
		CompilerOptions: program.SourceOnlyCompilerOptions(),
		SingleThreaded:  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	target := prefer_array_index_of.PreferArrayIndexOfRule
	var all []rule.RuleDiagnostic
	for _, demand := range []rule.EditDemand{rule.EditDemandAll, rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion} {
		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program: prog, File: file,
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{
					Name: target.Name, Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						if ctx.TypeChecker != nil {
							t.Fatal("source-only case received a type checker")
						}
						return target.Run(ctx, nil)
					},
				}}
			},
			Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, d)
			}},
		})
		if len(diagnostics) != 3 {
			t.Fatalf("demand %d: got %d diagnostics, want 3", demand, len(diagnostics))
		}
		if demand == rule.EditDemandAll {
			all = diagnostics
		}
		for i, diagnostic := range diagnostics {
			if diagnostic.Range != all[i].Range || !reflect.DeepEqual(diagnostic.Message, all[i].Message) {
				t.Errorf("demand %d changed diagnostic %d", demand, i)
			}
			wantFix := i == 0 && demand&rule.EditDemandAutofix != 0
			if (diagnostic.FixesPtr != nil) != wantFix || (wantFix && !reflect.DeepEqual(diagnostic.FixesPtr, all[i].FixesPtr)) {
				t.Errorf("demand %d: unexpected fixes for diagnostic %d", demand, i)
			}
			wantSuggestion := i == 1 && demand&rule.EditDemandSuggestion != 0
			if (diagnostic.Suggestions != nil) != wantSuggestion || (wantSuggestion && !reflect.DeepEqual(diagnostic.Suggestions, all[i].Suggestions)) {
				t.Errorf("demand %d: unexpected suggestions for diagnostic %d", demand, i)
			}
		}
	}
}
