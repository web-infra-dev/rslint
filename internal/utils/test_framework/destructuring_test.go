package test_framework_test

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

func firstCallNamed(t *testing.T, code, name string) *ast.Node {
	t.Helper()
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.ts",
		Path:     "/test.ts",
	}, code, core.ScriptKindTS)
	var found *ast.Node
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindCallExpression {
			expression := ast.SkipParentheses(node.AsCallExpression().Expression)
			if expression != nil && expression.Kind == ast.KindIdentifier &&
				expression.Text() == name {
				found = node
				return true
			}
		}
		return node.ForEachChild(visit)
	}
	sourceFile.AsNode().ForEachChild(visit)
	if found == nil {
		t.Fatalf("no %s call in %q", name, code)
	}
	return found
}

func TestStaticBindingForValue(t *testing.T) {
	tests := []struct {
		code string
		want string
	}{
		{`const [pending] = [make()];`, "pending"},
		{`const [, pending] = [other, make()];`, "pending"},
		{`const { pending } = { pending: make() };`, "pending"},
		{`const { source: pending } = { source: make() };`, "pending"},
		{`const { nested: [pending] } = { nested: [make()] };`, "pending"},
		{`const [pending = fallback] = [make()];`, "pending"},
		{`let pending; [pending] = [make()];`, "pending"},
		{`let pending; ({ pending } = { pending: make() });`, "pending"},
		{`let pending; ({ source: pending } = { source: make() });`, "pending"},
		{`let pending; ({ nested: [pending] } = { nested: [make()] });`, "pending"},
		{`let pending; [pending = fallback] = [make()];`, "pending"},
		{`const [...pending] = [make()];`, ""},
		{`let pending; [...pending] = [make()];`, ""},
		{`const [pending] = [...values, make()];`, ""},
		{`const { [key]: pending } = { [key]: make() };`, ""},
		{`let pending; ({ [key]: pending } = { [key]: make() });`, ""},
		{`const { pending } = { ...values, pending: make() };`, ""},
	}
	for _, test := range tests {
		t.Run(test.code, func(t *testing.T) {
			call := firstCallNamed(t, test.code, "make")
			binding := testFramework.StaticBindingForValue(call)
			got := ""
			if binding != nil {
				got = binding.Text()
			}
			if got != test.want {
				t.Errorf("StaticBindingForValue = %q, want %q", got, test.want)
			}
		})
	}
}
