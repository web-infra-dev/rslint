//go:build !windows && !js

package main

import (
	"os"
	"testing"
)

func createWorkingDirectoryAlias(t *testing.T, target, alias string) {
	t.Helper()
	if err := os.Symlink(target, alias); err != nil {
		t.Fatalf("create working-directory symlink: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Remove(alias); err != nil && !os.IsNotExist(err) {
			t.Errorf("remove working-directory symlink: %v", err)
		}
	})
}
