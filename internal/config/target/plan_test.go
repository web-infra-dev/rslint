// cspell:ignore gitdir

package target

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/cachedvfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/rules"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/pathutil"
)

type retargetingExplicitOwnerFS struct {
	vfs.FS
	targetPath       string
	targetParentPath string
	parentCalls      int
}

func (fsys *retargetingExplicitOwnerFS) CaseSensitivity() tspath.CaseSensitivity {
	return pathutil.CaseSensitivity(false)
}
func (fsys *retargetingExplicitOwnerFS) FileExists(filePath tspath.RootedFilePath) bool {
	return tspath.NormalizePath(filePath.AsString()) == "/repo/shared.ts"
}
func (fsys *retargetingExplicitOwnerFS) Realpath(filePath tspath.RootedPath) tspath.RootedPath {
	filePath = pathutil.Path(tspath.NormalizePath(filePath.AsString()))
	switch filePath.AsString() {
	case tspath.CombinePaths("/repo/Project", tspath.GetBaseFileName(fsys.targetPath)):
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
	plan, err := Resolve(withTestGitPolicy(Request{
		Config: rslintconfig.RslintConfig{{
			Settings: rslintconfig.Settings{"generation": "discovery"},
		}},
		ConfigDirectory: directory,
		ScanRoot:        directory,
		FS:              osvfs.FS(),
		Files:           freezeFilePathsForTest([]string{filePath}, osvfs.FS()),
		SingleThreaded:  true,
	}))
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
		plan.GitPolicy(),
	)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Files[0].Match() == nil {
		t.Fatal("explicit target lost its discovery match")
	}
	resolved := resolver.ResolveTargetWithMatch(plan.Files[0].Identity(), plan.Files[0].Match())
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
	plan, err := Resolve(withTestGitPolicy(Request{
		ConfigMap: map[string]rslintconfig.RslintConfig{
			"/repo/Project": {{}},
		},
		ConfigDirectory: "/repo",
		ScanRoot:        "/repo",
		FS:              fsys,
		Files:           freezeFilePathsForTest([]string{targetPath}, fsys),
		SingleThreaded:  true,
	}))
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
	if fsys.parentCalls != 1 {
		t.Fatalf("target parent Realpath calls = %d, want one frozen observation", fsys.parentCalls)
	}
}

func TestTargetPlanExplicitFileExistenceUsesFrozenIdentity(t *testing.T) {
	for _, aliasKind := range []string{"directory", "file"} {
		for _, exists := range []bool{true, false} {
			name := "captured missing"
			if exists {
				name = "captured present"
			}
			t.Run(aliasKind+"/"+name, func(t *testing.T) {
				fsys := osvfs.FS()
				root := tspath.NormalizePath(fsys.Realpath(pathutil.Path(t.TempDir())).AsString())
				first, replacement := tspath.ResolvePath(root, "first"), tspath.ResolvePath(root, "replacement")
				for _, directory := range []string{first, replacement} {
					if err := os.Mkdir(directory, 0o755); err != nil {
						t.Fatal(err)
					}
				}
				firstFile, replacementFile := tspath.ResolvePath(first, "source.ts"), tspath.ResolvePath(replacement, "source.ts")
				// A leaf symlink must initially resolve so its physical target can
				// be captured. The missing case removes that target after capture.
				if exists || aliasKind == "file" {
					if err := os.WriteFile(firstFile, []byte("export {};\n"), 0o644); err != nil {
						t.Fatal(err)
					}
				}
				if !exists {
					if err := os.WriteFile(replacementFile, []byte("export {};\n"), 0o644); err != nil {
						t.Fatal(err)
					}
				}
				alias, destination, replacementDestination := tspath.ResolvePath(root, "alias"), first, replacement
				path := tspath.ResolvePath(alias, "source.ts")
				if aliasKind == "file" {
					alias, destination, replacementDestination = tspath.ResolvePath(root, "alias.ts"), firstFile, replacementFile
					path = alias
				}
				if err := os.Symlink(destination, alias); err != nil {
					t.Skipf("%s symlink unavailable: %v", aliasKind, err)
				}
				identity := rslintconfig.FreezeFileIdentity(path, fsys)
				if identity.CanonicalPath != firstFile {
					t.Fatalf("captured file = %q, want %q", identity.CanonicalPath, firstFile)
				}
				request := withTestGitPolicy(Request{
					Config:          rslintconfig.RslintConfig{{}},
					ConfigDirectory: root,
					ScanRoot:        root,
					FS:              fsys,
					Files:           []rslintconfig.PathIdentity{identity},
					SingleThreaded:  true,
				})
				if err := os.Remove(alias); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(replacementDestination, alias); err != nil {
					t.Fatal(err)
				}
				if !exists && aliasKind == "file" {
					if err := os.Remove(firstFile); err != nil {
						t.Fatal(err)
					}
				}
				if fsys.FileExists(pathutil.File(path)) == exists || fsys.FileExists(pathutil.File(firstFile)) != exists {
					t.Fatal("alias and captured file must have opposite existence")
				}
				plan, err := Resolve(request)
				if err != nil {
					t.Fatal(err)
				}
				if len(plan.ExplicitFileOutcomes) != 1 || plan.ExplicitFileOutcomes[0].Path != path ||
					plan.ExplicitFileOutcomes[0].Exists != exists || plan.ExplicitFileOutcomes[0].Ignored {
					t.Fatalf("outcome crossed captured identity: %+v", plan.ExplicitFileOutcomes)
				}
				if (len(plan.Files) == 1) != exists || len(plan.Files) > 1 {
					t.Fatalf("selection crossed captured identity: %+v", plan.Files)
				}
				if exists && plan.Files[0].Identity() != identity {
					t.Fatalf("selected identity = %+v, want %+v", plan.Files[0].Identity(), identity)
				}
			})
		}
	}
}

func TestExplicitLintTargetSetPreservesVirtualAndLexicalFallback(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	path, canonical := tspath.ResolvePath(root, "source.ts"), tspath.ResolvePath(root, "physical/source.txt")
	for _, identity := range []rslintconfig.PathIdentity{
		{Path: path},
		{Path: path, CanonicalPath: canonical},
	} {
		key := identity.CanonicalPath
		if key == "" {
			key = path
		}
		fsys := utils.NewOverlayVFS(osvfs.FS(), map[string]string{key: "export {};\n"})
		set := newExplicitLintTargetSet([]rslintconfig.PathIdentity{identity}, fsys)
		selected := set.targetsForIdentities([]rslintconfig.PathIdentity{identity})
		if len(selected) != 1 || !selected[0].selectedBy(root, root, true, nil) {
			t.Fatalf("virtual explicit target was lost: identity=%+v outcomes=%+v", identity, set.outcomes())
		}
	}
}

func TestTargetPlanPreservesIgnoredBeforeMissingWarningPriority(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	target := tspath.CombinePaths(root, "node_modules/pkg/missing.ts")
	plan, err := Resolve(withTestGitPolicy(Request{
		Config:          rslintconfig.RslintConfig{{}},
		ConfigDirectory: root,
		ScanRoot:        root,
		FS:              osvfs.FS(),
		Files:           freezeFilePathsForTest([]string{target}, osvfs.FS()),
		SingleThreaded:  true,
	}))
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
	plan, err := Resolve(withTestGitPolicy(Request{
		Config:          entries,
		ConfigDirectory: configDirectory,
		ScanRoot:        scanRoot,
		FS:              osvfs.FS(),
		Directories:     rslintconfig.ResolveDirectoryIdentities([]string{scanRoot}, osvfs.FS()),
		SingleThreaded:  true,
	}))
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
	merged := rslintconfig.NewFileConfigResolverWithFS(entries, configDirectory, osvfs.FS(), rules.All()).
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

	_, err := Resolve(withTestGitPolicy(Request{
		ConfigMap:       configs,
		ConfigDirectory: tspath.NormalizePath(ownersRoot),
		FS:              fsys,
		Files:           freezeFilePathsForTest([]string{targetA, targetB}, fsys),
		SingleThreaded:  true,
	}))
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

	_, err := Resolve(withTestGitPolicy(Request{
		ConfigMap:       configs,
		ConfigDirectory: tspath.NormalizePath(aliasesRoot),
		FS:              fsys,
		Files:           freezeFilePathsForTest([]string{targetA, targetB}, fsys),
		SingleThreaded:  true,
	}))
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

func TestTargetPlanCarriesDiscoveryMatches(t *testing.T) {
	root, paths := setupDiscoveryFixture(t, []string{"root.ts", "pkg/a.ts", "pkg/explicit.ts"})
	child := tspath.ResolvePath(root, "pkg")
	configs := map[string]rslintconfig.RslintConfig{
		root:  {{Rules: rslintconfig.Rules{"no-console": "error"}}},
		child: {{Rules: rslintconfig.Rules{"no-debugger": "error"}}},
	}
	plan, err := Resolve(withTestGitPolicy(Request{
		ConfigMap: configs, ConfigDirectory: root, FS: osvfs.FS(),
		Files:       freezeFilePathsForTest([]string{paths["pkg/explicit.ts"]}, osvfs.FS()),
		Directories: rslintconfig.ResolveDirectoryIdentities([]string{root}, osvfs.FS()),
	}))
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Files) != 3 {
		t.Fatalf("targets = %v", plan.Files)
	}
	for _, file := range plan.Files {
		match := file.Match()
		if match == nil || !match.Selected() || match.GloballyIgnored() {
			t.Fatalf("%s lost its discovery match", file.Path)
		}
		resolver, err := rslintconfig.NewFileConfigResolverWithPathSpaces(configs[file.ConfigDirectory], file.ConfigDirectory, osvfs.FS(), plan.PathSpaces(), rules.All(), plan.GitPolicy())
		if err != nil {
			t.Fatal(err)
		}
		resolved := resolver.ResolveTargetWithMatch(file.Identity(), match)
		name := "no-console"
		if file.ConfigDirectory == child {
			name = "no-debugger"
		}
		if len(resolved.EnabledRules) != 1 || resolved.EnabledRules[0].Name != name {
			t.Fatalf("%s received another owner's rules", file.Path)
		}
	}
}

func TestTargetPlanRetainsNestedRepositoryEntryUnderIgnoredAncestor(t *testing.T) {
	root := tspath.NormalizePath(osvfs.FS().Realpath(pathutil.Path(t.TempDir())).AsString())
	nested := tspath.ResolvePath(root, "vendor/nested")
	for path, content := range map[string]string{
		".gitignore":             "vendor/\n",
		"root.ts":                "export {};\n",
		"vendor/nested/.git":     "gitdir: /unused/worktree\n",
		"vendor/nested/index.ts": "export {};\n",
		"vendor/nested/other.ts": "export {};\n",
	} {
		fullPath := tspath.ResolvePath(root, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(tspath.ResolvePath(root, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, roots := range [][]string{{root, nested}, {nested, root}, {root, nested, nested}} {
		plan, err := Resolve(withTestGitPolicy(Request{
			Config:          rslintconfig.RslintConfig{{Rules: rslintconfig.Rules{"no-debugger": "error"}}},
			ConfigDirectory: root, ScanRoot: root, FS: osvfs.FS(),
			Directories:    rslintconfig.ResolveDirectoryIdentities(roots, osvfs.FS()),
			SingleThreaded: true,
		}))
		if err != nil {
			t.Fatal(err)
		}
		if len(plan.Files) != 3 {
			t.Fatalf("roots %v lost or duplicated nested entry: %+v", roots, plan.Files)
		}
		if plan.GitPolicy() == nil {
			t.Fatal("plan lost Git policy")
		}
	}
	// A literal file is an independent entry, not permission to scan its siblings.
	if err := os.WriteFile(tspath.ResolvePath(root, ".gitignore"), []byte("vendor/nested/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	file := rslintconfig.FreezeFileIdentity(tspath.ResolvePath(nested, "index.ts"), osvfs.FS())
	plan, err := Resolve(withTestGitPolicy(Request{
		Config:          rslintconfig.RslintConfig{{Rules: rslintconfig.Rules{"no-debugger": "error"}}},
		ConfigDirectory: root, ScanRoot: root, FS: osvfs.FS(),
		Files: []rslintconfig.PathIdentity{file}, Directories: rslintconfig.ResolveDirectoryIdentities([]string{root}, osvfs.FS()),
		SingleThreaded: true,
	}))
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Files) != 2 || plan.Files[0].Path != tspath.ResolvePath(root, "root.ts") || plan.Files[1].Path != file.Path {
		t.Fatalf("literal nested input opened sibling traversal: %+v", plan.Files)
	}
}

type frozenDirectoryRequestFS struct {
	vfs.FS
	root   string
	frozen bool
}

func (fsys *frozenDirectoryRequestFS) Realpath(path tspath.RootedPath) tspath.RootedPath {
	if fsys.frozen && tspath.NormalizePath(path.AsString()) == fsys.root {
		panic("requested root resolved after freezing")
	}
	return fsys.FS.Realpath(path)
}

func TestTargetPlanDoesNotResolveFrozenDirectoryAgain(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	physical := tspath.ResolvePath(root, "physical")
	alias := tspath.ResolvePath(root, "alias")
	configDir := tspath.ResolvePath(root, "config")
	for _, directory := range []string{physical, configDir} {
		if err := os.Mkdir(directory, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(tspath.ResolvePath(physical, "input.ts"), []byte("export {};\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(physical, alias); err != nil {
		t.Skipf("directory symlink unavailable: %v", err)
	}
	fsys := &frozenDirectoryRequestFS{FS: osvfs.FS(), root: alias}
	directories := rslintconfig.ResolveDirectoryIdentities([]string{alias}, fsys)
	fsys.frozen = true
	plan, err := Resolve(withTestGitPolicy(Request{Config: rslintconfig.RslintConfig{{}}, ConfigDirectory: configDir, ScanRoot: configDir, FS: fsys, Directories: directories, SingleThreaded: true}))
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Files) != 1 || plan.Files[0].CanonicalParentPath != directories[0].CanonicalPath {
		t.Fatalf("frozen root was not retained: %+v", plan.Files)
	}
}

func TestTargetPlanIgnoredDirectoryDoesNotDependOnScanRoot(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	ignored := tspath.ResolvePath(root, "ignored")
	for path, content := range map[string]string{".gitignore": "ignored/\n", "ignored/nested/.git": "gitdir: /unused\n", "ignored/nested/index.ts": "export {};\n"} {
		full := tspath.ResolvePath(root, path)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(tspath.ResolvePath(root, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, cwd := range []string{root, ignored} {
		request := withTestGitPolicy(Request{Config: rslintconfig.RslintConfig{{}}, ConfigDirectory: root, ScanRoot: cwd, FS: osvfs.FS(), Directories: rslintconfig.ResolveDirectoryIdentities([]string{ignored}, osvfs.FS()), SingleThreaded: true})
		plan, err := Resolve(request)
		if err != nil {
			t.Fatal(err)
		}
		if len(plan.Files) != 0 {
			t.Fatalf("ignored root reopened when ScanRoot=%s: %+v", cwd, plan.Files)
		}
	}
}
