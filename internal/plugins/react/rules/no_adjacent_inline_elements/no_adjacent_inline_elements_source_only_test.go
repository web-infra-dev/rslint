package no_adjacent_inline_elements

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/cachedvfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/testutil"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestNoAdjacentInlineElementsSourceOnlyShadowing(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := tspath.NormalizePath(filepath.Join(tmpDir, "case.tsx"))
	source := `import { createElement } from "react";
function f(createElement) {
  return createElement("div", null, [createElement("a"), createElement("span")]);
}`
	if err := os.WriteFile(filePath, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	program, err := utils.CreateProgramFromOptionsLenient(true, &core.CompilerOptions{
		Target:          core.ScriptTargetESNext,
		Module:          core.ModuleKindCommonJS,
		ESModuleInterop: core.TSTrue,
		SkipLibCheck:    core.TSTrue,
	}, []string{filePath}, utils.CreateCompilerHost(tmpDir, bundled.WrapFS(cachedvfs.From(osvfs.FS()))))
	if err != nil {
		t.Fatal(err)
	}
	sourceProgram, err := lintprogram.NewFromBoundSources(program, program.SourceFiles())
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	testutil.LintProgram(t, testutil.LintProgramOptions{
		Program:                sourceProgram,
		ExcludedPathSubstrings: testutil.DefaultExcludedPathSubstrings,
		GetRulesForFile: func(sf *ast.SourceFile) []rule.ConfiguredRule {
			if sf.FileName() != filePath {
				return nil
			}
			return []rule.ConfiguredRule{{
				Name:     NoAdjacentInlineElementsRule.Name,
				Severity: rule.SeverityWarning,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return NoAdjacentInlineElementsRule.Run(ctx, nil)
				},
			}}
		},
		OnDiagnostic: func(rule.RuleDiagnostic) { count++ },
	})
	if count != 0 {
		t.Fatalf("source-only shadowed createElement diagnostics = %d, want 0", count)
	}
}
