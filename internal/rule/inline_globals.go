package rule

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
)

// inlineGlobalsKeywords lists the directive keywords that introduce a
// `/* global */` comment. "globals" is checked before "global" since it is
// the longer prefix.
var inlineGlobalsKeywords = [...]string{"globals", "global"}

// InlineGlobal describes one name declared by `/* global */` comments.
// Declared is the name's final inline state after all comments are applied.
// NameRanges contains one exact name range per comment that mentions the name,
// in source order. Repeating a name within one comment still contributes only
// its first range, matching ESLint's comment metadata.
type InlineGlobal struct {
	Name       string
	Declared   bool
	NameRanges []core.TextRange
}

type inlineGlobalName struct {
	name      string
	setting   string
	nameRange core.TextRange
}

// ParseInlineGlobals returns both the final name -> declared map and ordered
// declaration metadata for `/* global ... */` / `/* globals ... */` comments.
// A source-text candidate check keeps the shared comment store lazy unless such a
// directive may be present.
//
// Only real block-comment ranges supplied by the TypeScript scanner are read,
// so lookalike text in strings, templates, regexes, or line comments is ignored.
// Within a comment, duplicate names use the last setting and retain the first
// name range. Across comments, the last setting wins and every comment range is
// preserved. As in the existing globals API, only "off" un-declares a name.
func ParseInlineGlobals(sourceFile *ast.SourceFile, comments *CommentStore) (map[string]bool, []InlineGlobal) {
	if sourceFile == nil || sourceFile.Text() == "" || !mayContainInlineGlobalDirective(sourceFile.Text()) {
		return nil, nil
	}

	text := sourceFile.Text()
	sourceComments := comments.All()
	if len(sourceComments) == 0 {
		return nil, nil
	}
	var values map[string]bool
	var globals []InlineGlobal
	var globalIndexes map[string]int

	for _, comment := range sourceComments {
		entries := parseInlineGlobalComment(text, comment)
		if len(entries) == 0 {
			continue
		}

		// ESLint's parseStringConfig returns an object, so a repeated name in
		// one comment has one comment entry: its last setting and first range.
		commentEntries := make([]inlineGlobalName, 0, len(entries))
		commentIndexes := make(map[string]int, len(entries))
		for _, entry := range entries {
			if index, exists := commentIndexes[entry.name]; exists {
				commentEntries[index].setting = entry.setting
				continue
			}
			commentIndexes[entry.name] = len(commentEntries)
			commentEntries = append(commentEntries, entry)
		}

		if values == nil {
			values = make(map[string]bool)
			globalIndexes = make(map[string]int)
		}
		for _, entry := range commentEntries {
			declared := entry.setting != "off"
			values[entry.name] = declared

			if index, exists := globalIndexes[entry.name]; exists {
				globals[index].Declared = declared
				globals[index].NameRanges = append(globals[index].NameRanges, entry.nameRange)
				continue
			}
			globalIndexes[entry.name] = len(globals)
			globals = append(globals, InlineGlobal{
				Name:       entry.name,
				Declared:   declared,
				NameRanges: []core.TextRange{entry.nameRange},
			})
		}
	}

	return values, globals
}

func mayContainInlineGlobalDirective(text string) bool {
	for searchStart := 0; searchStart < len(text); {
		markerOffset := strings.Index(text[searchStart:], "/*")
		if markerOffset < 0 {
			return false
		}

		contentStart := searchStart + markerOffset + len("/*")
		contentStart, _ = trimECMAScriptWhitespaceRange(text, contentStart, len(text))
		for _, keyword := range inlineGlobalsKeywords {
			if !strings.HasPrefix(text[contentStart:], keyword) {
				continue
			}
			restStart := contentStart + len(keyword)
			if restStart == len(text) || strings.HasPrefix(text[restStart:], "*/") {
				return true
			}
			r, _ := utf8.DecodeRuneInString(text[restStart:])
			if isECMAScriptWhitespace(r) {
				return true
			}
		}

		searchStart = contentStart
	}
	return false
}

func parseInlineGlobalComment(text string, comment *ast.CommentRange) []inlineGlobalName {
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

	restStart, ok := matchInlineGlobalsDirectiveRange(text, contentStart, contentEnd)
	if !ok {
		return nil
	}
	if justificationStart := findDirectiveJustification(text, restStart, contentEnd); justificationStart >= 0 {
		contentEnd = justificationStart
	}
	restStart, contentEnd = trimECMAScriptWhitespaceRange(text, restStart, contentEnd)
	return parseGlobalNameListEntries(text, restStart, contentEnd)
}

// matchInlineGlobalsDirective reports whether comment content begins with the
// exact lower-case "global"/"globals" directive label followed by ECMAScript
// whitespace or end-of-string.
func matchInlineGlobalsDirective(content string) (string, bool) {
	start, end := trimECMAScriptWhitespaceRange(content, 0, len(content))
	restStart, ok := matchInlineGlobalsDirectiveRange(content, start, end)
	if !ok {
		return "", false
	}
	restStart, end = trimECMAScriptWhitespaceRange(content, restStart, end)
	return content[restStart:end], true
}

func matchInlineGlobalsDirectiveRange(text string, start int, end int) (int, bool) {
	for _, keyword := range inlineGlobalsKeywords {
		if !strings.HasPrefix(text[start:end], keyword) {
			continue
		}
		restStart := start + len(keyword)
		if restStart == end {
			return restStart, true
		}
		r, _ := utf8.DecodeRuneInString(text[restStart:end])
		if isECMAScriptWhitespace(r) {
			return restStart, true
		}
	}
	return 0, false
}

// MergeGlobals combines config-declared globals with inline `/* global */`
// comment globals into the single set exposed to rules as ctx.Globals. Inline
// settings win on conflict. Returns nil if both inputs are empty.
func MergeGlobals(configGlobals, inlineGlobals map[string]bool) map[string]bool {
	if len(configGlobals) == 0 {
		return inlineGlobals
	}
	if len(inlineGlobals) == 0 {
		return configGlobals
	}
	merged := make(map[string]bool, len(configGlobals)+len(inlineGlobals))
	for name, declared := range configGlobals {
		merged[name] = declared
	}
	for name, declared := range inlineGlobals {
		merged[name] = declared
	}
	return merged
}

// parseGlobalNameList parses ESLint's comma-and/or-whitespace separated
// "name[:setting]" syntax. It is kept as a map helper for focused parser tests.
func parseGlobalNameList(s string) map[string]string {
	names := make(map[string]string)
	for _, entry := range parseGlobalNameListEntries(s, 0, len(s)) {
		names[entry.name] = entry.setting
	}
	return names
}

type globalConfigRune struct {
	value rune
	start int
	end   int
}

func parseGlobalNameListEntries(text string, start int, end int) []inlineGlobalName {
	runes := normalizeGlobalConfigRunes(text, start, end)
	var entries []inlineGlobalName

	for index := 0; index < len(runes); {
		for index < len(runes) && (runes[index].value == ',' || isECMAScriptWhitespace(runes[index].value)) {
			index++
		}
		if index == len(runes) {
			break
		}

		tokenStart := index
		for index < len(runes) && runes[index].value != ',' && !isECMAScriptWhitespace(runes[index].value) {
			index++
		}
		tokenEnd := index

		nameEnd := tokenEnd
		for i := tokenStart; i < tokenEnd; i++ {
			if runes[i].value == ':' {
				nameEnd = i
				break
			}
		}
		if nameEnd == tokenStart {
			continue
		}

		setting := ""
		if nameEnd < tokenEnd {
			settingStart, settingEnd := nameEnd+1, tokenEnd
			for i := settingStart; i < tokenEnd; i++ {
				if runes[i].value == ':' {
					settingEnd = i
					break
				}
			}
			if settingStart < settingEnd {
				setting = text[runes[settingStart].start:runes[settingEnd-1].end]
			}
		}

		nameStartPos := runes[tokenStart].start
		nameEndPos := runes[nameEnd-1].end
		entries = append(entries, inlineGlobalName{
			name:      text[nameStartPos:nameEndPos],
			setting:   setting,
			nameRange: core.NewTextRange(nameStartPos, nameEndPos),
		})
	}

	return entries
}

// normalizeGlobalConfigRunes mirrors @eslint/plugin-kit's parseStringConfig:
// whitespace immediately around ':' and ',' is removed before tokens are
// split. Source positions stay attached so declaration ranges remain exact.
func normalizeGlobalConfigRunes(text string, start int, end int) []globalConfigRune {
	raw := make([]globalConfigRune, 0, end-start)
	for index := start; index < end; {
		r, size := utf8.DecodeRuneInString(text[index:end])
		raw = append(raw, globalConfigRune{value: r, start: index, end: index + size})
		index += size
	}

	normalized := make([]globalConfigRune, 0, len(raw))
	for index := 0; index < len(raw); {
		if !isECMAScriptWhitespace(raw[index].value) {
			normalized = append(normalized, raw[index])
			index++
			continue
		}

		whitespaceEnd := index + 1
		for whitespaceEnd < len(raw) && isECMAScriptWhitespace(raw[whitespaceEnd].value) {
			whitespaceEnd++
		}
		previousIsDelimiter := len(normalized) > 0 && (normalized[len(normalized)-1].value == ':' || normalized[len(normalized)-1].value == ',')
		nextIsDelimiter := whitespaceEnd < len(raw) && (raw[whitespaceEnd].value == ':' || raw[whitespaceEnd].value == ',')
		if !previousIsDelimiter && !nextIsDelimiter {
			normalized = append(normalized, raw[index:whitespaceEnd]...)
		}
		index = whitespaceEnd
	}
	return normalized
}

func findDirectiveJustification(text string, start int, end int) int {
	for index := start; index < end; {
		r, size := utf8.DecodeRuneInString(text[index:end])
		if !isECMAScriptWhitespace(r) {
			index += size
			continue
		}

		hyphenStart := index + size
		afterHyphens := hyphenStart
		for afterHyphens < end && text[afterHyphens] == '-' {
			afterHyphens++
		}
		if afterHyphens-hyphenStart >= 2 && afterHyphens < end {
			next, _ := utf8.DecodeRuneInString(text[afterHyphens:end])
			if isECMAScriptWhitespace(next) {
				return index
			}
		}
		index += size
	}
	return -1
}

func trimECMAScriptWhitespaceRange(text string, start int, end int) (int, int) {
	for start < end {
		r, size := utf8.DecodeRuneInString(text[start:end])
		if !isECMAScriptWhitespace(r) {
			break
		}
		start += size
	}
	for end > start {
		r, size := utf8.DecodeLastRuneInString(text[start:end])
		if !isECMAScriptWhitespace(r) {
			break
		}
		end -= size
	}
	return start, end
}

// ECMAScript's \s set is Unicode Zs plus ASCII spacing/line terminators,
// U+2028/U+2029, and BOM. unicode.IsSpace is not exact: it includes U+0085 and
// excludes BOM. TypeScript's internal stringutil helper also accepts U+0085
// and U+200B, so it cannot model ESLint's JavaScript regexp semantics here.
func isECMAScriptWhitespace(r rune) bool {
	if unicode.Is(unicode.Zs, r) {
		return true
	}
	switch r {
	case '\t', '\v', '\f', '\n', '\r', '\u2028', '\u2029', '\uFEFF':
		return true
	default:
		return false
	}
}
