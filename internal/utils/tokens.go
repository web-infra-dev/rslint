package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
)

// SourceToken is a source-backed token with its kind, byte span, and text.
type SourceToken struct {
	Kind       ast.Kind
	Start, End int
	Text       string
}

// TokensOfNode returns all parser tokens contained in node, in source order.
func TokensOfNode(sourceFile *ast.SourceFile, node *ast.Node) []SourceToken {
	tokens := []SourceToken{}
	sourceText := sourceFile.Text()
	ForEachToken(node, func(token *ast.Node) {
		trimmed := TrimNodeTextRange(sourceFile, token)
		if trimmed.Pos() >= trimmed.End() {
			return
		}
		tokens = append(tokens, SourceToken{
			Kind:  token.Kind,
			Start: trimmed.Pos(),
			End:   trimmed.End(),
			Text:  sourceText[trimmed.Pos():trimmed.End()],
		})
	}, sourceFile)
	return tokens
}

// TokenAtOrAfter returns the first non-trivia token at or after pos.
func TokenAtOrAfter(sourceFile *ast.SourceFile, pos int) (SourceToken, bool) {
	sourceText := sourceFile.Text()
	if pos < 0 {
		pos = 0
	}
	if pos >= len(sourceText) {
		return SourceToken{}, false
	}

	start := scanner.SkipTrivia(sourceText, pos)
	if start >= len(sourceText) {
		return SourceToken{}, false
	}

	tokenRange := scanner.GetRangeOfTokenAtPosition(sourceFile, start)
	if tokenRange.Pos() >= tokenRange.End() || tokenRange.End() > len(sourceText) {
		return SourceToken{}, false
	}
	return SourceToken{
		Kind:  scanner.ScanTokenAtPosition(sourceFile, start),
		Start: tokenRange.Pos(),
		End:   tokenRange.End(),
		Text:  sourceText[tokenRange.Pos():tokenRange.End()],
	}, true
}

// PreviousTokenBefore returns the last token in node whose end is not after pos.
func PreviousTokenBefore(sourceFile *ast.SourceFile, node *ast.Node, pos int) (SourceToken, bool) {
	var previous SourceToken
	found := false
	for _, token := range TokensOfNode(sourceFile, node) {
		if token.Start >= pos {
			break
		}
		if token.End <= pos {
			previous = token
			found = true
		}
	}
	return previous, found
}

// IsSameLine reports whether two positions are on the same ECMAScript line.
func IsSameLine(sourceFile *ast.SourceFile, a int, b int) bool {
	lineMap := sourceFile.ECMALineMap()
	return scanner.ComputeLineOfPosition(lineMap, a) == scanner.ComputeLineOfPosition(lineMap, b)
}
