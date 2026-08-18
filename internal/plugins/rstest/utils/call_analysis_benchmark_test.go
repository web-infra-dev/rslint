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

func benchmarkRstestSourceFile(b *testing.B, concurrent bool) (*ast.SourceFile, []*ast.Node) {
	b.Helper()
	var source strings.Builder
	for index := range 100 {
		if concurrent {
			fmt.Fprintf(
				&source,
				"test.concurrent(%q, () => expect(value).toMatchSnapshot());\n",
				fmt.Sprintf("case %d", index),
			)
		} else {
			fmt.Fprintf(
				&source,
				"test(%q, () => expect(value).toBe(%d));\n",
				fmt.Sprintf("case %d", index),
				index,
			)
		}
	}
	sourceFile := parser.ParseSourceFile(
		ast.SourceFileParseOptions{
			FileName: "/analysis-benchmark.test.ts",
			Path:     "/analysis-benchmark.test.ts",
		},
		source.String(),
		core.ScriptKindTS,
	)
	calls := make([]*ast.Node, 0, 300)
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindCallExpression {
			calls = append(calls, node)
		}
		return node.ForEachChild(visit)
	}
	sourceFile.Node.ForEachChild(visit)
	return sourceFile, calls
}

func BenchmarkRstestCallAnalysisParseExpectCalls(b *testing.B) {
	sourceFile, calls := benchmarkRstestSourceFile(b, false)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		analysis := newRstestCallAnalysis(rule.RuleContext{SourceFile: sourceFile})
		for _, call := range calls {
			analysis.ParseExpectCall(call)
		}
	}
}

func BenchmarkRstestCallAnalysisParseRegistrations(b *testing.B) {
	sourceFile, calls := benchmarkRstestSourceFile(b, false)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		analysis := newRstestCallAnalysis(rule.RuleContext{SourceFile: sourceFile})
		for _, call := range calls {
			analysis.ParseFnCall(call)
		}
	}
}

func BenchmarkRstestCallAnalysisConcurrentOwnership(b *testing.B) {
	sourceFile, calls := benchmarkRstestSourceFile(b, true)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		analysis := newRstestCallAnalysis(rule.RuleContext{SourceFile: sourceFile})
		context := &RstestConcurrentContext{
			ownership: collectRstestCallbackOwnership(analysis),
			modes:     map[*ast.Node]rstestConcurrentModeState{},
		}
		for _, call := range calls {
			context.IsInConcurrentTest(call)
		}
	}
}
