package utils

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

func TestGetStaticPropertyNameNumericLiteral(t *testing.T) {
	tests := []struct {
		name string
		code string
		want string
	}{
		{
			name: "hexadecimal above 2^53",
			code: "({ 0x1000000000000281: 0 })",
			want: "1152921504606847500",
		},
		{
			name: "binary above 2^53",
			code: "({ 0b1000000000000000000000000000000000000000000000000001010000001: 0 })",
			want: "1152921504606847500",
		},
		{
			name: "binary matching Acorn stepwise rounding",
			code: "({ 0b10100010000111101000011111100111101111100110100011110110000010100: 0 })",
			want: "23363847825694777000",
		},
		{
			name: "octal above 2^53",
			code: "({ 0o100000000000000001201: 0 })",
			want: "1152921504606847500",
		},
		{
			name: "uppercase prefix and numeric separators",
			code: "({ 0X1_0000_0000_0000_281: 0 })",
			want: "1152921504606847500",
		},
		{
			name: "parenthesized computed literal",
			code: "({ [(0x1000000000000281)]: 0 })",
			want: "1152921504606847500",
		},
		{
			name: "small radix literal",
			code: "({ 0xff: 0 })",
			want: "255",
		},
		{
			name: "decimal literal keeps normal normalization",
			code: "({ 1e21: 0 })",
			want: "1e+21",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
				FileName: "/test.ts",
				Path:     "/test.ts",
			}, tt.code, core.ScriptKindTS)
			property := findFirstNodeOfKind(t, sourceFile, ast.KindPropertyAssignment)
			nameNode := property.Name()
			if nameNode == nil {
				t.Fatal("property has no name")
			}
			got, ok := GetStaticPropertyName(nameNode)
			if !ok || got != tt.want {
				t.Fatalf("GetStaticPropertyName() = (%q, %v), want (%q, true)", got, ok, tt.want)
			}
		})
	}
}

func TestGetStaticPropertyNameDetachedNumericLiteral(t *testing.T) {
	factory := ast.NewNodeFactory(ast.NodeFactoryHooks{})
	node := factory.NewNumericLiteral("255", ast.TokenFlagsHexSpecifier)

	got, ok := GetStaticPropertyName(node)
	if !ok || got != "255" {
		t.Fatalf("GetStaticPropertyName() = (%q, %v), want (%q, true)", got, ok, "255")
	}
}

func TestRadixLiteralValueRejectsMalformedText(t *testing.T) {
	for _, raw := range []string{
		"",
		"123",
		"0q1",
		"0x",
		"0x_1",
		"0x1_",
		"0x1__0",
		"0b2",
		"0o8",
		"0x1n",
	} {
		t.Run(raw, func(t *testing.T) {
			if _, ok := radixLiteralValue(raw); ok {
				t.Fatalf("radixLiteralValue(%q) unexpectedly succeeded", raw)
			}
		})
	}
}

func TestRadixLiteralValueMatchesAcornRounding(t *testing.T) {
	value, ok := radixLiteralValue("0b10100010000111101000011111100111101111100110100011110110000010100")
	if !ok {
		t.Fatal("radixLiteralValue() unexpectedly failed")
	}
	if got, want := ecmascript.NumberToString(value), "23363847825694777000"; got != want {
		t.Fatalf("radixLiteralValue() = %q, want %q", got, want)
	}
}

func TestGetFunctionNameWithKindCore(t *testing.T) {
	tests := []struct {
		name string
		code string
		kind ast.Kind
		want string
	}{
		{
			name: "private method name is not quoted",
			code: "class C { #method() {} }",
			kind: ast.KindMethodDeclaration,
			want: "private method #method",
		},
		{
			name: "string key beginning with hash stays quoted",
			code: "class C { '#method'() {} }",
			kind: ast.KindMethodDeclaration,
			want: "method '#method'",
		},
		{
			name: "private getter name is not quoted",
			code: "class C { get #value() {} }",
			kind: ast.KindGetAccessor,
			want: "private getter #value",
		},
		{
			name: "parenthesized private class field arrow",
			code: "class C { static #field = ((() => {})) }",
			kind: ast.KindArrowFunction,
			want: "static private method #field",
		},
		{
			name: "parenthesized object property arrow",
			code: "const value = { field: ((() => {})) };",
			kind: ast.KindArrowFunction,
			want: "method 'field'",
		},
		{
			name: "parenthesized object property function",
			code: "const value = { field: ((function named() {})) };",
			kind: ast.KindFunctionExpression,
			want: "method 'field'",
		},
		{
			name: "parenthesized class field function",
			code: "class C { field = ((function named() {})) }",
			kind: ast.KindFunctionExpression,
			want: "method 'field'",
		},
		{
			name: "auto accessor arrow remains an arrow function",
			code: "class C { accessor value = () => {} }",
			kind: ast.KindArrowFunction,
			want: "arrow function",
		},
		{
			name: "parenthesized auto accessor function keeps its own name",
			code: "class C { accessor value = (function named() {}) }",
			kind: ast.KindFunctionExpression,
			want: "function 'named'",
		},
		{
			name: "auto accessor does not contribute static or private modifiers",
			code: "class C { static accessor #value = (async () => {}) }",
			kind: ast.KindArrowFunction,
			want: "async arrow function",
		},
		{
			name: "type assertion is not transparent",
			code: "class C { field = ((() => {}) as unknown) }",
			kind: ast.KindArrowFunction,
			want: "arrow function",
		},
		{
			name: "satisfies expression is not transparent",
			code: "class C { field = ((() => {}) satisfies unknown) }",
			kind: ast.KindArrowFunction,
			want: "arrow function",
		},
		{
			name: "comma expression is not transparent",
			code: "const value = { field: (0, (() => {})) };",
			kind: ast.KindArrowFunction,
			want: "arrow function",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
				FileName: "/test.ts",
				Path:     "/test.ts",
			}, tt.code, core.ScriptKindTS)
			node := findFirstNodeOfKind(t, sourceFile, tt.kind)
			if got := GetFunctionNameWithKindCore(node); got != tt.want {
				t.Fatalf("GetFunctionNameWithKindCore() = %q, want %q", got, tt.want)
			}
		})
	}
}
