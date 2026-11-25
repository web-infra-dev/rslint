package loader

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/cachedvfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/iovfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/config/target"
	"github.com/web-infra-dev/rslint/internal/utils"
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
	mu         sync.Mutex
	reads      map[string]int
	dirEntries map[string]int
}

func (f *programReadCountingFS) ReadFile(path string) (string, bool) {
	f.mu.Lock()
	f.reads[path]++
	f.mu.Unlock()
	return f.FS.ReadFile(path)
}

func (f *programReadCountingFS) GetAccessibleEntries(path string) vfs.Entries {
	f.mu.Lock()
	if f.dirEntries != nil {
		f.dirEntries[tspath.RemoveTrailingDirectorySeparator(tspath.NormalizePath(path))]++
	}
	f.mu.Unlock()
	return f.FS.GetAccessibleEntries(path)
}

func (f *programReadCountingFS) readCount(path string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.reads[path]
}

func (f *programReadCountingFS) dirEntryCount(path string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.dirEntries[tspath.RemoveTrailingDirectorySeparator(tspath.NormalizePath(path))]
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

func TestBuildContextMetadataDiscoveryPreservesParallelSourceReuse(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	const config = `{"compilerOptions":{"noLib":true,"noResolve":true},"files":["shared.ts"]}`
	writeProgramBuildFixture(t, root, map[string]string{
		"first.json":  config,
		"second.json": config,
		"shared.ts":   "export const shared = 1;\n",
	})
	baseFS := &programReadCountingFS{
		FS:    bundled.WrapFS(cachedvfs.From(osvfs.FS())),
		reads: make(map[string]int),
	}
	context := newBuildContext(baseFS)
	var configs []*tsoptions.ParsedCommandLine
	for _, name := range []string{"first.json", "second.json"} {
		parsed, err := context.parseConfig(root, tspath.ResolvePath(root, name))
		if err != nil {
			t.Fatalf("parse config %s: %v", name, err)
		}
		if parsed == nil || len(parsed.Errors) != 0 {
			t.Fatalf("metadata discovery did not parse valid config %s", name)
		}
		configs = append(configs, parsed)
	}
	fileName := tspath.ResolvePath(root, "shared.ts")
	if got := baseFS.readCount(fileName); got != 0 {
		t.Fatalf("metadata discovery read source %d times, want 0", got)
	}

	context.enableConcurrentProgramQueries()
	type buildResult struct {
		program *compiler.Program
		err     error
	}
	results := make(chan buildResult, len(configs))
	for _, parsed := range configs {
		go func() {
			program, err := utils.CreateProgramFromParsedConfigLenient(true, parsed, context.newCompilerHostWithCache(root))
			results <- buildResult{program: program, err: err}
		}()
	}
	first, second := <-results, <-results
	if first.err != nil || second.err != nil {
		t.Fatalf("build programs: first=%v second=%v", first.err, second.err)
	}
	firstSource, secondSource := first.program.GetSourceFile(fileName), second.program.GetSourceFile(fileName)
	if firstSource == nil || secondSource == nil {
		t.Fatal("shared source must be present in both Programs")
	}
	if firstSource != secondSource {
		t.Fatal("identical sources and parse options must reuse the AST across Programs")
	}
	if got := baseFS.readCount(fileName); got != 1 {
		t.Fatalf("source reads after metadata discovery and parallel construction = %d, want 1", got)
	}
}

func TestBuildContextSharesExtendedConfigWithoutFreezingConfigDir(t *testing.T) {
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
	context := newBuildContext(baseFS)
	configA, err := context.parseConfig(root, filepath.Join(root, "packages/a/tsconfig.json"))
	if err != nil {
		t.Fatalf("parse config A: %v", err)
	}
	configB, err := context.parseConfig(root, filepath.Join(root, "packages/b/tsconfig.json"))
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

func TestBuildContextExtendedConfigCrossCycleDoesNotDeadlock(t *testing.T) {
	root := t.TempDir()
	writeProgramBuildFixture(t, root, map[string]string{
		"root-a.json": `{"extends":"./a.json","files":[]}`,
		"root-b.json": `{"extends":"./b.json","files":[]}`,
		"a.json":      `{"extends":"./b.json"}`,
		"b.json":      `{"extends":"./a.json"}`,
	})

	context := newBuildContext(bundled.WrapFS(cachedvfs.From(osvfs.FS())))
	parse := func(name string, done chan<- struct{}) {
		host := context.newCompilerHostWithCache(root)
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

func TestBuildContextRegistersAndSharesProjectReferenceReads(t *testing.T) {
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
	context := newBuildContext(baseFS)
	if _, err := context.createProjectProgram(true, root, filepath.Join(root, "root-a.json")); err != nil {
		t.Fatalf("create root A Program: %v", err)
	}
	if _, err := context.createProjectProgram(true, root, filepath.Join(root, "root-b.json")); err != nil {
		t.Fatalf("create root B Program: %v", err)
	}

	referencePath := tspath.NormalizePath(filepath.Join(root, "shared/tsconfig.json"))
	if got := baseFS.readCount(referencePath); got != 1 {
		t.Fatalf("shared project-reference config reads = %d, want 1", got)
	}
}

func TestBuildContextCachesResolvedProjectReferencesAcrossPrograms(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	files := map[string]string{
		"tsconfig.base.json": `{"compilerOptions":{"strict":true}}`,
		"packages/shared/tsconfig.json": `{
			"extends":"../../tsconfig.base.json",
			"compilerOptions":{"composite":true,"outDir":"dist","rootDir":"src"},
			"include":["src/**/*"]
		}`,
		"packages/shared/src/index.ts": "export const sharedValue: number = 1;\n",
		"packages/shared/src/extra.ts": "export const extraValue: number = 2;\n",
	}
	const appCount = 8
	appDirs := make([]string, appCount)
	appConfigs := make([]string, appCount)
	for i := range appCount {
		name := string(rune('a' + i))
		dir := "packages/app-" + name
		files[dir+"/tsconfig.json"] = `{
			"compilerOptions":{"noLib":true},
			"files":["index.ts"],
			"references":[{"path":"../shared"},{"path":"../missing"}]
		}`
		files[dir+"/index.ts"] = "export const app = 1;\n"
		appDirs[i] = tspath.ResolvePath(root, dir)
		appConfigs[i] = tspath.ResolvePath(root, dir+"/tsconfig.json")
	}
	writeProgramBuildFixture(t, root, files)

	sharedConfig := tspath.ResolvePath(root, "packages/shared/tsconfig.json")
	sharedSrcDir := tspath.ResolvePath(root, "packages/shared/src")
	missingConfig := tspath.ResolvePath(root, "packages/missing/tsconfig.json")

	baseFS := &programReadCountingFS{
		FS:         bundled.WrapFS(cachedvfs.From(osvfs.FS())),
		reads:      make(map[string]int),
		dirEntries: make(map[string]int),
	}
	context := newBuildContext(baseFS)
	useCaseSensitive := baseFS.UseCaseSensitiveFileNames()
	sharedPath := tspath.ToPath(sharedConfig, root, useCaseSensitive)
	missingPath := tspath.ToPath(missingConfig, root, useCaseSensitive)

	firstProgram, err := context.createProjectProgram(true, appDirs[0], appConfigs[0])
	if err != nil {
		t.Fatalf("create app-a Program: %v", err)
	}
	secondProgram, err := context.createProjectProgram(true, appDirs[1], appConfigs[1])
	if err != nil {
		t.Fatalf("create app-b Program: %v", err)
	}

	firstRef, ok := firstProgram.GetResolvedProjectReferenceFor(sharedPath)
	if !ok || firstRef == nil {
		t.Fatalf("app-a did not resolve shared project reference: ok=%v ref=%v", ok, firstRef)
	}
	secondRef, ok := secondProgram.GetResolvedProjectReferenceFor(sharedPath)
	if !ok || secondRef != firstRef {
		t.Fatalf("app-b did not reuse cached *ParsedCommandLine: ok=%v same=%v", ok, secondRef == firstRef)
	}
	if missingRef, ok := secondProgram.GetResolvedProjectReferenceFor(missingPath); !ok || missingRef != nil {
		t.Fatalf("missing project reference = (%v, %v), want (nil, true)", missingRef, ok)
	}
	if got := baseFS.readCount(sharedConfig); got != 1 {
		t.Fatalf("shared project-reference config reads = %d, want 1", got)
	}
	if got := baseFS.dirEntryCount(sharedSrcDir); got != 1 {
		t.Fatalf("shared project-reference directory scans = %d, want 1", got)
	}
	if got := baseFS.readCount(missingConfig); got != 2 {
		t.Fatalf("missing project-reference config reads = %d, want 2 (nil results must not be permanently cached)", got)
	}

	// Cold concurrent Programs across different package working directories must
	// all converge on one immutable winning *ParsedCommandLine instance without
	// data races or duplicate cold parsing.
	concurrentFS := &programReadCountingFS{
		FS:         bundled.WrapFS(cachedvfs.From(osvfs.FS())),
		reads:      make(map[string]int),
		dirEntries: make(map[string]int),
	}
	concurrentContext := newBuildContext(concurrentFS)
	concurrentContext.enableConcurrentProgramQueries()
	type buildResult struct {
		program *compiler.Program
		err     error
	}
	start := make(chan struct{})
	results := make([]buildResult, appCount)
	var wg sync.WaitGroup
	wg.Add(appCount)
	for i := range appCount {
		go func() {
			defer wg.Done()
			<-start
			program, err := concurrentContext.createProjectProgram(false, appDirs[i], appConfigs[i])
			results[i] = buildResult{program: program, err: err}
		}()
	}
	close(start)
	wg.Wait()

	var winner *tsoptions.ParsedCommandLine
	for i, result := range results {
		if result.err != nil {
			t.Fatalf("concurrent app %d build failed: %v", i, result.err)
		}
		ref, ok := result.program.GetResolvedProjectReferenceFor(sharedPath)
		if !ok || ref == nil {
			t.Fatalf("concurrent app %d missing shared reference: ok=%v ref=%v", i, ok, ref)
		}
		if winner == nil {
			winner = ref
		} else if ref != winner {
			t.Fatalf("concurrent app %d observed distinct *ParsedCommandLine instance", i)
		}
		if missingRef, ok := result.program.GetResolvedProjectReferenceFor(missingPath); !ok || missingRef != nil {
			t.Fatalf("concurrent app %d missing reference = (%v, %v), want (nil, true)", i, missingRef, ok)
		}
	}
	if got := concurrentFS.readCount(sharedConfig); got != 1 {
		t.Fatalf("concurrent shared project-reference config reads = %d, want 1", got)
	}
	if got := concurrentFS.dirEntryCount(sharedSrcDir); got != 1 {
		t.Fatalf("concurrent shared project-reference directory scans = %d, want 1", got)
	}
	if got := concurrentFS.readCount(missingConfig); got < 1 {
		t.Fatalf("concurrent missing project-reference config reads = %d, want >= 1", got)
	}
}

func TestBuildContextProjectReferenceCycleDoesNotDeadlock(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	writeProgramBuildFixture(t, root, map[string]string{
		"packages/a/tsconfig.json": `{
			"compilerOptions":{"composite":true,"outDir":"dist","noLib":true},
			"files":["a.ts"],
			"references":[{"path":"../b"}]
		}`,
		"packages/b/tsconfig.json": `{
			"compilerOptions":{"composite":true,"outDir":"dist","noLib":true},
			"files":["b.ts"],
			"references":[{"path":"../a"}]
		}`,
		"packages/a/a.ts": "export const a = 1;\n",
		"packages/b/b.ts": "export const b = 2;\n",
	})

	dirA := tspath.ResolvePath(root, "packages/a")
	dirB := tspath.ResolvePath(root, "packages/b")
	configA := tspath.ResolvePath(root, "packages/a/tsconfig.json")
	configB := tspath.ResolvePath(root, "packages/b/tsconfig.json")

	context := newBuildContext(bundled.WrapFS(cachedvfs.From(osvfs.FS())))
	context.enableConcurrentProgramQueries()
	useCaseSensitive := context.FS().UseCaseSensitiveFileNames()
	pathA := tspath.ToPath(configA, root, useCaseSensitive)
	pathB := tspath.ToPath(configB, root, useCaseSensitive)

	type outcome struct {
		isA     bool
		program *compiler.Program
		err     error
	}
	const runs = 4
	done := make(chan outcome, runs)
	start := make(chan struct{})
	for i := range runs {
		isA := i%2 == 0
		go func() {
			<-start
			cwd, cfg := dirB, configB
			if isA {
				cwd, cfg = dirA, configA
			}
			program, err := context.createProjectProgram(false, cwd, cfg)
			done <- outcome{isA: isA, program: program, err: err}
		}()
	}
	close(start)

	var refA, refB *tsoptions.ParsedCommandLine
	for range runs {
		select {
		case result := <-done:
			if result.err != nil {
				t.Fatalf("cyclic project build failed (isA=%v): %v", result.isA, result.err)
			}
			gotA, okA := result.program.GetResolvedProjectReferenceFor(pathA)
			gotB, okB := result.program.GetResolvedProjectReferenceFor(pathB)
			if !okA || gotA == nil || !okB || gotB == nil {
				t.Fatalf("cyclic references incomplete: A=(%v,%v) B=(%v,%v)", gotA, okA, gotB, okB)
			}
			if refA == nil {
				refA = gotA
			} else if gotA != refA {
				t.Fatal("cyclic project reference A was not shared across Programs")
			}
			if refB == nil {
				refB = gotB
			} else if gotB != refB {
				t.Fatal("cyclic project reference B was not shared across Programs")
			}
		case <-time.After(5 * time.Second):
			t.Fatal("concurrent cyclic project-reference builds deadlocked")
		}
	}
}

func TestBuildContextWriteInvalidatesProjectReferenceGeneration(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	writeProgramBuildFixture(t, root, map[string]string{
		"packages/app/tsconfig.json": `{
			"compilerOptions":{"noLib":true},
			"files":["index.ts"],
			"references":[{"path":"../shared"},{"path":"../late"}]
		}`,
		"packages/app/index.ts":         "export const app = 1;\n",
		"packages/shared/tsconfig.json": `{"compilerOptions":{"composite":true,"strict":true,"noLib":true},"files":["index.ts"]}`,
		"packages/shared/index.ts":      "export const shared = 1;\n",
		"packages/late/index.ts":        "export const late = 1;\n",
	})

	appDir := tspath.ResolvePath(root, "packages/app")
	appConfig := tspath.ResolvePath(root, "packages/app/tsconfig.json")
	sharedConfig := tspath.ResolvePath(root, "packages/shared/tsconfig.json")
	lateConfig := tspath.ResolvePath(root, "packages/late/tsconfig.json")

	context := newBuildContext(bundled.WrapFS(osvfs.FS()))
	useCaseSensitive := context.FS().UseCaseSensitiveFileNames()
	sharedPath := tspath.ToPath(sharedConfig, root, useCaseSensitive)
	latePath := tspath.ToPath(lateConfig, root, useCaseSensitive)

	before, err := context.createProjectProgram(true, appDir, appConfig)
	if err != nil {
		t.Fatalf("create Program before write: %v", err)
	}
	sharedBefore, _ := before.GetResolvedProjectReferenceFor(sharedPath)
	if sharedBefore == nil || sharedBefore.CompilerOptions().Strict != core.TSTrue {
		t.Fatalf("shared strict before write = %v, want true", sharedBefore)
	}
	if lateBefore, ok := before.GetResolvedProjectReferenceFor(latePath); !ok || lateBefore != nil {
		t.Fatalf("late ref before write = (%v, %v), want (nil, true)", lateBefore, ok)
	}

	if err := context.FS().WriteFile(sharedConfig, `{"compilerOptions":{"composite":true,"strict":false,"noLib":true},"files":["index.ts"]}`); err != nil {
		t.Fatalf("update shared tsconfig: %v", err)
	}
	if err := context.FS().WriteFile(lateConfig, `{"compilerOptions":{"composite":true,"noLib":true},"files":["index.ts"]}`); err != nil {
		t.Fatalf("create late tsconfig: %v", err)
	}

	after, err := context.createProjectProgram(true, appDir, appConfig)
	if err != nil {
		t.Fatalf("create Program after write: %v", err)
	}
	sharedAfter, _ := after.GetResolvedProjectReferenceFor(sharedPath)
	if sharedAfter == nil {
		t.Fatal("shared project reference missing after write")
	}
	if sharedAfter == sharedBefore || sharedAfter.CompilerOptions().Strict != core.TSFalse {
		t.Fatalf("write did not invalidate cached shared project reference: same=%v strict=%v", sharedAfter == sharedBefore, sharedAfter.CompilerOptions().Strict)
	}
	if lateAfter, ok := after.GetResolvedProjectReferenceFor(latePath); !ok || lateAfter == nil {
		t.Fatalf("late project reference did not resolve after write: (%v, %v)", lateAfter, ok)
	}
}

func TestBuildContextProjectReferenceCacheIsolatesExactKeys(t *testing.T) {
	const (
		lowerConfig = "/repo/shared/tsconfig.json"
		upperConfig = "/repo/Shared/tsconfig.json"
		relConfig   = "missing/tsconfig.json"
	)
	base := newProgramMetadataTestFS(map[string]string{
		lowerConfig: `{"compilerOptions":{"composite":true,"noLib":true},"files":[]}`,
		upperConfig: `{"compilerOptions":{"composite":true,"noLib":true},"files":[]}`,
	})
	context := newBuildContext(base)
	hostA := context.newCompilerHost("/repo/a")
	hostB := context.newCompilerHost("/repo/b")

	// Same (fileName, path) across different host working directories must hit the cache.
	sharedPath := tspath.Path(lowerConfig)
	lowerRefA := hostA.GetResolvedProjectReference(lowerConfig, sharedPath)
	lowerRefB := hostB.GetResolvedProjectReference(lowerConfig, sharedPath)
	if lowerRefA == nil || lowerRefA != lowerRefB {
		t.Fatalf("identical (fileName, path) across different working directories missed cache: A=%p B=%p", lowerRefA, lowerRefB)
	}

	// Same canonical path with different fileName casing must keep distinct ConfigName entries.
	upperRef := hostB.GetResolvedProjectReference(upperConfig, sharedPath)
	if upperRef == nil || lowerRefA == upperRef {
		t.Fatalf("distinct fileName casing collided: lower=%p upper=%p", lowerRefA, upperRef)
	}
	if lowerRefA.ConfigName() != lowerConfig || upperRef.ConfigName() != upperConfig {
		t.Fatalf("preserved ConfigName mismatch: lower=%q upper=%q", lowerRefA.ConfigName(), upperRef.ConfigName())
	}

	// Same fileName with distinct canonical path must keep distinct SourceFile.Path entries.
	altPath := tspath.Path("/repo/shared-alias/tsconfig.json")
	altRef := hostB.GetResolvedProjectReference(lowerConfig, altPath)
	if altRef == nil || altRef == lowerRefA {
		t.Fatalf("distinct canonical path collided: lower=%p alt=%p", lowerRefA, altRef)
	}
	if altRef.ConfigFile.SourceFile.Path() != altPath {
		t.Fatalf("alt reference path = %q, want %q", altRef.ConfigFile.SourceFile.Path(), altPath)
	}

	// Relative config file names must bypass the cache map completely (even if an
	// entry were present under that key) so distinct working directories never collide.
	relPath := tspath.Path(relConfig)
	relKey := programProjectReferenceKey{fileName: relConfig, path: relPath}
	context.metadataFS.currentProjectReferenceGeneration().entries.Store(relKey, &cachedProjectReference{commandLine: lowerRefA})
	if ref := hostA.GetResolvedProjectReference(relConfig, relPath); ref != nil {
		t.Fatalf("relative config hit project-reference cache instead of bypassing: got %v, want nil", ref)
	}
	if got := base.readCount(relConfig); got != 1 {
		t.Fatalf("relative config read count = %d, want 1 (must bypass cache)", got)
	}
}

func TestBuildContextProjectReferenceFinalizesDiagnosticsBeforePublish(t *testing.T) {
	cases := []struct {
		name           string
		tsconfig       string
		extraFiles     map[string]string
		wantMappings   int
		wantOutDts     int
		wantEagerDiags bool
		wantLazyTS6059 bool
		wantShared     bool
	}{
		{
			name: "with_out_dir",
			tsconfig: `{
				"compilerOptions":{"composite":true,"rootDir":"src","outDir":"dist","noLib":true},
				"files":["src/index.ts","outside.ts"]
			}`,
			wantMappings:   2,
			wantOutDts:     2,
			wantEagerDiags: true,
			wantLazyTS6059: true,
			wantShared:     true,
		},
		{
			name: "in_place_valid_within_root_dir",
			tsconfig: `{
				"compilerOptions":{"composite":true,"rootDir":"src","noLib":true},
				"files":["src/index.ts"]
			}`,
			wantMappings:   1,
			wantOutDts:     1,
			wantEagerDiags: false,
			wantLazyTS6059: false,
			wantShared:     true,
		},
		{
			name: "in_place_outside_root_dir_unshared",
			tsconfig: `{
				"compilerOptions":{"composite":true,"rootDir":"src","noLib":true},
				"files":["src/index.ts","outside.ts"]
			}`,
			wantMappings:   2,
			wantOutDts:     2,
			wantEagerDiags: false,
			wantLazyTS6059: true,
			wantShared:     false,
		},
		{
			name: "out_dir_json_only_outside_root_dir_unshared",
			tsconfig: `{
				"compilerOptions":{"composite":true,"rootDir":"src","outDir":"dist","resolveJsonModule":true,"noLib":true},
				"files":["outside.json"]
			}`,
			extraFiles: map[string]string{
				"packages/shared/outside.json": `{"ok":true}`,
			},
			wantMappings:   1,
			wantOutDts:     0,
			wantEagerDiags: false,
			wantLazyTS6059: true,
			wantShared:     false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := tspath.NormalizePath(t.TempDir())
			fixture := map[string]string{
				"packages/shared/tsconfig.json": tc.tsconfig,
				"packages/shared/src/index.ts":  "export const inside = 1;\n",
				"packages/shared/outside.ts":    "export const outside = 2;\n",
			}
			for path, content := range tc.extraFiles {
				fixture[path] = content
			}
			writeProgramBuildFixture(t, root, fixture)

			sharedConfig := tspath.ResolvePath(root, "packages/shared/tsconfig.json")
			ctx := newBuildContext(bundled.WrapFS(cachedvfs.From(osvfs.FS())))
			sharedPath := tspath.ToPath(sharedConfig, root, ctx.FS().UseCaseSensitiveFileNames())

			if !tc.wantEagerDiags {
				// Without ParseInputOutputNames triggering CommonSourceDirectory,
				// loading must NOT eagerly call CommonSourceDirectory() or append TS6059.
				initial := ctx.newCompilerHost(root).GetResolvedProjectReference(sharedConfig, sharedPath)
				if initial == nil {
					t.Fatal("shared project reference resolved to nil")
				}
				if got := len(initial.GetConfigFileParsingDiagnostics()); got != 0 {
					t.Fatalf("unexpected eager diagnostics before CommonSourceDirectory: got %d, want 0", got)
				}
			}

			const workers = 16
			start := make(chan struct{})
			refs := make([]*tsoptions.ParsedCommandLine, workers)
			diagCountsBefore := make([]int, workers)
			diagCountsAfter := make([]int, workers)
			commonDirs := make([]string, workers)
			var wg sync.WaitGroup
			wg.Add(workers)
			for i := range workers {
				go func() {
					defer wg.Done()
					host := ctx.newCompilerHost(root)
					<-start
					ref := host.GetResolvedProjectReference(sharedConfig, sharedPath)
					refs[i] = ref
					if ref == nil {
						return
					}
					// Interleave readers of Errors, CommonSourceDirectory, and
					// ParseInputOutputNames under -race across all cases.
					switch i % 3 {
					case 0:
						diagCountsBefore[i] = len(ref.GetConfigFileParsingDiagnostics())
						commonDirs[i] = ref.CommonSourceDirectory()
						ref.ParseInputOutputNames()
					case 1:
						commonDirs[i] = ref.CommonSourceDirectory()
						diagCountsBefore[i] = len(ref.GetConfigFileParsingDiagnostics())
						ref.ParseInputOutputNames()
					default:
						ref.ParseInputOutputNames()
						diagCountsBefore[i] = len(ref.GetConfigFileParsingDiagnostics())
						commonDirs[i] = ref.CommonSourceDirectory()
					}
					diagCountsAfter[i] = len(ref.GetConfigFileParsingDiagnostics())
					_ = len(ref.SourceToProjectReference())
					_ = len(ref.OutputDtsToProjectReference())
				}()
			}
			close(start)
			wg.Wait()

			winner := refs[0]
			if winner == nil {
				t.Fatal("shared project reference resolved to nil")
			}
			if tc.wantEagerDiags && diagCountsBefore[0] < 1 {
				t.Fatalf("expected >= 1 rootDir diagnostic before publication, got %d", diagCountsBefore[0])
			}
			if tc.wantLazyTS6059 && diagCountsAfter[0] < 1 {
				t.Fatalf("expected >= 1 TS6059 diagnostic after CommonSourceDirectory, got %d", diagCountsAfter[0])
			}
			if !tc.wantLazyTS6059 && diagCountsAfter[0] != 0 {
				t.Fatalf("unexpected diagnostics for valid config: got %d, want 0", diagCountsAfter[0])
			}
			if commonDirs[0] == "" {
				t.Fatal("expected non-empty CommonSourceDirectory")
			}
			seenPtrs := map[*tsoptions.ParsedCommandLine]struct{}{winner: {}}
			for i := 1; i < workers; i++ {
				if tc.wantShared {
					if refs[i] != winner {
						t.Fatalf("worker %d observed distinct *ParsedCommandLine instance", i)
					}
				} else {
					if _, dup := seenPtrs[refs[i]]; dup {
						t.Fatalf("worker %d shared mutable *ParsedCommandLine instance with another caller", i)
					}
					seenPtrs[refs[i]] = struct{}{}
				}
				if tc.wantEagerDiags && diagCountsBefore[i] != diagCountsBefore[0] {
					t.Fatalf("worker %d saw %d diagnostics, want %d", i, diagCountsBefore[i], diagCountsBefore[0])
				}
				if diagCountsAfter[i] != diagCountsAfter[0] {
					t.Fatalf("worker %d saw %d diagnostics after CommonSourceDirectory, want %d", i, diagCountsAfter[i], diagCountsAfter[0])
				}
				if commonDirs[i] != commonDirs[0] {
					t.Fatalf("worker %d saw CommonSourceDirectory=%q, want %q", i, commonDirs[i], commonDirs[0])
				}
			}
			if got := len(winner.SourceToProjectReference()); got != tc.wantMappings {
				t.Fatalf("SourceToProjectReference size = %d, want %d", got, tc.wantMappings)
			}
			if got := len(winner.OutputDtsToProjectReference()); got != tc.wantOutDts {
				t.Fatalf("OutputDtsToProjectReference size = %d, want %d", got, tc.wantOutDts)
			}
		})
	}
}

func TestBuildContextConcurrentProgramsNoRaceOnInPlaceOutOfRootReference(t *testing.T) {
	cases := []struct {
		name           string
		sharedTSConfig string
		files          map[string]string
	}{
		{
			name: "explicit_root_dir",
			sharedTSConfig: `{
				"compilerOptions":{"composite":true,"rootDir":"src","noLib":true},
				"files":["src/index.ts","outside.ts"]
			}`,
			files: map[string]string{
				"packages/shared/src/index.ts": "export const inside = 1;\n",
				"packages/shared/outside.ts":   "export const outside = 2;\n",
			},
		},
		{
			name: "config_dir_root_dir",
			sharedTSConfig: `{
				"compilerOptions":{"composite":true,"noLib":true},
				"files":["src/index.ts","../outside.ts"]
			}`,
			files: map[string]string{
				"packages/shared/src/index.ts": "export const inside = 1;\n",
				"packages/outside.ts":          "export const outside = 2;\n",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := tspath.NormalizePath(t.TempDir())
			fixture := map[string]string{
				"packages/shared/tsconfig.json": tc.sharedTSConfig,
				"packages/app-a/tsconfig.json": `{
					"compilerOptions":{"noLib":true},
					"files":["index.ts"],
					"references":[{"path":"../shared"}]
				}`,
				"packages/app-b/tsconfig.json": `{
					"compilerOptions":{"noLib":true},
					"files":["index.ts"],
					"references":[{"path":"../shared"}]
				}`,
				"packages/app-a/index.ts": "export const a = 1;\n",
				"packages/app-b/index.ts": "export const b = 2;\n",
			}
			for path, content := range tc.files {
				fixture[path] = content
			}
			writeProgramBuildFixture(t, root, fixture)

			appADir := tspath.ResolvePath(root, "packages/app-a")
			appBDir := tspath.ResolvePath(root, "packages/app-b")
			appAConfig := tspath.ResolvePath(root, "packages/app-a/tsconfig.json")
			appBConfig := tspath.ResolvePath(root, "packages/app-b/tsconfig.json")
			sharedConfig := tspath.ResolvePath(root, "packages/shared/tsconfig.json")

			// Verify uncached baseline diagnostic timing for exact parity.
			t.Setenv("RSLINT_DISABLE_PROGRAM_METADATA_CACHE", "1")
			uncachedCtx := newBuildContext(bundled.WrapFS(cachedvfs.From(osvfs.FS())))
			uncachedProg, err := uncachedCtx.createProjectProgram(true, appADir, appAConfig)
			if err != nil {
				t.Fatalf("uncached program build: %v", err)
			}
			sharedPath := tspath.ToPath(sharedConfig, root, uncachedCtx.FS().UseCaseSensitiveFileNames())
			uncachedRef, ok := uncachedProg.GetResolvedProjectReferenceFor(sharedPath)
			if !ok || uncachedRef == nil {
				t.Fatalf("uncached shared ref missing: ok=%v ref=%v", ok, uncachedRef)
			}
			wantBefore := len(uncachedRef.GetConfigFileParsingDiagnostics())
			_ = uncachedRef.CommonSourceDirectory()
			wantAfter := len(uncachedRef.GetConfigFileParsingDiagnostics())
			if wantBefore != 0 || wantAfter < 1 {
				t.Fatalf("unexpected uncached baseline diagnostics: before=%d after=%d", wantBefore, wantAfter)
			}

			// Now run with metadata + project-reference caching enabled across two
			// concurrent Programs in one buildContext.
			t.Setenv("RSLINT_DISABLE_PROGRAM_METADATA_CACHE", "")
			cachedCtx := newBuildContext(bundled.WrapFS(cachedvfs.From(osvfs.FS())))
			cachedCtx.enableConcurrentProgramQueries()

			type buildOut struct {
				program *compiler.Program
				err     error
			}
			startBuild := make(chan struct{})
			outA := make(chan buildOut, 1)
			outB := make(chan buildOut, 1)
			go func() {
				<-startBuild
				p, e := cachedCtx.createProjectProgram(false, appADir, appAConfig)
				outA <- buildOut{program: p, err: e}
			}()
			go func() {
				<-startBuild
				p, e := cachedCtx.createProjectProgram(false, appBDir, appBConfig)
				outB <- buildOut{program: p, err: e}
			}()
			close(startBuild)
			resA, resB := <-outA, <-outB
			if resA.err != nil || resB.err != nil {
				t.Fatalf("concurrent program builds failed: A=%v B=%v", resA.err, resB.err)
			}

			refA, okA := resA.program.GetResolvedProjectReferenceFor(sharedPath)
			refB, okB := resB.program.GetResolvedProjectReferenceFor(sharedPath)
			if !okA || refA == nil || !okB || refB == nil {
				t.Fatalf("resolved shared refs missing: A=(%v,%v) B=(%v,%v)", refA, okA, refB, okB)
			}
			if refA == refB {
				t.Fatal("two Programs shared the same mutable in-place out-of-root reference")
			}

			// Concurrently trigger CommonSourceDirectory() (which appends TS6059)
			// on one Program while the other reads GetConfigFileParsingDiagnostics().
			startRace := make(chan struct{})
			var raceWG sync.WaitGroup
			raceWG.Add(2)
			var commonA string
			var diagsBBefore int
			go func() {
				defer raceWG.Done()
				<-startRace
				commonA = refA.CommonSourceDirectory()
			}()
			go func() {
				defer raceWG.Done()
				<-startRace
				diagsBBefore = len(refB.GetConfigFileParsingDiagnostics())
			}()
			close(startRace)
			raceWG.Wait()

			if commonA == "" {
				t.Fatal("refA.CommonSourceDirectory() returned empty string")
			}
			if diagsBBefore != wantBefore {
				t.Fatalf("refB diagnostics before its own CommonSourceDirectory() = %d, want %d", diagsBBefore, wantBefore)
			}
			diagsAAfter := refA.GetConfigFileParsingDiagnostics()
			if len(diagsAAfter) != wantAfter {
				t.Fatalf("refA diagnostics after CommonSourceDirectory() = %d, want %d", len(diagsAAfter), wantAfter)
			}
			if diagsAAfter[0].Code() != 6059 {
				t.Fatalf("refA diagnostic code = %d, want 6059", diagsAAfter[0].Code())
			}
			_ = refB.CommonSourceDirectory()
			if got := len(refB.GetConfigFileParsingDiagnostics()); got != wantAfter {
				t.Fatalf("refB diagnostics after CommonSourceDirectory() = %d, want %d", got, wantAfter)
			}
		})
	}
}

type overlappingGlobCountingFS struct {
	vfs.FS
	mu               sync.Mutex
	targetReads      map[string]int
	targetExists     map[string]int
	firstScanDir     string
	firstScanStarted chan struct{}
	expectedScans    int
	arrivedScans     int
	allScansArrived  chan struct{}
	watchedDirs      map[string]struct{}
}

func (f *overlappingGlobCountingFS) ReadFile(path string) (string, bool) {
	f.mu.Lock()
	if _, tracked := f.targetReads[path]; tracked {
		f.targetReads[path]++
	}
	f.mu.Unlock()
	return f.FS.ReadFile(path)
}

func (f *overlappingGlobCountingFS) FileExists(path string) bool {
	f.mu.Lock()
	if _, tracked := f.targetExists[path]; tracked {
		f.targetExists[path]++
	}
	f.mu.Unlock()
	return f.FS.FileExists(path)
}

func (f *overlappingGlobCountingFS) GetAccessibleEntries(path string) vfs.Entries {
	normalized := tspath.RemoveTrailingDirectorySeparator(tspath.NormalizePath(path))
	f.mu.Lock()
	_, watched := f.watchedDirs[normalized]
	if watched {
		delete(f.watchedDirs, normalized)
		if normalized == f.firstScanDir {
			close(f.firstScanStarted)
		}
		f.arrivedScans++
		if f.arrivedScans == f.expectedScans {
			close(f.allScansArrived)
		}
	}
	f.mu.Unlock()
	if watched {
		<-f.allScansArrived
	}
	return f.FS.GetAccessibleEntries(path)
}

func TestBuildContextConcurrentRootsShareExtendedConfigDuringSourceGlobScan(t *testing.T) {
	root := osvfs.FS().Realpath(tspath.NormalizePath(t.TempDir()))
	const rootCount = 8
	files := map[string]string{
		"grand.json": `{"compilerOptions":{"strict":true,"noLib":true}}`,
		"base.json":  `{"extends":"./grand"}`,
	}
	appDirs := make([]string, rootCount)
	appConfigs := make([]string, rootCount)
	watchedDirs := make(map[string]struct{}, rootCount)
	for i := range rootCount {
		name := string(rune('a' + i))
		dir := "packages/app-" + name
		files[dir+"/tsconfig.json"] = `{"extends":"../../base.json","include":["src/**/*"]}`
		files[dir+"/src/index.ts"] = "export const value: number = 1;\n"
		appDirs[i] = tspath.ResolvePath(root, dir)
		appConfigs[i] = tspath.ResolvePath(root, dir+"/tsconfig.json")
		watchedDirs[tspath.ResolvePath(root, dir+"/src")] = struct{}{}
	}
	writeProgramBuildFixture(t, root, files)

	baseJSON := tspath.ResolvePath(root, "base.json")
	grandJSON := tspath.ResolvePath(root, "grand.json")
	grandNoExt := tspath.ResolvePath(root, "grand")

	countingFS := &overlappingGlobCountingFS{
		FS: bundled.WrapFS(osvfs.FS()),
		targetReads: map[string]int{
			baseJSON:  0,
			grandJSON: 0,
		},
		targetExists: map[string]int{
			grandJSON:  0,
			grandNoExt: 0,
		},
		firstScanDir:     tspath.ResolvePath(root, "packages/app-a/src"),
		firstScanStarted: make(chan struct{}),
		expectedScans:    rootCount,
		allScansArrived:  make(chan struct{}),
		watchedDirs:      watchedDirs,
	}
	ctx := newBuildContext(countingFS)
	ctx.enableConcurrentProgramQueries()

	type parseOutcome struct {
		config *tsoptions.ParsedCommandLine
		err    error
	}
	outcomes := make([]parseOutcome, rootCount)
	var wg sync.WaitGroup
	wg.Add(rootCount)

	// Start root 0 first and pause it once it enters its source glob scan
	// (while parseConfig is still in flight), then launch roots 1..7 in
	// parallel so all 8 source glob scans overlap simultaneously.
	go func() {
		defer wg.Done()
		cfg, err := ctx.parseConfig(appDirs[0], appConfigs[0])
		outcomes[0] = parseOutcome{config: cfg, err: err}
	}()
	select {
	case <-countingFS.firstScanStarted:
	case <-time.After(10 * time.Second):
		t.Fatal("root 0 did not reach source glob scan")
	}

	startRest := make(chan struct{})
	for i := 1; i < rootCount; i++ {
		go func() {
			defer wg.Done()
			<-startRest
			cfg, err := ctx.parseConfig(appDirs[i], appConfigs[i])
			outcomes[i] = parseOutcome{config: cfg, err: err}
		}()
	}
	close(startRest)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("overlapping source glob scans deadlocked")
	}

	for i, out := range outcomes {
		if out.err != nil {
			t.Fatalf("root %d parseConfig failed: %v", i, out.err)
		}
		if out.config == nil {
			t.Fatalf("root %d parseConfig returned nil", i)
		}
		if len(out.config.GetConfigFileParsingDiagnostics()) != 0 {
			t.Fatalf("root %d had unexpected diagnostics: %d", i, len(out.config.GetConfigFileParsingDiagnostics()))
		}
		if out.config.CompilerOptions().Strict != core.TSTrue {
			t.Fatalf("root %d did not inherit strict=true from grand.json", i)
		}
		if len(out.config.FileNames()) != 1 {
			t.Fatalf("root %d matched %d files, want 1", i, len(out.config.FileNames()))
		}
	}

	countingFS.mu.Lock()
	baseReads := countingFS.targetReads[baseJSON]
	grandReads := countingFS.targetReads[grandJSON]
	grandExists := countingFS.targetExists[grandJSON]
	grandNoExtExists := countingFS.targetExists[grandNoExt]
	countingFS.mu.Unlock()

	if baseReads != 1 {
		t.Fatalf("base.json reads = %d, want 1", baseReads)
	}
	if grandReads != 1 {
		t.Fatalf("grand.json reads = %d, want 1", grandReads)
	}
	if grandExists != 1 {
		t.Fatalf("grand.json FileExists lookups = %d, want 1", grandExists)
	}
	if grandNoExtExists != 1 {
		t.Fatalf("grand FileExists lookups = %d, want 1", grandNoExtExists)
	}
}

func TestBuildContextProjectReferenceSymlinkedPackageJSONPermissionRecovery(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX directory permissions required")
	}

	cases := []struct {
		name           string
		sharedTSConfig string
		extraFiles     map[string]string
	}{
		{
			name: "direct_package_extends",
			sharedTSConfig: `{
				"extends":"@acme/tsconfig/base.json",
				"compilerOptions":{"composite":true,"noLib":true},
				"files":["src/value.ts"]
			}`,
		},
		{
			name: "two_hop_package_extends",
			sharedTSConfig: `{
				"extends":"../base.json",
				"compilerOptions":{"composite":true,"noLib":true},
				"files":["src/value.ts"]
			}`,
			extraFiles: map[string]string{
				"packages/base.json": `{"extends":"@acme/tsconfig/base.json"}`,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := osvfs.FS().Realpath(tspath.NormalizePath(t.TempDir()))
			fixture := map[string]string{
				"secret/package.json":                           `{"name":"@acme/tsconfig","exports":{"./base.json":"./configs/base.json"}}`,
				"node_modules/@acme/tsconfig/base.json":         `{"compilerOptions":{"rootDir":"${configDir}/src"}}`,
				"node_modules/@acme/tsconfig/configs/base.json": `{"compilerOptions":{"outDir":"${configDir}/dist","rootDir":"${configDir}/src"}}`,
				"packages/shared/tsconfig.json":                 tc.sharedTSConfig,
				"packages/shared/src/value.ts":                  "export const value: number = 1;\n",
				"packages/shared/dist/value.d.ts":               "export declare const value: number;\n",
				"packages/app/tsconfig.json": `{
					"compilerOptions":{"noLib":true,"disableSourceOfProjectReferenceRedirect":true},
					"files":["index.ts"],
					"references":[{"path":"../shared"}]
				}`,
				"packages/app/index.ts": "import { value } from \"../shared/src/value\";\nexport const doubled: number = value * 2;\n",
			}
			for path, content := range tc.extraFiles {
				fixture[path] = content
			}
			writeProgramBuildFixture(t, root, fixture)

			secretDir := filepath.Join(root, "secret")
			secretPkg := filepath.Join(secretDir, "package.json")
			pkgLink := filepath.Join(root, "node_modules/@acme/tsconfig/package.json")
			if err := os.Symlink(secretPkg, pkgLink); err != nil {
				t.Fatalf("Symlink(%s -> %s): %v", pkgLink, secretPkg, err)
			}

			if err := os.Chmod(secretDir, 0); err != nil {
				t.Fatalf("Chmod(0): %v", err)
			}
			t.Cleanup(func() { _ = os.Chmod(secretDir, 0o755) })
			if _, err := os.Stat(pkgLink); !errors.Is(err, fs.ErrPermission) {
				t.Skipf("OS did not enforce chmod 000 on symlink target directory (likely root): err=%v", err)
			}

			appDir := tspath.ResolvePath(root, "packages/app")
			appConfig := tspath.ResolvePath(root, "packages/app/tsconfig.json")
			appEntry := tspath.ResolvePath(root, "packages/app/index.ts")
			sharedConfig := tspath.ResolvePath(root, "packages/shared/tsconfig.json")

			ctx := newBuildContext(bundled.WrapFS(osvfs.FS()))
			useCaseSensitive := ctx.FS().UseCaseSensitiveFileNames()
			sharedPath := tspath.ToPath(sharedConfig, root, useCaseSensitive)

			// Program 1: package.json is behind chmod 000, so TypeScript falls
			// back to legacy base.json (no outDir -> expects src/value.d.ts -> TS6305).
			firstProg, err := ctx.createProjectProgram(true, appDir, appConfig)
			if err != nil {
				t.Fatalf("first program build: %v", err)
			}
			firstRef, ok := firstProg.GetResolvedProjectReferenceFor(sharedPath)
			if !ok || firstRef == nil {
				t.Fatalf("first program missing shared reference: ok=%v ref=%v", ok, firstRef)
			}
			if len(firstRef.GetConfigFileParsingDiagnostics()) != 0 {
				t.Fatalf("expected diagnostic-free fallback on first program, got %d diagnostics", len(firstRef.GetConfigFileParsingDiagnostics()))
			}
			if firstRef.CompilerOptions().OutDir != "" {
				t.Fatalf("first program unexpectedly resolved outDir=%q while package.json was inaccessible", firstRef.CompilerOptions().OutDir)
			}
			firstSF := firstProg.GetSourceFile(appEntry)
			if firstSF == nil {
				t.Fatal("first program missing app/index.ts")
			}
			firstDiags := firstProg.GetSemanticDiagnostics(context.Background(), firstSF)
			hasTS6305 := false
			for _, diag := range firstDiags {
				if diag.Code() == 6305 {
					hasTS6305 = true
					break
				}
			}
			if !hasTS6305 {
				t.Fatalf("expected TS6305 while package.json was inaccessible and fallback looked for src/value.d.ts, got %d diagnostics", len(firstDiags))
			}

			// Restore access to secretDir and verify the next Program picks up
			// package.json exports -> configs/base.json -> dist/value.d.ts cleanly.
			if err := os.Chmod(secretDir, 0o755); err != nil {
				t.Fatalf("Chmod(0755): %v", err)
			}

			secondProg, err := ctx.createProjectProgram(true, appDir, appConfig)
			if err != nil {
				t.Fatalf("second program build: %v", err)
			}
			secondRef, ok := secondProg.GetResolvedProjectReferenceFor(sharedPath)
			if !ok || secondRef == nil {
				t.Fatalf("second program missing shared reference: ok=%v ref=%v", ok, secondRef)
			}
			if secondRef == firstRef {
				t.Fatal("stale fallback project reference remained cached after package.json access was restored")
			}
			wantOutDir := tspath.ResolvePath(root, "packages/shared/dist")
			if got := secondRef.CompilerOptions().OutDir; got != wantOutDir {
				t.Fatalf("recovered outDir = %q, want %q", got, wantOutDir)
			}
			wantOutDts := tspath.ToPath(tspath.ResolvePath(root, "packages/shared/dist/value.d.ts"), tspath.ResolvePath(root, "packages/shared"), useCaseSensitive)
			if _, ok := secondRef.OutputDtsToProjectReference()[wantOutDts]; !ok {
				t.Fatalf("recovered OutputDtsToProjectReference missing %q: %v", wantOutDts, secondRef.OutputDtsToProjectReference())
			}
			secondSF := secondProg.GetSourceFile(appEntry)
			if secondSF == nil {
				t.Fatal("second program missing app/index.ts")
			}
			if secondDiags := secondProg.GetSemanticDiagnostics(context.Background(), secondSF); len(secondDiags) != 0 {
				t.Fatalf("expected 0 semantic diagnostics after access recovery, got %d (first code=%d)", len(secondDiags), secondDiags[0].Code())
			}

			// Program 3: healthy recovered reference should now be cached and reused.
			thirdProg, err := ctx.createProjectProgram(true, appDir, appConfig)
			if err != nil {
				t.Fatalf("third program build: %v", err)
			}
			thirdRef, ok := thirdProg.GetResolvedProjectReferenceFor(sharedPath)
			if !ok || thirdRef != secondRef {
				t.Fatalf("recovered healthy project reference was not cached: second=%p third=%p", secondRef, thirdRef)
			}
		})
	}
}

type transientInaccessibleFS struct {
	vfs.FS
	mu      sync.Mutex
	target  string
	blocked bool
}

func (f *transientInaccessibleFS) isBlocked(path string) bool {
	if path != f.target {
		return false
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.blocked
}

func (f *transientInaccessibleFS) FileExists(path string) bool {
	if f.isBlocked(path) {
		return false
	}
	return f.FS.FileExists(path)
}

func (f *transientInaccessibleFS) ReadFile(path string) (string, bool) {
	if f.isBlocked(path) {
		return "", false
	}
	return f.FS.ReadFile(path)
}

func (f *transientInaccessibleFS) setBlocked(blocked bool) {
	f.mu.Lock()
	f.blocked = blocked
	f.mu.Unlock()
}

func TestBuildContextProjectReferenceRetriesTransientExtendsAndInaccessibleDir(t *testing.T) {
	t.Run("extends_package_json_transient_read_failure_restores_outDir", func(t *testing.T) {
		root := osvfs.FS().Realpath(tspath.NormalizePath(t.TempDir()))
		writeProgramBuildFixture(t, root, map[string]string{
			"node_modules/@acme/tsconfig/package.json":      `{"name":"@acme/tsconfig","exports":{"./base.json":"./configs/base.json"}}`,
			"node_modules/@acme/tsconfig/configs/base.json": `{"compilerOptions":{"outDir":"${configDir}/dist","rootDir":"${configDir}/src"}}`,
			"packages/shared/tsconfig.json": `{
				"extends":"@acme/tsconfig/base.json",
				"compilerOptions":{"composite":true,"noLib":true},
				"files":["src/value.ts"]
			}`,
			"packages/shared/src/value.ts":    "export const value: number = 1;\n",
			"packages/shared/dist/value.d.ts": "export declare const value: number;\n",
		})

		pkgJSON := tspath.ResolvePath(root, "node_modules/@acme/tsconfig/package.json")
		sharedConfig := tspath.ResolvePath(root, "packages/shared/tsconfig.json")
		base := &transientProjectReferenceReadFS{
			FS:       bundled.WrapFS(osvfs.FS()),
			target:   pkgJSON,
			failLeft: 1,
		}
		context := newBuildContext(base)
		sharedPath := tspath.ToPath(sharedConfig, root, context.FS().UseCaseSensitiveFileNames())
		host := context.newCompilerHost(root)

		first := host.GetResolvedProjectReference(sharedConfig, sharedPath)
		if first == nil {
			t.Fatal("expected parsed config even when extends package.json read failed")
		}
		if first.CompilerOptions().OutDir != "" {
			t.Fatalf("first attempt unexpectedly inherited outDir=%q while package.json was unreadable", first.CompilerOptions().OutDir)
		}

		second := host.GetResolvedProjectReference(sharedConfig, sharedPath)
		if second == nil {
			t.Fatal("second attempt returned nil after package.json recovered")
		}
		if second == first {
			t.Fatalf("second attempt returned stale config after package.json recovered: first=%p second=%p", first, second)
		}
		wantOutDir := tspath.ResolvePath(root, "packages/shared/dist")
		if got := second.CompilerOptions().OutDir; got != wantOutDir {
			t.Fatalf("recovered outDir = %q, want %q", got, wantOutDir)
		}
		wantOutDts := tspath.ToPath(tspath.ResolvePath(root, "packages/shared/dist/value.d.ts"), tspath.ResolvePath(root, "packages/shared"), context.FS().UseCaseSensitiveFileNames())
		if _, ok := second.OutputDtsToProjectReference()[wantOutDts]; !ok {
			t.Fatalf("recovered OutputDtsToProjectReference missing %q: %v", wantOutDts, second.OutputDtsToProjectReference())
		}
		third := host.GetResolvedProjectReference(sharedConfig, sharedPath)
		if third != second {
			t.Fatalf("healthy config was not cached after recovery: second=%p third=%p", second, third)
		}
	})

	t.Run("relative_extended_config_transient_read_failure_retries", func(t *testing.T) {
		root := osvfs.FS().Realpath(tspath.NormalizePath(t.TempDir()))
		writeProgramBuildFixture(t, root, map[string]string{
			"packages/base.json": `{"compilerOptions":{"outDir":"${configDir}/dist","rootDir":"${configDir}/src"}}`,
			"packages/shared/tsconfig.json": `{
				"extends":"../base.json",
				"compilerOptions":{"composite":true,"noLib":true},
				"files":["src/value.ts"]
			}`,
			"packages/shared/src/value.ts": "export const value = 1;\n",
		})

		baseJSON := tspath.ResolvePath(root, "packages/base.json")
		sharedConfig := tspath.ResolvePath(root, "packages/shared/tsconfig.json")
		base := &transientProjectReferenceReadFS{
			FS:       bundled.WrapFS(osvfs.FS()),
			target:   baseJSON,
			failLeft: 1,
		}
		context := newBuildContext(base)
		sharedPath := tspath.ToPath(sharedConfig, root, context.FS().UseCaseSensitiveFileNames())
		host := context.newCompilerHost(root)

		first := host.GetResolvedProjectReference(sharedConfig, sharedPath)
		if first == nil {
			t.Fatal("first attempt returned nil")
		}
		if first.CompilerOptions().OutDir != "" {
			t.Fatalf("first attempt unexpectedly inherited outDir=%q", first.CompilerOptions().OutDir)
		}
		second := host.GetResolvedProjectReference(sharedConfig, sharedPath)
		wantOutDir := tspath.ResolvePath(root, "packages/shared/dist")
		if second == nil {
			t.Fatal("second attempt returned nil after base.json recovered")
		}
		if got := second.CompilerOptions().OutDir; got != wantOutDir {
			t.Fatalf("recovered outDir = %q, want %q", got, wantOutDir)
		}
	})

	t.Run("diagnostic_free_package_json_fallback_retries", func(t *testing.T) {
		root := osvfs.FS().Realpath(tspath.NormalizePath(t.TempDir()))
		// Both a legacy direct base.json (no outDir) and package.json exports -> configs/base.json (with outDir)
		// exist: unreadable package.json resolves legacy base.json with ZERO config diagnostics,
		// so only transientReadTrackerFS catches the failed ReadFile.
		writeProgramBuildFixture(t, root, map[string]string{
			"node_modules/@acme/tsconfig/package.json":      `{"name":"@acme/tsconfig","exports":{"./base.json":"./configs/base.json"}}`,
			"node_modules/@acme/tsconfig/base.json":         `{"compilerOptions":{"rootDir":"${configDir}/src"}}`,
			"node_modules/@acme/tsconfig/configs/base.json": `{"compilerOptions":{"outDir":"${configDir}/dist","rootDir":"${configDir}/src"}}`,
			"packages/shared/tsconfig.json": `{
				"extends":"@acme/tsconfig/base.json",
				"compilerOptions":{"composite":true,"noLib":true},
				"files":["src/value.ts"]
			}`,
			"packages/shared/src/value.ts": "export const value = 1;\n",
		})

		pkgJSON := tspath.ResolvePath(root, "node_modules/@acme/tsconfig/package.json")
		sharedConfig := tspath.ResolvePath(root, "packages/shared/tsconfig.json")
		base := &transientProjectReferenceReadFS{
			FS:       bundled.WrapFS(osvfs.FS()),
			target:   pkgJSON,
			failLeft: 1,
		}
		context := newBuildContext(base)
		sharedPath := tspath.ToPath(sharedConfig, root, context.FS().UseCaseSensitiveFileNames())
		host := context.newCompilerHost(root)

		first := host.GetResolvedProjectReference(sharedConfig, sharedPath)
		if first == nil {
			t.Fatal("expected diagnostic-free fallback without outDir on attempt 1, got nil")
		}
		if len(first.GetConfigFileParsingDiagnostics()) != 0 || first.CompilerOptions().OutDir != "" {
			t.Fatalf("expected diagnostic-free fallback without outDir on attempt 1, got diags=%d outDir=%q",
				len(first.GetConfigFileParsingDiagnostics()), first.CompilerOptions().OutDir)
		}
		second := host.GetResolvedProjectReference(sharedConfig, sharedPath)
		wantOutDir := tspath.ResolvePath(root, "packages/shared/dist")
		if second == nil {
			t.Fatal("second attempt returned nil after package.json recovered")
		}
		if second == first || second.CompilerOptions().OutDir != wantOutDir {
			t.Fatalf("expected retry to restore outDir=%q, got same=%v outDir=%q", wantOutDir, second == first, second.CompilerOptions().OutDir)
		}
	})

	t.Run("two_hop_extends_inaccessible_grandparent_does_not_poison_extended_cache", func(t *testing.T) {
		root := osvfs.FS().Realpath(tspath.NormalizePath(t.TempDir()))
		writeProgramBuildFixture(t, root, map[string]string{
			"packages/secret/grandparent.json": `{"compilerOptions":{"outDir":"${configDir}/dist","rootDir":"${configDir}/src"}}`,
			"packages/base.json":               `{"extends":"./secret/grandparent"}`,
			"packages/shared/tsconfig.json": `{
				"extends":"../base.json",
				"compilerOptions":{"composite":true,"noLib":true},
				"files":["src/value.ts"]
			}`,
			"packages/shared/src/value.ts": "export const value = 1;\n",
		})

		grandparent := tspath.ResolvePath(root, "packages/secret/grandparent.json")
		sharedConfig := tspath.ResolvePath(root, "packages/shared/tsconfig.json")
		base := &transientInaccessibleFS{
			FS:      bundled.WrapFS(osvfs.FS()),
			target:  grandparent,
			blocked: true,
		}
		context := newBuildContext(base)
		sharedPath := tspath.ToPath(sharedConfig, root, context.FS().UseCaseSensitiveFileNames())
		host := context.newCompilerHost(root)

		first := host.GetResolvedProjectReference(sharedConfig, sharedPath)
		if first == nil {
			t.Fatal("first attempt returned nil")
		}
		if first.CompilerOptions().OutDir != "" {
			t.Fatalf("first attempt unexpectedly inherited outDir=%q", first.CompilerOptions().OutDir)
		}
		if len(first.SourceToProjectReference()) != 1 {
			t.Fatalf("non-nil transient result did not finalize ParseInputOutputNames before returning: %d", len(first.SourceToProjectReference()))
		}
		base.setBlocked(false)
		second := host.GetResolvedProjectReference(sharedConfig, sharedPath)
		wantOutDir := tspath.ResolvePath(root, "packages/shared/dist")
		if second == nil {
			t.Fatal("second attempt returned nil after grandparent recovered")
		}
		if second == first || second.CompilerOptions().OutDir != wantOutDir {
			t.Fatalf("2-hop extended config cache was poisoned: same=%v outDir=%q want=%q",
				second == first, second.CompilerOptions().OutDir, wantOutDir)
		}
		third := host.GetResolvedProjectReference(sharedConfig, sharedPath)
		if third != second {
			t.Fatalf("healthy 2-hop config was not cached after recovery: second=%p third=%p", second, third)
		}
	})

	t.Run("two_hop_extends_inaccessible_package_dir_does_not_poison_extended_cache", func(t *testing.T) {
		root := osvfs.FS().Realpath(tspath.NormalizePath(t.TempDir()))
		writeProgramBuildFixture(t, root, map[string]string{
			"node_modules/@acme/tsconfig/package.json":      `{"name":"@acme/tsconfig","exports":{"./base.json":"./configs/base.json"}}`,
			"node_modules/@acme/tsconfig/configs/base.json": `{"compilerOptions":{"outDir":"${configDir}/dist","rootDir":"${configDir}/src"}}`,
			"packages/base.json":                            `{"extends":"@acme/tsconfig/base.json"}`,
			"packages/shared/tsconfig.json": `{
				"extends":"../base.json",
				"compilerOptions":{"composite":true,"noLib":true},
				"files":["src/value.ts"]
			}`,
			"packages/shared/src/value.ts": "export const value = 1;\n",
		})

		pkgJSON := tspath.ResolvePath(root, "node_modules/@acme/tsconfig/package.json")
		sharedConfig := tspath.ResolvePath(root, "packages/shared/tsconfig.json")
		base := &transientInaccessibleFS{
			FS:      bundled.WrapFS(osvfs.FS()),
			target:  pkgJSON,
			blocked: true,
		}
		context := newBuildContext(base)
		sharedPath := tspath.ToPath(sharedConfig, root, context.FS().UseCaseSensitiveFileNames())
		host := context.newCompilerHost(root)

		first := host.GetResolvedProjectReference(sharedConfig, sharedPath)
		if first == nil {
			t.Fatal("first attempt returned nil")
		}
		if first.CompilerOptions().OutDir != "" {
			t.Fatalf("first attempt unexpectedly inherited outDir=%q", first.CompilerOptions().OutDir)
		}
		base.setBlocked(false)
		second := host.GetResolvedProjectReference(sharedConfig, sharedPath)
		wantOutDir := tspath.ResolvePath(root, "packages/shared/dist")
		if second == nil {
			t.Fatal("second attempt returned nil after package.json recovered")
		}
		if second == first || second.CompilerOptions().OutDir != wantOutDir {
			t.Fatalf("2-hop package extends poisoned extended config cache: same=%v outDir=%q want=%q",
				second == first, second.CompilerOptions().OutDir, wantOutDir)
		}
	})

	t.Run("inaccessible_directory_file_exists_false_retries", func(t *testing.T) {
		root := osvfs.FS().Realpath(tspath.NormalizePath(t.TempDir()))
		writeProgramBuildFixture(t, root, map[string]string{
			"packages/shared/tsconfig.json": `{"compilerOptions":{"composite":true,"noLib":true},"files":["index.ts"]}`,
			"packages/shared/index.ts":      "export const shared = 1;\n",
		})

		sharedConfig := tspath.ResolvePath(root, "packages/shared/tsconfig.json")
		base := &transientInaccessibleFS{
			FS:      bundled.WrapFS(osvfs.FS()),
			target:  sharedConfig,
			blocked: true,
		}
		context := newBuildContext(base)
		sharedPath := tspath.ToPath(sharedConfig, root, context.FS().UseCaseSensitiveFileNames())
		host := context.newCompilerHost(root)

		if first := host.GetResolvedProjectReference(sharedConfig, sharedPath); first != nil {
			t.Fatalf("inaccessible directory lookup returned %v, want nil", first)
		}
		base.setBlocked(false)
		second := host.GetResolvedProjectReference(sharedConfig, sharedPath)
		if second == nil {
			t.Fatal("inaccessible directory nil result was cached after directory access recovered")
		}
		if third := host.GetResolvedProjectReference(sharedConfig, sharedPath); third != second {
			t.Fatalf("recovered reference was not cached: second=%p third=%p", second, third)
		}
	})
}

func TestMayMutateErrorsOnCommonSourceDirectoryZeroAllocAndEquivalence(t *testing.T) {
	t.Run("zero_alloc_on_1000_file_valid_in_place_config", func(t *testing.T) {
		const fileCount = 1000
		cases := []struct {
			name             string
			rootDir          string
			filePrefix       string
			useCaseSensitive bool
		}{
			{
				name:             "case_sensitive_posix",
				rootDir:          "/repo/packages/shared/src",
				filePrefix:       "/repo/packages/shared/src/mod_",
				useCaseSensitive: true,
			},
			{
				name:             "case_insensitive_mixed_case",
				rootDir:          "C:/Repo/Packages/Shared/Src/",
				filePrefix:       "c:/repo/packages/shared/src/mod_",
				useCaseSensitive: false,
			},
			{
				name:             "case_insensitive_non_ascii_exact_prefix",
				rootDir:          "/\u9879\u76ee/packages/shared/src",
				filePrefix:       "/\u9879\u76ee/packages/shared/src/mod_",
				useCaseSensitive: false,
			},
			{
				name:             "case_insensitive_non_ascii_mixed_ascii_casing",
				rootDir:          "/\u9879\u76ee/packages/shared/SRC",
				filePrefix:       "/\u9879\u76ee/packages/shared/src/mod_",
				useCaseSensitive: false,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				files := make([]string, fileCount)
				for i := range fileCount {
					files[i] = tc.filePrefix + strconv.Itoa(i) + ".ts"
				}
				cmd := tsoptions.NewParsedCommandLine(
					&core.CompilerOptions{
						Composite: core.TSTrue,
						RootDir:   tc.rootDir,
					},
					files,
					nil,
					tspath.ComparePathsOptions{
						CurrentDirectory:          "/repo/packages/shared",
						UseCaseSensitiveFileNames: tc.useCaseSensitive,
					},
				)
				if mayMutateErrorsOnCommonSourceDirectory(cmd) {
					t.Fatal("valid 1000-file in-place config reported outside files")
				}
				allocs := testing.AllocsPerRun(50, func() {
					if mayMutateErrorsOnCommonSourceDirectory(cmd) {
						t.Fatal("unexpected mutation flag")
					}
				})
				if allocs != 0 {
					t.Fatalf("mayMutateErrorsOnCommonSourceDirectory allocs/run = %v, want 0", allocs)
				}
			})
		}
	})

	t.Run("fast_path_soundness_and_tspath_equivalence", func(t *testing.T) {
		cases := []struct {
			root             string
			file             string
			useCaseSensitive bool
			wantFast         bool
			wantContains     bool
		}{
			{root: "/repo/src", file: "/repo/src/index.ts", useCaseSensitive: true, wantFast: true, wantContains: true},
			{root: "/repo/src/", file: "/repo/src/a/b.ts", useCaseSensitive: true, wantFast: true, wantContains: true},
			{root: "/repo/src", file: "/repo/src", useCaseSensitive: true, wantFast: true, wantContains: true},
			{root: "/repo/src", file: "/repo/src2/index.ts", useCaseSensitive: true, wantFast: false, wantContains: false},
			{root: "/repo/src", file: "/repo/src/../outside.ts", useCaseSensitive: true, wantFast: false, wantContains: false},
			{root: "/repo/src", file: "/repo/src/sub/../../outside.ts", useCaseSensitive: true, wantFast: false, wantContains: false},
			{root: "/repo/src", file: `/repo/src\..\outside.ts`, useCaseSensitive: true, wantFast: false, wantContains: false},
			{root: "/repo/src", file: "/repo/src/./index.ts", useCaseSensitive: true, wantFast: false, wantContains: true},
			{root: "/", file: "/index.ts", useCaseSensitive: true, wantFast: true, wantContains: true},
			{root: "/", file: "//server/share/index.ts", useCaseSensitive: true, wantFast: false, wantContains: false},
			{root: "C:/repo/src", file: "C:/repo/src/index.ts", useCaseSensitive: true, wantFast: true, wantContains: true},
			{root: "C:/repo/src", file: "c:/repo/src/index.ts", useCaseSensitive: true, wantFast: false, wantContains: true},
			{root: "C:/repo/src", file: "c:/REPO/src/index.ts", useCaseSensitive: false, wantFast: true, wantContains: true},
			{root: "C:/", file: "c:/index.ts", useCaseSensitive: false, wantFast: true, wantContains: true},
			{root: "C:", file: "C:/index.ts", useCaseSensitive: true, wantFast: false, wantContains: false},
			{root: "//", file: "//server/share/index.ts", useCaseSensitive: true, wantFast: false, wantContains: false},
			{root: "//", file: "//foo.ts", useCaseSensitive: true, wantFast: false, wantContains: false},
			{root: "//server/share/src", file: "//server/share/src/index.ts", useCaseSensitive: true, wantFast: true, wantContains: true},
			{root: "file:///repo/src", file: "file:///repo/src/index.ts", useCaseSensitive: false, wantFast: false, wantContains: true},
			{root: "/repo/src", file: "/repo/src/caf\u00e9.ts", useCaseSensitive: false, wantFast: true, wantContains: true},
			{root: "/repo/src", file: "/repo/src/caf\u00e9.ts", useCaseSensitive: true, wantFast: true, wantContains: true},
			{root: "/\u9879\u76ee/src", file: "/\u9879\u76ee/src/index.ts", useCaseSensitive: false, wantFast: true, wantContains: true},
			{root: "/\u9879\u76ee/src", file: "/\u9879\u76ee/src/caf\u00e9.ts", useCaseSensitive: false, wantFast: true, wantContains: true},
			{root: "/\u9879\u76ee/src", file: "/\u9879\u76ee/SRC/index.ts", useCaseSensitive: false, wantFast: true, wantContains: true},
			{root: "/\u9879\u76ee/caf\u00e9", file: "/\u9879\u76ee/CAF\u00c9/index.ts", useCaseSensitive: false, wantFast: false, wantContains: true},
			{root: "/\u9879\u76ee/src", file: "/\u9879\u76ee/src/../outside.ts", useCaseSensitive: false, wantFast: false, wantContains: false},
			{root: "src", file: "src/index.ts", useCaseSensitive: true, wantFast: false, wantContains: true},
		}

		for _, tc := range cases {
			opts := tspath.ComparePathsOptions{
				CurrentDirectory:          "/repo",
				UseCaseSensitiveFileNames: tc.useCaseSensitive,
			}
			cleanRoot, rootLen, ok := prepareCleanAbsoluteRoot(tc.root)
			gotFast := ok && containsCleanAbsolutePath(cleanRoot, rootLen, tc.file, tc.useCaseSensitive)
			gotContains := tspath.ContainsPath(tc.root, tc.file, opts)
			if gotFast != tc.wantFast {
				t.Errorf("fast path(%q, %q, cs=%v) = %v, want %v", tc.root, tc.file, tc.useCaseSensitive, gotFast, tc.wantFast)
			}
			if gotContains != tc.wantContains {
				t.Errorf("tspath.ContainsPath(%q, %q, cs=%v) = %v, want %v", tc.root, tc.file, tc.useCaseSensitive, gotContains, tc.wantContains)
			}
			if gotFast && !gotContains {
				t.Fatalf("fast path returned unsound true for (%q, %q, cs=%v)", tc.root, tc.file, tc.useCaseSensitive)
			}
		}
	})
}

type pureVFSProbeCountingFS struct {
	vfs.FS
	mu              sync.Mutex
	hidden          map[string]bool
	fileExistsCalls int
	dirExistsCalls  int
	readFileCalls   int
}

func (f *pureVFSProbeCountingFS) FileExists(path string) bool {
	f.mu.Lock()
	f.fileExistsCalls++
	hidden := f.hidden[path]
	f.mu.Unlock()
	if hidden {
		return false
	}
	return f.FS.FileExists(path)
}

func (f *pureVFSProbeCountingFS) DirectoryExists(path string) bool {
	f.mu.Lock()
	f.dirExistsCalls++
	f.mu.Unlock()
	return f.FS.DirectoryExists(path)
}

func (f *pureVFSProbeCountingFS) ReadFile(path string) (string, bool) {
	f.mu.Lock()
	f.readFileCalls++
	hidden := f.hidden[path]
	f.mu.Unlock()
	if hidden {
		return "", false
	}
	return f.FS.ReadFile(path)
}

func (f *pureVFSProbeCountingFS) setHidden(path string, hidden bool) {
	f.mu.Lock()
	f.hidden[path] = hidden
	f.mu.Unlock()
}

func (f *pureVFSProbeCountingFS) counts() (fileExists, dirExists, readFile int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.fileExistsCalls, f.dirExistsCalls, f.readFileCalls
}

func TestBuildContextProjectReferenceMissingPackageJSONPureVFSRevalidation(t *testing.T) {
	newVirtualFS := func(extra map[string]string, hidden map[string]bool) *pureVFSProbeCountingFS {
		mem := fstest.MapFS{
			"virtual-only/node_modules/@acme/tsconfig/package.json":      &fstest.MapFile{Data: []byte(`{"name":"@acme/tsconfig","exports":{"./base.json":"./configs/base.json"}}`)},
			"virtual-only/node_modules/@acme/tsconfig/base.json":         &fstest.MapFile{Data: []byte(`{"compilerOptions":{"rootDir":"${configDir}/src"}}`)},
			"virtual-only/node_modules/@acme/tsconfig/configs/base.json": &fstest.MapFile{Data: []byte(`{"compilerOptions":{"outDir":"${configDir}/dist","rootDir":"${configDir}/src"}}`)},
			"virtual-only/packages/shared/package.json":                  &fstest.MapFile{Data: []byte(`{"name":"@acme/shared"}`)},
			"virtual-only/packages/shared/src/value.ts":                  &fstest.MapFile{Data: []byte("export const value: number = 1;\n")},
		}
		for k, v := range extra {
			mem[k] = &fstest.MapFile{Data: []byte(v)}
		}
		return &pureVFSProbeCountingFS{
			FS:     bundled.WrapFS(iovfs.From(mem, true)),
			hidden: hidden,
		}
	}

	const (
		root         = "/virtual-only"
		sharedConfig = "/virtual-only/packages/shared/tsconfig.json"
		pkgJSON      = "/virtual-only/node_modules/@acme/tsconfig/package.json"
		wantOutDir   = "/virtual-only/packages/shared/dist"
	)

	t.Run("healthy_package_makes_zero_revalidation_probes_on_hit", func(t *testing.T) {
		virtualFS := newVirtualFS(map[string]string{
			"virtual-only/packages/shared/tsconfig.json": `{
				"extends":"@acme/tsconfig/base.json",
				"compilerOptions":{"composite":true,"noLib":true},
				"files":["src/value.ts"]
			}`,
		}, map[string]bool{})

		ctx := newBuildContext(virtualFS)
		sharedPath := tspath.ToPath(sharedConfig, root, ctx.FS().UseCaseSensitiveFileNames())
		host := ctx.newCompilerHost(root)

		first := host.GetResolvedProjectReference(sharedConfig, sharedPath)
		if first == nil {
			t.Fatal("first lookup returned nil")
		}
		if got := first.CompilerOptions().OutDir; got != wantOutDir {
			t.Fatalf("healthy outDir = %q, want %q", got, wantOutDir)
		}

		feBefore, deBefore, rfBefore := virtualFS.counts()
		second := host.GetResolvedProjectReference(sharedConfig, sharedPath)
		third := host.GetResolvedProjectReference(sharedConfig, sharedPath)
		feAfter, deAfter, rfAfter := virtualFS.counts()

		if second != first || third != first {
			t.Fatalf("healthy reference did not hit cache: first=%p second=%p third=%p", first, second, third)
		}
		if feAfter != feBefore || deAfter != deBefore || rfAfter != rfBefore {
			t.Fatalf("healthy cache hit performed extra VFS probes: fileExists=%d dirExists=%d readFile=%d",
				feAfter-feBefore, deAfter-deBefore, rfAfter-rfBefore)
		}
	})

	cases := []struct {
		name                   string
		files                  map[string]string
		wantProbesWhileHidden  int
		wantProbesAfterRecover int
	}{
		{
			name: "direct_package_extends",
			files: map[string]string{
				"virtual-only/packages/shared/tsconfig.json": `{
					"extends":"@acme/tsconfig/base.json",
					"compilerOptions":{"composite":true,"noLib":true},
					"files":["src/value.ts"]
				}`,
			},
			wantProbesWhileHidden:  1,
			wantProbesAfterRecover: 0,
		},
		{
			name: "two_hop_package_extends",
			files: map[string]string{
				"virtual-only/packages/base.json": `{"extends":"@acme/tsconfig/base.json"}`,
				"virtual-only/packages/shared/tsconfig.json": `{
					"extends":"../base.json",
					"compilerOptions":{"composite":true,"noLib":true},
					"files":["src/value.ts"]
				}`,
			},
			// /virtual-only/packages/base.json has no package.json in its ancestor
			// chain (/virtual-only/packages, /virtual-only, /), so self-name lookup
			// probes those 3 absent workspace manifests in addition to pkgJSON.
			wantProbesWhileHidden:  4,
			wantProbesAfterRecover: 3,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			virtualFS := newVirtualFS(tc.files, map[string]bool{pkgJSON: true})
			ctx := newBuildContext(virtualFS)
			sharedPath := tspath.ToPath(sharedConfig, root, ctx.FS().UseCaseSensitiveFileNames())
			host := ctx.newCompilerHost(root)

			// Lookup 1: package.json is absent in the pure in-memory VFS,
			// so legacy base.json resolves without outDir.
			first := host.GetResolvedProjectReference(sharedConfig, sharedPath)
			if first == nil {
				t.Fatal("first lookup returned nil")
			}
			if first.CompilerOptions().OutDir != "" {
				t.Fatalf("first lookup unexpectedly inherited outDir=%q", first.CompilerOptions().OutDir)
			}

			// Lookup 2: while package.json stays absent, hit revalidates only
			// the missing package.json paths via VFS and reuses cached fallback.
			feBefore, deBefore, rfBefore := virtualFS.counts()
			second := host.GetResolvedProjectReference(sharedConfig, sharedPath)
			feAfter, deAfter, rfAfter := virtualFS.counts()
			if second != first {
				t.Fatalf("expected cached fallback reuse while package.json remained missing: first=%p second=%p", first, second)
			}
			if feAfter-feBefore != tc.wantProbesWhileHidden || deAfter-deBefore != 0 || rfAfter-rfBefore != 0 {
				t.Fatalf("revalidation probe delta = (fileExists=%d, dirExists=%d, readFile=%d), want (%d, 0, 0)",
					feAfter-feBefore, deAfter-deBefore, rfAfter-rfBefore, tc.wantProbesWhileHidden)
			}

			// Reveal package.json purely in the in-memory VFS (no WriteFile invalidation,
			// no host filesystem path at all) and verify eviction + clean recovery.
			virtualFS.setHidden(pkgJSON, false)
			third := host.GetResolvedProjectReference(sharedConfig, sharedPath)
			if third == nil {
				t.Fatal("third lookup returned nil after package.json appeared in VFS")
			}
			if third == first {
				t.Fatal("stale fallback remained cached after package.json appeared in VFS")
			}
			if got := third.CompilerOptions().OutDir; got != wantOutDir {
				t.Fatalf("recovered outDir = %q, want %q", got, wantOutDir)
			}

			// Lookup 4: recovered entry only revalidates any remaining absent
			// ancestor package.json probes (0 for a package with its own manifest).
			feBefore, deBefore, rfBefore = virtualFS.counts()
			fourth := host.GetResolvedProjectReference(sharedConfig, sharedPath)
			feAfter, deAfter, rfAfter = virtualFS.counts()
			if fourth != third {
				t.Fatalf("recovered healthy reference was not cached: third=%p fourth=%p", third, fourth)
			}
			if feAfter-feBefore != tc.wantProbesAfterRecover || deAfter != deBefore || rfAfter != rfBefore {
				t.Fatalf("recovered hit probe delta = (fileExists=%d, dirExists=%d, readFile=%d), want (%d, 0, 0)",
					feAfter-feBefore, deAfter-deBefore, rfAfter-rfBefore, tc.wantProbesAfterRecover)
			}
		})
	}

	t.Run("workspace_self_name_manifest_recovery", func(t *testing.T) {
		const workspacePkgJSON = "/virtual-only/packages/shared/package.json"
		virtualFS := newVirtualFS(map[string]string{
			"virtual-only/packages/shared/package.json": `{"name":"@acme/tsconfig","exports":{"./base.json":"./local-base.json"}}`,
			"virtual-only/packages/shared/local-base.json": `{
				"compilerOptions":{"outDir":"${configDir}/local-dist","rootDir":"${configDir}/src"}
			}`,
			"virtual-only/packages/shared/tsconfig.json": `{
				"extends":"@acme/tsconfig/base.json",
				"compilerOptions":{"composite":true,"noLib":true},
				"files":["src/value.ts"]
			}`,
		}, map[string]bool{workspacePkgJSON: true})

		ctx := newBuildContext(virtualFS)
		sharedPath := tspath.ToPath(sharedConfig, root, ctx.FS().UseCaseSensitiveFileNames())
		host := ctx.newCompilerHost(root)

		// Lookup 1: local package.json is inaccessible, so resolution falls back to node_modules.
		first := host.GetResolvedProjectReference(sharedConfig, sharedPath)
		if first == nil {
			t.Fatal("first lookup returned nil")
		}
		if got := first.CompilerOptions().OutDir; got != wantOutDir {
			t.Fatalf("fallback outDir = %q, want %q", got, wantOutDir)
		}

		// Reveal workspace package.json: cached fallback must evict and resolve local self-name export.
		virtualFS.setHidden(workspacePkgJSON, false)
		recovered := host.GetResolvedProjectReference(sharedConfig, sharedPath)
		if recovered == nil || recovered == first {
			t.Fatalf("expected eviction after workspace package.json recovered: first=%p recovered=%p", first, recovered)
		}
		const wantLocalOutDir = "/virtual-only/packages/shared/local-dist"
		if got := recovered.CompilerOptions().OutDir; got != wantLocalOutDir {
			t.Fatalf("recovered workspace self-name outDir = %q, want %q", got, wantLocalOutDir)
		}
	})

	t.Run("nested_npm_preset_without_subdir_manifest_shares_leaf_config", func(t *testing.T) {
		const leafConfig = "/virtual-only/node_modules/@acme/leaf/base.json"
		const presetSubdirPkg = "/virtual-only/node_modules/@acme/presets/presets/package.json"
		extra := map[string]string{
			"virtual-only/node_modules/@acme/presets/package.json":          `{"name":"@acme/presets"}`,
			"virtual-only/node_modules/@acme/presets/presets/browser.json":  `{"extends":"@acme/leaf/base.json"}`,
			"virtual-only/node_modules/@acme/presets/presets/package.json":  `{"name":"@acme/leaf","exports":{"./base.json":"./override.json"}}`,
			"virtual-only/node_modules/@acme/presets/presets/override.json": `{"compilerOptions":{"outDir":"${configDir}/override-dist","rootDir":"${configDir}/src"}}`,
			"virtual-only/node_modules/@acme/leaf/package.json":             `{"name":"@acme/leaf"}`,
			"virtual-only/node_modules/@acme/leaf/base.json":                `{"compilerOptions":{"outDir":"${configDir}/dist","rootDir":"${configDir}/src"}}`,
		}
		for i := range 8 {
			extra[fmt.Sprintf("virtual-only/packages/p%d/package.json", i)] = fmt.Sprintf(`{"name":"@acme/p%d"}`, i)
			extra[fmt.Sprintf("virtual-only/packages/p%d/src/index.ts", i)] = "export const v = 1;\n"
			extra[fmt.Sprintf("virtual-only/packages/p%d/tsconfig.json", i)] = `{
				"extends":"@acme/presets/presets/browser.json",
				"compilerOptions":{"composite":true,"noLib":true},
				"files":["src/index.ts"]
			}`
		}
		baseFS := newVirtualFS(extra, map[string]bool{presetSubdirPkg: true})
		var leafExistsCalls int
		countingFS := &pathProbeCounterFS{
			FS: baseFS,
			onFileExists: func(p string) {
				if p == leafConfig {
					leafExistsCalls++
				}
			},
		}

		ctx := newBuildContext(countingFS)
		host := ctx.newCompilerHost(root)
		for i := range 8 {
			cfg := fmt.Sprintf("/virtual-only/packages/p%d/tsconfig.json", i)
			cfgPath := tspath.ToPath(cfg, root, ctx.FS().UseCaseSensitiveFileNames())
			parsed := host.GetResolvedProjectReference(cfg, cfgPath)
			if parsed == nil || parsed.CompilerOptions().OutDir != fmt.Sprintf("/virtual-only/packages/p%d/dist", i) {
				t.Fatalf("p%d outDir = %v, want /virtual-only/packages/p%d/dist", i, parsed, i)
			}
		}
		if leafExistsCalls != 1 {
			t.Fatalf("shared leaf config FileExists count across 8 references = %d, want 1", leafExistsCalls)
		}

		// Reveal presets/package.json: both the shared extended preset and the
		// cached project references (both cold-miss p0 and warm-hit p1) must
		// evict and resolve the new override.
		baseFS.setHidden(presetSubdirPkg, false)
		for _, i := range []int{0, 1} {
			cfg := fmt.Sprintf("/virtual-only/packages/p%d/tsconfig.json", i)
			cfgPath := tspath.ToPath(cfg, root, ctx.FS().UseCaseSensitiveFileNames())
			reloaded := host.GetResolvedProjectReference(cfg, cfgPath)
			wantOverride := fmt.Sprintf("/virtual-only/packages/p%d/override-dist", i)
			if reloaded == nil || reloaded.CompilerOptions().OutDir != wantOverride {
				t.Fatalf("p%d after presets/package.json appeared, outDir = %v, want %q", i, reloaded, wantOverride)
			}
		}
	})

	t.Run("parseConfig_warmed_extended_config_still_revalidates_for_project_references", func(t *testing.T) {
		virtualFS := newVirtualFS(map[string]string{
			"virtual-only/packages/base.json":        `{"extends":"@acme/tsconfig/base.json"}`,
			"virtual-only/packages/app/package.json": `{"name":"@acme/app"}`,
			"virtual-only/packages/app/src/index.ts": "export const app = 1;\n",
			"virtual-only/packages/app/tsconfig.json": `{
				"extends":"../base.json",
				"compilerOptions":{"noLib":true},
				"files":["src/index.ts"]
			}`,
			"virtual-only/packages/shared/tsconfig.json": `{
				"extends":"../base.json",
				"compilerOptions":{"composite":true,"noLib":true},
				"files":["src/value.ts"]
			}`,
		}, map[string]bool{pkgJSON: true})

		ctx := newBuildContext(virtualFS)
		// Warm ../base.json via top-level parseConfig while pkgJSON is hidden.
		appCfg, err := ctx.parseConfig("/virtual-only/packages/app", "/virtual-only/packages/app/tsconfig.json")
		if err != nil || appCfg == nil {
			t.Fatalf("parseConfig(app) failed: cfg=%v err=%v", appCfg, err)
		}
		if appCfg.CompilerOptions().OutDir != "" {
			t.Fatalf("initial app outDir = %q, want empty", appCfg.CompilerOptions().OutDir)
		}

		// Resolve shared via project reference cache: must inherit missing pkgJSON
		// from the parseConfig-warmed base.json entry.
		sharedPath := tspath.ToPath(sharedConfig, root, ctx.FS().UseCaseSensitiveFileNames())
		host := ctx.newCompilerHost(root)
		first := host.GetResolvedProjectReference(sharedConfig, sharedPath)
		if first == nil || first.CompilerOptions().OutDir != "" {
			t.Fatalf("initial shared ref = %v, want empty outDir", first)
		}

		// Reveal pkgJSON: both parseConfig and GetResolvedProjectReference must recover.
		virtualFS.setHidden(pkgJSON, false)
		recoveredApp, err := ctx.parseConfig("/virtual-only/packages/app", "/virtual-only/packages/app/tsconfig.json")
		if err != nil || recoveredApp == nil || recoveredApp.CompilerOptions().OutDir != "/virtual-only/packages/app/dist" {
			t.Fatalf("recovered parseConfig(app) = (%v, %v), want /virtual-only/packages/app/dist", recoveredApp, err)
		}
		recoveredShared := host.GetResolvedProjectReference(sharedConfig, sharedPath)
		if recoveredShared == nil || recoveredShared == first || recoveredShared.CompilerOptions().OutDir != wantOutDir {
			t.Fatalf("recovered shared ref = %v, want %q", recoveredShared, wantOutDir)
		}
	})

	t.Run("missing_manifest_probes_deduplicated_within_single_parse", func(t *testing.T) {
		const (
			layerCount      = 8
			presetSubdirPkg = "/virtual-only/node_modules/@acme/presets/presets/package.json"
		)
		extra := map[string]string{
			"virtual-only/node_modules/@acme/presets/package.json":         `{"name":"@acme/presets"}`,
			"virtual-only/node_modules/@acme/presets/presets/browser.json": `{"extends":"@acme/leaf/base.json"}`,
			"virtual-only/node_modules/@acme/leaf/package.json":            `{"name":"@acme/leaf"}`,
			"virtual-only/node_modules/@acme/leaf/base.json":               `{"compilerOptions":{"outDir":"${configDir}/dist","rootDir":"${configDir}/src"}}`,
		}
		// Build an 8-layer local extends chain ending at @acme/presets/presets/browser.json.
		for i := range layerCount {
			target := "@acme/presets/presets/browser.json"
			if i > 0 {
				target = fmt.Sprintf("./layer-%d.json", i-1)
			}
			extra[fmt.Sprintf("virtual-only/packages/layered/layer-%d.json", i)] = fmt.Sprintf(`{"extends":%q}`, target)
		}
		extra["virtual-only/packages/layered/package.json"] = `{"name":"@acme/layered"}`
		extra["virtual-only/packages/layered/src/index.ts"] = "export const layered = 1;\n"
		extra["virtual-only/packages/layered/tsconfig.json"] = fmt.Sprintf(`{
			"extends":"./layer-%d.json",
			"compilerOptions":{"composite":true,"noLib":true},
			"files":["src/index.ts"]
		}`, layerCount-1)

		var pkgChecks int
		baseFS := newVirtualFS(extra, map[string]bool{})
		countingFS := &pathProbeCounterFS{
			FS: baseFS,
			onFileExists: func(p string) {
				if p == presetSubdirPkg {
					pkgChecks++
				}
			},
		}
		ctx := newBuildContext(countingFS)
		host := ctx.newCompilerHost(root)
		layeredCfg := "/virtual-only/packages/layered/tsconfig.json"
		layeredPath := tspath.ToPath(layeredCfg, root, ctx.FS().UseCaseSensitiveFileNames())
		parsed := host.GetResolvedProjectReference(layeredCfg, layeredPath)
		if parsed == nil || parsed.CompilerOptions().OutDir != "/virtual-only/packages/layered/dist" {
			t.Fatalf("layered ref = %v, want outDir=/virtual-only/packages/layered/dist", parsed)
		}
		if pkgChecks != 1 {
			t.Fatalf("8-layer cold project reference checked %s %d times, want 1", presetSubdirPkg, pkgChecks)
		}

		// Eight distinct project references sharing the same nested npm preset
		// must probe presets/package.json at most once per reference (8 total).
		extraMulti := map[string]string{
			"virtual-only/node_modules/@acme/presets/package.json":         `{"name":"@acme/presets"}`,
			"virtual-only/node_modules/@acme/presets/presets/browser.json": `{"extends":"@acme/leaf/base.json"}`,
			"virtual-only/node_modules/@acme/leaf/package.json":            `{"name":"@acme/leaf"}`,
			"virtual-only/node_modules/@acme/leaf/base.json":               `{"compilerOptions":{"outDir":"${configDir}/dist","rootDir":"${configDir}/src"}}`,
		}
		const refCount = 8
		for i := range refCount {
			extraMulti[fmt.Sprintf("virtual-only/packages/r%d/package.json", i)] = fmt.Sprintf(`{"name":"@acme/r%d"}`, i)
			extraMulti[fmt.Sprintf("virtual-only/packages/r%d/src/index.ts", i)] = "export const v = 1;\n"
			extraMulti[fmt.Sprintf("virtual-only/packages/r%d/tsconfig.json", i)] = `{
				"extends":"@acme/presets/presets/browser.json",
				"compilerOptions":{"composite":true,"noLib":true},
				"files":["src/index.ts"]
			}`
		}
		pkgChecks = 0
		multiFS := &pathProbeCounterFS{
			FS: newVirtualFS(extraMulti, map[string]bool{}),
			onFileExists: func(p string) {
				if p == presetSubdirPkg {
					pkgChecks++
				}
			},
		}
		multiCtx := newBuildContext(multiFS)
		multiHost := multiCtx.newCompilerHost(root)
		for i := range refCount {
			cfg := fmt.Sprintf("/virtual-only/packages/r%d/tsconfig.json", i)
			cfgPath := tspath.ToPath(cfg, root, multiCtx.FS().UseCaseSensitiveFileNames())
			ref := multiHost.GetResolvedProjectReference(cfg, cfgPath)
			if ref == nil || ref.CompilerOptions().OutDir != fmt.Sprintf("/virtual-only/packages/r%d/dist", i) {
				t.Fatalf("r%d ref = %v, want /virtual-only/packages/r%d/dist", i, ref, i)
			}
		}
		if pkgChecks != refCount {
			t.Fatalf("8 distinct references sharing nested preset checked %s %d times, want %d", presetSubdirPkg, pkgChecks, refCount)
		}
	})
}

type blockingTransientReadFS struct {
	vfs.FS
	mu           sync.Mutex
	failTarget   string
	failRead     bool
	pauseDir     string
	paused       bool
	pauseEntered chan struct{}
	pauseRelease chan struct{}
}

func (f *blockingTransientReadFS) ReadFile(path string) (string, bool) {
	f.mu.Lock()
	shouldFail := path == f.failTarget && f.failRead
	f.mu.Unlock()
	if shouldFail {
		return "", false
	}
	return f.FS.ReadFile(path)
}

func (f *blockingTransientReadFS) GetAccessibleEntries(path string) vfs.Entries {
	normalized := tspath.RemoveTrailingDirectorySeparator(tspath.NormalizePath(path))
	f.mu.Lock()
	shouldPause := normalized == f.pauseDir && !f.paused
	if shouldPause {
		f.paused = true
	}
	f.mu.Unlock()
	if shouldPause {
		close(f.pauseEntered)
		<-f.pauseRelease
	}
	return f.FS.GetAccessibleEntries(path)
}

func (f *blockingTransientReadFS) setFailRead(fail bool) {
	f.mu.Lock()
	f.failRead = fail
	f.mu.Unlock()
}

func TestBuildContextProjectReferenceWaiterRetriesAfterTransientReadFailure(t *testing.T) {
	root := osvfs.FS().Realpath(tspath.NormalizePath(t.TempDir()))
	writeProgramBuildFixture(t, root, map[string]string{
		"node_modules/@acme/tsconfig/package.json":      `{"name":"@acme/tsconfig","exports":{"./base.json":"./configs/base.json"}}`,
		"node_modules/@acme/tsconfig/base.json":         `{"compilerOptions":{"rootDir":"${configDir}/src"}}`,
		"node_modules/@acme/tsconfig/configs/base.json": `{"compilerOptions":{"outDir":"${configDir}/dist","rootDir":"${configDir}/src"}}`,
		"packages/shared/tsconfig.json": `{
			"extends":"@acme/tsconfig/base.json",
			"compilerOptions":{"composite":true,"noLib":true},
			"include":["src/**/*"]
		}`,
		"packages/shared/src/value.ts":    "export const value: number = 1;\n",
		"packages/shared/dist/value.d.ts": "export declare const value: number;\n",
	})

	pkgJSON := tspath.ResolvePath(root, "node_modules/@acme/tsconfig/package.json")
	sharedConfig := tspath.ResolvePath(root, "packages/shared/tsconfig.json")
	sharedSrcDir := tspath.ResolvePath(root, "packages/shared/src")
	base := &blockingTransientReadFS{
		FS:           bundled.WrapFS(osvfs.FS()),
		failTarget:   pkgJSON,
		failRead:     true,
		pauseDir:     sharedSrcDir,
		pauseEntered: make(chan struct{}),
		pauseRelease: make(chan struct{}),
	}
	ctx := newBuildContext(base)
	sharedPath := tspath.ToPath(sharedConfig, root, ctx.FS().UseCaseSensitiveFileNames())

	firstDone := make(chan *tsoptions.ParsedCommandLine, 1)
	go func() {
		firstDone <- ctx.newCompilerHost(root).GetResolvedProjectReference(sharedConfig, sharedPath)
	}()

	select {
	case <-base.pauseEntered:
	case <-time.After(5 * time.Second):
		t.Fatal("first caller did not reach source glob scan")
	}

	// Restore read access while caller A is still scanning source files,
	// then launch caller B so it joins A's in-flight lookup.
	base.setFailRead(false)

	secondStarted := make(chan struct{})
	secondDone := make(chan *tsoptions.ParsedCommandLine, 1)
	go func() {
		host := ctx.newCompilerHost(root)
		close(secondStarted)
		secondDone <- host.GetResolvedProjectReference(sharedConfig, sharedPath)
	}()

	<-secondStarted
	// Because pauseDir pauses only on caller A's first scan, if caller B had
	// bypassed A's flight it would complete immediately before pauseRelease.
	select {
	case ref := <-secondDone:
		t.Fatalf("caller B completed before caller A released its flight: %v", ref)
	case <-time.After(25 * time.Millisecond):
	}
	close(base.pauseRelease)

	var first, second *tsoptions.ParsedCommandLine
	select {
	case first = <-firstDone:
	case <-time.After(5 * time.Second):
		t.Fatal("first caller timed out")
	}
	select {
	case second = <-secondDone:
	case <-time.After(5 * time.Second):
		t.Fatal("second caller timed out")
	}
	if first == nil || first.CompilerOptions().OutDir != "" {
		t.Fatalf("first caller (started before recovery) = %v, want fallback without outDir", first)
	}
	wantOutDir := tspath.ResolvePath(root, "packages/shared/dist")
	if second == nil {
		t.Fatal("concurrent waiter returned nil instead of retrying after transient read failure")
	}
	if second == first || second.CompilerOptions().OutDir != wantOutDir {
		t.Fatalf("concurrent waiter inherited failed transient read: same=%v outDir=%q want=%q",
			second == first, second.CompilerOptions().OutDir, wantOutDir)
	}
}

func TestIsNodeModulesPackageDir(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"/repo/node_modules/preset", true},
		{"/repo/node_modules/preset/", true},
		{"/repo/node_modules/@acme/preset", true},
		{"/repo/node_modules/@acme/preset/", true},
		{`C:\repo\node_modules\@acme\preset`, true},
		{"/repo/node_modules/a/node_modules/@acme/preset", true},
		{"node_modules/preset", true},
		{"node_modules/@acme/preset", true},
		{"/repo/node_modules", false},
		{"/repo/node_modules/", false},
		{"/repo/node_modules/@acme", false},
		{"/repo/node_modules/@acme/", false},
		{"/repo/node_modules/@acme/preset/base.json", false},
		{"/repo/node_modules/preset/base.json", false},
		{"/repo/node_modules//preset", false},
		{"/repo/node_modules/@acme//preset", false},
		{"/repo/my_node_modules/preset", false},
		{"/repo/node_modules_bak/preset", false},
	}
	for _, tc := range cases {
		if got := isNodeModulesPackageDir(tc.path); got != tc.want {
			t.Errorf("isNodeModulesPackageDir(%q) = %v, want %v", tc.path, got, tc.want)
		}
	}
}

func TestBuildContextProjectReferenceNearerPackageDirSymlinkRecovery(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX directory permissions required")
	}

	root := osvfs.FS().Realpath(tspath.NormalizePath(t.TempDir()))
	writeProgramBuildFixture(t, root, map[string]string{
		"node_modules/@acme/preset/package.json": `{"name":"@acme/preset"}`,
		"node_modules/@acme/preset/base.json":    `{"compilerOptions":{"rootDir":"${configDir}/src"}}`,
		"secret/preset/package.json":             `{"name":"@acme/preset"}`,
		"secret/preset/base.json":                `{"compilerOptions":{"outDir":"${configDir}/dist","rootDir":"${configDir}/src"}}`,
		"packages/shared/tsconfig.json": `{
			"extends":"@acme/preset/base.json",
			"compilerOptions":{"composite":true,"noLib":true},
			"files":["src/value.ts"]
		}`,
		"packages/shared/src/value.ts":    "export const value: number = 1;\n",
		"packages/shared/dist/value.d.ts": "export declare const value: number;\n",
		"packages/app/tsconfig.json": `{
			"compilerOptions":{"noLib":true,"disableSourceOfProjectReferenceRedirect":true},
			"files":["index.ts"],
			"references":[{"path":"../shared"}]
		}`,
		"packages/app/index.ts": "import { value } from \"../shared/src/value\";\nexport const doubled: number = value * 2;\n",
	})

	secretDir := filepath.Join(root, "secret")
	secretPreset := filepath.Join(secretDir, "preset")
	localScopeDir := filepath.Join(root, "packages/shared/node_modules/@acme")
	if err := os.MkdirAll(localScopeDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(%s): %v", localScopeDir, err)
	}
	localPresetLink := filepath.Join(localScopeDir, "preset")
	if err := os.Symlink(secretPreset, localPresetLink); err != nil {
		t.Fatalf("Symlink(%s -> %s): %v", localPresetLink, secretPreset, err)
	}
	if err := os.Chmod(secretDir, 0); err != nil {
		t.Fatalf("Chmod(0): %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(secretDir, 0o755) })
	if _, err := os.Stat(localPresetLink); !errors.Is(err, fs.ErrPermission) {
		t.Skipf("OS did not enforce chmod 000 on symlink target directory (likely root): err=%v", err)
	}

	appDir := tspath.ResolvePath(root, "packages/app")
	appConfig := tspath.ResolvePath(root, "packages/app/tsconfig.json")
	appEntry := tspath.ResolvePath(root, "packages/app/index.ts")
	sharedConfig := tspath.ResolvePath(root, "packages/shared/tsconfig.json")

	ctx := newBuildContext(bundled.WrapFS(osvfs.FS()))
	sharedPath := tspath.ToPath(sharedConfig, root, ctx.FS().UseCaseSensitiveFileNames())

	// Program 1: nearer package dir symlink is inaccessible -> selects root copy (no outDir -> TS6305).
	firstProg, err := ctx.createProjectProgram(true, appDir, appConfig)
	if err != nil {
		t.Fatalf("first program build: %v", err)
	}
	firstRef, ok := firstProg.GetResolvedProjectReferenceFor(sharedPath)
	if !ok || firstRef == nil || firstRef.CompilerOptions().OutDir != "" {
		t.Fatalf("first program shared ref = (%v, %v), want fallback without outDir", firstRef, ok)
	}

	// Restore access to nearer package dir symlink -> Program 2 must evict fallback and clear TS6305.
	if err := os.Chmod(secretDir, 0o755); err != nil {
		t.Fatalf("Chmod(0755): %v", err)
	}
	secondProg, err := ctx.createProjectProgram(true, appDir, appConfig)
	if err != nil {
		t.Fatalf("second program build: %v", err)
	}
	secondRef, ok := secondProg.GetResolvedProjectReferenceFor(sharedPath)
	if !ok || secondRef == nil || secondRef == firstRef {
		t.Fatalf("stale root-copy fallback remained cached after nearer package dir recovered: first=%p second=%p", firstRef, secondRef)
	}
	wantOutDir := tspath.ResolvePath(root, "packages/shared/dist")
	if got := secondRef.CompilerOptions().OutDir; got != wantOutDir {
		t.Fatalf("recovered outDir = %q, want %q", got, wantOutDir)
	}
	secondSF := secondProg.GetSourceFile(appEntry)
	if secondSF == nil {
		t.Fatal("second program missing app/index.ts")
	}
	if diags := secondProg.GetSemanticDiagnostics(context.Background(), secondSF); len(diags) != 0 {
		t.Fatalf("expected 0 semantic diagnostics after nearer package dir recovery, got %d (first code=%d)", len(diags), diags[0].Code())
	}
}

type pathProbeCounterFS struct {
	vfs.FS
	onFileExists func(path string)
}

func (f *pathProbeCounterFS) FileExists(path string) bool {
	if f.onFileExists != nil {
		f.onFileExists(path)
	}
	return f.FS.FileExists(path)
}

type transientProjectReferenceReadFS struct {
	vfs.FS
	mu       sync.Mutex
	target   string
	failLeft int
	reads    int
}

func (f *transientProjectReferenceReadFS) ReadFile(path string) (string, bool) {
	if path == f.target {
		f.mu.Lock()
		f.reads++
		shouldFail := f.failLeft > 0
		if shouldFail {
			f.failLeft--
		}
		f.mu.Unlock()
		if shouldFail {
			return "", false
		}
	}
	return f.FS.ReadFile(path)
}

func TestBuildContextProjectReferenceDoesNotCacheTransientReadFailure(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	writeProgramBuildFixture(t, root, map[string]string{
		"packages/shared/tsconfig.json": `{"compilerOptions":{"composite":true,"noLib":true},"files":["index.ts"]}`,
		"packages/shared/index.ts":      "export const shared = 1;\n",
	})

	sharedConfig := tspath.ResolvePath(root, "packages/shared/tsconfig.json")
	base := &transientProjectReferenceReadFS{
		FS:       bundled.WrapFS(osvfs.FS()),
		target:   sharedConfig,
		failLeft: 1,
	}
	context := newBuildContext(base)
	sharedPath := tspath.ToPath(sharedConfig, root, context.FS().UseCaseSensitiveFileNames())
	host := context.newCompilerHost(root)

	if first := host.GetResolvedProjectReference(sharedConfig, sharedPath); first != nil {
		t.Fatalf("transient read failure returned %v, want nil", first)
	}

	second := host.GetResolvedProjectReference(sharedConfig, sharedPath)
	if second == nil {
		t.Fatal("subsequent healthy read returned cached nil after transient failure")
	}
	third := host.GetResolvedProjectReference(sharedConfig, sharedPath)
	if third != second {
		t.Fatalf("subsequent cached read returned distinct instance: %p != %p", third, second)
	}

	base.mu.Lock()
	reads := base.reads
	base.mu.Unlock()
	if reads != 2 {
		t.Fatalf("underlying ReadFile calls = %d, want 2 (1 failed + 1 cached)", reads)
	}
}

type panicOnceProjectReferenceReadFS struct {
	vfs.FS
	mu      sync.Mutex
	target  string
	reads   int
	entered chan struct{}
	release chan struct{}
}

func (f *panicOnceProjectReferenceReadFS) ReadFile(path string) (string, bool) {
	if path == f.target {
		f.mu.Lock()
		f.reads++
		attempt := f.reads
		f.mu.Unlock()
		if attempt == 1 {
			close(f.entered)
			<-f.release
			panic("project reference read panic")
		}
	}
	return f.FS.ReadFile(path)
}

func TestBuildContextProjectReferenceSingleFlightPanicRecovery(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	writeProgramBuildFixture(t, root, map[string]string{
		"packages/shared/tsconfig.json": `{"compilerOptions":{"composite":true,"noLib":true},"files":["index.ts"]}`,
		"packages/shared/index.ts":      "export const shared = 1;\n",
	})

	sharedConfig := tspath.ResolvePath(root, "packages/shared/tsconfig.json")
	base := &panicOnceProjectReferenceReadFS{
		FS:      bundled.WrapFS(osvfs.FS()),
		target:  sharedConfig,
		entered: make(chan struct{}),
		release: make(chan struct{}),
	}
	context := newBuildContext(base)
	sharedPath := tspath.ToPath(sharedConfig, root, context.FS().UseCaseSensitiveFileNames())

	ownerPanicked := make(chan any, 1)
	go func() {
		defer func() {
			ownerPanicked <- recover()
		}()
		host := context.newCompilerHost(root)
		_ = host.GetResolvedProjectReference(sharedConfig, sharedPath)
	}()
	<-base.entered

	key := programProjectReferenceKey{fileName: sharedConfig, path: sharedPath}
	raw, ok := context.metadataFS.currentProjectReferenceGeneration().entries.Load(key)
	if !ok {
		t.Fatal("in-flight project reference entry not registered")
	}
	flight, ok := raw.(*projectReferenceFlight)
	if !ok {
		t.Fatalf("in-flight entry type = %T, want *projectReferenceFlight", raw)
	}

	const waiters = 4
	waiterDone := make(chan *tsoptions.ParsedCommandLine, waiters)
	for range waiters {
		go func() {
			host := context.newCompilerHost(root)
			waiterDone <- host.GetResolvedProjectReference(sharedConfig, sharedPath)
		}()
	}

	close(base.release)

	select {
	case recovered := <-ownerPanicked:
		if recovered == nil {
			t.Fatal("expected single-flight owner to panic")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("panicking single-flight owner did not exit")
	}

	select {
	case <-flight.ready:
	case <-time.After(5 * time.Second):
		t.Fatal("panic stranded single-flight waiters")
	}
	if flight.completed {
		t.Fatal("panicked flight was marked completed")
	}

	var winner *tsoptions.ParsedCommandLine
	for i := range waiters {
		select {
		case ref := <-waiterDone:
			if ref == nil {
				t.Fatalf("waiter %d received nil after panic recovery", i)
			}
			if winner == nil {
				winner = ref
			} else if ref != winner {
				t.Fatalf("waiter %d observed distinct *ParsedCommandLine after panic recovery", i)
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("waiter %d hung after owner panic", i)
		}
	}
}

func TestBuildContextWriteInvalidatesExtendedConfigGeneration(t *testing.T) {
	root := t.TempDir()
	writeProgramBuildFixture(t, root, map[string]string{
		"base.json": `{"compilerOptions":{"strict":true}}`,
		"a.json":    `{"extends":"./base.json","files":[]}`,
		"b.json":    `{"extends":"./base.json","files":[]}`,
	})

	context := newBuildContext(bundled.WrapFS(cachedvfs.From(osvfs.FS())))
	before, err := context.parseConfig(root, filepath.Join(root, "a.json"))
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
	after, err := context.parseConfig(root, filepath.Join(root, "b.json"))
	if err != nil {
		t.Fatalf("parse config after write: %v", err)
	}
	if got := after.CompilerOptions().Strict; got != core.TSFalse {
		t.Fatalf("strict after write = %v, want false", got)
	}
}

func TestBuildContextRequestIsolationAndEscapeHatch(t *testing.T) {
	const path = "/repo/package.json"
	base := newProgramMetadataTestFS(map[string]string{path: "first"})
	first := newBuildContext(base)
	if content, _ := first.FS().ReadFile(path); content != "first" {
		t.Fatalf("first request read %q", content)
	}
	base.set(path, "second")
	second := newBuildContext(base)
	if content, _ := second.FS().ReadFile(path); content != "second" {
		t.Fatalf("second request inherited stale metadata %q", content)
	}
	if content, _ := first.FS().ReadFile(path); content != "first" {
		t.Fatalf("first request snapshot changed to %q", content)
	}

	t.Setenv("RSLINT_DISABLE_PROGRAM_METADATA_CACHE", "1")
	disabled := newBuildContext(base)
	if disabled.FS() != vfs.FS(base) {
		t.Fatal("escape hatch must preserve the exact caller VFS")
	}
	if disabled.metadataFS != nil || disabled.extendedConfigCache != nil {
		t.Fatal("escape hatch left metadata caches enabled")
	}
	disabled.enableConcurrentProgramQueries()
	if hostFS := disabled.newCompilerHostWithCache("/").FS(); hostFS != disabled.FS() {
		t.Fatal("escape hatch enabled the parallel Program VFS")
	}

	root := tspath.NormalizePath(t.TempDir())
	writeProgramBuildFixture(t, root, map[string]string{
		"base.json":            `{"compilerOptions":{"strict":true}}`,
		"tsconfig.json":        `{"extends":"./base.json","files":[],"references":[{"path":"./shared"}]}`,
		"shared/tsconfig.json": `{"compilerOptions":{"composite":true,"noLib":true},"files":["index.ts"]}`,
		"shared/index.ts":      "export const shared = 1;\n",
	})
	disabled = newBuildContext(bundled.WrapFS(cachedvfs.From(osvfs.FS())))
	firstDisabled, err := disabled.createProjectProgram(true, root, tspath.ResolvePath(root, "tsconfig.json"))
	if err != nil {
		t.Fatalf("escape hatch must preserve upstream extends parsing: %v", err)
	}
	secondDisabled, err := disabled.createProjectProgram(true, root, tspath.ResolvePath(root, "tsconfig.json"))
	if err != nil {
		t.Fatalf("escape hatch second program build: %v", err)
	}
	sharedPath := tspath.ToPath(tspath.ResolvePath(root, "shared/tsconfig.json"), root, disabled.FS().UseCaseSensitiveFileNames())
	firstRef, _ := firstDisabled.GetResolvedProjectReferenceFor(sharedPath)
	secondRef, _ := secondDisabled.GetResolvedProjectReferenceFor(sharedPath)
	if firstRef == nil || secondRef == nil || firstRef == secondRef {
		t.Fatalf("escape hatch must disable project-reference caching: first=%p second=%p", firstRef, secondRef)
	}
}

func TestBuildContextScopesParallelFSToCompilerHosts(t *testing.T) {
	base := bundled.WrapFS(cachedvfs.From(osvfs.FS()))

	serial := newBuildContext(base)
	if hostFS := serial.newCompilerHostWithCache("/").FS(); hostFS != serial.FS() {
		t.Fatal("serial compiler host did not retain the context VFS")
	}

	parallel := newBuildContext(base)
	logicalFS := parallel.FS()
	parallel.enableConcurrentProgramQueries()
	firstHostFS := parallel.newCompilerHostWithCache("/").FS()
	parallel.enableConcurrentProgramQueries()
	secondHostFS := parallel.newCompilerHostWithCache("/").FS()
	if parallel.FS() != logicalFS {
		t.Fatal("parallel activation changed the context's non-Program VFS")
	}
	if firstHostFS != secondHostFS {
		t.Fatal("parallel compiler hosts did not share one stable VFS view")
	}
	programFS, ok := firstHostFS.(*parallelProgramFS)
	if !ok {
		t.Fatalf("parallel compiler host FS has type %T", firstHostFS)
	}
	if programFS.FS != logicalFS {
		t.Fatal("parallel compiler VFS does not delegate to the context VFS")
	}
}

func TestBuildContextPreservesBOMSourceThroughVFSWrappers(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	fileName := tspath.NormalizePath(filepath.Join(directory, "marked.ts"))
	if err := os.WriteFile(fileName, []byte("\uFEFFlet a = 1;\n"), 0o644); err != nil {
		t.Fatalf("write marked source: %v", err)
	}
	overlay := utils.NewOverlayVFS(osvfs.FS(), map[string]string{
		fileName: "let a = 1;\n",
	})
	metadata := newProgramMetadataFS(overlay)
	parallel := newParallelProgramFS(metadata)

	for name, fsys := range map[string]utils.BOMSource{
		"programMetadataFS": metadata,
		"parallelProgramFS": parallel,
	} {
		if has, err := fsys.ReadSourceBOM(fileName); has || err != nil {
			t.Errorf("%s returned (%t, %v) for unmarked overlay content", name, has, err)
		}
	}
}

func TestSessionRejectsNilFilesystemBeforeLoading(t *testing.T) {
	var typedNil *programMetadataFS
	tests := []struct {
		name string
		fs   vfs.FS
	}{
		{name: "nil"},
		{name: "typed nil", fs: typedNil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			session := NewSession(test.fs)
			if session.FS() != nil {
				t.Fatal("invalid session unexpectedly exposed a filesystem")
			}
			if _, err := session.BuildProjects(ProjectBuildRequest{Configs: nil, Scope: AllDeclared, SingleThreaded: true}); err == nil {
				t.Fatal("BuildProjects accepted an invalid session")
			}
			if _, err := session.LoadCLI(ProjectSet{}, target.Plan{}, "/repo", true); err == nil {
				t.Fatal("LoadCLI accepted an invalid session")
			}
			if _, err := session.LoadAPI(ProjectSet{}, target.Plan{}, "/repo", true); err == nil {
				t.Fatal("LoadAPI accepted an invalid session")
			}
		})
	}
}

func TestSessionRecordsEmptyInitialCompilerGeneration(t *testing.T) {
	session := &Session{context: &buildContext{parseCache: utils.NewParseCache()}}
	session.retainCompilerPrograms(nil)
	if !session.initialProgramsSet {
		t.Fatal("empty initial compiler generation was not recorded")
	}
}

type parallelScanGateFS struct {
	vfs.FS
	scanDir string

	mu            sync.Mutex
	active        int
	peak          int
	total         int
	firstEntered  chan struct{}
	firstRelease  chan struct{}
	waiterEntered chan struct{}
	waiterRelease chan struct{}
	waiterTarget  int
	waiterClosed  bool
}

func (f *parallelScanGateFS) GetAccessibleEntries(path string) vfs.Entries {
	if tspath.NormalizePath(path) == f.scanDir {
		f.mu.Lock()
		f.total++
		callNum := f.total
		f.active++
		if f.active > f.peak {
			f.peak = f.active
		}
		if callNum == 1 {
			close(f.firstEntered)
			f.mu.Unlock()
			<-f.firstRelease
		} else {
			if f.active >= f.waiterTarget && !f.waiterClosed {
				f.waiterClosed = true
				close(f.waiterEntered)
			}
			f.mu.Unlock()
			<-f.waiterRelease
		}
		entries := f.FS.GetAccessibleEntries(path)
		f.mu.Lock()
		f.active--
		f.mu.Unlock()
		return entries
	}
	return f.FS.GetAccessibleEntries(path)
}

func TestBuildContextMissingPresetWaitersRetryInParallel(t *testing.T) {
	root := osvfs.FS().Realpath(tspath.NormalizePath(t.TempDir()))
	writeProgramBuildFixture(t, root, map[string]string{
		"packages/shared/tsconfig.json": `{
			"extends":"./missing-base.json",
			"compilerOptions":{"composite":true,"noLib":true},
			"include":["src/**/*"]
		}`,
		"packages/shared/src/value.ts": "export const value: number = 1;\n",
	})

	sharedConfig := tspath.ResolvePath(root, "packages/shared/tsconfig.json")
	sharedSrcDir := tspath.ResolvePath(root, "packages/shared/src")
	const callerCount = 8
	gate := &parallelScanGateFS{
		FS:            bundled.WrapFS(osvfs.FS()),
		scanDir:       sharedSrcDir,
		firstEntered:  make(chan struct{}),
		firstRelease:  make(chan struct{}),
		waiterEntered: make(chan struct{}),
		waiterRelease: make(chan struct{}),
		// Under serialization, at most 1 goroutine can scan at a time (peak == 1).
		// Requiring >= 2 simultaneous retries proves waiters fan out in parallel
		// without needing production synchronization hooks.
		waiterTarget: 2,
	}
	ctx := newBuildContext(gate)
	sharedPath := tspath.ToPath(sharedConfig, root, ctx.FS().UseCaseSensitiveFileNames())

	var wg sync.WaitGroup
	results := make([]*tsoptions.ParsedCommandLine, callerCount)
	wg.Add(1)
	go func() {
		defer wg.Done()
		results[0] = ctx.newCompilerHost(root).GetResolvedProjectReference(sharedConfig, sharedPath)
	}()
	select {
	case <-gate.firstEntered:
	case <-time.After(5 * time.Second):
		t.Fatal("leader did not enter source scan")
	}

	waitersStarted := make(chan struct{}, callerCount-1)
	for i := 1; i < callerCount; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			host := ctx.newCompilerHost(root)
			waitersStarted <- struct{}{}
			results[idx] = host.GetResolvedProjectReference(sharedConfig, sharedPath)
		}(i)
	}
	for range callerCount - 1 {
		<-waitersStarted
	}
	time.Sleep(25 * time.Millisecond)
	close(gate.firstRelease)

	select {
	case <-gate.waiterEntered:
		close(gate.waiterRelease)
	case <-time.After(5 * time.Second):
		close(gate.waiterRelease)
		wg.Wait()
		t.Fatalf("waiters serialized after missing preset failure: peak simultaneous scans = %d, want >= %d", gate.peak, gate.waiterTarget)
	}
	wg.Wait()

	if gate.total != callerCount {
		t.Fatalf("total scans = %d, want %d", gate.total, callerCount)
	}
	if gate.peak < gate.waiterTarget {
		t.Fatalf("peak simultaneous scans = %d, want >= %d", gate.peak, gate.waiterTarget)
	}
	for i, res := range results {
		if res == nil || !hasTransientConfigLookupDiagnostic(res) {
			t.Fatalf("caller %d did not receive expected missing-extends diagnostic: %v", i, res)
		}
	}
}
