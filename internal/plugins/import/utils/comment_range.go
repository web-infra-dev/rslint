package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
)

const maxAdjacentComments = 100

// LineRangeWithComments returns the source range import/order moves when it
// swaps two whole import statements. It mirrors import-js's
// findStartOfLineWithComments/findEndOfLineWithComments: only single-line
// comments on the statement's ending line are attached to the statement.
func LineRangeWithComments(text string, node *ast.Node, lineStarts []core.TextPos, comments []*ast.CommentRange) (int, int) {
	return StartOfLineWithComments(text, node, lineStarts, comments),
		EndOfLineWithComments(text, node, lineStarts, comments)
}

// StartOfLineWithComments returns the beginning of the movable statement
// range, including adjacent same-line comments before it and indentation.
func StartOfLineWithComments(text string, node *ast.Node, lineStarts []core.TextPos, comments []*ast.CommentRange) int {
	if node == nil {
		return -1
	}
	start := scanner.SkipTrivia(text, node.Pos())
	endingLine := scanner.ComputeLineOfPosition(lineStarts, node.End())
	seen := 0
	for i := len(comments) - 1; i >= 0 && seen < maxAdjacentComments; i-- {
		comment := comments[i]
		if comment.End() > start {
			continue
		}
		seen++
		if scanner.ComputeLineOfPosition(lineStarts, comment.Pos()) != endingLine ||
			scanner.ComputeLineOfPosition(lineStarts, comment.End()) != endingLine ||
			!onlyHorizontalWhitespace(text, comment.End(), start) {
			break
		}
		start = comment.Pos()
	}
	for i := start - 1; i > 0; i-- {
		if text[i] != ' ' && text[i] != '\t' {
			break
		}
		start = i
	}
	return start
}

// EndOfLineWithComments returns the end of the movable statement range. The
// terminating newline is included when present.
func EndOfLineWithComments(text string, node *ast.Node, lineStarts []core.TextPos, comments []*ast.CommentRange) int {
	if node == nil {
		return -1
	}
	endingLine := scanner.ComputeLineOfPosition(lineStarts, node.End())
	end := node.End()
	seen := 0
	for _, comment := range comments {
		if comment.Pos() < end {
			continue
		}
		if seen == maxAdjacentComments {
			break
		}
		seen++
		if scanner.ComputeLineOfPosition(lineStarts, comment.Pos()) != endingLine ||
			scanner.ComputeLineOfPosition(lineStarts, comment.End()) != endingLine ||
			!onlyHorizontalWhitespace(text, end, comment.Pos()) {
			break
		}
		end = comment.End()
	}
	for end < len(text) {
		switch text[end] {
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

// SameLineEndWithComments returns the point immediately after the last
// adjacent, single-line trailing comment. Newline fixes insert at this point.
func SameLineEndWithComments(text string, node *ast.Node, lineStarts []core.TextPos, comments []*ast.CommentRange) int {
	if node == nil {
		return -1
	}
	endingLine := scanner.ComputeLineOfPosition(lineStarts, node.End())
	end := node.End()
	seen := 0
	for _, comment := range comments {
		if comment.Pos() < end {
			continue
		}
		if seen == maxAdjacentComments {
			break
		}
		seen++
		if scanner.ComputeLineOfPosition(lineStarts, comment.Pos()) != endingLine ||
			scanner.ComputeLineOfPosition(lineStarts, comment.End()) != endingLine ||
			!onlyHorizontalWhitespace(text, end, comment.Pos()) {
			break
		}
		end = comment.End()
	}
	return end
}

func onlyHorizontalWhitespace(text string, start int, end int) bool {
	if start < 0 || start > end || end > len(text) {
		return false
	}
	for _, char := range text[start:end] {
		if char != ' ' && char != '\t' && char != '\r' {
			return false
		}
	}
	return true
}
