package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
)

func buildProgramFromRoots(t *testing.T, cwd string, roots []string, fs vfs.FS) *compiler.Program {
	t.Helper()
	host := CreateCompilerHost(cwd, fs)
	prog, err := CreateProgramFromOptions(false, &core.CompilerOptions{}, roots, host)
	if err != nil {
		t.Fatal(err)
	}
	return prog
}

func keysEqual(a, b map[string]struct{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if _, ok := b[k]; !ok {
			return false
		}
	}
	return true
}

// referenceSet is an INDEPENDENT golden computation of the program-file set: a
// straight serial loop matching the original (pre-parallelization) semantics.
// It deliberately does not share code with CollectProgramFiles' chunked
// resolution, so asserting against it catches bugs that would pollute both the
// serial and parallel paths equally (e.g. a dropped or mis-indexed realpath
// merge that serial==parallel comparison alone cannot detect).
func referenceSet(programs []*compiler.Program, fs vfs.FS) map[string]struct{} {
	want := map[string]struct{}{}
	for _, prog := range programs {
		for _, sf := range prog.GetSourceFiles() {
			name := sf.FileName()
			if _, ok := want[name]; ok {
				continue
			}
			want[name] = struct{}{}
			if r := fs.Realpath(name); r != name {
				want[r] = struct{}{}
			}
		}
	}
	return want
}

// assertMatchesReference asserts both the serial and parallel CollectProgramFiles
// equal the independent reference. referenceSet runs first to warm the shared
// realpath cache, so all three observe a consistent resolution.
func assertMatchesReference(t *testing.T, programs []*compiler.Program, fs vfs.FS) (ref, serial, parallel map[string]struct{}) {
	t.Helper()
	ref = referenceSet(programs, fs)
	serial = CollectProgramFiles(programs, fs, true)
	parallel = CollectProgramFiles(programs, fs, false)
	if !keysEqual(serial, ref) {
		t.Fatalf("serial result != reference (serial=%d ref=%d)", len(serial), len(ref))
	}
	if !keysEqual(parallel, ref) {
		t.Fatalf("parallel result != reference (parallel=%d ref=%d)", len(parallel), len(ref))
	}
	return ref, serial, parallel
}

// Parallel and serial resolution must both equal the independent reference.
func TestCollectProgramFiles_MatchesReference(t *testing.T) {
	dir := t.TempDir()
	var roots []string
	for i := range 30 {
		fp := filepath.Join(dir, fmt.Sprintf("f%d.ts", i))
		if err := os.WriteFile(fp, []byte("export const x = 1;\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		roots = append(roots, fp)
	}
	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	prog := buildProgramFromRoots(t, dir, roots, fs)

	ref, _, _ := assertMatchesReference(t, []*compiler.Program{prog}, fs)

	// Every program source file name must be present.
	for _, sf := range prog.GetSourceFiles() {
		if _, ok := ref[sf.FileName()]; !ok {
			t.Fatalf("source file %s missing from set", sf.FileName())
		}
	}
}

// With a symlinked directory, Realpath resolves to a different path, so the set
// must contain the resolved real paths in BOTH modes (ground-truth assertion,
// not just serial==parallel). This is the branch that a "drop realpath" bug
// would silently disable.
func TestCollectProgramFiles_SymlinkResolved(t *testing.T) {
	dir := t.TempDir()
	realDir := filepath.Join(dir, "real")
	if err := os.MkdirAll(realDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for i := range 5 {
		if err := os.WriteFile(filepath.Join(realDir, fmt.Sprintf("a%d.ts", i)), []byte("export const x = 1;\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	linkDir := filepath.Join(dir, "link")
	if err := os.Symlink(realDir, linkDir); err != nil {
		t.Skipf("symlink unsupported: %v", err)
	}
	var roots []string
	for i := range 5 {
		roots = append(roots, filepath.Join(linkDir, fmt.Sprintf("a%d.ts", i)))
	}
	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	prog := buildProgramFromRoots(t, dir, roots, fs)

	ref, serial, parallel := assertMatchesReference(t, []*compiler.Program{prog}, fs)

	// The symlink must produce realpath expansion: at least one set entry is a
	// resolved path that is not any program source file name. Otherwise the
	// realpath branch was not exercised at all.
	names := map[string]struct{}{}
	for _, sf := range prog.GetSourceFiles() {
		names[sf.FileName()] = struct{}{}
	}
	expanded := 0
	for k := range ref {
		if _, isName := names[k]; !isName {
			expanded++
		}
	}
	if expanded == 0 {
		t.Fatalf("no realpath expansion observed (set=%d, files=%d); branch uncovered", len(ref), len(names))
	}
	// Every resolved real path must be present in both modes.
	for name := range names {
		if r := fs.Realpath(name); r != name {
			if _, ok := serial[r]; !ok {
				t.Fatalf("resolved path %s missing from serial set", r)
			}
			if _, ok := parallel[r]; !ok {
				t.Fatalf("resolved path %s missing from parallel set", r)
			}
		}
	}
}

// Multiple programs sharing files: dedup is correct and both programs' files
// (and their resolved paths) are merged, matching the reference.
func TestCollectProgramFiles_MultiProgramDedup(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"a.ts", "b.ts", "shared.ts"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("export const x = 1;\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	progA := buildProgramFromRoots(t, dir, []string{filepath.Join(dir, "a.ts"), filepath.Join(dir, "shared.ts")}, fs)
	progB := buildProgramFromRoots(t, dir, []string{filepath.Join(dir, "b.ts"), filepath.Join(dir, "shared.ts")}, fs)

	assertMatchesReference(t, []*compiler.Program{progA, progB}, fs)
}

// Empty / single-file inputs must not panic and must match the reference.
func TestCollectProgramFiles_EmptyAndSingle(t *testing.T) {
	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	if got := CollectProgramFiles(nil, fs, false); len(got) != 0 {
		t.Fatalf("empty programs should yield empty set, got %d", len(got))
	}

	dir := t.TempDir()
	fp := filepath.Join(dir, "only.ts")
	if err := os.WriteFile(fp, []byte("export const x = 1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	prog := buildProgramFromRoots(t, dir, []string{fp}, fs)
	assertMatchesReference(t, []*compiler.Program{prog}, fs)
}
