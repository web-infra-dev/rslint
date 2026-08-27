// TestNoUnnecessaryTemplateExpressionExtras locks in autofix branches and
// edge shapes that the upstream test suite does not exercise. Each case names
// the template-literal boundary it protects; the migrated upstream suite lives
// in no_unnecessary_template_expression_upstream_test.go.
package no_unnecessary_template_expression

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnnecessaryTemplateExpressionExtras(t *testing.T) {
	errorWithMessage := func() rule_tester.InvalidTestCaseError {
		return rule_tester.InvalidTestCaseError{MessageId: "noUnnecessaryTemplateExpression"}
	}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnnecessaryTemplateExpressionRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: every ECMAScript line terminator preserves intentional trailing whitespace ----
			{Code: "`trailing whitespace: ${' '}\r`;"},
			{Code: "`trailing whitespace: ${' '}\u2028`;"},
			{Code: "`trailing whitespace: ${' '}\u2029`;"},
		},
		[]rule_tester.InvalidTestCase{
			// Locks in upstream getReportDescriptors() literal arm: numeric and bigint values use JavaScript String conversion.
			{
				Code:   "`${0o25}-${0b1010}-${0x25}-${1n}`;",
				Output: []string{"`21-10-37-1`;"},
				Errors: []rule_tester.InvalidTestCaseError{
					errorWithMessage(), errorWithMessage(), errorWithMessage(), errorWithMessage(),
				},
			},
			// Locks in upstream getReportDescriptors() escaping arm: an unescaped backtick remains template text.
			{
				Code:   "const value = `back${'`'}tick`;",
				Output: []string{"const value = `back\\`tick`;"},
				Errors: []rule_tester.InvalidTestCaseError{errorWithMessage()},
			},
			// Locks in upstream getReportDescriptors() escaping arm: `${` cannot open a nested interpolation.
			{
				Code:   "const value = `dollar${'${x}'}sign`;",
				Output: []string{"const value = `dollar\\${x}sign`;"},
				Errors: []rule_tester.InvalidTestCaseError{errorWithMessage()},
			},
			// Locks in upstream endsWithUnescapedDollarSign() adjacency arm.
			{
				Code:   "const value = ` ${'$'}{} `;",
				Output: []string{"const value = ` \\${} `;"},
				Errors: []rule_tester.InvalidTestCaseError{errorWithMessage()},
			},
			// Locks in upstream getReportDescriptors() nested-template arm.
			{
				Code:   "const value = `use${`less`}`;",
				Output: []string{"const value = `useless`;"},
				Errors: []rule_tester.InvalidTestCaseError{errorWithMessage()},
			},
			// ---- Dimension 4: adjacent literal kinds share one template and produce non-overlapping fixes ----
			{
				Code:   "const value = `${'a'} ${true} ${/a/}`;",
				Output: []string{"const value = `a true /a/`;"},
				Errors: []rule_tester.InvalidTestCaseError{
					errorWithMessage(), errorWithMessage(), errorWithMessage(),
				},
			},
			// ---- Dimension 4: template literal type interpolation ----
			{
				Code:   "type Value = `pre-${'suffix'}`;",
				Output: []string{"type Value = `pre-suffix`;"},
				Errors: []rule_tester.InvalidTestCaseError{errorWithMessage()},
			},
			// Locks in upstream reportSingleInterpolation() weak-precedence parenthesizing arm.
			{
				Code:   "declare const value: string; `${value || ''}`.length;",
				Output: []string{"declare const value: string; (value || '').length;"},
				Errors: []rule_tester.InvalidTestCaseError{errorWithMessage()},
			},
		},
	)
}

func TestNoUnnecessaryTemplateExpressionEditDemand(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name string
		code string
	}{
		{name: "single interpolation", code: "declare const value: string; `${value}`;"},
		{name: "literal interpolation", code: "`${1}`;"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
			program, sourceFile, err := helper.CreateTestProgram(test.code, "edit-demand.ts", "tsconfig.json")
			if err != nil {
				t.Fatal(err)
			}
			run := func(demand rule.EditDemand) rule.RuleDiagnostic {
				var diagnostics []rule.RuleDiagnostic
				linter.LintSingleFile(linter.LintSingleFileOptions{
					Program:     lintprogram.NewFromCompiler(program),
					File:        sourceFile.FileName(),
					HasTypeInfo: true,
					GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
						return []rule.ConfiguredRule{{
							Name:     NoUnnecessaryTemplateExpressionRule.Name,
							Severity: rule.SeverityError,
							Run: func(ctx rule.RuleContext) rule.RuleListeners {
								return NoUnnecessaryTemplateExpressionRule.Run(ctx, nil)
							},
						}}
					},
					Consumer: rule.DiagnosticConsumer{
						Demand: demand,
						Report: func(diagnostic rule.RuleDiagnostic) {
							diagnostics = append(diagnostics, diagnostic)
						},
					},
				})
				if len(diagnostics) != 1 {
					t.Fatalf("demand %d: got %d diagnostics, want 1", demand, len(diagnostics))
				}
				return diagnostics[0]
			}

			none := run(rule.EditDemandNone)
			autofix := run(rule.EditDemandAutofix)
			suggestion := run(rule.EditDemandSuggestion)
			all := run(rule.EditDemandAll)
			withoutEdits := func(diagnostic rule.RuleDiagnostic) rule.RuleDiagnostic {
				diagnostic.FixesPtr = nil
				diagnostic.Suggestions = nil
				return diagnostic
			}
			for demand, diagnostic := range map[rule.EditDemand]rule.RuleDiagnostic{
				rule.EditDemandNone:       none,
				rule.EditDemandAutofix:    autofix,
				rule.EditDemandSuggestion: suggestion,
			} {
				if !reflect.DeepEqual(withoutEdits(diagnostic), withoutEdits(all)) {
					t.Errorf("demand %d changed diagnostic identity", demand)
				}
			}
			if none.FixesPtr != nil || suggestion.FixesPtr != nil {
				t.Fatal("a non-autofix demand materialized fixes")
			}
			if autofix.FixesPtr == nil || !reflect.DeepEqual(autofix.FixesPtr, all.FixesPtr) {
				t.Fatal("autofix and all-edits demands did not produce the same fix")
			}
		})
	}
}
