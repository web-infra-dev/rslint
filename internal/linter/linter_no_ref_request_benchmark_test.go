package linter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var benchmarkProgramResultSink programLintResult

// createLinterBenchmarkProgram performs all filesystem, parse, bind, and plan
// setup before benchmark timers start. It deliberately returns exact target
// paths so bundled libraries never enter the measured lint pass.
func createLinterBenchmarkProgram(
	b *testing.B,
	sources map[string]string,
) (*compiler.Program, []string) {
	b.Helper()
	dir := b.TempDir()
	includes := make([]string, 0, len(sources))
	targets := make([]string, 0, len(sources))
	for name, source := range sources {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			b.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
			b.Fatal(err)
		}
		includes = append(includes, "./"+name)
		targets = append(targets, tspath.NormalizePath(path))
	}
	tsconfig := `{"include":["` + strings.Join(includes, `","`) + `"]}`
	if err := os.WriteFile(filepath.Join(dir, "tsconfig.json"), []byte(tsconfig), 0o644); err != nil {
		b.Fatal(err)
	}

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	program, err := utils.CreateProgram(true, fs, dir, "tsconfig.json", utils.CreateCompilerHost(dir, fs))
	if err != nil {
		b.Fatal(err)
	}
	return program, targets
}

func benchmarkProgramLintOptions(
	program *compiler.Program,
	targets []string,
	rules []ConfiguredRule,
) runProgramOptions {
	opts := runProgramOptions{
		Program:         program,
		TargetFiles:     targets,
		HasTargetFiles:  true,
		ExcludePaths:    []string{},
		SingleThreaded:  true,
		GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule { return rules },
	}
	plan := prepareProgramLintPlan(opts)
	opts.PreparedPlan = &plan
	return opts
}

// BenchmarkLinterNoRefRequestsManySmallFiles is intentionally compatible with
// the pre-RefCollector main branch. It is the cross-revision zero-consumer
// guard: the measured rules neither declare nor request RefStore collection.
func BenchmarkLinterNoRefRequestsManySmallFiles(b *testing.B) {
	sources := make(map[string]string, 128)
	for index := range 128 {
		sources[fmt.Sprintf("file_%03d.ts", index)] = fmt.Sprintf("export const value%d = %d;\n", index, index)
	}
	program, targets := createLinterBenchmarkProgram(b, sources)
	rules := []ConfiguredRule{{
		Name:     "benchmark/no-op",
		Severity: rule.SeverityError,
		Run:      func(rule.RuleContext) rule.RuleListeners { return nil },
	}}
	opts := benchmarkProgramLintOptions(program, targets, rules)
	consumer := rule.DiagnosticConsumer{Report: func(rule.RuleDiagnostic) {}}
	benchmarkProgramResultSink = runLintRulesInProgram(opts, consumer)

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		benchmarkProgramResultSink = runLintRulesInProgram(opts, consumer)
	}
}
