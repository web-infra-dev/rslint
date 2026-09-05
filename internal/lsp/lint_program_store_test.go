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
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
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
) (lintProgramLoader, lintProjectMetadataLoader, func()) {
	return f.store.Request(
		context.Background(),
		uri,
		lspConfigTarget(uriToPath(uri), f.server.cwd, f.server.fs),
	)
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
	fixture := newLintProgramStoreFixture(t, "export const value = 1;\n")
	metadata, err := parseStandaloneLintProject(
		fixture.configPath,
		fixture.server.fs,
		fixture.server.fs,
	)
	if err != nil {
		t.Fatal(err)
	}
	configPath := tspath.NormalizePath(fixture.configPath)
	request := &lintProgramRequest{
		store: fixture.store,
		ctx:   context.Background(),
		uri:   fixture.sourceURI,
		target: lspConfigTarget(
			fixture.sourcePath,
			fixture.server.cwd,
			fixture.server.fs,
		),
		projectMetadata: map[string]*lintProjectMetadata{
			configPath: metadata,
		},
		transientMetadata: map[string]struct{}{
			configPath: {},
		},
	}

	if _, sourceFile, err := request.load(configPath); err != nil {
		t.Fatal(err)
	} else if sourceFile == nil {
		t.Fatal("transient Program did not contain its direct root")
	}
	request.finalize()
	if len(fixture.store.programs) != 0 {
		t.Fatal("Program built from metadata predating watcher coverage became resident")
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
	const content = "export const value = 1;\n"
	fixture := newLintProgramStoreFixture(t, content)
	fixture.load(t)
	fixture.server.lintPrograms = fixture.store

	dependencyURI := documentURIFromPath(
		filepath.Join(filepath.Dir(fixture.sourcePath), "dependency.ts"),
	)
	if err := fixture.server.handleDidChangeWatchedFiles(
		context.Background(),
		&lsproto.DidChangeWatchedFilesParams{
			Changes: []*lsproto.FileEvent{{
				Uri:  dependencyURI,
				Type: lsproto.FileChangeTypeChanged,
			}},
		},
	); err != nil {
		t.Fatal(err)
	}
	select {
	case <-fixture.server.refreshCh:
	default:
		t.Fatal("custom-project watcher invalidation did not schedule diagnostics")
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
		loader, _, finalize := fixture.store.Request(
			context.Background(),
			fixture.sourceURI,
			target,
		)
		defer finalize()
		_, sourceFile, err := loader(fixture.configPath)
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

	loader, _, finalize := store.Request(
		context.Background(),
		sourceURI,
		lspConfigTarget(sourcePath, workspace, server.fs),
	)
	if _, _, err := loader(configPath); err != nil {
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
