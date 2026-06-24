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

// TestHandleLint_RuleOptionsWithSchemaDrivenRule guards against a regression
// where rules migrated to Schema0/Schema1 + RunWithOptions panicked with a
// nil-pointer dereference when invoked through the req.RuleOptions path
// (used by packages/rslint-test-tools rule-tester harnesses): that path
// called the now-nil legacy Run callback directly instead of dispatching to
// RunWithOptions with validated, default-hydrated options.
func TestHandleLint_RuleOptionsWithSchemaDrivenRule(t *testing.T) {
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
		FileContents: map[string]string{
			targetFile: "console.log('a'); console.error('b');\n",
		},
		RuleOptions: map[string]interface{}{
			"no-console": map[string]interface{}{
				"allow": []interface{}{"log"},
			},
		},
	})
	if err != nil {
		t.Fatalf("HandleLint returned error: %v", err)
	}

	if len(response.Diagnostics) != 1 {
		t.Fatalf("expected exactly one diagnostic (console.error, with console.log allowed), got %d: %+v", len(response.Diagnostics), response.Diagnostics)
	}
	if response.Diagnostics[0].RuleName != "no-console" {
		t.Fatalf("expected diagnostic from no-console, got %q", response.Diagnostics[0].RuleName)
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
