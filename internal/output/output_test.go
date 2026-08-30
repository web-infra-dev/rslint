package output

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestParseFormat(t *testing.T) {
	tests := []struct {
		value string
		want  Format
	}{
		{"default", FormatDefault},
		{"jsonline", FormatJSONLine},
		{"github", FormatGitHub},
		{"gitlab", FormatGitLab},
	}
	for _, test := range tests {
		got, err := ParseFormat(test.value)
		if err != nil || got != test.want {
			t.Errorf("ParseFormat(%q) = %v, %v; want %v", test.value, got, err, test.want)
		}
	}
	if _, err := ParseFormat("stylish"); err == nil {
		t.Fatal("expected invalid format to fail")
	}
	if _, err := ParseFormat(""); err == nil {
		t.Fatal("expected an explicitly empty format to fail")
	}
}

func TestNewReportOwnsDiagnosticSnapshot(t *testing.T) {
	diagnostics := []rule.RuleDiagnostic{{RuleName: "before", Severity: rule.SeverityError}}
	report := NewReport(diagnostics, Metadata{Mode: ModeLint})
	diagnostics[0].Severity = rule.SeverityWarning
	diagnostics[0].RuleName = "after"
	if report.diagnostics[0].RuleName != "before" || report.Counts().Errors != 1 {
		t.Fatalf("report changed after caller mutation: diagnostic=%+v counts=%+v", report.diagnostics[0], report.Counts())
	}
}

func TestNewReportCounts(t *testing.T) {
	diagnostics := []rule.RuleDiagnostic{
		{RuleName: "no-debugger", Severity: rule.SeverityError},
		{RuleName: "TypeScript(TS2322)", Severity: rule.SeverityError, Origin: rule.DiagnosticOriginTypeScript},
		{RuleName: "TypeScript(TS9999)", Severity: rule.SeverityError},
		{RuleName: "no-console", Severity: rule.SeverityWarning},
		{RuleName: "off", Severity: rule.SeverityOff},
	}

	lintCounts := NewReport(diagnostics, Metadata{Mode: ModeLint}).Counts()
	if lintCounts != (Counts{Errors: 3, Warnings: 1, LintErrors: 3}) {
		t.Fatalf("lint counts = %+v", lintCounts)
	}

	typeCheckCounts := NewReport(diagnostics, Metadata{Mode: ModeLintAndTypeCheck}).Counts()
	if typeCheckCounts != (Counts{Errors: 3, Warnings: 1, LintErrors: 2, TypeErrors: 1}) {
		t.Fatalf("type-check counts = %+v", typeCheckCounts)
	}
}

func TestSummaryText(t *testing.T) {
	tests := []struct {
		name     string
		report   Report
		outcome  Outcome
		expected string
	}{
		{
			name: "lint passed",
			report: NewReport(nil, Metadata{
				Mode: ModeLint, Files: 2, Rules: 3, Threads: 4,
			}),
			outcome:  Outcome{Kind: OutcomePassed},
			expected: "success Lint passed in 12ms (2 files, 3 rules, 4 threads)\n",
		},
		{
			name: "lint passed with warning and fix",
			report: NewReport([]rule.RuleDiagnostic{
				{Severity: rule.SeverityWarning},
			}, Metadata{
				Mode: ModeLint, Files: 1, Rules: 1, Threads: 1, FixedIssues: 1,
			}),
			outcome:  Outcome{Kind: OutcomePassed},
			expected: "success Lint passed with 1 warning after applying 1 fix in 12ms (1 file, 1 rule, 1 thread)\n",
		},
		{
			name: "lint failed with fix",
			report: NewReport([]rule.RuleDiagnostic{
				{Severity: rule.SeverityError},
				{Severity: rule.SeverityWarning},
			}, Metadata{
				Mode: ModeLint, Files: 5, Rules: 7, Threads: 2, FixedIssues: 3,
			}),
			outcome:  Outcome{Kind: OutcomeDiagnosticsFailed},
			expected: "error   Lint failed with 1 error and 1 warning after applying 3 fixes in 12ms (5 files, 7 rules, 2 threads)\n",
		},
		{
			name: "lint and type check failed",
			report: NewReport([]rule.RuleDiagnostic{
				{Severity: rule.SeverityError},
				{Severity: rule.SeverityError},
				{Severity: rule.SeverityError, Origin: rule.DiagnosticOriginTypeScript},
				{Severity: rule.SeverityWarning},
			}, Metadata{
				Mode: ModeLintAndTypeCheck, Files: 9, Rules: 8, Threads: 2,
			}),
			outcome:  Outcome{Kind: OutcomeDiagnosticsFailed},
			expected: "error   Lint and type check failed with 2 lint errors, 1 TypeScript error, and 1 warning in 12ms (9 files, 8 rules, 2 threads)\n",
		},
		{
			name: "type check only failed",
			report: NewReport([]rule.RuleDiagnostic{
				{Severity: rule.SeverityError, Origin: rule.DiagnosticOriginTypeScript},
			}, Metadata{
				Mode: ModeTypeCheckOnly, Files: 1, Threads: 2,
			}),
			outcome:  Outcome{Kind: OutcomeDiagnosticsFailed},
			expected: "error   Type check failed with 1 TypeScript error in 12ms (1 file, 2 threads)\n",
		},
		{
			name: "warning limit exceeded",
			report: NewReport([]rule.RuleDiagnostic{
				{Severity: rule.SeverityWarning},
				{Severity: rule.SeverityWarning},
			}, Metadata{
				Mode: ModeLint, Files: 2, Rules: 3, Threads: 4,
			}),
			outcome: Outcome{Kind: OutcomeWarningLimitExceeded, WarningLimit: 0},
			expected: "error   Lint failed in 12ms: 2 warnings exceeded the configured limit of 0 " +
				"(2 files, 3 rules, 4 threads)\n",
		},
		{
			name:     "empty lint",
			report:   NewReport(nil, Metadata{Mode: ModeLint, Threads: 4}),
			outcome:  Outcome{Kind: OutcomePassed},
			expected: "success No files to lint in 12ms\n",
		},
		{
			name:     "empty combined",
			report:   NewReport(nil, Metadata{Mode: ModeLintAndTypeCheck, Threads: 4}),
			outcome:  Outcome{Kind: OutcomePassed},
			expected: "success No files to lint or type check in 12ms\n",
		},
		{
			name:     "empty type check",
			report:   NewReport(nil, Metadata{Mode: ModeTypeCheckOnly, Threads: 4}),
			outcome:  Outcome{Kind: OutcomePassed},
			expected: "success No files to type check in 12ms\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			w := bufio.NewWriter(&buf)
			renderSummary(w, test.report, test.outcome, 12*time.Millisecond, newColorScheme(false))
			if err := w.Flush(); err != nil {
				t.Fatal(err)
			}
			if got := buf.String(); got != test.expected {
				t.Fatalf("summary:\n got: %q\nwant: %q", got, test.expected)
			}
		})
	}
}

func TestSummaryDetailsAreOneDimSpan(t *testing.T) {
	tests := []struct {
		name    string
		report  Report
		details string
	}{
		{
			name: "lint details",
			report: NewReport(nil, Metadata{
				Mode: ModeLint, Files: 2, Rules: 3, Threads: 4,
			}),
			details: "(2 files, 3 rules, 4 threads)",
		},
		{
			name: "lint and type-check details",
			report: NewReport(nil, Metadata{
				Mode: ModeLintAndTypeCheck, Files: 2, Rules: 3, Threads: 1,
			}),
			details: "(2 files, 3 rules, 1 thread)",
		},
		{
			name: "type-check-only details",
			report: NewReport(nil, Metadata{
				Mode: ModeTypeCheckOnly, Files: 2, Threads: 4,
			}),
			details: "(2 files, 4 threads)",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			w := bufio.NewWriter(&buf)
			renderSummary(w, test.report, Outcome{Kind: OutcomePassed}, 12*time.Millisecond, newColorScheme(true))
			if err := w.Flush(); err != nil {
				t.Fatal(err)
			}
			wantSpan := "\x1b[2m" + test.details + "\x1b[22m"
			if !strings.Contains(buf.String(), wantSpan) {
				t.Fatalf("details are not one dim span:\n%s", buf.String())
			}
		})
	}
}

func TestLifecycleColorsOnlyLabelsAndDetails(t *testing.T) {
	colors := newColorScheme(true)
	options := Options{Format: FormatDefault, ColorEnabled: true}
	var buf bytes.Buffer
	if err := RenderStart(&buf, ModeLint, options); err != nil {
		t.Fatal(err)
	}
	if got, want := buf.String(), colors.StartText("%s", "start")+"   Linting...\n"; got != want {
		t.Fatalf("start output:\n got: %q\nwant: %q", got, want)
	}
	if got, want := buf.String(), "\x1b[36;1m"+"start"+"\x1b[0;22m   Linting...\n"; got != want {
		t.Fatalf("start ANSI contract:\n got: %q\nwant: %q", got, want)
	}

	buf.Reset()
	w := bufio.NewWriter(&buf)
	report := NewReport(nil, Metadata{Mode: ModeLint, Files: 2, Rules: 3, Threads: 4})
	renderSummary(w, report, Outcome{Kind: OutcomePassed}, 12*time.Millisecond, colors)
	if err := w.Flush(); err != nil {
		t.Fatal(err)
	}
	want := colors.SuccessText("%s", "success") + " Lint passed in 12ms " +
		colors.DimText("%s", "(2 files, 3 rules, 4 threads)") + "\n"
	if got := buf.String(); got != want {
		t.Fatalf("completed output:\n got: %q\nwant: %q", got, want)
	}
}

func TestLifecycleTextByMode(t *testing.T) {
	for _, test := range []struct {
		name       string
		mode       Mode
		start      string
		abortStart string
	}{
		{name: "lint", mode: ModeLint, start: "start   Linting...\n", abortStart: "error   Linting failed in "},
		{
			name:       "combined",
			mode:       ModeLintAndTypeCheck,
			start:      "start   Linting and type checking...\n",
			abortStart: "error   Linting and type checking failed in ",
		},
		{
			name:       "type-check-only",
			mode:       ModeTypeCheckOnly,
			start:      "start   Type checking...\n",
			abortStart: "error   Type checking failed in ",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			options := Options{Format: FormatDefault}
			var buf bytes.Buffer
			if err := RenderStart(&buf, test.mode, options); err != nil {
				t.Fatal(err)
			}
			if got := buf.String(); got != test.start {
				t.Fatalf("start output = %q, want %q", got, test.start)
			}

			buf.Reset()
			if err := RenderAbort(&buf, test.mode, time.Now().Add(-12*time.Millisecond), "boom", options); err != nil {
				t.Fatal(err)
			}
			if got := buf.String(); !strings.HasPrefix(got, test.abortStart) || !strings.HasSuffix(got, ": boom\n") {
				t.Fatalf("abort output = %q", got)
			}
		})
	}
}

func TestRenderValidationFailureLeavesLifecycleRecoverable(t *testing.T) {
	options := Options{Format: FormatDefault}
	var buf bytes.Buffer
	if err := RenderStart(&buf, ModeLint, options); err != nil {
		t.Fatal(err)
	}
	report := NewReport([]rule.RuleDiagnostic{{
		RuleName: "broken",
		FilePath: "index.ts",
		Severity: rule.SeverityError,
	}}, Metadata{Mode: ModeLint, StartedAt: time.Now()})
	err := Render(&buf, report, Outcome{Kind: OutcomeDiagnosticsFailed}, options)
	if err == nil {
		t.Fatal("Render() succeeded for a diagnostic without a source file")
	}
	if got, want := buf.String(), "start   Linting...\n"; got != want {
		t.Fatalf("validation failure partially rendered a report:\n got: %q\nwant: %q", got, want)
	}
	if err := RenderAbort(&buf, ModeLint, time.Now(), "writing lint report: "+err.Error(), options); err != nil {
		t.Fatal(err)
	}
	if got := buf.String(); !strings.Contains(got, "\nerror   Linting failed in <1ms: writing lint report: diagnostic") {
		t.Fatalf("validation failure did not leave a terminal lifecycle status: %q", got)
	}
}

func TestMachineFormatsHaveNoLifecycleOutput(t *testing.T) {
	for _, format := range []Format{FormatJSONLine, FormatGitHub, FormatGitLab} {
		options := Options{Format: format, ColorEnabled: true}
		var buf bytes.Buffer
		if err := RenderStart(&buf, ModeLint, options); err != nil {
			t.Fatal(err)
		}
		if err := RenderAbort(&buf, ModeLint, time.Now(), "boom", options); err != nil {
			t.Fatal(err)
		}
		if buf.Len() != 0 {
			t.Fatalf("%s lifecycle output = %q, want empty", format, buf.String())
		}
	}
}

func TestMachineFormatsHaveNoLeadingBlankLine(t *testing.T) {
	diagnostic, paths := createOutputTestDiagnostic(t, rule.SeverityWarning)
	report := NewReport([]rule.RuleDiagnostic{diagnostic}, Metadata{Mode: ModeLint})

	for _, format := range []Format{FormatJSONLine, FormatGitHub, FormatGitLab} {
		t.Run(format.String(), func(t *testing.T) {
			var buf bytes.Buffer
			if err := Render(&buf, report, Outcome{Kind: OutcomePassed}, Options{Format: format, ComparePaths: paths}); err != nil {
				t.Fatal(err)
			}
			if len(buf.Bytes()) == 0 || buf.Bytes()[0] == '\n' {
				t.Fatalf("unexpected leading blank line: %q", buf.String())
			}
			if strings.Contains(buf.String(), "Found ") {
				t.Fatalf("machine format contains summary: %q", buf.String())
			}
		})
	}
}

func TestGitHubEscapesEveryWorkflowCommandField(t *testing.T) {
	diagnostic, paths := createOutputTestDiagnostic(t, rule.SeverityWarning)
	diagnostic.RuleName = "rule%,:\r\n::warning"
	diagnostic.Message.Description = "message%\r\n::error"

	var buf bytes.Buffer
	if err := Render(&buf, NewReport([]rule.RuleDiagnostic{diagnostic}, Metadata{}), Outcome{Kind: OutcomePassed}, Options{
		Format: FormatGitHub, ComparePaths: paths,
	}); err != nil {
		t.Fatal(err)
	}
	want := "::warning file=index.ts,line=1,endLine=1,col=7,endColumn=12,title=rule%25%2C%3A%0D%0A%3A%3A" +
		"warning::message%25%0D%0A::error\n"
	if got := buf.String(); got != want {
		t.Fatalf("GitHub output:\n got: %q\nwant: %q", got, want)
	}
}

func TestJSONLineProtocol(t *testing.T) {
	diagnostic, paths := createOutputTestDiagnostic(t, rule.SeverityWarning)
	var buf bytes.Buffer
	if err := Render(&buf, NewReport([]rule.RuleDiagnostic{diagnostic}, Metadata{}), Outcome{Kind: OutcomePassed}, Options{
		Format: FormatJSONLine, ComparePaths: paths,
	}); err != nil {
		t.Fatal(err)
	}
	want := "{\"ruleName\":\"test-rule\",\"message\":\"test message\",\"filePath\":\"index.ts\",\"range\":{\"start\":{\"line\":1,\"column\":7},\"end\":{\"line\":1,\"column\":12}},\"severity\":\"warn\"}\n"
	if got := buf.String(); got != want {
		t.Fatalf("JSON line output:\n got: %q\nwant: %q", got, want)
	}
}

func TestQuietFiltersRenderingButNotCounts(t *testing.T) {
	diagnostic := rule.RuleDiagnostic{Severity: rule.SeverityWarning}
	report := NewReport([]rule.RuleDiagnostic{diagnostic}, Metadata{
		Mode: ModeLint, Threads: 1,
	})
	if report.Counts().Warnings != 1 {
		t.Fatalf("warning count = %d", report.Counts().Warnings)
	}

	for _, test := range []struct {
		format Format
		want   string
	}{
		{FormatJSONLine, ""},
		{FormatGitHub, ""},
		{FormatGitLab, "[]\n"},
	} {
		var buf bytes.Buffer
		if err := Render(&buf, report, Outcome{Kind: OutcomePassed}, Options{Format: test.format, Quiet: true}); err != nil {
			t.Fatal(err)
		}
		if got := buf.String(); got != test.want {
			t.Fatalf("%s quiet output = %q, want %q", test.format, got, test.want)
		}
	}

	var defaultBuf bytes.Buffer
	if err := Render(&defaultBuf, report, Outcome{Kind: OutcomePassed}, Options{Format: FormatDefault, Quiet: true}); err != nil {
		t.Fatal(err)
	}
	if strings.HasPrefix(defaultBuf.String(), "\n") {
		t.Fatalf("quiet warning-only default report has a stray blank line: %q", defaultBuf.String())
	}
	if !strings.Contains(defaultBuf.String(), "1 warning") {
		t.Fatalf("summary lost hidden warning count: %q", defaultBuf.String())
	}

	var lifecycle bytes.Buffer
	options := Options{Format: FormatDefault, Quiet: true}
	if err := RenderStart(&lifecycle, ModeLint, options); err != nil {
		t.Fatal(err)
	}
	if err := Render(&lifecycle, report, Outcome{
		Kind:         OutcomeWarningLimitExceeded,
		WarningLimit: 0,
	}, options); err != nil {
		t.Fatal(err)
	}
	if got := lifecycle.String(); strings.Contains(got, "\n\n") ||
		!strings.Contains(got, "1 warning exceeded the configured limit of 0") {
		t.Fatalf("quiet lifecycle output = %q", got)
	}
}

func TestGitLabEmptyAndFingerprintCollisions(t *testing.T) {
	var empty bytes.Buffer
	if err := Render(&empty, NewReport(nil, Metadata{}), Outcome{Kind: OutcomePassed}, Options{Format: FormatGitLab}); err != nil {
		t.Fatal(err)
	}
	if empty.String() != "[]\n" {
		t.Fatalf("empty GitLab report = %q", empty.String())
	}

	sequence := func() []string {
		state := newGitLabFingerprintState()
		return []string{
			state.fingerprint("f.ts", "rule", "msg", 1, 1, 1, 5),
			state.fingerprint("f.ts", "rule", "msg", 1, 1, 1, 5),
			// Its unsalted tuple equals the previous diagnostic's historical
			// colon-salted tuple. All emitted fingerprints must still be unique.
			state.fingerprint("f.ts", "rule", "msg:1", 1, 1, 5, 1),
		}
	}
	fingerprints := sequence()
	if fingerprints[0] == fingerprints[1] || fingerprints[0] == fingerprints[2] || fingerprints[1] == fingerprints[2] {
		t.Fatalf("fingerprints are not unique: %q", fingerprints)
	}
	if fresh := sequence(); !slices.Equal(fresh, fingerprints) {
		t.Fatalf("fingerprints are not deterministic: first=%q fresh=%q", fingerprints, fresh)
	}

	diagnostic, paths := createOutputTestDiagnostic(t, rule.SeverityError)
	var rendered bytes.Buffer
	if err := Render(&rendered, NewReport([]rule.RuleDiagnostic{diagnostic}, Metadata{}), Outcome{Kind: OutcomePassed}, Options{
		Format: FormatGitLab, ComparePaths: paths,
	}); err != nil {
		t.Fatal(err)
	}
	var issues []map[string]any
	if err := json.Unmarshal(rendered.Bytes(), &issues); err != nil || len(issues) != 1 {
		t.Fatalf("invalid GitLab JSON: %v, %s", err, rendered.String())
	}
}

func TestRenderValidatesAllDiagnosticsBeforeWriting(t *testing.T) {
	valid, paths := createOutputTestDiagnostic(t, rule.SeverityError)
	start := valid.Range.Pos()
	tests := []struct {
		name   string
		mutate func(*rule.RuleDiagnostic)
	}{
		{name: "missing source", mutate: func(d *rule.RuleDiagnostic) { d.SourceFile = nil }},
		{name: "negative start", mutate: func(d *rule.RuleDiagnostic) { d.Range = core.NewTextRange(-1, 0) }},
		{name: "reversed", mutate: func(d *rule.RuleDiagnostic) { d.Range = core.NewTextRange(start+1, start) }},
		{name: "past source", mutate: func(d *rule.RuleDiagnostic) {
			d.Range = core.NewTextRange(start, len(d.SourceFile.Text())+1)
		}},
	}

	for _, format := range []Format{FormatDefault, FormatJSONLine, FormatGitHub, FormatGitLab} {
		for _, test := range tests {
			t.Run(format.String()+"/"+test.name, func(t *testing.T) {
				bad := valid
				test.mutate(&bad)
				var buf bytes.Buffer
				err := Render(&buf, NewReport([]rule.RuleDiagnostic{valid, bad}, Metadata{}), Outcome{Kind: OutcomePassed}, Options{
					Format: format, ComparePaths: paths,
				})
				if err == nil {
					t.Fatal("expected invalid diagnostic to fail")
				}
				if buf.Len() != 0 {
					t.Fatalf("invalid report wrote partial output: %q", buf.String())
				}
			})
		}
	}
}

func TestRenderReturnsWriterError(t *testing.T) {
	want := errors.New("write failed")
	err := Render(failingWriter{err: want}, NewReport(nil, Metadata{}), Outcome{Kind: OutcomePassed}, Options{Format: FormatGitLab})
	if !errors.Is(err, want) {
		t.Fatalf("Render error = %v, want %v", err, want)
	}
}

func TestRenderRejectsUnknownFormat(t *testing.T) {
	err := Render(&bytes.Buffer{}, NewReport(nil, Metadata{}), Outcome{Kind: OutcomePassed}, Options{Format: Format(255)})
	if err == nil || !strings.Contains(err.Error(), "unsupported output format") {
		t.Fatalf("Render error = %v", err)
	}
}

type failingWriter struct{ err error }

func (w failingWriter) Write(_ []byte) (int, error) { return 0, w.err }

func createOutputTestDiagnostic(t *testing.T, severity rule.DiagnosticSeverity) (rule.RuleDiagnostic, tspath.ComparePathsOptions) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "tsconfig.json"), []byte(`{"include":["./index.ts"]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	source := "const value = 1;\n"
	if err := os.WriteFile(filepath.Join(dir, "index.ts"), []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	host := utils.CreateCompilerHost(dir, fs)
	program, err := utils.CreateProgram(true, fs, dir, "tsconfig.json", host)
	if err != nil {
		t.Fatal(err)
	}
	var sourceFile *ast.SourceFile
	for _, file := range program.GetSourceFiles() {
		if strings.HasSuffix(file.FileName(), "index.ts") {
			sourceFile = file
			break
		}
	}
	if sourceFile == nil {
		t.Fatal("source file not found")
		return rule.RuleDiagnostic{}, tspath.ComparePathsOptions{}
	}
	start := strings.Index(source, "value")
	return rule.RuleDiagnostic{
		RuleName:   "test-rule",
		SourceFile: sourceFile,
		FilePath:   sourceFile.FileName(),
		Range:      core.NewTextRange(start, start+len("value")),
		Message:    rule.RuleMessage{Description: "test message"},
		Severity:   severity,
	}, tspath.ComparePathsOptions{CurrentDirectory: dir, UseCaseSensitiveFileNames: true}
}
