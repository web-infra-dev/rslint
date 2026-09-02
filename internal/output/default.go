package output

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

type defaultFormatter struct {
	colors colorScheme
}

const statusLabelWidth = 7

// RenderStart writes the interactive lifecycle prefix for the default format.
// Machine-readable formats intentionally remain diagnostics-only.
func RenderStart(dst io.Writer, mode Mode, options Options) error {
	if options.Format != FormatDefault {
		return nil
	}
	colors := newColorScheme(options.ColorEnabled)
	return renderStatusLine(dst, "start", colors.StartText, startMessage(mode), "")
}

// RenderAbort writes a terminal status for a run that started but could not
// produce a complete report. Cancellation is filtered by the CLI and does not
// enter this presentation path.
func RenderAbort(dst io.Writer, mode Mode, startedAt time.Time, reason string, options Options) error {
	if options.Format != FormatDefault {
		return nil
	}
	colors := newColorScheme(options.ColorEnabled)
	message := fmt.Sprintf("%s failed in %s: %s", ongoingAction(mode), formatElapsed(time.Since(startedAt)), reason)
	return renderStatusLine(dst, "error", colors.ErrorText, message, "")
}

func (f *defaultFormatter) begin(w *bufio.Writer, _ Report, hasVisibleDiagnostics bool) error {
	if hasVisibleDiagnostics {
		return w.WriteByte('\n')
	}
	return nil
}

func (f *defaultFormatter) diagnostic(w *bufio.Writer, view diagnosticView) error {
	renderDefaultDiagnostic(w, view, f.colors)
	return nil
}

func (f *defaultFormatter) finish(w *bufio.Writer, report Report) error {
	// Preserve the existing timing boundary: diagnostics are flushed to the
	// real destination before the completed-status duration is measured.
	if err := w.Flush(); err != nil {
		return err
	}
	elapsed := time.Duration(0)
	if !report.summary.StartedAt.IsZero() {
		elapsed = time.Since(report.summary.StartedAt)
	}
	renderSummary(w, report, elapsed, f.colors)
	return nil
}

func renderSummary(w *bufio.Writer, report Report, elapsed time.Duration, colors colorScheme) {
	if report.outcome.Kind == OutcomePassed && report.summary.Files == 0 && report.counts.Errors == 0 && report.counts.Warnings == 0 {
		message := fmt.Sprintf("%s in %s", emptyMessage(report.mode), formatElapsed(elapsed))
		_ = renderStatusLine(w, "success", colors.SuccessText, message, "")
		return
	}

	subject := completedSubject(report.mode)
	fixes := ""
	if report.summary.FixedIssues > 0 {
		fixes = fmt.Sprintf(" after applying %d %s",
			report.summary.FixedIssues,
			pluralize(report.summary.FixedIssues, "fix", "fixes"),
		)
	}

	var label, message string
	var labelColor func(string, ...interface{}) string
	switch report.outcome.Kind {
	case OutcomeDiagnosticsFailed:
		label = "error"
		labelColor = colors.ErrorText
		message = fmt.Sprintf("%s failed with %s%s in %s",
			subject,
			findingSummary(report, colors),
			fixes,
			formatElapsed(elapsed),
		)
	case OutcomeWarningLimitExceeded:
		label = "error"
		labelColor = colors.ErrorText
		message = fmt.Sprintf("%s failed%s in %s: %s %s exceeded the configured limit of %d",
			subject,
			fixes,
			formatElapsed(elapsed),
			colors.WarnText("%d", report.counts.Warnings),
			pluralize(report.counts.Warnings, "warning", "warnings"),
			report.outcome.WarningLimit,
		)
	default:
		label = "success"
		labelColor = colors.SuccessText
		warningSummary := ""
		if report.counts.Warnings > 0 {
			warningSummary = fmt.Sprintf(" with %s %s",
				colors.WarnText("%d", report.counts.Warnings),
				pluralize(report.counts.Warnings, "warning", "warnings"),
			)
		}
		message = fmt.Sprintf("%s passed%s%s in %s",
			subject,
			warningSummary,
			fixes,
			formatElapsed(elapsed),
		)
	}

	_ = renderStatusLine(w, label, labelColor, message, colors.DimText("%s", summaryDetails(report)))
}

func renderStatusLine(
	w io.Writer,
	label string,
	labelColor func(string, ...interface{}) string,
	message string,
	details string,
) error {
	padding := strings.Repeat(" ", statusLabelWidth-len(label)+1)
	if details == "" {
		_, err := fmt.Fprintf(w, "%s%s%s\n", labelColor("%s", label), padding, message)
		return err
	}
	_, err := fmt.Fprintf(w, "%s%s%s %s\n", labelColor("%s", label), padding, message, details)
	return err
}

func startMessage(mode Mode) string {
	switch mode {
	case ModeLintAndTypeCheck:
		return "Linting and type checking..."
	case ModeTypeCheckOnly:
		return "Type checking..."
	default:
		return "Linting..."
	}
}

func ongoingAction(mode Mode) string {
	switch mode {
	case ModeLintAndTypeCheck:
		return "Linting and type checking"
	case ModeTypeCheckOnly:
		return "Type checking"
	default:
		return "Linting"
	}
}

func completedSubject(mode Mode) string {
	switch mode {
	case ModeLintAndTypeCheck:
		return "Lint and type check"
	case ModeTypeCheckOnly:
		return "Type check"
	default:
		return "Lint"
	}
}

func emptyMessage(mode Mode) string {
	switch mode {
	case ModeLintAndTypeCheck:
		return "No files to lint or type check"
	case ModeTypeCheckOnly:
		return "No files to type check"
	default:
		return "No files to lint"
	}
}

func findingSummary(report Report, colors colorScheme) string {
	findings := make([]string, 0, 3)
	appendFinding := func(
		count int,
		singular string,
		plural string,
		countColor func(string, ...interface{}) string,
	) {
		if count > 0 {
			findings = append(findings, fmt.Sprintf("%s %s",
				countColor("%d", count),
				pluralize(count, singular, plural),
			))
		}
	}

	switch report.mode {
	case ModeLintAndTypeCheck:
		appendFinding(report.counts.LintErrors, "lint error", "lint errors", colors.ErrorText)
		appendFinding(report.counts.TypeErrors, "TypeScript error", "TypeScript errors", colors.ErrorText)
	case ModeTypeCheckOnly:
		appendFinding(report.counts.TypeErrors, "TypeScript error", "TypeScript errors", colors.ErrorText)
	default:
		appendFinding(report.counts.Errors, "error", "errors", colors.ErrorText)
	}
	appendFinding(report.counts.Warnings, "warning", "warnings", colors.WarnText)

	switch len(findings) {
	case 0:
		return "no reported problems"
	case 1:
		return findings[0]
	case 2:
		return findings[0] + " and " + findings[1]
	default:
		return strings.Join(findings[:len(findings)-1], ", ") + ", and " + findings[len(findings)-1]
	}
}

func summaryDetails(report Report) string {
	details := []string{
		fmt.Sprintf("%d %s", report.summary.Files, pluralize(report.summary.Files, "file", "files")),
	}
	if report.mode != ModeTypeCheckOnly {
		details = append(details,
			fmt.Sprintf("%d %s", report.summary.Rules, pluralize(report.summary.Rules, "rule", "rules")),
		)
	}
	details = append(details,
		fmt.Sprintf("%d %s", report.summary.Threads, pluralize(report.summary.Threads, "thread", "threads")),
	)
	return "(" + strings.Join(details, ", ") + ")"
}

func formatElapsed(elapsed time.Duration) string {
	if elapsed < time.Millisecond {
		return "<1ms"
	}
	return elapsed.Round(time.Millisecond).String()
}

func pluralize(count int, singular, plural string) string {
	if count == 1 {
		return singular
	}
	return plural
}

func renderDefaultDiagnostic(w *bufio.Writer, view diagnosticView, colors colorScheme) {
	diagnostic := view.raw
	diagnosticStart := diagnostic.Range.Start
	diagnosticEnd := diagnostic.Range.End
	diagnosticStartLine := view.start.line
	diagnosticStartColumn := view.start.column
	diagnosticEndLine := view.end.line

	lineMap := diagnostic.Source.lineStarts
	text := diagnostic.Source.text

	codeboxStartLine := max(diagnosticStartLine-1, 0)
	codeboxEndLine := min(diagnosticEndLine+1, len(lineMap)-1)
	codeboxStart := lineMap[codeboxStartLine]
	codeboxEnd := len(text)
	if codeboxEndLine != len(lineMap)-1 {
		codeboxEnd = lineMap[codeboxEndLine+1] - 1
	}

	w.WriteByte(' ')
	w.WriteString(colors.RuleName(" %s ", diagnostic.RuleName))
	w.WriteString(" — ")

	severityColor := colors.WarnText
	if diagnostic.Severity == SeverityError {
		severityColor = colors.ErrorText
	}
	w.WriteString(severityColor("[%s] ", diagnostic.Severity.String()))

	messageLineStart := 0
	for i, char := range diagnostic.Message {
		if char == '\n' {
			w.WriteString(diagnostic.Message[messageLineStart : i+1])
			messageLineStart = i + 1
			if diagnostic.PreFormatted {
				w.WriteString("  ")
			} else {
				w.WriteString("    ")
				w.WriteString(colors.BorderText("│"))
				w.WriteString(strings.Repeat(" ", len(diagnostic.RuleName)+1))
			}
		}
	}
	if messageLineStart <= len(diagnostic.Message) {
		w.WriteString(diagnostic.Message[messageLineStart:])
	}

	w.WriteString("\n  ")
	w.WriteString(colors.BorderText("╭─┴──────────("))
	w.WriteByte(' ')
	location := fmt.Sprintf("%s:%d:%d", view.relativePath, diagnosticStartLine+1, diagnosticStartColumn+1)
	w.WriteString(colors.FileName("%s", location))
	w.WriteByte(' ')
	w.WriteString(colors.BorderText(")─────"))
	w.WriteByte('\n')

	indentSize := math.MaxInt
	line := codeboxStartLine
	lineIndentCalculated := false
	lastNonSpaceByteIndex := -1

	numLines := codeboxEndLine - codeboxStartLine + 1
	lineStarts := make([]int, numLines)
	lineEnds := make([]int, numLines)

	// Preserve the established code-frame rendering contract: source locations
	// follow the producer's ECMAScript line map, while displayed content advances
	// only at LF. This intentionally retains legacy CR/U+2028/U+2029 behavior.
	codeboxText := text[codeboxStart:codeboxEnd]
	for offset := 0; offset < len(codeboxText); {
		char, size := utf8.DecodeRuneInString(codeboxText[offset:])
		current := codeboxStart + offset
		next := current + size
		offset += size

		if char == '\n' {
			if line != codeboxEndLine {
				lineIndentCalculated = false
				lineEnds[line-codeboxStartLine] = max(lastNonSpaceByteIndex-lineMap[line], 0)
				lastNonSpaceByteIndex = -1
				line++
			}
			continue
		}

		if !lineIndentCalculated && !unicode.IsSpace(char) {
			lineIndentCalculated = true
			lineStarts[line-codeboxStartLine] = max(current-lineMap[line], 0)
			indentSize = min(indentSize, lineStarts[line-codeboxStartLine])
		}
		if lineIndentCalculated && !unicode.IsSpace(char) {
			lastNonSpaceByteIndex = next
		}
	}
	if line == codeboxEndLine {
		lineEnds[line-codeboxStartLine] = max(lastNonSpaceByteIndex-lineMap[line], 0)
	}
	if indentSize == math.MaxInt {
		indentSize = 0
	}

	diagnosticHighlightActive := false
	lastLineNumber := strconv.Itoa(codeboxEndLine + 1)
	shouldFold := codeboxEndLine-codeboxStartLine >= 4

	for line := codeboxStartLine; line <= codeboxEndLine; line++ {
		if shouldFold && codeboxStartLine+1 < line && line < codeboxEndLine-1 {
			w.WriteString("  ")
			w.WriteString(colors.BorderText("│ "))
			foldDots := strings.Repeat(".", len(lastLineNumber))
			w.WriteString(colors.DimText("%s", foldDots))
			w.WriteString(colors.BorderText(" │"))
			w.WriteByte('\n')

			line = codeboxEndLine - 1
			diagnosticHighlightActive = diagnosticStart < lineMap[line] && diagnosticEnd >= lineMap[line]
		}

		w.WriteString("  ")
		w.WriteString(colors.BorderText("│ "))
		if line == codeboxEndLine {
			w.WriteString(colors.DimText("%s", lastLineNumber))
		} else {
			number := strconv.Itoa(line + 1)
			if len(number) < len(lastLineNumber) {
				w.WriteByte(' ')
			}
			w.WriteString(colors.DimText("%s", number))
		}
		w.WriteString(colors.BorderText(" │"))
		w.WriteString("  ")

		lineTextStart := lineMap[line] + indentSize
		underlineStart := max(lineTextStart, lineMap[line]+lineStarts[line-codeboxStartLine])
		underlineEnd := underlineStart
		lineTextEnd := max(lineMap[line]+lineEnds[line-codeboxStartLine], lineTextStart)

		if diagnosticHighlightActive {
			underlineEnd = lineTextEnd
		} else if lineMap[line] <= diagnosticStart && (line == len(lineMap)-1 || diagnosticStart < lineMap[line+1]) {
			underlineStart = min(max(lineTextStart, diagnosticStart), lineTextEnd)
			underlineEnd = lineTextEnd
			diagnosticHighlightActive = true
		}
		if lineMap[line] <= diagnosticEnd && (line == len(lineMap)-1 || diagnosticEnd < lineMap[line+1]) {
			underlineEnd = min(max(underlineStart, diagnosticEnd), lineTextEnd)
			diagnosticHighlightActive = false
		}

		if underlineStart != underlineEnd {
			w.WriteString(text[lineTextStart:underlineStart])
			w.WriteString(severityColor("%s", text[underlineStart:underlineEnd]))
			w.WriteString(text[underlineEnd:lineTextEnd])
		} else if lineTextStart != lineTextEnd {
			w.WriteString(text[lineTextStart:lineTextEnd])
		}

		w.WriteByte('\n')
	}
	w.WriteString("  ")
	w.WriteString(colors.BorderText("╰────────────────────────────────"))
	w.WriteString("\n\n")
}
