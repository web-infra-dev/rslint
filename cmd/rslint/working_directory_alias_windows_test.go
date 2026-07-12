//go:build windows && !js

package main

// cspell:ignore mklink

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
)

func createWorkingDirectoryAlias(t *testing.T, target, alias string) {
	t.Helper()
	// Directory symlinks can require Developer Mode or elevated privileges on
	// Windows. A directory junction exercises the same path-identity boundary and
	// is available to an ordinary CI user.
	commandLine := fmt.Sprintf(`mklink /J "%s" "%s"`, alias, target)
	if output, err := exec.Command("cmd.exe", "/d", "/c", commandLine).CombinedOutput(); err != nil {
		t.Fatalf("create working-directory junction: %v\n%s", err, output)
	}
	t.Cleanup(func() {
		if err := os.Remove(alias); err != nil && !os.IsNotExist(err) {
			t.Errorf("remove working-directory junction: %v", err)
		}
	})
}
