package utils

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
)

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
