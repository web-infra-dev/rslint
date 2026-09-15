package utils

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
)

func TestIsWriteReferenceThroughSatisfiesExpression(t *testing.T) {
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/write-reference.ts",
		Path:     "/write-reference.ts",
	}, `(callback satisfies (() => void)) = replacement;`, core.ScriptKindTS)

	var callback *ast.Node
	sourceFile.AsNode().ForEachChild(func(node *ast.Node) bool {
		var visit func(*ast.Node) bool
		visit = func(current *ast.Node) bool {
			if current.Kind == ast.KindIdentifier && current.Text() == "callback" {
				callback = current
				return true
			}
			return current.ForEachChild(visit)
		}
		return visit(node)
	})
	if callback == nil {
		t.Fatal("callback identifier not found")
	}
	if !IsWriteReference(callback) {
		t.Fatal("satisfies-wrapped assignment target was not classified as a write")
	}
}
