package utils

import (
	"sort"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

// SourceToken is a source-backed token with its kind, byte span, and text.
type SourceToken struct {
	Kind       ast.Kind
	Start, End int
	Text       string
}

var sourceTokenIndexKey = ast.NewSourceFileDataKey[[]SourceToken]()

// Range returns the token's source span.
func (t SourceToken) Range() core.TextRange {
	return core.NewTextRange(t.Start, t.End)
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

// TokenBeforePosition returns the last non-trivia parser token whose end is
// not after pos. Reading the parsed tree preserves lexical context across
// templates and recognizes regular-expression literals as single tokens.
func TokenBeforePosition(sourceFile *ast.SourceFile, pos int) (SourceToken, bool) {
	tokens := ast.GetOrComputeSourceFileData(sourceFile, sourceTokenIndexKey, func(sourceFile *ast.SourceFile) []SourceToken {
		return TokensOfNode(sourceFile, sourceFile.AsNode())
	})
	index := sort.Search(len(tokens), func(index int) bool {
		return tokens[index].End > pos
	})
	if index == 0 {
		return SourceToken{}, false
	}
	return tokens[index-1], true
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

// SafeReplacementText adds a single leading or trailing space when replacing
// node with replacement would otherwise merge with an adjacent token.
func SafeReplacementText(sourceFile *ast.SourceFile, node *ast.Node, replacement string) string {
	nodeRange := TrimNodeTextRange(sourceFile, node)
	output := replacement

	if before, ok := TokenBeforePosition(sourceFile, nodeRange.Pos()); ok &&
		before.End == nodeRange.Pos() &&
		!CanTokenTextsBeAdjacent(before.Text, output) {
		output = " " + output
	}

	if after, ok := TokenAtOrAfter(sourceFile, nodeRange.End()); ok &&
		after.Start == nodeRange.End() &&
		!CanTokenTextsBeAdjacent(output, after.Text) {
		output += " "
	}

	return output
}

// CanTokenTextsBeAdjacent is a small source-text analogue of ESLint's
// astUtils.canTokensBeAdjacent. It returns false for token pairs that would
// merge into a different token if printed without whitespace.
func CanTokenTextsBeAdjacent(left string, right string) bool {
	left = ecmascript.StringTrim(left)
	right = ecmascript.StringTrim(right)
	if left == "" || right == "" {
		return true
	}

	leftRune, _ := utf8.DecodeLastRuneInString(left)
	rightRune, _ := utf8.DecodeRuneInString(right)
	if leftRune == utf8.RuneError || rightRune == utf8.RuneError {
		return true
	}

	if scanner.IsIdentifierPart(leftRune) && scanner.IsIdentifierPart(rightRune) {
		return false
	}
	if scanner.IsIdentifierPart(leftRune) && startsWithEscapedIdentifier(right) {
		return false
	}
	// Keep a regexp delimiter distinct from a word-like token, and keep a
	// trailing decimal point from swallowing the following word operator.
	// These boundaries occur when fixes move regexp and `1.` literals next to
	// statement keywords or operators such as `in`, `as`, and `satisfies`.
	if (scanner.IsIdentifierPart(leftRune) && rightRune == '/') ||
		(leftRune == '/' && scanner.IsIdentifierPart(rightRune)) ||
		(leftRune == '.' && scanner.IsIdentifierPart(rightRune)) {
		return false
	}
	if (leftRune == '+' && rightRune == '+') || (leftRune == '-' && rightRune == '-') {
		return false
	}
	if leftRune == '/' && (rightRune == '/' || rightRune == '*') {
		return false
	}
	return true
}

func startsWithEscapedIdentifier(text string) bool {
	if text == "" || text[0] != '\\' {
		return false
	}
	s := scanner.NewScanner()
	s.SetText(text)
	return s.Scan() == ast.KindIdentifier && s.TokenStart() == 0
}

// IsSameLine reports whether two positions are on the same ECMAScript line.
func IsSameLine(sourceFile *ast.SourceFile, a int, b int) bool {
	lineMap := sourceFile.ECMALineMap()
	return scanner.ComputeLineOfPosition(lineMap, a) == scanner.ComputeLineOfPosition(lineMap, b)
}
