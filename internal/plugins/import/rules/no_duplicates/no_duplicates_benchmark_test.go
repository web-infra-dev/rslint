package no_duplicates_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
)

var noDuplicatesBenchmarkDiagnosticCount int

func BenchmarkNoDuplicatesEditDemand(b *testing.B) {
	const (
		groupCount      = 32
		importsPerGroup = 8
		fileName        = "benchmark.ts"
	)

	var source strings.Builder
	for group := range groupCount {
		moduleName := "./module-" + strconv.Itoa(group)
		for duplicate := range importsPerGroup {
			source.WriteString("import { value_")
			source.WriteString(strconv.Itoa(group))
			source.WriteByte('_')
			source.WriteString(strconv.Itoa(duplicate))
			source.WriteString(" } from '")
			source.WriteString(moduleName)
			source.WriteString("';\n")
		}
	}

	program, sourceFile := createNoDuplicatesProgram(b, fileName, source.String())

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
					GetRulesForFile: noDuplicatesConfiguredRules,
					ExcludePaths:    []string{},
					Consumer:        consumer,
				})
			}
			b.StopTimer()

			const wantDiagnostics = groupCount * importsPerGroup
			if diagnosticCount != wantDiagnostics {
				b.Fatalf("got %d diagnostics, want %d", diagnosticCount, wantDiagnostics)
			}
			noDuplicatesBenchmarkDiagnosticCount = diagnosticCount
		})
	}
}
