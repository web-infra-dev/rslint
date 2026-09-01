package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/target"
	"github.com/web-infra-dev/rslint/internal/output"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type gatedStartWriter struct {
	entered chan<- string
	release <-chan struct{}
}

type cancelAfterWriteWriter struct {
	dst    *bytes.Buffer
	cancel context.CancelFunc
}

func (w cancelAfterWriteWriter) Write(p []byte) (int, error) {
	n, err := w.dst.Write(p)
	w.cancel()
	return n, err
}

func (w gatedStartWriter) Write(p []byte) (int, error) {
	w.entered <- string(p)
	<-w.release
	return len(p), nil
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

func TestAssembleLintReportProjectsCountsAndOutcomeOnce(t *testing.T) {
	errorDiagnostic, comparePaths := createTestDiagnostic(t, "const value = 1;\n", 6, 11)
	errorDiagnostic.Origin = rule.DiagnosticOriginTypeScript
	warningDiagnostic := errorDiagnostic
	warningDiagnostic.RuleName = "warning-rule"
	warningDiagnostic.Message.Description = "warning message"
	warningDiagnostic.Severity = rule.SeverityWarning
	warningDiagnostic.Origin = rule.DiagnosticOriginLint

	report, err := assembleLintReport(lintReportInput{
		Mode:        output.ModeLintAndTypeCheck,
		Diagnostics: []rule.RuleDiagnostic{errorDiagnostic, warningDiagnostic},
		MaxWarnings: 0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := report.Counts(), (output.Counts{
		Errors: 1, Warnings: 1, TypeErrors: 1,
	}); got != want {
		t.Fatalf("counts = %+v, want %+v", got, want)
	}
	if got := report.Outcome(); got.Kind != output.OutcomeDiagnosticsFailed {
		t.Fatalf("outcome = %+v, want diagnostic failure precedence", got)
	}

	var rendered bytes.Buffer
	if err := output.Render(&rendered, report, output.Options{
		Format:       output.FormatJSONLine,
		ComparePaths: comparePaths,
	}); err != nil {
		t.Fatal(err)
	}
	got := rendered.String()
	if !strings.Contains(got, `"message":"Test diagnostic"`) ||
		!strings.Contains(got, `"message":"warning message"`) ||
		!strings.Contains(got, `"line":1`) {
		t.Fatalf("projected machine report = %q", got)
	}
}

func TestAssembleLintReportRejectsIncompleteDomainDiagnostic(t *testing.T) {
	_, err := assembleLintReport(lintReportInput{
		Mode: output.ModeLint,
		Diagnostics: []rule.RuleDiagnostic{{
			RuleName: "broken",
			FilePath: "/repo/broken.ts",
			Severity: rule.SeverityError,
		}},
	})
	if err == nil || !strings.Contains(err.Error(), "has no source file") {
		t.Fatalf("assemble error = %v, want missing source rejection", err)
	}
}

func TestAssembleLintReportRejectsUnknownDiagnosticClassifications(t *testing.T) {
	diagnostic, _ := createTestDiagnostic(t, "const value = 1;\n", 6, 11)
	for _, test := range []struct {
		name   string
		mutate func(*rule.RuleDiagnostic)
		want   string
	}{
		{
			name: "severity",
			mutate: func(diagnostic *rule.RuleDiagnostic) {
				diagnostic.Severity = rule.DiagnosticSeverity(255)
			},
			want: "invalid diagnostic severity 255",
		},
		{
			name: "origin",
			mutate: func(diagnostic *rule.RuleDiagnostic) {
				diagnostic.Origin = rule.DiagnosticOrigin(255)
			},
			want: "invalid diagnostic origin 255",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			invalid := diagnostic
			test.mutate(&invalid)
			_, err := assembleLintReport(lintReportInput{
				Mode:        output.ModeLint,
				Diagnostics: []rule.RuleDiagnostic{invalid},
			})
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("assemble error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestRenderLintReportStopsBeforeCompletedStatusAfterCancellation(t *testing.T) {
	t.Run("canceled before report", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		var dst bytes.Buffer
		report, reportErr := assembleLintReport(lintReportInput{
			Mode:    output.ModeLint,
			Summary: &output.Summary{},
		})
		if reportErr != nil {
			t.Fatal(reportErr)
		}
		err := renderLintReport(
			ctx,
			&dst,
			report,
			output.Options{Format: output.FormatDefault},
		)
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("render error = %v, want context.Canceled", err)
		}
		if dst.Len() != 0 {
			t.Fatalf("canceled report wrote output: %q", dst.String())
		}
	})

	t.Run("canceled after diagnostics flush", func(t *testing.T) {
		diagnostic, comparePaths := createTestDiagnostic(t, "debugger;\n", 0, 8)
		report, reportErr := assembleLintReport(lintReportInput{
			Mode:        output.ModeLint,
			Diagnostics: []rule.RuleDiagnostic{diagnostic},
			Summary: &output.Summary{
				Files:     1,
				Rules:     1,
				Threads:   1,
				StartedAt: time.Now(),
			},
			MaxWarnings:   -1,
			IncludeSource: true,
		})
		if reportErr != nil {
			t.Fatal(reportErr)
		}
		ctx, cancel := context.WithCancel(context.Background())
		var dst bytes.Buffer
		err := renderLintReport(
			ctx,
			cancelAfterWriteWriter{dst: &dst, cancel: cancel},
			report,
			output.Options{
				Format:       output.FormatDefault,
				ComparePaths: comparePaths,
			},
		)
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("render error = %v, want context.Canceled", err)
		}
		got := dst.String()
		if !strings.Contains(got, "test-rule") {
			t.Fatalf("test did not reach the diagnostics flush: %q", got)
		}
		if strings.Contains(got, "Lint failed") || strings.Contains(got, "success") {
			t.Fatalf("canceled report committed a completed status: %q", got)
		}
	})
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

func TestTypeCheckOnlyTimingDoesNotEmitRuleTable(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(
		filepath.Join(dir, "index.ts"),
		[]byte("export const value: number = 1;\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(dir, "tsconfig.json"),
		[]byte(`{"files":["index.ts"]}`),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	code, stdout, stderr := runLintCommandForTest(t, dir, lintArgs{
		ConfigCatalog: explicitConfigCatalogForTest(dir, rslintconfig.RslintConfig{{
			LanguageOptions: &rslintconfig.LanguageOptions{
				ParserOptions: &rslintconfig.ParserOptions{
					Project: rslintconfig.ProjectPaths{"./tsconfig.json"},
				},
			},
		}}),
		TypeCheck:      true,
		TypeCheckOnly:  true,
		Timing:         true,
		Format:         "default",
		NoColor:        true,
		SingleThreaded: true,
	})
	if code != 0 {
		t.Fatalf("type-check-only exit = %d, stdout=%q stderr=%q", code, stdout, stderr)
	}
	if !strings.HasPrefix(stdout, "start   Type checking...\n") ||
		!strings.Contains(stdout, "success Type check passed in ") ||
		!strings.Contains(stdout, "(1 file, 1 thread)") ||
		strings.Contains(stdout, "rule") {
		t.Fatalf("type-check-only lifecycle output = %q", stdout)
	}
	if stderr != "" {
		t.Fatalf("type-check-only --timing emitted a rule table: %q", stderr)
	}
}

func TestDefaultStartWriterCompletesBeforePostStartWarnings(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "ignored.js")
	if err := os.WriteFile(filePath, []byte("debugger;\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	stdoutFile, err := os.CreateTemp(t.TempDir(), "stdout-*")
	if err != nil {
		t.Fatal(err)
	}
	defer stdoutFile.Close()
	stderrFile, err := os.CreateTemp(t.TempDir(), "stderr-*")
	if err != nil {
		t.Fatal(err)
	}
	defer stderrFile.Close()

	t.Chdir(dir)
	originalStdout, originalStderr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = stdoutFile, stderrFile
	defer func() { os.Stdout, os.Stderr = originalStdout, originalStderr }()

	startEntered := make(chan string, 1)
	releaseStart := make(chan struct{})
	released := false
	defer func() {
		if !released {
			close(releaseStart)
		}
	}()
	codeCh := make(chan int, 1)
	go func() {
		codeCh <- handleLintCommand(lintArgs{
			ConfigCatalog: explicitConfigCatalogForTest(dir, rslintconfig.RslintConfig{
				{Ignores: []string{"ignored.js"}},
				{Rules: rslintconfig.Rules{"no-debugger": "error"}},
			}),
			Format:         "default",
			NoColor:        true,
			SingleThreaded: true,
			AllowFiles:     []string{filePath},
			StartWriter: gatedStartWriter{
				entered: startEntered,
				release: releaseStart,
			},
		}, context.Background(), nil)
	}()

	select {
	case start := <-startEntered:
		if start != "start   Linting...\n" {
			t.Fatalf("start output = %q", start)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("start writer was not called")
	}
	if err := stderrFile.Sync(); err != nil {
		t.Fatal(err)
	}
	stderrInfo, err := stderrFile.Stat()
	if err != nil {
		t.Fatal(err)
	}
	if stderrInfo.Size() != 0 {
		t.Fatalf("post-start warning was written before start acknowledgement; stderr bytes=%d", stderrInfo.Size())
	}

	close(releaseStart)
	released = true
	var code int
	select {
	case code = <-codeCh:
	case <-time.After(5 * time.Second):
		t.Fatal("lint command did not complete after start acknowledgement")
	}
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if err := stderrFile.Sync(); err != nil {
		t.Fatal(err)
	}
	stderr, err := os.ReadFile(stderrFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(stderr), "warning: ignored.js is ignored because of a matching ignore pattern") {
		t.Fatalf("post-start warning missing after acknowledgement: %q", stderr)
	}
	if err := stdoutFile.Sync(); err != nil {
		t.Fatal(err)
	}
	stdout, err := os.ReadFile(stdoutFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(stdout), "success No files to lint in ") {
		t.Fatalf("terminal status missing after acknowledgement: %q", stdout)
	}
}

func TestLintReportFileCountUsesCanonicalUnion(t *testing.T) {
	targets := []target.File{
		{PathIdentity: rslintconfig.PathIdentity{Path: "/repo/a.ts", CanonicalPath: "/physical/a.ts"}},
		{PathIdentity: rslintconfig.PathIdentity{Path: "/repo/b.ts", CanonicalPath: "/physical/b.ts"}},
		// A second lexical spelling of the same physical target must not count.
		{PathIdentity: rslintconfig.PathIdentity{Path: "/alias/a.ts", CanonicalPath: "/physical/a.ts"}},
	}
	roots := []string{
		rslintconfig.ExactPathID("/physical/b.ts"),
		rslintconfig.ExactPathID("/physical/c.ts"),
		rslintconfig.ExactPathID("/physical/c.ts"),
	}

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
			if got := lintReportFileCount(test.mode, targets, roots); got != test.want {
				t.Fatalf("lintReportFileCount() = %d, want %d", got, test.want)
			}
		})
	}
}
