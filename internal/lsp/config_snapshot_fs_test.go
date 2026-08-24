package lsp

import (
	"fmt"
	"testing"
)

type caseInsensitiveConfigSnapshotFS struct {
	mockFS
	realpathCalls int
}

func (fsys *caseInsensitiveConfigSnapshotFS) UseCaseSensitiveFileNames() bool {
	return false
}

func (fsys *caseInsensitiveConfigSnapshotFS) Realpath(filePath string) string {
	fsys.realpathCalls++
	return fmt.Sprintf("%s#%d", filePath, fsys.realpathCalls)
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
