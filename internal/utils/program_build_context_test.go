package utils

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tsoptions"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
)

type programMetadataTestFS struct {
	vfs.FS
	mu       sync.Mutex
	contents map[string]string
	reads    map[string]int
}

func newProgramMetadataTestFS(contents map[string]string) *programMetadataTestFS {
	return &programMetadataTestFS{
		FS:       bundled.WrapFS(osvfs.FS()),
		contents: contents,
		reads:    make(map[string]int),
	}
}

func (f *programMetadataTestFS) ReadFile(path string) (string, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.reads[path]++
	content, ok := f.contents[path]
	return content, ok
}

func (f *programMetadataTestFS) WriteFile(path string, content string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.contents[path] = content
	return nil
}

func (f *programMetadataTestFS) set(path string, content string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.contents[path] = content
}

func (f *programMetadataTestFS) readCount(path string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.reads[path]
}

func TestProgramMetadataFSSelectiveExactPathCaching(t *testing.T) {
	const (
		pkg          = "/repo/package.json"
		pkgAlias     = "/repo/./package.json"
		upperPkg     = "/repo/PACKAGE.JSON"
		windowsPkg   = `C:\repo\package.json`
		source       = "/repo/source.ts"
		customConfig = "/repo/config.shared.json"
	)
	base := newProgramMetadataTestFS(map[string]string{
		pkg:          "",
		pkgAlias:     `{"alias":true}`,
		upperPkg:     `{"upper":true}`,
		windowsPkg:   `{"windows":true}`,
		source:       "export const value = 1",
		customConfig: `{"compilerOptions":{}}`,
	})
	fsys := newProgramMetadataFS(base)

	for range 2 {
		if content, ok := fsys.ReadFile(pkg); !ok || content != "" {
			t.Fatalf("empty package.json is a successful cached read: content=%q ok=%v", content, ok)
		}
		if _, ok := fsys.ReadFile(pkgAlias); !ok {
			t.Fatal("expected lexical package.json alias")
		}
		if _, ok := fsys.ReadFile(windowsPkg); !ok {
			t.Fatal("expected Windows-style package.json path")
		}
		if _, ok := fsys.ReadFile(upperPkg); !ok {
			t.Fatal("expected upper-case JSON fixture")
		}
		if _, ok := fsys.ReadFile(source); !ok {
			t.Fatal("expected source fixture")
		}
		if _, ok := fsys.ReadFile(customConfig); !ok {
			t.Fatal("expected unregistered JSON fixture")
		}
	}

	if got := base.readCount(pkg); got != 1 {
		t.Fatalf("package.json read count = %d, want 1", got)
	}
	if got := base.readCount(pkgAlias); got != 1 {
		t.Fatalf("lexical alias must have its own cache key; reads = %d, want 1", got)
	}
	if got := base.readCount(windowsPkg); got != 1 {
		t.Fatalf("Windows package.json read count = %d, want 1", got)
	}
	if got := base.readCount(upperPkg); got != 2 {
		t.Fatalf("non-exact PACKAGE.JSON must bypass cache; reads = %d, want 2", got)
	}
	if got := base.readCount(source); got != 2 {
		t.Fatalf("source reads must bypass metadata cache; reads = %d, want 2", got)
	}
	if got := base.readCount(customConfig); got != 2 {
		t.Fatalf("unregistered JSON reads = %d, want 2", got)
	}

	fsys.registerTSConfig(customConfig)
	for range 2 {
		if _, ok := fsys.ReadFile(customConfig); !ok {
			t.Fatal("expected registered config fixture")
		}
	}
	if got := base.readCount(customConfig); got != 3 {
		t.Fatalf("registered config should snapshot its next successful read; reads = %d, want 3", got)
	}
}

func TestProgramMetadataFSFailureRetryAndWriteInvalidation(t *testing.T) {
	const path = "/repo/package.json"
	base := newProgramMetadataTestFS(map[string]string{})
	fsys := newProgramMetadataFS(base)

	if _, ok := fsys.ReadFile(path); ok {
		t.Fatal("missing package.json unexpectedly succeeded")
	}
	base.set(path, `{"version":1}`)
	if content, ok := fsys.ReadFile(path); !ok || content != `{"version":1}` {
		t.Fatalf("failed read was cached: content=%q ok=%v", content, ok)
	}
	if got := base.readCount(path); got != 2 {
		t.Fatalf("failed reads must be retried; reads = %d, want 2", got)
	}

	if err := fsys.WriteFile(path, `{"version":2}`); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if content, ok := fsys.ReadFile(path); !ok || content != `{"version":2}` {
		t.Fatalf("write did not replace metadata generation: content=%q ok=%v", content, ok)
	}
	if got := base.readCount(path); got != 3 {
		t.Fatalf("post-write read count = %d, want 3", got)
	}
}

type blockingProgramMetadataFS struct {
	vfs.FS
	mu      sync.Mutex
	content string
	reads   int
	entered chan struct{}
	release chan struct{}
}

func (f *blockingProgramMetadataFS) ReadFile(string) (string, bool) {
	f.mu.Lock()
	f.reads++
	readNumber := f.reads
	content := f.content
	f.mu.Unlock()
	if readNumber == 1 {
		close(f.entered)
		<-f.release
	}
	return content, true
}

func (f *blockingProgramMetadataFS) set(content string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.content = content
}

func (f *blockingProgramMetadataFS) readCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.reads
}

func TestProgramMetadataFSConcurrentSingleFlightAndGenerationIsolation(t *testing.T) {
	const path = "/repo/package.json"
	base := &blockingProgramMetadataFS{
		FS:      bundled.WrapFS(osvfs.FS()),
		content: "old",
		entered: make(chan struct{}),
		release: make(chan struct{}),
	}
	fsys := newProgramMetadataFS(base)

	type result struct {
		content string
		ok      bool
	}
	first := make(chan result, 1)
	go func() {
		content, ok := fsys.ReadFile(path)
		first <- result{content: content, ok: ok}
	}()
	<-base.entered

	const waiters = 24
	waiterResults := make(chan result, waiters)
	for range waiters {
		go func() {
			content, ok := fsys.ReadFile(path)
			waiterResults <- result{content: content, ok: ok}
		}()
	}

	// An invalidation does not mutate the captured old generation. A new read
	// can complete against new bytes while the old single-flight is in flight.
	base.set("new")
	fsys.invalidate()
	if content, ok := fsys.ReadFile(path); !ok || content != "new" {
		t.Fatalf("new generation read: content=%q ok=%v", content, ok)
	}
	close(base.release)

	if got := <-first; !got.ok || got.content != "old" {
		t.Fatalf("in-flight old generation = %+v, want old successful snapshot", got)
	}
	for range waiters {
		if got := <-waiterResults; !got.ok || (got.content != "old" && got.content != "new") {
			t.Fatalf("concurrent result = %+v", got)
		}
	}
	if got := base.readCount(); got != 2 {
		t.Fatalf("one underlying read per generation = %d, want 2", got)
	}
	if content, _ := fsys.ReadFile(path); content != "new" {
		t.Fatalf("current generation regressed to %q", content)
	}
}

type panickingProgramMetadataFS struct {
	vfs.FS
}

func (f *panickingProgramMetadataFS) ReadFile(string) (string, bool) {
	panic("metadata read panic")
}

func TestProgramMetadataFSPanicReleasesEntry(t *testing.T) {
	const path = "/repo/package.json"
	fsys := newProgramMetadataFS(&panickingProgramMetadataFS{FS: bundled.WrapFS(osvfs.FS())})
	generation := fsys.currentGeneration()
	read := &programMetadataRead{ready: make(chan struct{})}
	generation.entries.Store(path, read)

	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("expected underlying VFS panic")
			}
		}()
		_, _ = fsys.readAndPublish(generation, path, read)
	}()

	select {
	case <-read.ready:
	default:
		t.Fatal("panic stranded concurrent metadata readers")
	}
	if _, loaded := generation.entries.Load(path); loaded {
		t.Fatal("panicking read remained cached")
	}
}

type programReadCountingFS struct {
	vfs.FS
	mu    sync.Mutex
	reads map[string]int
}

func (f *programReadCountingFS) ReadFile(path string) (string, bool) {
	f.mu.Lock()
	f.reads[path]++
	f.mu.Unlock()
	return f.FS.ReadFile(path)
}

func (f *programReadCountingFS) readCount(path string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.reads[path]
}

func writeProgramBuildFixture(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("MkdirAll(%s): %v", path, err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("WriteFile(%s): %v", path, err)
		}
	}
}

func TestProgramBuildContextSharesExtendedConfigWithoutFreezingConfigDir(t *testing.T) {
	root := t.TempDir()
	writeProgramBuildFixture(t, root, map[string]string{
		"tsconfig.base.json": `{
			"compilerOptions": {
				"paths": { "@pkg/*": ["${configDir}/src/*"] }
			},
			"excludes": ["generated"]
		}`,
		"packages/a/tsconfig.json": `{"extends":"../../tsconfig.base.json","files":[]}`,
		"packages/b/tsconfig.json": `{"extends":"../../tsconfig.base.json","files":[]}`,
	})

	baseFS := &programReadCountingFS{
		FS:    bundled.WrapFS(cachedvfs.From(osvfs.FS())),
		reads: make(map[string]int),
	}
	context := NewProgramBuildContext(baseFS)
	_, configA, err := context.parseConfig(root, filepath.Join(root, "packages/a/tsconfig.json"))
	if err != nil {
		t.Fatalf("parse config A: %v", err)
	}
	_, configB, err := context.parseConfig(root, filepath.Join(root, "packages/b/tsconfig.json"))
	if err != nil {
		t.Fatalf("parse config B: %v", err)
	}

	basePath := tspath.NormalizePath(filepath.Join(root, "tsconfig.base.json"))
	if got := baseFS.readCount(basePath); got != 1 {
		t.Fatalf("shared extended config reads = %d, want 1", got)
	}
	if len(configA.Errors) == 0 || len(configB.Errors) == 0 {
		t.Fatalf("extended-config diagnostics must survive cache hits: A=%d B=%d", len(configA.Errors), len(configB.Errors))
	}
	wantA := tspath.NormalizePath(filepath.Join(root, "packages/a/src/*"))
	wantB := tspath.NormalizePath(filepath.Join(root, "packages/b/src/*"))
	if got := configA.CompilerOptions().Paths.GetOrZero("@pkg/*"); len(got) != 1 || got[0] != wantA {
		t.Fatalf("config A ${configDir} paths = %v, want [%s]", got, wantA)
	}
	if got := configB.CompilerOptions().Paths.GetOrZero("@pkg/*"); len(got) != 1 || got[0] != wantB {
		t.Fatalf("config B ${configDir} paths = %v, want [%s]", got, wantB)
	}
}

func TestProgramBuildContextExtendedConfigCrossCycleDoesNotDeadlock(t *testing.T) {
	root := t.TempDir()
	writeProgramBuildFixture(t, root, map[string]string{
		"root-a.json": `{"extends":"./a.json","files":[]}`,
		"root-b.json": `{"extends":"./b.json","files":[]}`,
		"a.json":      `{"extends":"./b.json"}`,
		"b.json":      `{"extends":"./a.json"}`,
	})

	context := NewProgramBuildContext(bundled.WrapFS(cachedvfs.From(osvfs.FS())))
	parse := func(name string, done chan<- struct{}) {
		host := context.NewCompilerHost(root)
		context.registerTSConfig(filepath.Join(root, name))
		_, _ = tsoptions.GetParsedCommandLineOfConfigFile(
			filepath.Join(root, name),
			nil,
			nil,
			host,
			context.extendedConfigCache,
		)
		done <- struct{}{}
	}

	done := make(chan struct{}, 2)
	go parse("root-a.json", done)
	go parse("root-b.json", done)
	for range 2 {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("concurrent cross-cycle config parsing deadlocked")
		}
	}
}

func TestProgramBuildContextRegistersAndSharesProjectReferenceReads(t *testing.T) {
	root := t.TempDir()
	writeProgramBuildFixture(t, root, map[string]string{
		"root-a.json": `{
			"files":["a.ts"],
			"references":[{"path":"./shared"}]
		}`,
		"root-b.json": `{
			"files":["b.ts"],
			"references":[{"path":"./shared"}]
		}`,
		"a.ts":                 "export const a = 1",
		"b.ts":                 "export const b = 1",
		"shared/tsconfig.json": `{"compilerOptions":{"composite":true},"files":["index.ts"]}`,
		"shared/index.ts":      "export const shared = 1",
	})

	baseFS := &programReadCountingFS{
		FS:    bundled.WrapFS(cachedvfs.From(osvfs.FS())),
		reads: make(map[string]int),
	}
	context := NewProgramBuildContext(baseFS)
	if _, err := context.CreateProgramLenient(true, root, filepath.Join(root, "root-a.json")); err != nil {
		t.Fatalf("create root A Program: %v", err)
	}
	if _, err := context.CreateProgramLenient(true, root, filepath.Join(root, "root-b.json")); err != nil {
		t.Fatalf("create root B Program: %v", err)
	}

	referencePath := tspath.NormalizePath(filepath.Join(root, "shared/tsconfig.json"))
	if got := baseFS.readCount(referencePath); got != 1 {
		t.Fatalf("shared project-reference config reads = %d, want 1", got)
	}
}

func TestProgramBuildContextWriteInvalidatesExtendedConfigGeneration(t *testing.T) {
	root := t.TempDir()
	writeProgramBuildFixture(t, root, map[string]string{
		"base.json": `{"compilerOptions":{"strict":true}}`,
		"a.json":    `{"extends":"./base.json","files":[]}`,
		"b.json":    `{"extends":"./base.json","files":[]}`,
	})

	context := NewProgramBuildContext(bundled.WrapFS(cachedvfs.From(osvfs.FS())))
	_, before, err := context.parseConfig(root, filepath.Join(root, "a.json"))
	if err != nil {
		t.Fatalf("parse config before write: %v", err)
	}
	if got := before.CompilerOptions().Strict; got != core.TSTrue {
		t.Fatalf("strict before write = %v, want true", got)
	}

	basePath := tspath.NormalizePath(filepath.Join(root, "base.json"))
	if err := context.FS().WriteFile(basePath, `{"compilerOptions":{"strict":false}}`); err != nil {
		t.Fatalf("write extended config: %v", err)
	}
	_, after, err := context.parseConfig(root, filepath.Join(root, "b.json"))
	if err != nil {
		t.Fatalf("parse config after write: %v", err)
	}
	if got := after.CompilerOptions().Strict; got != core.TSFalse {
		t.Fatalf("strict after write = %v, want false", got)
	}
}

func TestProgramBuildContextRequestIsolationAndEscapeHatch(t *testing.T) {
	const path = "/repo/package.json"
	base := newProgramMetadataTestFS(map[string]string{path: "first"})
	first := NewProgramBuildContext(base)
	if content, _ := first.FS().ReadFile(path); content != "first" {
		t.Fatalf("first request read %q", content)
	}
	base.set(path, "second")
	second := NewProgramBuildContext(base)
	if content, _ := second.FS().ReadFile(path); content != "second" {
		t.Fatalf("second request inherited stale metadata %q", content)
	}
	if content, _ := first.FS().ReadFile(path); content != "first" {
		t.Fatalf("first request snapshot changed to %q", content)
	}

	t.Setenv("RSLINT_DISABLE_PROGRAM_METADATA_CACHE", "1")
	disabled := NewProgramBuildContext(base)
	if disabled.FS() != vfs.FS(base) {
		t.Fatal("escape hatch must preserve the exact caller VFS")
	}
	if disabled.metadataFS != nil || disabled.extendedConfigCache != nil {
		t.Fatal("escape hatch left metadata caches enabled")
	}

	root := t.TempDir()
	writeProgramBuildFixture(t, root, map[string]string{
		"base.json":     `{"compilerOptions":{"strict":true}}`,
		"tsconfig.json": `{"extends":"./base.json","files":[]}`,
	})
	disabled = NewProgramBuildContext(bundled.WrapFS(cachedvfs.From(osvfs.FS())))
	if _, err := disabled.CreateProgramLenient(true, root, filepath.Join(root, "tsconfig.json")); err != nil {
		t.Fatalf("escape hatch must preserve upstream extends parsing: %v", err)
	}
}
