package no_warning_comments

import (
	_ "embed"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed no_warning_comments.schema.json
var schemaJSON []byte

const charLimit = 40

// jsWhitespaceClass lists, as Go regexp bracket-expression members, every
// rune ECMAScript's WhiteSpace/LineTerminator grammar treats as `\s` (the
// same set utils.IsTriviaWhitespaceByte/IsTriviaWhitespaceRune recognize —
// keep the two definitions in sync). RE2 has no built-in equivalent: its
// `\s` is ASCII-only, so the upstream `[\s<decoration>]*` prefix class is
// rebuilt explicitly rather than reused.
const jsWhitespaceClass = `\t\n\v\f\r ` +
	`\x{00a0}\x{1680}\x{2000}-\x{200a}\x{2028}\x{2029}\x{202f}\x{205f}\x{3000}\x{feff}`

// NoWarningCommentsRule disallows specified warning terms in comments.
// https://eslint.org/docs/latest/rules/no-warning-comments
var NoWarningCommentsRule = rule.Rule{
	Name:   "no-warning-comments",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)
		escapedDecoration := escapeRegExp(strings.Join(opts.Decoration, ""))
		warningRegExps := make([]*regexp.Regexp, len(opts.Terms))
		for i, term := range opts.Terms {
			warningRegExps[i] = convertToRegExp(term, opts.Location, escapedDecoration)
		}

		text := ctx.SourceFile.Text()
		for _, comment := range ctx.Comments.All() {
			checkComment(ctx, text, comment, opts.Terms, warningRegExps)
		}

		return rule.RuleListeners{}
	},
}

type ruleOptions struct {
	Terms      []string
	Location   string
	Decoration []string
}

func parseOptions(options []any) ruleOptions {
	opts := ruleOptions{
		Terms:    []string{"todo", "fixme", "xxx"},
		Location: "start",
	}
	if len(options) == 0 {
		return opts
	}
	optsMap, _ := options[0].(map[string]any)
	if arr, ok := optsMap["terms"].([]any); ok {
		terms := make([]string, 0, len(arr))
		for _, v := range arr {
			if s, ok := v.(string); ok {
				terms = append(terms, s)
			}
		}
		opts.Terms = terms
	}
	if loc, ok := optsMap["location"].(string); ok {
		opts.Location = loc
	}
	if arr, ok := optsMap["decoration"].([]any); ok {
		decoration := make([]string, 0, len(arr))
		for _, v := range arr {
			if s, ok := v.(string); ok {
				decoration = append(decoration, s)
			}
		}
		opts.Decoration = decoration
	}
	return opts
}

// escapeRegExp mirrors the `escape-string-regexp` npm package upstream uses:
// backslash-escape the regexp metacharacters, and additionally escape `-` so
// a decoration/term string embedded inside a `[...]` character class can
// never be misread as forming a range.
func escapeRegExp(s string) string {
	var b strings.Builder
	for i := range len(s) {
		c := s[i]
		switch c {
		case '|', '\\', '{', '}', '(', ')', '[', ']', '^', '$', '+', '*', '?', '.':
			b.WriteByte('\\')
			b.WriteByte(c)
		case '-':
			b.WriteString(`\-`)
		default:
			b.WriteByte(c)
		}
	}
	return b.String()
}

// isWordChar mirrors JS `\w` (`[A-Za-z0-9_]`), which stays ASCII-only even
// under the `u` flag — the same definition RE2's `\b` uses, so the two
// engines agree on where a word boundary falls.
func isWordChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_'
}

// convertToRegExp ports the upstream rule's convertToRegExp 1:1: build a
// case-insensitive regexp that matches `term` as a whole word in the
// configured location. See the source comment in ESLint's
// lib/rules/no-warning-comments.js for the rationale behind each prefix /
// suffix case.
func convertToRegExp(term, location, escapedDecoration string) *regexp.Regexp {
	escaped := escapeRegExp(term)

	var prefix string
	if location == "start" {
		prefix = `^[` + jsWhitespaceClass + escapedDecoration + `]*`
	} else {
		if r, size := utf8.DecodeRuneInString(term); size > 0 && isWordChar(r) {
			prefix = `\b`
		}
	}

	suffix := ""
	if r, size := utf8.DecodeLastRuneInString(term); size > 0 && isWordChar(r) {
		suffix = `\b`
	}

	return regexp.MustCompile("(?i)" + prefix + escaped + suffix)
}

// selfConfigRegex mirrors upstream's hardcoded self-reference guard: a
// directive comment that mentions this rule's own name is exempt from
// matching, so a comment documenting `no-warning-comments` configuration
// doesn't accidentally trip its own configured terms.
var selfConfigRegex = regexp.MustCompile(`\bno-warning-comments\b`)

// eslintDirectivePattern ports astUtils.ESLINT_DIRECTIVE_PATTERN, used to
// recognize block-comment directives (`/*eslint ...*/`, `/*global ...*/`,
// `/*exported ...*/`) by their leading text.
var eslintDirectivePattern = regexp.MustCompile(`^(?:eslint[- ]|(?:globals?|exported) )`)

// isDirectiveComment ports astUtils.isDirectiveComment: a Line comment is a
// directive if its trimmed text starts with "eslint-"; a Block comment is a
// directive if its trimmed text matches eslintDirectivePattern.
func isDirectiveComment(kind ast.Kind, trimmedValue string) bool {
	switch kind {
	case ast.KindSingleLineCommentTrivia:
		return strings.HasPrefix(trimmedValue, "eslint-")
	case ast.KindMultiLineCommentTrivia:
		return eslintDirectivePattern.MatchString(trimmedValue)
	default:
		return false
	}
}

// commentValue extracts the text between a comment's delimiters — the same
// substring ESLint exposes as `comment.value` — without any trimming.
func commentValue(text string, comment *ast.CommentRange) string {
	switch comment.Kind {
	case ast.KindSingleLineCommentTrivia:
		return text[comment.Pos()+2 : comment.End()]
	case ast.KindMultiLineCommentTrivia:
		return text[comment.Pos()+2 : comment.End()-2]
	default:
		return ""
	}
}

// jsTrim trims ECMAScript WhiteSpace/LineTerminator runes from both ends,
// matching JS `String.prototype.trim()` exactly (Go's strings.TrimSpace
// diverges on U+0085 NEL and doesn't need to agree here).
func jsTrim(s string) string {
	start := utils.SkipLeadingWhitespace(s, 0, len(s))
	end := utils.SkipTrailingWhitespace(s, start, len(s))
	return s[start:end]
}

func checkComment(ctx rule.RuleContext, text string, comment *ast.CommentRange, terms []string, warningRegExps []*regexp.Regexp) {
	value := commentValue(text, comment)

	if isDirectiveComment(comment.Kind, jsTrim(value)) && selfConfigRegex.MatchString(value) {
		return
	}

	for i, re := range warningRegExps {
		if re.MatchString(value) {
			reportMatch(ctx, comment, terms[i], value)
		}
	}
}

// truncateForDisplay ports upstream's display-truncation loop: rebuild the
// trimmed, whitespace-collapsed comment word by word, stopping (with a
// trailing "...") the moment adding the next word would exceed charLimit
// UTF-16 code units — the same unit JS's String#length counts, so an astral
// character (outside the BMP) counts as 2.
func truncateForDisplay(comment string) string {
	var b strings.Builder
	truncated := false
	first := true

	for _, word := range strings.Fields(comment) {
		var candidate string
		if first {
			candidate = word
		} else {
			candidate = b.String() + " " + word
		}

		if utf16Len(candidate) <= charLimit {
			if first {
				b.WriteString(word)
				first = false
			} else {
				b.WriteByte(' ')
				b.WriteString(word)
			}
		} else {
			truncated = true
			break
		}
	}

	if truncated {
		return b.String() + "..."
	}
	return b.String()
}

// utf16Len returns the length of s in UTF-16 code units, matching JS
// String#length (an astral-plane rune counts as a surrogate pair, i.e. 2).
func utf16Len(s string) int {
	n := 0
	for _, r := range s {
		if r > 0xFFFF {
			n += 2
		} else {
			n++
		}
	}
	return n
}

func reportMatch(ctx rule.RuleContext, comment *ast.CommentRange, matchedTerm, value string) {
	display := truncateForDisplay(value)
	ctx.ReportRange(core.NewTextRange(comment.Pos(), comment.End()), rule.RuleMessage{
		Id:          "unexpectedComment",
		Description: fmt.Sprintf("Unexpected '%s' comment: '%s'.", matchedTerm, display),
		Data: map[string]string{
			"matchedTerm": matchedTerm,
			"comment":     display,
		},
	})
}
