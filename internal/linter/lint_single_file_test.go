package linter

import (
	"errors"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func TestLintSingleFileWithoutRuleHandlerIsNoOp(t *testing.T) {
	LintSingleFile(LintSingleFileOptions{})
}

func TestLintSingleFileCallThrowsPrecedesRulesAndDoesNotLeak(t *testing.T) {
	raw, paths := createTestProgramWithFiles(t, map[string]string{
		"input.ts": "first(); second(); other();",
	})
	sourceProgram := lintprogram.NewFromCompiler(raw)
	for _, enabled := range []bool{true, false, true} {
		checked := false
		LintSingleFile(LintSingleFileOptions{
			Program: sourceProgram,
			File:    paths["input.ts"],
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				configured := []rule.ConfiguredRule{{
					Name: "consumer-before-providers",
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						checked = true
						if (ctx.CallThrows != nil) != enabled {
							t.Fatalf("predicate present = %v, enabled = %v", ctx.CallThrows != nil, enabled)
						}
						if enabled {
							for i, statement := range ctx.SourceFile.Statements.Nodes {
								call := statement.AsExpressionStatement().Expression
								if got := ctx.CallThrows(call); got != (i < 2) {
									t.Errorf("call %d throws = %v, want %v", i, got, i < 2)
								}
							}
						}
						return nil
					},
				}}
				if enabled {
					for _, name := range []string{"first", "second"} {
						configured = append(configured, rule.ConfiguredRule{
							Name: name,
							CallThrows: func(node *ast.Node) bool {
								return node.AsCallExpression().Expression.Text() == name
							},
							Run: func(rule.RuleContext) rule.RuleListeners { return nil },
						})
					}
				}
				return configured
			},
		})
		if !checked {
			t.Fatal("consumer did not run")
		}
	}
}

func TestLintSingleFileLeavesSyntaxGateToCaller(t *testing.T) {
	directory := t.TempDir()
	writeTestFiles(t, directory, map[string]string{"broken.ts": "const value = ;\n"})
	targetPath := norm(directory, "broken.ts")
	programs := wrapTestPrograms(gapProgram(t, directory, []string{targetPath}))
	var runs atomic.Int32
	LintSingleFile(LintSingleFileOptions{
		Program: programs[0],
		File:    targetPath,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
				Name: "caller-owned-syntax-gate",
				Run: func(rule.RuleContext) rule.RuleListeners {
					runs.Add(1)
					return nil
				},
			}}
		},
	})
	if runs.Load() != 1 {
		t.Fatalf("LintSingleFile rule runs = %d, want 1", runs.Load())
	}
}

func TestLintSingleFileRejectsFileOutsideProgram(t *testing.T) {
	raw, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
	})
	missing := paths["a.ts"] + ".missing"
	defer func() {
		recovered := recover()
		err, ok := recovered.(error)
		if !ok || !errors.Is(err, errTargetNotInProgram) || !strings.Contains(err.Error(), missing) {
			t.Fatalf("LintSingleFile panic = %v, want missing-target invariant error", recovered)
		}
	}()
	LintSingleFile(LintSingleFileOptions{
		Program: lintprogram.NewFromCompiler(raw),
		File:    missing,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			t.Fatal("missing single-file target resolved rules")
			return nil
		},
	})
}
