package prefer_type_error_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_type_error"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferTypeErrorExtras(t *testing.T) {
	typeErrorUnavailable := invalid("if (typeof value === 'string') { throw new Error(); }")
	typeErrorUnavailable.Globals = map[string]any{"TypeError": "off"}
	typeErrorUnavailable.Output = []string{}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_type_error.PreferTypeErrorRule,
		[]rule_tester.ValidTestCase{
			valid("function f(Error) { if (typeof value === 'string') { throw new Error(); } }"),
			valid("function f(isNaN) { if (isNaN(value)) { throw new Error(); } }"),
			valid("function f(isFinite) { if (isFinite(value)) { throw new Error(); } }"),
			{
				Code:     "if (typeof value === 'string') { throw new Error(); }",
				FileName: "case.js",
				Globals:  map[string]any{"Error": "off"},
			},
			{
				Code:     "if (isNaN(value)) { throw new Error(); }",
				FileName: "case.js",
				Globals:  map[string]any{"isNaN": "off"},
			},
			valid("if (typeof value === 'string') throw new Error();"),
			{
				Code:            "if (typeof value === 'string') { throw new (Error as any)(); }",
				FileName:        "case.ts",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			},
			valid("if (value instanceof ns.Error) { throw new Error(); }"),
			valid("if (SomeThing['isArray'](value)) { throw new Error(); }"),
			// ESTree exposes optional chains as ChainExpression, which upstream
			// does not classify as a type-checking expression.
			valid("if (utils?.isArray(value)) { throw new Error(); }"),
			valid("if (utils.isArray?.(value)) { throw new Error(); }"),
			valid("if (wrapper?.utils.isArray(value)) { throw new Error(); }"),
			// tsgo uses BinaryExpression for ESTree SequenceExpression and
			// AssignmentExpression shapes. Those are not upstream binary checks.
			valid("if (Array.isArray(value), flag) { throw new Error(); }"),
			valid("if (flag = Array.isArray(value)) { throw new Error(); }"),
			valid("if (flag += Number.isNaN(value)) { throw new Error(); }"),
			valid("if (flag &&= Array.isArray(value)) { throw new Error(); }"),
			valid("if (flag ||= Array.isArray(value)) { throw new Error(); }"),
			valid("if (flag ??= Array.isArray(value)) { throw new Error(); }"),
			valid("{ throw new Error(); }"),
		},
		[]rule_tester.InvalidTestCase{
			invalid("if (typeof value === 'string') {} else { throw new Error(); }"),
			invalid("if (SomeThing[isArray](value)) { throw new Error(); }"),
			invalid("if (typeof value > 0) { throw new Error(); }"),
			invalid("if (value instanceof ns['Error']) { throw new Error(); }"),
			invalid("if (!!Array.isArray(value)) { throw new Error(); }"),
			invalid("if (Array.isArray(value) ?? Number.isFinite(value)) { throw new Error(); }"),
			typeErrorUnavailable,
			{
				Code:            "function f(TypeError) { if (typeof value === 'string') { throw new Error(); } }",
				FileName:        "case.js",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Output:          []string{},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId:   "prefer-type-error",
					Message:     preferTypeErrorMessage,
					Line:        1,
					Column:      68,
					EndLine:     1,
					EndColumn:   73,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				}},
			},
		},
	)
}

func TestPreferTypeErrorArtifactsFollowDemand(t *testing.T) {
	for _, testCase := range []struct {
		source string
		output string
	}{
		{
			source: "if (typeof value === 'string') { throw new Error(); }",
			output: "if (typeof value === 'string') { throw new TypeError(); }",
		},
		{
			source: "function f(TypeError) { if (typeof value === 'string') { throw new Error(); } }",
			output: "function f(TypeError) { if (typeof value === 'string') { throw new Error(); } }",
		},
	} {
		t.Run(testCase.source, func(t *testing.T) {
			program, sourceFile, err := rule_tester.NewProgramHelper(fixtures.GetRootDir()).CreateTestProgram(testCase.source, "edit-demand.js", "tsconfig.json")
			if err != nil {
				t.Fatal(err)
			}

			demands := []rule.EditDemand{
				rule.EditDemandNone,
				rule.EditDemandAutofix,
				rule.EditDemandSuggestion,
				rule.EditDemandAll,
			}
			diagnostics := make([]rule.RuleDiagnostic, len(demands))
			for index, demand := range demands {
				var found []rule.RuleDiagnostic
				linter.LintSingleFile(linter.LintSingleFileOptions{
					Program: lintprogram.NewFromCompiler(program),
					File:    sourceFile.FileName(),
					GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
						return []rule.ConfiguredRule{{
							Name:     prefer_type_error.PreferTypeErrorRule.Name,
							Severity: rule.SeverityError,
							Run: func(ctx rule.RuleContext) rule.RuleListeners {
								return prefer_type_error.PreferTypeErrorRule.Run(ctx, nil)
							},
						}}
					},
					Consumer: rule.DiagnosticConsumer{
						Demand: demand,
						Report: func(diagnostic rule.RuleDiagnostic) {
							found = append(found, diagnostic)
						},
					},
				})
				if len(found) != 1 {
					t.Fatalf("demand %d: expected one diagnostic, got %d", demand, len(found))
				}
				diagnostics[index] = found[0]
			}

			all := diagnostics[len(diagnostics)-1]
			for index, demand := range demands {
				diagnostic := diagnostics[index]
				if diagnostic.Range != all.Range ||
					!reflect.DeepEqual(diagnostic.Message, all.Message) ||
					diagnostic.Severity != all.Severity {
					t.Errorf("demand %d changed diagnostic identity", demand)
				}
				wantFix := testCase.source != testCase.output &&
					(demand == rule.EditDemandAutofix || demand == rule.EditDemandAll)
				if (diagnostic.FixesPtr != nil) != wantFix {
					t.Errorf("demand %d: unexpected fix artifacts", demand)
				}
				if wantFix && !reflect.DeepEqual(diagnostic.FixesPtr, all.FixesPtr) {
					t.Errorf("demand %d changed fix artifacts", demand)
				}
				if diagnostic.Suggestions != nil {
					t.Errorf("demand %d: autofix-only rule produced suggestions", demand)
				}
			}

			for index, demand := range demands {
				wantFix := testCase.source != testCase.output &&
					(demand == rule.EditDemandAutofix || demand == rule.EditDemandAll)
				expected := testCase.source
				if wantFix {
					expected = testCase.output
				}
				output, _, fixed := linter.ApplyRuleFixes(testCase.source, diagnostics[index:index+1])
				if output != expected || fixed != wantFix {
					t.Errorf("demand %d: got %q, want %q", demand, output, expected)
				}
			}
		})
	}
}
