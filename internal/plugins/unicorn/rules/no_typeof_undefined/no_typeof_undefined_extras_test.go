package no_typeof_undefined_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_typeof_undefined"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoTypeofUndefinedMultilineBoundaries(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_typeof_undefined.NoTypeofUndefinedRule,
		[]rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			invalidFixed(
				"function f(value) { throw typeof // comment\nvalue === \"undefined\"; }",
				"function f(value) { throw ( // comment\nvalue === undefined); }",
			),
			invalidFixed(
				"function f(value) { const check = typeof // comment\nvalue === \"undefined\"; return check; }",
				"function f(value) { const check = // comment\nvalue === undefined; return check; }",
			),
		},
	)
}

func TestNoTypeofUndefinedReviewBoundaries(t *testing.T) {
	parenthesizedGlobal := invalidGlobal("typeof (missing) === \"undefined\"", "(missing) === undefined")
	asGlobal := invalidGlobal("typeof (missing as unknown) === \"undefined\"", "(missing as unknown) === undefined")
	asGlobal.FileName = "case.ts"
	nonNullGlobal := invalidGlobal("typeof missing! === \"undefined\"", "missing! === undefined")
	nonNullGlobal.FileName = "case.ts"
	angleGlobal := invalidGlobal("typeof (<unknown>missing) === \"undefined\"", "(<unknown>missing) === undefined")
	angleGlobal.FileName = "case.ts"

	shadowedUndefined := invalidFixed(
		"function f(value, undefined) { return typeof value === \"undefined\"; }",
		"function f(value, undefined) { return value === undefined; }",
	)
	shadowedUndefined.Output = []string{}

	shadowedUndefinedGlobal := invalidGlobal(
		"function f(undefined) { return typeof missing === \"undefined\"; }",
		"function f(undefined) { return missing === undefined; }",
	)
	shadowedUndefinedGlobal.Output = []string{}
	shadowedUndefinedGlobal.Errors[0].Suggestions = []rule_tester.InvalidTestCaseSuggestion{}

	documentAll := invalidFixed("typeof document.all === \"undefined\"", "document.all === undefined")
	documentAll.Output = []string{}
	documentAll.Globals = map[string]any{"document": "readonly"}

	documentAllComputed := invalidFixed("typeof document[\"all\"] === \"undefined\"", "document[\"all\"] === undefined")
	documentAllComputed.Output = []string{}
	documentAllComputed.Globals = map[string]any{"document": "readonly"}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_typeof_undefined.NoTypeofUndefinedRule,
		[]rule_tester.ValidTestCase{
			valid("typeof (missing) === \"undefined\""),
			{Code: "typeof (missing as unknown) === \"undefined\"", FileName: "case.ts"},
			{Code: "typeof missing! === \"undefined\"", FileName: "case.ts"},
			{Code: "typeof (<unknown>missing) === \"undefined\"", FileName: "case.ts"},
		},
		[]rule_tester.InvalidTestCase{
			parenthesizedGlobal,
			asGlobal,
			nonNullGlobal,
			angleGlobal,
			invalidFixed("typeof ({}) === \"undefined\"", "({}) === undefined"),
			invalidFixed("typeof {} === \"undefined\"", "({}) === undefined"),
			invalidFixed("typeof function() {} === \"undefined\"", "(function() {}) === undefined"),
			invalidFixed("typeof class {} === \"undefined\"", "(class {}) === undefined"),
			shadowedUndefined,
			shadowedUndefinedGlobal,
			documentAll,
			documentAllComputed,
		},
	)
}

func TestNoTypeofUndefinedSuggestionData(t *testing.T) {
	for _, test := range []struct {
		code     string
		operator string
	}{
		{"typeof missing === \"undefined\"", "==="},
		{"typeof missing != \"undefined\"", "!=="},
	} {
		t.Run(test.operator, func(t *testing.T) {
			program, sourceFile, err := rule_tester.NewProgramHelper(fixtures.GetRootDir()).CreateTestProgram(test.code, "case.js", "tsconfig.json")
			if err != nil {
				t.Fatal(err)
			}
			var diagnostics []rule.RuleDiagnostic
			linter.LintSingleFile(linter.LintSingleFileOptions{
				Program: lintprogram.NewFromCompiler(program),
				File:    sourceFile.FileName(),
				GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
					return []rule.ConfiguredRule{{
						Name:     no_typeof_undefined.NoTypeofUndefinedRule.Name,
						Severity: rule.SeverityError,
						Run: func(ctx rule.RuleContext) rule.RuleListeners {
							return no_typeof_undefined.NoTypeofUndefinedRule.Run(ctx, []any{map[string]any{"checkGlobalVariables": true}})
						},
					}}
				},
				Consumer: rule.DiagnosticConsumer{
					Demand: rule.EditDemandSuggestion,
					Report: func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) },
				},
			})
			if len(diagnostics) != 1 || diagnostics[0].Suggestions == nil || len(*diagnostics[0].Suggestions) != 1 {
				t.Fatalf("unexpected diagnostics: %+v", diagnostics)
			}
			got := (*diagnostics[0].Suggestions)[0].Message.Data
			want := map[string]string{"operator": test.operator}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("suggestion data = %#v, want %#v", got, want)
			}
		})
	}
}
