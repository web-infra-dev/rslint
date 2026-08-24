package target

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
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/rules"
)

type retargetingExplicitOwnerFS struct {
	vfs.FS
	targetPath       string
	targetParentPath string
	parentCalls      int
}

func (fsys *retargetingExplicitOwnerFS) UseCaseSensitiveFileNames() bool { return false }
func (fsys *retargetingExplicitOwnerFS) FileExists(filePath string) bool {
	return tspath.NormalizePath(filePath) == fsys.targetPath
}
func (fsys *retargetingExplicitOwnerFS) Realpath(filePath string) string {
	filePath = tspath.NormalizePath(filePath)
	switch filePath {
	case fsys.targetPath:
		return "/repo/shared.ts"
	case fsys.targetParentPath:
		fsys.parentCalls++
		if fsys.parentCalls <= 3 {
			return "/repo/Project"
		}
		return "/moved"
	default:
		return filePath
	}
}

func TestPlanSelectsActiveOwnersAndCallerPaths(t *testing.T) {
	plan := Plan{Files: []File{
		{
			PathIdentity: rslintconfig.PathIdentity{
				Path:          "/repo/b/index.ts",
				CanonicalPath: "/physical/b.ts",
			},
			ConfigDirectory: "/repo/b",
		},
		{
			PathIdentity: rslintconfig.PathIdentity{
				Path:          "/repo/a/alias.ts",
				CanonicalPath: "/physical/a.ts",
			},
			ConfigDirectory: "/repo/a",
		},
		{
			PathIdentity: rslintconfig.PathIdentity{
				Path:          "/repo/a/second-alias.ts",
				CanonicalPath: "/physical/a.ts",
			},
			ConfigDirectory: "/repo/a",
		},
	}}

	owners := plan.ActiveOwners()
	if len(owners) != 2 || owners[0] != "/repo/a" || owners[1] != "/repo/b" {
		t.Fatalf("active owners = %v, want sorted unique owners", owners)
	}
	preferred := plan.PreferredCallerPaths()
	if got := preferred[rslintconfig.ExactPathID("/physical/a.ts")]; got != "/repo/a/alias.ts" {
		t.Fatalf("preferred caller path = %q, want first lexical target", got)
	}
}

func TestTargetPlanAppliesExecutionConfigOverFrozenPathSpaces(t *testing.T) {
	directory := tspath.NormalizePath(t.TempDir())
	filePath := tspath.CombinePaths(directory, "index.ts")
	if err := os.WriteFile(filePath, []byte("export {};\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	plan, err := Resolve(Request{
		Config: rslintconfig.RslintConfig{{
			Settings: rslintconfig.Settings{"generation": "discovery"},
		}},
		ConfigDirectory: directory,
		ScanRoot:        directory,
		FS:              osvfs.FS(),
		Files:           []string{filePath},
		SingleThreaded:  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Files) != 1 {
		t.Fatalf("planned targets = %+v", plan.Files)
	}
	executionConfig := rslintconfig.RslintConfig{{
		Settings: rslintconfig.Settings{"generation": "execution"},
	}}
	resolver, err := rslintconfig.NewFileConfigResolverWithPathSpaces(
		executionConfig,
		directory,
		osvfs.FS(),
		plan.PathSpaces(),
		rules.All(),
		false,
	)
	if err != nil {
		t.Fatal(err)
	}
	resolved := resolver.ResolveTarget(plan.Files[0].Identity())
	if resolved.MergedConfig == nil ||
		resolved.MergedConfig.Settings["generation"] != "execution" {
		t.Fatalf("execution config was replaced by discovery config: %+v", resolved.MergedConfig)
	}
}

func TestTargetPlanFreezesExplicitOwnerAndWarningOutcome(t *testing.T) {
	targetPath := "/repo/project/link.ts"
	fsys := &retargetingExplicitOwnerFS{
		FS:               &configPathSpaceFS{caseSensitive: false},
		targetPath:       targetPath,
		targetParentPath: tspath.GetDirectoryPath(targetPath),
	}
	plan, err := Resolve(Request{
		ConfigMap: map[string]rslintconfig.RslintConfig{
			"/repo/Project": {{}},
		},
		ConfigDirectory: "/repo",
		ScanRoot:        "/repo",
		FS:              fsys,
		Files:           []string{targetPath},
		SingleThreaded:  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Files) != 1 || plan.Files[0].ConfigDirectory != "/repo/Project" {
		t.Fatalf("explicit target lost its assigned owner: %+v", plan.Files)
	}
	if len(plan.ExplicitFileOutcomes) != 1 ||
		plan.ExplicitFileOutcomes[0].Ignored ||
		!plan.ExplicitFileOutcomes[0].Exists {
		t.Fatalf("explicit outcome = %+v", plan.ExplicitFileOutcomes)
	}
	if fsys.parentCalls != 2 {
		t.Fatalf("target parent Realpath calls = %d, want only the two identity reads", fsys.parentCalls)
	}
}

func TestTargetPlanPreservesIgnoredBeforeMissingWarningPriority(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	target := tspath.CombinePaths(root, "node_modules/pkg/missing.ts")
	plan, err := Resolve(Request{
		Config:          rslintconfig.RslintConfig{{}},
		ConfigDirectory: root,
		ScanRoot:        root,
		FS:              osvfs.FS(),
		Files:           []string{target},
		SingleThreaded:  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Files) != 0 || len(plan.ExplicitFileOutcomes) != 1 {
		t.Fatalf("plan = %+v", plan)
	}
	outcome := plan.ExplicitFileOutcomes[0]
	if !outcome.Ignored || outcome.Exists {
		t.Fatalf("ignored missing outcome = %+v", outcome)
	}
}

func TestTargetResolveSeparatesConfigDirectoryFromScanRoot(t *testing.T) {
	root := t.TempDir()
	configDirectory := filepath.Join(root, "config")
	scanRoot := filepath.Join(root, "workspace")
	target := filepath.Join(scanRoot, "src", "target.ts")
	for _, directory := range []string{
		configDirectory,
		filepath.Dir(target),
	} {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for path, content := range map[string]string{
		target: "export const target = 1;\n",
		filepath.Join(configDirectory, "config-only.ts"): "export const configOnly = 1;\n",
	} {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	configDirectory = tspath.NormalizePath(configDirectory)
	scanRoot = tspath.NormalizePath(scanRoot)
	target = tspath.NormalizePath(target)
	entries := rslintconfig.RslintConfig{{
		Files: []string{"../workspace/src/*.ts"},
		Rules: rslintconfig.Rules{"no-debugger": "error"},
	}}
	plan, err := Resolve(Request{
		Config:          entries,
		ConfigDirectory: configDirectory,
		ScanRoot:        scanRoot,
		FS:              osvfs.FS(),
		Directories:     []string{scanRoot},
		SingleThreaded:  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Files) != 1 || plan.Files[0].Path != target {
		t.Fatalf("targets = %+v, want only %q", plan.Files, target)
	}
	if got := plan.Files[0].ConfigDirectory; got != configDirectory {
		t.Fatalf("target config directory = %q, want %q", got, configDirectory)
	}
	resolvedTarget := plan.Files[0]
	merged := rslintconfig.NewFileConfigResolverWithFS(entries, configDirectory, osvfs.FS(), rules.All(), false).
		ConfigForTarget(resolvedTarget.Path, resolvedTarget.CanonicalPath)
	if merged == nil || merged.Rules["no-debugger"] == nil {
		t.Fatalf("external target %q did not retain config-relative matching: %#v", resolvedTarget.Path, merged)
	}
}

func TestTargetResolveRejectsPhysicalAliasesWithDifferentOwners(t *testing.T) {
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
	configs := map[string]rslintconfig.RslintConfig{
		ownerA: {{Rules: rslintconfig.Rules{"no-debugger": "error"}}},
		ownerB: {{Rules: rslintconfig.Rules{"no-console": "error"}}},
	}
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))

	_, err := Resolve(Request{
		ConfigMap:       configs,
		ConfigDirectory: tspath.NormalizePath(ownersRoot),
		FS:              fsys,
		Files:           []string{targetA, targetB},
		SingleThreaded:  true,
	})
	if err == nil || !strings.Contains(err.Error(), "governed by different configs") {
		t.Fatalf("ownership conflict error = %v", err)
	}
}

func TestTargetResolveRejectsSamePhysicalConfigAliasesAsDifferentOwners(t *testing.T) {
	physicalOwner := t.TempDir()
	physicalTarget := filepath.Join(physicalOwner, "target.ts")
	if err := os.WriteFile(physicalTarget, []byte("export const value = 1;\n"), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}

	aliasesRoot := t.TempDir()
	ownerA := filepath.Join(aliasesRoot, "owner-a")
	ownerB := filepath.Join(aliasesRoot, "owner-b")
	if err := os.Symlink(physicalOwner, ownerA); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if err := os.Symlink(physicalOwner, ownerB); err != nil {
		t.Skipf("second symlink unavailable: %v", err)
	}

	ownerA = tspath.NormalizePath(ownerA)
	ownerB = tspath.NormalizePath(ownerB)
	targetA := tspath.CombinePaths(ownerA, "target.ts")
	targetB := tspath.CombinePaths(ownerB, "target.ts")
	configs := map[string]rslintconfig.RslintConfig{
		ownerA: {{Rules: rslintconfig.Rules{"owner-a": "error"}}},
		ownerB: {{Rules: rslintconfig.Rules{"owner-b": "error"}}},
	}
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))

	_, err := Resolve(Request{
		ConfigMap:       configs,
		ConfigDirectory: tspath.NormalizePath(aliasesRoot),
		FS:              fsys,
		Files:           []string{targetA, targetB},
		SingleThreaded:  true,
	})
	if err == nil || !strings.Contains(err.Error(), "governed by different configs") {
		t.Fatalf("ownership conflict error = %v", err)
	}
}

// discoverFilesOutsideProgramsForTest preserves the historical discovery
// test matrix without publishing Program membership as a config-layer API.
// Production code discovers owned targets first; program/loader decides their
// membership afterward.
func discoverFilesOutsideProgramsForTest(
	config rslintconfig.RslintConfig,
	configDir string,
	fsys vfs.FS,
	programFiles map[string]struct{},
	allowFiles []string,
	allowDirs []string,
	singleThreaded bool,
) []string {
	targets := discoverLintFiles(config, configDir, fsys, allowFiles, allowDirs, singleThreaded)
	result := make([]string, 0, len(targets))
	for _, target := range targets {
		if _, exists := programFiles[target]; !exists {
			result = append(result, target)
		}
	}
	return result
}

func discoverFilesOutsideProgramsMultiConfigForTest(
	configMap map[string]rslintconfig.RslintConfig,
	fsys vfs.FS,
	programFiles map[string]struct{},
	allowFiles []string,
	allowDirs []string,
	singleThreaded bool,
) []string {
	if len(configMap) == 0 {
		return nil
	}

	seen := make(map[string]struct{})
	var result []string
	for _, target := range discoverLintTargetsMultiConfig(
		configMap,
		nil,
		fsys,
		allowFiles,
		allowDirs,
		singleThreaded,
	) {
		if _, exists := programFiles[target.Path]; exists {
			continue
		}
		if _, exists := seen[target.Path]; exists {
			continue
		}
		seen[target.Path] = struct{}{}
		result = append(result, target.Path)
	}

	sort.Strings(result)
	return result
}
