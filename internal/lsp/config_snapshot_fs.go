package lsp

import (
	"strings"
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
)

// configSnapshotFS freezes path bases for one discovery transaction. Git source
// and marker observations are cached by the request-local Git reader instead.
type configSnapshotFS struct {
	vfs.FS
	mu               sync.Mutex
	realpathSnapshot map[string]string
	caseSensitive    bool
}

func newConfigSnapshotFS(fsys vfs.FS) *configSnapshotFS {
	caseSensitive := true
	if fsys != nil {
		caseSensitive = fsys.UseCaseSensitiveFileNames()
	}
	return &configSnapshotFS{
		FS:               fsys,
		realpathSnapshot: make(map[string]string),
		caseSensitive:    caseSensitive,
	}
}

// Realpath freezes normalized path lookups for this config evaluation. Cache
// identity follows the filesystem's case sensitivity so alternate casing
// cannot observe two physical generations on a case-insensitive filesystem.
// Target capture anchors leaf resolution in one physical parent. Owner
// selection, Git projection, and authored-base matching that follow it must
// agree on the config directory's identity throughout this transaction.
func (fsys *configSnapshotFS) Realpath(filePath string) string {
	if fsys == nil || fsys.FS == nil {
		return ""
	}
	key := fsys.snapshotKey(filePath)
	fsys.mu.Lock()
	defer fsys.mu.Unlock()
	if realPath, ok := fsys.realpathSnapshot[key]; ok {
		return realPath
	}
	realPath := fsys.FS.Realpath(filePath)
	fsys.realpathSnapshot[key] = realPath
	return realPath
}

func (fsys *configSnapshotFS) snapshotKey(filePath string) string {
	key := tspath.NormalizePath(filePath)
	if !fsys.caseSensitive {
		key = strings.ToLower(key)
	}
	return key
}
