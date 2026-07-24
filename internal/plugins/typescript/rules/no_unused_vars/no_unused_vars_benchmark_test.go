package no_unused_vars

import (
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var noUnusedVarsBenchmarkDiagnosticCount int

func BenchmarkNoUnusedVarsImportEditDemand(b *testing.B) {
	const (
		importDeclarations  = 32
		specifiersPerImport = 8
		fileName            = "benchmark.ts"
	)

	var source strings.Builder
	for declaration := range importDeclarations {
		source.WriteString("import { ")
		for specifier := range specifiersPerImport {
			if specifier > 0 {
				source.WriteString(", ")
			}
			source.WriteString("unused_")
			source.WriteString(strconv.Itoa(declaration))
			source.WriteByte('_')
			source.WriteString(strconv.Itoa(specifier))
		}
		source.WriteString(" } from './module-")
		source.WriteString(strconv.Itoa(declaration))
		source.WriteString("';\n")
	}

	program, sourceFile := createNoUnusedVarsProgram(b, fileName, source.String())

	for _, config := range []struct {
		name    string
		options []any
	}{
		{name: "suggestions", options: rule.NormalizeOptions(nil)},
		{
			name: "autofix",
			options: rule.NormalizeOptions(map[string]interface{}{
				"enableAutofixRemoval": map[string]interface{}{"imports": true},
			}),
		},
	} {
		b.Run(config.name, func(b *testing.B) {
			getRules := noUnusedVarsConfiguredRules(config.options)
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
							GetRulesForFile: getRules,
							ExcludePaths:    []string{},
							Consumer:        consumer,
						})
					}
					b.StopTimer()

					const wantDiagnostics = importDeclarations * specifiersPerImport
					if diagnosticCount != wantDiagnostics {
						b.Fatalf("got %d diagnostics, want %d", diagnosticCount, wantDiagnostics)
					}
					noUnusedVarsBenchmarkDiagnosticCount = diagnosticCount
				})
			}
		})
	}
}

func createNoUnusedVarsProgram(t testing.TB, fileName string, code string) (*compiler.Program, *ast.SourceFile) {
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

func noUnusedVarsConfiguredRules(options []any) func(*ast.SourceFile) []linter.ConfiguredRule {
	return func(*ast.SourceFile) []linter.ConfiguredRule {
		return []linter.ConfiguredRule{{
			Name:             NoUnusedVarsRule.Name,
			Severity:         rule.SeverityError,
			RequiresTypeInfo: true,
			Run: func(ctx rule.RuleContext) rule.RuleListeners {
				return NoUnusedVarsRule.Run(ctx, options)
			},
		}}
	}
}
