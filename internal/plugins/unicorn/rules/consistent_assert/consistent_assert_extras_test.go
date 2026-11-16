package consistent_assert_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/consistent_assert"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestConsistentAssertExtras(t *testing.T) {
	aliasCode := "import {strict as verify} from 'node:assert';\nverify(value);"
	aliasOutput := "import {strict as verify} from 'node:assert';\nverify.ok(value);"

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&consistent_assert.ConsistentAssertRule,
		[]rule_tester.ValidTestCase{
			valid("import assert from 'node:test';\nassert(value);"),
			valid("const assert = value => value;\nassert(value);"),
			valid("import assert from 'node:assert';\nassert.call(null, value);"),
			valid("import assert from 'node:assert';\nconsume(assert);"),
			validTS("import assert from 'node:assert';\n(assert as unknown as ((value: unknown) => void))(value);"),
			validTS("import {type strict as assert} from 'assert';\nassert(value);"),
			valid("import {strict as assert} from 'node:assert/strict';\nassert(value);"),
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:            aliasCode,
				FileName:        "file.js",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Output:          []string{aliasOutput},
				Errors:          []rule_tester.InvalidTestCaseError{expectedError("verify", 2, 1)},
			},
			invalid(
				"import assert from 'node:assert';\n((assert))(value);",
				"import assert from 'node:assert';\n((assert.ok))(value);",
				"assert",
				2,
				3,
			),
		},
	)
}

// TestConsistentAssertEditDemand checks that the diagnostic is identical under
// every edit demand and that the `.ok` autofix appears only when autofixes are
// requested. The rule offers no suggestions.
func TestConsistentAssertEditDemand(t *testing.T) {
	code := "import assert from 'node:assert';\nassert(value);"
	output := "import assert from 'node:assert';\nassert.ok(value);"

	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, source, err := helper.CreateTestProgram(code, "edit-demand.ts", "tsconfig.json")
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
					Name: consistent_assert.ConsistentAssertRule.Name, Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return consistent_assert.ConsistentAssertRule.Run(ctx, nil)
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
		if diagnostic.Suggestions != nil && len(*diagnostic.Suggestions) != 0 {
			t.Errorf("demand %d produced unexpected suggestions", demand)
		}
		if demand != rule.EditDemandAutofix && demand != rule.EditDemandAll {
			if diagnostic.FixesPtr != nil && len(*diagnostic.FixesPtr) != 0 {
				t.Errorf("demand %d produced an unexpected autofix", demand)
			}
			continue
		}
		if diagnostic.FixesPtr == nil || len(*diagnostic.FixesPtr) != 1 || !reflect.DeepEqual(diagnostic.FixesPtr, all.FixesPtr) {
			t.Fatalf("demand %d: inconsistent autofixes", demand)
		}
		fixed, _, ok := linter.ApplyRuleFixes(code, []rule.RuleDiagnostic{diagnostic})
		if !ok || fixed != output {
			t.Errorf("demand %d: unexpected autofix %q", demand, fixed)
		}
	}
}
