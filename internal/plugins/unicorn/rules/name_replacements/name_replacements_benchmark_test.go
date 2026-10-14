package name_replacements_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/name_replacements"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func BenchmarkNameReplacementsIndependentBindings(b *testing.B) {
	for _, count := range []int{800, 1600, 3200, 6400} {
		b.Run(fmt.Sprintf("bindings-%d", count), func(b *testing.B) {
			var source strings.Builder
			for index := range count {
				fmt.Fprintf(&source, "function f%d() { let err = %d; use(err); }\n", index, index)
			}
			program, sourceFile, err := rule_tester.NewProgramHelper(fixtures.GetRootDir()).CreateTestProgram(source.String(), "benchmark.js", "tsconfig.json")
			if err != nil {
				b.Fatal(err)
			}
			sourceProgram := lintprogram.NewFromCompiler(program)
			b.ResetTimer()
			for range b.N {
				diagnostics := 0
				linter.LintSingleFile(linter.LintSingleFileOptions{
					Program: sourceProgram,
					File:    sourceFile.FileName(),
					GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
						return []rule.ConfiguredRule{{
							Name:     name_replacements.NameReplacementsRule.Name,
							Severity: rule.SeverityError,
							Run: func(ctx rule.RuleContext) rule.RuleListeners {
								return name_replacements.NameReplacementsRule.Run(ctx, nil)
							},
						}}
					},
					Consumer: rule.DiagnosticConsumer{
						Demand: rule.EditDemandNone,
						Report: func(rule.RuleDiagnostic) { diagnostics++ },
					},
				})
				if diagnostics != count {
					b.Fatalf("diagnostics = %d, want %d", diagnostics, count)
				}
			}
		})
	}
}
