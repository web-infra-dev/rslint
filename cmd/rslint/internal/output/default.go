package output

import (
	"bufio"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type defaultFormatter struct {
	colors        colorScheme
	suppressEmpty bool
	hasVisible    bool
}

func (f *defaultFormatter) begin(w *bufio.Writer, _ Report, hasVisibleDiagnostics bool) error {
	f.hasVisible = hasVisibleDiagnostics
	if hasVisibleDiagnostics {
		return w.WriteByte('\n')
	}
	return nil
}

func (f *defaultFormatter) diagnostic(w *bufio.Writer, view diagnosticView) error {
	renderDefaultDiagnostic(w, view, f.colors)
	return nil
}

func (f *defaultFormatter) fileWarning(w *bufio.Writer, view fileWarningView) error {
	fmt.Fprintf(w, "%s: %s %s\n", f.colors.WarnText("warning"), f.colors.FileName("%s", view.relativePath), view.raw.Message)
	return nil
}

func (f *defaultFormatter) finish(w *bufio.Writer, report Report) error {
	if f.suppressEmpty && !f.hasVisible {
		return nil
	}
	// Preserve the existing timing boundary: diagnostics are flushed to the
	// real destination before the summary duration is measured.
	if err := w.Flush(); err != nil {
		return err
	}
	elapsed := time.Duration(0)
	if !report.metadata.StartedAt.IsZero() {
		elapsed = time.Since(report.metadata.StartedAt).Round(time.Millisecond)
	}
	renderSummary(w, report, elapsed, f.colors)
	return nil
}

func renderSummary(w *bufio.Writer, report Report, elapsed time.Duration, colors colorScheme) {
	errorCountText := func(count int) string {
		if count == 0 {
			return colors.SuccessText("%d", count)
		}
		return colors.ErrorText("%d", count)
	}

	warningColor := colors.WarnText
	if report.counts.Warnings == 0 {
		warningColor = colors.SuccessText
	}

	var errorsSummary string
	switch report.metadata.Mode {
	case ModeTypeCheckOnly:
		errorsSummary = fmt.Sprintf("%s %s",
			errorCountText(report.counts.TypeErrors),
			pluralize(report.counts.TypeErrors, "type error", "type errors"),
		)
	case ModeLintAndTypeCheck:
		errorsSummary = fmt.Sprintf("%s %s, %s %s",
			errorCountText(report.counts.LintErrors),
			pluralize(report.counts.LintErrors, "lint error", "lint errors"),
			errorCountText(report.counts.TypeErrors),
			pluralize(report.counts.TypeErrors, "type error", "type errors"),
		)
	default:
		errorsSummary = fmt.Sprintf("%s %s",
			errorCountText(report.counts.Errors),
			pluralize(report.counts.Errors, "error", "errors"),
		)
	}

	if report.metadata.Mode == ModeTypeCheckOnly {
		details := fmt.Sprintf("(type-checked %d %s in %v using %d %s)",
			report.metadata.TypeCheckedFiles,
			pluralize(report.metadata.TypeCheckedFiles, "file", "files"),
			elapsed,
			report.metadata.Threads,
			pluralize(report.metadata.Threads, "thread", "threads"),
		)
		fmt.Fprintf(w, "Found %s %s\n", errorsSummary, colors.DimText("%s", details))
		return
	}

	details := fmt.Sprintf("(linted %d %s with %d %s",
		report.metadata.LintedFiles,
		pluralize(report.metadata.LintedFiles, "file", "files"),
		report.metadata.Rules,
		pluralize(report.metadata.Rules, "rule", "rules"),
	)
	if report.metadata.Mode == ModeLintAndTypeCheck {
		details += fmt.Sprintf(", type-checked %d %s",
			report.metadata.TypeCheckedFiles,
			pluralize(report.metadata.TypeCheckedFiles, "file", "files"),
		)
	}
	details += fmt.Sprintf(" in %v using %d %s",
		elapsed,
		report.metadata.Threads,
		pluralize(report.metadata.Threads, "thread", "threads"),
	)
	if report.metadata.FixedIssues > 0 {
		details += fmt.Sprintf(", fixed %d %s",
			report.metadata.FixedIssues,
			pluralize(report.metadata.FixedIssues, "issue", "issues"),
		)
	}
	details += ")"

	fmt.Fprintf(w, "Found %s and %s %s %s\n",
		errorsSummary,
		warningColor("%d", report.counts.Warnings),
		pluralize(report.counts.Warnings, "warning", "warnings"),
		colors.DimText("%s", details),
	)
}

func pluralize(count int, singular, plural string) string {
	if count == 1 {
		return singular
	}
	return plural
}

func renderDefaultDiagnostic(w *bufio.Writer, view diagnosticView, colors colorScheme) {
	diagnostic := view.raw
	diagnosticStart := diagnostic.Range.Pos()
	diagnosticEnd := diagnostic.Range.End()
	diagnosticStartLine := view.start.line
	diagnosticStartColumn := view.start.column
	diagnosticEndLine := view.end.line

	lineMap := scanner.GetECMALineStarts(diagnostic.SourceFile)
	text := diagnostic.SourceFile.Text()

	codeboxStartLine := max(diagnosticStartLine-1, 0)
	codeboxEndLine := min(diagnosticEndLine+1, len(lineMap)-1)

	codeboxStart := int(lineMap[codeboxStartLine])
	var codeboxEnd int
	if codeboxEndLine == len(lineMap)-1 {
		codeboxEnd = len(text)
	} else {
		codeboxEnd = int(lineMap[codeboxEndLine+1]) - 1
	}

	w.WriteByte(' ')
	w.WriteString(colors.RuleName(" %s ", diagnostic.RuleName))
	w.WriteString(" — ")

	severityColor := colors.WarnText
	if diagnostic.Severity == rule.SeverityError {
		severityColor = colors.ErrorText
	}
	w.WriteString(severityColor("[%s] ", diagnostic.Severity.String()))

	messageLineStart := 0
	for i, char := range diagnostic.Message.Description {
		if char == '\n' {
			w.WriteString(diagnostic.Message.Description[messageLineStart : i+1])
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
	if messageLineStart <= len(diagnostic.Message.Description) {
		w.WriteString(diagnostic.Message.Description[messageLineStart:])
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

	codeboxText := text[codeboxStart:codeboxEnd]
	for i := 0; i < len(codeboxText); {
		char, size := utf8.DecodeRuneInString(codeboxText[i:])
		current := codeboxStart + i
		next := current + size
		i += size

		if char == '\n' {
			if line != codeboxEndLine {
				lineIndentCalculated = false
				lineEnds[line-codeboxStartLine] = lastNonSpaceByteIndex - int(lineMap[line])
				lastNonSpaceByteIndex = -1
				line++
			}
			continue
		}

		if !lineIndentCalculated && !unicode.IsSpace(char) {
			lineIndentCalculated = true
			lineStarts[line-codeboxStartLine] = current - int(lineMap[line])
			indentSize = min(indentSize, lineStarts[line-codeboxStartLine])
		}

		if lineIndentCalculated && !unicode.IsSpace(char) {
			lastNonSpaceByteIndex = next
		}
	}
	if line == codeboxEndLine {
		lineEnds[line-codeboxStartLine] = lastNonSpaceByteIndex - int(lineMap[line])
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
			diagnosticHighlightActive = diagnosticStart < int(lineMap[line]) && diagnosticEnd >= int(lineMap[line])
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

		lineTextStart := int(lineMap[line]) + indentSize
		underlineStart := max(lineTextStart, int(lineMap[line])+lineStarts[line-codeboxStartLine])
		underlineEnd := underlineStart
		lineTextEnd := max(int(lineMap[line])+lineEnds[line-codeboxStartLine], lineTextStart)

		if diagnosticHighlightActive {
			underlineEnd = lineTextEnd
		} else if int(lineMap[line]) <= diagnosticStart && (line == len(lineMap)-1 || diagnosticStart < int(lineMap[line+1])) {
			underlineStart = min(max(lineTextStart, diagnosticStart), lineTextEnd)
			underlineEnd = lineTextEnd
			diagnosticHighlightActive = true
		}
		if int(lineMap[line]) <= diagnosticEnd && (line == len(lineMap)-1 || diagnosticEnd < int(lineMap[line+1])) {
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
