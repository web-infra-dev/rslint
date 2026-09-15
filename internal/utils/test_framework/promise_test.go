package test_framework_test

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
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

func firstPromiseTestNode(t *testing.T, code string, kind ast.Kind) *ast.Node {
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

func TestAbruptCompletionPropagatesFailure(t *testing.T) {
	tests := []struct {
		name string
		code string
		kind ast.Kind
		want bool
	}{
		{
			name: "catch does not intercept returned promise rejection",
			code: `function callback() { try { return pending; } catch (error) {} }`,
			kind: ast.KindReturnStatement,
			want: true,
		},
		{
			name: "catch intercepts awaited promise rejection",
			code: `async function callback() { try { return await pending; } catch (error) {} }`,
			kind: ast.KindAwaitExpression,
			want: false,
		},
		{
			name: "rethrow preserves awaited promise rejection",
			code: `async function callback() { try { await pending; } catch (error) { throw error; } }`,
			kind: ast.KindAwaitExpression,
			want: true,
		},
		{
			name: "finally return suppresses returned promise",
			code: `function callback() { try { return pending; } finally { return; } }`,
			kind: ast.KindReturnStatement,
			want: false,
		},
		{
			name: "uncaught throw propagates failure",
			code: `function callback() { throw new Error("no"); }`,
			kind: ast.KindThrowStatement,
			want: true,
		},
		{
			name: "catch suppresses throw",
			code: `function callback() { try { throw new Error("caught"); } catch {} }`,
			kind: ast.KindThrowStatement,
			want: false,
		},
		{
			name: "rethrow preserves thrown failure",
			code: `function callback() { try { throw new Error("caught"); } catch (error) { throw error; } }`,
			kind: ast.KindThrowStatement,
			want: true,
		},
		{
			name: "finally return suppresses throw",
			code: `function callback() { try { throw new Error("suppressed"); } finally { return; } }`,
			kind: ast.KindThrowStatement,
			want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			node := firstPromiseTestNode(t, test.code, test.kind)
			if got := testFramework.AbruptCompletionPropagatesFailure(node, nil); got != test.want {
				t.Errorf("AbruptCompletionPropagatesFailure = %t, want %t", got, test.want)
			}
		})
	}
}
