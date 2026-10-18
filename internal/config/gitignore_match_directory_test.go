package config

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/config/gitignore"
)

type gitignoreReadSpyFS struct {
	vfs.FS
	reads       []string
	directories []string
}

func (fsys *gitignoreReadSpyFS) ReadFile(path string) (string, bool) {
	fsys.reads = append(fsys.reads, tspath.NormalizePath(path))
	return fsys.FS.ReadFile(path)
}

func (fsys *gitignoreReadSpyFS) GetAccessibleEntries(path string) vfs.Entries {
	fsys.directories = append(fsys.directories, tspath.NormalizePath(path))
	return fsys.FS.GetAccessibleEntries(path)
}

func gitPolicyFixture(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for name, content := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return tspath.NormalizePath(osvfs.FS().Realpath(root))
}

func gitPolicyMatcher(t *testing.T, entries RslintConfig, owner string, fsys vfs.FS, policy *GitPolicy) TargetMatcher {
	t.Helper()
	spaces := NewPathSpaceSnapshot(map[string]RslintConfig{owner: entries}, fsys)
	matcher, err := NewTargetMatcherWithPathSpaces(entries, owner, fsys, spaces, policy)
	if err != nil {
		t.Fatal(err)
	}
	return matcher
}

func TestGitPolicyTargetChainUsesNearestRepository(t *testing.T) {
	for _, marker := range []string{"directory", "file", "none"} {
		t.Run(marker, func(t *testing.T) {
			root := gitPolicyFixture(t, map[string]string{
				".gitignore":                 "dist/\n",
				"repo/.gitignore":            "*.generated.ts\n",
				"repo/pkg/.gitignore":        "!keep.generated.ts\n",
				"repo/pkg/dist/app.ts":       "",
				"repo/pkg/drop.generated.ts": "",
				"repo/pkg/keep.generated.ts": "",
			})
			repo := tspath.CombinePaths(root, "repo")
			if marker == "directory" {
				if err := os.Mkdir(filepath.Join(repo, ".git"), 0755); err != nil {
					t.Fatal(err)
				}
			}
			if marker == "file" {
				if err := os.WriteFile(filepath.Join(repo, ".git"), []byte("gitdir: elsewhere\n"), 0644); err != nil {
					t.Fatal(err)
				}
			}
			entries := RslintConfig{{Files: []string{"**/*.ts"}}}
			// The config's unrelated directory does not select a Git upper bound.
			owner := tspath.CombinePaths(root, "external-config")
			paths := []string{"pkg/dist/app.ts", "pkg/drop.generated.ts", "pkg/keep.generated.ts"}
			identities := make([]PathIdentity, len(paths))
			for index, path := range paths {
				identities[index] = FreezeFileIdentity(tspath.CombinePaths(repo, path), osvfs.FS())
			}
			policy := CollectGitPolicy(entries, owner, osvfs.FS(), identities, nil)
			matcher := gitPolicyMatcher(t, entries, owner, osvfs.FS(), policy)
			for index, want := range []bool{marker == "none", true, false} {
				if got := matcher.MatchFile(identities[index]).GloballyIgnored(); got != want {
					t.Errorf("%s ignored=%v want=%v", paths[index], got, want)
				}
			}
			if !policy.CoversDirectory(tspath.CombinePaths(repo, "pkg")) {
				t.Fatal("exact target directory is not covered")
			}
			if policy.CoversDirectory(tspath.CombinePaths(repo, "unvisited")) {
				t.Fatal("root presence covered an unvisited directory")
			}
		})
	}
}

func TestGitPolicyNestedEmptyRepositoryCutsAncestorRules(t *testing.T) {
	root := gitPolicyFixture(t, map[string]string{
		".git": "", ".gitignore": "vendor/\n",
		"vendor/inner/.git":      "gitdir: elsewhere\n",
		"vendor/inner/source.ts": "",
	})
	inner := tspath.CombinePaths(root, "vendor/inner")
	entries := RslintConfig{{Files: []string{"**/*.ts"}}}
	dirs := ResolveDirectoryIdentities([]string{root, inner}, osvfs.FS())
	policy := CollectGitPolicy(entries, root, osvfs.FS(), nil, dirs)
	matcher := gitPolicyMatcher(t, entries, root, osvfs.FS(), policy)
	target := FreezeFileIdentity(tspath.CombinePaths(inner, "source.ts"), osvfs.FS())
	if matcher.MatchFile(target).GloballyIgnored() {
		t.Fatal("empty inner Git root borrowed outer ignore rules")
	}
	if !policy.CoversDirectory(inner) {
		t.Fatal("explicit inner root was lost under pruned ancestor")
	}
	// Merely knowing an explicit inner root never opens sibling traversal.
	if policy.CoversDirectory(tspath.CombinePaths(root, "vendor/sibling")) {
		t.Fatal("explicit inner target reopened unrelated siblings")
	}
}

func TestGitPolicyCombinesSourceStateBeforeAuthoredNegations(t *testing.T) {
	root := gitPolicyFixture(t, map[string]string{
		".git": "", ".gitignore": "a-dir/\ndist/\n",
		"foo/.gitignore":    "!a-dir/keep.ts\n",
		"foo/a-dir/keep.ts": "", "dist/.gitignore": "source.ts\n", "dist/source.ts": "",
	})
	source := FreezeFileIdentity(tspath.CombinePaths(root, "foo/a-dir/keep.ts"), osvfs.FS())
	entries := RslintConfig{{Files: []string{"**/*.ts"}}}
	policy := CollectGitPolicy(entries, root, osvfs.FS(), []PathIdentity{source}, nil)
	if !gitPolicyMatcher(t, entries, root, osvfs.FS(), policy).MatchFile(source).GloballyIgnored() {
		t.Fatal("child source negation restored an excluded parent")
	}
	entries = append(RslintConfig{{Ignores: []string{"!dist/**"}}}, entries...)
	spy := &gitignoreReadSpyFS{FS: osvfs.FS()}
	policy = CollectGitPolicy(entries, root, spy, nil, ResolveDirectoryIdentities([]string{root}, spy))
	matcher := gitPolicyMatcher(t, entries, root, spy, policy)
	dist := FreezeFileIdentity(tspath.CombinePaths(root, "dist/source.ts"), spy)
	if matcher.MatchFile(dist).GloballyIgnored() {
		t.Fatal("authored negation did not override Git decision")
	}
	for _, path := range spy.reads {
		if path == tspath.CombinePaths(root, "dist/.gitignore") {
			t.Fatal("authored negation made raw Git source reachable")
		}
	}
	if !policy.CoversDirectory(tspath.CombinePaths(root, "dist")) {
		t.Fatal("proven raw block was not covered")
	}
	if matcher.CanPruneDirectory(DirectoryIdentity{LexicalPath: tspath.CombinePaths(root, "dist"), CanonicalPath: tspath.CombinePaths(root, "dist")}) {
		t.Fatal("authored negation lost directory reach")
	}
}

func TestGitPolicyDirectoryPruningUsesGitAndAuthoredNegationReach(t *testing.T) {
	root := gitPolicyFixture(t, map[string]string{".git": "", ".gitignore": "build/\n", "build/keep.ts": "", "build/drop.ts": ""})
	for _, test := range []struct {
		ignore []string
		prune  bool
	}{
		{nil, true}, {[]string{"!build/keep.ts"}, false}, {[]string{"!other/keep.ts"}, true}, {[]string{"build/**", "!build/keep.ts"}, true},
	} {
		entries := RslintConfig{{Ignores: test.ignore}, {Files: []string{"**/*.ts"}}}
		policy := CollectGitPolicy(entries, root, osvfs.FS(), []PathIdentity{FreezeFileIdentity(tspath.CombinePaths(root, "build/keep.ts"), osvfs.FS())}, nil)
		matcher := gitPolicyMatcher(t, entries, root, osvfs.FS(), policy)
		path := tspath.CombinePaths(root, "build")
		if got := matcher.CanPruneDirectory(DirectoryIdentity{LexicalPath: path, CanonicalPath: path}); got != test.prune {
			t.Errorf("%v prune=%v want=%v", test.ignore, got, test.prune)
		}
	}
}

func TestGitPolicyUsesPhysicalParentAndOriginalLeafName(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink privileges differ")
	}
	root := gitPolicyFixture(t, map[string]string{
		"repo-a/.git": "", "repo-a/.gitignore": "src/alias.ts\n",
		"repo-a/src/source.ts": "", "repo-b/.git": "", "repo-b/source.ts": "",
	})
	a := tspath.CombinePaths(root, "repo-a")
	b := tspath.CombinePaths(root, "repo-b")
	alias := tspath.CombinePaths(root, "directory-alias")
	if err := os.Symlink(filepath.Join(a, "src"), alias); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(b, "source.ts"), filepath.Join(a, "src/alias.ts")); err != nil {
		t.Fatal(err)
	}
	identity := FreezeFileIdentity(tspath.CombinePaths(alias, "alias.ts"), osvfs.FS())
	if want := tspath.CombinePaths(a, "src/alias.ts"); GitPath(identity) != want {
		t.Fatalf("Git path=%s want=%s", GitPath(identity), want)
	}
	entries := RslintConfig{{Files: []string{"**/*.ts"}}}
	policy := CollectGitPolicy(entries, b, osvfs.FS(), []PathIdentity{identity}, nil)
	if !gitPolicyMatcher(t, entries, b, osvfs.FS(), policy).MatchFile(identity).GloballyIgnored() {
		t.Fatal("leaf symlink borrowed destination repository rules")
	}
}

func TestGitPolicySnapshotAndInputConfigurationAreImmutable(t *testing.T) {
	root := gitPolicyFixture(t, map[string]string{".git": "", ".gitignore": "dist/\n", "dist/file.ts": "", "src/file.ts": ""})
	reader := gitignore.NewReader(osvfs.FS())
	reader.Observe(reader.Ancestors(root), root)
	snapshot := reader.Snapshot()
	policy := NewGitPolicy(snapshot)
	for index := range snapshot.Sources {
		for p := range snapshot.Sources[index].Patterns {
			snapshot.Sources[index].Patterns[p].NodeGlob = "does-not-match"
		}
	}
	snapshot.Roots[0] = "/elsewhere"
	reader.Observe(reader.Ancestors(tspath.CombinePaths(root, "src")), tspath.CombinePaths(root, "src"))
	if policy.CoversDirectory(tspath.CombinePaths(root, "src")) {
		t.Fatal("new observation mutated prior generation")
	}
	entries := RslintConfig{{Files: []string{"**/*.ts"}}}
	before := append(RslintConfig(nil), entries...)
	target := FreezeFileIdentity(tspath.CombinePaths(root, "dist/file.ts"), osvfs.FS())
	if !gitPolicyMatcher(t, entries, root, osvfs.FS(), policy).MatchFile(target).GloballyIgnored() {
		t.Fatal("snapshot mutation changed compiled Git rules")
	}
	if !reflect.DeepEqual(entries, before) || entries.IsFileIgnored(target.Path, root) {
		t.Fatal("Git collection leaked into authored configuration")
	}
}

func TestTargetMatchRequiresSameGitPolicyGeneration(t *testing.T) {
	root := gitPolicyFixture(t, map[string]string{".git": "", ".gitignore": "dist/\n", "dist/file.ts": ""})
	entries := RslintConfig{{Files: []string{"**/*.ts"}}}
	identity := FreezeFileIdentity(tspath.CombinePaths(root, "dist/file.ts"), osvfs.FS())
	first := CollectGitPolicy(entries, root, osvfs.FS(), []PathIdentity{identity}, nil)
	second := CollectGitPolicy(entries, root, osvfs.FS(), []PathIdentity{identity}, nil)
	spaces := NewPathSpaceSnapshot(map[string]RslintConfig{root: entries}, osvfs.FS())
	matcher, err := NewTargetMatcherWithPathSpaces(entries, root, osvfs.FS(), spaces, first)
	if err != nil {
		t.Fatal(err)
	}
	match := matcher.MatchFile(identity)
	for _, test := range []struct {
		policy *GitPolicy
		reuse  bool
	}{{first, true}, {second, false}, {nil, false}} {
		resolver, err := NewFileConfigResolverWithPathSpaces(entries, root, osvfs.FS(), spaces, baseRuleCatalog(), test.policy)
		if err != nil {
			t.Fatal(err)
		}
		if _, reuse := resolver.reuseTargetMatch(identity, &match); reuse != test.reuse {
			t.Errorf("reuse=%v want=%v", reuse, test.reuse)
		}
	}
}

func TestExplicitDirectoryIdentitiesKeepPrunedNestedRoots(t *testing.T) {
	fsys := osvfs.FS()
	root := tspath.NormalizePath(t.TempDir())
	child := tspath.CombinePaths(root, "ignored/inner")
	identities := ResolveDirectoryIdentities([]string{root, child}, fsys)
	if len(identities) != 2 || !strings.HasSuffix(identities[1].LexicalPath, "/ignored/inner") {
		t.Fatalf("explicit roots=%v", identities)
	}
}

func TestGitPolicyAuthoredLeafNegationFindsVisitedNestedRepository(t *testing.T) {
	root := gitPolicyFixture(t, map[string]string{
		".git": "", ".gitignore": "dist/\n", "dist/.gitignore": "*.ts\n",
		"dist/inner/.git": "", "dist/inner/.gitignore": "*.generated.ts\n",
		"dist/inner/source.ts": "", "dist/inner/output.generated.ts": "",
		"dist/unrelated/.gitignore": "*.ts\n", "dist/unrelated/source.ts": "",
	})
	entries := RslintConfig{{Ignores: []string{"!dist/inner/source.ts"}}, {Files: []string{"**/*.ts"}}}
	spy := &gitignoreReadSpyFS{FS: osvfs.FS()}
	policy := CollectGitPolicy(entries, root, spy, nil, ResolveDirectoryIdentities([]string{root}, spy))
	matcher := gitPolicyMatcher(t, entries, root, spy, policy)
	if !policy.CoversDirectory(tspath.CombinePaths(root, "dist/inner")) {
		t.Fatal("target-visible nested repository was not observed")
	}
	for _, test := range []struct {
		path    string
		ignored bool
	}{{"dist/inner/source.ts", false}, {"dist/inner/output.generated.ts", true}, {"dist/unrelated/source.ts", true}} {
		target := FreezeFileIdentity(tspath.CombinePaths(root, test.path), spy)
		if got := matcher.MatchFile(target).GloballyIgnored(); got != test.ignored {
			t.Errorf("%s ignored=%v want=%v", test.path, got, test.ignored)
		}
	}
	for _, path := range spy.reads {
		if path == tspath.CombinePaths(root, "dist/.gitignore") || path == tspath.CombinePaths(root, "dist/unrelated/.gitignore") {
			t.Fatalf("hidden source read: %s", path)
		}
	}
}

func TestGitPolicyAuthoredBasesPruneSourcesAndRespectAbsoluteBlocks(t *testing.T) {
	root := gitPolicyFixture(t, map[string]string{
		".git": "", ".gitignore": "*.generated.ts\n",
		"packages/blocked/.gitignore": "*.ts\n", "packages/blocked/file.ts": "",
		"packages/visible/.gitignore": "*.log\n", "packages/visible/file.ts": "",
		"node_modules/.gitignore": "*.ts\n", "node_modules/file.ts": "",
	})
	entries := append(ConfigWithAuthoredPathBase(RslintConfig{{Ignores: []string{"blocked/**", "!blocked/file.ts"}}}, tspath.CombinePaths(root, "packages")), ConfigEntry{Files: []string{"**/*.ts"}})
	spy := &gitignoreReadSpyFS{FS: osvfs.FS()}
	policy := CollectGitPolicy(entries, root, spy, nil, ResolveDirectoryIdentities([]string{root}, spy))
	for _, path := range spy.reads {
		if path == tspath.CombinePaths(root, "packages/blocked/.gitignore") || path == tspath.CombinePaths(root, "node_modules/.gitignore") {
			t.Fatalf("pruned source read: %s", path)
		}
	}
	if !slices.Contains(spy.reads, tspath.CombinePaths(root, "packages/visible/.gitignore")) {
		t.Fatal("sibling source was pruned")
	}
	target := FreezeFileIdentity(tspath.CombinePaths(root, "packages/blocked/file.ts"), spy)
	if !gitPolicyMatcher(t, entries, root, spy, policy).MatchFile(target).GloballyIgnored() {
		t.Fatal("Git/author negation overrode authored absolute directory block")
	}
}

func TestGitPolicyGitNegationCannotOverrideAuthoredIgnore(t *testing.T) {
	root := gitPolicyFixture(t, map[string]string{".git": "", ".gitignore": "*.ts\n!keep.ts\n", "keep.ts": ""})
	entries := RslintConfig{{Ignores: []string{"keep.ts"}}, {Files: []string{"**/*.ts"}}}
	identity := FreezeFileIdentity(tspath.CombinePaths(root, "keep.ts"), osvfs.FS())
	for _, policy := range []*GitPolicy{
		CollectGitPolicy(entries, root, osvfs.FS(), []PathIdentity{identity}, nil),
		CollectGitPolicy(entries, root, osvfs.FS(), nil, ResolveDirectoryIdentities([]string{root}, osvfs.FS())),
	} {
		if !gitPolicyMatcher(t, entries, root, osvfs.FS(), policy).MatchFile(identity).GloballyIgnored() {
			t.Fatal("Git negation overrode authored ignore")
		}
	}
}

func TestGitPolicyRecursiveCollectionDoesNotFollowDirectorySymlinks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink privileges differ")
	}
	root := gitPolicyFixture(t, map[string]string{".git": "", "src/.gitignore": "*.generated.ts\n", "src/source.ts": "", "outside/.gitignore": "*.ts\n"})
	if err := os.Symlink(filepath.Join(root, "src"), filepath.Join(root, "src/cycle")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, "outside"), filepath.Join(root, "src/external")); err != nil {
		t.Fatal(err)
	}
	spy := &gitignoreReadSpyFS{FS: osvfs.FS()}
	policy := CollectGitPolicy(nil, root, spy, nil, ResolveDirectoryIdentities([]string{tspath.CombinePaths(root, "src")}, spy))
	if !policy.CoversDirectory(tspath.CombinePaths(root, "src")) {
		t.Fatal("explicit directory root not covered")
	}
	for _, path := range spy.reads {
		if strings.Contains(path, "/cycle/") || strings.Contains(path, "/external/") || path == tspath.CombinePaths(root, "outside/.gitignore") {
			t.Fatalf("followed descendant symlink: %s", path)
		}
	}
}

func TestGitPolicyOnlyWalksSuppliedDirectoryRootsAndExactChains(t *testing.T) {
	root := gitPolicyFixture(t, map[string]string{
		".git": "", ".gitignore": "*.generated.ts\n",
		"target/deep/.gitignore": "local.ts\n", "target/deep/local.ts": "",
		"exact/deep/.gitignore": "*.ts\n", "exact/deep/file.ts": "",
		"unrequested/.gitignore": "*.ts\n", "unrequested/file.ts": "",
	})
	spy := &gitignoreReadSpyFS{FS: osvfs.FS()}
	files := ResolveFileIdentities([]string{tspath.CombinePaths(root, "exact/deep/file.ts")}, spy)
	directories := ResolveDirectoryIdentities([]string{tspath.CombinePaths(root, "target")}, spy)
	CollectGitPolicy(nil, root, spy, files, directories)
	for _, path := range spy.directories {
		if _, within := RelativePathWithinConfigRoot(path, tspath.CombinePaths(root, "target"), true); !within {
			t.Fatalf("walked outside the supplied recursive root: %s", path)
		}
	}
	if !slices.Contains(spy.reads, tspath.CombinePaths(root, "exact/deep/.gitignore")) {
		t.Fatal("exact target ancestor source was not observed")
	}
	if slices.Contains(spy.reads, tspath.CombinePaths(root, "unrequested/.gitignore")) {
		t.Fatal("read an unrelated subtree")
	}
}

type gitPolicyNoRealpathFS struct{ vfs.FS }

func (fsys gitPolicyNoRealpathFS) Realpath(string) string { return "" }
func (fsys gitPolicyNoRealpathFS) GetAccessibleEntries(path string) vfs.Entries {
	entries := fsys.FS.GetAccessibleEntries(path)
	entries.Symlinks = nil
	return entries
}

func TestGitPolicyVFSWithoutRealpathUsesLexicalDirectoryIdentity(t *testing.T) {
	root := gitPolicyFixture(t, map[string]string{".git": "", "pkg/.gitignore": "*.ts\n", "pkg/file.ts": ""})
	fsys := gitPolicyNoRealpathFS{FS: osvfs.FS()}
	entries := RslintConfig{{Files: []string{"**/*.ts"}}}
	policy := CollectGitPolicy(entries, root, fsys, nil, ResolveDirectoryIdentities([]string{root}, fsys))
	target := FreezeFileIdentity(tspath.CombinePaths(root, "pkg/file.ts"), fsys)
	if !gitPolicyMatcher(t, entries, root, fsys, policy).MatchFile(target).GloballyIgnored() {
		t.Fatal("unknown realpath was treated as a directory symlink")
	}
}

type gitPolicyVolumeRootFS struct {
	vfs.FS
	root string
}

func (fsys gitPolicyVolumeRootFS) UseCaseSensitiveFileNames() bool { return false }
func (fsys gitPolicyVolumeRootFS) DirectoryExists(string) bool     { return false }
func (fsys gitPolicyVolumeRootFS) FileExists(string) bool          { return false }
func (fsys gitPolicyVolumeRootFS) Realpath(path string) string     { return path }
func (fsys gitPolicyVolumeRootFS) ReadFile(path string) (string, bool) {
	if path == tspath.CombinePaths(fsys.root, ".gitignore") {
		return "dist/\n", true
	}
	return "", false
}
func TestGitPolicyVolumeRootRulesApplyToFilesAndDirectories(t *testing.T) {
	for _, root := range []string{"C:/", "//server/share/", "//server/share"} {
		t.Run(root, func(t *testing.T) {
			fsys := gitPolicyVolumeRootFS{root: root}
			file := tspath.CombinePaths(root, "pkg/dist/input.ts")
			identity := PathIdentity{Path: file, CanonicalPath: file, CanonicalParentPath: tspath.GetDirectoryPath(file)}
			entries := RslintConfig{{Files: []string{"**/*.ts"}}}
			policy := CollectGitPolicy(entries, "/config", fsys, []PathIdentity{identity}, nil)
			matcher := gitPolicyMatcher(t, entries, "/config", fsys, policy)
			if !matcher.MatchFile(identity).GloballyIgnored() {
				t.Fatalf("volume-root .gitignore was lost for %s", file)
			}
			if !matcher.CanPruneDirectory(DirectoryIdentity{LexicalPath: identity.CanonicalParentPath, CanonicalPath: identity.CanonicalParentPath}) {
				t.Fatal("volume-root directory ignore was lost")
			}
			if !policy.CoversDirectory(identity.CanonicalParentPath) {
				t.Fatal("exact directory coverage was lost")
			}
		})
	}
}
