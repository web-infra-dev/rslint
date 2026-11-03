// Additional AST and edit coverage checked against Unicorn v79.0.0.
package no_useless_fallback_in_spread_test

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_useless_fallback_in_spread"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUselessFallbackInSpreadPreservesJSDocTypes(t *testing.T) {
	for _, testCase := range []struct{ name, expression, fixed string }{
		{"asserted fallback", "/** @type {object} */ (value %s {})", "/** @type {object} */ (value)"},
		{"nested grouping", "((/** @type {object} */ (((value %s {})))))", "/** @type {object} */ (value)"},
		{"nested casts", "/** @type {object} */ (/** @type {unknown} */ (value %s {}))", "/** @type {object} */ (/** @type {unknown} */ (value))"},
		{"asserted operand", "(/** @type {object} */ (value) %s {})", "/** @type {object} */ (value)"},
	} {
		for _, operator := range []string{"||", "??"} {
			t.Run(testCase.name+"/"+operator, func(t *testing.T) {
				const body = "// @ts-check\n/** @param {unknown} value */\nfunction copy(value) {\n  return {...%s};\n}\n"
				code := fmt.Sprintf(body, fmt.Sprintf(testCase.expression, operator))
				helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
				program, sourceFile, err := helper.CreateTestProgram(code, "check-types.js", "tsconfig.json")
				if err != nil {
					t.Fatal(err)
				}
				if diagnostics := program.GetSemanticDiagnostics(context.Background(), sourceFile); len(diagnostics) != 0 {
					t.Fatalf("input has %d checkJs diagnostics", len(diagnostics))
				}
				var diagnostics []rule.RuleDiagnostic
				linter.LintSingleFile(linter.LintSingleFileOptions{
					Program: lintprogram.NewFromCompiler(program), File: sourceFile.FileName(),
					GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
						return []rule.ConfiguredRule{{Name: no_useless_fallback_in_spread.NoUselessFallbackInSpreadRule.Name, Severity: rule.SeverityError,
							Run: func(ctx rule.RuleContext) rule.RuleListeners {
								return no_useless_fallback_in_spread.NoUselessFallbackInSpreadRule.Run(ctx, nil)
							},
						}}
					},
					Consumer: rule.DiagnosticConsumer{Demand: rule.EditDemandAll, Report: func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }},
				})
				if len(diagnostics) != 1 {
					t.Fatalf("got %d rule diagnostics, want 1", len(diagnostics))
				}
				output, _, fixed := linter.ApplyRuleFixes(code, diagnostics)
				if !fixed || output != fmt.Sprintf(body, testCase.fixed) {
					t.Errorf("unexpected fixed output:\n%s", output)
				}
				fixedProgram, fixedFile, err := helper.CreateTestProgram(output, "check-types.js", "tsconfig.json")
				if err != nil {
					t.Fatal(err)
				}
				if diagnostics := fixedProgram.GetSemanticDiagnostics(context.Background(), fixedFile); len(diagnostics) != 0 {
					for _, diagnostic := range diagnostics {
						t.Errorf("fix introduced checkJs diagnostic %d", diagnostic.Code())
					}
				}
			})
		}
	}
}

func TestNoUselessFallbackInSpreadArtifactsFollowDemand(t *testing.T) {
	for _, testCase := range []struct{ code, output string }{
		{"const object = {...(foo ?? {})};", "const object = {...foo};"},
		{"const object = {.../** @type {object} */ (value || {})};", "const object = {.../** @type {object} */ (value)};"},
		{"const object = {...(foo ?? {/* keep */})};", "const object = {...(foo ?? {/* keep */})};"},
	} {
		t.Run(testCase.code, func(t *testing.T) {
			helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
			program, sourceFile, err := helper.CreateTestProgram(testCase.code, "edit-demand.js", "tsconfig.json")
			if err != nil {
				t.Fatal(err)
			}
			diagnostics := make(map[rule.EditDemand]rule.RuleDiagnostic, 4)
			for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
				var got []rule.RuleDiagnostic
				linter.LintSingleFile(linter.LintSingleFileOptions{
					Program: lintprogram.NewFromCompiler(program), File: sourceFile.FileName(),
					GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
						return []rule.ConfiguredRule{{Name: no_useless_fallback_in_spread.NoUselessFallbackInSpreadRule.Name, Severity: rule.SeverityError,
							Run: func(ctx rule.RuleContext) rule.RuleListeners {
								return no_useless_fallback_in_spread.NoUselessFallbackInSpreadRule.Run(ctx, nil)
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
			// Applying fixes sorts their slice in place; keep the reference intact.
			allFixes := slices.Clone(all.Fixes())
			for demand, diagnostic := range diagnostics {
				if diagnostic.Range != baseline.Range || !reflect.DeepEqual(diagnostic.Message, baseline.Message) || diagnostic.Severity != baseline.Severity {
					t.Errorf("demand %d changed diagnostic identity", demand)
				}
				wantFix := testCase.output != testCase.code && (demand == rule.EditDemandAutofix || demand == rule.EditDemandAll)
				if (diagnostic.FixesPtr != nil) != wantFix || (wantFix && !reflect.DeepEqual(diagnostic.Fixes(), allFixes)) {
					t.Errorf("demand %d: unexpected autofix artifacts", demand)
				}
				wantOutput := testCase.code
				if wantFix {
					wantOutput = testCase.output
				}
				output, _, fixed := linter.ApplyRuleFixes(testCase.code, []rule.RuleDiagnostic{diagnostic})
				if output != wantOutput || fixed != wantFix {
					t.Errorf("demand %d: unexpected autofix %q", demand, output)
				}
				if diagnostic.Suggestions != nil {
					t.Errorf("demand %d produced unexpected suggestions", demand)
				}
			}
		})
	}
}

func TestNoUselessFallbackInSpreadExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_useless_fallback_in_spread.NoUselessFallbackInSpreadRule, []rule_tester.ValidTestCase{
		// AST boundaries, comments, precedence, and positions
		{Code: "const object = {...(foo ||= {})};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const object = {...(foo ??= {})};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const object = {...(foo ?? [])};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const object = {...(foo || Object())};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const object = {...((foo || {}) as object)};", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const object = {...((foo ?? {}) satisfies object)};", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const object = {...(foo || {})!};", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const object = {...(foo || ({} as object))};", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const element = <div {...(foo || {})} />;", FileName: "case.jsx", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const object = {...(foo ?? { [key]: 1 })};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Authored TS wrappers and sequence expressions remain significant.
		{Code: "const object = {...(a || ({} satisfies object)), ...(b ?? ({}!))};", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const object = {...(foo || ({}, {}))};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
	}, []rule_tester.InvalidTestCase{
		// AST boundaries, comments, precedence, and positions
		{Code: "const object = {...foo || {}};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {...foo};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 27, EndLine: 1, EndColumn: 29, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...foo ?? {}};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {...foo};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 27, EndLine: 1, EndColumn: 29, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const 名 = \"😀\"; const object = {...(foo || {})};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const 名 = \"😀\"; const object = {...foo};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 44, EndLine: 1, EndColumn: 46, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...(\n  value ??\n  {}\n)};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {...\n  value\n};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 3, Column: 3, EndLine: 3, EndColumn: 5, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...(foo?.bar ?? {})};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {...foo?.bar};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 33, EndLine: 1, EndColumn: 35, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...(foo?.[key]?.() || {})};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {...foo?.[key]?.()};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 39, EndLine: 1, EndColumn: 41, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "class C { #value; method() { return {...(this.#value || {})}; } }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"class C { #value; method() { return {...this.#value}; } }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 57, EndLine: 1, EndColumn: 59, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const element = <div value={{...(foo || {})}} />;", FileName: "case.jsx", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const element = <div value={{...foo}} />;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 41, EndLine: 1, EndColumn: 43, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...(foo as object ?? {})};", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {...foo as object};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 38, EndLine: 1, EndColumn: 40, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...((foo satisfies object) || {})};", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {...(foo satisfies object)};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 47, EndLine: 1, EndColumn: 49, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...(foo! ?? {})};", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {...foo!};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 29, EndLine: 1, EndColumn: 31, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...(<object>foo || {})};", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {...<object>foo};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 36, EndLine: 1, EndColumn: 38, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...(factory<object>() ?? {})};", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {...factory<object>()};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 42, EndLine: 1, EndColumn: 44, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...(/** @type {object} */ (foo) || {})};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {.../** @type {object} */ (foo)};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 52, EndLine: 1, EndColumn: 54, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {.../** @type {object} */ (foo || {})};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {.../** @type {object} */ (foo)};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 50, EndLine: 1, EndColumn: 52, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...(foo || /** @type {object} */ ({}))};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 51, EndLine: 1, EndColumn: 53, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...(/* before */ (foo ?? {}) /* after */)};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {.../* before */ foo /* after */};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 42, EndLine: 1, EndColumn: 44, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...((foo // keep\n) ?? {})};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {...(foo // keep\n)};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 2, Column: 6, EndLine: 2, EndColumn: 8, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...(foo || (/* keep */ {}))};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 40, EndLine: 1, EndColumn: 42, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...((first(), second) || {})};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {...(first(), second)};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 42, EndLine: 1, EndColumn: 44, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...(foo || {}) ?? {}};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {...(foo || {})};", "const object = {...foo};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 35, EndLine: 1, EndColumn: 37, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...(a ?? {...(b || {})}), ...(c || {})};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {...(a ?? {...b}), ...c};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 36, EndLine: 1, EndColumn: 38, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 52, EndLine: 1, EndColumn: 54, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...(foo || {} /* keep */)};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {...foo /* keep */};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 28, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		// JSX/yield operands, JSDoc satisfies, Unicode line breaks, and deep grouping.
		{Code: "const object = {.../** @satisfies {object} */ ((foo) ?? {})};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {.../** @satisfies {object} */ (foo)};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 57, EndLine: 1, EndColumn: 59, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...(/** @satisfies {object} */ (foo) || {})};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {.../** @satisfies {object} */ (foo)};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 57, EndLine: 1, EndColumn: 59, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "function* merge() { return {...((yield getValue()) || {})}; }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"function* merge() { return {...(yield getValue())}; }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 55, EndLine: 1, EndColumn: 57, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...(<Widget /> || {})};", FileName: "case.jsx", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {...<Widget />};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 35, EndLine: 1, EndColumn: 37, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...(factory<object> ?? {})};", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {...factory<object>};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 40, EndLine: 1, EndColumn: 42, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...((foo // keep\u2028) ?? {})};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {...(foo // keep\u2028)};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 2, Column: 6, EndLine: 2, EndColumn: 8, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...((((((((((((((((((((((((((((((((/* before */ foo ?? {} /* after */))))))))))))))))))))))))))))))))};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {.../* before */ foo /* after */};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 72, EndLine: 1, EndColumn: 74, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
	})
}
