package utils_test

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
)

// findFirstNodeOfKind parses code and returns the first node of the given
// kind in traversal order, so nested constructs resolve to the outermost one.
func findFirstNodeOfKind(t *testing.T, code string, kind ast.Kind) *ast.Node {
	t.Helper()
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.ts",
		Path:     "/test.ts",
	}, code, core.ScriptKindTS)

	var found *ast.Node
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Kind == kind {
			found = node
			return true
		}
		return node.ForEachChild(visit)
	}
	sourceFile.AsNode().ForEachChild(visit)
	if found == nil {
		t.Fatalf("no %v node in %q", kind, code)
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
		// Supported computed accessors.
		{`promise["then"](fn);`, true},
		{"promise[`catch`](fn);", true},
		// Dynamic keys are not statically knowable.
		{`promise[thenName](fn);`, false},
		{`promise.other(fn);`, false},
		// A bare call is not a member access.
		{`then(fn);`, false},
		// ESTree drops parentheses, so the Go port skips them.
		{`(promise.then)(fn);`, true},
		{`promise?.then(fn);`, true},
	}
	for _, test := range tests {
		t.Run(test.code, func(t *testing.T) {
			call := findFirstNodeOfKind(t, test.code, ast.KindCallExpression)
			if got := rstestUtils.IsPromiseChainCall(call); got != test.want {
				t.Errorf("IsPromiseChainCall = %t, want %t", got, test.want)
			}
		})
	}

	if rstestUtils.IsPromiseChainCall(nil) {
		t.Error("IsPromiseChainCall(nil) should be false")
	}
	notACall := findFirstNodeOfKind(t, `a.then;`, ast.KindPropertyAccessExpression)
	if rstestUtils.IsPromiseChainCall(notACall) {
		t.Error("IsPromiseChainCall should be false for a non-call node")
	}
}
