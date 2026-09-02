package utils

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
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
		if comma := strings.LastIndexByte(text[previousEnd:elementRange.Pos()], ','); comma >= 0 {
			start = previousEnd + comma
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
	between := text[start:end]
	comma := strings.IndexByte(between, ',')
	if comma < 0 {
		return end
	}
	comment := firstCommentOffset(between)
	if comment < 0 {
		return end
	}
	if comment > comma {
		return start + comment
	}
	if afterComma := firstCommentOffset(between[comma+1:]); afterComma >= 0 {
		return start + comma + 1 + afterComma
	}
	return end
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
