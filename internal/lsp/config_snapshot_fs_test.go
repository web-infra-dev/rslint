package lsp

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/utils/pathutil"
	"testing"
)

type caseInsensitiveConfigSnapshotFS struct {
	mockFS
	realpathCalls int
}

func (fsys *caseInsensitiveConfigSnapshotFS) CaseSensitivity() tspath.CaseSensitivity {
	return pathutil.CaseSensitivity(false)
}

func (fsys *caseInsensitiveConfigSnapshotFS) Realpath(filePath tspath.RootedPath) tspath.RootedPath {
	fsys.realpathCalls++
	return pathutil.Path(fmt.Sprintf("%s#%d", filePath, fsys.realpathCalls))
}

func TestConfigSnapshotFSCaseInsensitiveRealpathIdentity(t *testing.T) {
	underlying := &caseInsensitiveConfigSnapshotFS{}
	fsys := newConfigSnapshotFS(underlying)

	first := fsys.Realpath("/Project/Config")
	second := fsys.Realpath("/project/config")
	if first != second {
		t.Fatalf("alternate casing observed two snapshots: first %q, second %q", first, second)
	}
	if underlying.realpathCalls != 1 {
		t.Fatalf("Realpath calls = %d, want one case-insensitive observation", underlying.realpathCalls)
	}
}
