package test_framework

import (
	"slices"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
)

func parseFirstCall(t *testing.T, code string) (*ast.SourceFile, *ast.Node) {
	t.Helper()
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.ts",
		Path:     "/test.ts",
	}, code, core.ScriptKindTS)

	call := findFirstNodeOfKind(sourceFile, ast.KindCallExpression)
	if call == nil {
		t.Fatalf("no call expression in %q", code)
	}
	return sourceFile, call
}

// parseFirstNodeOfKind returns the first node of kind in traversal order, so
// nested constructs resolve to the outermost one.
func parseFirstNodeOfKind(t *testing.T, code string, kind ast.Kind) *ast.Node {
	t.Helper()
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.ts",
		Path:     "/test.ts",
	}, code, core.ScriptKindTS)

	node := findFirstNodeOfKind(sourceFile, kind)
	if node == nil {
		t.Fatalf("no %v node in %q", kind, code)
	}
	return node
}

func findFirstNodeOfKind(sourceFile *ast.SourceFile, kind ast.Kind) *ast.Node {
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
	return found
}

func TestGetMemberEntries(t *testing.T) {
	tests := []struct {
		code      string
		wantNames []string
	}{
		{
			code:      "describe.only()",
			wantNames: []string{"describe", "only"},
		},
		{
			code:      `describe[ "only" ]("math", () => {})`,
			wantNames: []string{"describe", "only"},
		},
		{
			code:      "describe[\n  \"only\"\n](\"math\", () => {})",
			wantNames: []string{"describe", "only"},
		},
		{
			code:      "test.concurrent.only()",
			wantNames: []string{"test", "concurrent", "only"},
		},
	}

	for _, test := range tests {
		t.Run(test.code, func(t *testing.T) {
			_, call := parseFirstCall(t, test.code)
			entries := GetMemberEntries(call)
			names := make([]string, len(entries))
			for i, entry := range entries {
				names[i] = entry.Name
			}
			if !slices.Equal(names, test.wantNames) {
				t.Fatalf("member names = %v, want %v", names, test.wantNames)
			}
		})
	}
}

func TestGetMemberEntriesRejectsDynamicElementAccess(t *testing.T) {
	_, call := parseFirstCall(t, "describe[getMode()]()")
	if entries := GetMemberEntries(call); entries != nil {
		t.Fatalf("dynamic element access entries = %#v, want nil", entries)
	}
}

func TestResolveFirstIdentifier(t *testing.T) {
	_, call := parseFirstCall(t, "((test).each)`table`('name', () => {})")
	identifier := ResolveFirstIdentifier(call.AsCallExpression().Expression)
	if identifier == nil || identifier.Kind != ast.KindIdentifier || identifier.AsIdentifier().Text != "test" {
		t.Fatalf("first identifier = %#v, want test", identifier)
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
		// Unsupported computed keys break the chain entirely, unlike
		// GetMemberEntries which truncates.
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
			_, call := parseFirstCall(t, test.code)
			if got := CalleeChainName(call.AsCallExpression().Expression); got != test.want {
				t.Errorf("CalleeChainName = %q, want %q", got, test.want)
			}
		})
	}

	if got := CalleeChainName(nil); got != "" {
		t.Errorf("CalleeChainName(nil) = %q, want empty", got)
	}
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
			if !IsFunction(parseFirstNodeOfKind(t, form.code, form.kind)) {
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
			if IsFunction(parseFirstNodeOfKind(t, form.code, form.kind)) {
				t.Errorf("IsFunction should be false for %s", form.name)
			}
		})
	}

	if IsFunction(nil) {
		t.Error("IsFunction(nil) should be false")
	}
}
