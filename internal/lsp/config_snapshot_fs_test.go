package lsp

import (
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/microsoft/typescript-go/shim/vfs"
)

type blockingSnapshotReadFS struct {
	vfs.FS
	started chan string

	mu       sync.Mutex
	releases map[string]chan struct{}
	reads    map[string]int
}

func (*blockingSnapshotReadFS) UseCaseSensitiveFileNames() bool { return true }

func (fsys *blockingSnapshotReadFS) ReadFile(filePath string) (string, bool) {
	fsys.mu.Lock()
	fsys.reads[filePath]++
	release := fsys.releases[filePath]
	fsys.mu.Unlock()
	fsys.started <- filePath
	<-release
	return filePath, true
}

func TestConfigSnapshotFSReadsDifferentGitignoresConcurrently(t *testing.T) {
	first := "/repo/one/.gitignore"
	second := "/repo/two/.gitignore"
	underlying := &blockingSnapshotReadFS{
		started: make(chan string, 4),
		releases: map[string]chan struct{}{
			first:  make(chan struct{}),
			second: make(chan struct{}),
		},
		reads: make(map[string]int),
	}
	fsys := newConfigSnapshotFS(underlying)

	firstDone := make(chan struct{})
	go func() {
		defer close(firstDone)
		fsys.ReadFile(first)
	}()
	if got := <-underlying.started; got != first {
		t.Fatalf("first read = %q, want %q", got, first)
	}

	secondDone := make(chan struct{})
	go func() {
		defer close(secondDone)
		fsys.ReadFile(second)
	}()
	select {
	case got := <-underlying.started:
		if got != second {
			t.Fatalf("second read = %q, want %q", got, second)
		}
	case <-time.After(time.Second):
		t.Fatal("second .gitignore read was serialized behind the first")
	}

	duplicateDone := make(chan struct{})
	go func() {
		defer close(duplicateDone)
		fsys.ReadFile(first)
	}()
	close(underlying.releases[first])
	<-firstDone
	<-duplicateDone
	close(underlying.releases[second])
	<-secondDone

	underlying.mu.Lock()
	defer underlying.mu.Unlock()
	if got := underlying.reads[first]; got != 1 {
		t.Fatalf("same source read %d times, want one snapshot", got)
	}
}

func TestConfigSnapshotFSPreservesPOSIXBackslashDirectories(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("backslash is a separator on Windows")
	}
	filePath := "/repo/literal\\directory/.gitignore"
	release := make(chan struct{})
	underlying := &blockingSnapshotReadFS{
		started: make(chan string, 1),
		releases: map[string]chan struct{}{
			filePath: release,
		},
		reads: make(map[string]int),
	}
	close(release)
	content, ok := newConfigSnapshotFS(underlying).ReadFile(filePath)
	if !ok || content != filePath {
		t.Fatalf("ReadFile() = (%q, %v), want exact POSIX path", content, ok)
	}
	if got := <-underlying.started; got != filePath {
		t.Fatalf("underlying read = %q, want %q", got, filePath)
	}
}
