package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
)

func writeProjectPlanFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func projectEntry(paths ...string) *LanguageOptions {
	if paths == nil {
		paths = []string{}
	}
	return &LanguageOptions{ParserOptions: &ParserOptions{Project: paths}}
}

func TestProjectPathResolverUsesTargetEffectiveProject(t *testing.T) {
	dir := tspath.NormalizePath(t.TempDir())
	for _, name := range []string{"tsconfig.json", "tsconfig-a.json", "tsconfig-b.json"} {
		writeProjectPlanFile(t, filepath.Join(dir, name), `{}`)
	}
	config := RslintConfig{
		{Files: []string{"**/*.ts"}, LanguageOptions: projectEntry("./tsconfig-a.json")},
		{Files: []string{"default/**"}},
		{Files: []string{"tests/**"}, LanguageOptions: projectEntry("./tsconfig-b.json")},
		{Files: []string{"inherit/**"}, LanguageOptions: &LanguageOptions{ParserOptions: &ParserOptions{Project: nil}}},
		{Files: []string{"clear/**"}, LanguageOptions: projectEntry()},
	}
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	resolver, err := NewProjectPathResolver(nil, config, dir, fsys, false)
	if err != nil {
		t.Fatal(err)
	}
	targets := []struct {
		name      string
		relative  string
		want      string
		wantState ProjectDeclarationState
		effective bool
	}{
		{name: "base", relative: "src/a.ts", want: "tsconfig-a.json", wantState: ProjectExplicit, effective: true},
		{name: "override", relative: "tests/a.ts", want: "tsconfig-b.json", wantState: ProjectExplicit, effective: true},
		{name: "nil inherits", relative: "inherit/a.ts", want: "tsconfig-a.json", wantState: ProjectExplicit, effective: true},
		{name: "empty clears then defaults", relative: "clear/a.ts", want: "tsconfig.json", wantState: ProjectExplicitEmpty, effective: true},
		{name: "matched unspecified defaults", relative: "default/a.js", want: "tsconfig.json", wantState: ProjectUnspecified, effective: true},
		{name: "no matching entry", relative: "outside.js", wantState: ProjectUnspecified},
	}
	for _, test := range targets {
		t.Run(test.name, func(t *testing.T) {
			path := tspath.ResolvePath(dir, test.relative)
			planned, err := resolver.ResolveLintTarget(DiscoveredLintTarget{
				Path: path, CanonicalPath: path, ConfigDirectory: dir,
			})
			if err != nil {
				t.Fatal(err)
			}
			if (planned.Effective != nil) != test.effective {
				t.Fatalf("effective config present = %t, want %t", planned.Effective != nil, test.effective)
			}
			if got := planned.Effective.ProjectState(); got != test.wantState {
				t.Fatalf("project state = %v, want %v", got, test.wantState)
			}
			if test.want == "" {
				if len(planned.ProjectPaths) != 0 {
					t.Fatalf("project paths = %v, want none", planned.ProjectPaths)
				}
				return
			}
			want := tspath.ResolvePath(dir, test.want)
			if !slices.Equal(planned.ProjectPaths, []string{want}) {
				t.Fatalf("project paths = %v, want [%s]", planned.ProjectPaths, want)
			}
		})
	}
}

func TestProjectPathResolverPreservesProjectShapeFromJSON(t *testing.T) {
	dir := tspath.NormalizePath(t.TempDir())
	for _, name := range []string{"tsconfig.json", "tsconfig-a.json"} {
		writeProjectPlanFile(t, filepath.Join(dir, name), `{}`)
	}
	var config RslintConfig
	if err := json.Unmarshal([]byte(`[
		{"files":["**/*.ts"],"languageOptions":{"parserOptions":{"project":["./tsconfig-a.json"]}}},
		{"files":["inherit/**"],"languageOptions":{"parserOptions":{}}},
		{"files":["clear/**"],"languageOptions":{"parserOptions":{"project":[]}}}
	]`), &config); err != nil {
		t.Fatal(err)
	}
	if config[2].LanguageOptions.ParserOptions.Project == nil {
		t.Fatal("JSON project:[] was collapsed into an unspecified project")
	}
	resolver, err := NewProjectPathResolver(
		nil,
		config,
		dir,
		bundled.WrapFS(cachedvfs.From(osvfs.FS())),
		false,
	)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		path string
		want string
	}{
		{path: "inherit/a.ts", want: "tsconfig-a.json"},
		{path: "clear/a.ts", want: "tsconfig.json"},
	}
	for _, test := range tests {
		target := tspath.ResolvePath(dir, test.path)
		planned, err := resolver.ResolveLintTarget(DiscoveredLintTarget{
			Path: target, CanonicalPath: target, ConfigDirectory: dir,
		})
		if err != nil {
			t.Fatal(err)
		}
		want := tspath.ResolvePath(dir, test.want)
		if !slices.Equal(planned.ProjectPaths, []string{want}) {
			t.Fatalf("%s projects = %v, want [%s]", test.path, planned.ProjectPaths, want)
		}
	}
}

func TestProjectPathResolverValidatesUnmatchedDeclarations(t *testing.T) {
	dir := tspath.NormalizePath(t.TempDir())
	_, err := NewProjectPathResolver(nil, RslintConfig{{
		Files:           []string{"never/**"},
		LanguageOptions: projectEntry("./missing.json"),
	}}, dir, bundled.WrapFS(cachedvfs.From(osvfs.FS())), false)
	if err == nil {
		t.Fatal("unmatched invalid project declaration was not rejected")
	}
}

func TestProjectPathResolverBuildsTypeCheckCatalogPerOwner(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	child := tspath.ResolvePath(root, "packages/child")
	empty := tspath.ResolvePath(root, "packages/empty")
	shared := tspath.ResolvePath(root, "tsconfig.shared.json")
	for _, path := range []string{
		shared,
		tspath.ResolvePath(child, "tsconfig.json"),
		tspath.ResolvePath(empty, "tsconfig.json"),
	} {
		writeProjectPlanFile(t, path, `{}`)
	}
	configs := map[string]RslintConfig{
		root:  {{LanguageOptions: projectEntry(shared)}},
		child: {{}},
		empty: {{LanguageOptions: projectEntry()}},
	}
	resolver, err := NewProjectPathResolver(
		configs,
		nil,
		root,
		bundled.WrapFS(cachedvfs.From(osvfs.FS())),
		false,
	)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		shared,
		tspath.ResolvePath(child, "tsconfig.json"),
		tspath.ResolvePath(empty, "tsconfig.json"),
	}
	got := resolver.CatalogProjectPaths()
	if !slices.Equal(got, want) {
		t.Fatalf("catalog projects = %v, want %v", got, want)
	}
}
