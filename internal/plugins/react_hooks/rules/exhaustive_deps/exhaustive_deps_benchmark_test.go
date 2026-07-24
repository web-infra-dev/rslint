package exhaustive_deps

import (
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var exhaustiveDepsBenchmarkDiagnosticCount int

func BenchmarkExhaustiveDepsEditDemand(b *testing.B) {
	const (
		hookCount     = 24
		propertyCount = 12
		fileName      = "benchmark.ts"
	)

	var source strings.Builder
	source.WriteString("function Component(props: {")
	for property := range propertyCount {
		source.WriteString(" value")
		source.WriteString(strconv.Itoa(property))
		source.WriteString(": number;")
	}
	source.WriteString(" }) {\n")
	for range hookCount {
		source.WriteString("  useEffect(() => { console.log(")
		for property := range propertyCount {
			if property > 0 {
				source.WriteString(", ")
			}
			source.WriteString("props.value")
			source.WriteString(strconv.Itoa(property))
		}
		source.WriteString("); }, []);\n")
	}
	source.WriteString("}\n")

	program, sourceFile := createExhaustiveDepsProgram(b, fileName, source.String())

	for _, config := range []struct {
		name    string
		options []any
	}{
		{name: "suggestions", options: rule.NormalizeOptions(nil)},
		{
			name: "dangerous_autofix",
			options: rule.NormalizeOptions(map[string]interface{}{
				"enableDangerousAutofixThisMayCauseInfiniteLoops": true,
			}),
		},
	} {
		b.Run(config.name, func(b *testing.B) {
			getRules := exhaustiveDepsConfiguredRules(config.options)
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

					if diagnosticCount != hookCount {
						b.Fatalf("got %d diagnostics, want %d", diagnosticCount, hookCount)
					}
					exhaustiveDepsBenchmarkDiagnosticCount = diagnosticCount
				})
			}
		})
	}
}

func createExhaustiveDepsProgram(t testing.TB, fileName string, code string) (*compiler.Program, *ast.SourceFile) {
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

func exhaustiveDepsConfiguredRules(options []any) func(*ast.SourceFile) []linter.ConfiguredRule {
	return func(*ast.SourceFile) []linter.ConfiguredRule {
		return []linter.ConfiguredRule{{
			Name:     ExhaustiveDepsRule.Name,
			Severity: rule.SeverityError,
			Run: func(ctx rule.RuleContext) rule.RuleListeners {
				return ExhaustiveDepsRule.Run(ctx, options)
			},
		}}
	}
}
