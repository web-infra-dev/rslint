package utils_test

import (
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

func TestParseRstestFnCallRespectsSourceOnlyNamespaceShadowing(t *testing.T) {
	code := `import { it as importedIt } from '@rstest/core';
importedIt('import');
const scenario = importedIt; scenario('alias');
namespace group {
  namespace importedIt {}
  importedIt('shadowed import');
  namespace it {}
  it('shadowed global');
}
namespace test {}
test('shadowed top-level global');
it('global');`

	probe := rule.Rule{
		Name: "rstest/source-only-fn-probe",
		Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
			analysis := rstestUtils.GetRstestCallAnalysis(ctx)
			return rule.RuleListeners{
				ast.KindCallExpression: func(node *ast.Node) {
					parsed := analysis.ParseFnCall(node)
					if parsed == nil || parsed.Kind != rstestUtils.RstestFnTypeTest {
						return
					}
					ctx.ReportNode(node, rule.RuleMessage{Id: "parsedTest", Description: parsed.Name})
				},
			}
		},
	}

	root := fixtures.GetRootDir()
	fileName := tspath.ResolvePath(root.Dir, "parse-rstest-fn-source-only.ts")
	fs := utils.NewOverlayVFS(root.FS, map[string]string{fileName: code})
	program, err := lintprogram.NewFromRoots(lintprogram.RootOptions{
		RootFileNames:   []string{fileName},
		Host:            utils.CreateCompilerHost(root.Dir, fs),
		CompilerOptions: &core.CompilerOptions{Module: core.ModuleKindESNext},
		SingleThreaded:  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if program.CanProvideTypeChecker(program.SourceFiles()[0]) {
		t.Fatal("expected source-only Program")
	}

	var diagnostics []rule.RuleDiagnostic
	linter.LintSingleFile(linter.LintSingleFileOptions{
		Program: program,
		File:    fileName,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
				Name:     probe.Name,
				Severity: rule.SeverityError,
				Run:      func(ctx rule.RuleContext) rule.RuleListeners { return probe.Run(ctx, nil) },
			}}
		},
		Consumer: rule.DiagnosticConsumer{Report: func(diagnostic rule.RuleDiagnostic) {
			diagnostics = append(diagnostics, diagnostic)
		}},
	})

	if len(diagnostics) != 3 {
		t.Fatalf("parsed %d Rstest test calls, want 3: %+v", len(diagnostics), diagnostics)
	}
	for index, diagnostic := range diagnostics {
		if diagnostic.Message.Description != "it" {
			t.Errorf("diagnostic %d resolved %q, want it", index, diagnostic.Message.Description)
		}
	}
}
