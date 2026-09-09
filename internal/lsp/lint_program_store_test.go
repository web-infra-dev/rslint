package lsp

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/lsp/lsproto"
	"github.com/microsoft/TypeScript/tsc/shim/project"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
)

type lintProgramStoreFixture struct {
	server     *Server
	store      *lintProgramStore
	configPath string
	sourcePath string
	sourceURI  lsproto.DocumentUri
	watchCalls int
}

func newLintProgramStoreFixture(t *testing.T, source string) *lintProgramStoreFixture {
	t.Helper()

	root := t.TempDir()
	sourcePath := filepath.Join(root, "src", "index.ts")
	if err := os.MkdirAll(filepath.Dir(sourcePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sourcePath, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(root, "tsconfig.json")
	if err := os.WriteFile(
		configPath,
		[]byte(`{"compilerOptions":{"noLib":true},"include":["src/**/*.ts"]}`),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	server := newTestServer()
	server.cwd = root
	server.fs = bundled.WrapFS(osvfs.FS())
	server.defaultLibraryPath = bundled.LibPath()
	server.initializeParams = &lsproto.InitializeParams{}
	sourceURI := documentURIFromPath(sourcePath)
	server.documents[sourceURI] = source

	fixture := &lintProgramStoreFixture{
		server:     server,
		configPath: configPath,
		sourcePath: tspath.NormalizePath(sourcePath),
		sourceURI:  sourceURI,
	}
	fixture.store = newLintProgramStore(server)
	fixture.store.coverage.watchFiles = func(
		context.Context,
		project.WatcherID,
		[]*lsproto.FileSystemWatcher,
	) error {
		fixture.watchCalls++
		return nil
	}
	return fixture
}

func (f *lintProgramStoreFixture) request(
	uri lsproto.DocumentUri,
) (func(string) (*compiler.Program, *ast.SourceFile, error), lintProjectMetadataLoader, func()) {
	loader, metadata, finalize := f.store.Request(
		context.Background(),
		uri,
		lspConfigTarget(uriToPath(uri), f.server.cwd, f.server.fs),
	)
	return func(path string) (*compiler.Program, *ast.SourceFile, error) {
		selected, _, err := metadata(path)
		if err != nil {
			return nil, nil, err
		}
		return loader(selected)
	}, metadata, finalize
}

func (f *lintProgramStoreFixture) load(t *testing.T) *compiler.Program {
	t.Helper()
	loader, _, finalize := f.request(f.sourceURI)
	program, _, err := loader(f.configPath)
	if err != nil {
		t.Fatalf("load lint Program: %v", err)
	}
	finalize()
	return program
}

func TestLintProgramStoreReusesAndUpdatesSource(t *testing.T) {
	const initial = "export const value = 1;\n"
	fixture := newLintProgramStoreFixture(t, initial)

	first := fixture.load(t)
	second := fixture.load(t)
	if second != first {
		t.Fatal("unchanged document rebuilt its standalone Program")
	}
	if fixture.watchCalls == 0 {
		t.Fatal("initial Program did not register dependency coverage")
	}

	const changed = "export const value = 2;\n"
	fixture.server.documents[fixture.sourceURI] = changed
	fixture.store.DidChange(fixture.sourceURI, changed)
	updated := fixture.load(t)
	if updated == first {
		t.Fatal("changed document did not advance the Program")
	}
	if updated.Host() != first.Host() {
		t.Fatal("source-only update replaced the stable compiler host")
	}
	sourceFile := updated.GetSourceFile(fixture.sourcePath)
	if sourceFile == nil {
		t.Fatal("updated Program does not contain the lint target")
		return
	}
	if sourceFile.Text() != changed {
		t.Fatalf("updated source text = %q, want %q", sourceFile.Text(), changed)
	}
}

func selectLintProgramRequestForTest(request *lintProgramRequest, rootDirectory string) (selectedLintProject, error) {
	request.prepareOverlay()
	selected, _, err := selectConfiguredLintProject(nil, rootDirectory, request.target, request.overlayFS, lintProjectLoaders{
		program: request.load, metadata: request.loadMetadata,
	})
	return selected, err
}

func TestLintProgramStoreProjectServiceReusesModeAndUpdatesSource(t *testing.T) {
	const original = "export const value = 1;\n"
	fixture := newLintProgramStoreFixture(t, original)
	loadService := func() *compiler.Program {
		t.Helper()
		request := fixture.store.request(context.Background(), fixture.sourceURI,
			lspConfigTarget(fixture.sourcePath, fixture.server.cwd, fixture.server.fs), true)
		selected, err := selectLintProgramRequestForTest(request, fixture.server.cwd)
		if err != nil {
			t.Fatal(err)
		}
		request.finalize()
		return selected.program
	}
	first := loadService()
	if second := loadService(); second != first {
		t.Fatal("unchanged service request rebuilt its configured Program")
	}
	legacy := fixture.load(t)
	if legacy == first {
		t.Fatal("service and legacy requests reused different reference modes")
	}
	if afterLegacy := loadService(); afterLegacy != first {
		t.Fatal("legacy request evicted the service-mode Program")
	}
	const changed = "export const value = 2;\n"
	fixture.server.documents[fixture.sourceURI] = changed
	fixture.store.DidChange(fixture.sourceURI, changed)
	updated := loadService()
	if updated == first || updated.GetSourceFile(fixture.sourcePath).Text() != changed {
		t.Fatal("service cache did not advance to the editor generation")
	}
	if first.GetSourceFile(fixture.sourcePath).Text() != original {
		t.Fatal("updating the service cache mutated a previous Program")
	}
	if !fixture.store.Invalidate() || len(fixture.store.programs) != 0 {
		t.Fatal("invalidation did not discard both Program modes")
	}
}

func TestLintProgramStoreProjectServiceIgnoresPreviouslyLoadedReferences(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/project_service.txtar")
	directory := tspath.NormalizePath(archive.Materialize(t, "warm-references"))
	server := newTestServer()
	server.cwd = directory
	server.fs = bundled.WrapFS(osvfs.FS())
	store := newLintProgramStore(server)
	server.lintPrograms = store
	store.coverage.watchFiles = func(context.Context, project.WatcherID, []*lsproto.FileSystemWatcher) error { return nil }
	load := func(relativePath string) (*compiler.Program, error) {
		fileName := tspath.ResolvePath(directory, relativePath)
		uri := documentURIFromPath(fileName)
		server.documents[uri] = "export const value = 1;\n"
		request := store.request(context.Background(), uri, lspConfigTarget(fileName, directory, server.fs), true)
		defer request.finalize()
		selected, err := selectLintProgramRequestForTest(request, directory)
		return selected.program, err
	}
	if program, err := load("app/second.ts"); err != nil || program != nil {
		t.Fatalf("cold reference should remain a project gap: program=%v error=%v", program, err)
	}
	legacyTarget := tspath.ResolvePath(directory, "shared/first.ts")
	legacyRequest := store.request(context.Background(), documentURIFromPath(legacyTarget), lspConfigTarget(legacyTarget, directory, server.fs), false)
	legacyMetadata, err := legacyRequest.metadata(tspath.ResolvePath(directory, "shared/tsconfig.json"))
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := legacyRequest.load(legacyMetadata); err != nil {
		t.Fatal(err)
	}
	legacyRequest.finalize()
	if program, err := load("app/second.ts"); err != nil || program != nil {
		t.Fatalf("legacy Program changed the service gap: program=%v error=%v", program, err)
	}
	first, err := load("shared/first.ts")
	if err != nil {
		t.Fatal(err)
	}
	second, err := load("app/second.ts")
	// rslint's direct-root discovery does not emulate TS Server's warm-project
	// exception to disableReferencedProjectLoad.
	if err != nil || second != nil || first == nil {
		t.Fatalf("loaded reference bypassed disabled traversal: program=%v error=%v", second, err)
	}
	fileName := tspath.ResolvePath(directory, "app/second.ts")
	uri := documentURIFromPath(fileName)
	entries := config.RslintConfig{{LanguageOptions: &config.LanguageOptions{
		ParserOptions: &config.ParserOptions{ProjectService: config.BoolPtr(true)},
	}}}
	snapshot := documentLintSnapshotForTest(server, uri, entries, directory, false, nil)
	environment := server.freezeSpeculativeLintEnvironment(uri, snapshot.target)
	const fixedContent = "export const fixed = 3;\n"
	generation, release, err := acquireSpeculativeGeneration(context.Background(), fixedContent, snapshot, environment)
	if release != nil {
		defer release()
	}
	if err != nil {
		t.Fatalf("speculative gap: %v", err)
	}
	if len(generation.Native.Programs) != 1 {
		t.Fatalf("speculative Programs=%d", len(generation.Native.Programs))
	}
	speculative := generation.Native.Programs[0]
	if speculative.Options().ConfigFilePath != "" || !speculative.Options().NoResolve.IsTrue() {
		t.Fatal("speculative generation bypassed disabled reference traversal")
	}
	if speculative.GetSourceFile(fileName).Text() != fixedContent || speculative.GetSourceFile(fileName) == first.GetSourceFile(fileName) {
		t.Fatal("speculative generation reused resident text")
	}
	if first.GetSourceFile(fileName).Text() != server.documents[uri] {
		t.Fatal("speculative generation changed resident source text")
	}
}

func TestLintProgramStoreProjectServiceReselectsAfterConfigChanges(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/project_service.txtar")
	directory := tspath.NormalizePath(archive.Materialize(t, "ancestor"))
	fileName := tspath.ResolvePath(directory, "pkg/target.ts")
	uri := documentURIFromPath(fileName)
	server := newTestServer()
	server.cwd = directory
	server.fs = bundled.WrapFS(osvfs.FS())
	store := newLintProgramStore(server)
	store.coverage.watchFiles = func(context.Context, project.WatcherID, []*lsproto.FileSystemWatcher) error { return nil }
	load := func(rootDir string) (selectedLintProject, error) {
		request := store.request(context.Background(), uri, lspConfigTarget(fileName, directory, server.fs), true)
		defer request.finalize()
		return selectLintProgramRequestForTest(request, rootDir)
	}
	first, err := load(directory)
	if err != nil || first.configPath != tspath.ResolvePath(directory, "tsconfig.json") {
		t.Fatalf("initial selection=%s error=%v", first.configPath, err)
	}
	if selected, err := load(tspath.ResolvePath(directory, "pkg")); err != nil || selected.program != nil {
		t.Fatalf("root boundary should leave a project gap: config=%s error=%v", selected.configPath, err)
	}
	nearestConfig := tspath.ResolvePath(directory, "pkg/tsconfig.json")
	configURI := documentURIFromPath(nearestConfig)
	const nearestContent = `{"compilerOptions":{"noLib":true,"strict":true},"files":["target.ts"]}`
	server.documents[configURI] = nearestContent
	store.DidOpen(configURI, nearestContent, true)
	nearest, err := load(directory)
	if err != nil || nearest.configPath != nearestConfig || !nearest.program.Options().Strict.IsTrue() {
		t.Fatalf("unsaved nearest config selection=%s error=%v", nearest.configPath, err)
	}
	delete(server.documents, configURI)
	store.DidClose(configURI)
	restored, err := load(directory)
	if err != nil || restored.configPath != first.configPath {
		t.Fatalf("closed config selection=%s error=%v", restored.configPath, err)
	}
	const parentContent = `{"compilerOptions":{"noLib":true,"strict":true},"files":["pkg/target.ts"]}`
	if err := os.WriteFile(first.configPath, []byte(parentContent), 0o644); err != nil {
		t.Fatal(err)
	}
	if !store.DidChangeWatchedFiles([]*lsproto.FileEvent{{Uri: documentURIFromPath(first.configPath), Type: lsproto.FileChangeTypeChanged}}) {
		t.Fatal("watched config change retained service state")
	}
	updated, err := load(directory)
	if err != nil || updated.program == restored.program || !updated.program.Options().Strict.IsTrue() {
		t.Fatalf("watched config did not update compiler options: %v", err)
	}
}

func TestLintProgramStoreProjectServiceFinalizesSelectedProject(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/project_service.txtar")
	directory := tspath.NormalizePath(archive.Materialize(t, "fallback-finalizer"))
	fileName := tspath.ResolvePath(directory, "target.ts")
	uri := documentURIFromPath(fileName)
	server := newTestServer()
	server.cwd = directory
	server.fs = bundled.WrapFS(osvfs.FS())
	store := newLintProgramStore(server)
	store.coverage.watchFiles = func(context.Context, project.WatcherID, []*lsproto.FileSystemWatcher) error { return nil }
	request := store.request(context.Background(), uri, lspConfigTarget(fileName, directory, server.fs), true)
	selected, err := selectLintProgramRequestForTest(request, directory)
	if err != nil {
		t.Fatal(err)
	}
	selectedKey := request.key(tspath.ResolvePath(directory, "tsconfig.json"))
	laterKey := request.key(tspath.ResolvePath(directory, "jsconfig.json"))
	if selected.configPath != tspath.ResolvePath(directory, "tsconfig.json") || store.programs[selectedKey] == nil || store.programs[laterKey] != nil {
		t.Fatalf("built a Program beyond the metadata-selected project: selected=%s Programs=%v", selected.configPath, store.programs)
	}
	// Emulate a checker's lazy dependency read outside the existing coverage.
	selected.program.Host().FS().FileExists(tspath.NormalizePath(filepath.Join(t.TempDir(), "lazy.d.ts")))
	request.finalize()
	if store.programs[selectedKey] != nil {
		t.Fatal("selected fallback retained lazy reads predating watcher coverage")
	}
	if store.programs[laterKey] != nil {
		t.Fatal("finalization created an unrelated project")
	}
}

func TestLintProgramStoreProjectServiceUpdatesReferencedSources(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/project_service.txtar")
	directory := tspath.NormalizePath(archive.Materialize(t, "reference-sources"))
	fileName := tspath.ResolvePath(directory, "app/src/main.ts")
	referencePath := tspath.ResolvePath(directory, "lib/src/value.ts")
	uri := documentURIFromPath(fileName)
	referenceURI := documentURIFromPath(referencePath)
	server := newTestServer()
	server.cwd = directory
	server.fs = bundled.WrapFS(osvfs.FS())
	const original = "export const value = 42;\n"
	server.documents[referenceURI] = original
	store := newLintProgramStore(server)
	store.coverage.watchFiles = func(context.Context, project.WatcherID, []*lsproto.FileSystemWatcher) error { return nil }
	load := func() *compiler.Program {
		request := store.request(context.Background(), uri, lspConfigTarget(fileName, directory, server.fs), true)
		defer request.finalize()
		selected, err := selectLintProgramRequestForTest(request, directory)
		if err != nil {
			t.Fatal(err)
		}
		return selected.program
	}
	first := load()
	if first.GetSourceFile(referencePath).Text() != original {
		t.Fatal("initial service Program did not use referenced editor source")
	}
	if diagnostics := first.GetSemanticDiagnostics(context.Background(), first.GetSourceFile(fileName)); len(diagnostics) != 0 {
		t.Fatalf("initial semantic diagnostics=%v", diagnostics)
	}
	const changed = "export const value = false;\n"
	server.documents[referenceURI] = changed
	store.DidChange(referenceURI, changed)
	updated := load()
	if updated == first || updated.GetSourceFile(referencePath).Text() != changed {
		t.Fatal("referenced editor change did not advance the consuming Program")
	}
	if first.GetSourceFile(referencePath).Text() != original {
		t.Fatal("reference update mutated the prior Program")
	}
	if !updated.IsSourceFromProjectReference(updated.GetSourceFile(referencePath).Path()) ||
		updated.GetSourceFile(tspath.ResolvePath(directory, "lib/dist/value.d.ts")) != nil {
		t.Fatal("incremental update lost project source-reference semantics")
	}
	foundMismatch := false
	for _, diagnostic := range updated.GetSemanticDiagnostics(context.Background(), updated.GetSourceFile(fileName)) {
		foundMismatch = foundMismatch || diagnostic.Code() == 2322
	}
	if !foundMismatch {
		t.Fatal("updated consumer did not report the new referenced boolean-to-number mismatch")
	}
}

func TestLintProgramStoreProjectServiceKeepsDisabledReferencesColdAfterInvalidation(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/project_service.txtar")
	for _, event := range []string{"source", "config-options", "config-membership", "cache-disabled"} {
		t.Run(event, func(t *testing.T) {
			directory := tspath.NormalizePath(archive.Materialize(t, "warm-references"))
			server := newTestServer()
			server.cwd = directory
			server.fs = bundled.WrapFS(osvfs.FS())
			store := newLintProgramStore(server)
			store.coverage.watchFiles = func(context.Context, project.WatcherID, []*lsproto.FileSystemWatcher) error { return nil }
			load := func(relativePath string) (selectedLintProject, error) {
				fileName := tspath.ResolvePath(directory, relativePath)
				request := store.request(context.Background(), documentURIFromPath(fileName), lspConfigTarget(fileName, directory, server.fs), true)
				defer request.finalize()
				return selectLintProgramRequestForTest(request, directory)
			}
			first, err := load("shared/first.ts")
			if err != nil {
				t.Fatal(err)
			}
			changedPath := tspath.ResolvePath(directory, "app/second.ts")
			content := "export const changed = 3;\n"
			switch event {
			case "config-options", "config-membership":
				changedPath = tspath.ResolvePath(directory, "shared/tsconfig.json")
				content = `{"compilerOptions":{"noLib":true,"composite":true,"strict":true},"files":["first.ts","../app/second.ts"]}`
				if event == "config-membership" {
					content = `{"compilerOptions":{"noLib":true,"composite":true},"files":["first.ts"]}`
				}
			case "cache-disabled":
				store.coverage.disabled = true
			}
			if err := os.WriteFile(changedPath, []byte(content), 0o644); err != nil {
				t.Fatal(err)
			}
			store.DidChangeWatchedFiles([]*lsproto.FileEvent{{Uri: documentURIFromPath(changedPath), Type: lsproto.FileChangeTypeChanged}})
			if len(store.programs) != 0 {
				t.Fatal("watch event retained a stale Program")
			}
			selected, err := load("app/second.ts")
			if err != nil || selected.program != nil {
				t.Fatalf("disabled reference traversal after %s: selected=%s error=%v", event, selected.configPath, err)
			}
			if len(store.programs) != 0 || first.program == nil {
				t.Fatal("unowned target constructed a Program or initial seed did not load")
			}
		})
	}
}

func TestLintProgramStoreProjectServiceDoesNotConstructUnownedProbes(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/project_service.txtar")
	directory := tspath.NormalizePath(archive.Materialize(t, "transient-service-project"))
	server := newTestServer()
	server.cwd = directory
	server.fs = bundled.WrapFS(osvfs.FS())
	store := newLintProgramStore(server)
	store.coverage.watchFiles = func(context.Context, project.WatcherID, []*lsproto.FileSystemWatcher) error { return nil }
	load := func(relativePath string) (selectedLintProject, error) {
		fileName := tspath.ResolvePath(directory, relativePath)
		request := store.request(context.Background(), documentURIFromPath(fileName), lspConfigTarget(fileName, directory, server.fs), true)
		defer request.finalize()
		return selectLintProgramRequestForTest(request, directory)
	}
	if selected, err := load("app/second.ts"); err != nil || selected.program != nil {
		t.Fatalf("parsed reference should leave a project gap: config=%s error=%v", selected.configPath, err)
	}
	if selected, err := load("shared/unowned.ts"); err != nil || selected.program != nil {
		t.Fatalf("non-containing project should leave a gap: config=%s error=%v", selected.configPath, err)
	}
	if len(store.programs) != 0 {
		t.Fatal("non-containing probe retained its Program")
	}
	selected, err := load("app/second.ts")
	if err != nil || selected.program != nil || len(store.programs) != 0 {
		t.Fatalf("unowned probe changed later ownership: selected=%s error=%v", selected.configPath, err)
	}
}

func TestLintProgramStoreProjectServiceCaseInsensitiveIdentity(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/project_service.txtar")
	const directory = "/repo"
	files := make(map[string]string)
	names, err := archive.FileNames("warm-references")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range names {
		content, err := archive.ReadFile("warm-references/" + name)
		if err != nil {
			t.Fatal(err)
		}
		// Exercise path identity under ordinary reference traversal, independent
		// of historical project loading.
		content = []byte(strings.ReplaceAll(string(content), `"disableReferencedProjectLoad":true`, `"disableReferencedProjectLoad":false`))
		files[tspath.ResolvePath(directory, name)] = string(content)
		if strings.HasPrefix(name, "shared/") {
			files[tspath.ResolvePath(directory, "SHARED/"+strings.TrimPrefix(name, "shared/"))] = string(content)
		}
	}
	server := newTestServer()
	server.cwd = directory
	server.fs = &exactCaseLSPProgramFS{FS: bundled.WrapFS(osvfs.FS()), files: files}
	store := newLintProgramStore(server)
	server.lintPrograms = store
	store.coverage.watchFiles = func(context.Context, project.WatcherID, []*lsproto.FileSystemWatcher) error { return nil }
	load := func(relativePath string) selectedLintProject {
		fileName := tspath.ResolvePath(directory, relativePath)
		request := store.request(context.Background(), documentURIFromPath(fileName), lspConfigTarget(fileName, directory, server.fs), true)
		defer request.finalize()
		selected, err := selectLintProgramRequestForTest(request, directory)
		if err != nil {
			t.Fatal(err)
		}
		return selected
	}
	first := load("SHARED/first.ts")
	second := load("app/second.ts")
	if second.program != first.program {
		t.Fatal("reference spelling changed the identity of a case-insensitive configured project")
	}
	fileName := tspath.ResolvePath(directory, "app/second.ts")
	target := lspConfigTarget(fileName, directory, server.fs)
	request := newStandaloneLintProjectRequestWithFS(target, server.fs)
	request.sourceReferences = true
	selected, _, err := selectConfiguredLintProject(nil, directory, target, server.fs, request.loaders())
	if err != nil || lintProgramLexicalPathID(selected.configPath, server.fs) != lintProgramLexicalPathID(first.configPath, server.fs) {
		t.Fatalf("speculative case-insensitive reference selection=%s error=%v", selected.configPath, err)
	}
}

func TestLintProgramStorePersistsWatcherProtectedProjectMetadata(t *testing.T) {
	fixture := newLintProgramStoreFixture(t, "export const value = 1;\n")
	_, loadMetadata, finalize := fixture.request(fixture.sourceURI)
	metadata, available, err := loadMetadata(fixture.configPath)
	if err != nil {
		t.Fatalf("load metadata: %v", err)
	}
	if !available {
		t.Fatal("project metadata was unavailable")
	}
	finalize()
	if metadata == nil || !metadata.Contains(fixture.sourcePath, fixture.sourcePath) {
		t.Fatalf("project metadata did not contain the configured source: %v", metadata)
	}
	_, loadMetadata, finalize = fixture.request(fixture.sourceURI)
	reused, available, err := loadMetadata(fixture.configPath)
	if err != nil {
		t.Fatalf("reuse metadata: %v", err)
	}
	if !available {
		t.Fatal("reused project metadata was unavailable")
	}
	finalize()
	if reused != metadata {
		t.Fatal("unchanged watcher-protected root metadata was reparsed")
	}
	if fixture.watchCalls == 0 {
		t.Fatal("resident root metadata has no watcher coverage")
	}
	if !fixture.store.Invalidate() {
		t.Fatal("a project-root change would not trigger diagnostics refresh")
	}

	if err := os.WriteFile(fixture.configPath, []byte(`{"files":[]}`), 0o644); err != nil {
		t.Fatalf("rewrite config: %v", err)
	}
	_, loadMetadata, finalize = fixture.request(fixture.sourceURI)
	metadata, available, err = loadMetadata(fixture.configPath)
	if err != nil {
		t.Fatalf("reload metadata: %v", err)
	}
	if !available {
		t.Fatal("reloaded project metadata was unavailable")
	}
	finalize()
	if metadata == nil || metadata.Contains(fixture.sourcePath, fixture.sourcePath) {
		t.Fatalf("invalidated project metadata leaked across requests: %v", metadata)
	}
}

func TestLintProgramStoreOpeningNewIncludedFileInvalidatesProjectMetadata(t *testing.T) {
	const content = "export const value = 1;\n"
	fixture := newLintProgramStoreFixture(t, content)
	_, loadMetadata, finalize := fixture.request(fixture.sourceURI)
	before, available, err := loadMetadata(fixture.configPath)
	if err != nil {
		t.Fatal(err)
	}
	if !available {
		t.Fatal("initial project metadata was unavailable")
	}
	finalize()

	newPath := filepath.Join(filepath.Dir(fixture.sourcePath), "new.ts")
	if err := os.WriteFile(newPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	newURI := documentURIFromPath(newPath)
	fixture.server.documents[newURI] = content
	fixture.store.DidOpen(newURI, content, true)
	if len(fixture.store.projectMetadata) != 0 {
		t.Fatal("newly included source retained stale project metadata")
	}

	_, loadMetadata, finalize = fixture.request(newURI)
	after, available, err := loadMetadata(fixture.configPath)
	if err != nil {
		t.Fatal(err)
	}
	if !available {
		t.Fatal("refreshed project metadata was unavailable")
	}
	finalize()
	if after == before || !after.Contains(newPath, "") {
		t.Fatal("newly included source was absent from refreshed project metadata")
	}
}

func TestLintProgramStoreDoesNotRetainNonContainingFallbackProgram(t *testing.T) {
	fixture := newLintProgramStoreFixture(t, "export const value = 1;\n")
	outsidePath := filepath.Join(filepath.Dir(filepath.Dir(fixture.sourcePath)), "outside.ts")
	const outsideContent = "export const outside = 1;\n"
	if err := os.WriteFile(outsidePath, []byte(outsideContent), 0o644); err != nil {
		t.Fatal(err)
	}
	outsideURI := documentURIFromPath(outsidePath)
	fixture.server.documents[outsideURI] = outsideContent

	loadProgram, loadMetadata, finalize := fixture.request(outsideURI)
	metadata, available, err := loadMetadata(fixture.configPath)
	if err != nil {
		t.Fatalf("load metadata: %v", err)
	}
	if !available {
		t.Fatal("project metadata was unavailable")
	}
	if metadata.Contains(outsidePath, "") {
		t.Fatal("outside target unexpectedly became a direct project root")
	}
	_, sourceFile, err := loadProgram(fixture.configPath)
	if err != nil {
		t.Fatalf("probe fallback Program: %v", err)
	}
	finalize()
	if sourceFile != nil {
		t.Fatalf("outside target unexpectedly entered the Program: %v", sourceFile)
	}
	if len(fixture.store.programs) != 0 {
		t.Fatal("non-containing fallback Program became resident")
	}
	if len(fixture.store.projectMetadata) != 1 {
		t.Fatalf("lightweight project metadata was not retained: %d", len(fixture.store.projectMetadata))
	}
}

func TestLintProgramStoreDoesNotRetainProgramWithTransientProjectMetadata(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/project_service.txtar")
	for _, transientReference := range []bool{false, true} {
		name := "root"
		if transientReference {
			name = "reference"
		}
		t.Run(name, func(t *testing.T) {
			directory := tspath.NormalizePath(archive.Materialize(t, "snapshot-references"))
			configPath := tspath.ResolvePath(directory, "app/tsconfig.json")
			refPath := tspath.ResolvePath(directory, "lib/tsconfig.json")
			server := newTestServer()
			server.cwd = tspath.ResolvePath(directory, "app")
			server.fs = bundled.WrapFS(osvfs.FS())
			store := newLintProgramStore(server)
			store.coverage.watchFiles = func(context.Context, project.WatcherID, []*lsproto.FileSystemWatcher) error { return nil }
			metadata, err := parseStandaloneLintProject(configPath, server.fs, server.fs)
			if err != nil {
				t.Fatal(err)
			}
			reference, err := parseStandaloneLintProject(refPath, server.fs, server.fs)
			if err != nil {
				t.Fatal(err)
			}
			target := lspConfigTarget(tspath.ResolvePath(directory, "app/target.ts"), server.cwd, server.fs)
			request := store.request(context.Background(), documentURIFromPath(target.Path), target, true)
			request.projectMetadata[configPath] = metadata
			request.projectMetadata[refPath] = reference
			transientPath := configPath
			if transientReference {
				transientPath = refPath
			}
			request.transientMetadata[transientPath] = struct{}{}
			if _, sourceFile, err := request.load(metadata); err != nil || sourceFile == nil {
				t.Fatalf("transient Program source=%v error=%v", sourceFile, err)
			}
			request.finalize()
			if len(store.programs) != 0 {
				t.Fatal("Program retained metadata predating stable watcher coverage")
			}
		})
	}
}

func TestLintProgramStoreProjectServiceTracksMissingReferences(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/project_service.txtar")
	directory := tspath.NormalizePath(archive.Materialize(t, "snapshot-references"))
	rootPath := tspath.ResolvePath(directory, "app/tsconfig.json")
	refPath := tspath.ResolvePath(directory, "lib/tsconfig.json")
	reference, err := os.ReadFile(refPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(refPath); err != nil {
		t.Fatal(err)
	}
	server := newTestServer()
	server.cwd = tspath.ResolvePath(directory, "app")
	server.fs = bundled.WrapFS(osvfs.FS())
	store := newLintProgramStore(server)
	store.coverage.watchFiles = func(context.Context, project.WatcherID, []*lsproto.FileSystemWatcher) error { return nil }
	target := lspConfigTarget(tspath.ResolvePath(directory, "app/target.ts"), server.cwd, server.fs)
	uri := documentURIFromPath(target.Path)
	load := func() (*compiler.Program, *lintProgramState) {
		t.Helper()
		request := store.request(context.Background(), uri, target, true)
		selected, err := selectLintProgramRequestForTest(request, server.cwd)
		if err != nil || selected.sourceFile == nil {
			t.Fatalf("selected source=%v error=%v", selected.sourceFile, err)
		}
		request.finalize()
		return selected.program, store.programs[request.key(rootPath)]
	}
	first, state := load()
	if state == nil {
		t.Fatal("missing reference prevented resident Program creation")
	}
	if _, tracked := state.failedLookups[lintProgramLexicalPathID(refPath, server.fs)]; !tracked {
		t.Fatal("missing external reference bypassed failed-lookup tracking")
	}
	if err := os.WriteFile(refPath, reference, 0o644); err != nil {
		t.Fatal(err)
	}
	// Discovery sees the new reference before its watcher event arrives. Its
	// snapshot must constrain the resident Program too, with no Session help.
	second, _ := load()
	if second == first {
		t.Fatal("new reference reused the Program from its missing-config generation")
	}
	found := false
	second.RangeResolvedProjectReference(func(_ tspath.Path, parsed, _ *tsoptions.ParsedCommandLine, _ int) bool {
		if parsed != nil && lintProgramLexicalPathID(parsed.ConfigName(), server.fs) == lintProgramLexicalPathID(refPath, server.fs) {
			found = true
		}
		return true
	})
	if !found {
		t.Fatal("created reference did not enter the rebuilt Program")
	}
}

func TestLintProgramStoreRebuildsForGraphChange(t *testing.T) {
	const initial = "export const value = 1;\n"
	fixture := newLintProgramStoreFixture(t, initial)
	first := fixture.load(t)

	dependencyPath := filepath.Join(filepath.Dir(fixture.sourcePath), "dependency.ts")
	if err := os.WriteFile(
		dependencyPath,
		[]byte("export const dependency = 1;\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	const changed = "import { dependency } from './dependency';\nexport const value = dependency;\n"
	fixture.server.documents[fixture.sourceURI] = changed
	fixture.store.DidChange(fixture.sourceURI, changed)

	updated := fixture.load(t)
	if updated == first {
		t.Fatal("graph-changing edit retained the old Program")
	}
	if source := updated.GetSourceFile(
		tspath.NormalizePath(dependencyPath),
	); source == nil {
		t.Fatal("graph-changing edit did not load the new dependency")
	}

	const changedDependency = "export const dependency = 2;\n"
	if err := os.WriteFile(dependencyPath, []byte(changedDependency), 0o644); err != nil {
		t.Fatal(err)
	}
	fixture.store.DidChangeWatchedFiles([]*lsproto.FileEvent{{
		Uri:  documentURIFromPath(dependencyPath),
		Type: lsproto.FileChangeTypeChanged,
	}})
	reloaded := fixture.load(t)
	dependency := reloaded.GetSourceFile(tspath.NormalizePath(dependencyPath))
	if dependency == nil || dependency.Text() != changedDependency {
		t.Fatal("watched dependency change was not visible after rebuild")
	}
}

func TestLintProgramStoreReopenSameContentStaysWarm(t *testing.T) {
	const content = "export const value = 1;\n"
	fixture := newLintProgramStoreFixture(t, content)
	first := fixture.load(t)

	delete(fixture.server.documents, fixture.sourceURI)
	fixture.store.DidClose(fixture.sourceURI)
	fixture.server.documents[fixture.sourceURI] = content
	fixture.store.DidOpen(fixture.sourceURI, content, true)

	if reopened := fixture.load(t); reopened != first {
		t.Fatal("close and reopen with unchanged content rebuilt the Program")
	}
}

func TestLintProgramStoreUnsavedFileSaveRemainsIncremental(t *testing.T) {
	const content = "export const value = 1;\n"
	fixture := newLintProgramStoreFixture(t, content)
	newPath := filepath.Join(filepath.Dir(fixture.sourcePath), "new.ts")
	newURI := documentURIFromPath(newPath)
	fixture.server.documents[newURI] = content
	fixture.store.DidOpen(newURI, content, false)

	loader, _, finalize := fixture.request(newURI)
	first, sourceFile, err := loader(fixture.configPath)
	if err != nil {
		t.Fatal(err)
	}
	finalize()
	if sourceFile == nil {
		t.Fatal("unsaved included file was not loaded from the editor overlay")
	}

	identityBeforeSave := lspFilesystemPathID(newPath, fixture.server.fs)
	if err := os.WriteFile(newPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	identityAfterSave := lspFilesystemPathID(newPath, fixture.server.fs)
	fixture.store.DidSave(newURI, true)
	loader, _, finalize = fixture.request(newURI)
	saved, _, err := loader(fixture.configPath)
	if err != nil {
		t.Fatal(err)
	}
	finalize()
	if identityBeforeSave == identityAfterSave && saved != first {
		t.Fatal("saving an unchanged open file rebuilt its Program")
	}

	const changed = "export const value = 2;\n"
	fixture.server.documents[newURI] = changed
	fixture.store.DidChange(newURI, changed)
	loader, _, finalize = fixture.request(newURI)
	updated, updatedSource, err := loader(fixture.configPath)
	if err != nil {
		t.Fatal(err)
	}
	finalize()
	if updated == saved || updatedSource == nil || updatedSource.Text() != changed {
		t.Fatal("saved file did not remain incrementally updateable")
	}
}

func TestLintProgramStoreUnsavedFileCloseInvalidates(t *testing.T) {
	const content = "export const value = 1;\n"
	fixture := newLintProgramStoreFixture(t, content)
	newPath := filepath.Join(filepath.Dir(fixture.sourcePath), "new.ts")
	newURI := documentURIFromPath(newPath)
	fixture.server.documents[newURI] = content
	fixture.store.DidOpen(newURI, content, false)

	loader, _, finalize := fixture.request(newURI)
	if _, sourceFile, err := loader(fixture.configPath); err != nil {
		t.Fatal(err)
	} else if sourceFile == nil {
		t.Fatal("unsaved included file was not loaded from the editor overlay")
	}
	finalize()
	delete(fixture.server.documents, newURI)
	fixture.store.DidClose(newURI)
	if len(fixture.store.programs) != 0 {
		t.Fatal("closing an unsaved file retained a Program built with its overlay")
	}
}

func TestLintProgramStoreUnsavedConfigChangeRebuilds(t *testing.T) {
	const content = "export const value = 1;\n"
	fixture := newLintProgramStoreFixture(t, content)
	first := fixture.load(t)

	configURI := documentURIFromPath(fixture.configPath)
	const changedConfig = `{"compilerOptions":{"noLib":true},"files":[]}`
	fixture.server.documents[configURI] = changedConfig
	fixture.store.DidChange(configURI, changedConfig)
	if len(fixture.store.programs) != 0 {
		t.Fatal("unsaved tsconfig change retained a dependent Program")
	}
	if rebuilt := fixture.load(t); rebuilt == first {
		t.Fatal("unsaved tsconfig change did not rebuild the Program")
	}
}

func TestLintProgramStoreWatchedChangesInvalidateOnlyWhenNeeded(t *testing.T) {
	const content = "export const value = 1;\n"
	fixture := newLintProgramStoreFixture(t, content)
	first := fixture.load(t)

	fixture.store.DidChangeWatchedFiles([]*lsproto.FileEvent{{
		Uri:  fixture.sourceURI,
		Type: lsproto.FileChangeTypeChanged,
	}})
	if afterSave := fixture.load(t); afterSave != first {
		t.Fatal("open-file save discarded the resident Program")
	}

	dependencyURI := documentURIFromPath(filepath.Join(filepath.Dir(fixture.sourcePath), "dependency.ts"))
	fixture.store.DidChangeWatchedFiles([]*lsproto.FileEvent{{
		Uri:  dependencyURI,
		Type: lsproto.FileChangeTypeChanged,
	}})
	if len(fixture.store.programs) != 0 {
		t.Fatal("external filesystem change retained a resident Program")
	}
	if afterDependencyChange := fixture.load(t); afterDependencyChange == first {
		t.Fatal("external filesystem change retained the old Program")
	}
}

func TestLintProgramStoreWatchedSymlinkSourceChangeRebuilds(t *testing.T) {
	const content = "export const value = 1;\n"
	fixture := newLintProgramStoreFixture(t, content)
	firstTarget := filepath.Join(filepath.Dir(fixture.sourcePath), "first.ts")
	if err := os.Rename(fixture.sourcePath, firstTarget); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(firstTarget, fixture.sourcePath); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	first := fixture.load(t)

	secondTarget := filepath.Join(filepath.Dir(fixture.sourcePath), "second.ts")
	if err := os.WriteFile(secondTarget, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(fixture.sourcePath); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(secondTarget, fixture.sourcePath); err != nil {
		t.Fatal(err)
	}
	fixture.store.DidChangeWatchedFiles([]*lsproto.FileEvent{{
		Uri:  fixture.sourceURI,
		Type: lsproto.FileChangeTypeChanged,
	}})
	if len(fixture.store.programs) != 0 {
		t.Fatal("watched symlink source change retained a generation-bound Program")
	}
	if rebuilt := fixture.load(t); rebuilt == first {
		t.Fatal("watched symlink source change retained the old Program")
	}
}

func TestLintProgramStoreWatchedChangeRefreshesCustomProjectDiagnostics(t *testing.T) {
	for _, loaded := range []bool{false, true} {
		t.Run(strconv.FormatBool(loaded), func(t *testing.T) {
			fixture := newLintProgramStoreFixture(t, "export const value = 1;\n")
			if loaded {
				fixture.load(t)
			}
			fixture.server.lintPrograms = fixture.store
			// A missing explicit path/glob can fail policy resolution before a
			// Program is loaded. A delivered creation event must still relint.
			customURI := documentURIFromPath(tspath.ResolvePath(fixture.server.cwd, "custom.json"))
			if err := fixture.server.handleDidChangeWatchedFiles(context.Background(), &lsproto.DidChangeWatchedFilesParams{
				Changes: []*lsproto.FileEvent{{Uri: customURI, Type: lsproto.FileChangeTypeCreated}},
			}); err != nil {
				t.Fatal(err)
			}
			select {
			case <-fixture.server.refreshCh:
			default:
				t.Fatal("custom-project watcher event did not schedule diagnostics")
			}
			delete(fixture.server.documents, fixture.sourceURI)
			if fixture.store.DidChangeWatchedFiles([]*lsproto.FileEvent{{Uri: customURI, Type: lsproto.FileChangeTypeChanged}}) {
				t.Fatal("empty store without open documents requested diagnostics")
			}
		})
	}
}

func TestLintProgramStoreWatcherFailureFallsBackToFreshPrograms(t *testing.T) {
	const content = "export const value = 1;\n"
	fixture := newLintProgramStoreFixture(t, content)
	fixture.store.coverage.watchFiles = func(
		context.Context,
		project.WatcherID,
		[]*lsproto.FileSystemWatcher,
	) error {
		return errors.New("watch registration failed")
	}

	first := fixture.load(t)
	if fixture.store.Usable() {
		t.Fatal("store remained enabled without dependency coverage")
	}
	if len(fixture.store.programs) != 0 {
		t.Fatal("store retained a Program after watcher failure")
	}
	if second := fixture.load(t); second == first {
		t.Fatal("disabled store did not fall back to fresh construction")
	}
}

func TestLintProgramStoreOpeningUnrelatedFileKeepsResidentProgram(t *testing.T) {
	const content = "export const value = 1;\n"
	fixture := newLintProgramStoreFixture(t, content)
	first := fixture.load(t)

	unrelatedPath := filepath.Join(filepath.Dir(filepath.Dir(fixture.sourcePath)), "other.ts")
	if err := os.WriteFile(unrelatedPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	unrelatedURI := documentURIFromPath(unrelatedPath)
	fixture.server.documents[unrelatedURI] = content
	fixture.store.DidOpen(unrelatedURI, content, true)

	if len(fixture.store.programs) != 1 {
		t.Fatal("opening an unrelated existing file discarded a resident Program")
	}
	if afterOpen := fixture.load(t); afterOpen != first {
		t.Fatal("opening an unrelated existing file rebuilt a resident Program")
	}
}

func TestLintProgramStoreOpeningNewIncludedFileRebuilds(t *testing.T) {
	const content = "export const value = 1;\n"
	fixture := newLintProgramStoreFixture(t, content)
	first := fixture.load(t)

	newPath := filepath.Join(filepath.Dir(fixture.sourcePath), "new.ts")
	if err := os.WriteFile(newPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	newURI := documentURIFromPath(newPath)
	fixture.server.documents[newURI] = content
	fixture.store.DidOpen(newURI, content, true)
	if len(fixture.store.programs) != 0 {
		t.Fatal("newly included source retained a Program built before the file existed")
	}

	loader, _, finalize := fixture.request(newURI)
	defer finalize()
	rebuilt, sourceFile, err := loader(fixture.configPath)
	if err != nil {
		t.Fatal(err)
	}
	if rebuilt == first {
		t.Fatal("newly included source did not rebuild the Program")
	}
	if sourceFile == nil || sourceFile.FileName() != tspath.NormalizePath(newPath) {
		t.Fatalf("newly included source missing from rebuilt Program: %v", sourceFile)
	}
}

func TestLintProgramStoreOpeningNewImportedFileRebuildsBeforeWatchEvent(t *testing.T) {
	const content = "import '../generated/value';\nexport const value = 1;\n"
	fixture := newLintProgramStoreFixture(t, content)
	if err := os.WriteFile(
		fixture.configPath,
		[]byte(`{"compilerOptions":{"noLib":true},"files":["src/index.ts"]}`),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	first := fixture.load(t)

	importedPath := filepath.Join(
		filepath.Dir(filepath.Dir(fixture.sourcePath)),
		"generated",
		"value.ts",
	)
	if err := os.MkdirAll(filepath.Dir(importedPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		importedPath,
		[]byte("export const generated = 1;\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	importedURI := documentURIFromPath(importedPath)
	fixture.server.documents[importedURI] = "export const generated = 1;\n"
	fixture.store.DidOpen(importedURI, fixture.server.documents[importedURI], true)
	if len(fixture.store.programs) != 0 {
		t.Fatal("newly resolved import retained a Program built while it was missing")
	}

	loader, _, finalize := fixture.request(importedURI)
	defer finalize()
	rebuilt, sourceFile, err := loader(fixture.configPath)
	if err != nil {
		t.Fatal(err)
	}
	if rebuilt == first {
		t.Fatal("newly resolved import did not rebuild the Program")
	}
	if sourceFile == nil ||
		sourceFile.FileName() != tspath.NormalizePath(importedPath) {
		t.Fatalf("newly resolved import missing from rebuilt Program: %v", sourceFile)
	}
}

func TestLintProgramStoreConflictingAliasesUseFreshPrograms(t *testing.T) {
	const content = "export const value = 1;\n"
	fixture := newLintProgramStoreFixture(t, content)
	fixture.load(t)

	aliasPath := filepath.Join(filepath.Dir(fixture.sourcePath), "alias.ts")
	if err := os.Symlink(fixture.sourcePath, aliasPath); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	aliasURI := documentURIFromPath(aliasPath)
	fixture.server.documents[aliasURI] = "export const value = 2;\n"

	load := func() *compiler.Program {
		t.Helper()
		loader, _, finalize := fixture.request(fixture.sourceURI)
		defer finalize()
		program, _, err := loader(fixture.configPath)
		if err != nil {
			t.Fatal(err)
		}
		return program
	}
	first := load()
	if len(fixture.store.programs) != 0 {
		t.Fatal("conflicting alias buffers retained a resident Program")
	}
	if second := load(); second == first {
		t.Fatal("conflicting alias buffers did not use fresh Programs")
	}
}

type realpathCountingFS struct {
	vfs.FS
	calls int
}

func (fs *realpathCountingFS) Realpath(path string) string {
	fs.calls++
	return fs.FS.Realpath(path)
}

type retargetingLintProgramFS struct {
	vfs.FS
	targetPath string
	firstPath  string
	laterPath  string
	targetCall int
}

func (fs *retargetingLintProgramFS) Realpath(filePath string) string {
	if tspath.NormalizePath(filePath) != fs.targetPath {
		return fs.FS.Realpath(filePath)
	}
	fs.targetCall++
	if fs.targetCall == 1 {
		return fs.firstPath
	}
	return fs.laterPath
}

func TestLintProgramStoreReusesFrozenTargetForResidentAndRebuild(t *testing.T) {
	const content = "export const value = 1;\n"
	fixture := newLintProgramStoreFixture(t, content)
	fixture.load(t)
	canonicalPath := tspath.NormalizePath(fixture.server.fs.Realpath(fixture.sourcePath))
	laterPath := tspath.NormalizePath(filepath.Join(filepath.Dir(fixture.sourcePath), "moved.ts"))
	if err := os.WriteFile(laterPath, []byte("export const moved = 2;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	retargetingFS := &retargetingLintProgramFS{
		FS:         fixture.server.fs,
		targetPath: fixture.sourcePath,
		firstPath:  canonicalPath,
		laterPath:  laterPath,
	}
	fixture.server.fs = retargetingFS
	target := lspConfigTarget(fixture.sourcePath, fixture.server.cwd, retargetingFS)

	load := func() *ast.SourceFile {
		t.Helper()
		loader, loadMetadata, finalize := fixture.store.Request(
			context.Background(),
			fixture.sourceURI,
			target,
		)
		defer finalize()
		metadata, _, err := loadMetadata(fixture.configPath)
		if err != nil {
			t.Fatal(err)
		}
		_, sourceFile, err := loader(metadata)
		if err != nil {
			t.Fatal(err)
		}
		return sourceFile
	}
	if sourceFile := load(); sourceFile == nil || sourceFile.Text() != content {
		t.Fatalf("resident Program source = %v", sourceFile)
	}
	fixture.store.Invalidate()
	if sourceFile := load(); sourceFile == nil || sourceFile.Text() != content {
		t.Fatalf("rebuilt Program source = %v", sourceFile)
	}
	if retargetingFS.targetCall != 1 {
		t.Fatalf("target Realpath calls = %d, want one frozen observation", retargetingFS.targetCall)
	}
}

func TestLintProgramStoreUnrelatedOpenDoesNotScanProgramSources(t *testing.T) {
	const content = "export const value = 1;\n"
	fixture := newLintProgramStoreFixture(t, content)
	for index := range 100 {
		path := filepath.Join(
			filepath.Dir(fixture.sourcePath),
			"file-"+strconv.Itoa(index)+".ts",
		)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	countingFS := &realpathCountingFS{FS: fixture.server.fs}
	fixture.server.fs = countingFS
	fixture.load(t)

	unrelatedPath := filepath.Join(filepath.Dir(filepath.Dir(fixture.sourcePath)), "outside.ts")
	if err := os.WriteFile(unrelatedPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	countingFS.calls = 0
	unrelatedURI := documentURIFromPath(unrelatedPath)
	fixture.server.documents[unrelatedURI] = content
	fixture.store.DidOpen(unrelatedURI, content, true)
	if countingFS.calls > 3 {
		t.Fatalf("unrelated open performed %d realpath calls; want O(1)", countingFS.calls)
	}
}

func TestLintProgramStoreWatchesExternalEmptyIncludeDirectory(t *testing.T) {
	root := t.TempDir()
	workspace := filepath.Join(root, "app")
	shared := filepath.Join(root, "shared")
	sourcePath := filepath.Join(workspace, "src", "index.ts")
	for _, directory := range []string{filepath.Dir(sourcePath), shared} {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	const content = "export const value = 1;\n"
	if err := os.WriteFile(sourcePath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(workspace, "tsconfig.json")
	if err := os.WriteFile(
		configPath,
		[]byte(`{"compilerOptions":{"noLib":true},"include":["src/**/*.ts","../shared/**/*.d.ts"]}`),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	server := newTestServer()
	server.cwd = workspace
	server.fs = bundled.WrapFS(osvfs.FS())
	server.defaultLibraryPath = bundled.LibPath()
	server.initializeParams = &lsproto.InitializeParams{}
	sourceURI := documentURIFromPath(sourcePath)
	server.documents[sourceURI] = content
	store := newLintProgramStore(server)
	var watcherGlobs []string
	store.coverage.watchFiles = func(
		_ context.Context,
		_ project.WatcherID,
		watchers []*lsproto.FileSystemWatcher,
	) error {
		for _, watcher := range watchers {
			watcherGlobs = append(
				watcherGlobs,
				project.FileSystemWatcherGlobString(watcher),
			)
		}
		return nil
	}

	loader, loadMetadata, finalize := store.Request(
		context.Background(),
		sourceURI,
		lspConfigTarget(sourcePath, workspace, server.fs),
	)
	metadata, _, err := loadMetadata(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := loader(metadata); err != nil {
		t.Fatal(err)
	}
	finalize()
	if realPath := server.fs.Realpath(shared); realPath != "" {
		shared = realPath
	}
	shared = strings.ToLower(filepath.ToSlash(shared))
	for _, glob := range watcherGlobs {
		watchedRoot := strings.TrimSuffix(
			strings.ToLower(glob),
			"/**/*",
		)
		if shared == watchedRoot || strings.HasPrefix(shared, watchedRoot+"/") {
			return
		}
	}
	t.Fatalf("external empty include directory %q is not covered by %v", shared, watcherGlobs)
}
