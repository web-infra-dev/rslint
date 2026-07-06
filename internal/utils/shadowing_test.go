package utils

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
)

func TestIsNameShadowedBetweenEnumDeclaration(t *testing.T) {
	source := `enum value {
  A = value.present,
}
value.outside;
`
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.ts",
		Path:     "/test.ts",
	}, source, core.ScriptKindTS)

	shadowed := AccessExpressionObject(findNodeWithText(t, sourceFile, "value.present"))
	if shadowed == nil || !IsNameShadowedBetween(shadowed, sourceFile.AsNode(), "value") {
		t.Fatal("expected enum declaration to shadow the namespace before the source-file boundary")
	}

	outsideRef := AccessExpressionObject(findNodeWithText(t, sourceFile, "value.outside"))
	if outsideRef == nil || IsNameShadowedBetween(outsideRef, sourceFile.AsNode(), "value") {
		t.Fatal("expected reference outside enum scope not to be shadowed")
	}
}
