package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/config/gitignore"
)

type gitignoreReadSpyFS struct {
	vfs.FS
	reads []string
}

func (fsys *gitignoreReadSpyFS) ReadFile(path string) (string, bool) {
	fsys.reads = append(fsys.reads, tspath.NormalizePath(path))
	return fsys.FS.ReadFile(path)
}

func TestCollectedGitignoreUsesItsOwnMatchDirectory(t *testing.T) {
	config := ConfigWithCollectedGitignore(
		RslintConfig{
			{Ignores: []string{"../workspace/config-ignored.ts"}},
			{Rules: Rules{"rule": "error"}},
		},
		[]gitignore.Pattern{{
			Glob:           "git-ignored.ts",
			NodeGlob:       "git-ignored.ts",
			MatchDirectory: "/repo/workspace",
		}},
		false,
	)

	for _, test := range []struct {
		name    string
		path    string
		ignored bool
	}{
		{name: "Git pattern uses invocation root", path: "/repo/workspace/git-ignored.ts", ignored: true},
		{name: "authored ignore uses config directory", path: "/repo/workspace/config-ignored.ts", ignored: true},
		{name: "same basename outside Git-ignore root", path: "/repo/config/git-ignored.ts", ignored: false},
		{name: "visible target", path: "/repo/workspace/visible.ts", ignored: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := config.IsFileIgnored(test.path, "/repo/config"); got != test.ignored {
				t.Fatalf("IsFileIgnored(%q) = %t, want %t", test.path, got, test.ignored)
			}
		})
	}
}

func TestCollectedGitignoreUsesConfigMatchSpaceBeforeLexicalFallback(t *testing.T) {
	config := ConfigWithCollectedGitignore(
		RslintConfig{{Rules: Rules{"rule": "error"}}},
		[]gitignore.Pattern{{
			Glob:                  "ignored.js",
			NodeGlob:              "ignored.js",
			MatchDirectory:        "/physical/workspace",
			LexicalMatchDirectory: "/repo/workspace",
		}},
		false,
	)

	if !config.IsFileIgnored("/physical/workspace/ignored.js", "/physical/config") {
		t.Fatal("Git pattern did not match the config-space projection")
	}
	if !config.IsFileIgnored("/repo/workspace/ignored.js", "/repo/config-link") {
		t.Fatal("Git pattern lost its lexical compatibility fallback")
	}
	if config.IsFileIgnored("/physical/config/ignored.js", "/physical/config") {
		t.Fatal("Git pattern escaped its projected invocation root")
	}
}

func TestCollectedGitignoreMatchDirectoryIsCaseInsensitiveOnWindows(t *testing.T) {
	config := ConfigWithCollectedGitignore(
		RslintConfig{{Rules: Rules{"rule": "error"}}},
		[]gitignore.Pattern{{
			Glob:           "src/generated.ts",
			NodeGlob:       "src/generated.ts",
			MatchDirectory: "C:/Repo/Workspace",
		}},
		true,
	)

	if !config.IsFileIgnored("c:/repo/workspace/SRC/GENERATED.TS", "D:/Config") {
		t.Fatal("case-insensitive Git pattern did not match the Windows path alias")
	}
	if config.IsFileIgnored("D:/Config/src/generated.ts", "D:/Config") {
		t.Fatal("Git pattern escaped its Windows drive/root")
	}
}

func TestCollectedGitignoreMatchDirectoryControlsDirectoryPruning(t *testing.T) {
	config := ConfigWithCollectedGitignore(
		RslintConfig{{Rules: Rules{"rule": "error"}}},
		[]gitignore.Pattern{{
			Glob:           "generated/**/*",
			NodeGlob:       "generated",
			DirectoryOnly:  true,
			MatchDirectory: "/repo/workspace",
		}},
		false,
	)
	resolver := newConfigTargetResolver(config, "/repo/config", nil)
	if !resolver.canPruneDirectory("/repo/workspace/generated", "") {
		t.Fatal("Git-ignore-rooted directory was not pruned")
	}
	if resolver.canPruneDirectory("/repo/config/generated", "") {
		t.Fatal("same relative directory outside the Git-ignore root was pruned")
	}
}

func TestCollectedGitignoreChoosesOneAuthoritativeScope(t *testing.T) {
	root := t.TempDir()
	physicalRoot := tspath.NormalizePath(filepath.Join(root, "physical"))
	physicalChild := tspath.CombinePaths(physicalRoot, "child")
	aliasRoot := tspath.NormalizePath(filepath.Join(root, "alias-root"))
	aliasChild := tspath.NormalizePath(filepath.Join(root, "alias-child"))
	for _, directory := range []string{physicalChild, tspath.CombinePaths(physicalChild, "nested")} {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	target := tspath.CombinePaths(physicalChild, "ignored.ts")
	if err := os.WriteFile(target, []byte("debugger;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(physicalRoot, aliasRoot); err != nil {
		t.Skipf("symlinks are unavailable: %v", err)
	}
	if err := os.Symlink(physicalChild, aliasChild); err != nil {
		t.Skipf("symlinks are unavailable: %v", err)
	}

	first := CollectedGitignorePatternsForRoot([]gitignore.Pattern{{
		Glob:     "child/ignored.ts",
		NodeGlob: "child/ignored.ts",
	}}, "/config", aliasRoot, osvfs.FS())
	second := CollectedGitignorePatternsForRoot([]gitignore.Pattern{{
		Glob:     "ignored.ts",
		NodeGlob: "ignored.ts",
		Negated:  true,
	}}, "/config", aliasChild, osvfs.FS())
	effective := ConfigWithCollectedGitignoreScopes(
		RslintConfig{{Rules: Rules{"rule": "error"}}},
		append(first, second...),
		[]string{aliasRoot, aliasChild},
		"/config",
		osvfs.FS(),
		false,
	)

	// The physical spelling falls under both physical roots. Stable scope order
	// assigns it to aliasRoot, so aliasChild's negation cannot override it.
	resolver := newConfigTargetResolver(effective, "/config", osvfs.FS())
	if !resolver.resolve(target, osvfs.FS().Realpath(target)).globallyIgnored {
		t.Fatal("physical target borrowed patterns from more than one Git scope")
	}
}

func TestConfigWithGitignoreForTargetsKeepsEmptyAliasScopeAuthoritative(t *testing.T) {
	root := t.TempDir()
	invocationRoot := tspath.NormalizePath(filepath.Join(root, "invocation"))
	physicalRoot := tspath.NormalizePath(filepath.Join(root, "physical"))
	aliasDirectory := tspath.CombinePaths(invocationRoot, "link")
	configDirectory := tspath.NormalizePath(filepath.Join(root, "config"))
	for _, directory := range []string{invocationRoot, physicalRoot, configDirectory} {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(physicalRoot, aliasDirectory); err != nil {
		t.Skipf("symlinks are unavailable: %v", err)
	}
	if err := os.WriteFile(filepath.Join(physicalRoot, ".gitignore"), []byte("ignored.ts\nblocked/\nreopen/*\n!reopen/keep/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{
		tspath.CombinePaths(physicalRoot, "ignored.ts"),
		tspath.CombinePaths(physicalRoot, "blocked", "file.ts"),
		tspath.CombinePaths(physicalRoot, "reopen", "keep", "file.ts"),
	} {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("debugger;\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	effective := ConfigWithGitignoreForTargetsFromRoot(
		RslintConfig{{Rules: Rules{"rule": "error"}}},
		configDirectory,
		invocationRoot,
		osvfs.FS(),
		nil,
		[]string{invocationRoot, physicalRoot},
	)
	resolver := newConfigTargetResolver(effective, configDirectory, osvfs.FS())
	physicalIgnored := tspath.CombinePaths(physicalRoot, "ignored.ts")
	aliasIgnored := tspath.CombinePaths(aliasDirectory, "ignored.ts")
	if !resolver.resolve(physicalIgnored, physicalIgnored).globallyIgnored {
		t.Fatal("physical scope did not apply its own .gitignore")
	}
	if resolver.resolve(aliasIgnored, physicalIgnored).globallyIgnored {
		t.Fatal("empty lexical scope borrowed an aliased physical scope's .gitignore")
	}

	physicalBlocked := tspath.CombinePaths(physicalRoot, "blocked")
	aliasBlocked := tspath.CombinePaths(aliasDirectory, "blocked")
	if !resolver.canPruneDirectory(physicalBlocked, physicalBlocked) {
		t.Fatal("physical scope did not prune its ignored directory")
	}
	if resolver.canPruneDirectory(aliasBlocked, physicalBlocked) {
		t.Fatal("directory pruning crossed from the physical scope into the empty alias scope")
	}
	physicalKeep := tspath.CombinePaths(physicalRoot, "reopen", "keep")
	aliasKeep := tspath.CombinePaths(aliasDirectory, "reopen", "keep")
	if !resolver.reopensDirectory(physicalKeep, physicalKeep) {
		t.Fatal("physical scope did not apply its own Git negation")
	}
	if resolver.reopensDirectory(aliasKeep, physicalKeep) {
		t.Fatal("directory reopening crossed from the physical scope into the empty alias scope")
	}
}

func TestGitignoreCollectionPruningUsesEachConfigEntryAuthoredBase(t *testing.T) {
	root := t.TempDir()
	configDirectory := tspath.NormalizePath(filepath.Join(root, "config"))
	invocationRoot := tspath.NormalizePath(filepath.Join(root, "workspace"))
	blockedDirectory := tspath.CombinePaths(invocationRoot, "blocked")
	for _, directory := range []string{configDirectory, blockedDirectory} {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	blockedGitignore := tspath.CombinePaths(blockedDirectory, ".gitignore")
	if err := os.WriteFile(blockedGitignore, []byte("ignored.ts\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tspath.CombinePaths(blockedDirectory, "ignored.ts"), []byte("debugger;\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	inlineOverride := ConfigWithAuthoredPathBase(
		RslintConfig{{Ignores: []string{"blocked/**"}}},
		invocationRoot,
	)
	entries := append(
		RslintConfig{{Rules: Rules{"no-debugger": "error"}}},
		inlineOverride...,
	)
	spy := &gitignoreReadSpyFS{FS: osvfs.FS()}
	_ = ConfigWithGitignoreForTargetsFromRoot(
		entries,
		configDirectory,
		invocationRoot,
		spy,
		nil,
		[]string{invocationRoot},
	)
	for _, read := range spy.reads {
		if read == blockedGitignore {
			t.Fatal("collection rebased an invocation-authored ignore to the external config directory")
		}
	}
}

func TestCollectedGitignoreScopesStayIndependentOnWindowsAndUNCPaths(t *testing.T) {
	for _, test := range []struct {
		name          string
		firstRoot     string
		secondRoot    string
		firstTarget   string
		secondTarget  string
		caseSensitive bool
	}{
		{
			name:         "Windows drive and case",
			firstRoot:    "C:/Repo/One",
			secondRoot:   "D:/Repo/Two",
			firstTarget:  "c:/repo/one/SHARED.TS",
			secondTarget: "d:/repo/two/SHARED.TS",
		},
		{
			name:          "UNC shares",
			firstRoot:     "//server/share-one/repo",
			secondRoot:    "//server/share-two/repo",
			firstTarget:   "//server/share-one/repo/shared.ts",
			secondTarget:  "//server/share-two/repo/shared.ts",
			caseSensitive: true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			patterns := []gitignore.Pattern{
				{Glob: "shared.ts", NodeGlob: "shared.ts", MatchDirectory: test.firstRoot},
				{Glob: "shared.ts", NodeGlob: "shared.ts", Negated: true, MatchDirectory: test.secondRoot},
			}
			effective := ConfigWithCollectedGitignoreScopes(
				RslintConfig{{Rules: Rules{"rule": "error"}}},
				patterns,
				[]string{test.firstRoot, test.secondRoot},
				"C:/config",
				nil,
				!test.caseSensitive,
			)
			if !effective.IsFileIgnored(test.firstTarget, "C:/config") {
				t.Fatal("first scope lost its positive pattern")
			}
			if effective.IsFileIgnored(test.secondTarget, "C:/config") {
				t.Fatal("a positive pattern crossed into the second scope")
			}
		})
	}
}

func TestConfigWithGitignoreForTargetsUsesExternalScanRoot(t *testing.T) {
	root := t.TempDir()
	configDir := tspath.NormalizePath(filepath.Join(root, "config"))
	scanRoot := tspath.NormalizePath(filepath.Join(root, "workspace"))
	for _, directory := range []string{configDir, scanRoot} {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	visible := tspath.CombinePaths(scanRoot, "visible.js")
	ignored := tspath.CombinePaths(scanRoot, "ignored.js")
	for _, filePath := range []string{visible, ignored} {
		if err := os.WriteFile(filePath, []byte("debugger;\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(scanRoot, ".gitignore"), []byte("ignored.js\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	entries := ConfigWithGitignoreForTargetsFromRoot(
		RslintConfig{{
			Files: []string{"../workspace/*.js"},
			Rules: Rules{"no-debugger": "error"},
		}},
		configDir,
		scanRoot,
		osvfs.FS(),
		[]string{visible, ignored},
		nil,
	)
	resolver := newConfigTargetResolver(entries, configDir, osvfs.FS())
	if decision := resolver.resolve(ignored, ""); !decision.globallyIgnored {
		t.Fatalf("external scan-root Git ignore did not exclude %q: %+v", ignored, decision)
	}
	if decision := resolver.resolve(visible, ""); decision.globallyIgnored || !decision.selected {
		t.Fatalf("external scan-root visible target was not selected: %+v", decision)
	}
}

func TestConfigWithGitignoreForTargetsMatchesPhysicalAndLexicalScanRoots(t *testing.T) {
	configDir := tspath.NormalizePath(t.TempDir())
	physicalScanRoot := tspath.NormalizePath(t.TempDir())
	aliasScanRoot := tspath.CombinePaths(tspath.NormalizePath(t.TempDir()), "workspace-alias")
	if err := os.Symlink(physicalScanRoot, aliasScanRoot); err != nil {
		t.Skipf("directory symlink unavailable: %v", err)
	}
	physicalIgnored := tspath.CombinePaths(physicalScanRoot, "ignored.js")
	aliasIgnored := tspath.CombinePaths(aliasScanRoot, "ignored.js")
	configVisible := tspath.CombinePaths(configDir, "ignored.js")
	for _, filePath := range []string{physicalIgnored, configVisible} {
		if err := os.WriteFile(filePath, []byte("debugger;\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(aliasScanRoot, ".gitignore"), []byte("ignored.js\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	physicalIgnored = tspath.NormalizePath(osvfs.FS().Realpath(physicalIgnored))

	entries := ConfigWithGitignoreForTargetsFromRoot(
		RslintConfig{{Rules: Rules{"no-debugger": "error"}}},
		configDir,
		aliasScanRoot,
		osvfs.FS(),
		[]string{physicalIgnored},
		nil,
	)
	resolver := newConfigTargetResolver(entries, configDir, osvfs.FS())
	for _, filePath := range []string{physicalIgnored, aliasIgnored} {
		if decision := resolver.resolve(filePath, ""); !decision.globallyIgnored {
			t.Fatalf("Git ignore did not cover scan-root identity %q: %+v", filePath, decision)
		}
	}
	if decision := resolver.resolve(configVisible, ""); decision.globallyIgnored {
		t.Fatalf("Git ignore escaped the scan root onto %q: %+v", configVisible, decision)
	}
}

func TestConfigWithGitignoreForTargetsRequestedPhysicalAncestorStartsIndependentScope(t *testing.T) {
	physicalParent := tspath.NormalizePath(t.TempDir())
	physicalWorkspace := tspath.CombinePaths(physicalParent, "workspace")
	if err := os.MkdirAll(physicalWorkspace, 0o755); err != nil {
		t.Fatal(err)
	}
	aliasWorkspace := tspath.CombinePaths(tspath.NormalizePath(t.TempDir()), "workspace-alias")
	if err := os.Symlink(physicalWorkspace, aliasWorkspace); err != nil {
		t.Skipf("directory symlink unavailable: %v", err)
	}
	sibling := tspath.CombinePaths(physicalParent, "sibling.ts")
	if err := os.WriteFile(sibling, []byte("debugger;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		tspath.CombinePaths(physicalParent, ".gitignore"),
		[]byte("sibling.ts\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	effective := ConfigWithGitignoreForTargetsFromRoot(
		RslintConfig{{Rules: Rules{"no-debugger": "error"}}},
		aliasWorkspace,
		aliasWorkspace,
		osvfs.FS(),
		nil,
		[]string{physicalParent},
	)
	resolver := newConfigTargetResolver(effective, aliasWorkspace, osvfs.FS())
	if decision := resolver.resolve(sibling, tspath.NormalizePath(osvfs.FS().Realpath(sibling))); !decision.globallyIgnored {
		t.Fatalf("requested physical ancestor lost its own Git scope: %+v", decision)
	}
}

func TestExternalConfigGitignoreMatchesRawLexicalTargetBeforeOwnerProjection(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	scanRoot := tspath.CombinePaths(root, "workspace")
	configDir := tspath.CombinePaths(root, "physical", "package")
	physicalTargetDir := tspath.CombinePaths(configDir, "src")
	aliasTargetDir := tspath.CombinePaths(scanRoot, "linked-src")
	for _, directory := range []string{scanRoot, physicalTargetDir} {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(physicalTargetDir, aliasTargetDir); err != nil {
		t.Skipf("directory symlink unavailable: %v", err)
	}
	physicalTarget := tspath.CombinePaths(physicalTargetDir, "ignored.js")
	aliasTarget := tspath.CombinePaths(aliasTargetDir, "ignored.js")
	physicalGeneratedDir := tspath.CombinePaths(physicalTargetDir, "generated")
	aliasGeneratedDir := tspath.CombinePaths(aliasTargetDir, "generated")
	physicalReopenedDir := tspath.CombinePaths(physicalTargetDir, "reopened")
	aliasReopenedDir := tspath.CombinePaths(aliasTargetDir, "reopened")
	for _, directory := range []string{physicalGeneratedDir, physicalReopenedDir} {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(physicalTarget, []byte("debugger;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(scanRoot, ".gitignore"),
		[]byte("linked-src/ignored.js\nlinked-src/generated/\n!linked-src/reopened/\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	entries := ConfigWithGitignoreForTargetsFromRoot(
		RslintConfig{{
			Files: []string{"src/**/*.js"},
			Rules: Rules{"no-debugger": "error"},
		}},
		configDir,
		scanRoot,
		osvfs.FS(),
		[]string{aliasTarget},
		[]string{aliasTargetDir},
	)
	resolver := newConfigTargetResolver(entries, configDir, osvfs.FS())
	canonicalTarget := tspath.NormalizePath(osvfs.FS().Realpath(aliasTarget))
	if decision := resolver.resolve(aliasTarget, canonicalTarget); !decision.globallyIgnored {
		t.Fatalf("Git ignore lost the caller-visible symlink path: %+v", decision)
	}
	if decision := resolver.resolve(physicalTarget, physicalTarget); decision.globallyIgnored {
		t.Fatalf("lexical Git pattern leaked onto the physical target spelling: %+v", decision)
	}
	canonicalGeneratedDir := tspath.NormalizePath(osvfs.FS().Realpath(aliasGeneratedDir))
	if !resolver.canPruneDirectory(aliasGeneratedDir, canonicalGeneratedDir) {
		t.Fatal("Git directory pattern lost the caller-visible symlink path while pruning")
	}
	if resolver.canPruneDirectory(physicalGeneratedDir, physicalGeneratedDir) {
		t.Fatal("lexical Git directory pattern leaked onto the physical directory spelling")
	}
	canonicalReopenedDir := tspath.NormalizePath(osvfs.FS().Realpath(aliasReopenedDir))
	if !resolver.reopensDirectory(aliasReopenedDir, canonicalReopenedDir) {
		t.Fatal("Git negation lost the caller-visible symlink path while reopening")
	}
	if resolver.reopensDirectory(physicalReopenedDir, physicalReopenedDir) {
		t.Fatal("lexical Git negation leaked onto the physical directory spelling")
	}
}

func TestExternalGitignoreDistinguishesFileAndDirectorySymlinks(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	gitRoot := tspath.CombinePaths(root, "workspace")
	configDir := tspath.CombinePaths(root, "config")
	outsideDir := tspath.CombinePaths(root, "outside")
	physicalDirectoryTarget := tspath.CombinePaths(gitRoot, "generated")
	for _, directory := range []string{gitRoot, configDir, outsideDir, physicalDirectoryTarget} {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	physicalFileTarget := tspath.CombinePaths(gitRoot, "ignored.js")
	visibleTarget := tspath.CombinePaths(gitRoot, "visible.js")
	physicalDirectoryFile := tspath.CombinePaths(physicalDirectoryTarget, "nested.js")
	for _, filePath := range []string{physicalFileTarget, visibleTarget, physicalDirectoryFile} {
		if err := os.WriteFile(filePath, []byte("debugger;\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(
		filepath.Join(gitRoot, ".gitignore"),
		[]byte("ignored.js\ngenerated/\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	fileAlias := tspath.CombinePaths(outsideDir, "file-alias.js")
	directoryAlias := tspath.CombinePaths(outsideDir, "directory-alias")
	if err := os.Symlink(physicalFileTarget, fileAlias); err != nil {
		t.Skipf("file symlink unavailable: %v", err)
	}
	if err := os.Symlink(physicalDirectoryTarget, directoryAlias); err != nil {
		t.Skipf("directory symlink unavailable: %v", err)
	}
	directoryAliasFile := tspath.CombinePaths(directoryAlias, "nested.js")

	entries := ConfigWithGitignoreForTargetsFromRoot(
		RslintConfig{{Rules: Rules{"no-debugger": "error"}}},
		configDir,
		gitRoot,
		osvfs.FS(),
		[]string{visibleTarget, fileAlias, directoryAliasFile},
		nil,
	)
	resolver := newConfigTargetResolver(entries, configDir, osvfs.FS())
	if decision := resolver.resolve(physicalFileTarget, physicalFileTarget); !decision.globallyIgnored {
		t.Fatalf("Git rule did not ignore its physical in-root target: %+v", decision)
	}
	if decision := resolver.resolve(
		fileAlias,
		tspath.NormalizePath(osvfs.FS().Realpath(fileAlias)),
	); decision.globallyIgnored {
		t.Fatalf("Git rule escaped its root through an external file symlink: %+v", decision)
	}
	if decision := resolver.resolve(
		directoryAliasFile,
		tspath.NormalizePath(osvfs.FS().Realpath(directoryAliasFile)),
	); !decision.globallyIgnored {
		t.Fatalf("Git rule lost a verified directory-alias target: %+v", decision)
	}
	if !resolver.canPruneDirectory(
		directoryAlias,
		tspath.NormalizePath(osvfs.FS().Realpath(directoryAlias)),
	) {
		t.Fatal("Git rule did not prune a verified directory alias into its root")
	}
}
