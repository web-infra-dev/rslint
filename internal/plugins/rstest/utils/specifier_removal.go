package utils

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
)

// SpecifierRemovalRange spans the specifier at index together with the comma
// that separates it from its neighbour. Comments stay with the specifier on
// their side of the separating comma.
func SpecifierRemovalRange(sourceFile *ast.SourceFile, elements []*ast.Node, index int) core.TextRange {
	elementRange := internalUtils.TrimNodeTextRange(sourceFile, elements[index])
	if index > 0 {
		text := sourceFile.Text()
		previousEnd := internalUtils.TrimNodeTextRange(sourceFile, elements[index-1]).End()
		start := previousEnd
		if comma := separatorComma(text, previousEnd, elementRange.Pos()); comma >= 0 {
			start = comma
		}
		limit := specifierListEnd(sourceFile, elements, index)
		return core.NewTextRange(start, commentsEnd(text, elementRange.End(), limit))
	}
	nextStart := internalUtils.TrimNodeTextRange(sourceFile, elements[index+1]).Pos()
	return core.NewTextRange(elementRange.Pos(), firstSpecifierRemovalEnd(sourceFile.Text(), elementRange.End(), nextStart))
}

func specifierListEnd(sourceFile *ast.SourceFile, elements []*ast.Node, index int) int {
	if index+1 < len(elements) {
		return internalUtils.TrimNodeTextRange(sourceFile, elements[index+1]).Pos()
	}
	if parent := elements[index].Parent; parent != nil {
		return internalUtils.TrimNodeTextRange(sourceFile, parent).End()
	}
	return len(sourceFile.Text())
}

func commentsEnd(text string, start int, limit int) int {
	if start < 0 || limit > len(text) || start >= limit {
		return start
	}
	end := start
	for position := start; position < limit; {
		rest := text[position:limit]
		switch {
		case rest[0] == ' ' || rest[0] == '\t' || rest[0] == '\r' || rest[0] == '\n':
			position++
		case strings.HasPrefix(rest, "//"):
			length := strings.IndexByte(rest, '\n')
			if length < 0 {
				length = len(rest)
			}
			position += length
			end = position
		case strings.HasPrefix(rest, "/*"):
			length := strings.Index(rest, "*/")
			if length < 0 {
				return end
			}
			position += length + 2
			end = position
		default:
			return end
		}
	}
	return end
}

func firstSpecifierRemovalEnd(text string, start int, end int) int {
	if start < 0 || end > len(text) || start >= end {
		return end
	}
	comma := separatorComma(text, start, end)
	if comma < 0 {
		return end
	}
	// Comments past the separator belong to the surviving specifier, so the
	// removal stops at the first of them. Without one the whole gap goes.
	if comment := firstCommentOffset(text[comma+1 : end]); comment >= 0 {
		return comma + 1 + comment
	}
	return end
}

// separatorComma returns the index of the comma separating two specifiers, or
// -1 when the range holds none. Comment trivia is skipped, so a comma written
// inside a comment is never mistaken for the separator.
func separatorComma(text string, start int, end int) int {
	if start < 0 || end > len(text) || start >= end {
		return -1
	}
	for position := start; position < end; {
		rest := text[position:end]
		switch {
		case rest[0] == ',':
			return position
		case strings.HasPrefix(rest, "//"):
			length := strings.IndexByte(rest, '\n')
			if length < 0 {
				return -1
			}
			position += length + 1
		case strings.HasPrefix(rest, "/*"):
			length := strings.Index(rest, "*/")
			if length < 0 {
				return -1
			}
			position += length + 2
		default:
			position++
		}
	}
	return -1
}

func firstCommentOffset(text string) int {
	offset := -1
	if line := strings.Index(text, "//"); line >= 0 {
		offset = line
	}
	if block := strings.Index(text, "/*"); block >= 0 && (offset < 0 || block < offset) {
		offset = block
	}
	return offset
}
