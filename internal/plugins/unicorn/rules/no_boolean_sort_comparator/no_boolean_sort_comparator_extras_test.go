package no_boolean_sort_comparator_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_boolean_sort_comparator"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoBooleanSortComparatorEdges(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_boolean_sort_comparator.NoBooleanSortComparatorRule, []rule_tester.ValidTestCase{
		{Code: "(rows?.sort)((a, b) => a > b);", FileName: "case.js"},
		{Code: "class Collection { #sort(compare) {} run() { this.#sort((a, b) => a > b); } }", FileName: "case.js"},
		{Code: "let sort = 'sort'; sort = other; rows[sort]((a, b) => a > b);", FileName: "case.js"},
		{Code: "rows.sort(Reflect?.has);", FileName: "case.js"},
		{Code: "function order(Boolean) { rows.sort((a, b) => Boolean(a)); }", FileName: "case.js"},
		{Code: "const compare = (a, b) => compare(a, b); rows.sort(compare);", FileName: "case.js"},
		{Code: "rows.sort((a, b) => { log(a, b); return a > b; });", FileName: "case.js"},
		{Code: "rows.sort(function () { return; });", FileName: "case.js"},
	}, []rule_tester.InvalidTestCase{
		invalidComparator("rows?.[`sort`]?.(((a, b) => a?.score > b?.score));", "case.js", 1, 19, 1, 48),
		invalidComparator("class Row { #score = 0; static sort(rows) { return rows.sort((a, b) => a.#score > b.#score); } }",
			"case.js", 1, 62, 1, 91, "class Row { #score = 0; static sort(rows) { return rows.sort((a, b) => a.#score - b.#score); } }"),
		invalidComparator("const sort = 'so' + 'rt'; rows[sort]((a, b) => a['rank'] >= b[`rank`]);", "case.js", 1, 38, 1, 70,
			"const sort = 'so' + 'rt'; rows[sort]((a, b) => a['rank'] - b[`rank`]);"),
		invalidComparator("rows.sort((a, b) => a.rank > b['rank']);", "case.js", 1, 11, 1, 39),
		invalidComparator("rows.sort((a = 0, b) => a > b);", "case.js", 1, 11, 1, 30),
		invalidComparator("rows.sort((a, b) => /x/.test(a.name));", "case.js", 1, 11, 1, 37),
		invalidComparator("rows.sort((compare<number>)!); function compare<T>(a: T, b: T): boolean { return a > b; }",
			"case.ts", 1, 11, 1, 29),
		invalidComparator("interface Comparator { (a: number, b: number): boolean } rows.sort(compare as Comparator);",
			"case.ts", 1, 68, 1, 89),
		invalidComparator("function compare(a: number, b: number): boolean; function compare(a: string, b: string): boolean; function compare(a: unknown, b: unknown) { return a === b; } rows.sort(compare);",
			"case.ts", 1, 170, 1, 177),
		invalidComparator("const view = <List rows={rows.toSorted((a, b) => a.score > b.score)} />;", "case.tsx", 1, 40, 1, 67,
			"const view = <List rows={rows.toSorted((a, b) => a.score - b.score)} />;"),
		invalidComparator("const label = '🌕'; rows.sort(((α, β) => (α) > (β)));", "case.js", 1, 32, 1, 51,
			"const label = '🌕'; rows.sort(((α, β) => α - β));"),
		// These suggestions would remove type syntax or change the comparator's bindings.
		invalidComparator("rows.sort((a, b) => (a > b) as boolean);", "case.ts", 1, 11, 1, 39),
		invalidComparator("rows.sort(/** @returns {boolean} */ (a, b) => a > b);", "case.js", 1, 37, 1, 52),
		invalidComparator("rows.sort((a, b) => /** @type {boolean} */ (a > b));", "case.js", 1, 11, 1, 51),
		invalidComparator("rows.sort(function (a, b) { return a[arguments[0]] > b[arguments[0]]; });", "case.js", 1, 11, 1, 72),
		invalidComparator("rows.sort(function compare(a, b) { return a[compare.key] > b[compare.key]; });", "case.js", 1, 11, 1, 77),
		invalidComparator("rows.sort(function (a, b) { return a[this.key] < b[this.key]; });", "case.js", 1, 11, 1, 64),
		// Subtraction cannot implement the intended comparator for these operands.
		invalidComparator("['a', 'b'].sort((a, b) => a < b);", "case.js", 1, 17, 1, 32),
		invalidComparator("new BigInt64Array(8).sort((a, b) => a > b);", "case.js", 1, 27, 1, 42),
		invalidComparator("const words: string[] = ['a', 'b']; words.toSorted((a, b) => a < b);", "case.ts", 1, 52, 1, 67),
	})
}

func TestNoBooleanSortComparatorSubtractionSafety(t *testing.T) {
	t.Run("type constraints", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_boolean_sort_comparator.NoBooleanSortComparatorRule, nil, []rule_tester.InvalidTestCase{
			invalidComparator("function f<T extends bigint>(rows: T[]) { return rows.sort((a, b) => a > b); }", "case.ts", 1, 60, 1, 75),
			invalidComparator("function f<T extends string>(rows: T[]) { return rows.sort((a, b) => a > b); }", "case.ts", 1, 60, 1, 75),
			invalidComparator("function f<T extends number | bigint>(rows: T[]) { return rows.sort((a, b) => a > b); }", "case.ts", 1, 69, 1, 84),
			invalidComparator("function f<T extends number>(rows: T[]) { return rows.sort((a, b) => a > b); }", "case.ts", 1, 60, 1, 75,
				"function f<T extends number>(rows: T[]) { return rows.sort((a, b) => a - b); }"),
		})
	})
	t.Run("without type information", func(t *testing.T) {
		testedRule := no_boolean_sort_comparator.NoBooleanSortComparatorRule
		testedRule.Run = func(ctx rule.RuleContext, options []any) rule.RuleListeners {
			ctx.TypeChecker = nil
			return no_boolean_sort_comparator.NoBooleanSortComparatorRule.Run(ctx, options)
		}
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &testedRule, nil, []rule_tester.InvalidTestCase{
			invalidComparator("[-2n, -1n].sort((a, b) => a > b);", "case.js", 1, 17, 1, 32),
			invalidComparator("BigInt64Array.of(1n, 2n).sort((a, b) => a > b);", "case.js", 1, 31, 1, 46),
			invalidComparator("BigUint64Array['from']([1n, 2n]).toSorted((a, b) => a > b);", "case.js", 1, 43, 1, 58),
			invalidComparator("const method = 'of'; BigInt64Array[method](1n, 2n).sort((a, b) => a > b);", "case.js", 1, 57, 1, 72),
			invalidComparator("const method = 'of'; Int32Array[method](1, 2).sort((a, b) => a > b);", "case.js", 1, 52, 1, 67,
				"const method = 'of'; Int32Array[method](1, 2).sort((a, b) => a - b);"),
			invalidComparator("function f(BigInt64Array) { const method = 'of'; return BigInt64Array[method](1, 2).sort((a, b) => a > b); }", "case.js", 1, 90, 1, 105,
				"function f(BigInt64Array) { const method = 'of'; return BigInt64Array[method](1, 2).sort((a, b) => a - b); }"),
		})
	})
}

func TestNoBooleanSortComparatorEditDemand(t *testing.T) {
	const code = "rows.sort((a, b) => a > b); rows.sort((a, b) => a === b);"
	program, file, err := rule_tester.NewProgramHelper(fixtures.GetRootDir()).CreateTestProgram(code, "case.js", "tsconfig.json")
	if err != nil {
		t.Fatal(err)
	}
	sourceProgram := lintprogram.NewFromCompiler(program)
	var baseline []rule.RuleDiagnostic
	for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program: sourceProgram, File: file.FileName(),
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{
					Name: no_boolean_sort_comparator.NoBooleanSortComparatorRule.Name, Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return no_boolean_sort_comparator.NoBooleanSortComparatorRule.Run(ctx, nil)
					},
				}}
			},
			Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }},
		})
		if len(diagnostics) != 2 {
			t.Fatalf("demand %d: got %d diagnostics, want 2", demand, len(diagnostics))
		}
		if demand == rule.EditDemandNone {
			baseline = diagnostics
		}
		for index, diagnostic := range diagnostics {
			if diagnostic.Range != baseline[index].Range || !reflect.DeepEqual(diagnostic.Message, baseline[index].Message) ||
				diagnostic.Severity != baseline[index].Severity || diagnostic.FixesPtr != nil {
				t.Errorf("demand %d changed diagnostic %d or provided an automatic fix: %#v", demand, index, diagnostic)
			}
			if index != 0 || (demand != rule.EditDemandSuggestion && demand != rule.EditDemandAll) {
				if diagnostic.Suggestions != nil {
					t.Errorf("demand %d: unexpected suggestions for diagnostic %d", demand, index)
				}
				continue
			}
			if diagnostic.Suggestions == nil || len(*diagnostic.Suggestions) != 1 {
				t.Fatalf("demand %d: want one suggestion", demand)
			}
			suggestion := (*diagnostic.Suggestions)[0]
			const replacement = "(a, b) => a - b"
			expectedMessage := rule.RuleMessage{
				Id: suggestionID, Description: "Replace with `" + replacement + "`.",
				Data: map[string]string{"replacement": replacement},
			}
			if !reflect.DeepEqual(suggestion.Message, expectedMessage) {
				t.Errorf("demand %d: unexpected suggestion message %#v", demand, suggestion.Message)
			}
			output, _, fixed := linter.ApplyRuleFixes(code, []rule.RuleSuggestion{suggestion})
			if !fixed || output != "rows.sort((a, b) => a - b); rows.sort((a, b) => a === b);" {
				t.Errorf("demand %d: unexpected suggestion output %q", demand, output)
			}
		}
	}
}
