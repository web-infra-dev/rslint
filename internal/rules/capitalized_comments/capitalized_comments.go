package capitalized_comments

import (
	_ "embed"
	"strings"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
	"github.com/web-infra-dev/rslint/internal/utils/unicode17"
)

//go:embed capitalized_comments.schema.json
var schemaJSON []byte

// defaultIgnorePattern ports astUtils.COMMENTS_IGNORE_PATTERN: words this rule
// always ignores at the start of a comment, regardless of any configured
// ignorePattern. It also accepts the `rslint` prefix this linter recognizes
// alongside `eslint`, so its own directive comments stay exempt and no fix
// rewrites one into a form DisableManager no longer recognizes.
var defaultIgnorePattern = esregexp.MustCompile(`^\s*(?:eslint|rslint|jshint\s+|jslint\s+|istanbul\s+|globals?\s+|exported\s+|jscs)`, "u")

// maybeURL ports upstream's MAYBE_URL: a comment that starts with something
// that looks like a `scheme://` URL is never reported.
var maybeURL = esregexp.MustCompile(`^\s*[^:/?#\s]+://[^?#]`, "u")

// CapitalizedCommentsRule enforces or disallows capitalization of the first
// letter of a comment.
// https://eslint.org/docs/latest/rules/capitalized-comments
var CapitalizedCommentsRule = rule.Rule{
	Name:   "capitalized-comments",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)
		text := ctx.SourceFile.Text()
		comments := ctx.Comments.All()

		// Only the inline and consecutive checks need the token stream, and
		// only for a comment that gets as far as them, so the walk that
		// answers it happens at most once per file and only on demand.
		var precedingEnds []int
		precedingEnd := func(index int) int {
			if precedingEnds == nil {
				precedingEnds = precedingTokenEnds(ctx.SourceFile, comments)
			}
			return precedingEnds[index]
		}

		for i := range comments {
			checkComment(ctx, text, comments, i, opts, precedingEnd)
		}

		return rule.RuleListeners{}
	},
}

type commentOptions struct {
	IgnorePattern             *esregexp.RegExp
	IgnoreInlineComments      bool
	IgnoreConsecutiveComments bool
}

type ruleOptions struct {
	Capitalize string
	Line       commentOptions
	Block      commentOptions
}

func parseOptions(options []any) ruleOptions {
	opts := ruleOptions{Capitalize: "always"}
	if len(options) > 0 {
		if s, ok := options[0].(string); ok {
			opts.Capitalize = s
		}
	}
	var raw map[string]any
	if len(options) > 1 {
		raw, _ = options[1].(map[string]any)
	}
	opts.Line = normalizeCommentOptions(raw, "line")
	opts.Block = normalizeCommentOptions(raw, "block")
	return opts
}

// normalizeCommentOptions ports getNormalizedOptions: options for `which`
// ("line" or "block") come from raw[which] when present, otherwise from raw
// itself — so a single flat options object applies to both line and block
// comments, and a `{line: {...}, block: {...}}` shape lets each diverge.
func normalizeCommentOptions(raw map[string]any, which string) commentOptions {
	source := raw
	if sub, ok := raw[which].(map[string]any); ok {
		source = sub
	}

	var result commentOptions
	if v, ok := source["ignoreInlineComments"].(bool); ok {
		result.IgnoreInlineComments = v
	}
	if v, ok := source["ignoreConsecutiveComments"].(bool); ok {
		result.IgnoreConsecutiveComments = v
	}
	if pattern, ok := source["ignorePattern"].(string); ok && pattern != "" {
		// Upstream builds the RegExp when the rule is created, so a pattern the
		// engine rejects surfaces as a configuration error. Rules here have no
		// channel for reporting one, so a pattern that fails to compile is
		// dropped and exempts nothing.
		if re, err := esregexp.Compile(`^\s*(?:`+pattern+`)`, "u"); err == nil {
			result.IgnorePattern = re
		}
	}
	return result
}

func checkComment(ctx rule.RuleContext, text string, comments []*ast.CommentRange, index int, opts ruleOptions, precedingEnd func(int) int) {
	comment := comments[index]
	value := utils.CommentValue(text, comment)

	commentOpts := opts.Line
	if comment.Kind == ast.KindMultiLineCommentTrivia {
		commentOpts = opts.Block
	}

	isInline := func() bool { return isInlineComment(ctx.SourceFile, comments, index, precedingEnd(index)) }
	isConsecutive := func() bool { return isConsecutiveComment(comments, index, precedingEnd(index)) }

	if isCommentValid(value, opts.Capitalize, commentOpts, isInline, isConsecutive) {
		return
	}

	messageId := "unexpectedUppercaseComment"
	description := "Comments should not begin with an uppercase character."
	if opts.Capitalize == "always" {
		messageId = "unexpectedLowercaseComment"
		description = "Comments should not begin with a lowercase character."
	}

	textRange := core.NewTextRange(comment.Pos(), comment.End())
	msg := rule.RuleMessage{Id: messageId, Description: description}

	ctx.ReportRangeWithDeferredFixes(textRange, msg, func() []rule.RuleFix {
		return buildFix(comment, value, opts.Capitalize)
	})
}

// isCommentValid ports isCommentValid 1:1: a comment is valid (not reported)
// if it hits any of the early-exit checks below, or if its first letter
// already has the configured case.
func isCommentValid(value string, capitalize string, opts commentOptions, isInline func() bool, isConsecutive func() bool) bool {
	// 1. Default ignore pattern.
	if defaultIgnorePattern.Test(value) {
		return true
	}

	withoutAsterisks := strings.ReplaceAll(value, "*", "")

	// 2. Custom ignore pattern.
	if opts.IgnorePattern != nil && opts.IgnorePattern.Test(withoutAsterisks) {
		return true
	}

	// 3. Inline comments.
	if opts.IgnoreInlineComments && isInline() {
		return true
	}

	// 4. Consecutive comments.
	if opts.IgnoreConsecutiveComments && isConsecutive() {
		return true
	}

	// 5. Does the comment start with a possible URL?
	if maybeURL.Test(withoutAsterisks) {
		return true
	}

	// 6. Is the initial word character a letter?
	wordCharsOnly := stripWhitespace(withoutAsterisks)
	if wordCharsOnly == "" {
		return true
	}
	firstChar, _ := utf8.DecodeRuneInString(wordCharsOnly)
	if !unicode17.IsLetter(firstChar) {
		return true
	}

	// 7. Check the case of the initial word character.
	firstCharStr := string(firstChar)
	isUppercase := firstCharStr != ecmascript.StringToLocaleLowerCase(firstCharStr)
	isLowercase := firstCharStr != ecmascript.StringToLocaleUpperCase(firstCharStr)

	if capitalize == "always" && isLowercase {
		return false
	}
	if capitalize == "never" && isUppercase {
		return false
	}
	return true
}

// stripWhitespace removes every ECMAScript `\s` character from s, matching
// upstream's `.replace(WHITESPACE, "")` with the global `/\s/gu` pattern.
func stripWhitespace(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if !ecmascript.IsWhiteSpaceOrLineTerminator(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// buildFix locates the same character isCommentValid found invalid — the
// first Unicode letter in the raw comment value — and replaces it with its
// upper/lowercase form, offset by the 2-byte comment opener (`//` or `/*`).
func buildFix(comment *ast.CommentRange, value string, capitalize string) []rule.RuleFix {
	idx := -1
	var char rune
	for i, r := range value {
		if unicode17.IsLetter(r) {
			idx = i
			char = r
			break
		}
	}
	if idx < 0 {
		return nil
	}

	charStart := comment.Pos() + 2 + idx
	charEnd := charStart + utf8.RuneLen(char)

	replacement := ecmascript.StringToLocaleLowerCase(string(char))
	if capitalize == "always" {
		replacement = ecmascript.StringToLocaleUpperCase(string(char))
	}

	return []rule.RuleFix{rule.RuleFixReplaceRange(core.NewTextRange(charStart, charEnd), replacement)}
}

// precedingTokenEnds returns, for each comment, the end of the last real token
// that ends before the comment starts, or -1 when no token precedes it.
// Tokens and comments both arrive in source order, so one walk of the token
// tree answers it for every comment at once, where asking per comment would
// read the file again for each one. Walking parser tokens rather than scanning
// keeps a regular-expression literal whole, which a standalone scanner splits
// into division punctuators.
func precedingTokenEnds(sourceFile *ast.SourceFile, comments []*ast.CommentRange) []int {
	ends := make([]int, len(comments))
	cursor := 0
	last := -1
	utils.ForEachToken(sourceFile.AsNode(), func(token *ast.Node) {
		tokenRange := utils.TrimNodeTextRange(sourceFile, token)
		if tokenRange.Pos() >= tokenRange.End() {
			return
		}
		for cursor < len(comments) && comments[cursor].Pos() < tokenRange.End() {
			ends[cursor] = last
			cursor++
		}
		last = tokenRange.End()
	}, sourceFile)
	for ; cursor < len(comments); cursor++ {
		ends[cursor] = last
	}
	return ends
}

// isInlineComment ports isInlineComment: a comment is inline when both the
// nearest preceding and following token-or-comment share its start/end line.
// Only block comments can ever be inline — a line comment always consumes
// the rest of its own line, so there can be no "next token on the same line".
func isInlineComment(sourceFile *ast.SourceFile, comments []*ast.CommentRange, index int, precedingEnd int) bool {
	prevEnd, ok := nearestBeforeEnd(comments, index, precedingEnd)
	if !ok {
		return false
	}
	nextStart, ok := nearestAfterStart(sourceFile, comments, index, comments[index].End())
	if !ok {
		return false
	}
	comment := comments[index]
	return utils.IsSameLine(sourceFile, comment.Pos(), prevEnd) && utils.IsSameLine(sourceFile, comment.End(), nextStart)
}

// isConsecutiveComment ports isConsecutiveComment: true when the nearest
// preceding token-or-comment is itself a comment.
func isConsecutiveComment(comments []*ast.CommentRange, index int, precedingEnd int) bool {
	if index == 0 {
		return false
	}
	// A real token past the previous comment's end sits between the two.
	return precedingEnd <= comments[index-1].End()
}

// nearestBeforeEnd returns the End() of whichever comes later: the nearest
// real token before the comment, or the immediately preceding sibling comment
// (comments[index-1]) when no real token sits between them.
func nearestBeforeEnd(comments []*ast.CommentRange, index int, precedingEnd int) (int, bool) {
	best := precedingEnd
	if index > 0 {
		if prevEnd := comments[index-1].End(); prevEnd <= comments[index].Pos() && prevEnd > best {
			best = prevEnd
		}
	}
	if best < 0 {
		return 0, false
	}
	return best, true
}

// nearestAfterStart is the mirror of nearestBeforeEnd for the position
// immediately following pos.
func nearestAfterStart(sourceFile *ast.SourceFile, comments []*ast.CommentRange, index int, pos int) (int, bool) {
	best := -1
	if tok, ok := utils.TokenAtOrAfter(sourceFile, pos); ok {
		best = tok.Start
	}
	if index+1 < len(comments) {
		if nextStart := comments[index+1].Pos(); nextStart >= pos && (best < 0 || nextStart < best) {
			best = nextStart
		}
	}
	if best < 0 {
		return 0, false
	}
	return best, true
}
