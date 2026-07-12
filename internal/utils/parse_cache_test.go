package utils

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/zeebo/xxh3"

	"github.com/web-infra-dev/rslint/internal/api"
)

type readCountingFS struct {
	vfs.FS
	mu    sync.Mutex
	reads map[string]int
}

func newReadCountingFS(base vfs.FS) *readCountingFS {
	return &readCountingFS{FS: base, reads: make(map[string]int)}
}

func (f *readCountingFS) ReadFile(path string) (string, bool) {
	f.mu.Lock()
	f.reads[path]++
	f.mu.Unlock()
	return f.FS.ReadFile(path)
}

func (f *readCountingFS) readCount(path string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.reads[path]
}

type memoryReadFS struct {
	vfs.FS
	mu        sync.Mutex
	contents  map[string]string
	realPaths map[string]string
	reads     map[string]int
}

func newMemoryReadFS(contents map[string]string) *memoryReadFS {
	return &memoryReadFS{
		FS:        bundled.WrapFS(osvfs.FS()),
		contents:  contents,
		realPaths: make(map[string]string),
		reads:     make(map[string]int),
	}
}

func (f *memoryReadFS) ReadFile(path string) (string, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.reads[path]++
	text, ok := f.contents[path]
	return text, ok
}

func (f *memoryReadFS) Realpath(path string) string {
	f.mu.Lock()
	defer f.mu.Unlock()
	if realpath := f.realPaths[path]; realpath != "" {
		return realpath
	}
	return path
}

func (f *memoryReadFS) set(path string, text string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.contents[path] = text
}

func (f *memoryReadFS) readCount(path string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.reads[path]
}

// valueReadFS deliberately has a non-pointer dynamic type. The source
// snapshot layer must bypass it because value equality is not FS identity.
type valueReadFS struct {
	vfs.FS
	source *memoryReadFS
}

func (f valueReadFS) ReadFile(path string) (string, bool) {
	return f.source.ReadFile(path)
}

type divergentConcurrentFS struct {
	vfs.FS
	path       string
	wantReads  int
	mu         sync.Mutex
	reads      int
	allStarted chan struct{}
	release    chan struct{}
}

func (f *divergentConcurrentFS) ReadFile(path string) (string, bool) {
	if path != f.path {
		return "", false
	}
	f.mu.Lock()
	f.reads++
	readNumber := f.reads
	if f.reads == f.wantReads {
		close(f.allStarted)
	}
	f.mu.Unlock()
	<-f.release
	return fmt.Sprintf("export const value = %d;\n", readNumber), true
}

type blockingGenerationFS struct {
	vfs.FS
	path    string
	mu      sync.Mutex
	text    string
	reads   int
	started chan struct{}
	release chan struct{}
}

func (f *blockingGenerationFS) ReadFile(path string) (string, bool) {
	if path != f.path {
		return "", false
	}
	f.mu.Lock()
	f.reads++
	first := f.reads == 1
	text := f.text
	f.mu.Unlock()
	if first {
		close(f.started)
		<-f.release
	}
	return text, true
}

func (f *blockingGenerationFS) setText(text string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.text = text
}

func testParseOpts(fileName string, emi ast.ExternalModuleIndicatorOptions) ast.SourceFileParseOptions {
	return ast.SourceFileParseOptions{
		FileName:                       fileName,
		Path:                           tspath.ToPath(fileName, "", true),
		ExternalModuleIndicatorOptions: emi,
	}
}

// assertParseEquivalent asserts two SourceFiles are equivalent parses of the
// same text: full-tree byte comparison via the --api encoder (which also
// pins I3 — the encoded header hash bytes stay zero) plus field-by-field
// parse diagnostics.
func assertParseEquivalent(t *testing.T, a, b *ast.SourceFile) {
	t.Helper()
	encA, errA := api.EncodeAST(a, "t")
	encB, errB := api.EncodeAST(b, "t")
	if errA != nil || errB != nil {
		t.Fatalf("EncodeAST failed: %v / %v", errA, errB)
	}
	if !bytes.Equal(encA, encB) {
		t.Fatalf("re-parsed AST encoding differs (len %d vs %d)", len(encA), len(encB))
	}
	da, db := a.Diagnostics(), b.Diagnostics()
	if len(da) != len(db) {
		t.Fatalf("diagnostics count differs: %d vs %d", len(da), len(db))
	}
	for i := range da {
		if da[i].Code() != db[i].Code() || da[i].Pos() != db[i].Pos() ||
			da[i].End() != db[i].End() || da[i].MessageKey() != db[i].MessageKey() {
			t.Fatalf("diagnostic %d differs: %v vs %v", i, da[i], db[i])
		}
	}
}

func TestParseCache_SameKeySameObject(t *testing.T) {
	c := &ParseCache{}
	opts := testParseOpts("/virtual/a.ts", ast.ExternalModuleIndicatorOptions{})
	text := "export const x: number = 1;\n"

	sf1 := c.acquire(opts, text)
	sf2 := c.acquire(opts, text)
	if sf1 != sf2 {
		t.Fatal("same (opts, text) must return the same shared object")
	}
}

func TestParseCache_DifferentContentNotShared(t *testing.T) {
	c := &ParseCache{}
	opts := testParseOpts("/virtual/a.ts", ast.ExternalModuleIndicatorOptions{})

	sf1 := c.acquire(opts, "export const x = 1;\n")
	sf2 := c.acquire(opts, "export const x = 2;\n")
	if sf1 == sf2 {
		t.Fatal("different content must not share an object")
	}
}

func TestParseCache_DifferentEMIOptionsNotShared(t *testing.T) {
	c := &ParseCache{}
	text := "const x = 1;\n"
	sf1 := c.acquire(testParseOpts("/virtual/a.ts", ast.ExternalModuleIndicatorOptions{}), text)
	sf2 := c.acquire(testParseOpts("/virtual/a.ts", ast.ExternalModuleIndicatorOptions{Force: true}), text)
	if sf1 == sf2 {
		t.Fatal("different ExternalModuleIndicatorOptions must not share an object")
	}
}

func TestCachingHost_SourceSnapshotSharedAcrossParseOptions(t *testing.T) {
	fileName := "/virtual/options.ts"
	fs := newMemoryReadFS(map[string]string{fileName: "const value = 1;\n"})
	cache := &ParseCache{}
	host := WithParseCache(CreateCompilerHost("/virtual", fs), cache)

	regular := host.GetSourceFile(testParseOpts(fileName, ast.ExternalModuleIndicatorOptions{}))
	forced := host.GetSourceFile(testParseOpts(fileName, ast.ExternalModuleIndicatorOptions{Force: true}))
	if regular == nil || forced == nil {
		t.Fatal("both parse-option variants must produce a SourceFile")
	}
	if regular == forced {
		t.Fatal("different parse options must retain distinct AST entries")
	}
	if got := fs.readCount(fileName); got != 1 {
		t.Fatalf("source text must be shared across parse options, reads = %d, want 1", got)
	}
}

func TestCachingHost_SourceSnapshotReadOnceAcrossPrograms(t *testing.T) {
	tmpDir := t.TempDir()
	fileName := tspath.NormalizePath(filepath.Join(tmpDir, "shared.ts"))
	if err := os.WriteFile(fileName, []byte("export const shared = 1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	fs := newReadCountingFS(bundled.WrapFS(cachedvfs.From(osvfs.FS())))
	cache := &ParseCache{}
	build := func() *compiler.Program {
		t.Helper()
		host := WithParseCache(CreateCompilerHost(tmpDir, fs), cache)
		program, err := CreateProgramFromOptionsLenient(true, &core.CompilerOptions{
			NoLib:     core.TSTrue,
			NoResolve: core.TSTrue,
		}, []string{fileName}, host)
		if err != nil {
			t.Fatal(err)
		}
		return program
	}

	firstProgram := build()
	secondProgram := build()
	first := firstProgram.GetSourceFile(fileName)
	second := secondProgram.GetSourceFile(fileName)
	if first == nil || second == nil {
		t.Fatal("shared source must be present in both Programs")
	}
	if first != second {
		t.Fatal("same source and parse inputs must share the AST across Programs")
	}
	if got := fs.readCount(fileName); got != 1 {
		t.Fatalf("same source must be read once per generation, reads = %d", got)
	}

	cache.InvalidateSourceSnapshots()
	third := build().GetSourceFile(fileName)
	if got := fs.readCount(fileName); got != 2 {
		t.Fatalf("new generation must read source again, reads = %d", got)
	}
	if third != first {
		t.Fatal("unchanged content after invalidation must reuse the content-keyed AST")
	}
}

func TestCachingHost_ChangedContentRequiresInvalidation(t *testing.T) {
	fileName := "/virtual/change.ts"
	oldText := "export const value = 1;\n"
	newText := "export const value = 2;\n"
	fs := newMemoryReadFS(map[string]string{fileName: oldText})
	cache := &ParseCache{}
	host := WithParseCache(CreateCompilerHost("/virtual", fs), cache)
	opts := testParseOpts(fileName, ast.ExternalModuleIndicatorOptions{})

	oldSource := host.GetSourceFile(opts)
	fs.set(fileName, newText)
	frozenSource := host.GetSourceFile(opts)
	if frozenSource != oldSource || frozenSource.Text() != oldText {
		t.Fatal("a generation must retain its first successful source snapshot")
	}
	if got := fs.readCount(fileName); got != 1 {
		t.Fatalf("generation hit must not re-read source, reads = %d", got)
	}

	cache.InvalidateSourceSnapshots()
	newSource := host.GetSourceFile(opts)
	if newSource == nil || newSource.Text() != newText {
		t.Fatalf("new generation did not observe changed source: %v", newSource)
	}
	if newSource == oldSource {
		t.Fatal("changed content must produce a new AST")
	}
	if got := fs.readCount(fileName); got != 2 {
		t.Fatalf("new generation must perform one new read, reads = %d", got)
	}
	value, ok := cache.currentSourceGeneration().entries.Load(fileName)
	if !ok {
		t.Fatal("changed source snapshot was not published")
	}
	snapshot := value.(sourceSnapshot)
	if snapshot.text != newText || snapshot.hash != xxh3.HashString128(newText) {
		t.Fatalf("snapshot text/hash pair is inconsistent: %+v", snapshot)
	}
}

func TestCachingHost_EmptySourceIsCached(t *testing.T) {
	fileName := "/virtual/empty.ts"
	fs := newMemoryReadFS(map[string]string{fileName: ""})
	cache := &ParseCache{}
	host := WithParseCache(CreateCompilerHost("/virtual", fs), cache)
	opts := testParseOpts(fileName, ast.ExternalModuleIndicatorOptions{})

	first := host.GetSourceFile(opts)
	second := host.GetSourceFile(opts)
	if first == nil || second != first || first.Text() != "" {
		t.Fatal("an empty successful read must be cached as a valid source")
	}
	if got := fs.readCount(fileName); got != 1 {
		t.Fatalf("empty source reads = %d, want 1", got)
	}
}

func TestCachingHost_ExactAliasesAreNotMerged(t *testing.T) {
	aliasPath := "/virtual/alias.ts"
	canonicalPath := "/virtual/canonical.ts"
	fs := newMemoryReadFS(map[string]string{
		aliasPath:     "export const fromAlias = true;\n",
		canonicalPath: "export const fromCanonical = true;\n",
	})
	fs.realPaths[aliasPath] = canonicalPath
	fs.realPaths[canonicalPath] = canonicalPath
	cache := &ParseCache{}
	host := WithParseCache(CreateCompilerHost("/virtual", fs), cache)

	alias := host.GetSourceFile(testParseOpts(aliasPath, ast.ExternalModuleIndicatorOptions{}))
	canonical := host.GetSourceFile(testParseOpts(canonicalPath, ast.ExternalModuleIndicatorOptions{}))
	if alias == nil || canonical == nil {
		t.Fatal("both aliases must produce a SourceFile")
	}
	if alias.Text() == canonical.Text() || alias == canonical {
		t.Fatal("paths with the same realpath but distinct overlay text must not be merged")
	}
	if fs.readCount(aliasPath) != 1 || fs.readCount(canonicalPath) != 1 {
		t.Fatalf("each exact alias must own one read, alias=%d canonical=%d", fs.readCount(aliasPath), fs.readCount(canonicalPath))
	}
}

func TestCachingHost_OverlaySymlinkAliasesWithSharedPathAreNotMerged(t *testing.T) {
	tmpDir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("resolve temp dir: %v", err)
	}
	canonicalPath := tspath.NormalizePath(filepath.Join(tmpDir, "canonical.ts"))
	aliasPath := tspath.NormalizePath(filepath.Join(tmpDir, "alias.ts"))
	if err := os.WriteFile(canonicalPath, []byte("export const fromDisk = true;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(canonicalPath, aliasPath); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	fs := newReadCountingFS(NewOverlayVFS(
		bundled.WrapFS(osvfs.FS()),
		map[string]string{
			canonicalPath: "export const fromCanonicalOverlay = true;\n",
			aliasPath:     "export const fromAliasOverlay = true;\n",
		},
	))
	if aliasRealpath, canonicalRealpath := fs.Realpath(aliasPath), fs.Realpath(canonicalPath); aliasRealpath != canonicalRealpath {
		t.Fatalf("fixture aliases do not share a realpath: %q != %q", aliasRealpath, canonicalRealpath)
	}

	cache := &ParseCache{}
	host := WithParseCache(CreateCompilerHost(tmpDir, fs), cache)
	canonicalOpts := testParseOpts(canonicalPath, ast.ExternalModuleIndicatorOptions{})
	aliasOpts := testParseOpts(aliasPath, ast.ExternalModuleIndicatorOptions{})
	// Exercise the strongest alias collision: compiler identity has already
	// canonicalized Path, while FileName still names the distinct overlays.
	aliasOpts.Path = canonicalOpts.Path

	canonical := host.GetSourceFile(canonicalOpts)
	alias := host.GetSourceFile(aliasOpts)
	if canonical == nil || alias == nil {
		t.Fatal("both overlay aliases must produce a SourceFile")
	}
	if canonical == alias || canonical.Text() == alias.Text() {
		t.Fatal("exact FileName overlays must remain distinct even when Path and realpath match")
	}
	if got := fs.readCount(canonicalPath); got != 1 {
		t.Fatalf("canonical overlay reads = %d, want 1", got)
	}
	if got := fs.readCount(aliasPath); got != 1 {
		t.Fatalf("alias overlay reads = %d, want 1", got)
	}
}

func TestCachingHost_DifferentFilesystemViewsBypassSourceSnapshots(t *testing.T) {
	fileName := "/virtual/view.ts"
	fsA := newMemoryReadFS(map[string]string{fileName: "export const view = 'a';\n"})
	fsB := newMemoryReadFS(map[string]string{fileName: "export const view = 'b';\n"})
	cache := &ParseCache{}
	hostA := WithParseCache(CreateCompilerHost("/virtual", fsA), cache)
	hostB := WithParseCache(CreateCompilerHost("/virtual", fsB), cache)
	opts := testParseOpts(fileName, ast.ExternalModuleIndicatorOptions{})

	a1, a2 := hostA.GetSourceFile(opts), hostA.GetSourceFile(opts)
	b1, b2 := hostB.GetSourceFile(opts), hostB.GetSourceFile(opts)
	if a1 == nil || b1 == nil || a1.Text() == b1.Text() {
		t.Fatal("different FS views of one path must retain their own content")
	}
	if a2 != a1 || b2 != b1 {
		t.Fatal("both FS views must still reuse their content-keyed AST entries")
	}
	if got := fsA.readCount(fileName); got != 1 {
		t.Fatalf("bound FS reads = %d, want 1", got)
	}
	if got := fsB.readCount(fileName); got != 2 {
		t.Fatalf("different FS must bypass source snapshots, reads = %d, want 2", got)
	}

	valueBacking := newMemoryReadFS(map[string]string{fileName: "export const view = 'value';\n"})
	valueFS := valueReadFS{FS: valueBacking.FS, source: valueBacking}
	valueCache := &ParseCache{}
	valueHost := WithParseCache(CreateCompilerHost("/virtual", valueFS), valueCache)
	valueFirst, valueSecond := valueHost.GetSourceFile(opts), valueHost.GetSourceFile(opts)
	if valueFirst == nil || valueSecond != valueFirst {
		t.Fatal("non-pointer FS bypass must retain AST-cache correctness")
	}
	if got := valueBacking.readCount(fileName); got != 2 {
		t.Fatalf("non-pointer FS must bypass source snapshots, reads = %d, want 2", got)
	}
}

func TestParseCache_ConcurrentFilesystemBindingKeepsViewsIsolated(t *testing.T) {
	fileName := "/virtual/concurrent-view.ts"
	textA := "export const view = 'a';\n"
	textB := "export const view = 'b';\n"
	fsA := newMemoryReadFS(map[string]string{fileName: textA})
	fsB := newMemoryReadFS(map[string]string{fileName: textB})
	cache := &ParseCache{}
	opts := testParseOpts(fileName, ast.ExternalModuleIndicatorOptions{})
	start := make(chan struct{})
	errs := make(chan error, 16)

	var wg sync.WaitGroup
	for i := range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			fs, want := vfs.FS(fsA), textA
			if i%2 == 1 {
				fs, want = fsB, textB
			}
			host := WithParseCache(CreateCompilerHost("/virtual", fs), cache)
			source := host.GetSourceFile(opts)
			if source == nil || source.Text() != want {
				errs <- fmt.Errorf("filesystem view %d returned %v, want %q", i%2, source, want)
			}
		}()
	}
	close(start)
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
}

// I3: the cache never writes file.Hash — the --api EncodeAST header relies
// on it staying zero.
func TestParseCache_HashStaysZero(t *testing.T) {
	c := &ParseCache{}
	sf := c.acquire(testParseOpts("/virtual/a.ts", ast.ExternalModuleIndicatorOptions{}), "let a = 1;\n")
	if sf.Hash != (xxh3.Uint128{}) {
		t.Fatalf("sf.Hash must stay zero, got %v", sf.Hash)
	}
	c.RetainOnly(nil)
	sf2 := c.acquire(testParseOpts("/virtual/a.ts", ast.ExternalModuleIndicatorOptions{}), "let a = 1;\n")
	if sf2.Hash != (xxh3.Uint128{}) {
		t.Fatalf("sf.Hash must stay zero after eviction/reparse, got %v", sf2.Hash)
	}
}

// T2/T3: a RetainOnly(empty) sweep clears everything; a re-acquired entry is
// a NEW object whose parse result is byte-equivalent to the evicted one.
func TestParseCache_EvictedReparseEquivalent(t *testing.T) {
	c := &ParseCache{}
	opts := testParseOpts("/virtual/a.ts", ast.ExternalModuleIndicatorOptions{})
	// Include a deliberate syntax error so parse diagnostics are non-empty
	// and the diagnostics part of the equivalence assertion has teeth.
	text := "export const x: = 1;\nfunction f() { return 2 }\n"

	before := c.acquire(opts, text)
	c.RetainOnly(nil) // empty live set: evict all
	after := c.acquire(opts, text)
	if before == after {
		t.Fatal("entry must have been evicted; expected a fresh object")
	}
	assertParseEquivalent(t, before, after)
}

// T1/T4: RetainOnly keeps entries whose object is referenced by any given
// program (pointer-stable across the sweep) and evicts the rest.
func TestParseCache_RetainOnlyKeepsLivePrograms(t *testing.T) {
	tmpDir := t.TempDir()
	for name, content := range map[string]string{
		"a.ts":      "export const a = 1;\n",
		"b.ts":      "export const b = 2;\n",
		"shared.ts": "export const s = 3;\n",
	} {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	cache := &ParseCache{}
	host := WithParseCache(CreateCompilerHost(tmpDir, fs), cache)

	progA, err := CreateProgramFromOptions(false, &core.CompilerOptions{},
		[]string{filepath.Join(tmpDir, "a.ts"), filepath.Join(tmpDir, "shared.ts")}, host)
	if err != nil {
		t.Fatal(err)
	}
	progB, err := CreateProgramFromOptions(false, &core.CompilerOptions{},
		[]string{filepath.Join(tmpDir, "b.ts"), filepath.Join(tmpDir, "shared.ts")}, host)
	if err != nil {
		t.Fatal(err)
	}

	findFile := func(p *compiler.Program, name string) *ast.SourceFile {
		t.Helper()
		want := tspath.NormalizePath(filepath.Join(tmpDir, name))
		for _, sf := range p.GetSourceFiles() {
			if tspath.NormalizePath(sf.FileName()) == want {
				return sf
			}
		}
		t.Fatalf("%s not found in program", name)
		return nil
	}

	// The shared file parsed once: both programs hold the same object.
	sharedA, sharedB := findFile(progA, "shared.ts"), findFile(progB, "shared.ts")
	if sharedA != sharedB {
		t.Fatal("shared.ts must be one shared object across programs")
	}

	// Sweep with both programs live: every program file keeps its pointer.
	cache.RetainOnly([]*compiler.Program{progA, progB})
	bText, _ := fs.ReadFile(filepath.Join(tmpDir, "b.ts"))
	bOpts := findFile(progB, "b.ts").ParseOptions()
	if got := cache.acquire(bOpts, bText); got != findFile(progB, "b.ts") {
		t.Fatal("live entry must keep its pointer across RetainOnly")
	}

	// T4: sweep with only progA live (progB "dropped from the live set"):
	// b.ts is evicted, a re-acquire yields a fresh but equivalent object;
	// shared.ts survives because progA still references it.
	oldB := findFile(progB, "b.ts")
	cache.RetainOnly([]*compiler.Program{progA})
	newB := cache.acquire(bOpts, bText)
	if newB == oldB {
		t.Fatal("entry only referenced by the dropped program must be evicted")
	}
	assertParseEquivalent(t, oldB, newB)
	sharedText, _ := fs.ReadFile(filepath.Join(tmpDir, "shared.ts"))
	if got := cache.acquire(sharedA.ParseOptions(), sharedText); got != sharedA {
		t.Fatal("entry referenced by a live program must survive the sweep")
	}
}

// T6: concurrent acquire against a concurrently sweeping RetainOnly must be
// race-free (run with -race) and every result must be a valid parse.
func TestParseCache_ConcurrentAcquireAndRetainOnly(t *testing.T) {
	c := &ParseCache{}
	opts := testParseOpts("/virtual/conc.ts", ast.ExternalModuleIndicatorOptions{})
	text := "export function f(n: number): number { return n * 2 }\n"
	ref := c.acquire(opts, text)

	stop := make(chan struct{})
	var sweeper sync.WaitGroup
	sweeper.Add(1)
	go func() {
		defer sweeper.Done()
		for {
			select {
			case <-stop:
				return
			default:
				c.RetainOnly(nil) // continuously evict everything
			}
		}
	}()

	var acquirers sync.WaitGroup
	for range 8 {
		acquirers.Add(1)
		go func() {
			defer acquirers.Done()
			for range 200 {
				if c.acquire(opts, text) == nil {
					t.Error("acquire returned nil")
					return
				}
			}
		}()
	}
	acquirers.Wait()
	close(stop)
	sweeper.Wait()

	final := c.acquire(opts, text)
	assertParseEquivalent(t, ref, final)
}

func TestCachingHost_ConcurrentMissesUseOneWinningSnapshot(t *testing.T) {
	const callers = 8
	fileName := "/virtual/concurrent-winner.ts"
	fs := &divergentConcurrentFS{
		FS:         bundled.WrapFS(osvfs.FS()),
		path:       fileName,
		wantReads:  callers,
		allStarted: make(chan struct{}),
		release:    make(chan struct{}),
	}
	cache := &ParseCache{}
	host := WithParseCache(CreateCompilerHost("/virtual", fs), cache)
	opts := testParseOpts(fileName, ast.ExternalModuleIndicatorOptions{})

	results := make([]*ast.SourceFile, callers)
	var wg sync.WaitGroup
	for i := range callers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i] = host.GetSourceFile(opts)
		}()
	}
	<-fs.allStarted
	close(fs.release)
	wg.Wait()

	winner := results[0]
	if winner == nil {
		t.Fatal("concurrent source lookup returned nil")
	}
	for i, result := range results[1:] {
		if result != winner {
			t.Fatalf("caller %d did not use the source/AST winner", i+1)
		}
	}
	value, ok := cache.currentSourceGeneration().entries.Load(fileName)
	if !ok {
		t.Fatal("winning source snapshot was not stored")
	}
	snapshot := value.(sourceSnapshot)
	if winner.Text() != snapshot.text || snapshot.hash != xxh3.HashString128(snapshot.text) {
		t.Fatal("winner AST and immutable source text/hash pair diverged")
	}
}

func TestCachingHost_LateOldGenerationMissCannotRepopulateNewGeneration(t *testing.T) {
	fileName := "/virtual/generation.ts"
	oldText := "export const generation = 'old';\n"
	newText := "export const generation = 'new';\n"
	fs := &blockingGenerationFS{
		FS:      bundled.WrapFS(osvfs.FS()),
		path:    fileName,
		text:    oldText,
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
	cache := &ParseCache{}
	host := WithParseCache(CreateCompilerHost("/virtual", fs), cache)
	opts := testParseOpts(fileName, ast.ExternalModuleIndicatorOptions{})

	oldResult := make(chan *ast.SourceFile, 1)
	go func() {
		oldResult <- host.GetSourceFile(opts)
	}()
	<-fs.started

	fs.setText(newText)
	cache.InvalidateSourceSnapshots()
	newSource := host.GetSourceFile(opts)
	if newSource == nil || newSource.Text() != newText {
		t.Fatalf("new generation did not observe new content: %v", newSource)
	}

	close(fs.release)
	oldSource := <-oldResult
	if oldSource == nil || oldSource.Text() != oldText {
		t.Fatalf("in-flight old generation did not finish with its captured text: %v", oldSource)
	}
	final := host.GetSourceFile(opts)
	if final != newSource || final.Text() != newText {
		t.Fatal("late old-generation store contaminated the current generation")
	}
}

func TestCachingHost_RetainOnlyDoesNotClearSourceSnapshot(t *testing.T) {
	fileName := "/virtual/retain-source.ts"
	text := "export const retained = true;\n"
	fs := newMemoryReadFS(map[string]string{fileName: text})
	cache := &ParseCache{}
	host := WithParseCache(CreateCompilerHost("/virtual", fs), cache)
	opts := testParseOpts(fileName, ast.ExternalModuleIndicatorOptions{})

	first := host.GetSourceFile(opts)
	cache.RetainOnly(nil)
	second := host.GetSourceFile(opts)
	if first == second {
		t.Fatal("RetainOnly(nil) must evict the AST entry")
	}
	if got := fs.readCount(fileName); got != 1 {
		t.Fatalf("RetainOnly must not clear source snapshots, reads = %d", got)
	}

	cache.InvalidateSourceSnapshots()
	third := host.GetSourceFile(opts)
	if got := fs.readCount(fileName); got != 2 {
		t.Fatalf("source invalidation must trigger a new read, reads = %d", got)
	}
	if third != second {
		t.Fatal("source invalidation must retain the unchanged content-keyed AST")
	}
}

func TestCachingHost_ConcurrentLookupAndInvalidation(t *testing.T) {
	fileName := "/virtual/invalidate-race.ts"
	text := "export const stable = true;\n"
	fs := newMemoryReadFS(map[string]string{fileName: text})
	cache := &ParseCache{}
	host := WithParseCache(CreateCompilerHost("/virtual", fs), cache)
	opts := testParseOpts(fileName, ast.ExternalModuleIndicatorOptions{})

	start := make(chan struct{})
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for range 200 {
				source := host.GetSourceFile(opts)
				if source == nil || source.Text() != text {
					t.Errorf("lookup returned an invalid source: %v", source)
					return
				}
			}
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-start
		for range 200 {
			cache.InvalidateSourceSnapshots()
		}
	}()
	close(start)
	wg.Wait()
}

// T9: a failed read returns nil and is not cached; once the file appears the
// host parses it normally.
func TestCachingHost_ReadFileFailureNotCached(t *testing.T) {
	tmpDir := t.TempDir()
	fs := newReadCountingFS(bundled.WrapFS(osvfs.FS())) // no cachedvfs: exists-cache would hide the file's later appearance
	cache := &ParseCache{}
	host := WithParseCache(CreateCompilerHost(tmpDir, fs), cache)

	target := filepath.Join(tmpDir, "late.ts")
	opts := testParseOpts(tspath.NormalizePath(target), ast.ExternalModuleIndicatorOptions{})
	if sf := host.GetSourceFile(opts); sf != nil {
		t.Fatal("missing file must yield nil")
	}
	if err := os.WriteFile(target, []byte("export const late = true;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	sf := host.GetSourceFile(opts)
	if sf == nil {
		t.Fatal("file appeared but GetSourceFile still returns nil (negative caching?)")
	}
	// Compiler hosts and vfs use TypeScript-normalized paths on every platform.
	if got := fs.readCount(opts.FileName); got != 2 {
		t.Fatalf("failed reads must be retried, reads = %d, want 2", got)
	}
}

func TestCachingHost_BypassReadFailureNotCached(t *testing.T) {
	fileName := "/virtual/bypass-late.ts"
	cache := &ParseCache{}
	boundFS := newMemoryReadFS(map[string]string{
		fileName: "export const bound = true;\n",
	})
	// Binding happens when the wrapper is created; no read is needed.
	_ = WithParseCache(CreateCompilerHost("/virtual", boundFS), cache)

	bypassFS := newMemoryReadFS(map[string]string{})
	host := WithParseCache(CreateCompilerHost("/virtual", bypassFS), cache)
	opts := testParseOpts(fileName, ast.ExternalModuleIndicatorOptions{})
	if source := host.GetSourceFile(opts); source != nil {
		t.Fatal("missing file on bypass FS must yield nil")
	}
	bypassFS.set(fileName, "export const appeared = true;\n")
	source := host.GetSourceFile(opts)
	if source == nil || source.Text() != "export const appeared = true;\n" {
		t.Fatalf("bypass FS did not retry the failed read: %v", source)
	}
	if got := bypassFS.readCount(fileName); got != 2 {
		t.Fatalf("failed bypass reads must be retried, reads = %d, want 2", got)
	}
}

func TestCachingHost_InvalidSourceEntryFallsBackSafely(t *testing.T) {
	fileName := "/virtual/invalid-entry.ts"
	text := "export const valid = true;\n"
	fs := newMemoryReadFS(map[string]string{fileName: text})
	cache := &ParseCache{}
	cache.currentSourceGeneration().entries.Store(fileName, "invalid internal value")
	host := WithParseCache(CreateCompilerHost("/virtual", fs), cache)

	source := host.GetSourceFile(testParseOpts(fileName, ast.ExternalModuleIndicatorOptions{}))
	if source == nil || source.Text() != text {
		t.Fatalf("invalid internal entry did not fall back to the fresh read: %v", source)
	}
	if got := fs.readCount(fileName); got != 1 {
		t.Fatalf("fallback reads = %d, want 1", got)
	}
}

func TestParseCache_NilInvalidationIsSafe(t *testing.T) {
	var cache *ParseCache
	cache.InvalidateSourceSnapshots()
}

func TestParseCache_NilRetainOnlyIsSafe(t *testing.T) {
	var cache *ParseCache
	cache.RetainOnly(nil)
	new(ParseCache).RetainOnly([]*compiler.Program{nil})
}

func TestNewParseCache_EnabledInitializesSourceGeneration(t *testing.T) {
	t.Setenv("RSLINT_DISABLE_PARSE_CACHE", "")
	cache := NewParseCache()
	if cache == nil {
		t.Fatal("enabled parse cache must be constructed")
	}
	initial := cache.sourceGeneration.Load()
	if initial == nil {
		t.Fatal("constructor must eagerly publish a source generation")
	}
	cache.InvalidateSourceSnapshots()
	if current := cache.sourceGeneration.Load(); current == nil || current == initial {
		t.Fatal("invalidation must publish a distinct non-nil generation")
	}
}

func TestNewParseCache_DisableEnvironmentBypassesBothLayers(t *testing.T) {
	t.Setenv("RSLINT_DISABLE_PARSE_CACHE", "1")
	cache := NewParseCache()
	if cache != nil {
		t.Fatal("RSLINT_DISABLE_PARSE_CACHE must return a nil cache")
	}

	fileName := "/virtual/disabled.ts"
	fs := newMemoryReadFS(map[string]string{fileName: "export const disabled = true;\n"})
	baseHost := CreateCompilerHost("/virtual", fs)
	host := WithParseCache(baseHost, cache)
	if host != baseHost {
		t.Fatal("disabled cache must leave the compiler host unwrapped")
	}
	opts := testParseOpts(fileName, ast.ExternalModuleIndicatorOptions{})
	first, second := host.GetSourceFile(opts), host.GetSourceFile(opts)
	if first == nil || second == nil || first == second {
		t.Fatal("disabled cache must perform independent source reads/parses")
	}
	if got := fs.readCount(fileName); got != 2 {
		t.Fatalf("disabled cache reads = %d, want 2", got)
	}
	cache.InvalidateSourceSnapshots()
}

func TestWithParseCache_NilPassthrough(t *testing.T) {
	fs := bundled.WrapFS(osvfs.FS())
	host := CreateCompilerHost(t.TempDir(), fs)
	if WithParseCache(host, nil) != host {
		t.Fatal("nil cache must return the host unchanged")
	}
}
