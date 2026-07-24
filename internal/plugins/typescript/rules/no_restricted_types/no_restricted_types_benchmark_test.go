package no_restricted_types

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

var noRestrictedTypesBenchmarkDiagnosticCount int

func BenchmarkNoRestrictedTypesEditDemand(b *testing.B) {
	const (
		violationCount = 64
		fileName       = "benchmark.ts"
	)

	var source strings.Builder
	for index := range violationCount {
		source.WriteString("type Alias")
		source.WriteString(strconv.Itoa(index))
		source.WriteString(" = Banned;\n")
	}

	program, sourceFile := createNoRestrictedTypesProgram(b, fileName, source.String())
	options := rule.NormalizeOptions(map[string]interface{}{
		"types": map[string]interface{}{
			"Banned": map[string]interface{}{
				"fixWith": "Allowed",
				"message": "Use an allowed domain type.",
				"suggest": []interface{}{
					"AllowedOne",
					"AllowedTwo",
					"AllowedThree",
					"AllowedFour",
				},
			},
		},
	})
	getRules := noRestrictedTypesConfiguredRules(options)

	for _, test := range []struct {
		name   string
		demand rule.EditDemand
	}{
		{name: "diagnostics_only", demand: rule.EditDemandNone},
		{name: "autofix", demand: rule.EditDemandAutofix},
		{name: "suggestions", demand: rule.EditDemandSuggestion},
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

			if diagnosticCount != violationCount {
				b.Fatalf("got %d diagnostics, want %d", diagnosticCount, violationCount)
			}
			noRestrictedTypesBenchmarkDiagnosticCount = diagnosticCount
		})
	}
}

func createNoRestrictedTypesProgram(t testing.TB, fileName string, code string) (*compiler.Program, *ast.SourceFile) {
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

func noRestrictedTypesConfiguredRules(options []any) func(*ast.SourceFile) []linter.ConfiguredRule {
	return func(*ast.SourceFile) []linter.ConfiguredRule {
		return []linter.ConfiguredRule{{
			Name:     NoRestrictedTypesRule.Name,
			Severity: rule.SeverityError,
			Run: func(ctx rule.RuleContext) rule.RuleListeners {
				return NoRestrictedTypesRule.Run(ctx, options)
			},
		}}
	}
}
