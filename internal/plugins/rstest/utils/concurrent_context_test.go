package utils_test

import (
	"fmt"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestRstestConcurrentContextIsSharedAndLazy(t *testing.T) {
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
	if len(analysis.Callbacks().Functions) == 0 {
		t.Fatal("test callback fixture was not collected")
	}

	var markerCall *ast.Node
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindCallExpression {
			call := node.AsCallExpression()
			if call != nil && call.Expression != nil &&
				call.Expression.Kind == ast.KindIdentifier &&
				call.Expression.AsIdentifier().Text == "marker" {
				markerCall = node
				return true
			}
		}
		return node.ForEachChild(visit)
	}
	sourceFile.Node.ForEachChild(visit)
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
	Name:             "rstest/concurrent-ownership-probe",
	RequiresTypeInfo: true,
	Schema:           rule.EmptyArraySchema,
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
		},
		[]rule_tester.InvalidTestCase{
			{Code: `test.concurrent("x", () => marker());`, Errors: reported},
			{Code: `const t = test.concurrent; t("x", () => marker());`, Errors: reported},
			{Code: `describe.concurrent("s", () => test("x", () => marker()));`, Errors: reported},
			{Code: `describe.sequential("s", () => test.concurrent("x", () => marker()));`, Errors: reported},
			{Code: `test.concurrent("x", callback); function callback() { marker(); }`, Errors: reported},
			{Code: `describe.concurrent("s", suite); function suite() { test("x", callback); } function callback() { marker(); }`, Errors: reported},
			{Code: `test.sequential("a", callback); test.concurrent("b", callback); function callback() { marker(); }`, Errors: reported},
			// A closure declared inside a concurrent callback runs as part of
			// that concurrent test whenever it runs at all.
			{Code: `test.concurrent("x", () => { const helper = () => marker(); });`, Errors: reported},
			{Code: `describe.concurrent("s", () => { beforeEach(() => marker()); });`, Errors: reported},
		},
	)
}
