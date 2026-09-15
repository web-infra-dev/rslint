package linter

import (
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// These tests exercise the "type-check only" code path: LintPlan=nil combined
// with TypeCheck=true. The contract is:
//
//   1. Phase 1 (lint rules) is skipped entirely — no rule diagnostics, no
//      LintedFileCount, no ExecutedRules.
//   2. Phase 2 (type-check) still produces tsc-aligned diagnostics.
//   3. A lint plan does not constrain Phase 2 diagnostics. This locks the
//      contract documented in website/docs/en/guide/type-checking.md
//      (type-check mirrors `tsgo --noEmit`, ignoring lint targets).
//   4. Source-only Programs expose no Phase 2 capability.

// triggerOnIdentifierRule reports a warning on every identifier — used to
// confirm rules really would have fired if Phase 1 had run.
func triggerOnIdentifierRule() []rule.ConfiguredRule {
	return []rule.ConfiguredRule{
		{
			Name:     "would-have-fired",
			Severity: rule.SeverityError,
			Run: func(ctx rule.RuleContext) rule.RuleListeners {
				return rule.RuleListeners{
					ast.KindIdentifier: func(node *ast.Node) {
						ctx.ReportNode(node, rule.RuleMessage{Id: "x", Description: "would have fired"})
					},
				}
			},
		},
	}
}

// classifyDiagnostics splits a flat slice into (TypeScript(TS…) entries, lint entries).
func classifyDiagnostics(diags []rule.RuleDiagnostic) (tsDiags, lintDiags []rule.RuleDiagnostic) {
	for _, d := range diags {
		if strings.HasPrefix(d.RuleName, "TypeScript(") {
			tsDiags = append(tsDiags, d)
		} else {
			lintDiags = append(lintDiags, d)
		}
	}
	return
}

func TestTypeCheckOnly_NoLintDiagnostics(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		// triggerOnIdentifierRule would fire many times on this file.
		"a.ts": "const a = 1; const b = 2;",
	})

	var diags []rule.RuleDiagnostic
	_, err := RunLinter(RunLinterOptions{
		TypeCheckOnlyPrograms: wrapTestPrograms(program),
		SingleThreaded:        true,
		TypeCheck:             true,
		Consumer: rule.DiagnosticConsumer{
			Report: func(d rule.RuleDiagnostic) { diags = append(diags, d) },
		},
	})
	if err != nil {
		t.Fatalf("RunLinter returned error: %v", err)
	}

	_, lintDiags := classifyDiagnostics(diags)
	if len(lintDiags) != 0 {
		t.Fatalf("expected no lint diagnostics with LintPlan=nil, got %d: %+v", len(lintDiags), lintDiags)
	}
}

func TestTypeCheckOnly_StillReportsTSErrors(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const x: number = 'hello';", // TS2322 type mismatch
	})

	var diags []rule.RuleDiagnostic
	_, err := RunLinter(RunLinterOptions{
		TypeCheckOnlyPrograms: wrapTestPrograms(program),
		SingleThreaded:        true,
		TypeCheck:             true,
		Consumer: rule.DiagnosticConsumer{
			Report: func(d rule.RuleDiagnostic) { diags = append(diags, d) },
		},
	})
	if err != nil {
		t.Fatalf("RunLinter returned error: %v", err)
	}

	tsDiags, _ := classifyDiagnostics(diags)
	if len(tsDiags) == 0 {
		t.Fatal("expected TS diagnostics under type-check-only mode, got none")
	}
	foundTS2322 := false
	for _, d := range tsDiags {
		if strings.Contains(d.RuleName, "TS2322") {
			foundTS2322 = true
			break
		}
	}
	if !foundTS2322 {
		t.Errorf("expected TS2322 (type mismatch), got: %+v", tsDiags)
	}
}

func TestTypeCheckOnly_LintedFileCountIsZero(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
		"b.ts": "const b = 2;",
	})

	result, err := RunLinter(RunLinterOptions{
		TypeCheckOnlyPrograms: wrapTestPrograms(program),
		SingleThreaded:        true,
		TypeCheck:             true,
		Consumer: rule.DiagnosticConsumer{
			Report: func(rule.RuleDiagnostic) {},
		},
	})
	if err != nil {
		t.Fatalf("RunLinter returned error: %v", err)
	}
	if result.LintedFileCount != 0 {
		t.Errorf("expected LintedFileCount=0 when Phase 1 is skipped, got %d", result.LintedFileCount)
	}
}

func TestTypeCheckOnly_ExecutedRulesIsEmpty(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
	})

	result, err := RunLinter(RunLinterOptions{
		TypeCheckOnlyPrograms: wrapTestPrograms(program),
		SingleThreaded:        true,
		TypeCheck:             true,
		Consumer: rule.DiagnosticConsumer{
			Report: func(rule.RuleDiagnostic) {},
		},
	})
	if err != nil {
		t.Fatalf("RunLinter returned error: %v", err)
	}
	if len(result.ExecutedRules) != 0 {
		t.Errorf("expected ExecutedRules to be empty when Phase 1 is skipped, got %v", result.ExecutedRules)
	}
	if result.ExecutedRules == nil {
		t.Error("expected ExecutedRules to be a writable, non-nil empty map")
	}
}

// TestTypeCheckOnly_BaselineLintWouldFire is a sanity check: with the same
// Program but a real lint plan, lint diagnostics WOULD fire — proving the
// absence above is due to the missing plan, not an unlintable file.
func TestTypeCheckOnly_BaselineLintWouldFire(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1; const b = 2;",
	})

	var diags []rule.RuleDiagnostic
	programs := wrapTestPrograms(program)
	lintPlan := mustPrepareLintPlan(t, PrepareLintPlanOptions{
		Programs:         programs,
		TargetsByProgram: [][]string{{paths["a.ts"]}},
		SingleThreaded:   true,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return triggerOnIdentifierRule()
		},
	})
	_, err := RunLinter(RunLinterOptions{
		SingleThreaded: true,
		LintPlan:       lintPlan,
		Consumer: rule.DiagnosticConsumer{
			Report: func(d rule.RuleDiagnostic) { diags = append(diags, d) },
		},
	})
	if err != nil {
		t.Fatalf("RunLinter returned error: %v", err)
	}
	_, lintDiags := classifyDiagnostics(diags)
	if len(lintDiags) == 0 {
		t.Fatal("baseline expectation broken: rule was supposed to fire with a lint plan")
	}
}

// TestTypeCheckOnly_EmptyLintPlanDoesNotConstrainTypeCheck verifies that an
// empty Phase 1 projection does not suppress Program-wide diagnostics.
func TestTypeCheckOnly_EmptyLintPlanDoesNotConstrainTypeCheck(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const x: number = 'hello';",
	})

	var diags []rule.RuleDiagnostic
	programs := wrapTestPrograms(program)
	lintPlan := mustPrepareLintPlan(t, PrepareLintPlanOptions{
		Programs:         programs,
		TargetsByProgram: [][]string{nil},
		SingleThreaded:   true,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			t.Fatal("empty target projection resolved rules")
			return nil
		},
	})
	_, err := RunLinter(RunLinterOptions{
		SingleThreaded: true,
		LintPlan:       lintPlan,
		TypeCheck:      true,
		Consumer: rule.DiagnosticConsumer{
			Report: func(d rule.RuleDiagnostic) { diags = append(diags, d) },
		},
	})
	if err != nil {
		t.Fatalf("RunLinter returned error: %v", err)
	}
	tsDiags, _ := classifyDiagnostics(diags)
	if len(tsDiags) == 0 {
		t.Fatal("empty lint plan suppressed type-check diagnostics")
	}
}

func TestTypeCheck_IncludesZeroTargetProgramFromLintPlan(t *testing.T) {
	lintProgram, lintPaths := createTestProgramWithFiles(t, map[string]string{
		"lint.ts": "const lintTarget = 1;",
	})
	typeCheckProgram, typeCheckPaths := createTestProgramWithFiles(t, map[string]string{
		"type-error.ts": "const value: number = 'wrong';",
	})
	programs := wrapTestPrograms(lintProgram, typeCheckProgram)
	lintPlan := mustPrepareLintPlan(t, PrepareLintPlanOptions{
		Programs:         programs,
		TargetsByProgram: [][]string{{lintPaths["lint.ts"]}, nil},
		SingleThreaded:   true,
		GetRulesForFile:  func(*ast.SourceFile) []rule.ConfiguredRule { return noopRule() },
	})

	var diagnostics []rule.RuleDiagnostic
	result, err := RunLinter(RunLinterOptions{
		SingleThreaded: true,
		LintPlan:       lintPlan,
		TypeCheck:      true,
		Consumer: rule.DiagnosticConsumer{
			Report: func(diagnostic rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, diagnostic)
			},
		},
	})
	if err != nil {
		t.Fatalf("RunLinter returned error: %v", err)
	}
	if result.LintedFileCount != 1 {
		t.Fatalf("LintedFileCount = %d, want 1", result.LintedFileCount)
	}
	typeDiagnostics, _ := classifyDiagnostics(diagnostics)
	if len(typeDiagnostics) == 0 || typeDiagnostics[0].FilePath != typeCheckPaths["type-error.ts"] {
		t.Fatalf("zero-target Program type diagnostics = %+v", typeDiagnostics)
	}
}

// TestTypeCheckOnly_SkipsSourceOnlyProgram verifies that source-only Programs
// remain excluded from type-check when Phase 1 is skipped.
func TestTypeCheckOnly_SkipsSourceOnlyProgram(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const x: number = 'hello';",
	})
	file := program.GetSourceFile(paths["a.ts"])
	if file == nil {
		t.Fatal("fixture Program did not contain a.ts")
	}
	sourceOnly := mustSourceOnlyTestProgram(t, program, []*ast.SourceFile{file})

	var diags []rule.RuleDiagnostic
	_, err := RunLinter(RunLinterOptions{
		TypeCheckOnlyPrograms: testPrograms(sourceOnly),
		SingleThreaded:        true,
		TypeCheck:             true,
		Consumer: rule.DiagnosticConsumer{
			Report: func(d rule.RuleDiagnostic) { diags = append(diags, d) },
		},
	})
	if err != nil {
		t.Fatalf("RunLinter returned error: %v", err)
	}
	tsDiags, _ := classifyDiagnostics(diags)
	if len(tsDiags) != 0 {
		t.Errorf("expected 0 TS diagnostics for a source-only Program, got %d: %+v", len(tsDiags), tsDiags)
	}
}

// TestTypeCheckOnly_TypeCheckFalseProducesNothing closes the matrix: even
// when LintPlan is nil, if TypeCheck is also false there should be
// no diagnostics at all (no work done in either phase).
func TestTypeCheckOnly_TypeCheckFalseProducesNothing(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const x: number = 'hello';",
	})

	var diags []rule.RuleDiagnostic
	_, err := RunLinter(RunLinterOptions{
		TypeCheckOnlyPrograms: wrapTestPrograms(program),
		SingleThreaded:        true,
		TypeCheck:             false,
		Consumer: rule.DiagnosticConsumer{
			Report: func(d rule.RuleDiagnostic) { diags = append(diags, d) },
		},
	})
	if err != nil {
		t.Fatalf("RunLinter returned error: %v", err)
	}
	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics when both phases are off, got %d: %+v", len(diags), diags)
	}
}
