package linter

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func checkerProbeRule(checkerWasNil *bool) []ConfiguredRule {
	return []ConfiguredRule{{
		Name:     "checker-probe",
		Severity: rule.SeverityWarning,
		Run: func(ctx rule.RuleContext) rule.RuleListeners {
			*checkerWasNil = ctx.TypeChecker == nil
			return nil
		},
	}}
}

func runProgramCapabilityProbe(
	t *testing.T,
	sourceProgram *lintprogram.Program,
	rules RuleHandler,
) *LintResult {
	t.Helper()
	programs := []*lintprogram.Program{sourceProgram}
	targets := sourceProgram.RootFileNames()
	lintPlan := mustPrepareLintPlan(t, PrepareLintPlanOptions{
		Programs:         programs,
		TargetsByProgram: [][]string{targets},
		SingleThreaded:   true,
		GetRulesForFile:  rules,
	})
	result, err := RunLinter(RunLinterOptions{
		SingleThreaded: true,
		LintPlan:       lintPlan,
		Consumer:       rule.DiagnosticConsumer{Report: func(rule.RuleDiagnostic) {}},
	})
	if err != nil {
		t.Fatalf("RunLinter: %v", err)
	}
	return result
}

func TestCompilerCapableProgramProvidesTypeChecker(t *testing.T) {
	raw, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const x = 1;",
	})
	var checkerWasNil bool
	runProgramCapabilityProbe(t, lintprogram.NewFromCompiler(raw), func(*ast.SourceFile) []ConfiguredRule {
		return checkerProbeRule(&checkerWasNil)
	})
	if checkerWasNil {
		t.Fatal("compiler-capable Program did not provide a TypeChecker")
	}
}

func TestSourceOnlyProgramWithholdsTypeChecker(t *testing.T) {
	raw, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const x = 1;",
	})
	file := raw.GetSourceFile(paths["a.ts"])
	if file == nil {
		t.Fatal("fixture Program did not contain a.ts")
	}
	sourceOnly := mustSourceOnlyTestProgram(t, raw, []*ast.SourceFile{file})
	var checkerWasNil bool
	result := runProgramCapabilityProbe(t, sourceOnly, func(*ast.SourceFile) []ConfiguredRule {
		return checkerProbeRule(&checkerWasNil)
	})
	if !checkerWasNil || result.LintedFileCount != 1 {
		t.Fatalf("source-only result: checkerWasNil=%t files=%d", checkerWasNil, result.LintedFileCount)
	}
}

func TestSourceOnlyProgramFiltersTypeAwareRule(t *testing.T) {
	raw, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const x = 1;",
	})
	file := raw.GetSourceFile(paths["a.ts"])
	if file == nil {
		t.Fatal("fixture Program did not contain a.ts")
	}
	sourceOnly := mustSourceOnlyTestProgram(t, raw, []*ast.SourceFile{file})
	ruleRan := false
	result := runProgramCapabilityProbe(t, sourceOnly, func(*ast.SourceFile) []ConfiguredRule {
		return []ConfiguredRule{{
			Name:             "type-aware-probe",
			Severity:         rule.SeverityWarning,
			RequiresTypeInfo: true,
			Run: func(rule.RuleContext) rule.RuleListeners {
				ruleRan = true
				return nil
			},
		}}
	})
	if ruleRan {
		t.Fatal("type-aware rule ran for a source-only Program")
	}
	if _, retained := result.ExecutedRules["type-aware-probe"]; retained {
		t.Fatal("filtered type-aware rule appeared in ExecutedRules")
	}
}

func TestCompilerCapableProgramRunsTypeAwareRule(t *testing.T) {
	raw, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const x = 1;",
	})
	ruleRan := false
	result := runProgramCapabilityProbe(t, lintprogram.NewFromCompiler(raw), func(*ast.SourceFile) []ConfiguredRule {
		return []ConfiguredRule{{
			Name:             "type-aware-probe",
			Severity:         rule.SeverityWarning,
			RequiresTypeInfo: true,
			Run: func(ctx rule.RuleContext) rule.RuleListeners {
				if ctx.TypeChecker == nil {
					t.Fatal("type-aware rule received a nil TypeChecker")
				}
				ruleRan = true
				return nil
			},
		}}
	})
	if !ruleRan {
		t.Fatal("type-aware rule did not run for a compiler-capable Program")
	}
	if _, retained := result.ExecutedRules["type-aware-probe"]; !retained {
		t.Fatal("type-aware rule missing from ExecutedRules")
	}
}
