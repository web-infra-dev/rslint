package test_framework_test

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

func firstCall(t *testing.T, code string) *ast.Node {
	t.Helper()
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.ts",
		Path:     "/test.ts",
	}, code, core.ScriptKindTS)

	var found *ast.Node
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindCallExpression {
			found = node
			return true
		}
		return node.ForEachChild(visit)
	}
	sourceFile.AsNode().ForEachChild(visit)
	if found == nil {
		t.Fatalf("no call expression in %q", code)
	}
	return found
}

func TestIsPromiseChainCall(t *testing.T) {
	tests := []struct {
		code string
		want bool
	}{
		{`promise.then(fn);`, true},
		{`promise.then(fn, onError);`, true},
		{`promise.then(a, b, c);`, false},
		{`promise.then();`, false},
		{`promise.catch(fn);`, true},
		{`promise.catch(a, b);`, false},
		{`promise.finally(fn);`, true},
		{`promise.finally(a, b);`, false},
		{`promise["then"](fn);`, true},
		{"promise[`catch`](fn);", true},
		{`promise[thenName](fn);`, false},
		{`promise.other(fn);`, false},
		{`then(fn);`, false},
		{`(promise.then)(fn);`, true},
		{`promise?.then(fn);`, true},
	}
	for _, test := range tests {
		t.Run(test.code, func(t *testing.T) {
			if got := testFramework.IsPromiseChainCall(firstCall(t, test.code)); got != test.want {
				t.Errorf("IsPromiseChainCall = %t, want %t", got, test.want)
			}
		})
	}

	if testFramework.IsPromiseChainCall(nil) {
		t.Error("IsPromiseChainCall(nil) should be false")
	}
}
