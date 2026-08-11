package utils

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
)

func TestGetRequireCall(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		expression  string
		literalOnly bool
		want        bool
	}{
		{name: "bare string literal", expression: `require("pkg")`, literalOnly: true, want: true},
		{name: "parenthesized call callee and argument", expression: `(((require))(("pkg")))`, literalOnly: true, want: true},
		{name: "template literal accepted by broad mode", expression: "require(`pkg`)", literalOnly: true, want: true},
		{name: "dynamic argument accepted when unchecked", expression: `require(name)`, want: true},
		{name: "dynamic argument rejected in literal mode", expression: `require(name)`, literalOnly: true},
		{name: "optional call remains caller policy", expression: `require?.("pkg")`, literalOnly: true, want: true},
		{name: "member callee", expression: `loader.require("pkg")`, literalOnly: true},
		{name: "no arguments", expression: `require()`, literalOnly: true},
		{name: "multiple arguments", expression: `require("pkg", options)`, literalOnly: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			source := "const value = " + test.expression + ";"
			sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
				FileName: "/test.ts",
				Path:     "/test.ts",
			}, source, core.ScriptKindTS)
			declaration := sourceFile.Statements.Nodes[0].AsVariableStatement().
				DeclarationList.AsVariableDeclarationList().Declarations.Nodes[0].AsVariableDeclaration()

			got := GetRequireCall(declaration.Initializer, test.literalOnly) != nil
			if got != test.want {
				t.Fatalf("GetRequireCall(%s, %v) = %v, want %v", test.expression, test.literalOnly, got, test.want)
			}
		})
	}
}
