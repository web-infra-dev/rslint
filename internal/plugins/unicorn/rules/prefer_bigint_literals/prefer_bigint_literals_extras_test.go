package prefer_bigint_literals_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_bigint_literals"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestPreferBigIntLiteralsExtras(t *testing.T) {
	root := fixtures.GetRootDir()
	root.FS = utils.NewOverlayVFS(root.FS, map[string]string{
		tspath.ResolvePath(root.Dir, "tsconfig.es2019.json"): `{"extends": "./tsconfig.json", "compilerOptions": {"target": "es2019"}}`,
	})
	typed := fixed("const value = (BigInt(-1) as bigint) ** 2n;", "const value = ((-1n) as bigint) ** 2n;", at(1, 16, 1, 26))
	typed.FileName = "case.ts"
	jsx := fixed(`const view = <span title="😀">{BigInt((2))}</span>;`, `const view = <span title="😀">{2n}</span>;`, at(1, 32, 1, 43))
	jsx.FileName = "case.jsx"
	withStatement := fixed("with (BigInt(1)) BigInt(2);", "with (1n) BigInt(2);", at(1, 7, 1, 16))
	withStatement.LanguageOptions.SourceType = "script"

	rule_tester.RunRuleTester(root, "tsconfig.json", t,
		&prefer_bigint_literals.PreferBigIntLiteralsRule,
		[]rule_tester.ValidTestCase{
			{Code: `BigInt(1e309); BigInt(-"1"); BigInt(++value);`, FileName: "case.js"},
			{Code: `BigInt("-0x1"); BigInt("+0o1"); BigInt("\u00851");`, FileName: "case.js"},
			{Code: `globalThis.BigInt(1); globalThis["BigInt"](1); (globalThis?.BigInt)(1);`, FileName: "case.js"},
			{Code: `BigInt<number>(1); BigInt(1 as number); (BigInt as BigIntConstructor)(1);`, FileName: "case.ts"},
			{Code: "const value = BigInt(1);", FileName: "case.ts", TSConfig: "tsconfig.es2019.json"},
			{
				Code: "function BigInt(value) { return value; } BigInt(1);", FileName: "case.js",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			},
			{
				Code: "BigInt(1);", FileName: "case.js",
				Globals: map[string]any{"BigInt": "off"},
			},
			{
				Code: "BigInt(1);", FileName: "case.js",
				Globals:         map[string]any{"BigInt": "readonly"},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2019},
			},
		},
		[]rule_tester.InvalidTestCase{
			fixed("BigInt(-(-1))", "1n", at(1, 1, 1, 14)),
			fixed(`BigInt("\uFEFF0x0f\u00A0")`, "15n", at(1, 1, 1, 27)),
			fixed("BigInt(9007199254740991); BigInt(9007199254740992);", "9007199254740991n; 9007199254740992n;", at(1, 1, 1, 25), at(1, 27, 1, 51)),
			fixed(`B\u0069gInt(0X20000000000000)`, "0X20000000000000n", at(1, 1, 1, 30)),
			suggested("BigInt(9007199254740991.1)", "9007199254740991n", at(1, 1, 1, 27)),
			suggested("BigInt(-1e-999)", "0n", at(1, 1, 1, 16)),
			suggested("BigInt(0b0010)", "2n", at(1, 1, 1, 15)),
			// The Number spelling round-trips to 1000000000000000128, which is the
			// integer BigInt receives. Upstream would autofix to a different value.
			suggested("BigInt(1000000000000000100)", "1000000000000000128n", at(1, 1, 1, 28)),
			fixed("function f(BigInt) { BigInt(1); }\nBigInt(2);", "function f(BigInt) { BigInt(1); }\n2n;", at(2, 1, 2, 10)),
			withStatement,
			fixed("function f() {return(BigInt)(1)in obj}", "function f() {return 1n in obj}", at(1, 21, 1, 32)),
			fixed("if (ready) BigInt(-2); use(BigInt(-2));", "if (ready) (-2n); use((-2n));", at(1, 12, 1, 22), at(1, 28, 1, 38)),
			fixed("function f() { return\nBigInt(-1) }", "function f() { return\n;(-1n) }", at(2, 1, 2, 11)),
			fixed("\"use strict\"\nBigInt(-1)", "\"use strict\"\n;(-1n)", at(2, 1, 2, 11)),
			fixed("/** @type {bigint} */ (BigInt(-1))", "/** @type {bigint} */ (-1n)", at(1, 24, 1, 34)),
			fixed("BigInt(1)?.toString()", "1n?.toString()", at(1, 1, 1, 10)),
			typed,
			jsx,
		},
	)
}

func TestPreferBigIntLiteralsLegacyOctalPrecision(t *testing.T) {
	checkLegacySuggestion(t, "BigInt(01000000000000000000000)", "9223372036854775808n")
}

func checkLegacySuggestion(t *testing.T, source, replacement string) {
	t.Helper()
	lint := lintBigIntSource(t, source, rule.LanguageOptions{SourceType: "script"})
	diagnostics := lint(rule.EditDemandAll)
	if len(diagnostics) != 1 {
		t.Fatalf("%q: got %d diagnostics, want 1", source, len(diagnostics))
	}
	diagnostic := diagnostics[0]
	want := at(1, 1, 1, len(source)+1)
	if diagnostic.Message.Id != want.MessageId || diagnostic.Message.Description != want.Message ||
		diagnostic.Range != core.NewTextRange(0, len(source)) {
		t.Errorf("%q: unexpected diagnostic: %#v", source, diagnostic)
	}
	if diagnostic.FixesPtr != nil || diagnostic.Suggestions == nil || len(*diagnostic.Suggestions) != 1 {
		t.Fatalf("%q: expected only one suggestion, got %#v", source, diagnostic)
	}
	suggestion := (*diagnostic.Suggestions)[0]
	if suggestion.Message.Id != "prefer-bigint-literals/suggestion" {
		t.Errorf("%q: unexpected suggestion ID %q", source, suggestion.Message.Id)
	}
	if output, _, changed := linter.ApplyRuleFixes(source, []rule.RuleSuggestion{suggestion}); !changed || output != replacement {
		t.Errorf("%q: suggestion gives %q, want %q", source, output, replacement)
	}
}

func TestPreferBigIntLiteralsEditDemand(t *testing.T) {
	const source = "BigInt(1);\nBigInt(/* units */-2);\nBigInt(1e21);"
	lint := lintBigIntSource(t, source, rule.LanguageOptions{})

	all := lint(rule.EditDemandAll)
	for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
		diagnostics := all
		if demand != rule.EditDemandAll {
			diagnostics = lint(demand)
		}
		if len(diagnostics) != 3 {
			t.Fatalf("demand %d: got %d diagnostics, want 3", demand, len(diagnostics))
		}
		for i, want := range all {
			diagnostic := diagnostics[i]
			wantsFix := i == 0 && demand&rule.EditDemandAutofix != 0
			wantsSuggestion := i != 0 && demand&rule.EditDemandSuggestion != 0
			if (diagnostic.FixesPtr != nil) != wantsFix || (diagnostic.Suggestions != nil) != wantsSuggestion {
				t.Fatalf("demand %d, diagnostic %d has unexpected edits: %#v", demand, i, diagnostic)
			}
			if !wantsFix {
				want.FixesPtr = nil
			}
			if !wantsSuggestion {
				want.Suggestions = nil
			}
			if !reflect.DeepEqual(diagnostic, want) {
				t.Errorf("demand %d, diagnostic %d differs from all-edits result", demand, i)
			}
		}
	}
	for i, replacement := range []string{"`-2n`", "a bigint literal"} {
		suggestions := *all[i+1].Suggestions
		if len(suggestions) != 1 {
			t.Fatalf("diagnostic %d has %d suggestions, want 1", i+1, len(suggestions))
		}
		want := rule.RuleMessage{
			Id: "prefer-bigint-literals/suggestion", Description: "Replace with " + replacement + ".",
			Data: map[string]string{"replacement": replacement},
		}
		if !reflect.DeepEqual(suggestions[0].Message, want) {
			t.Errorf("suggestion message = %#v, want %#v", suggestions[0].Message, want)
		}
	}
}

func lintBigIntSource(t *testing.T, source string, languageOptions rule.LanguageOptions) func(rule.EditDemand) []rule.RuleDiagnostic {
	t.Helper()
	root := fixtures.GetRootDir()
	file := tspath.ResolvePath(root.Dir, "case.js")
	fs := utils.NewOverlayVFS(root.FS, map[string]string{file: source})
	host := utils.CreateCompilerHost(root.Dir, fs)
	compiler, err := utils.CreateProgramLenient(true, fs, root.Dir, "tsconfig.json", host)
	if err != nil {
		t.Fatal(err)
	}
	if compiler.GetSourceFile(file) == nil {
		t.Fatalf("missing source file %q", file)
	}
	sourceProgram := program.NewFromCompiler(compiler)
	bigintRule := &prefer_bigint_literals.PreferBigIntLiteralsRule
	options := rule_tester.ResolveTestCaseOptions(t, bigintRule, nil)
	return func(demand rule.EditDemand) []rule.RuleDiagnostic {
		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program: sourceProgram,
			File:    file,
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{
					Name: bigintRule.Name, Severity: rule.SeverityError,
					Environment: &rule.RuleEnvironment{LanguageOptions: languageOptions},
					Run:         func(ctx rule.RuleContext) rule.RuleListeners { return bigintRule.Run(ctx, options) },
				}}
			},
			Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(diagnostic rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, diagnostic)
			}},
		})
		return diagnostics
	}
}
