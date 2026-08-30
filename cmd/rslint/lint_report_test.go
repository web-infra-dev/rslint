package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/target"
	"github.com/web-infra-dev/rslint/internal/output"
)

type realpathSpyFS struct {
	vfs.FS
	calls []string
}

func (fsys *realpathSpyFS) Realpath(path string) string {
	fsys.calls = append(fsys.calls, path)
	return path
}

func TestLintReportOutcome(t *testing.T) {
	tests := []struct {
		name        string
		counts      output.Counts
		maxWarnings int
		want        output.Outcome
	}{
		{
			name:        "warnings below limit",
			counts:      output.Counts{Warnings: 1},
			maxWarnings: 2,
			want:        output.Outcome{Kind: output.OutcomePassed},
		},
		{
			name:        "warnings equal limit",
			counts:      output.Counts{Warnings: 2},
			maxWarnings: 2,
			want:        output.Outcome{Kind: output.OutcomePassed},
		},
		{
			name:        "warnings exceed limit",
			counts:      output.Counts{Warnings: 3},
			maxWarnings: 2,
			want: output.Outcome{
				Kind:         output.OutcomeWarningLimitExceeded,
				WarningLimit: 2,
			},
		},
		{
			name:        "disabled warning limit",
			counts:      output.Counts{Warnings: 3},
			maxWarnings: -1,
			want:        output.Outcome{Kind: output.OutcomePassed},
		},
		{
			name:        "diagnostics take priority over warning limit",
			counts:      output.Counts{Errors: 1, Warnings: 3},
			maxWarnings: 0,
			want:        output.Outcome{Kind: output.OutcomeDiagnosticsFailed},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := lintReportOutcome(test.counts, test.maxWarnings); got != test.want {
				t.Fatalf("lintReportOutcome(%+v, %d) = %+v, want %+v", test.counts, test.maxWarnings, got, test.want)
			}
		})
	}
}

func TestCLIWarningLimitStatusAndMachineCompatibility(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "index.js")
	if err := os.WriteFile(filePath, []byte("debugger;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	baseArgs := lintArgs{
		ConfigCatalog: explicitConfigCatalogForTest(dir, rslintconfig.RslintConfig{{
			Files: []string{"*.js"},
			Rules: rslintconfig.Rules{"no-debugger": "warn"},
		}}),
		MaxWarnings:    0,
		NoColor:        true,
		SingleThreaded: true,
		AllowFiles:     []string{filePath},
	}

	defaultArgs := baseArgs
	defaultArgs.Format = "default"
	code, stdout, stderr := runLintCommandForTest(t, dir, defaultArgs)
	if code != 1 {
		t.Fatalf("default warning-limit exit = %d, want 1; stdout=%q stderr=%q", code, stdout, stderr)
	}
	if !strings.HasPrefix(stdout, "start   Linting...\n\n") ||
		!strings.Contains(stdout, "error   Lint failed in ") ||
		!strings.Contains(stdout, "1 warning exceeded the configured limit of 0") {
		t.Fatalf("default warning-limit output = %q", stdout)
	}
	if strings.Contains(stdout, "success") || strings.Contains(stderr, "too many warnings") {
		t.Fatalf("default warning-limit produced conflicting output: stdout=%q stderr=%q", stdout, stderr)
	}

	machineArgs := baseArgs
	machineArgs.Format = "jsonline"
	machineArgs.Timing = true
	code, stdout, stderr = runLintCommandForTest(t, dir, machineArgs)
	if code != 1 {
		t.Fatalf("machine warning-limit exit = %d, want 1; stdout=%q stderr=%q", code, stdout, stderr)
	}
	if strings.Contains(stdout, "start   ") || strings.Contains(stdout, "Lint failed") || !strings.HasPrefix(stdout, "{") {
		t.Fatalf("machine stdout changed shape: %q", stdout)
	}
	legacy := "Rslint found too many warnings (maximum: 0)."
	if !strings.Contains(stderr, legacy) {
		t.Fatalf("machine warning-limit stderr lost legacy message: %q", stderr)
	}
	if table := strings.Index(stderr, "Rule "); table < strings.Index(stderr, legacy) {
		t.Fatalf("timing table was not emitted last: %q", stderr)
	}
}

func TestCLIEmptyDefaultCompletesButMachineFormatStaysSilent(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "ignored.js")
	if err := os.WriteFile(filePath, []byte("debugger;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	baseArgs := lintArgs{
		ConfigCatalog: explicitConfigCatalogForTest(dir, rslintconfig.RslintConfig{
			{Ignores: []string{"ignored.js"}},
			{Rules: rslintconfig.Rules{"no-debugger": "error"}},
		}),
		NoColor:        true,
		SingleThreaded: true,
		AllowFiles:     []string{filePath},
	}

	defaultArgs := baseArgs
	defaultArgs.Format = "default"
	code, stdout, _ := runLintCommandForTest(t, dir, defaultArgs)
	if code != 0 || !strings.HasPrefix(stdout, "start   Linting...\n") ||
		!strings.Contains(stdout, "success No files to lint in ") {
		t.Fatalf("default empty result: code=%d stdout=%q", code, stdout)
	}

	machineArgs := baseArgs
	machineArgs.Format = "jsonline"
	code, stdout, _ = runLintCommandForTest(t, dir, machineArgs)
	if code != 0 || stdout != "" {
		t.Fatalf("machine empty result changed: code=%d stdout=%q", code, stdout)
	}
}

func TestLintReportFileCountUsesCanonicalUnion(t *testing.T) {
	targets := []target.File{
		{PathIdentity: rslintconfig.PathIdentity{Path: "/repo/a.ts", CanonicalPath: "/physical/a.ts"}},
		{PathIdentity: rslintconfig.PathIdentity{Path: "/repo/b.ts", CanonicalPath: "/physical/b.ts"}},
		// A second lexical spelling of the same physical target must not count.
		{PathIdentity: rslintconfig.PathIdentity{Path: "/alias/a.ts", CanonicalPath: "/physical/a.ts"}},
	}
	roots := []string{"/physical/b.ts", "/physical/c.ts", "/physical/c.ts"}

	tests := []struct {
		name string
		mode output.Mode
		want int
	}{
		{name: "lint targets", mode: output.ModeLint, want: 2},
		{name: "type-check roots", mode: output.ModeTypeCheckOnly, want: 2},
		{name: "combined union", mode: output.ModeLintAndTypeCheck, want: 3},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fsys := &realpathSpyFS{}
			if got := lintReportFileCount(test.mode, targets, roots, fsys); got != test.want {
				t.Fatalf("lintReportFileCount() = %d, want %d", got, test.want)
			}
			wantRealpathCalls := 0
			if test.mode != output.ModeLint {
				wantRealpathCalls = len(roots)
			}
			if len(fsys.calls) != wantRealpathCalls {
				t.Fatalf("Realpath calls = %v, want %d type-check root calls and no frozen-target calls", fsys.calls, wantRealpathCalls)
			}
		})
	}
}
