package prefer_comparison_matcher_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	jestRule "github.com/web-infra-dev/rslint/internal/plugins/jest/rules/prefer_comparison_matcher"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	rstestRule "github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/prefer_comparison_matcher"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestPreferComparisonMatcherSourceOnlyReferences(t *testing.T) {
	for _, target := range []*rule.Rule{&jestRule.PreferComparisonMatcherRule, &rstestRule.PreferComparisonMatcherRule} {
		module := "@jest/globals"
		if strings.HasPrefix(target.Name, "rstest/") {
			module = "@rstest/core"
		}
		for _, test := range []struct {
			code  string
			count int
		}{
			{`import { expect as check } from '` + module + `'; check(a > b).toBe(true);`, 1},
			{`const { expect: check } = require('` + module + `'); check(a > b).toBe(true);`, 1},
			{`function f(expect: any) { expect(a > b).toBe(true); }`, 0},
			{`const expect = custom(); expect(a > b).toBe(true);`, 0},
			{`import { expect } from 'other'; expect(a > b).toBe(true);`, 0},
		} {
			root := fixtures.GetRootDir()
			fileName := tspath.ResolvePath(root.Dir, "references.ts")
			fs := utils.NewOverlayVFS(root.FS, map[string]string{fileName: test.code})
			program, err := lintprogram.NewFromRoots(lintprogram.RootOptions{
				RootFileNames: []string{fileName}, Host: utils.CreateCompilerHost(root.Dir, fs),
				CompilerOptions: &core.CompilerOptions{Module: core.ModuleKindESNext}, SingleThreaded: true,
			})
			if err != nil {
				t.Fatal(err)
			}
			count := 0
			linter.LintSingleFile(linter.LintSingleFileOptions{
				Program: program, File: fileName,
				GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
					return []rule.ConfiguredRule{{Name: target.Name, Severity: rule.SeverityError, Run: func(ctx rule.RuleContext) rule.RuleListeners {
						if ctx.TypeChecker != nil {
							t.Fatal("expected no type checker")
						}
						return target.Run(ctx, nil)
					}}}
				},
				Consumer: rule.DiagnosticConsumer{Report: func(rule.RuleDiagnostic) { count++ }},
			})
			if count != test.count {
				t.Errorf("%s: %s: got %d diagnostics, want %d", target.Name, test.code, count, test.count)
			}
		}
	}
}

func TestPreferComparisonMatcherEditDemand(t *testing.T) {
	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, source, err := helper.CreateTestProgram(`expect(2 > 1).toBe(false);`, "demand.ts", "tsconfig.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, target := range []*rule.Rule{&jestRule.PreferComparisonMatcherRule, &rstestRule.PreferComparisonMatcherRule} {
		t.Run(target.Name, func(t *testing.T) {
			var baseline rule.RuleDiagnostic
			for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
				var diagnostics []rule.RuleDiagnostic
				linter.LintSingleFile(linter.LintSingleFileOptions{
					Program: lintprogram.NewFromCompiler(program), File: source.FileName(),
					GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
						return []rule.ConfiguredRule{{Name: target.Name, Severity: rule.SeverityError, Run: func(ctx rule.RuleContext) rule.RuleListeners { return target.Run(ctx, nil) }}}
					},
					Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }},
				})
				if len(diagnostics) != 1 {
					t.Fatalf("demand %d: %d diagnostics", demand, len(diagnostics))
				}
				diagnostic := diagnostics[0]
				wantsFix := target == &jestRule.PreferComparisonMatcherRule && (demand == rule.EditDemandAutofix || demand == rule.EditDemandAll)
				wantsSuggestion := target == &rstestRule.PreferComparisonMatcherRule && (demand == rule.EditDemandSuggestion || demand == rule.EditDemandAll)
				if (diagnostic.FixesPtr != nil) != wantsFix || (diagnostic.Suggestions != nil) != wantsSuggestion {
					t.Fatalf("demand %d: unexpected edit materialization: %#v", demand, diagnostic)
				}
				if wantsSuggestion && len(*diagnostic.Suggestions) != 1 {
					t.Fatal("expected one suggestion")
				}
				diagnostic.FixesPtr, diagnostic.Suggestions = nil, nil
				if demand == rule.EditDemandNone {
					baseline = diagnostic
				} else if !reflect.DeepEqual(baseline, diagnostic) {
					t.Fatalf("demand %d changed diagnostic", demand)
				}
			}
		})
	}
}
