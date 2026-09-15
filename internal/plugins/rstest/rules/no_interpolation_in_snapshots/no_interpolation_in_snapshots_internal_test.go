package no_interpolation_in_snapshots

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
)

func TestSourceMayContainInterpolatedSnapshot(t *testing.T) {
	for _, testCase := range []struct {
		source string
		want   bool
	}{
		{source: `expect(value).toMatchInlineSnapshot("plain")`, want: false},
		{source: "expect(value).toMatchInlineSnapshot(`${value}`)", want: true},
		{source: `const text = "${notCode}"`, want: true},
	} {
		sourceFile := parser.ParseSourceFile(
			ast.SourceFileParseOptions{
				FileName: "/source.test.ts",
				Path:     "/source.test.ts",
			},
			testCase.source,
			core.ScriptKindTS,
		)
		if got := sourceMayContainInterpolatedSnapshot(sourceFile); got != testCase.want {
			t.Errorf(
				"sourceMayContainInterpolatedSnapshot(%q) = %t, want %t",
				testCase.source,
				got,
				testCase.want,
			)
		}
	}
}

func TestMayContainInterpolatedInlineSnapshot(t *testing.T) {
	for _, testCase := range []struct {
		source string
		want   bool
	}{
		{
			source: "expect(value).toMatchInlineSnapshot(`${value}`)",
			want:   true,
		},
		{
			source: "expect(value)[\"toMatchInlineSnapshot\"](`${value}`)",
			want:   true,
		},
		{
			source: "expect(value).to.be.a(\"string\").and.toMatchInlineSnapshot(`${value}`)",
			want:   true,
		},
		{
			source: "expect(value).toMatchInlineSnapshot((`${value}`))",
			want:   true,
		},
		{
			source: "expect(value).toThrowErrorMatchingInlineSnapshot(`${value}`)",
			want:   true,
		},
		{
			source: "expect(value).toMatchInlineSnapshot(\"plain\", `case ${id}`)",
			want:   true,
		},
		{
			source: "expect(value).toMatchSnapshot(`${value}`)",
			want:   false,
		},
		{
			source: "custom.toMatchInlineSnapshot(`${value}`)",
			want:   true,
		},
		{
			source: "expect(value).toBe(`${value}`)",
			want:   false,
		},
	} {
		sourceFile := parser.ParseSourceFile(
			ast.SourceFileParseOptions{
				FileName: "/source.test.ts",
				Path:     "/source.test.ts",
			},
			testCase.source,
			core.ScriptKindTS,
		)
		var outerCall *ast.Node
		var visit func(*ast.Node) bool
		visit = func(node *ast.Node) bool {
			if node.Kind == ast.KindCallExpression &&
				rstestUtils.FindTopMostCallExpression(node) == node {
				outerCall = node
			}
			return node.ForEachChild(visit)
		}
		sourceFile.Node.ForEachChild(visit)
		if outerCall == nil {
			t.Fatalf("call not found for %q", testCase.source)
		}
		if got := mayContainInterpolatedInlineSnapshot(outerCall); got != testCase.want {
			t.Errorf(
				"mayContainInterpolatedInlineSnapshot(%q) = %t, want %t",
				testCase.source,
				got,
				testCase.want,
			)
		}
	}
}
