package utils

import (
	"bytes"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/zeebo/xxh3"

	"github.com/web-infra-dev/rslint/internal/api"
)

func testParseOpts(fileName string, emi ast.ExternalModuleIndicatorOptions) ast.SourceFileParseOptions {
	return ast.SourceFileParseOptions{
		FileName:                       fileName,
		Path:                           tspath.ToPath(fileName, "", true),
		ExternalModuleIndicatorOptions: emi,
	}
}

// assertParseEquivalent asserts two SourceFiles are equivalent parses of the
// same text: full-tree byte comparison via the --api encoder (which also
// pins I2 — the encoded header hash bytes stay zero) plus field-by-field
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

// I2: the cache never writes file.Hash — the --api EncodeAST header relies
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

// T9: a failed read returns nil and is not cached; once the file appears the
// host parses it normally.
func TestCachingHost_ReadFileFailureNotCached(t *testing.T) {
	tmpDir := t.TempDir()
	fs := bundled.WrapFS(osvfs.FS()) // no cachedvfs: exists-cache would hide the file's later appearance
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
}

func TestWithParseCache_NilPassthrough(t *testing.T) {
	fs := bundled.WrapFS(osvfs.FS())
	host := CreateCompilerHost(t.TempDir(), fs)
	if WithParseCache(host, nil) != host {
		t.Fatal("nil cache must return the host unchanged")
	}
}
