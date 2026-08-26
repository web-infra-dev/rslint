package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/target"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/program/loader"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rules"
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
	loaded, err := programSession.LoadCLI(
		loader.ProjectSet{},
		target.Plan{Files: []target.File{{PathIdentity: rslintconfig.PathIdentity{Path: targetFile,
			CanonicalPath: targetFile}, ConfigDirectory: tmpDir,
		}}},
		tmpDir,
		false,
	)
	if err != nil || len(loaded.Programs) != 1 {
		t.Fatalf("load source-only Program: programs=%d err=%v", len(loaded.Programs), err)
	}

	cfg := rslintconfig.RslintConfig{
		rslintconfig.ConfigEntry{
			Files:   []string{"**/*.ts"},
			Rules:   rslintconfig.Rules{"@typescript-eslint/no-unsafe-member-access": "error"},
			Plugins: []string{"@typescript-eslint"},
		},
	}
	// Deliberately bypass the config resolver's type-info gate.
	configuredRules, _ := rslintconfig.ResolveEnabledRules(rules.All(), cfg, targetFile, tmpDir, false)
	if len(configuredRules) != 1 || configuredRules[0].Name != "@typescript-eslint/no-unsafe-member-access" || !configuredRules[0].RequiresTypeInfo {
		t.Fatalf("fixture did not resolve the expected type-aware rule: %+v", configuredRules)
	}
	lintPlan, err := linter.PrepareLintPlan(linter.PrepareLintPlanOptions{
		Programs:         loaded.Programs,
		TargetsByProgram: loaded.TargetsByProgram,
		GetRulesForFile: func(*ast.SourceFile) []linter.ConfiguredRule {
			return configuredRules
		},
	})
	if err != nil {
		t.Fatalf("PrepareLintPlan: %v", err)
	}
	result, err := linter.RunLinter(linter.RunLinterOptions{
		LintPlan: lintPlan,
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
