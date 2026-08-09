package linter

import (
	"fmt"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

var (
	benchmarkReferenceCountSink   int
	benchmarkDeclarationCountSink int
	benchmarkImportCountSink      int
)

type refBenchmarkQuery uint8

const (
	refBenchmarkReferences refBenchmarkQuery = iota
	refBenchmarkDeclarations
	refBenchmarkImports
	refBenchmarkAll
)

func benchmarkRefRule(query refBenchmarkQuery, streaming bool, fallbackAt int) ConfiguredRule {
	var needs rule.RefNeeds
	switch query {
	case refBenchmarkReferences:
		needs = rule.RefNeedReferences
	case refBenchmarkDeclarations:
		needs = rule.RefNeedBindingDeclarations
	case refBenchmarkImports:
		needs = rule.RefNeedImportBindings
	case refBenchmarkAll:
		needs = rule.RefNeedsAll
	default:
		panic("unknown RefStore benchmark query")
	}

	configured := ConfiguredRule{
		Name:     "benchmark/ref-store",
		Severity: rule.SeverityError,
		Run: func(ctx rule.RuleContext) rule.RuleListeners {
			if streaming {
				ctx.RequestRefs(needs)
			}
			tracked := ctx.SourceFile.Locals["tracked"]
			queryStore := func() {
				switch query {
				case refBenchmarkReferences:
					benchmarkReferenceCountSink = len(ctx.Refs.References(tracked))
				case refBenchmarkDeclarations:
					benchmarkDeclarationCountSink = len(ctx.Refs.BindingDeclarations())
				case refBenchmarkImports:
					benchmarkImportCountSink = len(ctx.Refs.ImportBindings())
				case refBenchmarkAll:
					benchmarkReferenceCountSink = len(ctx.Refs.References(tracked))
					benchmarkDeclarationCountSink = len(ctx.Refs.BindingDeclarations())
					benchmarkImportCountSink = len(ctx.Refs.ImportBindings())
				}
			}

			listeners := rule.RuleListeners{
				rule.ListenerOnFileFinalize(): func(*ast.Node) { queryStore() },
			}
			if fallbackAt > 0 {
				seen := 0
				listeners[ast.KindIdentifier] = func(*ast.Node) {
					seen++
					if seen == fallbackAt {
						queryStore()
					}
				}
			}
			return listeners
		},
	}
	if streaming {
		configured.Needs = rule.RuleNeeds{Refs: needs}
	}
	return configured
}

func runRefStoreBenchmark(
	b *testing.B,
	programOptions runProgramOptions,
	rules []ConfiguredRule,
) {
	programOptions.GetRulesForFile = func(*ast.SourceFile) []ConfiguredRule { return rules }
	plan := prepareProgramLintPlan(programOptions)
	programOptions.PreparedPlan = &plan
	consumer := rule.DiagnosticConsumer{Report: func(rule.RuleDiagnostic) {}}
	benchmarkProgramResultSink = runLintRulesInProgram(programOptions, consumer)

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		benchmarkProgramResultSink = runLintRulesInProgram(programOptions, consumer)
	}
}

func BenchmarkLinterRefStoreReferenceCollection(b *testing.B) {
	source := `export {};
let tracked = 0;
function work(input: number) {
` + strings.Repeat("  tracked += input;\n  consume(tracked, input);\n", 1000) + "}\n"
	program, targets := createLinterBenchmarkProgram(b, map[string]string{"large.ts": source})
	file := program.GetSourceFile(targets[0])
	if file == nil || file.Locals["tracked"] == nil {
		b.Fatal("benchmark source was not bound")
	}
	identifierCount := 0
	var countIdentifiers func(*ast.Node) bool
	countIdentifiers = func(node *ast.Node) bool {
		if node.Kind == ast.KindIdentifier {
			identifierCount++
		}
		node.ForEachChild(countIdentifiers)
		return false
	}
	file.AsNode().ForEachChild(countIdentifiers)
	base := runProgramOptions{
		Program:        program,
		TargetFiles:    targets,
		HasTargetFiles: true,
		ExcludePaths:   []string{},
		SingleThreaded: true,
	}

	for _, testCase := range []struct {
		name       string
		streaming  bool
		fallbackAt int
	}{
		{name: "finalize/lazy-prepass"},
		{name: "finalize/streaming", streaming: true},
		{name: "fallback-10pct/lazy-prepass", fallbackAt: identifierCount / 10},
		{name: "fallback-10pct/streaming", streaming: true, fallbackAt: identifierCount / 10},
		{name: "fallback-50pct/lazy-prepass", fallbackAt: identifierCount / 2},
		{name: "fallback-50pct/streaming", streaming: true, fallbackAt: identifierCount / 2},
		{name: "fallback-90pct/lazy-prepass", fallbackAt: identifierCount * 9 / 10},
		{name: "fallback-90pct/streaming", streaming: true, fallbackAt: identifierCount * 9 / 10},
	} {
		b.Run(testCase.name, func(b *testing.B) {
			rules := []ConfiguredRule{benchmarkRefRule(refBenchmarkReferences, testCase.streaming, testCase.fallbackAt)}
			runRefStoreBenchmark(b, base, rules)
		})
	}
}

func BenchmarkLinterRefStoreFactCollections(b *testing.B) {
	var source strings.Builder
	source.WriteString("export {};\nlet tracked = 0;\n")
	for index := range 300 {
		fmt.Fprintf(&source, "import { value as imported%d } from './missing%d';\n", index, index)
	}
	for index := range 1000 {
		fmt.Fprintf(&source, "const local%d = imported%d; tracked += local%d;\n", index, index%300, index)
	}
	program, targets := createLinterBenchmarkProgram(b, map[string]string{"facts.ts": source.String()})
	base := runProgramOptions{
		Program:        program,
		TargetFiles:    targets,
		HasTargetFiles: true,
		ExcludePaths:   []string{},
		SingleThreaded: true,
	}

	for _, query := range []struct {
		name  string
		query refBenchmarkQuery
	}{
		{name: "declarations", query: refBenchmarkDeclarations},
		{name: "imports", query: refBenchmarkImports},
		{name: "all", query: refBenchmarkAll},
	} {
		b.Run(query.name+"/lazy-prepass", func(b *testing.B) {
			runRefStoreBenchmark(b, base, []ConfiguredRule{benchmarkRefRule(query.query, false, 0)})
		})
		b.Run(query.name+"/streaming", func(b *testing.B) {
			runRefStoreBenchmark(b, base, []ConfiguredRule{benchmarkRefRule(query.query, true, 0)})
		})
	}
}
