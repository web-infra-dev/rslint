package no_inline_comments

import (
	_ "embed"
	"regexp"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
)

//go:embed no_inline_comments.schema.json
var schemaJSON []byte

var unexpectedInlineCommentMessage = rule.RuleMessage{
	Id:          "unexpectedInlineComment",
	Description: "Unexpected comment inline with code.",
}

// eslintDirectivePattern mirrors upstream astUtils.isDirectiveComment's
// ESLINT_DIRECTIVE_PATTERN: a Block comment starting with "eslint-"/"eslint ",
// "global "/"globals ", or "exported " is treated as an ESLint directive, not
// a stray inline comment. Line comments only ever match the "eslint-" arm
// (checked separately below), since upstream applies a narrower `.startsWith`
// test to them instead of this regex.
var eslintDirectivePattern = regexp.MustCompile(`^(?:eslint[- ]|(?:globals?|exported) )`)

type options struct {
	ignorePattern *regexp.Regexp
}

func parseOptions(raw []any) options {
	opts := options{}
	if len(raw) == 0 {
		return opts
	}
	m, _ := raw[0].(map[string]any)
	if pattern, ok := m["ignorePattern"].(string); ok && pattern != "" {
		if re, err := regexp.Compile(pattern); err == nil {
			opts.ignorePattern = re
		}
	}
	return opts
}

// commentInnerText returns the comment's value the way ESTree exposes it on
// Comment.value: the source text with the leading "//" / "/*" and (for block
// comments) trailing "*/" delimiters stripped.
func commentInnerText(sourceText string, comment *ast.CommentRange) string {
	raw := sourceText[comment.Pos():comment.End()]
	switch comment.Kind {
	case ast.KindSingleLineCommentTrivia:
		return strings.TrimPrefix(raw, "//")
	case ast.KindMultiLineCommentTrivia:
		return strings.TrimSuffix(strings.TrimPrefix(raw, "/*"), "*/")
	default:
		return raw
	}
}

// isDirectiveComment mirrors upstream astUtils.isDirectiveComment(node).
func isDirectiveComment(comment *ast.CommentRange, value string) bool {
	trimmed := strings.TrimSpace(value)
	switch comment.Kind {
	case ast.KindSingleLineCommentTrivia:
		return strings.HasPrefix(trimmed, "eslint-")
	case ast.KindMultiLineCommentTrivia:
		return eslintDirectivePattern.MatchString(trimmed)
	default:
		return false
	}
}

// NoInlineCommentsRule disallows comments placed on the same line as code on
// either side of them.
//
// https://eslint.org/docs/latest/rules/no-inline-comments
var NoInlineCommentsRule = rule.Rule{
	Name:   "no-inline-comments",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
		opts := parseOptions(rawOptions)
		sourceFile := ctx.SourceFile
		text := sourceFile.Text()
		lineStarts := scanner.GetECMALineStarts(sourceFile)

		lineText := func(line int) string {
			start := int(lineStarts[line])
			end := len(text)
			if line+1 < len(lineStarts) {
				end = int(lineStarts[line+1])
			}
			return strings.TrimRight(text[start:end], "\r\n")
		}

		check := func(comment *ast.CommentRange) {
			startLine := scanner.ComputeLineOfPosition(lineStarts, comment.Pos())
			endLine := scanner.ComputeLineOfPosition(lineStarts, comment.End())
			startCol := comment.Pos() - int(lineStarts[startLine])
			endCol := comment.End() - int(lineStarts[endLine])

			startLineText := lineText(startLine)
			endLineText := lineText(endLine)
			if startCol > len(startLineText) {
				startCol = len(startLineText)
			}
			if endCol > len(endLineText) {
				endCol = len(endLineText)
			}

			preamble := strings.TrimSpace(startLineText[:startCol])
			postamble := strings.TrimSpace(endLineText[endCol:])
			isPreambleEmpty := preamble == ""
			isPostambleEmpty := postamble == ""

			// Nothing on both sides.
			if isPreambleEmpty && isPostambleEmpty {
				return
			}

			value := commentInnerText(text, comment)

			// Matches the ignore pattern.
			if opts.ignorePattern != nil && opts.ignorePattern.MatchString(value) {
				return
			}

			// JSX exception: a comment that is the sole content of a JSX
			// expression container (`{/* comment */}`, `{// comment\n}`) is
			// not "inline with code" — there is no code, only the braces.
			if (isPreambleEmpty || preamble == "{") && (isPostambleEmpty || postamble == "}") {
				enclosing := ast.GetNodeAtPosition(sourceFile, comment.Pos(), false)
				if enclosing != nil && enclosing.Kind == ast.KindJsxExpression &&
					enclosing.AsJsxExpression().Expression == nil {
					return
				}
			}

			// Don't report ESLint directive comments.
			if isDirectiveComment(comment, value) {
				return
			}

			ctx.ReportRange(comment.TextRange, unexpectedInlineCommentMessage)
		}

		for _, comment := range ctx.Comments.All() {
			check(comment)
		}

		return rule.RuleListeners{}
	},
}
