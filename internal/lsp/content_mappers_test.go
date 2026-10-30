package lsp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/lsp/lsproto"
	"github.com/microsoft/TypeScript/tsc/shim/project"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/program/contentmappers"
	"github.com/web-infra-dev/rslint/internal/testutil/contentmappertest"
)

func TestInitializeReadsRunExternalCode(t *testing.T) {
	t.Parallel()
	enabled, disabled := true, false
	tests := []struct {
		name    string
		options *lsproto.InitializationOptionsOrNull
		want    bool
	}{
		{name: "absent"},
		{name: "null", options: &lsproto.InitializationOptionsOrNull{}},
		{name: "unset", options: &lsproto.InitializationOptionsOrNull{InitializationOptions: &lsproto.InitializationOptions{}}},
		{name: "false", options: &lsproto.InitializationOptionsOrNull{InitializationOptions: &lsproto.InitializationOptions{RunExternalCode: &disabled}}},
		{name: "true", options: &lsproto.InitializationOptionsOrNull{InitializationOptions: &lsproto.InitializationOptions{RunExternalCode: &enabled}}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := newTestServer()
			if _, err := s.handleInitialize(context.Background(), &lsproto.InitializeParams{InitializationOptions: tt.options}); err != nil {
				t.Fatal(err)
			}
			if s.contentMappers == nil {
				t.Fatal("initialize must create the content mapper Runtime")
			}
			if got := s.contentMappers.Enabled(); got != tt.want {
				t.Fatalf("Enabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

const mappedSource = "import { answer } from './data" + contentmappertest.Extension + "';\nexport const doubled = answer * 2;\n"

func mappedConfig(strict bool) string {
	strictOption := "false"
	if strict {
		strictOption = "true"
	}
	return `{"compilerOptions":{"noLib":true,"strict":` + strictOption + `},"include":["src/**/*.ts"],` +
		`"contentMappers":[{"package":"fake-mapper","extensions":["` + contentmappertest.Extension + `"]}]}`
}

// newMappedLintProgramStoreFixture is a store fixture whose source imports a
// file only the fake mapper provides.
func newMappedLintProgramStoreFixture(t *testing.T) (*lintProgramStoreFixture, *contentmappertest.Spawner) {
	t.Helper()
	fixture := newLintProgramStoreFixture(t, mappedSource)
	root := filepath.Dir(fixture.configPath)
	contentmappertest.WritePackage(t, root, "fake-mapper")
	if err := os.WriteFile(fixture.configPath, []byte(mappedConfig(false)), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "data"+contentmappertest.Extension), []byte("answer = 42\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	spawner := &contentmappertest.Spawner{Output: "export const answer: number = 42;\n"}
	fixture.server.contentMappers = contentmappers.New(contentmappers.Options{Enabled: true, Spawner: spawner})
	t.Cleanup(func() { _ = fixture.server.contentMappers.Close() })
	return fixture, spawner
}

func TestLintProgramStoreKeepsMapperAcrossRebuilds(t *testing.T) {
	fixture, spawner := newMappedLintProgramStoreFixture(t)
	first := fixture.load(t)

	configURI := documentURIFromPath(fixture.configPath)
	fixture.server.documents[configURI] = mappedConfig(true)
	fixture.store.DidChange(configURI, mappedConfig(true))
	if rebuilt := fixture.load(t); rebuilt == first {
		t.Fatal("tsconfig change did not rebuild the Program")
	}
	if spawner.Spawns() != 1 || spawner.Closes() != 0 {
		t.Fatalf("after the rebuild: %d mappers started, %d stopped; want 1, 0", spawner.Spawns(), spawner.Closes())
	}
	if len(fixture.store.mapperStates) != 1 {
		t.Fatalf("%d states hold mapper projects after the rebuild, want 1", len(fixture.store.mapperStates))
	}

	delete(fixture.server.documents, fixture.sourceURI)
	fixture.store.DidClose(fixture.sourceURI)
	if spawner.Closes() != 1 {
		t.Fatalf("stopped %d mappers after the last Program was released, want 1", spawner.Closes())
	}
}

func TestStandaloneLintRequestsKeepMapperRunning(t *testing.T) {
	fixture, spawner := newMappedLintProgramStoreFixture(t)
	fs := bundled.WrapFS(osvfs.FS())
	for range 2 {
		request := newStandaloneLintProjectRequestWithFS(
			lspConfigTarget(fixture.sourcePath, fixture.server.cwd, fs), fs, fixture.server.contentMappers,
		)
		metadata, err := request.metadata(fixture.configPath)
		if err != nil {
			t.Fatal(err)
		}
		if _, sourceFile, err := request.program(metadata); err != nil || sourceFile == nil {
			t.Fatalf("standalone request did not lint the target: %v", err)
		}
	}
	if spawner.Spawns() != 1 || spawner.Closes() != 0 {
		t.Fatalf("%d mappers started, %d stopped; want 1, 0", spawner.Spawns(), spawner.Closes())
	}
}

func TestLintProgramStoreWatchesDynamicMapperConfig(t *testing.T) {
	for _, event := range []lsproto.FileChangeType{
		lsproto.FileChangeTypeChanged, lsproto.FileChangeTypeCreated, lsproto.FileChangeTypeDeleted,
	} {
		t.Run(fmt.Sprint(event), func(t *testing.T) {
			fixture, spawner := newMappedLintProgramStoreFixture(t)
			root := filepath.Dir(fixture.configPath)
			manifest := filepath.Join(root, "node_modules", "fake-mapper", "package.json")
			if err := os.WriteFile(manifest, []byte(`{"name":"fake-mapper","version":"1.0.0","typescript":{"contentMapper":{"exec":["node","server.js"],"dynamicConfig":true}}}`), 0o644); err != nil {
				t.Fatal(err)
			}
			dependency := tspath.NormalizePath(filepath.Join(t.TempDir(), "mapper.config.ts"))
			spawner.ConfigFile = dependency
			writeConfig := func(output string) {
				t.Helper()
				if err := os.WriteFile(dependency, []byte(output), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			writeConfig("export const answer: number = 42;")
			registered := false
			fixture.store.coverage.watchFiles = func(_ context.Context, _ project.WatcherID, watchers []*lsproto.FileSystemWatcher) error {
				for _, watcher := range watchers {
					encoded, _ := json.Marshal(watcher)
					if strings.Contains(strings.ToLower(string(encoded)), strings.ToLower(filepath.Dir(dependency))) {
						registered = true
						// Change during registration: the retry must reopen the mapper.
						writeConfig("export const answer: string = 'registered';")
					}
				}
				return nil
			}
			first := fixture.load(t)
			if !registered {
				t.Fatal("mapper dependency outside the workspace was not watched")
			}
			mappedPath := tspath.NormalizePath(filepath.Join(root, "src", "data"+contentmappertest.Extension))
			assertOutput := func(program *compiler.Program, expected string) {
				t.Helper()
				source := program.GetSourceFile(mappedPath)
				if source == nil || !strings.Contains(source.Text(), expected) {
					t.Fatalf("mapped source did not contain %q: %v", expected, source)
				}
			}
			assertOutput(first, "registered")
			if fixture.load(t) != first {
				t.Fatal("unchanged Program was not retained")
			}
			// An open .ts dependency must bypass the source-overlay shortcut.
			uri := documentURIFromPath(dependency)
			fixture.server.documents[uri] = "unsaved editor content"
			writeConfig("export const answer: boolean = true;")
			if !fixture.store.DidChangeWatchedFiles([]*lsproto.FileEvent{{Uri: uri, Type: event}}) {
				t.Fatal("mapper dependency change did not request diagnostic refresh")
			}
			rebuilt := fixture.load(t)
			if rebuilt == first {
				t.Fatal("mapper dependency change reused the Program")
			}
			assertOutput(rebuilt, "boolean")
			if spawner.Spawns() != 1 {
				t.Fatal("refresh restarted the mapper process")
			}
		})
	}
}
