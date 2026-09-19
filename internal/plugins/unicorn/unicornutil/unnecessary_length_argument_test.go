package unicornutil

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
)

func TestUnnecessaryLengthArgumentHasSideEffect(t *testing.T) {
	for _, test := range []struct {
		expression string
		want       bool
	}{
		{expression: "start + 1"},
		{expression: "index ?? 0"},
		{expression: "flag ? 1 : 2"},
		{expression: "getStart()", want: true},
		{expression: "start + getOffset()", want: true},
		{expression: "array = other", want: true},
		{expression: "options.start", want: true},
		{expression: "tag`value`", want: true},
		{expression: "() => getStart()"},
	} {
		t.Run(test.expression, func(t *testing.T) {
			sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
				FileName: "/test.js",
				Path:     tspath.Path("/test.js"),
			}, "const value = "+test.expression+";", core.ScriptKindJS)
			declaration := sourceFile.Statements.Nodes[0].AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes[0].AsVariableDeclaration()
			if got := hasSideEffect(declaration.Initializer, true); got != test.want {
				t.Fatalf("hasSideEffect(%q) = %v, want %v", test.expression, got, test.want)
			}
		})
	}
}
