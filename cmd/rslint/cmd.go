package main

import (
	"bufio"
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
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
)

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

	filePath := tspath.ConvertToRelativePath(d.SourceFile.FileName(), comparePathOptions)
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
		FilePath: tspath.ConvertToRelativePath(d.SourceFile.FileName(), comparePathOptions),
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
	filePath := tspath.ConvertToRelativePath(d.SourceFile.FileName(), comparePathOptions)
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

	// Iterate by runes to correctly handle multi-byte UTF-8 characters,
	// but track byte positions for string slicing
	codeboxText := text[codeboxStart:codeboxEnd]
	bytePos := codeboxStart
	for _, char := range codeboxText {
		charBytes := utf8.RuneLen(char)
		current, next := bytePos, bytePos+charBytes
		bytePos = next

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

func (f *repeatedFlag) String() string   { return strings.Join(*f, ", ") }
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
		f := d.SourceFile.FileName()
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

func runCMD() int {
	flag.Usage = func() { fmt.Fprint(os.Stderr, usage) }

	var (
		init        bool
		help        bool
		config      string
		configStdin bool
		fix         bool
		typeCheck   bool

		traceOut       string
		cpuprofOut     string
		singleThreaded bool
		format         string
		noColor        bool
		forceColor     bool
		quiet          bool
		maxWarnings    int
		startTimeMs    int64
		ruleFlags      repeatedFlag
	)
	flag.StringVar(&format, "format", "default", "output format")
	flag.StringVar(&config, "config", "", "which rslint config to use")
	flag.BoolVar(&configStdin, "config-stdin", false, "read config from stdin (used internally by JS config loader)")
	flag.BoolVar(&init, "init", false, "initialize a default config in the current directory")
	flag.BoolVar(&fix, "fix", false, "automatically fix problems")
	flag.BoolVar(&typeCheck, "type-check", false, "enable TypeScript type checking")
	flag.BoolVar(&help, "help", false, "show help")
	flag.BoolVar(&help, "h", false, "show help")
	flag.BoolVar(&noColor, "no-color", false, "disable colored output")
	flag.BoolVar(&forceColor, "force-color", false, "force colored output")
	flag.BoolVar(&quiet, "quiet", false, "report errors only")
	flag.IntVar(&maxWarnings, "max-warnings", -1, "Number of warnings to trigger nonzero exit code")

	flag.StringVar(&traceOut, "trace", "", "file to put trace to")
	flag.StringVar(&cpuprofOut, "cpuprof", "", "file to put cpu profiling to")
	flag.BoolVar(&singleThreaded, "singleThreaded", false, "run in single threaded mode")
	flag.Int64Var(&startTimeMs, "start-time", 0, "internal: epoch milliseconds from Node.js entry point")
	flag.Var(&ruleFlags, "rule", "rule override, e.g. 'no-console: error' (repeatable)")

	flag.Parse()

	// Collect file/directory arguments for targeted linting (e.g. rslint file1.ts src/)
	var allowFiles []string
	var allowDirs []string
	if args := flag.Args(); len(args) > 0 {
		for _, arg := range args {
			absPath, err := filepath.Abs(arg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error resolving path %s: %v\n", arg, err)
				return 1
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
			// and CollectProgramFiles's Realpath'd keys in create_program.go.
			normalized := tspath.NormalizePath(absPath)
			info, statErr := os.Stat(absPath)
			if statErr == nil && info.IsDir() {
				allowDirs = append(allowDirs, normalized)
			} else {
				allowFiles = append(allowFiles, normalized)
			}
		}
	}

	if help {
		flag.Usage()
		return 0
	}

	// Override color detection based on flags
	if noColor {
		color.NoColor = true
	}
	if forceColor {
		color.NoColor = false
	}

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

	if init {
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

		payload, parseErr := parseConfigPayload(data)
		if parseErr != nil {
			fmt.Fprintf(os.Stderr, "%v\n", parseErr)
			return 1
		}

		if payload.IsMultiConfig {
			// Multi-config format
			configMap = payload.ConfigMap

			// Inject .gitignore patterns as global ignores for each config.
			// Each config independently reads its own .gitignore tree:
			// ReadGitignoreAsGlobs walks UP (ancestor inheritance) and DOWN
			// (nested .gitignore) from each configDir. Sibling configs are
			// fully isolated — they never share gitignore patterns.
			//
			// Config ignores are passed so that directories which are
			// directory-level blocked (e.g. **/tests/**) are pruned during
			// the .gitignore scan. This is safe because isDirPathBlocked is
			// the same function used by the linter — blocked dirs' files
			// are never linted, so their .gitignore patterns are irrelevant.
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
			if singleThreaded {
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

			for configDir, entries := range configMap {
				progs, exitCode := createProgramsForConfig(configDir, entries, singleThreaded, fs, seenTsConfigs)
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

			// Inject .gitignore patterns as global ignores. Run gitignore
			// reading in parallel with createProgramsForConfig — they're
			// independent (createProgramsForConfig only reads
			// languageOptions.parserOptions.project, not Ignores).
			var (
				progs    []*compiler.Program
				exitCode int
			)
			rslintConfig, progs, exitCode = parallelGitignoreAndPrograms(
				rslintConfig, currentDirectory, fs, singleThreaded, nil,
			)
			if exitCode != 0 {
				return exitCode
			}
			programs = append(programs, progs...)
		}
	} else {
		// Load configuration from file (JSON config path, isJSConfig stays false)
		rslintConfig, _, currentDirectory = rslintconfig.LoadConfigurationWithFallback(config, currentDirectory, fs)

		// Inject .gitignore patterns as global ignores. Run gitignore reading
		// in parallel with createProgramsForConfig (see comment above).
		var (
			progs    []*compiler.Program
			exitCode int
		)
		rslintConfig, progs, exitCode = parallelGitignoreAndPrograms(
			rslintConfig, currentDirectory, fs, singleThreaded, nil,
		)
		if exitCode != 0 {
			return exitCode
		}
		programs = append(programs, progs...)
	}

	// Apply --rule CLI overrides by appending a synthetic ConfigEntry (no files = matches all).
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
	cwd := currentDirectory

	comparePathOptions := tspath.ComparePathsOptions{
		CurrentDirectory:          cwd,
		UseCaseSensitiveFileNames: fs.UseCaseSensitiveFileNames(),
	}

	// No args → implicit CWD scoping (same as `rslint .`).
	// Only applies to multi-config stdin path. In this mode, configs may include
	// parent or nested configs, so Programs may contain files outside CWD.
	// Without scoping, all those files would be linted unexpectedly.
	if len(allowFiles) == 0 && len(allowDirs) == 0 && configStdin && configMap != nil {
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
			gapFiles = rslintconfig.DiscoverGapFilesMultiConfig(configMap, fs, programFiles, allowFiles, allowDirs, singleThreaded)
		} else {
			gapFiles = rslintconfig.DiscoverGapFiles(rslintConfig, currentDirectory, fs, programFiles, allowFiles, allowDirs, singleThreaded)
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

			if len(gapFiles) > 0 {
				fallback, _ := createFallbackProgram(gapFiles, singleThreaded, cwd, fs)
				if fallback != nil {
					programs = append(programs, fallback)
					fallbackProgramIndex = len(programs) - 1
				}
			}
		}
	}

	// createPrograms rebuilds programs (needed for multi-pass --fix re-linting).
	// Returns the program slice and the index of the fallback gap-file program
	// (or -1 if none), so callers can mark it skipped during type-check.
	createPrograms := func() ([]*compiler.Program, int, error) {
		var baseProgs []*compiler.Program
		if configMap != nil {
			seen := make(map[string]struct{})
			for configDir, entries := range configMap {
				progs, exitCode := createProgramsForConfig(configDir, entries, singleThreaded, fs, seen)
				if exitCode != 0 {
					return nil, -1, fmt.Errorf("failed to create programs for %s", configDir)
				}
				baseProgs = append(baseProgs, progs...)
			}
		} else {
			progs, exitCode := createProgramsForConfig(currentDirectory, rslintConfig, singleThreaded, fs, nil)
			if exitCode != 0 {
				return nil, -1, errors.New("failed to create programs")
			}
			baseProgs = append(baseProgs, progs...)
		}

		// Rebuild fallback Program for gap files (content may have changed after fixes).
		fallbackIdx := -1
		if len(capturedGapFiles) > 0 {
			fallback, _ := createFallbackProgram(capturedGapFiles, singleThreaded, cwd, fs)
			if fallback != nil {
				baseProgs = append(baseProgs, fallback)
				fallbackIdx = len(baseProgs) - 1
			}
		}

		return baseProgs, fallbackIdx, nil
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

	lintResult, err := linter.RunLinter(linter.RunLinterOptions{
		Programs:              programs,
		SingleThreaded:        singleThreaded,
		Scope:                 linter.FileScope{Files: allowFiles, Dirs: allowDirs},
		PerProgramFilter:      toFileFilters(fileFilters),
		GetRulesForFile:       getRulesForFile,
		TypeInfoFiles:         typeInfoFiles,
		TypeCheck:             typeCheck,
		SkipTypeCheckPrograms: skipTypeCheck,
		OnDiagnostic: func(d rule.RuleDiagnostic) {
			diagnosticsChan <- d
		},
	})

	close(diagnosticsChan)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error running linter: %v\n", err)
		return 1
	}

	lintedfileCount := lintResult.LintedFileCount

	wg.Wait()

	// Emit per-file warnings for CLI-specified files that won't be linted.
	// Distinguish between "ignored by pattern" vs "no matching configuration"
	// vs "not found in project", aligned with ESLint v10's warning behavior.
	if len(allowFiles) > 0 {
		programFiles := make(map[string]struct{})
		for _, prog := range programs {
			for _, sf := range prog.GetSourceFiles() {
				programFiles[sf.FileName()] = struct{}{}
			}
		}
		// Cache FindNearestConfig results by directory to avoid redundant lookups
		// when many files are in the same directory (e.g., lint-staged).
		type cachedConfig struct {
			cfgDir string
			cfg    rslintconfig.RslintConfig
		}
		dirConfigCache := make(map[string]*cachedConfig)

		for _, f := range allowFiles {
			_, inProgram := programFiles[f]

			if !inProgram {
				// File not in any Program — warn and skip
				relPath := tspath.ConvertToRelativePath(f, comparePathOptions)
				fmt.Fprintf(os.Stderr, "warning: %s was not found in the project, skipping\n", relPath)
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
				relPath := tspath.ConvertToRelativePath(f, comparePathOptions)
				fmt.Fprintf(os.Stderr, "warning: %s is ignored because of a matching ignore pattern\n", relPath)
			}
		}
	}
	if (len(allowFiles) > 0 || len(allowDirs) > 0) && lintedfileCount == 0 {
		return 0
	}

	// Phase 2: Apply fixes if --fix flag is enabled.
	// Uses multi-pass fixing: after applying fixes, rebuild programs and re-lint
	// to catch cascading issues (e.g. ban-types fix triggers no-inferrable-types).
	// After fixing, allDiags is replaced with remaining (unfixed) diagnostics.
	const maxFixPasses = 10
	if fix && len(allDiags) > 0 {
		diagnosticsByFile := groupDiagsByFile(allDiags)
		fixedCount += applyFixPass(diagnosticsByFile)

		// Re-lint → fix → re-lint → fix → ... until stable or maxFixPasses.
		// Skip if nothing was fixed in the first pass (no need to re-lint).
		for pass := 1; pass < maxFixPasses && fixedCount > 0; pass++ {
			newPrograms, newFallbackIdx, err := createPrograms()
			if err != nil || len(newPrograms) == 0 {
				break
			}

			// Re-lint: collect remaining diagnostics.
			// Rebuild file filters for the new programs (ownership + ignores).
			fixFileFilters := buildFileFilters(newPrograms, configMap, programConfigDirs, rslintConfig, currentDirectory)
			fixSkipMask := buildTypeCheckSkipMask(len(newPrograms), newFallbackIdx)
			var passDiags []rule.RuleDiagnostic
			passResult, _ := linter.RunLinter(linter.RunLinterOptions{
				Programs:              newPrograms,
				SingleThreaded:        singleThreaded,
				Scope:                 linter.FileScope{Files: allowFiles, Dirs: allowDirs},
				PerProgramFilter:      toFileFilters(fixFileFilters),
				GetRulesForFile:       getRulesForFile,
				TypeInfoFiles:         typeInfoFiles,
				TypeCheck:             typeCheck,
				SkipTypeCheckPrograms: fixSkipMask,
				OnDiagnostic: func(d rule.RuleDiagnostic) {
					diagsMu.Lock()
					passDiags = append(passDiags, d)
					diagsMu.Unlock()
				},
			})
			if passResult != nil {
				for name := range passResult.ExecutedRules {
					lintResult.ExecutedRules[name] = struct{}{}
				}
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
			printDiagnostic(d, w, comparePathOptions, format)
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
	if !singleThreaded {
		threadsCount = runtime.GOMAXPROCS(0)
	}
	if format == "default" {
		// Build the errors summary part.
		// When type-check is enabled and there are type errors, split the display.
		var errorsSummary string
		if typeCheck {
			errorsSummary = fmt.Sprintf("%s %s, %s %s",
				errorsColorFunc("%d", lintErrorsCount),
				pluralize(lintErrorsCount, "lint error", "lint errors"),
				errorsColorFunc("%d", typeErrorsCount),
				pluralize(typeErrorsCount, "type error", "type errors"),
			)
		} else {
			errorsSummary = fmt.Sprintf("%s %s",
				errorsColorFunc("%d", errorsCount),
				pluralize(errorsCount, "error", "errors"),
			)
		}

		if fix && fixedCount > 0 {
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
