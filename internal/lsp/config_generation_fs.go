package lsp

import (
	"strings"
	"sync"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
)

// configGenerationFS gives one config-discovery transaction a stable view of
// every .gitignore file it observes. cachedvfs deliberately does not cache
// ReadFile, so retaining only the discovery filesystem would otherwise let a
// rejected refresh change target admission underneath the committed catalog.
//
// Snapshot preparation collects every config-scoped .gitignore source that can
// govern the generation. Once frozen, an unobserved .gitignore is treated as
// absent: this prevents a file created after prepare from leaking into the
// committed generation.
type configGenerationFS struct {
	vfs.FS

	mu                sync.Mutex
	gitignoreSnapshot map[string]configGenerationFile
	frozen            bool
	caseSensitive     bool
}

type configGenerationFile struct {
	content string
	exists  bool
}

func newConfigGenerationFS(fsys vfs.FS) *configGenerationFS {
	caseSensitive := true
	if fsys != nil {
		caseSensitive = fsys.UseCaseSensitiveFileNames()
	}
	return &configGenerationFS{
		FS:                fsys,
		gitignoreSnapshot: make(map[string]configGenerationFile),
		caseSensitive:     caseSensitive,
	}
}

func (fsys *configGenerationFS) ReadFile(filePath string) (string, bool) {
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
	if fsys.frozen {
		return "", false
	}
	content, exists := fsys.FS.ReadFile(filePath)
	fsys.gitignoreSnapshot[key] = configGenerationFile{content: content, exists: exists}
	return content, exists
}

func (fsys *configGenerationFS) freeze() {
	if fsys == nil {
		return
	}
	fsys.mu.Lock()
	fsys.frozen = true
	fsys.mu.Unlock()
}
