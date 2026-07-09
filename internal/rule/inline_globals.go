package rule

import (
	"regexp"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
)

// inlineGlobalsKeywords lists the directive keywords that introduce a
// `/* global */` comment. "globals" is checked before "global" since it is
// the longer prefix.
var inlineGlobalsKeywords = [...]string{"globals", "global"}

// ParseInlineGlobals scans `/* global ... */` / `/* globals ... */` block
// comments and returns the set of names they declare (name -> declared).
//
// This is the DisableManager-equivalent preprocessing step for globals: it
// runs once per file, over the same real comment tokens DisableManager
// consumes (not a regex over raw source text, so lookalike text inside a
// string literal or another comment can't be mistaken for a directive), and
// its result is merged with config globals before being handed to rules —
// no rule should parse `/* global */` comments itself.
//
// Mirrors ESLint's inline-config handling: only the setting "off" un-declares
// a name; a bare name or any other setting (e.g. "writable") declares it.
func ParseInlineGlobals(sourceFile *ast.SourceFile, comments []*ast.CommentRange) map[string]bool {
	if sourceFile.Text() == "" || len(comments) == 0 {
		return nil
	}

	text := sourceFile.Text()
	var globals map[string]bool

	for _, comment := range comments {
		// `/* global */` is only recognized as a block comment, matching
		// ESLint's convention (and rslint's prior behavior).
		if comment.Kind != ast.KindMultiLineCommentTrivia {
			continue
		}
		content := strings.TrimSpace(text[comment.Pos()+2 : comment.End()-2])

		rest, ok := matchInlineGlobalsDirective(content)
		if !ok {
			continue
		}

		if globals == nil {
			globals = make(map[string]bool)
		}
		for name, setting := range parseGlobalNameList(rest) {
			globals[name] = setting != "off"
		}
	}

	return globals
}

// matchInlineGlobalsDirective reports whether comment content begins with
// the "global"/"globals" directive keyword (followed by whitespace or
// end-of-string, so e.g. "globalConfig setup" is not mistaken for one) and
// returns the remainder to parse as a name list.
func matchInlineGlobalsDirective(content string) (string, bool) {
	for _, kw := range inlineGlobalsKeywords {
		if !strings.HasPrefix(content, kw) {
			continue
		}
		rest := content[len(kw):]
		if rest == "" || rest[0] == ' ' || rest[0] == '\t' || rest[0] == '\n' || rest[0] == '\r' {
			return strings.TrimSpace(rest), true
		}
	}
	return "", false
}

// MergeGlobals combines config-declared globals with inline `/* global */`
// comment globals into the single set exposed to rules as ctx.Globals —
// mirroring ESLint's addDeclaredGlobals, which layers inline comments on top
// of config (inline wins on conflict). Returns nil if both are empty.
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

// globalSettingSeparatorPattern collapses whitespace around a `:` or `,`
// separator (e.g. "foo : writable ,  bar" -> "foo:writable,bar") before
// splitting — mirrors ESLint's own parseStringConfig, which does the same
// normalization so spaces around a separator don't get mistaken for part of
// the name/setting.
var globalSettingSeparatorPattern = regexp.MustCompile(`\s*([:,])\s*`)

// parseGlobalNameList parses a comma-and/or-whitespace separated
// "name[:setting]" list, e.g. "foo, bar:writable baz:off", returning each
// name mapped to its raw setting ("" when none was given).
func parseGlobalNameList(s string) map[string]string {
	collapsed := globalSettingSeparatorPattern.ReplaceAllString(strings.TrimSpace(s), "$1")

	names := make(map[string]string)
	for _, part := range strings.FieldsFunc(collapsed, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t' || r == '\n' || r == '\r'
	}) {
		name, setting, _ := strings.Cut(part, ":")
		if name != "" {
			names[name] = setting
		}
	}
	return names
}
