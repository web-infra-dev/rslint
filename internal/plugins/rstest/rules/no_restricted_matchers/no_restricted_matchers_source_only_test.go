// TestNoRestrictedMatchersResolvesInSourceOnlyProgram verifies that the
// binder distinguishes Rstest expect sources from foreign and local bindings
// when no TypeChecker is available.
package no_restricted_matchers

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

func TestNoRestrictedMatchersResolvesInSourceOnlyProgram(t *testing.T) {
	if NoRestrictedMatchersRule.RequiresTypeInfo {
		t.Fatal("rstest/no-restricted-matchers must not require type information")
	}

	code := `
expect(globalValue).toBe(1);

import { expect as check, test as rstestTest } from '@rstest/core';
import * as rstest from '@rstest/playwright';
const { expect: requiredExpect } = require('rstack/test');

check(importedValue).toBe(1);
rstest.expect(browserValue).toBe(1);
requiredExpect(requiredValue).toBe(1);
rstestTest('context receiver', (ctx) => { ctx.expect(contextValue).toBe(1); });
rstestTest('context alias', ({ expect: contextExpect }) => { contextExpect(contextValue).toBe(1); });

import { expect as vitestExpect } from 'vitest';
import { expect as jestExpect } from '@jest/globals';

vitestExpect(foreignValue).toBe(1);
jestExpect(foreignValue).toBe(1);
function localAssertion() {
  const expect = createAssertionLibrary();
  expect(localValue).toBe(1);
}
`
	want := []string{"toBe", "toBe", "toBe", "toBe", "toBe", "toBe"}

	root := fixtures.GetRootDir()
	fileName := tspath.ResolvePath(root.Dir, "no-restricted-matchers-source-only.ts")
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
				Name:     NoRestrictedMatchersRule.Name,
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return NoRestrictedMatchersRule.Run(ctx, restrictions("toBe", nil))
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
		t.Fatalf("reported %d matcher chains, want %d: %v", len(got), len(want), got)
	}
	for index, expected := range want {
		if got[index].text != expected {
			t.Errorf("diagnostic %d covers %q, want %q", index, got[index].text, expected)
		}
	}
}
