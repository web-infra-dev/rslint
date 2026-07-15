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
		expected string
	}{
		{
			name: "lint zero plural",
			report: NewReport(nil, Metadata{
				Mode: ModeLint, LintedFiles: 2, Rules: 3, Threads: 4,
			}),
			expected: "Found 0 errors and 0 warnings (linted 2 files with 3 rules in 12ms using 4 threads)\n",
		},
		{
			name: "lint singular with fix",
			report: NewReport([]rule.RuleDiagnostic{
				{Severity: rule.SeverityError},
				{Severity: rule.SeverityWarning},
			}, Metadata{
				Mode: ModeLint, LintedFiles: 1, Rules: 1, Threads: 1, FixedIssues: 1,
			}),
			expected: "Found 1 error and 1 warning (linted 1 file with 1 rule in 12ms using 1 thread, fixed 1 issue)\n",
		},
		{
			name: "lint and type-check",
			report: NewReport([]rule.RuleDiagnostic{
				{RuleName: "no-debugger", Severity: rule.SeverityError},
				{RuleName: "TypeScript(TS2322)", Severity: rule.SeverityError, Origin: rule.DiagnosticOriginTypeScript},
			}, Metadata{
				Mode: ModeLintAndTypeCheck, LintedFiles: 1, TypeCheckedFiles: 2, Rules: 0, Threads: 2,
			}),
			expected: "Found 1 lint error, 1 type error and 0 warnings (linted 1 file with 0 rules, type-checked 2 files in 12ms using 2 threads)\n",
		},
		{
			name: "type-check only",
			report: NewReport([]rule.RuleDiagnostic{
				{RuleName: "TypeScript(TS2322)", Severity: rule.SeverityError, Origin: rule.DiagnosticOriginTypeScript},
			}, Metadata{
				Mode: ModeTypeCheckOnly, TypeCheckedFiles: 1, Threads: 2,
			}),
			expected: "Found 1 type error (type-checked 1 file in 12ms using 2 threads)\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			w := bufio.NewWriter(&buf)
			renderSummary(w, test.report, 12*time.Millisecond, newColorScheme(false))
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
			name: "lint and fixed details",
			report: NewReport(nil, Metadata{
				Mode: ModeLint, LintedFiles: 2, Rules: 3, Threads: 4, FixedIssues: 5,
			}),
			details: "(linted 2 files with 3 rules in 12ms using 4 threads, fixed 5 issues)",
		},
		{
			name: "lint and type-check details",
			report: NewReport(nil, Metadata{
				Mode: ModeLintAndTypeCheck, LintedFiles: 1, TypeCheckedFiles: 2, Rules: 3, Threads: 1,
			}),
			details: "(linted 1 file with 3 rules, type-checked 2 files in 12ms using 1 thread)",
		},
		{
			name: "type-check-only details",
			report: NewReport(nil, Metadata{
				Mode: ModeTypeCheckOnly, TypeCheckedFiles: 2, Threads: 4,
			}),
			details: "(type-checked 2 files in 12ms using 4 threads)",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			w := bufio.NewWriter(&buf)
			renderSummary(w, test.report, 12*time.Millisecond, newColorScheme(true))
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

func TestMixedSummaryColorsErrorCountsIndependently(t *testing.T) {
	report := NewReport([]rule.RuleDiagnostic{
		{Severity: rule.SeverityError, Origin: rule.DiagnosticOriginTypeScript},
	}, Metadata{
		Mode: ModeLintAndTypeCheck, LintedFiles: 1, TypeCheckedFiles: 1, Threads: 1,
	})
	colors := newColorScheme(true)
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	renderSummary(w, report, 12*time.Millisecond, colors)
	if err := w.Flush(); err != nil {
		t.Fatal(err)
	}

	got := buf.String()
	if want := colors.SuccessText("%d", 0) + " lint errors"; !strings.Contains(got, want) {
		t.Fatalf("zero lint-error count is not green:\n%s", got)
	}
	if want := colors.ErrorText("%d", 1) + " type error"; !strings.Contains(got, want) {
		t.Fatalf("non-zero type-error count is not red:\n%s", got)
	}
	if unwanted := colors.ErrorText("%d", 0) + " lint errors"; strings.Contains(got, unwanted) {
		t.Fatalf("zero lint-error count is red:\n%s", got)
	}
}

func TestMachineFormatsHaveNoLeadingBlankLine(t *testing.T) {
	diagnostic, paths := createOutputTestDiagnostic(t, rule.SeverityWarning)
	report := NewReport([]rule.RuleDiagnostic{diagnostic}, Metadata{Mode: ModeLint})

	for _, format := range []Format{FormatJSONLine, FormatGitHub, FormatGitLab} {
		t.Run(format.String(), func(t *testing.T) {
			var buf bytes.Buffer
			if err := Render(&buf, report, Options{Format: format, ComparePaths: paths}); err != nil {
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
	if err := Render(&buf, NewReport([]rule.RuleDiagnostic{diagnostic}, Metadata{}), Options{
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
	if err := Render(&buf, NewReport([]rule.RuleDiagnostic{diagnostic}, Metadata{}), Options{
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
		if err := Render(&buf, report, Options{Format: test.format, Quiet: true}); err != nil {
			t.Fatal(err)
		}
		if got := buf.String(); got != test.want {
			t.Fatalf("%s quiet output = %q, want %q", test.format, got, test.want)
		}
	}

	var defaultBuf bytes.Buffer
	if err := Render(&defaultBuf, report, Options{Format: FormatDefault, Quiet: true}); err != nil {
		t.Fatal(err)
	}
	if defaultBuf.Len() != 0 {
		t.Fatalf("quiet warning-only default output = %q, want empty", defaultBuf.String())
	}
}

func TestGitLabEmptyAndFingerprintCollisions(t *testing.T) {
	var empty bytes.Buffer
	if err := Render(&empty, NewReport(nil, Metadata{}), Options{Format: FormatGitLab}); err != nil {
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
	if err := Render(&rendered, NewReport([]rule.RuleDiagnostic{diagnostic}, Metadata{}), Options{
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
				err := Render(&buf, NewReport([]rule.RuleDiagnostic{valid, bad}, Metadata{}), Options{
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
	err := Render(failingWriter{err: want}, NewReport(nil, Metadata{}), Options{Format: FormatGitLab})
	if !errors.Is(err, want) {
		t.Fatalf("Render error = %v, want %v", err, want)
	}
}

func TestRenderRejectsUnknownFormat(t *testing.T) {
	err := Render(&bytes.Buffer{}, NewReport(nil, Metadata{}), Options{Format: Format(255)})
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

func TestFileWarningsUseFormatterAndCountWithoutSyntheticRange(t *testing.T) {
	dir := t.TempDir()
	warning := FileWarning{
		FilePath: filepath.Join(dir, `literal\\file.ts`),
		Message:  "is ignored because of a matching ignore pattern",
	}
	paths := tspath.ComparePathsOptions{CurrentDirectory: dir, UseCaseSensitiveFileNames: true}
	report := NewReportWithFileWarnings(nil, []FileWarning{warning}, Metadata{Mode: ModeLint})
	if got := report.Counts().Warnings; got != 1 {
		t.Fatalf("warning count = %d, want 1", got)
	}

	tests := []struct {
		format Format
		want   []string
		not    []string
	}{
		{FormatDefault, []string{"literal\\\\file.ts", warning.Message, "1 warning"}, nil},
		{FormatJSONLine, []string{`"filePath":"literal\\\\file.ts"`, `"severity":"warning"`, warning.Message}, []string{`"range"`, `"ruleName"`}},
		{FormatGitHub, []string{"::warning file=literal\\\\file.ts::", warning.Message}, []string{"line="}},
		{FormatGitLab, []string{`"check_name":"rslint/file-warning"`, `"begin":1`, warning.Message}, []string{`"positions"`}},
	}
	for _, test := range tests {
		t.Run(test.format.String(), func(t *testing.T) {
			var rendered strings.Builder
			if err := Render(&rendered, report, Options{Format: test.format, ComparePaths: paths}); err != nil {
				t.Fatal(err)
			}
			for _, want := range test.want {
				if !strings.Contains(rendered.String(), want) {
					t.Fatalf("output %q does not contain %q", rendered.String(), want)
				}
			}
			for _, unwanted := range test.not {
				if strings.Contains(rendered.String(), unwanted) {
					t.Fatalf("output %q unexpectedly contains %q", rendered.String(), unwanted)
				}
			}
		})
	}

	var quiet strings.Builder
	if err := Render(&quiet, report, Options{Format: FormatJSONLine, ComparePaths: paths, Quiet: true}); err != nil {
		t.Fatal(err)
	}
	if quiet.Len() != 0 {
		t.Fatalf("quiet warning output = %q, want empty", quiet.String())
	}
	if report.Counts().Warnings != 1 {
		t.Fatal("quiet rendering must not change warning counts")
	}
}
