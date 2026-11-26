package utils

// cspell:ignore synctest

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"testing/iotest"
	"testing/synctest"
	"time"
	"unicode/utf16"

	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
)

// writeTestFile writes bytes to a file under dir and returns the path in the
// normalized form the VFS layers key on.
func writeTestFile(t *testing.T, dir string, name string, contents []byte) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, contents, 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return tspath.NormalizePath(path)
}

// TestSourceHasBOMFromDisk covers the answer for text that came off disk. The
// file reader decodes the mark away before anyone sees the text, so the file's
// own bytes are the only witness left — including for UTF-16, which is decoded
// to UTF-8 and loses its mark the same way.
func TestSourceHasBOMFromDisk(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	utf8Marked := writeTestFile(t, dir, "utf8-bom.ts", []byte(BOM+"let a = 1;\n"))
	clean := writeTestFile(t, dir, "clean.ts", []byte("let a = 1;\n"))
	empty := writeTestFile(t, dir, "empty.ts", nil)
	short := writeTestFile(t, dir, "short.ts", []byte{0xFF})
	bomOnly := writeTestFile(t, dir, "bom-only.ts", []byte{0xFE, 0xFF})
	utf16LE := writeTestFile(t, dir, "utf16-le.ts", []byte{0xFF, 0xFE, 'a', 0x00})
	utf16BE := writeTestFile(t, dir, "utf16-be.ts", []byte{0xFE, 0xFF, 0x00, 'a'})
	missing := tspath.NormalizePath(filepath.Join(dir, "missing.ts"))

	fs := osvfs.FS()
	for _, tt := range []struct {
		name string
		path string
		want bool
	}{
		{"utf-8 mark", utf8Marked, true},
		{"no mark", clean, false},
		{"empty file", empty, false},
		{"one byte", short, false},
		{"two-byte UTF-16 mark", bomOnly, true},
		{"utf-16 little endian mark", utf16LE, true},
		{"utf-16 big endian mark", utf16BE, true},
		{"unreadable path", missing, false},
	} {
		if got, err := ReadSourceBOM(fs, tt.path); got != tt.want || err != nil {
			t.Errorf("ReadSourceBOM(%s) = (%v, %v), want (%v, nil)", tt.name, got, err, tt.want)
		}
	}
}

// TestSourceHasBOMFromOverlay locks in that caller-supplied content answers for
// itself. Consulting the file behind an overlay would report a mark on a buffer
// that does not have one, and miss a mark on one that does.
func TestSourceHasBOMFromOverlay(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	markedOnDisk := writeTestFile(t, dir, "marked.ts", []byte(BOM+"let a = 1;\n"))
	cleanOnDisk := writeTestFile(t, dir, "clean.ts", []byte("let a = 1;\n"))

	fs := NewOverlayVFS(osvfs.FS(), map[string]string{
		markedOnDisk: "let a = 1;\n",
		cleanOnDisk:  BOM + "let a = 1;\n",
	})

	if SourceHasBOM(fs, markedOnDisk) {
		t.Error("a clean overlay over a marked file reports a mark")
	}
	if !SourceHasBOM(fs, cleanOnDisk) {
		t.Error("a marked overlay over a clean file reports no mark")
	}

	// The text handed to the parser never carries the mark, whichever side it
	// came from, so offsets measured against it agree across both shapes.
	if text, _ := fs.ReadFile(cleanOnDisk); text != "let a = 1;\n" {
		t.Errorf("overlay ReadFile = %q, want the text without its mark", text)
	}
	if size := fs.Stat(cleanOnDisk).Size(); size != int64(len("let a = 1;\n")) {
		t.Errorf("overlay Stat size = %d, want the size without the mark", size)
	}

	t.Run("Windows realpath alias", func(t *testing.T) {
		t.Parallel()
		const alias = "C:/Link/source.ts"
		const canonical = "c:/Physical/source.ts"
		base := newMemoryReadFS(nil)
		base.realPaths[alias] = "C:/Physical/source.ts"
		files := map[string]string{canonical: BOM + "let value = 1;"}
		overlay := NewOverlayVFS(base, files)
		if !SourceHasBOM(overlay, alias) {
			t.Fatal("realpath drive spelling lost the overlay BOM")
		}
		if text, ok := overlay.ReadFile(alias); !ok || text != "let value = 1;" {
			t.Fatalf("realpath drive spelling lost the overlay text: %q, %v", text, ok)
		}
		files[canonical] = "let value = 20;"
		info := overlay.Stat(alias)
		if SourceHasBOM(overlay, alias) || info == nil || info.Size() != int64(len(files[canonical])) {
			t.Fatal("BOM and Stat did not observe the updated caller map")
		}
	})
}

func TestRestoreSourceBOM(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	markedPath := writeTestFile(t, dir, "marked.ts", []byte(BOM+"const value = 1;"))
	doubleMarkedPath := writeTestFile(t, dir, "double-marked.ts", []byte(BOM+BOM+"const value = 1;"))
	cleanPath := writeTestFile(t, dir, "clean.ts", []byte("const value = 1;"))
	fs := osvfs.FS()

	for _, tt := range []struct {
		name string
		path string
		text string
		want string
	}{
		{
			name: "restore stripped encoding mark",
			path: markedPath,
			text: "const value = 1;",
			want: BOM + "const value = 1;",
		},
		{
			name: "leave unmarked source unchanged",
			path: cleanPath,
			text: "const value = 1;",
			want: "const value = 1;",
		},
		{
			name: "preserve leading content mark separately",
			path: doubleMarkedPath,
			text: BOM + "const value = 1;",
			want: BOM + BOM + "const value = 1;",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got, err := RestoreSourceBOM(fs, tt.path, tt.text); got != tt.want || err != nil {
				t.Errorf("RestoreSourceBOM() = (%q, %v), want (%q, nil)", got, err, tt.want)
			}
		})
	}
}

func TestDiskBOMRetriesDescriptorExhaustion(t *testing.T) {
	t.Parallel()

	path := writeTestFile(t, t.TempDir(), "source.ts", []byte(BOM))
	processLimit := syscall.EMFILE
	if runtime.GOOS == "windows" {
		processLimit = syscall.Errno(4) // ERROR_TOO_MANY_OPEN_FILES
	}
	for _, tc := range []struct {
		name     string
		openErrs []error
		wantErr  error
		wantBOM  bool
	}{
		{"descriptor released", []error{processLimit, nil}, nil, true},
		{"retry denied", []error{processLimit, os.ErrPermission}, os.ErrPermission, false},
		{"retry I/O error", []error{processLimit, syscall.EIO}, syscall.EIO, false},
		{"file disappeared", []error{processLimit, os.ErrNotExist}, nil, false},
		{"first open denied", []error{os.ErrPermission}, os.ErrPermission, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			synctest.Test(t, func(t *testing.T) {
				attempts := 0
				has, err := readDiskBOM(path, func(path string) (*os.File, error) {
					if attempts >= len(tc.openErrs) {
						t.Fatal("unexpected extra open")
					}
					openErr := tc.openErrs[attempts]
					attempts++
					if openErr != nil {
						return nil, &os.PathError{Op: "open", Path: path, Err: openErr}
					}
					return os.Open(path)
				})
				if attempts != len(tc.openErrs) || has != tc.wantBOM || !errors.Is(err, tc.wantErr) {
					t.Errorf("got (%t, %v) in %d attempts, want (%t, %v) in %d", has, err, attempts, tc.wantBOM, tc.wantErr, len(tc.openErrs))
				}
			})
		})
	}
}

func TestBOMHeaderReadError(t *testing.T) {
	// A read failure after two bytes is distinct from the end of a short file.
	reader := io.MultiReader(bytes.NewReader([]byte{0xff, 0xfe}), iotest.ErrReader(syscall.EIO))
	if has, err := readBOMHeader(reader); has || !errors.Is(err, syscall.EIO) {
		t.Fatalf("incomplete read returned (%t, %v), want I/O error", has, err)
	}
}

func TestHealthyDiskBOMReadsProceedDuringDescriptorRetry(t *testing.T) {
	path := writeTestFile(t, t.TempDir(), "marked.ts", []byte(BOM))
	synctest.Test(t, func(t *testing.T) {
		start := func(open func(string) (*os.File, error)) <-chan bool {
			done := make(chan bool, 1)
			go func() {
				has, err := readDiskBOM(path, open)
				done <- has && err == nil
			}()
			return done
		}
		checkHealthy := func() {
			t.Helper()
			done := start(os.Open)
			synctest.Wait()
			select {
			case ok := <-done:
				if !ok {
					t.Fatal("healthy read lost the BOM")
				}
			default:
				t.Fatal("healthy read blocked behind an unrelated file")
			}
		}
		slow := make(chan struct{})
		retry := make(chan struct{})
		defer close(slow)
		defer close(retry)
		start(func(path string) (*os.File, error) {
			<-slow
			return os.Open(path)
		})
		synctest.Wait()
		retryStarted := false
		attempts := 0
		start(func(path string) (*os.File, error) {
			attempts++
			if attempts == 1 {
				return nil, syscall.EMFILE
			}
			retryStarted = true
			<-retry
			return os.Open(path)
		})
		synctest.Wait()
		checkHealthy()
		time.Sleep(10 * time.Millisecond)
		synctest.Wait()
		if !retryStarted {
			t.Fatal("failed open did not retry")
		}
		checkHealthy()
	})
}

func TestDiskBOMRetriesDescriptorsReleasedElsewhere(t *testing.T) {
	path := writeTestFile(t, t.TempDir(), "marked.ts", []byte(BOM))
	synctest.Test(t, func(t *testing.T) {
		available := make(chan struct{})
		done := make(chan bool, 1)
		go func() {
			has, err := readDiskBOM(path, func(path string) (*os.File, error) {
				select {
				case <-available:
					return os.Open(path)
				default:
					return nil, syscall.EMFILE
				}
			})
			done <- has && err == nil
		}()
		synctest.Wait()
		time.Sleep(50 * time.Millisecond)
		close(available)
		if !<-done {
			t.Fatal("failed open did not recover the BOM after an external descriptor release")
		}
	})
}

func TestDiskBOMRetryStopsWhileAnotherOpenIsBlocked(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		blocked := make(chan struct{})
		defer close(blocked)
		go readDiskBOM("unrelated.ts", func(string) (*os.File, error) {
			<-blocked
			return nil, os.ErrPermission
		})
		synctest.Wait()

		done := make(chan error, 1)
		go func() {
			attempts := 0
			_, err := readDiskBOM("target.ts", func(string) (*os.File, error) {
				attempts++
				if attempts == 1 {
					return nil, syscall.EMFILE
				}
				return nil, syscall.ENFILE
			})
			done <- err
		}()
		time.Sleep(2 * time.Second)
		synctest.Wait()
		select {
		case err := <-done:
			if !errors.Is(err, syscall.ENFILE) {
				t.Fatalf("got %v, want the descriptor-limit error", err)
			}
		default:
			t.Fatal("descriptor recovery waited indefinitely for an unrelated open")
		}
	})
}

type failingBOMSource struct {
	vfs.FS
	err error
}

func (fs failingBOMSource) ReadSourceBOM(string) (bool, error) {
	return false, fs.err
}

func TestRestoreSourceBOMForwardsUnderlyingError(t *testing.T) {
	t.Parallel()

	dir := tspath.NormalizePath(t.TempDir())
	path := tspath.ResolvePath(dir, "source.ts")
	base := failingBOMSource{FS: osvfs.FS(), err: &os.PathError{Op: "open", Path: path, Err: syscall.EMFILE}}
	overlayPath := tspath.ResolvePath(dir, "overlay.ts")
	inner := NewOverlayVFS(base, map[string]string{overlayPath: BOM + "x;"})
	overlay := NewOverlayVFS(inner, nil)
	if text, err := RestoreSourceBOM(overlay, path, "x;"); text != "" || !errors.Is(err, syscall.EMFILE) {
		t.Errorf("unavailable BOM returned (%q, %v)", text, err)
	}
	if text, err := RestoreSourceBOM(overlay, overlayPath, "x;"); text != BOM+"x;" || err != nil {
		t.Errorf("overlay BOM returned (%q, %v)", text, err)
	}
}

type countingBOMSource struct {
	vfs.FS
	calls      atomic.Int64
	fallbackFn func(path string) (bool, error)
}

func (f *countingBOMSource) ReadSourceBOM(path string) (bool, error) {
	f.calls.Add(1)
	if f.fallbackFn != nil {
		return f.fallbackFn(path)
	}
	return ReadSourceBOM(f.FS, path)
}

func encodeTestUTF16(text string, order binary.ByteOrder) []byte {
	units := utf16.Encode([]rune(text))
	out := make([]byte, len(units)*2)
	for i, u := range units {
		order.PutUint16(out[i*2:], u)
	}
	return out
}

func TestBOMTrackingDiskFSParityAndZeroReopen(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cases := []struct {
		name     string
		fileName string
		raw      []byte
		wantText string
		wantBOM  bool
	}{
		{
			name:     "utf-8 BOM",
			fileName: "utf8-bom.ts",
			raw:      []byte(BOM + "#!/usr/bin/env node\nlet a = 1;\n"),
			wantText: "#!/usr/bin/env node\nlet a = 1;\n",
			wantBOM:  true,
		},
		{
			name:     "utf-8 BOM only",
			fileName: "utf8-bom-only.ts",
			raw:      []byte(BOM),
			wantText: "",
			wantBOM:  true,
		},
		{
			name:     "double utf-8 BOM",
			fileName: "double-bom.ts",
			raw:      []byte(BOM + BOM + "let a = 1;\n"),
			wantText: BOM + "let a = 1;\n",
			wantBOM:  true,
		},
		{
			name:     "utf-16le BOM",
			fileName: "utf16le.ts",
			raw:      append([]byte{0xFF, 0xFE}, encodeTestUTF16("const s = '日本😀';\n", binary.LittleEndian)...),
			wantText: "const s = '日本😀';\n",
			wantBOM:  true,
		},
		{
			name:     "utf-16le BOM only",
			fileName: "utf16le-only.ts",
			raw:      []byte{0xFF, 0xFE},
			wantText: "",
			wantBOM:  true,
		},
		{
			name:     "utf-16le BOM odd trailing byte",
			fileName: "utf16le-odd.ts",
			raw:      []byte{0xFF, 0xFE, 'a', 0x00, 'x'},
			wantText: "a",
			wantBOM:  true,
		},
		{
			name:     "utf-16be BOM",
			fileName: "utf16be.ts",
			raw:      append([]byte{0xFE, 0xFF}, encodeTestUTF16("const s = '日本';\n", binary.BigEndian)...),
			wantText: "const s = '日本';\n",
			wantBOM:  true,
		},
		{
			name:     "utf-16be BOM only",
			fileName: "utf16be-only.ts",
			raw:      []byte{0xFE, 0xFF},
			wantText: "",
			wantBOM:  true,
		},
		{
			name:     "no BOM",
			fileName: "clean.ts",
			raw:      []byte("const 日本 = 1;\n"),
			wantText: "const 日本 = 1;\n",
			wantBOM:  false,
		},
		{
			name:     "empty file",
			fileName: "empty.ts",
			raw:      nil,
			wantText: "",
			wantBOM:  false,
		},
		{
			name:     "one byte",
			fileName: "short.ts",
			raw:      []byte{0xFF},
			wantText: "\xFF",
			wantBOM:  false,
		},
		{
			name:     "truncated utf-8 BOM",
			fileName: "truncated-bom.ts",
			raw:      []byte{0xEF, 0xBB},
			wantText: "\xEF\xBB",
			wantBOM:  false,
		},
	}

	fallback := &countingBOMSource{FS: osvfs.FS()}
	fsys := BundledFS(CachedFS(newBOMTrackingDiskFS(fallback, nil)))
	baselineFS := osvfs.FS()

	type readRecord struct {
		path string
		text string
	}
	records := make([]readRecord, len(cases))

	for i, tc := range cases {
		path := writeTestFile(t, dir, tc.fileName, tc.raw)

		wantFile, wantFileOK := baselineFS.ReadFile(path)
		wantDiskBOM, wantDiskBOMErr := ReadSourceBOM(baselineFS, path)
		if !wantFileOK || wantDiskBOMErr != nil {
			t.Fatalf("%s: baseline osvfs failed: ok=%v err=%v", tc.name, wantFileOK, wantDiskBOMErr)
		}

		gotText, gotOK := fsys.ReadFile(path)
		if !gotOK || gotText != tc.wantText || gotText != wantFile {
			t.Fatalf("%s: ReadFile = (%q, %v), want (%q, true) and osvfs %q", tc.name, gotText, gotOK, tc.wantText, wantFile)
		}
		if tc.wantBOM != wantDiskBOM {
			t.Fatalf("%s: test expectation BOM=%v disagrees with osvfs BOM=%v", tc.name, tc.wantBOM, wantDiskBOM)
		}

		records[i] = readRecord{path: path, text: gotText}
	}

	// Mutate and remove files on disk after ReadFile to prove ReadSourceBOM and
	// RestoreSourceBOM serve the snapshot captured during ReadFile with zero
	// extra disk opens.
	for i, rec := range records {
		if i%2 == 0 {
			if err := os.Remove(rec.path); err != nil {
				t.Fatalf("remove %s: %v", rec.path, err)
			}
		} else {
			flipped := []byte("overwritten-without-bom")
			if !cases[i].wantBOM {
				flipped = []byte(BOM + "overwritten-with-bom")
			}
			if err := os.WriteFile(rec.path, flipped, 0o644); err != nil {
				t.Fatalf("overwrite %s: %v", rec.path, err)
			}
		}
	}

	for i, tc := range cases {
		rec := records[i]
		gotBOM, err := ReadSourceBOM(fsys, rec.path)
		if err != nil || gotBOM != tc.wantBOM {
			t.Errorf("%s: ReadSourceBOM = (%v, %v), want (%v, nil)", tc.name, gotBOM, err, tc.wantBOM)
		}
		if SourceHasBOM(fsys, rec.path) != tc.wantBOM {
			t.Errorf("%s: SourceHasBOM = %v, want %v", tc.name, !tc.wantBOM, tc.wantBOM)
		}
		wantRestore := rec.text
		if tc.wantBOM {
			wantRestore = BOM + rec.text
		}
		restored, err := RestoreSourceBOM(fsys, rec.path, rec.text)
		if err != nil || restored != wantRestore {
			t.Errorf("%s: RestoreSourceBOM = (%q, %v), want (%q, nil)", tc.name, restored, err, wantRestore)
		}
	}

	if opens := fallback.calls.Load(); opens != 0 {
		t.Fatalf("fallback disk opens = %d, want 0 after ReadFile", opens)
	}
}

func TestBOMTrackingDiskFSUnreadFallbackAndErrors(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	marked := writeTestFile(t, dir, "marked.ts", []byte(BOM+"let x = 1;\n"))
	clean := writeTestFile(t, dir, "clean.ts", []byte("let x = 1;\n"))
	missing := tspath.NormalizePath(filepath.Join(dir, "missing.ts"))

	fallback := &countingBOMSource{FS: osvfs.FS()}
	fsys := BundledFS(CachedFS(newBOMTrackingDiskFS(fallback, nil)))

	// Unread files fall back to disk probing.
	if has, err := ReadSourceBOM(fsys, marked); !has || err != nil {
		t.Fatalf("unread marked file = (%v, %v), want (true, nil)", has, err)
	}
	if has, err := ReadSourceBOM(fsys, clean); has || err != nil {
		t.Fatalf("unread clean file = (%v, %v), want (false, nil)", has, err)
	}
	if fallback.calls.Load() != 2 {
		t.Fatalf("unread fallback opens = %d, want 2", fallback.calls.Load())
	}

	// Missing files and directories return ok=false on ReadFile and do not error
	// on missing-file ReadSourceBOM.
	if _, ok := fsys.ReadFile(missing); ok {
		t.Fatal("ReadFile succeeded for missing file")
	}
	if has, err := ReadSourceBOM(fsys, missing); has || err != nil {
		t.Fatalf("missing ReadSourceBOM = (%v, %v), want (false, nil)", has, err)
	}
	if _, ok := fsys.ReadFile(tspath.NormalizePath(dir)); ok {
		t.Fatal("ReadFile succeeded for directory")
	}

	// Underlying BOMSource errors must pass through BundledFS(CachedFS(tracker))
	// for unread paths, failed reads, and reads that fail after a previous hit.
	t.Run("error forwarding and failed read invalidation", func(t *testing.T) {
		t.Parallel()

		failNext := false
		emfileErr := &os.PathError{Op: "open", Path: marked, Err: syscall.EMFILE}
		base := failingBOMSource{FS: osvfs.FS(), err: emfileErr}
		errTracker := newBOMTrackingDiskFS(base, func(rootName string, rest string) ([]byte, error) {
			if failNext {
				return nil, os.ErrPermission
			}
			return os.ReadFile(filepath.Join(rootName, rest))
		})
		wrapped := BundledFS(CachedFS(errTracker))

		// Unread path forwards underlying BOMSource error.
		if text, err := RestoreSourceBOM(wrapped, marked, "let x = 1;\n"); text != "" || !errors.Is(err, syscall.EMFILE) {
			t.Fatalf("unread RestoreSourceBOM = (%q, %v), want EMFILE", text, err)
		}

		// Failed initial ReadFile must not cache false and mask the error.
		failNext = true
		if _, ok := wrapped.ReadFile(marked); ok {
			t.Fatal("ReadFile unexpectedly succeeded while failing")
		}
		if text, err := RestoreSourceBOM(wrapped, marked, "let x = 1;\n"); text != "" || !errors.Is(err, syscall.EMFILE) {
			t.Fatalf("post-failure RestoreSourceBOM = (%q, %v), want EMFILE", text, err)
		}

		// Successful ReadFile caches BOM; a later failed ReadFile invalidates it.
		failNext = false
		if text, ok := wrapped.ReadFile(marked); !ok || text != "let x = 1;\n" {
			t.Fatalf("ReadFile = (%q, %v), want success", text, ok)
		}
		if text, err := RestoreSourceBOM(wrapped, marked, "let x = 1;\n"); text != BOM+"let x = 1;\n" || err != nil {
			t.Fatalf("cached RestoreSourceBOM = (%q, %v), want marked text", text, err)
		}

		failNext = true
		if _, ok := wrapped.ReadFile(marked); ok {
			t.Fatal("second ReadFile unexpectedly succeeded while failing")
		}
		if text, err := RestoreSourceBOM(wrapped, marked, "let x = 1;\n"); text != "" || !errors.Is(err, syscall.EMFILE) {
			t.Fatalf("invalidated RestoreSourceBOM = (%q, %v), want EMFILE", text, err)
		}
	})

	// Non-absolute ReadFile panics just like osvfs.FS().
	t.Run("non-absolute path panics", func(t *testing.T) {
		t.Parallel()
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("ReadFile on relative path did not panic")
			}
		}()
		NewDiskFSWithBOM().ReadFile("relative/file.ts")
	})
}

func TestBundledFSReadSourceBOM(t *testing.T) {
	t.Parallel()

	sentinel := &os.PathError{Op: "open", Path: "/sentinel.ts", Err: syscall.EMFILE}
	fsys := BundledFS(CachedFS(failingBOMSource{FS: osvfs.FS(), err: sentinel}))

	if bundled.Embedded {
		libFile := tspath.CombinePaths(bundled.LibPath(), "lib.d.ts")
		if _, ok := fsys.ReadFile(libFile); !ok {
			t.Fatalf("ReadFile(%q) failed", libFile)
		}
		if has, err := ReadSourceBOM(fsys, libFile); has || err != nil {
			t.Fatalf("bundled ReadSourceBOM(%q) = (%v, %v), want (false, nil)", libFile, has, err)
		}
	}

	if has, err := ReadSourceBOM(fsys, "/sentinel.ts"); has || !errors.Is(err, syscall.EMFILE) {
		t.Fatalf("non-bundled ReadSourceBOM = (%v, %v), want (false, EMFILE)", has, err)
	}
}

func TestBOMTrackingDiskFSInvalidationAndFixGenerations(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	fileA := writeTestFile(t, dir, "a.ts", []byte(BOM+"var a = 1;\n"))
	fileB := writeTestFile(t, dir, "b.ts", []byte(BOM+"var b = 1;\n"))
	emptyFile := writeTestFile(t, dir, "empty.ts", nil)

	subDir := filepath.Join(dir, "nested")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}
	nestedFile := writeTestFile(t, subDir, "child.ts", []byte(BOM+"export {};\n"))

	fallback := &countingBOMSource{FS: osvfs.FS()}
	fsys := BundledFS(CachedFS(newBOMTrackingDiskFS(fallback, nil)))

	// CachedFS must not promote ClearCache without invalidating bomTrackingDiskFS.
	if _, hasClearCache := CachedFS(NewDiskFSWithBOM()).(interface{ ClearCache() }); hasClearCache {
		t.Fatal("CachedFS unexpectedly exposed ClearCache")
	}

	// Trailing-slash ReadFile normalizes to the same key.
	if _, ok := fsys.ReadFile(fileB + "/"); !ok {
		t.Fatal("ReadFile(fileB/) failed")
	}
	if has, err := ReadSourceBOM(fsys, fileB); !has || err != nil || fallback.calls.Load() != 0 {
		t.Fatalf("ReadSourceBOM(fileB) after trailing-slash read = (%v, %v) calls=%d, want (true, nil) calls=0", has, err, fallback.calls.Load())
	}

	// WriteFile invalidates cached BOM in both directions.
	if _, ok := fsys.ReadFile(fileA); !ok {
		t.Fatal("ReadFile(fileA) failed")
	}
	if has, err := ReadSourceBOM(fsys, fileA); !has || err != nil {
		t.Fatalf("initial ReadSourceBOM(fileA) = (%v, %v), want (true, nil)", has, err)
	}
	if err := fsys.WriteFile(fileA, "let a = 1;\n"); err != nil {
		t.Fatalf("WriteFile(fileA) unmarked: %v", err)
	}
	if has, err := ReadSourceBOM(fsys, fileA); has || err != nil {
		t.Fatalf("post-WriteFile unmarked ReadSourceBOM(fileA) = (%v, %v), want (false, nil)", has, err)
	}
	if err := fsys.WriteFile(fileA, BOM+"let a = 1;\n"); err != nil {
		t.Fatalf("WriteFile(fileA) marked: %v", err)
	}
	if has, err := ReadSourceBOM(fsys, fileA); !has || err != nil {
		t.Fatalf("post-WriteFile marked ReadSourceBOM(fileA) = (%v, %v), want (true, nil)", has, err)
	}

	// AppendFile invalidates cached BOM (empty file gaining a BOM).
	if _, ok := fsys.ReadFile(emptyFile); !ok {
		t.Fatal("ReadFile(emptyFile) failed")
	}
	if has, err := ReadSourceBOM(fsys, emptyFile); has || err != nil {
		t.Fatalf("initial ReadSourceBOM(emptyFile) = (%v, %v), want (false, nil)", has, err)
	}
	if err := fsys.AppendFile(emptyFile, BOM+"const x = 1;\n"); err != nil {
		t.Fatalf("AppendFile(emptyFile): %v", err)
	}
	if has, err := ReadSourceBOM(fsys, emptyFile); !has || err != nil {
		t.Fatalf("post-AppendFile ReadSourceBOM(emptyFile) = (%v, %v), want (true, nil)", has, err)
	}

	// Remove invalidates both direct files and descendants of a removed directory.
	if _, ok := fsys.ReadFile(nestedFile); !ok {
		t.Fatal("ReadFile(nestedFile) failed")
	}
	if has, err := ReadSourceBOM(fsys, nestedFile); !has || err != nil {
		t.Fatalf("initial ReadSourceBOM(nestedFile) = (%v, %v), want (true, nil)", has, err)
	}
	if err := fsys.Remove(tspath.NormalizePath(subDir)); err != nil {
		t.Fatalf("Remove(subDir): %v", err)
	}
	if has, err := ReadSourceBOM(fsys, nestedFile); has || err != nil {
		t.Fatalf("post-Remove ReadSourceBOM(nestedFile) = (%v, %v), want (false, nil)", has, err)
	}

	// Multi-pass --fix: in-memory overlay supersedes pass 1's cached disk BOM
	// for edited files while preserving pass 1's snapshot for unedited siblings,
	// and a fresh generation observes newly written disk contents.
	gen1 := NewCachedDiskFS()
	if _, ok := gen1.ReadFile(fileA); !ok {
		t.Fatal("gen1 ReadFile(fileA) failed")
	}
	if _, ok := gen1.ReadFile(fileB); !ok {
		t.Fatal("gen1 ReadFile(fileB) failed")
	}

	pass2 := NewOverlayVFS(gen1, map[string]string{
		fileA: "let a = 1;\n",
	})
	if has, err := ReadSourceBOM(pass2, fileA); has || err != nil {
		t.Fatalf("pass2 overlay ReadSourceBOM(fileA) = (%v, %v), want (false, nil)", has, err)
	}
	if has, err := ReadSourceBOM(pass2, fileB); !has || err != nil {
		t.Fatalf("pass2 sibling ReadSourceBOM(fileB) = (%v, %v), want (true, nil)", has, err)
	}

	if err := os.WriteFile(fileA, []byte("let a = 1;\n"), 0o644); err != nil {
		t.Fatalf("commit fileA: %v", err)
	}
	gen2 := NewCachedDiskFS()
	if text, ok := gen2.ReadFile(fileA); !ok || text != "let a = 1;\n" {
		t.Fatalf("gen2 ReadFile(fileA) = (%q, %v), want unmarked committed text", text, ok)
	}
	if has, err := ReadSourceBOM(gen2, fileA); has || err != nil {
		t.Fatalf("gen2 ReadSourceBOM(fileA) = (%v, %v), want (false, nil)", has, err)
	}
}

type stubCaseSensitivityFS struct {
	vfs.FS
	caseSensitive bool
}

func (f stubCaseSensitivityFS) UseCaseSensitiveFileNames() bool {
	return f.caseSensitive
}

func TestBOMTrackingDiskFSCaseSensitivityAndWindowsPaths(t *testing.T) {
	t.Parallel()

	t.Run("case-insensitive Windows drive paths", func(t *testing.T) {
		t.Parallel()

		fallback := &countingBOMSource{
			FS: stubCaseSensitivityFS{FS: osvfs.FS(), caseSensitive: false},
			fallbackFn: func(path string) (bool, error) {
				if path == `c:/repo/src/file.ts` {
					return true, nil
				}
				return false, nil
			},
		}
		tracker := newBOMTrackingDiskFS(
			fallback,
			func(rootName string, rest string) ([]byte, error) {
				if rootName != "C:/" || rest != "Repo/Src/File.ts" {
					return nil, os.ErrNotExist
				}
				return []byte(BOM + "export const x = 1;\n"), nil
			},
		)
		fsys := BundledFS(CachedFS(tracker))

		if text, ok := fsys.ReadFile(`C:\Repo\Src\File.ts`); !ok || text != "export const x = 1;\n" {
			t.Fatalf("ReadFile Windows path = (%q, %v)", text, ok)
		}

		for _, query := range []string{
			`C:\Repo\Src\File.ts`,
			`C:/Repo/Src/File.ts`,
			`c:/Repo/Src/File.ts`,
			`c:\Repo\Src\File.ts`,
			`C:/Repo/Src/File.ts/`,
			`C:/Repo/other/../Src/File.ts`,
		} {
			if has, err := ReadSourceBOM(fsys, query); !has || err != nil {
				t.Errorf("ReadSourceBOM(%q) = (%v, %v), want (true, nil)", query, has, err)
			}
		}
		if opens := fallback.calls.Load(); opens != 0 {
			t.Fatalf("same-case Windows path fallback opens = %d, want 0", opens)
		}

		// Alternate directory/file casing safely misses the exact-case cache and
		// falls back to the underlying FS instead of colliding with distinct-case
		// siblings.
		if has, err := ReadSourceBOM(fsys, `c:/repo/src/file.ts`); !has || err != nil {
			t.Fatalf("alternate-case ReadSourceBOM = (%v, %v), want (true, nil)", has, err)
		}
		if opens := fallback.calls.Load(); opens != 1 {
			t.Fatalf("alternate-case fallback opens = %d, want 1", opens)
		}
	})

	t.Run("mixed-volume case-sensitive siblings on case-insensitive OS", func(t *testing.T) {
		t.Parallel()

		fallback := &countingBOMSource{
			FS:         stubCaseSensitivityFS{FS: osvfs.FS(), caseSensitive: false},
			fallbackFn: func(string) (bool, error) { return false, nil },
		}
		tracker := newBOMTrackingDiskFS(
			fallback,
			func(rootName string, rest string) ([]byte, error) {
				if rootName != "/" {
					return nil, os.ErrNotExist
				}
				switch rest {
				case "Volumes/CaseSensitive/Model.ts":
					return []byte(BOM + "export const upper = 1;\n"), nil
				case "Volumes/CaseSensitive/model.ts":
					return []byte("export const lower = 1;\n"), nil
				default:
					return nil, os.ErrNotExist
				}
			},
		)
		fsys := BundledFS(CachedFS(tracker))

		const upperPath = "/Volumes/CaseSensitive/Model.ts"
		const lowerPath = "/Volumes/CaseSensitive/model.ts"

		if text, ok := fsys.ReadFile(upperPath); !ok || text != "export const upper = 1;\n" {
			t.Fatalf("ReadFile(%q) = (%q, %v)", upperPath, text, ok)
		}
		if text, ok := fsys.ReadFile(lowerPath); !ok || text != "export const lower = 1;\n" {
			t.Fatalf("ReadFile(%q) = (%q, %v)", lowerPath, text, ok)
		}

		if has, err := ReadSourceBOM(fsys, upperPath); !has || err != nil {
			t.Fatalf("ReadSourceBOM(%q) = (%v, %v), want (true, nil)", upperPath, has, err)
		}
		if has, err := ReadSourceBOM(fsys, lowerPath); has || err != nil {
			t.Fatalf("ReadSourceBOM(%q) = (%v, %v), want (false, nil)", lowerPath, has, err)
		}
		if opens := fallback.calls.Load(); opens != 0 {
			t.Fatalf("fallback opens = %d, want 0", opens)
		}
	})

	t.Run("case-sensitive POSIX paths stay distinct", func(t *testing.T) {
		t.Parallel()

		fallback := &countingBOMSource{
			FS:         stubCaseSensitivityFS{FS: osvfs.FS(), caseSensitive: true},
			fallbackFn: func(string) (bool, error) { return false, nil },
		}
		tracker := newBOMTrackingDiskFS(
			fallback,
			func(rootName string, rest string) ([]byte, error) {
				if rootName != "/" || rest != "Repo/File.ts" {
					return nil, os.ErrNotExist
				}
				return []byte(BOM + "export const x = 1;\n"), nil
			},
		)
		fsys := BundledFS(CachedFS(tracker))

		if text, ok := fsys.ReadFile("/Repo/File.ts"); !ok || text != "export const x = 1;\n" {
			t.Fatalf("ReadFile POSIX path = (%q, %v)", text, ok)
		}
		if has, err := ReadSourceBOM(fsys, "/Repo/File.ts"); !has || err != nil {
			t.Fatalf("exact case ReadSourceBOM = (%v, %v), want (true, nil)", has, err)
		}
		if fallback.calls.Load() != 0 {
			t.Fatalf("exact case fallback opens = %d, want 0", fallback.calls.Load())
		}
		if has, err := ReadSourceBOM(fsys, "/Repo/file.ts"); has || err != nil {
			t.Fatalf("different case ReadSourceBOM = (%v, %v), want (false, nil)", has, err)
		}
		if fallback.calls.Load() != 1 {
			t.Fatalf("different case fallback opens = %d, want 1", fallback.calls.Load())
		}
	})
}

func TestBOMTrackingDiskFSConcurrent(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	const fileCount = 16
	paths := make([]string, fileCount)
	for i := range fileCount {
		content := "export const v = 1;\n"
		if i%2 == 0 {
			content = BOM + content
		}
		paths[i] = writeTestFile(t, dir, "file-"+string(rune('a'+i))+".ts", []byte(content))
	}
	mutablePath := writeTestFile(t, dir, "mutable.ts", []byte(BOM+"export const m = 1;\n"))

	fsys := NewCachedDiskFS()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for round := range 20 {
			content := "export const m = 1;\n"
			if round%2 == 0 {
				content = BOM + content
			}
			_ = fsys.WriteFile(mutablePath, content)
			_, _ = fsys.ReadFile(mutablePath)
			_, _ = ReadSourceBOM(fsys, mutablePath)
		}
	}()
	for g := range 8 {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for round := range 20 {
				idx := (worker + round) % fileCount
				path := paths[idx]
				wantBOM := idx%2 == 0
				text, ok := fsys.ReadFile(path)
				if !ok || text != "export const v = 1;\n" {
					t.Errorf("worker %d ReadFile(%s) = (%q, %v)", worker, path, text, ok)
					return
				}
				gotBOM, err := ReadSourceBOM(fsys, path)
				if err != nil || gotBOM != wantBOM {
					t.Errorf("worker %d ReadSourceBOM(%s) = (%v, %v), want (%v, nil)", worker, path, gotBOM, err, wantBOM)
					return
				}
				wantText := text
				if wantBOM {
					wantText = BOM + text
				}
				if restored, err := RestoreSourceBOM(fsys, path, text); err != nil || restored != wantText {
					t.Errorf("worker %d RestoreSourceBOM(%s) = (%q, %v), want (%q, nil)", worker, path, restored, err, wantText)
					return
				}
				_, _ = ReadSourceBOM(fsys, mutablePath)
			}
		}(g)
	}
	wg.Wait()
}

func TestBOMTrackingDiskFSConcurrentMutationRace(t *testing.T) {
	t.Parallel()

	t.Run("WriteFile invalidates in-flight ReadFile", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := writeTestFile(t, dir, "race.ts", []byte(BOM+"let a = 1;\n"))

		readStarted := make(chan struct{})
		releaseRead := make(chan struct{})
		tracker := newBOMTrackingDiskFS(osvfs.FS(), func(rootName string, rest string) ([]byte, error) {
			raw, err := os.ReadFile(filepath.Join(rootName, rest))
			close(readStarted)
			<-releaseRead
			return raw, err
		})

		type readResult struct {
			text string
			ok   bool
		}
		done := make(chan readResult, 1)
		go func() {
			text, ok := tracker.ReadFile(path)
			done <- readResult{text: text, ok: ok}
		}()

		<-readStarted
		if err := tracker.WriteFile(path, "let a = 2;\n"); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
		close(releaseRead)

		res := <-done
		if !res.ok || res.text != "let a = 1;\n" {
			t.Fatalf("in-flight ReadFile = (%q, %v), want (%q, true)", res.text, res.ok, "let a = 1;\n")
		}
		if has, err := ReadSourceBOM(tracker, path); has || err != nil {
			t.Fatalf("ReadSourceBOM after concurrent WriteFile = (%v, %v), want (false, nil)", has, err)
		}
	})

	t.Run("Remove invalidates in-flight ReadFile", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := writeTestFile(t, dir, "removed.ts", []byte(BOM+"let a = 1;\n"))

		readStarted := make(chan struct{})
		releaseRead := make(chan struct{})
		tracker := newBOMTrackingDiskFS(osvfs.FS(), func(rootName string, rest string) ([]byte, error) {
			raw, err := os.ReadFile(filepath.Join(rootName, rest))
			close(readStarted)
			<-releaseRead
			return raw, err
		})

		done := make(chan bool, 1)
		go func() {
			_, ok := tracker.ReadFile(path)
			done <- ok
		}()

		<-readStarted
		if err := tracker.Remove(path); err != nil {
			t.Fatalf("Remove: %v", err)
		}
		close(releaseRead)
		if !<-done {
			t.Fatal("in-flight ReadFile unexpectedly failed")
		}

		if has, err := ReadSourceBOM(tracker, path); has || err != nil {
			t.Fatalf("ReadSourceBOM after concurrent Remove = (%v, %v), want (false, nil)", has, err)
		}
	})

	t.Run("stale failed ReadFile does not evict newer epoch", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := tspath.NormalizePath(filepath.Join(dir, "created.ts"))

		readStarted := make(chan struct{})
		releaseRead := make(chan struct{})
		var calls atomic.Int32
		fallback := &countingBOMSource{FS: osvfs.FS()}
		tracker := newBOMTrackingDiskFS(fallback, func(rootName string, rest string) ([]byte, error) {
			if calls.Add(1) == 1 {
				close(readStarted)
				<-releaseRead
				return nil, os.ErrNotExist
			}
			return os.ReadFile(filepath.Join(rootName, rest))
		})

		done := make(chan bool, 1)
		go func() {
			_, ok := tracker.ReadFile(path)
			done <- ok
		}()

		<-readStarted
		if err := tracker.WriteFile(path, BOM+"let created = 1;\n"); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
		if text, ok := tracker.ReadFile(path); !ok || text != "let created = 1;\n" {
			t.Fatalf("newer ReadFile = (%q, %v), want success", text, ok)
		}
		close(releaseRead)
		if <-done {
			t.Fatal("stale ReadFile unexpectedly succeeded")
		}

		if has, err := ReadSourceBOM(tracker, path); !has || err != nil {
			t.Fatalf("ReadSourceBOM after stale failure = (%v, %v), want (true, nil)", has, err)
		}
		if opens := fallback.calls.Load(); opens != 0 {
			t.Fatalf("fallback opens = %d, want 0 (newer epoch entry evicted)", opens)
		}
	})
}

func TestBOMTrackingDiskFSSymlinkAliasInvalidation(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	realPath := writeTestFile(t, dir, "real.ts", []byte(BOM+"let a = 1;\n"))
	linkPath := tspath.NormalizePath(filepath.Join(dir, "link.ts"))
	if err := os.Symlink(realPath, linkPath); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	fsys := NewCachedDiskFS()
	if text, ok := fsys.ReadFile(linkPath); !ok || text != "let a = 1;\n" {
		t.Fatalf("ReadFile(linkPath) = (%q, %v), want (%q, true)", text, ok, "let a = 1;\n")
	}
	if has, err := ReadSourceBOM(fsys, linkPath); !has || err != nil {
		t.Fatalf("initial ReadSourceBOM(linkPath) = (%v, %v), want (true, nil)", has, err)
	}

	if err := fsys.WriteFile(realPath, "let a = 2;\n"); err != nil {
		t.Fatalf("WriteFile(realPath): %v", err)
	}
	if has, err := ReadSourceBOM(fsys, linkPath); has || err != nil {
		t.Fatalf("post-WriteFile ReadSourceBOM(linkPath) = (%v, %v), want (false, nil)", has, err)
	}
}

func TestBOMTrackingDiskFSRelativeRemoveInvalidation(t *testing.T) {
	dir := t.TempDir()
	subDir := filepath.Join(dir, "subdir")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("mkdir subdir: %v", err)
	}
	childPath := writeTestFile(t, subDir, "child.ts", []byte(BOM+"export const x = 1;\n"))

	t.Chdir(dir)

	fsys := NewCachedDiskFS()
	if text, ok := fsys.ReadFile(childPath); !ok || text != "export const x = 1;\n" {
		t.Fatalf("ReadFile(childPath) = (%q, %v), want success", text, ok)
	}
	if has, err := ReadSourceBOM(fsys, childPath); !has || err != nil {
		t.Fatalf("initial ReadSourceBOM(childPath) = (%v, %v), want (true, nil)", has, err)
	}

	if err := fsys.Remove("subdir"); err != nil {
		t.Fatalf("Remove(\"subdir\"): %v", err)
	}
	if has, err := ReadSourceBOM(fsys, childPath); has || err != nil {
		t.Fatalf("post-Remove ReadSourceBOM(childPath) = (%v, %v), want (false, nil)", has, err)
	}
}

func TestReadDiskFileWithBOM(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cases := []struct {
		name     string
		fileName string
		raw      []byte
		want     string
	}{
		{
			name:     "utf-8 BOM",
			fileName: "utf8-bom.ts",
			raw:      []byte(BOM + "let a = 1;\n"),
			want:     BOM + "let a = 1;\n",
		},
		{
			name:     "utf-8 BOM only",
			fileName: "utf8-bom-only.ts",
			raw:      []byte(BOM),
			want:     BOM,
		},
		{
			name:     "double utf-8 BOM",
			fileName: "double-bom.ts",
			raw:      []byte(BOM + BOM + "let a = 1;\n"),
			want:     BOM + BOM + "let a = 1;\n",
		},
		{
			name:     "utf-16le BOM",
			fileName: "utf16le.ts",
			raw:      append([]byte{0xFF, 0xFE}, encodeTestUTF16("const s = '日本😀';\n", binary.LittleEndian)...),
			want:     BOM + "const s = '日本😀';\n",
		},
		{
			name:     "utf-16be BOM",
			fileName: "utf16be.ts",
			raw:      append([]byte{0xFE, 0xFF}, encodeTestUTF16("const s = '日本';\n", binary.BigEndian)...),
			want:     BOM + "const s = '日本';\n",
		},
		{
			name:     "unmarked",
			fileName: "clean.ts",
			raw:      []byte("const x = 1;\n"),
			want:     "const x = 1;\n",
		},
		{
			name:     "empty",
			fileName: "empty.ts",
			raw:      nil,
			want:     "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			path := writeTestFile(t, dir, tc.fileName, tc.raw)
			got, ok := ReadDiskFileWithBOM(path)
			if !ok || got != tc.want {
				t.Fatalf("ReadDiskFileWithBOM(%s) = (%q, %v), want (%q, true)", tc.name, got, ok, tc.want)
			}
		})
	}

	t.Run("missing file", func(t *testing.T) {
		t.Parallel()
		missing := tspath.NormalizePath(filepath.Join(dir, "missing.ts"))
		if got, ok := ReadDiskFileWithBOM(missing); ok || got != "" {
			t.Fatalf("ReadDiskFileWithBOM(missing) = (%q, %v), want (\"\", false)", got, ok)
		}
	})

	t.Run("directory", func(t *testing.T) {
		t.Parallel()
		if got, ok := ReadDiskFileWithBOM(tspath.NormalizePath(dir)); ok || got != "" {
			t.Fatalf("ReadDiskFileWithBOM(dir) = (%q, %v), want (\"\", false)", got, ok)
		}
	})

	t.Run("non-absolute path panics", func(t *testing.T) {
		t.Parallel()
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("ReadDiskFileWithBOM on relative path did not panic")
			}
		}()
		ReadDiskFileWithBOM("relative/file.ts")
	})
}
