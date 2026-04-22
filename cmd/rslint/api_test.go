package main

import (
	"os"
	"path/filepath"
	"testing"

	api "github.com/web-infra-dev/rslint/internal/api"
)

func preserveWorkingDirectory(t *testing.T) {
	t.Helper()

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	// HandleLint mutates the process CWD when WorkingDirectory is set.
	// Register testing's built-in cleanup to restore the original directory.
	t.Chdir(oldWD)
}

func TestHandleLint_DefaultsToLintAllFiles(t *testing.T) {
	fixturesDir, err := filepath.Abs(filepath.Join("..", "..", "packages", "rslint", "fixtures"))
	if err != nil {
		t.Fatalf("resolve fixtures dir: %v", err)
	}
	configPath := filepath.Join(fixturesDir, "rslint.json")

	preserveWorkingDirectory(t)

	handler := &IPCHandler{}
	response, err := handler.HandleLint(api.LintRequest{
		Config:           configPath,
		WorkingDirectory: fixturesDir,
		RuleOptions: map[string]interface{}{
			"@typescript-eslint/no-unsafe-member-access": "error",
		},
	})
	if err != nil {
		t.Fatalf("HandleLint returned error: %v", err)
	}

	if response.FileCount <= 0 {
		t.Fatalf("expected lint to process at least one file, got %d", response.FileCount)
	}
	if len(response.Diagnostics) == 0 {
		t.Fatal("expected lint to report diagnostics when no explicit files filter is provided")
	}
}

func TestHandleLint_ExplicitFilesRestrictsScope(t *testing.T) {
	fixturesDir, err := filepath.Abs(filepath.Join("..", "..", "packages", "rslint", "fixtures"))
	if err != nil {
		t.Fatalf("resolve fixtures dir: %v", err)
	}
	configPath := filepath.Join(fixturesDir, "rslint.json")
	targetFile := filepath.Join(fixturesDir, "src", "index.ts")

	preserveWorkingDirectory(t)

	handler := &IPCHandler{}
	response, err := handler.HandleLint(api.LintRequest{
		Config:           configPath,
		Files:            []string{targetFile},
		WorkingDirectory: fixturesDir,
		RuleOptions: map[string]interface{}{
			"@typescript-eslint/no-unsafe-member-access": "error",
		},
	})
	if err != nil {
		t.Fatalf("HandleLint returned error: %v", err)
	}

	if response.FileCount != 1 {
		t.Fatalf("expected lint to process exactly one file, got %d", response.FileCount)
	}
	if len(response.Diagnostics) == 0 {
		t.Fatal("expected lint to report diagnostics for the explicitly requested file")
	}
	for _, diagnostic := range response.Diagnostics {
		if diagnostic.FilePath != "src/index.ts" {
			t.Fatalf("expected diagnostics to be limited to src/index.ts, got %q", diagnostic.FilePath)
		}
	}
}
