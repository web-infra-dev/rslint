package config

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
)

func TestLintTargetPlanSelectsActiveConfigsAndCallerPaths(t *testing.T) {
	configA := RslintConfig{{Rules: Rules{"no-debugger": "error"}}}
	configB := RslintConfig{{Rules: Rules{"no-console": "error"}}}
	configs := map[string]RslintConfig{
		"/repo/a": configA,
		"/repo/b": configB,
	}
	plan := LintTargetPlan{Targets: []DiscoveredLintTarget{
		{
			Path:            "/repo/a/alias.ts",
			CanonicalPath:   "/physical/a.ts",
			ConfigDirectory: "/repo/a",
		},
		{
			Path:            "/repo/a/second-alias.ts",
			CanonicalPath:   "/physical/a.ts",
			ConfigDirectory: "/repo/a",
		},
	}}

	active := plan.ActiveConfigs(configs)
	if len(active) != 1 || active["/repo/a"] == nil {
		t.Fatalf("active configs = %v, want only /repo/a", active)
	}
	preferred := plan.PreferredCallerPaths()
	if got := preferred[exactPathID("/physical/a.ts")]; got != "/repo/a/alias.ts" {
		t.Fatalf("preferred caller path = %q, want first lexical target", got)
	}
}

func TestResolveLintTargetPlanRejectsPhysicalAliasesWithDifferentOwners(t *testing.T) {
	sharedDir := t.TempDir()
	sharedTarget := filepath.Join(sharedDir, "target.ts")
	if err := os.WriteFile(sharedTarget, []byte("export const value = 1;\n"), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}

	ownersRoot := t.TempDir()
	ownerA := filepath.Join(ownersRoot, "owner-a")
	ownerB := filepath.Join(ownersRoot, "owner-b")
	for _, owner := range []string{ownerA, ownerB} {
		if err := os.MkdirAll(owner, 0o755); err != nil {
			t.Fatalf("mkdir owner: %v", err)
		}
	}
	targetA := filepath.Join(ownerA, "target.ts")
	targetB := filepath.Join(ownerB, "target.ts")
	if err := os.Symlink(sharedTarget, targetA); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if err := os.Symlink(sharedTarget, targetB); err != nil {
		t.Skipf("second symlink unavailable: %v", err)
	}

	ownerA = tspath.NormalizePath(ownerA)
	ownerB = tspath.NormalizePath(ownerB)
	targetA = tspath.NormalizePath(targetA)
	targetB = tspath.NormalizePath(targetB)
	configs := map[string]RslintConfig{
		ownerA: {{Rules: Rules{"no-debugger": "error"}}},
		ownerB: {{Rules: Rules{"no-console": "error"}}},
	}
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))

	_, err := ResolveLintTargetPlan(
		configs,
		nil,
		tspath.NormalizePath(ownersRoot),
		nil,
		fsys,
		[]string{targetA, targetB},
		nil,
		true,
	)
	if err == nil || !strings.Contains(err.Error(), "governed by different configs") {
		t.Fatalf("ownership conflict error = %v", err)
	}
}

// discoverFilesOutsideProgramsForTest preserves the historical discovery
// test matrix without publishing Program membership as a config-layer API.
// Production code discovers owned targets first; program/loader decides their
// membership afterward.
func discoverFilesOutsideProgramsForTest(
	config RslintConfig,
	configDir string,
	fsys vfs.FS,
	programFiles map[string]struct{},
	allowFiles []string,
	allowDirs []string,
	singleThreaded bool,
) []string {
	targets := DiscoverLintFiles(config, configDir, fsys, allowFiles, allowDirs, singleThreaded)
	result := make([]string, 0, len(targets))
	for _, target := range targets {
		if _, exists := programFiles[target]; !exists {
			result = append(result, target)
		}
	}
	return result
}

func discoverFilesOutsideProgramsMultiConfigForTest(
	configMap map[string]RslintConfig,
	fsys vfs.FS,
	programFiles map[string]struct{},
	allowFiles []string,
	allowDirs []string,
	singleThreaded bool,
) []string {
	if len(configMap) == 0 {
		return nil
	}

	index := newConfigDirectoryIndex(configMap, fsys)
	configDirs := make([]string, 0, len(configMap))
	for configDir := range configMap {
		configDirs = append(configDirs, configDir)
	}
	sort.Strings(configDirs)

	seen := make(map[string]struct{})
	var result []string
	for _, configDir := range configDirs {
		targets := discoverLintTargetsForConfigInMap(
			configMap,
			index,
			nil,
			configDir,
			fsys,
			allowFiles,
			allowDirs,
			singleThreaded,
		)
		for _, target := range targets {
			if _, exists := programFiles[target.Path]; exists {
				continue
			}
			if _, exists := seen[target.Path]; exists {
				continue
			}
			seen[target.Path] = struct{}{}
			result = append(result, target.Path)
		}
	}

	sort.Strings(result)
	return result
}
