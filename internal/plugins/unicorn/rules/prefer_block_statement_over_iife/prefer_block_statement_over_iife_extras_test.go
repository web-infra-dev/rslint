package prefer_block_statement_over_iife_test

import (
	"reflect"
	"slices"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_block_statement_over_iife"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestPreferBlockStatementOverIIFEExtras(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// Call arguments and parameter initializers can have side effects.
		{Code: "(() => { run(); })(sideEffect());", FileName: "case.js"},
		{Code: "((value = sideEffect()) => { run(); })();", FileName: "case.js"},
		{Code: "void (() => { run(); })(); new (function () { run(); })();", FileName: "case.js"},
		{Code: "((() => { run(); }) satisfies () => void)(); (() => { run(); })() as void;", FileName: "case.ts"},
		{Code: "(() => { 'custom directive'; run(); })();", FileName: "case.js"},
		{Code: "(() => { for (var key in items) run(key); })();", FileName: "case.js"},
		{Code: "(function () { use(() => { return value; }); return; })();", FileName: "case.js"},
		{Code: "(() => { eval?.(eval(code)); })();", FileName: "case.js"},
		{Code: "(() => { (<Function>eval)(code); })();", FileName: "case.ts"},
		{Code: "(() => { (eval satisfies Function)(code); })();", FileName: "case.ts"},
		{Code: "(() => { (eval<string>)(code); })();", FileName: "case.ts"},

		// Keys and decorators evaluate in the IIFE's scope.
		{Code: "(() => { const obj = { [eval(code)]() {} }; run(obj); })();", FileName: "case.js"},
		{Code: "(function () { const obj = { [this.name]() {} }; run(obj); })();", FileName: "case.js"},
		{Code: "(function () { class C { @decorate(this) method() {} } run(C); })();", FileName: "case.ts"},
		{Code: "(function () { class C { method(@decorate(this) value: number) {} } run(C); })();", FileName: "case.ts"},

		// These replacements can change bindings even though Unicorn v79 reports them.
		{Code: "function outer() { (() => { function inner() {} })(); return typeof inner; }", FileName: "case.cjs"},
		{Code: `(function () { const getArgs = () => eval("arguments.length"); run(getArgs()); })();`, FileName: "case.js"},
		{Code: "(function () { render(<this.View />); })();", FileName: "case.tsx"},
		{Code: "(function () { render(<arguments.View />); })();", FileName: "case.tsx"},
	}
	invalid := []rule_tester.InvalidTestCase{
		{
			Code: "const 𐐀 = 1; ((() => { run(𐐀); })()); // after", FileName: "case.js",
			Output: []string{"const 𐐀 = 1; { run(𐐀); } // after"},
			Errors: iifeDiagnostic(1, 16, 1, 38),
		},
		{
			Code: "#!/usr/bin/env node\n// Run initialization\n\u00a0((() => { run(); })()); // outside", FileName: "case.js",
			Output: []string{"#!/usr/bin/env node\n// Run initialization\n\u00a0{ run(); } // outside"},
			Errors: iifeDiagnostic(3, 3, 3, 23),
		},
		{
			Code: "if (ready) (() => {})(); else wait();", FileName: "case.js",
			Output: []string{"if (ready) {} else wait();"},
			Errors: iifeDiagnostic(1, 12, 1, 24),
		},
		{
			Code: `(() => { ("use strict"); run(); "use strict"; })();`, FileName: "case.js",
			Output: []string{`{ ("use strict"); run(); "use strict"; }`},
			Errors: iifeDiagnostic(1, 1, 1, 51),
		},
		{
			Code: "(function () { const obj = { method() { var value = eval(code); return this.x + value; } }; run(obj); })();", FileName: "case.js",
			Output: []string{"{ const obj = { method() { var value = eval(code); return this.x + value; } }; run(obj); }"},
			Errors: iifeDiagnostic(1, 1, 1, 107),
		},
		{
			Code: "(function () { class C { #method() { return this; } } run(C); })();", FileName: "case.js",
			Output: []string{"{ class C { #method() { return this; } } run(C); }"},
			Errors: iifeDiagnostic(1, 1, 1, 67),
		},
		{
			Code: "(function () { const nested = () => { function local() {} return local; }; run(nested); })();", FileName: "case.js",
			LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			Output:          []string{"{ const nested = () => { function local() {} return local; }; run(nested); }"},
			Errors:          iifeDiagnostic(1, 1, 1, 93),
		},
		{
			Code: "(() => { const run = () => eval(code); run(); })();", FileName: "case.js",
			Output: []string{"{ const run = () => eval(code); run(); }"},
			Errors: iifeDiagnostic(1, 1, 1, 51),
		},
		{
			Code: "(() => { (0, eval)(code); globalThis.eval(code); })();", FileName: "case.js",
			Output: []string{"{ (0, eval)(code); globalThis.eval(code); }"},
			Errors: iifeDiagnostic(1, 1, 1, 54),
		},
		{
			Code: "(function () { console.log(import.meta.url); })();", FileName: "case.js",
			Output: []string{"{ console.log(import.meta.url); }"},
			Errors: iifeDiagnostic(1, 1, 1, 50),
		},
		{
			Code: "/** @type {Function} */ ((() => { run(); })());", FileName: "case.js",
			Output: []string{"/** @type {Function} */ { run(); }"},
			Errors: iifeDiagnostic(1, 26, 1, 46),
		},
		{
			Code: "(function () { render(<arguments arguments={1} />, <View.arguments />); })();", FileName: "case.tsx",
			Output: []string{"{ render(<arguments arguments={1} />, <View.arguments />); }"},
			Errors: iifeDiagnostic(1, 1, 1, 77),
		},
		{
			// Replacing the IIFE with its body would leave T unbound.
			Code: "(<T>() => { run<T>(); })<string>();", FileName: "case.ts",
			Output: []string{},
			Errors: iifeDiagnostic(1, 1, 1, 35),
		},
		{
			Code: "(() => { using resource = acquire(); use(resource); })();", FileName: "case.ts",
			Output: []string{"{ using resource = acquire(); use(resource); }"},
			Errors: iifeDiagnostic(1, 1, 1, 57),
		},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t,
		&prefer_block_statement_over_iife.PreferBlockStatementOverIIFERule, valid, invalid)
}

func TestPreferBlockStatementOverIIFELegacyTarget(t *testing.T) {
	root := fixtures.GetRootDir()
	root.FS = utils.NewOverlayVFS(root.FS, map[string]string{
		tspath.ResolvePath(root.Dir, "tsconfig.es5.json"): `{"extends":"./tsconfig.json","compilerOptions":{"target":"es5"}}`,
	})
	rule_tester.RunRuleTester(root, "tsconfig.es5.json", t,
		&prefer_block_statement_over_iife.PreferBlockStatementOverIIFERule, nil,
		[]rule_tester.InvalidTestCase{
			{
				Code: "export {};\n(() => { function example() {} example(); })();", FileName: "case.ts",
				Output: []string{},
				Errors: iifeDiagnostic(2, 1, 2, 47),
			},
			{
				// The declaration remains in the nested function after fixing.
				Code: "export {};\n(() => { use(() => { function nested() {} nested(); }); })();", FileName: "case.ts",
				Output: []string{"export {};\n{ use(() => { function nested() {} nested(); }); }"},
				Errors: iifeDiagnostic(2, 1, 2, 61),
			},
		})
}

func TestPreferBlockStatementOverIIFEEditDemand(t *testing.T) {
	const source = "(() => { run(); })();\n(/* keep */ () => {})();"
	const fixed = "{ run(); }\n(/* keep */ () => {})();"
	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, file, err := helper.CreateTestProgram(source, "edit-demand.js", "tsconfig.json")
	if err != nil {
		t.Fatal(err)
	}
	r := prefer_block_statement_over_iife.PreferBlockStatementOverIIFERule
	var all []rule.RuleDiagnostic
	for _, demand := range []rule.EditDemand{rule.EditDemandAll, rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion} {
		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program: lintprogram.NewFromCompiler(program),
			File:    file.FileName(),
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{
					Name: r.Name, Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners { return r.Run(ctx, nil) },
				}}
			},
			Consumer: rule.DiagnosticConsumer{
				Demand: demand,
				Report: func(diagnostic rule.RuleDiagnostic) { diagnostics = append(diagnostics, diagnostic) },
			},
		})
		if len(diagnostics) != 2 {
			t.Fatalf("demand %v: got %d diagnostics, want 2", demand, len(diagnostics))
		}
		if demand == rule.EditDemandAll {
			all = diagnostics
		}
		wantFix := demand == rule.EditDemandAll || demand == rule.EditDemandAutofix
		for i, got := range diagnostics {
			want := all[i]
			if got.Range != want.Range || got.RuleName != want.RuleName ||
				got.Severity != want.Severity || !reflect.DeepEqual(got.Message, want.Message) {
				t.Errorf("demand %v: changed diagnostic %d", demand, i)
			}
			if (got.FixesPtr != nil) != (i == 0 && wantFix) || got.Suggestions != nil {
				t.Errorf("demand %v: unexpected edits for diagnostic %d", demand, i)
			}
			if wantFix && !slices.Equal(got.Fixes(), want.Fixes()) {
				t.Errorf("demand %v: changed fixes for diagnostic %d", demand, i)
			}
		}
		wantOutput := source
		if wantFix {
			wantOutput = fixed
		}
		if output, _, applied := linter.ApplyRuleFixes(source, diagnostics); output != wantOutput || applied != wantFix {
			t.Errorf("demand %v: unexpected output %q", demand, output)
		}
	}
}
