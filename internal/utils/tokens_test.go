package utils

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
)

func TestSourceTokenRange(t *testing.T) {
	token := SourceToken{Kind: ast.KindIdentifier, Start: 3, End: 8, Text: "value"}
	r := token.Range()
	if r.Pos() != token.Start || r.End() != token.End {
		t.Fatalf("Range() = [%d,%d), want [%d,%d)", r.Pos(), r.End(), token.Start, token.End)
	}
}
