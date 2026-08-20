package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/linter"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/program/loader"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Program capability is the framework gate for RequiresTypeInfo rules. Config
// resolution can return the complete rule set without knowing how source
// services were assembled.
func TestGate_LinterFiltersTypeAwareRuleOnSourceOnlyProgram(t *testing.T) {
	tmpDir := tspath.NormalizePath(t.TempDir())
	targetFile := tspath.NormalizePath(filepath.Join(tmpDir, "source-only.ts"))

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	fs = utils.NewOverlayVFS(fs, map[string]string{targetFile: "let a: any = 10;\na.b = 20;\n"})

	programSession := loader.NewSession(fs)
	lintProjectPlan := rslintconfig.LintProjectPlan{Targets: []rslintconfig.PlannedLintTarget{{
		Target: rslintconfig.DiscoveredLintTarget{
			Path:            targetFile,
			CanonicalPath:   targetFile,
			ConfigDirectory: tmpDir,
		},
		MatchPath: targetFile,
	}}}
	projectSet, err := programSession.SelectProjects(lintProjectPlan, nil, false, false)
	if err != nil {
		t.Fatalf("select source-only target: %v", err)
	}
	loaded, err := programSession.LoadCLI(
		projectSet,
		lintProjectPlan,
		tmpDir,
		false,
	)
	if err != nil || len(loaded.Programs) != 1 {
		t.Fatalf("load source-only Program: programs=%d err=%v", len(loaded.Programs), err)
	}

	rslintconfig.RegisterAllRules()
	cfg := rslintconfig.RslintConfig{
		rslintconfig.ConfigEntry{
			Files:   []string{"**/*.ts"},
			Rules:   rslintconfig.Rules{"@typescript-eslint/no-unsafe-member-access": "error"},
			Plugins: []string{"@typescript-eslint"},
		},
	}
	// Deliberately bypass the config resolver's type-info gate.
	rules, _ := rslintconfig.GlobalRuleRegistry.GetEnabledRules(cfg, targetFile, tmpDir, false)
	if len(rules) != 1 || rules[0].Name != "@typescript-eslint/no-unsafe-member-access" || !rules[0].RequiresTypeInfo {
		t.Fatalf("fixture did not resolve the expected type-aware rule: %+v", rules)
	}
	result, err := linter.RunLinter(linter.RunLinterOptions{
		Programs: loaded.Programs,
		Scope:    linter.FileScope{Files: []string{targetFile}},
		GetRulesForFile: func(*ast.SourceFile) []linter.ConfiguredRule {
			return rules
		},
		Consumer: rule.DiagnosticConsumer{
			Report: func(rule.RuleDiagnostic) {},
		},
	})
	if err != nil {
		t.Fatalf("RunLinter: %v", err)
	}
	if _, ran := result.ExecutedRules["@typescript-eslint/no-unsafe-member-access"]; ran {
		t.Fatal("type-aware rule bypassed the Program capability filter")
	}
}

func writeProgramTestFiles(t *testing.T, directory string, files map[string]string) {
	t.Helper()
	for relativePath, content := range files {
		fileName := filepath.Join(directory, relativePath)
		if err := os.MkdirAll(filepath.Dir(fileName), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(fileName), err)
		}
		if err := os.WriteFile(fileName, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", fileName, err)
		}
	}
}

func createTestProgram(t *testing.T, files map[string]string) *compiler.Program {
	t.Helper()
	directory := t.TempDir()
	writeProgramTestFiles(t, directory, files)
	if err := os.WriteFile(filepath.Join(directory, "tsconfig.json"), []byte(`{"include":["**/*.ts"]}`), 0o644); err != nil {
		t.Fatalf("write tsconfig: %v", err)
	}
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	program, err := utils.CreateProgram(
		true,
		fsys,
		directory,
		"tsconfig.json",
		utils.CreateCompilerHost(directory, fsys),
	)
	if err != nil {
		t.Fatalf("create Program: %v", err)
	}
	return program
}

func createLenientSyntacticProgram(t *testing.T) (*lintprogram.Program, string) {
	t.Helper()
	directory := t.TempDir()
	target := tspath.NormalizePath(filepath.Join(directory, "target.ts"))
	writeProgramTestFiles(t, directory, map[string]string{"target.ts": "export const broken = ;\n"})
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	program, err := utils.CreateProgramFromOptionsLenient(
		true,
		&core.CompilerOptions{NoLib: core.TSTrue},
		[]string{target},
		utils.CreateCompilerHost(directory, fsys),
	)
	if err != nil {
		t.Fatalf("create lenient Program: %v", err)
	}
	return lintprogram.NewFromCompiler(program), target
}

func TestCollectTargetSyntacticDiagnosticsDefersOnlyCatalogPrograms(t *testing.T) {
	catalogProgram, catalogTarget := createLenientSyntacticProgram(t)
	lintOnlyProgram, lintOnlyTarget := createLenientSyntacticProgram(t)
	programs := []*lintprogram.Program{catalogProgram, lintOnlyProgram}
	targets := [][]string{{catalogTarget}, {lintOnlyTarget}}

	diagnostics := collectTargetSyntacticDiagnostics(
		programs,
		programs[:1],
		targets,
		true,
		false,
	)
	if len(diagnostics) != 1 || diagnostics[0].FilePath != lintOnlyTarget {
		t.Fatalf("lint-only syntax diagnostics = %+v, want only %q", diagnostics, lintOnlyTarget)
	}

	if diagnostics := collectTargetSyntacticDiagnostics(programs, nil, targets, true, false); len(diagnostics) != 0 {
		t.Fatalf("nil catalog must preserve all-program type-check coverage, got %+v", diagnostics)
	}
}
