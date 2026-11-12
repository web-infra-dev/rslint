package prefer_math_constants_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_math_constants"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestPreferMathConstantsExtras(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// Written precision matters even when the parsed value is the same.
		{Code: "[3.140, 2.7180, 3.1415926535897930, 314e-2, 0o10, 3.];", FileName: "file.js"},
		// A binding element stores the non-computed key separately from its name.
		{Code: "const {3.14: radius} = values;", FileName: "file.js"},
		{Code: "({3.14: radius} = values);", FileName: "file.js"},
		{Code: "type Value = Shape[3.14]; type Shape = { 3.14(): void };", FileName: "file.ts"},
		{Code: "/** @type {-3.14} */ let angle;", FileName: "file.js"},
	}
	compactReturn := approximationCase("function f() { return.6931; }", ".6931", "Math.LN2", "file.js")
	compactReturn.Errors[0].Suggestions[0].Output = "function f() { return Math.LN2; }"
	invalid := []rule_tester.InvalidTestCase{
		compactReturn,
		approximationCase("const {[(-3.14)]: radius} = values;", "3.14", "Math.PI", "file.js"),
		approximationCase("class Shape { #radius = +3.14; }", "3.14", "Math.PI", "file.js"),
		approximationCase("class Shape { get [3.14]() { return 1; } }", "3.14", "Math.PI", "file.ts"),
		approximationCase("const angle = { value: 3.14 } as const;", "3.14", "Math.PI", "file.ts"),
		approximationCase("enum Angles { Pi = 3.14, Half = 1 }", "3.14", "Math.PI", "file.ts"),
		{
			Code:     "const result = (\n  \"😀\", object?.[(.693_1)]\n);",
			FileName: "file.js",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "prefer-math-constants/error",
				Message:   "Prefer `Math.LN2` over the approximate value `.693_1`.",
				Line:      2,
				Column:    19,
				EndLine:   2,
				EndColumn: 25,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "prefer-math-constants/suggestion",
					Output:    "const result = (\n  \"😀\", object?.[(Math.LN2)]\n);",
				}},
			}},
		},
	}

	// These cases still report the approximation, but applying the suggestion
	// would select the wrong Math or break TypeScript's constant requirements.
	disabledMath := approximationCase("const pi = 3.14;", "3.14", "Math.PI", "file.js")
	disabledMath.Globals = map[string]any{"Math": "off"}
	for _, test := range []rule_tester.InvalidTestCase{
		disabledMath,
		approximationCase("function f(Math) { return 3.14; }", "3.14", "Math.PI", "file.js"),
		approximationCase("import { CustomMath as Math } from 'library'; const angle = 3.14;", "3.14", "Math.PI", "file.ts"),
		approximationCase("const angle = ((-3.14)) as const;", "3.14", "Math.PI", "file.ts"),
		approximationCase("const angle = <const>.6931;", ".6931", "Math.LN2", "file.ts"),
		approximationCase("const angle = /** @type {const} */ (3.14);", "3.14", "Math.PI", "file.js"),
		approximationCase("const enum Angles { Tau = 2 * 3.14 }", "3.14", "Math.PI", "file.ts"),
		approximationCase("enum Angles { Pi = 3.14, Tau }", "3.14", "Math.PI", "file.ts"),
		approximationCase("declare const angle = 3.14;", "3.14", "Math.PI", "file.ts"),
		approximationCase("interface Angles { [-3.14](): void }", "3.14", "Math.PI", "file.ts"),
		approximationCase("type Angles = { [3.14]: number };", "3.14", "Math.PI", "file.ts"),
	} {
		test.Errors[0].Suggestions = nil
		invalid = append(invalid, test)
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_math_constants.PreferMathConstantsRule, valid, invalid)
}

func TestPreferMathConstantsEditDemand(t *testing.T) {
	const source = "function f() { return.6931; }"
	const replacement = "function f() { return Math.LN2; }"
	root := fixtures.GetRootDir()
	fileName := tspath.ResolvePath(root.Dir, "edit-demand.js")
	fs := utils.NewOverlayVFS(root.FS, map[string]string{fileName: source})
	compiler, err := utils.CreateProgram(true, fs, root.Dir, "tsconfig.json", utils.CreateCompilerHost(root.Dir, fs))
	if err != nil {
		t.Fatal(err)
	}
	if compiler.GetSourceFile(fileName) == nil {
		t.Fatal("missing edit-demand source file")
	}
	r := prefer_math_constants.PreferMathConstantsRule
	lint := func(demand rule.EditDemand) rule.RuleDiagnostic {
		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program: program.NewFromCompiler(compiler), File: fileName,
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{Name: r.Name, Severity: rule.SeverityError, Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return r.Run(ctx, nil)
				}}}
			},
			Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, d)
			}},
		})
		if len(diagnostics) != 1 {
			t.Fatalf("demand %d: got %d diagnostics, want 1", demand, len(diagnostics))
		}
		return diagnostics[0]
	}
	all := lint(rule.EditDemandAll)
	if all.FixesPtr != nil || all.Suggestions == nil || len(*all.Suggestions) != 1 {
		t.Fatalf("expected one suggestion and no autofix, got %#v", all)
	}
	suggestions := *all.Suggestions
	if got := suggestions[0].Message.Description; got != "Replace `.6931` with `Math.LN2`." {
		t.Fatalf("unexpected suggestion message %q", got)
	}
	if output, _, applied := linter.ApplyRuleFixes(source, suggestions); !applied || output != replacement {
		t.Fatalf("suggestion produced %q, want %q", output, replacement)
	}
	if got := lint(rule.EditDemandSuggestion); !reflect.DeepEqual(got, all) {
		t.Fatalf("suggestion-only demand changed diagnostic: %#v", got)
	}
	all.Suggestions = nil
	for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix} {
		if got := lint(demand); !reflect.DeepEqual(got, all) {
			t.Errorf("demand %d: unexpected diagnostic %#v", demand, got)
		}
	}
}
