package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
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

const spaces = "                                                                                                    "

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

	startLine, startColumn := scanner.GetECMALineAndCharacterOfPosition(d.SourceFile, diagnosticStart)
	endLine, endColumn := scanner.GetECMALineAndCharacterOfPosition(d.SourceFile, diagnosticEnd)

	filePath := tspath.ConvertToRelativePath(d.SourceFile.FileName(), comparePathOptions)
	output := fmt.Sprintf(
		"::%s file=%s,line=%d,endLine=%d,col=%d,endColumn=%d,title=%s::%s\n",
		severity,
		escapeProperty(filePath),
		startLine+1,
		endLine+1,
		startColumn+1,
		endColumn+1,
		d.RuleName,
		escapeData(d.Message.Description),
	)
	w.WriteString(output)
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

	startLine, startColumn := scanner.GetECMALineAndCharacterOfPosition(d.SourceFile, diagnosticStart)
	endLine, endColumn := scanner.GetECMALineAndCharacterOfPosition(d.SourceFile, diagnosticEnd)

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
				Column: startColumn + 1,
			},
			End: Location{
				Line:   endLine + 1,
				Column: endColumn + 1,
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

	diagnosticStartLine, diagnosticStartColumn := scanner.GetECMALineAndCharacterOfPosition(d.SourceFile, diagnosticStart)
	diagnosticEndline, _ := scanner.GetECMALineAndCharacterOfPosition(d.SourceFile, diagnosticEnd)

	lineMap := scanner.GetECMALineStarts(d.SourceFile)
	text := d.SourceFile.Text()

	codeboxStartLine := max(diagnosticStartLine-1, 0)
	codeboxEndLine := min(diagnosticEndline+1, len(lineMap)-1)

	codeboxStart := scanner.GetECMAPositionOfLineAndCharacter(d.SourceFile, codeboxStartLine, 0)
	var codeboxEndColumn int
	if codeboxEndLine == len(lineMap)-1 {
		codeboxEndColumn = len(text) - int(lineMap[len(lineMap)-1])
	} else {
		codeboxEndColumn = int(lineMap[codeboxEndLine+1]-lineMap[codeboxEndLine]) - 1
	}
	codeboxEnd := scanner.GetECMAPositionOfLineAndCharacter(d.SourceFile, codeboxEndLine, codeboxEndColumn)

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

	// Message handling
	messageLineStart := 0
	for i, char := range d.Message.Description {
		if char == '\n' {
			w.WriteString(d.Message.Description[messageLineStart : i+1])
			messageLineStart = i + 1
			w.WriteString("    ")
			w.WriteString(colors.BorderText("│"))
			w.WriteString(spaces[:len(d.RuleName)+1])
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

	lineStarts := make([]int, 13)
	lineEnds := make([]int, 13)

	if codeboxEndLine-codeboxStartLine >= len(lineEnds) {
		w.WriteString("  ")
		w.WriteString(colors.BorderText("│"))
		w.WriteString("  Error range is too big. Skipping code block printing.\n  ")
		w.WriteString(colors.BorderText("╰────────────────────────────────"))
		w.WriteByte('\n')
		w.WriteByte('\n')
		return
	}

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

	for line := codeboxStartLine; line <= codeboxEndLine; line++ {
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

const usage = `🚀 Rslint - Rocket Speed Linter

Usage:
  rslint [OPTIONS]

Options:
  --init				Initialize a default config in the current directory.
  --config PATH         Which rslint config file to use. Defaults to rslint.json.
  --format FORMAT       Output format: default | jsonline | github
  --fix                 Automatically fix problems
  --no-color            Disable colored output
  --force-color         Force colored output
  --quiet               Report errors only 
  --max-warnings Int    Number of warnings to trigger nonzero exit code
  -h, --help            Show help
`

func runCMD() int {
	flag.Usage = func() { fmt.Fprint(os.Stderr, usage) }

	var (
		init   bool
		help   bool
		config string
		fix    bool

		traceOut       string
		cpuprofOut     string
		singleThreaded bool
		format         string
		noColor        bool
		forceColor     bool
		quiet          bool
		maxWarnings    int
	)
	flag.StringVar(&format, "format", "default", "output format")
	flag.StringVar(&config, "config", "", "which rslint config to use")
	flag.BoolVar(&init, "init", false, "initialize a default config in the current directory")
	flag.BoolVar(&fix, "fix", false, "automatically fix problems")
	flag.BoolVar(&help, "help", false, "show help")
	flag.BoolVar(&help, "h", false, "show help")
	flag.BoolVar(&noColor, "no-color", false, "disable colored output")
	flag.BoolVar(&forceColor, "force-color", false, "force colored output")
	flag.BoolVar(&quiet, "quiet", false, "report errors only")
	flag.IntVar(&maxWarnings, "max-warnings", -1, "Number of warnings to trigger nonzero exit code")

	flag.StringVar(&traceOut, "trace", "", "file to put trace to")
	flag.StringVar(&cpuprofOut, "cpuprof", "", "file to put cpu profiling to")
	flag.BoolVar(&singleThreaded, "singleThreaded", false, "run in single threaded mode")

	flag.Parse()

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
	timeBefore := time.Now()

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
	var tsConfigs []string
	// Load rslint configuration and determine which rules to enable
	rslintConfig, tsConfigs, currentDirectory = rslintconfig.LoadConfigurationWithFallback(config, currentDirectory, fs)

	host := utils.CreateCompilerHost(currentDirectory, fs)

	comparePathOptions := tspath.ComparePathsOptions{
		CurrentDirectory:          host.GetCurrentDirectory(),
		UseCaseSensitiveFileNames: host.FS().UseCaseSensitiveFileNames(),
	}
	programs := []*compiler.Program{}
	for _, configFileName := range tsConfigs {
		program, err := utils.CreateProgram(singleThreaded, fs, currentDirectory, configFileName, host)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating TS program: %v", err)
			return 1
		}
		programs = append(programs, program)

	}

	var wg sync.WaitGroup

	diagnosticsChan := make(chan rule.RuleDiagnostic, 4096)
	errorsCount := 0
	warningsCount := 0
	fixedCount := 0

	// Store diagnostics by file for fixing
	var diagnosticsByFile map[string][]rule.RuleDiagnostic
	if fix {
		diagnosticsByFile = make(map[string][]rule.RuleDiagnostic)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		w := bufio.NewWriterSize(os.Stdout, 4096*100)
		defer w.Flush()
		for d := range diagnosticsChan {
			switch d.Severity {
			case rule.SeverityError:
				errorsCount++
			case rule.SeverityWarning:
				warningsCount++
			}

			// Store diagnostics by file for fixing
			if fix {
				fileName := d.SourceFile.FileName()
				diagnosticsByFile[fileName] = append(diagnosticsByFile[fileName], d)
			}

			if errorsCount+warningsCount == 1 {
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
	}()

	lintedfileCount, err := linter.RunLinter(
		programs,
		singleThreaded,
		nil,
		utils.ExcludePaths,

		func(sourceFile *ast.SourceFile) []linter.ConfiguredRule {
			activeRules := rslintconfig.GlobalRuleRegistry.GetEnabledRules(rslintConfig, sourceFile.FileName())
			return activeRules
		},
		func(d rule.RuleDiagnostic) {
			diagnosticsChan <- d
		},
	)

	close(diagnosticsChan)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error running linter: %v\n", err)
		return 1
	}

	wg.Wait()

	// Apply fixes if --fix flag is enabled
	if fix && len(diagnosticsByFile) > 0 {
		for fileName, fileDiagnostics := range diagnosticsByFile {
			// Only apply fixes for diagnostics that have fixes
			diagnosticsWithFixes := make([]rule.RuleDiagnostic, 0)
			for _, d := range fileDiagnostics {
				if len(d.Fixes()) > 0 {
					diagnosticsWithFixes = append(diagnosticsWithFixes, d)
				}
			}

			if len(diagnosticsWithFixes) > 0 {
				// Read the original file content
				originalContent := diagnosticsWithFixes[0].SourceFile.Text()

				// Apply fixes
				fixedContent, unapplied, wasFixed := linter.ApplyRuleFixes(originalContent, diagnosticsWithFixes)

				if wasFixed {
					// Write the fixed content back to the file
					err := os.WriteFile(fileName, []byte(fixedContent), 0644)
					if err != nil {
						fmt.Fprintf(os.Stderr, "error writing fixed file %s: %v\n", fileName, err)
					} else {
						fixedCount += len(diagnosticsWithFixes) - len(unapplied)
					}
				}
			}
		}
	}

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

	errorsText := "errors"
	if errorsCount == 1 {
		errorsText = "error"
	}

	warningsText := "warnings"
	if warningsCount == 1 {
		warningsText = "warning"
	}

	filesText := "files"
	if lintedfileCount == 1 {
		filesText = "file"
	}
	threadsCount := 1
	if !singleThreaded {
		threadsCount = runtime.GOMAXPROCS(0)
	}
	if format == "default" {
		if fix && fixedCount > 0 {
			fixText := "issues"
			if fixedCount == 1 {
				fixText = "issue"
			}
			fmt.Fprintf(
				os.Stdout,
				"Found %s %s and %s %s %s(linted %s %s in %s using %s threads, fixed %s %s)%s\n",
				errorsColorFunc("%d", errorsCount),
				errorsText,
				warningsColorFunc("%d", warningsCount),
				warningsText,
				colors.DimText(""),
				colors.BoldText("%d", lintedfileCount),
				filesText,
				colors.BoldText("%v", time.Since(timeBefore).Round(time.Millisecond)),
				colors.BoldText("%d", threadsCount),
				colors.SuccessText("%d", fixedCount),
				fixText,
				color.New().SprintFunc()(""), // Reset
			)
		} else {
			fmt.Fprintf(
				os.Stdout,
				"Found %s %s and %s %s %s(linted %s %s with in %s using %s threads)%s\n",
				errorsColorFunc("%d", errorsCount),
				errorsText,
				warningsColorFunc("%d", warningsCount),
				warningsText,
				colors.DimText(""),
				colors.BoldText("%d", lintedfileCount),
				filesText,
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
