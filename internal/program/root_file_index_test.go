package program_test

import (
	"sync"
	"testing"

	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
)

type rootFileAliasFS struct {
	vfs.FS
	realPaths     map[string]string
	realpathMu    sync.Mutex
	realpathCalls map[string]int
}

type caseInsensitiveRootFS struct {
	*rootFileAliasFS
}

func (*caseInsensitiveRootFS) UseCaseSensitiveFileNames() bool {
	return false
}

func (fsys *rootFileAliasFS) Realpath(fileName string) string {
	fsys.realpathMu.Lock()
	fsys.realpathCalls[fileName]++
	fsys.realpathMu.Unlock()
	if realPath := fsys.realPaths[fileName]; realPath != "" {
		return realPath
	}
	return fileName
}

func (fsys *rootFileAliasFS) realpathCallCount(fileName string) int {
	fsys.realpathMu.Lock()
	defer fsys.realpathMu.Unlock()
	return fsys.realpathCalls[fileName]
}

func TestRootFileIndexUsesDirectAndPhysicalIdentity(t *testing.T) {
	const (
		rootAlias = "/repo-link/src/index.ts"
		target    = "/repo/src/index.ts"
		other     = "/repo/src/other.ts"
	)
	fsys := &rootFileAliasFS{
		FS:            osvfs.FS(),
		realPaths:     map[string]string{rootAlias: target},
		realpathCalls: make(map[string]int),
	}
	rootFileNames := []string{rootAlias}
	index := lintprogram.NewRootFileIndex(rootFileNames, fsys)
	rootFileNames[0] = other
	if !index.Contains(rootAlias, target) {
		t.Fatal("exact root was not recognized")
	}
	if !index.Contains(target, target) {
		t.Fatal("physical alias of a root was not recognized")
	}
	if got := fsys.realpathCallCount(target); got != 0 {
		t.Fatalf("already-canonical target was resolved %d extra time(s)", got)
	}
	if index.Contains(other, other) {
		t.Fatal("non-root file was recognized as a direct root")
	}

	// A shared parsed config may be inspected by independent config-owner
	// frontiers. Lock in concurrent read safety for the lazily built indexes.
	var workers sync.WaitGroup
	for range 8 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for range 100 {
				if !index.Contains(target, target) {
					t.Error("concurrent root lookup lost membership")
					return
				}
			}
		}()
	}
	workers.Wait()
}

func TestRootFileIndexUsesCanonicalIdentityInsteadOfGlobalCaseFlag(t *testing.T) {
	const (
		root     = "/repo/Source/Index.ts"
		query    = "/REPO/source/index.ts"
		physical = "/physical/repo/source/index.ts"
	)

	t.Run("distinct physical paths remain distinct", func(t *testing.T) {
		base := &rootFileAliasFS{
			FS:            osvfs.FS(),
			realPaths:     make(map[string]string),
			realpathCalls: make(map[string]int),
		}
		index := lintprogram.NewRootFileIndex([]string{root}, &caseInsensitiveRootFS{base})
		if index.Contains(query, query) {
			t.Fatal("global case behavior merged distinct canonical paths")
		}
	})

	t.Run("same physical path is recognized", func(t *testing.T) {
		base := &rootFileAliasFS{
			FS: osvfs.FS(),
			realPaths: map[string]string{
				root:  physical,
				query: physical,
			},
			realpathCalls: make(map[string]int),
		}
		index := lintprogram.NewRootFileIndex([]string{root}, &caseInsensitiveRootFS{base})
		if !index.Contains(query, "") {
			t.Fatal("paths with the same canonical identity were not matched")
		}
		if got := base.realpathCallCount(query); got != 1 {
			t.Fatalf("query was resolved %d time(s), want one", got)
		}
	})
}
