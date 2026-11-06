package utils

// cspell:ignore synctest

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
	"testing/iotest"
	"testing/synctest"
	"time"

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
