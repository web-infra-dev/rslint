package max_nested_describe_test

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/max_nested_describe"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestMaxNestedDescribeSourceOnly(t *testing.T) {
	code := `import { describe as suite } from '@rstest/core';
suite('one', () => { suite('inline', () => {}); });

function registerFunctionScope() {
  const body = () => { suite('function inner', () => {}); };
  suite('function outer', body);
}
registerFunctionScope();

{
  const body = () => { suite('block inner', () => {}); };
  suite('block outer', body);
}

class StaticScope {
  static {
    const body = () => { suite('static inner', () => {}); };
    suite('static outer', body);
  }
}

let reassigned = () => { suite('stale inner', () => {}); };
reassigned = () => {};
suite('reassigned outer', reassigned);

var unwritten = () => { suite('unwritten inner', () => {}); };
suite('unwritten outer', unwritten);

function wrap(fn) { return fn; }
function combinedBody() {
  suite('combined middle', wrap(() => { suite('combined inner', () => {}); }));
}
suite('combined outer', combinedBody);

function destructured() { suite('destructured inner', () => {}); }
({ destructured } = replacements);
suite('destructured outer', destructured);`
	root := fixtures.GetRootDir()
	fileName := tspath.ResolvePath(root.Dir, "max-nested-describe-source-only.ts")
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

	var diagnostics []rule.RuleDiagnostic
	lintPlan, err := linter.PrepareLintPlan(linter.PrepareLintPlanOptions{
		Programs:         []*lintprogram.Program{sourceProgram},
		TargetsByProgram: [][]string{{fileName}},
		SingleThreaded:   true,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
				Name:     max_nested_describe.MaxNestedDescribeRule.Name,
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return max_nested_describe.MaxNestedDescribeRule.Run(ctx, maxOption(1))
				},
			}}
		},
	})
	if err != nil {
		t.Fatalf("PrepareLintPlan: %v", err)
	}
	if _, err := linter.RunLinter(linter.RunLinterOptions{
		SingleThreaded: true,
		LintPlan:       lintPlan,
		Consumer: rule.DiagnosticConsumer{Report: func(diagnostic rule.RuleDiagnostic) {
			diagnostics = append(diagnostics, diagnostic)
		}},
	}); err != nil {
		t.Fatalf("RunLinter: %v", err)
	}
	if len(diagnostics) != 7 {
		t.Fatalf("got %d diagnostics, want 7: %v", len(diagnostics), diagnostics)
	}
	wantLines := []int{2, 5, 11, 17, 26, 31, 31}
	wantDepths := []string{
		"Too many nested describe calls (2) - maximum allowed is 1",
		"Too many nested describe calls (2) - maximum allowed is 1",
		"Too many nested describe calls (2) - maximum allowed is 1",
		"Too many nested describe calls (2) - maximum allowed is 1",
		"Too many nested describe calls (2) - maximum allowed is 1",
		"Too many nested describe calls (2) - maximum allowed is 1",
		"Too many nested describe calls (3) - maximum allowed is 1",
	}
	for index, diagnostic := range diagnostics {
		if diagnostic.Message.Description != wantDepths[index] {
			t.Errorf("diagnostic %d = %q, want %q", index, diagnostic.Message.Description, wantDepths[index])
		}
		line, _ := scanner.GetECMALineAndUTF16CharacterOfPosition(
			diagnostic.SourceFile,
			diagnostic.Range.Pos(),
		)
		if got := line + 1; got != wantLines[index] {
			t.Errorf("diagnostic %d starts on line %d, want %d", index, got, wantLines[index])
		}
	}
}
