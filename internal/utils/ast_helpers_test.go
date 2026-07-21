package utils

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
)

func TestOutermostParenthesizedExpression(t *testing.T) {
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.ts",
		Path:     "/test.ts",
	}, "consume((((value)))); consume(plain);", core.ScriptKindTS)

	value := findNodeWithText(t, sourceFile, "value")
	outer := OutermostParenthesizedExpression(value)
	if got := TrimmedNodeText(sourceFile, outer); got != "(((value)))" {
		t.Fatalf("OutermostParenthesizedExpression(value) = %q, want %q", got, "(((value)))")
	}
	if outer.Parent == nil || !ast.IsCallExpression(outer.Parent) {
		t.Fatal("outermost parentheses should remain the direct call argument")
	}

	plain := findNodeWithText(t, sourceFile, "plain")
	if got := OutermostParenthesizedExpression(plain); got != plain {
		t.Fatal("unparenthesized node should be returned unchanged")
	}
	if got := OutermostParenthesizedExpression(nil); got != nil {
		t.Fatal("nil input should return nil")
	}
}
