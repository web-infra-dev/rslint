package loader

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"testing"

	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
)

func TestSelectProjectsPrefetchPreservesPhysicalRootOrder(t *testing.T) {
	dir := tspath.NormalizePath(t.TempDir())
	physicalPath := tspath.ResolvePath(dir, "physical.ts")
	aliasPath := tspath.ResolvePath(dir, "alias.ts")
	writeProgramTestFiles(t, dir, map[string]string{
		"physical.ts":         `export const value = 1;`,
		"tsconfig-alias.json": `{"files":["alias.ts"],"compilerOptions":{"noLib":true}}`,
		"tsconfig-exact.json": `{"files":["physical.ts"],"compilerOptions":{"noLib":true}}`,
	})
	if err := os.Symlink(physicalPath, aliasPath); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	baseFS := cachedvfs.From(osvfs.FS())
	fsys := bundled.WrapFS(baseFS)
	canonicalPath := tspath.NormalizePath(fsys.Realpath(physicalPath))
	if aliasPath == physicalPath || canonicalPath == "" ||
		tspath.NormalizePath(fsys.Realpath(aliasPath)) != canonicalPath {
		t.Fatalf("fixture must expose distinct lexical paths for one physical file")
	}
	config := rslintconfig.RslintConfig{lintProjectEntry(
		[]string{"**/*.ts"},
		"./tsconfig-alias.json",
		"./tsconfig-exact.json",
	)}
	resolver, err := rslintconfig.NewProjectPathResolver(nil, config, dir, fsys, false)
	if err != nil {
		t.Fatal(err)
	}
	lintPlan, err := resolver.ResolveLintProjectPlan(rslintconfig.LintTargetPlan{Targets: []rslintconfig.DiscoveredLintTarget{{
		Path:            physicalPath,
		CanonicalPath:   canonicalPath,
		ConfigDirectory: dir,
	}}})
	if err != nil {
		t.Fatal(err)
	}
	session := NewSession(fsys)
	set, err := session.SelectProjects(lintPlan, nil, true, true)
	if err != nil {
		t.Fatal(err)
	}
	if set.Len() != 1 {
		t.Fatalf("selected Programs = %d, want one", set.Len())
	}
	wantConfig := tspath.ResolvePath(dir, "tsconfig-alias.json")
	if got := tspath.NormalizePath(set.compilerPrograms[0].Options().ConfigFilePath); got != wantConfig {
		t.Fatalf("prefetched physical owner = %q, want earlier alias %q", got, wantConfig)
	}
	loaded, err := session.LoadAPI(set, lintPlan, dir, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.TargetsByProgram) != 1 || len(loaded.TargetsByProgram[0]) != 1 {
		t.Fatalf("physical target projection = %v", loaded.TargetsByProgram)
	}
}

func TestSelectProjectsPrefetchResolvesMissingTargetCanonicalPath(t *testing.T) {
	dir := tspath.NormalizePath(t.TempDir())
	physicalPath := tspath.ResolvePath(dir, "physical.ts")
	aliasPath := tspath.ResolvePath(dir, "alias.ts")
	writeProgramTestFiles(t, dir, map[string]string{
		"physical.ts":   `export const value = 1;`,
		"tsconfig.json": `{"files":["alias.ts"],"compilerOptions":{"noLib":true}}`,
	})
	if err := os.Symlink(physicalPath, aliasPath); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	if tspath.NormalizePath(fsys.Realpath(aliasPath)) != tspath.NormalizePath(fsys.Realpath(physicalPath)) {
		t.Fatal("fixture aliases do not resolve to one physical file")
	}
	lintPlan := rslintconfig.LintProjectPlan{Targets: []rslintconfig.PlannedLintTarget{{
		Target: rslintconfig.DiscoveredLintTarget{
			Path:            physicalPath,
			ConfigDirectory: dir,
		},
		ProjectPaths: []string{tspath.ResolvePath(dir, "tsconfig.json")},
	}}}

	session := NewSession(fsys)
	set, err := session.SelectProjects(lintPlan, nil, true, true)
	if err != nil {
		t.Fatal(err)
	}
	if set.Len() != 1 || set.targetBinding == nil ||
		!slices.Equal(set.targetBinding.owners, []int{0}) {
		t.Fatalf("missing-canonical alias binding: set=%d binding=%+v", set.Len(), set.targetBinding)
	}
	loaded, err := session.LoadAPI(set, lintPlan, dir, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.TargetsByProgram) != 1 ||
		!slices.Equal(loaded.TargetsByProgram[0], []string{aliasPath}) {
		t.Fatalf("missing-canonical alias projection = %v", loaded.TargetsByProgram)
	}
	if got := loaded.TargetPathBySourcePath[aliasPath]; got != physicalPath {
		t.Fatalf("missing-canonical caller path = %q, want %q", got, physicalPath)
	}
}

func TestSelectProjectsPrefetchKeepsDistinctCanonicalCaseIdentities(t *testing.T) {
	const (
		dir        = "/repo"
		upper      = "/repo/Source.ts"
		lower      = "/repo/source.ts"
		configPath = "/repo/tsconfig.json"
	)
	fsys := &exactCaseProgramFS{
		FS: osvfs.FS(),
		files: map[string]string{
			upper:      "export const upper = 1;\n",
			lower:      "export const lower = 2;\n",
			configPath: `{"files":["Source.ts"],"compilerOptions":{"noLib":true}}`,
		},
	}
	config := rslintconfig.RslintConfig{lintProjectEntry([]string{"**/*.ts"}, "./tsconfig.json")}
	resolver, err := rslintconfig.NewProjectPathResolver(nil, config, dir, fsys, false)
	if err != nil {
		t.Fatal(err)
	}
	lintPlan, err := resolver.ResolveLintProjectPlan(rslintconfig.LintTargetPlan{Targets: []rslintconfig.DiscoveredLintTarget{{
		Path:            lower,
		CanonicalPath:   lower,
		ConfigDirectory: dir,
	}}})
	if err != nil {
		t.Fatal(err)
	}
	session := NewSession(fsys)
	set, err := session.SelectProjects(lintPlan, nil, true, true)
	if err != nil {
		t.Fatal(err)
	}
	if set.Len() != 0 || set.targetBinding == nil || !slices.Equal(set.targetBinding.owners, []int{-1}) {
		t.Fatalf("case-distinct source was claimed by prefetched project: set=%d binding=%+v", set.Len(), set.targetBinding)
	}
	loaded, err := session.LoadAPI(set, lintPlan, dir, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Programs) != 1 || len(loaded.TargetsByProgram) != 1 ||
		!slices.Equal(loaded.TargetsByProgram[0], []string{lower}) {
		t.Fatalf("case-distinct target did not remain source-only: %v", loaded.TargetsByProgram)
	}
}

func TestSelectProjectsDirectHintIsBoundedAndCannotChooseOwner(t *testing.T) {
	previousGOMAXPROCS := runtime.GOMAXPROCS(2)
	t.Cleanup(func() { runtime.GOMAXPROCS(previousGOMAXPROCS) })
	dir := tspath.NormalizePath(t.TempDir())
	targetPath := tspath.ResolvePath(dir, "nested/target.ts")
	laterOnlyPath := tspath.ResolvePath(dir, "nested/later-only.ts")
	unrelatedPath := tspath.ResolvePath(dir, "unrelated.ts")
	writeProgramTestFiles(t, dir, map[string]string{
		"nested/target.ts":     `export const target = 1;`,
		"nested/later-only.ts": `export const later = 1;`,
		"unrelated.ts":         `export const unrelated = 1;`,
		"tsconfig-first.json":  `{"files":["nested/target.ts"],"compilerOptions":{"noLib":true}}`,
		"nested/tsconfig.json": `{"files":["target.ts","later-only.ts"],"compilerOptions":{"noLib":true}}`,
		"tsconfig-unused.json": `{"files":["unrelated.ts"],"compilerOptions":{"noLib":true}}`,
	})
	config := rslintconfig.RslintConfig{lintProjectEntry(
		[]string{"**/*.ts"},
		"./tsconfig-first.json",
		"./nested/tsconfig.json",
		"./tsconfig-unused.json",
	)}

	for _, test := range []struct {
		name           string
		prefetch       bool
		singleThreaded bool
		wantHintRead   bool
	}{
		{name: "demand parallel", wantHintRead: true},
		{name: "directory prefetch parallel", prefetch: true, wantHintRead: true},
		{name: "directory prefetch single threaded", prefetch: true, singleThreaded: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			counting := &programReadCountingFS{
				FS:    bundled.WrapFS(cachedvfs.From(osvfs.FS())),
				reads: make(map[string]int),
			}
			fsys := vfs.FS(counting)
			resolver, err := rslintconfig.NewProjectPathResolver(nil, config, dir, fsys, false)
			if err != nil {
				t.Fatal(err)
			}
			lintPlan, err := resolver.ResolveLintProjectPlan(rslintconfig.LintTargetPlan{Targets: []rslintconfig.DiscoveredLintTarget{
				testLintTarget(fsys, dir, targetPath),
			}})
			if err != nil {
				t.Fatal(err)
			}
			set, err := NewSession(fsys).SelectProjects(lintPlan, nil, test.prefetch, test.singleThreaded)
			if err != nil {
				t.Fatal(err)
			}
			if set.Len() != 1 {
				t.Fatalf("selected Programs = %d, want one", set.Len())
			}
			wantConfig := tspath.ResolvePath(dir, "tsconfig-first.json")
			if got := tspath.NormalizePath(set.compilerPrograms[0].Options().ConfigFilePath); got != wantConfig {
				t.Fatalf("hint chose owner %q, want earlier declaration %q", got, wantConfig)
			}
			if got := counting.readCount(laterOnlyPath); (got > 0) != test.wantHintRead {
				t.Fatalf("hint-only source reads = %d, want hint read %t", got, test.wantHintRead)
			}
			if got := counting.readCount(unrelatedPath); got != 0 {
				t.Fatalf("complete direct hint prefetched unrelated source %d time(s)", got)
			}
		})
	}
}

func TestSelectProjectsUsesFullPrefetchForDistinctDirectHints(t *testing.T) {
	previousGOMAXPROCS := runtime.GOMAXPROCS(2)
	t.Cleanup(func() { runtime.GOMAXPROCS(previousGOMAXPROCS) })
	dir := tspath.NormalizePath(t.TempDir())
	aTarget := tspath.ResolvePath(dir, "a/target.ts")
	bTarget := tspath.ResolvePath(dir, "b/target.ts")
	aOnly := tspath.ResolvePath(dir, "a/hint-only.ts")
	bOnly := tspath.ResolvePath(dir, "b/hint-only.ts")
	writeProgramTestFiles(t, dir, map[string]string{
		"a/target.ts":        `export const a = 1;`,
		"b/target.ts":        `export const b = 1;`,
		"a/hint-only.ts":     `export const ah = 1;`,
		"b/hint-only.ts":     `export const bh = 1;`,
		"tsconfig-root.json": `{"files":["a/target.ts","b/target.ts"],"compilerOptions":{"noLib":true}}`,
		"a/tsconfig.json":    `{"files":["target.ts","hint-only.ts"],"compilerOptions":{"noLib":true}}`,
		"b/tsconfig.json":    `{"files":["target.ts","hint-only.ts"],"compilerOptions":{"noLib":true}}`,
	})
	config := rslintconfig.RslintConfig{lintProjectEntry(
		[]string{"**/*.ts"},
		"./tsconfig-root.json",
		"./a/tsconfig.json",
		"./b/tsconfig.json",
	)}
	for _, test := range []struct {
		name         string
		prefetch     bool
		wantHintRead bool
	}{
		{name: "demand"},
		{name: "directory prefetch", prefetch: true, wantHintRead: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			counting := &programReadCountingFS{
				FS:    bundled.WrapFS(cachedvfs.From(osvfs.FS())),
				reads: make(map[string]int),
			}
			fsys := vfs.FS(counting)
			resolver, err := rslintconfig.NewProjectPathResolver(nil, config, dir, fsys, false)
			if err != nil {
				t.Fatal(err)
			}
			lintPlan, err := resolver.ResolveLintProjectPlan(rslintconfig.LintTargetPlan{Targets: []rslintconfig.DiscoveredLintTarget{
				testLintTarget(fsys, dir, aTarget),
				testLintTarget(fsys, dir, bTarget),
			}})
			if err != nil {
				t.Fatal(err)
			}
			set, err := NewSession(fsys).SelectProjects(lintPlan, nil, test.prefetch, false)
			if err != nil {
				t.Fatal(err)
			}
			if set.Len() != 1 {
				t.Fatalf("selected Programs = %d, want one umbrella owner", set.Len())
			}
			wantConfig := tspath.ResolvePath(dir, "tsconfig-root.json")
			if got := tspath.NormalizePath(set.compilerPrograms[0].Options().ConfigFilePath); got != wantConfig {
				t.Fatalf("selected config = %q, want umbrella %q", got, wantConfig)
			}
			if set.targetBinding == nil || !slices.Equal(set.targetBinding.owners, []int{0, 0}) {
				t.Fatalf("umbrella binding = %+v, want owners [0 0]", set.targetBinding)
			}
			for _, source := range []string{aOnly, bOnly} {
				if got := counting.readCount(source); (got > 0) != test.wantHintRead {
					t.Fatalf("distinct hint source %q reads = %d, want reads %t", source, got, test.wantHintRead)
				}
			}
		})
	}
}

func lintProjectEntry(files []string, projects ...string) rslintconfig.ConfigEntry {
	if projects == nil {
		projects = []string{}
	}
	return rslintconfig.ConfigEntry{
		Files: files,
		LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{
			Project: projects,
		}},
	}
}

func resolveEffectivePipelinePlan(
	t *testing.T,
	config rslintconfig.RslintConfig,
	dir string,
	fsys *cachedvfs.FS,
	targets ...string,
) (*rslintconfig.ProjectPathResolver, rslintconfig.LintProjectPlan) {
	t.Helper()
	resolver, err := rslintconfig.NewProjectPathResolver(nil, config, dir, bundled.WrapFS(fsys), false)
	if err != nil {
		t.Fatal(err)
	}
	targetPlan := rslintconfig.LintTargetPlan{Targets: make([]rslintconfig.DiscoveredLintTarget, len(targets))}
	for index, target := range targets {
		target = tspath.NormalizePath(target)
		targetPlan.Targets[index] = rslintconfig.DiscoveredLintTarget{
			Path: target, CanonicalPath: target, ConfigDirectory: dir,
		}
	}
	lintPlan, err := resolver.ResolveLintProjectPlan(targetPlan)
	if err != nil {
		t.Fatal(err)
	}
	return resolver, lintPlan
}

func TestSelectProjectsKeepsTypeCheckCatalogOutOfLintOwnership(t *testing.T) {
	dir := tspath.NormalizePath(t.TempDir())
	writeProgramTestFiles(t, dir, map[string]string{
		"a.ts":              `export const a = 1;`,
		"catalog-main.ts":   `import "./a";`,
		"tsconfig-a.json":   `{"files":["a.ts"],"compilerOptions":{"noLib":true}}`,
		"tsconfig-cat.json": `{"files":["catalog-main.ts"],"compilerOptions":{"noLib":true}}`,
	})
	config := rslintconfig.RslintConfig{
		lintProjectEntry([]string{"catalog/**"}, "./tsconfig-cat.json"),
		lintProjectEntry([]string{"a.ts"}, "./tsconfig-a.json"),
	}
	baseFS := cachedvfs.From(osvfs.FS())
	resolver, lintPlan := resolveEffectivePipelinePlan(t, config, dir, baseFS, filepath.Join(dir, "a.ts"))
	session := NewSession(bundled.WrapFS(baseFS))
	set, err := session.SelectProjects(lintPlan, resolver.CatalogProjectPaths(), true, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(set.TypeCheckPrograms()) != 2 || set.Len() != 2 {
		t.Fatalf("program roles = all:%d typecheck:%d, want 2/2", set.Len(), len(set.TypeCheckPrograms()))
	}
	binding, err := session.LoadAPI(set, lintPlan, dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(binding.TargetsByProgram) != 2 || len(binding.TargetsByProgram[0]) != 0 || len(binding.TargetsByProgram[1]) != 1 {
		t.Fatalf("catalog-only project stole the lint target: %v", binding.TargetsByProgram)
	}
}

func TestSelectProjectsSeparatesLintOnlyDefaultFromTypeCheckCatalog(t *testing.T) {
	dir := tspath.NormalizePath(t.TempDir())
	writeProgramTestFiles(t, dir, map[string]string{
		"a.ts":            `export const a = 1;`,
		"b.ts":            `export const b = 1;`,
		"tsconfig-a.json": `{"files":["a.ts"],"compilerOptions":{"noLib":true}}`,
		"tsconfig.json":   `{"files":["b.ts"],"compilerOptions":{"noLib":true}}`,
	})
	config := rslintconfig.RslintConfig{
		lintProjectEntry([]string{"a.ts"}, "./tsconfig-a.json"),
		lintProjectEntry([]string{"b.ts"}),
	}
	baseFS := cachedvfs.From(osvfs.FS())
	resolver, lintPlan := resolveEffectivePipelinePlan(t, config, dir, baseFS, filepath.Join(dir, "b.ts"))
	if got := resolver.CatalogProjectPaths(); !slices.Equal(got, []string{tspath.ResolvePath(dir, "tsconfig-a.json")}) {
		t.Fatalf("type-check catalog = %v", got)
	}
	session := NewSession(bundled.WrapFS(baseFS))
	set, err := session.SelectProjects(lintPlan, resolver.CatalogProjectPaths(), false, true)
	if err != nil {
		t.Fatal(err)
	}
	if set.Len() != 2 || len(set.TypeCheckPrograms()) != 1 {
		t.Fatalf("program roles = all:%d typecheck:%d, want 2/1", set.Len(), len(set.TypeCheckPrograms()))
	}
	binding, err := session.LoadAPI(set, lintPlan, dir, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(binding.TargetsByProgram) != 2 || len(binding.TargetsByProgram[0]) != 0 || len(binding.TargetsByProgram[1]) != 1 {
		t.Fatalf("lint-only default project was not selected independently: %v", binding.TargetsByProgram)
	}
}

func TestSelectProjectsCarriesCompleteBindingIntoLoad(t *testing.T) {
	dir := tspath.NormalizePath(t.TempDir())
	writeProgramTestFiles(t, dir, map[string]string{
		"main.ts":       `import "./imported";`,
		"direct.ts":     `export const direct = 1;`,
		"imported.ts":   `export const imported = 1;`,
		"gap.ts":        `export const gap = 1;`,
		"tsconfig.json": `{"files":["main.ts","direct.ts"],"compilerOptions":{"noLib":true}}`,
	})
	config := rslintconfig.RslintConfig{
		lintProjectEntry([]string{"**/*.ts"}, "./tsconfig.json"),
	}
	baseFS := cachedvfs.From(osvfs.FS())
	_, lintPlan := resolveEffectivePipelinePlan(
		t,
		config,
		dir,
		baseFS,
		filepath.Join(dir, "direct.ts"),
		filepath.Join(dir, "imported.ts"),
		filepath.Join(dir, "gap.ts"),
	)
	session := NewSession(bundled.WrapFS(baseFS))
	set, err := session.SelectProjects(lintPlan, nil, false, true)
	if err != nil {
		t.Fatal(err)
	}
	if set.targetBinding == nil || !slices.Equal(set.targetBinding.owners, []int{0, 0, -1}) {
		t.Fatalf("complete owners = %v, want [0 0 -1]", set.targetBinding)
	}
	binding, err := session.LoadAPI(set, lintPlan, dir, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(binding.Programs) != 2 ||
		len(binding.TargetsByProgram) != 2 ||
		len(binding.TargetsByProgram[0]) != 2 ||
		len(binding.TargetsByProgram[1]) != 1 {
		t.Fatalf("complete binding projection = %v", binding.TargetsByProgram)
	}
}

func TestSelectProjectsPreservesExplicitEmptyTypeCheckCatalog(t *testing.T) {
	dir := tspath.NormalizePath(t.TempDir())
	writeProgramTestFiles(t, dir, map[string]string{
		"target.ts":     `export const target = 1;`,
		"tsconfig.json": `{"files":["target.ts"],"compilerOptions":{"noLib":true}}`,
	})
	baseFS := cachedvfs.From(osvfs.FS())
	_, lintPlan := resolveEffectivePipelinePlan(
		t,
		rslintconfig.RslintConfig{lintProjectEntry([]string{"target.ts"}, "./tsconfig.json")},
		dir,
		baseFS,
		filepath.Join(dir, "target.ts"),
	)
	session := NewSession(bundled.WrapFS(baseFS))
	set, err := session.SelectProjects(lintPlan, []string{}, false, true)
	if err != nil {
		t.Fatal(err)
	}
	if set.TypeCheckPrograms() == nil || len(set.TypeCheckPrograms()) != 0 {
		t.Fatalf("explicit empty type-check catalog = %#v, want non-nil empty", set.TypeCheckPrograms())
	}
	loaded, err := session.LoadAPI(set, lintPlan, dir, true)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.TypeCheckPrograms == nil || len(loaded.TypeCheckPrograms) != 0 {
		t.Fatalf("loaded empty type-check catalog = %#v, want non-nil empty", loaded.TypeCheckPrograms)
	}
	if len(loaded.Programs) != 1 || len(loaded.TargetsByProgram) != 1 || len(loaded.TargetsByProgram[0]) != 1 {
		t.Fatalf("lint projection was lost: programs=%d targets=%v", len(loaded.Programs), loaded.TargetsByProgram)
	}
}

func TestSelectProjectsPreservesPerOwnerProjectOrder(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	child := tspath.ResolvePath(root, "child")
	writeProgramTestFiles(t, root, map[string]string{
		"root-target.ts":        `export const rootTarget = 1;`,
		"child/child-target.ts": `export const childTarget = 1;`,
		"a-main.ts":             `import "./root-target"; import "./child/child-target";`,
		"b-main.ts":             `import "./root-target"; import "./child/child-target";`,
		"tsconfig-a.json":       `{"files":["a-main.ts"],"compilerOptions":{"noLib":true}}`,
		"tsconfig-b.json":       `{"files":["b-main.ts"],"compilerOptions":{"noLib":true}}`,
	})
	baseFS := cachedvfs.From(osvfs.FS())
	fsys := bundled.WrapFS(baseFS)
	configs := map[string]rslintconfig.RslintConfig{
		root:  {lintProjectEntry([]string{"**/*.ts"}, "./tsconfig-a.json", "./tsconfig-b.json")},
		child: {lintProjectEntry([]string{"**/*.ts"}, "../tsconfig-b.json", "../tsconfig-a.json")},
	}
	resolver, err := rslintconfig.NewProjectPathResolver(configs, nil, root, fsys, false)
	if err != nil {
		t.Fatal(err)
	}
	targetPlan := rslintconfig.LintTargetPlan{Targets: []rslintconfig.DiscoveredLintTarget{
		{Path: tspath.ResolvePath(root, "root-target.ts"), CanonicalPath: tspath.ResolvePath(root, "root-target.ts"), ConfigDirectory: root},
		{Path: tspath.ResolvePath(child, "child-target.ts"), CanonicalPath: tspath.ResolvePath(child, "child-target.ts"), ConfigDirectory: child},
	}}
	lintPlan, err := resolver.ResolveLintProjectPlan(targetPlan)
	if err != nil {
		t.Fatal(err)
	}
	for _, prefetch := range []bool{false, true} {
		name := "focused"
		if prefetch {
			name = "prefetched"
		}
		t.Run(name, func(t *testing.T) {
			session := NewSession(fsys)
			set, err := session.SelectProjects(lintPlan, nil, prefetch, true)
			if err != nil {
				t.Fatal(err)
			}
			if set.targetBinding == nil || !slices.Equal(set.targetBinding.owners, []int{0, 1}) {
				t.Fatalf("per-owner binding = %+v, want owners [0 1]", set.targetBinding)
			}
			loaded, err := session.LoadAPI(set, lintPlan, root, true)
			if err != nil {
				t.Fatal(err)
			}
			if len(loaded.TargetsByProgram) != 2 || len(loaded.TargetsByProgram[0]) != 1 || len(loaded.TargetsByProgram[1]) != 1 {
				t.Fatalf("per-owner target projection = %v", loaded.TargetsByProgram)
			}
		})
	}
}
