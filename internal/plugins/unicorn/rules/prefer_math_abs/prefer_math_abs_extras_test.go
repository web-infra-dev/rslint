package prefer_math_abs_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_math_abs"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestPreferMathAbsExtras(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// An assertion can explicitly allow BigInt even when the checker simplifies the type.
		{Code: "function run(value: any | bigint) { return value < 0 ? -value : value; }", FileName: "file.ts"},
		{Code: "function run(value) { return (value as any & bigint) < 0 ? -value : value; }", FileName: "file.ts"},
		{Code: "value > limit || value < -(limit as any | bigint);", FileName: "file.ts"},
		// No Math.abs suggestion for a property whose inferred type is BigInt.
		{Code: "const state = {offset: 1n}; state.offset < 0 ? -state.offset : state.offset;", FileName: "file.ts"},
		{Code: "getValue() < 0 ? -getValue() : getValue();", FileName: "file.js"},
		{Code: "value === 0 ? -value : value; value in object ? -value : value;", FileName: "file.js"},
	}
	disabledMath := absCase("value > limit || value < -limit;", "", "file.js", 1, 1, 1, 32)
	disabledMath.Globals = map[string]any{"Math": "off"}
	invalid := []rule_tester.InvalidTestCase{
		absCase("function f() { return.5 < 0x0 ? -.5 : .5; }", "function f() { return Math.abs(.5); }", "file.js", 1, 22, 1, 41),
		absCase(
			"const result = <span>{\"😀\"}{(offset) < 0e9 ?\r\n  (-(offset)) : offset\r\n}</span>;",
			"const result = <span>{\"😀\"}{Math.abs((offset))\r\n}</span>;",
			"file.jsx", 1, 29, 2, 23,
		),
		absCase("delta > limit || delta < (0 - limit);", "Math.abs(delta) > limit;", "file.js", 1, 1, 1, 37),
		absCase("(value < -1 * limit) || (limit < value);", "Math.abs(value) > limit;", "file.js", 1, 1, 1, 40),
		absCase("value > limit || value < limit * -0x1;", "Math.abs(value) > limit;", "file.js", 1, 1, 1, 38),
		absCase("value > '4' || value < -`4`;", "Math.abs(value) > '4';", "file.js", 1, 1, 1, 28),
		absCase("state?.['offset'] < 0 ? -state.offset : state?.offset;", "", "file.js", 1, 1, 1, 54),
		absCase("obj['offset' as string] < 0 ? -obj.offset : obj.offset;", "", "file.ts", 1, 1, 1, 55),
		absCase("class A { #offset = 1; abs() { return this.#offset < 0 ? -this.#offset : this.#offset; } }", "", "file.js", 1, 39, 1, 86),
		absCase("function f(Math, value) { return value < 0 ? -value : value; }", "", "file.js", 1, 34, 1, 60),
		disabledMath,
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_math_abs.PreferMathAbsRule, valid, invalid)
}

func TestPreferMathAbsSourceOnlyBigInt(t *testing.T) {
	const expression = "value < 0 ? -value : value"
	for _, test := range []struct {
		name, filename, code string
		wantSuggestion       bool
	}{
		{
			name:     "generic alias of bigint",
			filename: "alias.ts",
			code:     "type Id<T> = T; function abs(value: Id<bigint>) { return " + expression + "; } abs(-5n);",
		},
		{
			name:           "generic alias of number",
			filename:       "alias.ts",
			code:           "type Id<T> = T; function abs(value: Id<number>) { return " + expression + "; }",
			wantSuggestion: true,
		},
		{
			name:     "distinct instantiations with a dependent default",
			filename: "alias.ts",
			code: "type Id<T> = T; type Value<T, Result = Id<T>> = Result; " +
				"function abs(value: Value<Id<number>> | Value<Id<Id<bigint>>>) { return " + expression + "; }",
		},
		{
			name:           "unused bigint type argument",
			filename:       "alias.ts",
			code:           "type Select<First, Second> = Second; function abs(value: Select<bigint, number>) { return " + expression + "; }",
			wantSuggestion: true,
		},
		{
			name:     "global BigInt",
			filename: "binding.js",
			code:     "const value = BigInt(1); " + expression + ";",
		},
		{
			name:           "shadowed BigInt",
			filename:       "binding.js",
			code:           "function BigInt(x) { return Number(x); } const value = BigInt(1); " + expression + ";",
			wantSuggestion: true,
		},
		{
			name:     "shadowed BigInt with a bigint return annotation",
			filename: "binding.js",
			code:     "/** @returns {bigint} */ function BigInt() { return -5n; } const value = BigInt(); " + expression + ";",
		},
		{
			name:     "global BigInt method",
			filename: "binding.js",
			code:     "const value = (BigInt).asUintN(8, 1n); " + expression + ";",
		},
		{
			name:           "shadowed BigInt method",
			filename:       "binding.js",
			code:           "const BigInt = { asIntN: (bits, x) => Number(x) }; const value = BigInt.asIntN(8, 1); " + expression + ";",
			wantSuggestion: true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			diagnostics := sourceOnlyLint(t, test.filename, test.code)(rule.EditDemandAll)
			if !test.wantSuggestion {
				if len(diagnostics) != 0 {
					t.Fatalf("expected no diagnostics, got %#v", diagnostics)
				}
				return
			}
			if len(diagnostics) != 1 {
				t.Fatalf("expected one diagnostic, got %#v", diagnostics)
			}
			diagnostic := diagnostics[0]
			if diagnostic.Message.Id != "prefer-math-abs" || diagnostic.FixesPtr != nil ||
				diagnostic.Suggestions == nil || len(*diagnostic.Suggestions) != 1 {
				t.Fatalf("expected a prefer-math-abs diagnostic with one suggestion, got %#v", diagnostic)
			}
			if got := test.code[diagnostic.Range.Pos():diagnostic.Range.End()]; got != expression {
				t.Errorf("reported %q, want %q", got, expression)
			}
			want := strings.Replace(test.code, expression, "Math.abs(value)", 1)
			if got, _, applied := linter.ApplyRuleFixes(test.code, *diagnostic.Suggestions); !applied || got != want {
				t.Errorf("suggestion produced %q, want %q", got, want)
			}
		})
	}
}

func TestPreferMathAbsEditDemand(t *testing.T) {
	const code = `const amount = /** @type {bigint} */ (read());
amount < 0 ? -amount : amount;
function abs(value) { return(value < 0 ? -value : value); }
function beyond(value, limit) { return value > limit || value < -limit; }
`
	const expected = `const amount = /** @type {bigint} */ (read());
amount < 0 ? -amount : amount;
function abs(value) { return (Math.abs(value)); }
function beyond(value, limit) { return Math.abs(value) > limit; }
`
	lint := sourceOnlyLint(t, "edit-demand.js", code)
	all := lint(rule.EditDemandAll)
	if len(all) != 2 {
		t.Fatalf("got %d diagnostics, want 2", len(all))
	}
	var suggestions []rule.RuleSuggestion
	for _, diagnostic := range all {
		if diagnostic.FixesPtr != nil || diagnostic.Suggestions == nil || len(*diagnostic.Suggestions) != 1 {
			t.Fatalf("expected one suggestion and no autofix, got %#v", diagnostic)
		}
		suggestion := (*diagnostic.Suggestions)[0]
		if suggestion.Message.Id != "prefer-math-abs/suggestion" || suggestion.Message.Description != "Replace with `Math.abs()`." {
			t.Fatalf("unexpected suggestion message %#v", suggestion.Message)
		}
		suggestions = append(suggestions, suggestion)
	}
	if output, _, applied := linter.ApplyRuleFixes(code, suggestions); !applied || output != expected {
		t.Fatalf("suggestions produced %q, want %q", output, expected)
	}
	if got := lint(rule.EditDemandSuggestion); !reflect.DeepEqual(got, all) {
		t.Fatalf("suggestion demand changed diagnostics: %#v", got)
	}
	for i := range all {
		all[i].Suggestions = nil
	}
	for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix} {
		if got := lint(demand); !reflect.DeepEqual(got, all) {
			t.Errorf("demand %d: unexpected diagnostics %#v", demand, got)
		}
	}
}

func sourceOnlyLint(t *testing.T, name, code string) func(rule.EditDemand) []rule.RuleDiagnostic {
	t.Helper()
	root := fixtures.GetRootDir()
	filename := tspath.ResolvePath(root.Dir, name)
	fs := utils.NewOverlayVFS(root.FS, map[string]string{filename: code})
	sourceProgram, err := program.NewFromRoots(program.RootOptions{
		RootFileNames:   []string{filename},
		Host:            utils.CreateCompilerHost(root.Dir, fs),
		CompilerOptions: program.SourceOnlyCompilerOptions(),
		SingleThreaded:  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	r := prefer_math_abs.PreferMathAbsRule
	return func(demand rule.EditDemand) []rule.RuleDiagnostic {
		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program: sourceProgram,
			File:    filename,
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{
					Name:     r.Name,
					Severity: rule.SeverityError,
					Run:      func(ctx rule.RuleContext) rule.RuleListeners { return r.Run(ctx, nil) },
				}}
			},
			Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(diagnostic rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, diagnostic)
			}},
		})
		return diagnostics
	}
}
