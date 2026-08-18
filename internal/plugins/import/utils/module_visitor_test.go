package utils

import (
	"slices"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
)

func TestVisitModulesCommonJSUsesESTreeStringLiterals(t *testing.T) {
	t.Parallel()

	const source = `
require("plain");
require?.("optional");
((require))((("parenthesized")));
require(` + "`template`" + `);
require(1);
require(name);
require("first", "second");
loader.require("member");
import((("dynamic")));
import(` + "`dynamic-template`" + `);
`
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/modules.ts",
		Path:     "/modules.ts",
	}, source, core.ScriptKindTS)

	var visited []string
	listeners := VisitModules(func(module *ast.StringLiteralLike, _ *ast.Node) {
		visited = append(visited, module.Text())
	}, VisitModulesOptions{Commonjs: true, ESModule: true})
	callListener := listeners[ast.KindCallExpression]
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node.Kind == ast.KindCallExpression {
			callListener(node)
		}
		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	walk(sourceFile.AsNode())

	want := []string{"plain", "optional", "parenthesized", "dynamic"}
	if !slices.Equal(visited, want) {
		t.Fatalf("visited modules = %v, want %v", visited, want)
	}
}
