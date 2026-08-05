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

func TestIsFunction(t *testing.T) {
	functionForms := []struct {
		name string
		code string
		kind ast.Kind
	}{
		{"function declaration", `function f() {}`, ast.KindFunctionDeclaration},
		{"function expression", `const f = function () {};`, ast.KindFunctionExpression},
		{"arrow function", `const f = () => {};`, ast.KindArrowFunction},
		{"method declaration", `class C { m() {} }`, ast.KindMethodDeclaration},
		{"constructor", `class C { constructor() {} }`, ast.KindConstructor},
		{"get accessor", `class C { get p() { return 1; } }`, ast.KindGetAccessor},
		{"set accessor", `class C { set p(v) {} }`, ast.KindSetAccessor},
	}
	for _, form := range functionForms {
		t.Run(form.name, func(t *testing.T) {
			node := findFirstNodeOfKind(t, form.code, form.kind)
			if !rstestUtils.IsFunction(node) {
				t.Errorf("IsFunction should be true for %s", form.name)
			}
		})
	}

	nonFunctionForms := []struct {
		name string
		code string
		kind ast.Kind
	}{
		{"call expression", `f();`, ast.KindCallExpression},
		{"class declaration", `class C {}`, ast.KindClassDeclaration},
		{"identifier", `x;`, ast.KindIdentifier},
		{"property access", `a.b;`, ast.KindPropertyAccessExpression},
	}
	for _, form := range nonFunctionForms {
		t.Run(form.name, func(t *testing.T) {
			node := findFirstNodeOfKind(t, form.code, form.kind)
			if rstestUtils.IsFunction(node) {
				t.Errorf("IsFunction should be false for %s", form.name)
			}
		})
	}

	if rstestUtils.IsFunction(nil) {
		t.Error("IsFunction(nil) should be false")
	}
}

func TestCalleeChainName(t *testing.T) {
	tests := []struct {
		code string
		want string
	}{
		{`foo();`, "foo"},
		{`foo.bar();`, "foo.bar"},
		{`foo.bar.baz();`, "foo.bar.baz"},
		{`foo["bar"]();`, "foo.bar"},
		{"foo[`bar`]();", "foo.bar"},
		// Identifier keys are supported accessor names upstream, matching
		// getNodeName's treatment of computed identifier properties.
		{`foo[bar]();`, "foo.bar"},
		// Unsupported computed keys break the chain entirely.
		{`foo[bar.baz]();`, ""},
		{`foo[1]();`, ""},
		{"foo[`bar${x}`]();", ""},
		// Calls and news are peeled.
		{`foo().bar();`, "foo.bar"},
		{`new Foo().bar();`, "Foo.bar"},
		// new (require("x")).y() parses entirely as a NewExpression, so the
		// outer call has to be forced with parentheses to exercise the peel.
		{`(new (require("x")).y)();`, "require.y"},
		// Parentheses and optional chaining.
		{`(foo.bar)();`, "foo.bar"},
		{`foo?.bar();`, "foo.bar"},
		// Tagged templates resolve through their tag.
		{"foo.bar`x`();", "foo.bar"},
		// A non-identifier root yields no name, and property access on it
		// keeps the empty left side.
		{`"str".includes();`, ""},
	}
	for _, test := range tests {
		t.Run(test.code, func(t *testing.T) {
			call := findFirstNodeOfKind(t, test.code, ast.KindCallExpression)
			got := rstestUtils.CalleeChainName(call.AsCallExpression().Expression)
			if got != test.want {
				t.Errorf("CalleeChainName = %q, want %q", got, test.want)
			}
		})
	}

	if got := rstestUtils.CalleeChainName(nil); got != "" {
		t.Errorf("CalleeChainName(nil) = %q, want empty", got)
	}
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
