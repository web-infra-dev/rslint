// TestNoConditionalTestsResolvesInSourceOnlyProgram locks in that the rule
// keeps discriminating real Rstest registrations from same-named bindings when
// rslint builds a Program with no TypeChecker — the generation used for files
// no tsconfig project owns. The rule does not require type information, so it
// runs there, and its registration resolution has to survive on the file's
// binder scope graph alone.
package no_conditional_tests_test

import (
	"sort"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/no_conditional_tests"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestNoConditionalTestsResolvesInSourceOnlyProgram(t *testing.T) {
	if no_conditional_tests.NoConditionalTestsRule.RequiresTypeInfo {
		t.Fatal("rstest/no-conditional-tests must not require type information")
	}

	code := `
import { test } from 'vitest';
import { test as rstestTest, describe } from '@rstest/core';
import * as rstest from '@rstest/core';

const localTest = (name: string, fn: () => void) => {};

if (flag) { test('vitest lookalike', () => {}); }
if (flag) { localTest('local lookalike', () => {}); }
if (flag) { rstestTest('renamed import', () => {}); }
if (flag) { rstest.test('namespace import', () => {}); }
if (flag) { describe('plain import', () => {}); }
`
	// The identifier each diagnostic must anchor to, in source order. The two
	// lookalikes are absent: neither is an Rstest registration, and reporting
	// one would be the false positive this rule's resolution exists to avoid.
	// A namespace registration anchors to its receiver, the same node a
	// checker-backed run reports.
	want := []string{"rstestTest", "rstest", "describe"}

	root := fixtures.GetRootDir()
	fileName := tspath.ResolvePath(root.Dir, "no-conditional-tests-source-only.ts")
	fs := utils.NewOverlayVFS(root.FS, map[string]string{fileName: code})
	host := utils.CreateCompilerHost(root.Dir, fs)
	sourceProgram, err := lintprogram.NewFromRoots(lintprogram.RootOptions{
		RootFileNames:   []string{fileName},
		Host:            host,
		CompilerOptions: &core.CompilerOptions{Module: core.ModuleKindESNext},
		SingleThreaded:  true,
	})
	if err != nil {
		t.Fatalf("NewFromRoots: %v", err)
	}
	if sourceProgram.CanProvideTypeChecker(sourceProgram.SourceFiles()[0]) {
		t.Fatal("expected a source-only Program with no TypeChecker")
	}

	lintPlan, err := linter.PrepareLintPlan(linter.PrepareLintPlanOptions{
		Programs:         []*lintprogram.Program{sourceProgram},
		TargetsByProgram: [][]string{{fileName}},
		SingleThreaded:   true,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
				Name:     no_conditional_tests.NoConditionalTestsRule.Name,
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return no_conditional_tests.NoConditionalTestsRule.Run(ctx, nil)
				},
			}}
		},
	})
	if err != nil {
		t.Fatalf("PrepareLintPlan: %v", err)
	}

	type reported struct {
		pos  int
		text string
	}
	var got []reported
	if _, err := linter.RunLinter(linter.RunLinterOptions{
		SingleThreaded: true,
		LintPlan:       lintPlan,
		Consumer: rule.DiagnosticConsumer{
			Report: func(diagnostic rule.RuleDiagnostic) {
				got = append(got, reported{
					pos:  diagnostic.Range.Pos(),
					text: code[diagnostic.Range.Pos():diagnostic.Range.End()],
				})
			},
		},
	}); err != nil {
		t.Fatalf("RunLinter: %v", err)
	}
	sort.Slice(got, func(left, right int) bool { return got[left].pos < got[right].pos })

	if len(got) != len(want) {
		t.Fatalf("reported %d registrations, want %d: %v", len(got), len(want), got)
	}
	for index, expected := range want {
		if got[index].text != expected {
			t.Errorf("diagnostic %d anchored to %q, want %q", index, got[index].text, expected)
		}
	}
}
