package reactutil

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
)

func TestGetJsxStringLiteralValue(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name, code, want string
	}{
		{name: "entity forms token", code: `<iframe sandbox="allow&#x2D;forms" />`, want: "allow-forms"},
		{name: "entity forms delimiter", code: `<iframe sandbox="allow-forms&#x20;allow-modals" />`, want: "allow-forms allow-modals"},
		{name: "plain string", code: `<iframe sandbox="allow-popups" />`, want: "allow-popups"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			sourceFile := parser.ParseSourceFile(
				ast.SourceFileParseOptions{FileName: "/component.tsx", Path: "/component.tsx"},
				testCase.code,
				core.ScriptKindTSX,
			)
			var initializer *ast.Node
			var visit func(*ast.Node) bool
			visit = func(node *ast.Node) bool {
				if node.Kind == ast.KindJsxAttribute && GetJsxPropName(node) == "sandbox" {
					initializer = node.AsJsxAttribute().Initializer
					return true
				}
				node.ForEachChild(visit)
				return false
			}
			visit(sourceFile.AsNode())

			got, ok := GetJsxStringLiteralValue(sourceFile, initializer)
			if !ok || got != testCase.want {
				t.Fatalf("GetJsxStringLiteralValue() = %q, %v; want %q, true", got, ok, testCase.want)
			}
		})
	}
}
