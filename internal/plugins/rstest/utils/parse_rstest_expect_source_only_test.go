package utils_test

import (
	"sort"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestParseRstestExpectCallResolvesInSourceOnlyProgram(t *testing.T) {
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

	probe := rule.Rule{
		Name: "rstest/source-only-expect-probe",
		Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
			analysis := rstestUtils.GetRstestCallAnalysis(ctx)
			return rule.RuleListeners{
				ast.KindCallExpression: func(node *ast.Node) {
					parsed := analysis.ParseExpectCall(node)
					if parsed == nil || parsed.MatcherEntry == nil {
						return
					}
					ctx.ReportNode(parsed.MatcherEntry.Node, rule.RuleMessage{
						Id:          "parsedExpect",
						Description: "parsed Rstest expect",
					})
				},
			}
		},
	}

	root := fixtures.GetRootDir()
	fileName := tspath.ResolvePath(root.Dir, "parse-rstest-expect-source-only.ts")
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
				Name:     probe.Name,
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return probe.Run(ctx, nil)
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
		Consumer: rule.DiagnosticConsumer{Report: func(diagnostic rule.RuleDiagnostic) {
			got = append(got, reported{
				pos:  diagnostic.Range.Pos(),
				text: code[diagnostic.Range.Pos():diagnostic.Range.End()],
			})
		}},
	}); err != nil {
		t.Fatalf("RunLinter: %v", err)
	}
	sort.Slice(got, func(left, right int) bool { return got[left].pos < got[right].pos })

	if len(got) != 6 {
		t.Fatalf("parsed %d Rstest expect calls, want 6: %v", len(got), got)
	}
	for index, diagnostic := range got {
		if diagnostic.text != "toBe" {
			t.Errorf("diagnostic %d covers %q, want toBe", index, diagnostic.text)
		}
	}
}
