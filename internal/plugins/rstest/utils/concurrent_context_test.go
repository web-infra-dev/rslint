package utils_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/binder"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
)

// TestRstestConcurrentContextIsSharedPerFile covers the part only this test
// covers: every rule asking for the concurrent context of one file gets the
// same instance, so the ownership index is built once rather than per rule.
// Laziness itself is asserted by TestRstestConcurrentContextBuildsOwnershipLazily,
// which runs inside the package and can read the ownership field.
func TestRstestConcurrentContextIsSharedPerFile(t *testing.T) {
	sourceFile := parser.ParseSourceFile(
		ast.SourceFileParseOptions{
			FileName: "/concurrent-context.test.ts",
			Path:     "/concurrent-context.test.ts",
		},
		`test.concurrent("x", () => marker());`,
		core.ScriptKindTS,
	)
	cache := rule.NewFileCache()
	ctx := rule.RuleContext{SourceFile: sourceFile}.WithFileCache(cache)
	analysis := rstestUtils.GetRstestCallAnalysis(ctx)
	first := rstestUtils.GetRstestConcurrentContext(ctx, analysis)
	second := rstestUtils.GetRstestConcurrentContext(ctx, analysis)
	if first != second {
		t.Fatal("contexts for one file did not share concurrent ownership")
	}

	markerCall := findCallByCalleeName(sourceFile, "marker")
	if markerCall == nil {
		t.Fatal("marker call not found")
	}
	if !first.IsInConcurrentTest(markerCall) {
		t.Fatal("concurrent ownership was not resolved on first query")
	}
}

var expectContextProvenanceProbe = rule.Rule{
	Name:             "rstest/expect-context-provenance-probe",
	RequiresTypeInfo: true,
	Schema:           rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				parsed := analysis.ParseExpectCall(node)
				if parsed == nil {
					return
				}
				ctx.ReportNode(node, probeMessage(
					"contextProvenance",
					fmt.Sprintf("context=%t", parsed.FromTestContext),
				))
			},
		}
	},
}

func TestParsedRstestExpectCallPreservesContextProvenance(t *testing.T) {
	contextTrue := []rule_tester.InvalidTestCaseError{{MessageId: "contextProvenance", Message: "context=true"}}
	contextFalse := []rule_tester.InvalidTestCaseError{{MessageId: "contextProvenance", Message: "context=false"}}
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &expectContextProvenanceProbe,
		[]rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			{Code: `test("x", ({ expect }) => expect(1).toBe(1));`, Errors: contextTrue},
			{Code: `test("x", ({ expect: check }) => check(1).toBe(1));`, Errors: contextTrue},
			{Code: `test("x", ctx => ctx.expect(1).toBe(1));`, Errors: contextTrue},
			{Code: `expect(1).toBe(1);`, Errors: contextFalse},
			{Code: `import { expect as check } from "@rstest/core"; check(1).toBe(1);`, Errors: contextFalse},
			{Code: `import * as rstest from "@rstest/core"; rstest.expect(1).toBe(1);`, Errors: contextFalse},
			{Code: `import { expect } from "@rstest/playwright"; expect(1).toBe(1);`, Errors: contextFalse},
			{Code: `import.meta.rstest.expect(1).toBe(1);`, Errors: contextFalse},
		},
	)
}

var executionModeProbe = rule.Rule{
	Name:             "rstest/execution-mode-probe",
	RequiresTypeInfo: true,
	Schema:           rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				parsed := analysis.ParseFnCall(node)
				if parsed == nil || parsed.ExecutionMode == rstestUtils.RstestExecutionDefault {
					return
				}
				ctx.ReportNode(node, probeMessage(
					"executionMode",
					fmt.Sprintf("mode=%d", parsed.ExecutionMode),
				))
			},
		}
	},
}

func TestRstestExecutionModeSurvivesAliasesAndUsesRuntimePriority(t *testing.T) {
	concurrent := []rule_tester.InvalidTestCaseError{{MessageId: "executionMode", Message: "mode=1"}}
	sequential := []rule_tester.InvalidTestCaseError{{MessageId: "executionMode", Message: "mode=2"}}
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &executionModeProbe,
		[]rule_tester.ValidTestCase{{Code: `test("x", cb);`}},
		[]rule_tester.InvalidTestCase{
			{Code: `test.concurrent("x", cb);`, Errors: concurrent},
			{Code: `test.sequential("x", cb);`, Errors: sequential},
			{Code: `const t = test.concurrent; t("x", cb);`, Errors: concurrent},
			{Code: `const suite = describe.sequential; suite("x", cb);`, Errors: sequential},
			{Code: `test.concurrent.sequential("x", cb);`, Errors: concurrent},
			{Code: `test.sequential.concurrent("x", cb);`, Errors: concurrent},
			{Code: `import.meta.rstest.test.concurrent("x", cb);`, Errors: concurrent},
		},
	)
}

var concurrentOwnershipProbe = rule.Rule{
	Name:   "rstest/concurrent-ownership-probe",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		concurrentContext := rstestUtils.GetRstestConcurrentContext(ctx, analysis)
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				if call == nil || call.Expression == nil || call.Expression.Kind != ast.KindIdentifier ||
					call.Expression.AsIdentifier().Text != "marker" ||
					!concurrentContext.IsInConcurrentTest(node) {
					return
				}
				ctx.ReportNode(node, probeMessage("concurrent", "concurrent=true"))
			},
		}
	},
}

func TestRstestConcurrentCallbackOwnership(t *testing.T) {
	reported := []rule_tester.InvalidTestCaseError{{MessageId: "concurrent", Message: "concurrent=true"}}
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &concurrentOwnershipProbe,
		[]rule_tester.ValidTestCase{
			{Code: `test("x", () => marker());`},
			{Code: `describe.concurrent("s", () => test.sequential("x", () => marker()));`},
			{Code: `function helper() { marker(); } test.concurrent("x", () => helper());`},
			{Code: `let callback = () => marker(); callback = () => {}; test.concurrent("x", callback);`},
			{Code: `function callback() { marker(); } callback = () => {}; test.concurrent("x", callback);`},
		},
		[]rule_tester.InvalidTestCase{
			{Code: `test.concurrent("x", () => marker());`, Errors: reported},
			{Code: `const t = test.concurrent; t("x", () => marker());`, Errors: reported},
			{Code: `describe.concurrent("s", () => test("x", () => marker()));`, Errors: reported},
			{Code: `describe.sequential("s", () => test.concurrent("x", () => marker()));`, Errors: reported},
			{Code: `test.concurrent("x", callback); function callback() { marker(); }`, Errors: reported},
			{Code: `describe.concurrent("s", suite); function suite() { test("x", callback); } function callback() { marker(); }`, Errors: reported},
			{Code: `test.sequential("a", callback); test.concurrent("b", callback); function callback() { marker(); }`, Errors: reported},
			{Code: `function register() { const callback = () => marker(); test.concurrent("x", callback); } register();`, Errors: reported},
			{Code: `{ const callback = () => marker(); test.concurrent("x", callback); }`, Errors: reported},
			{Code: `class C { static { const callback = () => marker(); test.concurrent("x", callback); } }`, Errors: reported},
			// A closure declared inside a concurrent callback runs as part of
			// that concurrent test whenever it runs at all.
			{Code: `test.concurrent("x", () => { const helper = () => marker(); });`, Errors: reported},
			{Code: `describe.concurrent("s", () => { beforeEach(() => marker()); });`, Errors: reported},
		},
	)
}

// findCallByCalleeName returns the first call whose callee is the identifier
// name, which the ownership fixtures use as a stand-in for an assertion.
func findCallByCalleeName(sourceFile *ast.SourceFile, name string) *ast.Node {
	var found *ast.Node
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindCallExpression {
			call := node.AsCallExpression()
			if call != nil && call.Expression != nil &&
				call.Expression.Kind == ast.KindIdentifier &&
				call.Expression.AsIdentifier().Text == name {
				found = node
				return true
			}
		}
		return node.ForEachChild(visit)
	}
	sourceFile.Node.ForEachChild(visit)
	return found
}

// TestRstestCallbackOwnershipSourceOnlyScopesAndWrites verifies that the
// binder resolves callback references without a TypeChecker, while mutable or
// reassigned bindings are never treated as stable callback ownership.
func TestRstestCallbackOwnershipSourceOnlyScopesAndWrites(t *testing.T) {
	for _, testCase := range []struct {
		name       string
		code       string
		concurrent bool
	}{
		{
			name:       "top level declaration is attributed",
			code:       `function cb() { marker(); } test.concurrent("x", cb);`,
			concurrent: true,
		},
		{
			name:       "top level arrow initializer is attributed",
			code:       `const cb = () => { marker(); }; test.concurrent("x", cb);`,
			concurrent: true,
		},
		{
			name:       "function scoped declaration is attributed",
			code:       `function register() { function cb() { marker(); } test.concurrent("x", cb); } register();`,
			concurrent: true,
		},
		{
			name:       "block scoped arrow initializer is attributed",
			code:       `{ const cb = () => { marker(); }; test.concurrent("x", cb); }`,
			concurrent: true,
		},
		{
			name:       "class static block arrow initializer is attributed",
			code:       `class C { static { const cb = () => { marker(); }; test.concurrent("x", cb); } }`,
			concurrent: true,
		},
		{
			name:       "nested shadow does not hide top level declaration",
			code:       `function cb() { marker(); } function other() { function cb() {} } test.concurrent("x", cb);`,
			concurrent: true,
		},
		{
			name:       "mutable initializer is not attributed",
			code:       `let cb = () => { marker(); }; test.concurrent("x", cb);`,
			concurrent: false,
		},
		{
			name:       "reassigned function declaration is not attributed",
			code:       `function cb() { marker(); } cb = () => {}; test.concurrent("x", cb);`,
			concurrent: false,
		},
		{
			name:       "write after registration conservatively rejects attribution",
			code:       `function cb() { marker(); } test.concurrent("x", cb); cb = () => {};`,
			concurrent: false,
		},
		{
			name:       "destructuring write is not attributed",
			code:       `function cb() { marker(); } ({ cb } = replacements); test.concurrent("x", cb);`,
			concurrent: false,
		},
		{
			name:       "satisfies wrapped write is not attributed",
			code:       `function cb() { marker(); } (cb satisfies (() => void)) = () => {}; test.concurrent("x", cb);`,
			concurrent: false,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			sourceFile := parser.ParseSourceFile(
				ast.SourceFileParseOptions{
					FileName: "/ownership.test.ts",
					Path:     "/ownership.test.ts",
				},
				testCase.code,
				core.ScriptKindTS,
			)
			binder.BindSourceFile(sourceFile)
			_, refsInit, _ := rule.ResolveLanguageDefaults(sourceFile.FileName(), rule.LanguageOptions{})
			ctx := rule.RuleContext{
				SourceFile: sourceFile,
				Refs:       rule.NewRefStore(sourceFile, &core.CompilerOptions{}, nil, refsInit),
			}.WithFileCache(rule.NewFileCache())
			context := rstestUtils.GetRstestConcurrentContext(ctx, rstestUtils.GetRstestCallAnalysis(ctx))
			marker := findCallByCalleeName(sourceFile, "marker")
			if marker == nil {
				t.Fatal("marker call not found")
			}
			if got := context.IsInConcurrentTest(marker); got != testCase.concurrent {
				t.Fatalf("IsInConcurrentTest = %t, want %t", got, testCase.concurrent)
			}
		})
	}
}

func TestRstestConcurrentCallbackOwnershipSourceOnlyProgram(t *testing.T) {
	code := `function register() {
  const callback = () => marker();
  test.concurrent("function", callback);
}
register();
{
  const callback = () => marker();
  test.concurrent("block", callback);
}
class C {
  static {
    const callback = () => marker();
    test.concurrent("static", callback);
  }
}
let reassigned = () => marker();
reassigned = () => {};
test.concurrent("reassigned", reassigned);
function replaced() { marker(); }
replaced = () => {};
test.concurrent("replaced", replaced);
var mutable = () => marker();
test.concurrent("mutable", mutable);
function destructured() { marker(); }
({ destructured } = replacements);
test.concurrent("destructured", destructured);`

	root := fixtures.GetRootDir()
	fileName := tspath.ResolvePath(root.Dir, "concurrent-ownership-source-only.ts")
	fs := internalUtils.NewOverlayVFS(root.FS, map[string]string{fileName: code})
	host := internalUtils.CreateCompilerHost(root.Dir, fs)
	sourceProgram, err := lintprogram.NewFromRoots(lintprogram.RootOptions{
		RootFileNames:   []string{fileName},
		Host:            host,
		CompilerOptions: &core.CompilerOptions{Module: core.ModuleKindESNext},
		SingleThreaded:  true,
	})
	if err != nil {
		t.Fatalf("NewFromRoots: %v", err)
	}
	if sourceProgram.CanProvideTypeChecker(sourceProgram.SourceFiles()[0]) {
		t.Fatal("expected a source-only Program with no TypeChecker")
	}

	lintPlan, err := linter.PrepareLintPlan(linter.PrepareLintPlanOptions{
		Programs:         []*lintprogram.Program{sourceProgram},
		TargetsByProgram: [][]string{{fileName}},
		SingleThreaded:   true,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
				Name:     concurrentOwnershipProbe.Name,
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return concurrentOwnershipProbe.Run(ctx, nil)
				},
			}}
		},
	})
	if err != nil {
		t.Fatalf("PrepareLintPlan: %v", err)
	}

	var lines []int
	if _, err := linter.RunLinter(linter.RunLinterOptions{
		SingleThreaded: true,
		LintPlan:       lintPlan,
		Consumer: rule.DiagnosticConsumer{Report: func(diagnostic rule.RuleDiagnostic) {
			line, _ := scanner.GetECMALineAndUTF16CharacterOfPosition(
				diagnostic.SourceFile,
				diagnostic.Range.Pos(),
			)
			lines = append(lines, line+1)
		}},
	}); err != nil {
		t.Fatalf("RunLinter: %v", err)
	}
	want := []int{2, 7, 12}
	if !slices.Equal(lines, want) {
		t.Fatalf("reported lines %v, want %v", lines, want)
	}
}
