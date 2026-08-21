package utils

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
)

func TestForEachVariableDeclarationBinding(t *testing.T) {
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.ts",
		Path:     "/test.ts",
	}, "const simple = 0, {a, b: renamed, nested: {c}, ...rest} = object, [first, , ...tail] = array;", core.ScriptKindTS)

	statement := sourceFile.Statements.Nodes[0].AsVariableStatement()
	if statement == nil || statement.DeclarationList == nil {
		t.Fatal("test source did not produce a variable declaration list")
	}

	type binding struct {
		name        string
		declaration string
	}
	var got []binding
	ForEachVariableDeclarationBinding(statement.DeclarationList, func(declaration *ast.Node, _ *ast.Node, name string) {
		got = append(got, binding{
			name:        name,
			declaration: TrimmedNodeText(sourceFile, declaration),
		})
	})

	want := []binding{
		{name: "simple", declaration: "simple = 0"},
		{name: "a", declaration: "{a, b: renamed, nested: {c}, ...rest} = object"},
		{name: "renamed", declaration: "{a, b: renamed, nested: {c}, ...rest} = object"},
		{name: "c", declaration: "{a, b: renamed, nested: {c}, ...rest} = object"},
		{name: "rest", declaration: "{a, b: renamed, nested: {c}, ...rest} = object"},
		{name: "first", declaration: "[first, , ...tail] = array"},
		{name: "tail", declaration: "[first, , ...tail] = array"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ForEachVariableDeclarationBinding() = %#v, want %#v", got, want)
	}

	// Malformed callers and optional callbacks are deliberately no-ops because
	// declaration collectors often run while recovering from parser errors.
	ForEachVariableDeclarationBinding(nil, func(*ast.Node, *ast.Node, string) {
		t.Fatal("nil declaration list should not invoke callback")
	})
	ForEachVariableDeclarationBinding(sourceFile.AsNode(), func(*ast.Node, *ast.Node, string) {
		t.Fatal("non-declaration-list node should not invoke callback")
	})
	ForEachVariableDeclarationBinding(statement.DeclarationList, nil)
}
