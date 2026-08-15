package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func writeProgramTestFiles(t *testing.T, directory string, files map[string]string) {
	t.Helper()
	for relativePath, content := range files {
		fileName := filepath.Join(directory, relativePath)
		if err := os.MkdirAll(filepath.Dir(fileName), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(fileName), err)
		}
		if err := os.WriteFile(fileName, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", fileName, err)
		}
	}
}

func createTestProgram(t *testing.T, files map[string]string) *compiler.Program {
	t.Helper()
	directory := t.TempDir()
	writeProgramTestFiles(t, directory, files)
	if err := os.WriteFile(filepath.Join(directory, "tsconfig.json"), []byte(`{"include":["**/*.ts"]}`), 0o644); err != nil {
		t.Fatalf("write tsconfig: %v", err)
	}
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	program, err := utils.CreateProgram(
		true,
		fsys,
		directory,
		"tsconfig.json",
		utils.CreateCompilerHost(directory, fsys),
	)
	if err != nil {
		t.Fatalf("create Program: %v", err)
	}
	return program
}
