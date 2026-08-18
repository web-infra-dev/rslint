package utils

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
)

func TestGetRequireCall(t *testing.T) {
	t.Parallel()
	if GetRequireCall(nil, false) != nil {
		t.Fatal("GetRequireCall(nil) returned a call")
	}

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

func TestGetStaticRequireCall(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		expression string
		want       bool
	}{
		{name: "string literal", expression: `require("pkg")`, want: true},
		{name: "type arguments are transparent", expression: `require<unknown>("pkg")`, want: true},
		{name: "parentheses are transparent", expression: `(((require))(("pkg")))`, want: true},
		{name: "template literal is not an ESTree Literal", expression: "require(`pkg`)"},
		{name: "dynamic argument", expression: `require(name)`},
		{name: "optional call", expression: `require?.("pkg")`},
		{name: "member callee", expression: `loader.require("pkg")`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			initializer := parseInitializer(t, test.expression)
			if got := GetStaticRequireCall(initializer) != nil; got != test.want {
				t.Fatalf("GetStaticRequireCall(%s) = %v, want %v", test.expression, got, test.want)
			}
		})
	}
}

func TestGetRequireCallWithStringLiteralArgument(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		expression string
		want       bool
	}{
		{name: "string", expression: `require("pkg")`, want: true},
		{name: "parenthesized string", expression: `((require))((("pkg")))`, want: true},
		{name: "optional call remains caller policy", expression: `require?.("pkg")`, want: true},
		{name: "template literal", expression: "require(`pkg`)"},
		{name: "number", expression: `require(1)`},
		{name: "dynamic", expression: `require(name)`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			initializer := parseInitializer(t, test.expression)
			if got := GetRequireCallWithStringLiteralArgument(initializer) != nil; got != test.want {
				t.Fatalf("GetRequireCallWithStringLiteralArgument(%s) = %v, want %v", test.expression, got, test.want)
			}
		})
	}
}

func TestGetRequireCallWithLiteralArgument(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		expression string
		want       bool
	}{
		{name: "string", expression: `require("pkg")`, want: true},
		{name: "number", expression: `require(1)`, want: true},
		{name: "bigint", expression: `require(1n)`, want: true},
		{name: "boolean", expression: `require(true)`, want: true},
		{name: "null", expression: `require(null)`, want: true},
		{name: "regexp", expression: `require(/pkg/)`, want: true},
		{name: "parenthesized literal", expression: `require(((1)))`, want: true},
		{name: "template literal", expression: "require(`pkg`)"},
		{name: "dynamic argument", expression: `require(name)`},
		{name: "optional call", expression: `require?.(1)`},
		{name: "member callee", expression: `loader.require(1)`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			initializer := parseInitializer(t, test.expression)
			if got := GetRequireCallWithLiteralArgument(initializer) != nil; got != test.want {
				t.Fatalf("GetRequireCallWithLiteralArgument(%s) = %v, want %v", test.expression, got, test.want)
			}
		})
	}
}

func TestFindStaticRequireCallInChain(t *testing.T) {
	t.Parallel()
	if FindStaticRequireCallInChain(nil) != nil {
		t.Fatal("FindStaticRequireCallInChain(nil) returned a call")
	}

	tests := []struct {
		name       string
		expression string
		want       bool
	}{
		{name: "direct", expression: `require("pkg")`, want: true},
		{name: "property and call chain", expression: `require("pkg").factory()`, want: true},
		{name: "type arguments in chain", expression: `require<unknown>("pkg").factory()`, want: true},
		{name: "element and call chain", expression: `require("pkg")["factory"]()`, want: true},
		{name: "parenthesized chain", expression: `((require("pkg"))).factory`, want: true},
		{name: "require passed as argument", expression: `load(require("pkg"))`},
		{name: "optional require call", expression: `require?.("pkg").factory`},
		{name: "optional member", expression: `require("pkg")?.factory`},
		{name: "optional member call", expression: `require("pkg").factory?.()`},
		{name: "call after optional member", expression: `require("pkg")?.factory()`},
		{name: "typescript assertion is opaque", expression: `(require("pkg") as unknown).factory`},
		{name: "other require member", expression: `loader.require("pkg").factory`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			initializer := parseInitializer(t, test.expression)
			if got := FindStaticRequireCallInChain(initializer) != nil; got != test.want {
				t.Fatalf("FindStaticRequireCallInChain(%s) = %v, want %v", test.expression, got, test.want)
			}
		})
	}
}

func parseInitializer(t testing.TB, expression string) *ast.Node {
	t.Helper()
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.ts",
		Path:     "/test.ts",
	}, "const value = "+expression+";", core.ScriptKindTS)
	return sourceFile.Statements.Nodes[0].AsVariableStatement().
		DeclarationList.AsVariableDeclarationList().Declarations.Nodes[0].AsVariableDeclaration().Initializer
}
