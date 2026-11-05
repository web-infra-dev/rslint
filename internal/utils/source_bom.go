package utils

import (
	"errors"
	"io"
	"os"
	"runtime"
	"sync"
	"syscall"

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
// short to contain a mark return false. Descriptor exhaustion after an
// exclusive retry returns an error.
func ReadSourceBOM(fileSystem vfs.FS, path string) (bool, error) {
	if source, ok := fileSystem.(BOMSource); ok {
		return source.ReadSourceBOM(path)
	}
	if source, ok := fileSystem.(interface{ SourceHasBOM(path string) bool }); ok {
		return source.SourceHasBOM(path), nil
	}
	return readDiskBOM(path, fileBytesStartWithBOM)
}

var diskBOMReads sync.RWMutex

func readDiskBOM(path string, probe func(string) (bool, error)) (bool, error) {
	diskBOMReads.RLock()
	has, err := probe(path)
	diskBOMReads.RUnlock()
	if isFileDescriptorLimit(err) {
		// The writer waits for every active probe to close its file and
		// prevents new probes from borrowing the available descriptor.
		diskBOMReads.Lock()
		has, err = probe(path)
		diskBOMReads.Unlock()
		if isFileDescriptorLimit(err) {
			return false, err
		}
	}
	return has, nil
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

// fileBytesStartWithBOM reports whether the bytes of path begin with a byte
// order mark. UTF-16 marks count: those files are decoded to UTF-8 and lose
// their mark exactly as UTF-8 ones do. A path with no bytes to read — a
// virtual file, a bundled lib — has nothing to speak for it.
func fileBytesStartWithBOM(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	var head [3]byte
	n, err := io.ReadFull(file, head[:])
	if n >= 3 && head[0] == 0xEF && head[1] == 0xBB && head[2] == 0xBF {
		return true, nil
	}
	return n >= 2 && ((head[0] == 0xFF && head[1] == 0xFE) || (head[0] == 0xFE && head[1] == 0xFF)), err
}
