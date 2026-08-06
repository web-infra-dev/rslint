package utils

import (
	"io"
	"os"

	"github.com/microsoft/typescript-go/shim/vfs"
)

// BOM is the Unicode byte order mark, U+FEFF, in the UTF-8 encoding source
// text uses by the time rslint sees it.
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
	SourceHasBOM(path string) bool
}

// SourceHasBOM reports whether the source text for path began with a byte
// order mark. A VFS that tracks the mark answers directly; anything else falls
// back to the file's own bytes, which is the right answer for every path whose
// text came off disk unchanged.
func SourceHasBOM(fileSystem vfs.FS, path string) bool {
	if source, ok := fileSystem.(BOMSource); ok {
		return source.SourceHasBOM(path)
	}
	return fileBytesStartWithBOM(path)
}

// fileBytesStartWithBOM reports whether the bytes of path begin with a byte
// order mark. UTF-16 marks count: those files are decoded to UTF-8 and lose
// their mark exactly as UTF-8 ones do. A path with no bytes to read — a
// virtual file, a bundled lib — has nothing to speak for it.
func fileBytesStartWithBOM(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	var head [3]byte
	n, err := io.ReadFull(file, head[:])
	if err != nil && n < 2 {
		return false
	}
	if n >= 3 && head[0] == 0xEF && head[1] == 0xBB && head[2] == 0xBF {
		return true
	}
	return (head[0] == 0xFF && head[1] == 0xFE) || (head[0] == 0xFE && head[1] == 0xFF)
}
