package utils

import (
	"errors"
	"io"
	"os"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/vfs"
)

// BOM is U+FEFF encoded in the UTF-8 strings rslint uses for source text.
const BOM = "\uFEFF"

// BOMSource is implemented by a VFS layer that knows whether the text it hands
// out for a path began with a byte order mark. Source text always reaches the
// linter with the mark already removed — ts-go's file reader decodes it away,
// and [OverlayVFS] strips it from caller-supplied content so both shapes agree
// — which leaves the layer that removed it as the only one able to answer.
//
// A wrapping VFS must forward this explicitly. Wrappers embed the vfs.FS
// interface, and interface embedding promotes only that interface's own
// methods.
type BOMSource interface {
	ReadSourceBOM(path string) (bool, error)
}

// SourceHasBOM supplies native rules with their boolean view of the source.
func SourceHasBOM(fileSystem vfs.FS, path string) bool {
	has, err := ReadSourceBOM(fileSystem, path)
	return err == nil && has
}

// ReadSourceBOM asks the VFS layer that supplied the text, or probes the
// original file when its text came from disk. Missing files and headers too
// short to contain a mark return false. Other file errors reach the caller.
func ReadSourceBOM(fileSystem vfs.FS, path string) (bool, error) {
	if source, ok := fileSystem.(BOMSource); ok {
		return source.ReadSourceBOM(path)
	}
	if source, ok := fileSystem.(interface{ SourceHasBOM(path string) bool }); ok {
		return source.SourceHasBOM(path), nil
	}
	return diskBOMReads.read(path, os.Open)
}

// active counts opens in flight and files held until close. closed lets a
// failed opener notice a file released during its own open attempt.
type bomFileReads struct {
	mu      sync.Mutex
	active  int
	closed  uint64
	changed chan struct{} // allocated only when a failed open needs to wait
}

var diskBOMReads bomFileReads

func (r *bomFileReads) read(path string, openFile func(string) (*os.File, error)) (bool, error) {
	file, err := r.open(path, openFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	has, readErr := readBOMHeader(file)
	closeErr := file.Close()
	r.mu.Lock()
	r.finish(closeErr == nil)
	r.mu.Unlock()
	if readErr != nil {
		return false, readErr
	}
	return has && closeErr == nil, closeErr
}

func (r *bomFileReads) open(path string, openFile func(string) (*os.File, error)) (*os.File, error) {
	finalTry := false
	r.mu.Lock()
	for {
		before := r.closed
		r.active++
		r.mu.Unlock()
		file, err := openFile(path)
		if err == nil {
			return file, nil
		}
		r.mu.Lock()
		r.finish(false)
		if !isFileDescriptorLimit(err) || (finalTry && r.closed == before && r.active == 0) {
			r.mu.Unlock()
			return nil, err
		}
		woke := false
		if r.closed == before && r.active != 0 {
			if r.changed == nil {
				r.changed = make(chan struct{})
			}
			changed := r.changed
			r.mu.Unlock()
			select {
			case <-changed:
				woke = true
			case <-time.After(time.Second):
				// Descriptors released outside BOM probes have no notification.
				// Recheck even if a peer is still blocked in I/O.
			}
			r.mu.Lock()
		}
		// A notification without a close means the active probes drained.
		// Give the OS one last chance when no probe could release a file.
		finalTry = r.closed == before && (woke || r.active == 0)
	}
}

// finish unregisters an attempt under mu. A failed open wakes peers only when
// it was the last attempt, so an exhausted batch can return the OS error.
func (r *bomFileReads) finish(closed bool) {
	r.active--
	if closed {
		r.closed++
	}
	if r.changed != nil && (closed || r.active == 0) {
		close(r.changed)
		r.changed = nil
	}
}

func isFileDescriptorLimit(err error) bool {
	// CreateFile returns the native Windows ERROR_TOO_MANY_OPEN_FILES (4).
	// syscall.EMFILE on Windows is a separate, synthetic errno.
	const windowsTooManyOpenFiles = syscall.Errno(4)
	return errors.Is(err, syscall.EMFILE) || errors.Is(err, syscall.ENFILE) ||
		runtime.GOOS == "windows" && errors.Is(err, windowsTooManyOpenFiles)
}

// RestoreSourceBOM reconstructs the complete source text after a VFS or parser
// has removed its encoding byte order mark. It deliberately does not inspect
// text for a leading U+FEFF: a source can contain both an encoding mark and a
// real U+FEFF as its first character, in which case both must be preserved.
func RestoreSourceBOM(fileSystem vfs.FS, path string, text string) (string, error) {
	has, err := ReadSourceBOM(fileSystem, path)
	if err != nil {
		return "", err
	}
	if has {
		return BOM + text, nil
	}
	return text, nil
}

// A short file can still consist of a complete two-byte UTF-16 mark.
func readBOMHeader(reader io.Reader) (bool, error) {
	var head [3]byte
	n, err := io.ReadFull(reader, head[:])
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		return false, err
	}
	if n == 3 && head == [3]byte{0xEF, 0xBB, 0xBF} {
		return true, nil
	}
	return n >= 2 && ((head[0] == 0xFF && head[1] == 0xFE) || (head[0] == 0xFE && head[1] == 0xFF)), nil
}
