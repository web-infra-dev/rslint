package utils

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
)

// CommentValue extracts the text between a comment's delimiters — the same
// substring ESLint exposes as `comment.value` — without any trimming.
func CommentValue(text string, comment *ast.CommentRange) string {
	switch comment.Kind {
	case ast.KindSingleLineCommentTrivia:
		return text[comment.Pos()+2 : comment.End()]
	case ast.KindMultiLineCommentTrivia:
		// A block comment left unterminated at end of file still parses, and
		// then has no closing delimiter to strip.
		end := comment.End()
		if end-comment.Pos() >= 4 && text[end-2:end] == "*/" {
			end -= 2
		}
		return text[comment.Pos()+2 : end]
	default:
		return ""
	}
}

// eslintDirectivePattern ports astUtils.ESLINT_DIRECTIVE_PATTERN, used to
// recognize block-comment directives (`/*eslint ...*/`, `/*global ...*/`,
// `/*exported ...*/`) by their leading text.
var eslintDirectivePattern = esregexp.MustCompile(`^(?:eslint[- ]|(?:globals?|exported) )`, "u")

// IsDirectiveComment ports astUtils.isDirectiveComment: a Line comment is a
// directive if its trimmed text starts with "eslint-"; a Block comment is a
// directive if its trimmed text matches eslintDirectivePattern. Both forms
// also accept the "rslint-" prefix this linter recognizes alongside "eslint-".
//
// "rslint-" is checked outside eslintDirectivePattern rather than added to it:
// the pattern's other alternatives — "eslint " (space), "global "/"globals ",
// "exported " — have no rslint-prefixed counterpart. rslint-disable and
// rslint-enable, with their -line and -next-line forms, are the whole of what
// this linter recognizes on top of upstream's set, and all of them carry the
// dash.
//
// trimmedValue is what [CommentValue] returns, run through
// ecmascript.StringTrim.
func IsDirectiveComment(kind ast.Kind, trimmedValue string) bool {
	isLintDirective := strings.HasPrefix(trimmedValue, "eslint-") ||
		strings.HasPrefix(trimmedValue, "rslint-")

	switch kind {
	case ast.KindSingleLineCommentTrivia:
		return isLintDirective
	case ast.KindMultiLineCommentTrivia:
		return isLintDirective || eslintDirectivePattern.Test(trimmedValue)
	default:
		return false
	}
}
