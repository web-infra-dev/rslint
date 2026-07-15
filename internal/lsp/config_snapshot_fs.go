package lsp

import (
	"strings"
	"sync"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
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
		caseSensitive:     caseSensitive,
	}
}

func (fsys *configSnapshotFS) ReadFile(filePath string) (string, bool) {
	if fsys == nil || fsys.FS == nil {
		return "", false
	}
	filePath = tspath.NormalizePath(filePath)
	if tspath.GetBaseFileName(filePath) != ".gitignore" {
		return fsys.FS.ReadFile(filePath)
	}

	key := filePath
	if !fsys.caseSensitive {
		key = strings.ToLower(key)
	}
	fsys.mu.Lock()
	defer fsys.mu.Unlock()
	if file, ok := fsys.gitignoreSnapshot[key]; ok {
		return file.content, file.exists
	}
	content, exists := fsys.FS.ReadFile(filePath)
	fsys.gitignoreSnapshot[key] = configSnapshotFile{content: content, exists: exists}
	return content, exists
}
