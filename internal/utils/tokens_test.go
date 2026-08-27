package utils

import (
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
)

func TestSourceTokenRange(t *testing.T) {
	token := SourceToken{Kind: ast.KindIdentifier, Start: 3, End: 8, Text: "value"}
	r := token.Range()
	if r.Pos() != token.Start || r.End() != token.End {
		t.Fatalf("Range() = [%d,%d), want [%d,%d)", r.Pos(), r.End(), token.Start, token.End)
	}
}

func TestTokenBeforePositionRecoversRegularExpressionLiteral(t *testing.T) {
	code := "/before/giu\nnext()"
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.ts",
		Path:     "/test.ts",
	}, code, core.ScriptKindTS)

	token, ok := TokenBeforePosition(sourceFile, strings.Index(code, "next"))
	if !ok {
		t.Fatal("TokenBeforePosition() did not find the regular expression")
	}
	if token.Kind != ast.KindRegularExpressionLiteral ||
		token.Text != "/before/giu" ||
		token.Start != 0 ||
		token.End != len("/before/giu") {
		t.Fatalf("TokenBeforePosition() = %#v, want the complete regular-expression token", token)
	}
}

func TestTokenBeforePositionDoesNotTreatDivisionAsRegularExpression(t *testing.T) {
	code := "left / right\nnext()"
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.ts",
		Path:     "/test.ts",
	}, code, core.ScriptKindTS)

	token, ok := TokenBeforePosition(sourceFile, strings.Index(code, "next"))
	if !ok {
		t.Fatal("TokenBeforePosition() did not find the division operand")
	}
	if token.Kind != ast.KindIdentifier || token.Text != "right" {
		t.Fatalf("TokenBeforePosition() = %#v, want the identifier after division", token)
	}
}

func TestTokenBeforePositionAfterInterpolatedTemplate(t *testing.T) {
	code := "`${seed}`; function *f(){yield(1)<a}"
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.ts",
		Path:     "/test.ts",
	}, code, core.ScriptKindTS)

	token, ok := TokenBeforePosition(sourceFile, strings.Index(code, "(1)"))
	if !ok || token.Kind != ast.KindYieldKeyword || token.Text != "yield" {
		t.Fatalf("TokenBeforePosition() = %#v, %v, want the yield keyword", token, ok)
	}
}

func TestCanTokenTextsBeAdjacentFixBoundaries(t *testing.T) {
	for _, tt := range []struct {
		left, right string
	}{
		{left: "return", right: "/a/"},
		{left: "/a/", right: "instanceof"},
		{left: "1.", right: "satisfies"},
	} {
		if CanTokenTextsBeAdjacent(tt.left, tt.right) {
			t.Errorf("CanTokenTextsBeAdjacent(%q, %q) = true, want false", tt.left, tt.right)
		}
	}
}
