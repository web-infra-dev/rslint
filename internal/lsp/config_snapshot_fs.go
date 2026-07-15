package lsp

import (
	"sync"

	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/hostpath"
)

// configSnapshotFS gives one config-discovery transaction a stable view of
// every .gitignore file it observes. cachedvfs deliberately does not cache
// ReadFile, while a committed JS evaluator may revisit one source for later
// exact files. Caching those bytes keeps the generation internally consistent;
// the evaluator retains this filesystem until a later transaction replaces it.
type configSnapshotFS struct {
	vfs.FS

	mu                sync.Mutex
	gitignoreSnapshot map[string]*configSnapshotState
	caseSensitive     bool
}

type configSnapshotFile struct {
	content string
	exists  bool
}

type configSnapshotState struct {
	ready chan struct{}
	file  configSnapshotFile
}

func newConfigSnapshotFS(fsys vfs.FS) *configSnapshotFS {
	caseSensitive := true
	if fsys != nil {
		caseSensitive = fsys.UseCaseSensitiveFileNames()
	}
	return &configSnapshotFS{
		FS:                fsys,
		gitignoreSnapshot: make(map[string]*configSnapshotState),
		caseSensitive:     caseSensitive,
	}
}

func (fsys *configSnapshotFS) ReadFile(filePath string) (string, bool) {
	if fsys == nil || fsys.FS == nil {
		return "", false
	}
	filePath = hostpath.NormalizeForRoot(filePath, filePath)
	if hostpath.BaseForRoot(filePath, filePath) != ".gitignore" {
		return fsys.FS.ReadFile(filePath)
	}

	key := hostpath.Identity(filePath, hostpath.DirectoryForRoot(filePath, filePath), fsys.caseSensitive)
	fsys.mu.Lock()
	if state, ok := fsys.gitignoreSnapshot[key]; ok {
		fsys.mu.Unlock()
		<-state.ready
		return state.file.content, state.file.exists
	}
	state := &configSnapshotState{ready: make(chan struct{})}
	fsys.gitignoreSnapshot[key] = state
	fsys.mu.Unlock()

	// Different .gitignore sources are independent and may be read in
	// parallel. Only readers of this exact source wait for its first snapshot.
	content, exists := fsys.FS.ReadFile(filePath)
	state.file = configSnapshotFile{content: content, exists: exists}
	close(state.ready)
	return content, exists
}
