package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
)

// LineRangeWithComments returns the [start, end) source-text range that covers
// `node` plus any comments on the same physical line(s).  Mirrors
// upstream's `findStartOfLineWithComments` + `findEndOfLineWithComments` from
// `eslint-plugin-import/src/rules/order.js`.
//
// Specifically:
//   - `start` is the column-0 position of the first line that contains the
//     node, INCLUDING any `//` or `/* */` comments that appear on that line
//     before the node (token-level scan, so `import /*...*/ a from 'a'` keeps
//     the leading comment with the import).
//   - `end` is the position right after the trailing `\n` (or end-of-file)
//     of the last line that contains the node OR a same-line trailing
//     comment.  `import a from 'a'; // trailing` keeps the trailing comment.
//
// CRLF endings (`\r\n`) are handled — the returned `end` skips both bytes.
func LineRangeWithComments(text string, node *ast.Node, lineStarts []core.TextPos, factory *ast.NodeFactory) (int, int) {
	return startOfLineWithComments(text, node, lineStarts, factory),
		endOfLineWithComments(text, node, lineStarts, factory)
}

// startOfLineWithComments walks back from node.Pos() to the column-0 of the
// first line that contains the node, pulling in any `//` or `/* */` comments
// that appear on the same line *before* the node.
func startOfLineWithComments(text string, node *ast.Node, lineStarts []core.TextPos, factory *ast.NodeFactory) int {
	// Skip the node's own leading trivia to find the first significant token.
	tokenStart := scanner.SkipTrivia(text, node.Pos())
	startLine := scanner.ComputeLineOfPosition(lineStarts, tokenStart)
	startOfStartLine := int(lineStarts[startLine])

	// Scan leading comment ranges *before* the node. Any whose end-line equals
	// startLine and whose own start sits at or after startOfStartLine is on
	// the same line *before* the node — pull it into the range.
	earliest := tokenStart
	for cr := range scanner.GetLeadingCommentRanges(factory, text, node.Pos()) {
		crEndLine := scanner.ComputeLineOfPosition(lineStarts, cr.End())
		if crEndLine == startLine && cr.Pos() >= startOfStartLine && cr.Pos() < earliest {
			earliest = cr.Pos()
		}
	}
	if earliest <= startOfStartLine {
		return startOfStartLine
	}
	// Walk back across spaces/tabs from earliest to column 0 of that line.
	i := earliest - 1
	for i >= startOfStartLine && (text[i] == ' ' || text[i] == '\t') {
		i--
	}
	if i < startOfStartLine {
		return startOfStartLine
	}
	return earliest
}

// endOfLineWithComments returns the position right after the trailing newline
// of the last line containing the node or a same-line trailing comment.
func endOfLineWithComments(text string, node *ast.Node, lineStarts []core.TextPos, factory *ast.NodeFactory) int {
	nodeEndLine := scanner.ComputeLineOfPosition(lineStarts, node.End())
	end := node.End()

	// Pull in any trailing comments that start on the same line as the node's
	// end. Trailing comments may extend onto subsequent lines (block
	// comments) — we still include them as a single contiguous unit.
	for cr := range scanner.GetTrailingCommentRanges(factory, text, node.End()) {
		crStartLine := scanner.ComputeLineOfPosition(lineStarts, cr.Pos())
		if crStartLine == nodeEndLine {
			if cr.End() > end {
				end = cr.End()
				// Update nodeEndLine to track the comment's end line for
				// chained `// a // b`-style cases.
				nodeEndLine = scanner.ComputeLineOfPosition(lineStarts, cr.End())
			}
		} else {
			break
		}
	}

	// Eat trailing whitespace and the line-terminator. Handles `\r\n`.
	for end < len(text) {
		c := text[end]
		switch c {
		case ' ', '\t', '\r':
			end++
		case '\n':
			return end + 1
		default:
			return end
		}
	}
	return end
}
