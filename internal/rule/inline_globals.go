package rule

import (
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

// inlineGlobalsKeywords lists the directive keywords that introduce a
// `/* global */` comment. "globals" is checked before "global" since it is
// the longer prefix.
var inlineGlobalsKeywords = []string{"globals", "global"}

// InlineGlobal describes one name declared by `/* global */` comments.
// Access is the name's final inline setting after all comments are applied.
// NameRanges contains one exact name range per comment that mentions the name,
// in source order. Repeating a name within one comment still contributes only
// its first range, matching ESLint's comment metadata.
type InlineGlobal struct {
	Name       string
	Access     utils.GlobalAccess
	NameRanges []core.TextRange
}

type inlineGlobalName struct {
	name       string
	setting    string
	hasSetting bool
	nameRange  core.TextRange
}

// ParseInlineGlobals returns both the final name -> access map and ordered
// declaration metadata for `/* global ... */` / `/* globals ... */` comments.
// A source-text candidate check keeps the shared comment store lazy unless such a
// directive may be present.
//
// Only real block-comment ranges supplied by the TypeScript scanner are read,
// so lookalike text in strings, templates, regexes, or line comments is ignored.
// Within a comment, duplicate names use the last setting and retain the first
// name range. Across comments, the last setting wins and every comment range is
// preserved. A name written without a setting is readonly; a setting that
// spells none of ESLint's three levels is ignored along with its name range,
// leaving the name's earlier inline setting — or the config or built-in
// setting — in place.
func ParseInlineGlobals(sourceFile *ast.SourceFile, comments *CommentStore) (map[string]utils.GlobalAccess, []InlineGlobal) {
	if sourceFile == nil || sourceFile.Text() == "" || !mayContainDirective(sourceFile.Text(), inlineGlobalsKeywords) {
		return nil, nil
	}

	text := sourceFile.Text()
	sourceComments := comments.All()
	if len(sourceComments) == 0 {
		return nil, nil
	}
	var globals []InlineGlobal
	var entryIndexes map[string]int

	for _, comment := range sourceComments {
		commentNames := parseInlineGlobalComment(text, comment)
		if len(commentNames) == 0 {
			continue
		}

		// ESLint's parseStringConfig returns an object, so a repeated name in
		// one comment has one comment entry: its last setting and first range.
		commentEntries := make([]inlineGlobalName, 0, len(commentNames))
		commentIndexes := make(map[string]int, len(commentNames))
		for _, entry := range commentNames {
			if index, exists := commentIndexes[entry.name]; exists {
				commentEntries[index].setting = entry.setting
				commentEntries[index].hasSetting = entry.hasSetting
				continue
			}
			commentIndexes[entry.name] = len(commentEntries)
			commentEntries = append(commentEntries, entry)
		}

		if entryIndexes == nil {
			entryIndexes = make(map[string]int)
		}
		for _, entry := range commentEntries {
			// ESLint reports the bad directive and moves on, so an unusable
			// setting contributes neither an access level nor a name range.
			access, valid := utils.NormalizeInlineGlobalAccess(entry.setting, entry.hasSetting)
			if !valid {
				continue
			}

			if index, exists := entryIndexes[entry.name]; exists {
				globals[index].Access = access
				globals[index].NameRanges = append(globals[index].NameRanges, entry.nameRange)
				continue
			}
			entryIndexes[entry.name] = len(globals)
			globals = append(globals, InlineGlobal{
				Name:       entry.name,
				Access:     access,
				NameRanges: []core.TextRange{entry.nameRange},
			})
		}
	}

	if len(globals) == 0 {
		return nil, nil
	}
	values := make(map[string]utils.GlobalAccess, len(globals))
	for _, global := range globals {
		values[global.Name] = global.Access
	}

	return values, globals
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

	restStart, ok := matchDirectiveLabelRange(text, contentStart, contentEnd, inlineGlobalsKeywords)
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
	restStart, ok := matchDirectiveLabelRange(content, start, end, inlineGlobalsKeywords)
	if !ok {
		return "", false
	}
	restStart, end = trimECMAScriptWhitespaceRange(content, restStart, end)
	return content[restStart:end], true
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
		for index < len(runes) && (runes[index].value == ',' || ecmascript.IsWhiteSpaceOrLineTerminator(runes[index].value)) {
			index++
		}
		if index == len(runes) {
			break
		}

		tokenStart := index
		for index < len(runes) && runes[index].value != ',' && !ecmascript.IsWhiteSpaceOrLineTerminator(runes[index].value) {
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
		hasSetting := nameEnd < tokenEnd
		if hasSetting {
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
			name:       text[nameStartPos:nameEndPos],
			setting:    setting,
			hasSetting: hasSetting,
			nameRange:  core.NewTextRange(nameStartPos, nameEndPos),
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
		if !ecmascript.IsWhiteSpaceOrLineTerminator(raw[index].value) {
			normalized = append(normalized, raw[index])
			index++
			continue
		}

		whitespaceEnd := index + 1
		for whitespaceEnd < len(raw) && ecmascript.IsWhiteSpaceOrLineTerminator(raw[whitespaceEnd].value) {
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
