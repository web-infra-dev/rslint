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

let { expect: replacedExpect } = require('@rstest/core');
replacedExpect = createAssertionLibrary();
replacedExpect(foreignValue).toBe(1);
let { expect: replacedMeta } = import.meta.rstest;
[replacedMeta] = replacements;
replacedMeta(foreignValue).toBe(1);
rstestTest('reassigned context', ({ expect: localExpect }) => {
  localExpect = createAssertionLibrary();
  localExpect(foreignValue).toBe(1);
});

var { expect: reinitialized } = require('@rstest/core');
var reinitialized = createAssertionLibrary();
reinitialized(foreignValue).toBe(1);
var { expect: redeclared } = require('@rstest/core');
var redeclared;
redeclared(importedValue).toBe(1);
var { expect: loopExpect } = require('@rstest/core');
for (var loopExpect of replacements) {}
loopExpect(foreignValue).toBe(1);

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

	if len(got) != 7 {
		t.Fatalf("parsed %d Rstest expect calls, want 7: %v", len(got), got)
	}
	for index, diagnostic := range got {
		if diagnostic.text != "toBe" {
			t.Errorf("diagnostic %d covers %q, want toBe", index, diagnostic.text)
		}
	}
}

func TestParseRstestExpectCallRejectsWrittenGlobalInSourceOnlyProgram(t *testing.T) {
	for _, code := range []string{
		`expect = replacement; expect(value).toBe(1);`,
		`expect ||= replacement; expect(value).toBe(1);`,
		`expect++; expect(value).toBe(1);`,
		`[expect] = replacements; expect(value).toBe(1);`,
		`({ expect } = replacement); expect(value).toBe(1);`,
		`for (expect of replacements) {} expect(value).toBe(1);`,
		`for (expect in replacements) {} expect(value).toBe(1);`,
		`expect(value).toBe(1); expect = replacement;`,
	} {
		if got := sourceOnlyParsedExpectCount(t, code); got != 0 {
			t.Errorf("parsed %d Rstest expect calls, want 0 for %q", got, code)
		}
	}
}

func TestParseRstestExpectCallKeepsGlobalForUnrelatedWritesInSourceOnlyProgram(t *testing.T) {
	for _, code := range []string{
		`function mutate(expect: any) { expect = replacement; } expect(value).toBe(1);`,
		`{ let expect; expect = replacement; } expect(value).toBe(1);`,
		`expect.customMatcher = matcher; expect(value).toBe(1);`,
	} {
		if got := sourceOnlyParsedExpectCount(t, code); got != 1 {
			t.Errorf("parsed %d Rstest expect calls, want 1 for %q", got, code)
		}
	}
}

func sourceOnlyParsedExpectCount(t *testing.T, code string) int {
	t.Helper()
	root := fixtures.GetRootDir()
	fileName := tspath.ResolvePath(root.Dir, "parse-rstest-expect-global-write.ts")
	fs := utils.NewOverlayVFS(root.FS, map[string]string{fileName: code})
	program, err := lintprogram.NewFromRoots(lintprogram.RootOptions{
		RootFileNames:   []string{fileName},
		Host:            utils.CreateCompilerHost(root.Dir, fs),
		CompilerOptions: &core.CompilerOptions{Module: core.ModuleKindESNext},
		SingleThreaded:  true,
	})
	if err != nil {
		t.Fatalf("NewFromRoots: %v", err)
	}
	if program.CanProvideTypeChecker(program.SourceFiles()[0]) {
		t.Fatal("expected source-only program")
	}

	count := 0
	linter.LintSingleFile(linter.LintSingleFileOptions{
		Program: program,
		File:    fileName,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
				Name: "rstest/source-only-expect-write-probe",
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					analysis := rstestUtils.GetRstestCallAnalysis(ctx)
					return rule.RuleListeners{
						ast.KindCallExpression: func(node *ast.Node) {
							if analysis.ParseExpectCall(node) != nil {
								count++
							}
						},
					}
				},
			}}
		},
		Consumer: rule.DiagnosticConsumer{Report: func(rule.RuleDiagnostic) {}},
	})
	return count
}

func TestRstestExpectCustomizationResolvesInSourceOnlyProgram(t *testing.T) {
	code := `
import { expect as check } from '@rstest/core';
import * as core from 'rstack/test';

const verify = check;
const again = verify;
again.extend({ toBe() {} });
const checkEquality = core.expect;
const { addEqualityTesters: add } = checkEquality;
add([tester]);
probe();
`

	probe := rule.Rule{
		Name: "rstest/source-only-expect-customization-probe",
		Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
			analysis := rstestUtils.GetRstestCallAnalysis(ctx)
			return rule.RuleListeners{
				ast.KindCallExpression: func(node *ast.Node) {
					callee := node.Expression()
					if callee == nil || callee.Kind != ast.KindIdentifier || callee.Text() != "probe" {
						return
					}
					ctx.ReportNode(node, probeMessage("customization", fmt.Sprintf(
						"toBe=%t equality=%t",
						analysis.IsExpectMatcherOverridden("toBe"),
						analysis.HasCustomEqualityTesters(),
					)))
				},
			}
		},
	}

	root := fixtures.GetRootDir()
	fileName := tspath.ResolvePath(root.Dir, "expect-customization-source-only.ts")
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
				Run:      func(ctx rule.RuleContext) rule.RuleListeners { return probe.Run(ctx, nil) },
			}}
		},
	})
	if err != nil {
		t.Fatalf("PrepareLintPlan: %v", err)
	}
	var got []rule.RuleDiagnostic
	if _, err := linter.RunLinter(linter.RunLinterOptions{
		SingleThreaded: true,
		LintPlan:       lintPlan,
		Consumer: rule.DiagnosticConsumer{Report: func(diagnostic rule.RuleDiagnostic) {
			got = append(got, diagnostic)
		}},
	}); err != nil {
		t.Fatalf("RunLinter: %v", err)
	}
	if len(got) != 1 || got[0].Message.Description != "toBe=true equality=true" {
		t.Fatalf("source-only customization = %+v", got)
	}
}
