package main

import (
	"bytes"
	"go/format"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Every production Go file under shim must come from the pinned compiler and
// its manifests. Compare in memory so that running the test never edits them.
func TestGeneratedShims(t *testing.T) {
	t.Chdir("../..")
	generated := make(map[string]bool)
	generateShims(func(name string, source []byte) error {
		generated[filepath.FromSlash(name)] = true
		actual, err := os.ReadFile(name)
		if err != nil {
			t.Error(err)
			return nil
		}
		if !bytes.Equal(actual, source) {
			// Some checked-in shims have been run through gofmt.
			want, wantErr := format.Source(source)
			got, gotErr := format.Source(actual)
			if wantErr != nil || gotErr != nil || !bytes.Equal(want, got) {
				t.Errorf("%s differs from go run ./tools/gen_shims", name)
			}
		}
		return nil
	})
	err := filepath.WalkDir("shim", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go") && !generated[name] {
			t.Errorf("%s has no generator", name)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
