package utils

import (
	"fmt"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/web-infra-dev/rslint/internal/rule"
)

var benchmarkRstestCallAnalysis *RstestCallAnalysis

func benchmarkAnalysisSourceFile() *ast.SourceFile {
	var source strings.Builder
	source.WriteString(`import { expect, test } from "@rstest/core";`)
	for index := range 200 {
		fmt.Fprintf(
			&source,
			"\nhelper%d(); test(%q, () => expect(%d).toBe(%d));",
			index,
			fmt.Sprintf("case %d", index),
			index,
			index,
		)
	}
	return parser.ParseSourceFile(
		ast.SourceFileParseOptions{
			FileName: "/analysis-benchmark.test.ts",
			Path:     "/analysis-benchmark.test.ts",
		},
		source.String(),
		core.ScriptKindTS,
	)
}

func BenchmarkRstestCallAnalysisFiveRules(b *testing.B) {
	sourceFile := benchmarkAnalysisSourceFile()
	b.ReportAllocs()

	b.Run("independent", func(b *testing.B) {
		for b.Loop() {
			ctx := rule.RuleContext{SourceFile: sourceFile}
			for range 5 {
				benchmarkRstestCallAnalysis = newRstestCallAnalysis(ctx)
			}
		}
	})
	b.Run("shared", func(b *testing.B) {
		for b.Loop() {
			ctx := rule.RuleContext{SourceFile: sourceFile}.
				WithFileCache(rule.NewFileCache())
			for range 5 {
				benchmarkRstestCallAnalysis = GetRstestCallAnalysis(ctx)
			}
		}
	})
}
