package main

import (
	"bufio"
	"context"
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
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
)

// lintArgs is the parsed-and-validated input to executeLintPipeline.
// Built by parseLintFlags (consuming os.Args) and then optionally
// overlaid with overrides from an IPC `init` payload inside runCLI.
type lintArgs struct {
	Init           bool
	Config         string
	ConfigStdin    bool // true → executeLintPipeline reads stdin as a config payload (set by runCLI)
	Fix            bool
	TypeCheck      bool
	TypeCheckOnly  bool
	TraceOut       string
	CpuprofOut     string
	SingleThreaded bool
	Format         string
	NoColor        bool
	ForceColor     bool
	Quiet          bool
	MaxWarnings    int
	StartTimeMs    int64
	RuleFlags      []string
	// Positional args resolved into existing-dir vs file paths.
	AllowFiles []string
	AllowDirs  []string
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
}

// setupColors initializes the color scheme based on environment and flags
func setupColors() *ColorScheme {
	// Handle color flags and environment variables
	if os.Getenv("NO_COLOR") != "" {
		color.NoColor = true
	}
	if os.Getenv("FORCE_COLOR") != "" {
		color.NoColor = false
	}

	// GitHub Actions specific handling
	if os.Getenv("GITHUB_ACTIONS") != "" {
		color.NoColor = false // Enable colors in GitHub Actions
	}

	// Create color functions
	ruleNameColor := color.New(color.FgHiGreen).SprintfFunc()
	fileNameColor := color.New(color.FgCyan, color.Italic).SprintfFunc()
	errorTextColor := color.New(color.FgRed, color.Bold).SprintfFunc()
	successColor := color.New(color.FgGreen, color.Bold).SprintfFunc()
	dimColor := color.New(color.Faint).SprintfFunc()
	boldColor := color.New(color.Bold).SprintfFunc()
	borderColor := color.New(color.Faint).SprintfFunc()
	WarnColor := color.New(color.FgYellow).SprintfFunc()

	return &ColorScheme{
		RuleName:    ruleNameColor,
		FileName:    fileNameColor,
		ErrorText:   errorTextColor,
		SuccessText: successColor,
		DimText:     dimColor,
		BoldText:    boldColor,
		BorderText:  borderColor,
		WarnText:    WarnColor,
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
			FilePath:   d.File().FileName(),
			SourceFile: d.File(),
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

func printDiagnostic(d rule.RuleDiagnostic, w *bufio.Writer, comparePathOptions tspath.ComparePathsOptions, format string) {
	switch format {
	case "default":
		printDiagnosticDefault(d, w, comparePathOptions)
	case "jsonline":
		printDiagnosticJsonLine(d, w, comparePathOptions)
	case "github":
		printDiagnosticGitHub(d, w, comparePathOptions)
	default:
		panic("not supported format " + format)
	}
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
  --config PATH         Which rslint config file to use.
  --format FORMAT       Output format: default | jsonline | github
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

// applyFixPass applies auto-fixes for all files in diagnosticsByFile,
// writes fixed content to disk, and returns the number of issues fixed.
func applyFixPass(diagnosticsByFile map[string][]rule.RuleDiagnostic) int {
	fixed := 0
	for fileName, fileDiagnostics := range diagnosticsByFile {
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
				fmt.Fprintf(os.Stderr, "error writing fixed file %s: %v\n", fileName, err)
			} else {
				fixed += len(diagnosticsWithFixes) - len(unapplied)
			}
		}
	}
	return fixed
}

// allowFileWarningKind categorizes why a CLI-specified file won't have lint
// rules applied. These are Phase-1 (lint) concepts; Phase 2 (type-check) is
// not affected by either case (it runs program-wide regardless of CLI scope
// and rslint ignores).
type allowFileWarningKind int

const (
	allowFileNotInProgram allowFileWarningKind = iota
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
	case allowFileNotInProgram:
		return fmt.Sprintf("warning: %s was not found in the project, skipping", rel)
	case allowFileIgnored:
		return fmt.Sprintf("warning: %s is ignored because of a matching ignore pattern", rel)
	}
	return ""
}

// collectAllowFileWarnings explains, for each CLI-specified file in
// allowFiles, why it won't be visited by Phase 1 (lint). A file lands in
// the result if it is outside every Program, or if it is in some Program
// but the nearest config's `ignores` patterns exclude it. Returns nil for
// empty allowFiles.
//
// This is a Phase-1 concern only. In --type-check-only mode the lint phase
// is skipped, so callers MUST gate emission on `!typeCheckOnly` — otherwise
// users see misleading messages like "is ignored because of a matching
// ignore pattern" while Phase 2 happily reports type errors for that same
// file. See website/docs/en/guide/type-checking.md ("type-check is not
// constrained by `files`/`ignores`").
// In compatOnlyMode the caller passes a non-nil `compatLintedFiles`
// set populated from `buildCompatFileInputs`, since `programs` is
// empty in that fast path and using empty `programs` as the
// "was-linted" oracle wrongly flags every CLI-passed file as
// `allowFileNotInProgram`. Outside compat-only mode the caller
// passes nil and the oracle falls back to the program-file walk.
func collectAllowFileWarnings(
	allowFiles []string,
	programs []*compiler.Program,
	compatLintedFiles map[string]struct{},
	configMap map[string]rslintconfig.RslintConfig,
	rslintConfig rslintconfig.RslintConfig,
	currentDirectory string,
) []allowFileWarning {
	if len(allowFiles) == 0 {
		return nil
	}
	programFiles := make(map[string]struct{})
	for _, prog := range programs {
		for _, sf := range prog.GetSourceFiles() {
			programFiles[sf.FileName()] = struct{}{}
		}
	}
	// In compatOnlyMode the file set the dispatcher actually linted
	// is the source of truth. Merge it into `programFiles` so the
	// `inProgram` lookup below treats compat-dispatched files the
	// same as program-resident files. Outside compat-only mode the
	// map is nil; nothing to merge.
	for f := range compatLintedFiles {
		programFiles[f] = struct{}{}
	}

	// Cache FindNearestConfig results by directory to avoid redundant lookups
	// when many files are in the same directory (e.g., lint-staged).
	type cachedConfig struct {
		cfgDir string
		cfg    rslintconfig.RslintConfig
	}
	dirConfigCache := make(map[string]*cachedConfig)

	var out []allowFileWarning
	for _, f := range allowFiles {
		if _, inProgram := programFiles[f]; !inProgram {
			out = append(out, allowFileWarning{Path: f, Kind: allowFileNotInProgram})
			continue
		}
		// File is in a Program — check if config would assign rules
		var merged *rslintconfig.MergedConfig
		if configMap != nil {
			dir := tspath.GetDirectoryPath(f)
			cached, ok := dirConfigCache[dir]
			if !ok {
				cfgDir, cfg := rslintconfig.FindNearestConfig(f, configMap)
				cached = &cachedConfig{cfgDir: cfgDir, cfg: cfg}
				dirConfigCache[dir] = cached
			}
			if cached.cfg != nil {
				merged = cached.cfg.GetConfigForFile(f, cached.cfgDir)
			}
		} else {
			merged = rslintConfig.GetConfigForFile(f, currentDirectory)
		}
		if merged == nil {
			out = append(out, allowFileWarning{Path: f, Kind: allowFileIgnored})
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

// parseLintFlags parses argv (typically os.Args[1:]) through Go's global
// flag set into a lintArgs.
//
// Returns:
//   - args: the parsed lintArgs (always populated, even on partial failure)
//   - help: true when `--help` / `-h` was set; caller should print usage and exit 0
//   - fatalExitCode: non-zero when parsing failed in a way the caller cannot
//     recover from (e.g., a positional path couldn't be resolved). Caller
//     should propagate this exit code without proceeding into the lint
//     pipeline. Zero on success / on the help path.
//
// Called once per process by runCLI. The IPC `init` payload then overlays
// a small subset (positional files, working directory, force-color); the
// rest of the flag-derived state (--fix, --format, --type-check, …) is
// taken as-is from the user-forwarded argv.
func parseLintFlags(argv []string) (args lintArgs, help bool, fatalExitCode int) {
	// Use a fresh FlagSet so parseLintFlags is callable multiple times
	// per process without panicking on "flag redefined". flag.CommandLine
	// retains every Var registration globally; calling parseLintFlags
	// twice (e.g. from tests, or from any future caller that runs the
	// CLI loop more than once) would otherwise panic. fs is also
	// configured with ContinueOnError so a parse failure becomes a
	// returnable error rather than an os.Exit(2) that bypasses our
	// own cleanup (IPC shutdown, stdout drain).
	fs := flag.NewFlagSet("rslint", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() { fmt.Fprint(os.Stderr, usage) }

	var ruleFlags repeatedFlag

	fs.StringVar(&args.Format, "format", "default", "output format")
	fs.StringVar(&args.Config, "config", "", "which rslint config to use")
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
		// Translate to a fatal exit so caller bails before IPC handshake.
		fatalExitCode = 2
		return args, help, fatalExitCode
	}
	args.RuleFlags = []string(ruleFlags)

	// --type-check-only implies --type-check and skips all lint rules.
	// Reject incompatible flag combinations before doing any work.
	if code, msg := validateTypeCheckOnlyFlags(args.TypeCheckOnly, args.Fix, []string(ruleFlags)); code != 0 {
		fmt.Fprintln(os.Stderr, msg)
		fatalExitCode = code
		return args, help, fatalExitCode
	}
	if args.TypeCheckOnly {
		args.TypeCheck = true
	}

	// Collect file/directory arguments for targeted linting (e.g. rslint file1.ts src/)
	if positional := fs.Args(); len(positional) > 0 {
		var allowFiles, allowDirs []string
		for _, arg := range positional {
			absPath, err := filepath.Abs(arg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error resolving path %s: %v\n", arg, err)
				args.AllowFiles = nil
				args.AllowDirs = nil
				fatalExitCode = 1
				return args, help, fatalExitCode
			}
			// NOTE: we intentionally do NOT call filepath.EvalSymlinks here.
			// EvalSymlinks resolves symlinks (macOS /tmp → /private/tmp) and
			// Windows 8.3 short names to long names, but the rest of the
			// pipeline (os.Getwd, TypeScript file names, configDir) uses
			// unresolved CWD-based paths. Resolving only file args would
			// create a format mismatch causing failures in gap file detection,
			// config matching, dir scoping, and gitignore checks.
			normalized := tspath.NormalizePath(absPath)
			info, statErr := os.Stat(absPath)
			if statErr == nil && info.IsDir() {
				allowDirs = append(allowDirs, normalized)
			} else {
				allowFiles = append(allowFiles, normalized)
			}
		}
		args.AllowFiles = allowFiles
		args.AllowDirs = allowDirs
	}

	return args, help, fatalExitCode
}

// allRulesAreCompat reports whether every rule referenced by every entry in
// the supplied configuration set is an eslint-plugin (compat) rule. Returns
// false when:
//   - any rule is a native (Go-implemented) rule;
//   - any rule name isn't registered at all (we treat unknown as "could
//     be native" to stay safe);
//   - the config has zero rules in it (no rules → nothing to compat-only).
//
// Used as the safety gate for skipping the ts-go program build: if all the
// configured rules are compat, the worker pool handles everything and Go
// never needs the AST / type checker, so program creation is pure waste.
func allRulesAreCompat(rslintConfig rslintconfig.RslintConfig, configMap map[string]rslintconfig.RslintConfig) bool {
	check := func(cfg rslintconfig.RslintConfig) bool {
		anyRule := false
		for _, entry := range cfg {
			for ruleName := range entry.Rules {
				anyRule = true
				r, ok := rslintconfig.GlobalRuleRegistry.GetRule(ruleName)
				if !ok || !r.IsEslintPluginRule {
					return false
				}
			}
		}
		return anyRule
	}
	if configMap != nil {
		if len(configMap) == 0 {
			return false
		}
		for _, cfg := range configMap {
			if !check(cfg) {
				return false
			}
		}
		return true
	}
	return check(rslintConfig)
}

// configHasFilesField reports whether at least one entry in the config set
// declares a `files` glob — the prerequisite for DiscoverGapFiles to walk
// the filesystem and return matched files (no `files` field → tsconfig-only
// behavior, which doesn't apply when we're skipping the ts-go program).
func configHasFilesField(rslintConfig rslintconfig.RslintConfig, configMap map[string]rslintconfig.RslintConfig) bool {
	check := func(cfg rslintconfig.RslintConfig) bool {
		for _, entry := range cfg {
			if len(entry.Files) > 0 {
				return true
			}
		}
		return false
	}
	if configMap != nil {
		for _, cfg := range configMap {
			if check(cfg) {
				return true
			}
		}
		return false
	}
	return check(rslintConfig)
}

// shouldUseCompatOnlyFastPath reports whether the compat-only fast path
// applies: a compat dispatcher exists, no type-check is requested, every
// active rule is an eslint-plugin rule, and the config carries `files`
// globs. When true the caller skips the ts-go Program build entirely
// (plugin rules run in the worker pool via oxc-parser). Pass either a
// single rslintConfig (configMap nil) or a configMap (rslintConfig nil) —
// allRulesAreCompat / configHasFilesField accept both shapes.
//
// Centralizes the guard previously copy-pasted at the three config-load
// branches (multi-config / single-or-stdin / JSON-config) so a future
// condition can't be added to two of them and silently skipped in the third.
func shouldUseCompatOnlyFastPath(
	args lintArgs,
	rslintConfig rslintconfig.RslintConfig,
	configMap map[string]rslintconfig.RslintConfig,
	compatDispatcher linter.CompatBatchHandler,
) bool {
	return !args.TypeCheck && !args.TypeCheckOnly && compatDispatcher != nil &&
		allRulesAreCompat(rslintConfig, configMap) &&
		configHasFilesField(rslintConfig, configMap)
}

// buildCompatFileInputs assembles the per-file inputs for DispatchCompat
// from a discovered file list and the user's resolved config(s). For
// each file we:
//
//  1. Read its source text from disk (the worker pool needs the text
//     to feed oxc-parser).
//  2. Find the nearest config (multi-config) or use the single config
//     (legacy / non-stdin), then run the same active-rules resolution
//     RunLinter's `getRulesForFile` does — preserving severity, options,
//     language options, settings, and the owning config directory.
//  3. Filter to compat (IsEslintPluginRule) entries; if none survive,
//     the file is skipped entirely.
//
// Files that can't be read or that yield zero compat rules drop out
// silently — matching the pre-refactor RunLinter behavior where files
// with empty `rules` (linter.go:204) skipped the per-file lint loop.
func buildCompatFileInputs(
	files []string,
	allowFiles []string,
	allowDirs []string,
	rslintConfig rslintconfig.RslintConfig,
	configMap map[string]rslintconfig.RslintConfig,
	currentDirectory string,
	enforcePlugins bool,
	fsys vfs.FS,
) []linter.CompatFileEntry {
	if len(files) == 0 {
		return nil
	}
	// Honor CLI scope: when --files or --dirs args were passed, restrict
	// to that subset. RunLinter applies the same gate in its inner loop;
	// here we apply it up-front so we don't even read text for files we
	// won't lint.
	// File-scope matching reuses the linter's exported helpers so the
	// compat-only path and RunLinter can't drift on symlink / case handling.
	// allowFiles are NormalizePath'd by the caller (same shape as the paths
	// in `files`), so IsFileAllowed's string-equality fast path hits; the
	// os.SameFile fallback covers symlink mismatches.
	allowFileInfos := linter.PrecomputeAllowFileInfos(allowFiles)
	caseSensitive := fsys.UseCaseSensitiveFileNames()

	out := make([]linter.CompatFileEntry, 0, len(files))
	for _, p := range files {
		if len(allowFiles) > 0 || len(allowDirs) > 0 {
			fileOK := (len(allowFiles) > 0 && linter.IsFileAllowed(p, allowFiles, allowFileInfos)) ||
				(len(allowDirs) > 0 && linter.IsDirAllowed(p, allowDirs, caseSensitive))
			if !fileOK {
				continue
			}
		}

		text, ok := fsys.ReadFile(p)
		if !ok {
			continue
		}

		var cfg rslintconfig.RslintConfig
		var cwd string
		if configMap != nil {
			cwd, cfg = rslintconfig.FindNearestConfig(p, configMap)
			if cfg == nil {
				continue
			}
		} else {
			cfg = rslintConfig
			cwd = currentDirectory
		}

		activeRules, _ := rslintconfig.GlobalRuleRegistry.GetEnabledRules(cfg, p, cwd, enforcePlugins)
		if len(activeRules) == 0 {
			continue
		}

		ruleMap, sevMap, langOpts, settings, configKey, ok :=
			linter.CompatRuleMaps(activeRules)
		if !ok {
			continue
		}
		out = append(out, linter.CompatFileEntry{
			Path:            p,
			Text:            text,
			Rules:           ruleMap,
			Severity:        sevMap,
			LanguageOptions: langOpts,
			Settings:        settings,
			ConfigKey:       configKey,
		})
	}
	return out
}

// executeLintPipeline runs the full CLI lint flow against the given args.
// Called only by runCLI (the unified IPC entry). Direct-binary invocation
// is no longer supported — the binary is an internal npm artifact.
//
// The pipeline:
//  1. Resolve cwd, optionally apply --init / profiling
//  2. Load config (JSON file OR stdin payload when args.ConfigStdin)
//  3. Build TS Programs (via createProgramsForConfig / multi-config)
//  4. Discover gap files + create fallback Program
//  5. Run linter (Phase 1) — native rules + IsEslintPluginRule rules
//     dispatched via `compatDispatcher` (nil → silently skipped)
//  6. Multi-pass --fix loop (rebuild Programs, re-lint, re-fix)
//  7. Print diagnostics, summary, exit code
//
// `compatDispatcher` is the ESLint-plugin rule dispatcher: in CLI mode
// it reverse-RPCs to the Node parent's WorkerPool over IPC; in LSP mode
// it sends `rslint/lintCompatBatch` server→client requests to the
// editor, which runs the rules in its own WorkerPool and replies with
// diagnostics; `--api` passes nil and any IsEslintPluginRule rule is
// silently skipped. Passing nil is valid for native-rules-only flows.
//
// `ctx` is consulted at file-loop boundaries inside each Program. When
// it fires, the partial diagnostics already streamed are kept and the
// pipeline returns early; the exit code maps Canceled/DeadlineExceeded
// to a SIGINT-like 130 in the caller (runCLI). Pass context.Background
// (or nil) for uncancellable runs.
//
// Exit codes match ESLint conventions: 0 (clean), 1 (lint errors / too
// many warnings), other (config / runtime errors written to stderr).
func executeLintPipeline(ctx context.Context, args lintArgs, compatDispatcher linter.CompatBatchHandler) int {
	// Override color detection based on flags
	// --no-color must win over TTY-derived forceColor. The CLI wrapper
	// sends forceColor=true whenever stdout is a TTY (cli.ts), so without
	// this precedence an explicit --no-color on a TTY would be overridden
	// and color emitted anyway.
	if args.NoColor {
		color.NoColor = true
	} else if args.ForceColor {
		color.NoColor = false
	}

	enableVirtualTerminalProcessing()
	timeBefore := resolveStartTime(args.StartTimeMs)

	if args.TraceOut != "" {
		f, err := os.Create(args.TraceOut)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating trace file: %v\n", err)
			return 1
		}
		defer f.Close()
		trace.Start(f)
		defer trace.Stop()
	}
	if args.CpuprofOut != "" {
		f, err := os.Create(args.CpuprofOut)
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

	if args.Init {
		if err := rslintconfig.InitDefaultConfig(currentDirectory); err != nil {
			fmt.Fprintf(os.Stderr, "error initializing config: %v\n", err)
			return 1
		}
		return 0
	}

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))

	// Initialize rule registry with all available rules
	rslintconfig.RegisterAllRules()
	var rslintConfig rslintconfig.RslintConfig

	// configMap holds per-directory configs for multi-config (monorepo) support.
	// Only populated in the configStdin path; nil otherwise (single-config mode).
	var configMap map[string]rslintconfig.RslintConfig

	programs := []*compiler.Program{}
	// programConfigDirs tracks which configDir each program belongs to (parallel to programs).
	// Used for ownership-based file deduplication in multi-config mode.
	var programConfigDirs []string

	allowFiles := args.AllowFiles
	allowDirs := args.AllowDirs

	// compatOnlyMode is set by the config-load branches below when every
	// configured rule is an eslint-plugin rule and no type-check is
	// requested. In that case program build is skipped entirely (see the
	// per-branch fast paths) and DispatchCompat replaces RunLinter
	// downstream. The flag is also consulted by the DiscoverGapFiles
	// block to skip the fallback Program build — the gap file list
	// itself becomes the compat dispatch input.
	var compatOnlyMode bool

	// applyCLIRuleOverrides materializes `--rule` CLI flags into a
	// synthetic ConfigEntry and appends it to the active config(s).
	// It MUST be called inside every config-load branch BEFORE the
	// compat-only fast-path evaluates `allRulesAreCompat`, otherwise
	// a user `--rule no-console:error` on top of an all-plugin config
	// is silently dropped: the fast-path sees only the original
	// (plugin-only) rules and routes to DispatchCompat, which
	// filters every entry where `IsEslintPluginRule == false`. Native
	// rules from `--rule` then never execute — false negative.
	//
	// `cliRulesApplied` guards against double-application; the
	// trailing block at the bottom of this function also calls this
	// helper as a safety net for any branch that forgot to call it
	// (currently none, but future branches inherit the guarantee).
	cliRulesApplied := false
	applyCLIRuleOverrides := func() error {
		if cliRulesApplied || len(args.RuleFlags) == 0 {
			return nil
		}
		cliRulesApplied = true
		cliEntry, err := rslintconfig.BuildCLIRuleEntry(args.RuleFlags)
		if err != nil {
			return err
		}
		if cliEntry == nil {
			return nil
		}
		if configMap != nil {
			for dir, cfg := range configMap {
				configMap[dir] = append(cfg, *cliEntry)
			}
		} else {
			rslintConfig = append(rslintConfig, *cliEntry)
		}
		return nil
	}

	if args.ConfigStdin {
		// Read config JSON from stdin (sent by JS config loader / IPC init).
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

		payload, parseErr := parseConfigPayload(data)
		if parseErr != nil {
			fmt.Fprintf(os.Stderr, "%v\n", parseErr)
			return 1
		}

		if payload.IsMultiConfig {
			// Multi-config format
			configMap = payload.ConfigMap

			// Apply --rule CLI overrides BEFORE the fast-path check
			// below so any user-introduced native rule turns off
			// compatOnlyMode (see helper definition above).
			if err := applyCLIRuleOverrides(); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				return 1
			}

			// Compat-only fast path: when every rule across every config
			// is an eslint-plugin rule (and no type-check is requested),
			// skip all ts-go Program build. Plugin rules run in the
			// worker pool with oxc-parser; the Go side never consumes the
			// AST. Gitignore still has to merge (DiscoverGapFiles below
			// consults the merged ignores), but we can do it serially
			// since there's no Program build to overlap with.
			if shouldUseCompatOnlyFastPath(args, nil, configMap, compatDispatcher) {
				for configDir, entries := range configMap {
					configMap[configDir] = readGitignoreOnly(entries, configDir, fs)
				}
				compatOnlyMode = true
				// programs / programConfigDirs stay empty;
				// DispatchCompat replaces RunLinter downstream.
				goto afterProgramBuild
			}

			// Inject .gitignore patterns as global ignores for each config.
			// Each config independently reads its own .gitignore tree:
			// ReadGitignoreAsGlobs walks UP (ancestor inheritance) and DOWN
			// (nested .gitignore) from each configDir. Sibling configs are
			// fully isolated — they never share gitignore patterns.
			//
			// Concurrency:
			//   - When singleThreaded is set, both stages run sequentially in
			//     the main goroutine (no goroutines spawned at all).
			//   - Otherwise, gitignore reads run in parallel across configs
			//     (independent FS reads via cachedvfs, which is concurrent-
			//     safe). createProgramsForConfig still runs serially in the
			//     main goroutine — typescript-go's API is invoked one config
			//     at a time. The two stages overlap: gitignore goroutines
			//     run alongside the createPrograms loop.
			//
			// configMap is NOT mutated by gitignore goroutines; results are
			// collected via channel and merged in the main goroutine after
			// the createPrograms loop, so the createPrograms loop sees the
			// pre-augmentation entries (createProgramsForConfig only reads
			// languageOptions.parserOptions.project — Ignores entries are
			// no-ops for it; verified in LoadTsConfigsFromRslintConfig).
			type giResult struct {
				configDir string
				globs     []string
			}
			var (
				giResults chan giResult
				giWG      sync.WaitGroup
			)
			if args.SingleThreaded {
				// Inline serial gitignore reads.
				giResults = nil
				for configDir, entries := range configMap {
					configIgnores := rslintconfig.ExtractConfigIgnores(entries)
					globs := rslintconfig.ReadGitignoreAsGlobs(configDir, fs, configIgnores)
					if len(globs) > 0 {
						configMap[configDir] = append(
							rslintconfig.RslintConfig{{Ignores: globs}},
							configMap[configDir]...,
						)
					}
				}
			} else {
				giResults = make(chan giResult, len(configMap))
				for configDir, entries := range configMap {
					configIgnores := rslintconfig.ExtractConfigIgnores(entries)
					giWG.Add(1)
					go func(dir string, ignores []string) {
						defer giWG.Done()
						globs := rslintconfig.ReadGitignoreAsGlobs(dir, fs, ignores)
						giResults <- giResult{configDir: dir, globs: globs}
					}(configDir, configIgnores)
				}
				go func() { giWG.Wait(); close(giResults) }()
			}

			seenTsConfigs := make(map[string]struct{})

			// Deterministic config iteration: Go map ranges are randomized,
			// so `seenTsConfigs` ownership (first-config-wins) used to be
			// non-deterministic, producing different `programConfigDirs`
			// order across runs and ultimately different `fileOwner` maps
			// (programs.go:192-204). In monorepos with shared tsconfigs
			// that yielded run-to-run diagnostic flicker. Sorting the
			// configDir keys pins the ownership outcome.
			configDirs := make([]string, 0, len(configMap))
			for configDir := range configMap {
				configDirs = append(configDirs, configDir)
			}
			sort.Strings(configDirs)
			for _, configDir := range configDirs {
				entries := configMap[configDir]
				progs, exitCode := createProgramsForConfig(configDir, entries, args.SingleThreaded, fs, seenTsConfigs)
				if exitCode != 0 {
					return exitCode
				}
				for range progs {
					programConfigDirs = append(programConfigDirs, configDir)
				}
				programs = append(programs, progs...)
			}

			// Drain gitignore results (parallel path only) and merge into
			// configMap. Must complete before DiscoverGapFiles, which relies
			// on the augmented configMap.
			if giResults != nil {
				for r := range giResults {
					if len(r.globs) > 0 {
						configMap[r.configDir] = append(
							rslintconfig.RslintConfig{{Ignores: r.globs}},
							configMap[r.configDir]...,
						)
					}
				}
			}
		} else {
			// Legacy single-config format
			rslintConfig = payload.SingleConfig
			currentDirectory = payload.SingleConfigDir

			// Apply --rule CLI overrides BEFORE fast-path check.
			if err := applyCLIRuleOverrides(); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				return 1
			}

			// Compat-only fast path: when every rule across the config is
			// an eslint-plugin rule and the config has `files` globs, skip
			// the ts-go Program build entirely. Plugin rules run in the
			// worker pool with their own oxc-parser; the AST + type
			// checker the Program would provide are never consumed. We
			// still need gitignore-merged ignores for DiscoverGapFiles
			// below — keep just that half.
			if shouldUseCompatOnlyFastPath(args, rslintConfig, nil, compatDispatcher) {
				rslintConfig = readGitignoreOnly(rslintConfig, currentDirectory, fs)
				compatOnlyMode = true
				// programs stays empty; DispatchCompat will be invoked
				// later in lieu of RunLinter.
			} else {
				// Inject .gitignore patterns as global ignores. Run gitignore
				// reading in parallel with createProgramsForConfig — they're
				// independent (createProgramsForConfig only reads
				// languageOptions.parserOptions.project, not Ignores).
				var (
					progs    []*compiler.Program
					exitCode int
				)
				rslintConfig, progs, exitCode = parallelGitignoreAndPrograms(
					rslintConfig, currentDirectory, fs, args.SingleThreaded, nil,
				)
				if exitCode != 0 {
					return exitCode
				}
				programs = append(programs, progs...)
			}
		}
	} else {
		// Load configuration from file (JSON config path, isJSConfig stays false)
		rslintConfig, _, currentDirectory = rslintconfig.LoadConfigurationWithFallback(args.Config, currentDirectory, fs)

		// Apply --rule CLI overrides BEFORE fast-path check.
		if err := applyCLIRuleOverrides(); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return 1
		}

		// Compat-only fast path: same shape as the stdin-singleConfig branch
		// above. See that comment for the rationale.
		if shouldUseCompatOnlyFastPath(args, rslintConfig, nil, compatDispatcher) {
			rslintConfig = readGitignoreOnly(rslintConfig, currentDirectory, fs)
			compatOnlyMode = true
		} else {
			// Inject .gitignore patterns as global ignores. Run gitignore reading
			// in parallel with createProgramsForConfig (see comment above).
			var (
				progs    []*compiler.Program
				exitCode int
			)
			rslintConfig, progs, exitCode = parallelGitignoreAndPrograms(
				rslintConfig, currentDirectory, fs, args.SingleThreaded, nil,
			)
			if exitCode != 0 {
				return exitCode
			}
			programs = append(programs, progs...)
		}
	}

afterProgramBuild:

	// Apply --rule CLI overrides if not already applied by one of the
	// config-load branches above. The guard (`cliRulesApplied`) makes
	// this a no-op when the branch already called the helper before
	// the fast-path check; it stays as a safety net so any future
	// config-load branch that forgets the early call still produces
	// correct lint output (just without the compat-only fast-path
	// opt-out for `--rule`-introduced native rules).
	if err := applyCLIRuleOverrides(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	// Use CWD for display paths (not any config directory).
	// In multi-config mode, currentDirectory was never reassigned from os.Getwd(),
	// so it already holds the normalized CWD.
	cwd := currentDirectory

	comparePathOptions := tspath.ComparePathsOptions{
		CurrentDirectory:          cwd,
		UseCaseSensitiveFileNames: fs.UseCaseSensitiveFileNames(),
	}

	// No args → implicit CWD scoping (same as `rslint .`).
	// Only applies to multi-config stdin path. In this mode, configs may include
	// parent or nested configs, so Programs may contain files outside CWD.
	// Without scoping, all those files would be linted unexpectedly.
	if len(allowFiles) == 0 && len(allowDirs) == 0 && args.ConfigStdin && configMap != nil {
		allowDirs = []string{cwd}
	}

	// --- Gap file discovery and fallback Program ---
	// Discover files that match config `files` patterns but are not in any
	// tsconfig Program. These "gap" files get a fallback Program (AST-only,
	// no type info) and only run non-type-aware rules.
	var typeInfoFiles map[string]struct{}
	var capturedGapFiles []string // retained for --fix rebuild
	fallbackProgramIndex := -1    // index of the gap-file fallback program (if any) within `programs`

	{
		programFiles := utils.CollectProgramFiles(programs, fs)

		var gapFiles []string
		if configMap != nil {
			gapFiles = rslintconfig.DiscoverGapFilesMultiConfig(configMap, fs, programFiles, allowFiles, allowDirs, args.SingleThreaded)
		} else {
			gapFiles = rslintconfig.DiscoverGapFiles(rslintConfig, currentDirectory, fs, programFiles, allowFiles, allowDirs, args.SingleThreaded)
		}

		// CLI file args bypass config `files` patterns (ESLint behavior):
		// if a user explicitly passes a file, lint it even if no config entry
		// has a matching `files` pattern.
		if gapFiles != nil && len(allowFiles) > 0 {
			gapSet := make(map[string]struct{}, len(gapFiles))
			for _, f := range gapFiles {
				gapSet[f] = struct{}{}
			}
			for _, f := range allowFiles {
				nf := tspath.NormalizePath(f)
				if _, inProgram := programFiles[nf]; inProgram {
					continue
				}
				if _, alreadyGap := gapSet[nf]; alreadyGap {
					continue
				}
				// Check if config would assign any rules to this file.
				var merged *rslintconfig.MergedConfig
				if configMap != nil {
					cfgDir, cfg := rslintconfig.FindNearestConfig(nf, configMap)
					if cfg != nil {
						merged = cfg.GetConfigForFile(nf, cfgDir)
					}
				} else {
					merged = rslintConfig.GetConfigForFile(nf, currentDirectory)
				}
				if merged != nil {
					gapFiles = append(gapFiles, nf)
				}
			}
		}

		if gapFiles != nil {
			// Build type-info set from existing (tsconfig) Programs BEFORE
			// appending the fallback, so fallback files are NOT in this set.
			typeInfoFiles = utils.CollectProgramFiles(programs, fs)
			capturedGapFiles = gapFiles

			if len(gapFiles) > 0 && !compatOnlyMode {
				// Build the fallback Program only when at least one native
				// rule will run on these files. In compat-only mode the
				// worker pool handles every file via oxc-parser without
				// touching the Go-side AST, so building this Program is
				// pure waste — the ts-go parse for thousands of files is
				// the dominant Go-side cost we want to drop.
				fallback, _ := createFallbackProgram(gapFiles, args.SingleThreaded, cwd, fs)
				if fallback != nil {
					programs = append(programs, fallback)
					fallbackProgramIndex = len(programs) - 1
				}
			}
		}
	}

	// createPrograms rebuilds programs (needed for multi-pass --fix re-linting).
	// Returns the program slice, the parallel `programConfigDirs` slice
	// keyed by program index (so callers can drive `buildFileFilters`
	// without falling back to the stale first-pass slice), and the
	// index of the fallback gap-file program (or -1 if none).
	//
	// The configDir iteration order MUST match the first-pass logic
	// at lines 1204-1226 — `sort.Strings(configDirs)` — because
	// `seenTsConfigs`/`seen` performs first-config-wins ownership
	// dedup. A re-pass that visits configDirs in Go's randomized map
	// order produces a `newPrograms` slice whose index ordering is
	// not parallel to the first-pass `programConfigDirs` snapshot.
	// `buildFileFilters(newPrograms, configMap, programConfigDirs, ...)`
	// then reads `programConfigDirs[i]` as the owner of `newPrograms[i]`
	// and mis-assigns files to the wrong config in shared-tsconfig
	// monorepos — silently flipping diagnostics on a --fix re-lint.
	// Re-sorting here keeps both runs aligned.
	createPrograms := func() ([]*compiler.Program, []string, int, error) {
		var baseProgs []*compiler.Program
		var baseProgDirs []string
		if configMap != nil {
			seen := make(map[string]struct{})
			configDirs := make([]string, 0, len(configMap))
			for configDir := range configMap {
				configDirs = append(configDirs, configDir)
			}
			sort.Strings(configDirs)
			for _, configDir := range configDirs {
				entries := configMap[configDir]
				progs, exitCode := createProgramsForConfig(configDir, entries, args.SingleThreaded, fs, seen)
				if exitCode != 0 {
					return nil, nil, -1, fmt.Errorf("failed to create programs for %s", configDir)
				}
				for range progs {
					baseProgDirs = append(baseProgDirs, configDir)
				}
				baseProgs = append(baseProgs, progs...)
			}
		} else {
			progs, exitCode := createProgramsForConfig(currentDirectory, rslintConfig, args.SingleThreaded, fs, nil)
			if exitCode != 0 {
				return nil, nil, -1, errors.New("failed to create programs")
			}
			for range progs {
				baseProgDirs = append(baseProgDirs, currentDirectory)
			}
			baseProgs = append(baseProgs, progs...)
		}

		// Rebuild fallback Program for gap files (content may have changed after fixes).
		fallbackIdx := -1
		if len(capturedGapFiles) > 0 {
			fallback, _ := createFallbackProgram(capturedGapFiles, args.SingleThreaded, cwd, fs)
			if fallback != nil {
				baseProgs = append(baseProgs, fallback)
				// Fallback program has no owning config; mark it with an
				// empty string so the parallel slice stays correctly
				// aligned by index. buildFileFilters' caller-side logic
				// treats empty owner as "no per-config ownership filter".
				baseProgDirs = append(baseProgDirs, "")
				fallbackIdx = len(baseProgs) - 1
			}
		}

		return baseProgs, baseProgDirs, fallbackIdx, nil
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

	enforcePlugins := args.ConfigStdin // JS/TS configs loaded via stdin require plugin declarations
	getRulesForFile := func(sourceFile *ast.SourceFile) []linter.ConfiguredRule {
		filePath := sourceFile.FileName()
		if configMap != nil {
			cfgDir, cfg := rslintconfig.FindNearestConfig(filePath, configMap)
			if cfg == nil {
				return nil
			}
			return rslintconfig.GlobalRuleRegistry.GetActiveRulesForFile(cfg, filePath, cfgDir, enforcePlugins, typeInfoFiles)
		}
		return rslintconfig.GlobalRuleRegistry.GetActiveRulesForFile(rslintConfig, filePath, currentDirectory, enforcePlugins, typeInfoFiles)
	}

	// Build per-program file filters combining:
	//   - multi-config ownership deduplication (ESLint v10 aligned)
	//   - config `ignores` exclusion (applies to rules and counts)
	fileFilters := buildFileFilters(programs, configMap, programConfigDirs, rslintConfig, currentDirectory)

	// fallback gap program (if any) is excluded from --type-check: its
	// CompilerOptions are synthesized defaults, not the user's tsconfig,
	// so semantic diagnostics there would be unreliable. This honors the
	// commitment in website/docs/en/guide/type-checking.md ("Files Without
	// tsconfig Coverage").
	skipTypeCheck := buildTypeCheckSkipMask(len(programs), fallbackProgramIndex)

	// In --type-check-only mode, skip the lint phase entirely by passing
	// nil for GetRulesForFile. RunLinter's Phase 1 is gated on this being
	// non-nil; Phase 2 (type-check) runs independently.
	var rulesForFile linter.RuleHandler
	if !args.TypeCheckOnly {
		rulesForFile = getRulesForFile
	}

	// Suggestions mode: enabled in jsonline (consumers are programmatic and
	// want every actionable hint); off in default/github (the human-readable
	// summary doesn't render suggestions).
	suggestionsMode := "off"
	if args.Format == "jsonline" {
		suggestionsMode = "eager"
	}

	var lintResult *linter.LintResult
	// Set populated only in compatOnlyMode; used by
	// collectAllowFileWarnings below so a CLI-specified file that
	// actually went through DispatchCompat is NOT mis-reported as
	// "not found in the project". Outside compat-only mode the set
	// stays nil and the warning logic falls back to its existing
	// program-files oracle.
	// dispatchCompatPass builds the compat file inputs and dispatches them to
	// the worker pool — the compat-only fast path shared by Phase 1 and every
	// --fix re-lint pass (only the OnDiagnostic sink differs). It bypasses
	// RunLinter (which would iterate a ts-go Program we never built). Returns
	// the file set so Phase 1 can derive compatLintedFiles.
	dispatchCompatPass := func(
		onDiag linter.DiagnosticHandler,
	) (*linter.LintResult, []linter.CompatFileEntry, error) {
		files := buildCompatFileInputs(
			capturedGapFiles, allowFiles, allowDirs,
			rslintConfig, configMap, currentDirectory, enforcePlugins, fs,
		)
		res, dErr := linter.DispatchCompat(linter.DispatchCompatOptions{
			Files:           files,
			Dispatcher:      compatDispatcher,
			OnDiagnostic:    onDiag,
			CollectFixes:    args.Fix,
			SuggestionsMode: suggestionsMode,
			Ctx:             ctx,
		})
		return res, files, dErr
	}

	var compatLintedFiles map[string]struct{}
	if compatOnlyMode {
		var files []linter.CompatFileEntry
		lintResult, files, err = dispatchCompatPass(
			func(d rule.RuleDiagnostic) { diagnosticsChan <- d },
		)
		compatLintedFiles = make(map[string]struct{}, len(files))
		for _, f := range files {
			compatLintedFiles[f.Path] = struct{}{}
		}
	} else {
		lintResult, err = linter.RunLinter(linter.RunLinterOptions{
			Programs:              programs,
			SingleThreaded:        args.SingleThreaded,
			Scope:                 linter.FileScope{Files: allowFiles, Dirs: allowDirs},
			PerProgramFilter:      toFileFilters(fileFilters),
			GetRulesForFile:       rulesForFile,
			TypeInfoFiles:         typeInfoFiles,
			TypeCheck:             args.TypeCheck,
			SkipTypeCheckPrograms: skipTypeCheck,
			OnDiagnostic: func(d rule.RuleDiagnostic) {
				diagnosticsChan <- d
			},
			CompatRuleDispatcher: compatDispatcher,
			CollectFixes:         args.Fix,
			SuggestionsMode:      suggestionsMode,
			Ctx:                  ctx,
		})
	}

	close(diagnosticsChan)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error running linter: %v\n", err)
		return 1
	}

	lintedfileCount := lintResult.LintedFileCount

	wg.Wait()

	// Emit per-file warnings for CLI-specified files that won't be linted.
	// Distinguishes "not found in project" vs "ignored by pattern", aligned
	// with ESLint v10's warning behavior. Skipped in --type-check-only mode:
	// these are lint-phase concepts and would mislead users about Phase 2
	// (which runs program-wide regardless of CLI scope and rslint ignores).
	if !args.TypeCheckOnly {
		warnings := collectAllowFileWarnings(allowFiles, programs, compatLintedFiles, configMap, rslintConfig, currentDirectory)
		for _, w := range warnings {
			fmt.Fprintln(os.Stderr, formatAllowFileWarning(w, comparePathOptions))
		}
	}
	scopeRestricted := len(allowFiles) > 0 || len(allowDirs) > 0
	if shouldShortCircuitOutput(args.TypeCheckOnly, args.TypeCheck, scopeRestricted, lintedfileCount) {
		return 0
	}

	// Phase 2: Apply fixes if --fix flag is enabled.
	// Uses multi-pass fixing: after applying fixes, rebuild programs and re-lint
	// to catch cascading issues (e.g. ban-types fix triggers no-inferrable-types).
	// After fixing, allDiags is replaced with remaining (unfixed) diagnostics.
	//
	// compatOnlyMode skips ts-go program rebuilds. When ALL active rules
	// are plugin rules (no native rule needs Phase 2's typechecker), we
	// reuse the same DispatchCompat path Phase 1 took:
	//   - Files (capturedGapFiles + allowFiles + allowDirs) are stable
	//     across passes — fixes change FILE CONTENTS, not the file SET.
	//   - The worker re-reads disk via `readFileSync(filePath, 'utf8')`
	//     on every dispatch, so the second pass naturally sees fixed
	//     text without any host-side coordination.
	//   - `buildCompatFileInputs` is pure over its inputs, so calling it
	//     once per pass is cheap (no I/O on the host side; only the
	//     allow-file gate + rule resolution).
	// This eliminates the createPrograms / RunLinter cost per pass —
	// dominant overhead in all-plugin configs where Phase 1 already
	// skipped program build for the same reason.
	const maxFixPasses = 10
	if args.Fix && len(allDiags) > 0 {
		diagnosticsByFile := groupDiagsByFile(allDiags)
		fixedCount += applyFixPass(diagnosticsByFile)

		// Re-lint → fix → re-lint → fix → ... until stable or maxFixPasses.
		// Skip if nothing was fixed in the first pass (no need to re-lint).
		for pass := 1; pass < maxFixPasses && fixedCount > 0; pass++ {
			var passDiags []rule.RuleDiagnostic
			onDiag := func(d rule.RuleDiagnostic) {
				diagsMu.Lock()
				passDiags = append(passDiags, d)
				diagsMu.Unlock()
			}
			var passResult *linter.LintResult

			if compatOnlyMode {
				// Compat-only fast path — same dispatch as Phase 1 (no ts-go
				// programs to rebuild; the worker pool handles everything).
				// capturedGapFiles is Phase 1's discovered set; fixes change
				// content, not set membership, so reusing it is correct.
				passResult, _, _ = dispatchCompatPass(onDiag)
			} else {
				// Native path: a native rule depends on type info, so we
				// rebuild programs (slow but unavoidable — fixes can change
				// types that downstream rules see).
				newPrograms, newProgramConfigDirs, newFallbackIdx, err := createPrograms()
				if err != nil || len(newPrograms) == 0 {
					break
				}
				// IMPORTANT: pass `newProgramConfigDirs`, NOT the
				// first-pass `programConfigDirs` snapshot. The latter
				// was built from a (now sorted) initial pass; the
				// retry's program slice is also sorted (createPrograms
				// re-sorts), so the indices align — but only if we
				// use the freshly-emitted parallel slice. Falling
				// back to the cached first-pass slice was the bug
				// flagged in review G2.
				fixFileFilters := buildFileFilters(newPrograms, configMap, newProgramConfigDirs, rslintConfig, currentDirectory)
				fixSkipMask := buildTypeCheckSkipMask(len(newPrograms), newFallbackIdx)
				passResult, _ = linter.RunLinter(linter.RunLinterOptions{
					Programs:              newPrograms,
					SingleThreaded:        args.SingleThreaded,
					Scope:                 linter.FileScope{Files: allowFiles, Dirs: allowDirs},
					PerProgramFilter:      toFileFilters(fixFileFilters),
					GetRulesForFile:       getRulesForFile,
					TypeInfoFiles:         typeInfoFiles,
					TypeCheck:             args.TypeCheck,
					SkipTypeCheckPrograms: fixSkipMask,
					OnDiagnostic:          onDiag,
					CompatRuleDispatcher:  compatDispatcher,
					CollectFixes:          args.Fix,
					SuggestionsMode:       suggestionsMode,
					Ctx:                   ctx,
				})
			}
			if passResult != nil {
				for name := range passResult.ExecutedRules {
					lintResult.ExecutedRules[name] = struct{}{}
				}
				// Carry the re-lint pass's compat dispatch failures into the
				// aggregate the exit code reads (the post-fix state), not the
				// stale pre-fix value — previously dropped, so a worker
				// failure during the fix pass produced a false success exit.
				lintResult.CompatDispatchErrors = passResult.CompatDispatchErrors
			}

			// Replace allDiags with latest post-fix diagnostics.
			allDiags = passDiags

			passFixed := applyFixPass(groupDiagsByFile(passDiags))
			if passFixed == 0 {
				break // Stable — allDiags reflect final state
			}
			fixedCount += passFixed
		}
	}

	// Phase 3: Print diagnostics and count errors/warnings.
	// allDiags contains: original diagnostics (no fix), or remaining after fix.
	errorsCount := 0
	warningsCount := 0
	typeErrorsCount := 0
	{
		w := bufio.NewWriterSize(os.Stdout, 4096*100)
		for i, d := range allDiags {
			switch d.Severity {
			case rule.SeverityError:
				errorsCount++
				if args.TypeCheck && strings.HasPrefix(d.RuleName, "TypeScript(") {
					typeErrorsCount++
				}
			case rule.SeverityWarning:
				warningsCount++
			}

			if i == 0 {
				w.WriteByte('\n')
			}
			// Only print Error message when quiet is true
			if args.Quiet && d.Severity != rule.SeverityError {
				continue
			}
			printDiagnostic(d, w, comparePathOptions, args.Format)
			if w.Available() < 4096 {
				w.Flush()
			}
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
	if !args.SingleThreaded {
		threadsCount = runtime.GOMAXPROCS(0)
	}
	if args.Format == "default" {
		// Build the errors summary part.
		// When type-check is enabled and there are type errors, split the display.
		var errorsSummary string
		switch {
		case args.TypeCheckOnly:
			// Lint phase was skipped; only type errors are possible.
			errorsSummary = fmt.Sprintf("%s %s",
				errorsColorFunc("%d", typeErrorsCount),
				pluralize(typeErrorsCount, "type error", "type errors"),
			)
		case args.TypeCheck:
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

		if args.TypeCheckOnly {
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
				color.New().SprintFunc()(""),
			)
		} else if args.Fix && fixedCount > 0 {
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
				color.New().SprintFunc()(""), // Reset
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
				color.New().SprintFunc()(""), // Reset
			)
		}
	}

	tooManyWarnings := args.MaxWarnings >= 0 && warningsCount > args.MaxWarnings

	if errorsCount == 0 && tooManyWarnings {
		fmt.Fprintf(os.Stderr, "Rslint found too many warnings (maximum: %d).\n", args.MaxWarnings)
	}

	// Compat dispatcher failures (Node parent crash, IPC timeout, LSP
	// client error, plugin import error) are runner-level failures, not
	// lint findings. The error is already logged to stderr; exit 2
	// (distinct from the lint-error exit 1) so the user notices the infra
	// problem rather than debugging "missing diagnostics", and CI can tell
	// a flaky runner apart from real findings.
	if lintResult.CompatDispatchErrors > 0 {
		fmt.Fprintf(os.Stderr,
			"Rslint: compat dispatcher failed for %d program(s).\n",
			lintResult.CompatDispatchErrors)
		return 2
	}

	// Exit with non-zero status code if errors were found
	if errorsCount > 0 || tooManyWarnings {
		return 1
	}
	return 0
}
