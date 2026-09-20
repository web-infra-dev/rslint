package prefer_lowercase_title

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

func TestNewRuleParsesEachCallOnce(t *testing.T) {
	parseCount := 0
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.ts",
		Path:     "/test.ts",
	}, "describe(); nested();", core.ScriptKindTS)
	var calls []*ast.Node
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
		Name: "test/prefer-lowercase-title",
		Prepare: func(ctx rule.RuleContext) Runtime {
			return Runtime{
				Parse: func(node *ast.Node) *testFramework.ParsedCall {
					parseCount++
					if node != describeCall {
						return nil
					}
					return &testFramework.ParsedCall{Kind: testFramework.FnKindDescribe}
				},
			}
		},
	})

	listeners := r.Run(rule.RuleContext{}, nil)
	listeners[ast.KindCallExpression](describeCall)
	listeners[ast.KindCallExpression](nestedCall)
	listeners[rule.ListenerOnExit(ast.KindCallExpression)](nestedCall)
	listeners[rule.ListenerOnExit(ast.KindCallExpression)](describeCall)

	if parseCount != 4 {
		t.Fatalf("Parse called %d times, want 4 (2 enter + 2 exit)", parseCount)
	}
}

func TestResolveOptionsIgnoreExpansion(t *testing.T) {
	cfg := Config{
		DescribeAliases: []string{"describe", "fdescribe", "xdescribe"},
		TestAliases:     []string{"test", "xtest"},
		ItAliases:       []string{"it", "xit", "fit"},
	}
	opts := resolveOptions([]any{map[string]interface{}{
		"ignore": []interface{}{"describe", "it"},
	}}, cfg)

	for _, name := range []string{"describe", "fdescribe", "xdescribe", "it", "xit", "fit"} {
		if _, ok := opts.ignoredNames[name]; !ok {
			t.Errorf("expected %q to be in ignoredNames", name)
		}
	}
	if _, ok := opts.ignoredNames["test"]; ok {
		t.Error("expected test to NOT be in ignoredNames")
	}
	if _, ok := opts.ignoredNames["xtest"]; ok {
		t.Error("expected xtest to NOT be in ignoredNames")
	}
}

func TestResolveOptionsDefaults(t *testing.T) {
	cfg := Config{}
	opts := resolveOptions(nil, cfg)
	if opts.ignoreTopLevelDescribe {
		t.Error("expected ignoreTopLevelDescribe to default to false")
	}
	if opts.ignoreTodos {
		t.Error("expected ignoreTodos to default to false")
	}
	if len(opts.allowedPrefixes) != 0 {
		t.Error("expected allowedPrefixes to be empty")
	}
}
