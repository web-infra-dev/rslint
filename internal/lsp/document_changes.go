package lsp

import (
	"fmt"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/lsp/lsproto"
)

// applyDocumentChanges applies LSP content changes in order. Partial ranges are
// interpreted against the document produced by the preceding change, as
// required by the LSP protocol. The returned content is transactional: malformed
// input leaves the original document unchanged.
func applyDocumentChanges(
	content string,
	changes []lsproto.TextDocumentContentChangePartialOrWholeDocument,
) (string, error) {
	original := content
	for i, change := range changes {
		if (change.Partial == nil) == (change.WholeDocument == nil) {
			return original, fmt.Errorf("content change %d must contain exactly one change kind", i)
		}

		if whole := change.WholeDocument; whole != nil {
			content = whole.Text
			continue
		}

		partial := change.Partial
		lineStarts := computeLSPLineStarts(content)
		start := lspUTF16PositionToOffset(content, lineStarts, partial.Range.Start)
		end := lspUTF16PositionToOffset(content, lineStarts, partial.Range.End)
		if start > end {
			return original, fmt.Errorf("content change %d has reversed range", i)
		}

		var updated strings.Builder
		updated.Grow(len(content) - (end - start) + len(partial.Text))
		updated.WriteString(content[:start])
		updated.WriteString(partial.Text)
		updated.WriteString(content[end:])
		content = updated.String()
	}

	return content, nil
}

func lspUTF16PositionToOffset(content string, lineStarts []int, position lsproto.Position) int {
	line := uint64(position.Line)
	if line >= uint64(len(lineStarts)) {
		return len(content)
	}

	lineIndex := int(line)
	lineStart := lineStarts[lineIndex]
	lineEnd := len(content)
	if line+1 < uint64(len(lineStarts)) {
		lineEnd = lineStarts[lineIndex+1]
	}

	offset := lineStart
	character := uint64(position.Character)
	var utf16Offset uint64
	for offset < lineEnd {
		r, size := utf8.DecodeRuneInString(content[offset:lineEnd])
		units := uint64(utf16.RuneLen(r))
		if utf16Offset+units > character {
			break
		}
		utf16Offset += units
		offset += size
	}
	return offset
}

// computeLSPLineStarts deliberately recognizes only CR, LF, and CRLF. Unicode
// line/paragraph separators are ECMAScript line breaks, but the LSP protocol
// does not treat them as document lines.
func computeLSPLineStarts(content string) []int {
	lineStarts := make([]int, 1, strings.Count(content, "\n")+1)
	for offset := 0; offset < len(content); {
		switch content[offset] {
		case '\r':
			offset++
			if offset < len(content) && content[offset] == '\n' {
				offset++
			}
			lineStarts = append(lineStarts, offset)
		case '\n':
			offset++
			lineStarts = append(lineStarts, offset)
		default:
			offset++
		}
	}
	return lineStarts
}
