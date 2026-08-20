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

func TestRootFileIndexUsesAuthoritativeIdentityForCaseAliases(t *testing.T) {
	const (
		root  = "/repo/Source/Index.ts"
		query = "/REPO/source/index.ts"
	)

	t.Run("same physical file", func(t *testing.T) {
		const physical = "/physical/Source/Index.ts"
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
			t.Fatal("filesystem-proven case alias was not recognized")
		}
		if got := base.realpathCallCount(query); got != 1 {
			t.Fatalf("query Realpath calls = %d, want 1", got)
		}
	})

	t.Run("distinct physical files", func(t *testing.T) {
		base := &rootFileAliasFS{
			FS: osvfs.FS(),
			realPaths: map[string]string{
				root:  "/physical/Source/Index.ts",
				query: "/physical/source/index.ts",
			},
			realpathCalls: make(map[string]int),
		}
		index := lintprogram.NewRootFileIndex([]string{root}, &caseInsensitiveRootFS{base})
		if index.Contains(query, "") {
			t.Fatal("case-folded but physically distinct file was recognized")
		}
	})
}

func TestRootFileIndexesResolvePhysicalRootsAsOneLazyBatch(t *testing.T) {
	const (
		firstRoot      = "/repo/first.ts"
		secondRoot     = "/repo/second.ts"
		firstPhysical  = "/physical/first.ts"
		secondPhysical = "/physical/second.ts"
	)
	fsys := &rootFileAliasFS{
		FS: osvfs.FS(),
		realPaths: map[string]string{
			firstRoot:  firstPhysical,
			secondRoot: secondPhysical,
		},
		realpathCalls: make(map[string]int),
	}
	resolver := lintprogram.NewPathIdentityResolver(fsys, false, nil)
	indexes := lintprogram.NewRootFileIndexes(
		[][]string{{firstRoot}, {secondRoot}},
		resolver,
	)

	if !indexes[0].Contains(firstRoot, firstPhysical) {
		t.Fatal("exact root was not recognized")
	}
	if got := fsys.realpathCallCount(firstRoot) + fsys.realpathCallCount(secondRoot); got != 0 {
		t.Fatalf("exact lookup resolved %d physical root(s), want none", got)
	}
	if !indexes[0].Contains(firstPhysical, firstPhysical) {
		t.Fatal("physical alias of the first root was not recognized")
	}
	if got := fsys.realpathCallCount(firstRoot); got != 1 {
		t.Fatalf("first root Realpath calls = %d, want 1", got)
	}
	if got := fsys.realpathCallCount(secondRoot); got != 1 {
		t.Fatalf("second root Realpath calls = %d, want shared batch to resolve it once", got)
	}
	if !indexes[1].Contains(secondPhysical, secondPhysical) {
		t.Fatal("physical alias of the second root was not recognized")
	}
	if got := fsys.realpathCallCount(secondRoot); got != 1 {
		t.Fatalf("second root was resolved %d times after the shared batch", got)
	}
}
