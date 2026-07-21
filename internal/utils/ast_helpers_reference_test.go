package utils

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
)

func TestIsIdentifierInTypeReference(t *testing.T) {
	source := `
type Alias = TypeOnly;
interface Interface extends Heritage.Nested {}
class Implements implements Impl.Nested {}
class Extends extends Runtime.Base {}
const asserted = runtimeValue as Assertion;
type Query = typeof Queried.member;
`
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.ts",
		Path:     "/test.ts",
	}, source, core.ScriptKindTS)

	tests := []struct {
		name string
		want bool
	}{
		{name: "TypeOnly", want: true},
		{name: "Heritage", want: true},
		{name: "Impl", want: true},
		{name: "Runtime", want: false},
		{name: "runtimeValue", want: false},
		{name: "Assertion", want: true},
		{name: "Queried", want: true},
	}
	for _, test := range tests {
		node := findIdentifierByText(t, sourceFile.AsNode(), test.name)
		if got := IsIdentifierInTypeReference(node); got != test.want {
			t.Errorf("IsIdentifierInTypeReference(%s) = %v, want %v", test.name, got, test.want)
		}
	}
}

func TestIsNonReferenceIdentifierJsxNames(t *testing.T) {
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.tsx",
		Path:     "/test.tsx",
	}, `const view = <><div data-id="x" /><Component /><NS.Member /></>;`, core.ScriptKindTSX)

	tests := []struct {
		name string
		want bool
	}{
		{name: "div", want: true},
		{name: "data-id", want: true},
		{name: "Component", want: false},
		{name: "NS", want: false},
		{name: "Member", want: true},
	}
	for _, test := range tests {
		node := findIdentifierByText(t, sourceFile.AsNode(), test.name)
		if got := IsNonReferenceIdentifier(node); got != test.want {
			t.Errorf("IsNonReferenceIdentifier(%s) = %v, want %v", test.name, got, test.want)
		}
	}
}

func findIdentifierByText(t *testing.T, root *ast.Node, text string) *ast.Node {
	t.Helper()
	var found *ast.Node
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node == nil || found != nil {
			return
		}
		if ast.IsIdentifier(node) && node.Text() == text {
			found = node
			return
		}
		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return found != nil
		})
	}
	walk(root)
	if found == nil {
		t.Fatalf("identifier %q not found", text)
	}
	return found
}
