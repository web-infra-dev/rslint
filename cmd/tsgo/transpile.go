package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/microsoft/typescript-go/shim/tsoptions"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type transpileOptions struct {
	Config          string
	File            string
	Module          string
	Target          string
	Jsx             string
	InlineSourceMap bool
	SourceMap       bool
	TypeCheck       bool
}

func runTranspile(opts transpileOptions) int {
	if opts.File == "" {
		fmt.Fprintln(os.Stderr, "error: --file is required when using --transpile")
		return 2
	}
	absFile, err := filepath.Abs(opts.File)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error resolving file path: %v\n", err)
		return 1
	}
	absFile = tspath.NormalizePath(absFile)

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	currentDirectory, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error resolving cwd: %v\n", err)
		return 1
	}
	currentDirectory = tspath.NormalizePath(currentDirectory)
	host := utils.CreateCompilerHost(currentDirectory, fs)

	var baseParsed *tsoptions.ParsedCommandLine
	if opts.Config != "" {
		parsed, diags := tsoptions.GetParsedCommandLineOfConfigFile(opts.Config, &core.CompilerOptions{}, host, nil)
		if len(diags) > 0 {
			printTranspileDiagnostics(diags, currentDirectory)
			return 1
		}
		if parsed == nil {
			fmt.Fprintf(os.Stderr, "error parsing tsconfig: %s\n", opts.Config)
			return 1
		}
		baseParsed = parsed
	}

	var compilerOptions core.CompilerOptions
	if baseParsed != nil && baseParsed.CompilerOptions() != nil {
		compilerOptions = *baseParsed.CompilerOptions()
	}
	if err := applyTranspileOverrides(&compilerOptions, opts, absFile); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 2
	}

	comparePathsOptions := tspath.ComparePathsOptions{
		UseCaseSensitiveFileNames: fs.UseCaseSensitiveFileNames(),
		CurrentDirectory:          currentDirectory,
	}
	attachConfig := baseParsed != nil && shouldAttachConfig(baseParsed, absFile, comparePathsOptions)
	fileNames := []string{absFile}
	if attachConfig {
		fileNames = baseParsed.FileNames()
	}
	parsed := tsoptions.NewParsedCommandLine(&compilerOptions, fileNames, comparePathsOptions)
	if attachConfig {
		parsed.ConfigFile = baseParsed.ConfigFile
		parsed.Raw = baseParsed.Raw
		parsed.SetTypeAcquisition(baseParsed.TypeAcquisition())
	}

	program := compiler.NewProgram(compiler.ProgramOptions{
		Config:         parsed,
		SingleThreaded: core.TSTrue,
		Host:           host,
	})
	if program == nil {
		fmt.Fprintln(os.Stderr, "error creating program")
		return 1
	}
	sourceFile := program.GetSourceFile(absFile)
	if sourceFile == nil {
		fmt.Fprintf(os.Stderr, "error loading source file: %s\n", absFile)
		return 1
	}

	if opts.TypeCheck {
		diagnostics := compiler.GetDiagnosticsOfAnyProgram(
			context.Background(),
			program,
			nil,
			false,
			program.GetBindDiagnostics,
			program.GetSemanticDiagnostics,
		)
		if len(diagnostics) > 0 {
			printTranspileDiagnostics(diagnostics, currentDirectory)
			if hasErrorDiagnostics(diagnostics) {
				return 1
			}
		}
	}

	var output string
	emitResult := program.Emit(context.Background(), compiler.EmitOptions{
		TargetSourceFile: sourceFile,
		EmitOnly:         compiler.EmitOnlyJs,
		WriteFile: func(fileName string, text string, writeByteOrderMark bool, data *compiler.WriteFileData) error {
			if isJsOutput(fileName) {
				output = text
			}
			return nil
		},
	})
	if emitResult != nil && len(emitResult.Diagnostics) > 0 && !opts.TypeCheck {
		printTranspileDiagnostics(emitResult.Diagnostics, currentDirectory)
	}
	if emitResult != nil && emitResult.EmitSkipped {
		fmt.Fprintln(os.Stderr, "emit skipped")
		return 1
	}
	if output == "" {
		fmt.Fprintln(os.Stderr, "no output produced")
		return 1
	}
	fmt.Fprint(os.Stdout, output)
	return 0
}

func applyTranspileOverrides(options *core.CompilerOptions, opts transpileOptions, filePath string) error {
	options.NoEmit = core.TSFalse
	options.NoEmitOnError = core.TSFalse
	options.Composite = core.TSFalse
	if opts.TypeCheck {
		options.NoCheck = core.TSFalse
	}
	options.EmitDeclarationOnly = core.TSFalse
	if !opts.TypeCheck {
		options.AllowNonTsExtensions = core.TSTrue
		options.AllowImportingTsExtensions = core.TSTrue
	}

	if opts.Module != "" {
		moduleKind, err := parseModuleKind(opts.Module)
		if err != nil {
			return err
		}
		options.Module = moduleKind
	} else if options.Module == core.ModuleKindNone {
		options.Module = core.ModuleKindCommonJS
	}

	if opts.Target != "" {
		target, err := parseScriptTarget(opts.Target)
		if err != nil {
			return err
		}
		options.Target = target
	} else if options.Target == core.ScriptTargetNone {
		options.Target = core.ScriptTargetES2020
	}

	if opts.Jsx != "" {
		jsx, err := parseJsxEmit(opts.Jsx)
		if err != nil {
			return err
		}
		options.Jsx = jsx
	} else if options.Jsx == core.JsxEmitNone && strings.HasSuffix(strings.ToLower(filePath), ".tsx") {
		options.Jsx = core.JsxEmitReactJSX
	}

	if opts.InlineSourceMap {
		options.InlineSourceMap = core.TSTrue
		options.InlineSources = core.TSTrue
	}
	if opts.SourceMap {
		options.SourceMap = core.TSTrue
	}
	return nil
}

func parseModuleKind(value string) (core.ModuleKind, error) {
	switch strings.ToLower(value) {
	case "commonjs", "cjs":
		return core.ModuleKindCommonJS, nil
	case "es2015", "es6":
		return core.ModuleKindES2015, nil
	case "es2020":
		return core.ModuleKindES2020, nil
	case "es2022":
		return core.ModuleKindES2022, nil
	case "esnext", "esm":
		return core.ModuleKindESNext, nil
	case "node16":
		return core.ModuleKindNode16, nil
	case "node18":
		return core.ModuleKindNode18, nil
	case "node20":
		return core.ModuleKindNode20, nil
	case "nodenext":
		return core.ModuleKindNodeNext, nil
	case "preserve":
		return core.ModuleKindPreserve, nil
	default:
		return core.ModuleKindNone, fmt.Errorf("unknown module kind: %s", value)
	}
}

func parseScriptTarget(value string) (core.ScriptTarget, error) {
	switch strings.ToLower(value) {
	case "es3":
		return core.ScriptTargetES3, nil
	case "es5":
		return core.ScriptTargetES5, nil
	case "es2015", "es6":
		return core.ScriptTargetES2015, nil
	case "es2016":
		return core.ScriptTargetES2016, nil
	case "es2017":
		return core.ScriptTargetES2017, nil
	case "es2018":
		return core.ScriptTargetES2018, nil
	case "es2019":
		return core.ScriptTargetES2019, nil
	case "es2020":
		return core.ScriptTargetES2020, nil
	case "es2021":
		return core.ScriptTargetES2021, nil
	case "es2022":
		return core.ScriptTargetES2022, nil
	case "es2023":
		return core.ScriptTargetES2023, nil
	case "es2024":
		return core.ScriptTargetES2024, nil
	case "esnext":
		return core.ScriptTargetESNext, nil
	case "json":
		return core.ScriptTargetJSON, nil
	default:
		return core.ScriptTargetNone, fmt.Errorf("unknown script target: %s", value)
	}
}

func parseJsxEmit(value string) (core.JsxEmit, error) {
	switch strings.ToLower(value) {
	case "none":
		return core.JsxEmitNone, nil
	case "preserve":
		return core.JsxEmitPreserve, nil
	case "react-native":
		return core.JsxEmitReactNative, nil
	case "react":
		return core.JsxEmitReact, nil
	case "react-jsx":
		return core.JsxEmitReactJSX, nil
	case "react-jsxdev":
		return core.JsxEmitReactJSXDev, nil
	default:
		return core.JsxEmitNone, fmt.Errorf("unknown jsx emit: %s", value)
	}
}

func isJsOutput(fileName string) bool {
	lower := strings.ToLower(fileName)
	return strings.HasSuffix(lower, ".js") || strings.HasSuffix(lower, ".mjs") || strings.HasSuffix(lower, ".cjs")
}

func printTranspileDiagnostics(diagnostics []*ast.Diagnostic, baseDir string) {
	for _, diag := range diagnostics {
		printTranspileDiagnostic(diag, baseDir)
	}
}

func formatTranspileDiagnostics(diagnostics []*ast.Diagnostic, baseDir string) string {
	if len(diagnostics) == 0 {
		return ""
	}
	var builder strings.Builder
	builder.WriteString("TSError: ⨯ Unable to compile TypeScript:\n")
	for _, diag := range diagnostics {
		formatTranspileDiagnostic(&builder, diag, baseDir)
	}
	return builder.String()
}

func hasErrorDiagnostics(diagnostics []*ast.Diagnostic) bool {
	for _, diag := range diagnostics {
		if diag.Category() == 1 {
			return true
		}
	}
	return false
}

func printTranspileDiagnostic(diag *ast.Diagnostic, baseDir string) {
	if diag == nil {
		return
	}
	severity := diag.Category().Name()
	if diag.Code() != 0 {
		fmt.Fprintf(os.Stderr, "%s TS%d: %s\n", severity, diag.Code(), diag.Message())
	} else {
		fmt.Fprintf(os.Stderr, "%s: %s\n", severity, diag.Message())
	}

	file := diag.File()
	if file == nil {
		fmt.Fprintln(os.Stderr)
		return
	}
	fileName := formatDiagnosticFileName(file, baseDir)

	pos := diag.Pos()
	end := diag.End()
	if pos < 0 || end < 0 {
		fmt.Fprintf(os.Stderr, "  at %s\n\n", fileName)
		return
	}

	startLine, startColumn := scanner.GetECMALineAndCharacterOfPosition(file, pos)
	endLine, endColumn := scanner.GetECMALineAndCharacterOfPosition(file, end)
	fmt.Fprintf(os.Stderr, "  at %s:%d:%d\n", fileName, startLine+1, startColumn+1)

	lineStarts := scanner.GetECMALineStarts(file)
	text := file.Text()
	if len(lineStarts) == 0 || text == "" {
		fmt.Fprintln(os.Stderr)
		return
	}

	contextStart := maxInt(0, startLine-1)
	contextEnd := minInt(len(lineStarts)-1, endLine+1)
	lineNumberWidth := len(strconv.Itoa(contextEnd + 1))

	for line := contextStart; line <= contextEnd; line++ {
		lineStart := int(lineStarts[line])
		lineEnd := len(text)
		if line+1 < len(lineStarts) {
			lineEnd = int(lineStarts[line+1])
		}
		lineText := strings.TrimRight(text[lineStart:lineEnd], "\r\n")
		fmt.Fprintf(os.Stderr, "  %*d | %s\n", lineNumberWidth, line+1, lineText)

		if line < startLine || line > endLine {
			continue
		}

		underlineStart := 0
		if line == startLine {
			underlineStart = startColumn
		}
		underlineEnd := len(lineText)
		if line == endLine {
			underlineEnd = endColumn
		}
		underlineStart = clampInt(underlineStart, 0, len(lineText))
		underlineEnd = clampInt(underlineEnd, underlineStart, len(lineText))
		underlineLen := underlineEnd - underlineStart
		if underlineLen == 0 {
			underlineLen = 1
		}
		fmt.Fprintf(
			os.Stderr,
			"  %*s | %s%s\n",
			lineNumberWidth,
			"",
			strings.Repeat(" ", underlineStart),
			strings.Repeat("^", underlineLen),
		)
	}
	fmt.Fprintln(os.Stderr)
}

func formatTranspileDiagnostic(builder *strings.Builder, diag *ast.Diagnostic, baseDir string) {
	if diag == nil {
		return
	}
	severity := strings.ToLower(diag.Category().Name())
	if severity == "" {
		severity = "error"
	}

	file := diag.File()
	if file == nil {
		if diag.Code() != 0 {
			fmt.Fprintf(builder, "%s TS%d: %s\n\n", severity, diag.Code(), diag.Message())
		} else {
			fmt.Fprintf(builder, "%s: %s\n\n", severity, diag.Message())
		}
		return
	}
	fileName := formatDiagnosticFileName(file, baseDir)

	pos := diag.Pos()
	end := diag.End()
	if pos < 0 || end < 0 {
		if diag.Code() != 0 {
			fmt.Fprintf(builder, "%s - %s TS%d: %s\n\n", fileName, severity, diag.Code(), diag.Message())
		} else {
			fmt.Fprintf(builder, "%s - %s: %s\n\n", fileName, severity, diag.Message())
		}
		return
	}

	startLine, startColumn := scanner.GetECMALineAndCharacterOfPosition(file, pos)
	endLine, endColumn := scanner.GetECMALineAndCharacterOfPosition(file, end)
	if diag.Code() != 0 {
		fmt.Fprintf(
			builder,
			"%s:%d:%d - %s TS%d: %s\n",
			fileName,
			startLine+1,
			startColumn+1,
			severity,
			diag.Code(),
			diag.Message(),
		)
	} else {
		fmt.Fprintf(
			builder,
			"%s:%d:%d - %s: %s\n",
			fileName,
			startLine+1,
			startColumn+1,
			severity,
			diag.Message(),
		)
	}

	lineStarts := scanner.GetECMALineStarts(file)
	text := file.Text()
	if len(lineStarts) == 0 || text == "" {
		builder.WriteString("\n")
		return
	}

	contextStart := maxInt(0, startLine-1)
	contextEnd := minInt(len(lineStarts)-1, endLine+1)
	lineNumberWidth := len(strconv.Itoa(contextEnd + 1))

	for line := contextStart; line <= contextEnd; line++ {
		lineStart := int(lineStarts[line])
		lineEnd := len(text)
		if line+1 < len(lineStarts) {
			lineEnd = int(lineStarts[line+1])
		}
		lineText := strings.TrimRight(text[lineStart:lineEnd], "\r\n")
		fmt.Fprintf(builder, "  %*d | %s\n", lineNumberWidth, line+1, lineText)

		if line < startLine || line > endLine {
			continue
		}

		underlineStart := 0
		if line == startLine {
			underlineStart = startColumn
		}
		underlineEnd := len(lineText)
		if line == endLine {
			underlineEnd = endColumn
		}
		underlineStart = clampInt(underlineStart, 0, len(lineText))
		underlineEnd = clampInt(underlineEnd, underlineStart, len(lineText))
		underlineLen := underlineEnd - underlineStart
		if underlineLen == 0 {
			underlineLen = 1
		}
		fmt.Fprintf(
			builder,
			"  %*s | %s%s\n",
			lineNumberWidth,
			"",
			strings.Repeat(" ", underlineStart),
			strings.Repeat("~", underlineLen),
		)
	}
	builder.WriteString("\n")
}

func shouldAttachConfig(parsed *tsoptions.ParsedCommandLine, filePath string, options tspath.ComparePathsOptions) bool {
	if parsed == nil || parsed.ConfigFile == nil {
		return false
	}
	files := parsed.FileNames()
	if len(files) == 0 {
		return false
	}
	for _, candidate := range files {
		if tspath.ComparePaths(candidate, filePath, options) == 0 {
			return true
		}
	}
	return false
}

func clampInt(value int, minValue int, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func formatDiagnosticFileName(file *ast.SourceFile, baseDir string) string {
	if file == nil {
		return ""
	}
	name := file.FileName()
	if baseDir == "" {
		return name
	}
	rel, err := filepath.Rel(filepath.FromSlash(baseDir), filepath.FromSlash(name))
	if err != nil {
		return name
	}
	return filepath.ToSlash(rel)
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
