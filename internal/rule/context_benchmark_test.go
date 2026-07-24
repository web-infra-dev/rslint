package rule

import (
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
)

var combinedEditBenchmarkDiagnosticCount int

func BenchmarkRuleContextCombinedEditDemand(b *testing.B) {
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/benchmark.ts",
		Path:     "/benchmark.ts",
	}, "const value = 1;", core.ScriptKindTS)
	textRange := core.NewTextRange(0, 5)
	message := RuleMessage{Description: "benchmark diagnostic"}

	for _, test := range []struct {
		name   string
		demand EditDemand
	}{
		{name: "diagnostics_only", demand: EditDemandNone},
		{name: "autofix", demand: EditDemandAutofix},
		{name: "suggestions", demand: EditDemandSuggestion},
		{name: "all_edits", demand: EditDemandAll},
	} {
		b.Run("eager/"+test.name, func(b *testing.B) {
			diagnosticCount := 0
			ctx := RuleContext{
				SourceFile:     sourceFile,
				DisableManager: NewDisableManager(sourceFile, NewCommentStore(sourceFile)),
			}.WithDiagnosticConsumer("benchmark", SeverityWarning, DiagnosticConsumer{
				Demand: test.demand,
				Report: func(RuleDiagnostic) {
					diagnosticCount++
				},
			})

			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				fixes := buildCombinedBenchmarkFixes(textRange)
				suggestions := buildCombinedBenchmarkSuggestions(textRange)
				ctx.ReportRangeWithFixesAndSuggestions(textRange, message, fixes, suggestions)
			}
			b.StopTimer()

			if diagnosticCount != b.N {
				b.Fatalf("got %d diagnostics, want %d", diagnosticCount, b.N)
			}
			combinedEditBenchmarkDiagnosticCount = diagnosticCount
		})

		b.Run("deferred/"+test.name, func(b *testing.B) {
			diagnosticCount := 0
			ctx := RuleContext{
				SourceFile:     sourceFile,
				DisableManager: NewDisableManager(sourceFile, NewCommentStore(sourceFile)),
			}.WithDiagnosticConsumer("benchmark", SeverityWarning, DiagnosticConsumer{
				Demand: test.demand,
				Report: func(RuleDiagnostic) {
					diagnosticCount++
				},
			})
			buildFixes := func() []RuleFix {
				return buildCombinedBenchmarkFixes(textRange)
			}
			buildSuggestions := func() []RuleSuggestion {
				return buildCombinedBenchmarkSuggestions(textRange)
			}

			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				ctx.ReportRangeWithDeferredFixesAndSuggestions(
					textRange,
					message,
					buildFixes,
					buildSuggestions,
				)
			}
			b.StopTimer()

			if diagnosticCount != b.N {
				b.Fatalf("got %d diagnostics, want %d", diagnosticCount, b.N)
			}
			combinedEditBenchmarkDiagnosticCount = diagnosticCount
		})
	}
}

func buildCombinedBenchmarkFixes(textRange core.TextRange) []RuleFix {
	fixes := make([]RuleFix, 32)
	for i := range fixes {
		fixes[i] = RuleFix{
			Range: textRange,
			Text:  strings.Repeat("fix-"+strconv.Itoa(i), 2),
		}
	}
	return fixes
}

func buildCombinedBenchmarkSuggestions(textRange core.TextRange) []RuleSuggestion {
	suggestions := make([]RuleSuggestion, 16)
	for i := range suggestions {
		suggestions[i] = RuleSuggestion{
			Message: RuleMessage{
				Description: "suggestion " + strconv.Itoa(i),
			},
			FixesArr: []RuleFix{{
				Range: textRange,
				Text:  strings.Repeat("suggestion-fix-"+strconv.Itoa(i), 2),
			}},
		}
	}
	return suggestions
}
