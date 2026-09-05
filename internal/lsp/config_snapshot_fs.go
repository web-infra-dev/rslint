package lsp

import (
	"strings"
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
)

// configSnapshotFS gives one config-discovery transaction a stable view of
// every .gitignore file it observes. cachedvfs deliberately does not cache
// ReadFile, while snapshot preparation may visit one source for multiple
// ownership scopes. Caching those bytes keeps all scopes in the candidate
// catalog consistent; the resulting config entries contain materialized ignore
// patterns and do not retain this filesystem after commit.
type configSnapshotFS struct {
	vfs.FS

	mu                sync.Mutex
	gitignoreSnapshot map[string]configSnapshotFile
	realpathSnapshot  map[string]string
	caseSensitive     bool
}

type configSnapshotFile struct {
	content string
	exists  bool
}

func newConfigSnapshotFS(fsys vfs.FS) *configSnapshotFS {
	caseSensitive := true
	if fsys != nil {
		caseSensitive = fsys.UseCaseSensitiveFileNames()
	}
	return &configSnapshotFS{
		FS:                fsys,
		gitignoreSnapshot: make(map[string]configSnapshotFile),
		realpathSnapshot:  make(map[string]string),
		caseSensitive:     caseSensitive,
	}
}

// Realpath freezes normalized path lookups for this config evaluation. Cache
// identity follows the filesystem's case sensitivity so alternate casing
// cannot observe two physical generations on a case-insensitive filesystem.
// It is intentionally scoped outside target capture: FreezeFileIdentity
// must observe its parent-before/file/parent-after sequence directly, while
// owner selection, Git projection, and authored-base matching that follow it
// must agree with one another.
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

func (fsys *configSnapshotFS) ReadFile(filePath string) (string, bool) {
	if fsys == nil || fsys.FS == nil {
		return "", false
	}
	filePath = tspath.NormalizePath(filePath)
	if tspath.GetBaseFileName(filePath) != ".gitignore" {
		return fsys.FS.ReadFile(filePath)
	}

	key := fsys.snapshotKey(filePath)
	fsys.mu.Lock()
	defer fsys.mu.Unlock()
	if file, ok := fsys.gitignoreSnapshot[key]; ok {
		return file.content, file.exists
	}
	content, exists := fsys.FS.ReadFile(filePath)
	fsys.gitignoreSnapshot[key] = configSnapshotFile{content: content, exists: exists}
	return content, exists
}

func (fsys *configSnapshotFS) snapshotKey(filePath string) string {
	key := tspath.NormalizePath(filePath)
	if !fsys.caseSensitive {
		key = strings.ToLower(key)
	}
	return key
}
