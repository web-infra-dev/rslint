package linter

import (
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// These tests exercise the "type-check only" code path: GetRulesForFile=nil
// combined with TypeCheck=true. The contract is:
//
//   1. Phase 1 (lint rules) is skipped entirely — no rule diagnostics, no
//      LintedFileCount, no ExecutedRules.
//   2. Phase 2 (type-check) still produces tsc-aligned diagnostics.
//   3. PerProgramFilter is a Phase-1-only concern; it does NOT suppress
//      Phase 2 diagnostics. This locks the contract documented in
//      website/docs/en/guide/type-checking.md (type-check mirrors
//      `tsgo --noEmit`, ignoring rslint-side filters).
//   4. SkipTypeCheckPrograms continues to gate Phase 2 per-program.

// triggerOnIdentifierRule reports a warning on every identifier — used to
// confirm rules really would have fired if Phase 1 had run.
func triggerOnIdentifierRule() []ConfiguredRule {
	return []ConfiguredRule{
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
		Programs:        []*compiler.Program{program},
		SingleThreaded:  true,
		ExcludePaths:    utils.ExcludePaths,
		GetRulesForFile: nil, // <-- type-check-only path
		TypeCheck:       true,
		Consumer: rule.DiagnosticConsumer{
			Report: func(d rule.RuleDiagnostic) { diags = append(diags, d) },
		},
	})
	if err != nil {
		t.Fatalf("RunLinter returned error: %v", err)
	}

	_, lintDiags := classifyDiagnostics(diags)
	if len(lintDiags) != 0 {
		t.Fatalf("expected no lint diagnostics with GetRulesForFile=nil, got %d: %+v", len(lintDiags), lintDiags)
	}
}

func TestTypeCheckOnly_StillReportsTSErrors(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const x: number = 'hello';", // TS2322 type mismatch
	})

	var diags []rule.RuleDiagnostic
	_, err := RunLinter(RunLinterOptions{
		Programs:        []*compiler.Program{program},
		SingleThreaded:  true,
		ExcludePaths:    utils.ExcludePaths,
		GetRulesForFile: nil,
		TypeCheck:       true,
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
		Programs:        []*compiler.Program{program},
		SingleThreaded:  true,
		ExcludePaths:    utils.ExcludePaths,
		GetRulesForFile: nil,
		TypeCheck:       true,
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
		Programs:        []*compiler.Program{program},
		SingleThreaded:  true,
		ExcludePaths:    utils.ExcludePaths,
		GetRulesForFile: nil,
		TypeCheck:       true,
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
// program but a real GetRulesForFile, lint diagnostics WOULD fire — proving
// the absence above is due to nil, not due to the file being un-lintable.
func TestTypeCheckOnly_BaselineLintWouldFire(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1; const b = 2;",
	})

	var diags []rule.RuleDiagnostic
	_, err := RunLinter(RunLinterOptions{
		Programs:        []*compiler.Program{program},
		SingleThreaded:  true,
		ExcludePaths:    utils.ExcludePaths,
		GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule { return triggerOnIdentifierRule() },
		TypeCheck:       false,
		Consumer: rule.DiagnosticConsumer{
			Report: func(d rule.RuleDiagnostic) { diags = append(diags, d) },
		},
	})
	if err != nil {
		t.Fatalf("RunLinter returned error: %v", err)
	}
	_, lintDiags := classifyDiagnostics(diags)
	if len(lintDiags) == 0 {
		t.Fatal("baseline expectation broken: rule was supposed to fire when GetRulesForFile is non-nil")
	}
}

// TestTypeCheckOnly_PerProgramFilterIgnoredByTypeCheck locks Claim B from
// the design discussion: a PerProgramFilter that rejects ALL files still
// does not suppress type-check diagnostics. This mirrors `tsgo --noEmit`
// semantics — type-check scope is the tsconfig-determined program, not the
// rslint-side ignore set.
func TestTypeCheckOnly_PerProgramFilterIgnoredByTypeCheck(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const x: number = 'hello';",
	})

	rejectAll := func(string) bool { return false }

	var diags []rule.RuleDiagnostic
	_, err := RunLinter(RunLinterOptions{
		Programs:         []*compiler.Program{program},
		SingleThreaded:   true,
		ExcludePaths:     utils.ExcludePaths,
		PerProgramFilter: []FileFilter{rejectAll},
		GetRulesForFile:  nil,
		TypeCheck:        true,
		Consumer: rule.DiagnosticConsumer{
			Report: func(d rule.RuleDiagnostic) { diags = append(diags, d) },
		},
	})
	if err != nil {
		t.Fatalf("RunLinter returned error: %v", err)
	}
	tsDiags, _ := classifyDiagnostics(diags)
	if len(tsDiags) == 0 {
		t.Fatal("PerProgramFilter rejecting all files should NOT suppress type-check diagnostics, but got none")
	}
}

// TestTypeCheckOnly_RespectsSkipMask verifies SkipTypeCheckPrograms still
// gates Phase 2 even when Phase 1 is skipped (i.e. gap fallback programs
// remain excluded from type-check, as committed in type-checking.md).
func TestTypeCheckOnly_RespectsSkipMask(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const x: number = 'hello';",
	})

	var diags []rule.RuleDiagnostic
	_, err := RunLinter(RunLinterOptions{
		Programs:              []*compiler.Program{program},
		SingleThreaded:        true,
		ExcludePaths:          utils.ExcludePaths,
		GetRulesForFile:       nil,
		TypeCheck:             true,
		SkipTypeCheckPrograms: []bool{true}, // skip the only program
		Consumer: rule.DiagnosticConsumer{
			Report: func(d rule.RuleDiagnostic) { diags = append(diags, d) },
		},
	})
	if err != nil {
		t.Fatalf("RunLinter returned error: %v", err)
	}
	tsDiags, _ := classifyDiagnostics(diags)
	if len(tsDiags) != 0 {
		t.Errorf("expected 0 TS diagnostics with SkipTypeCheckPrograms[0]=true, got %d: %+v", len(tsDiags), tsDiags)
	}
}

// TestTypeCheckOnly_TypeCheckFalseProducesNothing closes the matrix: even
// when GetRulesForFile is nil, if TypeCheck is also false there should be
// no diagnostics at all (no work done in either phase).
func TestTypeCheckOnly_TypeCheckFalseProducesNothing(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const x: number = 'hello';",
	})

	var diags []rule.RuleDiagnostic
	_, err := RunLinter(RunLinterOptions{
		Programs:        []*compiler.Program{program},
		SingleThreaded:  true,
		ExcludePaths:    utils.ExcludePaths,
		GetRulesForFile: nil,
		TypeCheck:       false,
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
