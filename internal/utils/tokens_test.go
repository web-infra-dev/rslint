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
		want        bool
	}{
		{left: "return", right: "/a/", want: false},
		{left: "/a/", right: "instanceof", want: false},
		{left: "return", right: "/a/ > value", want: false},
		{left: "value > /a/", right: "instanceof", want: false},
		{left: "/", right: "0b111", want: true},
		{left: "foo/", right: "bar", want: true},
		{left: "foo", right: "/bar", want: true},
		{left: "1.", right: "satisfies", want: false},
		{left: "return", right: `"a"`, want: true},
		{left: `"a"`, right: "instanceof", want: true},
		{left: "return", right: "`a`", want: true},
	} {
		if got := CanTokenTextsBeAdjacent(tt.left, tt.right); got != tt.want {
			t.Errorf("CanTokenTextsBeAdjacent(%q, %q) = %v, want %v", tt.left, tt.right, got, tt.want)
		}
	}
}
