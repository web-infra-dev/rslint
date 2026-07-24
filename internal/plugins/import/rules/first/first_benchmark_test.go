package first_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/first"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var firstBenchmarkDiagnosticCount int

func BenchmarkFirstEditDemand(b *testing.B) {
	const (
		initialStatements   = 64
		misplacedImports    = 24
		statementsPerImport = 4
		fileName            = "benchmark.ts"
	)

	var source strings.Builder
	for i := range initialStatements {
		writeFirstBenchmarkStatement(&source, "initial", i, 0)
	}
	for importIndex := range misplacedImports {
		for statementIndex := range statementsPerImport {
			writeFirstBenchmarkStatement(&source, "between", importIndex, statementIndex)
		}
		source.WriteString("import { imported_")
		source.WriteString(strconv.Itoa(importIndex))
		source.WriteString(" } from './module-")
		source.WriteString(strconv.Itoa(importIndex))
		source.WriteString("';\n")
	}

	program, sourceFile := createFirstProgram(b, fileName, source.String())

	for _, test := range []struct {
		name   string
		demand rule.EditDemand
	}{
		{name: "diagnostics_only", demand: rule.EditDemandNone},
		{name: "all_edits", demand: rule.EditDemandAll},
	} {
		b.Run(test.name, func(b *testing.B) {
			diagnosticCount := 0
			consumer := rule.DiagnosticConsumer{
				Demand: test.demand,
				Report: func(rule.RuleDiagnostic) {
					diagnosticCount++
				},
			}

			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				diagnosticCount = 0
				linter.LintSingleFile(linter.LintSingleFileOptions{
					Program:         program,
					File:            sourceFile.FileName(),
					HasTypeInfo:     true,
					GetRulesForFile: firstConfiguredRules,
					ExcludePaths:    []string{},
					Consumer:        consumer,
				})
			}
			b.StopTimer()

			if diagnosticCount != misplacedImports {
				b.Fatalf("got %d diagnostics, want %d", diagnosticCount, misplacedImports)
			}
			firstBenchmarkDiagnosticCount = diagnosticCount
		})
	}
}

func writeFirstBenchmarkStatement(source *strings.Builder, prefix string, firstIndex int, secondIndex int) {
	source.WriteString("const ")
	source.WriteString(prefix)
	source.WriteByte('_')
	source.WriteString(strconv.Itoa(firstIndex))
	source.WriteByte('_')
	source.WriteString(strconv.Itoa(secondIndex))
	source.WriteString(" = { value: ")
	source.WriteString(strconv.Itoa(firstIndex + secondIndex))
	source.WriteString(", nested: { enabled: true } };\n")
}

func createFirstProgram(t testing.TB, fileName string, code string) (*compiler.Program, *ast.SourceFile) {
	t.Helper()

	rootDir := fixtures.GetRootDir()
	fs := utils.NewOverlayVFSForFile(tspath.ResolvePath(rootDir, fileName), code)
	host := utils.CreateCompilerHost(rootDir, fs)
	program, err := utils.CreateProgram(true, fs, rootDir, "tsconfig.json", host)
	if err != nil {
		t.Fatalf("failed to create program: %v", err)
	}
	sourceFile := program.GetSourceFile(fileName)
	if sourceFile == nil {
		t.Fatalf("source file %q not found", fileName)
	}
	return program, sourceFile
}

func firstConfiguredRules(*ast.SourceFile) []linter.ConfiguredRule {
	return []linter.ConfiguredRule{{
		Name:     first.FirstRule.Name,
		Severity: rule.SeverityError,
		Run: func(ctx rule.RuleContext) rule.RuleListeners {
			return first.FirstRule.Run(ctx, nil)
		},
	}}
}
