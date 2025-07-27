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
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/fatih/color"
	"github.com/typescript-eslint/rslint/internal/linter"
	"github.com/typescript-eslint/rslint/internal/rule"
	"github.com/typescript-eslint/rslint/internal/utils"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	rslintconfig "github.com/typescript-eslint/rslint/internal/config"
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
	errorTextColor := color.New(color.FgRed, color.Underline).SprintfFunc()
	successColor := color.New(color.FgGreen, color.Bold).SprintfFunc()
	dimColor := color.New(color.Faint).SprintfFunc()
	boldColor := color.New(color.Bold).SprintfFunc()
	borderColor := color.New(color.Faint).SprintfFunc()

	return &ColorScheme{
		RuleName:    ruleNameColor,
		FileName:    fileNameColor,
		ErrorText:   errorTextColor,
		SuccessText: successColor,
		DimText:     dimColor,
		BoldText:    boldColor,
		BorderText:  borderColor,
	}
}

func printDiagnostic(d rule.RuleDiagnostic, w *bufio.Writer, comparePathOptions tspath.ComparePathsOptions, format string) {
	switch format {
	case "default":
		printDiagnosticDefault(d, w, comparePathOptions)
	case "jsonline":
		printDiagnosticJsonLine(d, w, comparePathOptions)
	default:
		panic(fmt.Sprintf("not supported format %s", format))
	}
}

// print as [jsonline](https://jsonlines.org/) format which can be used for lsp
func printDiagnosticJsonLine(d rule.RuleDiagnostic, w *bufio.Writer, comparePathOptions tspath.ComparePathsOptions) {
	diagnosticStart := d.Range.Pos()
	diagnosticEnd := d.Range.End()

	startLine, startColumn := scanner.GetLineAndCharacterOfPosition(d.SourceFile, diagnosticStart)
	endLine, endColumn := scanner.GetLineAndCharacterOfPosition(d.SourceFile, diagnosticEnd)

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
	}

	jsonBytes, err := json.Marshal(diagnostic)
	if err != nil {
		type ErrorObject struct {
			Error string `json:"error"`
		}
		errorObject := ErrorObject{Error: fmt.Sprintf("Failed to marshal diagnostic: %s", err)}
		errorBytes, _ := json.Marshal(errorObject) // Ignoring error since struct is simple
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

	diagnosticStartLine, diagnosticStartColumn := scanner.GetLineAndCharacterOfPosition(d.SourceFile, diagnosticStart)
	diagnosticEndline, _ := scanner.GetLineAndCharacterOfPosition(d.SourceFile, diagnosticEnd)

	lineMap := d.SourceFile.LineMap()
	text := d.SourceFile.Text()

	codeboxStartLine := max(diagnosticStartLine-1, 0)
	codeboxEndLine := min(diagnosticEndline+1, len(lineMap)-1)

	codeboxStart := scanner.GetPositionOfLineAndCharacter(d.SourceFile, codeboxStartLine, 0)
	var codeboxEndColumn int
	if codeboxEndLine == len(lineMap)-1 {
		codeboxEndColumn = len(text) - int(lineMap[len(lineMap)-1])
	} else {
		codeboxEndColumn = int(lineMap[codeboxEndLine+1]-lineMap[codeboxEndLine]) - 1
	}
	codeboxEnd := scanner.GetPositionOfLineAndCharacter(d.SourceFile, codeboxEndLine, codeboxEndColumn)

	// Rule name with conditional coloring
	w.WriteByte(' ')
	w.WriteString(colors.RuleName(" %s ", d.RuleName))
	w.WriteString(" — ")

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
	lastNonSpaceIndex := -1

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

	for i, char := range text[codeboxStart:codeboxEnd] {
		if char == '\n' {
			if line != codeboxEndLine {
				lineIndentCalculated = false
				lineEnds[line-codeboxStartLine] = lastNonSpaceIndex - int(lineMap[line]) + codeboxStart
				lastNonSpaceIndex = -1
				line++
			}
			continue
		}

		if !lineIndentCalculated && !unicode.IsSpace(char) {
			lineIndentCalculated = true
			lineStarts[line-codeboxStartLine] = i - int(lineMap[line]) + codeboxStart
			indentSize = min(indentSize, lineStarts[line-codeboxStartLine])
		}

		if lineIndentCalculated && !unicode.IsSpace(char) {
			lastNonSpaceIndex = i + 1
		}
	}
	if line == codeboxEndLine {
		lineEnds[line-codeboxStartLine] = lastNonSpaceIndex - int(lineMap[line]) + codeboxStart
	}

	diagnosticHighlightActive := false
	lastLineNumber := strconv.Itoa(codeboxEndLine + 1)
	for line := codeboxStartLine; line <= codeboxEndLine; line++ {
		w.WriteString("  ")
		w.WriteString(colors.BorderText("│ "))
		if line == codeboxEndLine {
			w.WriteString(lastLineNumber)
		} else {
			number := strconv.Itoa(line + 1)
			if len(number) < len(lastLineNumber) {
				w.WriteByte(' ')
			}
			w.WriteString(number)
		}
		w.WriteString(colors.BorderText(" │"))
		w.WriteString("  ")

		lineTextStart := int(lineMap[line]) + indentSize
		underlineStart := max(lineTextStart, int(lineMap[line])+lineStarts[line-codeboxStartLine])
		underlineEnd := underlineStart
		lineTextEnd := max(int(lineMap[line])+lineEnds[line-codeboxStartLine], lineTextStart)

		if diagnosticHighlightActive {
			underlineEnd = lineTextEnd
		} else if int(lineMap[line]) <= diagnosticStart && (line == len(lineMap) || diagnosticStart < int(lineMap[line+1])) {
			underlineStart = min(max(lineTextStart, diagnosticStart), lineTextEnd)
			underlineEnd = lineTextEnd
			diagnosticHighlightActive = true
		}
		if int(lineMap[line]) <= diagnosticEnd && (line == len(lineMap) || diagnosticEnd < int(lineMap[line+1])) {
			underlineEnd = min(max(underlineStart, diagnosticEnd), lineTextEnd)
			diagnosticHighlightActive = false
		}

		if underlineStart != underlineEnd {
			w.WriteString(text[lineTextStart:underlineStart])
			w.WriteString(colors.ErrorText("%s", text[underlineStart:underlineEnd]))
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

const usage = `✨ Rslint - Rocket Speed Linter

Usage:
    rslint [OPTIONS]

Options:
    --config PATH     Which rslint config file to use. Defaults to rslint.json.
	--list-files      List matched files
	--format FORMAT   Output format: default | jsonline
	--ipc             Run in IPC mode (for JS integration)
	--no-color        Disable colored output
	--force-color     Force colored output
	-h, --help        Show help
`

// read config and deserialize the jsonc result
func loadRslintConfig(configPath string, currentDirectory string, fs vfs.FS) (rslintconfig.RslintConfig, string) {
	configFileName := tspath.ResolvePath(currentDirectory, configPath)
	if !fs.FileExists(configFileName) {
		fmt.Fprintf(os.Stderr, "error: rslint config file %q doesn't exist\n", configFileName)
		os.Exit(1)
	}

	data, ok := fs.ReadFile(configFileName)
	if !ok {
		fmt.Fprintf(os.Stderr, "error reading rslint config file %q\n", configFileName)
		os.Exit(1)
	}

	var config rslintconfig.RslintConfig
	// Use JSONC parser to support comments and trailing commas
	if err := utils.ParseJSONC([]byte(data), &config); err != nil {
		fmt.Fprintf(os.Stderr, "error parsing rslint config file %q: %v\n", configFileName, err)
		os.Exit(1)
	}
	currentDirectory = tspath.GetDirectoryPath(configFileName)
	return config, currentDirectory
}
func loadTsConfigFromRslintConfig(rslintConfig rslintconfig.RslintConfig, currentDirectory string, fs vfs.FS) []string {
	tsConfig := []string{}
	for _, entry := range rslintConfig {

		for _, config := range entry.LanguageOptions.ParserOptions.Project {
			tsconfigPath := tspath.ResolvePath(currentDirectory, config)

			if !fs.FileExists(tsconfigPath) {
				fmt.Fprintf(os.Stderr, "error: tsconfig file %q doesn't exist\n", tsconfigPath)
				os.Exit(1)
			}
			tsConfig = append(tsConfig, tsconfigPath)

		}
	}
	return tsConfig
}
func runCMD() int {
	flag.Usage = func() { fmt.Fprint(os.Stderr, usage) }

	var (
		help      bool
		config    string
		listFiles bool

		traceOut       string
		cpuprofOut     string
		singleThreaded bool
		format         string
		ipcMode        bool
		noColor        bool
		forceColor     bool
	)
	flag.StringVar(&format, "format", "default", "output format")
	flag.StringVar(&config, "config", "", "which rslint config to use")
	flag.BoolVar(&listFiles, "list-files", false, "list matched files")
	flag.BoolVar(&help, "help", false, "show help")
	flag.BoolVar(&help, "h", false, "show help")
	flag.BoolVar(&ipcMode, "ipc", false, "run in IPC mode (for JS integration)")
	flag.BoolVar(&noColor, "no-color", false, "disable colored output")
	flag.BoolVar(&forceColor, "force-color", false, "force colored output")

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

	// Check if we need to run in IPC mode
	if ipcMode {
		return runAPI()
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

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))

	// Initialize rule registry with all available rules
	rslintconfig.RegisterAllTypeSriptEslintPluginRules()
	var rslintConfig rslintconfig.RslintConfig
	var cwd string
	var tsConfigs []string
	// Load rslint configuration and determine which rules to enable
	if config != "" {
		rslintConfig, cwd = loadRslintConfig(config, currentDirectory, fs)
		tsConfigs = loadTsConfigFromRslintConfig(rslintConfig, cwd, fs)
	} else {
		// Try to load default config files in order of preference
		defaultConfigs := []string{"rslint.json", "rslint.jsonc"}
		configLoaded := false
		for _, defaultConfig := range defaultConfigs {
			defaultConfigPath := tspath.ResolvePath(currentDirectory, defaultConfig)
			if fs.FileExists(defaultConfigPath) {
				rslintConfig, cwd = loadRslintConfig(defaultConfig, currentDirectory, fs)
				tsConfigs = loadTsConfigFromRslintConfig(rslintConfig, cwd, fs)
				configLoaded = true
				break
			}
		}
		if !configLoaded {
			fmt.Fprintf(os.Stderr, "error: no rslint config file found. Expected rslint.json or rslint.jsonc\n")
			return 1
		}
	}

	if len(tsConfigs) == 0 {
		fmt.Fprintf(os.Stderr, "error: no TypeScript configuration found in rslint config\n")
		return 1
	}

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

	files := []*ast.SourceFile{}
	for _, program := range programs {
		cwdPath := string(tspath.ToPath("", currentDirectory, program.Host().FS().UseCaseSensitiveFileNames()).EnsureTrailingDirectorySeparator())
		var matchedFiles strings.Builder
		for _, file := range program.SourceFiles() {
			p := string(file.Path())
			if strings.Contains(p, "/node_modules/") {
				continue
			}
			if fileName, matched := strings.CutPrefix(p, cwdPath); matched {
				if listFiles {
					matchedFiles.WriteString("Found file: ")
					matchedFiles.WriteString(fileName)
					matchedFiles.WriteByte('\n')
				}
				files = append(files, file)
			}
		}
		if listFiles {
			os.Stdout.WriteString(matchedFiles.String())
		}
		slices.SortFunc(files, func(a *ast.SourceFile, b *ast.SourceFile) int {
			return len(b.Text()) - len(a.Text())
		})
	}

	var wg sync.WaitGroup

	diagnosticsChan := make(chan rule.RuleDiagnostic, 4096)
	errorsCount := 0

	wg.Add(1)
	go func() {
		defer wg.Done()
		w := bufio.NewWriterSize(os.Stdout, 4096*100)
		defer w.Flush()
		for d := range diagnosticsChan {
			errorsCount++
			if errorsCount == 1 {
				w.WriteByte('\n')
			}
			printDiagnostic(d, w, comparePathOptions, format)
			if w.Available() < 4096 {
				w.Flush()
			}
		}
	}()

	err = linter.RunLinter(
		programs,
		singleThreaded,
		files,
		func(sourceFile *ast.SourceFile) []linter.ConfiguredRule {
			activeRules := rslintconfig.GlobalRuleRegistry.GetEnabledRules(rslintConfig, sourceFile.FileName())
			return utils.Map(activeRules, func(r rule.Rule) linter.ConfiguredRule {
				return linter.ConfiguredRule{
					Name: r.Name,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return r.Run(ctx, nil)
					},
				}
			})
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

	colors := setupColors()
	var errorsColorFunc func(string, ...interface{}) string
	if errorsCount == 0 {
		errorsColorFunc = colors.SuccessText
	} else {
		errorsColorFunc = colors.BoldText
	}

	errorsText := "errors"
	if errorsCount == 1 {
		errorsText = "error"
	}
	filesText := "files"
	if len(files) == 1 {
		filesText = "file"
	}
	threadsCount := 1
	if !singleThreaded {
		threadsCount = runtime.GOMAXPROCS(0)
	}
	if format == "default" {
		fmt.Fprintf(
			os.Stdout,
			"Found %s %s %s(linted %s %s with in %s using %s threads)%s\n",
			errorsColorFunc("%d", errorsCount),
			errorsText,
			colors.DimText(""),
			colors.BoldText("%d", len(files)),
			filesText,
			colors.BoldText("%v", time.Since(timeBefore).Round(time.Millisecond)),
			colors.BoldText("%d", threadsCount),
			color.New().SprintFunc()(""), // Reset
		)
	}

	// Exit with non-zero status code if errors were found
	if errorsCount > 0 {
		return 1
	}
	return 0
}
