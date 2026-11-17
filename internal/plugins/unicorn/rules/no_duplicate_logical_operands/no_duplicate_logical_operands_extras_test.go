package no_duplicate_logical_operands_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_duplicate_logical_operands"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoDuplicateLogicalOperandsExtras(t *testing.T) {
	script := rule.LanguageOptions{SourceType: "script"}
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t,
		&no_duplicate_logical_operands.NoDuplicateLogicalOperandsRule,
		[]rule_tester.ValidTestCase{
			// Only a same-operator left operand is flattened, as ESTree does.
			valid(`foo && bar || bar`, "file.js"),
			valid(`(foo && bar) || bar`, "file.js"),
			valid(`foo ?? bar ?? bar`, "file.js"),
			valid(`foo &&= foo`, "file.js"),
			valid(`foo ||= foo`, "file.js"),

			// Different member paths and keys.
			valid(`foo.bar && foo.baz`, "file.js"),
			valid(`foo[bar] && foo.bar`, "file.js"),
			valid(`foo[0] && foo[1]`, "file.js"),
			valid(`foo[bar] && foo[baz]`, "file.js"),
			valid(`a.b.c && a.b.d`, "file.js"),
			valid(`class A {#foo; foo; m() { return this.#foo && this.foo; }}`, "file.js"),

			// Optional members anywhere in either member path, including behind wrappers.
			valid(`foo?.bar.baz && foo?.bar.baz`, "file.js"),
			valid(`foo?.bar!.baz && foo?.bar!.baz`, "file.ts"),
			valid(`(foo?.bar).baz && (foo?.bar).baz`, "file.js"),
			valid(`foo?.bar! && foo?.bar!`, "file.ts"),
			valid(`foo.bar && foo?.bar`, "file.js"),

			// Computed keys outside ESTree's Identifier, ThisExpression and Literal.
			valid(`foo[-1] && foo[-1]`, "file.js"),
			valid(`foo[bar.baz] && foo[bar.baz]`, "file.js"),
			valid(`foo().bar && foo().bar`, "file.js"),

			// Non-reference operands.
			valid(`function f() { return new.target && new.target; }`, "file.js"),
			valid(`!foo && !foo`, "file.js"),
			valid(`typeof foo && typeof foo`, "file.js"),

			// Any enclosing `with` statement disables the rule, even across functions.
			{Code: `with (scope) { (function () { return foo || foo; }); }`, FileName: "file.js", LanguageOptions: script},
			{Code: `with (foo && foo) {}`, FileName: "file.js", LanguageOptions: script},

			// A following line can continue the right operand into a call or member.
			valid("foo && foo\n(bar)", "file.js"),
			valid("foo && foo\n[0]", "file.js"),
		},
		[]rule_tester.InvalidTestCase{
			// Parentheses are transparent for matching and retained in the edit.
			fixed(`((foo)) && ((foo))`, "file.js", `((foo))`, 14, 17),
			fixed(`(foo && foo)`, "file.js", `(foo)`, 9, 12),
			fixed(`(foo && bar) && bar`, "file.js", `(foo && bar)`, 17, 20),
			fixed(`foo && (bar) && bar`, "file.js", `foo && (bar)`, 17, 20),
			fixed(`foo || bar && bar`, "file.js", `foo || bar`, 15, 18),
			fixed(`a ?? (foo || foo)`, "file.js", `a ?? (foo)`, 14, 17),
			fixed(`(foo || foo) ?? a`, "file.js", `(foo) ?? a`, 9, 12),
			fixed(`if (a && (b || b)) {}`, "file.js", `if (a && (b)) {}`, 16, 17),
			// Columns count UTF-16 code units, not bytes.
			fixed(`ü && ü`, "file.js", `ü`, 6, 7),

			// A TypeScript wrapper on the adjacent operand selects a suggestion.
			suggested(`foo && bar! && bar`, "file.ts", `foo && bar!`, 16, 19),
			suggested(`(foo as Foo) || foo`, "file.ts", `(foo as Foo)`, 17, 20),
			suggested(`(foo as any).bar && foo.bar`, "file.ts", `(foo as any).bar`, 21, 28),
			suggested(`foo!.bar && foo.bar`, "file.ts", `foo!.bar`, 13, 20),
			suggested(`foo[bar as any] && foo[bar]`, "file.ts", `foo[bar as any]`, 20, 28),

			// Instantiation expressions are TypeScript wrappers, too.
			suggested(`(foo<string>) && (foo<string>)`, "file.ts", `(foo<string>)`, 19, 30),
			suggested(`foo[bar<string>] && foo[bar<string>]`, "file.ts", `foo[bar<string>]`, 21, 37),

			// Static and simple computed keys.
			suggested(`foo.bar && foo['bar']`, "file.js", `foo.bar`, 12, 22),
			suggested(`foo[0] && foo["0"]`, "file.js", `foo[0]`, 11, 19),
			suggested(`foo[this] && foo[this]`, "file.js", `foo[this]`, 14, 23),
			suggested(`foo[null] || foo[null]`, "file.js", `foo[null]`, 14, 23),
			suggested(`foo[1n] && foo[1n]`, "file.js", `foo[1n]`, 12, 19),
			suggested(`foo[/a/] && foo[/a/]`, "file.js", `foo[/a/]`, 13, 21),
			suggested(`foo[bar] && foo[(bar)]`, "file.js", `foo[bar]`, 13, 23),
			suggested(`foo[bar][baz] && foo[bar][baz]`, "file.js", `foo[bar][baz]`, 18, 31),
			suggested(
				`class A extends B { m() { return super[foo] && super[foo]; } }`, "file.js",
				`class A extends B { m() { return super[foo]; } }`, 48, 58,
			),

			// Reference contexts are found through parentheses and TypeScript wrappers.
			reported(`(object.method || object.method)?.()`, "file.js", 19, 32),
			reported(`((object.method || object.method) as any)()`, "file.ts", 20, 33),
			reported(`((object.method || object.method)<string>)()`, "file.ts", 20, 33),
			reported(`(object.method || object.method)<string>()`, "file.ts", 19, 32),
			reported("(object.tag && object.tag)!`text`", "file.ts", 16, 26),
			reported(`delete (value || value)!`, "file.ts", 18, 23),
			// rslint difference: `typeof` hides a ReferenceError from a bare
			// identifier, so `typeof (missing || missing)` is not fixed to
			// `typeof (missing)`.
			reported(`typeof (missing || missing)`, "file.js", 20, 27),
			reported(`typeof (missing && missing);`, "file.js", 20, 27),
			reported(`typeof ((missing || missing) as unknown)`, "file.ts", 21, 28),
			// Member reads still throw under `typeof`, and a nested pair keeps a
			// logical expression as the `typeof` operand.
			suggested(`typeof (a.b || a.b)`, "file.js", `typeof (a.b)`, 16, 19),
			fixed(`typeof (x || x || y)`, "file.js", `typeof (x || y)`, 14, 15),
			fixed(`typeof (missing || missing).length`, "file.js", `typeof (missing).length`, 20, 27),
			// Other operand positions keep their edits.
			fixed(`call(foo || foo)`, "file.js", `call(foo)`, 13, 16),
			{
				Code: `(foo || foo)(foo || foo)`, FileName: "file.js", Output: []string{`(foo || foo)(foo)`},
				Errors: []rule_tester.InvalidTestCaseError{diagnostic(1, 9, 1, 12), diagnostic(1, 21, 1, 24)},
			},
			fixed(`void (foo || foo)`, "file.js", `void (foo)`, 14, 17),
			fixed(`(foo || foo).bar`, "file.js", `(foo).bar`, 9, 12),
			suggested(`(a.b || a.b).c()`, "file.js", `(a.b).c()`, 9, 12),

			// Multiline right operand.
			{
				Code: "foo\n  && foo", FileName: "file.js", Output: []string{"foo"},
				Errors: []rule_tester.InvalidTestCaseError{diagnostic(2, 6, 2, 9)},
			},

			// Comments outside the logical expression are untouched by the edit.
			fixed(`foo && foo // trailing`, "file.js", `foo // trailing`, 8, 11),
			fixed(`/* lead */ foo && foo`, "file.js", `/* lead */ foo`, 19, 22),
			fixed(`/** @type {boolean} */ (foo) && foo`, "file.js", `/** @type {boolean} */ (foo)`, 33, 36),
			fixed(`((foo /* c */)) && foo`, "file.js", `((foo /* c */))`, 20, 23),
			reported(`foo && (/* c */ foo)`, "file.js", 17, 20),
			// Comments also withhold the suggestion for member operands.
			reported(`foo.bar /* keep */ && foo.bar`, "file.js", 23, 30),

			// The retained text ends where the duplicate began, so ASI is unchanged.
			fixed("foo && foo\n;(bar)", "file.js", "foo\n;(bar)", 8, 11),

			// A `with` statement that does not enclose the expression does not matter.
			{
				Code: "with (scope) {}\nfoo && foo", FileName: "file.js", LanguageOptions: script,
				Output: []string{"with (scope) {}\nfoo"},
				Errors: []rule_tester.InvalidTestCaseError{diagnostic(2, 8, 2, 11)},
			},

			// Each adjacent pair reports, the outer expression first in traversal
			// order; the overlapping fixes apply over two passes.
			{
				Code: `foo && bar && bar && bar`, FileName: "file.js",
				Output: []string{`foo && bar && bar`, `foo && bar`},
				Errors: []rule_tester.InvalidTestCaseError{diagnostic(1, 22, 1, 25), diagnostic(1, 15, 1, 18)},
			},
		},
	)
}

// TestNoDuplicateLogicalOperandsEditDemand checks that the diagnostic is
// identical under every demand and only the requested edit kind materializes.
func TestNoDuplicateLogicalOperandsEditDemand(t *testing.T) {
	for _, test := range []struct {
		code, output    string
		fix, suggestion bool
	}{
		{"foo && foo", "foo", true, false},
		{"foo.bar && foo.bar", "foo.bar", false, true},
		{"foo /* keep */ && foo", "", false, false},
	} {
		t.Run(test.code, func(t *testing.T) {
			helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
			program, source, err := helper.CreateTestProgram(test.code, "edit-demand.ts", "tsconfig.json")
			if err != nil {
				t.Fatal(err)
			}
			diagnostics := make(map[rule.EditDemand]rule.RuleDiagnostic, 4)
			for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
				var got []rule.RuleDiagnostic
				linter.LintSingleFile(linter.LintSingleFileOptions{
					Program: lintprogram.NewFromCompiler(program), File: source.FileName(),
					GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
						return []rule.ConfiguredRule{{
							Name: no_duplicate_logical_operands.NoDuplicateLogicalOperandsRule.Name, Severity: rule.SeverityError,
							Run: func(ctx rule.RuleContext) rule.RuleListeners {
								return no_duplicate_logical_operands.NoDuplicateLogicalOperandsRule.Run(ctx, nil)
							},
						}}
					},
					Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) { got = append(got, d) }},
				})
				if len(got) != 1 {
					t.Fatalf("demand %d: got %d diagnostics", demand, len(got))
				}
				diagnostics[demand] = got[0]
			}
			baseline, all := diagnostics[rule.EditDemandNone], diagnostics[rule.EditDemandAll]
			for demand, diagnostic := range diagnostics {
				if diagnostic.Range != baseline.Range || !reflect.DeepEqual(diagnostic.Message, baseline.Message) || diagnostic.Severity != baseline.Severity {
					t.Errorf("demand %d changed diagnostic identity", demand)
				}
				if wantFix := test.fix && (demand == rule.EditDemandAutofix || demand == rule.EditDemandAll); wantFix {
					if diagnostic.FixesPtr == nil || len(*diagnostic.FixesPtr) != 1 || !reflect.DeepEqual(diagnostic.FixesPtr, all.FixesPtr) {
						t.Fatalf("demand %d: inconsistent autofixes", demand)
					}
					output, _, fixed := linter.ApplyRuleFixes(test.code, []rule.RuleDiagnostic{diagnostic})
					if !fixed || output != test.output {
						t.Errorf("demand %d: unexpected autofix %q", demand, output)
					}
				} else if diagnostic.FixesPtr != nil && len(*diagnostic.FixesPtr) != 0 {
					t.Errorf("demand %d produced an unexpected autofix", demand)
				}
				if wantSuggestion := test.suggestion && (demand == rule.EditDemandSuggestion || demand == rule.EditDemandAll); !wantSuggestion {
					if diagnostic.Suggestions != nil && len(*diagnostic.Suggestions) != 0 {
						t.Errorf("demand %d produced unexpected suggestions", demand)
					}
					continue
				}
				if diagnostic.Suggestions == nil || len(*diagnostic.Suggestions) != 1 || !reflect.DeepEqual(diagnostic.Suggestions, all.Suggestions) {
					t.Fatalf("demand %d: inconsistent suggestions", demand)
				}
				suggestion := (*diagnostic.Suggestions)[0]
				if suggestion.Message.Id != suggestionID || suggestion.Message.Description != "Remove the duplicate operand." {
					t.Errorf("demand %d: wrong suggestion message %+v", demand, suggestion.Message)
				}
				output, _, fixed := linter.ApplyRuleFixes(test.code, []rule.RuleSuggestion{suggestion})
				if !fixed || output != test.output {
					t.Errorf("demand %d: unexpected suggestion output %q", demand, output)
				}
			}
		})
	}
}
