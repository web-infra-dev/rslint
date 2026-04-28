package benchmark

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var (
	benchmarkSyntaxSink   atomic.Int64
	benchmarkTypeSink     atomic.Int64
	benchmarkSemanticSink int
)

func BenchmarkLinterSyntaxRules(b *testing.B) {
	program := createBenchmarkProgramInDir(b, filepath.Join(b.TempDir(), "syntax"), 32, false)
	rules := benchmarkSyntaxRules()
	getRules := func(*ast.SourceFile) []linter.ConfiguredRule { return rules }
	onDiag := func(rule.RuleDiagnostic) {}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		linter.RunLinterInProgram(
			program,
			nil,
			nil,
			utils.ExcludePaths,
			getRules,
			false,
			onDiag,
			nil,
			nil,
		)
	}
}

func BenchmarkLinterTypeAwareRules(b *testing.B) {
	program := createBenchmarkProgramInDir(b, filepath.Join(b.TempDir(), "type-aware"), 32, false)
	rules := benchmarkTypeAwareRules()
	typeInfoFiles := benchmarkTypeInfoFiles(program)
	getRules := func(*ast.SourceFile) []linter.ConfiguredRule { return rules }
	onDiag := func(rule.RuleDiagnostic) {}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		linter.RunLinterInProgram(
			program,
			nil,
			nil,
			utils.ExcludePaths,
			getRules,
			false,
			onDiag,
			typeInfoFiles,
			nil,
		)
	}
}

func BenchmarkLinterSemanticDiagnostics(b *testing.B) {
	projectDir := createBenchmarkProjectInDir(b, filepath.Join(b.TempDir(), "semantic"), 32, true)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		// Reload the program outside the timed region so each iteration measures
		// a fresh semantic pass instead of checker-side diagnostic caches.
		b.StopTimer()
		program := loadBenchmarkProgramInDir(b, projectDir)
		sourceFiles := benchmarkRootSourceFiles(b, program)
		b.StartTimer()

		total := 0
		for _, sourceFile := range sourceFiles {
			total += len(program.GetSemanticDiagnostics(ctx, sourceFile))
		}
		benchmarkSemanticSink = total
	}
}

func benchmarkSyntaxRules() []linter.ConfiguredRule {
	return []linter.ConfiguredRule{
		{
			Name:     "bench-syntax-vars",
			Severity: rule.SeverityWarning,
			Run: func(ctx rule.RuleContext) rule.RuleListeners {
				return rule.RuleListeners{
					ast.KindVariableDeclarationList: func(node *ast.Node) {
						if node.Flags&ast.NodeFlagsBlockScoped == 0 {
							benchmarkSyntaxSink.Add(1)
						}
					},
					ast.KindBinaryExpression: func(node *ast.Node) {
						bin := node.AsBinaryExpression()
						if bin == nil {
							return
						}
						op := bin.OperatorToken.Kind
						if op == ast.KindEqualsEqualsToken || op == ast.KindExclamationEqualsToken {
							benchmarkSyntaxSink.Add(1)
						}
					},
				}
			},
		},
		{
			Name:     "bench-syntax-control-flow",
			Severity: rule.SeverityWarning,
			Run: func(ctx rule.RuleContext) rule.RuleListeners {
				return rule.RuleListeners{
					ast.KindDebuggerStatement: func(node *ast.Node) {
						benchmarkSyntaxSink.Add(1)
					},
					ast.KindForStatement: func(node *ast.Node) {
						benchmarkSyntaxSink.Add(1)
					},
					ast.KindSwitchStatement: func(node *ast.Node) {
						benchmarkSyntaxSink.Add(1)
					},
				}
			},
		},
	}
}

func benchmarkTypeAwareRules() []linter.ConfiguredRule {
	return []linter.ConfiguredRule{
		{
			Name:             "bench-type-identifiers",
			Severity:         rule.SeverityWarning,
			RequiresTypeInfo: true,
			Run: func(ctx rule.RuleContext) rule.RuleListeners {
				return rule.RuleListeners{
					ast.KindIdentifier: func(node *ast.Node) {
						if ctx.TypeChecker == nil {
							return
						}
						if ctx.TypeChecker.GetSymbolAtLocation(node) != nil {
							benchmarkTypeSink.Add(1)
						}
					},
				}
			},
		},
		{
			Name:             "bench-type-expressions",
			Severity:         rule.SeverityWarning,
			RequiresTypeInfo: true,
			Run: func(ctx rule.RuleContext) rule.RuleListeners {
				return rule.RuleListeners{
					ast.KindPropertyAccessExpression: func(node *ast.Node) {
						if ctx.TypeChecker == nil {
							return
						}
						if ctx.TypeChecker.GetTypeAtLocation(node) != nil {
							benchmarkTypeSink.Add(1)
						}
					},
					ast.KindCallExpression: func(node *ast.Node) {
						if ctx.TypeChecker == nil {
							return
						}
						callExpr := node.AsCallExpression()
						if callExpr == nil {
							return
						}
						if ctx.TypeChecker.GetTypeAtLocation(callExpr.Expression) != nil {
							benchmarkTypeSink.Add(1)
						}
					},
				}
			},
		},
	}
}

func createBenchmarkProgramInDir(b *testing.B, configDir string, fileCount int, injectTypeError bool) *compiler.Program {
	b.Helper()

	createBenchmarkProjectInDir(b, configDir, fileCount, injectTypeError)
	return loadBenchmarkProgramInDir(b, configDir)
}

func createBenchmarkProjectInDir(b *testing.B, configDir string, fileCount int, injectTypeError bool) string {
	b.Helper()

	srcDir := filepath.Join(configDir, "src")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		b.Fatalf("failed to create benchmark source dir %s: %v", srcDir, err)
	}

	for i := range fileCount {
		filePath := filepath.Join(srcDir, "file"+strconv.Itoa(i)+".ts")
		if err := os.WriteFile(filePath, []byte(benchmarkFileContents(i, injectTypeError)), 0o644); err != nil {
			b.Fatalf("failed to write %s: %v", filePath, err)
		}
	}

	tsconfig := `{"compilerOptions":{"strict":true,"target":"esnext","module":"esnext"},"include":["src/**/*.ts"]}`
	if err := os.WriteFile(filepath.Join(configDir, "tsconfig.json"), []byte(tsconfig), 0o644); err != nil {
		b.Fatalf("failed to write tsconfig.json: %v", err)
	}

	return configDir
}

func loadBenchmarkProgramInDir(b *testing.B, configDir string) *compiler.Program {
	b.Helper()

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	host := utils.CreateCompilerHost(configDir, fs)
	program, err := utils.CreateProgram(true, fs, configDir, "tsconfig.json", host)
	if err != nil {
		b.Fatalf("failed to create program: %v", err)
	}

	return program
}

func benchmarkTypeInfoFiles(program *compiler.Program) map[string]struct{} {
	typeInfoFiles := make(map[string]struct{}, len(program.CommandLine().FileNames()))
	for _, fileName := range program.CommandLine().FileNames() {
		typeInfoFiles[fileName] = struct{}{}
	}
	return typeInfoFiles
}

func benchmarkRootSourceFiles(b *testing.B, program *compiler.Program) []*ast.SourceFile {
	b.Helper()

	rootFiles := program.CommandLine().FileNames()
	sourceFiles := make([]*ast.SourceFile, 0, len(rootFiles))
	for _, fileName := range rootFiles {
		sourceFile := program.GetSourceFile(tspath.NormalizePath(fileName))
		if sourceFile == nil {
			b.Fatalf("source file %s not found in program", fileName)
		}
		sourceFiles = append(sourceFiles, sourceFile)
	}
	return sourceFiles
}

func benchmarkFileContents(index int, injectTypeError bool) string {
	suffix := strconv.Itoa(index)
	typeErrorLine := ""
	if injectTypeError {
		typeErrorLine = "  const brokenValue: number = 'not-a-number';\n"
	}

	return "type Status" + suffix + " = 'open' | 'closed';\n" +
		"interface User" + suffix + " { id: number; name: string; nested?: { value: number } }\n" +
		"const records" + suffix + ": User" + suffix + "[] = [];\n" +
		"export async function process" + suffix + "(input: string, users: User" + suffix + "[]): Promise<number> {\n" +
		"  var total = 0;\n" +
		"  for (var i = 0; i < users.length; i++) {\n" +
		"    const user = users[i];\n" +
		"    if (input == '') {\n" +
		"      debugger;\n" +
		"    }\n" +
		"    switch (user.id) {\n" +
		"    case 1:\n" +
		"      total += user.id;\n" +
		"      break;\n" +
		"    default:\n" +
		"      total += user.nested?.value ?? 0;\n" +
		"    }\n" +
		"    const payload: any = { value: Promise.resolve(user.name) };\n" +
		"    const resolved = await payload.value;\n" +
		"    if (resolved != null) {\n" +
		"      total += resolved.length;\n" +
		"    }\n" +
		typeErrorLine +
		"  }\n" +
		"  records" + suffix + ".push(...users);\n" +
		"  return total + records" + suffix + ".length;\n" +
		"}\n"
}
