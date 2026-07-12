package main

import (
	"bufio"
	"cmp"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/fatih/color"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/term"
)

// lintArgs is the parsed CLI flag set, decoupled from the global flag
// package so each entry point can build it: parseLintFlags from argv, and
// runCLI additionally from the IPC init-handshake payload.
type lintArgs struct {
	Init           bool
	Config         string
	ConfigStdin    bool // true → executeLintPipeline reads stdin as a config payload
	Fix            bool
	TypeCheck      bool
	TypeCheckOnly  bool
	TraceOut       string
	CpuprofOut     string
	SingleThreaded bool
	Format         string
	NoColor        bool
	ForceColor     bool
	// StdoutIsTTY is the TTY fact for the real output destination, reported
	// by the Node host via the IPC init payload (the Go process's own stdout
	// is an IPC pipe and says nothing about the user's terminal). False when
	// unknown (old peer, wasm build).
	StdoutIsTTY bool
	Quiet       bool
	MaxWarnings int
	StartTimeMs int64
	RuleFlags   []string
	// Positional args resolved into existing-dir vs file paths.
	AllowFiles []string
	AllowDirs  []string
	// EslintPlugins carries the {prefix, ruleNames} metadata for ESLint
	// plugins mounted via the config's object-form `plugins`. Used to register
	// placeholder rules so plugin rule names resolve; the live plugins run
	// in the Node worker.
	EslintPlugins []rslintconfig.EslintPluginEntry
}

// ColorScheme contains all the color functions for different UI elements
type ColorScheme struct {
	RuleName    func(format string, a ...interface{}) string
	FileName    func(format string, a ...interface{}) string
	ErrorText   func(format string, a ...interface{}) string
	SuccessText func(format string, a ...interface{}) string
	DimText     func(format string, a ...interface{}) string
	BoldText    func(format string, a ...interface{}) string
	BorderText  func(format string, a ...interface{}) string
	WarnText    func(format string, a ...interface{}) string
	// Reset is the trailing SGR reset sequence the summary lines emit (empty
	// when colors are off) — byte-identical to the inline
	// color.New().SprintFunc()("") emitters it replaces.
	Reset string
}

// newPinnedColor builds a color object pinned to the resolved decision.
// color.New seeds a per-object noColor switch from the NO_COLOR env at
// creation time, which would silently override the pipeline-entry decision
// (e.g. NO_COLOR set but --force-color given) — Enable/DisableColor makes
// the object follow color.NoColor unconditionally.
func newPinnedColor(attrs ...color.Attribute) *color.Color {
	c := color.New(attrs...)
	if color.NoColor {
		c.DisableColor()
	} else {
		c.EnableColor()
	}
	return c
}

// setupColors builds the SprintfFunc scheme used by the formatters. It is a
// pure factory: the on/off decision is owned by term.ResolveColorEnabled,
// applied once at pipeline entry and propagated here via color.NoColor.
func setupColors() *ColorScheme {
	return &ColorScheme{
		RuleName:    newPinnedColor(color.FgHiGreen).SprintfFunc(),
		FileName:    newPinnedColor(color.FgCyan, color.Italic).SprintfFunc(),
		ErrorText:   newPinnedColor(color.FgRed, color.Bold).SprintfFunc(),
		SuccessText: newPinnedColor(color.FgGreen, color.Bold).SprintfFunc(),
		DimText:     newPinnedColor(color.Faint).SprintfFunc(),
		BoldText:    newPinnedColor(color.Bold).SprintfFunc(),
		BorderText:  newPinnedColor(color.Faint).SprintfFunc(),
		WarnText:    newPinnedColor(color.FgYellow).SprintfFunc(),
		Reset:       newPinnedColor().SprintFunc()(""),
	}
}

// reportSyntacticErrors renders syntax errors with code snippets (like tsc --pretty).
// Returns true if syntactic errors were found and reported.
func reportSyntacticErrors(err error, w *bufio.Writer, comparePathOptions tspath.ComparePathsOptions) bool {
	var syntacticErr *utils.SyntacticError
	if !errors.As(err, &syntacticErr) {
		return false
	}
	rendered := false
	for _, d := range syntacticErr.Diagnostics {
		if d.File() == nil {
			continue
		}
		diag := rule.RuleDiagnostic{
			RuleName:   fmt.Sprintf("TypeScript(TS%d)", d.Code()),
			SourceFile: d.File(),
			FilePath:   d.File().FileName(),
			Range:      d.Loc(),
			Message:    rule.RuleMessage{Description: d.String()},
			Severity:   rule.SeverityError,
		}
		printDiagnosticDefault(diag, w, comparePathOptions)
		rendered = true
	}
	w.Flush()
	return rendered
}

func printDiagnostic(d rule.RuleDiagnostic, w *bufio.Writer, comparePathOptions tspath.ComparePathsOptions, format string, gitlabState *gitlabReportState) {
	switch format {
	case "default":
		printDiagnosticDefault(d, w, comparePathOptions)
	case "jsonline":
		printDiagnosticJsonLine(d, w, comparePathOptions)
	case "github":
		printDiagnosticGitHub(d, w, comparePathOptions)
	case "gitlab":
		printDiagnosticGitLab(d, w, comparePathOptions, gitlabState)
	default:
		panic("not supported format " + format)
	}
}

// gitlabReportState tracks the streaming state needed to emit the gitlab
// format as a single JSON array (one open bracket, comma-separated entries,
// one close bracket) while diagnostics are still printed one at a time, plus
// a fingerprint occurrence counter so two diagnostics that would otherwise
// hash identically (rare) don't collapse into one entry in the GitLab MR
// widget, which merges issues sharing a fingerprint.
type gitlabReportState struct {
	wroteFirst   bool
	fingerprints map[string]int
}

func newGitlabReportState() *gitlabReportState {
	return &gitlabReportState{fingerprints: make(map[string]int)}
}

// finish closes the JSON array. Must be called exactly once after all
// diagnostics have been printed.
func (s *gitlabReportState) finish(w *bufio.Writer) {
	if s.wroteFirst {
		w.WriteString("]\n")
	} else {
		w.WriteString("[]\n")
	}
}

// gitlabFingerprint derives a stable identifier for a diagnostic from its
// location and content. md5 is used only to derive a short opaque key (no
// cryptographic resistance is needed), and the seen map breaks ties
// deterministically instead of the random salt some other gitlab formatters
// use, which keeps report output reproducible across runs.
func gitlabFingerprint(seen map[string]int, filePath, ruleName, message string, startLine, startColumn, endLine, endColumn int) string {
	input := fmt.Sprintf("%s:%s:%s:%d:%d:%d:%d", filePath, ruleName, message, startLine, startColumn, endLine, endColumn)
	digest := func(s string) string {
		sum := md5.Sum([]byte(s)) //nolint:gosec // opaque identifier, not a security boundary
		return hex.EncodeToString(sum[:])
	}

	fingerprint := digest(input)
	if n, ok := seen[fingerprint]; ok {
		seen[fingerprint] = n + 1
		return digest(input + ":" + strconv.Itoa(n))
	}
	seen[fingerprint] = 1
	return fingerprint
}

// print as [GitLab Code Quality report](https://docs.gitlab.com/ci/testing/code_quality/)
// format: a single JSON array of issues, suitable for the `codequality`
// report artifact consumed by GitLab CI merge request widgets.
func printDiagnosticGitLab(d rule.RuleDiagnostic, w *bufio.Writer, comparePathOptions tspath.ComparePathsOptions, state *gitlabReportState) {
	diagnosticStart := d.Range.Pos()
	diagnosticEnd := d.Range.End()

	startLine, startColumn := scanner.GetECMALineAndUTF16CharacterOfPosition(d.SourceFile, diagnosticStart)
	endLine, endColumn := scanner.GetECMALineAndUTF16CharacterOfPosition(d.SourceFile, diagnosticEnd)

	filePath := tspath.ConvertToRelativePath(d.FilePath, comparePathOptions)

	var severity string
	switch d.Severity {
	case rule.SeverityError:
		severity = "major"
	case rule.SeverityWarning:
		severity = "minor"
	default:
		severity = "info"
	}

	type gitlabPosition struct {
		Line   int `json:"line"`
		Column int `json:"column"`
	}
	type gitlabPositions struct {
		Begin gitlabPosition `json:"begin"`
		End   gitlabPosition `json:"end"`
	}
	type gitlabLines struct {
		Begin int `json:"begin"`
		End   int `json:"end"`
	}
	type gitlabLocation struct {
		Path      string          `json:"path"`
		Lines     gitlabLines     `json:"lines"`
		Positions gitlabPositions `json:"positions"`
	}
	type gitlabIssue struct {
		Description string         `json:"description"`
		CheckName   string         `json:"check_name"`
		Fingerprint string         `json:"fingerprint"`
		Severity    string         `json:"severity"`
		Location    gitlabLocation `json:"location"`
	}

	beginLine, beginColumn := startLine+1, int(startColumn)+1
	endLineNum, endColumnNum := endLine+1, int(endColumn)+1

	issue := gitlabIssue{
		Description: d.Message.Description,
		CheckName:   d.RuleName,
		Fingerprint: gitlabFingerprint(state.fingerprints, filePath, d.RuleName, d.Message.Description, beginLine, beginColumn, endLineNum, endColumnNum),
		Severity:    severity,
		Location: gitlabLocation{
			Path: filePath,
			Lines: gitlabLines{
				Begin: beginLine,
				End:   endLineNum,
			},
			Positions: gitlabPositions{
				Begin: gitlabPosition{Line: beginLine, Column: beginColumn},
				End:   gitlabPosition{Line: endLineNum, Column: endColumnNum},
			},
		},
	}

	jsonBytes, err := json.Marshal(issue)
	if err != nil {
		// gitlabIssue contains only strings and ints, so Marshal cannot fail
		// in practice; skip rather than risk corrupting the JSON array.
		return
	}

	if state.wroteFirst {
		w.WriteByte(',')
	} else {
		w.WriteByte('[')
		state.wroteFirst = true
	}
	w.Write(jsonBytes)
}

// print as [Workflow commands for GitHub Actions](https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands) format
func printDiagnosticGitHub(d rule.RuleDiagnostic, w *bufio.Writer, comparePathOptions tspath.ComparePathsOptions) {
	var severity string
	switch d.Severity {
	case rule.SeverityError:
		severity = "error"
	case rule.SeverityWarning:
		severity = "warning"
	default:
		severity = "notice"
	}

	diagnosticStart := d.Range.Pos()
	diagnosticEnd := d.Range.End()

	startLine, startColumn := scanner.GetECMALineAndUTF16CharacterOfPosition(d.SourceFile, diagnosticStart)
	endLine, endColumn := scanner.GetECMALineAndUTF16CharacterOfPosition(d.SourceFile, diagnosticEnd)

	filePath := tspath.ConvertToRelativePath(d.FilePath, comparePathOptions)
	output := fmt.Sprintf(
		"::%s file=%s,line=%d,endLine=%d,col=%d,endColumn=%d,title=%s::%s\n",
		severity,
		escapeProperty(filePath),
		startLine+1,
		endLine+1,
		int(startColumn)+1,
		int(endColumn)+1,
		d.RuleName,
		escapeData(d.Message.Description),
	)
	w.WriteString(output)
}

func pluralize(count int, singular, plural string) string {
	if count == 1 {
		return singular
	}
	return plural
}

func escapeData(str string) string {
	// https://github.com/biomejs/biome/blob/4416573f4d709047a28407d99381810b7bc7dcc7/crates/biome_diagnostics/src/display_github.rs#L85C4-L85C15
	str = strings.ReplaceAll(str, "%", "%25")
	str = strings.ReplaceAll(str, "\r", "%0D")
	str = strings.ReplaceAll(str, "\n", "%0A")
	return str
}
func escapeProperty(str string) string {
	// https://github.com/biomejs/biome/blob/4416573f4d709047a28407d99381810b7bc7dcc7/crates/biome_diagnostics/src/display_github.rs#L103
	str = strings.ReplaceAll(str, "%", "%25")
	str = strings.ReplaceAll(str, "\r", "%0D")
	str = strings.ReplaceAll(str, "\n", "%0A")
	str = strings.ReplaceAll(str, ":", "%3A")
	str = strings.ReplaceAll(str, ",", "%2C")
	return str
}

// print as [jsonline](https://jsonlines.org/) format which can be used for lsp
func printDiagnosticJsonLine(d rule.RuleDiagnostic, w *bufio.Writer, comparePathOptions tspath.ComparePathsOptions) {
	diagnosticStart := d.Range.Pos()
	diagnosticEnd := d.Range.End()

	startLine, startColumn := scanner.GetECMALineAndUTF16CharacterOfPosition(d.SourceFile, diagnosticStart)
	endLine, endColumn := scanner.GetECMALineAndUTF16CharacterOfPosition(d.SourceFile, diagnosticEnd)

	type Location struct {
		Line   int `json:"line"`
		Column int `json:"column"`
	}

	type Range struct {
		Start Location `json:"start"`
		End   Location `json:"end"`
	}

	type Diagnostic struct {
		RuleName string `json:"ruleName"`
		Message  string `json:"message"`
		FilePath string `json:"filePath"`
		Range    Range  `json:"range"`
		Severity string `json:"severity"`
	}

	diagnostic := Diagnostic{
		RuleName: d.RuleName,
		Message:  d.Message.Description,
		FilePath: tspath.ConvertToRelativePath(d.FilePath, comparePathOptions),
		Range: Range{
			Start: Location{
				Line:   startLine + 1, // Convert to 1-based indexing
				Column: int(startColumn) + 1,
			},
			End: Location{
				Line:   endLine + 1,
				Column: int(endColumn) + 1,
			},
		},
		Severity: d.Severity.String(),
	}

	jsonBytes, err := json.Marshal(diagnostic)
	if err != nil {
		type ErrorObject struct {
			Error string `json:"error"`
		}
		errorObject := ErrorObject{Error: fmt.Sprintf("Failed to marshal diagnostic: %s", err)}

		errorBytes, _ := json.Marshal(errorObject) //nolint:errchkjson
		w.Write(errorBytes)
		w.WriteByte('\n')
		return
	}

	w.Write(jsonBytes)
	w.WriteByte('\n')
}

// print a normal logger
func printDiagnosticDefault(d rule.RuleDiagnostic, w *bufio.Writer, comparePathOptions tspath.ComparePathsOptions) {
	colors := setupColors()

	diagnosticStart := d.Range.Pos()
	diagnosticEnd := d.Range.End()

	diagnosticStartLine, diagnosticStartColumn := scanner.GetECMALineAndUTF16CharacterOfPosition(d.SourceFile, diagnosticStart)
	diagnosticEndline, _ := scanner.GetECMALineAndUTF16CharacterOfPosition(d.SourceFile, diagnosticEnd)

	lineMap := scanner.GetECMALineStarts(d.SourceFile)
	text := d.SourceFile.Text()

	codeboxStartLine := max(diagnosticStartLine-1, 0)
	codeboxEndLine := min(diagnosticEndline+1, len(lineMap)-1)

	codeboxStart := int(lineMap[codeboxStartLine])
	var codeboxEnd int
	if codeboxEndLine == len(lineMap)-1 {
		codeboxEnd = len(text)
	} else {
		codeboxEnd = int(lineMap[codeboxEndLine+1]) - 1
	}

	// Rule name with conditional coloring
	w.WriteByte(' ')
	w.WriteString(colors.RuleName(" %s ", d.RuleName))
	w.WriteString(" — ")

	// Severity level with conditional coloring
	severityColor := colors.WarnText
	if d.Severity == rule.SeverityError {
		severityColor = colors.ErrorText
	}
	w.WriteString(severityColor("[%s] ", d.Severity.String()))

	// Message handling — multi-line continuation:
	// - PreFormatted (e.g. tsc diagnostics): 2-space indent, message already has indentation
	// - Lint rules: │ border aligned after rule name
	messageLineStart := 0
	for i, char := range d.Message.Description {
		if char == '\n' {
			w.WriteString(d.Message.Description[messageLineStart : i+1])
			messageLineStart = i + 1
			if d.PreFormatted {
				w.WriteString("  ")
			} else {
				w.WriteString("    ")
				w.WriteString(colors.BorderText("│"))
				w.WriteString(strings.Repeat(" ", len(d.RuleName)+1))
			}
		}
	}
	if messageLineStart <= len(d.Message.Description) {
		w.WriteString(d.Message.Description[messageLineStart:len(d.Message.Description)])
	}

	// File path with conditional coloring
	w.WriteString("\n  ")
	w.WriteString(colors.BorderText("╭─┴──────────("))
	w.WriteByte(' ')
	filePath := tspath.ConvertToRelativePath(d.FilePath, comparePathOptions)
	location := fmt.Sprintf("%s:%d:%d", filePath, diagnosticStartLine+1, diagnosticStartColumn+1)
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

	// Iterate by runes to correctly handle multi-byte UTF-8 characters.
	// Use utf8.DecodeRuneInString to get the true byte width of each rune
	// (including invalid UTF-8 bytes, which decode as RuneError with size=1).
	// `for _, char := range str` plus `utf8.RuneLen(char)` is unsafe here
	// because invalid bytes yield RuneError (U+FFFD) and RuneLen returns 3
	// (the encoded length of U+FFFD), throwing off the byte counter and
	// eventually slicing past len(text).
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
	// If no non-space content was seen anywhere in the codebox,
	// `indentSize` was never updated from the math.MaxInt sentinel.
	// Clamping to 0 prevents `lineMap[line] + indentSize` from overflowing
	// int in the render loop below (which would wrap to a large negative
	// number and slice out of bounds).
	if indentSize == math.MaxInt {
		indentSize = 0
	}

	diagnosticHighlightActive := false
	lastLineNumber := strconv.Itoa(codeboxEndLine + 1)
	// Fold when codebox spans 5+ lines: show first 2 + "..." + last 2 (same as tsc)
	shouldFold := codeboxEndLine-codeboxStartLine >= 4

	for line := codeboxStartLine; line <= codeboxEndLine; line++ {
		// Fold: skip middle lines, show first 2 and last 2
		if shouldFold && codeboxStartLine+1 < line && line < codeboxEndLine-1 {
			w.WriteString("  ")
			w.WriteString(colors.BorderText("│ "))
			foldDots := strings.Repeat(".", len(lastLineNumber))
			w.WriteString(colors.DimText("%s", foldDots))
			w.WriteString(colors.BorderText(" │"))
			w.WriteByte('\n')

			line = codeboxEndLine - 1
			// Update highlight state for the jumped-to line
			diagnosticHighlightActive = diagnosticStart < int(lineMap[line]) && diagnosticEnd >= int(lineMap[line])
			// Fall through to render this line
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

// repeatedFlag collects multiple values for the same flag (e.g. --rule used multiple times).
type repeatedFlag []string

func (f *repeatedFlag) String() string     { return strings.Join(*f, ", ") }
func (f *repeatedFlag) Set(v string) error { *f = append(*f, v); return nil }

const usage = `🚀 Rslint - Rocket Speed Linter

Usage:
  rslint [OPTIONS] [files...]

Options:
  --init                Initialize a default config in the current directory.
  -c, --config PATH     Which rslint config file to use.
  --format FORMAT       Output format: default | jsonline | github | gitlab
  --fix                 Automatically fix problems
  --type-check          Enable TypeScript type checking
  --type-check-only     Run only TypeScript type checking (skip all lint rules)
  --no-color            Disable colored output
  --force-color         Force colored output
  --quiet               Report errors only
  --max-warnings Int    Number of warnings to trigger nonzero exit code
  --rule RULE           Rule override, e.g. 'no-console: error' (repeatable)
  -h, --help            Show help
`

// groupDiagsByFile groups a flat slice of diagnostics by their source file name.
func groupDiagsByFile(diags []rule.RuleDiagnostic) map[string][]rule.RuleDiagnostic {
	m := make(map[string][]rule.RuleDiagnostic)
	for _, d := range diags {
		f := d.FilePath
		m[f] = append(m[f], d)
	}
	return m
}

// remapDiagnosticTargetPaths keeps diagnostics in the caller's target path
// space when a TypeScript Program represents that target by another lexical or
// canonical source-file path. SourceFile remains unchanged because ranges and
// fixes are defined against its text; FilePath controls display and disk writes.
func remapDiagnosticTargetPaths(diags []rule.RuleDiagnostic, targetPathBySourcePath map[string]string, filesystems ...vfs.FS) {
	if len(targetPathBySourcePath) == 0 {
		return
	}
	var fsys vfs.FS
	if len(filesystems) > 0 {
		fsys = filesystems[0]
	}
	for i := range diags {
		targetPath := targetPathBySourcePath[diags[i].FilePath]
		if targetPath == "" {
			targetPath = targetPathBySourcePath[canonicalFilesystemPathID(diags[i].FilePath, fsys)]
		}
		if targetPath != "" {
			diags[i].FilePath = targetPath
		}
	}
}

type typeScriptDiagnosticDedupeKey struct {
	path     string
	ruleName string
	pos      int
	end      int
	message  string
}

// deduplicateTypeScriptDiagnostics joins the lint-target syntax path and the
// program-wide type-check path. A file governed by a config without a tsconfig
// can be parsed by the fallback Program while also belonging to another
// config's real Program; --type-check legitimately visits both, but the same
// TypeScript diagnostic must be reported once.
func deduplicateTypeScriptDiagnostics(
	diags []rule.RuleDiagnostic,
	fsys vfs.FS,
	preferredCallerTargets ...map[string]string,
) []rule.RuleDiagnostic {
	if len(diags) == 0 {
		return diags
	}
	var callerTargetByCanonicalPath map[string]string
	if len(preferredCallerTargets) > 0 {
		callerTargetByCanonicalPath = preferredCallerTargets[0]
	}
	if len(diags) == 1 {
		if strings.HasPrefix(diags[0].RuleName, "TypeScript(TS") {
			canonicalID := canonicalFilesystemPathID(diags[0].FilePath, fsys)
			if callerTarget := callerTargetByCanonicalPath[canonicalID]; callerTarget != "" {
				diags[0].FilePath = callerTarget
			}
		}
		return diags
	}
	bestIndex := make(map[typeScriptDiagnosticDedupeKey]int)
	keys := make([]typeScriptDiagnosticDedupeKey, len(diags))
	for i, diagnostic := range diags {
		if !strings.HasPrefix(diagnostic.RuleName, "TypeScript(TS") {
			continue
		}
		key := typeScriptDiagnosticDedupeKey{
			path:     canonicalFilesystemPathID(diagnostic.FilePath, fsys),
			ruleName: diagnostic.RuleName,
			pos:      diagnostic.Range.Pos(),
			end:      diagnostic.Range.End(),
			message:  diagnostic.Message.Description,
		}
		keys[i] = key
		current, exists := bestIndex[key]
		if !exists || preferTypeScriptDiagnostic(diagnostic, diags[current], callerTargetByCanonicalPath[key.path], fsys) {
			bestIndex[key] = i
		}
	}

	result := diags[:0]
	for i, diagnostic := range diags {
		if !strings.HasPrefix(diagnostic.RuleName, "TypeScript(TS") {
			result = append(result, diagnostic)
			continue
		}
		key := keys[i]
		if bestIndex[key] != i {
			continue
		}
		if callerTarget := callerTargetByCanonicalPath[key.path]; callerTarget != "" {
			diagnostic.FilePath = callerTarget
		}
		result = append(result, diagnostic)
	}
	return result
}

func preferTypeScriptDiagnostic(candidate rule.RuleDiagnostic, current rule.RuleDiagnostic, callerTarget string, fsys vfs.FS) bool {
	if callerTarget != "" {
		candidateIsCaller := exactFilesystemPathID(candidate.FilePath) == exactFilesystemPathID(callerTarget)
		currentIsCaller := exactFilesystemPathID(current.FilePath) == exactFilesystemPathID(callerTarget)
		if candidateIsCaller != currentIsCaller {
			return candidateIsCaller
		}
	}
	candidatePath := tspath.NormalizePath(candidate.FilePath)
	currentPath := tspath.NormalizePath(current.FilePath)
	if candidatePath != currentPath {
		return candidatePath < currentPath
	}
	return false
}

// applyFixPass applies auto-fixes for all files in diagnosticsByFile,
// writes fixed content to disk, and returns the number of issues fixed. Write
// failures are returned after all independent files have been attempted.
func applyFixPass(diagnosticsByFile map[string][]rule.RuleDiagnostic) (int, error) {
	fixed := 0
	fileNames := make([]string, 0, len(diagnosticsByFile))
	for fileName := range diagnosticsByFile {
		fileNames = append(fileNames, fileName)
	}
	sort.Strings(fileNames)

	var writeErrors []error
	for _, fileName := range fileNames {
		fileDiagnostics := diagnosticsByFile[fileName]
		var diagnosticsWithFixes []rule.RuleDiagnostic
		for _, d := range fileDiagnostics {
			if len(d.Fixes()) > 0 {
				diagnosticsWithFixes = append(diagnosticsWithFixes, d)
			}
		}
		if len(diagnosticsWithFixes) == 0 {
			continue
		}

		originalContent := diagnosticsWithFixes[0].SourceFile.Text()
		fixedContent, unapplied, wasFixed := linter.ApplyRuleFixes(originalContent, diagnosticsWithFixes)

		if wasFixed {
			err := os.WriteFile(fileName, []byte(fixedContent), 0644)
			if err != nil {
				writeErrors = append(writeErrors, fmt.Errorf("write fixed file %q: %w", fileName, err))
			} else {
				fixed += len(diagnosticsWithFixes) - len(unapplied)
			}
		}
	}
	return fixed, errors.Join(writeErrors...)
}

// allowFileWarningKind categorizes why a CLI-specified file won't have lint
// rules applied. These are Phase-1 (lint) concepts; Phase 2 (type-check) is
// not affected by either case (it runs program-wide regardless of CLI scope
// and rslint ignores).
type allowFileWarningKind int

const (
	allowFileNotFound allowFileWarningKind = iota
	allowFileIgnored
)

// allowFileWarning is the structured form of a "this CLI-specified file
// won't be linted" diagnostic. Returned by collectAllowFileWarnings so the
// emission policy (skip in --type-check-only, format with relative paths,
// etc.) can be unit-tested without capturing stderr.
type allowFileWarning struct {
	Path string // absolute, normalized — matches what was put in allowFiles
	Kind allowFileWarningKind
}

// formatAllowFileWarning renders an allowFileWarning for stderr emission,
// resolving the absolute path against comparePathOptions. Returns "" for
// unknown kinds so callers can safely ignore the empty case.
func formatAllowFileWarning(w allowFileWarning, opts tspath.ComparePathsOptions) string {
	rel := tspath.ConvertToRelativePath(w.Path, opts)
	switch w.Kind {
	case allowFileNotFound:
		return fmt.Sprintf("warning: %s was not found, skipping", rel)
	case allowFileIgnored:
		return fmt.Sprintf("warning: %s is ignored because of a matching ignore pattern", rel)
	}
	return ""
}

// collectAllowFileWarnings explains, for each CLI-specified file in
// allowFiles, why it won't be visited by Phase 1 (lint). Program membership
// is deliberately not consulted: lint targets are resolved before type-info
// binding, so a file outside every tsconfig can still be linted via fallback.
// Returns nil for empty allowFiles.
//
// This is a Phase-1 concern only. In --type-check-only mode the lint phase
// is skipped, so callers MUST gate emission on `!typeCheckOnly` — otherwise
// users see misleading messages like "is ignored because of a matching
// ignore pattern" while Phase 2 happily reports type errors for that same
// file. See website/docs/en/guide/type-checking.md ("type-check is not
// constrained by `files`/`ignores`").
func collectAllowFileWarnings(
	allowFiles []string,
	configMap map[string]rslintconfig.RslintConfig,
	rslintConfig rslintconfig.RslintConfig,
	currentDirectory string,
	useCaseSensitive bool,
	filesystems ...vfs.FS,
) []allowFileWarning {
	if len(allowFiles) == 0 {
		return nil
	}

	// Reuse owner resolution by directory when many explicit files share a
	// directory (for example, lint-staged invocations).
	type cachedConfig struct {
		cfgDir string
		cfg    rslintconfig.RslintConfig
	}
	dirConfigCache := make(map[string]*cachedConfig)
	var fsys vfs.FS
	if len(filesystems) > 0 {
		fsys = filesystems[0]
	}
	var configOwnerResolver *rslintconfig.ConfigOwnerResolver
	if configMap != nil {
		configOwnerResolver = rslintconfig.NewConfigOwnerResolver(configMap, fsys)
	}

	var out []allowFileWarning
	for _, f := range allowFiles {
		var cfgDir string
		var cfg rslintconfig.RslintConfig
		if configMap != nil {
			dir := tspath.GetDirectoryPath(f)
			cacheKey := exactFilesystemPathID(dir)
			cached, ok := dirConfigCache[cacheKey]
			if !ok {
				var foundDir string
				var foundCfg rslintconfig.RslintConfig
				foundDir, foundCfg = configOwnerResolver.Resolve(f)
				cached = &cachedConfig{cfgDir: foundDir, cfg: foundCfg}
				dirConfigCache[cacheKey] = cached
			}
			cfgDir = cached.cfgDir
			cfg = cached.cfg
		} else {
			cfgDir = currentDirectory
			cfg = rslintConfig
		}
		matchFile := f
		matchDir := cfgDir
		if fsys != nil && cfgDir != "" {
			canonicalPath := authoritativeFilesystemPath(f, fsys)
			matchFile = configPathForLintTarget(resolvedLintTarget{
				Path:           tspath.NormalizePath(f),
				CanonicalPath:  canonicalPath,
				OwnerConfigDir: cfgDir,
			}, fsys)
			matchDir = authoritativeFilesystemPath(cfgDir, fsys)
		}
		if rslintconfig.IsDefaultExcludedPath(matchFile, matchDir, useCaseSensitive) ||
			(cfg != nil && cfg.IsFileIgnored(matchFile, matchDir)) {
			out = append(out, allowFileWarning{Path: f, Kind: allowFileIgnored})
			continue
		}

		if _, err := os.Stat(f); err != nil {
			out = append(out, allowFileWarning{Path: f, Kind: allowFileNotFound})
			continue
		}
	}
	return out
}

// shouldShortCircuitOutput returns true when rslint should bail early
// without printing diagnostics or a summary. The short-circuit exists so
// that e.g. `rslint nonexistent-file.ts` returns 0 with no spurious output
// when Phase 1 visited zero files.
//
// Any type-check mode (`--type-check` or `--type-check-only`) must NOT take
// the short-circuit: Phase 2 runs program-wide and is not gated by the CLI
// Scope/PerProgramFilter that drives lintedFileCount, so lintedFileCount==0
// is a normal state in which Phase 2 may still have produced diagnostics.
// Short-circuiting there would silently drop type errors that the user
// explicitly asked for — see website/docs/en/guide/type-checking.md.
func shouldShortCircuitOutput(typeCheckOnly, typeCheck, scopeRestricted bool, lintedFileCount int32) bool {
	if typeCheckOnly || typeCheck {
		return false
	}
	return scopeRestricted && lintedFileCount == 0
}

// validateTypeCheckOnlyFlags rejects --type-check-only combined with flags
// whose semantics depend on the lint phase that this mode disables. Returns
// (0, "") when the combination is valid (or --type-check-only isn't set);
// otherwise returns (exitCode > 0, stderr message). Pulled out as a pure
// function so the policy can be exercised in unit tests.
func validateTypeCheckOnlyFlags(typeCheckOnly, fix bool, ruleFlags []string) (int, string) {
	if !typeCheckOnly {
		return 0, ""
	}
	if fix {
		return 2, "error: --fix cannot be combined with --type-check-only (no lint rules run, nothing to fix)"
	}
	if len(ruleFlags) > 0 {
		return 2, "error: --rule cannot be combined with --type-check-only (no lint rules run)"
	}
	return 0, ""
}

func cloneConfigMap(configMap map[string]rslintconfig.RslintConfig) map[string]rslintconfig.RslintConfig {
	if configMap == nil {
		return nil
	}
	cloned := make(map[string]rslintconfig.RslintConfig, len(configMap))
	for dir, cfg := range configMap {
		cloned[dir] = slices.Clone(cfg)
	}
	return cloned
}

// resolveStartTime returns the start time for timing output.
// If startTimeMs (epoch millis from the Node.js entry point) is positive,
// it is used so the reported duration covers end-to-end execution.
// Otherwise falls back to time.Now().
func resolveStartTime(startTimeMs int64) time.Time {
	if startTimeMs > 0 {
		return time.UnixMilli(startTimeMs)
	}
	return time.Now()
}

// parseLintFlags parses the lint CLI flags out of argv into a lintArgs.
// It uses a fresh FlagSet (not the global flag.CommandLine) so it is
// callable more than once per process, and ContinueOnError so a bad flag
// returns a fatal exit code instead of os.Exit-ing past caller cleanup.
// A non-zero fatalExitCode means the caller should return it immediately
// (the diagnostic was already printed to stderr).
func parseLintFlags(argv []string) (args lintArgs, help bool, fatalExitCode int) {
	fs := flag.NewFlagSet("rslint", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() { fmt.Fprint(os.Stderr, usage) }

	var ruleFlags repeatedFlag

	fs.StringVar(&args.Format, "format", "default", "output format")
	fs.StringVar(&args.Config, "config", "", "which rslint config to use")
	fs.StringVar(&args.Config, "c", "", "which rslint config to use")
	fs.BoolVar(&args.ConfigStdin, "config-stdin", false, "read config from stdin (used internally by JS config loader)")
	fs.BoolVar(&args.Init, "init", false, "initialize a default config in the current directory")
	fs.BoolVar(&args.Fix, "fix", false, "automatically fix problems")
	fs.BoolVar(&args.TypeCheck, "type-check", false, "enable TypeScript type checking")
	fs.BoolVar(&args.TypeCheckOnly, "type-check-only", false, "run only TypeScript type checking (skip all lint rules)")
	fs.BoolVar(&help, "help", false, "show help")
	fs.BoolVar(&help, "h", false, "show help")
	fs.BoolVar(&args.NoColor, "no-color", false, "disable colored output")
	fs.BoolVar(&args.ForceColor, "force-color", false, "force colored output")
	fs.BoolVar(&args.Quiet, "quiet", false, "report errors only")
	fs.IntVar(&args.MaxWarnings, "max-warnings", -1, "Number of warnings to trigger nonzero exit code")

	fs.StringVar(&args.TraceOut, "trace", "", "file to put trace to")
	fs.StringVar(&args.CpuprofOut, "cpuprof", "", "file to put cpu profiling to")
	fs.BoolVar(&args.SingleThreaded, "singleThreaded", false, "run in single threaded mode")
	fs.Int64Var(&args.StartTimeMs, "start-time", 0, "internal: epoch milliseconds from Node.js entry point")
	fs.Var(&ruleFlags, "rule", "rule override, e.g. 'no-console: error' (repeatable)")

	if err := fs.Parse(argv); err != nil {
		// ContinueOnError: fs already printed the diagnostic to stderr.
		return args, help, 2
	}
	args.RuleFlags = []string(ruleFlags)

	// --type-check-only implies --type-check and skips all lint rules.
	// Reject incompatible flag combinations before doing any work.
	if code, msg := validateTypeCheckOnlyFlags(args.TypeCheckOnly, args.Fix, args.RuleFlags); code != 0 {
		fmt.Fprintln(os.Stderr, msg)
		return args, help, code
	}
	if args.TypeCheckOnly {
		args.TypeCheck = true
	}

	// Collect file/directory arguments for targeted linting (e.g. rslint file1.ts src/)
	if positional := fs.Args(); len(positional) > 0 {
		for _, arg := range positional {
			absPath, err := filepath.Abs(arg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error resolving path %s: %v\n", arg, err)
				return lintArgs{}, help, 1
			}
			// NOTE: we intentionally do NOT call filepath.EvalSymlinks here.
			// EvalSymlinks resolves symlinks (macOS /tmp → /private/tmp) and
			// Windows 8.3 short names to long names, but the rest
			// of the pipeline (os.Getwd, TypeScript file names, configDir)
			// uses unresolved CWD-based paths. Resolving only file args would
			// create a format mismatch causing failures in gap file detection,
			// config matching, dir scoping, and gitignore checks.
			// Edge cases (e.g. user passes a symlink-resolved absolute path)
			// are handled by isFileAllowed's os.SameFile fallback in linter.go
			// and program file index realpath aliases during target binding.
			normalized := tspath.NormalizePath(absPath)
			info, statErr := os.Stat(absPath)
			if statErr == nil && info.IsDir() {
				args.AllowDirs = append(args.AllowDirs, normalized)
			} else {
				args.AllowFiles = append(args.AllowFiles, normalized)
			}
		}
	}

	return args, help, 0
}

// executeLintPipeline runs the full lint flow (config load → program build →
// lint target plan/fallback binding → lint → optional --fix loop → report) and
// returns the process exit code. Shared by the IPC entry (runCLI) and the wasm
// native fallback.
func executeLintPipeline(args lintArgs, ctx context.Context, dispatch linter.EslintPluginDispatcher) int {
	// Unpack into locals so the pipeline body below stays verbatim — only the
	// flag-parse front matter lives in parseLintFlags.
	init := args.Init
	config := args.Config
	configStdin := args.ConfigStdin
	fix := args.Fix
	typeCheck := args.TypeCheck
	typeCheckOnly := args.TypeCheckOnly
	traceOut := args.TraceOut
	cpuprofOut := args.CpuprofOut
	singleThreaded := args.SingleThreaded
	format := args.Format
	quiet := args.Quiet
	maxWarnings := args.MaxWarnings
	startTimeMs := args.StartTimeMs
	ruleFlags := args.RuleFlags
	allowFiles := args.AllowFiles
	allowDirs := args.AllowDirs

	// The single color decision for the whole run. It unconditionally
	// overwrites fatih/color's package-init value, whose isatty component
	// keyed off the Go process's own stdout — an IPC pipe in every native
	// lint run, never the user's terminal (its NO_COLOR/TERM components are
	// re-derived in the tiers below). Nothing after this line mutates
	// color.NoColor.
	forceColorEnv, forceColorEnvSet := os.LookupEnv("FORCE_COLOR")
	_, noColorEnvSet := os.LookupEnv("NO_COLOR")
	color.NoColor = !term.ResolveColorEnabled(term.ColorInputs{
		NoColorFlag:      args.NoColor,
		ForceColorFlag:   args.ForceColor,
		ForceColorEnv:    forceColorEnv,
		ForceColorEnvSet: forceColorEnvSet,
		NoColorEnvSet:    noColorEnvSet,
		GithubActionsEnv: os.Getenv("GITHUB_ACTIONS"),
		Term:             os.Getenv("TERM"),
		StdoutIsTTY:      args.StdoutIsTTY,
	})

	enableVirtualTerminalProcessing()
	timeBefore := resolveStartTime(startTimeMs)

	if traceOut != "" {
		f, err := os.Create(traceOut)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating trace file: %v\n", err)
			return 1
		}
		defer f.Close()
		trace.Start(f)
		defer trace.Stop()
	}
	if cpuprofOut != "" {
		f, err := os.Create(cpuprofOut)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating cpuprof file: %v\n", err)
			return 1
		}
		defer f.Close()
		err = pprof.StartCPUProfile(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error starting cpu profiling: %v\n", err)
			return 1
		}
		defer pprof.StopCPUProfile()
	}

	currentDirectory, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting current directory: %v\n", err)
		return 1
	}
	currentDirectory = tspath.NormalizePath(currentDirectory)
	workingDirectory := currentDirectory

	if init {
		if err := rslintconfig.InitDefaultConfig(currentDirectory); err != nil {
			fmt.Fprintf(os.Stderr, "error initializing config: %v\n", err)
			return 1
		}
		return 0
	}

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))

	// Run-scoped parse cache shared by every Program built in this pipeline
	// (initial build, gap fallback, --fix rebuilds). It is passed explicitly and
	// discarded after the invocation; no package-level cache is involved.
	parseCache := utils.NewParseCache()

	// Initialize rule registry with all available rules
	rslintconfig.RegisterAllRules()
	// Register placeholder rules for mounted ESLint plugins so their rule
	// names resolve (and route to the Node worker) instead of being dropped.
	rslintconfig.RegisterEslintPluginRules(args.EslintPlugins)
	var rslintConfig rslintconfig.RslintConfig

	// configMap holds per-directory configs for multi-config (monorepo) support.
	// Only populated in the configStdin path; nil otherwise (single-config mode).
	var configMap map[string]rslintconfig.RslintConfig

	// originalConfigDir maps each normalized configMap key back to the raw
	// configDirectory the JS host sent, so the eslint-plugin wire configKey
	// round-trips raw (byte-matching the worker's plugin map key). nil outside
	// multi-config mode (single-config / JSON configs never mount plugins).
	var originalConfigDir map[string]string
	var configTargetScopes map[string]rslintconfig.LintDiscoveryScope

	// Program-wide type checking builds every configured project. Plain linting
	// waits for target discovery and builds only the projects owned by configs
	// that govern at least one selected target.
	var realProgramSet lintProgramSet
	buildAllPrograms := typeCheck || typeCheckOnly
	needsLintTargets := !typeCheckOnly

	if configStdin {
		// Read config JSON from stdin (sent by JS config loader).
		// Read up to maxConfigSize+1 so we can detect truncation.
		const maxConfigSize = 50 << 20 // 50 MB
		data, err := io.ReadAll(io.LimitReader(os.Stdin, maxConfigSize+1))
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading config from stdin: %v\n", err)
			return 1
		}
		if len(data) > maxConfigSize {
			fmt.Fprintf(os.Stderr, "error: config from stdin exceeds maximum size of %d bytes\n", maxConfigSize)
			return 1
		}

		payload, parseErr := parseConfigPayload(data, fs)
		if parseErr != nil {
			fmt.Fprintf(os.Stderr, "%v\n", parseErr)
			return 1
		}

		if payload.IsMultiConfig {
			// Multi-config format
			configMap = payload.ConfigMap
			originalConfigDir = payload.OriginalConfigDir
			configTargetScopes = payload.ConfigTargetScopes

			// Inject .gitignore patterns as global ignores for each config.
			// Each config independently reads its own .gitignore tree:
			// The shared gitignore collector walks UP (ancestor inheritance) and DOWN
			// (nested .gitignore) from each configDir. Sibling configs are
			// fully isolated — they never share gitignore patterns.
			//
			// Directories excluded by global config ignores are pruned during
			// the .gitignore scan because files below them cannot be linted.
			//
			// configMap is not mutated by gitignore workers. Results are collected
			// and merged on this goroutine before target discovery.
			type giResult struct {
				configDir string
				config    rslintconfig.RslintConfig
			}
			var (
				giResults []giResult
				giMu      sync.Mutex
			)
			giWG := core.NewWorkGroup(singleThreaded)
			if needsLintTargets {
				for configDir, entries := range configMap {
					giWG.Queue(func() {
						augmented := rslintconfig.ConfigWithGitignore(entries, configDir, fs, nil)
						giMu.Lock()
						giResults = append(giResults, giResult{configDir, augmented})
						giMu.Unlock()
					})
				}
			}

			var programErr error
			if buildAllPrograms {
				realProgramSet, programErr = createProgramSetForConfigs(configMap, singleThreaded, fs, parseCache)
			}

			// Join the gitignore reads and merge into configMap. Must complete
			// before target discovery, which relies on the augmented configMap.
			giWG.RunAndWait()
			for _, r := range giResults {
				configMap[r.configDir] = r.config
			}
			if programErr != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", programErr)
				return 1
			}
		} else {
			// Legacy single-config format
			rslintConfig = payload.SingleConfig
			currentDirectory = payload.SingleConfigDir

			if typeCheckOnly {
				realProgramSet, err = createProgramSetForConfig(currentDirectory, rslintConfig, singleThreaded, fs, parseCache)
			} else if buildAllPrograms {
				rslintConfig, realProgramSet, err = parallelGitignoreAndPrograms(
					rslintConfig, currentDirectory, fs, singleThreaded, parseCache,
				)
			} else {
				rslintConfig = rslintconfig.ConfigWithGitignore(rslintConfig, currentDirectory, fs, nil)
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				return 1
			}
		}
	} else {
		// Load configuration from file (JSON config path, isJSConfig stays false)
		loader := rslintconfig.NewConfigLoader(fs, currentDirectory)
		rslintConfig, currentDirectory, err = loader.LoadRslintConfiguration(config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return 1
		}
		if config != "" {
			// Explicit --config follows flat-config invocation semantics: file,
			// ignore, project, and implicit no-args target scope are rooted at
			// the cwd where rslint was invoked, not the config file's directory.
			currentDirectory = workingDirectory
		}

		if typeCheckOnly {
			realProgramSet, err = createProgramSetForConfig(currentDirectory, rslintConfig, singleThreaded, fs, parseCache)
		} else if buildAllPrograms {
			rslintConfig, realProgramSet, err = parallelGitignoreAndPrograms(
				rslintConfig, currentDirectory, fs, singleThreaded, parseCache,
			)
		} else {
			rslintConfig = rslintconfig.ConfigWithGitignore(rslintConfig, currentDirectory, fs, nil)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return 1
		}
	}

	targetConfigMap := cloneConfigMap(configMap)
	targetRslintConfig := slices.Clone(rslintConfig)

	// Apply --rule CLI overrides by appending a synthetic ConfigEntry. Target
	// discovery below intentionally uses the pre-override config snapshots
	// above, so --rule overlays already-selected lint targets without widening
	// discovery by itself.
	if len(ruleFlags) > 0 {
		cliEntry, err := rslintconfig.BuildCLIRuleEntry(ruleFlags)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return 1
		}
		if cliEntry != nil {
			if configMap != nil {
				for dir, cfg := range configMap {
					configMap[dir] = append(cfg, *cliEntry)
				}
			} else {
				rslintConfig = append(rslintConfig, *cliEntry)
			}
		}
	}

	// Use CWD for display paths (not any config directory).
	// In multi-config mode, currentDirectory was never reassigned from os.Getwd(),
	// so it already holds the normalized CWD.
	cwd := workingDirectory

	comparePathOptions := tspath.ComparePathsOptions{
		CurrentDirectory:          cwd,
		UseCaseSensitiveFileNames: true,
	}

	// No args → implicit CWD scoping (same as `rslint .`), matching ESLint.
	// This keeps an explicit --config outside the current directory from
	// widening the scanned root to the config file's directory.
	if len(allowFiles) == 0 && len(allowDirs) == 0 {
		allowDirs = []string{cwd}
	}

	// --- Lint target discovery and fallback Program binding ---
	programs := realProgramSet.Programs
	programConfigMap := configMap
	buildSingleConfigPrograms := buildAllPrograms
	var (
		targetPlan                 lintTargetPlan
		typeInfoFiles              map[string]struct{}
		targetsByProgram           [][]string
		targetPathBySourcePath     map[string]string
		configPathBySourcePath     map[string]string
		ownerConfigDirBySourcePath map[string]string
	)
	// --type-check-only is program-wide and pays no lint-target discovery,
	// fallback, config-resolution, or Program-binding cost.
	if !typeCheckOnly {
		targetPlan, err = resolveLintTargetPlan(
			targetConfigMap,
			targetRslintConfig,
			currentDirectory,
			configTargetScopes,
			fs,
			allowFiles,
			allowDirs,
			singleThreaded,
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return 1
		}
		if !buildAllPrograms {
			if configMap != nil {
				programConfigMap = configsForLintTargetPlan(configMap, targetPlan)
				realProgramSet, err = createProgramSetForConfigs(programConfigMap, singleThreaded, fs, parseCache)
			} else if len(targetPlan.Targets) > 0 {
				buildSingleConfigPrograms = true
				realProgramSet, err = createProgramSetForConfig(currentDirectory, rslintConfig, singleThreaded, fs, parseCache)
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				return 1
			}
		}
		binding, err := bindLintTargetPlan(realProgramSet, targetPlan, currentDirectory, fs, parseCache, singleThreaded)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return 1
		}
		programs = binding.Programs
		typeInfoFiles = binding.TypeInfoFiles
		targetsByProgram = binding.TargetsByProgram
		targetPathBySourcePath = binding.TargetPathBySourcePath
		configPathBySourcePath = binding.ConfigPathBySourcePath
		ownerConfigDirBySourcePath = binding.OwnerConfigDirBySourcePath
	}

	// Initial build (including any fallback) is complete. Evict entries for
	// parsed files that ended up in no Program.
	parseCache.RetainOnly(programs)

	// Rebuild real Programs and bind the original stable target plan again on
	// every fix pass. A target can move between a tsconfig Program and fallback
	// when fixes change the import graph.
	createPrograms := func() (lintTargetBinding, error) {
		var rebuilt lintProgramSet
		var err error
		if configMap != nil {
			rebuilt, err = createProgramSetForConfigs(programConfigMap, singleThreaded, fs, parseCache)
		} else if buildSingleConfigPrograms {
			rebuilt, err = createProgramSetForConfig(currentDirectory, rslintConfig, singleThreaded, fs, parseCache)
		}
		if err != nil {
			return lintTargetBinding{}, err
		}
		return bindLintTargetPlan(rebuilt, targetPlan, currentDirectory, fs, parseCache, singleThreaded)
	}

	// Phase 1: Collect all diagnostics (no printing yet).
	// Like ESLint, diagnostics are collected first, then printed at the end.
	// This ensures --fix only shows remaining unfixed issues.
	var allDiags []rule.RuleDiagnostic
	var diagsMu sync.Mutex
	fixedCount := 0

	diagnosticsChan := make(chan rule.RuleDiagnostic, 4096)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for d := range diagnosticsChan {
			allDiags = append(allDiags, d)
		}
	}()

	enforcePlugins := configStdin // JS/TS configs loaded via stdin require plugin declarations
	fileConfigResolver := newLintConfigResolver(lintConfigResolverOptions{
		ConfigMap:                  configMap,
		Config:                     rslintConfig,
		CurrentDirectory:           currentDirectory,
		EnforcePlugins:             enforcePlugins,
		TypeInfoFiles:              typeInfoFiles,
		ConfigPathBySourcePath:     configPathBySourcePath,
		OwnerConfigDirBySourcePath: ownerConfigDirBySourcePath,
		SourceMappingsCanonical:    true,
		FS:                         fs,
	})
	getRulesForFile := func(sourceFile *ast.SourceFile) []linter.ConfiguredRule {
		return fileConfigResolver.ActiveRulesForFile(sourceFile.FileName())
	}

	// Target discovery already excluded default paths, global ignores, and
	// .gitignore entries. Target ownership and deduplication were already
	// resolved in targetsByProgram.
	// Programs not backed by a real tsconfig are excluded from --type-check:
	// their CompilerOptions are synthesized defaults, not the user's tsconfig,
	// so semantic diagnostics there would be unreliable. This includes the
	// non-project-backed fallback used for selected files outside tsconfig coverage and
	// for projects with no tsconfig at all, honoring the "Gap files" contract
	// in website/docs/en/guide/type-checking.md.
	skipTypeCheck := buildTypeCheckSkipMask(programs)
	syntaxDiagnostics, syntaxErrorFiles := collectTargetSyntacticDiagnostics(
		programs,
		targetsByProgram,
		skipTypeCheck,
		typeCheck,
		typeCheckOnly,
	)
	for _, diagnostic := range syntaxDiagnostics {
		diagnosticsChan <- diagnostic
	}

	// In --type-check-only mode, skip the lint phase entirely by passing
	// nil for GetRulesForFile. RunLinter's Phase 1 is gated on this being
	// non-nil; Phase 2 (type-check) runs independently.
	var rulesForFile linter.RuleHandler
	if !typeCheckOnly {
		rulesForFile = getRulesForFile
	}

	runOpts := linter.RunLinterOptions{
		Programs:              programs,
		SingleThreaded:        singleThreaded,
		Scope:                 linter.FileScope{Files: allowFiles, Dirs: allowDirs},
		TargetFiles:           targetsByProgram,
		GetRulesForFile:       rulesForFile,
		TypeInfoFiles:         typeInfoFiles,
		SyntaxErrorFiles:      syntaxErrorFiles,
		TypeCheck:             typeCheck,
		SkipTypeCheckPrograms: skipTypeCheck,
		OnDiagnostic: func(d rule.RuleDiagnostic) {
			diagnosticsChan <- d
		},
	}
	// Dispatch eslint-plugin rules to the Node worker in parallel with the
	// native lint pass; results are awaited + merged before output / --fix.
	// ONLY when plugins are actually configured — otherwise the whole reverse-
	// dispatch (including buildPluginFileInputs' extra per-file rule resolution
	// over every file) is skipped so the native-only path pays nothing for the
	// feature.
	hasEslintPlugins := len(args.EslintPlugins) > 0
	pluginResolver := pluginConfigResolver{
		lintResolver:      fileConfigResolver,
		originalConfigDir: originalConfigDir,
	}
	var pluginCh <-chan []rule.RuleDiagnostic
	if hasEslintPlugins {
		pluginInputs := buildPluginFileInputs(runOpts, pluginResolver)
		pluginCh = dispatchPluginLintAsync(ctx, dispatch, pluginInputs, fix, pluginSuggestionsMode(fix))
	}

	lintResult, err := linter.RunLinter(runOpts)

	close(diagnosticsChan)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error running linter: %v\n", err)
		return 1
	}

	lintedfileCount := lintResult.LintedFileCount

	wg.Wait()
	// Merge eslint-plugin diagnostics (dispatched in parallel) now that the
	// native diagnostics goroutine has drained.
	if pluginCh != nil {
		allDiags = append(allDiags, (<-pluginCh)...)
	}
	remapDiagnosticTargetPaths(allDiags, targetPathBySourcePath, fs)

	// Emit per-file warnings for CLI-specified files that won't be linted.
	// Distinguishes "not found in project" vs "ignored by pattern", aligned
	// with ESLint v10's warning behavior. Skipped in --type-check-only mode:
	// these are lint-phase concepts and would mislead users about Phase 2
	// (which runs program-wide regardless of CLI scope and rslint ignores).
	if !typeCheckOnly {
		warnings := collectAllowFileWarnings(allowFiles, configMap, rslintConfig, currentDirectory, fs.UseCaseSensitiveFileNames(), fs)
		for _, w := range warnings {
			fmt.Fprintln(os.Stderr, formatAllowFileWarning(w, comparePathOptions))
		}
	}
	scopeRestricted := len(allowFiles) > 0 || len(allowDirs) > 0
	if shouldShortCircuitOutput(typeCheckOnly, typeCheck, scopeRestricted, lintedfileCount) {
		return 0
	}

	// Phase 2: Apply fixes if --fix flag is enabled.
	// Uses multi-pass fixing: after applying fixes, rebuild programs and re-lint
	// to catch cascading issues (e.g. no-wrapper-object-types fix triggers no-inferrable-types).
	// After fixing, allDiags is replaced with remaining (unfixed) diagnostics.
	const maxFixPasses = 10
	if fix && len(allDiags) > 0 {
		diagnosticsByFile := groupDiagsByFile(allDiags)
		passFixed, fixErr := applyFixPass(diagnosticsByFile)
		// Replace the entire source generation after every write attempt and
		// before any Program rebuild. os.WriteFile may truncate or partially
		// mutate a file even when it ultimately returns an error, and whole-
		// generation invalidation also covers caller/source/symlink aliases.
		parseCache.InvalidateSourceSnapshots()
		if fixErr != nil {
			fmt.Fprintf(os.Stderr, "error applying fixes: %v\n", fixErr)
			return 1
		}
		fixedCount += passFixed

		// Re-lint → fix → re-lint → fix → ... until stable or maxFixPasses.
		// Skip if nothing was fixed in the first pass (no need to re-lint).
		for pass := 1; pass <= maxFixPasses && fixedCount > 0; pass++ {
			newBinding, err := createPrograms()
			if err != nil {
				fmt.Fprintf(os.Stderr, "error rebuilding Programs after fixes: %v\n", err)
				return 1
			}
			newPrograms := newBinding.Programs
			if len(newPrograms) == 0 {
				fmt.Fprintln(os.Stderr, "error rebuilding Programs after fixes: no Program returned")
				return 1
			}

			// Evict cache entries no longer referenced by any live program:
			// previous-round ASTs of rewritten files and this round's dedup
			// losers. The live set is the union of this round's programs and
			// the initial ones — the initial slice stays referenced until the
			// end of this function, so its objects are alive regardless and
			// keeping their entries costs nothing because RetainOnly only deletes.
			parseCache.RetainOnly(append(slices.Clone(newPrograms), programs...))

			// Re-lint using the fresh binding derived from the stable target plan.
			fixTargetsByProgram := newBinding.TargetsByProgram
			fixTargetPathBySourcePath := newBinding.TargetPathBySourcePath
			fixSkipMask := buildTypeCheckSkipMask(newPrograms)
			fixTypeInfoFiles := newBinding.TypeInfoFiles
			fixConfigResolver := newLintConfigResolver(lintConfigResolverOptions{
				ConfigMap:                  configMap,
				Config:                     rslintConfig,
				CurrentDirectory:           currentDirectory,
				EnforcePlugins:             enforcePlugins,
				TypeInfoFiles:              fixTypeInfoFiles,
				ConfigPathBySourcePath:     newBinding.ConfigPathBySourcePath,
				OwnerConfigDirBySourcePath: newBinding.OwnerConfigDirBySourcePath,
				SourceMappingsCanonical:    true,
				FS:                         fs,
			})
			fixGetRulesForFile := func(sourceFile *ast.SourceFile) []linter.ConfiguredRule {
				return fixConfigResolver.ActiveRulesForFile(sourceFile.FileName())
			}
			var fixRulesForFile linter.RuleHandler
			if !typeCheckOnly {
				fixRulesForFile = fixGetRulesForFile
			}
			var passDiags []rule.RuleDiagnostic
			fixSyntaxDiagnostics, fixSyntaxErrorFiles := collectTargetSyntacticDiagnostics(
				newPrograms,
				fixTargetsByProgram,
				fixSkipMask,
				typeCheck,
				typeCheckOnly,
			)
			passDiags = append(passDiags, fixSyntaxDiagnostics...)
			fixRunOpts := linter.RunLinterOptions{
				Programs:              newPrograms,
				SingleThreaded:        singleThreaded,
				Scope:                 linter.FileScope{Files: allowFiles, Dirs: allowDirs},
				TargetFiles:           fixTargetsByProgram,
				GetRulesForFile:       fixRulesForFile,
				TypeInfoFiles:         fixTypeInfoFiles,
				SyntaxErrorFiles:      fixSyntaxErrorFiles,
				TypeCheck:             typeCheck,
				SkipTypeCheckPrograms: fixSkipMask,
				OnDiagnostic: func(d rule.RuleDiagnostic) {
					diagsMu.Lock()
					passDiags = append(passDiags, d)
					diagsMu.Unlock()
				},
			}
			// Re-dispatch plugin rules each pass (only when configured): the
			// worker re-reads the post-fix file content, and merging here keeps
			// plugin diagnostics from being lost when allDiags is replaced.
			var fixPluginCh <-chan []rule.RuleDiagnostic
			if hasEslintPlugins {
				fixPluginInputs := buildPluginFileInputs(fixRunOpts, pluginConfigResolver{
					lintResolver:      fixConfigResolver,
					originalConfigDir: originalConfigDir,
				})
				fixPluginCh = dispatchPluginLintAsync(ctx, dispatch, fixPluginInputs, fix, pluginSuggestionsMode(fix))
			}
			passResult, passErr := linter.RunLinter(fixRunOpts)
			var fixPluginDiags []rule.RuleDiagnostic
			if fixPluginCh != nil {
				fixPluginDiags = <-fixPluginCh
			}
			if passErr != nil {
				fmt.Fprintf(os.Stderr, "error running linter after fixes: %v\n", passErr)
				return 1
			}
			if passResult != nil {
				for name := range passResult.ExecutedRules {
					lintResult.ExecutedRules[name] = struct{}{}
				}
			}
			// Merge this pass's plugin diagnostics before applying fixes so
			// plugin fixes participate and plugin diagnostics survive.
			passDiags = append(passDiags, fixPluginDiags...)
			remapDiagnosticTargetPaths(passDiags, fixTargetPathBySourcePath, fs)

			// Replace allDiags with latest post-fix diagnostics.
			allDiags = passDiags
			if pass == maxFixPasses {
				// The maximum number of write passes has already run (the initial
				// pass plus maxFixPasses-1 loop passes). This extra pass is the
				// required final verification of the bytes written by pass 10.
				break
			}

			passFixed, fixErr := applyFixPass(groupDiagsByFile(passDiags))
			// See the first fix pass above: invalidate before inspecting the
			// result so a partially successful write can never feed a rebuild.
			parseCache.InvalidateSourceSnapshots()
			if fixErr != nil {
				fmt.Fprintf(os.Stderr, "error applying fixes: %v\n", fixErr)
				return 1
			}
			if passFixed == 0 {
				break // Stable — allDiags reflect final state
			}
			fixedCount += passFixed
		}
	}

	allDiags = deduplicateTypeScriptDiagnostics(allDiags, fs, preferredCallerTargetPaths(targetPlan))

	// Diagnostics arrive in completion order — programs and, within a
	// program, file shards run in parallel — so impose a deterministic
	// order before printing. The key is (file, start position) only,
	// deliberately with NO end/rule tie-break: ESLint orders same-start
	// diagnostics by emission order (parent reported before nested child),
	// and a file's diagnostics are all emitted by a single worker, so under
	// a STABLE sort this key is already fully deterministic. Keep this
	// comparator in sync with the --api one in api.go (same policy over
	// api.Diagnostic).
	slices.SortStableFunc(allDiags, func(a, b rule.RuleDiagnostic) int {
		if c := strings.Compare(a.FilePath, b.FilePath); c != 0 {
			return c
		}
		return cmp.Compare(a.Range.Pos(), b.Range.Pos())
	})

	// Phase 3: Print diagnostics and count errors/warnings.
	// allDiags contains: original diagnostics (no fix), or remaining after fix.
	errorsCount := 0
	warningsCount := 0
	typeErrorsCount := 0
	{
		w := bufio.NewWriterSize(os.Stdout, 4096*100)
		var gitlabState *gitlabReportState
		if format == "gitlab" {
			gitlabState = newGitlabReportState()
		}
		for i, d := range allDiags {
			switch d.Severity {
			case rule.SeverityError:
				errorsCount++
				if typeCheck && strings.HasPrefix(d.RuleName, "TypeScript(") {
					typeErrorsCount++
				}
			case rule.SeverityWarning:
				warningsCount++
			}

			if i == 0 {
				w.WriteByte('\n')
			}
			// Only print Error message when quiet is true
			if quiet && d.Severity != rule.SeverityError {
				continue
			}
			printDiagnostic(d, w, comparePathOptions, format, gitlabState)
			if w.Available() < 4096 {
				w.Flush()
			}
		}
		if gitlabState != nil {
			gitlabState.finish(w)
		}
		w.Flush()
	}

	lintErrorsCount := errorsCount - typeErrorsCount

	colors := setupColors()
	var errorsColorFunc func(string, ...interface{}) string
	if errorsCount == 0 {
		errorsColorFunc = colors.SuccessText
	} else {
		errorsColorFunc = colors.ErrorText
	}

	var warningsColorFunc func(string, ...interface{}) string
	if warningsCount == 0 {
		warningsColorFunc = colors.SuccessText
	} else {
		warningsColorFunc = colors.WarnText
	}

	warningsText := pluralize(warningsCount, "warning", "warnings")
	filesText := pluralize(int(lintedfileCount), "file", "files")
	rulesCount := len(lintResult.ExecutedRules)
	rulesText := pluralize(rulesCount, "rule", "rules")
	threadsCount := 1
	if !singleThreaded {
		threadsCount = runtime.GOMAXPROCS(0)
	}
	if format == "default" {
		// Build the errors summary part.
		// When type-check is enabled and there are type errors, split the display.
		var errorsSummary string
		switch {
		case typeCheckOnly:
			// Lint phase was skipped; only type errors are possible.
			errorsSummary = fmt.Sprintf("%s %s",
				errorsColorFunc("%d", typeErrorsCount),
				pluralize(typeErrorsCount, "type error", "type errors"),
			)
		case typeCheck:
			errorsSummary = fmt.Sprintf("%s %s, %s %s",
				errorsColorFunc("%d", lintErrorsCount),
				pluralize(lintErrorsCount, "lint error", "lint errors"),
				errorsColorFunc("%d", typeErrorsCount),
				pluralize(typeErrorsCount, "type error", "type errors"),
			)
		default:
			errorsSummary = fmt.Sprintf("%s %s",
				errorsColorFunc("%d", errorsCount),
				pluralize(errorsCount, "error", "errors"),
			)
		}

		if typeCheckOnly {
			// type-check-only: omit lint-file/rule/warning columns since no
			// lint ran. Report the type-checked file count derived from
			// non-skipped programs' root files (tsconfig include/files);
			// transitive .d.ts imports are excluded for user readability.
			seen := make(map[string]struct{})
			for i, prog := range programs {
				if i < len(skipTypeCheck) && skipTypeCheck[i] {
					continue
				}
				for _, fn := range prog.CommandLine().FileNames() {
					seen[fn] = struct{}{}
				}
			}
			typeCheckedFileCount := len(seen)
			fmt.Fprintf(
				os.Stdout,
				"Found %s %s(type-checked %s %s in %s using %s threads)%s\n",
				errorsSummary,
				colors.DimText(""),
				colors.BoldText("%d", typeCheckedFileCount),
				pluralize(typeCheckedFileCount, "file", "files"),
				colors.BoldText("%v", time.Since(timeBefore).Round(time.Millisecond)),
				colors.BoldText("%d", threadsCount),
				colors.Reset,
			)
		} else if fix && fixedCount > 0 {
			fixText := pluralize(fixedCount, "issue", "issues")
			fmt.Fprintf(
				os.Stdout,
				"Found %s and %s %s %s(linted %s %s with %s %s in %s using %s threads, fixed %s %s)%s\n",
				errorsSummary,
				warningsColorFunc("%d", warningsCount),
				warningsText,
				colors.DimText(""),
				colors.BoldText("%d", lintedfileCount),
				filesText,
				colors.BoldText("%d", rulesCount),
				rulesText,
				colors.BoldText("%v", time.Since(timeBefore).Round(time.Millisecond)),
				colors.BoldText("%d", threadsCount),
				colors.SuccessText("%d", fixedCount),
				fixText,
				colors.Reset,
			)
		} else {
			fmt.Fprintf(
				os.Stdout,
				"Found %s and %s %s %s(linted %s %s with %s %s in %s using %s threads)%s\n",
				errorsSummary,
				warningsColorFunc("%d", warningsCount),
				warningsText,
				colors.DimText(""),
				colors.BoldText("%d", lintedfileCount),
				filesText,
				colors.BoldText("%d", rulesCount),
				rulesText,
				colors.BoldText("%v", time.Since(timeBefore).Round(time.Millisecond)),
				colors.BoldText("%d", threadsCount),
				colors.Reset,
			)
		}
	}

	tooManyWarnings := maxWarnings >= 0 && warningsCount > maxWarnings

	if errorsCount == 0 && tooManyWarnings {
		fmt.Fprintf(os.Stderr, "Rslint found too many warnings (maximum: %d).\n", maxWarnings)
	}

	// Exit with non-zero status code if errors were found
	if errorsCount > 0 || tooManyWarnings {
		return 1
	}
	return 0
}
