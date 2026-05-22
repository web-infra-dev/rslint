// Package stylisticutil collects AST / source-text helpers shared by
// `@stylistic/eslint-plugin` rule ports. Each helper exists in this package
// because at least two rules in the plugin needed the same shape (extraction
// is mandated by the "duplicate-across-rules" rule in
// agents/port-rule/references/PORT_RULE.md — not an optional refactor).
package stylisticutil

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
)

// SameLineByPos reports whether two byte positions lie on the same source
// line. Uses the pre-computed ECMA line map (binary search via
// scanner.ComputeLineOfPosition) rather than scanning newlines ourselves —
// O(log n) instead of O(distance).
func SameLineByPos(sf *ast.SourceFile, a, b int) bool {
	lineStarts := sf.ECMALineMap()
	return scanner.ComputeLineOfPosition(lineStarts, a) ==
		scanner.ComputeLineOfPosition(lineStarts, b)
}

// CommentsExistBetween reports whether the byte range [low, high) contains a
// `//` or `/*` sequence. Mirrors ESLint's
// `sourceCode.commentsExistBetween(left, right)` predicate.
//
// **Safety invariant**: callers MUST pass a range that lies between two
// parser-recognized tokens — i.e. the byte range is pure trivia (whitespace
// and comments only, no string / regex / template literal content). Without
// this invariant a substring like `*/` inside a string would be falsely
// classified as a block-comment terminator. brace-style and arrow-parens
// both satisfy this invariant by definition: the trivia between e.g. `)` and
// `{`, or `(` and the param, cannot contain a string literal.
func CommentsExistBetween(text string, low, high int) bool {
	if low < 0 {
		low = 0
	}
	if high > len(text) {
		high = len(text)
	}
	if low >= high {
		return false
	}
	s := text[low:high]
	return strings.Contains(s, "//") || strings.Contains(s, "/*")
}

// FindPrevTokenEnd scans tokens forward from `searchStart` (a known earlier
// position guaranteed to lie before any token of interest) until it reaches
// the token starting at `targetPos`, and returns the end position of the
// previous token. Returns (-1, false) when no token starts at exactly
// `targetPos` in the forward stream (defensive — should not happen on
// well-formed input).
//
// Forward scanning is used instead of backward byte-walking because line
// comments aren't recognizable backward without re-scanning the line start
// (a backward `t` could be the last char of a token or of a `// ...` body),
// so a forward scan from a known earlier position is the only reliable way
// to locate a "previous real token" boundary.
//
// Cost is O(tokens between searchStart and targetPos). For nested
// constructs, callers typically pass `node.Parent.Pos()` as `searchStart`,
// which bounds the scan to the immediate enclosing AST node.
func FindPrevTokenEnd(sf *ast.SourceFile, searchStart, targetPos int) (int, bool) {
	if searchStart >= targetPos {
		return -1, false
	}
	s := scanner.GetScannerForSourceFile(sf, searchStart)
	prevEnd := -1
	for {
		if s.Token() == ast.KindEndOfFile {
			return -1, false
		}
		start := s.TokenStart()
		if start == targetPos {
			if prevEnd < 0 {
				return -1, false
			}
			return prevEnd, true
		}
		if start > targetPos {
			return -1, false
		}
		prevEnd = s.TokenEnd()
		s.Scan()
	}
}
