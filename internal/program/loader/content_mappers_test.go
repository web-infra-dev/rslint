package loader

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/cachedvfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/program/contentmappers"
	"github.com/web-infra-dev/rslint/internal/testutil/contentmappertest"
)

func writeContentMapperFiles(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	contentmappertest.WritePackage(t, dir, "fake-mapper")
	for name, content := range files {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

const contentMapperTSConfig = `{
  "compilerOptions": { "strict": true, "noEmit": true },
  "files": ["index.ts"],
  "contentMappers": [{ "package": "fake-mapper", "extensions": [".fake"] }]
}`

func TestBuildContextRunsContentMappersOnlyWhenEnabled(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeContentMapperFiles(t, dir, map[string]string{
		"tsconfig.json": contentMapperTSConfig,
		"index.ts":      "import { answer } from './data.fake';\nexport const doubled: number = answer * 2;\n",
		"data.fake":     "answer = 42\n",
	})
	root := tspath.NormalizePath(dir)

	for _, enabled := range []bool{false, true} {
		spawner := &contentmappertest.Spawner{Output: "export const answer: number = 42;\n"}
		mappers := contentmappers.New(contentmappers.Options{Enabled: enabled, Spawner: spawner})
		context := newBuildContext(bundled.WrapFS(cachedvfs.From(osvfs.FS())), mappers)
		program, err := context.createProjectProgram(true, root, "tsconfig.json")
		if err != nil {
			t.Fatal(err)
		}
		mapped := program.GetSourceFile(root + "/data.fake")
		if enabled != (mapped != nil && mapped.ContentMapper() != "") {
			t.Fatalf("enabled=%v: mapped file present=%v", enabled, mapped != nil)
		}
		wantIgnored := []string{}
		if !enabled {
			wantIgnored = []string{root + "/tsconfig.json"}
		}
		if got := mappers.IgnoredConfigs(); !slices.Equal(got, wantIgnored) {
			t.Fatalf("enabled=%v: IgnoredConfigs() = %v, want %v", enabled, got, wantIgnored)
		}
		if err := mappers.Close(); err != nil {
			t.Fatal(err)
		}
		if spawner.Spawns() != spawner.Closes() {
			t.Fatalf("enabled=%v: %d mappers started, %d stopped", enabled, spawner.Spawns(), spawner.Closes())
		}
	}
}

func TestBuildContextParsesProjectReferencesWithContentMapperOptions(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeContentMapperFiles(t, dir, map[string]string{
		"tsconfig.json": `{
  "compilerOptions": { "noEmit": true },
  "files": ["index.ts"],
  "references": [{ "path": "./lib" }]
}`,
		"index.ts":          "export {};\n",
		"lib/tsconfig.json": `{ "compilerOptions": { "composite": true }, "files": ["index.ts"], "contentMappers": [{ "package": "fake-mapper", "extensions": [".fake"] }] }`,
		"lib/index.ts":      "export const value = 1;\n",
	})
	root := tspath.NormalizePath(dir)

	mappers := contentmappers.New(contentmappers.Options{})
	defer mappers.Close()
	context := newBuildContext(bundled.WrapFS(cachedvfs.From(osvfs.FS())), mappers)
	if _, err := context.createProjectProgram(true, root, "tsconfig.json"); err != nil {
		t.Fatal(err)
	}
	if got, want := mappers.IgnoredConfigs(), []string{root + "/lib/tsconfig.json"}; !slices.Equal(got, want) {
		t.Fatalf("IgnoredConfigs() = %v, want %v", got, want)
	}
}
