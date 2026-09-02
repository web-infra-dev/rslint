package rule

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
)

// inlineExportedKeywords lists the directive keyword that introduces an
// `/* exported */` comment.
var inlineExportedKeywords = []string{"exported"}

// InlineExported describes one name declared by `/* exported */` comments.
// NameRanges contains one exact name range per comment that lists the name, in
// source order. Repeating a name within one comment still contributes only its
// first range, matching ESLint's comment metadata.
type InlineExported struct {
	Name       string
	NameRanges []core.TextRange
}

type inlineExportedName struct {
	name      string
	nameRange core.TextRange
}

// ParseInlineExported returns both the declared name set and ordered
// declaration metadata for `/* exported name1, name2 */` comments. A
// source-text candidate check keeps the shared comment store lazy unless such a
// directive may be present.
//
// Only real block-comment ranges supplied by the TypeScript scanner are read,
// so lookalike text in strings, templates, regexes, or line comments is ignored
// — ESLint accepts this directive in block comments only. The value is split on
// commas, and each entry is trimmed and unwrapped of one matching layer of
// single or double quotes, mirroring @eslint/plugin-kit's parseListConfig; a
// trailing ` -- justification ` is dropped before the list is read.
func ParseInlineExported(sourceFile *ast.SourceFile, comments *CommentStore) (map[string]bool, []InlineExported) {
	if sourceFile == nil || sourceFile.Text() == "" || !mayContainDirective(sourceFile.Text(), inlineExportedKeywords) {
		return nil, nil
	}

	text := sourceFile.Text()
	sourceComments := comments.All()
	if len(sourceComments) == 0 {
		return nil, nil
	}
	var exported []InlineExported
	var entryIndexes map[string]int

	for _, comment := range sourceComments {
		commentNames := parseInlineExportedComment(text, comment)
		if len(commentNames) == 0 {
			continue
		}

		// ESLint's parseListConfig returns an object, so a repeated name in one
		// comment has one comment entry, carrying its first range.
		commentSeen := make(map[string]bool, len(commentNames))
		if entryIndexes == nil {
			entryIndexes = make(map[string]int)
		}
		for _, entry := range commentNames {
			if commentSeen[entry.name] {
				continue
			}
			commentSeen[entry.name] = true

			if index, exists := entryIndexes[entry.name]; exists {
				exported[index].NameRanges = append(exported[index].NameRanges, entry.nameRange)
				continue
			}
			entryIndexes[entry.name] = len(exported)
			exported = append(exported, InlineExported{
				Name:       entry.name,
				NameRanges: []core.TextRange{entry.nameRange},
			})
		}
	}

	if len(exported) == 0 {
		return nil, nil
	}
	names := make(map[string]bool, len(exported))
	for _, entry := range exported {
		names[entry.Name] = true
	}

	return names, exported
}

func parseInlineExportedComment(text string, comment *ast.CommentRange) []inlineExportedName {
	if comment == nil || comment.Kind != ast.KindMultiLineCommentTrivia {
		return nil
	}

	start, end := comment.Pos(), comment.End()
	if start < 0 || end > len(text) || end-start < len("/*") || text[start:start+2] != "/*" {
		return nil
	}

	contentStart, contentEnd := start+2, end
	if contentEnd-contentStart >= 2 && text[contentEnd-2:contentEnd] == "*/" {
		contentEnd -= 2
	}
	contentStart, contentEnd = trimECMAScriptWhitespaceRange(text, contentStart, contentEnd)
	if contentStart == contentEnd {
		return nil
	}

	restStart, ok := matchDirectiveLabelRange(text, contentStart, contentEnd, inlineExportedKeywords)
	if !ok {
		return nil
	}
	if justificationStart := findDirectiveJustification(text, restStart, contentEnd); justificationStart >= 0 {
		contentEnd = justificationStart
	}
	return parseExportedNameListEntries(text, restStart, contentEnd)
}

// parseExportedNameListEntries splits the directive value on commas only —
// unlike `/* global */`, whitespace never separates two names, so
// `/* exported a b */` declares the single name "a b". Source positions stay
// attached so declaration ranges remain exact.
func parseExportedNameListEntries(text string, start int, end int) []inlineExportedName {
	var entries []inlineExportedName

	for start <= end {
		itemEnd := end
		if comma := strings.IndexByte(text[start:end], ','); comma >= 0 {
			itemEnd = start + comma
		}

		nameStart, nameEnd := trimECMAScriptWhitespaceRange(text, start, itemEnd)
		nameStart, nameEnd = stripMatchingQuoteRange(text, nameStart, nameEnd)
		if nameStart < nameEnd {
			entries = append(entries, inlineExportedName{
				name:      text[nameStart:nameEnd],
				nameRange: core.NewTextRange(nameStart, nameEnd),
			})
		}

		if itemEnd == end {
			break
		}
		start = itemEnd + 1
	}

	return entries
}

// stripMatchingQuoteRange removes one matching layer of single or double
// quotes, the unwrapping parseListConfig performs on every entry. An unbalanced
// or lone quote is part of the name.
func stripMatchingQuoteRange(text string, start int, end int) (int, int) {
	if end-start >= 2 {
		first, last := text[start], text[end-1]
		if (first == '\'' || first == '"') && first == last {
			return start + 1, end - 1
		}
	}
	return start, end
}
