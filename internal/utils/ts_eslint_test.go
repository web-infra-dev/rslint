package utils

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
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

func TestGetStaticPropertyNameRegularExpressionLiteral(t *testing.T) {
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.ts",
		Path:     "/test.ts",
	}, `({ [/a/mi]: 0 })`, core.ScriptKindTS)
	property := findFirstNodeOfKind(t, sourceFile, ast.KindPropertyAssignment)
	nameNode := property.Name()
	if nameNode == nil {
		t.Fatal("property has no name")
	}
	got, ok := GetStaticPropertyName(nameNode)
	if !ok || got != "/a/im" {
		t.Fatalf("GetStaticPropertyName() = (%q, %v), want (%q, true)", got, ok, "/a/im")
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

func TestGetFunctionNameWithKindCoreJSDocCasts(t *testing.T) {
	tests := []struct {
		name string
		code string
		kind ast.Kind
		want string
	}{
		{
			name: "class field arrow",
			code: "class C { field = /** @type {Function} */ (() => {}) }",
			kind: ast.KindArrowFunction,
			want: "method 'field'",
		},
		{
			name: "static private async class field arrow",
			code: "class C { static #field = /** @type {Function} */ (async () => {}) }",
			kind: ast.KindArrowFunction,
			want: "static private async method #field",
		},
		{
			name: "class field async generator function prefers owner name",
			code: "class C { field = /** @type {Function} */ (async function* named() {}) }",
			kind: ast.KindFunctionExpression,
			want: "async generator method 'field'",
		},
		{
			name: "object property arrow",
			code: "const value = { field: /** @type {Function} */ (() => {}) };",
			kind: ast.KindArrowFunction,
			want: "method 'field'",
		},
		{
			name: "dynamic class field falls back to function name",
			code: "class C { [key] = /** @type {Function} */ (function named() {}) }",
			kind: ast.KindFunctionExpression,
			want: "method 'named'",
		},
		{
			name: "nested type and satisfies casts",
			code: "class C { field = /** @type {Function} */ (/** @satisfies {Function} */ (() => {})) }",
			kind: ast.KindArrowFunction,
			want: "method 'field'",
		},
		{
			name: "auto accessor remains a plain arrow function",
			code: "class C { static accessor #field = /** @type {Function} */ (async () => {}) }",
			kind: ast.KindArrowFunction,
			want: "async arrow function",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourceFile := parseJSFile(t, tt.code)
			node := findFirstNodeOfKind(t, sourceFile, tt.kind)
			if got := GetFunctionNameWithKindCore(node); got != tt.want {
				t.Fatalf("GetFunctionNameWithKindCore() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsGuaranteedSafeLiteral(t *testing.T) {
	t.Run("nil node", func(t *testing.T) {
		if IsGuaranteedSafeLiteral(nil) {
			t.Fatal("IsGuaranteedSafeLiteral(nil) = true, want false")
		}
	})

	t.Run("source file root has nil parent", func(t *testing.T) {
		sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
			FileName: "/test.ts",
			Path:     "/test.ts",
		}, `const x = 1;`, core.ScriptKindTS)
		if IsGuaranteedSafeLiteral(sourceFile.AsNode()) {
			t.Fatal("IsGuaranteedSafeLiteral(sourceFile) = true, want false")
		}
	})

	t.Run("detached numeric literal has nil parent", func(t *testing.T) {
		factory := ast.NewNodeFactory(ast.NodeFactoryHooks{})
		node := factory.NewNumericLiteral("1", ast.TokenFlagsNone)
		if IsGuaranteedSafeLiteral(node) {
			t.Fatal("IsGuaranteedSafeLiteral(detached numeric literal) = true, want false")
		}
	})

	tests := []struct {
		name string
		tsx  bool
		code string
		kind ast.Kind
		want bool
	}{
		{
			name: "property assignment string key is not safe expression literal",
			code: `const o = { "a": 1 };`,
			kind: ast.KindStringLiteral,
			want: false,
		},
		{
			name: "import specifier string property name is not safe expression literal",
			code: `import { "a" as b } from "./mod";`,
			kind: ast.KindStringLiteral,
			want: false,
		},
		{
			name: "string literal inside literal type is not safe expression literal",
			code: `type T = "a";`,
			kind: ast.KindStringLiteral,
			want: false,
		},
		{
			name: "numeric literal inside literal type is not safe expression literal",
			code: `type N = 1;`,
			kind: ast.KindNumericLiteral,
			want: false,
		},
		{
			name: "import module specifier string literal is not safe expression literal",
			code: `import "./mod";`,
			kind: ast.KindStringLiteral,
			want: false,
		},
		{
			name: "true keyword inside literal type is not safe expression literal",
			code: `type T = true;`,
			kind: ast.KindTrueKeyword,
			want: false,
		},
		{
			name: "false keyword inside literal type is not safe expression literal",
			code: `type F = false;`,
			kind: ast.KindFalseKeyword,
			want: false,
		},
		{
			name: "null keyword inside literal type is not safe expression literal",
			code: `type N = null;`,
			kind: ast.KindNullKeyword,
			want: false,
		},
		{
			name: "jsx attribute string literal is not safe expression literal",
			tsx:  true,
			code: `const el = <Comp a="x" />;`,
			kind: ast.KindStringLiteral,
			want: false,
		},
		{
			name: "numeric literal inside with statement is not safe expression literal",
			code: `with (obj) { const x = 1; }`,
			kind: ast.KindNumericLiteral,
			want: false,
		},
		{
			name: "tagged template literal is not safe primitive literal",
			code: "const x = tag`hello`;",
			kind: ast.KindNoSubstitutionTemplateLiteral,
			want: false,
		},
		{
			name: "untagged no-substitution template literal is safe",
			code: "const x = `hello`;",
			kind: ast.KindNoSubstitutionTemplateLiteral,
			want: true,
		},
		{
			name: "expression numeric literal is safe",
			code: `const x = 1;`,
			kind: ast.KindNumericLiteral,
			want: true,
		},
		{
			name: "expression string literal is safe",
			code: `const x = "a";`,
			kind: ast.KindStringLiteral,
			want: true,
		},
		{
			name: "expression true keyword is safe",
			code: `const x = true;`,
			kind: ast.KindTrueKeyword,
			want: true,
		},
		{
			name: "expression false keyword is safe",
			code: `const x = false;`,
			kind: ast.KindFalseKeyword,
			want: true,
		},
		{
			name: "bigint literal inside literal type is not safe expression literal",
			code: `type B = 1n;`,
			kind: ast.KindBigIntLiteral,
			want: false,
		},
		{
			name: "expression bigint literal is safe",
			code: `const x = 1n;`,
			kind: ast.KindBigIntLiteral,
			want: true,
		},
		{
			name: "expression null keyword is safe",
			code: `const x = null;`,
			kind: ast.KindNullKeyword,
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := ast.SourceFileParseOptions{
				FileName: "/test.ts",
				Path:     "/test.ts",
			}
			scriptKind := core.ScriptKindTS
			if tt.tsx {
				opts.FileName = "/test.tsx"
				opts.Path = "/test.tsx"
				scriptKind = core.ScriptKindTSX
			}
			sourceFile := parser.ParseSourceFile(opts, tt.code, scriptKind)
			node := findFirstNodeOfKind(t, sourceFile, tt.kind)
			if got := IsGuaranteedSafeLiteral(node); got != tt.want {
				t.Fatalf("IsGuaranteedSafeLiteral() = %v, want %v", got, tt.want)
			}
		})
	}
}
