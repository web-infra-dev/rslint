package no_identical_title

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

func TestNewRuleParsesEachCallOnce(t *testing.T) {
	parseCount := 0
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.ts",
		Path:     "/test.ts",
	}, "describe(); nested();", core.ScriptKindTS)
	calls := make([]*ast.Node, 0, 2)
	var collectCalls func(*ast.Node) bool
	collectCalls = func(node *ast.Node) bool {
		if node.Kind == ast.KindCallExpression {
			calls = append(calls, node)
		}
		return node.ForEachChild(collectCalls)
	}
	sourceFile.AsNode().ForEachChild(collectCalls)
	if len(calls) != 2 {
		t.Fatalf("parsed %d CallExpressions, want 2", len(calls))
	}
	describeCall, nestedCall := calls[0], calls[1]
	r := NewRule(Config{
		Name: "test/no-identical-title",
		Parse: func(node *ast.Node, ctx rule.RuleContext) *ParsedCall {
			parseCount++
			if node != describeCall {
				return nil
			}
			return &ParsedCall{Call: &testFramework.ParsedCall{Kind: testFramework.FnKindDescribe}}
		},
	})

	listeners := r.Run(rule.RuleContext{}, nil)
	listeners[ast.KindCallExpression](describeCall)
	listeners[ast.KindCallExpression](nestedCall)
	listeners[rule.ListenerOnExit(ast.KindCallExpression)](nestedCall)
	listeners[rule.ListenerOnExit(ast.KindCallExpression)](describeCall)

	if parseCount != 2 {
		t.Fatalf("Parse called %d times, want once for each CallExpression", parseCount)
	}
}
