package output

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"math"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/microsoft/typescript-go/shim/tspath"
)

func newTestReport(diagnostics []Diagnostic, summary Summary) Report {
	return newTestReportForMode(ModeLint, diagnostics, 0, summary)
}

func newTestReportForMode(mode Mode, diagnostics []Diagnostic, typeErrors int, summary Summary) Report {
	counts := Counts{}
	for _, diagnostic := range diagnostics {
		switch diagnostic.Severity {
		case SeverityError:
			counts.Errors++
		case SeverityWarning:
			counts.Warnings++
		}
	}
	counts.TypeErrors = typeErrors
	counts.LintErrors = counts.Errors - typeErrors
	outcome := Outcome{Kind: OutcomePassed}
	if counts.Errors > 0 {
		outcome.Kind = OutcomeDiagnosticsFailed
	}
	return NewReport(mode, diagnostics, counts, &summary, outcome)
}

func renderTest(dst io.Writer, report Report, outcome Outcome, options Options) error {
	report.outcome = outcome
	return Render(dst, report, options)
}

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
	lineStarts := []int{0}
	source, err := NewDiagnosticSource("before", lineStarts)
	if err != nil {
		t.Fatal(err)
	}
	diagnostics := []Diagnostic{{RuleName: "before", Source: source, Severity: SeverityError}}
	counts := Counts{Errors: 1, LintErrors: 1}
	report := NewReport(ModeLint, diagnostics, counts, &Summary{}, Outcome{Kind: OutcomeDiagnosticsFailed})
	lineStarts[0] = 1
	diagnostics[0].Severity = SeverityWarning
	diagnostics[0].RuleName = "after"
	if report.diagnostics[0].RuleName != "before" || report.Counts() != counts ||
		report.Outcome().Kind != OutcomeDiagnosticsFailed ||
		report.diagnostics[0].Source.lineStarts[0] != 0 {
		t.Fatalf("report changed after caller mutation: diagnostic=%+v counts=%+v", report.diagnostics[0], report.Counts())
	}
}

func TestNewDiagnosticSourceValidatesStructuralLineStarts(t *testing.T) {
	for _, test := range []struct {
		name       string
		text       string
		lineStarts []int
	}{
		{name: "missing"},
		{name: "nonzero first", text: "x", lineStarts: []int{1}},
		{name: "negative", text: "x", lineStarts: []int{0, -1}},
		{name: "past text", text: "x", lineStarts: []int{0, 2}},
		{name: "duplicate", text: "x", lineStarts: []int{0, 0}},
		{name: "decreasing", text: "abc", lineStarts: []int{0, 2, 1}},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := NewDiagnosticSource(test.text, test.lineStarts); err == nil {
				t.Fatal("invalid line starts were accepted")
			}
		})
	}

	for _, test := range []struct {
		name       string
		text       string
		lineStarts []int
	}{
		{name: "empty", lineStarts: []int{0}},
		{name: "LF", text: "a\nb", lineStarts: []int{0, 2}},
		{name: "CR", text: "a\rb", lineStarts: []int{0, 2}},
		{name: "CRLF", text: "a\r\nb", lineStarts: []int{0, 3}},
		{name: "line separator", text: "a\u2028b", lineStarts: []int{0, 4}},
		{name: "paragraph separator", text: "a\u2029b", lineStarts: []int{0, 4}},
		{name: "trailing line", text: "x\n", lineStarts: []int{0, 2}},
		// The producer owns line-break semantics. Output accepts any bounded,
		// ordered map and separately verifies each projected diagnostic line.
		{name: "producer metadata", text: "a\nb", lineStarts: []int{0, 3}},
	} {
		t.Run("valid "+test.name, func(t *testing.T) {
			if _, err := NewDiagnosticSource(test.text, test.lineStarts); err != nil {
				t.Fatalf("valid line starts were rejected: %v", err)
			}
		})
	}
}

func TestRenderDefaultAcceptsStructurallySafeProducerMap(t *testing.T) {
	rendered := renderDefaultLineBreakFixture(t, "a\nb", []int{0, 3}, 0, 0)
	if !strings.Contains(rendered, "test message") {
		t.Fatalf("rendered output = %q", rendered)
	}
}

func TestRenderDefaultPreservesLegacyLineBreakRendering(t *testing.T) {
	tests := []struct {
		name           string
		separator      string
		legacyCombined bool
	}{
		{name: "LF", separator: "\n"},
		{name: "CRLF", separator: "\r\n"},
		{name: "CR", separator: "\r", legacyCombined: true},
		{name: "line separator", separator: "\u2028", legacyCombined: true},
		{name: "paragraph separator", separator: "\u2029", legacyCombined: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			text := "a" + test.separator + "b\n"
			secondLineStart := 1 + len(test.separator)
			rendered := renderDefaultLineBreakFixture(
				t,
				text,
				[]int{0, secondLineStart, len(text)},
				secondLineStart,
				1,
			)

			if test.legacyCombined {
				if !strings.Contains(rendered, "  │ 1 │  a"+test.separator+"b\n") ||
					!strings.Contains(rendered, "  │ 2 │  \n") {
					t.Fatalf("legacy combined line rendering changed: %q", rendered)
				}
				return
			}
			if !strings.Contains(rendered, "  │ 1 │  a\n") ||
				!strings.Contains(rendered, "  │ 2 │  b\n") {
				t.Fatalf("legacy LF line rendering changed: %q", rendered)
			}
			if strings.Contains(rendered, "\r") {
				t.Fatalf("CRLF terminator leaked into rendered content: %q", rendered)
			}
		})
	}
}

func renderDefaultLineBreakFixture(
	t *testing.T,
	text string,
	lineStarts []int,
	rangeStart int,
	line int,
) string {
	t.Helper()
	source, err := NewDiagnosticSource(text, lineStarts)
	if err != nil {
		t.Fatal(err)
	}
	diagnostic := Diagnostic{
		FilePath: "/repo/test.ts",
		RuleName: "test-rule",
		Message:  "test message",
		Range: TextRange{
			Start: rangeStart,
			End:   rangeStart + 1,
		},
		Start:    Position{Line: line},
		End:      Position{Line: line, Column: 1},
		Source:   source,
		Severity: SeverityError,
	}
	report := NewReport(
		ModeLint,
		[]Diagnostic{diagnostic},
		Counts{Errors: 1, LintErrors: 1},
		&Summary{Files: 1, Rules: 1, Threads: 1},
		Outcome{Kind: OutcomeDiagnosticsFailed},
	)
	var rendered bytes.Buffer
	if err := Render(&rendered, report, Options{
		Format: FormatDefault,
		ComparePaths: tspath.ComparePathsOptions{
			CurrentDirectory:          "/repo",
			UseCaseSensitiveFileNames: true,
		},
	}); err != nil {
		t.Fatal(err)
	}
	return rendered.String()
}

func TestOutcomeFailed(t *testing.T) {
	if (Outcome{Kind: OutcomePassed}).Failed() {
		t.Fatal("passed outcome reported failure")
	}
	if !(Outcome{Kind: OutcomeDiagnosticsFailed}).Failed() ||
		!(Outcome{Kind: OutcomeWarningLimitExceeded}).Failed() {
		t.Fatal("failed outcome reported success")
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
			report: newTestReport(nil, Summary{
				Files: 2, Rules: 3, Threads: 4,
			}),
			outcome:  Outcome{Kind: OutcomePassed},
			expected: "success Lint passed in 12ms (2 files, 3 rules, 4 threads)\n",
		},
		{
			name: "lint passed with warnings",
			report: newTestReport([]Diagnostic{
				{Severity: SeverityWarning},
				{Severity: SeverityWarning},
			}, Summary{
				Files: 2, Rules: 3, Threads: 4,
			}),
			outcome:  Outcome{Kind: OutcomePassed},
			expected: "success Lint passed with 2 warnings in 12ms (2 files, 3 rules, 4 threads)\n",
		},
		{
			name: "lint passed after applying fixes",
			report: newTestReport(nil, Summary{
				Files: 2, Rules: 3, Threads: 4, FixedIssues: 2,
			}),
			outcome:  Outcome{Kind: OutcomePassed},
			expected: "success Lint passed after applying 2 fixes in 12ms (2 files, 3 rules, 4 threads)\n",
		},
		{
			name: "lint passed with warning and fix",
			report: newTestReport([]Diagnostic{
				{Severity: SeverityWarning},
			}, Summary{
				Files: 1, Rules: 1, Threads: 1, FixedIssues: 1,
			}),
			outcome:  Outcome{Kind: OutcomePassed},
			expected: "success Lint passed with 1 warning after applying 1 fix in 12ms (1 file, 1 rule, 1 thread)\n",
		},
		{
			name: "lint failed with fix",
			report: newTestReport([]Diagnostic{
				{Severity: SeverityError},
				{Severity: SeverityWarning},
			}, Summary{
				Files: 5, Rules: 7, Threads: 2, FixedIssues: 3,
			}),
			outcome:  Outcome{Kind: OutcomeDiagnosticsFailed},
			expected: "error   Lint failed with 1 error and 1 warning after applying 3 fixes in 12ms (5 files, 7 rules, 2 threads)\n",
		},
		{
			name: "lint and type check failed",
			report: newTestReportForMode(ModeLintAndTypeCheck, []Diagnostic{
				{Severity: SeverityError},
				{Severity: SeverityError},
				{Severity: SeverityError},
				{Severity: SeverityWarning},
			}, 1, Summary{
				Files: 9, Rules: 8, Threads: 2,
			}),
			outcome:  Outcome{Kind: OutcomeDiagnosticsFailed},
			expected: "error   Lint and type check failed with 2 lint errors, 1 TypeScript error, and 1 warning in 12ms (9 files, 8 rules, 2 threads)\n",
		},
		{
			name: "lint and type check passed",
			report: newTestReportForMode(ModeLintAndTypeCheck, nil, 0, Summary{
				Files: 9, Rules: 8, Threads: 2,
			}),
			outcome:  Outcome{Kind: OutcomePassed},
			expected: "success Lint and type check passed in 12ms (9 files, 8 rules, 2 threads)\n",
		},
		{
			name: "lint and type check failed after applying fixes",
			report: newTestReportForMode(ModeLintAndTypeCheck, []Diagnostic{
				{Severity: SeverityError},
				{Severity: SeverityError},
			}, 1, Summary{
				Files: 9, Rules: 8, Threads: 2, FixedIssues: 2,
			}),
			outcome:  Outcome{Kind: OutcomeDiagnosticsFailed},
			expected: "error   Lint and type check failed with 1 lint error and 1 TypeScript error after applying 2 fixes in 12ms (9 files, 8 rules, 2 threads)\n",
		},
		{
			name: "lint and type check warning limit exceeded",
			report: newTestReportForMode(ModeLintAndTypeCheck, []Diagnostic{
				{Severity: SeverityWarning},
				{Severity: SeverityWarning},
			}, 0, Summary{
				Files: 9, Rules: 8, Threads: 2,
			}),
			outcome: Outcome{Kind: OutcomeWarningLimitExceeded, WarningLimit: 1},
			expected: "error   Lint and type check failed in 12ms: 2 warnings exceeded the configured limit of 1 " +
				"(9 files, 8 rules, 2 threads)\n",
		},
		{
			name: "type check only failed",
			report: newTestReportForMode(ModeTypeCheckOnly, []Diagnostic{
				{Severity: SeverityError},
			}, 1, Summary{
				Files: 1, Threads: 2,
			}),
			outcome:  Outcome{Kind: OutcomeDiagnosticsFailed},
			expected: "error   Type check failed with 1 TypeScript error in 12ms (1 file, 2 threads)\n",
		},
		{
			name: "type check only passed",
			report: newTestReportForMode(ModeTypeCheckOnly, nil, 0, Summary{
				Files: 1, Threads: 2,
			}),
			outcome:  Outcome{Kind: OutcomePassed},
			expected: "success Type check passed in 12ms (1 file, 2 threads)\n",
		},
		{
			name: "warning limit exceeded",
			report: newTestReport([]Diagnostic{
				{Severity: SeverityWarning},
				{Severity: SeverityWarning},
			}, Summary{
				Files: 2, Rules: 3, Threads: 4,
			}),
			outcome: Outcome{Kind: OutcomeWarningLimitExceeded, WarningLimit: 0},
			expected: "error   Lint failed in 12ms: 2 warnings exceeded the configured limit of 0 " +
				"(2 files, 3 rules, 4 threads)\n",
		},
		{
			name: "warning limit exceeded after applying fixes",
			report: newTestReport([]Diagnostic{
				{Severity: SeverityWarning},
				{Severity: SeverityWarning},
			}, Summary{
				Files: 2, Rules: 3, Threads: 4, FixedIssues: 1,
			}),
			outcome: Outcome{Kind: OutcomeWarningLimitExceeded, WarningLimit: 0},
			expected: "error   Lint failed after applying 1 fix in 12ms: 2 warnings exceeded the configured limit of 0 " +
				"(2 files, 3 rules, 4 threads)\n",
		},
		{
			name:     "empty lint",
			report:   newTestReport(nil, Summary{Threads: 4}),
			outcome:  Outcome{Kind: OutcomePassed},
			expected: "success No files to lint in 12ms\n",
		},
		{
			name:     "empty combined",
			report:   newTestReportForMode(ModeLintAndTypeCheck, nil, 0, Summary{Threads: 4}),
			outcome:  Outcome{Kind: OutcomePassed},
			expected: "success No files to lint or type check in 12ms\n",
		},
		{
			name:     "empty type check",
			report:   newTestReportForMode(ModeTypeCheckOnly, nil, 0, Summary{Threads: 4}),
			outcome:  Outcome{Kind: OutcomePassed},
			expected: "success No files to type check in 12ms\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			w := bufio.NewWriter(&buf)
			test.report.outcome = test.outcome
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
			name: "lint details",
			report: newTestReport(nil, Summary{
				Files: 2, Rules: 3, Threads: 4,
			}),
			details: "(2 files, 3 rules, 4 threads)",
		},
		{
			name: "lint and type-check details",
			report: newTestReportForMode(ModeLintAndTypeCheck, nil, 0, Summary{
				Files: 2, Rules: 3, Threads: 1,
			}),
			details: "(2 files, 3 rules, 1 thread)",
		},
		{
			name: "type-check-only details",
			report: newTestReportForMode(ModeTypeCheckOnly, nil, 0, Summary{
				Files: 2, Threads: 4,
			}),
			details: "(2 files, 4 threads)",
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

func TestLifecycleColors(t *testing.T) {
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
	report := newTestReport(nil, Summary{Files: 2, Rules: 3, Threads: 4})
	renderSummary(w, report, 12*time.Millisecond, colors)
	if err := w.Flush(); err != nil {
		t.Fatal(err)
	}
	want := colors.SuccessText("%s", "success") + " Lint passed in 12ms " +
		colors.DimText("%s", "(2 files, 3 rules, 4 threads)") + "\n"
	if got := buf.String(); got != want {
		t.Fatalf("completed output:\n got: %q\nwant: %q", got, want)
	}
	if got, want := buf.String(), "\x1b[32;1m"+"success"+"\x1b[0;22m Lint passed in 12ms "+
		"\x1b[2m(2 files, 3 rules, 4 threads)\x1b[22m\n"; got != want {
		t.Fatalf("success ANSI contract:\n got: %q\nwant: %q", got, want)
	}

	buf.Reset()
	w = bufio.NewWriter(&buf)
	report = newTestReport([]Diagnostic{{Severity: SeverityError}}, Summary{
		Files: 2, Rules: 3, Threads: 4,
	})
	renderSummary(w, report, 12*time.Millisecond, colors)
	if err := w.Flush(); err != nil {
		t.Fatal(err)
	}
	if got, want := buf.String(), "\x1b[31;1m"+"error"+"\x1b[0;22m   Lint failed with "+
		"\x1b[31;1m1\x1b[0;22m error in 12ms "+
		"\x1b[2m(2 files, 3 rules, 4 threads)\x1b[22m\n"; got != want {
		t.Fatalf("error ANSI contract:\n got: %q\nwant: %q", got, want)
	}

	buf.Reset()
	w = bufio.NewWriter(&buf)
	report = newTestReportForMode(ModeLintAndTypeCheck, []Diagnostic{
		{Severity: SeverityError},
		{Severity: SeverityError},
		{Severity: SeverityError},
		{Severity: SeverityWarning},
	}, 1, Summary{Files: 2, Rules: 3, Threads: 4})
	renderSummary(w, report, 12*time.Millisecond, colors)
	if err := w.Flush(); err != nil {
		t.Fatal(err)
	}
	want = colors.ErrorText("%s", "error") + "   Lint and type check failed with " +
		colors.ErrorText("%d", 2) + " lint errors, " +
		colors.ErrorText("%d", 1) + " TypeScript error, and " +
		colors.WarnText("%d", 1) + " warning in 12ms " +
		colors.DimText("%s", "(2 files, 3 rules, 4 threads)") + "\n"
	if got := buf.String(); got != want {
		t.Fatalf("mixed diagnostics ANSI contract:\n got: %q\nwant: %q", got, want)
	}

	buf.Reset()
	w = bufio.NewWriter(&buf)
	report = newTestReport([]Diagnostic{
		{Severity: SeverityWarning},
		{Severity: SeverityWarning},
	}, Summary{Files: 2, Rules: 3, Threads: 4})
	renderSummary(w, report, 12*time.Millisecond, colors)
	if err := w.Flush(); err != nil {
		t.Fatal(err)
	}
	want = colors.SuccessText("%s", "success") + " Lint passed with " +
		colors.WarnText("%d", 2) + " warnings in 12ms " +
		colors.DimText("%s", "(2 files, 3 rules, 4 threads)") + "\n"
	if got := buf.String(); got != want {
		t.Fatalf("warning success ANSI contract:\n got: %q\nwant: %q", got, want)
	}

	buf.Reset()
	w = bufio.NewWriter(&buf)
	report.outcome = Outcome{Kind: OutcomeWarningLimitExceeded, WarningLimit: 0}
	renderSummary(w, report, 12*time.Millisecond, colors)
	if err := w.Flush(); err != nil {
		t.Fatal(err)
	}
	want = colors.ErrorText("%s", "error") + "   Lint failed in 12ms: " +
		colors.WarnText("%d", 2) + " warnings exceeded the configured limit of 0 " +
		colors.DimText("%s", "(2 files, 3 rules, 4 threads)") + "\n"
	if got := buf.String(); got != want {
		t.Fatalf("warning limit ANSI contract:\n got: %q\nwant: %q", got, want)
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
	report := newTestReport([]Diagnostic{{
		RuleName: "broken",
		FilePath: "index.ts",
		Severity: SeverityError,
	}}, Summary{StartedAt: time.Now()})
	err := renderTest(&buf, report, Outcome{Kind: OutcomeDiagnosticsFailed}, options)
	if err == nil {
		t.Fatal("renderTest() succeeded for a diagnostic without a source file")
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
	diagnostic, paths := createOutputTestDiagnostic(t, SeverityWarning)
	report := newTestReport([]Diagnostic{diagnostic}, Summary{})

	for _, format := range []Format{FormatJSONLine, FormatGitHub, FormatGitLab} {
		t.Run(format.String(), func(t *testing.T) {
			var buf bytes.Buffer
			if err := renderTest(&buf, report, Outcome{Kind: OutcomePassed}, Options{Format: format, ComparePaths: paths}); err != nil {
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
	diagnostic, paths := createOutputTestDiagnostic(t, SeverityWarning)
	diagnostic.RuleName = "rule%,:\r\n::warning"
	diagnostic.Message = "message%\r\n::error"

	var buf bytes.Buffer
	if err := renderTest(&buf, newTestReport([]Diagnostic{diagnostic}, Summary{}), Outcome{Kind: OutcomePassed}, Options{
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
	diagnostic, paths := createOutputTestDiagnostic(t, SeverityWarning)
	var buf bytes.Buffer
	if err := renderTest(&buf, newTestReport([]Diagnostic{diagnostic}, Summary{}), Outcome{Kind: OutcomePassed}, Options{
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
	diagnostic := Diagnostic{Severity: SeverityWarning}
	report := newTestReport([]Diagnostic{diagnostic}, Summary{
		Threads: 1,
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
		if err := renderTest(&buf, report, Outcome{Kind: OutcomePassed}, Options{Format: test.format, Quiet: true}); err != nil {
			t.Fatal(err)
		}
		if got := buf.String(); got != test.want {
			t.Fatalf("%s quiet output = %q, want %q", test.format, got, test.want)
		}
	}

	var defaultBuf bytes.Buffer
	if err := renderTest(&defaultBuf, report, Outcome{Kind: OutcomePassed}, Options{Format: FormatDefault, Quiet: true}); err != nil {
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
	if err := renderTest(&lifecycle, report, Outcome{
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
	if err := renderTest(&empty, newTestReport(nil, Summary{}), Outcome{Kind: OutcomePassed}, Options{Format: FormatGitLab}); err != nil {
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

	diagnostic, paths := createOutputTestDiagnostic(t, SeverityError)
	var rendered bytes.Buffer
	if err := renderTest(&rendered, newTestReport([]Diagnostic{diagnostic}, Summary{}), Outcome{Kind: OutcomePassed}, Options{
		Format: FormatGitLab, ComparePaths: paths,
	}); err != nil {
		t.Fatal(err)
	}
	var issues []map[string]any
	if err := json.Unmarshal(rendered.Bytes(), &issues); err != nil || len(issues) != 1 {
		t.Fatalf("invalid GitLab JSON: %v, %s", err, rendered.String())
	}
}

func TestRenderDefaultRequiresSummary(t *testing.T) {
	report := NewReport(ModeLint, nil, Counts{}, nil, Outcome{Kind: OutcomePassed})
	var rendered bytes.Buffer
	if err := Render(&rendered, report, Options{Format: FormatDefault}); err == nil {
		t.Fatal("default render accepted a diagnostics-only report")
	}
	if rendered.Len() != 0 {
		t.Fatalf("invalid default report wrote %q", rendered.String())
	}
}

func TestRenderDefaultValidatesAllSourcesBeforeWriting(t *testing.T) {
	valid, paths := createOutputTestDiagnostic(t, SeverityError)
	start := valid.Range.Start
	tests := []struct {
		name   string
		mutate func(*Diagnostic)
	}{
		{name: "missing source", mutate: func(d *Diagnostic) { d.Source = nil }},
		{name: "negative start", mutate: func(d *Diagnostic) { d.Range = TextRange{Start: -1} }},
		{name: "reversed", mutate: func(d *Diagnostic) { d.Range = TextRange{Start: start + 1, End: start} }},
		{name: "past source", mutate: func(d *Diagnostic) {
			d.Range = TextRange{Start: start, End: len(d.Source.text) + 1}
		}},
		{name: "inconsistent start", mutate: func(d *Diagnostic) { d.Start.Line++ }},
		{name: "inconsistent end", mutate: func(d *Diagnostic) { d.End.Line++ }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bad := valid
			test.mutate(&bad)
			var buf bytes.Buffer
			err := renderTest(&buf, newTestReport([]Diagnostic{valid, bad}, Summary{}), Outcome{Kind: OutcomePassed}, Options{
				Format: FormatDefault, ComparePaths: paths,
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

func TestMachineFormatsConsumeProjectedLocationsWithoutSource(t *testing.T) {
	diagnostic, paths := createOutputTestDiagnostic(t, SeverityError)
	diagnostic.Source = nil
	report := newTestReport([]Diagnostic{diagnostic}, Summary{})
	for _, format := range []Format{FormatJSONLine, FormatGitHub, FormatGitLab} {
		t.Run(format.String(), func(t *testing.T) {
			var rendered bytes.Buffer
			if err := Render(&rendered, report, Options{Format: format, ComparePaths: paths}); err != nil {
				t.Fatal(err)
			}
			if rendered.Len() == 0 {
				t.Fatal("machine formatter emitted no diagnostic")
			}
		})
	}
}

func TestRenderValidatesAllProjectedLocationsBeforeWriting(t *testing.T) {
	valid, paths := createOutputTestDiagnostic(t, SeverityError)
	for _, test := range []struct {
		name   string
		mutate func(*Diagnostic)
	}{
		{name: "negative", mutate: func(diagnostic *Diagnostic) { diagnostic.End.Column = -1 }},
		{name: "overflow", mutate: func(diagnostic *Diagnostic) { diagnostic.End.Column = math.MaxInt }},
	} {
		t.Run(test.name, func(t *testing.T) {
			invalid := valid
			test.mutate(&invalid)
			for _, format := range []Format{FormatDefault, FormatJSONLine, FormatGitHub, FormatGitLab} {
				t.Run(format.String(), func(t *testing.T) {
					var rendered bytes.Buffer
					err := Render(
						&rendered,
						newTestReport([]Diagnostic{valid, invalid}, Summary{}),
						Options{Format: format, ComparePaths: paths},
					)
					if err == nil {
						t.Fatal("invalid projected location was accepted")
					}
					if rendered.Len() != 0 {
						t.Fatalf("invalid report wrote partial output: %q", rendered.String())
					}
				})
			}
		})
	}
}

func TestRenderReturnsWriterError(t *testing.T) {
	want := errors.New("write failed")
	err := renderTest(failingWriter{err: want}, newTestReport(nil, Summary{}), Outcome{Kind: OutcomePassed}, Options{Format: FormatGitLab})
	if !errors.Is(err, want) {
		t.Fatalf("Render error = %v, want %v", err, want)
	}
}

func TestRenderRejectsUnknownFormat(t *testing.T) {
	err := renderTest(&bytes.Buffer{}, newTestReport(nil, Summary{}), Outcome{Kind: OutcomePassed}, Options{Format: Format(255)})
	if err == nil || !strings.Contains(err.Error(), "unsupported output format") {
		t.Fatalf("Render error = %v", err)
	}
}

type failingWriter struct{ err error }

func (w failingWriter) Write(_ []byte) (int, error) { return 0, w.err }

func createOutputTestDiagnostic(t *testing.T, severity Severity) (Diagnostic, tspath.ComparePathsOptions) {
	t.Helper()
	dir := t.TempDir()
	source := "const value = 1;\n"
	projectedSource, err := NewDiagnosticSource(source, []int{0, len(source)})
	if err != nil {
		t.Fatal(err)
	}
	start := strings.Index(source, "value")
	return Diagnostic{
		RuleName: "test-rule",
		FilePath: filepath.Join(dir, "index.ts"),
		Range: TextRange{
			Start: start,
			End:   start + len("value"),
		},
		Start:    Position{Line: 0, Column: start},
		End:      Position{Line: 0, Column: start + len("value")},
		Source:   projectedSource,
		Message:  "test message",
		Severity: severity,
	}, tspath.ComparePathsOptions{CurrentDirectory: dir, UseCaseSensitiveFileNames: true}
}
