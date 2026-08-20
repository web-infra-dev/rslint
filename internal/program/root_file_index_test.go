package program_test

import (
	"sync"
	"testing"

	"github.com/microsoft/typescript-go/shim/tspath"
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

func TestRootFileIndexUsesCanonicalFilesystemIdentity(t *testing.T) {
	const (
		root     = "/repo/Source/Index.ts"
		query    = "/REPO/source/index.ts"
		physical = "/repo/source/Index.ts"
	)
	base := &rootFileAliasFS{
		FS:            osvfs.FS(),
		realPaths:     map[string]string{root: physical, query: physical},
		realpathCalls: make(map[string]int),
	}
	index := lintprogram.NewRootFileIndex([]string{root}, &caseInsensitiveRootFS{base})
	if !index.Contains(query, "") {
		t.Fatal("paths with the same canonical identity were not matched")
	}
	if got := base.realpathCallCount(query); got != 1 {
		t.Fatalf("canonical target was resolved %d time(s), want one", got)
	}

	distinct := &rootFileAliasFS{
		FS:            osvfs.FS(),
		realPaths:     map[string]string{root: root, query: query},
		realpathCalls: make(map[string]int),
	}
	index = lintprogram.NewRootFileIndex([]string{root}, &caseInsensitiveRootFS{distinct})
	if index.Contains(query, "") {
		t.Fatal("case-folded paths with distinct canonical identities were merged")
	}
}

func TestRootFileIndexCanonicalIdentityAcrossPathForms(t *testing.T) {
	tests := []struct {
		name      string
		root      string
		query     string
		rootReal  string
		queryReal string
		want      bool
	}{
		{
			name:      "Windows drive casing",
			root:      `C:\Repo\Src\Index.ts`,
			query:     `c:\repo\src\index.ts`,
			rootReal:  "C:/Repo/Src/Index.ts",
			queryReal: "C:/Repo/Src/Index.ts",
			want:      true,
		},
		{
			name:      "UNC server and share casing",
			root:      `\\SERVER\SHARE\Repo\Index.ts`,
			query:     `\\server\share\repo\index.ts`,
			rootReal:  "//SERVER/SHARE/Repo/Index.ts",
			queryReal: "//SERVER/SHARE/Repo/Index.ts",
			want:      true,
		},
		{
			name:      "different Windows drives",
			root:      "C:/Repo/Index.ts",
			query:     "D:/Repo/Index.ts",
			rootReal:  "C:/Repo/Index.ts",
			queryReal: "D:/Repo/Index.ts",
		},
		{
			name:      "different UNC shares",
			root:      "//server/share-a/Repo/Index.ts",
			query:     "//server/share-b/Repo/Index.ts",
			rootReal:  "//server/share-a/Repo/Index.ts",
			queryReal: "//server/share-b/Repo/Index.ts",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fsys := &rootFileAliasFS{
				FS: osvfs.FS(),
				realPaths: map[string]string{
					tspath.NormalizePath(test.root):  test.rootReal,
					tspath.NormalizePath(test.query): test.queryReal,
				},
				realpathCalls: make(map[string]int),
			}
			index := lintprogram.NewRootFileIndex([]string{test.root}, fsys)
			if got := index.Contains(test.query, ""); got != test.want {
				t.Fatalf("Contains(%q) = %t, want %t", test.query, got, test.want)
			}
		})
	}
}
