package utils_test

import (
	"fmt"
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
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestParseRstestExpectCallResolvesAcrossProgramModes(t *testing.T) {
	for _, typed := range []bool{false, true} {
		for _, identityFirst := range []bool{false, true} {
			t.Run(fmt.Sprintf("typed=%t/identityFirst=%t", typed, identityFirst), func(t *testing.T) {
				testRstestExpectCallProgram(t, typed, identityFirst)
			})
		}
	}
}

func testRstestExpectCallProgram(t *testing.T, typed, identityFirst bool) {
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
		Name: "rstest/expect-cache-probe",
		Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
			analysis := rstestUtils.GetRstestCallAnalysis(ctx)
			return rule.RuleListeners{
				ast.KindCallExpression: func(node *ast.Node) {
					var parsed *rstestUtils.ParsedRstestExpectCall
					if identityFirst {
						isExpect := analysis.IsExpectCall(node)
						parsed = analysis.ParseExpectCall(node)
						if isExpect != (parsed != nil) {
							t.Fatalf("identity/full parse mismatch at %d", node.Pos())
						}
					} else {
						parsed = analysis.ParseExpectCall(node)
						if analysis.IsExpectCall(node) != (parsed != nil) {
							t.Fatalf("full parse/identity mismatch at %d", node.Pos())
						}
					}
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
	var sourceProgram *lintprogram.Program
	if typed {
		rawProgram, sourceFile, err := rule_tester.NewProgramHelper(root).CreateTestProgram(
			code,
			"parse-rstest-expect-source-only.ts",
			"tsconfig.json",
		)
		if err != nil {
			t.Fatal(err)
		}
		sourceProgram = lintprogram.NewFromCompiler(rawProgram)
		fileName = sourceFile.FileName()
	} else {
		fs := utils.NewOverlayVFS(root.FS, map[string]string{fileName: code})
		host := utils.CreateCompilerHost(root.Dir, fs)
		var err error
		sourceProgram, err = lintprogram.NewFromRoots(lintprogram.RootOptions{
			RootFileNames:   []string{fileName},
			Host:            host,
			CompilerOptions: &core.CompilerOptions{Module: core.ModuleKindESNext},
			SingleThreaded:  true,
		})
		if err != nil {
			t.Fatalf("NewFromRoots: %v", err)
		}
	}
	if got := sourceProgram.CanProvideTypeChecker(sourceProgram.GetSourceFile(fileName)); got != typed {
		t.Fatalf("CanProvideTypeChecker = %t, want %t", got, typed)
	}

	lintPlan, err := linter.PrepareLintPlan(linter.PrepareLintPlanOptions{
		Programs:         []*lintprogram.Program{sourceProgram},
		TargetsByProgram: [][]string{{fileName}},
		SingleThreaded:   true,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
				Name:             probe.Name,
				Severity:         rule.SeverityError,
				RequiresTypeInfo: typed,
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
