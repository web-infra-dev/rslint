package prefer_modern_math_apis_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_modern_math_apis"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestPreferModernMathApisExtras(t *testing.T) {
	t.Run("AST boundaries and ignored patterns", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_modern_math_apis.PreferModernMathApisRule,
			[]rule_tester.ValidTestCase{
				{Code: "(Math.log as Function)(x) / Math.LN2;", FileName: "input.ts"},
				{Code: "(Math as typeof Math).log(x) / Math.LN2;", FileName: "input.ts"},
				{Code: "(Math?.log)(x) / Math.LN2;", FileName: "input.ts"},
				{Code: "(Math.log(x) as number) / Math.LN2;", FileName: "input.ts"},
				{Code: "Math.log(x) / (Math.LN2 as number);", FileName: "input.ts"},
				{Code: "Math.sqrt((x ** 2) as number);", FileName: "input.ts"},
				{Code: "Math.sqrt(x ** (2 as number));", FileName: "input.ts"},
				{Code: "Math.sqrt(x ** +2);", FileName: "input.ts"},
				{Code: "Math.sqrt(x * y);", FileName: "input.ts"},
				{Code: "Math.sqrt(obj[key()] * obj[key()]);", FileName: "input.ts"},
				{Code: "class Point { #x; f() { return Math.sqrt(this.#x * this.x); } }", FileName: "input.ts"},
				{Code: "Math.log(x) * Math.LN2; Math.LOG10E / Math.log(x);", FileName: "input.ts"},
				{Code: "Math.sqrt(-1 * -1);", FileName: "input.ts"},
				{Code: "Math.sqrt((+0) * (-0));", FileName: "input.ts"},
			}, []rule_tester.InvalidTestCase{})
	})
	t.Run("Reference comparisons and expression wrappers", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_modern_math_apis.PreferModernMathApisRule,
			[]rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
				{Code: "Math.sqrt(`4` * \"4\");", FileName: "input.ts",
					Output: []string{"Math.abs(`4`);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.sqrt(`4` * `4`);", FileName: "input.ts",
					Output: []string{"Math.abs(`4`);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.sqrt(obj[\"x\" + \"y\"] * obj.xy);", FileName: "input.ts",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.abs(obj[\"x\" + \"y\"]);"}}},
					}},
				{Code: "Math.sqrt(obj?.x * obj.x);", FileName: "input.ts",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.abs(obj?.x);"}}},
					}},
				{Code: "class Point { #x; f() { return Math.sqrt(this.#x * this.#x); } }", FileName: "input.ts",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 37, EndLine: 1, EndColumn: 41, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "class Point { #x; f() { return Math.abs(this.#x); } }"}}},
					}},
				{Code: "class Point extends Parent { f() { return Math.sqrt(super.x * super.x); } }", FileName: "input.ts",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 48, EndLine: 1, EndColumn: 52, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "class Point extends Parent { f() { return Math.abs(super.x); } }"}}},
					}},
				{Code: "Math.sqrt(/x/gi * /x/gi);", FileName: "input.ts",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.abs(/x/gi);"}}},
					}},
				{Code: "Math.sqrt(1n * 1n);", FileName: "input.ts",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.abs(1n);"}}},
					}},
				{Code: "Math.sqrt((value<number>) * value);", FileName: "input.ts",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.abs((value<number>));"}}},
					}},
				{Code: "Math.sqrt((value<number>).x * value.x);", FileName: "input.ts",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.abs((value<number>).x);"}}},
					}},
				{Code: "void (Math.log(value) / Math.LN2)<number>;", FileName: "input.ts",
					Output: []string{"void (Math.log2(value))<number>;"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 33},
					}},
				{Code: "const value = Infinity<number>; Math.sqrt(value ** 2);", FileName: "input.ts",
					Output: []string{"const value = Infinity<number>; Math.abs(value);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 38, EndLine: 1, EndColumn: 42},
					}},
				{Code: "((Math.sqrt))(x * x);", FileName: "input.ts",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "((Math.abs))(x);"}}},
					}},
				{Code: "function f(Math) { return Math.log(x) / Math.LN2; }", FileName: "input.ts",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 27, EndLine: 1, EndColumn: 49, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "function f(Math) { return Math.log2(x); }"}}},
					}},
				{Code: "function f(Infinity) { return Math.sqrt(Infinity ** 2); }", FileName: "input.ts",
					Errors: []rule_tester.InvalidTestCaseError{
						// A parameter named Infinity is not necessarily the builtin value.
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 36, EndLine: 1, EndColumn: 40, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "function f(Infinity) { return Math.abs(Infinity); }"}}},
					}},
			})
	})
	t.Run("Literal conversion and binding limits", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_modern_math_apis.PreferModernMathApisRule,
			[]rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
				{Code: "Math.log(\"\\uFEFF 1\\u00a0\") / Math.LN2;", FileName: "input.ts",
					Output: []string{"Math.log2(\"\\uFEFF 1\\u00a0\");"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 38},
					}},
				{Code: "Math.log(\"0x1\") / Math.LN2;", FileName: "input.ts",
					Output: []string{"Math.log2(\"0x1\");"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 27},
					}},
				{Code: "Math.log(\"+0x1\") / Math.LN2;", FileName: "input.ts",
					Output: []string{"Math.log2(\"+0x1\");"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 28},
					}},
				{Code: "Math.log(+\"1\") / Math.LN2;", FileName: "input.ts",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2(+\"1\");"}}},
					}},
				{Code: "Math.log(-true) / Math.LN2;", FileName: "input.ts",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2(-true);"}}},
					}},
				{Code: "Math.log(-`1`) / Math.LN2;", FileName: "input.ts",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2(-`1`);"}}},
					}},
				{Code: "Math.log(!Infinity) / Math.LN2;", FileName: "input.ts",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 31, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2(!Infinity);"}}},
					}},
				{Code: "let value = 1; Math.log(value) / Math.LN2;", FileName: "input.ts",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 16, EndLine: 1, EndColumn: 42, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "let value = 1; Math.log2(value);"}}},
					}},
				{Code: "const {value} = 1; Math.log(value) / Math.LN2;", FileName: "input.ts",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 20, EndLine: 1, EndColumn: 46, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const {value} = 1; Math.log2(value);"}}},
					}},
				{Code: "const [value] = 0; Math.sqrt(value ** 2);", FileName: "input.ts",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 25, EndLine: 1, EndColumn: 29, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const [value] = 0; Math.abs(value);"}}},
					}},
				{Code: "Math.sqrt(false * false);", FileName: "input.ts",
					Output: []string{"Math.abs(false);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.sqrt(null * null);", FileName: "input.ts",
					Output: []string{"Math.abs(null);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.sqrt(0x10 * 16);", FileName: "input.ts",
					Output: []string{"Math.abs(0x10);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.sqrt(1_000 * 1000);", FileName: "input.ts",
					Output: []string{"Math.abs(1_000);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "Math.sqrt(obj[\"x\" as string] * obj.x);", FileName: "input.ts",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.abs(obj[\"x\" as string]);"}}},
					}},
			})
	})
	t.Run("Comments and UTF-16 diagnostic ranges", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_modern_math_apis.PreferModernMathApisRule,
			[]rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
				{Code: "const 名 = \"😀\"; Math.sqrt(名 * 名);", FileName: "input.js",
					Output: []string{"const 名 = \"😀\"; Math.abs(名);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 22, EndLine: 1, EndColumn: 26},
					}},
				{Code: "const 名 = \"😀\"; Math.log(名) / Math.LN2;", FileName: "input.js",
					Output: []string{"const 名 = \"😀\"; Math.log2(名);"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 17, EndLine: 1, EndColumn: 39},
					}},
				{Code: "Math.sqrt(\n  (first ** 2)\n    + (second ** 2)\n);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot(\n  (first)\n, (second)\n);"}}},
					}},
				{Code: "Math.sqrt((first ** 2 + second ** 2) + third ** 2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot(first, second, third);"}}},
					}},
				{Code: "Math.sqrt(first ** 2 + second ** 2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot(first, second);"}}},
					}},
				{Code: "Math.sqrt((/*keep*/ first ** 2 + second ** 2));", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot(/*keep*/ first, second);"}}},
					}},
				{Code: "Math.sqrt(first /*keep*/ * first);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
				{Code: "const result = Math.sqrt(0 * 0 /*keep*/) + 1;", FileName: "input.js",
					Output: []string{"const result = Math.abs(0 /*keep*/) + 1;"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 21, EndLine: 1, EndColumn: 25},
					}},
				{Code: "/*before*/ Math.log((/*inside*/ 1)) / Math.LN2; /*after*/", FileName: "input.js",
					Output: []string{"/*before*/ Math.log2((/*inside*/ 1)); /*after*/"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 12, EndLine: 1, EndColumn: 47},
					}},
				{Code: "/** @type {number} */ (Math.log(value) / Math.LN2);", FileName: "input.js",
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 24, EndLine: 1, EndColumn: 50, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "/** @type {number} */ (Math.log2(value));"}}},
					}},
				{Code: "void /** @type {number} */ (Math.log(value) / Math.LN2);", FileName: "input.js",
					Output: []string{"void /** @type {number} */ (Math.log2(value));"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 29, EndLine: 1, EndColumn: 55},
					}},
				{Code: "Math.sqrt(/** @type {number} */ (4) * 4);", FileName: "input.js",
					Output: []string{"Math.abs(/** @type {number} */ (4));"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
					}},
			})
	})
}

func TestPreferModernMathApisEditDemand(t *testing.T) {
	source := strings.Join([]string{
		"Math.log(1) / Math.LN2;",
		"Math.log(value) / Math.LN2;",
		"Math.sqrt(4 * 4);",
		"Math.sqrt(first ** 2 + second ** 2);",
		"Math.sqrt(4 * /* keep */ 4);",
		"Math.log(/* keep */ 1) / Math.LN2;",
	}, "\n")
	root := fixtures.GetRootDir()
	fileName := tspath.ResolvePath(root.Dir, "edit-demand.js")
	fs := utils.NewOverlayVFS(root.FS, map[string]string{fileName: source})
	host := utils.CreateCompilerHost(root.Dir, fs)
	compilerProgram, err := utils.CreateProgram(true, fs, root.Dir, "tsconfig.json", host)
	if err != nil {
		t.Fatal(err)
	}
	if compilerProgram.GetSourceFile(fileName) == nil {
		t.Fatal("missing edit-demand source file")
	}
	results := make(map[rule.EditDemand][]rule.RuleDiagnostic)
	for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program: program.NewFromCompiler(compilerProgram), File: fileName,
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{Name: prefer_modern_math_apis.PreferModernMathApisRule.Name, Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return prefer_modern_math_apis.PreferModernMathApisRule.Run(ctx, nil)
					},
				}}
			},
			Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) { results[demand] = append(results[demand], d) }},
		})
		if len(results[demand]) != 6 {
			t.Fatalf("demand %d: got %d diagnostics, want 6", demand, len(results[demand]))
		}
	}
	for demand, diagnostics := range results {
		for i, d := range diagnostics {
			metadata := d
			metadata.FixesPtr, metadata.Suggestions = nil, nil
			if !reflect.DeepEqual(metadata, results[rule.EditDemandNone][i]) {
				t.Errorf("demand %d, diagnostic %d changed metadata", demand, i)
			}
			wantsFix := demand&rule.EditDemandAutofix != 0 && (i == 0 || i == 2)
			wantsSuggestion := demand&rule.EditDemandSuggestion != 0 && (i == 1 || i == 3)
			if (d.FixesPtr != nil) != wantsFix || (d.Suggestions != nil) != wantsSuggestion {
				t.Fatalf("demand %d, diagnostic %d has unexpected edits: %#v", demand, i, d)
			}
			if wantsFix && !reflect.DeepEqual(d.FixesPtr, results[rule.EditDemandAll][i].FixesPtr) {
				t.Errorf("demand %d changed fixes", demand)
			}
			if wantsSuggestion && !reflect.DeepEqual(d.Suggestions, results[rule.EditDemandAll][i].Suggestions) {
				t.Errorf("demand %d changed suggestions", demand)
			}
		}
	}
	for _, demand := range []rule.EditDemand{rule.EditDemandAutofix, rule.EditDemandAll} {
		output, _, fixed := linter.ApplyRuleFixes(source, results[demand])
		want := strings.ReplaceAll(strings.ReplaceAll(source, "Math.log(1) / Math.LN2", "Math.log2(1)"), "Math.sqrt(4 * 4)", "Math.abs(4)")
		if !fixed || output != want {
			t.Errorf("demand %d: output = %q, want %q", demand, output, want)
		}
	}
	for _, demand := range []rule.EditDemand{rule.EditDemandSuggestion, rule.EditDemandAll} {
		for _, test := range []struct {
			index                 int
			method, before, after string
		}{
			{1, "log2", "Math.log(value) / Math.LN2", "Math.log2(value)"},
			{3, "hypot", "Math.sqrt(first ** 2 + second ** 2)", "Math.hypot(first, second)"},
		} {
			suggestions := *results[demand][test.index].Suggestions
			replacement := "Math." + test.method + "(…)"
			wantMessage := rule.RuleMessage{Id: "suggestion", Description: "Replace with `" + replacement + "`.", Data: map[string]string{"replacement": replacement}}
			if len(suggestions) != 1 || !reflect.DeepEqual(suggestions[0].Message, wantMessage) {
				t.Fatalf("demand %d: unexpected suggestions: %#v", demand, suggestions)
			}
			output, _, fixed := linter.ApplyRuleFixes(source, suggestions)
			if !fixed || output != strings.Replace(source, test.before, test.after, 1) {
				t.Errorf("demand %d: suggestion output = %q", demand, output)
			}
		}
	}
}

func TestPreferModernMathApisJSDoc(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_modern_math_apis.PreferModernMathApisRule, nil, []rule_tester.InvalidTestCase{
		{Code: "Math.sqrt(/** @type {number} */ (a ** 2 + b ** 2));", FileName: "jsdoc.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot(/** @type {number} */ a, b);"}}},
		}},
		{Code: "Math.sqrt((/** @type {number} */ (a ** 2 + b ** 2)) + c ** 2);", FileName: "jsdoc.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot(/** @type {number} */ a, b, c);"}}},
		}},
		{Code: "Math.log(/** @type {number} */ (1)) / Math.LN2;", FileName: "jsdoc.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 47},
		}},
	})
}

// Additional AST shapes and conservative fixes for shadowed builtin names.
func TestPreferModernMathApisBoundaries(t *testing.T) {
	t.Run("heritage, context and binding shapes", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_modern_math_apis.PreferModernMathApisRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
			{Code: "class C extends Math.sqrt(value ** 2) {}", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 22, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "class C extends Math.abs(value) {}"}}},
				}},
			{Code: "class C extends (Math.log(value) / Math.LN2) {}", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 18, EndLine: 1, EndColumn: 44, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "class C extends (Math.log2(value)) {}"}}},
				}},
			{Code: "class C { static value = Math.sqrt(4 * 4); constructor(public x = Math.log(1) / Math.LN2) {} }", FileName: "input.ts",
				Output: []string{"class C { static value = Math.abs(4); constructor(public x = Math.log2(1)) {} }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 31, EndLine: 1, EndColumn: 35},
					{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 67, EndLine: 1, EndColumn: 89},
				}},
			{Code: "const value = 1; (function value(){ return Math.log(value) / Math.LN2; })();", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 44, EndLine: 1, EndColumn: 70, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const value = 1; (function value(){ return Math.log2(value); })();"}}},
				}},
			{Code: "const value = 1; (class value { static result = Math.log(value) / Math.LN2; });", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 49, EndLine: 1, EndColumn: 75, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const value = 1; (class value { static result = Math.log2(value); });"}}},
				}},
			{Code: "const value = 1; try {} catch (value) { Math.log(value) / Math.LN2; }", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 41, EndLine: 1, EndColumn: 67, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const value = 1; try {} catch (value) { Math.log2(value); }"}}},
				}},
			{Code: "const value = 1; function f({value}) { return Math.log(value) / Math.LN2; }", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 47, EndLine: 1, EndColumn: 73, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const value = 1; function f({value}) { return Math.log2(value); }"}}},
				}},
			{Code: "const value = 1; for (const value of values) { Math.log(value) / Math.LN2; }", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 48, EndLine: 1, EndColumn: 74, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const value = 1; for (const value of values) { Math.log2(value); }"}}},
				}},
			{Code: "namespace A { const value = 1; Math.log(value) / Math.LN2; }", FileName: "input.ts",
				Output: []string{"namespace A { const value = 1; Math.log2(value); }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 32, EndLine: 1, EndColumn: 58},
				}},
			{Code: "const value = 1; type value = string; Math.log(value) / Math.LN2;", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 39, EndLine: 1, EndColumn: 65, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const value = 1; type value = string; Math.log2(value);"}}},
				}},
			{Code: "function f(undefined) { return Math.log(undefined) / Math.LN2; }", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 32, EndLine: 1, EndColumn: 62, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "function f(undefined) { return Math.log2(undefined); }"}}},
				}},
			{Code: "function f(NaN) { return Math.sqrt(NaN ** 2); }", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 31, EndLine: 1, EndColumn: 35, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "function f(NaN) { return Math.abs(NaN); }"}}},
				}},
			{Code: "function f(Number) { return Math.log(Number.NaN) / Math.LN2; }", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 29, EndLine: 1, EndColumn: 60, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "function f(Number) { return Math.log2(Number.NaN); }"}}},
				}},
			{Code: "function f(Math) { return Math.sqrt(4 * 4); }", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 32, EndLine: 1, EndColumn: 36, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "function f(Math) { return Math.abs(4); }"}}},
				}},
		})
	})
	t.Run("JSDoc and spelling boundaries", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_modern_math_apis.PreferModernMathApisRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
			{Code: "Math.sqrt(/** @satisfies {number} */ (a ** 2 + b ** 2));", FileName: "input.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot(/** @satisfies {number} */ a, b);"}}},
				}},
			{Code: "Math.sqrt(/** @type {number} */ ((a ** 2 + b ** 2)));", FileName: "input.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot(/** @type {number} */ a, b);"}}},
				}},
			{Code: "Math.sqrt((a ** 2 + b ** 2) + /** @type {number} */ (c ** 2 + d ** 2));", FileName: "input.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot(a, b, /** @type {number} */ c, d);"}}},
				}},
			{Code: "void /** @type {number} */ (Math.log(value) / Math.LN2);", FileName: "input.js",
				Output: []string{"void /** @type {number} */ (Math.log2(value));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 29, EndLine: 1, EndColumn: 55},
				}},
			{Code: "Math.sqrt(a ** 0x2 + b ** 2.0 + c ** 2e0);", FileName: "input.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot(a, b, c);"}}},
				}},
			{Code: "Math.sqrt((a ** 2 /* tail */) + b ** 2);", FileName: "input.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.hypot((a /* tail */), b);"}}},
				}},
			{Code: "Math.sqrt(/** @type {number} */ (0 ** 2) + 0 ** 2);", FileName: "input.js",
				Output: []string{"Math.hypot(/** @type {number} */ (0), 0);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.hypot(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
				}},
		})
	})
	t.Run("computed keys and reference equality", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_modern_math_apis.PreferModernMathApisRule, []rule_tester.ValidTestCase{
			{Code: "Math.sqrt(obj[String.raw`x`] * obj.x);", FileName: "input.ts"},
		}, []rule_tester.InvalidTestCase{
			{Code: "Math.sqrt(obj[(1, 2)] * obj[2]);", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.abs(obj[(1, 2)]);"}}},
				}},
			{Code: "Math.sqrt(obj[true ? \"x\" : \"y\"] * obj.x);", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.abs(obj[true ? \"x\" : \"y\"]);"}}},
				}},
			{Code: "Math.sqrt(obj[(\"x\" satisfies string)] * obj.x);", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.abs(obj[(\"x\" satisfies string)]);"}}},
				}},
			{Code: "Math.sqrt((2.0) * 0x2);", FileName: "input.ts",
				Output: []string{"Math.abs((2.0));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10},
				}},
			{Code: "Math.sqrt(this * this);", FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.abs(this);"}}},
				}},
		})
	})
}

func TestPreferModernMathApisGlobals(t *testing.T) {
	t.Run("disabled Math", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_modern_math_apis.PreferModernMathApisRule, nil, []rule_tester.InvalidTestCase{
			{Code: "Math.sqrt(4 * 4);", Globals: map[string]any{"Math": "off"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.abs(4);"}}},
			}},
		})
	})
	t.Run("disabled Number", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_modern_math_apis.PreferModernMathApisRule, nil, []rule_tester.InvalidTestCase{
			{Code: "Math.log(Number.NaN) / Math.LN2;", Globals: map[string]any{"Number": "off"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 32, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2(Number.NaN);"}}},
			}},
		})
	})
	t.Run("disabled Infinity", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_modern_math_apis.PreferModernMathApisRule, nil, []rule_tester.InvalidTestCase{
			{Code: "Math.sqrt(Infinity ** 2);", Globals: map[string]any{"Infinity": "off"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.abs(Infinity);"}}},
			}},
		})
	})
	t.Run("disabled NaN", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_modern_math_apis.PreferModernMathApisRule, nil, []rule_tester.InvalidTestCase{
			{Code: "Math.sqrt(NaN ** 2);", Globals: map[string]any{"NaN": "off"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.abs(NaN);"}}},
			}},
		})
	})
	t.Run("disabled undefined", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_modern_math_apis.PreferModernMathApisRule, nil, []rule_tester.InvalidTestCase{
			{Code: "Math.log(undefined) / Math.LN2;", Globals: map[string]any{"undefined": "off"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.log2(…)` over `Math.log(…) / Math.LN2`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 31, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "Math.log2(undefined);"}}},
			}},
		})
	})
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_modern_math_apis.PreferModernMathApisRule, nil, []rule_tester.InvalidTestCase{
		{Code: "type Math = unknown; Math.sqrt(4 * 4);", FileName: "global.ts", Output: []string{"type Math = unknown; Math.abs(4);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 27, EndLine: 1, EndColumn: 31},
		}},
		{Code: "type Number = unknown; Math.sqrt(Number.NaN ** 2);", FileName: "global.ts", Output: []string{"type Number = unknown; Math.abs(Number.NaN);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 29, EndLine: 1, EndColumn: 33},
		}},
		{Code: "const Infinity = 4; Math.sqrt(Infinity * Infinity);", FileName: "global.ts", Output: []string{"const Infinity = 4; Math.abs(Infinity);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 26, EndLine: 1, EndColumn: 30},
		}},
		{Code: "const value = Infinity; function f(Infinity) { return Math.sqrt(value ** 2); }", FileName: "global.ts", Output: []string{"const value = Infinity; function f(Infinity) { return Math.abs(value); }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 60, EndLine: 1, EndColumn: 64},
		}},
		{Code: "/** @typedef {any} Math */ Math.sqrt(4 * 4);", FileName: "global.js", Output: []string{"/** @typedef {any} Math */ Math.abs(4);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-modern-math-apis", Message: "Prefer `Math.abs(…)` over `Math.sqrt(…)`.", Line: 1, Column: 33, EndLine: 1, EndColumn: 37},
		}},
	})
}
