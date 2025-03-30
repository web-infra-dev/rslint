package main

import (
	"bufio"
	"flag"
	"fmt"
	"math"
	"os"
	"runtime/pprof"
	"runtime/trace"
	"strconv"
	"sync"
	"time"
	"unicode"

	"none.none/tsgolint/internal/linter"
	"none.none/tsgolint/internal/rule"
	"none.none/tsgolint/internal/utils"

	"none.none/tsgolint/internal/rules/await_thenable"
	"none.none/tsgolint/internal/rules/no_array_delete"
	"none.none/tsgolint/internal/rules/no_base_to_string"
	"none.none/tsgolint/internal/rules/no_duplicate_type_constituents"
	"none.none/tsgolint/internal/rules/no_floating_promises"
	"none.none/tsgolint/internal/rules/no_for_in_array"
	"none.none/tsgolint/internal/rules/no_implied_eval"
	"none.none/tsgolint/internal/rules/no_misused_promises"
	"none.none/tsgolint/internal/rules/no_redundant_type_constituents"
	"none.none/tsgolint/internal/rules/no_unnecessary_type_assertion"
	"none.none/tsgolint/internal/rules/no_unsafe_argument"
	"none.none/tsgolint/internal/rules/no_unsafe_assignment"
	"none.none/tsgolint/internal/rules/no_unsafe_call"
	"none.none/tsgolint/internal/rules/no_unsafe_enum_comparison"
	"none.none/tsgolint/internal/rules/no_unsafe_member_access"
	"none.none/tsgolint/internal/rules/no_unsafe_return"
	"none.none/tsgolint/internal/rules/no_unsafe_unary_minus"
	"none.none/tsgolint/internal/rules/only_throw_error"
	"none.none/tsgolint/internal/rules/prefer_promise_reject_errors"
	"none.none/tsgolint/internal/rules/require_await"
	"none.none/tsgolint/internal/rules/restrict_plus_operands"
	"none.none/tsgolint/internal/rules/restrict_template_expressions"
	"none.none/tsgolint/internal/rules/switch_exhaustiveness_check"
	"none.none/tsgolint/internal/rules/unbound_method"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
)

var traceOut = flag.String("trace", "", "File to put trace to")
var cpuprofOut = flag.String("cpuprof", "", "File to put cpu profiling to")
var singleThreaded = flag.Bool("singleThreaded", false, "Run in single threaded mode.")

var lineEnds = make([]int, 13)

func printDiagnostic(d rule.RuleDiagnostic, w *bufio.Writer, comparePathOptions tspath.ComparePathsOptions) {
	diagnosticStart := d.Range.Pos()
	diagnosticEnd := d.Range.End()

	diagnosticStartLine, diagnosticStartColumn := scanner.GetLineAndCharacterOfPosition(d.SourceFile, diagnosticStart)
	diagnosticEndline, _ := scanner.GetLineAndCharacterOfPosition(d.SourceFile, diagnosticEnd)

	lineMap := d.SourceFile.LineMap()
	text := d.SourceFile.Text

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

	w.Write([]byte{' ', 0x1b, '[', '7', 'm', 0x1b, '[', '1', 'm', 0x1b, '[', '3', '8', ';', '5', ';', '3', '7', 'm', ' '})
	w.WriteString(d.RuleName)
	w.WriteString(" \x1b[0m — ")
	w.WriteString(d.Message.Description)
	w.WriteString("\n  \x1b[2m╭─┴──────────(\x1b[0m \x1b[38;5;45m")
	w.WriteString(tspath.ConvertToRelativePath(d.SourceFile.FileName(), comparePathOptions))
	w.WriteByte(':')
	w.Write([]byte(strconv.Itoa(diagnosticStartLine + 1)))
	w.WriteByte(':')
	w.Write([]byte(strconv.Itoa(diagnosticStartColumn + 1)))
	w.WriteString("\x1b[0m \x1b[2m)──\x1b[0m\n")

	indentSize := math.MaxInt
	line := codeboxStartLine
	lineIndentCalculated := false
	lastNonSpaceIndex := -1

	if codeboxEndLine-codeboxStartLine >= len(lineEnds) {
		w.WriteString("  \x1b[2m│\x1b[0m  Error range is too big. Skipping code block printing.\n  \x1b[2m╰────────────────────────────────\x1b[0m\n\n")
		return
	}

	var lineStarts = make([]int, 13)

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
		w.WriteString("  \x1b[2m│ ")
		if line == codeboxEndLine {
			w.WriteString(lastLineNumber)
		} else {
			number := strconv.Itoa(line + 1)
			if len(number) < len(lastLineNumber) {
				w.WriteByte(' ')
			}
			w.WriteString(number)
		}
		w.Write([]byte(" │\x1b[0m  "))

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
			w.Write([]byte{0x1b, '[', '3', '8', ';', '5', ';', '1', '9', '6', 'm'})
			w.Write([]byte{0x1b, '[', '4', 'm'})
			w.Write([]byte{0x1b, '[', '4', ':', '3', 'm'})
			w.Write([]byte{0x1b, '[', '5', '8', ':', '5', ':', '1', '9', '6', 'm'})
			w.WriteString(text[underlineStart:underlineEnd])
			w.Write([]byte{0x1b, '[', '0', 'm'})
			w.WriteString(text[underlineEnd:lineTextEnd])
		} else if lineTextStart != lineTextEnd {
			w.WriteString(text[lineTextStart:lineTextEnd])
		}

		w.WriteByte('\n')
	}
	w.WriteString("  \x1b[2m╰────────────────────────────────\x1b[0m\n\n")
}

func main() {
	enableVirtualTerminalProcessing()
	timeBefore := time.Now()
	flag.Parse()

	if *traceOut != "" {
		f, err := os.Create(*traceOut)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		trace.Start(f)
		defer trace.Stop()
	}
	if *cpuprofOut != "" {
		f, err := os.Create(*cpuprofOut + ".pg.gz")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		err = pprof.StartCPUProfile(f)
		if err != nil {
			panic(err)
		}
		defer pprof.StopCPUProfile()
	}

	currentDirectory, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
		os.Exit(1)
	}
	configFileName := tspath.ResolvePath(currentDirectory, "tsconfig.eslint.json")

	var files []string

	var rules = []rule.Rule{
		await_thenable.AwaitThenableRule,
		no_array_delete.NoArrayDeleteRule,
		no_base_to_string.NoBaseToStringRule,
		no_duplicate_type_constituents.NoDuplicateTypeConstituentsRule,
		no_floating_promises.NoFloatingPromisesRule,
		no_for_in_array.NoForInArrayRule,
		no_implied_eval.NoImpliedEvalRule,
		no_misused_promises.NoMisusedPromisesRule,
		no_redundant_type_constituents.NoRedundantTypeConstituentsRule,
		no_unnecessary_type_assertion.NoUnnecessaryTypeAssertionRule,
		no_unsafe_argument.NoUnsafeArgumentRule,
		no_unsafe_assignment.NoUnsafeAssignmentRule,
		no_unsafe_call.NoUnsafeCallRule,
		no_unsafe_enum_comparison.NoUnsafeEnumComparisonRule,
		no_unsafe_member_access.NoUnsafeMemberAccessRule,
		no_unsafe_return.NoUnsafeReturnRule,
		no_unsafe_unary_minus.NoUnsafeUnaryMinusRule,
		only_throw_error.OnlyThrowErrorRule,
		prefer_promise_reject_errors.PreferPromiseRejectErrorsRule,
		require_await.RequireAwaitRule,
		restrict_plus_operands.RestrictPlusOperandsRule,
		restrict_template_expressions.RestrictTemplateExpressionsRule,
		switch_exhaustiveness_check.SwitchExhaustivenessCheckRule,
		unbound_method.UnboundMethodRule,
	}

	diagnosticsChan := make(chan rule.RuleDiagnostic, 4096)
	errorsCount := 0

	fs := bundled.WrapFS(osvfs.FS())
	host := utils.CreateCompilerHost(currentDirectory, fs)

	comparePathOptions := tspath.ComparePathsOptions{
		CurrentDirectory:          host.GetCurrentDirectory(),
		UseCaseSensitiveFileNames: host.FS().UseCaseSensitiveFileNames(),
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		w := bufio.NewWriterSize(os.Stdout, 4096*100)
		defer w.Flush()
		for d := range diagnosticsChan {
			errorsCount++
			printDiagnostic(d, w, comparePathOptions)
		}
		fmt.Fprintf(w, "Total errors: %v\n", errorsCount)
	}()

	err = linter.RunLinter(
		*singleThreaded,
		fs,
		files,
		func(sourceFile *ast.SourceFile) []linter.ConfiguredRule {
			return utils.Map(rules, func(r rule.Rule) linter.ConfiguredRule {
				return linter.ConfiguredRule{
					Name: r.Name,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return r.Run(ctx, nil)
					},
				}
			})
		},
		currentDirectory,
		configFileName,
		func(d rule.RuleDiagnostic) {
			diagnosticsChan <- d
		},
		host,
	)

	close(diagnosticsChan)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running linter: %v\n", err)
		return
		// os.Exit(1)
	}

	wg.Wait()

	fmt.Printf("Execution time: %v\n", time.Since(timeBefore))
}
