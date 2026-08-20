package loader

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
)

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
	session := NewSession(fsys)
	set, err := session.SelectProjects(lintPlan, nil, false, true)
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
}
