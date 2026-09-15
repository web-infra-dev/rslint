package prefer_to_contain

import (
	"sort"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestPreferToContainResolvesInSourceOnlyProgram(t *testing.T) {
	if PreferToContainRule.RequiresTypeInfo {
		t.Fatal("rstest/prefer-to-contain must not require type information")
	}
	code := `
expect(globalList.includes(item)).toBe(true);
import { expect as check, test as rstestTest } from '@rstest/core';
import * as rstest from '@rstest/playwright';
const { expect: requiredExpect } = require('rstack/test');
check(importedList.includes(item)).toBe(true);
rstest.expect(browserList.includes(item)).toBe(true);
requiredExpect(requiredList.includes(item)).toBe(true);
rstestTest('context', ({ expect: localExpect }) => { localExpect(contextList.includes(item)).toBe(true); });
import { expect as vitestExpect } from 'vitest';
vitestExpect(foreignList.includes(item)).toBe(true);
function local() { const expect = createExpect(); expect(localList.includes(item)).toBe(true); }
function shadowNaN(NaN: any) { expect([1].includes(NaN)).toBe(true); }
function shadowNumber(Number: any) { expect([1].includes(Number.NaN)).toBe(true); }
function shadowGlobalThis(globalThis: any) { expect([1].includes(globalThis.NaN)).toBe(true); }
Number = { NaN: 1 } as any; expect([1].includes(Number.NaN)).toBe(true);
type LocalNaN = NaN; expect([NaN].includes(NaN)).toBe(true);
interface Number {} expect([Number.NaN].includes(Number.NaN)).toBe(true);
`
	root := fixtures.GetRootDir()
	fileName := tspath.ResolvePath(root.Dir, "prefer-to-contain-source-only.ts")
	fs := utils.NewOverlayVFS(root.FS, map[string]string{fileName: code})
	host := utils.CreateCompilerHost(root.Dir, fs)
	sourceProgram, err := lintprogram.NewFromRoots(lintprogram.RootOptions{
		RootFileNames: []string{fileName}, Host: host,
		CompilerOptions: &core.CompilerOptions{Module: core.ModuleKindESNext}, SingleThreaded: true,
	})
	if err != nil {
		t.Fatalf("NewFromRoots: %v", err)
	}
	if sourceProgram.CanProvideTypeChecker(sourceProgram.SourceFiles()[0]) {
		t.Fatal("expected a source-only Program with no TypeChecker")
	}
	lintPlan, err := linter.PrepareLintPlan(linter.PrepareLintPlanOptions{
		Programs: []*lintprogram.Program{sourceProgram}, TargetsByProgram: [][]string{{fileName}}, SingleThreaded: true,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{Name: PreferToContainRule.Name, Severity: rule.SeverityError, Run: func(ctx rule.RuleContext) rule.RuleListeners {
				return PreferToContainRule.Run(ctx, nil)
			}}}
		},
	})
	if err != nil {
		t.Fatalf("PrepareLintPlan: %v", err)
	}
	var positions []int
	if _, err := linter.RunLinter(linter.RunLinterOptions{
		SingleThreaded: true, LintPlan: lintPlan,
		Consumer: rule.DiagnosticConsumer{Report: func(diagnostic rule.RuleDiagnostic) { positions = append(positions, diagnostic.Range.Pos()) }},
	}); err != nil {
		t.Fatalf("RunLinter: %v", err)
	}
	sort.Ints(positions)
	if len(positions) != 8 {
		t.Fatalf("reported %d assertions, want 8 at %v", len(positions), positions)
	}
}
