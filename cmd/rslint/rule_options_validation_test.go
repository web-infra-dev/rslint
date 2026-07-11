package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The rule-options validation step runs after configuration is fully
// resolved and before any linting starts: a config with schema-invalid rule
// options must fail fast, report every failure at once, and never produce
// lint diagnostics.
func TestCLIInvalidRuleOptionsFailFastBeforeLinting(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "rslint.jsonc"), []byte(`[
		{
			"files": ["*.js"],
			"rules": {
				"no-console": ["error", { "allow": "warn" }],
				"eqeqeq": ["error", "sometimes"],
				"no-debugger": "error"
			}
		}
	]`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "app.js"), []byte("debugger;\n"), 0o644); err != nil {
		t.Fatalf("write app.js: %v", err)
	}

	code, stdout, stderr := runLintPipelineForTest(t, dir, lintArgs{
		Config:         "rslint.jsonc",
		Format:         "default",
		NoColor:        true,
		SingleThreaded: true,
	})
	if code != 1 {
		t.Fatalf("expected exit code 1 for invalid rule options, got code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	// Every failure is reported, not just the first.
	if !strings.Contains(stderr, `invalid options for rule "no-console"`) {
		t.Errorf("expected stderr to name no-console, got %q", stderr)
	}
	if !strings.Contains(stderr, `invalid options for rule "eqeqeq"`) {
		t.Errorf("expected stderr to name eqeqeq, got %q", stderr)
	}
	// Linting never started: the valid no-debugger rule produced no diagnostic.
	if strings.Contains(stdout, "no-debugger") {
		t.Errorf("expected no lint diagnostics before validation passes, stdout=%q", stdout)
	}
}

// The same config with schema-valid options must sail through the validation
// step and lint normally.
func TestCLIValidRuleOptionsLintNormally(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "rslint.jsonc"), []byte(`[
		{
			"files": ["*.js"],
			"rules": {
				"no-console": ["error", { "allow": ["warn"] }],
				"eqeqeq": ["error", "smart"],
				"no-debugger": "error"
			}
		}
	]`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "app.js"), []byte("debugger;\n"), 0o644); err != nil {
		t.Fatalf("write app.js: %v", err)
	}

	code, stdout, stderr := runLintPipelineForTest(t, dir, lintArgs{
		Config:         "rslint.jsonc",
		Format:         "default",
		NoColor:        true,
		SingleThreaded: true,
	})
	if strings.Contains(stderr, "invalid options") {
		t.Fatalf("expected schema validation to pass, stderr=%q", stderr)
	}
	if code != 1 || !strings.Contains(stdout, "no-debugger") {
		t.Fatalf("expected the lint itself to run and flag no-debugger, got code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
}
